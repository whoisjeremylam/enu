package main

import (
	"encoding/json"
	"testing"
	"time"
)

func TestGetDividend(t *testing.T) {
	go InitTesting()

	// Make URL from base URL
	var url = baseURL + "/asset/dividend/556dfb84e6f08480f066d1719cefba25"
	var result map[string]interface{}

	var send = map[string]interface{}{
		"nonce": time.Now().Unix(),
	}

	assetJsonBytes, err := json.Marshal(send)
	if err != nil {
		t.Errorf("TestGetPayment(): Unable to create payload")
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

func TestGetAsset(t *testing.T) {
	go InitTesting()

	// Make URL from base URL
	var url = baseURL + "/asset/470cf33c0069f14c8e57aaf5823605da"
	var result map[string]interface{}

	var send = map[string]interface{}{
		"nonce": time.Now().Unix(),
	}

	assetJsonBytes, err := json.Marshal(send)
	if err != nil {
		t.Errorf("TestGetAsset(): Unable to create payload")
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
