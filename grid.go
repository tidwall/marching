package marching

import (
	"fmt"
	"image"
	"image/color"
	"math"

	"github.com/fogleman/gg"
)

type Case byte

type Cell struct {
	Case        Case
	CenterAbove bool
}

type Grid struct {
	Cells  []Cell
	Width  int
	Height int
}

func NewGrid(values []float64, width, height int, level float64) *Grid {
	if len(values) != width*height {
		panic("number of values are not equal to width multiplied by height")
	}
	if width <= 2 || height <= 2 {
		panic("width or height is not greater than zero")
	}
	cells := make([]Cell, (width-1)*(height-1))
	var j int
	for y := 0; y < height-1; y++ {
		for x := 0; x < width-1; x++ {
			var cell Cell
			// top-left
			if values[y*width+x] < level {
				cell.Case |= 0x8
			}
			// top-right
			if values[y*width+x+1] < level {
				cell.Case |= 0x4
			}
			// bottom-right
			if values[(y+1)*width+x+1] < level {
				cell.Case |= 0x2
			}
			// bottom-left
			if values[(y+1)*width+x] < level {
				cell.Case |= 0x1
			}
			if (values[y*width+x]+values[y*width+x+1]+
				values[(y+1)*width+x+1]+values[(y+1)*width+x])/4 >= level {
				cell.CenterAbove = true
			}
			cells[j] = cell
			j++
		}
	}
	return &Grid{
		Cells:  cells,
		Width:  width - 1,
		Height: height - 1,
	}
}

type ImageOptions struct {
	Rounded     bool
	Marks       bool
	LineWidth   float64
	Color       color.Color
	ExpandEdges bool
}

//////////////////////////////
// lineGatherer
type lineGatherer struct {
	lines []line
}

type point struct{ x, y float64 }

