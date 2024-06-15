package payment

type Customer struct {
	Id        string
	Name      string
	AccPrefix string
}

func NewCustomer(id string, name string, accPrefix string) Customer {
	c := Customer{
		Id:        id,
		Name:      name,
		AccPrefix: accPrefix,
	}

	return c
}
