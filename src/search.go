package main

import (
	"log"
	"math"
	"os"
	"sync/atomic"

	mmap "github.com/edsrzf/mmap-go"
)

var (
	VectorDataset mmap.MMap
	Labels        mmap.MMap
	TotalVectors  uint32
)

// Contadores atômicos para diagnóstico sem custo em produção
var (
	totalSearches  atomic.Int64
	totalVisited   atomic.Int64
	cappedSearches atomic.Int64 // buscas que atingiram MaxKDVisited
)

const VectorSize = 14

func openLargeFile(vectorFilePath string, labelFilePath string) error {
	var err error

	vectorfile, err := os.Open(vectorFilePath)
	if err != nil {
		return err
	}

	VectorDataset, err = mmap.Map(vectorfile, mmap.RDONLY, 0)
	if err != nil {
		return err
	}

	labelFile, err := os.Open(labelFilePath)
	if err != nil {
		return err
	}

	Labels, err = mmap.Map(labelFile, mmap.RDONLY, 0)
	if err != nil {
		return err
	}

	TotalVectors = uint32(len(VectorDataset) / VectorSize)

	log.Printf("dataset carregado: %d vetores", TotalVectors)
	log.Printf("memoria vetores: %.2f MB", float64(len(VectorDataset))/(1024*1024))
	log.Printf("memoria labels:  %.2f MB", float64(len(Labels))/(1024*1024))

	// Pre-warming: força a leitura de 1 byte de cada página OS (4KB) para carregar na RAM
	log.Println("Iniciando pre-warming do dataset...")
	tempSum := uint64(0)
	for i := 0; i < len(VectorDataset); i += 4096 {
		tempSum += uint64(VectorDataset[i])
	}
	for i := 0; i < len(Labels); i += 4096 {
		tempSum += uint64(Labels[i])
	}
	log.Printf("Pre-warming concluído (checksum de páginas: %d)", tempSum)

	return nil
}

func releaseLargeFile() {
	if VectorDataset != nil {
		_ = VectorDataset.Unmap()
		VectorDataset = nil
	}
	if Labels != nil {
		_ = Labels.Unmap()
		Labels = nil
	}
	log.Println("Arquivos binários desmapeados e memória liberada com sucesso")
}

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

	// Bucket baseado no mapeamento adaptativo pré-calculado.
	bucketID := int(BucketMap[vector[0]])

	var minBaseDist uint32
	if vector[5] == 255 {
		minBaseDist += uint32(255 * 255)
	}
	if vector[6] == 255 {
		minBaseDist += uint32(255 * 255)
	}

	visited := 0
	SearchKDTree(KDTrees[bucketID], vector, &topVectors, &topLabels, minBaseDist, &visited)

	// Atualiza contadores atômicos de diagnóstico
	totalSearches.Add(1)
	totalVisited.Add(int64(visited))
	if visited >= MaxKDVisited {
		cappedSearches.Add(1)
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
