package marching

import (
	"sort"
	"time"
)

const multi = 8

type Point struct {
	X, Y float64
}
type Rect struct {
	Min, Max Point
}
type Polygon []Point

func (p Polygon) Rect() Rect {
	var bbox Rect
	for i, p := range p {
		if i == 0 {
			bbox.Min = p
			bbox.Max = p
		} else {
			if p.X < bbox.Min.X {
				bbox.Min.X = p.X
			} else if p.X > bbox.Max.X {
				bbox.Max.X = p.X
			}
			if p.Y < bbox.Min.Y {
				bbox.Min.Y = p.Y
			} else if p.Y > bbox.Max.Y {
				bbox.Max.Y = p.Y
			}
		}
	}
	return bbox
}

type PathOptions struct{}

// Paths convert the grid into a series of closed paths.
func (grid *Grid) Paths(width, height float64, opts *PathOptions) ([]Polygon, map[int]Point) {
	aboveMap := make(map[int]Point)
	width2f := float64(grid.Width * multi)
	height2f := float64(grid.Height * multi)
	lg := newLineGatherer(int(width2f), int(height2f))
	count := lg.addGrid(grid)
	var paths []Polygon

	if count == 0 {
		// having no lines means that the entire grid is above or below the level.
		// we need to make at least one big path.
		paths = make([]Polygon, 1)
		if lg.above {
			// create one path that encompased the entire rect. clockwise.
			paths[0] = []Point{{0, 0}, {width, 0}, {width, height}, {0, height}, {0, 0}}
		} else {
			// create one path that encompased the entire rect. counter-clockwise.
			//	paths[0] = []Point{{0, 0}, {0, height}, {width, height}, {width, 0}, {0, 0}}
		}
	} else {
		paths = make([]Polygon, count)
		var i int
		for _, line := range lg.lines {
			if line.deleted {
				continue
			}
			path := Polygon(make([]Point, len(line.points)))
			for j, point := range line.points {
				path[j] = Point{float64(point.x) / width2f * width, float64(point.y) / height2f * height}
			}
			if line.aboved {
				above := Point{float64(line.above.x) / width2f * width, float64(line.above.y) / height2f * height}
				aboveMap[i] = above
				if path.PointInside(above) != path.IsClockwise() {
					path.ReverseWinding()
				}
			}
			paths[i] = path
			i++
		}
	}
	return paths, aboveMap
}

type lineGatherer struct {
	lines         []line
	width, height int
	above         bool // at least one grid item is above
}

func (lg *lineGatherer) Len() int {
	return len(lg.lines)
}

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

func (lg *lineGatherer) Swap(a, b int) {
	lg.lines[a], lg.lines[b] = lg.lines[b], lg.lines[a]
}

type point struct {
	x, y int
}

type line struct {
	points  []point
	above   point
	aboved  bool
	deleted bool
}

var (
	ctxStart interface{} = "start"
	ctxEnd   interface{} = "end"
)

func (l line) first() point { return l.points[0] }
func (l line) last() point  { return l.points[len(l.points)-1] }
func newLineGatherer(width, height int) *lineGatherer {
	return &lineGatherer{
		width:  width,
		height: height,
	}
}

func (lg *lineGatherer) appendLines(i, j int) {
	if !lg.lines[i].aboved {
		lg.lines[i].aboved = lg.lines[j].aboved
		lg.lines[i].above = lg.lines[j].above
	}
	lg.lines[i].points = append(lg.lines[i].points, lg.lines[j].points[1:]...)
	lg.lines[j].deleted = true
}

func (lg *lineGatherer) addSegment(ax, ay, bx, by int, aboveX, aboveY int, hasAbove bool) {
	lg.lines = append(lg.lines, line{
		points: []point{{ax, ay}, {bx, by}},
		above:  point{aboveX, aboveY},
		aboved: hasAbove,
	})
}

