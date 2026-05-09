package main

import (
	"compress/gzip"
	"encoding/json"
	"log"
	"math"
	"os"
)

var (
	VectorDataset []float32
	Labels        []uint8
)

const VectorSize = 14

func openLargeFile(path string) {
	f, _ := os.Open(path)

	gzReader, err := gzip.NewReader(f)

	if err != nil {
		log.Fatal(err)
	}

	defer gzReader.Close()

	decoder := json.NewDecoder(gzReader)

	if _, err := decoder.Token(); err != nil {
		log.Fatal(err)
	}

	VectorDataset = make([]float32, 0, 3000000*VectorSize)
	Labels = make([]uint8, 0, 3000000)

	for decoder.More() {
		var row struct {
			Vector []float32 `json:"vector"`
			Label  string    `json:"label"`
		}

		if err := decoder.Decode(&row); err != nil {
			continue
		}

		VectorDataset = append(VectorDataset, row.Vector...)
		if row.Label == "fraud" {
			Labels = append(Labels, 1)
		} else {
			Labels = append(Labels, 0)
		}

	}

}

// i will implement a simple KNN search
func search(vector [14]float32) (float32, bool) {
	var topVectors [5]float32
	var topLabels [5]uint8

	for i := range topVectors {
		topVectors[i] = math.MaxFloat32
		topLabels[i] = 255
	}

	totalRegistros := len(Labels)

	if totalRegistros == 0 {
		return 0.0, false
	}

	for i := range totalRegistros {
		offset := i * VectorSize
		var dist float32

		for j := range 14 {
			diff := vector[j] - VectorDataset[offset+j]

			dist += diff * diff
		}

		if dist < topVectors[4] {
			insertDislocated(&topVectors, &topLabels, dist, Labels[i])
		}
	}

	var count uint8
	for _, l := range topLabels {
		count += l
	}

	precision := float32(count) / 5.0

	return precision, count >= 3
}

func insertDislocated(topVectors *[5]float32, labels *[5]uint8, d float32, l uint8) {
	for i := 0; i < 5; i++ {
		if d < topVectors[i] {
			for j := 4; j > i; j-- {
				topVectors[j] = topVectors[j-1]
				labels[j] = labels[j-1]
			}

			topVectors[i] = d
			labels[i] = l
			return
		}
	}
}
