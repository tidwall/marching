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

func savePathsPNG(grid *Grid, paths []Polygon, aboveMap map[int]Point, width, height int, filePath string) error {
	gc := gg.NewContext(width, height)
	gc.SetColor(color.White)
	gc.DrawRectangle(0, 0, float64(width), float64(height))
	gc.Fill()

	//if len(paths) > 1 {
	//	for i := 0; i < len(paths[2]); i++ {
	//		paths[2][i].X += 20
	//	}
	//}

	for _, path := range paths {
		if len(path) > 2 {
			gc.MoveTo(path[0].X, path[0].Y)
			for i := 1; i < len(path); i++ {
				gc.LineTo(path[i].X, path[i].Y)
			}
			//			gc.ClosePath()
		}
	}

	//	gc.SetFillRuleEvenOdd()
	gc.SetColor(color.NRGBA{0x88, 0xAA, 0xCC, 0xFF})
	gc.Fill()

	for _, path := range paths {
		if len(path) > 2 {
			gc.MoveTo(path[0].X, path[0].Y)
			for i := 1; i < len(path); i++ {
				gc.LineTo(path[i].X, path[i].Y)
			}
			//			gc.ClosePath()
		}
	}
	gc.SetLineWidth(4)
	gc.SetColor(color.NRGBA{0xCC, 0xAA, 0x88, 0xFF})
	gc.Stroke()

	//opts := Options{PixelPlane: true}

	// draw outline
	if true {
		for i, path := range paths {
			rect := path.Rect()
			gc.MoveTo(rect.Min.X, rect.Min.Y)
			gc.LineTo(rect.Max.X, rect.Min.Y)
			gc.LineTo(rect.Max.X, rect.Max.Y)
			gc.LineTo(rect.Min.X, rect.Max.Y)
			gc.LineTo(rect.Min.X, rect.Min.Y)
			gc.SetLineWidth(1)
			//if i == 2 {
			//	reverseWinding(path)
			//}
			if pathIsClockwise(path) {
				gc.SetColor(color.NRGBA{0, 0, 0xff, 0xFF})
			} else {
				gc.SetColor(color.NRGBA{0xff, 0, 0, 0xFF})
			}
			gc.Stroke()
			gc.DrawString(fmt.Sprintf("%d", i), rect.Min.X+2, rect.Min.Y+12)
			gc.Fill()
			if above, ok := aboveMap[i]; ok {
				inside := pnpoly(path, above)

				if !inside {
					gc.SetColor(color.NRGBA{0, 0, 0, 0xFF})
				}
				gc.DrawLine(above.X, above.Y, (rect.Max.X-rect.Min.X)/2+rect.Min.X, (rect.Max.Y-rect.Min.Y)/2+rect.Min.Y)
				gc.Stroke()
				gc.DrawCircle(above.X, above.Y, 6)
				gc.DrawCircle((rect.Max.X-rect.Min.X)/2+rect.Min.X, (rect.Max.Y-rect.Min.Y)/2+rect.Min.Y, 3)
				gc.Fill()
			}
		}
	}
	return gc.SavePNG(filePath)
}
