package main

import (
	"container/list"
)

func (s *State) TrueLegalSuccessors() *list.List {
	l := s.LegalSuccessors()

	for e := l.Front(); e != nil; {
		successor := e.Value.(*State)
		successorResults := successor.LegalSuccessors()
		invalid, castled := false, false

		// it is illegal to put oneself in check...
		for f := successorResults.Front(); f != nil; f = f.Next() {
			successorResult := f.Value.(*State)
			if successorResult.LostKing() {
				invalid = true
				break
			}
		}

		// Let's see if we castled...
		differences := 0
		for i := 0; i < 8; i++ {
			for j := 0; j < 8; j++ {
				if s.GetSquare(i, j) != successor.GetSquare(i, j) {
					differences++
				}
			}
		}
		if differences == 4 {
			castled = true
		}

		if castled {
			// it is illegal to castle out of check...
			if s.InCheck() {
				invalid = true
			// it is illegal to castle through check...
			} else if s.castledThroughCheck(successor) {
				invalid = true
			}
		}

		next := e.Next()
		if invalid {
			l.Remove(e)
		}
		e = next
	}

	return l
}

func (s *State) castledThroughCheck(t *State) bool {
	var row, col int
	if s.ToMove == White {
		row = 0
	} else {
		row = 7
	}

	if t.GetPiece(row, 2) == King {
		col = 3
	} else {
		col = 5
	}

	rookPostCastle := t.GetSquare(row, col)

	// If your opponent can take the rook that moved immediately
	// following a castle, then you castled through check.
	l := t.LegalSuccessors()
	for e := l.Front(); e != nil; e = e.Next() {
		if e.Value.(*State).GetSquare(row, col) != rookPostCastle {
			return true
		}
	}

	return false
}

func (s *State) LegalSuccessors() *list.List {
	l := list.New()

	for i := 0; i < 8; i++ {
		for j := 0; j < 8; j++ {
			if s.GetColor(i, j) == s.ToMove {
				switch (s.GetPiece(i, j)) {
				case Pawn:
					PushPawns(l, s, i, j)
				case Knight:
					PushKnights(l, s, i, j)
				case Bishop:
					PushDiagonals(l, s, i, j)
				case Rook:
					PushStraights(l, s, i, j)
				case Queen:
					PushDiagonals(l, s, i, j)
					PushStraights(l, s, i, j)
				case King:
					PushKings(l, s, i, j)
				}
			}
		}
	}

	return l
}

func PushPawns(l *list.List, s *State, row, col int) {
	var forward, forward_2 int

	if s.GetColor(row, col) == White {
		forward = 1
	} else {
		forward = -1
	}
	forward_2 = forward << 1

	if s.GetPiece(row + forward, col) == Empty {
		PushPawnMoveResult(l, s, row, col, forward, 0, false)
		if (s.GetColor(row, col) == White && row == 1) ||
		   (s.GetColor(row, col) == Black && row == 6) {
			if s.GetPiece(row + forward_2, col) == Empty {
				   PushPawnMoveResult(l, s, row, col, forward_2, 0, false)
			}
		}
	}

	// regular captures...
	if col - 1 >= 0 && s.GetPiece(row + forward, col - 1) != Empty &&
	   s.GetColor(row + forward, col - 1) != s.GetColor(row, col) {
		   PushPawnMoveResult(l, s, row, col, forward, -1, false)
	}

	if col + 1 < 8 && s.GetPiece(row + forward, col + 1) != Empty &&
	   s.GetColor(row + forward, col + 1) != s.GetColor(row, col) {
		   PushPawnMoveResult(l, s, row, col, forward, 1, false)
	}

	// en passant captures...
	if (s.ToMove == Black && row == 3) || (s.ToMove == White && row == 4) {
		if col - 1 >= 0 && s.GetMoved(row, col - 1) &&
			s.GetPiece(row, col - 1) == Pawn &&
			s.GetColor(row, col - 1) != s.GetColor(row, col) &&
			s.Predecessor.GetPiece(row, col - 1) == Empty {
			PushPawnMoveResult(l, s, row, col, forward, -1, true)
		}

		if col + 1 < 8 && s.GetMoved(row, col + 1) &&
			s.GetPiece(row, col + 1) == Pawn &&
			s.GetColor(row, col + 1) != s.GetColor(row, col) &&
			s.Predecessor.GetPiece(row, col + 1) == Empty {
			PushPawnMoveResult(l, s, row, col, forward, 1, true)
		}
	}
}

