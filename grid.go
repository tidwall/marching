package marching

// Cell represents a single isoline square.
type Cell struct {
	// Case field can be a value between 0-15.
	Case byte
	// CenterAbove field indicates that the value in center of the cell
	// is above the level that was passed to NewGrid().
	CenterAbove bool
}

// Grid represents a grid of isoline cells.
type Grid struct {
	// Cells field is an array of isoline cells in pixel coordinates where
	// the cell located at position (0,0) is the top-left cell and is at
	// index zero. The cell at position (Width,Height) is the bottom-right cell
	// and is the last item in the Cell array.
	Cells []Cell
	// Width is the width of the grid. This value is one less than the
	// original width of the values that were passed to NewGrid().
	Width int
	// Height is the height of the grid. This value is one less than the
	// original height of the values that were passed to NewGrid().
	Height int
}

// NewGrid generates a grid of isoline cells from a series of values.
// The resulting Grid contains cells with case indexes compared to the
// level param.
// The complexity param can be used to increase or decrease the number
// of grid cells.
// Using a complexity of zero will result in a grid with the default
// number of cells.
func NewGrid(values []float64, width, height int, level float64, complexity int) *Grid {
	if len(values) != width*height {
		panic("number of values are not equal to width multiplied by height")
	}
	if width <= 2 || height <= 2 {
		panic("width or height are not greater than or equal to two")
	}
	var pcmplx uint
	var ncmplx uint
	var gwidth int
	var gheight int
	if complexity > 0 {
		pcmplx = uint(complexity)
		gwidth = (width - 1) << pcmplx
		gheight = (height - 1) << pcmplx
	} else if complexity < 0 {
		ncmplx = uint(1 << uint(complexity*-1))
		gwidth = (width - 1) >> ncmplx
		gheight = (height - 1) >> ncmplx
	} else {
		gwidth = width - 1
		gheight = height - 1
	}
	println(gwidth, gheight)
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
			} else if complexity > 0 {
				vals[0] = values[((y>>pcmplx)+0)*width+((x>>pcmplx)+0)]
				vals[1] = values[((y>>pcmplx)+0)*width+((x>>pcmplx)+1)]
				vals[2] = values[((y>>pcmplx)+1)*width+((x>>pcmplx)+1)]
				vals[3] = values[((y>>pcmplx)+1)*width+((x>>pcmplx)+0)]
				if complexity > 0 {
					rx := x % (1 << pcmplx)
					ry := y % (1 << pcmplx)
					sx := float64(rx) / float64(int(1<<pcmplx))
					sy := float64(ry) / float64(int(1<<pcmplx))
					ex := sx + 1/float64(int(1<<pcmplx))
					ey := sy + 1/float64(int(1<<pcmplx))
					vals = [4]float64{
						bilinearInterpolation(vals, sx, sy),
						bilinearInterpolation(vals, ex, sy),
						bilinearInterpolation(vals, ex, ey),
						bilinearInterpolation(vals, sx, ey),
					}
				}
			} else {
				vals[0] = values[((y+0)<<ncmplx)*width+((x+0)<<ncmplx)]
				vals[1] = values[((y+0)<<ncmplx)*width+((x+1)<<ncmplx)]
				vals[2] = values[((y+1)<<ncmplx)*width+((x+1)<<ncmplx)]
				vals[3] = values[((y+1)<<ncmplx)*width+((x+0)<<ncmplx)]
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
			cell.CenterAbove = (vals[0]+vals[1]+vals[2]+vals[3])/4 >= level
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
