package marching

import (
	"image"
	"image/color"

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
	Rounded bool
	Marks   bool
}

func rp(width, size float64) float64 {
	return width / 256 * size
}

func (grid *Grid) drawCell(
	cell Cell, x, y int,
	gc *gg.Context,
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
	var centerx = offsetx + cellw*float64(x) + cellw*0.5
	var centery = offsety + cellh*float64(y) + cellh*0.5

	leftx, lefty, rightx, righty, topx, topy, bottomx, bottomy = leftx, lefty, rightx, righty, topx, topy, bottomx, bottomy
	switch cell.Case {
	default:
		panic("invalid case")
	case 0:

	case 1:
		gc.MoveTo(bottomx, bottomy)
		if opts.Rounded {
			gc.QuadraticTo(centerx, centery, leftx, lefty)
		} else {
			gc.LineTo(leftx, lefty)
		}
	case 2:
		gc.MoveTo(rightx, righty)
		if opts.Rounded {
			gc.QuadraticTo(centerx, centery, bottomx, bottomy)
		} else {
			gc.LineTo(bottomx, bottomy)
		}
	case 3:
		gc.MoveTo(rightx, righty)
		if opts.Rounded {
			gc.QuadraticTo(centerx, centery, leftx, lefty)
		} else {
			gc.LineTo(leftx, lefty)
		}
	case 4:
		gc.MoveTo(topx, topy)
		if opts.Rounded {
			gc.QuadraticTo(centerx, centery, rightx, righty)
		} else {
			gc.LineTo(rightx, righty)
		}
	case 5:
		if !cell.CenterAbove {
			gc.MoveTo(topx, topy)
			if opts.Rounded {
				gc.QuadraticTo(centerx, centery, rightx, righty)
			} else {
				gc.LineTo(rightx, righty)
			}
			gc.MoveTo(bottomx, bottomy)
			if opts.Rounded {
				gc.QuadraticTo(centerx, centery, leftx, lefty)
			} else {
				gc.LineTo(leftx, lefty)
			}
		} else {
			gc.MoveTo(bottomx, bottomy)
			if opts.Rounded {
				gc.QuadraticTo(centerx, centery, rightx, righty)
			} else {
				gc.LineTo(rightx, righty)
			}
			gc.MoveTo(topx, topy)
			if opts.Rounded {
				gc.QuadraticTo(centerx, centery, leftx, lefty)
			} else {
				gc.LineTo(leftx, lefty)
			}
		}
	case 6:
		gc.MoveTo(topx, topy)
		if opts.Rounded {
			gc.QuadraticTo(centerx, centery, bottomx, bottomy)
		} else {
			gc.LineTo(bottomx, bottomy)
		}
	case 7:
		gc.MoveTo(topx, topy)
		if opts.Rounded {
			gc.QuadraticTo(centerx, centery, leftx, lefty)
		} else {
			gc.LineTo(leftx, lefty)
		}
	case 8:
		gc.MoveTo(leftx, lefty)
		if opts.Rounded {
			gc.QuadraticTo(centerx, centery, topx, topy)
		} else {
			gc.LineTo(topx, topy)
		}
	case 9:
		gc.MoveTo(bottomx, bottomy)
		if opts.Rounded {
			gc.QuadraticTo(centerx, centery, topx, topy)
		} else {
			gc.LineTo(topx, topy)
		}
	case 10:
		if !cell.CenterAbove {
			gc.MoveTo(bottomx, bottomy)
			if opts.Rounded {
				gc.QuadraticTo(centerx, centery, rightx, righty)
			} else {
				gc.LineTo(rightx, righty)
			}
			gc.MoveTo(topx, topy)
			if opts.Rounded {
				gc.QuadraticTo(centerx, centery, leftx, lefty)
			} else {
				gc.LineTo(leftx, lefty)
			}
		} else {
			gc.MoveTo(topx, topy)
			if opts.Rounded {
				gc.QuadraticTo(centerx, centery, rightx, righty)
			} else {
				gc.LineTo(rightx, righty)
			}
			gc.MoveTo(bottomx, bottomy)
			if opts.Rounded {
				gc.QuadraticTo(centerx, centery, leftx, lefty)
			} else {
				gc.LineTo(leftx, lefty)
			}
		}
	case 11:
		gc.MoveTo(rightx, righty)
		if opts.Rounded {
			gc.QuadraticTo(centerx, centery, topx, topy)
		} else {
			gc.LineTo(topx, topy)
		}
	case 12:
		gc.MoveTo(leftx, lefty)
		if opts.Rounded {
			gc.QuadraticTo(centerx, centery, rightx, lefty)
		} else {
			gc.LineTo(rightx, righty)
		}
	case 13:
		gc.MoveTo(bottomx, bottomy)
		if opts.Rounded {
			gc.QuadraticTo(centerx, centery, rightx, righty)
		} else {
			gc.LineTo(rightx, righty)
		}
	case 14:
		gc.MoveTo(bottomx, bottomy)
		if opts.Rounded {
			gc.QuadraticTo(centerx, centery, leftx, lefty)
		} else {
			gc.LineTo(leftx, lefty)
		}
	case 15:
	}
}

