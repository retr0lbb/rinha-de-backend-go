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
func search(vector [14]float32) (float32, bool) {
	var topVectors [5]float32
	var topLabels [5]uint8

	for i := range topVectors {
		topVectors[i] = math.MaxFloat32
		topLabels[i] = 255
	}

	if TotalVectors == 0 {
		return 0.0, false
	}

	for i := 0; i < int(TotalVectors); i++ {
		offset := i * 14
		var dist float32

		for j := 0; j < 14; j++ {

			dbValue := VectorDataset[j+offset]

			convertedBackValueFromVector := float32(dbValue) / 254.0

			if convertedBackValueFromVector == 255 {
				continue
			}

			diff := vector[j] - convertedBackValueFromVector

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
