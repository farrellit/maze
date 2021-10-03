package main

import (
	"math/rand"
	"testing"
	"time"
	"bytes"
	"fmt"
)

type CoordTest struct {
	c Coord
	s string
	cof map[int]Coord
}

func TestCoord(t *testing.T){
	tests := []CoordTest{
		CoordTest{
			Coord{3,3},
			"(3,3)",
			map[int]Coord{
				12: {0,4},
			},
		},
	}
	t.Run("String", func(t *testing.T){
		for _,  test := range tests {
			if s:= test.c.String(); s != test.s {
				t.Errorf("%+v: wanted %s, got %s",test.c,test.s,s)
			}
		}
	})
	t.Run("Dims.CoordOf", func(t *testing.T){
		for _,  test := range tests {
			t.Run(test.c.String(), func(t *testing.T){
				for idx, exp := range test.cof {
					t.Run(
						fmt.Sprintf("%d=>%s", idx, exp.String()),
						func(t *testing.T){
							if c := Dims(test.c).CoordOf(idx); c != exp {
								t.Errorf("Exptected %s, got %s", &exp, &c)
							}
						},
					)
				}
			})
		}
	})
}

type GridTest struct {
	g Grid
	dims Dims
	idxs []IdxTest
}

type IdxTest struct {
	pos Coord
	idx int
}

func TestGrid(t *testing.T){
	gridTests := []GridTest{
		GridTest{
			dims:Dims{10,10},
			idxs : []IdxTest{
				IdxTest{Coord{0,0},0},
				IdxTest{Coord{9,0},9},
				IdxTest{Coord{9,9},99},
			},
		},
	}
	t.Run("Init", func(t *testing.T){
		for _, test := range gridTests{
			t.Run(test.dims.String(), func(t *testing.T){
				test.g.Init(test.dims)
				if len(test.g.g) != test.dims.X * test.dims.Y {
					t.Errorf("Grid did not have expected dimensions: %d vs %d",
					len(test.g.g), test.dims.X * test.dims.Y)
				}
			})
		}
	})
	t.Run("Idx", func(t *testing.T){
		for _, test := range gridTests{
			test.g.Init(test.dims)
			t.Run(test.dims.String(), func(t *testing.T){
				for _, idxt := range test.idxs {
					t.Run(fmt.Sprintf("%+v", idxt),func(t *testing.T){
						if i := test.g.Idx(idxt.pos); i != idxt.idx {
							t.Errorf("Expected %d, got %d", idxt.idx, i)
						}
					})
				}
			})
		}
	})
}


func TestMaze(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	t.Run("NewMaze", func(t *testing.T) {
		m := NewMaze(100, 100)
		t.Run("RandomLoc", func(t *testing.T) {
			x, y := rand.Intn(m.x), rand.Intn(m.y)
			if l := m.At(x, y); l.X != x || l.Y != y {
				t.Errorf("(%d,%d) did not have correct coordinates: %+v",
					x, y, l)
			}
		})
	})
}

func TestConsoleRenderer(t *testing.T) {
	t.Run("Draw", func(t *testing.T) {
		m := NewMaze(100, 100)
		var b bytes.Buffer
		d := ConsoleRenderer{
			dest: &b,
		}
		d.Draw(m)
		t.Log("\n" + string(b.Bytes()))
	})
}

func TestWalkingCreator(t *testing.T){
	m := NewMaze(50,50)
	wc := &WalkingCreator{}
	wc.Fill(&m.grid, Coord{0,0}, Coord{49,49})
	var b bytes.Buffer
	d := ConsoleRenderer{
		dest: &b,
	}
	d.Draw(m)
	t.Log("\n" + string(b.Bytes()))
}