func (p1 point) veryClose(p2 point) bool {
	const maxRelativeError = 0.00001
	if p1 == p2 {
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

type line struct {
	points []point
	edge   bool
}

func (l *line) first() point { return l.points[0] }
func (l *line) last() point  { return l.points[len(l.points)-1] }
func newLineGatherer() *lineGatherer {
	return &lineGatherer{}
}

func (lg *lineGatherer) appendLines(i, j int) {
	lg.lines[i].points = append(lg.lines[i].points, lg.lines[j].points[1:]...)
	lg.lines = append(lg.lines[:j], lg.lines[j+1:]...)
}

func (lg *lineGatherer) DrawLine(ax, ay, bx, by float64, edge bool) {
	lg.lines = append(lg.lines, line{[]point{{ax, ay}, {bx, by}}, edge})
}
func (lg *lineGatherer) Copy() *lineGatherer {
	nlg := newLineGatherer()
	nlg.lines = make([]line, len(lg.lines))
	for i, line := range lg.lines {
		nline := line
		nline.points = make([]point, len(line.points))
		copy(nline.points, line.points)
		nlg.lines[i] = nline
	}
	return nlg
}
func (lg *lineGatherer) ReduceLines(edges bool) {
again:
	for i := 0; i < len(lg.lines); i++ {
		if !edges && lg.lines[i].edge {
			continue
		}
		for j := 0; j < len(lg.lines); j++ {
			if i == j {
				continue
			}
			if !edges && lg.lines[j].edge {
				continue
			}
			if lg.lines[j].first().veryClose(lg.lines[i].last()) {
				lg.appendLines(i, j)
				goto again
			}
			if lg.lines[j].last().veryClose(lg.lines[i].first()) {
				lg.appendLines(j, i)
				goto again
			}
			if lg.lines[j].last().veryClose(lg.lines[i].last()) ||
				lg.lines[j].first().veryClose(lg.lines[i].first()) {
				// reverse the line and try again
				s := lg.lines[j].points
				for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
					s[i], s[j] = s[j], s[i]
				}
				goto again
			}
		}
	}
}

var marksLineColors = []color.Color{
	color.RGBA{0xff, 0x00, 0x00, 0xff}, // red
	color.RGBA{0x00, 0xff, 0x00, 0xff}, // green
	color.RGBA{0x00, 0x00, 0xff, 0xff}, // blue
	color.RGBA{0xff, 0xff, 0x00, 0xff}, // yellow
	color.RGBA{0x00, 0xff, 0xff, 0xff}, // cyan
	color.RGBA{0xff, 0x00, 0xff, 0xff}, // magenta
	color.RGBA{0x66, 0x66, 0x66, 0xff}, // dark-gray
	color.RGBA{0xCC, 0xCC, 0xCC, 0xff}, // light-gray
}

func (lg *lineGatherer) DrawToContext(gc *gg.Context, marks bool) {
	for i, line := range lg.lines {
		if marks {
			fmt.Printf("drawing line: %d   {%v,%v}\n", i, line.first(), line.last())
		}
		if len(line.points) > 0 {
			if marks {
				gc.SetColor(marksLineColors[i%len(marksLineColors)])
			}
			gc.MoveTo(line.points[0].x, line.points[0].y)
			for i := 1; i < len(line.points); i++ {
				gc.LineTo(line.points[i].x, line.points[i].y)
			}
			if marks {
				gc.Stroke()
			}
		}
	}
}

type lineKey struct {
	x, y float64
}

type drawer interface {
	DrawLine(ax, ay, bx, by float64, edge bool)
}

//////////////////////////////
// drawCell
func (grid *Grid) drawCell(
	cell Cell, x, y int,
	gc drawer,
	widthf, heightf float64,
	opts *ImageOptions,
) {
	var cellw, cellh float64
	var offsetx, offsety float64
	if opts.Marks {
		cellh = heightf / float64(grid.Width+1)
		cellw = widthf / float64(grid.Height+1)
		offsetx = cellw / 2
		offsety = cellh / 2
	} else {
		cellh = heightf / float64(grid.Width)
		cellw = widthf / float64(grid.Height)
	}
	var leftx = offsetx + cellw*float64(x)
	var lefty = offsety + cellh*float64(y) + cellh*0.5

	var rightx = offsetx + cellw*float64(x) + cellw
	var righty = offsety + cellh*float64(y) + cellh*0.5
	var topx = offsetx + cellw*float64(x) + cellw*0.5
	var topy = offsety + cellh*float64(y)
	var bottomx = offsetx + cellw*float64(x) + cellw*0.5
	var bottomy = offsety + cellh*float64(y) + cellh
	//var centerx = offsetx + cellw*float64(x) + cellw*0.5
	//var centery = offsety + cellh*float64(y) + cellh*0.5

	switch cell.Case {
	default:
		panic("invalid case")
	case 0:

	case 1:
		gc.DrawLine(bottomx, bottomy, leftx, lefty, false)
	case 2:
		gc.DrawLine(rightx, righty, bottomx, bottomy, false)
	case 3:
		gc.DrawLine(rightx, righty, leftx, lefty, false)
	case 4:
		gc.DrawLine(topx, topy, rightx, righty, false)
	case 5:
		if !cell.CenterAbove {
			gc.DrawLine(topx, topy, rightx, righty, false)
			gc.DrawLine(bottomx, bottomy, leftx, lefty, false)
		} else {
			gc.DrawLine(bottomx, bottomy, rightx, righty, false)
			gc.DrawLine(topx, topy, leftx, lefty, false)
		}
	case 6:
		gc.DrawLine(topx, topy, bottomx, bottomy, false)
	case 7:
		gc.DrawLine(topx, topy, leftx, lefty, false)
	case 8:
		gc.DrawLine(leftx, lefty, topx, topy, false)
	case 9:
		gc.DrawLine(bottomx, bottomy, topx, topy, false)
	case 10:
		if !cell.CenterAbove {
			gc.DrawLine(bottomx, bottomy, rightx, righty, false)
			gc.DrawLine(topx, topy, leftx, lefty, false)
		} else {
			gc.DrawLine(rightx, righty, topx, topy, false)
			gc.DrawLine(leftx, lefty, bottomx, bottomy, false)
		}
	case 11:
		gc.DrawLine(rightx, righty, topx, topy, false)
	case 12:
		gc.DrawLine(leftx, lefty, rightx, righty, false)
	case 13:
		gc.DrawLine(bottomx, bottomy, rightx, righty, false)
	case 14:
		gc.DrawLine(leftx, lefty, bottomx, bottomy, false)
	case 15:
	}

	if y == 0 {
		// top
		if cell.Case&0x8 == 0 {
			ax := offsetx + cellw*float64(x) - cellw/2
			ay := offsety
			bx := ax + cellw
			by := ay
			if opts.ExpandEdges {
				gc.DrawLine(ax, ay, ax, ay-cellh/2, true)
				gc.DrawLine(ax, ay-cellh/2, bx, by-cellh/2, true)
				gc.DrawLine(bx, by-cellh/2, bx, by, true)
			} else {
				if x == 0 {
					gc.DrawLine(ax+cellw/2, ay+cellh/2, ax+cellw/2, ay, true)
					gc.DrawLine(ax+cellw/2, ay, bx, by, true)
				} else {
					gc.DrawLine(ax, ay, bx, by, true)
				}
			}
		}
	} else if y == grid.Height-1 {
		// bottom
		if cell.Case&0x2 == 0 {
			ax := offsetx + cellw*float64(x+1) + cellw/2
			ay := offsety + float64(grid.Height)*cellh
			bx := ax - cellw
			by := ay
			if opts.ExpandEdges {
				gc.DrawLine(ax, ay, ax, ay+cellh/2, true)
				gc.DrawLine(ax, ay+cellh/2, bx, by+cellh/2, true)
				gc.DrawLine(bx, by+cellh/2, bx, by, true)
			} else {
				if x == grid.Width-1 {
					gc.DrawLine(ax-cellw/2, ay-cellh/2, ax-cellw/2, ay, true)
					gc.DrawLine(ax-cellw/2, ay, bx, by, true)
				} else {
					gc.DrawLine(ax, ay, bx, by, true)
				}
			}
		}
	}
	if x == 0 {
		// left
		if cell.Case&0x1 == 0 {
			ax := offsetx
			ay := offsety + cellh*float64(y+1) + cellh/2
			bx := ax
			by := ay - cellh
			if opts.ExpandEdges {
				gc.DrawLine(ax, ay, ax-cellw/2, ay, true)
				gc.DrawLine(ax-cellw/2, ay, bx-cellw/2, by, true)
				gc.DrawLine(bx-cellw/2, by, bx, by, true)
			} else {
				if y == grid.Height-1 {
					gc.DrawLine(ax+cellw/2, ay-cellh/2, ax, ay-cellh/2, true)
					gc.DrawLine(ax, ay-cellh/2, bx, by, true)
				} else {
					gc.DrawLine(ax, ay, bx, by, true)
				}
			}
		}
	} else if x == grid.Width-1 {
		// right
		if cell.Case&0x4 == 0 {
			ax := offsetx + float64(grid.Width)*cellw
			ay := offsety + cellh*float64(y) - cellh/2
			bx := ax
			by := ay + cellh
			if opts.ExpandEdges {
				gc.DrawLine(ax, ay, ax+cellw/2, ay, true)
				gc.DrawLine(ax+cellw/2, ay, bx+cellw/2, by, true)
				gc.DrawLine(bx+cellw/2, by, bx, by, true)
			} else {
				if y == 0 {
					gc.DrawLine(ax-cellw/2, ay+cellh/2, ax, ay+cellh/2, true)
					gc.DrawLine(ax, ay+cellh/2, bx, by, true)
				} else {
					gc.DrawLine(ax, ay, bx, by, true)
				}
			}
		}
	}

}

func (grid *Grid) Image(width, height int, opts *ImageOptions) *image.RGBA {
	widthf, heightf := float64(width), float64(height)
	if opts == nil {
		opts = &ImageOptions{}
	}
	if opts.LineWidth == 0 {
		opts.LineWidth = 1
	}
	if opts.Color == nil {
		opts.Color = color.Black
	}
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	gc := gg.NewContextForRGBA(img)

	if opts.Marks {
		var rp float64
		if widthf < heightf {
			rp = widthf / 256
		} else {
			rp = heightf / 256
		}
		// draw background
		gc.Clear()
		gc.SetColor(color.White)
		gc.MoveTo(0, 0)
		gc.LineTo(widthf, 0)
		gc.LineTo(widthf, heightf)
		gc.LineTo(0, heightf)
		gc.LineTo(0, 0)
		gc.Fill()

		// draw value outlines
		gc.SetColor(color.RGBA{0xCC, 0xCC, 0xCC, 0xFF})
		gc.SetLineWidth(rp * 2)
		gc.MoveTo(0, 0)
		gc.LineTo(widthf, 0)
		gc.LineTo(widthf, heightf)
		gc.LineTo(0, heightf)
		gc.LineTo(0, 0)
		gc.Stroke()
		gc.SetLineWidth(rp * 1)
		cellh := heightf / float64(grid.Width+1)
		cellw := widthf / float64(grid.Height+1)
		for y := cellh; y < heightf; y += cellh {
			gc.MoveTo(0, y)
			gc.LineTo(widthf, y)
			gc.Stroke()
		}
		for x := cellw; x < widthf; x += cellw {
			gc.MoveTo(x, 0)
			gc.LineTo(x, heightf)
			gc.Stroke()
		}

		// draw grid outlines
		gc.SetColor(color.RGBA{0x99, 0x99, 0x99, 0xFF})
		gc.SetLineWidth(rp * 4)
		gc.MoveTo(cellw/2, cellh/2)
		gc.LineTo(widthf-cellw/2, cellh/2)
		gc.LineTo(widthf-cellw/2, heightf-cellh/2)
		gc.LineTo(cellw/2, heightf-cellh/2)
		gc.LineTo(cellw/2, cellh/2)
		gc.Stroke()
		for y := cellh + cellh/2; y < heightf; y += cellh {
			gc.MoveTo(cellw/2, y)
			gc.LineTo(widthf-cellw/2, y)
			gc.Stroke()
		}
		for x := cellw + cellw/2; x < widthf; x += cellw {
			gc.MoveTo(x, cellh/2)
			gc.LineTo(x, heightf-cellh/2)
			gc.Stroke()
		}

		// draw cell outlines
		for y := 0; y < grid.Height; y++ {
			for x := 0; x < grid.Width; x++ {
				cell := grid.Cells[y*grid.Height+x]

				gc.SetLineWidth(rp * 1)
				gc.SetColor(color.White)
				gc.DrawCircle(cellw/2+cellw*float64(x), cellh/2+cellh*float64(y), rp*8)
				gc.DrawCircle(cellw/2+cellw*float64(x+1), cellh/2+cellh*float64(y), rp*8)
				gc.DrawCircle(cellw/2+cellw*float64(x+1), cellh/2+cellh*float64(y+1), rp*8)
				gc.DrawCircle(cellw/2+cellw*float64(x), cellh/2+cellh*float64(y+1), rp*8)
				gc.Fill()

				gc.SetColor(color.Black)
				gc.DrawCircle(cellw/2+cellw*float64(x), cellh/2+cellh*float64(y), rp*8)
				gc.DrawCircle(cellw/2+cellw*float64(x+1), cellh/2+cellh*float64(y), rp*8)
				gc.DrawCircle(cellw/2+cellw*float64(x+1), cellh/2+cellh*float64(y+1), rp*8)
				gc.DrawCircle(cellw/2+cellw*float64(x), cellh/2+cellh*float64(y+1), rp*8)
				gc.Stroke()

				//top-left
				if cell.Case&0x8 != 0 {
					gc.SetColor(color.Black)
					gc.DrawCircle(cellw/2+cellw*float64(x), cellh/2+cellh*float64(y), rp*8)
					gc.Fill()
				}
				// top-right
				if cell.Case&0x4 != 0 {
					gc.SetColor(color.Black)
					gc.DrawCircle(cellw/2+cellw*float64(x+1), cellh/2+cellh*float64(y), rp*8)
					gc.Fill()
				}
				// bottom-right
				if cell.Case&0x2 != 0 {
					gc.SetColor(color.Black)
					gc.DrawCircle(cellw/2+cellw*float64(x+1), cellh/2+cellh*float64(y+1), rp*8)
					gc.Fill()
				}
				// bottom-left
				if cell.Case&0x1 != 0 {
					gc.SetColor(color.Black)
					gc.DrawCircle(cellw/2+cellw*float64(x), cellh/2+cellh*float64(y+1), rp*8)
					gc.Fill()
				}
			}
		}
		gc.SetColor(color.RGBA{0x1b, 0xa3, 0xe5, 0xff})
		gc.SetLineWidth(rp * 6)
	} else {
		gc.SetColor(opts.Color)
		gc.SetLineWidth(opts.LineWidth)
	}
	lg := newLineGatherer()
	for y := 0; y < grid.Height; y++ {
		for x := 0; x < grid.Width; x++ {
			cell := grid.Cells[y*grid.Height+x]
			grid.drawCell(cell, x, y, lg, widthf, heightf, opts)
		}
	}

	lg.ReduceLines(false)

	var fill, stroke *lineGatherer
	stroke = lg
	fill = lg.Copy()
	fill.ReduceLines(true)
	if opts.Marks {
		stroke.ReduceLines(true)
	}

	fill.DrawToContext(gc, false)
	if opts.Marks {
		gc.Fill()
	}
	stroke.DrawToContext(gc, opts.Marks)
	if !opts.Marks {
		gc.Stroke()
	}
	return img

}
