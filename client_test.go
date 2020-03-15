package deromerchant

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const (
	apiKey    = "bfe737bcdc5d8886a03be6e6c34c545d85ab8fa39052b9e3be36d3626c180a6f"
	secretKey = "b3cef2080cf82a010acba9bd00c9bd5797ec07767fbd7c08702a921d67c8155a"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		options        *ClientOptions
		expectedClient *Client
		expectError    bool
	}{
		// Client with default scheme, host and API version
		{
			options: &ClientOptions{
				APIKey:    apiKey,
				SecretKey: secretKey,
			},
			expectedClient: &Client{
				scheme:     defaultScheme,
				host:       defaultHost,
				apiVersion: defaultAPIVersion,
				apiKey:     apiKey,
				secretKey:  secretKey,
			},
			expectError: false,
		},
		// Client with custom scheme, host and API version
		{
			options: &ClientOptions{
				Scheme:     "http",
				Host:       "localhost:8080",
				APIVersion: "v1",
				APIKey:     apiKey,
				SecretKey:  secretKey,
			},
			expectedClient: &Client{
				scheme:     "http",
				host:       "localhost:8080",
				apiVersion: "v1",
				apiKey:     apiKey,
				secretKey:  secretKey,
			},
			expectError: false,
		},
		// Client with invalid URL
		{
			options: &ClientOptions{
				Scheme:     "1234",
				Host:       "localhost:8080:9090",
				APIVersion: "///v1",
				APIKey:     apiKey,
				SecretKey:  secretKey,
			},
			expectedClient: nil,
			expectError:    true,
		},
	}

	for _, test := range tests {
		c, err := NewClient(test.options)
		if err == nil {
			if test.expectError {
				t.Error("Expected error")
			}
		} else {
			if !test.expectError {
				t.Errorf("Error not expected. Got: %v\n", err)
				t.FailNow()
			}
		}

		if c != nil {
			if test.expectedClient == nil {
				t.Errorf("\nExpected nil Client.\nGot:\n%+v\n", *c)
				t.FailNow()
			}

			test.expectedClient.baseURL = c.baseURL
			test.expectedClient.HTTPClient = c.HTTPClient

			if *c != *test.expectedClient {
				t.Errorf("\nExpected Client:\n%+v\nGot:\n%+v\n", *&test.expectedClient, *c)
			}
		}
	}
}

func TestNewRequest(t *testing.T) {
	c, err := NewClient(&ClientOptions{
		Scheme:     "http",
		Host:       "localhost:8080",
		APIVersion: "v1",
		APIKey:     apiKey,
		SecretKey:  secretKey,
	})
	if err != nil {
		t.Fatal(err)
	}

	type Payload struct {
		AString string  `json:"aString,omitempty"`
		AnInt   int     `json:"anInt,omitempty"`
		AFloat  float64 `json:"aFloat,omitempty"`
	}

	tests := []struct {
		method      string
		endpoint    string
		queryParams map[string]interface{}
		payload     interface{}
		expectError bool
	}{
		{"GET", "/test", map[string]interface{}{"one": 1, "two": "two"}, nil, false},
		{"post", "/posttest", nil, &Payload{"one", 1, 1.0}, false},
		{http.MethodPost, "/anotherpost", map[string]interface{}{"i": "am", "a": "test"}, &Payload{"two", 2, 2.0}, false},
	}

	for _, test := range tests {
		req, err := c.NewRequest(test.method, test.endpoint, test.queryParams, test.payload)
		if err == nil {
			if test.expectError {
				t.Error("Expected error")
				t.FailNow()
			}
		}

		if req != nil {
			method := strings.ToUpper(test.method)
			if req.Method != method {
				t.Errorf("Expected method: %s. Got: %s\n", method, req.Method)
			}

			scheme := strings.ToLower(c.scheme)
			if req.URL.Scheme != scheme {
				t.Errorf("Expected scheme: %s. Got: %s\n", scheme, req.URL.Scheme)
			}

			if req.URL.Host != c.host {
				t.Errorf("Expected host: %s. Got: %s\n", c.host, req.URL.Host)
			}

			path := fmt.Sprintf("/api/%s%s", c.apiVersion, test.endpoint)
			if req.URL.Path != path {
				t.Errorf("Expected path: %s. Got: %s\n", path, req.URL.Path)
			}

			if test.queryParams != nil {
				for k, v := range test.queryParams {
					actualVal := req.URL.Query().Get(k)
					expectedVal := fmt.Sprintf("%v", v)
					if actualVal != expectedVal {
						t.Errorf("Expected query param %s = %v. Got %s = %v\n", k, v, k, actualVal)
					}
				}
			}

			if test.payload != nil {
				if req.Body == nil {
					t.Error("Expected request body not to be nil")
					t.FailNow()
				}

				b, err := req.GetBody()
				if err != nil {
					t.Fatal(err)
				}

				body, err := ioutil.ReadAll(b)
				if err != nil {
					t.Fatal(err)
				}

				payloadBytes, err := json.Marshal(&test.payload)
				if err != nil {
					t.Fatal(err)
				}

				if !bytes.Equal(body, payloadBytes) {
					t.Error("Expected request body to be json Marshaled payload")
				}

				if h := req.Header.Get("Content-Type"); h != "application/json" {
					t.Errorf("Expected header Content-Type: application/json. Got: %s\n", h)
				}

				if h := req.Header.Get("Accept"); h != "application/json" {
					t.Errorf("Expected header Accept: application/json. Got: %s\n", h)
				}
			} else {
				if req.Body != nil {
					t.Error("Expected request body to be nil")
					t.FailNow()
				}
			}

			if h := req.Header.Get("User-Agent"); h != "DeroMerchant_Client_Golang/1.0" {
				t.Errorf("Expected header User-Agent: DeroMerchant_Client_Golang/1.0. Got %s\n", h)
			}

			if h := req.Header.Get("X-API-Key"); h != c.apiKey {
				t.Errorf("Expected header X-API-Key: %s. Got %s\n", c.apiKey, h)
			}
		}
	}
}

