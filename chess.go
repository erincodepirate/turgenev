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
	"bytes"
	"container/list"
)

// Type of a chessman, including empty squares
type Piece byte
const (
	Empty Piece = iota
	Pawn
	Knight
	Bishop
	Rook
	Queen
	King
)

// Color of a chessman ('None' is for empty squares).
type Color byte
const (
	None Color = iota
	White
	Black
)

// From the right, the square byte has 3 bits for the piece, 2 bits for
// the color, and one bit for whether or not the piece has moved.
type square byte
const (
	pieceMask = 0x07
	colorMask = 0x18
	movedMask = 0x20
)

// The fundamental representation of the state of the board
type State struct {
	board []square
	toMove Color
	predecessor *State
}

// Create a new state (an empty board)
func CreateState() *State {
	var s *State = new(State)
	s.board = make([]square, 64)
	return s
}

// Return an exact copy of the given state
func CopyState(s *State) *State {
	var t *State = CreateState()
	copy(t.board, s.board)
	t.toMove = s.toMove
	t.predecessor = s.predecessor
	return t
}

// Return the player to move
func (s *State) GetToMove() Color {
	return s.toMove
}

// Set the player to move
func (s *State) SetToMove(player Color) {
	s.toMove = player
}

// Return a state's immediate predecessor
func (s *State) GetPredecessor() *State {
	return s.predecessor
}

// Set a state's immediate predecessor
func (s *State) SetPredecessor(p *State) {
	s.predecessor = p
}

// Return the color of the piece in square (row, col)
func (s *State) GetColor(row, col int) Color {
	return Color((s.board[(row << 3) + col] & colorMask) >> 3)
}

// Set the color of the piece in square (row, col)
func (s *State) SetColor(row, col int, color Color) {
	s.board[(row << 3) + col] &= pieceMask | movedMask
	s.board[(row << 3) + col] |= (square(color) << 3)
}

// Return the piece in square (row, col)
func (s *State) GetPiece(row, col int) Piece {
	return Piece(s.board[(row << 3) + col] & pieceMask)
}

// Set the piece in square (row, col)
func (s *State) SetPiece(row, col int, piece Piece) {
	s.board[(row << 3) + col] &= colorMask | movedMask
	s.board[(row << 3) + col] |= square(piece)
}

// Return true iff square (row, col) contains a piece that has moved
// (or pawn that has advanced two squares -- used for en passant captures)
func (s *State) GetMoved(row, col int) bool {
	return (s.board[(row << 3) + col] & movedMask) != 0
}

// Set whether square (row, col) contains a piece that's moved
func (s *State) SetMoved(row, col int, moved bool) {
	if moved {
		s.board[(row << 3) + col] |= movedMask
	} else {
		s.board[(row << 3) + col] &= pieceMask | colorMask
	}
}

// Clear the given square (piece = Empty, color = None, moved = false)
func (s *State) ClearSquare(row, col int) {
	s.board[(row << 3) + col] = 0
}

// Return the value of square (row, col)
func (s *State) getSquare(row, col int) square {
	return s.board[(row << 3) + col]
}

// Set square (row, col) to the given value
func (s *State) setSquare(row, col int, sqr square) {
	s.board[(row << 3) + col] = sqr
}

// Return the state of a freshly set board
func InitialState() *State {
	s := CreateState()

	s.SetPiece(0, 0, Rook);
	s.SetPiece(7, 0, Rook);
	s.SetPiece(0, 7, Rook);
	s.SetPiece(7, 7, Rook);

	s.SetPiece(0, 1, Knight);
	s.SetPiece(7, 1, Knight);
	s.SetPiece(0, 6, Knight);
	s.SetPiece(7, 6, Knight);

	s.SetPiece(0, 2, Bishop);
	s.SetPiece(7, 2, Bishop);
	s.SetPiece(0, 5, Bishop);
	s.SetPiece(7, 5, Bishop);

	s.SetPiece(0, 3, Queen);
	s.SetPiece(7, 3, Queen);

	s.SetPiece(0, 4, King);
	s.SetPiece(7, 4, King);

	for i := 0; i < 8; i++ {
		for j := 0; j < 8; j++ {
			if i < 2 {
				s.SetColor(i, j, White)
			} else if i > 5 {
				s.SetColor(i, j, Black)
			}
			if i == 1 || i == 6 {
				s.SetPiece(i, j, Pawn)
			}
			if i > 1 && i < 6 {
				s.SetPiece(i, j, Empty)
			}
			s.SetMoved(i, j, false)
		}
	}

	s.SetToMove(White)
	return s
}