func PushKnights(l *list.List, s *State, row, col int) {
	color := s.ToMove

	if row + 2 < 8 && col + 1 < 8 &&
	   s.GetColor(row + 2, col + 1) != color {
		PushMoveResult(l, s, row, col, 2, 1)
	}

	if row + 2 < 8 && col - 1 >= 0 &&
	   s.GetColor(row + 2, col - 1) != color {
		PushMoveResult(l, s, row, col, 2, -1)
	}

	if row + 1 < 8 && col + 2 < 8 &&
	   s.GetColor(row + 1, col + 2) != color {
		PushMoveResult(l, s, row, col, 1, 2)
	}

	if row + 1 < 8 && col - 2 >= 0 &&
	   s.GetColor(row + 1, col - 2) != color {
		PushMoveResult(l, s, row, col, 1, -2)
	}

	if row - 1 >= 0 && col + 2 < 8 &&
	   s.GetColor(row - 1, col + 2) != color {
		PushMoveResult(l, s, row, col, -1, 2)
	}

	if row - 1 >= 0 && col - 2 >= 0 &&
	   s.GetColor(row - 1, col - 2) != color {
		PushMoveResult(l, s, row, col, -1, -2)
	}

	if row - 2 >= 0 && col + 1 < 8 &&
	   s.GetColor(row - 2, col + 1) != color {
		PushMoveResult(l, s, row, col, -2, 1)
	}

	if row - 2 >= 0 && col - 1 >= 0 &&
	   s.GetColor(row - 2, col - 1) != color {
		PushMoveResult(l, s, row, col, -2, -1)
	}
}

func PushDiagonals(l *list.List, s *State, row, col int) {
	player := s.ToMove

	for i, j := row + 1, col + 1; i < 8 && j < 8; i, j = i + 1, j + 1 {
		if color := s.GetColor(i, j); color != None {
			if color == Opponent(player) {
				PushMoveResult(l, s, row, col, i - row, j - col)
			}
			break
		} else {
			PushMoveResult(l, s, row, col, i - row, j - col)
		}
	}

	for i, j := row + 1, col - 1; i < 8 && j >= 0; i, j = i + 1, j - 1 {
		if color := s.GetColor(i, j); color != None {
			if color == Opponent(player) {
				PushMoveResult(l, s, row, col, i - row, j - col)
			}
			break
		} else {
			PushMoveResult(l, s, row, col, i - row, j - col)
		}
	}

	for i, j := row - 1, col + 1; i >= 0 && j < 8; i, j = i - 1, j + 1 {
		if color := s.GetColor(i, j); color != None {
			if color == Opponent(player) {
				PushMoveResult(l, s, row, col, i - row, j - col)
			}
			break
		} else {
			PushMoveResult(l, s, row, col, i - row, j - col)
		}
	}

	for i, j := row - 1, col - 1; i >= 0 && j >= 0; i, j = i - 1, j - 1 {
		if color := s.GetColor(i, j); color != None {
			if color == Opponent(player) {
				PushMoveResult(l, s, row, col, i - row, j - col)
			}
			break
		} else {
			PushMoveResult(l, s, row, col, i - row, j - col)
		}
	}
}

func PushStraights(l *list.List, s *State, row, col int) {
	player := s.ToMove

	for i := row + 1; i < 8; i++ {
		if color := s.GetColor(i, col); color != None {
			if color == Opponent(player) {
				PushMoveResult(l, s, row, col, i - row, 0)
			}
			break
		} else {
			PushMoveResult(l, s, row, col, i - row, 0)
		}
	}

	for i := row - 1; i >= 0; i-- {
		if color := s.GetColor(i, col); color != None {
			if color == Opponent(player) {
				PushMoveResult(l, s, row, col, i - row, 0)
			}
			break
		} else {
			PushMoveResult(l, s, row, col, i - row, 0)
		}
	}

	for i := col + 1; i < 8; i++ {
		if color := s.GetColor(row, i); color != None {
			if color == Opponent(player) {
				PushMoveResult(l, s, row, col, 0, i - col)
			}
			break
		} else {
			PushMoveResult(l, s, row, col, 0, i - col)
		}
	}

	for i := col - 1; i >= 0; i-- {
		if color := s.GetColor(row, i); color != None {
			if color == Opponent(player) {
				PushMoveResult(l, s, row, col, 0, i - col)
			}
			break
		} else {
			PushMoveResult(l, s, row, col, 0, i - col)
		}
	}
}

