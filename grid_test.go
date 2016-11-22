package marching

import (
	"image"
	"image/draw"
	"image/png"
	"os"
	"testing"
	"time"
)

var (
	testAValues = []float64{
		1, 1, 1, 1, 1,
		1, 2, 3, 2, 1,
		1, 3, 3, 3, 1,
		1, 2, 3, 2, 1,
		1, 1, 1, 1, 1,
	}
	testAWidth          = 5
	testAHeight         = 5
	testALevel  float64 = 2
	testACases          = []byte{
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

func TestTerrarium(t *testing.T) {
	f, err := os.Open("12_770_1644-12_774_1647.png")
	//f, err := os.Open("768.png")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	img1, err := png.Decode(f)
	if err != nil {
		t.Fatal(err)
	}
	img := image.NewRGBA(img1.Bounds())
	draw.Draw(img, img.Bounds(), img1, image.ZP, draw.Src)
	width, height := img.Bounds().Size().X, img.Bounds().Size().Y
	values := make([]float64, width*height)
	for i, j := 0, 0; i < len(img.Pix); i, j = i+4, j+1 {
		red := float64(img.Pix[i+0])
		green := float64(img.Pix[i+1])
		blue := float64(img.Pix[i+2])
		meters := (red*256 + green + blue/256) - 32768
		values[j] = meters
	}
	start := time.Now()
	start2 := time.Now()
	grid := NewGrid(values, width, height, 700, 0)
	println("** NewGrid:", time.Now().Sub(start2).String())
	paths, aboveMap := grid.Paths(float64(width), float64(height), nil)
	println(time.Now().Sub(start).String())
	if err := savePathsPNG(grid, paths, aboveMap, width, height, "terrarium.png"); err != nil {
		t.Fatal(err)
	}
	return
}
func TestGrid(t *testing.T) {
	//grid := NewGrid(testAValues, testAWidth, testAHeight, testALevel)
	start := time.Now()
	start2 := time.Now()
	values, width, height, level := testBValues, testBWidth, testBHeight, testBLevel
	complexity := 0
	grid := NewGrid(values, width, height, level, complexity)
	println("** NewGrid:", time.Now().Sub(start2).String())
	paths, aboveMap := grid.Paths(500, 500, nil)
	println(time.Now().Sub(start).String())
	if err := savePathsPNG(grid, paths, aboveMap, 500, 500, "testpaths.png"); err != nil {
		t.Fatal(err)
	}
	return
	//if len(grid.Cells) != (width-1)*(height-1) {
	//	t.Fatalf("expected %v, got %v", (width-1)*(height-1), len(grid.Cells))
	//}
	println(grid.Cells[0].Case, len(grid.Cells))
	//	return
	/*
		if len(grid.Cells) != len(testACases) {
			t.Fatalf("expected %v, got %v", len(testACases), len(grid.Cells))
		}
			if false {
				for i := 0; i < len(grid.Cells); i++ {
					if testACases[i] != grid.Cells[i].Case {
						t.Fatalf("expected %v, got %v for #%d", testACases[i], grid.Cells[i].Case, i)

					}
				}
			}
	*/
	/*
		paths := grid.Paths(500, 500, nil)
		println(time.Now().Sub(start).String())
		println(paths)
		return
		img := grid.Image(500, 500, &ImageOptions{
			Marks: true, //false, //true,
			//FillColor:   color.NRGBA{0xff, 0, 0, 0xff},
			//StrokeColor: color.NRGBA{0, 0, 0, 0xff},
			//NoStroke:    true,
			//LineWidth: 10,
			//ExpandEdges: true,
		})
		println(time.Now().Sub(start).String())
		var buf bytes.Buffer
		if err := png.Encode(&buf, img); err != nil {
			t.Fatal(err)
		}
		ioutil.WriteFile("testgrid.png", buf.Bytes(), 0600)
	*/
}

/*
func TestSpline(t *testing.T) {
	X := []float64{
		0, 1, 2, 3, 4, 5, 6, 7, 8, 9,
	}
	Y := []float64{
		5, 20, 10, 13, 4, 1, 8, 12, 14, 9,
	}
	s := spline.Spline{}

}
*/
