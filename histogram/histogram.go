package main

import (
	"fmt"
	"os"
	"bufio"
	"io"
	"strings"
	"errors"
	"runtime"
)

// ----------------- data type ------------------

type Pixel struct {
	r int
	g int
	b int
}

type PPMImage struct {
	header         string
	rgb_comp_color int
	width          int
	height         int
	data           []Pixel
	size           int
}

// index, value pair of histogram
type HVPair struct {
	index int
	value float32
}

// ----------------- data type ----------------------------

// ----------------- problem solving function -------------
func count(img PPMImage, key, r, g, b int, out chan <- HVPair) {
	count := 0
	for i := 0; i < img.size; i++ {
		if (img.data[i].r == r && img.data[i].g == g && img.data[i].b == b) {
			count++;
		}
	}

	h := float32(count) / float32(img.size)
	ret := HVPair{index: key, value:h }
	// write to output channel
	out <- ret
}

func Histogram(img PPMImage, out chan <- HVPair) {
	for index, r := 0, 0; r <= 3; r++ {
		for g := 0; g <= 3; g++ {
			for b := 0; b <= 3; b++ {
				// start goroutine for each count
				go count(img, index, r, g, b, out)
				index++
			}
		}
	}
}

// --------------------------------------------------------

func read_input(rd io.Reader) (PPMImage, error) {
	reader := bufio.NewReader(rd)

	img := PPMImage{rgb_comp_color: 255, header: "P6" }

	line, err := reader.ReadString('\n')

	if strings.Trim(line, " \r\n") != img.header {
		return img, errors.New("Invalid image format (must be 'P6')\n")
	}

	for true {
		line, _ := reader.ReadString('\n')

		if strings.HasPrefix(line, "#") {
			continue
		} else {
			_, err := fmt.Sscanf(line, "%d %d", &img.width, &img.height)

			if err != nil {
				return img, errors.New("Invalid image size (error loading)\n")
			}

			break
		}
	}

	line, err = reader.ReadString('\n')
	if err != nil {
		var rgb_comp int
		_, err := fmt.Sscanf(line, "%d", &rgb_comp)

		if err != nil {
			return img, errors.New("Invalid rgb component (error loading)\n")
		}

		if ( rgb_comp != img.rgb_comp_color) {
			return img, errors.New("Image does not have 8-bits components\n");
		}
	}

	// read pixel data
	img.size = img.width * img.height
	img.data = make([]Pixel, img.size)
	for i := 0; i < img.size; i++ {
		// read 3 pixel
		r, _ := reader.ReadByte()
		g, _ := reader.ReadByte()
		b, _ := reader.ReadByte()

		//fmt.Println(r, g, b)
		// set pixel with pre-processing
		img.data[i] = Pixel{
			r: ((int(r) * 4) / 256),
			g: ((int(g) * 4) / 256),
			b: ((int(b) * 4) / 256) }
	}

	return img, nil
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	data, err := read_input(os.Stdin)

	if err != nil {
		fmt.Printf("%v", err)
		os.Exit(1)
	}
	//fmt.Println(data)

	hist := make([]float32, 64)
	result_chan := make(chan HVPair, 64)

	go Histogram(data, result_chan)

	// wait & receive all output hv pair
	received := 0
	for hv := range result_chan {
		hist[hv.index] = hv.value
		received++
		if ( received == 64 ) {
			close(result_chan)
			break
		}
	}

	// print output
	for _, h := range hist {
		fmt.Printf("%0.3f ", h);
	}
	fmt.Println()
}
