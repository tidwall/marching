package marching

import (
	"math"
	"sort"
)

// Paths generates linestring paths from a sample of values.
func Paths(values []float64, width, height int, level float64) [][][2]float64 {
	cells := makeCells(values, width, height, level)
	paths := makePaths(cells, width, height, level)
	interpolatePaths(paths, values, level, width, height)
	offsetPaths(paths, -0.5, -0.5)
	return paths
}

// multi is used as a grid cell multiplier for integer space.
// It's nescessary to have enough divisible subspace to identity
// where a point is "above" the requested level and to discover
// connection points without approximation.
// This value must be divisible by 16.
const multi = 16

// Cell represents a single isoline square.
type cellT struct {
	// Case field can be a value between 0-15.
	Case byte
	// CenterAbove field indicates that the value in center of the cell
	// is above the level that was passed to makeCells().
	CenterAbove bool
	// Values are the four corner values
	Values [4]float64
}

// makeCells ...
func makeCells(values []float64, width, height int, level float64) []cellT {
	if len(values) != width*height {
		panic("number of values are not equal to width multiplied by height")
	}
	if width <= 2 || height <= 2 {
		panic("width or height are not greater than or equal to two")
	}
	gwidth := width - 1   // grid width
	gheight := height - 1 // grid height
	cells := make([]cellT, gwidth*gheight)
	var j int
	for y := 0; y < gheight; y++ {
		for x := 0; x < gwidth; x++ {
			var cell cellT
			// one-to-one value lookups
			cell.Values[0] = values[(y+0)*width+(x+0)]
			cell.Values[1] = values[(y+0)*width+(x+1)]
			cell.Values[2] = values[(y+1)*width+(x+0)]
			cell.Values[3] = values[(y+1)*width+(x+1)]
			if cell.Values[0] < level {
				// top-left
				cell.Case |= 0x8
			}
			if cell.Values[1] < level {
				// top-right
				cell.Case |= 0x4
			}
			if cell.Values[2] < level {
				// bottom-left
				cell.Case |= 0x1
			}
			if cell.Values[3] < level {
				// bottom-right
				cell.Case |= 0x2
			}
			// determine if center of the cell is above the level. this is used
			// to swap saddle points when needed.
			cell.CenterAbove = (cell.Values[0]+cell.Values[1]+
				cell.Values[2]+cell.Values[3])/4 >= level
			cells[j] = cell
			j++
			//fmt.Printf("%dx%d\n", x, y)
		}
	}
	return cells
}

// polygon is a helpful wrapper around a 3x deep array and provides
// variuos handy functions.
type polygon [][2]float64

// rect returns the outmost boundaries of a polygon.
func (p polygon) rect() (min, max []float64) {
	if len(p) > 0 {
		min = []float64{p[0][0], p[0][1]}
		max = []float64{p[0][0], p[0][1]}
		for i := 1; i < len(p); i++ {
			if p[i][0] < min[0] {
				min[0] = p[i][0]
			} else if p[i][0] > max[0] {
				max[0] = p[i][0]
			}
			if p[i][1] < min[1] {
				min[1] = p[i][1]
			} else if p[i][1] > max[1] {
				max[1] = p[i][1]
			}
		}
	}
	return
}

// http://stackoverflow.com/a/1165943/424124
func (p polygon) isClockwise() bool {
	var signedArea float64
	for i := 0; i < len(p); i++ {
		if i == len(p)-1 {
			signedArea += (p[i][0]*p[0][1] - p[0][0]*p[i][1])
		} else {
			signedArea += (p[i][0]*p[i+1][1] - p[i+1][0]*p[i][1])
		}
	}
	return (signedArea / 2) > 0
}

// reverseWinding reverses the winding of a path
func (p polygon) reverseWinding() {
	for i, j := 0, len(p)-1; i < j; i, j = i+1, j-1 {
		p[i], p[j] = p[j], p[i]
	}
}

// pointInside tests if a point is inside a polygon
func (p polygon) pointInside(test [2]float64) bool {
	var c bool
	for i, j := 0, len(p)-1; i < len(p); j, i = i, i+1 {
		if ((p[i][1] > test[1]) != (p[j][1] > test[1])) &&
			(test[0] < (p[j][0]-p[i][0])*(test[1]-p[i][1])/(p[j][1]-p[i][1])+p[i][0]) {
			c = !c
		}
	}
	return c
}

