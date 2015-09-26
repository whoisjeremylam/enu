package main

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/vennd/enu/counterpartycrypto"
	//	"github.com/vennd/enu/log"
)

func TestWalletCreate(t *testing.T) {
	go InitTesting()

	// Make URL from base URL
	var url = baseURL + "/wallet"
	var wallet counterpartycrypto.CounterpartyWallet

	var send = map[string]interface{}{
		"nonce": time.Now().Unix(),
	}

	assetJsonBytes, err := json.Marshal(send)
	if err != nil {
		t.Errorf("TestWalletCreate(): Unable to create payload")
	}

	responseData, statusCode, err := DoEnuAPITesting("POST", url, assetJsonBytes)

	// deserialise the response if the status is 0
	if err != nil && statusCode != 0 {
		t.Errorf("Error in API call. Error: %s, statusCode: %d\n", err, statusCode)
	}

	if err := json.Unmarshal(responseData, &wallet); err != nil {
		t.Errorf("Error in API call. Unable to unmarshal responseData. Error: %s", err)
	}
}

func TestWalletBalance(t *testing.T) {
	go InitTesting()

	// Make URL from base URL
	var url = baseURL + "/wallet/balances/1GaZfh9VhxL4J8tBt2jrDvictZEKc8kcHx"
	var wallet counterpartycrypto.CounterpartyWallet

	var send = map[string]interface{}{
		"nonce": time.Now().Unix(),
	}

	assetJsonBytes, err := json.Marshal(send)
	if err != nil {
		t.Errorf("TestWalletBalance(): Unable to create payload")
	}

	responseData, statusCode, err := DoEnuAPITesting("GET", url, assetJsonBytes)

	// deserialise the response if the status is 0
	if err != nil && statusCode != 0 {
		t.Errorf("Error in API call. Error: %s, statusCode: %d\n", err, statusCode)
	}

	if err := json.Unmarshal(responseData, &wallet); err != nil {
		t.Errorf("Error in API call. Unable to unmarshal responseData. Error: %s", err)
	}
}
