# DERO Merchant Go SDK
Library with bindings for the [DERO Merchant REST API](https://merchant.dero.io/docs) for accepting DERO payments on a Golang backend.

## Requirements
- A store registered on your [DERO Merchant Dashboard](https://merchant.dero.io/dashboard) to receive an API Key and a Secret Key, required to send requests to the API.
- A Golang web server.

## Installation
`go get -u github.com/peppinux/dero-merchant-go-sdk`

## Usage
### Import
`import deromerchant "github.com/peppinux/dero-merchant-go-sdk"`

### Setup
```go
dmClient, err := deromerchant.NewClient(&deromerchant.ClientOptions{
        Scheme:     "https",                              // OPTIONAL. Default: https
        Host:       "merchant.dero.io",                   // OPTIONAL. Default: merchant.dero.io
        APIVersion: "v1",                                 // OPTIONAL. Default: v1
        APIKey:     "API_KEY_OF_YOUR_STORE_GOES_HERE",    // REQUIRED
        SecretKey:  "SECRET_KEY_OF_YOUR_STORE_GOES_HERE", // REQUIRED
})
if err != nil {
        // Bad options were provided.
        // dmClient is therefore nil and trying to call its methods will result in a nil pointer dereference.
        // As such, make sure to handle error.
        log.Fatalln("Error creating new DERO Merchant client", err)
}

_, err = dmClient.Ping()
if err != nil {
        // Server is offline OR bad Scheme/Host/APIVersion were provided.
        // Handle error.
        log.Fatalln("Error pinging DERO Merchant server", err)
}
```

### Create a Payment
```go
// p, err := dmClient.CreatePayment("USD", 1) // USD value will be converted to DERO
// p, err := dmClient.CreatePayment("EUR", 100) // Same thing goes for EUR and other currencies supported by the CoinGecko API V3
p, err := dmClient.CreatePayment("DERO", 10)
if err != nil {
        apiErr, ok := err.(*deromerchant.APIError)
        if ok {
                // Handle APIError
        }
        // Handle error
}

fmt.Printf("%+v\n", p)
/*
        Object of type *deromerchant.Payment
        &{
                PaymentID:09052ec05347670f76cc07ce9c88deb6ce2bf71105eb284fc805de83439ce980
                Status:pending
                Currency:DERO
                CurrencyAmount:10
                ExchangeRate:1
                DeroAmount:10.000000000000
                AtomicDeroAmount:10000000000000
                IntegratedAddress:dETiaFw6kkrSQ8BByamH8P9iNUCfYsLnUHTL9KftUBRZZEt44i86djtWr9sMpudU955wnLMwcv2YuNGDuTbQwrwDe2tRVt3yXdtCwhHBbXUz8jPtozbqcG7H6gLKgDnE66ZQ6wtEtJct5u
                CreationTime:2020-01-29 17:36:20.040876 +0000 UTC
                TTL:60
    	}
*/
```

### Get a Payment from its ID
```go
paymentID := "09052ec05347670f76cc07ce9c88deb6ce2bf71105eb284fc805de83439ce980"
p, err := dmClient.GetPayment(paymentID)
if err != nil {
        apiErr, ok := err.(*deromerchant.APIError)
        if ok {
                // Handle APIError
        }
        // Handle error
}

fmt.Printf("%+v\n", p)
/*
        Object of type *deromerchant.Payment
        &{
                PaymentID:09052ec05347670f76cc07ce9c88deb6ce2bf71105eb284fc805de83439ce980 
                Status:pending 
                Currency:DERO 
                CurrencyAmount:10 
                ExchangeRate:1 
                DeroAmount:10.000000000000 
                AtomicDeroAmount:10000000000000 
                IntegratedAddress:dETiaFw6kkrSQ8BByamH8P9iNUCfYsLnUHTL9KftUBRZZEt44i86djtWr9sMpudU955wnLMwcv2YuNGDuTbQwrwDe2tRVt3yXdtCwhHBbXUz8jPtozbqcG7H6gLKgDnE66ZQ6wtEtJct5u 
                CreationTime:2020-01-29 17:36:20.040876 +0000 UTC 
                TTL:48
        }
*/
```

### Get an array of Payments from their IDs
```go
paymentIDs := []string{
        "09052ec05347670f76cc07ce9c88deb6ce2bf71105eb284fc805de83439ce980",
        "38ad8cf0c5da388fe9b5b44f6641619659c99df6cdece60c6e202acd78e895b1",
}
// ps, err := dmClient.GetPayments([]string{"09052ec05347670f76cc07ce9c88deb6ce2bf71105eb284fc805de83439ce980", "38ad8cf0c5da388fe9b5b44f6641619659c99df6cdece60c6e202acd78e895b1"})
ps, err := dmClient.GetPayments(paymentIDs)
if err != nil {
        apiErr, ok := err.(*deromerchant.APIError)
        if ok {
                // Handle APIError
        }
        // Handle error
}

for _, p := range ps {
        fmt.Printf("%+v\n", p)
}
/*
        Objects of type *deromerchant.Payment
        &{
                PaymentID:38ad8cf0c5da388fe9b5b44f6641619659c99df6cdece60c6e202acd78e895b1 
                Status:paid 
                Currency:DERO 
                CurrencyAmount:10 
                ExchangeRate:1 
                DeroAmount:10.000000000000 
                AtomicDeroAmount:10000000000000 
                IntegratedAddress:dETiaFw6kkrSQ8BByamH8P9iNUCfYsLnUHTL9KftUBRZZEt44i86djtWr9sMpudU955wnLMwcv2YuNGDuTbQwrwDe2tRbFua6e8dW1xcFY6wPTBwHDPNN2eC4gdDNzhJWUL79pD2Tn2ksE 
                CreationTime:2020-01-16 16:49:59.131189 +0000 UTC 
                TTL:0
        }
        &{
                PaymentID:09052ec05347670f76cc07ce9c88deb6ce2bf71105eb284fc805de83439ce980 
                Status:pending 
                Currency:DERO 
                CurrencyAmount:10 
                ExchangeRate:1 
                DeroAmount:10.000000000000 
                AtomicDeroAmount:10000000000000 
                IntegratedAddress:dETiaFw6kkrSQ8BByamH8P9iNUCfYsLnUHTL9KftUBRZZEt44i86djtWr9sMpudU955wnLMwcv2YuNGDuTbQwrwDe2tRVt3yXdtCwhHBbXUz8jPtozbqcG7H6gLKgDnE66ZQ6wtEtJct5u 
                CreationTime:2020-01-29 17:36:20.040876 +0000 UTC 
                TTL:43
        }
*/
```

### Get an array of filtered Payments
_Not detailed because this endpoint was created for an internal usecase._
```go
resp, err := GetFilteredPayments(limit int, page int, sortBy string, orderBy string, statusFilter string, currencyFilter string)
if err != nil {
        apiErr, ok := err.(*deromerchant.APIError)
        if ok {
                // Handle APIError
        }
        // Handle error
}

fmt.Println("%+v\n", resp) // Object of type *deromerchant.GetFilteredPaymentsResponse
```

### Get Pay helper page URL
```go
paymentID := "09052ec05347670f76cc07ce9c88deb6ce2bf71105eb284fc805de83439ce980"
payURL := dmClient.GetPayHelperURL(paymentID)

fmt.Println(payURL) // https://merchant.dero.io/pay/09052ec05347670f76cc07ce9c88deb6ce2bf71105eb284fc805de83439ce980
```

### Verify Webhook Signature and Parse Webhook Request
When using Webhooks to receive Payment status updates, it is highly suggested to verify the HTTP requests are actually sent by the DERO Merchant server thorugh the X-Signature header.

This library offers a function for such verification, along with an utility function to parse the payload of the request into an fitting struct.

**Example using the _net/http_ standard library**
```go
const webhookSecretKey = "THE_WEBHOOK_SECRET_KEY_OF_YOUR_STORE_GOES_HERE"

http.HandleFunc("/dero_merchant_webhook_example", func(w http.ResponseWriter, r *http.Request) {
        valid, err := deromerchant.VerifyWebhookSignature(r, webhookSecretKey)
        if err != nil {
                // Don't trust the request.
                // Handle error.
                return
        }

        if !valid {
                // DON'T trust the request.
                // Make sure you copied the webhook secret key right.
                return
        }

        e, err := deromerchant.ParseWebhookRequest(r)
        if err != nil {
                // Handle error.
                return
        }
        fmt.Printf("%+v\n", e)
        /*
        	Object of type *deromerchant.PaymentUpdateEvent
        	&{
                    PaymentID:09052ec05347670f76cc07ce9c88deb6ce2bf71105eb284fc805de83439ce980
                    Status:paid
        	}
        */
})
```

**Example using the _net/http_ standard library and the single function VerifyAndParseWebhookRequest**
```go
const webhookSecretKey = "THE_WEBHOOK_SECRET_KEY_OF_YOUR_STORE_GOES_HERE"

http.HandleFunc("/dero_merchant_webhook_example", func(w http.ResponseWriter, r *http.Request) {
        valid, e, err := deromerchant.VerifyAndParseWebhookRequest(r, webhookSecretKey)
        if err != nil {
                // Error occured while verifying the signature or parsing the request
                return
        }

        if !valid {
                // Request signature not valid
                return
        }

        if e == nil {
                // Request not parsed
                return
        }

        fmt.Printf("%+v\n", e)
        /*
        	Object of type *deromerchant.PaymentUpdateEvent
        	&{
                    PaymentID:09052ec05347670f76cc07ce9c88deb6ce2bf71105eb284fc805de83439ce980
                    Status:paid
        	}
        */
})
```
