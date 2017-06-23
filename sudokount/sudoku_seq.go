/*
Sudoku: Based on Peter Norvig constraint propagation search/solution (http://norvig.com/sudoku.html)

Board is not limited to 9x9 (square=3), It support square=n and grid=n^2xn^2
*/

package main

import (
	"fmt"
	"io"
	"os"
	"errors"
)

//  Solution mode: true = count all possible solution, false = find single solution
const SUDOKU_COUNT_MODE = true
const MAX_N = 8                        // dim = n^4

// ---------------   data type and structure --------------
type Cell struct {
	r, c int
}

// Cell value hold the possible digit for that cell
type Cell_Value struct {
	v []int
}

type UnitList map[int][][][]Cell
type Peers map[int][][]Cell
type Values [][]Cell_Value

type Board struct {
	unit_size int     // unit size = n
	dim       int     // dim = grid width/height = n*n = N
	data      [][]int // total data/cell = n^2 x n^2 = n^4
}

type Sudoku struct {
	n              int    // unit size
	dim            int    // dimension
	peer_size      int
	total_solution int64
	grid           Board  // initial state of board
	unit_list      UnitList
	peers          Peers
	values         Values // partially solved or solved state of board
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

// value range from 1 to Dim
func (cv *Cell_Value) set(pos int) {
	if pos <= len(cv.v) {
		cv.v[pos - 1] = 1
	}
}

func (cv *Cell_Value) unset(pos int) {
	if pos <= len(cv.v) {
		cv.v[pos - 1] = 0
	}
}

func (cv Cell_Value) get(pos int) int {
	ret := 0
	if pos <= len(cv.v) {
		ret = cv.v[pos - 1]
	}
	return ret
}

func (cv Cell_Value) count() int {
	t := 0
	for _, bit := range cv.v {
		t = t + bit
	}
	return t
}

// digit used for this set, set_bit = 1
func (cv Cell_Value) digit_get() int {
	for t, bit := range cv.v {
		if bit == 1 {
			return (t + 1)
		}
	}
	return -1;
}

// Solved: If values grid hold 1 digit per cell
func (item Values) solved() bool {
	dim, solved := len(item), true

	for i := 0; solved && i < dim; i++ {
		for j := 0; j < dim; j++ {
			if item[i][j].count() != 1 {
				solved = false
				break
			}
		}
	}
	return solved
}

func (item Values) clone() Values {
	dim := len(item)
	dst := make([][]Cell_Value, dim)
	for i := 0; i < dim; i++ {
		dst[i] = make([]Cell_Value, dim)
		for j := 0; j < dim; j++ {
			dst[i][j].v = make([]int, dim)
			copy(dst[i][j].v, item[i][j].v)
		}
	}
	return dst
}

// --------------------------------------------------------

// ------------- Problem solving functions ----------------

// Create Unit List of each cell, 3 unit set (row wise,column wise, square unit)
// Example:
// row unit => [row][1], [row][2]...[row][unit_size]
// col unit => [1][col], [2][col]...[unit_size][col]
// square unit => [n]x[n] grid
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
func initValues(dim int) ([][]Cell_Value) {
	values := make([][]Cell_Value, dim)
	for r := 0; r < dim; r++ {
		values[r] = make([]Cell_Value, dim)
		for c := 0; c < dim; c++ {
			values[r][c].v = make([]int, dim)
		}
	}
	return values
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

// create values, and bit-mask possible values
func makeValues(dim int) ([][]Cell_Value) {
	values := initValues(dim)
	for i := 0; i < dim; i++ {
		for j := 0; j < dim; j++ {
			for k := 1; k <= dim; k++ {
				values[i][j].set(k)
			}
		}
	}
	return values
}

// Eliminate values
func parse_grid(s Sudoku) bool {
	for i := 0; i < s.dim; i++ {
		for j := 0; j < s.dim; j++ {
			if (s.grid.data[i][j] > 0 && !assign(s, i, j, s.grid.data[i][j]) ) {
				return false
			}
		}
	}

	return true
}

func assign(s Sudoku, i, j, d int) bool {
	for v := 1; v <= s.dim; v++ {
		if v != d && !eliminate(s, i, j, v) {
			return false
		}
	}
	return true
}

// Remove cell values from initial set and peer
func eliminate(s Sudoku, i, j, d int) bool {

	if ( s.values[i][j].get(d) == 0) {
		return true
	}

	s.values[i][j].unset(d)

	count := s.values[i][j].count()

	if count == 0 {
		// contradict
		return false
	} else if count == 1 {
		// if a cell has only one value then remove this value(d) from peers
		for k := 0; k < s.peer_size; k++ {
			if ( !eliminate(s, s.peers[i][j][k].r, s.peers[i][j][k].c, s.values[i][j].digit_get())) {
				return false
			}
		}
	}

	// walk through 3 unit (row, col, square), check d is assign only once in 3 units
	for k := 0; k < 3; k++ {
		cont, pos := 0, 0
		u := s.unit_list[i][j][k]
		for x := 0; x < s.dim; x++ {
			if ( s.values[u[x].r][u[x].c].get(d) == 1) {
				cont++
				pos = x
			}
		}

		if cont == 0 {
			// contradict
			return false
		} else if cont == 1 {
			if !assign(s, u[pos].r, u[pos].c, d) {
				return false
			}
		}
	}

	return true
}

func search(s *Sudoku, status bool) bool {

	if !status {
		return status
	}

	if (s.values.solved()) {
		s.total_solution++
		return (SUDOKU_COUNT_MODE == false)
	}

	// find min I, J
	min, minI, minJ, ret := s.dim + 1, -1, -1, false

	for i := 0; i < s.dim; i++ {
		for j := 0; j < s.dim; j++ {
			used := s.values[i][j].count()
			if used > 1 && used < min {
				min, minI, minJ = used, i, j
			}
		}
	}

	for k := 1; k <= s.dim; k++ {
		if s.values[minI][minJ].get(k) == 1 {
			// backup values
			values_bkp := s.values.clone()

			if search(s, assign(*s, minI, minJ, k)) {
				ret = true
			} else {
				s.values = values_bkp
			}
		}
	}

	return ret
}

func solve(s *Sudoku) bool {
	return search(s, true)
}

func create_sudoku(b Board) Sudoku {
	r := Sudoku{}
	r.n = b.unit_size
	r.dim = b.dim
	r.grid = b
	r.total_solution = 0
	r.unit_list = makeUnitList(b)
	r.peer_size, r.peers = makePeers(b, r.unit_list)
	r.values = makeValues(r.dim)
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

	if n < 3 || n > MAX_N {
		return Board{}, errors.New("Invalid box size.")
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

func print_solution(s Sudoku) {
	fmt.Println(s.n)
	for r := 0; r < s.dim; r++ {
		for c := 0; c < s.dim; c++ {
			fmt.Printf("%d ", s.values[r][c].digit_get())
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

	if parse_grid(s) {
		// call solver function
		solve(&s)

		if SUDOKU_COUNT_MODE {
			// print total number of solution
			fmt.Println(s.total_solution)
		} else {
			// print a solved board
			print_solution(s)
		}
	} else {
		fmt.Println("Could not load puzzle.")
	}
}
