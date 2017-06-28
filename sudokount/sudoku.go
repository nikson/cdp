/*
Sudoku: Based on Peter Norvig constraint propagation search/solution (http://norvig.com/sudoku.html)

Board is not limited to 9x9 (square=3), It support square=n and grid=n^2xn^2

Parallel constraint propagation / parallel backtrack
*/

package main

import (
	"fmt"
	"io"
	"os"
	"errors"
	"runtime"
	"sync/atomic"
	//_ "github.com/pkg/profile"
)

//  Solution mode: true = count all possible solution, false = find single solution
const SUDOKU_COUNT_MODE = true
const MAX_N = 8                        // dim = n^4

// ---------------   data type and structure --------------
type Cell struct {
	r, c int
}

// MAX_N=8 => 8^2=64 bits number (order of count RTL) is representing a cells possible values
// Ex: 10011 = {5, 2, 1}
type CellValue struct {
	bits uint64
}

type UnitList [][][][]Cell
type Peers [][][]Cell
type Values [][]CellValue

type Board struct {
	unit_size int     // unit size = n
	dim       int     // dim = grid width/height = n*n = N
	data      [][]int // total data/cell = n^2 x n^2 = n^4
}

type Sudoku struct {
	n              int    // unit size
	dim            int    // dimension
	peer_size      int
	total_solution uint64
	grid           Board  // initial state of board
	unit_list      UnitList
	peers          Peers
	values         Values // partially solved or solved state of board
}

type TaskResult struct {
	status bool
	count  uint64
}

type Promise chan TaskResult

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
func (cv *CellValue) set(pos int) {
	cv.bits = cv.bits | ( 1 << uint8(pos - 1) )
}

func (cv *CellValue) unset(pos int) {
	// (shift pos, negate = mask) & bits
	cv.bits = cv.bits & (^( 1 << uint8(pos - 1) ))
}

func (cv CellValue) get(pos int) int {
	ret := (cv.bits >> uint8(pos - 1) ) & 1
	return int(ret)
}

func (cv CellValue) count() int {
	r := 0
	for tmp := cv.bits; tmp > 0; tmp = tmp >> 1 {
		r = r + int(1 & (tmp >> uint8(0)))
	}
	return r

	// N.B.: time cosuming operation, 40% of execution time spent for strconv operation
	//str := strconv.FormatUint(cv.bits, 2)
	//return strings.Count(str, "1")
}

