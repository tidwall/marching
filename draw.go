package marching

import (
	"fmt"
	"image"
	"image/color"

	"github.com/fogleman/gg"
)

type ImageOptions struct {
	//Rounded     bool
	Marks       bool
	LineWidth   float64
	StrokeColor color.Color
	FillColor   color.Color
	NoStroke    bool
	NoFill      bool
	ExpandEdges bool
	//Spline      float64
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
	if opts != nil && opts.Marks {
		cellw = widthf / float64(grid.Width+1)
		cellh = heightf / float64(grid.Height+1)
		offsetx = cellw / 2
		offsety = cellh / 2
	} else {
		cellw = widthf / float64(grid.Width)
		cellh = heightf / float64(grid.Height)
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
		gc.DrawLine(bottomx, bottomy, leftx, lefty, false, false)
	case 2:
		gc.DrawLine(rightx, righty, bottomx, bottomy, false, false)
	case 3:
		gc.DrawLine(rightx, righty, leftx, lefty, false, false)
	case 4:
		gc.DrawLine(topx, topy, rightx, righty, false, false)
	case 5:
		if !cell.CenterAbove {
			gc.DrawLine(topx, topy, rightx, righty, false, false)
			gc.DrawLine(bottomx, bottomy, leftx, lefty, false, false)
		} else {
			gc.DrawLine(bottomx, bottomy, rightx, righty, false, false)
			gc.DrawLine(topx, topy, leftx, lefty, false, false)
		}
	case 6:
		gc.DrawLine(topx, topy, bottomx, bottomy, false, false)
	case 7:
		gc.DrawLine(topx, topy, leftx, lefty, false, false)
	case 8:
		gc.DrawLine(leftx, lefty, topx, topy, false, false)
	case 9:
		gc.DrawLine(bottomx, bottomy, topx, topy, false, false)
	case 10:
		if !cell.CenterAbove {
			gc.DrawLine(bottomx, bottomy, rightx, righty, false, false)
			gc.DrawLine(topx, topy, leftx, lefty, false, false)
		} else {
			gc.DrawLine(rightx, righty, topx, topy, false, false)
			gc.DrawLine(leftx, lefty, bottomx, bottomy, false, false)
		}
	case 11:
		gc.DrawLine(rightx, righty, topx, topy, false, false)
	case 12:
		gc.DrawLine(leftx, lefty, rightx, righty, false, false)
	case 13:
		gc.DrawLine(bottomx, bottomy, rightx, righty, false, false)
	case 14:
		gc.DrawLine(leftx, lefty, bottomx, bottomy, false, false)
	case 15:
	}

	if y == 0 {
		// top
		if cell.Case&0x8 == 0 {
			ax := offsetx + cellw*float64(x) - cellw/2
			ay := offsety
			bx := ax + cellw
			by := ay
			if opts != nil && opts.ExpandEdges {
				if x == 0 {
					gc.DrawLine(ax+cellw/2, ay+cellh/2, ax, ay+cellh/2, true, true)
					gc.DrawLine(ax, ay+cellh/2, ax, ay-cellh/2, true, false)
					gc.DrawLine(ax, ay-cellh/2, bx, by-cellh/2, true, false)
					gc.DrawLine(bx, by-cellh/2, bx, by, true, true)
				} else {
					gc.DrawLine(ax, ay, ax, by-cellh/2, true, true)
					gc.DrawLine(ax, by-cellh/2, ax+cellw, by-cellh/2, true, false)
					gc.DrawLine(ax+cellw, by-cellh/2, bx, by, true, true)
				}
			} else {
				if x == 0 {
					gc.DrawLine(ax+cellw/2, ay+cellh/2, ax+cellw/2, ay, true, false)
					gc.DrawLine(ax+cellw/2, ay, bx, by, true, false)
				} else {
					gc.DrawLine(ax, ay, bx, by, true, false)
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
			if opts != nil && opts.ExpandEdges {
				if x == grid.Width-1 {
					gc.DrawLine(ax-cellw/2, ay-cellh/2, ax, ay-cellh/2, true, true)
					gc.DrawLine(ax, ay-cellh/2, ax, ay+cellh/2, true, false)
					gc.DrawLine(ax, ay+cellh/2, bx, by+cellh/2, true, false)
					gc.DrawLine(bx, by+cellh/2, bx, by, true, true)
				} else {
					gc.DrawLine(ax, ay, ax, ay+cellh/2, true, true)
					gc.DrawLine(ax, ay+cellh/2, bx, by+cellh/2, true, false)
					gc.DrawLine(bx, by+cellh/2, bx, by, true, true)
				}
			} else {
				if x == grid.Width-1 {
					gc.DrawLine(ax-cellw/2, ay-cellh/2, ax-cellw/2, ay, true, false)
					gc.DrawLine(ax-cellw/2, ay, bx, by, true, false)
				} else {
					gc.DrawLine(ax, ay, bx, by, true, false)
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
			if opts != nil && opts.ExpandEdges {
				if y == grid.Height-1 {
					gc.DrawLine(ax+cellw/2, ay-cellh/2, ax+cellw/2, ay, true, true)
					gc.DrawLine(ax+cellw/2, ay, ax-cellw/2, ay, true, false)
					gc.DrawLine(ax-cellw/2, ay, bx-cellw/2, by, true, false)
					gc.DrawLine(bx-cellw/2, by, bx, by, true, true)
				} else {
					gc.DrawLine(ax, ay, ax-cellw/2, ay, true, true)
					gc.DrawLine(ax-cellw/2, ay, bx-cellw/2, by, true, false)
					gc.DrawLine(bx-cellw/2, by, bx, by, true, true)
				}
			} else {
				if y == grid.Height-1 {
					gc.DrawLine(ax+cellw/2, ay-cellh/2, ax, ay-cellh/2, true, false)
					gc.DrawLine(ax, ay-cellh/2, bx, by, true, false)
				} else {
					gc.DrawLine(ax, ay, bx, by, true, false)
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
			if opts != nil && opts.ExpandEdges {
				if y == 0 {
					gc.DrawLine(ax-cellw/2, ay+cellh/2, ax-cellw/2, ay, true, true)
					gc.DrawLine(ax-cellw/2, ay, ax+cellw/2, ay, true, false)
					gc.DrawLine(ax+cellw/2, ay, bx+cellw/2, by, true, false)
					gc.DrawLine(bx+cellw/2, by, bx, by, true, true)
				} else {
					gc.DrawLine(ax, ay, ax+cellw/2, ay, true, true)
					gc.DrawLine(ax+cellw/2, ay, bx+cellw/2, by, true, false)
					gc.DrawLine(bx+cellw/2, by, bx, by, true, true)
				}
			} else {
				if y == 0 {
					gc.DrawLine(ax-cellw/2, ay+cellh/2, ax, ay+cellh/2, true, false)
					gc.DrawLine(ax, ay+cellh/2, bx, by, true, false)
				} else {
					gc.DrawLine(ax, ay, bx, by, true, false)
				}
			}
		}
	}
}

func (grid *Grid) drawMarksGG(img *image.RGBA, rp, widthf, heightf float64) {
	gc := gg.NewContextForRGBA(img)

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
	cellw := widthf / float64(grid.Width+1)
	cellh := heightf / float64(grid.Height+1)
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

	circleSize := 5.0
	// draw cell outlines
	for y := 0; y < grid.Height; y++ {
		for x := 0; x < grid.Width; x++ {
			cell := grid.Cells[y*grid.Width+x]
			gc.SetLineWidth(rp * 1)
			gc.SetColor(color.White)
			gc.DrawCircle(cellw/2+cellw*float64(x), cellh/2+cellh*float64(y), rp*circleSize)
			gc.DrawCircle(cellw/2+cellw*float64(x+1), cellh/2+cellh*float64(y), rp*circleSize)
			gc.DrawCircle(cellw/2+cellw*float64(x+1), cellh/2+cellh*float64(y+1), rp*circleSize)
			gc.DrawCircle(cellw/2+cellw*float64(x), cellh/2+cellh*float64(y+1), rp*circleSize)
			gc.Fill()

			gc.SetColor(color.Black)
			gc.DrawCircle(cellw/2+cellw*float64(x), cellh/2+cellh*float64(y), rp*circleSize)
			gc.DrawCircle(cellw/2+cellw*float64(x+1), cellh/2+cellh*float64(y), rp*circleSize)
			gc.DrawCircle(cellw/2+cellw*float64(x+1), cellh/2+cellh*float64(y+1), rp*circleSize)
			gc.DrawCircle(cellw/2+cellw*float64(x), cellh/2+cellh*float64(y+1), rp*circleSize)
			gc.Stroke()

			//top-left
			if cell.Case&0x8 != 0 {
				gc.SetColor(color.Black)
				gc.DrawCircle(cellw/2+cellw*float64(x), cellh/2+cellh*float64(y), rp*circleSize)
				gc.Fill()
			}
			// top-right
			if cell.Case&0x4 != 0 {
				gc.SetColor(color.Black)
				gc.DrawCircle(cellw/2+cellw*float64(x+1), cellh/2+cellh*float64(y), rp*circleSize)
				gc.Fill()
			}
			// bottom-right
			if cell.Case&0x2 != 0 {
				gc.SetColor(color.Black)
				gc.DrawCircle(cellw/2+cellw*float64(x+1), cellh/2+cellh*float64(y+1), rp*circleSize)
				gc.Fill()
			}
			// bottom-left
			if cell.Case&0x1 != 0 {
				gc.SetColor(color.Black)
				gc.DrawCircle(cellw/2+cellw*float64(x), cellh/2+cellh*float64(y+1), rp*circleSize)
				gc.Fill()
			}
		}
	}
	gc.SetColor(color.RGBA{0x1b, 0xa3, 0xe5, 0xff})
	gc.SetLineWidth(rp * 6)
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
			gc.MoveTo(line.points[0].X, line.points[0].Y)
			for i := 1; i < len(line.points); i++ {
				//gc.QuadraticTo(line.points[i].x, line.points[i].y, line.points[i+1].x, line.points[i+1].y)
				gc.LineTo(line.points[i].X, line.points[i].Y)
			}
			if marks {
				gc.Stroke()
			}
		}
	}
}

type drawer interface {
	DrawLine(ax, ay, bx, by float64, edge, connect bool)
}

func (lg *lineGatherer) DrawLine(ax, ay, bx, by float64, edge, connect bool) {
	lg.lines = append(lg.lines, line{[]Point{{ax, ay}, {bx, by}}})
}
func (lg *lineGatherer) drawCopy() *lineGatherer {
	nlg := newLineGatherer()
	nlg.lines = make([]line, len(lg.lines))
	for i, line := range lg.lines {
		nline := line
		nline.points = make([]Point, len(line.points))
		copy(nline.points, line.points)
		nlg.lines[i] = nline
	}
	return nlg
}

// http://stackoverflow.com/a/1165943/424124
func pathIsClockwise(path []Point) bool {
	var total float64
	for i := 1; i < len(path); i++ {
		total += (path[i].X - path[i-1].X) * (path[i].Y - path[i-1].Y)
	}
	fmt.Printf("%v\n", total)
	return total < 0
}
func reverseWinding(path []Point) []Point {
	for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
		path[i], path[j] = path[j], path[i]
	}
	return path
}

func savePathsPNG(paths [][]Point, width, height int, filePath string) error {
	gc := gg.NewContext(width, height)
	gc.SetColor(color.White)
	gc.DrawRectangle(0, 0, float64(width), float64(height))
	gc.Fill()

	for _, path := range paths {
		if len(path) > 2 {
			gc.MoveTo(path[0].X, path[0].Y)
			for i := 1; i < len(path); i++ {
				gc.LineTo(path[i].X, path[i].Y)
			}
			gc.ClosePath()
		}
	}
	gc.SetFillRuleEvenOdd()
	gc.SetColor(color.NRGBA{0x88, 0xAA, 0xCC, 0xFF})
	gc.Fill()

	for _, path := range paths {
		if len(path) > 2 {
			gc.MoveTo(path[0].X, path[0].Y)
			for i := 1; i < len(path); i++ {
				gc.LineTo(path[i].X, path[i].Y)
			}
			gc.ClosePath()
		}
	}
	gc.SetLineWidth(2)
	gc.SetColor(color.NRGBA{0xCC, 0xAA, 0x88, 0xFF})
	gc.Stroke()
	return gc.SavePNG(filePath)
}
