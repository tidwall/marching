package marching

func makeSegments(cells []cellT, width, height int, level float64) [][2]float64 {
	width, height = width-1, height-1
	var segs [][2]float64
	var i int
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			cell := cells[i]
			//vals := cell.Values
			//fmt.Printf("%v,%v %v\n", x, y, cell)
			a := [2]float64{float64(x + 1), float64(y + 1)}
			b := [2]float64{float64(x + 1), float64(y + 1)}
			c := [2]float64{float64(x + 1), float64(y + 1)}
			d := [2]float64{float64(x + 1), float64(y + 1)}
			// o == above
			// • == below
			switch cell.Case {
			default:
				println(cell.Case)
				panic("invalid case")
			case 0:
				// o---------o
				// |         |
				// |         |
				// |         |
				// o---------o
			case 15:
				// •---------•
				// |         |
				// |         |
				// |         |
				// •---------•
			case 1:
				// o---------o
				// |         |
				// |\        |
				// | \       |
				// •---------o
				// BOTTOM -> LEFT
				a = bottom(a, 3, cell, level)
				b = left(b, 0, cell, level)
			case 2:
				// o---------o
				// |         |
				// |        /|
				// |       / |
				// o---------•
				// RIGHT -> BOTTOM
				a = right(a, 1, cell, level)
				b = bottom(b, 2, cell, level)
			case 3:
				// o---------o
				// |         |
				// |---------|
				// |         |
				// •---------•
				// RIGHT -> LEFT
				a = right(a, 1, cell, level)
				b = left(b, 0, cell, level)
			case 4:
				// o---------•
				// |       \ |
				// |        \|
				// |         |
				// o---------o
				// TOP -> RIGHT
				a = top(a, 0, cell, level)
				b = right(b, 3, cell, level)
			case 5:
				if !cell.CenterAbove {
					// center below
					// o---------•
					// | /       |
					// |/       /|
					// |       / |
					// •---------o
					// TOP -> LEFT
					a = top(a, 0, cell, level)
					b = left(b, 0, cell, level)
					// BOTTOM -> RIGHT
					c = bottom(c, 3, cell, level)
					d = right(d, 3, cell, level)
				} else {
					// center above
					// o---------•
					// |       \ |
					// |\       \|
					// | \       |
					// •---------o
					// TOP -> RIGHT
					a = top(a, 0, cell, level)
					b = right(b, 3, cell, level)
					// BOTTOM -> LEFT
					c = bottom(c, 3, cell, level)
					d = left(d, 0, cell, level)
				}
			case 6:
				// o---------•
				// |    |    |
				// |    |    |
				// |    |    |
				// o---------•
				// TOP -> BOTTOM
				a = top(a, 0, cell, level)
				b = bottom(b, 2, cell, level)
			case 7:
				// o---------•
				// | /       |
				// |/        |
				// |         |
				// •---------•
				// TOP -> LEFT
				a = top(a, 0, cell, level)
				b = left(b, 0, cell, level)
			case 8:
				// •---------o
				// | /       |
				// |/        |
				// |         |
				// o---------o
				// LEFT -> TOP
				a = left(a, 2, cell, level)
				b = top(b, 1, cell, level)
			case 9:
				// •---------o
				// |    |    |
				// |    |    |
				// |    |    |
				// •---------o
				// BOTTOM -> TOP
				a = bottom(a, 3, cell, level)
				b = top(b, 1, cell, level)
			case 10:
				if !cell.CenterAbove {
					// center below
					// •---------o
					// |       \ |
					// |\       \|
					// | \       |
					// o---------•
					// RIGHT -> TOP
					a = right(a, 1, cell, level)
					b = top(b, 1, cell, level)
					// LEFT -> BOTTOM
					c = left(c, 2, cell, level)
					d = bottom(d, 2, cell, level)
				} else {
					// center above
					// •---------o
					// | /       |
					// |/       /|
					// |       / |
					// o---------•
					// TOP -> LEFT
					a = top(a, 1, cell, level)
					b = left(b, 2, cell, level)
					// BOTTOM -> RIGHT
					c = bottom(c, 2, cell, level)
					d = right(d, 1, cell, level)
				}
			case 11:
				// •---------o
				// |       \ |
				// |        \|
				// |         |
				// •---------•
				// RIGHT -> TOP
				a = right(a, 1, cell, level)
				b = top(b, 1, cell, level)
			case 12:
				// •---------•
				// |         |
				// |---------|
				// |         |
				// o---------o
				// LEFT -> RIGHT
				a = left(a, 2, cell, level)
				b = right(b, 3, cell, level)
			case 13:
				// •---------•
				// |         |
				// |        /|
				// |       / |
				// •---------o
				// BOTTOM -> RIGHT
				//a[1] = a[1] + 0.5
				a = bottom(a, 3, cell, level)
				b = right(b, 3, cell, level)
			case 14:
				// •---------•
				// |         |
				// |\        |
				// | \       |
				// o---------•
				// LEFT -> BOTTOM
				a = left(a, 2, cell, level)
				b = bottom(b, 2, cell, level)
			}
			if a != b {
				segs = append(segs, a, b)
			}
			if c != d {
				segs = append(segs, c, d)
			}
			i++
		}
	}

	return segs
}

func left(p [2]float64, corner int, cell cellT, level float64) [2]float64 {
	p[0] = p[0] - 0.5
	switch corner {
	default:
		panic("invalid corner, must be 0 or 2")
	case 0:
		p[1] = p[1] + 0.5 - lint(cell.Values[0], cell.Values[2], level)
	case 2:
		p[1] = p[1] - 0.5 + lint(cell.Values[2], cell.Values[0], level)
	}
	return p
}
func right(p [2]float64, corner int, cell cellT, level float64) [2]float64 {
	p[0] = p[0] + 0.5
	switch corner {
	default:
		panic("invalid corner, must be 1 or 3")
	case 1:
		p[1] = p[1] + 0.5 - lint(cell.Values[1], cell.Values[3], level)
	case 3:
		p[1] = p[1] - 0.5 + lint(cell.Values[3], cell.Values[1], level)
	}
	return p
}
func bottom(p [2]float64, corner int, cell cellT, level float64) [2]float64 {
	switch corner {
	default:
		panic("invalid corner, must be 2 or 3")
	case 2:
		p[0] = p[0] + 0.5 - lint(cell.Values[2], cell.Values[3], level)
	case 3:
		p[0] = p[0] - 0.5 + lint(cell.Values[3], cell.Values[2], level)
	}
	p[1] = p[1] + 0.5
	return p
}
func top(p [2]float64, corner int, cell cellT, level float64) [2]float64 {
	// var q float64
	// if cell.Values[0] < cell.Values[1] {
	// 	q = lint(cell.Values[1], cell.Values[0], level)
	// 	//fmt.Printf("%v\n",
	// } else {
	// 	q = lint(cell.Values[1], cell.Values[0], level)
	// 	//fmt.Printf("%v\n", lint(cell.Values[1], cell.Values[0], level))
	// }
	// p[0] = p[0] + 0.7
	// fmt.Printf("%v\n", q)
	switch corner {
	default:
		panic("invalid corner, must be 0 or 1")
	case 0:
		p[0] = p[0] + 0.5 - lint(cell.Values[0], cell.Values[1], level)
	case 1:
		p[0] = p[0] - 0.5 + lint(cell.Values[1], cell.Values[0], level)
	}
	p[1] = p[1] - 0.5
	return p
}
