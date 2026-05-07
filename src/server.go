package main

import (
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"slices"
	"time"
)

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/", handleHello)

	log.Fatal(http.ListenAndServe(":8080", mux))
}

func handleHello(w http.ResponseWriter, _ *http.Request) {
	wc, err := w.Write([]byte("Hello, World!\n"))

	if err != nil {
		slog.Error("Error Writing the response", "err", err)
		return
	}

	fmt.Printf("%d bytes written\n", wc)
}

// Constante para o formato da string de data que vem no JSON (ajuste se necessário)
const TimeFormat = time.RFC3339

func handleVectorizePayload(payload Payload) [14]float32 {
	const (
		MaxAmount            = 10000
		MaxInstallments      = 12
		AmountVsAvgRatio     = 10
		MaxMinutes           = 1440
		MaxKm                = 1000
		MaxTxCount24h        = 20
		MaxMerchantAvgAmount = 10000
	)

	var vetor [14]float32

	// 0 & 1: Transação básica
	vetor[0] = limiter(payload.transaction.amount, float32(MaxAmount))
	vetor[1] = limiter(float32(payload.transaction.installments), float32(MaxInstallments))

	// 2: Razão vs Média do Cliente
	if payload.customer.avg_amount > 0 {
		ratio := payload.transaction.amount / payload.customer.avg_amount
		vetor[2] = limiter(ratio, float32(AmountVsAvgRatio))
	}

	// 3 & 4: Tempo (Parsing da string de data)
	t, err := time.Parse(TimeFormat, payload.transaction.requestedAt)
	if err == nil {
		t = t.UTC()
		vetor[3] = float32(t.Hour()) / 23.0
		vetor[4] = float32(t.Weekday()) / 6.0
	}

	// 5 & 6: Última Transação (Como a struct está embutida, checamos se o campo chave está vazio)
	if payload.last_transaction.timestamp == "" {
		vetor[5] = -1
		vetor[6] = -1
	} else {
		lastT, err := time.Parse(TimeFormat, payload.last_transaction.timestamp)
		if err == nil {
			diff := t.Sub(lastT).Minutes()
			vetor[5] = limiter(float32(diff), float32(MaxMinutes))
		}
		vetor[6] = limiter(float32(payload.last_transaction.km_from_current), float32(MaxKm))
	}

	// 7 & 8: Terminal e Frequência
	vetor[7] = limiter(float32(payload.terminal.km_from_home), float32(MaxKm))
	vetor[8] = limiter(float32(payload.customer.tx_count_24h), float32(MaxTxCount24h))

	// 9 & 10: Flags Binárias (Convertendo bool para 0/1)
	if payload.terminal.is_online {
		vetor[9] = 1
	}
	if payload.terminal.card_present {
		vetor[10] = 1
	}

	// 11: Merchant Desconhecido (Lógica invertida: 1 se for novo/desconhecido)
	isKnown := slices.Contains(payload.customer.known_merchants, payload.merchant.id)
	if !isKnown {
		vetor[11] = 1
	}

	// 12: Risco MCC (Exemplo usando um valor padrão, substitua pela sua lógica de mapa)
	vetor[12] = 0.5

	// 13: Média do Mercante
	vetor[13] = limiter(payload.merchant.avg_amount, float32(MaxMerchantAvgAmount))

	return vetor
}