// Return the opponent of the given player
func Opponent(player Color) Color {
	switch (player) {
	case White:
		return Black
	case Black:
		return White
	}

	return None
}

// Return true iff the player to move is in check
func (s *State) InCheck() bool {
	cs := CopyState(s)
	cs.SetToMove(Opponent(s.GetToMove()))
	il := cs.Successors()

	for e := il.Front(); e != nil; e = e.Next() {
		if e.Value.(*State).LostKing() {
			return true
		}
	}

	return false
}

// Return true iff there are fewer than two kings on the board
func (s *State) LostKing() bool {
	kings := 0

	for i := 0; i < 8; i++ {
		for j := 0; j < 8; j++ {
			if s.GetPiece(i, j) == King {
				kings++
			}
		}
	}

	return kings < 2
}

// Return a list of states that can (strictly) legally follow
func (s *State) LegalSuccessors() *list.List {
	l := s.Successors()

	for e := l.Front(); e != nil; {
		successor := e.Value.(*State)
		successorResults := successor.Successors()
		valid, castled := true, false

		// it is illegal to put oneself in check...
		for f := successorResults.Front(); f != nil; f = f.Next() {
			successorResult := f.Value.(*State)
			if successorResult.LostKing() {
				valid = false
				break
			}
		}

		// Let's see if we castled...
		differences := 0
		for i := 0; i < 8; i++ {
			for j := 0; j < 8; j++ {
				if s.getSquare(i, j) != successor.getSquare(i, j) {
					differences++
				}
			}
		}
		if differences == 4 {
			castled = true
		}

		if valid && castled {
			// it is illegal to castle out of check...
			if s.InCheck() {
				valid = false
			// it is illegal to castle through check...
			} else if s.castledThroughCheck(successor) {
				valid = false
			}
		}

		next := e.Next()
		if !valid {
			l.Remove(e)
		}
		e = next
	}

	return l
}

// Return a list of states that could legally follow if not for restrictions
// on putting oneself in check or castling through/out-of check. Offered
// because it's faster than its stricter counterpart (and front-end),
// LegalSuccessors(), while being sufficient for many search purposes.
func (s *State) Successors() *list.List {
	l := list.New()

	for i := 0; i < 8; i++ {
		for j := 0; j < 8; j++ {
			if s.GetColor(i, j) == s.GetToMove() {
				switch (s.GetPiece(i, j)) {
				case Pawn:
					pushPawns(l, s, i, j)
				case Knight:
					pushKnights(l, s, i, j)
				case Bishop:
					pushDiagonals(l, s, i, j)
				case Rook:
					pushStraights(l, s, i, j)
				case Queen:
					pushDiagonals(l, s, i, j)
					pushStraights(l, s, i, j)
				case King:
					pushKings(l, s, i, j)
				}
			}
		}
	}

	return l
}

// Return a representation of the state as a Unicode string
func (s *State) UnicodeKey() string {
	var buffer bytes.Buffer

	board := s.board
	for i := 0; i < 64; i++ {
		buffer.WriteRune(rune(0xAE + board[i]))
	}
	buffer.WriteRune(rune(0xAE + s.toMove))

	return buffer.String()
}

// Return a state corresponding to the given string (the state will be
// identical to the one that made the hash except for having a null
// predecessor)
// func StateFromUnicodeKey(key string) *State {
// 	s := CreateState()
//
// 	return s
// }

// Return true iff moving from s to t involves castling through check
func (s *State) castledThroughCheck(t *State) bool {
	var row, col int
	if s.GetToMove() == White {
		row = 0
	} else {
		row = 7
	}

	if t.GetPiece(row, 2) == King {
		col = 3
	} else {
		col = 5
	}

	rookPostCastle := t.getSquare(row, col)

	// If your opponent can take the rook that moved immediately
	// following a castle, then you castled through check.
	l := t.Successors()
	for e := l.Front(); e != nil; e = e.Next() {
		if e.Value.(*State).getSquare(row, col) != rookPostCastle {
			return true
		}
	}

	return false
}

