package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"
)

var logging bool

type Trans Coord // translation along x and y

func (t Trans) Translate(c Coord) Coord {
	return Coord{c.X + t.X, c.Y + t.Y}
}

func (t *Trans) String() string {
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
	seed int64
}

type walkingCreatorRun struct {
	g                 *Grid
	finish            Coord
	start             Coord
	reached           bool
	reverse           bool    // after some passes, we complete the maze by tracing back from the end
	reverse_locations []Coord // when reverse solving, we can't backtrack to anywhere -we need to choose one of these
}

func (r *walkingCreatorRun) previousCandidateFilterFactory(
	cur Coord,
) func(Coord) bool { //Is this neighbor a possible previous?  We're tracing backwards to solve the maze
	return func(c Coord) (b bool) {
		var msg string
		if logging {
			defer func() {
			log.Printf("Cur %s: Filter result for %s: %t (%s)",
				cur.String(),
				c.String(), b, msg)
			}()
		}
		if r.g.At(c).Passable {
			// if we're working backwards and we've found a passable  candidate, we've
			// complated the maze.
			for _, revc := range r.reverse_locations {
				if revc == c {
					msg = "This is part of the reverse solution we're working on so it goes in a circle"
					return false
				}
			}
			msg = fmt.Sprintf(
				"We're reverse completing and its' passible and not in reverse locations, so our maze is solved at %s",
				&c,
			)
			r.reached = true
			return false
		}
		//If this square has neighbors that are Passable and not in list of
		//reverse locations, it is a solution candidate
		oneigh, _ := r.g.Neighbors(c)
		for _, cn := range oneigh { //next candidates orth. neighs
			if cn == cur {
				continue
				//we  already know it's orthogonal to the current point!
			}
			if cn == r.finish && cur != r.finish {
				msg = fmt.Sprintf(
					"We're wokring backwards and %s brings us back to the finish (%s)", &c, &cn)
				return false
			}
			for _, pc := range r.reverse_locations {
				if pc == cn {
					msg = fmt.Sprintf("%s neighbors reverse location %s so it brings us back to the finish", &c, &pc)
					return false
				}
			}
			if r.g.At(cn).Passable {
				// we already filtered out the passable reverse locations
				//TODO: This is a preferred reverse solution
				msg = fmt.Sprintf("%s has a neighbor %s that would complete the reverse solution we're working on",
					&c, &cn)
				return true
			}
		}
		msg = "no disqualifying factors"
		return true
	}
}
func (r *walkingCreatorRun) nextCandidateFilterFactory(
	cur Coord,
) func(Coord) bool {
	return func(c Coord) (b bool) { //Is this neighbor a possible next?
		var msg string
		if logging { defer func() {
			log.Printf("Cur %s: Filter result for %s: %t (%s)",
				cur.String(),
				c.String(), b, msg)
		}() }
		if c == r.finish {
			msg = "it's the finish (reached set to true)"
			r.reached = true
			return false
		} else if r.g.At(c).Passable { //invalidates the start
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
			if cn == r.finish {
				msg = fmt.Sprintf("%s brings us to the finish(%s)", &c, &cn)
				return true
			}
			if r.g.At(cn).Passable {
				msg = fmt.Sprintf(
					"one of %s's orthogonal neighbors, %s, is aready passable so this would loop",
					c.String(), cn.String())
				return false
			}
		}
		msg = "no disqualifying factors"
		return true
	}
}

func (wc *WalkingCreator) Fill(grid *Grid, start, finish Coord) {
	var max_passes = grid.Len() * 12
	if wc.seed == 0 {
		wc.seed = time.Now().UnixNano()
	}
	rand.Seed(wc.seed)
	grid.Update(MakePassable, start, finish)
	grid.Update(func(l Loc) Loc {
		l.Special = l.Special | Start
		return l
	}, start)
	grid.Update(func(l Loc) Loc {
		l.Special = l.Special | Finish
		return l
	}, finish)
	r := walkingCreatorRun{
		start:             start,
		finish:            finish,
		g:                 grid,
		reverse_locations: make([]Coord, 0),
	}
	cur := start
	for !r.reached {
		if max_passes < -10000 {
			log.Printf("Failed to reverse complete maze after 10000 reverse iterations")
			// this should probably be a failure of somethign
			return
		} else if max_passes == 0 {
			if logging { log.Printf("Max passes reached at %s, reverse completings", &cur) }
			grid.Update(func(l Loc) Loc { l.Special = l.Special | MaxPasses; return l }, cur)
			r.reverse = true
			cur = finish
			r.reverse_locations = append(r.reverse_locations, finish)
		}
		max_passes--
		var nexts coordCandidates
		nexts.cand, _ = grid.Neighbors(cur)
		if r.reverse {
			nexts.filter(r.previousCandidateFilterFactory(cur))
		} else {
			nexts.filter(r.nextCandidateFilterFactory(cur))
		}
		if len(nexts.dest) > 0 {
			// TODO: if we're reverse solving, we should prefer candidates that complete the maze
			// we can choose a random qualified candidate
			next := nexts.dest[rand.Intn(len(nexts.dest))]
			grid.Update(MakePassable, next)
			if r.reverse {
				grid.Update(func(l Loc) Loc { l.Special = l.Special | Reverse; return l }, next)
			}
			if r.reverse {
				r.reverse_locations = append(r.reverse_locations, cur)
			}
			cur = next
			continue
		}
		// we don't seem to have any possible positions
		if r.reverse {
			if len(r.reverse_locations) == 0 {
				panic("No reverse destinations!  This makes no sense")
			}
			last := cur
			cur = r.reverse_locations[rand.Intn(len(r.reverse_locations))]
			if logging {
				log.Printf("No candidates  backwards from %s, going to previous part of backwards path %s", last.String(), cur.String())
			}
			continue
		}
		if logging { log.Printf("No candidates onward from %s", cur.String()) }
		grid.Update(func(l Loc)Loc{l.Special = l.Special | CreateEnd; return l }, cur)
		// so we'll "backtrack"
		i := r.g.Idx(cur)
		for !(i != r.g.Idx(cur) && i != r.g.Idx(start) &&
			i != r.g.Idx(finish) && r.g.AtIdx(i).Passable) {
			i = rand.Intn(grid.Len())
		}
		cur = grid.CoordOf(i)
		if logging { log.Printf("Backtracking to %s: %+v", &cur, grid.At(cur)) }
	}
}
