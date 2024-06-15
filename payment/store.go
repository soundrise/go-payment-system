//go:generate mockgen -source=store.go -destination=mocks/store_mock.go -package=payment

package payment

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"regexp"
	"sync"
	"time"
)

const (
	AccountPrefix               = "BY"
	AccountStateEmissionPrefix  = "SE"                                                       // STATE_EMISSION (for quick access to special account)
	AccountStateTerminatePrefix = "ST"                                                       // STATE_TERMINATE (for quick access to special account)
	AccountStateEmissionNumber  = AccountStateEmissionPrefix + "00MMMM00000000000000000001"  // special emission account number
	AccountStateTerminateNumber = AccountStateTerminatePrefix + "00MMMM00000000000000000002" // special terminate account number
	AccountCharacters           = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	AccountNumbers              = "0123456789"
	AccountLength               = 26
)

func init() {
	source := rand.NewSource(time.Now().UnixNano())
	rand.New(source)
}

type Store interface {
	CreateAccount(c Customer, accType string, currencyCode string, amount float32) error
	GetSpecialAccount(accountPrefix string) (Account, error)
	FindAccount(c Customer, currencyCode string) (Account, error)
	CloseAccount(ac Account) error
	Emit(amount float32) error
	Terminate(acc Account, a float32) error
	Transfer(s Account, d Account, amount float32) error
	TransferJson(json []byte) error
	PrintStore() error
	PrintStoreJson() error
	Lock() error
	Unlock() error
	DumpStore() ([]byte, error)
	Restore(json []byte) error
}

/*
This struct realize the Store interface
and represent a bank store and gives the methods to work with it.
*/
type PaymentSystem struct {
	store map[string]map[string]Account
	mu    sync.Mutex
}

func NewPaymentSystem() *PaymentSystem {
	ps := &PaymentSystem{
		store: make(map[string]map[string]Account),
	}

	return ps
}

func (ps *PaymentSystem) Lock() error {
	ps.mu.Lock()

	return nil
}

func (ps *PaymentSystem) Unlock() error {
	ps.mu.Unlock()

	return nil
}

/*
This func makes a copy of the database to restore.
*/
func (ps *PaymentSystem) DumpStore() ([]byte, error) {
	jsonMap, err := json.Marshal(ps.store)
	if err != nil {
		return nil, err
	}

	s := make([]byte, len(jsonMap))

	copy(s, jsonMap)

	return s, nil
}

/*
This func restore database(store map) when transaction is aborted.
*/
func (ps *PaymentSystem) Restore(oldStore []byte) error {
	ps.store = make(map[string]map[string]Account)
	s := ps.store

	err := json.Unmarshal(oldStore, &s)
	if err != nil {
		return err
	}

	return nil
}

/*
This func create a new account with params assined to customer.
*/
func (ps *PaymentSystem) CreateAccount(c Customer, accType string, currencyCode string, amount float32) error {
	log.Printf("Try to create account for cusromer %s currency %s amount %.2f \n", c.Id, currencyCode, amount)

	if amount < 0 {
		log.Println("Creating account is imposible: amount < 0")

		return fmt.Errorf("Creating account  is imposible: amount < 0 \n")
	}

	if c.Id == "" {
		log.Println("Creating account is imposible: without client ID")

		return fmt.Errorf("Creating account is imposible: without client ID\n")
	}

	sid := ""
	aid := ""

	if accType == AccountStateEmissionPrefix {
		sid = AccountStateEmissionPrefix
		aid = AccountStateEmissionNumber
	} else if accType == AccountStateTerminatePrefix {
		sid = AccountStateTerminatePrefix
		aid = AccountStateTerminateNumber
	} else {
		sid = generateIdentifier(AccountLength)
		aid = GenerateAccountNumber()
	}

	na := NewAccount(c.Id, currencyCode, aid, amount)

	list, ok := ps.store[c.Id]
	if ok {
		list[sid] = na
	} else {
		accounts := make(map[string]Account)
		accounts[sid] = na
		ps.store[c.Id] = accounts
	}

	log.Printf("Account with number:%s is created successfully\n", na.Num)

	return nil
}

