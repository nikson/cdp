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
	"runtime"
)

// ------------------ Data type -----------

type Board struct {
	data [][]int
	size int
}

type RowChunk struct {
	row_id int   // row_id of current chunk
	value  []int // current data set
	result []int // result data set
	top    []int // above row of row_value
	bottom []int // bottom row of row_value
}

func (rc RowChunk)  update() RowChunk {
	size := len(rc.value)
	rc.result = make([]int, size)

	// Optimization: Use dynamic algo (DP) for counting life rules in a single loop
	// DP algo:
	// neighbour = head_3x3 - tail_3x3 - current[row][col]; head = row + 1, tail = row - 2;
	// head - tail = row + 1 - row + 2 =  3x3 grid sum
	head := rc.top[0] + rc.bottom[0] + rc.value[0]
	temp := []int{head}

	for k, tail := 0, 0; k < size; k++ {
		// update head
		if ( k + 1 < size ) {
			head = (rc.top[k + 1] + rc.bottom[k + 1] + rc.value[k + 1] ) + head
		}
		// add head in temp list
		temp = append(temp, head)

		if (k - 2 >= 0) {
			tail = temp[0]
			temp = temp[1:]
		}

		neighbour := head - tail - rc.value[k]

		rc.result[k] = game_of_life_status(rc.value[k], neighbour);
	}

	return rc
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

func (b Board) neighbour(row int, col int) int {
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

// -------------------- problem solving functions ----------------------

// Pseudo:
// step 1: split data in row-wise (easy case)
// step 2: apply game rules, determine survival, return the result
// setp 3: merge row into final board
// step 4: return final board, if playing continue goto step 1 using final-board

func play(src Board) Board {
	dst := src.clone()

	for r := 0; r < src.size; r++ {
		for c := 0; c < src.size; c++ {
			// count the neighbour
			count := src.neighbour(r, c)
			life := game_of_life_status(src.data[r][c], count)
			// update new board by life value
			dst.data[r][c] = life
		}
	}

	return dst
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

func play_parallel(board Board, in chan <- RowChunk, out <-chan RowChunk) Board {

	// board splitter, split the board row wise
	go split(board, in)

	// merge process data and wait until all row processed
	return merge(board.size, out)
}

func split(b Board, data_in chan <- RowChunk) {

	for i := 0; i < b.size; i++ {
		chunk := RowChunk{}
		// working row id
		chunk.row_id = i
		// working row value
		chunk.value = b.data[i]

		if chunk.row_id > 0 {
			chunk.top = b.data[chunk.row_id - 1]
		} else {
			chunk.top = make([]int, b.size)
		}

		if chunk.row_id + 1 < b.size {
			chunk.bottom = b.data[chunk.row_id + 1]
		} else {
			chunk.bottom = make([]int, b.size)
		}

		// write chunk data in worker channel queue
		data_in <- chunk
	}

}

func merge(total int, data_out <-chan RowChunk) Board {
	dst := NewBoard(total)

	// Don't use range, channel will not close in each timestep
	for i := 0; i < total; i++ {
		// get processed data from worker channel queue
		item, ok := <-data_out
		if ok {
			dst.data[item.row_id] = item.result
		}
	}

	return dst
}

func init_worker_pool(pool_size int, state <-chan bool, out <-chan RowChunk, in chan <- RowChunk) {
	for i := 0; i < pool_size; i++ {
		go func() {
			for ; ; {
				select {
				case signal := <-state:
				// exit from loop
					if signal {
						break
					}
				default:
					data := <-out
					in <- data.update()
				}
			}
		}()
	}
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

func print_rowchunk(rc RowChunk) {
	//rc := RowChunk{}
	//rc.top = []int{0, 1, 1, 1}
	//rc.bottom = []int{1, 1, 0, 0}
	//rc.value = []int{1, 0, 1, 0}
	//rr := rc.update()
	//fmt.Println(rc)

	fmt.Println("Row: ", rc.row_id)
	fmt.Println("T: ", rc.top)
	fmt.Println("V: ", rc.value)
	fmt.Println("B: ", rc.bottom)
	fmt.Println("R: ", rc.result)
	fmt.Println("______________________________")
}

// -------------- input/output ----------------

func main() {
	// Flag to use mulitcore
	runtime.GOMAXPROCS(runtime.NumCPU())
	cpu := runtime.NumCPU();

	board, step, err := read_input(os.Stdin)

	//if err == nil {
	//	fmt.Println("Inital board")
	//	print_board(board)
	//}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	wstate := make(chan bool)
	data_in := make(chan RowChunk)
	data_out := make(chan RowChunk)

	go init_worker_pool(cpu, wstate, data_in, data_out);

	for i := 0; i < step; i++ {
		// clone board inside the func and return calculated fresh copy
		//next := play(board)
		// concurrent and parrallel processing
		next := play_parallel(board, data_in, data_out)
		// reference new board
		board = next
	}

	// signal to shutdown worker pool
	go func() {
		for i := 0; i < cpu; i++ {
			wstate <- true
		}
		close(data_in)
		close(data_out)
		close(wstate)
	}()

	// print final board
	print_board(board)
}
