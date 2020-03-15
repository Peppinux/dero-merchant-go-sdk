package deromerchant

import (
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const (
	validAPIKey    = "bfe737bcdc5d8886a03be6e6c34c545d85ab8fa39052b9e3be36d3626c180a6f"
	validSecretKey = "b3cef2080cf82a010acba9bd00c9bd5797ec07767fbd7c08702a921d67c8155a"

	invalidAPIKey    = "7960a3f4b301de77d773deb4b9cdd7f74ef096e18b8cdef27610f57b304fadcc"
	invalidSecretKey = "23b6000f6aba43f8a15a09995bd3c33ab76f9c58a6df02b4a3c229e0cb53c6fa"
)

func TestCreatePayment(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-API-Key") != validAPIKey {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		req, err := ioutil.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}
		defer r.Body.Close()

		reqSignature := r.Header.Get("X-Signature")
		key, err := hex.DecodeString(validSecretKey)
		if err != nil {
			t.Fatal(err)
		}

		s, err := signMessage(req, key)
		if err != nil {
			t.Fatal(err)
		}

		signature := hex.EncodeToString(s)
		if signature != reqSignature {
			err := sendErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
			if err != nil {
				t.Fatal(err)
			}
			return
		}

		var paymentReq createPaymentRequest
		err = json.Unmarshal(req, &paymentReq)
		if err != nil {
			t.Fatal(err)
		}

		paymentResp := &Payment{
			Currency:       paymentReq.Currency,
			CurrencyAmount: paymentReq.Amount,
		}

		resp, err := json.Marshal(&paymentResp)
		if err != nil {
			t.Fatal(err)
		}

		w.WriteHeader(http.StatusCreated)
		w.Write(resp)
	}))
	defer ts.Close()

	tests := []struct {
		apiKey         string
		secretKey      string
		currency       string
		amount         float64
		expectError    bool
		expectAPIError bool
	}{
		{apiKey: validAPIKey, secretKey: validSecretKey, currency: "DERO", amount: 10, expectError: false, expectAPIError: false},
		{apiKey: validAPIKey, secretKey: invalidSecretKey, currency: "USD", amount: 0.999, expectError: true, expectAPIError: true},
		{apiKey: invalidAPIKey, secretKey: validSecretKey, currency: "EUR", amount: 123.456, expectError: true, expectAPIError: false},
		{apiKey: invalidAPIKey, secretKey: invalidSecretKey, currency: "DERO", amount: 100, expectError: true, expectAPIError: false},
	}

	for _, test := range tests {
		c, err := NewClient(&ClientOptions{
			APIKey:    test.apiKey,
			SecretKey: test.secretKey,
		})
		if err != nil {
			t.Fatal(err)
		}
		c.baseURL = ts.URL // Override Client's base URL to point to fake server

		resp, err := c.CreatePayment(test.currency, test.amount)
		if err == nil {
			if test.expectError {
				t.Error("Expected error")
			}
			if test.expectAPIError {
				t.Error("Expected API Error")
			}

			if resp.Currency != test.currency || resp.CurrencyAmount != test.amount {
				t.Errorf("Expected currency: %s and amount: %f. Got: %s and %f\n", test.currency, test.amount, resp.Currency, resp.CurrencyAmount)
			}
		} else {
			if !test.expectError {
				t.Errorf("Error not expected. Got: %v\n", err)
				t.FailNow()
			}

			apiErr, ok := err.(*APIError)
			if ok && !test.expectAPIError {
				t.Errorf("API Error not expected. Got: %v\n", apiErr)
			}
			if !ok && test.expectAPIError {
				t.Errorf("Expected API Error. Got: %v\n", err)
			}
		}
	}
}

