package deromerchant

import "fmt"

// APIError represents the error object returned by the server when a request fails.
type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("DeroMerchant Client: API Error %d: %s", e.Code, e.Message)
}
