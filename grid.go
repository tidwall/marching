package marching

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

func NewGrid(values []float64, width, height int, level float64, complexity int) *Grid {
	if len(values) != width*height {
		panic("number of values are not equal to width multiplied by height")
	}
	if width < 2 || height < 2 {
		panic("width or height are not greater than or equal to two")
	}
	if complexity < 0 {
		panic("invalid complexity")
	}
	cmplx := uint(complexity)
	gwidth := (width - 1) << cmplx
	gheight := (height - 1) << cmplx

	cells := make([]Cell, gwidth*gheight)
	var j int
	for y := 0; y < gheight; y++ {
		for x := 0; x < gwidth; x++ {
			vals := [4]float64{
				values[((y>>cmplx)+0)*width+((x>>cmplx)+0)],
				values[((y>>cmplx)+0)*width+((x>>cmplx)+1)],
				values[((y>>cmplx)+1)*width+((x>>cmplx)+1)],
				values[((y>>cmplx)+1)*width+((x>>cmplx)+0)],
			}
			if complexity > 0 {
				rx := x % (1 << cmplx)
				ry := y % (1 << cmplx)
				sx := float64(rx) / float64(int(1<<cmplx))
				sy := float64(ry) / float64(int(1<<cmplx))
				ex := sx + 1/float64(int(1<<cmplx))
				ey := sy + 1/float64(int(1<<cmplx))
				vals = [4]float64{
					bilinearInterpolation(vals, sx, sy),
					bilinearInterpolation(vals, ex, sy),
					bilinearInterpolation(vals, ex, ey),
					bilinearInterpolation(vals, sx, ey),
				}
			}
			center := bilinearInterpolation(vals, 0.5, 0.5)
			var cell Cell
			for i := 0; i < 4; i++ {
				if vals[i] < level {
					cell.Case |= 1 << uint(4-i-1)
				}
			}
			cell.CenterAbove = center >= level
			cells[j] = cell
			j++
		}
	}
	return &Grid{
		Cells:  cells,
		Width:  gwidth,
		Height: gheight,
	}
}

func bilinearInterpolation(vals [4]float64, x, y float64) float64 {
	return vals[3]*(1-x)*y + vals[2]*x*y + vals[0]*(1-x)*(1-y) + vals[1]*x*(1-y)
}

/*
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
	points  []point
	edge    bool
	connect bool
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

func (lg *lineGatherer) DrawLine(ax, ay, bx, by float64, edge, connect bool) {
	lg.lines = append(lg.lines, line{[]point{{ax, ay}, {bx, by}}, edge, connect})
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

func (lg *lineGatherer) RemoveEdges() {
	for i := 0; i < len(lg.lines); i++ {
		if lg.lines[i].edge {
			lg.lines = append(lg.lines[:i], lg.lines[i+1:]...)
			i--
			continue
		}
	}
}

func (lg *lineGatherer) ReduceLines(edges bool) {
again1:
	for i := 0; i < len(lg.lines); i++ {
		if lg.lines[i].edge && lg.lines[i].connect && len(lg.lines[i].points) == 2 {
			for j := 0; j < len(lg.lines); j++ {
				if i != j && lg.lines[j].edge && lg.lines[j].connect && len(lg.lines[j].points) == 2 {
					if (lg.lines[i].points[0].veryClose(lg.lines[j].points[0]) &&
						lg.lines[i].points[1].veryClose(lg.lines[j].points[1])) ||
						(lg.lines[i].points[1].veryClose(lg.lines[j].points[0]) &&
							lg.lines[i].points[0].veryClose(lg.lines[j].points[1])) {
						if i < j {
							lg.lines = append(lg.lines[:j], lg.lines[j+1:]...)
							lg.lines = append(lg.lines[:i], lg.lines[i+1:]...)
							goto again1
						} else {
							lg.lines = append(lg.lines[:i], lg.lines[i+1:]...)
							lg.lines = append(lg.lines[:j], lg.lines[j+1:]...)
							goto again1
						}
					}
				}
			}
		}
	}
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
				//gc.QuadraticTo(line.points[i].x, line.points[i].y, line.points[i+1].x, line.points[i+1].y)
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
	DrawLine(ax, ay, bx, by float64, edge, connect bool)
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
			if opts.ExpandEdges {
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
			if opts.ExpandEdges {
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
			if opts.ExpandEdges {
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
			if opts.ExpandEdges {
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

func (grid *Grid) Image(width, height int, opts *ImageOptions) *image.RGBA {
	widthf, heightf := float64(width), float64(height)
	if opts == nil {
		opts = &ImageOptions{}
	}
	if opts.LineWidth == 0 {
		opts.LineWidth = 1
	}
	if opts.StrokeColor == nil {
		opts.StrokeColor = color.Black
	}
	if opts.FillColor == nil {
		opts.FillColor = color.NRGBA{0, 0, 0, 0x77}
	}
	var rp float64
	if widthf < heightf {
		rp = widthf / 256
	} else {
		rp = heightf / 256
	}
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	if opts.Marks {
		grid.drawMarksGG(img, rp, widthf, heightf)
	}
	gc := gg.NewContextForRGBA(img)
	//return img

	lg := newLineGatherer()
	for y := 0; y < grid.Height; y++ {
		for x := 0; x < grid.Width; x++ {
			cell := grid.Cells[y*grid.Height+x]
			grid.drawCell(cell, x, y, lg, widthf, heightf, opts)
		}
	}
	lg.ReduceLines(false)
	// draw fill
	if !opts.NoFill {
		fill := lg.Copy()
		fill.ReduceLines(true)
		if !opts.Marks {
			gc.SetColor(opts.FillColor)
		} else {
			gc.SetColor(color.NRGBA{0, 0, 0, 0x11})
		}
		fill.DrawToContext(gc, opts.Marks)
		gc.Fill()
	}
	// return img
	// draw stroke
	if !opts.NoStroke {
		stroke := lg.Copy()
		if opts.Marks {
			stroke.ReduceLines(true)
			gc.SetLineWidth(opts.LineWidth)
		} else {
			stroke.RemoveEdges()
			gc.SetLineWidth(opts.LineWidth)
		}
		gc.SetColor(opts.StrokeColor)
		stroke.DrawToContext(gc, opts.Marks)
		gc.Stroke()
	}

	return img
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
}
*/
