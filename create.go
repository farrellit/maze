package main

import (
	"math/rand"
	"time"
	"log"
	"fmt"
)

type Trans Coord // translation along x and y

func (t Trans) Translate(c Coord) Coord {
	return Coord{c.X + t.X, c.Y + t.Y}
}

func (t *Trans)String() string {
	c := Coord(*t)
	return c.String()
}

func (t *Trans) Rand() {
	switch rand.Intn(4) {
	case 0:
		*t = Trans{-1, 0}
	case 1:
		*t = Trans{0, -1}
	case 2:
		*t = Trans{1, 0}
	case 3:
		*t = Trans{0, 1}
	}
}

type MazeCreator interface {
	Fill(grid *Grid, start, finish Coord)
}

type WalkingCreator struct {
}

type walkingCreatorRun struct {
	g *Grid
	finish Coord
	start Coord
	reached bool
}

func (r *walkingCreatorRun)nextCandidateFilterFactory(
	cur Coord,
)( func(Coord)bool){
	return func(c Coord)(b bool){ //Is this neighbor a possible next?
			var msg string
			defer func(){
				log.Printf("Filter result for %s: %t (%s)",
				c.String(), b, msg)
			}()
			// if we've reached the finish, 
			// we can stop iterating.  This isn't a valid next,
			// because we're already "next to" it.
			if c == r.finish {
				msg = "it's the finish (reached set to true)"
				r.reached = true
				return false
			}
			//for this to be a candidate is has to be  impassible now
			if r.g.At(c).Passable {
				msg = "already passable"
				return false
			}
			//if this square already has orthogonal neighbors that 
			//are passable (and not us), we'll reject it as well
			oneigh, _ := r.g.Neighbors(c)
			for _, cn := range oneigh { //next candidates orth. neighs
				if cn == cur {
					continue
					//we  already know it's orthogonal to the current point
				}
				if r.g.At(cn).Passable {
					msg = fmt.Sprintf(
						"one of %s's orthogonal neighbors, %s, is aready passable",
						c.String(), cn.String())
					return false
				}
			}
			msg = "no disqualifying factors"
			return true
		}
}

func (wc *WalkingCreator) Fill(grid *Grid, start, finish Coord) {
	var max_passes = 1000
	rand.Seed(time.Now().UnixNano())
	grid.Update(MakePassable, start, finish)
	r := walkingCreatorRun{
		start: start,
		finish: finish,
		g: grid,
	}
	cur := start
	for ! r.reached {
		if max_passes == 0 {
			log.Printf("Max passes reached at %s", &cur)
			return
		} else {
			max_passes--
		}
		var nexts coordCandidates
		nexts.cand, _ = grid.Neighbors(cur)
		nexts.filter(r.nextCandidateFilterFactory(cur))
		if len(nexts.dest) > 0 {
			// we can choose a random qualified candidate
			next := nexts.dest[rand.Intn(len(nexts.dest))]
			grid.Update(MakePassable,next)
			cur = next
		} else if ! r.reached  {
			// we don't seem to have any possible positions, 
			log.Printf("No candidates onward from %s", cur.String())
			// so we'll "backtrack"
			i := r.g.Idx(cur)
			for !(
				i != r.g.Idx(cur) && i != r.g.Idx(start) &&
				i != r.g.Idx(finish) && r.g.AtIdx(i).Passable) {
				i = rand.Intn(grid.Len())
			}
			cur = grid.CoordOf(i)
			log.Printf("Backtracking to %s: %+v", &cur, grid.At(cur))
		}
	}
}

/*
func (wc *WalkingCreator) FillOld( grid Grid, start, finish Coord) {
	rand.Seed(time.Now().UnixNano())
	grid.Update(func(l Loc)Loc{
		l.Passable = true
		return l
	}, start)
	var trans Trans
	var pos = start
	var reached = false
	for iters := 0;!reached && iters < 100; iters++ {
		trans.Rand()
		mag := rand.Intn(8)
		next := pos
		log.Printf("Next translation is %s x %d",
			&trans, mag)
		for i := 0; i < mag; i++ {
			next = trans.Translate(next)
			log.Printf("Next location is %s", &next)
			if ! grid.Within(next) {
				break
			}
			nextloc := grid.At(next)
			if nextloc.Passable {
				break
			}
			os, ds := grid.Neighbors(next)
			var passables = 0
			checks := append(os, ds...)
			log.Printf("checking locations %+v", checks)
			for _, p := range checks {
				if p == pos {
					// one passable square we can touch -
					//   our current path position and the finish!
					continue
				} else if p == finish {
					log.Printf("%s can reach the finish!", &next)
					reached = true
				}
				loc := grid.At(p)
				if loc.Passable {
					passables++
				}
			}
			if passables > 0 {
				log.Printf("%s touches too many passables (%d).  finding another candidate", &next, passables)
				break
			}
			grid.Update(MakePassable, next)
			pos = next
		}
	}
}
*/
