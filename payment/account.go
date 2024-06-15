package payment

const (
	BYN = "BYN"
	RU  = "RU"
	USD = "USD"
	EUR = "EUR"
)

const (
	Active  = "active"
	Blocked = "blocked"
)

type Account struct {
	CustomerId   string  `json:"customer_id"`
	Num          string  `json:"num"`
	CurrencyCode string  `json:"currency_code"`
	Status       string  `json:"status"`
	Balance      float32 `json:"balance"`
	Description  string  `json:"desc"`
}

func NewAccount(cid string, currencyCode string, accountNum string, amount float32) Account {
	a := Account{
		CustomerId:   cid,
		Num:          accountNum,
		CurrencyCode: currencyCode,
		Status:       Active,
		Balance:      amount,
	}

	return a
}
