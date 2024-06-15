package payment_test

import (
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/soundrise/go-payment-system/payment"
	mocks "github.com/soundrise/go-payment-system/payment/mocks"
	"github.com/stretchr/testify/assert"
)

func TestPaymentSystem_CreateAccount(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	// gov := payment.NewCustomer("0", "GOVERNMENT", payment.AccountStateEmissionPrefix)
	customer1 := payment.NewCustomer("1", "Customer One", payment.AccountPrefix)

	expectedCurrencyCode := payment.BYN
	// expectedUSDCurrencyCode := payment.USD
	expectedAmount := float32(5.0)
	expectedNegAmount := float32(-5.0)

	invalidCustomer := payment.Customer{
		Name: "inval",
	}

	testCases := []struct {
		desc             string
		mockExpectations func(s *mocks.MockStore, c *mocks.MockController)
		customer         payment.Customer
		currencyCode     string
		amount           float32
		expectedErr      error
	}{
		{
			desc:         "given a valid customer's params, create valid account and nil returned",
			customer:     customer1,
			currencyCode: payment.BYN,
			amount:       float32(0.0),
			mockExpectations: func(s *mocks.MockStore, c *mocks.MockController) {
				s.EXPECT().CreateAccount(customer1, customer1.AccPrefix, expectedCurrencyCode, expectedAmount).Return(nil).Times(1)

				c.EXPECT().Add(s.CreateAccount(customer1, customer1.AccPrefix, expectedCurrencyCode, expectedAmount)).AnyTimes()
			},
		},

		{
			desc:         "given a valid customer's params, except amount < 0, error returned",
			customer:     customer1,
			currencyCode: expectedCurrencyCode,
			amount:       expectedNegAmount,
			mockExpectations: func(s *mocks.MockStore, c *mocks.MockController) {
				s.EXPECT().CreateAccount(customer1, customer1.AccPrefix, expectedCurrencyCode, expectedNegAmount).Return(fmt.Errorf("something went wrong")).Times(1)

				c.EXPECT().Add(s.CreateAccount(customer1, customer1.AccPrefix, expectedCurrencyCode, expectedNegAmount)).AnyTimes()
			},
			expectedErr: fmt.Errorf("Creating account  is imposible: amount < 0 \n"),
		},
		{
			desc:         "given a invalid customer's params, error returned",
			customer:     invalidCustomer,
			currencyCode: expectedCurrencyCode,
			amount:       expectedAmount,
			mockExpectations: func(s *mocks.MockStore, c *mocks.MockController) {
				s.EXPECT().CreateAccount(invalidCustomer, invalidCustomer.AccPrefix, expectedCurrencyCode, expectedAmount).Return(fmt.Errorf("something went wrong")).Times(1)

				c.EXPECT().Add(s.CreateAccount(invalidCustomer, invalidCustomer.AccPrefix, expectedCurrencyCode, expectedAmount)).AnyTimes()
			},
			expectedErr: fmt.Errorf("Creating account is imposible: without client ID\n"),
		},
	}

	for _, tt := range testCases {
		tt := tt
		t.Run(tt.desc, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockPaymentSystem := mocks.NewMockStore(mockCtrl)
			mockPaymentController := mocks.NewMockController(mockCtrl)

			if tt.mockExpectations != nil {
				tt.mockExpectations(mockPaymentSystem, mockPaymentController)
			}

			ps := payment.NewPaymentSystem()
			pc := payment.NewPaymentController(ps)

			var gotError error

			pc.Add(func() error {
				gotError = ps.CreateAccount(tt.customer, tt.customer.AccPrefix, tt.currencyCode, tt.amount)

				return gotError
			})

			pc.Run()

			if tt.expectedErr != nil {
				assert.EqualError(t, gotError, tt.expectedErr.Error())
			} else {
				assert.NoError(t, gotError)
			}
		})
	}
}