func TestGetPayment(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-API-Key") != validAPIKey {
			err := sendErrorResponse(w, http.StatusForbidden, "Forbidden")
			if err != nil {
				t.Fatal(err)
			}
			return
		}

		paymentID := strings.Split(r.URL.Path, "/")[2]
		if len(paymentID) == 0 {
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		}

		paymentResp := &Payment{
			PaymentID: paymentID,
		}

		resp, err := json.Marshal(&paymentResp)
		if err != nil {
			t.Fatal(err)
		}

		w.WriteHeader(http.StatusOK)
		w.Write(resp)
	}))
	defer ts.Close()

	tests := []struct {
		apiKey         string
		paymentID      string
		expectError    bool
		expectAPIError bool
	}{
		{apiKey: validAPIKey, paymentID: "f30d35de693c2f6cee02f3a099c9bf4cdb75d0b42c5527a0bae967a4521c56cb", expectError: false, expectAPIError: false},
		{apiKey: invalidAPIKey, paymentID: "ee4709228d728851919a93ae2e9a9d9f34f4ddce06908c951525b9b14ed110b4", expectError: true, expectAPIError: true},
		{apiKey: "", paymentID: "da5048bc58500d8ae7569578e573972bd80aae78c97d9378705319f5ec8b74b7", expectError: true, expectAPIError: true},
		{apiKey: validAPIKey, paymentID: "", expectError: true, expectAPIError: false},
	}

	for _, test := range tests {
		c, err := NewClient(&ClientOptions{
			APIKey: test.apiKey,
		})
		if err != nil {
			t.Fatal(err)
		}
		c.baseURL = ts.URL // Override Client's base URL to point to fake server

		resp, err := c.GetPayment(test.paymentID)
		if err == nil {
			if test.expectError {
				t.Error("Expected error")
			}
			if test.expectAPIError {
				t.Error("Expected API Error")
			}

			if resp.PaymentID != test.paymentID {
				t.Errorf("Expected Payment ID: %s. Got: %s\n", test.paymentID, resp.PaymentID)
			}
		} else {
			if !test.expectError {
				t.Errorf("Error not expected. Got: %v\n", err)
				t.FailNow()
			}

			apiErr, ok := err.(*APIError)
			if ok && !test.expectAPIError {
				t.Errorf("API Error not expected. Got: %v\n", apiErr)
			}
			if !ok && test.expectAPIError {
				t.Errorf("Expected API Error. Got: %v\n", err)
			}
		}
	}
}

func TestGetPayments(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-API-Key") != validAPIKey {
			err := sendErrorResponse(w, http.StatusForbidden, "Forbidden")
			if err != nil {
				t.Fatal(err)
			}
			return
		}

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}
		defer r.Body.Close()

		var paymentReq []string
		err = json.Unmarshal(body, &paymentReq)
		if err != nil {
			t.Fatal(err)
		}

		if len(paymentReq) == 0 {
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		}

		var paymentResp []*Payment
		for _, payid := range paymentReq {
			paymentResp = append(paymentResp, &Payment{
				PaymentID: payid,
			})
		}

		resp, err := json.Marshal(&paymentResp)
		if err != nil {
			t.Fatal(err)
		}

		w.WriteHeader(http.StatusOK)
		w.Write(resp)
	}))
	defer ts.Close()

	tests := []struct {
		apiKey         string
		paymentIDs     []string
		expectError    bool
		expectAPIError bool
	}{
		{apiKey: validAPIKey, expectError: false, expectAPIError: false, paymentIDs: []string{"f30d35de693c2f6cee02f3a099c9bf4cdb75d0b42c5527a0bae967a4521c56cb", "ee4709228d728851919a93ae2e9a9d9f34f4ddce06908c951525b9b14ed110b4", "da5048bc58500d8ae7569578e573972bd80aae78c97d9378705319f5ec8b74b7"}},
		{apiKey: invalidAPIKey, expectError: true, expectAPIError: true, paymentIDs: []string{"ee4709228d728851919a93ae2e9a9d9f34f4ddce06908c951525b9b14ed110b4", "da5048bc58500d8ae7569578e573972bd80aae78c97d9378705319f5ec8b74b7"}},
		{apiKey: "", expectError: true, expectAPIError: true, paymentIDs: []string{"da5048bc58500d8ae7569578e573972bd80aae78c97d9378705319f5ec8b74b7", "f30d35de693c2f6cee02f3a099c9bf4cdb75d0b42c5527a0bae967a4521c56cb"}},
		{apiKey: validAPIKey, expectError: true, expectAPIError: false, paymentIDs: []string{}},
	}

	for _, test := range tests {
		c, err := NewClient(&ClientOptions{
			APIKey: test.apiKey,
		})
		if err != nil {
			t.Fatal(err)
		}
		c.baseURL = ts.URL // Override Client's base URL to point to fake server

		resp, err := c.GetPayments(test.paymentIDs)
		if err == nil {
			if test.expectError {
				t.Error("Expected error")
			}
			if test.expectAPIError {
				t.Error("Expected API Error")
			}

			if expLen, actLen := len(test.paymentIDs), len(resp); expLen != actLen {
				t.Errorf("Expected %d payments. Got %d\n", expLen, actLen)
			}
			for i, respPayment := range resp {
				if respPayment.PaymentID != test.paymentIDs[i] {
					t.Errorf("Expected Payment ID: %s in position %d of slice. Got: %s", test.paymentIDs[i], i, respPayment.PaymentID)
					t.FailNow()
				}
			}
		} else {
			if !test.expectError {
				t.Errorf("Error not expected. Got: %v\n", err)
				t.FailNow()
			}

			apiErr, ok := err.(*APIError)
			if ok && !test.expectAPIError {
				t.Errorf("API Error not expected. Got: %v\n", apiErr)
			}
			if !ok && test.expectAPIError {
				t.Errorf("Expected API Error. Got: %v\n", err)
			}
		}
	}
}

// GetFilteredPayments not tested because not intended for public use. API route was created for internal uses.