// Function for quick access to special accounts.
func (ps *PaymentSystem) GetSpecialAccount(accountPrefix string) (Account, error) {
	var res Account

	acc, ok := ps.store["0"][accountPrefix]
	if ok {
		res = acc
	} else {
		log.Println("Special account not found")

		return res, fmt.Errorf("special account not found %s\n", accountPrefix)
	}

	if accountPrefix == AccountStateEmissionPrefix {
		log.Printf("EmissionAccount Number  is %s\n", res.Num)
	} else if accountPrefix == AccountStateTerminatePrefix {
		log.Printf("TerminateAccount Number is %s\n", res.Num)
	} else {
		log.Println("Special account not found")

		return res, fmt.Errorf("special account not found %s\n", accountPrefix)
	}

	err := VerifyAccountNumber(res.Num)
	if err != nil {
		return res, err
	}

	return res, nil
}

/*
This func returns the account from store by customer and account number.
*/
func (ps *PaymentSystem) GetAccountByNumber(c Customer, accountNum string) (Account, error) {
	var res Account

	_, ok := ps.store[c.Id]
	if ok {
		for _, v := range ps.store[c.Id] {
			if v.Num == accountNum {
				res = v
			}
		}
	} else {
		log.Println("Account not found!")
		// panic("Account not found!")
		return res, fmt.Errorf("Account not found!%s", accountNum)
	}

	return res, nil
}

/*
This func returns the account from store by customer and currency code.
*/
func (ps *PaymentSystem) FindAccount(c Customer, currencyCode string) (Account, error) {
	var res Account

	for _, v := range ps.store[c.Id] {
		if v.CurrencyCode == currencyCode {
			res = v
		}
	}

	err := VerifyAccountNumber(res.Num)
	if err != nil {
		return res, err
	}

	return res, nil
}

/*
This func blocks the account.
*/
func (ps *PaymentSystem) CloseAccount(ac Account) error {
	log.Printf("Try to block account %s \n", ac.Num)

	var res Account

	var oldKey string

	_, ok := ps.store[ac.CustomerId]

	if ok {
		for k, v := range ps.store[ac.CustomerId] {
			if v.Num == ac.Num && v.CustomerId == ac.CustomerId {
				res = v
				oldKey = k

				break
			}
		}

		res.Status = Blocked
		ps.store[ac.CustomerId][oldKey] = res

		log.Printf("Account %s was closed successfully. CustomerId: %s\n", res.Num, res.CustomerId)

		return nil
	}

	log.Printf("Account %s is not closed. The reason: not founded. CustomerId: %s\n", res.Num, res.CustomerId)

	return fmt.Errorf("Account %s is not closed. The reason: not founded. CustomerId: %s\n", res.Num, res.CustomerId)
}

/*
This func activate the account.
*/
func (ps *PaymentSystem) ActivateAccount(ac Account) error {
	log.Printf("Try to activate account %s \n", ac.Num)

	var res Account

	var oldKey string

	_, ok := ps.store[ac.CustomerId]

	if ok {
		for k, v := range ps.store[ac.CustomerId] {
			if v.Num == ac.Num && v.CustomerId == ac.CustomerId {
				res = v
				oldKey = k

				break
			}
		}

		res.Status = Active
		ps.store[ac.CustomerId][oldKey] = res

		log.Printf("Account %s was activate successfully. CustomerId: %s\n", res.Num, res.CustomerId)

		return nil
	}

	log.Printf("Account %s is not activated. The reason: not founded. CustomerId: %s\n", res.Num, res.CustomerId)

	return fmt.Errorf("Account %s is not activated. The reason: not founded. CustomerId: %s\n", res.Num, res.CustomerId)
}

