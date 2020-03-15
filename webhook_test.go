package deromerchant

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func createWebhookRequest(url string, e *PaymentUpdateEvent, webhookSecretKey string) (*http.Request, error) {
	json, err := json.Marshal(&e)
	if err != nil {
		return nil, err
	}
	body := bytes.NewBuffer(json)

	req, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		return nil, err
	}

	k, err := hex.DecodeString(webhookSecretKey)
	if err != nil {
		return nil, err
	}

	s, err := signMessage(json, k)
	if err != nil {
		return nil, err
	}

	signature := hex.EncodeToString(s)
	req.Header.Add("X-Signature", signature)

	return req, nil
}

func TestVerifyAndParseWebhookRequest(t *testing.T) {
	const (
		validWebhookSecretKey   = "010f2b45384c57bd388bccb520722abd8d5a61f66ca71fcd25bf7942d067ca73"
		invalidWebhookSecretKey = "1e9a0eefcff11530a1bc247672e9ebcb712fc6ab82e0b54b0e586c8adc33b0c0"
	)

	var (
		actualValid bool
		actualEvent *PaymentUpdateEvent
		actualErr   error
	)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		actualValid, actualEvent, actualErr = VerifyAndParseWebhookRequest(r, validWebhookSecretKey)
	}))
	defer ts.Close()

	tests := []struct {
		e             *PaymentUpdateEvent
		key           string
		expectedValid bool
		expectError   bool
	}{
		{e: &PaymentUpdateEvent{PaymentID: "6c8dd967897d8c46879d75236027f4791816146bed38a259a1dbdb8e047c10b4", Status: "paid"}, key: validWebhookSecretKey, expectedValid: true, expectError: false},
		{e: &PaymentUpdateEvent{PaymentID: "7b28c71eb2a0880bcb60aa7b9090da64e662608706049aa11ba724043dcf0d3e", Status: "expired"}, key: invalidWebhookSecretKey, expectedValid: false, expectError: true},
		{e: nil, key: invalidWebhookSecretKey, expectedValid: false, expectError: true},
		{e: nil, key: "", expectedValid: false, expectError: true},
	}

	httpClient := &http.Client{
		Timeout: time.Second,
	}

	for _, test := range tests {
		req, err := createWebhookRequest(ts.URL, test.e, test.key)
		if err != nil {
			t.Fatal(err)
		}

		_, err = httpClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}

		if actualValid != test.expectedValid {
			t.Errorf("Expected valid: %t. Got: %t\n", test.expectedValid, actualValid)
		}

		if actualErr == nil {
			if test.expectError {
				t.Error("Expected error")
				t.FailNow()
			}

			if *actualEvent != *test.e {
				t.Errorf("\nExpected event:\n%+v\nGot:\n%+v\n", *test.e, *actualEvent)
			}
		} else {
			if !test.expectError {
				t.Errorf("Error not expected. Got: %v\n", actualErr)
			}
		}
	}
}
