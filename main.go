package main

import (
	"fmt"
	"io"
	"strings"
	"sync"
	//"log"
	"os"
)

func main() {
	m := NewMaze(80, 40)
	wc := &WalkingCreator{}
	wc.Fill(&m.grid, Coord{0, 0}, Coord{m.x - 1, m.y - 1})
	d := ConsoleRenderer{
		dest: os.Stdout,
	}
	d.Draw(m)
	fmt.Print("")
}

type Coord struct {
	X, Y int
}

func (c *Coord) String() string {
	if c == nil {
		return "(undefined,undefined)"
	}
	return fmt.Sprintf("(%d,%d)", c.X, c.Y)
}

type Dims Coord

func (d *Dims) String() string {
	if d == nil {
		return "undefined"
	}
	return fmt.Sprintf("%dx%d", d.X, d.Y)
}

func (d Dims) CoordOf(i int) Coord {
	// return the coord of offset i
	return Coord{i % d.X, i / d.X}
}

const (
	Start = 1 << iota
	Finish
	Reverse
	MaxPasses
)

type Loc struct {
	Coord
	Passable bool
	Special  uint
}

func MakePassable(l Loc) Loc {
	l.Passable = true
	return l
}

type OutOfBoundsError struct {
	loc  Coord
	dims Dims
	l    int
}

func (oobe *OutOfBoundsError) Error() string {
	return fmt.Sprintf("%s is Out of bounds of %dx%d (%d) Grid",
		&oobe.loc, oobe.dims.X, oobe.dims.Y, oobe.l)
}

func (g *Grid) oobe(loc Coord) *OutOfBoundsError {
	return &OutOfBoundsError{
		loc:  loc,
		dims: g.dims,
		l:    len(g.g),
	}
}

type Grid struct {
	g    []Loc
	dims Dims
}

func (g *Grid) Len() int {
	return len(g.g)
}

func (g *Grid) Update(f func(Loc) Loc, cs ...Coord) {
	for _, c := range cs {
		g.oobPanic(c)
		i := g.Idx(c)
		g.g[i] = f(g.g[i])
	}
}

func (g *Grid) Within(c Coord) (b bool) {
	/*defer func() {
		log.Printf("%s Grid: Coord %s Within? %t",
			g.dims.String(),
			c.String(),
			b)
	}()*/
	if c.X < 0 || c.Y < 0 || c.Y >= g.dims.Y || c.X >= g.dims.X {
		return false
	}
	return g.WithinIdx(g.Idx(c))
}

func (g *Grid) WithinIdx(i int) (b bool) {
	/*
		defer func() {
			log.Printf("%s Grid: Index %d Within? %t",
				g.dims.String(),
				i,
				b)
			}()*/
	return i >= 0 && i < len(g.g)
}

func (g *Grid) oobPanic(c Coord) {
	g.oobPanicIdx(g.Idx(c))
}

func (g *Grid) oobPanicIdx(i int) {
	if !g.WithinIdx(i) {
		panic(g.oobe(g.dims.CoordOf(i)))
	}
}

func (g *Grid) CoordOf(idx int) Coord {
	return g.dims.CoordOf(idx)
}

func (g *Grid) AtIdx(idx int) Loc {
	g.oobPanicIdx(idx)
	return g.g[idx]
}

func (g *Grid) At(c Coord) Loc {
	g.oobPanic(c)
	return g.g[g.Idx(c)]
}

func (g *Grid) Init(dims Dims) *Grid {
	g.dims = dims
	g.g = make([]Loc, dims.X*dims.Y)
	for j := 0; j < dims.Y; j++ {
		for i := 0; i < dims.X; i++ {
			g.g[g.Idx(Coord{i, j})].X = i
			g.g[g.Idx(Coord{i, j})].Y = j
		}
	}
	return g
}

func (g *Grid) Idx(loc Coord) int {
	return loc.Y*g.dims.X + loc.X
}

type coordCandidates struct {
	cand []Coord
	dest []Coord
}

