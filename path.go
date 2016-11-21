package marching

import (
	"fmt"
	"sort"
	"time"

	"github.com/tidwall/poly"
)

type PathOptions struct{}

// Paths convert the grid into a series of closed paths.
func (grid *Grid) Paths(width, height float64, opts *PathOptions) ([]poly.Polygon, map[int]poly.Point) {
	aboveMap := make(map[int]poly.Point)
	width2f := float64(grid.Width * 2)
	height2f := float64(grid.Height * 2)

	lg := newLineGatherer(int(width2f), int(height2f))
	count := lg.addGrid(grid)
	var paths []poly.Polygon

	if count == 0 {
		// having no lines means that the entire grid is above or below the level.
		// we need to make at least one big path.
		paths = make([]poly.Polygon, 1)
		if lg.above {
			// create one path that encompased the entire rect. clockwise.
			paths[0] = []poly.Point{{0, 0}, {width, 0}, {width, height}, {0, height}, {0, 0}}
		} else {
			// create one path that encompased the entire rect. counter-clockwise.
			//	paths[0] = []Point{{0, 0}, {0, height}, {width, height}, {width, 0}, {0, 0}}
		}
	} else {
		//	opts := &poly.Options{PixelPlane: true}
		paths = make([]poly.Polygon, count)
		var i int
		for _, line := range lg.lines {
			if line.deleted {
				continue
			}
			path := poly.Polygon(make([]poly.Point, len(line.points)))
			for j, point := range line.points {
				path[j] = poly.Point{float64(point.x) / width2f * width, float64(point.y) / height2f * height}
			}
			if line.aboved {
				above := poly.Point{float64(line.above.x) / width2f * width, float64(line.above.y) / height2f * height}
				aboveMap[i] = above
				if false {
					var aboveInside bool
					//	:= above.Inside(path, nil, opts)
					clockwise := pathIsClockwise(path)
					var clockwiseI int
					if clockwise {
						clockwiseI = 1
					}
					var aboveInsideI int
					if aboveInside {
						aboveInsideI = 1
					}
					fmt.Printf("line: %d, clockwise: %v, aboveInside: %v, abovePoint: %v\n", i, clockwiseI, aboveInsideI, above)

					if false {
						if pathIsClockwise(path) {
							//fmt.Printf("line %d: clockwise %v: %v, aboveInside: %v\n", i, 1, above, aboveInside)
							// if above is outside{
							//      reverseWinding(path)
							// }
						} else {
							//fmt.Printf("line %d: clockwise %v: %v, aboveInside: %v\n", i, 0, above, aboveInside)
							// if above is inside{
							//      reverseWinding(path)
							// }
						}
					}
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
	// close paths
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
		lg.above = true
	} else if cell.Case != 15 {
		var leftx = x * 2
		var lefty = y*2 + 1
		var rightx = x*2 + 2
		var righty = y*2 + 1
		var topx = x*2 + 1
		var topy = y * 2
		var bottomx = x*2 + 1
		var bottomy = y*2 + 2
		//var centerx = x*2+1
		//var centery =y*2+1

		switch cell.Case {
		default:
			panic("invalid case")
		case 1:
			lg.addSegment(bottomx, bottomy, leftx, lefty, rightx, topy, true)
		case 2:
			lg.addSegment(rightx, righty, bottomx, bottomy, leftx, topy, true)
		case 3:
			lg.addSegment(rightx, righty, leftx, lefty, topx, topy, true)
		case 4:
			lg.addSegment(topx, topy, rightx, righty, leftx, bottomy, true)
		case 5:
			if !cell.CenterAbove {
				// center below
				// o---------•
				// | /       |
				// |/       /|
				// |       / |
				// •---------o
				lg.addSegment(topx, topy, leftx, lefty, leftx, topy, true)
				lg.addSegment(bottomx, bottomy, rightx, righty, rightx, bottomy, true)
			} else {
				// center above
				// o---------•
				// |       \ |
				// |\       \|
				// | \       |
				// •---------o
				lg.addSegment(topx, topy, rightx, righty, leftx, topy, true)
				lg.addSegment(bottomx, bottomy, leftx, lefty, rightx, bottomy, true)
			}
		case 6:
			lg.addSegment(topx, topy, bottomx, bottomy, leftx, lefty, true)
		case 7:
			lg.addSegment(topx, topy, leftx, lefty, leftx, topy, true)
		case 8:
			lg.addSegment(leftx, lefty, topx, topy, rightx, bottomy, true)
		case 9:
			lg.addSegment(bottomx, bottomy, topx, topy, rightx, righty, true)
		case 10:
			if !cell.CenterAbove {
				// center below
				// •---------o
				// |       \ |
				// |\       \|
				// | \       |
				// o---------•
				lg.addSegment(rightx, righty, topx, topy, rightx, topy, true)
				lg.addSegment(leftx, lefty, bottomx, bottomy, leftx, bottomy, false)
			} else {
				// center above
				// •---------o
				// | /       |
				// |/       /|
				// |       / |
				// o---------•
				lg.addSegment(topx, topy, leftx, lefty, rightx, topy, true)
				lg.addSegment(bottomx, bottomy, rightx, righty, leftx, bottomy, true)
			}
		case 11:
			lg.addSegment(rightx, righty, topx, topy, rightx, topy, true)
		case 12:
			lg.addSegment(leftx, lefty, rightx, righty, bottomx, bottomy, true)
		case 13:
			lg.addSegment(bottomx, bottomy, rightx, righty, rightx, bottomy, true)
		case 14:
			lg.addSegment(leftx, lefty, bottomx, bottomy, leftx, bottomy, true)
		}
	}

	if y == 0 {
		// top
		if cell.Case&0x8 == 0 {
			ax := 2*x - 1
			ay := 0
			bx := ax + 2
			by := ay
			if x == 0 {
				lg.addSegment(ax+2/2, ay+2/2, ax+2/2, ay, 0, 0, false)
				lg.addSegment(ax+2/2, ay, bx, by, 0, 0, false)
			} else {
				lg.addSegment(ax, ay, bx, by, 0, 0, false)
			}
		}
	} else if y == gridHeight-1 {
		// bottom
		if cell.Case&0x2 == 0 {
			ax := 2*(x+1) + 2/2
			ay := gridHeight * 2
			bx := ax - 2
			by := ay
			if x == gridWidth-1 {
				lg.addSegment(ax-2/2, ay-2/2, ax-2/2, ay, 0, 0, false)
				lg.addSegment(ax-2/2, ay, bx, by, 0, 0, false)
			} else {
				lg.addSegment(ax, ay, bx, by, 0, 0, false)
			}
		}
	}
	if x == 0 {
		// left
		if cell.Case&0x1 == 0 {
			ax := 0
			ay := 2*(y+1) + 2/2
			bx := ax
			by := ay - 2
			if y == gridHeight-1 {
				lg.addSegment(ax+2/2, ay-2/2, ax, ay-2/2, 0, 0, false)
				lg.addSegment(ax, ay-2/2, bx, by, 0, 0, false)
			} else {
				lg.addSegment(ax, ay, bx, by, 0, 0, false)
			}
		}
	} else if x == gridWidth-1 {
		// right
		if cell.Case&0x4 == 0 {
			ax := gridWidth * 2
			ay := 2*y - 2/2
			bx := ax
			by := ay + 2
			if y == 0 {
				lg.addSegment(ax-2/2, ay+2/2, ax, ay+2/2, 0, 0, false)
				lg.addSegment(ax, ay+2/2, bx, by, 0, 0, false)
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
func pathIsClockwise(path poly.Polygon) bool {
	var signedArea float64
	for i := 0; i < len(path); i++ {
		if i == len(path)-1 {
			signedArea += (path[i].X*path[0].Y - path[0].X*path[i].Y)
		} else {
			signedArea += (path[i].X*path[i+1].Y - path[i+1].X*path[i].Y)
		}
	}
	return (signedArea / 2) > 0
}
func reverseWinding(path poly.Polygon) {
	for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
		path[i], path[j] = path[j], path[i]
	}
}
