package main

import (
	"fmt"
	"github.com/ajstarks/svgo"
	"io"
)

type SVGRenderer struct {
	dest  io.Writer
	scale int // size of each location

}

func (c *Coord) Ints() (int, int) {
	return c.X, c.Y
}

//bordered position
func (sr *SVGRenderer) PosInts(x, y int) (int, int) {
	return (x + 1) * sr.scale, (y + 1) * sr.scale
}

func (sr *SVGRenderer) Draw(m *Maze) {
	canvas := svg.New(sr.dest)
	canvas.Start((m.x+2)*sr.scale, (m.y+2)*sr.scale)
	// border
	canvas.Rect(sr.scale/2, sr.scale/2,
		(m.x+1)*sr.scale, (m.y+1)*sr.scale,
		fmt.Sprintf("stroke: black; stroke-width: %d; fill: black", sr.scale),
	)
	i, _ := m.Iter()
	for loc := range i {
		on, _ := m.grid.Neighbors(loc.Coord)
		var x, y = sr.PosInts(loc.X, loc.Y)
		var msg, textstyle string
		textstyle =fmt.Sprintf("font-size: %d;", sr.scale/2-1)
		//d := &Dims{x, y}
		if loc.Passable {
			/*canvas.Circle(
				x+(sr.scale/2),y+(sr.scale/2),
				sr.scale/2,
				fmt.Sprintf("stroke-width: %d; stroke: white; stroke-linecap: round" +
					"fill: white", sr.scale),
			)
			*/
			for _, n := range on {
				if m.grid.At(n).Passable && n.X >= loc.X && n.Y >= loc.Y {
					otherx, othery := sr.PosInts(n.X, n.Y)
					canvas.Line(
						x+(sr.scale/2),y+(sr.scale/2),
						otherx+(sr.scale/2), othery+(sr.scale/2),
						fmt.Sprintf("stroke-width: %d; fill: white; stroke: white; stroke-linecap: round", sr.scale * 4 / 5 ),
					)
				}
			}
			if loc.Special&Start != 0 {
				msg = "S"
				textstyle += "fill: green;"
			} else if loc.Special&Finish != 0 {
				msg = "F"
				textstyle += "fill: light-blue;"
			} else if loc.Special&MaxPasses != 0 {
				msg = "E"
			} else if loc.Special&Reverse != 0 {
				msg = "r"
			} else if loc.Special & CreateEnd != 0 {
				textstyle += "fill: pink;"
				msg = "e"
			} else {
				//msg += d.String()
				textstyle += "visibility: hidden"
			}
		} else {
			//msg += d.String()
			textstyle = "visibility: hidden"
		}
		if msg != "" {
			canvas.Text(x+sr.scale/2, y+sr.scale/2, msg, textstyle+ "; dominant-baseline:middle; text-anchor:middle")
		}
	}
	canvas.End()
}
