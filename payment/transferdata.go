package payment

type TransferData struct {
	S      Account `json:"source"`
	D      Account `json:"dectination"`
	Amount float32 `json:"amount"`
}

func NewTransferData(s Account, d Account, amount float32) TransferData {
	t := TransferData{
		S:      s,
		D:      d,
		Amount: amount,
	}

	return t
}
