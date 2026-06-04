package main

import (
	"compress/gzip"
	"encoding/binary"
	"encoding/json"
	"log"
	"os"
	"rinha-de-backend-retr0lbb/utils"
)

// NumBuckets deve coincidir com o valor em src/kdtree.go.
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

	// Stage 1: carrega todos os vetores em memória e constrói histograma
	// por valor individual de vector[0] (256 valores possíveis).
	// Necessário antes de calcular os limites dos buckets.
	var allRows []Row
	var hist [256]uint32

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

		allRows = append(allRows, r)
		hist[r.Vector[0]]++
	}

	total := uint32(len(allRows))
	log.Printf("Total de vetores processados: %d", total)

	// Histograma detalhado de vector[0] para diagnóstico
	log.Println("=== Histograma de vector[0] ===")
	for v := 0; v < 256; v++ {
		if hist[v] > 0 {
			log.Printf("  v=%3d: %7d (%.2f%%)", v, hist[v], float64(hist[v])/float64(total)*100)
		}
	}
	log.Println("================================")

	// Stage 2: calcula o mapa de buckets de forma adaptativa.
	//
	// Algoritmo: percorre os 256 valores possíveis de vector[0] em ordem e
	// acumula contagens. Quando o acumulado >= target por bucket, avança para
	// o próximo bucket. O excedente é carregado para o próximo bucket (carry-over),
	// garantindo que um único valor com muitos vetores distribui corretamente
	// entre múltiplos buckets.
	targetPerBucket := total / NumBuckets
	if targetPerBucket == 0 {
		targetPerBucket = 1
	}

	var bucketMap [256]uint8
	currentBucket := uint8(0)
	accumulated := uint32(0)

	for v := 0; v < 256; v++ {
		bucketMap[v] = currentBucket
		accumulated += hist[v]
		// Avança o bucket (possivelmente mais de uma vez) enquanto o
		// acumulado supera o target e ainda há buckets disponíveis.
		for currentBucket < NumBuckets-1 && accumulated >= targetPerBucket {
			currentBucket++
			accumulated -= targetPerBucket
		}
	}

	// Log dos limites calculados
	log.Println("=== Mapa de Buckets Calculado ===")
	bucketRanges := [NumBuckets][2]int{}
	for b := range NumBuckets {
		bucketRanges[b] = [2]int{256, -1} // min, max
	}
	for v := 0; v < 256; v++ {
		b := int(bucketMap[v])
		if v < bucketRanges[b][0] {
			bucketRanges[b][0] = v
		}
		if v > bucketRanges[b][1] {
			bucketRanges[b][1] = v
		}
	}
	for b := range NumBuckets {
		if bucketRanges[b][1] >= 0 {
			log.Printf("  Bucket %2d: v0 [%3d–%3d]", b, bucketRanges[b][0], bucketRanges[b][1])
		} else {
			log.Printf("  Bucket %2d: vazio", b)
		}
	}
	log.Println("=================================")

	// Salva o mapa de buckets (256 bytes) para ser carregado pelo servidor.
	bucketMapFile, err := os.Create("files/bucket_map.bin")
	if err != nil {
		log.Fatal(err)
	}
	if _, err := bucketMapFile.Write(bucketMap[:]); err != nil {
		log.Fatal(err)
	}
	bucketMapFile.Close()
	log.Println("bucket_map.bin salvo com sucesso")

	// Stage 3: distribui os vetores nos buckets usando o mapa calculado
	var buckets [NumBuckets][]Row
	for _, r := range allRows {
		bid := bucketMap[r.Vector[0]]
		buckets[bid] = append(buckets[bid], r)
	}

	// Log da distribuição resultante
	log.Println("=== Distribuição dos Buckets (process) ===")
	for i := range NumBuckets {
		pct := 0.0
		if total > 0 {
			pct = float64(len(buckets[i])) / float64(total) * 100
		}
		log.Printf("  Bucket %2d (v0 [%3d–%3d]): %7d vetores (%.1f%%)",
			i, bucketRanges[i][0], bucketRanges[i][1], len(buckets[i]), pct)
	}
	log.Println("==========================================")

	// Escreve os arquivos binários
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
	for i := range NumBuckets {
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
