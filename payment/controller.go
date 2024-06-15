//go:generate mockgen -source=controller.go -destination=mocks/controller_mock.go -package=payment

package payment

import (
	"log"
	"time"
)

type HandlerFunc func() error

type Controller interface {
	Add(h HandlerFunc)
	Run()
}

// Controller responses to perform (Add()) and execute(Run()) number of functions.
type PaymentController struct {
	ps      Store
	workers chan HandlerFunc
}

func NewPaymentController(ps Store) PaymentController {
	pc := PaymentController{
		ps:      ps,
		workers: make(chan HandlerFunc, 100),
	}

	return pc
}

// This func perform transaction or batch of transactions.
func (pc PaymentController) Add(h HandlerFunc) {
	pc.workers <- h
}

// This func execute transaction or batch of transactions
// and manage the concurency
// If at list one transaction function will give an error - transaction will aborted.
func (pc PaymentController) Run(end chan bool) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("Recovered in f", r)
		}
	}()

	go func() {
		for {
			select {
			case worker, ok := <-pc.workers:
				{
					pc.ps.Lock()

					err := worker()

					pc.ps.Unlock()

					if err != nil {
						log.Printf("error occurred")

						continue
					} else {
						log.Println("task is done")
					}

					if ok && len(pc.workers) == 0 {
						end <- true

						return
					}
				}

			default:
				// log.Println("processing")
				time.Sleep(time.Millisecond * 50)
			}
		}
	}()

	time.Sleep(time.Millisecond * 10)
}
