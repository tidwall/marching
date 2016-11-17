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
	if width <= 2 || height <= 2 {
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
