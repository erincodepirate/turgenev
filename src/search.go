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

var (
	WallTime time.Duration
)

func (s *State) NegamaxST(depth int) *State {
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

func (s *State) Value() int {
	value := 0

	value += s.MaterialAdvantage()
	value += s.PositionAppeal()

	return value
}

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
