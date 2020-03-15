package deromerchant

import (
	"fmt"
	"net/http"
	"time"
)

// Payment represents a Payment created on/fetched from DERO Merchant server.
// It holds the the unmarshalled JSON response of a CreatePayment/GetPayment request.
type Payment struct {
	PaymentID         string    `json:"paymentID"`
	Status            string    `json:"status"`
	Currency          string    `json:"currency"`
	CurrencyAmount    float64   `json:"currencyAmount"`
	ExchangeRate      float64   `json:"exchangeRate"`
	DeroAmount        string    `json:"deroAmount"`
	AtomicDeroAmount  uint64    `json:"atomicDeroAmount"`
	IntegratedAddress string    `json:"integratedAddress"`
	CreationTime      time.Time `json:"creationTime"`
	TTL               int       `json:"ttl"`
}

type createPaymentRequest struct {
	Currency string  `json:"currency"`
	Amount   float64 `json:"amount"`
}

// CreatePayment sends a POST request to the /payment endpoint and returns the response as a Payment.
// It is used to create a new Payment on the DERO Merchant server and receive its details.
// Function can return an APIError if the request makes it to the server but something goes wrong.
func (c *Client) CreatePayment(currency string, amount float64) (*Payment, error) {
	payload := &createPaymentRequest{
		Currency: currency,
		Amount:   amount,
	}

	req, err := c.NewRequest(http.MethodPost, "/payment", nil, payload)
	if err != nil {
		return nil, err
	}

	var resp *Payment
	err = c.SendSignedRequest(req, &resp)
	if err != nil {
		apiErr, ok := err.(*APIError)
		if ok {
			return nil, apiErr
		}

		return nil, err
	}

	return resp, nil
}

// GetPayment sends a GET request to the /payment/:paymentID endpoint and returns the response as a Payment.
// It is used to get a Payment's details from its Payment ID from the DERO Merchant server.
// Function can return an APIError if the request makes it to the server but something goes wrong.
func (c *Client) GetPayment(paymentID string) (*Payment, error) {
	endpoint := fmt.Sprintf("/payment/%s", paymentID)
	req, err := c.NewRequest(http.MethodGet, endpoint, nil, nil)
	if err != nil {
		return nil, err
	}

	var resp *Payment
	err = c.SendRequest(req, &resp)
	if err != nil {
		apiErr, ok := err.(*APIError)
		if ok {
			return nil, apiErr
		}

		return nil, err
	}

	return resp, nil
}

// GetPayments sends a POST request to the /payments endpoint and returns the response as a slice of Payment(s).
// It is used to get multiple Payments' details from their Paymnet IDs from the DERO Merchant server.
// Function can return an APIError if the request makes it to the server but something goes wrong.
func (c *Client) GetPayments(paymentIDs []string) ([]*Payment, error) {
	req, err := c.NewRequest(http.MethodPost, "/payments", nil, paymentIDs)
	if err != nil {
		return nil, err
	}

	var resp []*Payment
	err = c.SendRequest(req, &resp)
	if err != nil {
		apiErr, ok := err.(*APIError)
		if ok {
			return nil, apiErr
		}

		return nil, err
	}

	return resp, nil
}

// GetFilteredPaymentsResponse is a struct that holds the unmarshalled JSON response of a GetFilteredPayments request.
type GetFilteredPaymentsResponse struct {
	Limit         int        `json:"limit"`
	Page          int        `json:"page"`
	TotalPayments int        `json:"totalPayments"`
	TotalPages    int        `json:"totalPages"`
	Payments      []*Payment `json:"payments"`
}

// GetFilteredPayments sends a GET request to the /payments endpoint and returns the response as a GetFilteredPaymentsResponse.
// It is used internally by DERO Merchant and should be of particurarly useful outside.
// It gets multiple Payments' details based on filters from the DERO Merchant server.
// Function can return an APIError if the request makes it to the server but something goes wrong.
func (c *Client) GetFilteredPayments(limit, page int, sortBy, orderBy, statusFilter, currencyFilter string) (*GetFilteredPaymentsResponse, error) {
	queryParams := map[string]interface{}{
		"limit":    limit,
		"page":     page,
		"sort_by":  sortBy,
		"order_by": orderBy,
		"status":   statusFilter,
		"currency": currencyFilter,
	}

	req, err := c.NewRequest(http.MethodGet, "/payments", queryParams, nil)
	if err != nil {
		return nil, err
	}

	var resp *GetFilteredPaymentsResponse
	err = c.SendRequest(req, &resp)
	if err != nil {
		apiErr, ok := err.(*APIError)
		if ok {
			return nil, apiErr
		}

		return nil, err
	}

	return resp, nil
}
