package main

import (
	"container/list"
	"fmt"
	"runtime"
	"time"
)

const (
	PosInfinity =  1 << 24
	NegInfinity = -1 << 24
	UNSET = 1 << 30
)

var (
	Depth int
	TreeNodes int
	Threads int
	Processors int
	WallTime time.Duration
)

type StateValuePair struct {
	state *State
	value int
}

func SVPList(l *list.List) *list.List {
	svp := list.New()

	for e := l.Front(); e != nil; e = e.Next() {
		p := new(StateValuePair)
		p.state = e.Value.(*State)
		p.value = UNSET
		svp.PushBack(p)
	}

	return svp
}

func Lookup(children *list.List) int {
	count := 0

	for e := children.Front(); e != nil; e = e.Next() {
		child := e.Value.(*StateValuePair)
		if tableValue, exists := TTable[child.state.Hash()]; exists {
			child.value = tableValue
			count++
		}
	}

	return count
}

func (s *State) NegamaxST2(depth int) *State {
	start := time.Now()

	children, bestValue := s.TrueLegalSuccessors(), NegInfinity
	var choice *State

	Depth, TreeNodes, Threads, Processors = depth, 0, 1, runtime.NumCPU()

	for e := children.Front(); e != nil; e = e.Next() {
		child, value := e.Value.(*State), 0
		grandchildren := child.TrueLegalSuccessors()

		// If checkmate is possible, don't beat around the bush...
		if grandchildren.Len() == 0 && child.InCheck() {
			WallTime = time.Since(start)
			return child
		}

		// Avoid repeated states...
		if Contains(PastStates, child.Hash()) {
			PrintLog("Devaluing (strongly) repeat state from " +
				 MoveString(s, child, Coordinate) + ".\n")
			value = NegInfinity

		// Avoid stalemates...
		} else if grandchildren.Len() == 0 {
			PrintLog("Devaluing (strongly) stalemate from " +
				 MoveString(s, child, Coordinate) + ".\n")
			value = NegInfinity
		} else {
			value = -child.Negamax(depth - 1, NegInfinity, PosInfinity)
		}

		// Avoid repeated moves...
		if Contains(PastMoves, MoveString(s, child, Coordinate)) {
			PrintLog("Devaluing (weakly) repeat move " +
				 MoveString(s, child, Coordinate) + ".\n")
			value -= MaterialValue(Pawn >> 1)
		}

		if value >= bestValue {
			bestValue = value
			choice = child
		}
	}

	PastMoves.PushFront(MoveString(s, choice, Coordinate))
	PastStates.PushFront(choice.Hash())

	WallTime = time.Since(start)
	return choice
}

func (s *State) NegamaxMT(depth int) *State {
	start := time.Now()

	children, bestValue := s.LegalSuccessors(), NegInfinity
	var choice *State
	resultChannel := make(chan *StateValuePair)

	Depth, TreeNodes, Threads, Processors = depth, 0, 0, runtime.NumCPU()
	runtime.GOMAXPROCS(Processors)

	for e := children.Front(); e != nil; e = e.Next() {
		child := e.Value.(*State)
		go child.NegamaxThread(depth - 1, resultChannel)
		Threads++
	}

	for i := 0; i < children.Len(); i++ {
		rp := <-resultChannel

		if rp.value >= bestValue {
			bestValue = rp.value
			choice = rp.state
		}
	}

	WallTime = time.Since(start)
	return choice
}

func (s *State) NegamaxThread(depth int, ch chan *StateValuePair) {
	children, bestValue := s.LegalSuccessors(), NegInfinity
	TreeNodes++

	for e := children.Front(); e != nil; e = e.Next() {
		child := e.Value.(*State)
		value := -child.Negamax(depth - 1, NegInfinity, -bestValue)
		if value >= bestValue {
			bestValue = value
		}
	}

	rp := new(StateValuePair)
	rp.state = s
	rp.value = -bestValue
	ch <- rp
}

func (s *State) NegamaxST(depth int) *State {
	start := time.Now()

	children, bestValue := s.LegalSuccessors(), NegInfinity
	var choice *State

	Depth, TreeNodes, Threads, Processors = depth, 0, 1, runtime.NumCPU()

	for e := children.Front(); e != nil; e = e.Next() {
		child := e.Value.(*State)
		value := -child.Negamax(depth - 1, NegInfinity, PosInfinity)
		if value >= bestValue {
			bestValue = value
			choice = child
		}
	}

	WallTime = time.Since(start)
	return choice
}

func (s *State) Negamax(depth, alpha, beta int) int {
	TreeNodes++
	if depth == 0 || s.LostKing() {
		return s.Value()
	}

	children := s.LegalSuccessors()

	for e := children.Front(); e != nil; e = e.Next() {
		child := e.Value.(*State)
		value := -child.Negamax(depth - 1, -beta, -alpha)
		if value >= beta {
			return value
		}
		if value >= alpha {
			alpha = value
		}
	}

	return alpha
}

func (s *State) Value() int {
	value := 0

	value += s.MaterialAdvantage()
	value += s.PositionAppeal()

	return value
}

func (s *State) PositionAppeal() int {
	value, player := 0, s.ToMove

	// we like to control the center of the board...
	for i := 3; i < 5; i++ {
		for j := 2; j < 6; j++ {
			color := s.GetColor(i, j)
			if color == player {
				piece := s.GetPiece(i, j)
				value += (MaterialValue(piece) >> 6)
			} else if color == Opponent(player) {
				piece := s.GetPiece(i, j)
				value -= (MaterialValue(piece) >> 6)
			}
		}
	}

	return value
}

func (s *State) MaterialAdvantage() int {
	value, bishops, enemyBishops := 0, 0, 0
	king, enemyKing := false, false
	player := s.ToMove

	for i := 0; i < 8; i++ {
		for j := 0; j < 8; j++ {
			color := s.GetColor(i, j)
			if color == player {
				piece := s.GetPiece(i, j)
				value += MaterialValue(piece)
				if piece == Bishop {
					bishops++
				} else if piece == King {
					king = true
				}
			} else if color == Opponent(player) {
				piece := s.GetPiece(i, j)
				value -= MaterialValue(piece)
				if piece == Bishop {
					enemyBishops++
				} else if piece == King {
					enemyKing = true
				}
			}
		}
	}

	if !king {
		return NegInfinity
	}
	if !enemyKing {
		return PosInfinity
	}

	if bishops > 1 {
		value += 040
	}
	if enemyBishops > 1 {
		value -= 040
	}

	return value
}

func MaterialValue(piece Piece) int {
	switch (piece) {
	case Pawn:
		return 0100
	case Knight:
		return 0300
	case Bishop:
		return 0300
	case Rook:
		return 0500
	case Queen:
		return 01100
	}

	return 0
}
