package main

import (
	"encoding/json"
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	err := loadMCCscores("files/mcc_risk.json")
	openLargeFile("files/vectors.bin", "files/labels.bin")

	log.Println("Dataset carregado com sucesso!")

	if err != nil {
		log.Fatalf("Erro ao carregar MCC scores: %v", err)
	}

	mux.HandleFunc("/ready", handleReady)
	mux.HandleFunc("/fraud-score", handleAnalyze)
	log.Fatal(http.ListenAndServe(":9999", mux))
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

	//handler
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
