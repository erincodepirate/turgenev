package main

import (
	"container/list"
	"fmt"
	"os"
	"time"
	"unicode"
)

const Log = "/tmp/turgenev.log"

type MoveRepresentation byte

const (
	Coordinate MoveRepresentation = iota
	Algebraic
)

type IOMode byte

const (
	TUI IOMode = iota
	Xboard
)

var (
	Mode IOMode = TUI
	Orientation Color = White
)

func Prompt(s *State) (next *State, a Action) {
	moveMap, choice := StringsToStates(s), "FIRST"

	for moveMap[choice] == nil {
		if choice != "FIRST" && Mode == Xboard {
			if IsMove(choice) {
				fmt.Printf("Illegal move: " + choice + "\n")
				PrintLog("\t\t\tOUTPUT: Illegal move: " + choice + "\n")
			}
		}

		switch choice {
		case "xboard":
			Mode = Xboard
			fmt.Printf("\n")
		case "tui":
			Mode = TUI
			fmt.Printf("\nSwitched to TUI mode...\n\n")
		case "help":
			PrintHelp()
		case "reprint":
			PrintState(s, Orientation)
		case "rotate":
			Orientation = Opponent(Orientation)
			PrintState(s, Orientation)
		case "moves":
			if Mode == TUI { PrintPossibleMoves(s) }
		case "white":
			fallthrough
		case "switch":
			next, a = nil, SetCompWhite
			return
		case "quit":
			if Mode == TUI { fmt.Printf("\nBye!\n\n") }
			os.Exit(0)
		default:
			if choice != "FIRST" && Mode == TUI {
				fmt.Printf("\nI didn't understand that. Type 'help' " +
					   "for a list of things I understand.\n\n")
			}
		}

		if Mode == TUI { fmt.Printf("Your move: ") }
		fmt.Scanf("%s", &choice)
		PrintLog("INPUT: " + choice + "\n")
	}

	next, a = moveMap[choice], MakeMove
	return
}

func StringsToStates(start *State) map[string]*State {
	successors := start.LegalSuccessors()
	m := make(map[string]*State)

	for e := successors.Front(); e != nil; e = e.Next() {
		m[MoveString(start, e.Value.(*State), Coordinate)] = e.Value.(*State)
		m[MoveString(start, e.Value.(*State), Algebraic)] = e.Value.(*State)
	}

	return m
}

func PrintResults(final *State) {
	if final.InCheck() {
		if final.ToMove == Black {
			if Mode == TUI {
				fmt.Printf("\nCheckmate. White wins.\n\n")
			} else {
				fmt.Printf("result 1-0 {white mates}\n")
			}
			PrintLog("\t\t\tOUTPUT: result 1-0 {white mates}\n")
		} else {
			if Mode == TUI {
				fmt.Printf("\nCheckmate. Black wins.\n\n")
			} else {
				fmt.Printf("result 0-1 {black mates}\n")
			}
			PrintLog("\t\t\tOUTPUT: result 0-1 {black mates}\n")
		}
	} else {
		if Mode == TUI {
			fmt.Printf("\nStalemate. Nobody wins.\n\n")
		} else {
			fmt.Printf("result 1/2-1/2 {stalemate}\n")
		}
		PrintLog("\t\t\tOUTPUT: result 1/2-1/2 {stalemate}\n")
	}
}

func PrintHelp() {
	fmt.Printf("\n\tTURGENEV COMMANDS\n\n")

	fmt.Printf("help\t\tPrint this menu\n")
	fmt.Printf("moves\t\tPrint the possible moves (in coordinate notation)\n")
	fmt.Printf("reprint\t\tPrint the board again\n")
	fmt.Printf("rotate\t\tView the board from the other side\n")
	fmt.Printf("switch\t\tTrade places with the computer\n\n")

	fmt.Printf("tui\t\tSwitch to human TUI IO mode\n")
	fmt.Printf("xboard\t\tSwitch to xboard IO mode\n\n")

	fmt.Printf("quit\t\tExit the program\n\n")
}

func IsMove(s string) bool {
	// obviously this can be improved...

	switch s {
	case "hard":
		fallthrough
	case "easy":
		fallthrough
	case "quit":
		fallthrough
	case "white":
		return false
	}

	if len(s) == 4 || len(s) == 5 {
		return true
	}
	return false
}

func PrintPossibleMoves(s *State) {
	moves, i := MoveList(s, Coordinate), 1

	fmt.Printf("\nThe possible moves are:\n")
	for e := moves.Front(); e != nil; e = e.Next() {
		fmt.Printf("\t%s", e.Value.(string))
		if i % 5 == 0 {
			fmt.Printf("\n")
		}
		i++
	}
	fmt.Printf("\n")
	if (i - 1) % 5 != 0 {
		fmt.Printf("\n")
	}
}

func PrintLog(str string) {
	fd, err := os.OpenFile(Log, os.O_RDWR | os.O_APPEND, 0666)
	if err != nil {
		fd, err = os.Create(Log)
	}
	if err != nil {
		panic(err)
	}
	defer fd.Close()

	now := time.Now().Unix()

	_, err = fd.WriteString(fmt.Sprintf("%d  +  %d\t: %s", SessionStart,
					    now - SessionStart, str))
	if err != nil { panic(err) }
}

