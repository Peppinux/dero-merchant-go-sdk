package deromerchant

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
)

var (
	// ErrNoWebhookSignature is returned by VerifyWebhookSignature if webhook request has no X-Signature header.
	ErrNoWebhookSignature = errors.New("DeroMerchant: webhook request has no signature header")
	// ErrInvalidSignature is returned by VerifyWebhookSignature is the signature provided in the X-Signature header does not match the signature generated using the Webhook Secret Key provided as a parameter.
	ErrInvalidSignature = errors.New("DeroMerchant: webhook request has invalid signature")
)

// VerifyWebhookSignature returns whether the signature of a webhook request payload, sent in the X-Signature header, is valid or not.
// Function can return defined errrors ErrNoWebhookSignature or ErrInvalidSignature.
// Requests not verified by this function should not be considered valid.
func VerifyWebhookSignature(req *http.Request, webhookSecretKey string) (bool, error) {
	h := req.Header.Get("X-Signature")
	if h == "" {
		return false, ErrNoWebhookSignature
	}

	signature, err := hex.DecodeString(h)
	if err != nil {
		return false, err
	}

	key, err := hex.DecodeString(webhookSecretKey)
	if err != nil {
		return false, err
	}

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return false, err
	}

	req.Body = ioutil.NopCloser(bytes.NewBuffer(body)) // Make body readable again

	valid, err := validMAC(body, signature, key)
	if err != nil {
		return false, err
	}

	if !valid {
		return false, ErrInvalidSignature
	}

	return true, nil
}

// PaymentUpdateEvent is a struct that holds the unmarshalled JSON data of a webhook request.
type PaymentUpdateEvent struct {
	PaymentID string `json:"paymentID,omitempty"`
	Status    string `json:"status,omitempty"`
}

// ParseWebhookRequest parses the body of a webhook request and returns it as a PaymentUpdateEvent object.
// It should be used after the request has been verified by VerifyWebhookSignature.
func ParseWebhookRequest(req *http.Request) (*PaymentUpdateEvent, error) {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}

	req.Body = ioutil.NopCloser(bytes.NewBuffer(body)) // Make body readable again

	var e *PaymentUpdateEvent
	err = json.Unmarshal(body, &e)
	if err != nil {
		return nil, err
	}

	return e, nil
}

// VerifyAndParseWebhookRequest both verifies and parses a webhook request.
// It is an alternative to calling VerifyWebhookSignature and ParseWebhookRequest individually.
func VerifyAndParseWebhookRequest(req *http.Request, webhookSecretKey string) (bool, *PaymentUpdateEvent, error) {
	valid, err := VerifyWebhookSignature(req, webhookSecretKey)
	if err != nil {
		return valid, nil, err
	}

	e, err := ParseWebhookRequest(req)
	if err != nil {
		return valid, nil, err
	}

	return valid, e, nil
}
