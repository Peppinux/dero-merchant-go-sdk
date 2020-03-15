package deromerchant

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPing(t *testing.T) {
	const (
		validAPIKey   = "bfe737bcdc5d8886a03be6e6c34c545d85ab8fa39052b9e3be36d3626c180a6f"
		invalidAPIKey = "7960a3f4b301de77d773deb4b9cdd7f74ef096e18b8cdef27610f57b304fadcc"
	)

	pingResp := &PingResponse{
		Ping: "pong",
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKey := r.Header.Get("X-API-Key")
		if apiKey == "" {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		if apiKey != validAPIKey {
			err := sendErrorResponse(w, http.StatusForbidden, "Forbidden")
			if err != nil {
				t.Fatal(err)
			}
			return
		}

		jsonResp, err := json.Marshal(&pingResp)
		if err != nil {
			t.Fatal(err)
		}

		w.WriteHeader(http.StatusOK)
		_, err = w.Write(jsonResp)
	}))
	defer ts.Close()

	tests := []struct {
		apiKey         string
		expectResp     bool
		expectError    bool
		expectAPIError bool
	}{
		{apiKey: validAPIKey, expectResp: true, expectError: false, expectAPIError: false},
		{apiKey: "", expectResp: false, expectError: true, expectAPIError: false},
		{apiKey: invalidAPIKey, expectResp: false, expectError: true, expectAPIError: true},
	}

	for _, test := range tests {
		c, err := NewClient(&ClientOptions{
			APIKey: test.apiKey,
		})
		if err != nil {
			t.Fatal(err)
		}
		c.baseURL = ts.URL // Override Client's base URL to point to fake server

		resp, err := c.Ping()
		if err == nil {
			if test.expectError {
				t.Error("Expected error")
			}

			if test.expectAPIError {
				t.Error("Expected API Error")
			}

			if test.expectResp {
				if *resp != *pingResp {
					t.Errorf("Expected ping response. Got: %v\n", resp)
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
