package database

import "testing"

func TestGetNonceByAccesKey(t *testing.T) {
	// Code to test nonce check
	nonce := GetNonceByAccessKey("73a7b844c80c3c5cf532d1dd843321b1c733c0c67e5b5ab162ca283da4cfc182")

	if nonce == 0 {
		t.Errorf("Unable to retrieve nonce. Expected != 0, got: %d\n", nonce)
	}
}
