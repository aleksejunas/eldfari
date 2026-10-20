package main

import (
	"container/heap"
	"math"

	"raketti/sim"
)

const (
	cell = 4.0 // grid resolution in px

	// shipRadius is the clearance the ship's centre needs to fit: the ship is
	// 21 px tall (nose 12 px above the centre) and 16 px wide.
	shipRadius = 12.0

	// settle is how far a route point may move to the most open spot nearby.
	// Without it the start point, just above the start pad, would always be
	// the narrowest place on the route.
	settle = 60.0

	// maxPassage is the widest clearance that counts as a passage. Wider
	// places are open caves and are never the hard part.
	maxPassage = 300.0
)

// grid holds, for every cell, the distance from its centre to the nearest
// solid cell (terrain or outside the world), and the width of the passage the
// cell is in.
type grid struct {
	w, h    int
	clear   []float64
	passage []float64 // clear, if there are walls on two opposite sides; else +Inf
}

func newGrid(l *sim.Level) *grid {
	g := &grid{w: int(math.Ceil(float64(l.W) / cell)), h: int(math.Ceil(float64(l.H) / cell))}
	g.clear = make([]float64, g.w*g.h)
	// Distance transform: Dijkstra from all solid cells at once. Steps are
	// straight or diagonal, so distances are within a few percent of the truth.
	q := &pq{}
	for i := range g.clear {
		x, y := g.centre(i)
		if l.Solid(x, y) {
			g.clear[i] = 0
			heap.Push(q, item{i, 0})
		} else {
			g.clear[i] = math.Inf(1)
		}
	}
	for q.Len() > 0 {
		it := heap.Pop(q).(item)
		if it.v > g.clear[it.i] {
			continue
		}
		g.neighbours(it.i, func(n int, step float64) {
			if d := it.v + step; d < g.clear[n] {
				g.clear[n] = d
				heap.Push(q, item{n, d})
			}
		})
	}
	g.passage = make([]float64, len(g.clear))
	for i, c := range g.clear {
		g.passage[i] = math.Inf(1)
		if c > 0 && c <= maxPassage && !math.IsInf(g.localWidth(i), 1) {
			g.passage[i] = c
		}
	}
	return g
}

func (g *grid) centre(i int) (float64, float64) {
	return (float64(i%g.w) + 0.5) * cell, (float64(i/g.w) + 0.5) * cell
}

func (g *grid) index(p sim.Pt) int {
	x := min(max(int(p.X/cell), 0), g.w-1)
	y := min(max(int(p.Y/cell), 0), g.h-1)
	return y*g.w + x
}

// openest returns the cell with the most clearance within settle px of p.
func (g *grid) openest(p sim.Pt) int {
	best := g.index(p)
	r := int(settle / cell)
	cx, cy := best%g.w, best/g.w
	for y := max(cy-r, 0); y <= min(cy+r, g.h-1); y++ {
		for x := max(cx-r, 0); x <= min(cx+r, g.w-1); x++ {
			dx, dy := x-cx, y-cy
			if i := y*g.w + x; dx*dx+dy*dy <= r*r && g.clear[i] > g.clear[best] {
				best = i
			}
		}
	}
	return best
}

// neighbours calls f for the 8 neighbours of cell i with the step length in px.
func (g *grid) neighbours(i int, f func(n int, step float64)) {
	x, y := i%g.w, i/g.w
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			nx, ny := x+dx, y+dy
			if (dx == 0 && dy == 0) || nx < 0 || ny < 0 || nx >= g.w || ny >= g.h {
				continue
			}
			step := cell
			if dx != 0 && dy != 0 {
				step = cell * math.Sqrt2
			}
			f(ny*g.w+nx, step)
		}
	}
}

// distances returns the shortest path length in px from cell src to every
// cell, moving only through cells where the ship fits.
func (g *grid) distances(src int) []float64 {
	dist := make([]float64, len(g.clear))
	for i := range dist {
		dist[i] = math.Inf(1)
	}
	dist[src] = 0
	q := &pq{{src, 0}}
	for q.Len() > 0 {
		it := heap.Pop(q).(item)
		if it.v > dist[it.i] {
			continue
		}
		g.neighbours(it.i, func(n int, step float64) {
			if g.clear[n] < shipRadius {
				return
			}
			if d := it.v + step; d < dist[n] {
				dist[n] = d
				heap.Push(q, item{n, d})
			}
		})
	}
	return dist
}