func MoveList(s *State, mr MoveRepresentation) *list.List {
	successors, moves := s.LegalSuccessors(), list.New()

	for e := successors.Front(); e != nil; e = e.Next() {
		moves.PushBack(MoveString(s, e.Value.(*State), mr))
	}

	return moves
}

func MoveString(s1, s2 *State, mr MoveRepresentation) string {
	differences := 0
	for i := 0; i < 8; i++ {
		for j := 0; j < 8; j++ {
			if s1.GetSquare(i, j) != s2.GetSquare(i, j) {
				differences++
			}
		}
	}

	switch (differences) {
	case 2:
		return RegularMoveString(s1, s2, mr)
	case 3:
		return EnPassantMoveString(s1, s2, mr)
	case 4:
		return CastleMoveString(s1, s2, mr)
	}

	return ""
}

func RegularMoveString(s1, s2 *State, mr MoveRepresentation) string {
	var move string = ""
	var r1, c1, r2, c2 int

	for i := 0; i < 8; i++ {
		for j := 0; j < 8; j++ {
			if s1.GetSquare(i, j) != s2.GetSquare(i, j) {
				if s2.GetPiece(i, j) == Empty {
					r1, c1 = i, j
				} else {
					r2, c2 = i, j
				}
			}
		}
	}

	if mr == Coordinate {
		move = fmt.Sprintf("%c%c%c%c", File(c1), Rank(r1),
				   File(c2), Rank(r2))

		if s1.GetPiece(r1, c1) != s2.GetPiece(r2, c2) {
			move = fmt.Sprintf("%s%c", move,
					   unicode.ToLower(s2.GetRune(r2, c2)))
		}

		return move
	}

	if s1.GetPiece(r1, c1) != Pawn {
		move = fmt.Sprintf("%c%s", unicode.ToUpper(s1.GetRune(r1, c1)),
				   move)
		// Possibly redundant:
		//move = fmt.Sprintf("%s%c%c", move, File(c1), Rank(r1))
	}

	if s1.GetPiece(r2, c2) != Empty {
		if s1.GetPiece(r1, c1) == Pawn {
			move = fmt.Sprintf("%c%s", File(c1), move)
		}
		move = fmt.Sprintf("%s%c", move, 'x')
	}

	move = fmt.Sprintf("%s%c%c", move, File(c2), Rank(r2))

	if s1.GetPiece(r1, c1) == Pawn && (r2 == 7 || r2 == 0) {
		move = fmt.Sprintf("%s%c", move, s2.GetRune(r2, c2))
	}

	return move
}

func EnPassantMoveString(s1, s2 *State, mr MoveRepresentation) string {
	var r1, c1, r2, c2 int

	for i := 0; i < 8; i++ {
		for j := 0; j < 8; j++ {
			if s1.GetSquare(i, j) != s2.GetSquare(i, j) {
				if s2.GetPiece(i, j) != Empty {
					r2, c2 = i, j
				} else if (s1.GetColor(i, j) == Black && i == 3) ||
					  (s1.GetColor(i, j) == White && i == 4) {
					r1, c1 = i, j
				}
			}
		}
	}

	if mr == Coordinate {
		return fmt.Sprintf("%c%c%c%c", File(c1), Rank(r1),
				   File(c2), Rank(r2))
	}

	return fmt.Sprintf("%cx%c%c", File(c1), File(c2), Rank(r2))
}

func CastleMoveString(s1, s2 *State, mr MoveRepresentation) string {
	if mr != Coordinate {
		if s1.GetSquare(0, 0) != s2.GetSquare(0, 0) ||
		   s1.GetSquare(7, 0) != s2.GetSquare(7, 0) {
			return "O-O-O"
		}
		return "O-O"
	}

	switch {
	case s1.GetSquare(0, 0) != s2.GetSquare(0, 0):
		return "e1c1"
	case s1.GetSquare(7, 0) != s2.GetSquare(7, 0):
		return "e8c8"
	case s1.GetSquare(0, 7) != s2.GetSquare(0, 7):
		return "e1g1"
	}
	return "e8g8"
}

func PrintState(s *State, player Color) {
	fmt.Printf("\n");
	for i := 7; i >= 0; i-- {
		for j := 0; j < 8; j++ {
			if player == White {
				fmt.Printf("%c ", s.GetRune(i, j))
			} else {
				fmt.Printf("%c ", s.GetRune(7 - i, 7 - j))
			}
		}
		fmt.Printf("\n");
	}
	fmt.Printf("\n");
}

func Rank(r int) rune {
	return rune(r + '1')
}

func File(c int) rune {
	return rune(c + 'a')
}

func (s *State) GetRune(row, col int) rune {
	var r rune = '.'

	switch (s.GetPiece(row, col)) {
	case Pawn:
		r = 'P'
	case Knight:
		r = 'N'
	case Bishop:
		r = 'B'
	case Rook:
		r = 'R'
	case Queen:
		r = 'Q'
	case King:
		r = 'K'
	}

	if s.GetPiece(row, col) != Empty &&
	   s.GetColor(row, col) == Black {
		r = unicode.ToLower(r)
	}

	return r
}
