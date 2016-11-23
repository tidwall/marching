// Package marching allows for generating isoline cells from a grid
// of values as specified in https://en.wikipedia.org/wiki/Marching_squares.
package marching

import "sort"

// Cell represents a single isoline square.
type Cell struct {
	// Case field can be a value between 0-15.
	Case byte
	// CenterAbove field indicates that the value in center of the cell
	// is above the level that was passed to NewGrid().
	CenterAbove bool
}

// Grid represents a grid of isoline cells.
type Grid struct {
	// Cells field is an array of isoline cells in pixel coordinates where
	// the cell located at position (0,0) is the top-left cell and is at
	// index zero. The cell at position (Width,Height) is the bottom-right cell
	// and is the last item in the Cell array.
	Cells []Cell
	// Width is the width of the grid. This value is one less than the
	// original width of the values that were passed to NewGrid().
	Width int
	// Height is the height of the grid. This value is one less than the
	// original height of the values that were passed to NewGrid().
	Height int

	values     []float64 // copy of the original values
	level      float64   // contour level
	complexity int       // original complexity
}

// NewGrid generates a grid of isoline cells from a series of values.
// The resulting Grid contains cells with case indexes compared to the
// level param.
// The complexity param can be used to increase or decrease the number
// of grid cells.
// Using a complexity of zero will result in a grid with the default
// number of cells.
func NewGrid(values []float64, width, height int, level float64, complexity int) *Grid {
	if len(values) != width*height {
		panic("number of values are not equal to width multiplied by height")
	}
	if width <= 2 || height <= 2 {
		panic("width or height are not greater than or equal to two")
	}
	var pcmplx uint // positive complexity
	var ncmplx uint // negative complexity
	var gwidth int  // grid width
	var gheight int // grid height
	if complexity > 0 {
		pcmplx = uint(complexity)
		gwidth = (width - 1) << pcmplx
		gheight = (height - 1) << pcmplx
	} else if complexity < 0 {
		ncmplx = uint(1 << uint(complexity*-1))
		gwidth = (width - 1) >> ncmplx
		gheight = (height - 1) >> ncmplx
	} else {
		gwidth = width - 1
		gheight = height - 1
	}
	cells := make([]Cell, gwidth*gheight)
	var vals [4]float64
	var j int
	for y := 0; y < gheight; y++ {
		for x := 0; x < gwidth; x++ {
			var cell Cell
			if complexity == 0 {
				// one-to-one value lookups
				vals[0] = values[(y+0)*width+(x+0)]
				vals[1] = values[(y+0)*width+(x+1)]
				vals[2] = values[(y+1)*width+(x+1)]
				vals[3] = values[(y+1)*width+(x+0)]
			} else if complexity > 0 {
				// using high complexity. location values
				// and use bilinear interpolation to convert
				// to subvalues.
				vals[0] = values[((y>>pcmplx)+0)*width+((x>>pcmplx)+0)]
				vals[1] = values[((y>>pcmplx)+0)*width+((x>>pcmplx)+1)]
				vals[2] = values[((y>>pcmplx)+1)*width+((x>>pcmplx)+1)]
				vals[3] = values[((y>>pcmplx)+1)*width+((x>>pcmplx)+0)]
				rx := x % (1 << pcmplx)
				ry := y % (1 << pcmplx)
				sx := float64(rx) / float64(int(1<<pcmplx))
				sy := float64(ry) / float64(int(1<<pcmplx))
				ex := sx + 1/float64(int(1<<pcmplx))
				ey := sy + 1/float64(int(1<<pcmplx))
				vals = [4]float64{
					bilinearInterpolation(vals, sx, sy),
					bilinearInterpolation(vals, ex, sy),
					bilinearInterpolation(vals, ex, ey),
					bilinearInterpolation(vals, sx, ey),
				}
			} else {
				// using degraded complexity. locate nearest
				// known values. this is a lossy operation and
				// favors faster and simpler geometries over
				// accuracy.
				vals[0] = values[((y+0)<<ncmplx)*width+((x+0)<<ncmplx)]
				vals[1] = values[((y+0)<<ncmplx)*width+((x+1)<<ncmplx)]
				vals[2] = values[((y+1)<<ncmplx)*width+((x+1)<<ncmplx)]
				vals[3] = values[((y+1)<<ncmplx)*width+((x+0)<<ncmplx)]
			}
			if vals[0] < level {
				// top-left
				cell.Case |= 0x8
			}
			if vals[1] < level {
				// top-right
				cell.Case |= 0x4
			}
			if vals[2] < level {
				// bottom-right
				cell.Case |= 0x2
			}
			if vals[3] < level {
				// bottom-left
				cell.Case |= 0x1
			}
			// determine if center of the cell is above the level. this is used
			// to swap saddle points when needed.
			cell.CenterAbove = (vals[0]+vals[1]+vals[2]+vals[3])/4 >= level
			cells[j] = cell
			j++
		}
	}
	return &Grid{
		Cells:      cells,
		Width:      gwidth,
		Height:     gheight,
		values:     append([]float64(nil), values...),
		level:      level,
		complexity: complexity,
	}
}

