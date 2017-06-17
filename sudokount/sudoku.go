/*
Sudoku: Based on Peter Norvig constraint propagation search/solution (http://norvig.com/sudoku.html)

Board is not limited to 9x9 (square=3), It support square=n and grid=n^2xn^2
*/

package main

import (
	"fmt"
	"io"
	"os"
)

// count all possible solution
const SUDOKU_COUNT_MODE = true

// ---------------   data type and structure --------------

type Cell struct {
	r, c int
}

type UnitList map[int][][][]Cell
type Peers map[int][][]Cell
type Values map[int][]Cell

type Board struct {
	unit_size int     // unit size = n
	dim       int     // dim = grid widht/height = n*n = N
	data      [][]int // total data/cell = n^2 x n^2 = n^4
}

type Sudoku struct {
	n         int // UnitList
	dim       int // dimension
	peer_size int
	sol_count int64
	grid      Board
	unit_list UnitList
	peers     Peers
	values    Values
}

func NewBoard(sz int) Board {
	N := sz * sz
	board := Board{unit_size: sz, dim: N, data: make([][]int, N)}

	for i := 0; i < board.dim; i++ {
		board.data[i] = make([]int, board.dim)
	}
	return board
}

func (b Board) clone() Board {
	// ToDo: deep copy
	board := Board{unit_size: b.unit_size, dim: b.dim, data: make([][]int, b.dim)}
	for i := 0; i < board.dim; i++ {
		board.data[i] = make([]int, board.dim)
		copy(board.data[i], b.data[i])
	}

	return b
}

func (b Board) get(r, c int) int {
	return b.data[r][c]
}
func (b Board) set(r, c, v int) {
	b.data[r][c] = v
}

// --------------------------------------------------------

// ------------- Problem solving functions ----------------
func eliminate() {}
func assign() {}

func search(b Board, status bool) bool {
	return false
}

func solve(b Board) bool {
	return search(b, false)
}

// Create Unit List of each cell, 3 unit set (row wise,column wise, square unit)
// Example:
// row unit => [row][1], [row][2]...[row][unit_size]
// col unit => [1][col], [2][col]...[unit_size][col]
// square unit => [n]x[n] grid
func makeUnitList(b Board) (map[int][][][]Cell) {
	ul := initUnitList(b.dim)
	// Populate
	for i := 0; i < b.dim; i++ {
		ibase := (i / b.unit_size) * b.unit_size
		for j := 0; j < b.dim; j++ {
			for pos := 0; pos < b.dim; pos++ {
				// row unit
				ul[i][j][0][pos].r = i
				ul[i][j][0][pos].c = pos
				// column unit
				ul[i][j][1][pos].r = pos
				ul[i][j][1][pos].c = j
			}

			// square unit
			jbase := (j / b.unit_size) * b.unit_size
			for pos, k := 0, 0; k < b.unit_size; k++ {
				for l := 0; l < b.unit_size; l++ {
					ul[i][j][2][pos].r = ibase + k
					ul[i][j][2][pos].c = jbase + l
					pos++
				}
			}
		}
	}
	return ul
}

// Create peer set of each cell
func makePeers(b Board, unit_list map[int][][][]Cell) (int, map[int][][]Cell) {
	// peer_size = 3 * dim - [1 + 1 + (2*n - 1)] = 3*dim - 2*n - 1
	// Here [1 + 1 + (2*n - 1)] = [1 cell from row unit + 1 cell from column unit + (2*n-1) cell from square
	pr_size := 3 * b.dim - 2 * b.unit_size - 1
	pr := initPeers(b.dim, pr_size)
	// Populate
	for i := 0; i < b.dim; i++ {
		for j := 0; j < b.dim; j++ {
			pos := 0

			for k := 0; k < b.dim; k++ {
				// row, remove cell (c = j column) = 1
				if unit_list[i][j][0][k].c != j {
					pr[i][j][pos] = unit_list[i][j][0][k]
					pos++
				}

				// column, remove cell (r = i row) = 1
				if unit_list[i][j][1][k].r != i {
					pr[i][j][pos] = unit_list[i][j][1][k]
					pos++
				}

				// square, remove all (r != i && c != j) = (2*n - 1 )
				cell := unit_list[i][j][2][k]
				if cell.r != i && cell.c != j {
					pr[i][j][pos] = cell
					pos++
				}
			}
		}
	}

	return pr_size, pr
}

// dim = dimension = n*n, n: unit_size
func initUnitList(dim int) (map[int][][][]Cell) {
	// [row][col][u][dim] = cell type, u = unit len = 3 (row+col+square)
	unit_list := make(map[int][][][]Cell, dim)

	for r := 0; r < dim; r++ {
		// col init
		unit_list[r] = make([][][]Cell, dim)

		for c := 0; c < dim; c++ {
			// unit size = 3
			unit_list[r][c] = make([][]Cell, 3)

			// 0 : row unit, 1: col unit, 2: box unit
			unit_list[r][c][0] = make([]Cell, dim)
			unit_list[r][c][1] = make([]Cell, dim)
			unit_list[r][c][2] = make([]Cell, dim)
		}
	}
	return unit_list
}

// dim = dimension = n*n, n: unit_size
// peer_size = 3 * dim - [1 + 1 + (2*n - 1)] = 3*dim - 2*n - 1
// Here [1 + 1 + (2*n - 1)] = [1 cell from row unit + 1 cell from column unit + (2*n-1) cell from square
func initPeers(dim, peer_size int) (map[int][][]Cell) {
	peers := make(map[int][][]Cell, dim)
	for r := 0; r < dim; r++ {
		peers[r] = make([][]Cell, dim)
		for c := 0; c < dim; c++ {
			peers[r][c] = make([]Cell, peer_size)
		}
	}
	return peers
}

// dim = dimension = n*n, n: unit_size
func initValues(dim int) (map[int][]Cell) {
	values := make(map[int][]Cell, dim)
	for r := 0; r < dim; r++ {
		values[r] = make([]Cell, dim)
	}
	return values
}

func create_sudoku(b Board) Sudoku {
	r := Sudoku{}
	r.n = b.unit_size
	r.dim = b.dim
	r.grid = b
	// r.peer_size = 3 * r.dim - 2 * r.n - 1
	r.unit_list = makeUnitList(b)
	r.peer_size, r.peers = makePeers(b, r.unit_list)
	r.values = initValues(r.dim)
	return r
}

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
	for r := 0; r < board.dim; r++ {
		for c := 0; c < board.dim; c++ {
			_, err := fmt.Scan(&board.data[r][c])
			if err != nil {
				return Board{}, err
			}
		}
	}

	return board, nil
}

func print_board(b Board) {
	fmt.Println(b.unit_size)
	for r := 0; r < b.dim; r++ {
		for c := 0; c < b.dim; c++ {
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

	//print_board(board)

	s := create_sudoku(board)
	// solve(board)
}
