package main

type Payload struct {
	id string

	transaction struct {
		amount       float32 //valor da transacao
		installments int     //numero de parcelas
		requestedAt  string  //timestamp da requisicao
	}

	customer struct {
		avg_amount      float32  //media de gasto do usuario do cartao
		tx_count_24h    int      //contagem de pagamenos nas ultimas 24 horas
		known_merchants []string //comerciantes ja usados pelo usuario
	}

	merchant struct {
		id         string
		mcc        int     //Codigo de categoria do mercante
		avg_amount float32 //ticket medio do mercante
	}

	terminal struct {
		is_online    bool    // se o pagamento for online
		card_present bool    // se o pagamento for presencial
		km_from_home float64 // distancia em km do endereco do portador
	}

	last_transaction struct { //pode ser null
		timestamp       string  //tempo
		km_from_current float64 //distancia entre ultima transacao
	}
}

type ResponsePayload struct {
	approved    bool
	fraud_score float32
}