// bilinearInterpolation return the value that is contained between four
// points. The x and y params must be between 0.0 - 1.0.
func bilinearInterpolation(vals [4]float64, x, y float64) float64 {
	return vals[3]*(1-x)*y + vals[2]*x*y + vals[0]*(1-x)*(1-y) + vals[1]*x*(1-y)
}

// Paths convert the grid into a series of closed paths.
// Each path is a series of XY coordinate points where X is at
// index zero and Y is at index one.
// All paths follow the non-zero winding rule which makes that paths that are
// clockwise are above the level param that was passed to NewGrid(), and paths
// that are counter-clockwise are below the level. In other words the
// clockwise paths are polygons and counter clockwise paths are holes.
// The IsClockwise(path) function can be used to determine the winding
// direction.
func (grid *Grid) Paths(width, height float64) [][][]float64 {
	return grid.pathsWithOptions(width, height, 0, nil)
}

// IsClockwise returns true if the path is clockwise.
func IsClockwise(path [][]float64) bool {
	return polygon(path).isClockwise()
}

// multi is used as a grid cell multiplier for integer space.
// It's nescessary to have enough divisible subspace to identity
// where a point is "above" the requested level and to discover
// connection points without approximation.
// This value must be divisible by 16.
const multi = 16

// polygon is a helpful wrapper around a 3x deep array and provides
// variuos handy functions.
type polygon [][]float64

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