func TestSendRequest(t *testing.T) {
	c, err := NewClient(&ClientOptions{
		Scheme:     "http",
		Host:       "localhost:8080",
		APIVersion: "v1",
		APIKey:     apiKey,
		SecretKey:  secretKey,
	})
	if err != nil {
		t.Fatal(err)
	}

	type Payload struct {
		ID      int    `json:"id,omitempty"`
		Message string `json:"message,omitempty"`
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { // Returns the Payload sent in the request as the response. Includes the ID if specified in the query params.
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		defer r.Body.Close()
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}

		var payload Payload
		err = json.Unmarshal(body, &payload)
		if err != nil {
			err := sendErrorResponse(w, http.StatusBadRequest, "Bad Request")
			if err != nil {
				t.Fatal(err)
			}
			return
		}

		excludeID := r.URL.Query().Get("exclude_id")
		if excludeID == "true" {
			payload.ID = 0
		}

		encodedResp, err := json.Marshal(&payload)
		if err != nil {
			t.Fatal(err)
		}

		w.WriteHeader(http.StatusOK)
		w.Write(encodedResp)
	}))
	defer ts.Close()
	c.baseURL = ts.URL // Override Client's base URL to point to fake server

	// Valid request with payload and query params
	q := map[string]interface{}{"exclude_id": true}
	p := &Payload{ID: 1, Message: "Hello world"}

	req, err := c.NewRequest(http.MethodPost, "/", q, p)
	if err != nil {
		t.Fatal(err)
	}

	var actualResp Payload
	err = c.SendRequest(req, &actualResp)
	if err != nil {
		t.Errorf("Expected error to be nil. Got: %v\n", err)
	}

	expectedResp := Payload{Message: p.Message}
	if actualResp != expectedResp {
		t.Errorf("\nExpected response:\n%+v\nGot:\n%+v\n", expectedResp, actualResp)
	}

	// Valid request with payload and no query params
	req, err = c.NewRequest(http.MethodPost, "/", nil, p)
	if err != nil {
		t.Fatal(err)
	}

	err = c.SendRequest(req, &actualResp)
	if err != nil {
		t.Errorf("Expected error to be nil. Got: %v\n", err)
	}

	expectedResp = *p
	if actualResp != expectedResp {
		t.Errorf("\nExpected response:\n%+v\nGot:\n%+v\n", expectedResp, actualResp)
	}

	// Request with no payload to get API error
	req, err = c.NewRequest(http.MethodPost, "/", nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	err = c.SendRequest(req, &actualResp)
	if err == nil {
		t.Error("Expected error")
		t.FailNow()
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Errorf("Expected API Error. Got %v\n", err)
		t.FailNow()
	}

	expectedAPIErr := APIError{
		Code:    http.StatusBadRequest,
		Message: "Bad Request",
	}

	if *apiErr != expectedAPIErr {
		t.Errorf("\nExpected API Error:\n%+v\nGot:\n%+v\n", expectedAPIErr, apiErr)
	}

	// Request with invalid method to get error
	req, err = c.NewRequest(http.MethodGet, "/", nil, p)
	if err != nil {
		t.Fatal(err)
	}

	err = c.SendRequest(req, &actualResp)

	expectedErr := fmt.Errorf("DeroMerchant Client: error %d returned by %s", http.StatusMethodNotAllowed, req.URL.String())
	if err.Error() != expectedErr.Error() {
		t.Errorf("Expected error: %v. Got %v", expectedErr, err)
	}

	// Server returns 404 error (not API Error)
	ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()
	c.baseURL = ts.URL // Override Client's base URL to point to fake server

	req, err = c.NewRequest(http.MethodGet, "/", nil, p)
	if err != nil {
		t.Fatal(err)
	}

	err = c.SendRequest(req, &actualResp)

	expectedErr = fmt.Errorf("DeroMerchant Client: error 404: page %s not found", req.URL.String())
	if err.Error() != expectedErr.Error() {
		t.Errorf("Expected error: %v. Got %v\n", expectedErr, err)
	}
}

