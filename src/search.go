package main

import (
	"log"
	"math"
	"os"
)

var (
	VectorDataset []uint8
	Labels        []uint8
	TotalVectors  uint32
)

const VectorSize = 14

func openLargeFile(vectorFilePath string, labelFilePath string) {
	var err error

	VectorDataset, err = os.ReadFile(vectorFilePath)

	if err != nil {
		log.Fatal(err)
	}

	Labels, err = os.ReadFile(labelFilePath)

	if err != nil {
		log.Fatal(err)
	}

	TotalVectors = uint32(len(VectorDataset) / VectorSize)

	log.Printf(
		"dataset carregado: %d vetores",
		TotalVectors,
	)

	log.Printf(
		"memoria vetores: %.2f MB",
		float64(len(VectorDataset))/(1024*1024),
	)

	log.Printf(
		"memoria labels: %.2f MB",
		float64(len(Labels))/(1024*1024),
	)
}

// i will implement a simple KNN search i changed from float32 to uint8 so it can be more memory efficient
// usando o mesmo payload tem que dar 0.4 true
func search(vector [14]uint8) (float32, bool) {
	var topVectors [5]uint32
	var topLabels [5]uint8

	for i := range topVectors {
		topVectors[i] = math.MaxUint32
		topLabels[i] = 255
	}

	if TotalVectors == 0 {
		return 0.0, false
	}

	for i := 0; i < int(TotalVectors); i++ {
		offset := i * 14
		var dist uint32

		for j := range 14 {

			dbValue := VectorDataset[j+offset]

			if vector[j] == 255 || dbValue == 255 {
				dist += 255 * 255
				continue
			}

			diff := int16(vector[j]) - int16(dbValue)

			dist += uint32(diff * diff)
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

	return precision, precision < 0.6
}

func insertDislocated(topVectors *[5]uint32, labels *[5]uint8, d uint32, l uint8) {
	for i := range 5 {
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
