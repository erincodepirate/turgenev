package main

import "bytes"

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

type Color byte

const (
	None Color = iota
	White
	Black
)

type Square byte

type State struct {
	board []Square
	ToMove Color
	Predecessor *State
}

func (s *State) Hash() string {
	var buffer bytes.Buffer

	board := s.board
	for i := 0; i < 64; i++ {
		buffer.WriteRune(rune(0xAE + board[i]))
	}

	return buffer.String()
}

func CreateState() *State {
	var s *State = new(State)
	s.board = make([]Square, 64)
	return s
}

// Return a copy of the given state
func CopyState(s *State) *State {
	var t *State = CreateState()
	copy(t.board, s.board)
	t.Predecessor = s
	return t
}

// Return the color of the piece in square (row, col)
func (s *State) GetColor(row, col int) Color {
	return Color((s.board[(row << 3) + col] & 0x18) >> 3)
}

// Set the color of the piece in square (row, col)
func (s *State) SetColor(row, col int, color Color) {
	s.board[(row << 3) + col] &= 0x27
	s.board[(row << 3) + col] |= (Square(color) << 3)
}

// Return the piece in square (row, col)
func (s *State) GetPiece(row, col int) Piece {
	return Piece(s.board[(row << 3) + col] & 0x7)
}

// Set the piece in square (row, col)
func (s *State) SetPiece(row, col int, piece Piece) {
	s.board[(row << 3) + col] &= 0x38
	s.board[(row << 3) + col] |= Square(piece)
}

// Return true iff square (row, col) contains a piece that has moved
// (or pawn that has advanced two squares -- used for en passant captures)
func (s *State) GetMoved(row, col int) bool {
	return (s.board[(row << 3) + col] & 0x20) != 0
}

// Set whether square (row, col) contains a piece that's moved
func (s *State) SetMoved(row, col int, moved bool) {
	if moved {
		s.board[(row << 3) + col] |= 0x20
	} else {
		s.board[(row << 3) + col] &= 0x1F
	}
}

// Return the value of square (row, col)
func (s *State) GetSquare(row, col int) Square {
	return s.board[(row << 3) + col]
}

// Set square (row, col) to the given value
func (s *State) SetSquare(row, col int, sqr Square) {
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

	s.ToMove = White
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
	cs.ToMove = Opponent(s.ToMove)
	il := cs.LegalSuccessors()

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