// makePaths ...
func makePaths(cells []cellT, width, height int, level float64) [][][2]float64 {
	width--
	height--
	// widthM and heightM are used to help translate the lineGatherer points to
	// the graphics pixel coordinates space.
	widthM := float64(width * multi)
	heightM := float64(width * multi)
	lg := newLineGatherer(int(widthM), int(heightM))

	// add the grid. this will produce all the lines that will in-turn become
	// the return paths. count is the valid non-deleted lines that lineGatherer
	// processed.
	count := lg.addCells(cells, width, height, level)
	var paths [][][2]float64
	if count == 0 {
		// having no lines means that the entire grid is above or below the level.
		// we need to make at least one big path.
		if lg.above {
			// create one path that encompased the entire rect. clockwise.
			paths = append(paths,
				[][2]float64{
					{0, 0},
					{float64(width), 0},
					{float64(width), float64(height)},
					{0, float64(height)},
					{0, 0}},
			)
		} else {
			// create one path that encompased the entire rect. counter-clockwise.
			//	paths[0] = [][]float64{{0, 0}, {0, height}, {width, height}, {width, 0}, {0, 0}}
		}
	} else {
		// we have lines. let's turn them to valid paths that the caller can use
		paths = make([][][2]float64, count)
		var i int
		for _, line := range lg.lines {
			if line.deleted {
				// ignore deleted lines
				continue
			}
			// wrap the path in a polygon type.
			path := polygon(make([][2]float64, len(line.points)))
			for j, point := range line.points {
				// add each point and translate to callers coordinates space.
				path[j] = [2]float64{float64(point.x) / widthM * float64(width), float64(point.y) / heightM * float64(height)}
			}
			if line.aboved {
				// the line contains a point that idenities an above level
				// position. this point can be used to determine if the
				// winding direction of the path is correct.
				above := [2]float64{float64(line.above.x) / widthM * float64(width), float64(line.above.y) / heightM * float64(height)}
				// if aboveMap != nil {
				// 	// the caller is requesting to store this point for
				// 	// later use.
				// 	aboveMap[i] = above
				// }
				if path.pointInside(above) != path.isClockwise() {
					// the point must be inside and the path must be clockwise,
					// or the point must be outside and path must be
					// counter-clockwise. let's reverse the winding of the
					// path to ensure that this is the case.
					path.reverseWinding()
				}
			}
			// unwrap and assign to return array
			paths[i] = path
			i++
		}
	}
	// offset the paths by half pixel to align with in the input width/height
	for i := 0; i < len(paths); i++ {
		for j := 0; j < len(paths[i]); j++ {
			paths[i][j][0] = paths[i][j][0] + 0.5
			paths[i][j][1] = paths[i][j][1] + 0.5
		}
	}
	return paths
}

//
// lineGatherer is responsible for converting grid cells into lines that can
// later be used for generating closed paths.
type lineGatherer struct {
	lines         []line // all lines
	width, height int    // line boundary
	above         bool   // at least one grid item is above
}

// Len implements sort.Interface
func (lg *lineGatherer) Len() int {
	return len(lg.lines)
}

// Less implements sort.Interface
func (lg *lineGatherer) Less(a, b int) bool {
	pointA := lg.lines[a].last()
	pointB := lg.lines[b].last()
	if pointA.y < pointB.y {
		return true
	}
	if pointA.x < pointB.x {
		return true
	}
	pointA = lg.lines[a].first()
	pointB = lg.lines[b].first()
	if pointA.y < pointB.y {
		return true
	}
	if pointA.x < pointB.x {
		return true
	}
	return false
}

// Swap implements sort.Interface
func (lg *lineGatherer) Swap(a, b int) {
	lg.lines[a], lg.lines[b] = lg.lines[b], lg.lines[a]
}

// point represents a simple XY point used by line.
type point struct{ x, y int }

// line represents a series of points used by lineGatherer.
type line struct {
	points  []point // series of points
	above   point   // a point that is above the grid level
	aboved  bool    // is the above point usable
	deleted bool    // is the line marked for delete
}

// first returns the first point in the line
func (l line) first() point { return l.points[0] }

// last returns the last point in in the line
func (l line) last() point { return l.points[len(l.points)-1] }

// newLineGatherer creates a lineGatherer object.
// the width and height must be divisible by 16
func newLineGatherer(width, height int) *lineGatherer {
	if width%16 != 0 || height%16 != 0 {
		panic("width and height must be divisible by 16")
	}
	return &lineGatherer{
		width:  width,
		height: height,
	}
}

