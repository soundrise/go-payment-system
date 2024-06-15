package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/soundrise/go-payment-system/payment"
)

func main() { //nolint:funlen
	log.Println("Payment System started!")

	/*
		init store (ps)
	*/
	ps := payment.NewPaymentSystem()

	/*
		init payment controller
	*/

	pc := payment.NewPaymentController(ps)

	/*
		init government
	*/
	var government []payment.Customer

	/*
		init customers
	*/
	var customers []payment.Customer

	// 1. Government
	gov := payment.NewCustomer("0", "GOVERNMENT", payment.AccountStateEmissionPrefix)
	// create default account (EMISSION)
	government = append(government, gov)

	gov = payment.NewCustomer("0", "GOVERNMENT", payment.AccountStateTerminatePrefix)
	// create default accounts (TERMINATE)
	government = append(government, gov)

	// 2. Customer 1
	customer1 := payment.NewCustomer("1", "Customer One", payment.AccountPrefix)
	customers = append(customers, customer1)
	// 3. Customer 2
	customer2 := payment.NewCustomer("2", "Customer Two", payment.AccountPrefix)
	customers = append(customers, customer2)

	customer2 = payment.NewCustomer("2", "Customer Two", payment.AccountPrefix)
	customers = append(customers, customer2)

	for _, c := range government {
		c := c

		pc.Add(func() error {
			return ps.CreateAccount(c, c.AccPrefix, payment.BYN, 0)
		})
	}

	// emit amount of money to special account
	pc.Add(func() error {
		return ps.Emit(2000000)
	})

	// emit again
	pc.Add(func() error {
		return ps.Emit(1500000)
	})

	// end emission

	// print store
	pc.Add(ps.PrintStoreJson)

	//
	for i, c := range customers {
		c := c
		i := i

		pc.Add(func() error {
			if i == 1 {
				return ps.CreateAccount(c, c.AccPrefix, payment.USD, 550)
			} else {
				return ps.CreateAccount(c, c.AccPrefix, payment.BYN, 1000)
			}
		})
	}

	// to debug
	pc.Add(ps.PrintStoreJson)

	/*
		Run payment service operations
	*/

	// get special accounts and print numbers
	pc.Add(func() error {
		_, err := ps.GetSpecialAccount(payment.AccountStateEmissionPrefix)
		if err != nil {
			return err
		}

		return nil
	})
	pc.Add(func() error {
		_, err := ps.GetSpecialAccount(payment.AccountStateTerminatePrefix)
		if err != nil {
			return err
		}

		return nil
	})

	// transactions

	// close acc for customer two
	pc.Add(func() error {
		acc, err := ps.FindAccount(customer2, payment.USD)
		if err != nil {
			return err
		}

		return ps.CloseAccount(acc)
	})

	pc.Add(ps.PrintStoreJson)

	// create new account in USD for customer 1
	pc.Add(func() error {
		return ps.CreateAccount(customer1, customer1.AccPrefix, payment.USD, 0)
	})

	pc.Add(ps.PrintStoreJson)

	// Transfer money from customer2 account to terminate account
	pc.Add(func() error {
		s, err := ps.FindAccount(customer2, payment.BYN)
		if err != nil {
			return err
		}

		return ps.Terminate(s, 200)
	})

	pc.Add(ps.PrintStoreJson)

	/*
		Transfer amount of money between two accounts
	*/

	// Transfer money from customer2 account to not valid account
	pc.Add(func() error {
		s, err1 := ps.FindAccount(customer2, payment.BYN)

		if err1 != nil {
			return err1
		}

		d := payment.Account{
			Num: "NOT_VALID",
		}

		return ps.Transfer(s, d, 250)
	})

	pc.Add(ps.PrintStoreJson)

	// Transfer money from customer2 account to customer1 account
	pc.Add(func() error {
		s, err1 := ps.FindAccount(customer2, payment.BYN)
		if err1 != nil {
			return err1
		}

		d, err2 := ps.FindAccount(customer1, payment.BYN)

		if err2 != nil {
			return err2
		}

		return ps.Transfer(s, d, 300)
	})

	pc.Add(ps.PrintStoreJson)

	// Transfer money from customer1 account to customer2 account by using JSON object
	pc.Add(func() error {
		s, err1 := ps.FindAccount(customer1, payment.BYN)
		if err1 != nil {
			return err1
		}

		d, err2 := ps.FindAccount(customer2, payment.BYN)

		if err2 != nil {
			return err2
		}

		t := payment.NewTransferData(s, d, 1)

		jsonData, err := json.Marshal(t)
		if err != nil {
			log.Printf("Error: cannot serialize JSON object. Source account: %s; Dest account %s\n", s.Num, d.Num)

			return fmt.Errorf("Error: cannot serialize JSON object")
		}

		return ps.TransferJson(jsonData)
	})

	pc.Add(ps.PrintStoreJson)

	e := make(chan bool)

	pc.Run(e)

	// time.Sleep(time.Millisecond * 10)

	select {
	case <-e:
		{
			log.Println("Payment System finished!")

			return
		}
	}
}
