package main

import (
	"rinha-de-backend-retr0lbb/src/config"
	"testing"
	"time"
)

func TestVectorizeFunction(t *testing.T) {
	payload := Payload{
		ID: "1",

		Transaction: Transaction{
			Amount:       1000,
			Installments: 3,
			RequestedAt:  time.Now().Format(time.RFC3339),
		},

		Customer: Customer{
			AvgAmount:      100,
			TxCount24h:     1,
			KnownMerchants: []string{},
		},

		Merchant: Merchant{
			ID:        "1",
			MCC:       "1",
			AvgAmount: 300,
		},

		Terminal: Terminal{
			IsOnline:    true,
			CardPresent: false,
			KmFromHome:  19.09123,
		},

		LastTransaction: nil,
	}

	vector := HandleVectorizePayload(payload)

	t.Log(vector)

	if len(vector) != 14 {
		t.Errorf("expected vector length 14, got %d", len(vector))
	}

	if vector[0] != float32(limiter(payload.Transaction.Amount, config.MaxAmount)) {
		t.Errorf("Expected Normalized transaction value to be %f and got %f", limiter(payload.Transaction.Amount, config.MaxAmount), vector[0])
	}
}
