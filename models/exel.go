package models

type FilteredTrader struct {
	Signer      string
	WinRate     float64
	BoughtUsd   string
	SoldUsd     string
	BoughtCount int
	SoldCount   int
	PlUsd       string
	Kind        string
}