/*
This func emit amount of money to special emission account.
*/
func (ps *PaymentSystem) Emit(amount float32) error {
	log.Printf("Try to emit amount: %.2f to spec emission account \n", amount)

	if amount <= 0 {
		log.Println("Emission is imposible: amount <= 0")

		return fmt.Errorf("Emission is imposible: amount <= 0 \n")
	}

	e, err := ps.GetSpecialAccount(AccountStateEmissionPrefix)
	if err != nil {
		log.Println("Emission is imposible: account not found!")

		return err
	}

	e.Balance += amount

	var oldKey string

	_, ok := ps.store[e.CustomerId]
	if ok {
		for k, v := range ps.store[e.CustomerId] {
			if v.Num == e.Num && v.CustomerId == e.CustomerId {
				oldKey = k

				break
			}
		}

		ps.store[e.CustomerId][oldKey] = e

		log.Printf("Emission %2.f to account %s was done successfully. CustomerId: %s\n", amount, e.Num, e.CustomerId)
	} else {
		log.Println("Emission is imposible: account not found!")

		return fmt.Errorf("Emission is imposible: spec emission account not found! \n")
	}

	return nil
}

/*
This func transfers amount of money from source account to terminate account.
*/
func (ps *PaymentSystem) Terminate(s Account, amount float32) error {
	log.Printf("Try to terminate amount: %.2f from account %s\n", amount, s.Num)

	if !accountAvailable(s) {
		log.Printf("Termination is imposible: account %s is not valid or blocked", s.Num)

		return fmt.Errorf("Termination is imposible: account %s is not valid or blocked", s.Num)
	}

	if amount <= 0 {
		log.Println("Termination is imposible: amount <= 0")

		return fmt.Errorf("Termination is imposible: amount <= 0 \n")
	}

	t, err := ps.GetSpecialAccount(AccountStateTerminatePrefix)
	if err != nil {
		log.Println("Termination is imposible: account not found!")

		return err
	}

	var res Account

	var oldKey, oldKey1 string

	// find key and value for source account

	_, ok := ps.store[s.CustomerId]
	if !ok {
		log.Println("Termination is imposible: account not found!")
		// panic("Terminate is imposible: account not found!")
		return fmt.Errorf("Termination is imposible: account not found! %s\n", s.Num)
	}

	// find key for terminate account
	_, ok1 := ps.store[t.CustomerId]
	if !ok1 {
		log.Println("Termination is imposible: account not found!")
		// panic("Terminate is imposible: account not found!")
		return fmt.Errorf("Termination is imposible: spec terminate account not found! \n")
	}

	if ok {
		for k, v := range ps.store[s.CustomerId] {
			if v.Num == s.Num && v.CustomerId == s.CustomerId {
				oldKey = k
				res = v

				break
			}
		}

		res.Balance -= amount
		ps.store[s.CustomerId][oldKey] = res

		log.Printf("Amount: %.2f was terminated from account %s\n", amount, s.Num)
	}

	if ok1 {
		for k, v := range ps.store[t.CustomerId] {
			if v.Num == t.Num && v.CustomerId == t.CustomerId {
				oldKey1 = k

				break
			}
		}

		t.Balance += amount
		ps.store[t.CustomerId][oldKey1] = t

		log.Printf("Amount: %.2f was transferred to special terminate account %s\n", amount, t.Num)
	}

	return nil
}

func accountAvailable(s Account) bool {
	return VerifyAccountNumber(s.Num) == nil && s.Status == Active
}