// Add states to l that follow from s due to the movement of a pawn from the
// given position. Only "mechanical" legality is necessary; the moves made may
// include putting oneself in check, etc.
func pushPawns(l *list.List, s *State, row, col int) {
	var forward, forward_2 int

	if s.GetColor(row, col) == White {
		forward = 1
	} else {
		forward = -1
	}
	forward_2 = forward << 1

	if s.GetPiece(row + forward, col) == Empty {
		pushPawnMoveResult(l, s, row, col, forward, 0, false)
		if (s.GetColor(row, col) == White && row == 1) ||
		   (s.GetColor(row, col) == Black && row == 6) {
			if s.GetPiece(row + forward_2, col) == Empty {
				   pushPawnMoveResult(l, s, row, col, forward_2, 0, false)
			}
		}
	}

	// regular captures...
	if col - 1 >= 0 && s.GetPiece(row + forward, col - 1) != Empty &&
	   s.GetColor(row + forward, col - 1) != s.GetColor(row, col) {
		   pushPawnMoveResult(l, s, row, col, forward, -1, false)
	}

	if col + 1 < 8 && s.GetPiece(row + forward, col + 1) != Empty &&
	   s.GetColor(row + forward, col + 1) != s.GetColor(row, col) {
		   pushPawnMoveResult(l, s, row, col, forward, 1, false)
	}

	// en passant captures...
	if (s.GetToMove() == Black && row == 3) || (s.GetToMove() == White && row == 4) {
		if col - 1 >= 0 && s.GetMoved(row, col - 1) &&
			s.GetPiece(row, col - 1) == Pawn &&
			s.GetColor(row, col - 1) != s.GetColor(row, col) &&
			s.GetPredecessor().GetPiece(row, col - 1) == Empty {
			pushPawnMoveResult(l, s, row, col, forward, -1, true)
		}

		if col + 1 < 8 && s.GetMoved(row, col + 1) &&
			s.GetPiece(row, col + 1) == Pawn &&
			s.GetColor(row, col + 1) != s.GetColor(row, col) &&
			s.GetPredecessor().GetPiece(row, col + 1) == Empty {
			pushPawnMoveResult(l, s, row, col, forward, 1, true)
		}
	}
}

// Add states to l that follow from s due to the movement of a knight. Only
// "mechanical" legality is necessary; the moves made may include putting
// oneself in check, etc.
func pushKnights(l *list.List, s *State, row, col int) {
	color := s.GetToMove()

	if row + 2 < 8 && col + 1 < 8 &&
	   s.GetColor(row + 2, col + 1) != color {
		pushMoveResult(l, s, row, col, 2, 1)
	}

	if row + 2 < 8 && col - 1 >= 0 &&
	   s.GetColor(row + 2, col - 1) != color {
		pushMoveResult(l, s, row, col, 2, -1)
	}

	if row + 1 < 8 && col + 2 < 8 &&
	   s.GetColor(row + 1, col + 2) != color {
		pushMoveResult(l, s, row, col, 1, 2)
	}

	if row + 1 < 8 && col - 2 >= 0 &&
	   s.GetColor(row + 1, col - 2) != color {
		pushMoveResult(l, s, row, col, 1, -2)
	}

	if row - 1 >= 0 && col + 2 < 8 &&
	   s.GetColor(row - 1, col + 2) != color {
		pushMoveResult(l, s, row, col, -1, 2)
	}

	if row - 1 >= 0 && col - 2 >= 0 &&
	   s.GetColor(row - 1, col - 2) != color {
		pushMoveResult(l, s, row, col, -1, -2)
	}

	if row - 2 >= 0 && col + 1 < 8 &&
	   s.GetColor(row - 2, col + 1) != color {
		pushMoveResult(l, s, row, col, -2, 1)
	}

	if row - 2 >= 0 && col - 1 >= 0 &&
	   s.GetColor(row - 2, col - 1) != color {
		pushMoveResult(l, s, row, col, -2, -1)
	}
}