func (lg *lineGatherer) reduceLines() int {
	sort.Sort(lg)
	for {
		var connectionMade bool
		for i := 0; i < len(lg.lines); i++ {
			if lg.lines[i].deleted {
				continue
			}
			for j := 0; j < len(lg.lines); j++ {
				if i == j {
					continue
				}
				if lg.lines[j].deleted {
					continue
				}
				if lg.lines[j].first() == lg.lines[i].last() {
					lg.appendLines(i, j)
					connectionMade = true
					j--
					continue
				}
				if lg.lines[j].last() == lg.lines[i].first() {
					lg.appendLines(j, i)
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
	var count int
	// close and count the paths and count
	for i := 0; i < len(lg.lines); i++ {
		if lg.lines[i].deleted {
			continue
		}
		// make sure that the paths close at exact points
		if lg.lines[i].first() != lg.lines[i].last() {
			// the path does not close
			if lg.lines[i].first() == lg.lines[i].last() {
				// the starting and ending points of the path are very close,
				// just switch assign the first to the last.
				lg.lines[i].points[len(lg.lines[i].points)-1] = lg.lines[i].points[0]
			} else {
				// add a point to the end
				lg.lines[i].points = append(lg.lines[i].points, lg.lines[i].points[0])
			}
		}
		count++
	}
	println(">>", count)
	return count
}

type lineKey struct {
	x, y float64
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

		switch cell.Case {
		default:
			panic("invalid case")
		case 1:
			// o---------o
			// |         |
			// |\        |
			// | \       |
			// •---------o
			lg.addSegment(bottomx, bottomy, leftx, lefty, rightx-1, topy+1, true)
		case 2:
			// o---------o
			// |         |
			// |        /|
			// |       / |
			// o---------•
			lg.addSegment(rightx, righty, bottomx, bottomy, leftx+1, topy+1, true)
		case 3:
			// o---------o
			// |         |
			// |---------|
			// |         |
			// •---------•
			lg.addSegment(rightx, righty, leftx, lefty, topx, topy+1, true)
		case 4:
			// o---------•
			// |       \ |
			// |        \|
			// |         |
			// o---------o
			lg.addSegment(topx, topy, rightx, righty, leftx+1, bottomy-1, true)
		case 5:
			if !cell.CenterAbove {
				// center below
				// o---------•
				// | /       |
				// |/       /|
				// |       / |
				// •---------o
				lg.addSegment(topx, topy, leftx, lefty, leftx+1, topy+1, true)
				lg.addSegment(bottomx, bottomy, rightx, righty, rightx-1, bottomy-1, true)
			} else {
				// center above
				// o---------•
				// |       \ |
				// |\       \|
				// | \       |
				// •---------o
				lg.addSegment(topx, topy, rightx, righty, leftx+1, topy+1, true)
				lg.addSegment(bottomx, bottomy, leftx, lefty, rightx-1, bottomy-1, true)
			}
		case 6:
			// o---------•
			// |    |    |
			// |    |    |
			// |    |    |
			// o---------•
			lg.addSegment(topx, topy, bottomx, bottomy, leftx+1, lefty, true)
		case 7:
			// o---------•
			// | /       |
			// |/        |
			// |         |
			// •---------•
			lg.addSegment(topx, topy, leftx, lefty, leftx+1, topy+1, true)
		case 8:
			// •---------o
			// | /       |
			// |/        |
			// |         |
			// o---------o
			lg.addSegment(leftx, lefty, topx, topy, rightx-1, bottomy-1, true)
		case 9:
			// •---------o
			// |    |    |
			// |    |    |
			// |    |    |
			// •---------o
			lg.addSegment(bottomx, bottomy, topx, topy, rightx-1, righty, true)
		case 10:
			if !cell.CenterAbove {
				// center below
				// •---------o
				// |       \ |
				// |\       \|
				// | \       |
				// o---------•
				lg.addSegment(rightx, righty, topx, topy, rightx-1, topy+1, true)
				lg.addSegment(leftx, lefty, bottomx, bottomy, leftx+1, bottomy-1, false)
			} else {
				// center above
				// •---------o
				// | /       |
				// |/       /|
				// |       / |
				// o---------•
				lg.addSegment(topx, topy, leftx, lefty, rightx-1, topy+1, true)
				lg.addSegment(bottomx, bottomy, rightx, righty, leftx+1, bottomy-1, true)
			}
		case 11:
			// •---------o
			// |       \ |
			// |        \|
			// |         |
			// •---------•
			lg.addSegment(rightx, righty, topx, topy, rightx-1, topy+1, true)
		case 12:
			// •---------•
			// |         |
			// |---------|
			// |         |
			// o---------o
			lg.addSegment(leftx, lefty, rightx, righty, bottomx, bottomy-1, true)
		case 13:
			// •---------•
			// |         |
			// |        /|
			// |       / |
			// •---------o
			lg.addSegment(bottomx, bottomy, rightx, righty, rightx-1, bottomy-1, true)
		case 14:
			// •---------•
			// |         |
			// |\        |
			// | \       |
			// o---------•
			lg.addSegment(leftx, lefty, bottomx, bottomy, leftx+1, bottomy-1, true)
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
				lg.addSegment(ax-multi/2, ay+multi/2, ax, ay+multi/2, 0, 0, false)
				lg.addSegment(ax, ay+multi/2, bx, by, 0, 0, false)
			} else {
				lg.addSegment(ax, ay, bx, by, 0, 0, false)
			}
		}
	}
}

func (lg *lineGatherer) addGrid(grid *Grid) int {
	start := time.Now()
	gwidth, gheight := grid.Width, grid.Height

	for y := 0; y < grid.Height; y++ {
		for x := 0; x < grid.Width; x++ {
			cell := grid.Cells[y*grid.Width+x]
			lg.addCell(cell, x, y, lg.width, lg.height, gwidth, gheight)
		}
	}
	println("** addCells:", time.Now().Sub(start).String())
	start = time.Now()
	count := lg.reduceLines()
	println("** reduceLines:", time.Now().Sub(start).String())
	return count
}

// http://stackoverflow.com/a/1165943/424124
func (p Polygon) IsClockwise() bool {
	var signedArea float64
	for i := 0; i < len(p); i++ {
		if i == len(p)-1 {
			signedArea += (p[i].X*p[0].Y - p[0].X*p[i].Y)
		} else {
			signedArea += (p[i].X*p[i+1].Y - p[i+1].X*p[i].Y)
		}
	}
	return (signedArea / 2) > 0
}

func (p Polygon) ReverseWinding() {
	for i, j := 0, len(p)-1; i < j; i, j = i+1, j-1 {
		p[i], p[j] = p[j], p[i]
	}
}

func (p Polygon) PointInside(test Point) bool {
	var c bool
	for i, j := 0, len(p)-1; i < len(p); j, i = i, i+1 {
		if ((p[i].Y > test.Y) != (p[j].Y > test.Y)) &&
			(test.X < (p[j].X-p[i].X)*(test.Y-p[i].Y)/(p[j].Y-p[i].Y)+p[i].X) {
			c = !c
		}
	}
	return c
}
