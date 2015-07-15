package main

import "testing"
import "github.com/vennd/enulib"

func TestGetNonceByAccesKey(t *testing.T) {
	// Code to test nonce check
	nonce := enulib.GetNonceByAccessKey("73a7b844c80c3c5cf532d1dd843321b1c733c0c67e5b5ab162ca283da4cfc182")
	//	log.Printf("Nonce for %s is %d", "73a7b844c80c3c5cf532d1dd843321b1c733c0c67e5b5ab162ca283da4cfc182", nonce)
}
