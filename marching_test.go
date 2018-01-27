package marching

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"math"
	"testing"

	"github.com/fogleman/gg"
)

var (
	// testAValues = []float64{
	// 	1, 1, 1, 1, 1,
	// 	1, 2, 3, 2, 1,
	// 	1, 3, 0, 3, 1,
	// 	1, 2, 3, 2, 1,
	// 	1, 1, 1, 1, 1,
	// }
	testAValues = []float64{
		3, 3, 3, 3, 3,
		3, 3, 3, 3, 3,
		3, 1, 1, 1, 3,
		3, 1, 1, 1, 3,
		3, 3, 3, 3, 3,
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
	values = testAValues
	width = testAWidth
	height = testAHeight
	level = testALevel
	paths := Paths(values, width, height, level, true)
	testSavePaths(paths, values,
		float64(width), float64(height),
		256, 256,
		"testpaths.png")
}

func generateSamples(img image.Image) (values []float64, min, max float64) {
	rgba, ok := img.(*image.RGBA)
	if !ok {
		// convert to RGBA
		rgba = image.NewRGBA(img.Bounds())
		draw.Draw(rgba, img.Bounds(), img, image.Pt(0, 0), draw.Src)
		return generateSamples(rgba)
	}
	values = make([]float64, len(rgba.Pix)/4)
	for i := 0; i < len(rgba.Pix); i += 4 {
		red := float64(rgba.Pix[i+0])
		green := float64(rgba.Pix[i+1])
		blue := float64(rgba.Pix[i+2])
		value := (red*256 + green + blue/256) - 32768
		values[i/4] = value
		if i == 0 {
			min, max = value, value
		} else if value < min {
			min = value
		} else if value > max {
			max = value
		}
	}
	return
}

func testSavePaths(paths [][][2]float64, values []float64,
	orgWidth, orgHeight float64,
	imgWidth, imgHeight float64,
	filePath string) error {
	gc := gg.NewContext(int(imgWidth), int(imgHeight))
	gc.SetColor(color.White)
	gc.DrawRectangle(0, 0, imgWidth, imgHeight)
	gc.Fill()

	// draw cell grid
	gc.SetDash(1, 2)
	gc.SetLineWidth(0.25)
	gc.SetColor(color.Black)
	for y := 0; y < int(orgWidth)-1; y++ {
		for x := 0; x < int(orgHeight)-1; x++ {
			gc.MoveTo((float64(x)+0.5)/orgWidth*imgWidth, (float64(y)+0.5)/orgHeight*imgHeight)
			gc.LineTo((float64(x+1)+0.5)/orgWidth*imgWidth, (float64(y)+0.5)/orgHeight*imgHeight)
			gc.LineTo((float64(x+1)+0.5)/orgWidth*imgWidth, (float64(y+1)+0.5)/orgHeight*imgHeight)
			gc.LineTo((float64(x)+0.5)/orgWidth*imgWidth, (float64(y+1)+0.5)/orgHeight*imgHeight)
			gc.LineTo((float64(x)+0.5)/orgWidth*imgWidth, (float64(y)+0.5)/orgHeight*imgHeight)
			gc.Stroke()
		}
	}

	// draw sample values
	for y := 0; y < int(orgWidth); y++ {
		for x := 0; x < int(orgHeight); x++ {
			value := values[y*int(orgWidth)+x]
			var s string
			if value == math.Floor(value) {
				s = fmt.Sprintf("%.0f", value)
			} else {
				s = fmt.Sprintf("%.1f", value)
			}
			sw, sh := gc.MeasureString(s)
			gc.DrawString(s,
				float64(x)/orgWidth*imgWidth+imgWidth/orgWidth/2-sw/2,
				float64(y)/orgHeight*imgHeight+imgHeight/orgHeight/2+sh/2-2,
			)
		}
	}

	gc.SetColor(color.NRGBA{0xCC, 0x66, 0x66, 0x20})
	for _, path := range paths {
		if len(path) < 2 {
			continue
		}
		for i := 0; i < len(path); i++ {
			if i == 0 {
				gc.MoveTo(path[i][0]/orgWidth*imgWidth, path[i][1]/orgHeight*imgHeight)
			} else {
				gc.LineTo(path[i][0]/orgWidth*imgWidth, path[i][1]/orgHeight*imgHeight)
			}
		}
		gc.ClosePath()
	}
	gc.Fill()

	gc.SetDash()
	gc.SetFillRuleWinding()
	gc.SetLineWidth(2)
	gc.SetColor(color.NRGBA{0xCC, 0x66, 0x66, 0xFF})
	for _, path := range paths {
		if len(path) < 2 {
			continue
		}
		for i := 0; i < len(path); i++ {
			if i == 0 {
				gc.MoveTo(path[i][0]/orgWidth*imgWidth, path[i][1]/orgHeight*imgHeight)
			} else {
				gc.LineTo(path[i][0]/orgWidth*imgWidth, path[i][1]/orgHeight*imgHeight)
			}
		}
		gc.ClosePath()
	}
	gc.Stroke()

	gc.SetColor(color.NRGBA{0xCC, 0x66, 0x66, 0xFF})
	for _, path := range paths {
		if len(path) < 2 {
			continue
		}
		for i := 0; i < len(path); i++ {
			gc.DrawCircle(path[i][0]/orgWidth*imgWidth, path[i][1]/orgHeight*imgHeight, 3)
			gc.Fill()
		}
	}

	return gc.SavePNG(filePath)
}

func TestWall(t *testing.T) {
	tequals := func(a, b int) {
		if a != b {
			panic("mismatch")
		}
	}
	tequals(wallIndexForSide(top, 1, 1, 3), 8)
	tequals(wallIndexForSide(top, 1, 2, 3), 15)
	tequals(wallIndexForSide(bottom, 1, 1, 3), 15)
	tequals(wallIndexForSide(bottom, 1, 2, 3), 22)
	tequals(wallIndexForSide(left, 1, 1, 3), 11)
	tequals(wallIndexForSide(left, 1, 2, 3), 18)
	tequals(wallIndexForSide(right, 1, 1, 3), 12)
	tequals(wallIndexForSide(right, 1, 2, 3), 19)
}