// Add states to l that follow from s due to the movement of a bishop or queen
// along diagonals. Only "mechanical" legality is necessary; the moves made
// may include putting oneself in check, etc.
func pushDiagonals(l *list.List, s *State, row, col int) {
	player := s.GetToMove()

	for i, j := row + 1, col + 1; i < 8 && j < 8; i, j = i + 1, j + 1 {
		if color := s.GetColor(i, j); color != None {
			if color == Opponent(player) {
				pushMoveResult(l, s, row, col, i - row, j - col)
			}
			break
		} else {
			pushMoveResult(l, s, row, col, i - row, j - col)
		}
	}

	for i, j := row + 1, col - 1; i < 8 && j >= 0; i, j = i + 1, j - 1 {
		if color := s.GetColor(i, j); color != None {
			if color == Opponent(player) {
				pushMoveResult(l, s, row, col, i - row, j - col)
			}
			break
		} else {
			pushMoveResult(l, s, row, col, i - row, j - col)
		}
	}

	for i, j := row - 1, col + 1; i >= 0 && j < 8; i, j = i - 1, j + 1 {
		if color := s.GetColor(i, j); color != None {
			if color == Opponent(player) {
				pushMoveResult(l, s, row, col, i - row, j - col)
			}
			break
		} else {
			pushMoveResult(l, s, row, col, i - row, j - col)
		}
	}

	for i, j := row - 1, col - 1; i >= 0 && j >= 0; i, j = i - 1, j - 1 {
		if color := s.GetColor(i, j); color != None {
			if color == Opponent(player) {
				pushMoveResult(l, s, row, col, i - row, j - col)
			}
			break
		} else {
			pushMoveResult(l, s, row, col, i - row, j - col)
		}
	}
}

// Add states to l that follow from s due to the movement of a rook or queen
// in a straight line. Only "mechanical" legality is necessary; the moves made
// may include putting oneself in check, etc.
func pushStraights(l *list.List, s *State, row, col int) {
	player := s.GetToMove()

	for i := row + 1; i < 8; i++ {
		if color := s.GetColor(i, col); color != None {
			if color == Opponent(player) {
				pushMoveResult(l, s, row, col, i - row, 0)
			}
			break
		} else {
			pushMoveResult(l, s, row, col, i - row, 0)
		}
	}

	for i := row - 1; i >= 0; i-- {
		if color := s.GetColor(i, col); color != None {
			if color == Opponent(player) {
				pushMoveResult(l, s, row, col, i - row, 0)
			}
			break
		} else {
			pushMoveResult(l, s, row, col, i - row, 0)
		}
	}

	for i := col + 1; i < 8; i++ {
		if color := s.GetColor(row, i); color != None {
			if color == Opponent(player) {
				pushMoveResult(l, s, row, col, 0, i - col)
			}
			break
		} else {
			pushMoveResult(l, s, row, col, 0, i - col)
		}
	}

	for i := col - 1; i >= 0; i-- {
		if color := s.GetColor(row, i); color != None {
			if color == Opponent(player) {
				pushMoveResult(l, s, row, col, 0, i - col)
			}
			break
		} else {
			pushMoveResult(l, s, row, col, 0, i - col)
		}
	}
}

