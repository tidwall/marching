package marching

import (
	"fmt"
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
