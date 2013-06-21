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
	"fmt"
	"time"
)

var (
	SessionStart int64 = time.Now().Unix()
	verbose bool = false
)

type Action byte

const (
	MakeMove Action = iota
	SetCompWhite
)

func main() {
	s := InitialState()

	for {
		if Mode == TUI { PrintState(s, Orientation) }
		c, a := Prompt(s)

		if a == MakeMove {
			s = c
			if Mode == TUI { PrintState(s, Orientation) }
			if s.LegalSuccessors().Len() == 0 {
				verbose = true
				break
			}
		}

		if Mode == TUI { fmt.Printf("Thinking... ") }
		c = s.NegamaxST(4)

		if c == nil {
			PrintLog("Break point B\n")
			break
		}

		if Mode == TUI {
			fmt.Printf("My move: ")
		} else {
			fmt.Printf("move ")
		}
		fmt.Println(MoveString(s, c, Coordinate))
		PrintLog("\t\t\tOUTPUT: move " + MoveString(s, c, Coordinate) + "\n")

		s = c
		if s.LegalSuccessors().Len() == 0 {
			PrintLog("Break point C\n")
			break
		}
	}

	PrintResults(s)

	for {}
}