// Add states to l that follow from s due to the movement of a king from the
// given location. Only "mechanical" legality is necessary; the moves made may
// include putting oneself in check or castling through/out-of check, etc.
func pushKings(l *list.List, s *State, row, col int) {
	player := s.GetToMove()

	if row - 1 >= 0 {
		if s.GetColor(row - 1, col) != player {
			pushMoveResult(l, s, row, col, -1, 0)
		}

		if col - 1 >= 0 && s.GetColor(row - 1, col - 1) != player {
			pushMoveResult(l, s, row, col, -1, -1)
		}

		if col + 1 < 8 && s.GetColor(row - 1, col + 1) != player {
			pushMoveResult(l, s, row, col, -1, 1)
		}
	}

	if row + 1 < 8 {
		if s.GetColor(row + 1, col) != player {
			pushMoveResult(l, s, row, col, 1, 0)
		}

		if col - 1 >= 0 && s.GetColor(row + 1, col - 1) != player {
			pushMoveResult(l, s, row, col, 1, -1)
		}

		if col + 1 < 8 && s.GetColor(row + 1, col + 1) != player {
			pushMoveResult(l, s, row, col, 1, 1)
		}
	}

	if col - 1 >= 0 && s.GetColor(row, col - 1) != player {
		pushMoveResult(l, s, row, col, 0, -1)
	}

	if col + 1 < 8 && s.GetColor(row, col + 1) != player {
		pushMoveResult(l, s, row, col, 0, 1)
	}

	if (player == White && row != 0) || (player == Black && row != 7) ||
	   s.GetMoved(row, col) {
		return
	}

	// King's side castling...
	if s.GetColor(row, 5) == None && s.GetColor(row, 6) == None &&
	   !s.GetMoved(row, 7) {
		cs := CopyState(s)
		cs.SetPredecessor(s)
		cs.setSquare(row, 6, s.getSquare(row, 4))
		cs.setSquare(row, 5, s.getSquare(row, 7))
		cs.SetMoved(row, 5, true)
		cs.SetMoved(row, 6, true)

		cs.ClearSquare(row, 4)
		cs.ClearSquare(row, 7)

		cs.SetToMove(Opponent(player))
		l.PushBack(cs)
	}

	// Queen's side castling...
	if s.GetColor(row, 1) == None && s.GetColor(row, 2) == None &&
	   s.GetColor(row, 3) == None && !s.GetMoved(row, 7) {
		cs := CopyState(s)
		cs.SetPredecessor(s)
		cs.setSquare(row, 2, s.getSquare(row, 4))
		cs.setSquare(row, 3, s.getSquare(row, 0))
		cs.SetMoved(row, 2, true)
		cs.SetMoved(row, 3, true)

		cs.ClearSquare(row, 4)
		cs.ClearSquare(row, 0)

		cs.SetToMove(Opponent(player))
		l.PushBack(cs)
	}
}

// Helper for the pushSomePiece functions. Appends the particular state
// created by moving a piece from (r, c) to (r + dr, c + dc)
func pushMoveResult(l *list.List, s *State, r, c, dr, dc int) {
	cs := CopyState(s)
	cs.SetPredecessor(s)
	cs.setSquare(r + dr, c + dc, s.getSquare(r, c))
	cs.SetMoved(r + dr, c + dc, true)
	cs.ClearSquare(r, c)
	cs.SetToMove(Opponent(s.GetToMove()))
	l.PushBack(cs)
}

// Similar to pushMoveResult(), this function is used by pushPawns() to
// append the particular state resulting from moving a pawn from (r, c) to
// (r + dr, c + dc). The difference is that this function also deals with
// pawn-specific behaviors like promotion and en passant captures.
func pushPawnMoveResult(l *list.List, s *State, r, c, dr, dc int, ep bool) {
	cs, player := CopyState(s), s.GetColor(r, c)
	cs.SetPredecessor(s)
	cs.SetToMove(Opponent(s.GetToMove()))
	cs.setSquare(r + dr, c + dc, s.getSquare(r, c))
	cs.ClearSquare(r, c)

	// For en passant captures, we have an extra square to clean up...
	if ep {
		cs.ClearSquare(r, c + dc)
	}

	// For pawns, "moved" means "advanced 2 rows"
	// (for en passant -- otherwise we wouldn't care...)
	if dr == 2 || dr == -2 {
		cs.SetMoved(r + dr, c + dc, true)
	}

	// Pawn Promotion!
	// The reason for pushing some things at the front and some at the
	// back below is that it helps (very slightly) alpha-beta pruning
	// by putting the more desirable options toward the front of the
	// list (which is doubly linked, so there's no penalty).
	if (r + dr == 7 && player == White) ||
	   (r + dr == 0 && player == Black) {
		ps := CopyState(cs)
		ps.SetPiece(r + dr, c + dc, Knight)
		l.PushFront(ps)
		ps = CopyState(cs)
		ps.SetPiece(r + dr, c + dc, Queen)
		l.PushFront(ps)
		ps = CopyState(cs)
		ps.SetPiece(r + dr, c + dc, Rook)
		l.PushBack(ps)
		ps = CopyState(cs)
		ps.SetPiece(r + dr, c + dc, Bishop)
		l.PushBack(ps)
	} else {
		l.PushBack(cs)
	}
}

