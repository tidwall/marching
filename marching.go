package marching

const (
	doALGO2  = true
	doLERP   = true
	doOFFSET = true
)

type calcCellT struct {
	x, y    int
	corners [4]float64
	level   float64
	gwidth  int // grid width
	gheight int // grid height
}

type endpointT struct {
	point   [2]float64
	pathIdx int
	next    *endpointT
}

func Paths(samples []float64, width, height int, level float64) [][][2]float64 {
	if len(samples) != width*height {
		panic("number of values are not equal to width multiplied by height")
	}
	if width <= 2 || height <= 2 {
		panic("width or height are not greater than or equal to two")
	}
	var cell calcCellT
	cell.level = level
	cell.gwidth = width - 1   // grid width
	cell.gheight = height - 1 // grid height

	endpoints := make(map[int]*endpointT) // wall -> point
	var llpaths []*endpointT              // unique paths
	var sides [2][2]side
	var numLines int
	// store each line in an endpoints hash
	for cell.y = 0; cell.y < cell.gheight; cell.y++ {
		for cell.x = 0; cell.x < cell.gwidth; cell.x++ {
			cell.corners[0] = samples[(cell.y+0)*width+(cell.x+0)]
			cell.corners[1] = samples[(cell.y+0)*width+(cell.x+1)]
			cell.corners[2] = samples[(cell.y+1)*width+(cell.x+0)]
			cell.corners[3] = samples[(cell.y+1)*width+(cell.x+1)]
			var casz byte
			if cell.corners[0] < level {
				// top-left
				casz |= 0x8
			}
			if cell.corners[1] < level {
				// top-right
				casz |= 0x4
			}
			if cell.corners[2] < level {
				// bottom-left
				casz |= 0x1
			}
			if cell.corners[3] < level {
				// bottom-right
				casz |= 0x2
			}
			//fmt.Printf("%d%d\n", cell.x, cell.y)
			numLines = 1
			switch casz {
			case 0:
				// o---------o
				// |         |
				// |         |
				// |         |
				// o---------o
				// all is above
				numLines = 0
			case 15:
				// •---------•
				// |         |
				// |         |
				// |         |
				// •---------•
				// all is below
				numLines = 0
			case 1:
				// o---------o
				// |         |
				// |\        |
				// | \       |
				// •---------o
				sides[0][0], sides[0][1] = bottom, left
			case 2:
				// o---------o
				// |         |
				// |        /|
				// |       / |
				// o---------•
				sides[0][0], sides[0][1] = right, bottom
			case 3:
				// o---------o
				// |         |
				// |---------|
				// |         |
				// •---------•
				sides[0][0], sides[0][1] = right, left
			case 4:
				// o---------•
				// |       \ |
				// |        \|
				// |         |
				// o---------o
				sides[0][0], sides[0][1] = top, right
			case 5:
				numLines = 2
				// determine if center of the cell is above the level. this is used
				// to swap saddle points when needed.
				above := (cell.corners[0]+cell.corners[1]+
					cell.corners[2]+cell.corners[3])/4 >= level
				if !above {
					// center below
					// o---------•
					// | /       |
					// |/       /|
					// |       / |
					// •---------o
					sides[0][0], sides[0][1] = top, left
					sides[1][0], sides[1][1] = bottom, right
				} else {
					// center above
					// o---------•
					// |       \ |
					// |\       \|
					// | \       |
					// •---------o
					sides[0][0], sides[0][1] = top, right
					sides[1][0], sides[1][1] = bottom, left
				}
			case 6:
				// o---------•
				// |    |    |
				// |    |    |
				// |    |    |
				// o---------•
				sides[0][0], sides[0][1] = top, bottom
			case 7:
				// o---------•
				// | /       |
				// |/        |
				// |         |
				// •---------•
				sides[0][0], sides[0][1] = top, left
			case 8:
				// •---------o
				// | /       |
				// |/        |
				// |         |
				// o---------o
				sides[0][0], sides[0][1] = left, top
			case 9:
				// •---------o
				// |    |    |
				// |    |    |
				// |    |    |
				// •---------o
				sides[0][0], sides[0][1] = bottom, top
			case 10:
				numLines = 2
				// determine if center of the cell is above the level. this is used
				// to swap saddle points when needed.
				above := (cell.corners[0]+cell.corners[1]+
					cell.corners[2]+cell.corners[3])/4 >= level
				if !above {
					// center below
					// •---------o
					// |       \ |
					// |\       \|
					// | \       |
					// o---------•
					sides[0][0], sides[0][1] = right, top
					sides[1][0], sides[1][1] = left, bottom
				} else {
					// center above
					// •---------o
					// | /       |
					// |/       /|
					// |       / |
					// o---------•
					sides[0][0], sides[0][1] = left, top
					sides[1][0], sides[1][1] = right, bottom
				}
			case 11:
				// •---------o
				// |       \ |
				// |        \|
				// |         |
				// •---------•
				sides[0][0], sides[0][1] = right, top
			case 12:
				// •---------•
				// |         |
				// |---------|
				// |         |
				// o---------o
				sides[0][0], sides[0][1] = left, right
			case 13:
				// •---------•
				// |         |
				// |        /|
				// |       / |
				// •---------o
				sides[0][0], sides[0][1] = bottom, right
			case 14:
				// •---------•
				// |         |
				// |\        |
				// | \       |
				// o---------•
				sides[0][0], sides[0][1] = left, bottom
			}
			// add each side to the endpoints hash
			// only calculate the points as needed
			for i := 0; i < numLines; i++ {
				wallA := wallIndexForSide(sides[i][0], cell.x, cell.y, cell.gwidth)
				wallB := wallIndexForSide(sides[i][1], cell.x, cell.y, cell.gwidth)
				ptA := endpoints[wallA]
				ptB := endpoints[wallB]
				if ptA == nil {
					ptA = new(endpointT)
					ptA.point = cell.calcPoint(sides[i][0])
					endpoints[wallA] = ptA
					if ptB == nil {
						ptA.pathIdx = len(llpaths)
						llpaths = append(llpaths, ptA)
					} else {
						ptA.pathIdx = ptB.pathIdx
						llpaths[ptA.pathIdx] = ptA
					}
				}
				if ptB == nil {
					ptB = new(endpointT)
					ptB.point = cell.calcPoint(sides[i][1])
					endpoints[wallB] = ptB
					ptB.pathIdx = ptA.pathIdx
				}
				ptA.next = ptB

				if ptA.pathIdx != ptB.pathIdx {
					// Must joined two different paths.
					// drop the previous path
					llpaths[ptB.pathIdx] = nil
					// update the pathB indexes
					pt := ptB
					for pt != nil && pt.pathIdx != ptA.pathIdx {
						pt.pathIdx = ptA.pathIdx
						pt = pt.next
					}
				}
			}
		}
	}

	var paths [][][2]float64
	for _, llpath := range llpaths {
		var path [][2]float64
		pt := llpath
		var last [2]float64
		for pt != nil {
			if pt.point != last {
				path = append(path, pt.point)
				last = pt.point
			}
			pt = pt.next
			if pt == llpath {
				if pt.point != last {
					path = append(path, pt.point)
					last = pt.point
				}
				break
			}
		}
		if len(path) > 1 {
			paths = append(paths, path)
		}
	}
	return paths
}

