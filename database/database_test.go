package database

import "testing"

func TestUpdateNonceByAccesKey(t *testing.T) {
	// Code to test update nonce
	
	nonce := int64(100000)

	accessKey := "71625888dc50d8915b871912aa6bbdce67fd1ed77d409ef1cf0726c6d9d7cf16"

	err := UpdateNonce(accessKey, nonce)

	if err != nil {
			t.Errorf(err.Error())
		}
}	

func TestGetNonceByAccesKey(t *testing.T) {
	// Code to test nonce check

	// non existing acessKey	
	//nonce := GetNonceByAccessKey("73a7b844c80c3c5cf532d1dd843321b1c733c0c67e5b5ab162ca283da4cfc182")
	nonce := GetNonceByAccessKey("71625888dc50d8915b871912aa6bbdce67fd1ed77d409ef1cf0726c6d9d7cf16")

	if nonce == 0 {
		t.Errorf("Unable to retrieve nonce. Expected != 0, got: %d\n", nonce)
	}
}


	