func (grid *Grid) Image(width, height int, opts *ImageOptions) *image.RGBA {
	widthf, heightf := float64(width), float64(height)
	if opts == nil {
		opts = &ImageOptions{}
	}
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	gc := gg.NewContextForRGBA(img)

	if opts.Marks {

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
		gc.SetLineWidth(rp(widthf, 2))
		gc.MoveTo(0, 0)
		gc.LineTo(widthf, 0)
		gc.LineTo(widthf, heightf)
		gc.LineTo(0, heightf)
		gc.LineTo(0, 0)
		gc.Stroke()
		gc.SetLineWidth(rp(widthf, 1))
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
		gc.SetColor(color.RGBA{0xb6, 0xe4, 0x38, 0xFF})
		gc.SetLineWidth(rp(widthf, 4))
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

				gc.SetLineWidth(rp(widthf, 1))
				gc.SetColor(color.White)
				gc.DrawCircle(cellw/2+cellw*float64(x), cellh/2+cellh*float64(y), rp(widthf, 8))
				gc.DrawCircle(cellw/2+cellw*float64(x+1), cellh/2+cellh*float64(y), rp(widthf, 8))
				gc.DrawCircle(cellw/2+cellw*float64(x+1), cellh/2+cellh*float64(y+1), rp(widthf, 8))
				gc.DrawCircle(cellw/2+cellw*float64(x), cellh/2+cellh*float64(y+1), rp(widthf, 8))
				gc.Fill()

				gc.SetColor(color.Black)
				gc.DrawCircle(cellw/2+cellw*float64(x), cellh/2+cellh*float64(y), rp(widthf, 8))
				gc.DrawCircle(cellw/2+cellw*float64(x+1), cellh/2+cellh*float64(y), rp(widthf, 8))
				gc.DrawCircle(cellw/2+cellw*float64(x+1), cellh/2+cellh*float64(y+1), rp(widthf, 8))
				gc.DrawCircle(cellw/2+cellw*float64(x), cellh/2+cellh*float64(y+1), rp(widthf, 8))
				gc.Stroke()

				//top-left
				if cell.Case&0x8 != 0 {
					gc.SetColor(color.Black)
					gc.DrawCircle(cellw/2+cellw*float64(x), cellh/2+cellh*float64(y), rp(widthf, 8))
					gc.Fill()
				}
				// top-right
				if cell.Case&0x4 != 0 {
					gc.SetColor(color.Black)
					gc.DrawCircle(cellw/2+cellw*float64(x+1), cellh/2+cellh*float64(y), rp(widthf, 8))
					gc.Fill()
				}
				// bottom-right
				if cell.Case&0x2 != 0 {
					gc.SetColor(color.Black)
					gc.DrawCircle(cellw/2+cellw*float64(x+1), cellh/2+cellh*float64(y+1), rp(widthf, 8))
					gc.Fill()
				}
				// bottom-left
				if cell.Case&0x1 != 0 {
					gc.SetColor(color.Black)
					gc.DrawCircle(cellw/2+cellw*float64(x), cellh/2+cellh*float64(y+1), rp(widthf, 8))
					gc.Fill()
				}
			}
		}
		gc.SetColor(color.RGBA{0x1b, 0xa3, 0xe5, 0xFF})
		gc.SetLineWidth(rp(widthf, 4))
	} else {
		gc.SetColor(color.RGBA{0x1b, 0xa3, 0xe5, 0xFF})
		gc.SetLineWidth(rp(widthf, 4))
	}
	for y := 0; y < grid.Height; y++ {
		for x := 0; x < grid.Width; x++ {
			cell := grid.Cells[y*grid.Height+x]
			grid.drawCell(cell, x, y, gc, widthf, heightf, opts)
		}
	}
	gc.Stroke()
	return img
}

func drawCircle(gc *gg.Context, x, y, radius float64) {
	gc.MoveTo(x+radius, y)
	gc.Clear()
	gc.DrawCircle(x, y, radius)

}
