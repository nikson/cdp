/*
Sudoku: Based on Peter Norvig constraint propagation search/solution
*/


package main

import (
	"fmt"
	"io"
	"os"
)

// ---------------   data type and structure --------------
type Board struct {
	// box size = n, total data/cell = n^2 x n^2 = n^4
	box_size  int
	grid_size int
	data      [][]int
}

func NewBoard(sz int) Board {
	board := Board{box_size: sz, grid_size: sz * sz, data: make([][]int, sz * sz)}
	for i := 0; i < board.grid_size; i++ {
		board.data[i] = make([]int, board.grid_size)
	}
	return board
}

func (b Board) clone() Board {
	// ToDo: deep copy
	board := Board{box_size: b.box_size, grid_size: b.grid_size, data: make([][]int, b.grid_size)}
	for i := 0; i < board.grid_size; i++ {
		board.data[i] = make([]int, board.grid_size)
		copy(board.data[i], b.data[i])
	}

	return b
}
// --------------------------------------------------------

// ------------- Problem solving functions ----------------
func eliminate() {}
func assign() {}
func search() {}
func solve() {}
// --------------------------------------------------------

// ---------------- read input/output ---------------------
func read_board(rd io.Reader) (Board, error) {
	// read size, n
	var n int
	_, err := fmt.Scan(&n)

	if err != nil {
		return Board{}, err
	}
	// create board
	board := NewBoard(n)

	// read the cells of grid
	for r := 0; r < board.grid_size; r++ {
		for c := 0; c < board.grid_size; c++ {
			_, err := fmt.Scan(&board.data[r][c])
			if err != nil {
				return Board{}, err
			}
		}
	}

	return board, nil
}

func print_board(b Board) {
	fmt.Println(b.box_size)
	for r := 0; r < b.grid_size; r++ {
		for c := 0; c < b.grid_size; c++ {
			fmt.Printf("%d ", b.data[r][c])
		}
		fmt.Println()
	}
}
// --------------------------------------------------------
func main() {

	board, err := read_board(os.Stdin)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	print_board(board)
}