func PushKings(l *list.List, s *State, row, col int) {
	player := s.ToMove

	if row - 1 >= 0 {
		if s.GetColor(row - 1, col) != player {
			PushMoveResult(l, s, row, col, -1, 0)
		}

		if col - 1 >= 0 && s.GetColor(row - 1, col - 1) != player {
			PushMoveResult(l, s, row, col, -1, -1)
		}

		if col + 1 < 8 && s.GetColor(row - 1, col + 1) != player {
			PushMoveResult(l, s, row, col, -1, 1)
		}
	}

	if row + 1 < 8 {
		if s.GetColor(row + 1, col) != player {
			PushMoveResult(l, s, row, col, 1, 0)
		}

		if col - 1 >= 0 && s.GetColor(row + 1, col - 1) != player {
			PushMoveResult(l, s, row, col, 1, -1)
		}

		if col + 1 < 8 && s.GetColor(row + 1, col + 1) != player {
			PushMoveResult(l, s, row, col, 1, 1)
		}
	}

	if col - 1 >= 0 && s.GetColor(row, col - 1) != player {
		PushMoveResult(l, s, row, col, 0, -1)
	}

	if col + 1 < 8 && s.GetColor(row, col + 1) != player {
		PushMoveResult(l, s, row, col, 0, 1)
	}

	if (player == White && row != 0) || (player == Black && row != 7) ||
	   s.GetMoved(row, col) {
		return
	}

	if s.GetColor(row, 5) == None && s.GetColor(row, 6) == None &&
	   !s.GetMoved(row, 7) {
		cs := CopyState(s)
		cs.SetSquare(row, 6, s.GetSquare(row, 4))
		cs.SetSquare(row, 5, s.GetSquare(row, 7))
		cs.SetPiece(row, 4, Empty); cs.SetColor(row, 4, None)
		cs.SetPiece(row, 7, Empty); cs.SetColor(row, 7, None)
		cs.SetMoved(row, 5, true)
		cs.SetMoved(row, 6, true)
		cs.ToMove = Opponent(player)
		l.PushBack(cs)
	}

	if s.GetColor(row, 1) == None && s.GetColor(row, 2) == None &&
	   s.GetColor(row, 3) == None && !s.GetMoved(row, 7) {
		cs := CopyState(s)
		cs.SetSquare(row, 2, s.GetSquare(row, 4))
		cs.SetSquare(row, 3, s.GetSquare(row, 0))
		cs.SetPiece(row, 4, Empty); cs.SetColor(row, 4, None)
		cs.SetPiece(row, 0, Empty); cs.SetColor(row, 0, None)
		cs.SetMoved(row, 2, true)
		cs.SetMoved(row, 3, true)
		cs.ToMove = Opponent(player)
		l.PushBack(cs)
	}
}

func PushMoveResult(l *list.List, s *State, r, c, dr, dc int) {
	cs := CopyState(s)
	cs.SetSquare(r + dr, c + dc, s.GetSquare(r, c))
	cs.SetPiece(r, c, Empty)
	cs.SetColor(r, c, None)
	cs.SetMoved(r + dr, c + dc, true)
	cs.ToMove = Opponent(s.ToMove)
	l.PushBack(cs)
}

func PushPawnMoveResult(l *list.List, s *State, r, c, dr, dc int, ep bool) {
	cs, player := CopyState(s), s.GetColor(r, c)
	cs.SetSquare(r + dr, c + dc, s.GetSquare(r, c))
	cs.SetPiece(r, c, Empty)
	cs.SetColor(r, c, None)

	// en passant extras...
	if ep {
		cs.SetPiece(r, c + dc, Empty)
		cs.SetColor(r, c + dc, None)
	}

	// for pawns, moved means advanced 2 rows (for en passant)
	if dr == 2 || dr == -2 {
		cs.SetMoved(r + dr, c + dc, true)
	}

	// Because of the way this state-copy happens, there is the possibility
	// that en passant captures will be broken immediately following pawn
	// promotions (a rare enough contingency to ignore for now). This is due
	// to CopyState() setting ps.Predecessor to the intermediate value cs.
	cs.ToMove = Opponent(s.ToMove)
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

