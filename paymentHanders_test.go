package main

import (
	"encoding/json"
	"testing"
	"time"

	//	"github.com/vennd/enu/log"
)

func TestGetPaymentsByAddress(t *testing.T) {
	go InitTesting()

	// Make URL from base URL
	var url = baseURL + "/payment/address/unittesting1"
	var result []map[string]interface{}

	var send = map[string]interface{}{
		"nonce": time.Now().Unix(),
	}

	assetJsonBytes, err := json.Marshal(send)
	if err != nil {
		t.Errorf("TestGetPaymentsByAddress(): Unable to create payload")
	}

	responseData, statusCode, err := DoEnuAPITesting("GET", url, assetJsonBytes)

	// deserialise the response if the status is 0
	if err != nil && statusCode != 0 {
		t.Errorf("Error in API call. Error: %s, statusCode: %d\n", err, statusCode)
	}

	if err := json.Unmarshal(responseData, &result); err != nil {
		t.Errorf("Error in API call. Unable to unmarshal responseData. Error: %s", err)
	}
}
