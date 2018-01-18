// bilinearInterpolation return the value that is contained between four
import "math"

// points. The x and y params must be between 0.0 - 1.0.
func bilinearInterpolation(vals [4]float64, x, y float64) float64 {
	return vals[3]*(1-x)*y + vals[2]*x*y + vals[0]*(1-x)*(1-y) + vals[1]*x*(1-y)
}

// linearInterpolation return the value that is contained between two
// points. The x params must be between 0.0 - 1.0.
func linearInterpolation(vals [2]float64, x float64) float64 {
	return (vals[1]-vals[0])*x + vals[0]
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

func Curve(paths [][][2]float64) [][][2]float64 {
	var npaths [][][2]float64
	for i := 0; i < len(paths); i++ {
		var fn = newCurveFn(paths[i])
		var npath [][2]float64
		length := pathLength(paths[i])
		segments := int(math.Ceil(length))
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