func TestSendSignedRequest(t *testing.T) {
	c, err := NewClient(&ClientOptions{
		Scheme:     "http",
		Host:       "localhost:8080",
		APIVersion: "v1",
		APIKey:     apiKey,
		SecretKey:  secretKey,
	})
	if err != nil {
		t.Fatal(err)
	}

	type Payload struct {
		ID      int    `json:"id,omitempty"`
		Message string `json:"message,omitempty"`
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Body != nil {
			clientSignature := r.Header.Get("X-Signature")

			defer r.Body.Close()
			body, err := ioutil.ReadAll(r.Body)

			key, err := hex.DecodeString(c.secretKey)
			if err != nil {
				t.Fatal(err)
			}

			s, err := signMessage(body, key)
			if err != nil {
				t.Fatal(err)
			}

			signature := hex.EncodeToString(s)

			if clientSignature != signature {
				t.Errorf("Expected signature: %s. Got from client signature: %s\n", signature, clientSignature)
			}
		}
	}))
	defer ts.Close()
	c.baseURL = ts.URL // Override Client's base URL to point to fake server

	p := &Payload{ID: 1, Message: "Hello world"}

	req, err := c.NewRequest(http.MethodPost, "/", nil, p)
	if err != nil {
		t.Fatal(err)
	}

	err = c.SendSignedRequest(req, nil)
	if err != nil {
		t.Errorf("Expected error to be nil. Got: %v\n", err)
	}
}

func TestGetPayHelperURL(t *testing.T) {
	var (
		scheme     = "http"
		host       = "localhost:8080"
		apiVersion = "v1"
		PaymentID  = "bce3fffd584e4ada6bd11ef288b1f853d1ed815f6c8b8338bdec804e3885b871"
	)

	expectedURL := fmt.Sprintf("%s://%s/pay/%s", scheme, host, PaymentID)

	c, err := NewClient(&ClientOptions{
		Scheme:     scheme,
		Host:       host,
		APIVersion: apiVersion,
		APIKey:     apiKey,
		SecretKey:  secretKey,
	})
	if err != nil {
		t.Fatal(err)
	}

	url := c.GetPayHelperURL(PaymentID)

	if url != expectedURL {
		t.Errorf("Expected URL: %s Got: %s\n", expectedURL, url)
	}
}

func sendErrorResponse(w http.ResponseWriter, code int, message string) error {
	resp := &errorResponse{
		Error: &APIError{
			Code:    code,
			Message: message,
		},
	}

	encodedResp, err := json.Marshal(&resp)
	if err != nil {
		return err
	}

	w.WriteHeader(code)
	w.Write(encodedResp)
	return nil
}
