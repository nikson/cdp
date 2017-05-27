/**
	Author: Nikson Kanti Paul
*/


package main

import (
	"io"
	"bufio"
	"os"
	"fmt"
	"errors"
)

// ------------------ Data type -----------

type Board struct {
	data [][]int
	size int
}

func NewBoard(sz int) Board {
	// Be careful: slice always reference original underlaying array
	board := Board{data: make([][]int, sz), size: sz}
	for i := 0; i < sz; i++ {
		board.data[i] = make([]int, sz)
	}
	return board
}

func (src Board) clone() Board {
	dst := Board{data: make([][]int, src.size), size: src.size}
	for i := 0; i < src.size; i++ {
		row := make([]int, src.size)
		copy(row, src.data[i])  // copy(dest, src) !!!
		dst.data[i] = row
	}
	return dst
}

func (b Board) get(row int, col int) int {
	return b.data[row][col]
}

func (b Board) set(row int, col int, value int) Board {
	b.data[row][col] = value
	return b
}


// -------------------- problem solving functions ----------------------

func play(src Board) Board {
	dst := src.clone()

	for r := 0; r < src.size; r++ {
		for c := 0; c < src.size; c++ {
			// count the neighbour
			count := neighbour(src, r, c)
			life := game_of_life_status(src.data[r][c], count)
			// update new board by life value
			dst.data[r][c] = life
		}
	}

	return dst
}

// Count the neighbour of [row,col] player
func neighbour(b Board, row int, col int) int {

	// init count, row start, row end, col start, col end
	count, rs, re, cs, cr := 0, row, row, col, col

	if row > 0 {
		rs = row - 1
	}
	if row + 1 < b.size {
		re = row + 1
	}
	if col > 0 {
		cs = col - 1
	}
	if col + 1 < b.size {
		cr = col + 1
	}

	for k := rs; k <= re; k++ {
		for l := cs; l <= cr; l++ {
			count += b.data[k][l];
		}
	}

	// minus self from count
	count -= b.data[row][col]
	return count
}

// game of life rules of new life, died or survive
func game_of_life_status(current_status int, neighbour int) int {
	// current status of life (0/1), current status remain same if neighbour=2
	life := current_status

	if neighbour < 2 || neighbour > 3 {
		// died for suffocation
		life = 0
	} else if neighbour == 3 {
		// add new born in village
		life = 1
	}
	//else if neighbour == 2 {
	//	// Hasta la vista = see you again in next round
	//	life := current_status
	//}

	return life
}

// ------------  input, output -----------

func read_input(rd io.Reader) (Board, int, error) {
	scanner := bufio.NewScanner(rd)
	// data structure for input data types

	var size, step int
	//var line string

	if scanner.Scan() {
		line := scanner.Text()
		_, err := fmt.Sscanf(line, "%d %d", &size, &step)
		if err != nil {
			return Board{}, 0, errors.New("Invalid parameter")
		}
	}

	board := NewBoard(size)

	// FIXME: there is a flaw in C sequential code, first line is empyt line, because of that i'm ignoring first line
	for i := 1; i < size; i++ {
		if scanner.Scan() {
			line := scanner.Text()
			for k, c := range line {
				v := 0
				if string(c) == "x" {
					v = 1
				}
				board.set(i, k, v)
			}
		}
	}

	return board, step, nil
}

func print_board(b Board) {
	for i := 0; i < b.size; i++ {
		for k := 0; k < b.size; k++ {
			if ( b.data[i][k] == 1 ) {
				fmt.Print("x")
			} else {
				fmt.Print(" ")
			}

		}
		fmt.Println()
	}
}

// -------------- input/output ----------------

func main() {

	board, step, err := read_input(os.Stdin)

	//if err == nil {
	//	fmt.Println("Inital board")
	//	print_board(board)
	//}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	for i := 0; i < step; i++ {
		// clone board inside the func and return calculated fresh copy
		next := play(board)
		// reference new board
		board = next
	}

	// print final board
	print_board(board)
}