package deromerchant

import (
	"crypto/hmac"
	"crypto/sha256"
)

func signMessage(msg, key []byte) ([]byte, error) {
	mac := hmac.New(sha256.New, key)
	_, err := mac.Write(msg)
	if err != nil {
		return nil, err
	}

	return mac.Sum(nil), nil
}

func validMAC(message, messageMAC, key []byte) (bool, error) {
	msgSignature, err := signMessage(message, key)
	if err != nil {
		return false, err
	}

	return hmac.Equal(messageMAC, msgSignature), nil
}
