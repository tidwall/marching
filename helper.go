package marching

import (
	"image"
	"math"
	"sort"
)

// TerrariumValues ...
func TerrariumValues(img *image.RGBA) []float64 {
	width, height := img.Bounds().Size().X, img.Bounds().Size().Y
	values := make([]float64, width*height)
	for i, j := 0, 0; i < len(img.Pix); i, j = i+4, j+1 {
		values[j] = (float64(img.Pix[i+0])*256 + float64(img.Pix[i+1]) +
			float64(img.Pix[i+2])/256) - 32768
	}
	return values
}

func valueForXY(values []float64, width, height int, x, y int) float64 {
	if x < 0 {
		x = 0
	} else if x > width-1 {
		x = width - 1
	}
	if y < 0 {
		y = 0
	} else if y > height-1 {
		y = height - 1
	}
	return values[y*width+x]
}

func BilinearInterpolationValues(values []float64, width, height int) (nvalues []float64, nwidth, nheight int) {
	nwidth = width*2 - 1
	nheight = height*2 - 1
	nvalues = make([]float64, nwidth*nheight)

	for y := 0; y < nheight; y++ {
		for x := 0; x < nwidth; x++ {
			var v float64
			if y%2 == 0 {
				if x%2 == 0 {
					v = valueForXY(values, width, height, x/2, y/2)
				} else {
					v1 := valueForXY(values, width, height, x/2, y/2)
					v2 := valueForXY(values, width, height, x/2+1, y/2)
					v = (v1 + v2) / 2
				}
			} else {
				if x%2 == 0 {
					v1 := valueForXY(values, width, height, x/2, y/2)
					v2 := valueForXY(values, width, height, x/2, y/2+1)
					v = (v1 + v2) / 2
				} else {
					v1 := valueForXY(values, width, height, x/2, y/2)
					v2 := valueForXY(values, width, height, x/2+1, y/2)
					v3 := valueForXY(values, width, height, x/2, y/2+1)
					v4 := valueForXY(values, width, height, x/2+1, y/2+1)
					v = (v1 + v2 + v3 + v4) / 4
				}
			}
			nvalues[y*nheight+x] = v
		}
	}
	return nvalues, nwidth, nheight
}

// bilinearInterpolation return the value that is contained between four
// points. The x and y params must be between 0.0 - 1.0.
func bilinearInterpolation(vals [4]float64, x, y float64) float64 {
	return vals[3]*(1-x)*y + vals[2]*x*y + vals[0]*(1-x)*(1-y) + vals[1]*x*(1-y)
}

// linearInterpolation return the value that is contained between two
// points. The x params must be between 0.0 - 1.0.
func linearInterpolation(vals [2]float64, x float64) float64 {
	return (vals[1]-vals[0])*x + vals[0]
}

func SimplifyPaths(paths [][][2]float64, amount float64) [][][2]float64 {
	paths = reducePathPoints(paths)
	var npaths [][][2]float64
	for i := 0; i < len(paths); i++ {
		var npath [][2]float64
		npath = simplifyPath(paths[i], amount)
		npaths = append(npaths, npath)
	}
	return npaths
}
func simplifyPath(points [][2]float64, amount float64) [][2]float64 {
	if len(points) <= 3 {
		return points
	}
	npoints := append([][2]float64{}, points...)
	count := int(float64(len(npoints)) * (1 - amount))
	for ; count > 0; count-- {
		//println(count, len(npoints))
		var minArea float64
		var minIndex int
		for i := 0; i < len(npoints)-2; i++ {
			area := area(npoints[i+0], npoints[i+1], npoints[i+2])
			if i == 0 || area < minArea {
				minArea = area
				minIndex = i + 1
			}
		}
		//fmt.Printf("%d %f\n", minIndex, minArea)
		npoints = append(npoints[:minIndex], npoints[minIndex+1:]...)
		if npoints[0] != npoints[len(npoints)-1] {
			npoints = append(npoints, npoints[0])
		}

	}
	return npoints

	type triangle struct {
		points  [][2]float64
		area    float64
		segment int
	}
	var triangles []triangle
	for i := 0; i < len(points)-2; i++ {
		area := area(points[i+0], points[i+1], points[i+2])
		triangles = append(triangles, triangle{
			points: points[i : i+3], area: area,
			segment: i,
		})
	}
	sort.Slice(triangles, func(i, j int) bool {
		return triangles[i].area < triangles[j].area
	})
	triangles = triangles[len(triangles)-int(float64(len(triangles))*amount):]
	sort.Slice(triangles, func(i, j int) bool {
		return triangles[i].segment < triangles[j].segment
	})
	// var npoints [][2]float64
	// for i := 0; i < len(triangles); i++ {
	// 	if len(npoints) > 0 {
	// 		if npoints[len(npoints)-1] == triangles[i].points[0] {
	// 			npoints = npoints[:len(npoints)-1]
	// 		}
	// 	}
	// 	npoints = append(npoints, triangles[i].points...)
	// }
	if npoints[0] != npoints[len(npoints)-1] {
		npoints = append(npoints, npoints[0])
	}

	return npoints
}

func interpolate(p0, p1 [2]float64, t float64) [2]float64 {
	return [2]float64{p0[0] + t*(p1[0]-p0[0]), p0[1] + t*(p1[1]-p0[1])}
}

func newCurveFn(pts [][2]float64) func(t float64) [2]float64 {
	if len(pts) < 2 {
		return nil
	} else if len(pts) == 2 {
		return func(t float64) [2]float64 {
			return interpolate(pts[0], pts[1], t)
		}
	} else if len(pts) == 3 {
		return func(t float64) [2]float64 {
			var a, b = interpolate(pts[0], pts[1], t), interpolate(pts[1], pts[2], t)
			return interpolate(a, b, t)
		}
	}
	var midPts = append(pts, make([][2]float64, len(pts)*(len(pts)-1)/2)...)
	return func(t float64) [2]float64 {
		for m, n := len(pts), len(pts)-1; n > 0; n-- {
			for i := 0; i < n; i++ {
				midPts[m+i] = interpolate(midPts[m+i-n-1], midPts[m+i-n], t)
			}
			m += n
		}
		return midPts[len(midPts)-1]
	}
}

func lineLength(a, b [2]float64) float64 {
	return math.Sqrt((a[0]-b[0])*(a[0]-b[0]) + (a[1]-b[1])*(a[1]-b[1]))
}

func pathLength(path [][2]float64) float64 {
	var length float64
	for i := 0; i < len(path)-1; i++ {
		length += lineLength(path[i], path[i+1])
	}
	return length
}

func Curve(values []float64, width, height int, level float64) [][][2]float64 {
	paths := Lines(values, width, height, level)
	var npaths [][][2]float64
	for i := 0; i < len(paths); i++ {
		var fn = newCurveFn(paths[i])
		var npath [][2]float64
		length := pathLength(paths[i])
		segments := int(math.Ceil(length) / 4)
		if segments < 4 {
			segments = 4
		}
		npath = append(npath, paths[i][0])
		for t := 1; t <= segments; t++ {
			var cur = fn(float64(t) / float64(segments))
			npath = append(npath, [2]float64{cur[0], cur[1]})
		}
		npaths = append(npaths, npath)
	}
	return npaths
}
