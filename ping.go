package deromerchant

import (
	"net/http"
)

// PingResponse is a struct that holds the unmarshalled JSON response of a Ping request.
type PingResponse struct {
	Ping string `json:"ping"`
}

// Ping sends a GET request to the /ping endpoint and returns the response as a PingResponse.
// It is used to check whether server is online or offline. Is the second case, it may also be due to bad Scheme/Host/APIVersion client options.
func (c *Client) Ping() (*PingResponse, error) {
	req, err := c.NewRequest(http.MethodGet, "/ping", nil, nil)
	if err != nil {
		return nil, err
	}

	var resp *PingResponse
	err = c.SendRequest(req, &resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
