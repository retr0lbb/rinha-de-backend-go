package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"runtime"
	"time"
)

// searchSem limita o número de buscas KD-Tree concorrentes.
// Com GOMAXPROCS=1 e 0.45 CPU, mais de 16 buscas simultâneas
// causa thrashing sem ganho de throughput.
var searchSem = make(chan struct{}, 16)

func main() {
	// Com cgroup de 0.45 CPU, usar mais de 1 thread OS causa context-switch
	// overhead sem ganho real. GOMAXPROCS=1 elimina thrashing.
	runtime.GOMAXPROCS(1)

	mux := http.NewServeMux()

	err := loadMCCscores("files/mcc_risk.json")
	if err != nil {
		log.Fatalf("Erro ao carregar MCC scores: %v", err)
	}

	err = openLargeFile("files/vectors.bin", "files/labels.bin")
	if err != nil {
		log.Fatalf("Erro ao carregar arquivos bin: %v", err)
	}

	err = loadBuckets("files/buckets.bin")
	if err != nil {
		log.Fatalf("Erro ao carregar buckets: %v", err)
	}

	BuildKDTrees()

	// Estimar footprint de memória da KD-Tree
	kdTreeMB := float64(len(KDTreeNodes)*16) / (1024 * 1024)
	log.Printf("KD-Tree: %d nós | footprint: %.2f MB", len(KDTreeNodes), kdTreeMB)
	log.Println("Dataset carregado com sucesso!")

	mux.HandleFunc("/ready", handleReady)
	mux.HandleFunc("/fraud-score", handleAnalyze)
	mux.HandleFunc("/debug/stats", handleStats)

	srv := &http.Server{
		Addr:              ":9999",
		Handler:           mux,
		ReadTimeout:       2 * time.Second,
		WriteTimeout:      3 * time.Second,
		IdleTimeout:       30 * time.Second,
		ReadHeaderTimeout: 500 * time.Millisecond,
	}

	log.Printf("Servidor iniciado em :9999 (GOMAXPROCS=%d)", runtime.GOMAXPROCS(0))
	log.Fatal(srv.ListenAndServe())
}

func handleReady(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func handleAnalyze(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req Payload
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "JSON invalido", http.StatusBadRequest)
		return
	}

	// Limitar concorrência: aguarda slot disponível no semáforo
	searchSem <- struct{}{}
	defer func() { <-searchSem }()

	vector := HandleVectorizePayload(req)
	score, approved := search(vector)

	res := ResponsePayload{
		Approved:   approved,
		FraudScore: score,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}

// handleStats expõe métricas de diagnóstico em texto simples.
// Útil para monitorar goroutines, throughput e taxa de cap da KD-Tree.
func handleStats(w http.ResponseWriter, r *http.Request) {
	searches := totalSearches.Load()
	visited := totalVisited.Load()
	capped := cappedSearches.Load()

	avgVisited := 0.0
	if searches > 0 {
		avgVisited = float64(visited) / float64(searches)
	}

	capPct := 0.0
	if searches > 0 {
		capPct = float64(capped) / float64(searches) * 100
	}

	fmt.Fprintf(w,
		"goroutines:       %d\n"+
			"sem_occupancy:    %d/%d\n"+
			"total_searches:   %d\n"+
			"avg_nodes_visited: %.1f\n"+
			"capped_searches:  %d (%.1f%% atingiram MaxKDVisited=%d)\n",
		runtime.NumGoroutine(),
		len(searchSem), cap(searchSem),
		searches,
		avgVisited,
		capped, capPct, MaxKDVisited,
	)
}
