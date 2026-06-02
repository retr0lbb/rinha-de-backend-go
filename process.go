package main

import (
	"compress/gzip"
	"encoding/binary"
	"encoding/json"
	"log"
	"os"
	"rinha-de-backend-retr0lbb/utils"
)

type Row struct {
	Vector [14]byte
	Label  byte
}

func main() {
	in, err := os.Open("files/references.json.gz")
	if err != nil {
		log.Fatal(err)
	}
	defer in.Close()

	gz, err := gzip.NewReader(in)
	if err != nil {
		log.Fatal("GZIP Failed", err)
	}
	defer gz.Close()

	decoder := json.NewDecoder(gz)
	_, err = decoder.Token()
	if err != nil {
		log.Fatal(err)
	}

	var buckets [8][]Row

	for decoder.More() {
		var row struct {
			Vector [14]float32 `json:"vector"`
			Label  string      `json:"label"`
		}

		if err := decoder.Decode(&row); err != nil {
			continue
		}

		var r Row
		for i, v := range row.Vector {
			r.Vector[i] = utils.Quantize(v)
		}

		if row.Label == "fraud" {
			r.Label = 1
		} else {
			r.Label = 0
		}

		// Calculate bucket ID
		bucketID := 0
		if r.Vector[9] != 0 {
			bucketID |= 4
		}
		if r.Vector[10] != 0 {
			bucketID |= 2
		}
		if r.Vector[11] != 0 {
			bucketID |= 1
		}

		buckets[bucketID] = append(buckets[bucketID], r)
	}

	vecFiles, err := os.Create("files/vectors.bin")
	if err != nil {
		log.Fatal(err)
	}
	defer vecFiles.Close()

	lbFiles, err := os.Create("files/labels.bin")
	if err != nil {
		log.Fatal(err)
	}
	defer lbFiles.Close()

	bucketFiles, err := os.Create("files/buckets.bin")
	if err != nil {
		log.Fatal(err)
	}
	defer bucketFiles.Close()

	var offset uint32 = 0
	for i := 0; i < 8; i++ {
		count := uint32(len(buckets[i]))
		
		// Write header: 4 bytes offset, 4 bytes count
		err := binary.Write(bucketFiles, binary.LittleEndian, offset)
		if err != nil {
			log.Fatal(err)
		}
		err = binary.Write(bucketFiles, binary.LittleEndian, count)
		if err != nil {
			log.Fatal(err)
		}

		for _, r := range buckets[i] {
			vecFiles.Write(r.Vector[:])
			lbFiles.Write([]byte{r.Label})
		}
		offset += count
	}

	log.Println("done")
}
