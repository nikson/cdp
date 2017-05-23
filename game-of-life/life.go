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
	// start-end row, start-end col
	start_r, end_r, start_c, end_c := row - 1, row + 1, col - 1, col + 1

	if start_r < 0 {
		start_r = row
	}

	if end_r >= b.size {
		end_r = end_r - 1
	}

	if start_c < 0 {
		start_c = col
	}

	if end_c >= b.size {
		end_c = end_c - 1
	}

	count := 0
	for k := start_r; k <= end_r; k++ {
		for i := start_c; i <= end_c; i++ {
			count += b.data[k][i]
		}
	}

	//for _, i := range b.data[start_r:end_r+1] {
	//
	//	for _, k := range i[start_c : end_c+1] {
	//		count += k;
	//	}
	//}

	// minus self from counting
	count -= b.data[row][col]
	return count
}

// game of life rules of new life, died or survive
func game_of_life_status(current_status int, neighbour int) int {
	// current status of life (0/1), current status remain same if neighbour=2
	life := current_status

	if (neighbour < 2 || neighbour > 3) {
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

	for i := 0; i < size; i++ {
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