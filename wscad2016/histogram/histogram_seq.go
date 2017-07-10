package main

import (
	"fmt"
	"os"
	"bufio"
	"io"
	"strings"
	"errors"
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

// ----------------- data type ------------------

func Histogram(img PPMImage) []float32 {
	hist := make([]float32, 64)

	index, count := 0, 0

	for j := 0; j <= 3; j++ {
		for k := 0; k <= 3; k++ {
			for l := 0; l <= 3; l++ {
				for i := 0; i < img.size; i++ {
					if (img.data[i].r == j && img.data[i].g == k && img.data[i].b == l) {
						count++;
					}
				}
				hist[index] = float32(count) / float32(img.size)
				index++
				count = 0
			}
		}
	}

	return hist
}

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
	data, err := read_input(os.Stdin)

	if err != nil {
		fmt.Printf("%v", err)
		os.Exit(1)
	}
	//fmt.Println(data)

	result := Histogram(data)

	// print output
	for _, h := range result {
		fmt.Printf("%0.3f ", h);
	}
	fmt.Println()
}