//  binary(d) = 10000 = 5
func (cv CellValue) digit_get() int {
	pos := 0
	for tmp := cv.bits; tmp > 0; tmp = tmp >> 1 {
		pos++
		r := int(1 & (tmp >> uint8(0)))
		if (r == 1) {
			return pos
		}
	}
	return -1

	// N.B.: time cosuming operation, 40% of execution time spent for strconv operation
	//str := strconv.FormatUint(cv.bits, 2)
	//pos := strings.Index(str, "1")
	//if pos > -1 {
	//	return (len(str) - pos)
	//}
	//return -1;
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
	dst := make([][]CellValue, dim)
	for i := 0; i < dim; i++ {
		dst[i] = make([]CellValue, dim)
		copy(dst[i], item[i])
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
func makeUnitList(unit_size int) ([][][][]Cell) {
	// [row][col][u][dim] = cell type, u = unit len = 3 (row+col+square)
	dim := unit_size * unit_size
	ul := make([][][][]Cell, dim)
	// Prepare unit list
	for r := 0; r < dim; r++ {
		ul[r] = make([][][]Cell, dim)
		// square unit row base
		ibase := (r / unit_size) * unit_size
		for c := 0; c < dim; c++ {
			ul[r][c] = make([][]Cell, 3)
			// 0 : row unit, 1: col unit, 2: square unit
			ul[r][c][0] = make([]Cell, dim)
			ul[r][c][1] = make([]Cell, dim)
			ul[r][c][2] = make([]Cell, dim)

			for pos := 0; pos < dim; pos++ {
				// row unit
				ul[r][c][0][pos].r = r
				ul[r][c][0][pos].c = pos
				// column unit
				ul[r][c][1][pos].r = pos
				ul[r][c][1][pos].c = c
			}

			// square unit
			// square unit col base
			jbase := (c / unit_size) * unit_size
			for pos, k := 0, 0; k < unit_size; k++ {
				for l := 0; l < unit_size; l++ {
					ul[r][c][2][pos].r = ibase + k
					ul[r][c][2][pos].c = jbase + l
					pos++
				}
			}
		}
	}
	return ul
}

// Create peers set of each cell
// dim = dimension = n*n, n: unit_size
// peer_size = 3 * dim - [1 + 1 + (2*n - 1)] = 3*dim - 2*n - 1
// Here [1 + 1 + (2*n - 1)] = [1 cell from row unit + 1 cell from column unit + (2*n-1) cell from square
func makePeers(unit_size int, unit_list [][][][]Cell) (int, [][][]Cell) {
	dim := unit_size * unit_size
	// peer_size = 3 * dim - [1 + 1 + (2*n - 1)] = 3*dim - 2*n - 1
	// Here [1 + 1 + (2*n - 1)] = [1 cell from row unit + 1 cell from column unit + (2*n-1) cell from square
	peer_size := 3 * dim - 2 * unit_size - 1
	peers := make([][][]Cell, dim)
	// prepare peers
	for r := 0; r < dim; r++ {
		peers[r] = make([][]Cell, dim)
		for c := 0; c < dim; c++ {
			peers[r][c] = make([]Cell, peer_size)
			pos := 0

			for k := 0; k < dim; k++ {
				// row, remove cell (c = j column) = 1
				if unit_list[r][c][0][k].c != c {
					peers[r][c][pos] = unit_list[r][c][0][k]
					pos++
				}

				// column, remove cell (r = i row) = 1
				if unit_list[r][c][1][k].r != r {
					peers[r][c][pos] = unit_list[r][c][1][k]
					pos++
				}

				// square, remove all (r != i && c != j) = (2*n - 1 )
				cell := unit_list[r][c][2][k]
				if cell.r != r && cell.c != c {
					peers[r][c][pos] = cell
					pos++
				}
			}
		}
	}

	return peer_size, peers
}

// dim = dimension = n*n, n: unit_size
// create possible cell values
func makeValues(dim int) ([][]CellValue) {
	//values := initValues(dim)

	// create default value for CellValue
	v := CellValue{bits: 0 }
	for k := 1; k <= dim; k++ {
		v.set(k)
	}

	// create a row of CellValue
	v_row := make([]CellValue, dim)
	for k := 0; k < dim; k++ {
		v_row[k] = v
	}

	// prepare Values
	values := make([][]CellValue, dim)
	for r := 0; r < dim; r++ {
		values[r] = make([]CellValue, dim)
		// copy default cell_value row
		copy(values[r], v_row)
	}
	return values
}

// validate inputs and values
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

// try to assign value 'd' for specified (r,c)
func assign(s Sudoku, i, j, d int) bool {
	for v := 1; v <= s.dim; v++ {
		if v != d && !eliminate(s, i, j, v) {
			return false
		}
	}
	return true
}

// Remove 'd' from other values list and peers
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

// find the cell which has minimum number of available option to choose.
// minimum cell should be choosen before larger value set
func find_minimum_option_cell(s Sudoku) (int, int) {
	// find min row, col
	min, r, c := (s.dim + 1), -1, -1

	// select cell whihc has minimum number of available option to choose
	for i := 0; i < s.dim; i++ {
		for j := 0; j < s.dim; j++ {
			remain := s.values[i][j].count()
			if remain > 1 && remain < min {
				min, r, c = remain, i, j
			}
		}
	}

	return r, c
}

func search(s Sudoku, status bool) (bool, Values) {

	if !status {
		return status, s.values
	}

	if (s.values.solved()) {
		// global counter for sequential program
		solution_counter++
		s.total_solution++
		// write a solution
		return (SUDOKU_COUNT_MODE == false), s.values
	}

	minI, minJ := find_minimum_option_cell(s)
	ret := false

	//// find min I, J
	//min, minI, minJ, ret := (s.dim + 1), -1, -1, false
	//
	//// select minimum remaining values from possible cell values
	//for i := 0; i < s.dim; i++ {
	//	for j := 0; j < s.dim; j++ {
	//		remain := s.values[i][j].count()
	//		if remain > 1 && remain < min {
	//			min, minI, minJ = remain, i, j
	//		}
	//	}
	//}

	for k := 1; k <= s.dim; k++ {
		if s.values[minI][minJ].get(k) == 1 {
			value_bkp := s.values.clone()

			// search for subtree, backtracking
			ret, s.values = search(s, assign(s, minI, minJ, k))

			// ret == false, reset values from backup
			if !ret {
				s.values = value_bkp
			}

			// set nil, to flag for GC
			value_bkp = nil
		}
	}

	return ret, s.values
}

// create new search in different thread
func fork_subtask_async(row, col, d int, s Sudoku) Promise {
	promise := make(Promise, 1)
	s_temp := s
	// create clean copy of values
	s_temp.values = s.values.clone()

	// start new goroutine for searching subtree
	go func(s1 Sudoku) {
		// streaming: <- <-  (push<- <-pop)
		promise <- <-search_parallel(s1, assign(s1, row, col, d))
		// flag for GC
		s1.values = nil
	}(s_temp)

	return promise
}

// Search concurrently and parallel
// FixMe: status is unnecessary here, keep it for signature similarity
func search_parallel(s Sudoku, status bool) Promise {
	// Non blocking channel, must be a single buffered channel
	promise := make(Promise, 1)

	if !status {
		promise <- TaskResult{status: false, count: 0 }
		return promise
	}

	if (s.values.solved()) {
		promise <- TaskResult{status: false, count: 1 }
		return promise
	}

	minI, minJ := find_minimum_option_cell(s)

	// subtasks future result channels array
	subtask := []Promise{}
	ret := TaskResult{status: false, count: 0 }

	for k := 1; k <= s.dim; k++ {
		if s.values[minI][minJ].get(k) == 1 {
			p := fork_subtask_async(minI, minJ, k, s)
			subtask = append(subtask, p)
		}
	}

	// wait for all subtask result, blocking here
	for t := 0; t < len(subtask); t++ {
		ret.count = ret.count + (<-subtask[t]).count
	}

	promise <- ret

	return promise
}

func solve(s Sudoku) (bool, Values) {
	ret, v := false, Values{}

	if !SUDOKU_COUNT_MODE {
		ret, v = search(s, true)
	} else {
		data := <-search_parallel(s, true)
		ret, v = data.count > 0, Values{}
		solution_counter = data.count
	}
	return ret, v
}

func create_sudoku(b Board) Sudoku {
	r := Sudoku{}
	r.n = b.unit_size
	r.dim = b.dim
	r.grid = b
	r.total_solution = 0
	r.unit_list = makeUnitList(b.unit_size)
	r.peer_size, r.peers = makePeers(b.unit_size, r.unit_list)
	r.values = makeValues(r.dim)
	return r
}

// ----------------- Parallel & Concurrent approach --------------------
func atomic_inc() {
	atomic.AddUint64(&solution_counter, 1)
}

func atomic_get() uint64 {
	return atomic.LoadUint64(&solution_counter)
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

var solution_counter uint64

func main() {
	//defer profile.Start(profile.CPUProfile, profile.ProfilePath("."), profile.NoShutdownHook).Stop()
	//defer profile.Start(profile.MemProfile, profile.ProfilePath("."), profile.NoShutdownHook).Stop()

	runtime.GOMAXPROCS(runtime.NumCPU())

	board, err := read_board(os.Stdin)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	//print_board(board)

	s, ret := create_sudoku(board), false

	solution_counter = 0

	if parse_grid(s) {
		// call solver function
		ret, s.values = solve(s)

		if ret && !SUDOKU_COUNT_MODE {
			// print a solved board
			print_solution(s)
		} else {
			// print total number of solution
			fmt.Println(solution_counter)
		}
	} else {
		fmt.Println("Could not load puzzle.")
	}
}
