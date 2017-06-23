package main

import (
	//"fmt"
	"testing"
	//"strconv"
)

func Test_unit_list(t *testing.T) {
	unit_list := initUnitList(9)

	t.Log("unit_list size: ", len(unit_list))
	t.Log("unit_list[r] size: ", len(unit_list[0]))
	t.Log("unit_list size[r][c]: ", len(unit_list[0][0]))
	t.Log("unit_list size[r][c][u=0]: ", len(unit_list[0][0][0]))
	t.Log("unit_list size[r][c][u=1]: ", len(unit_list[0][0][1]))
	t.Log("unit_list size[r][c][u=2]: ", len(unit_list[0][0][2]))
}

func Test_peers(t *testing.T) {
	peers := initPeers(9, 20)

	t.Log("peers[r] size: ", len(peers))
	t.Log("peers[r][c] size: ", len(peers[0][0]))
}

func Test_values(t *testing.T) {
	values := makeValues(9)

	t.Log("value[r] size: ", len(values))
	t.Log("values[r][c].v size: ", len(values[0][0].v))
	t.Log("values[r][c].v value: ", values[0][0].v)
}

func Test_bit_set(t *testing.T) {
	bit := 0
	bit = bit | ( 1 << 0 )
	t.Logf("bit << 0: %b", bit)

	bit = bit | ( 1 << 1 )
	t.Logf("bit << 1: %b", bit)

	//t.Log("bit << 0 :", strconv.FormatInt(bit, 2))
}

// func Test_bit_get(t *testing.T) {}

// func Test_bit_digit(t *testing.T) {}

//func main() {
//	fmt.Println("hello world")
//}