func (cc *coordCandidates) filter(f func(Coord) bool) {
	cc.dest = make([]Coord, 0)
	for _, c := range cc.cand {
		if f(c) {
			cc.dest = append(cc.dest, c)
		}
	}
}

// return orthogonals and diagonals separately
func (g *Grid) Neighbors(loc Coord) ([]Coord, []Coord) {
	ccs := []*coordCandidates{
		{
			cand: []Coord{ // orthogonals
				Coord{loc.X - 1, loc.Y},
				Coord{loc.X, loc.Y - 1},
				Coord{loc.X + 1, loc.Y},
				Coord{loc.X, loc.Y + 1},
			},
		},
		{
			cand: []Coord{ // diagonals
				Coord{loc.X - 1, loc.Y - 1},
				Coord{loc.X - 1, loc.Y + 1},
				Coord{loc.X + 1, loc.Y + 1},
				Coord{loc.X + 1, loc.Y - 1},
			},
		},
	}
	for _, cc := range ccs {
		cc.filter(g.Within)
	}
	/*
		log.Printf("For %s Neighbors: %+v, %+v",
			&loc, ccs[0].dest, ccs[1].dest)
	*/
	return ccs[0].dest, ccs[1].dest
}

type Maze struct {
	grid Grid
	l    sync.RWMutex
	x, y int
}

func (m *Maze) At(x, y int) Loc {
	m.l.RLock()
	l := m.grid.g[m.grid.Idx(Coord{x, y})]
	m.l.RUnlock()
	return l
}

func (m *Maze) Iter() (r chan Loc, cancel chan struct{}) {
	r = make(chan Loc)
	cancel = make(chan struct{}, 1)
	go func() {
		for y := 0; y < m.y; y++ {
			for x := 0; x < m.x; x++ {
				l := m.At(x, y)
				select {
				case r <- l:
					continue
				case <-cancel:
					close(r)
					return
				}
			}
		}
		close(r)
	}()
	return
}

func (m *Maze) idx(x, y int) int {
	return m.grid.Idx(Coord{x, y})
}

func NewMaze(x, y int) (m *Maze) {
	m = &Maze{
		x: x,
		y: y,
	}
	m.grid.Init(Dims{x, y})
	return
}

type Renderer interface {
	Draw(*Maze) error
}

type ConsoleRenderer struct {
	dest io.Writer
}

func (cr *ConsoleRenderer) Draw(m *Maze) {
	var bordercolor = "\033[1;48;5;94;38;5;94m"
	var opencolor = "\033[0;m"
	var clear = "\033[0m"
	var wall = "\u2588"
	// header
	fmt.Fprintf(cr.dest, "%s\n", m.grid.dims.String())
	fmt.Fprint(cr.dest, bordercolor, strings.Repeat(wall, m.x+2), bordercolor, clear, "\n", bordercolor, wall)
	// iterate over the elements, adding prefix and suffix to each lines wiht `oldx` rolls over
	i, _ := m.Iter()
	var oldx int = 0
	for loc := range i {
		if oldx > loc.X {
			// newline and borders
			fmt.Fprint(cr.dest, bordercolor, wall, clear, "\n", bordercolor, wall)
		}
		oldx = loc.X
		//var s = fmt.Sprintf("\033[1;38;5;135m%s\033[0m", "\u2588")
		var s = fmt.Sprint(bordercolor, wall)
		if loc.Passable {
			s = fmt.Sprint(clear, opencolor, func() string {
				if loc.Special&Start != 0 {
					return "\033[32m" + "S"
				}
				if loc.Special&Finish != 0 {
					return "\033[32m" + "F"
				}
				if loc.Special&MaxPasses != 0 {
					return "\033[38;5;219m" + " "
				}
				if loc.Special&Reverse != 0 {
					return "\033[38;5;212m" + " "
				}
				return " "
			}(), clear)
		}
		fmt.Fprint(cr.dest, s)
	}
	//footer
	fmt.Fprint(cr.dest, bordercolor, wall, clear, "\n", bordercolor, strings.Repeat(" ", m.x+2), clear, "\n")
}
