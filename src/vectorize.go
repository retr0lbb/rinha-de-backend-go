package main

import (
	"encoding/json"
	"os"
	"rinha-de-backend-retr0lbb/utils"
	"slices"
	"time"
)

// mudar para melhor performance depois
const TimeFormat = time.RFC3339

var MCCScores map[string]float32

func loadMCCscores(path string) error {
	file, err := os.Open(path)

	if err != nil {
		return err
	}

	defer file.Close()

	decoder := json.NewDecoder(file)

	return decoder.Decode(&MCCScores)
}

func HandleVectorizePayload(payload Payload) [14]uint8 {
	const (
		MaxAmount            = 10000
		MaxInstallments      = 12
		AmountVsAvgRatio     = 10
		MaxMinutes           = 1440
		MaxKm                = 1000
		MaxTxCount24h        = 20
		MaxMerchantAvgAmount = 10000
	)

	var vetor [14]uint8

	// 0 & 1: Transação básica
	vetor[0] = utils.Quantize(float32(limiter(payload.Transaction.Amount, MaxAmount)))
	vetor[1] = utils.Quantize(float32(limiter(float64(payload.Transaction.Installments), float64(MaxInstallments))))

	// 2: Razão vs Média do Cliente
	if payload.Customer.AvgAmount > 0 {
		ratio := payload.Transaction.Amount / payload.Customer.AvgAmount
		vetor[2] = utils.Quantize(float32(limiter(ratio, float64(AmountVsAvgRatio))))
	}

	// 3 & 4: Tempo
	t, err := time.Parse(time.RFC3339, payload.Transaction.RequestedAt)
	if err == nil {
		t = t.UTC()

		vetor[3] = utils.Quantize(float32(t.Hour()) / 23.0)
		vetor[4] = utils.Quantize(float32(t.Weekday()) / 6.0)
	}

	// 5 & 6: Última transação
	if payload.LastTransaction == nil {
		vetor[5] = utils.Quantize(-1)
		vetor[6] = utils.Quantize(-1)
	} else {
		lastT, err := time.Parse(
			time.RFC3339,
			payload.LastTransaction.Timestamp,
		)

		if err == nil {
			diff := t.Sub(lastT).Minutes()

			vetor[5] = utils.Quantize(float32(limiter(
				float64(diff),
				float64(MaxMinutes),
			)))
		}

		vetor[6] = utils.Quantize(float32(limiter(
			float64(payload.LastTransaction.KmFromCurrent),
			float64(MaxKm),
		)))
	}

	// 7 & 8
	vetor[7] = utils.Quantize(float32(limiter(
		float64(payload.Terminal.KmFromHome),
		float64(MaxKm),
	)))

	vetor[8] = utils.Quantize(float32(limiter(
		float64(payload.Customer.TxCount24h),
		float64(MaxTxCount24h),
	)))

	// 9 & 10
	if payload.Terminal.IsOnline {
		vetor[9] = 1
	}

	if payload.Terminal.CardPresent {
		vetor[10] = 1
	}

	// 11
	isKnown := slices.Contains(
		payload.Customer.KnownMerchants,
		payload.Merchant.ID,
	)

	if !isKnown {
		vetor[11] = 1
	}

	score, ok := MCCScores[payload.Merchant.MCC]
	if !ok {
		score = 0.5 // Valor padrão se não encontrar o MCC
	}

	// 12 merchant score
	vetor[12] = utils.Quantize(score)

	// 13
	vetor[13] = utils.Quantize(float32(limiter(
		payload.Merchant.AvgAmount,
		float64(MaxMerchantAvgAmount),
	)))

	return vetor
}
