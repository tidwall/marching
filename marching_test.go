package marching

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"testing"

	"github.com/fogleman/gg"
)

var (
	testAValues = []float64{
		1, 1, 1, 1, 1,
		1, 1, 3, 2, 1,
		1, 3, 3, 3, 1,
		1, 2, 3, 2, 1,
		1, 1, 1, 1, 1,
	}
	testAWidth          = 5
	testAHeight         = 5
	testALevel  float64 = 2

	testACases = []byte{
		13, 12, 12, 14,
		9, 0, 0, 6,
		9, 0, 0, 6,
		11, 3, 3, 7,
	}

	testBValues = []float64{
		2, 1, 3, 3, 1, 3,
		1, 3, 3, 3, 1, 1,
		1, 3, 0, 3, 3, 1,
		3, 3, 3, 3, 3, 1,
		1, 1, 3, 1, 1, 1,
		2, 1, 3, 1, 1, 2,
	}
	testBWidth          = 6
	testBHeight         = 6
	testBLevel  float64 = 2
)

func TestMarching(t *testing.T) {
	var values []float64
	var width, height int
	var level float64
	if false {
		f, _ := os.Open("tiles/14_3098_6576.png")
		img, _ := png.Decode(f)
		values = TerrariumValues(img.(*image.RGBA))
		var min, max float64
		for i, v := range values {
			if i == 0 {
				min, max = v, v
			} else if v < min {
				min = v
			} else if v > max {
				max = v
			}
		}
		level = math.Floor((min + max) / 2)
		width, height = 256, 256
	} else {
		values = testAValues
		width = testAWidth
		height = testAHeight
		level = testALevel
	}

	//values, width, height = BilinearInterpolationValues(values, width, height)
	// values, width, height = BilinearInterpolationValues(values, width, height)
	// values, width, height = BilinearInterpolationValues(values, width, height)
	// values, width, height = BilinearInterpolationValues(values, width, height)

	//paths := Lines(values, width, height, level)
	paths := Curve(values, width, height, level)
	//paths = SimplifyPaths(paths, 0.20)
	testSavePaths(paths, values,
		float64(width), float64(height),
		256, 256,
		"testpaths.png")

	// if err := testSavePathsPNG(nil, paths, nil, 500, 500, "testpaths.png"); err != nil {
	// 	t.Fatal(err)
	// }

}

func testSavePaths(paths [][][2]float64, values []float64,
	orgWidth, orgHeight float64,
	imgWidth, imgHeight float64,
	filePath string) error {
	gc := gg.NewContext(int(imgWidth), int(imgHeight))
	gc.SetColor(color.White)
	gc.DrawRectangle(0, 0, imgWidth, imgHeight)
	gc.Fill()

	// draw grid
	for y := 0; y < int(orgWidth); y++ {
		for x := 0; x < int(orgHeight); x++ {
			value := values[y*int(orgWidth)+x]
			gc.MoveTo(float64(x)/orgWidth*imgWidth, float64(y)/orgHeight*imgHeight)
			gc.LineTo(float64(x+1)/orgWidth*imgWidth, float64(y)/orgHeight*imgHeight)
			gc.LineTo(float64(x+1)/orgWidth*imgWidth, float64(y+1)/orgHeight*imgHeight)
			gc.LineTo(float64(x)/orgWidth*imgWidth, float64(y+1)/orgHeight*imgHeight)
			gc.LineTo(float64(x)/orgWidth*imgWidth, float64(y)/orgHeight*imgHeight)
			//gc.ClosePath()
			gc.SetLineWidth(0.25)
			gc.SetColor(color.NRGBA{0xCC, 0xCC, 0xCC, 0xFF})
			gc.Stroke()
			var s string
			if value == math.Floor(value) {
				s = fmt.Sprintf("%.0f", value)
			} else {
				s = fmt.Sprintf("%.1f", value)
			}
			sw, sh := gc.MeasureString(s)
			gc.DrawString(s,
				float64(x)/orgWidth*imgWidth+imgWidth/orgWidth/2-sw/2,
				float64(y)/orgHeight*imgHeight+imgHeight/orgHeight/2+sh/2,
			)
		}
	}

	// for _, path := range paths {
	// 	if len(path) > 2 {
	// 		gc.MoveTo(path[0][0]/orgWidth*imgWidth, path[0][1]/orgHeight*imgHeight)
	// 		for i := 1; i < len(path); i++ {
	// 			gc.LineTo(path[i][0]/orgWidth*imgWidth, path[i][1]/orgHeight*imgHeight)
	// 		}
	// 		gc.ClosePath()
	// 	}
	// }

	// //	gc.SetFillRuleEvenOdd()
	// gc.SetColor(color.NRGBA{0x88, 0xAA, 0xCC, 0xFF})
	// gc.Fill()

	for _, path := range paths {
		if len(path) > 2 {
			gc.MoveTo(path[0][0]/orgWidth*imgWidth, path[0][1]/orgHeight*imgHeight)
			for i := 1; i < len(path); i++ {
				gc.LineTo(path[i][0]/orgWidth*imgWidth, path[i][1]/orgHeight*imgHeight)
			}
			//gc.ClosePath()
			gc.SetLineWidth(2)
			gc.SetColor(color.NRGBA{0xCC, 0x66, 0x66, 0xFF})
			gc.Stroke()
			for i := 0; i < len(path); i++ {
				gc.DrawCircle(path[i][0]/orgWidth*imgWidth, path[i][1]/orgHeight*imgHeight, 2)
				gc.Fill()

			}
		}
	}

	//opts := Options{PixelPlane: true}

	// // draw outline
	// if true {
	// 	for i, path := range paths {
	// 		min, max := polygon(path).rect()
	// 		gc.MoveTo(min[0], min[1])
	// 		gc.LineTo(max[0], min[1])
	// 		gc.LineTo(max[0], max[1])
	// 		gc.LineTo(min[0], max[1])
	// 		gc.LineTo(min[0], min[1])
	// 		gc.SetLineWidth(1)
	// 		//if i == 2 {
	// 		//	reverseWinding(path)
	// 		//}
	// 		if polygon(path).isClockwise() {
	// 			gc.SetColor(color.NRGBA{0, 0, 0xff, 0xFF})
	// 		} else {
	// 			gc.SetColor(color.NRGBA{0xff, 0, 0, 0xFF})
	// 		}
	// 		gc.Stroke()
	// 		gc.DrawString(fmt.Sprintf("%d", i), min[0]+2, min[1]+12)
	// 		gc.Fill()
	// 	}
	// }
	return gc.SavePNG(filePath)
}
