package deromerchant

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Client is a struct that holds all the information needed to perform a request to the DERO Merchant REST API.
// Client also has methods that make use of such information to perform said requests and return the response (or error) in fitting structs.
// Use NewClient to create a new Client.
type Client struct {
	scheme     string
	host       string
	apiVersion string
	baseURL    string
	HTTPClient *http.Client

	apiKey    string
	secretKey string
}

// ClientOptions is a struct that holds the required options for the initialization of a new Client.
// ClientOptions have to be passed as an argument of the NewClient function.
// Scheme, Host and APIVersion are optional. If not provided, they will be filled with default values.
type ClientOptions struct {
	Scheme     string
	Host       string
	APIVersion string

	APIKey    string
	SecretKey string
}

const (
	defaultScheme     = "https"
	defaultHost       = "merchant.dero.io"
	defaultAPIVersion = "v1"
)

// NewClient returns a new Client.
// ClientOptions API Key and Secret Key are required. Scheme, Host and APIVersion will be filled with default values if not provided.
func NewClient(o *ClientOptions) (*Client, error) {
	c := &Client{
		scheme:     o.Scheme,
		host:       o.Host,
		apiVersion: o.APIVersion,
		apiKey:     o.APIKey,
		secretKey:  o.SecretKey,
	}

	if c.scheme == "" {
		c.scheme = defaultScheme
	}
	if c.host == "" {
		c.host = defaultHost
	}
	if c.apiVersion == "" {
		c.apiVersion = defaultAPIVersion
	}

	c.baseURL = fmt.Sprintf("%s://%s/api/%s", c.scheme, c.host, c.apiVersion)

	_, err := url.ParseRequestURI(c.baseURL)
	if err != nil {
		return nil, err
	}

	c.HTTPClient = &http.Client{
		Timeout: time.Second * 10,
	}

	return c, nil
}

// NewRequest returns a new request ready to be sent with SendRequest or SendSignedRequest.
func (c *Client) NewRequest(method, endpoint string, queryParams map[string]interface{}, payload interface{}) (*http.Request, error) {
	url := c.baseURL + endpoint

	var body io.Reader
	if payload != nil {
		b, err := json.Marshal(&payload)
		if err != nil {
			return nil, err
		}
		body = bytes.NewBuffer(b)
	}

	method = strings.ToUpper(method)
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	for k, v := range queryParams {
		val := fmt.Sprintf("%v", v)
		q.Add(k, val)
	}
	req.URL.RawQuery = q.Encode()

	req.Header.Set("User-Agent", "DeroMerchant_Client_Golang/1.0")
	req.Header.Set("X-API-Key", c.apiKey)
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
	}

	return req, nil
}

type errorResponse struct {
	Error *APIError `json:"error"`
}

// SendRequest sends a request to the API.
func (c *Client) SendRequest(req *http.Request, respBody interface{}) error {
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		var errResp errorResponse
		err := json.Unmarshal(b, &errResp)
		if err != nil {
			if resp.StatusCode == http.StatusNotFound {
				return fmt.Errorf("DeroMerchant Client: error 404: page %s not found", req.URL.String())
			}

			return fmt.Errorf("DeroMerchant Client: error %d returned by %s", resp.StatusCode, req.URL.String())
		}

		return errResp.Error
	}

	if respBody != nil {
		err = json.Unmarshal(b, respBody)
		if err != nil {
			return err
		}
	}

	return nil
}

// SendSignedRequest sends a signed request to the API.
// The signature is generated using the Secret Key to create a MAC of the request body.
// Signature is then sent along with the request in the X-Sginature header.
func (c *Client) SendSignedRequest(req *http.Request, respBody interface{}) error {
	if req.Body != nil {
		b, err := req.GetBody()
		if err != nil {
			return err
		}

		body, err := ioutil.ReadAll(b)
		if err != nil {
			return err
		}

		key, err := hex.DecodeString(c.secretKey)
		if err != nil {
			return err
		}

		s, err := signMessage(body, key)
		if err != nil {
			return err
		}

		signature := hex.EncodeToString(s)
		req.Header.Set("X-Signature", signature)
	}

	return c.SendRequest(req, respBody)
}

// GetPayHelperURL returns the URL of the Pay helper page of paymentID.
// Example: https://merchant.dero.io/pay/38ad8cf0c5da388fe9b5b44f6641619659c99df6cdece60c6e202acd78e895b1
func (c *Client) GetPayHelperURL(paymentID string) string {
	return fmt.Sprintf("%s://%s/pay/%s", c.scheme, c.host, paymentID)
}
