package main

import (
	"testing"
	"time"
)

func NewValidPayload() Payload {
	return Payload{
		ID: "1",
		Transaction: Transaction{
			Amount:       10000,
			Installments: 12,
			RequestedAt:  time.Now().Format(time.RFC3339),
		},
		Customer: Customer{
			AvgAmount:      10000,
			TxCount24h:     20,
			KnownMerchants: []string{},
		},
		Merchant: Merchant{
			ID:        "NULL",
			MCC:       "NULL",
			AvgAmount: 10000,
		},
		Terminal: Terminal{
			IsOnline:    true,
			CardPresent: false,
			KmFromHome:  1000,
		},
		LastTransaction: nil,
	}
}

func TestVectorization(t *testing.T) {
	payload := NewValidPayload()

	result := HandleVectorizePayload(payload)

	t.Log(result)

	if len(result) < 14 {
		t.Error("vector must contains 14 dimensions")
	}

	if result[5] != 255 {
		t.Errorf("Data not quantized correctly expected vector[5] to be 255 received: %d", result[5])
	}

	if result[6] != 255 {
		t.Errorf("Data not quantized correctly expected vector[6] to be 255 received: %d", result[6])
	}

	if result[9] != 254 || result[10] != 0 || result[11] != 254 {
		t.Errorf("Data not quantized correctly on boolean values expected 254 received: %d, %d, %d", result[9], result[10], result[11])
	}
}
