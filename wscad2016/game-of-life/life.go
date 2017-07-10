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
	// optimization: work with byte array
	data [][]byte
	size int
}

func NewBoard(sz int) Board {
	// Be careful: slice always reference original underlaying array
	board := Board{data: make([][]byte, sz), size: sz}
	for i := 0; i < sz; i++ {
		board.data[i] = make([]byte, sz)
	}
	return board
}

type RowChunk struct {
	row_id int    // row_id of current chunk
	value  []byte // current data set
	result []byte // result data set
	top    []byte // above row of row_value
	bottom []byte // bottom row of row_value
}

func NewRowChunk(id, size int) RowChunk {
	rc := RowChunk{}
	rc.row_id = id;
	rc.value = make([]byte, size)
	rc.result = make([]byte, size)
	rc.top = make([]byte, size)
	rc.bottom = make([]byte, size)
	return rc
}

func (rc RowChunk)  play() RowChunk {
	size := len(rc.value)

	// Optimization: Use dynamic algo for counting life rules in a single loop
	// algo:
	// neighbour = head_3x3 - tail_3x3 - current[row][col]; head = row + 1, tail = row - 2;
	// head - tail = row + 1 - row + 2 =  3x3 grid sum
	head := 0

	if size > 0 {
		head = count_X(rc.top[0], rc.bottom[0], rc.value[0])
	}

	// temporary hold the last 3 head value
	queue := [3]int{ 0, 0, head }

	for k, tail := 0, 0; k < size; k++ {
		// calculate next head
		if ( k + 1 < size ) {
			head = count_X(rc.top[k + 1], rc.bottom[k + 1], rc.value[k + 1]) + head
		}
		
		// Optimization: using array instaed of slice, slice resizing consuming 35% of exeuction time 
		// pop head 
		tail = queue[0]
		// push head 
		queue[0] = queue[1]
		queue[1] = queue[2]
		queue[2] = head 
				
		current := 0
		if rc.value[k] == 120 {
			current = 1
		}

		neighbour := head - tail - current

		rc.result[k] = set_life_status(game_of_life_status(current, neighbour))
	}

	return rc
}

// -------------------- problem solving functions ----------------------

// Pseudo:
// step 1: split data in row-wise (easy case)
// step 2: apply game rules, determine survival, return the result
// setp 3: merge row into final board
// step 4: return final board, if playing continue goto step 1 using final-board

func count_X(x, y, z byte) int {
	count := 0;
	// 'x' = 120
	if ( x == 120) {
		count = count + 1
	}
	if ( y == 120) {
		count = count + 1
	}
	if ( z == 120) {
		count = count + 1
	}
	return count
}

func set_life_status(state int) byte {
	life := byte(' ')
	if state == 1 {
		life = byte('x')
	}

	return life
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
		// working row id
		chunk := NewRowChunk(i, b.size)

		// working row value
		copy(chunk.value, b.data[i])

		// default top, bottom row is empty array
		if chunk.row_id > 0 {
			copy(chunk.top, b.data[chunk.row_id - 1])
		}

		if chunk.row_id + 1 < b.size {
			copy(chunk.bottom, b.data[chunk.row_id + 1])
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
			copy(dst.data[item.row_id], item.result)
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
					data, ok := <-out
					if ok {
						in <- data.play()
					}
				}
			}
		}()
	}
}

func shutdown_workers(pool_size int, state chan bool, out chan RowChunk, in chan RowChunk){
	for i := 0; i < pool_size; i++ {
		state <- true
	}
	close(out)
	close(in)
	close(state)
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
			board.data[i] = []byte(line)
		}
	}

	return board, step, nil
}

func print_board(b Board) {
	for i := 0; i < b.size; i++ {
		fmt.Println(string(b.data[i]))
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
	worker_pool_size := cpu*8

	go init_worker_pool(worker_pool_size, wstate, data_in, data_out)

	for i := 0; i < step; i++ {
		// clone board inside the func and return calculated fresh copy
		//next := play(board)
		// concurrent and parrallel processing
		next := play_parallel(board, data_in, data_out)
		// reference new board
		board = next
	}

	// signal to shutdown worker pool
	go shutdown_workers(worker_pool_size, wstate, data_in, data_out)
	
	// print final board
	print_board(board)
}
