package marching

import (
	"math"
	"time"
)

type Point struct {
	X, Y float64
}
type PathOptions struct{}

// Paths convert the grid into a series of closed paths.
func (grid *Grid) Paths(width, height float64, opts *PathOptions) [][]Point {
	lg := newLineGatherer(width, height)
	lg.addGrid(grid, width, height)
	paths := make([][]Point, len(lg.lines))
	for i, line := range lg.lines {
		path := make([]Point, len(line.points))
		for j, point := range line.points {
			path[j] = Point{point.x, point.y}
		}
		paths[i] = path
	}
	return paths
}

type lineGatherer struct {
	lines []*line
}
type point struct {
	x, y float64
}

type line struct {
	points []*point
}

const maxRelativeError = 0.00001

func (p1 *point) veryClose(p2 *point) bool {
	if *p1 == *p2 {
		return true
	}
	if math.Abs((p1.x-p2.x)/p2.x) > maxRelativeError {
		return false
	}
	if math.Abs((p1.y-p2.y)/p2.y) > maxRelativeError {
		return false
	}
	return true
}

var (
	ctxStart interface{} = "start"
	ctxEnd   interface{} = "end"
)

func (l *line) first() *point { return l.points[0] }
func (l *line) last() *point  { return l.points[len(l.points)-1] }
func newLineGatherer(width, height float64) *lineGatherer {
	return &lineGatherer{}
}

func (lg *lineGatherer) appendLines(i, j int) {
	lg.lines[i].points = append(lg.lines[i].points, lg.lines[j].points[1:]...)
	lg.lines = append(lg.lines[:j], lg.lines[j+1:]...)
}

func (lg *lineGatherer) addSegment(ax, ay, bx, by float64) {
	/*
		pa := Point{ax, ay}
		pb := Point{bx, by}
		if false {
			mina, maxa := pa.expandBounds(pb)
			minb, maxb := pb.expandBounds(pa)

			mina, maxa = mina, maxa
			minb, maxb = minb, maxb
			for i := range lg.lines {
				if lg.lines[i].first().veryClose(pa) {
					lg.lines[i].points = append([]Point{pb}, lg.lines[i].points...)
					return
				}
				if lg.lines[i].first().veryClose(pb) {
					lg.lines[i].points = append([]Point{pa}, lg.lines[i].points...)
					return
				}
				if lg.lines[i].last().veryClose(pa) {
					lg.lines[i].points = append(lg.lines[i].points, pb)
					return
				}
				if lg.lines[i].last().veryClose(pb) {
					lg.lines[i].points = append(lg.lines[i].points, pa)
					return
				}
			}
		}
	*/
	lg.lines = append(lg.lines, &line{[]*point{&point{ax, ay}, &point{bx, by}}})
}