// joinLines will combine append the line at index j to the line at index i.
func (lg *lineGatherer) joinLines(i, j int) {
	if !lg.lines[i].aboved {
		// if line[i] does not have an above point, then it will use the
		// the above point of line[j]
		lg.lines[i].aboved = lg.lines[j].aboved
		lg.lines[i].above = lg.lines[j].above
	}
	lg.lines[i].points = append(lg.lines[i].points, lg.lines[j].points[1:]...)
	// mark line[j] for deletion
	lg.lines[j].deleted = true
}

// addSegment will add a two point line segment to the series of lines.
func (lg *lineGatherer) addSegment(ax, ay, bx, by int, aboveX, aboveY int, hasAbove bool) {
	lg.lines = append(lg.lines, line{
		points: []point{{ax, ay}, {bx, by}},
		above:  point{aboveX, aboveY},
		aboved: hasAbove,
	})
}

// reduceLines will take all line segments and generate closed paths.
// The return value is the final number of non-deleted lines.
func (lg *lineGatherer) reduceLines(height float64) int {
	// sort the lines by Y then X
	sort.Sort(lg)
	for {
		var connectionMade bool
		for i := 0; i < len(lg.lines); i++ {
			if lg.lines[i].deleted {
				// ignore deleted lines
				continue
			}
			for j := 0; j < len(lg.lines); j++ {
				if i == j {
					// ignore same lines
					continue
				}
				if lg.lines[j].deleted {
					// ignore deleted lines
					continue
				}
				if lg.lines[j].first() == lg.lines[i].last() {
					// join line[j] to line[i]
					lg.joinLines(i, j)
					connectionMade = true
					j--
					continue
				}
				if lg.lines[j].last() == lg.lines[i].first() {
					// join line[i] to line[j]
					lg.joinLines(j, i)
					connectionMade = true
					i--
					break
				}
				if lg.lines[j].last() == lg.lines[i].last() ||
					lg.lines[j].first() == lg.lines[i].first() {
					// reverse the line and try again
					s := lg.lines[j].points
					for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
						s[i], s[j] = s[j], s[i]
					}
					connectionMade = true
					j--
					continue
				}
			}
		}
		if !connectionMade {
			break
		}
	}
	// count the paths
	for i, line := range lg.lines {
		if line.deleted {
			continue
		}
		lg.lines[i] = line
		if len(lg.lines[i].points) <= 2 {
			lg.lines[i].deleted = true
		}
	}
	var count int
	for _, line := range lg.lines {
		if line.deleted {
			continue
		}
		if len(line.points) < 3 {
			line.deleted = true
		}
		count++
	}
	return count
}

