// Thanks to http://www.jokecamp.com/blog/examples-of-creating-base64-hashes-using-hmac-sha256-in-different-languages/#go

package main

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"

	"github.com/gorilla/securecookie"
)

func ComputeHmac512(message []byte, secret string) string {
	key := []byte(secret)
	h := hmac.New(sha512.New, key)
	h.Write(message)
	return hex.EncodeToString(h.Sum(nil))
}

// Generates a 64 character random string that can be used as a secret or an access key
func GenerateKey() string {
	return hex.EncodeToString(securecookie.GenerateRandomKey(32))
}

func GeneratePaymentId() string {
	return hex.EncodeToString(securecookie.GenerateRandomKey(16))
}
