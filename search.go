// Copyright 2013 Chad Williamson.

// This file is part of Turgenev, a chess program.

// Turgenev is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// Turgenev is distributed in the hope that it will be useful, but
// WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU
// General Public License for more details.

// You should have received a copy of the GNU General Public License along
// with Turgenev. If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"time"
)

const (
	PosInfinity =  1 << 24
	NegInfinity = -1 << 24
	UNSET = 1 << 30
)

// Duration of the last search
var WallTime time.Duration

// SearchFunction is a type common to all searches used for passing such
// functions to GameLoop() (for example) as parameters.
type SearchFunction func(*State, int) *State

// NegamaxST() is a single-threaded negamax search with alpha-beta pruning.
func NegamaxST(s *State, depth int) *State {
	start := time.Now()

	children, bestValue := s.LegalSuccessors(), NegInfinity
	var choice *State

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

// Negamax() is the inner recursive part of the negamax search.
func (s *State) Negamax(depth, alpha, beta int) int {
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

// Value() is the State evaluation function, which returns an integer.
func (s *State) Value() int {
	value := 0

	value += s.MaterialAdvantage()
	value += s.PositionAppeal()

	return value
}

// PositionAppeal() is one element of the evaluation function which returns
// an integer expressing the favorability of material distribution on the
// board.
func (s *State) PositionAppeal() int {
	value, player := 0, s.GetToMove()

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

// MaterialAdvantage() is one element of the evaluation function which
// returns an integer expressing the favorability of the material on the
// board (regardless of its location on the board).
func (s *State) MaterialAdvantage() int {
	value, bishops, enemyBishops := 0, 0, 0
	king, enemyKing := false, false
	player := s.GetToMove()

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

// MaterialValue() defines the value of each Piece
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

// UCTSearch() is the UCT algorithm
func UCTSearch(s *State) *State {
// create root node v0 with state s0
// while within computational budget do
//   v1 = TreePolicy(v0)
//   delta = DefaultPolicy(s(v1))
//   BackupNegaMax(v1, delta)
// return a(bestChild(v0, 0))
}

// TreePolicy returns the best state to UCTSearch
func TreePolicy(v *State) *State {
// while v is nonterminal do
//   if v not fully expanded then
//      return Expand(v)
//   else
//      v = BestChild(v, e)
//   return v
}

// Expand() expands the states and returns values to TreePolicy()
func Expand(v *State) *State {
// choose a element of untried actions from A(s(v))
// add new child v_prime to v
//   with s(v_prime) = f(s(v),a)
//   and a(v_prime) = a
// return v_prime
}

// BestChild() Chooses the best child state and returns it to UTCSearch()
func BestChild(v *State, c int) {
//  return argmax v_prime element of children of v Q(v_prime)/N(v_prime) + c(sqrt((2ln * N(v))/N(v_prime)
}

// DefaultPolicy() is used by UTCSearch to determine the value assigned to a move
func DefaultPolicy(s *State) {
// while s is non-terminal do
//   choose a element of A(s) uniformly at random
//   s = f(s, a)
// return reward for state s
}

// BackupNegamax() is used to pick the best move from UTCSearch()
func BackupNegamax(v *State, d int)
{
// while v is not null do
//   N(v) = N(v) + 1
//   Q(v) = Q(v) + delta
//   d = -d
}

}


