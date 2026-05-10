package main

import (
	"compress/gzip"
	"encoding/json"
	"log"
	"math"
	"os"
)

var (
	VectorDataset []uint8
	Labels        []uint8
)

const (
	VectorSize    = 14
	ExpectedRows  = 3000000
	QuantizeScale = 255.0
)

func openLargeFile(path string) {
	f, err := os.Open(path)

	if err != nil {
		log.Fatal(err)
	}

	defer f.Close()

	gzReader, err := gzip.NewReader(f)

	if err != nil {
		log.Fatal(err)
	}

	defer gzReader.Close()

	decoder := json.NewDecoder(gzReader)

	if _, err := decoder.Token(); err != nil {
		log.Fatal(err)
	}

	VectorDataset = make([]uint8, 0, ExpectedRows*VectorSize)
	Labels = make([]uint8, 0, ExpectedRows)

	for decoder.More() {
		var row struct {
			Vector [14]float32 `json:"vector"`
			Label  string      `json:"label"`
		}

		if err := decoder.Decode(&row); err != nil {
			continue
		}

		for _, v := range row.Vector {
			if v < 0 {
				v = 0
			}

			if v > 1 {
				v = 1
			}

			VectorDataset = append(VectorDataset, uint8(v*QuantizeScale))
		}

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

	for i := 0; i < totalRegistros; i++ {
		offset := i * 14
		var dist float32

		for j := 0; j < 14; j++ {

			dbValue := float32(
				VectorDataset[offset+j],
			) / QuantizeScale

			diff := vector[j] - dbValue

			dist += diff * diff
		}

		if dist < topVectors[4] {
			insertDislocated(&topVectors, &topLabels, dist, Labels[i])
		}
	}

	var fraudCount uint8
	for _, l := range topLabels {
		fraudCount += l
	}

	precision := float32(fraudCount) / 5.0

	return precision, fraudCount >= 3
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
