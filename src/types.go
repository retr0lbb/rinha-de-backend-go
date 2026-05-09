package main

// Dica: use time.Time para datas se for fazer cálculos

type Transaction struct {
	Amount       float64 `json:"amount"` // Use float64 para evitar perda de precisão
	Installments int     `json:"installments"`
	RequestedAt  string  `json:"requested_at"` // Ajustado para snake_case
}

type Customer struct {
	AvgAmount      float64  `json:"avg_amount"`      // Ajustado
	TxCount24h     int      `json:"tx_count_24h"`    // Ajustado
	KnownMerchants []string `json:"known_merchants"` // Ajustado
}

type Merchant struct {
	ID        string  `json:"id"`         // Corrigido
	MCC       string  `json:"mcc"`        // No JSON veio como string "5912", mudei para string
	AvgAmount float64 `json:"avg_amount"` // Adicionado tag
}

type Terminal struct {
	IsOnline    bool    `json:"is_online"`    // Adicionado tag
	CardPresent bool    `json:"card_present"` // Adicionado tag
	KmFromHome  float64 `json:"km_from_home"` // Adicionado tag
}

type LastTransaction struct {
	Timestamp     string  `json:"timestamp"`
	KmFromCurrent float64 `json:"km_from_current"`
}

type Payload struct {
	ID              string           `json:"id"`
	Transaction     Transaction      `json:"transaction"`
	Customer        Customer         `json:"customer"`
	Merchant        Merchant         `json:"merchant"`
	Terminal        Terminal         `json:"terminal"`
	LastTransaction *LastTransaction `json:"last_transaction"` // Ponteiro para permitir null
}

type ResponsePayload struct {
	Approved   bool    `json:"approved"`
	FraudScore float32 `json:"fraud_score"`
}

type MerchantScore struct {
	ID    string  `json:"id"`
	Score float32 `json:"score"`
}
