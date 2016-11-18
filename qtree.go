package marching

const qtreeMaxFill = 16

type qtreeItem interface {
	Pos(ctx interface{}) (x, y float64)
}

type qtree struct {
	ctx        interface{}
	root       *qtreeNode
	minX, minY float64
	maxX, maxY float64
}

type qtreeNode struct {
	items []qtreeItem
	quads [4]*qtreeNode
}

func newQTree(minX, minY, maxX, maxY float64, ctx interface{}) *qtree {
	return &qtree{
		ctx:  ctx,
		root: &qtreeNode{},
		minX: minX, minY: minY,
		maxX: maxX, maxY: maxY,
	}
}

func (tr *qtree) insert(item qtreeItem) {
	x, y := item.Pos(tr.ctx)
	tr.root.insert(item, x, y, tr.minX, tr.minY, tr.maxX, tr.maxY)
}

func (tr *qtree) remove(item qtreeItem) qtreeItem {
	x, y := item.Pos(tr.ctx)
	return tr.root.remove(item, x, y, tr.minX, tr.minY, tr.maxX, tr.maxY)
}

func (tr *qtree) count() int {
	return tr.root.count()
}

func (tr *qtree) scan(iterator func(item qtreeItem, ctx interface{}) bool) {
	tr.root.scan(tr.ctx, iterator)
}
func (tr *qtree) search(minX, minY, maxX, maxY float64, iterator func(item qtreeItem, ctx interface{}) bool) {
	tr.root.search(tr.ctx, minX, minY, maxX, maxY, tr.minX, tr.minY, tr.maxX, tr.maxY, iterator)
}
func qtreeSplitQuad(x, y, minX, minY, maxX, maxY float64) (quad int, qminX, qminY, qmaxX, qmaxY float64) {
	midX := (maxX-minX)/2 + minX
	midY := (maxY-minY)/2 + minY
	if y < midY {
		if x < midX {
			quad, maxX, maxY = 0, midX, midY
		} else {
			quad, minX, maxY = 1, midX, midY
		}
	} else {
		if x < midX {
			quad, maxX, minY = 2, midX, midY
		} else {
			quad, minX, minY = 3, midX, midY
		}
	}
	return quad, minX, minY, maxX, maxY
}

func (n *qtreeNode) insert(item qtreeItem, x, y, minX, minY, maxX, maxY float64) {
	if len(n.items) < qtreeMaxFill {
		n.items = append(n.items, item)
		return
	}
	quad, qminX, qminY, qmaxX, qmaxY := qtreeSplitQuad(x, y, minX, minY, maxX, maxY)
	if n.quads[quad] == nil {
		n.quads[quad] = &qtreeNode{}
	}
	n.quads[quad].insert(item, x, y, qminX, qminY, qmaxX, qmaxY)
}
func (n *qtreeNode) remove(item qtreeItem, x, y, minX, minY, maxX, maxY float64) qtreeItem {
	for i, titem := range n.items {
		if titem == item {
			if len(n.items) == 1 {
				n.items = nil
			} else {
				n.items[i] = n.items[len(n.items)-1]
				n.items[len(n.items)-1] = nil
				n.items = n.items[:len(n.items)-1]
			}
			return titem
		}
	}
	quad, qminX, qminY, qmaxX, qmaxY := qtreeSplitQuad(x, y, minX, minY, maxX, maxY)
	if n.quads[quad] != nil {
		return n.quads[quad].remove(item, x, y, qminX, qminY, qmaxX, qmaxY)
	}
	return nil
}

func (n *qtreeNode) count() int {
	count := len(n.items)
	for quad := 0; quad < 4; quad++ {
		if n.quads[quad] != nil {
			count += n.quads[quad].count()
		}
	}
	return count
}
func (n *qtreeNode) scan(ctx interface{}, iterator func(item qtreeItem, ctx interface{}) bool) bool {
	for _, item := range n.items {
		if !iterator(item, ctx) {
			return false
		}
	}
	for quad := 0; quad < 4; quad++ {
		if n.quads[quad] != nil {
			if !n.quads[quad].scan(ctx, iterator) {
				return false
			}
		}
	}
	return true
}

func (n *qtreeNode) search(
	ctx interface{},
	rminX, rminY, rmaxX, rmaxY float64,
	minX, minY, maxX, maxY float64,
	iterator func(item qtreeItem, ctx interface{}) bool,
) bool {
	for _, item := range n.items {
		x, y := item.Pos(ctx)
		if x >= rminX && x <= rmaxX && y >= rminY && y <= rmaxY {
			if !iterator(item, ctx) {
				return false
			}
		}
	}
	var used [4]bool
	points := [8]float64{rminX, rminY, rmaxX, rminY, rmaxX, rmaxY, rminX, rmaxY}
	for i := 0; i < 8; i += 2 {
		quad, qminX, qminY, qmaxX, qmaxY := qtreeSplitQuad(points[i+0], points[i+1], minX, minY, maxX, maxY)
		if !used[quad] {
			used[quad] = true
			if n.quads[quad] != nil {
				if !n.quads[quad].search(ctx,
					rminX, rminY, rmaxX, rmaxY,
					qminX, qminY, qmaxX, qmaxY,
					iterator) {
					return false
				}
			}
		}
	}
	return true
}
