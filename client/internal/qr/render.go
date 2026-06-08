package qr

import (
	"fmt"
	"strings"

	qrcode "github.com/skip2/go-qrcode"
)

// Image is a machine-friendly QR bitmap the microcontroller can paint on a display.
type Image struct {
	Size    int   `json:"size"`
	Modules [][]int `json:"modules"`
	ASCII   string  `json:"ascii,omitempty"`
}

// Render encodes content into a square module matrix (1 = dark, 0 = light).
func Render(content string) (*Image, error) {
	if content == "" {
		return nil, fmt.Errorf("qr content is empty")
	}

	code, err := qrcode.New(content, qrcode.Medium)
	if err != nil {
		return nil, fmt.Errorf("encode qr: %w", err)
	}

	bitmap := code.Bitmap()
	size := len(bitmap)
	modules := make([][]int, size)
	var ascii strings.Builder

	for y := 0; y < size; y++ {
		row := make([]int, size)
		for x := 0; x < size; x++ {
			if bitmap[y][x] {
				row[x] = 1
				ascii.WriteByte('#')
			} else {
				row[x] = 0
				ascii.WriteByte(' ')
			}
		}
		modules[y] = row
		if y < size-1 {
			ascii.WriteByte('\n')
		}
	}

	return &Image{
		Size:    size,
		Modules: modules,
		ASCII:   ascii.String(),
	}, nil
}
