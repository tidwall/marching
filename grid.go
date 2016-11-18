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
	var vals [4]float64
	var j int
	for y := 0; y < gheight; y++ {
		for x := 0; x < gwidth; x++ {
			var cell Cell
			if complexity == 0 {
				vals[0] = values[(y+0)*width+(x+0)]
				vals[1] = values[(y+0)*width+(x+1)]
				vals[2] = values[(y+1)*width+(x+1)]
				vals[3] = values[(y+1)*width+(x+0)]
			} else {
				vals[0] = values[((y>>cmplx)+0)*width+((x>>cmplx)+0)]
				vals[1] = values[((y>>cmplx)+0)*width+((x>>cmplx)+1)]
				vals[2] = values[((y>>cmplx)+1)*width+((x>>cmplx)+1)]
				vals[3] = values[((y>>cmplx)+1)*width+((x>>cmplx)+0)]
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
			}
			if vals[0] < level {
				cell.Case |= 0x8
			}
			if vals[1] < level {
				cell.Case |= 0x4
			}
			if vals[2] < level {
				cell.Case |= 0x2
			}
			if vals[3] < level {
				cell.Case |= 0x1
			}
			center := (vals[0] + vals[1] + vals[2] + vals[3]) / 4
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