// bottleneck returns the widest way from cell a to cell b: the largest
// clearance c such that a path exists where every cell has at least c, and the
// cell where that narrowest point is. Only passages count (see localWidth), and
// not near a or b: a pad or a crystal in a corner is not a narrow passage.
func (g *grid) bottleneck(a, b int) (float64, int) {
	width := func(i int) float64 {
		if g.near(i, a) || g.near(i, b) {
			return math.Inf(1)
		}
		return g.passage[i]
	}
	best := make([]float64, len(g.clear)) // widest bottleneck found to each cell
	from := make([]int, len(g.clear))
	for i := range best {
		best[i] = -1
		from[i] = -1
	}
	best[a] = width(a)
	q := &pq{{a, -best[a]}} // pq pops the smallest value, so store -width
	for q.Len() > 0 {
		it := heap.Pop(q).(item)
		if -it.v < best[it.i] {
			continue
		}
		if it.i == b {
			break
		}
		g.neighbours(it.i, func(n int, _ float64) {
			if g.clear[n] < shipRadius {
				return
			}
			if w := math.Min(best[it.i], width(n)); w > best[n] {
				best[n], from[n] = w, it.i
				heap.Push(q, item{n, -w})
			}
		})
	}
	// Walk back from b to find the cell that limits the width.
	// Many cells on the path can share the narrowest clearance (along the
	// ceiling of a low room, for example). Pick the one between the closest
	// pair of walls: the real passage.
	where, wall := b, math.Inf(1)
	for i := b; i != -1; i = from[i] {
		if width(i) <= best[b]+cell {
			if lw := g.localWidth(i); lw < wall {
				where, wall = i, lw
			}
		}
	}
	return best[b], where
}

// near reports whether cells i and j are within settle px of each other.
func (g *grid) near(i, j int) bool {
	xi, yi := g.centre(i)
	xj, yj := g.centre(j)
	return math.Hypot(xi-xj, yi-yj) <= settle
}

// localWidth is the distance between the two walls on opposite sides of cell
// i (up and down, left and right, or along a diagonal), using the closest
// pair. It is +Inf if no pair is within 2 × maxPassage: a cell near the floor
// of a big open cave has a wall on one side only.
func (g *grid) localWidth(i int) float64 {
	reach := 2 * maxPassage
	dist := func(dx, dy int) float64 {
		step := cell * math.Hypot(float64(dx), float64(dy))
		x, y := i%g.w, i/g.w
		for d := step; d <= reach; d += step {
			x, y = x+dx, y+dy
			if x < 0 || y < 0 || x >= g.w || y >= g.h || g.clear[y*g.w+x] == 0 {
				return d
			}
		}
		return math.Inf(1)
	}
	w := math.Inf(1)
	for _, d := range [][2]int{{0, 1}, {1, 0}, {1, 1}, {1, -1}} {
		w = math.Min(w, dist(d[0], d[1])+dist(-d[0], -d[1]))
	}
	return w
}

type stats struct {
	gap        float64 // narrowest gap the route must pass, px (2 × clearance)
	gapAt      sim.Pt  // where that gap is
	route      float64 // route length in px: start, every crystal, back to start
	longestLeg float64 // longest single leg of the route, px
	path       []int   // cells of the route, for drawing
	ok         bool    // false if some crystal can't be reached
}

// analyse visits the crystals in nearest-first order (by path length) from the
// start pad and back. It is a reasonable route, not always the shortest one.
func analyse(l *sim.Level) (*grid, stats) {
	g := newGrid(l)
	points := append([]sim.Pt{l.Start}, l.Crystals...)

	cells := make([]int, len(points))
	dists := make([][]float64, len(points))
	for i, p := range points {
		cells[i] = g.openest(p)
		dists[i] = g.distances(cells[i])
	}

	st := stats{gap: math.Inf(1), ok: true}
	visited := make([]bool, len(points))
	visited[0] = true
	cur := 0
	for range l.Crystals {
		next := -1
		for j := range points {
			if !visited[j] && (next == -1 || dists[cur][cells[j]] < dists[cur][cells[next]]) {
				next = j
			}
		}
		st.leg(g, dists, cells, cur, next)
		visited[next], cur = true, next
	}
	st.leg(g, dists, cells, cur, 0)
	return g, st
}

// leg adds the way from point a to point b to the stats.
func (st *stats) leg(g *grid, dists [][]float64, cells []int, a, b int) {
	d := dists[a][cells[b]]
	if math.IsInf(d, 1) {
		st.ok = false
		return
	}
	st.route += d
	st.longestLeg = math.Max(st.longestLeg, d)
	if w, at := g.bottleneck(cells[a], cells[b]); 2*w < st.gap {
		x, y := g.centre(at)
		st.gap, st.gapAt = 2*w, sim.Pt{X: x, Y: y}
	}
	st.path = append(st.path, g.walk(dists[a], cells[b])...)
}

// walk follows decreasing distance from cell to the source of dist.
func (g *grid) walk(dist []float64, cell int) []int {
	path := []int{cell}
	for dist[cell] > 0 {
		next := cell
		g.neighbours(cell, func(n int, _ float64) {
			if dist[n] < dist[next] {
				next = n
			}
		})
		if next == cell {
			break
		}
		cell = next
		path = append(path, cell)
	}
	return path
}

type item struct {
	i int
	v float64
}

// pq is a min-heap of items by value, for container/heap.
type pq []item

func (q pq) Len() int           { return len(q) }
func (q pq) Less(i, j int) bool { return q[i].v < q[j].v }
func (q pq) Swap(i, j int)      { q[i], q[j] = q[j], q[i] }
func (q *pq) Push(x any)        { *q = append(*q, x.(item)) }
func (q *pq) Pop() any {
	old := *q
	it := old[len(old)-1]
	*q = old[:len(old)-1]
	return it
}
