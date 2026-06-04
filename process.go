package main

import (
	"compress/gzip"
	"encoding/binary"
	"encoding/json"
	"log"
	"os"
	"rinha-de-backend-retr0lbb/utils"
)

// NumBuckets deve coincidir com o valor em src/kdtree.go
const NumBuckets = 16

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

	var buckets [NumBuckets][]Row

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

		// Bucket baseado no valor da transação quantizado (vector[0] ∈ [0, 254])
		// Shift de 4 bits → 16 buckets, cada um cobrindo ~6.25% do range de amount.
		// Muito mais balanceado do que flags binárias (IsOnline/CardPresent/!isKnown).
		bucketID := int(r.Vector[0]) >> 4
		buckets[bucketID] = append(buckets[bucketID], r)
	}

	// Log de distribuição dos buckets para diagnóstico
	var total int
	for i := 0; i < NumBuckets; i++ {
		total += len(buckets[i])
	}
	log.Printf("Total de vetores processados: %d", total)
	log.Println("=== Distribuição dos Buckets (process) ===")
	for i := 0; i < NumBuckets; i++ {
		pct := 0.0
		if total > 0 {
			pct = float64(len(buckets[i])) / float64(total) * 100
		}
		// Faixa de amount: bucket i cobre amount quantizado [i*16, i*16+15]
		// Equivalente a amount original [i*MaxAmount/16, (i+1)*MaxAmount/16)
		log.Printf("  Bucket %2d (amount_q %3d–%3d): %7d vetores (%.1f%%)",
			i, i*16, i*16+15, len(buckets[i]), pct)
	}
	log.Println("==========================================")

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
	for i := 0; i < NumBuckets; i++ {
		count := uint32(len(buckets[i]))

		// Header: 4 bytes offset, 4 bytes count
		if err := binary.Write(bucketFiles, binary.LittleEndian, offset); err != nil {
			log.Fatal(err)
		}
		if err := binary.Write(bucketFiles, binary.LittleEndian, count); err != nil {
			log.Fatal(err)
		}

		for _, r := range buckets[i] {
			vecFiles.Write(r.Vector[:])
			lbFiles.Write([]byte{r.Label})
		}
		offset += count
	}

	log.Println("done — binários gerados com sucesso")
}
