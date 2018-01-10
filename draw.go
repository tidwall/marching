package marching

import (
	"image"
	"image/color"

	"github.com/fogleman/gg"
)

type DrawOptions struct {
	StrokeColor color.Color
	FillColor   color.Color
	LineWidth   float64
	NoFill      bool
	NoStroke    bool
	Simplify    int
}

func (grid *Grid) Draw(dst *image.RGBA, x, y, width, height float64, opts *DrawOptions) {
	var simplify int
	if opts != nil {
		simplify = opts.Simplify
	}
	paths := grid.pathsWithOptions(width, height, simplify, nil)
	gc := gg.NewContextForRGBA(dst)

	// fill
	if opts == nil || !opts.NoFill && opts.FillColor != nil {
		for _, path := range paths {
			gc.MoveTo(path[0][0], path[0][1])
			for i := 1; i < len(path); i++ {
				gc.LineTo(path[i][0], path[i][1])
			}
		}
		gc.SetColor(opts.FillColor)
		gc.Fill()
	}

	// stroke
	if opts == nil || !opts.NoStroke {
		for _, path := range paths {
			moveto := true
			for i := 0; i < len(path)-1; i++ {
				if path[i][0] == path[i+1][0] {
					if path[i][0] == 0 || path[i][0] == width {
						moveto = true
						continue
					}
				}
				if path[i][1] == path[i+1][1] {
					if path[i][1] == 0 || path[i][1] == height {
						moveto = true
						continue
					}
				}
				if moveto {
					gc.MoveTo(path[i][0], path[i][1])
					moveto = false
				}
				gc.LineTo(path[i+1][0], path[i+1][1])
			}
		}
		if opts != nil && opts.LineWidth != 0 {
			gc.SetLineWidth(opts.LineWidth)
		} else {
			gc.SetLineWidth(1)
		}
		if opts != nil && opts.StrokeColor != nil {
			gc.SetColor(opts.StrokeColor)
		} else {
			gc.SetColor(color.NRGBA{0, 0, 0, 0x11})
		}
		gc.Stroke()
	}
}

func onedge(point []float64) bool {
	return point[0] == 0 || point[0] == 1 || point[1] == 0 || point[1] == 1
}

// Lines returns individual linestrings that terminate at the edge.
// This operation calls Paths and loops through each one breaking them
// up into smaller segments.
func (grid *Grid) Lines(width, height float64) [][][]float64 {
	var lines [][][]float64
	paths := grid.Paths(width, height)
	for i := 0; i < len(paths); i++ {
		path := paths[i]
		var s int
		for j := 1; j < len(path); j++ {
			point := path[j]
			if onedge(point) {
				line := path[s : j+1]
				if len(line) > 1 && (len(line) > 2 || !onedge(line[0]) || !onedge(line[1])) {
					lines = append(lines, line)
				}
				s = j
			}
		}
		line := path[s:]
		if len(line) > 1 && (len(line) > 2 || !onedge(line[0]) || !onedge(line[1])) {
			lines = append(lines, line)
		}
	}
	return lines
}
