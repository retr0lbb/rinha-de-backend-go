package main

import (
	"compress/gzip"
	"encoding/json"
	"log"
	"os"
)

// for processing the files
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

	vecFiles, err := os.Create("files/vectors.bin")
	lbFiles, _ := os.Create("files/labels.bin")

	if err != nil {
		log.Fatal(err)
	}
	defer vecFiles.Close()
	defer lbFiles.Close()

	for decoder.More() {
		var row struct {
			Vector [14]float32 `json:"vector"`
			Label  string      `json:"label"`
		}

		if err := decoder.Decode(&row); err != nil {
			continue
		}

		var quantized [14]byte

		for i, v := range row.Vector {
			quantized[i] = quantize(v)
		}

		vecFiles.Write(quantized[:])

		if row.Label == "fraud" {
			lbFiles.Write([]byte{1})
		} else {
			lbFiles.Write([]byte{0})
		}
	}

	log.Println("done")
}
