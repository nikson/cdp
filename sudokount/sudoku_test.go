package main

import (
	//"fmt"
	"testing"
	//"strconv"
)

func Test_unit_list(t *testing.T) {
	unit_list := makeUnitList(3)

	t.Log("unit_list size: ", len(unit_list))
	t.Log("unit_list[r] size: ", len(unit_list[0]))
	t.Log("unit_list size[r][c]: ", len(unit_list[0][0]))
	t.Log("unit_list size[r][c][u=0]: ", len(unit_list[0][0][0]))
	t.Log("unit_list size[r][c][u=1]: ", len(unit_list[0][0][1]))
	t.Log("unit_list size[r][c][u=2]: ", len(unit_list[0][0][2]))
}

func Test_peers(t *testing.T) {
	unit_list := makeUnitList(3)
	size, peers := makePeers(3, unit_list)

	t.Log("peers size: ", size)
	t.Log("peers[r] size: ", len(peers))
}

func Test_values(t *testing.T) {
	values := makeValues(9)

	t.Log("value[r] size: ", len(values))
	t.Log("values[r][c].v value: ", values[0][0].bits)
}

func Test_set(t *testing.T) {
	cv := CellValue{bits: 3 }
	t.Logf("bits: %b", cv.bits)        // 11 = {2, 1}

	cv.set(5)
	t.Logf("bits: %b", cv.bits)        // 10011 = {5, 2, 1}
	//t.Log("bit << 0 :", strconv.FormatInt(bit, 2))
}

func Test_unset(t *testing.T) {
	cv := CellValue{bits: 3 }

	cv.set(5)
	t.Logf("bits: %b", cv.bits)        // 10011 = {5, 2, 1}
	cv.unset(2)
	t.Logf("bits: %b", cv.bits)        // 10011 = {5, 1}
}

func Test_get(t *testing.T) {
	cv := CellValue{bits: 3 }

	t.Logf("bits: %b", cv.bits)
	t.Log("1st bit", cv.get(1))                // 1
	t.Log("2nd bit", cv.get(2))                // 1
	t.Log("3rd bit", cv.get(3))                // 0
	t.Log("5rd bit", cv.get(5))                // 0
	t.Log("1st bit", cv.get(1))                // 1
}

func Test_count(t *testing.T) {
	cv := CellValue{bits: 5 }                // 101
	t.Log("count: ", cv.count(), cv.count() == 2)
}
func Test_digit_get(t *testing.T) {
	cv := CellValue{bits: 4 }        // 4 = 100 => digit 3
	t.Log("digit: ", cv.digit_get())
	cv.unset(3)
	cv.set(5)
	t.Log("digit: ", cv.digit_get())
}
