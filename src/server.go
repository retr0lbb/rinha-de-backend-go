package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/ready", handleReady)
	mux.HandleFunc("POST /fraud-score", handleAnalyze)

	log.Fatal(http.ListenAndServe(":9999", mux))
}

func handleReady(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func handleAnalyze(w http.ResponseWriter, r *http.Request) {
	var req Payload
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "JSON invalido", http.StatusBadRequest)
		return
	}

	//handler
	vector := HandleVectorizePayload(req)

	fmt.Println(vector)

	res := ResponsePayload{
		Approved:   true,
		FraudScore: 1.0,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	json.NewEncoder(w).Encode(res)
}