type side byte

const (
	top    side = 0
	left   side = 1
	right  side = 2
	bottom side = 3
)

func (cell *calcCellT) calcPos(coord int, a, b int) (pos float64) {
	if cell.corners[a] < cell.corners[b] {
		pos = float64(coord) + lerp(cell.corners[a], cell.corners[b], cell.level)
	} else {
		pos = float64(coord) + 1 - lerp(cell.corners[b], cell.corners[a], cell.level)
	}
	return
}
func (cell *calcCellT) calcPoint(side side) (point [2]float64) {
	switch side {
	case top:
		point[0] = cell.calcPos(cell.x, 0, 1)
		point[1] = float64(cell.y)
	case right:
		point[0] = float64(cell.x) + 1
		point[1] = cell.calcPos(cell.y, 1, 3)
	case bottom:
		point[0] = cell.calcPos(cell.x, 2, 3)
		point[1] = float64(cell.y) + 1
	case left:
		point[0] = float64(cell.x)
		point[1] = cell.calcPos(cell.y, 0, 2)
	}
	if !doOFFSET {
		point[0] += 0.5
		point[1] += 0.5
	}
	return
}

func lerp(below, above, level float64) float64 {
	if doLERP {
		return (1.0 - ((level - above) / (below - above)))
	}
	return 0.5
}

// ��---0---•
// |       |
// 1       2
// |       |
// •---3---•
//
// •---0---•---1---•
// |       |       |
// 2       3       4
// |       |       |
// •---5---•---6---•
//
// •---0---•---1---•---2---•
// |       |       |       |
// 3       4       5       6
// |       |       |       |
// •---7---•---8---•---9---•
//
// •---0---•---1---•---2---•---3---•
// |       |       |       |       |
// 4       5       6       7       8
// |       |       |       |       |
// •---9---•--10---•---11--•--12---•
//
// •---0---•---1---•---2---•
// |       |       |       |
// 3       4       5       6
// |       |       |       |
// •---7---•---8---•---9---•
// |       |       |       |
// 10     11       12     13
// |       |       |       |
// •--14---•---15--•---16--•
// |       |       |       |
// 17     18       19     20
// |       |       |       |
// •--21---•---22--•---23--•

func wallIndexForSide(side side, x, y, width int) int {
	n := width*2 + 1 // rowscan
	switch side {
	case top:
		return y*n + x
	case left:
		return y*n + width + x
	case right:
		return y*n + width + x + 1
	case bottom:
		return (y+1)*n + x
	}
	panic("invalid side")
}
