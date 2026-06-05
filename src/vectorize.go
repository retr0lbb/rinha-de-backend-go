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
	t, err := parseRFC3339Fast(payload.Transaction.RequestedAt)
	if err == nil {
		t = t.UTC()

		vetor[3] = utils.Quantize(float32(t.Hour()) / 23.0)
		vetor[4] = utils.Quantize(float32(t.Weekday()) / 6.0)
	}

	// 5 & 6: Última transação APENAS ESSES QUE TEM O -1
	if payload.LastTransaction == nil {
		vetor[5] = utils.Quantize(-1)
		vetor[6] = utils.Quantize(-1)
	} else {
		lastT, err := parseRFC3339Fast(payload.LastTransaction.Timestamp)

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
		vetor[9] = utils.Quantize(1)
	}

	if payload.Terminal.CardPresent {
		vetor[10] = utils.Quantize(1)
	}

	// 11
	isKnown := slices.Contains(
		payload.Customer.KnownMerchants,
		payload.Merchant.ID,
	)

	if !isKnown {
		vetor[11] = utils.Quantize(1)
	}

	score, ok := MCCScores[payload.Merchant.MCC]
	if !ok {
		score = 0.5
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

func parseRFC3339Fast(s string) (time.Time, error) {
	if len(s) < 19 {
		return time.Parse(time.RFC3339, s)
	}
	// Validação básica de delimitadores
	if s[4] != '-' || s[7] != '-' || s[10] != 'T' || s[13] != ':' || s[16] != ':' {
		return time.Parse(time.RFC3339, s)
	}
	year := int(s[0]-'0')*1000 + int(s[1]-'0')*100 + int(s[2]-'0')*10 + int(s[3]-'0')
	month := time.Month(int(s[5]-'0')*10 + int(s[6]-'0'))
	day := int(s[8]-'0')*10 + int(s[9]-'0')
	hour := int(s[11]-'0')*10 + int(s[12]-'0')
	min := int(s[14]-'0')*10 + int(s[15]-'0')
	sec := int(s[17]-'0')*10 + int(s[18]-'0')

	loc := time.UTC
	idx := 19
	if idx < len(s) && s[idx] == '.' {
		idx++
		for idx < len(s) && s[idx] >= '0' && s[idx] <= '9' {
			idx++
		}
	}
	if idx < len(s) {
		char := s[idx]
		if char == '+' || char == '-' {
			if len(s) >= idx+6 && s[idx+3] == ':' {
				sign := 1
				if char == '-' {
					sign = -1
				}
				ohour := int(s[idx+1]-'0')*10 + int(s[idx+2]-'0')
				omin := int(s[idx+4]-'0')*10 + int(s[idx+5]-'0')
				offset := sign * (ohour*3600 + omin*60)
				loc = time.FixedZone("", offset)
			} else {
				return time.Parse(time.RFC3339, s)
			}
		}
	}

	return time.Date(year, month, day, hour, min, sec, 0, loc), nil
}