func (lg *lineGatherer) reduceLines() {
	for {
		var connectionMade bool
		for i := 0; i < len(lg.lines); i++ {
			for j := 0; j < len(lg.lines); j++ {
				if i == j {
					continue
				}
				if lg.lines[j].first().veryClose(lg.lines[i].last()) {
					lg.appendLines(i, j)
					connectionMade = true
					j--
					continue
				}
				if lg.lines[j].last().veryClose(lg.lines[i].first()) {
					lg.appendLines(j, i)
					connectionMade = true
					i--
					break
				}
				if lg.lines[j].last().veryClose(lg.lines[i].last()) ||
					lg.lines[j].first().veryClose(lg.lines[i].first()) {
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
	// close paths
	for i := 0; i < len(lg.lines); i++ {
		// make sure that the paths close at exact points
		if lg.lines[i].first() != lg.lines[i].last() {
			// the path does not close
			if lg.lines[i].first().veryClose(lg.lines[i].last()) {
				// the starting and ending points of the path are very close,
				// just switch assign the first to the last.
				lg.lines[i].points[len(lg.lines[i].points)-1] = lg.lines[i].points[0]
			} else {
				// add a point to the end
				lg.lines[i].points = append(lg.lines[i].points, lg.lines[i].points[0])
			}
		}
	}
	println(">>", len(lg.lines))
}

type lineKey struct {
	x, y float64
}

func (lg *lineGatherer) addCell(
	cell Cell,
	x, y, width, height float64,
	gridWidth, gridHeight float64,
) {
	var offsetx, offsety float64
	var cellw = width / gridWidth
	var cellh = height / gridHeight
	var leftx = offsetx + cellw*x
	var lefty = offsety + cellh*y + cellh*0.5

	var rightx = offsetx + cellw*x + cellw
	var righty = offsety + cellh*y + cellh*0.5
	var topx = offsetx + cellw*x + cellw*0.5
	var topy = offsety + cellh*y
	var bottomx = offsetx + cellw*x + cellw*0.5
	var bottomy = offsety + cellh*y + cellh
	//var centerx = offsetx + cellw*float64(x) + cellw*0.5
	//var centery = offsety + cellh*float64(y) + cellh*0.5
	if cell.Case != 0 && cell.Case != 15 {
		switch cell.Case {
		default:
			panic("invalid case")
		case 1:
			lg.addSegment(bottomx, bottomy, leftx, lefty)
		case 2:
			lg.addSegment(rightx, righty, bottomx, bottomy)
		case 3:
			lg.addSegment(rightx, righty, leftx, lefty)
		case 4:
			lg.addSegment(topx, topy, rightx, righty)
		case 5:
			if !cell.CenterAbove {
				lg.addSegment(topx, topy, rightx, righty)
				lg.addSegment(bottomx, bottomy, leftx, lefty)
			} else {
				lg.addSegment(bottomx, bottomy, rightx, righty)
				lg.addSegment(topx, topy, leftx, lefty)
			}
		case 6:
			lg.addSegment(topx, topy, bottomx, bottomy)
		case 7:
			lg.addSegment(topx, topy, leftx, lefty)
		case 8:
			lg.addSegment(leftx, lefty, topx, topy)
		case 9:
			lg.addSegment(bottomx, bottomy, topx, topy)
		case 10:
			if !cell.CenterAbove {
				lg.addSegment(bottomx, bottomy, rightx, righty)
				lg.addSegment(topx, topy, leftx, lefty)
			} else {
				lg.addSegment(rightx, righty, topx, topy)
				lg.addSegment(leftx, lefty, bottomx, bottomy)
			}
		case 11:
			lg.addSegment(rightx, righty, topx, topy)
		case 12:
			lg.addSegment(leftx, lefty, rightx, righty)
		case 13:
			lg.addSegment(bottomx, bottomy, rightx, righty)
		case 14:
			lg.addSegment(leftx, lefty, bottomx, bottomy)
		}
	}

	if y == 0 {
		// top
		if cell.Case&0x8 == 0 {
			ax := offsetx + cellw*x - cellw/2
			ay := offsety
			bx := ax + cellw
			by := ay
			if x == 0 {
				lg.addSegment(ax+cellw/2, ay+cellh/2, ax+cellw/2, ay)
				lg.addSegment(ax+cellw/2, ay, bx, by)
			} else {
				lg.addSegment(ax, ay, bx, by)
			}
		}
	} else if y == gridHeight-1 {
		// bottom
		if cell.Case&0x2 == 0 {
			ax := offsetx + cellw*(x+1) + cellw/2
			ay := offsety + gridHeight*cellh
			bx := ax - cellw
			by := ay
			if x == gridWidth-1 {
				lg.addSegment(ax-cellw/2, ay-cellh/2, ax-cellw/2, ay)
				lg.addSegment(ax-cellw/2, ay, bx, by)
			} else {
				lg.addSegment(ax, ay, bx, by)
			}
		}
	}
	if x == 0 {
		// left
		if cell.Case&0x1 == 0 {
			ax := offsetx
			ay := offsety + cellh*(y+1) + cellh/2
			bx := ax
			by := ay - cellh
			if y == gridHeight-1 {
				lg.addSegment(ax+cellw/2, ay-cellh/2, ax, ay-cellh/2)
				lg.addSegment(ax, ay-cellh/2, bx, by)
			} else {
				lg.addSegment(ax, ay, bx, by)
			}
		}
	} else if x == gridWidth-1 {
		// right
		if cell.Case&0x4 == 0 {
			ax := offsetx + gridWidth*cellw
			ay := offsety + cellh*y - cellh/2
			bx := ax
			by := ay + cellh
			if y == 0 {
				lg.addSegment(ax-cellw/2, ay+cellh/2, ax, ay+cellh/2)
				lg.addSegment(ax, ay+cellh/2, bx, by)
			} else {
				lg.addSegment(ax, ay, bx, by)
			}
		}
	}
}

func (lg *lineGatherer) addGrid(grid *Grid, width, height float64) {
	start := time.Now()
	gwidth, gheight := float64(grid.Width), float64(grid.Height)

	for y := 0; y < grid.Height; y++ {
		for x := 0; x < grid.Width; x++ {
			cell := grid.Cells[y*grid.Width+x]
			lg.addCell(cell, float64(x), float64(y), width, height, gwidth, gheight)
		}
	}
	println("** addCells:", time.Now().Sub(start).String())
	start = time.Now()
	lg.reduceLines()
	println("** reduceLines:", time.Now().Sub(start).String())
}