/*
This func transfers amount of money from source account to destination account.
*/
func (ps *PaymentSystem) Transfer(s Account, d Account, amount float32) error {
	log.Printf("Try to transfer amount: %.2f from account %s to account %s\n", amount, s.Num, d.Num)

	if !accountAvailable(s) {
		log.Printf("Transfer is imposible: account %s is not valid or blocked", s.Num)

		return fmt.Errorf("Transfer is imposible: account %s is not valid or blocked", s.Num)
	}

	if !accountAvailable(d) {
		log.Printf("Transfer is imposible: account %s is not valid or blocked", d.Num)

		return fmt.Errorf("Transfer is imposible: account %s is not valid or blocked", d.Num)
	}

	if amount <= 0 {
		log.Println("Transfer is imposible: amount < 0")

		return fmt.Errorf("Transfer is imposible: amount < 0 \n")
	}

	if s.CurrencyCode != d.CurrencyCode {
		log.Println("Transfer is imposible: different currency!")

		return fmt.Errorf("Transfer is imposible: different currency! \n")
	}

	var res Account

	var oldKey, oldKey1 string

	// find key and value for source account
	_, ok := ps.store[s.CustomerId]
	if !ok {
		log.Println("Transfer is imposible: account not found!")
		// panic("Transfer is imposible: account not found!")
		return fmt.Errorf("Transfer is imposible: account not found! %s\n", s.Num)
	}
	// find key for terminate account
	_, ok1 := ps.store[d.CustomerId]

	if !ok1 {
		log.Println("Transfer is imposible: account not found!")
		// panic("Transfer is imposible: account not found!")
		return fmt.Errorf("Transfer is imposible: account not found! %s\n", d.Num)
	}

	if ok {
		for k, v := range ps.store[s.CustomerId] {
			if v.Num == s.Num && v.CustomerId == s.CustomerId {
				oldKey = k
				res = v

				break
			}
		}

		res.Balance -= amount
		ps.store[s.CustomerId][oldKey] = res

		log.Printf("Amount: %.2f was transferred from account %s\n", amount, s.Num)
	}

	// find key for terminate account

	if ok1 {
		for k, v := range ps.store[d.CustomerId] {
			if v.Num == d.Num && v.CustomerId == d.CustomerId {
				oldKey1 = k

				break
			}
		}

		d.Balance += amount
		ps.store[d.CustomerId][oldKey1] = d

		log.Printf("Amount: %.2f was transferfed to account %s\n", amount, d.Num)
	}

	return nil
}

func (ps *PaymentSystem) TransferJson(data []byte) error {
	// deserialize data
	var t TransferData

	err := json.Unmarshal(data, &t)
	if err != nil {
		log.Printf("Error: cannot deserialize JSON object.")

		return err
	}

	return ps.Transfer(t.S, t.D, t.Amount)
}

func (ps *PaymentSystem) PrintStoreJson() error {
	fmt.Println("______________________________________________________")
	fmt.Println()
	fmt.Println("\nStore Info:")
	fmt.Println("______________________________________________________")

	obj, err := json.MarshalIndent(ps.store, "", "    ")
	if err != nil {
		log.Printf("error while converting store to json in func PrintStoreJson: %v", err)

		return err
	}

	fmt.Println(string(obj))
	fmt.Println("______________________________________________________")

	return nil
}

func (ps *PaymentSystem) PrintStore() error {
	log.Println("______________________________________________________")
	log.Println()
	log.Println("\nStore Info:")
	log.Println("______________________________________________________")

	for k, v := range ps.store {
		log.Println()
		log.Printf("Client Id: %s\n", k)

		for _, a := range v {
			log.Printf("account num: %v\n", a)
		}

		log.Println("______________________________________________________")
	}

	return nil
}

/*
Функция генерирует в случайном порядке строку  с указанной длинной n.
*/
func generateIdentifier(n int) string {
	b := make([]byte, n)
	for i := range b {
		switch i {
		case 0, 1:
			b[i] = AccountNumbers[rand.Intn(len(AccountNumbers))]
		case 2, 3, 4, 5:
			b[i] = AccountCharacters[rand.Intn(len(AccountCharacters))]
		default:
			b[i] = AccountNumbers[rand.Intn(len(AccountNumbers))]
		}
	}

	return string(b)
}

/*
Функция генерирует номер счера в формате IBAN (BY01WOJJ06438416642840916051).
*/
func GenerateAccountNumber() string {
	accId := generateIdentifier(AccountLength)

	return AccountPrefix + accId
}

/*
Функция проверяет номер счета на соответствие формату IBAN.
*/
func VerifyAccountNumber(accNumber string) error {
	reg, err := regexp.Compile(`(BY|SE|ST)[0-9]{2}[A-Z]{4}[0-9]{4}[0-9]{16}$`)
	if err != nil {
		return err
	}

	res := reg.MatchString(accNumber)

	if res {
		// log.Println("IBAN number is valid")
		return nil
	}

	return fmt.Errorf("not valid accont number")
}
