package models

type TopTraderResponse struct {
	Data []Trader `json:"data"`
}

type Trader struct {
	ID         string     `json:"id"`
	Type       string     `json:"type"`
	Attributes Attributes `json:"attributes"`
}

type Attributes struct {
	BoughtUsd       string      `json:"boughtUsd"`
	BoughtToken     string      `json:"boughtToken"`
	BoughtCount     int         `json:"boughtCount"`
	SoldUsd         string      `json:"soldUsd"`
	SoldToken       string      `json:"soldToken"`
	SoldCount       int         `json:"soldCount"`
	RemainingTokens interface{} `json:"remainingTokens"`
	PlUsd           string      `json:"plUsd"`
	Kind            string      `json:"kind"`
	Signer          string      `json:"signer"`
}