func (lg *lineGatherer) addCell(
	cell cellT,
	x, y, width, height int,
	gridWidth, gridHeight int,
) {
	if cell.Case == 0 {
		// o---------o
		// |         |
		// |         |
		// |         |
		// o---------o
		// all is above
		lg.above = true
	} else if cell.Case == 15 {
		// •---------•
		// |         |
		// |         |
		// |         |
		// •---------•
	} else {
		var leftx = x * multi
		var lefty = y*multi + multi/2
		var rightx = x*multi + multi
		var righty = y*multi + multi/2
		var topx = x*multi + multi/2
		var topy = y * multi
		var bottomx = x*multi + multi/2
		var bottomy = y*multi + multi
		var one = multi / 8
		switch cell.Case {
		default:
			panic("invalid case")
		case 1:
			// o---------o
			// |         |
			// |\        |
			// | \       |
			// •---------o
			lg.addSegment(bottomx, bottomy, leftx, lefty, rightx-one, topy+one, true)
		case 2:
			// o---------o
			// |         |
			// |        /|
			// |       / |
			// o---------•
			lg.addSegment(rightx, righty, bottomx, bottomy, leftx+one, topy+one, true)
		case 3:
			// o---------o
			// |         |
			// |---------|
			// |         |
			// •---------•
			lg.addSegment(rightx, righty, leftx, lefty, topx, topy+one, true)
		case 4:
			// o---------•
			// |       \ |
			// |        \|
			// |         |
			// o---------o
			lg.addSegment(topx, topy, rightx, righty, leftx+one, bottomy-one, true)
		case 5:
			if !cell.CenterAbove {
				// center below
				// o---------•
				// | /       |
				// |/       /|
				// |       / |
				// •---------o
				lg.addSegment(topx, topy, leftx, lefty, leftx+one, topy+one, true)
				lg.addSegment(bottomx, bottomy, rightx, righty, rightx-one, bottomy-one, true)
			} else {
				// center above
				// o---------•
				// |       \ |
				// |\       \|
				// | \       |
				// •---------o
				lg.addSegment(topx, topy, rightx, righty, leftx+one, topy+one, true)
				lg.addSegment(bottomx, bottomy, leftx, lefty, rightx-one, bottomy-one, true)
			}
		case 6:
			// o---------•
			// |    |    |
			// |    |    |
			// |    |    |
			// o---------•
			lg.addSegment(topx, topy, bottomx, bottomy, leftx+one, lefty, true)
		case 7:
			// o---------•
			// | /       |
			// |/        |
			// |         |
			// •---------•
			lg.addSegment(topx, topy, leftx, lefty, leftx+one, topy+one, true)
		case 8:
			// •---------o
			// | /       |
			// |/        |
			// |         |
			// o---------o
			lg.addSegment(leftx, lefty, topx, topy, rightx-one, bottomy-one, true)
		case 9:
			// •---------o
			// |    |    |
			// |    |    |
			// |    |    |
			// •---------o
			lg.addSegment(bottomx, bottomy, topx, topy, rightx-one, righty, true)
		case 10:
			if !cell.CenterAbove {
				// center below
				// •---------o
				// |       \ |
				// |\       \|
				// | \       |
				// o---------•
				lg.addSegment(rightx, righty, topx, topy, rightx-one, topy+one, true)
				lg.addSegment(leftx, lefty, bottomx, bottomy, leftx+one, bottomy-one, false)
			} else {
				// center above
				// •---------o
				// | /       |
				// |/       /|
				// |       / |
				// o---------•
				lg.addSegment(topx, topy, leftx, lefty, rightx-one, topy+one, true)
				lg.addSegment(bottomx, bottomy, rightx, righty, leftx+one, bottomy-one, true)
			}
		case 11:
			// •---------o
			// |       \ |
			// |        \|
			// |         |
			// •---------•
			lg.addSegment(rightx, righty, topx, topy, rightx-one, topy+one, true)
		case 12:
			// •---------•
			// |         |
			// |---------|
			// |         |
			// o---------o
			lg.addSegment(leftx, lefty, rightx, righty, bottomx, bottomy-one, true)
		case 13:
			// •---------•
			// |         |
			// |        /|
			// |       / |
			// •---------o
			lg.addSegment(bottomx, bottomy, rightx, righty, rightx-one, bottomy-one, true)
		case 14:
			// •---------•
			// |         |
			// |\        |
			// | \       |
			// o---------•
			lg.addSegment(leftx, lefty, bottomx, bottomy, leftx+one, bottomy-one, true)
		}
	}
}

// addGrid will add the cells from a grid and reduce the lines
func (lg *lineGatherer) addCells(cells []cellT, width, height int, level float64) int {
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			cell := cells[y*width+x]
			lg.addCell(cell, x, y, lg.width, lg.height, width, height)
		}
	}
	return lg.reduceLines(level)
}

func interpolatePaths(paths [][][2]float64, values []float64, level float64, width, height int) {
	for i := 0; i < len(paths); i++ {
		for j := 0; j < len(paths[i]); j++ {
			p := paths[i][j]
			p[0] = round(p[0], 1)
			p[1] = round(p[1], 1)
			if math.Floor(p[1]) == p[1] {
				v1 := values[int(p[1]-1)*width+int(p[0])]
				v2 := values[int(p[1])*width+int(p[0])]
				q := (1.0 - ((level - v2) / (v1 - v2)))
				p[1] = p[1] - 0.5 + q
			}
			if math.Floor(p[0]) == p[0] {
				v1 := values[int(p[1])*width+int(p[0]-1)]
				v2 := values[int(p[1])*width+int(p[0])]
				q := (1.0 - ((level - v2) / (v1 - v2)))
				p[0] = p[0] - 0.5 + q
			}
			paths[i][j] = p
		}
	}
}

func round(f float64, places int) float64 {
	shift := math.Pow(10, float64(places))
	return math.Floor((f*shift)+.5) / shift
}

func offsetPaths(paths [][][2]float64, xoffset, yoffset float64) {
	for i := 0; i < len(paths); i++ {
		for j := 0; j < len(paths[i]); j++ {
			paths[i][j][0] += xoffset
			paths[i][j][1] += yoffset
		}
	}
}