func TestPaymentSystem_Emit(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	gov := payment.NewCustomer("0", "GOVERNMENT", payment.AccountStateEmissionPrefix)
	// customer1 := payment.NewCustomer("1", "Customer One", payment.AccountPrefix)

	expectedCurrencyCode := payment.BYN
	// expectedUSDCurrencyCode := payment.USD
	expectedAmount := float32(5000.0)
	expectedNegAmount := float32(-5000.0)

	invalidCustomer := payment.Customer{
		Id:        "33",
		Name:      "inval",
		AccPrefix: payment.AccountPrefix,
	}

	testCases := []struct {
		desc             string
		mockExpectations func(s *mocks.MockStore, c *mocks.MockController)
		customer         payment.Customer
		currencyCode     string
		amount           float32
		expectedErr      error
	}{
		{
			desc:         "given a valid customer's params to create special emissive account and nil returned",
			customer:     gov,
			currencyCode: expectedCurrencyCode,
			amount:       expectedAmount,
			mockExpectations: func(s *mocks.MockStore, c *mocks.MockController) {
				s.EXPECT().CreateAccount(gov, gov.AccPrefix, expectedCurrencyCode, expectedAmount).Return(nil).Times(1)

				c.EXPECT().Add(s.CreateAccount(gov, gov.AccPrefix, expectedCurrencyCode, expectedAmount)).AnyTimes()

				s.EXPECT().Emit(expectedAmount).Return(nil).Times(1)

				c.EXPECT().Add(s.Emit(expectedAmount)).AnyTimes()
			},
		},
		{
			desc:         "given a valid customer's params except amount < 0, error returned",
			customer:     gov,
			currencyCode: expectedCurrencyCode,
			amount:       expectedNegAmount,
			mockExpectations: func(s *mocks.MockStore, c *mocks.MockController) {
				s.EXPECT().CreateAccount(gov, gov.AccPrefix, expectedCurrencyCode, expectedAmount).Return(nil).Times(1)

				c.EXPECT().Add(s.CreateAccount(gov, gov.AccPrefix, expectedCurrencyCode, expectedAmount)).AnyTimes()

				s.EXPECT().Emit(expectedNegAmount).Return(fmt.Errorf("something went wrong")).Times(1)

				c.EXPECT().Add(s.Emit(expectedNegAmount)).AnyTimes()
			},
			expectedErr: fmt.Errorf("Emission is imposible: amount <= 0 \n"),
		},
		{
			desc:         "given a invalid special emit account, emit error returned",
			customer:     invalidCustomer,
			currencyCode: expectedCurrencyCode,
			amount:       expectedAmount,
			mockExpectations: func(s *mocks.MockStore, c *mocks.MockController) {
				s.EXPECT().CreateAccount(invalidCustomer, invalidCustomer.AccPrefix, expectedCurrencyCode, expectedAmount).Return(nil).Times(1)

				c.EXPECT().Add(s.CreateAccount(invalidCustomer, invalidCustomer.AccPrefix, expectedCurrencyCode, expectedAmount)).AnyTimes()

				s.EXPECT().Emit(expectedAmount).Return(fmt.Errorf("something went wrong")).Times(1)

				c.EXPECT().Add(s.Emit(expectedAmount)).AnyTimes()
			},
			expectedErr: fmt.Errorf("special account not found SE\n"),
		},
	}

	for _, tt := range testCases {
		tt := tt
		t.Run(tt.desc, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockPaymentSystem := mocks.NewMockStore(mockCtrl)
			mockPaymentController := mocks.NewMockController(mockCtrl)

			if tt.mockExpectations != nil {
				tt.mockExpectations(mockPaymentSystem, mockPaymentController)
			}

			ps := payment.NewPaymentSystem()
			pc := payment.NewPaymentController(ps)

			// var gotError error
			pc.Add(func() error {
				gotError := ps.CreateAccount(tt.customer, tt.customer.AccPrefix, tt.currencyCode, tt.amount)

				return gotError
			})

			pc.Run()

			var gotError error

			pc.Add(func() error {
				gotError = ps.Emit(tt.amount)

				return gotError
			})
			pc.Run()

			if tt.expectedErr != nil {
				assert.EqualError(t, gotError, tt.expectedErr.Error())
			} else {
				assert.NoError(t, gotError)
			}
		})
	}
}
