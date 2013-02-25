package main

import (
	"bufio"
	"container/list"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

var (
	PastStates *list.List = list.New()
	PastMoves *list.List = list.New()
	TTable map[string]int = make(map[string]int, 1024)
)

func LoadTTable(path string) {
	PrintLog("Loading transposition table from " + path + "\n")

	fd, err := os.Open(path)
	if err != nil {
		PrintLog("Failed to open transposition table file.\n")
		return
	}
	defer fd.Close()

	line, count := "", 0
	r := bufio.NewReader(fd)

	for err == nil {
		line, err = r.ReadString('\n')
		if (err != nil) { break }
		fields := strings.Fields(line)
		TTable[fields[0]], err = strconv.Atoi(fields[1])
		count++
	}

	if err == io.EOF || err == nil {
		PrintLog(fmt.Sprintf(
			"Transposition table loaded successfully (%d entries).\n", count))
	} else {
		PrintLog(fmt.Sprintf(
			"Error loading transposition table (near line %d).\n", count))
	}
}

func DumpTTable(path string) {
	PrintLog("Writing transposition table to " + path + "\n")

	fd, err := os.Create(path)
	if err != nil {
		PrintLog("Failed to create transposition table file.\n")
		return
	}
	defer fd.Close()

	count := 0
	for k, v := range TTable {
		_, err = fd.WriteString(fmt.Sprintf("%s %d\n", k, v))
		if err != nil { break }
		count++
	}

	if err == nil {
		PrintLog(fmt.Sprintf(
			"Transposition table written successfully (%d entries).\n", count))
	} else {
		PrintLog(fmt.Sprintf(
			"Error writing transposition table (near line %d).\n", count))
	}
}

func PrintTTable() {
	fmt.Printf("\n")
	for k, v := range TTable {
		fmt.Printf("%s %d\n", k, v)
	}
	fmt.Printf("\n")
}

func Contains(l *list.List, hash string) bool {
	for e := l.Front(); e != nil; e = e.Next() {
		if hash == e.Value.(string) {
			return true
		}
	}

	return false
}
