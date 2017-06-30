package main

import (
	"fmt"
	"os"
	"bufio"
	"io"
	"strings"
	"errors"
	"sync"
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

func count(img *PPMImage, hist []float32, key, r, g, b int, wg *sync.WaitGroup) {
	defer wg.Done()

	count := 0
	for i, k := 0, img.size; i < k; i++ {
		// forward read
		if (img.data[i].r == r && img.data[i].g == g && img.data[i].b == b) {
			count++;
		}
		// backward read
		k--
		if ( (i < k) && (img.data[k].r == r && img.data[k].g == g && img.data[k].b == b)) {
			count++;
		}

	}

	h := float32(count) / float32(img.size)
	//return HVPair{index: key, value:h }
	// hist is slice, and pass-by-reference
	hist[key] = h
}

func Histogram(img *PPMImage, wg *sync.WaitGroup) []float32 {
	hist := make([]float32, 64)
	wg.Add(64)

	for index, r := 0, 0; r <= 3; r++ {
		for g := 0; g <= 3; g++ {
			for b := 0; b <= 3; b++ {
				// start goroutine for each count
				go count(img, hist, index, r, g, b, wg)

				index++
			}
		}
	}

	return hist
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

func read_input_parallel(rd io.Reader, cpu int, wg *sync.WaitGroup) (PPMImage, error) {
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

	// read pixel data parallel
	number_of_chunk := cpu * 16                // create more chunk than cpus to utilize OS context switching
	block_size := img.size / number_of_chunk
	// handle small case where block_size is 0 when img size is less than number_of_chunk
	if block_size == 0 {
		block_size = img.size
	}

	for i, total := 0, 0; total < img.size; i++ {
		chunk := make([]byte, block_size * 3)        // 3 = rgb = 1 pixel
		//size, err := reader.Read(chunk)	// Issue, can be: n < len(chunk)
		size, err := io.ReadFull(reader, chunk)

		if err != nil {
			return img, errors.New("Pixel reading failed\n");
		}

		wg.Add(1)
		go process_pixels(&img, (block_size * i), size, chunk, wg)
		total = total + (size / 3)
	}

	return img, nil
}

func process_pixels(img *PPMImage, start_index, size int, chunkdata []byte, wg *sync.WaitGroup) {
	defer wg.Done()
	for i, pos := 0, start_index; i < size; i = i + 3 {
		// read 3 pixel
		r, g, b := chunkdata[i], chunkdata[i + 1], chunkdata[i + 2]

		// set pixel with pre-processing
		img.data[pos] = Pixel{
			r: ((int(r) * 4) / 256),
			g: ((int(g) * 4) / 256),
			b: ((int(b) * 4) / 256) }
		pos++
	}

	chunkdata = nil
}

// --------------------------------------------------------

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	wg := sync.WaitGroup{}

	data, err := read_input_parallel(os.Stdin, runtime.NumCPU(), &wg)
	wg.Wait()

	if err != nil {
		fmt.Printf("%v", err)
		os.Exit(1)
	}
	//fmt.Println(data)

	result := Histogram(&data, &wg)

	wg.Wait()

	// print output
	for _, h := range result {
		fmt.Printf("%0.3f ", h);
	}
	fmt.Println()
}
