package database

import (
	"github.com/vennd/enu/consts"
	"testing"
)

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

func TestUserKeyExists(t *testing.T) {
	exists := UserKeyExists("71625888dc50d8915b871912aa6bbdce67fd1ed77d409ef1cf0726c6d9d7cf16")
	notExists := UserKeyExists("narebeko")

	if exists == false {
		t.Errorf("User test key doesn't exist. Expected: true, got: %t\n", exists)
	}

	if notExists == true {
		t.Errorf("User test key doesn't exist. Expected: false, got: %t\n", exists)
	}
}

func TestCreateUserKey(t *testing.T) {
	// Create a user key
	key, _, err := CreateUserKey(777, "", consts.CounterpartyBlockchainId, "", "")
	if err != nil {
		t.Errorf("Unable to create user: %s\n", err.Error())
	}

	// Update user key with all possible statuses
	for _, value := range consts.AccessKeyStatuses {
		err2 := UpdateUserKeyStatus(key, value)
		if err2 != nil {
			t.Errorf("Unable to update userKey status: %s\n", err2.Error())
		}

		status := GetStatusByUserKey(key)
		if status != value {
			t.Errorf("User key status not set correctly. Expected: %s, got: %s\n", value, status)
		}
	}

	// Disable the user key that we created previously
	err3 := UpdateUserKeyStatus(key, consts.AccessKeyInvalidStatus)
	if err3 != nil {
		t.Errorf("Unable to update userKey status: %s\n", err3.Error())
	}

	// Attempt to set status to an invalid value
	err4 := UpdateUserKeyStatus(key, "this_should_not_work")
	if err4 == nil {
		t.Errorf("userKey status could be updated to an invalid value: %s\n", "this_should_not_work")
	}

	// Get the blockchainId and check it was set correctly
	blockchainId := GetBlockchainIdByUserKey(key)
	if blockchainId != consts.CounterpartyBlockchainId {
		t.Errorf("User key status not set correctly. Expected: %s, got: %s\n", consts.CounterpartyBlockchainId, blockchainId)
	}
}
