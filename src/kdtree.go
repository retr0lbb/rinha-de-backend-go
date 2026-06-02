package main

import (
	"encoding/binary"
	"log"
	"math"
	"os"
	"sort"
)

type BucketInfo struct {
	Offset uint32
	Count  uint32
}

type KDNode struct {
	VectorIdx uint32
	Left      uint32
	Right     uint32
	SplitDim  uint8
	SplitVal  uint8
	_         uint16
}

var Buckets [8]BucketInfo
var KDTreeNodes []KDNode
var KDTrees [8]uint32

func loadBuckets(bucketFilePath string) error {
	file, err := os.Open(bucketFilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	for i := 0; i < 8; i++ {
		err = binary.Read(file, binary.LittleEndian, &Buckets[i].Offset)
		if err != nil {
			return err
		}
		err = binary.Read(file, binary.LittleEndian, &Buckets[i].Count)
		if err != nil {
			return err
		}
	}

	log.Println("Buckets loaded successfully")
	return nil
}

func BuildKDTrees() {
	KDTreeNodes = make([]KDNode, 0, TotalVectors)
	for i := 0; i < 8; i++ {
		count := Buckets[i].Count
		offset := Buckets[i].Offset
		if count == 0 {
			KDTrees[i] = math.MaxUint32
			continue
		}

		indices := make([]uint32, count)
		for j := uint32(0); j < count; j++ {
			indices[j] = offset + j
		}

		KDTrees[i] = buildFlatTree(indices, 0)
	}
	log.Println("KD-Trees built successfully")
}

func buildFlatTree(indices []uint32, depth int) uint32 {
	if len(indices) == 0 {
		return math.MaxUint32
	}

	dim := uint8(depth % 14)

	sort.Slice(indices, func(i, j int) bool {
		idx1 := indices[i] * VectorSize
		idx2 := indices[j] * VectorSize
		return VectorDataset[idx1+uint32(dim)] < VectorDataset[idx2+uint32(dim)]
	})

	mid := len(indices) / 2
	midIdx := indices[mid]
	midVal := VectorDataset[midIdx*VectorSize+uint32(dim)]

	nodeIdx := uint32(len(KDTreeNodes))
	KDTreeNodes = append(KDTreeNodes, KDNode{
		VectorIdx: midIdx,
		Left:      math.MaxUint32,
		Right:     math.MaxUint32,
		SplitDim:  dim,
		SplitVal:  midVal,
	})

	leftIdx := buildFlatTree(indices[:mid], depth+1)
	rightIdx := buildFlatTree(indices[mid+1:], depth+1)

	KDTreeNodes[nodeIdx].Left = leftIdx
	KDTreeNodes[nodeIdx].Right = rightIdx

	return nodeIdx
}

func SearchKDTree(nodeIdx uint32, query [14]uint8, topVectors *[5]uint32, topLabels *[5]uint8) {
	if nodeIdx == math.MaxUint32 {
		return
	}

	node := &KDTreeNodes[nodeIdx]
	offset := node.VectorIdx * VectorSize
	worstDist := topVectors[4]
	var dist uint32

	// Dimensions 0-4 (no sentinels)
	for j := range uint32(5) {
		diff := int16(query[j]) - int16(VectorDataset[offset+j])
		dist += uint32(diff * diff)
	}

	if dist < worstDist {
		// Dimensions 5-6 (sentinel 255)
		for j := uint32(5); j < 7; j++ {
			if query[j] == 255 || VectorDataset[offset+j] == 255 {
				dist += uint32(255 * 255)
			} else {
				diff := int16(query[j]) - int16(VectorDataset[offset+j])
				dist += uint32(diff * diff)
			}
		}

		if dist < worstDist {
			// Dimensions 7-13 (no sentinels)
			for j := uint32(7); j < 14; j++ {
				diff := int16(query[j]) - int16(VectorDataset[offset+j])
				dist += uint32(diff * diff)
			}

			if dist < worstDist {
				insertDislocated(topVectors, topLabels, dist, Labels[node.VectorIdx])
				worstDist = topVectors[4]
			}
		}
	}

	diff := int16(query[node.SplitDim]) - int16(node.SplitVal)

	var first, second uint32
	if diff < 0 {
		first = node.Left
		second = node.Right
	} else {
		first = node.Right
		second = node.Left
	}

	SearchKDTree(first, query, topVectors, topLabels)

	var axisDist uint32
	if (node.SplitDim == 5 || node.SplitDim == 6) && (query[node.SplitDim] == 255 || node.SplitVal == 255) {
		axisDist = uint32(255 * 255)
	} else {
		axisDist = uint32(diff * diff)
	}

	if axisDist < topVectors[4] {
		SearchKDTree(second, query, topVectors, topLabels)
	}
}