// pathWithOptions return all polygon paths. When aboveMap is not nil the map
// will be fill with points that are above the grid level where the map key is
// the index of the path in the return values.
func (grid *Grid) pathsWithOptions(
	width, height float64,
	simplify int,
	aboveMap map[int][]float64,
) [][][]float64 {
	// widthM and heightM are used to help translate the lineGatherer points to
	// the graphics pixel coordinates space.
	widthM := float64(grid.Width * multi)
	heightM := float64(grid.Height * multi)
	lg := newLineGatherer(int(widthM), int(heightM))

	// add the grid. this will produce all the lines that will in-turn become
	// the return paths. count is the valid non-deleted lines that lineGatherer
	// processed.
	count := lg.addGrid(grid, simplify)

	var paths [][][]float64
	if count == 0 {
		// having no lines means that the entire grid is above or below the level.
		// we need to make at least one big path.
		if lg.above {
			// create one path that encompased the entire rect. clockwise.
			paths = append(paths,
				[][]float64{{0, 0}, {width, 0}, {width, height}, {0, height}, {0, 0}},
			)
		} else {
			// create one path that encompased the entire rect. counter-clockwise.
			//	paths[0] = [][]float64{{0, 0}, {0, height}, {width, height}, {width, 0}, {0, 0}}
		}
	} else {
		// we have lines. let's turn them to valid paths that the caller can use
		paths = make([][][]float64, count)
		var i int
		for _, line := range lg.lines {
			if line.deleted {
				// ignore deleted lines
				continue
			}
			// wrap the path in a polygon type.
			path := polygon(make([][]float64, len(line.points)))
			for j, point := range line.points {
				// add each point and translate to callers coordinates space.
				path[j] = []float64{float64(point.x) / widthM * width, float64(point.y) / heightM * height}
			}
			if line.aboved {
				// the line contains a point that idenities an above level
				// position. this point can be used to determine if the
				// winding direction of the path is correct.
				above := []float64{float64(line.above.x) / widthM * width, float64(line.above.y) / heightM * height}
				if aboveMap != nil {
					// the caller is requesting to store this point for
					// later use.
					aboveMap[i] = above
				}
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
	return paths
}

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

// last returns the last point in in the lin
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
func (lg *lineGatherer) reduceLines(simplify int) int {
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
	// close and count the paths and count
	for i, line := range lg.lines {
		if line.deleted {
			continue
		}
		// make sure that the paths close at exact points
		if line.first() != line.last() {
			// the path does not close
			if line.first() == line.last() {
				// the starting and ending points of the path are very close,
				// just switch assign the first to the last.
				line.points[len(line.points)-1] = line.points[0]
			} else {
				// add a point to the end
				line.points = append(line.points, line.points[0])
			}
		}
		// finally simplify the lines by removing any continuation points,
		// or rather triangles that have no volume.
		if len(line.points) > 2 {
			var points []point
			a, b, c := line.points[0], line.points[1], line.points[2]
			points = append(points, a)
			i := 3
			for {
				area := (a.x*(b.y-c.y) + b.x*(c.y-a.y) + c.x*(a.y-b.y)) / 2
				if area < 0 {
					area *= -1
				}
				if area < multi*4*simplify {
					// do not add b
					if i+1 >= len(line.points) {
						break
					}
					a, b, c = a, line.points[i], line.points[i+1]
					i += 2
				} else {
					// add b
					points = append(points, b)
					if i >= len(line.points) {
						break
					}
					a, b, c = b, c, line.points[i]
					i++
				}
			}
			points = append(points, c)
			line.points = append([]point(nil), points...)
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
	cell Cell,
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

	// connect the edges, if needed
	if y == 0 {
		// top
		if cell.Case&0x8 == 0 {
			ax := x*multi - multi/2
			ay := 0
			bx := ax + multi
			by := ay
			if x == 0 {
				// top-left corner
				lg.addSegment(ax+multi/2, ay+multi/2, ax+multi/2, ay, 0, 0, false)
				lg.addSegment(ax+multi/2, ay, bx, by, 0, 0, false)
			} else {
				lg.addSegment(ax, ay, bx, by, 0, 0, false)
			}
		}
	} else if y == gridHeight-1 {
		// bottom
		if cell.Case&0x2 == 0 {
			ax := x*multi + multi + multi/2
			ay := gridHeight * multi
			bx := ax - multi
			by := ay
			if x == gridWidth-1 {
				// bottom-right corner
				lg.addSegment(ax-multi/2, ay-multi/2, ax-multi/2, ay, 0, 0, false)
				lg.addSegment(ax-multi/2, ay, bx, by, 0, 0, false)
			} else {
				lg.addSegment(ax, ay, bx, by, 0, 0, false)
			}
		}
	}
	if x == 0 {
		// left
		if cell.Case&0x1 == 0 {
			ax := 0
			ay := y*multi + multi + multi/2
			bx := ax
			by := ay - multi
			if y == gridHeight-1 {
				// bottom-left corner
				lg.addSegment(ax+multi/2, ay-multi/2, ax, ay-multi/2, 0, 0, false)
				lg.addSegment(ax, ay-multi/2, bx, by, 0, 0, false)
			} else {
				lg.addSegment(ax, ay, bx, by, 0, 0, false)
			}
		}
	} else if x == gridWidth-1 {
		// right
		if cell.Case&0x4 == 0 {
			ax := gridWidth * multi
			ay := y*multi - multi/2
			bx := ax
			by := ay + multi
			if y == 0 {
				// top-right corner
				lg.addSegment(ax-multi/2, ay+multi/2, ax, ay+multi/2, 0, 0, false)
				lg.addSegment(ax, ay+multi/2, bx, by, 0, 0, false)
			} else {
				lg.addSegment(ax, ay, bx, by, 0, 0, false)
			}
		}
	}
}

// addGrid will add the cells from a grid and reduce the lines
func (lg *lineGatherer) addGrid(grid *Grid, simplify int) int {
	gwidth, gheight := grid.Width, grid.Height
	for y := 0; y < grid.Height; y++ {
		for x := 0; x < grid.Width; x++ {
			cell := grid.Cells[y*grid.Width+x]
			lg.addCell(cell, x, y, lg.width, lg.height, gwidth, gheight)
		}
	}
	return lg.reduceLines(simplify)
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
func (p polygon) pointInside(test []float64) bool {
	var c bool
	for i, j := 0, len(p)-1; i < len(p); j, i = i, i+1 {
		if ((p[i][1] > test[1]) != (p[j][1] > test[1])) &&
			(test[0] < (p[j][0]-p[i][0])*(test[1]-p[i][1])/(p[j][1]-p[i][1])+p[i][0]) {
			c = !c
		}
	}
	return c
}
