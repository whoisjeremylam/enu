package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/vennd/enu/counterpartycrypto"
	"github.com/vennd/enu/log"
)

var isInit = false
var apiKey = "71625888dc50d8915b871912aa6bbdce67fd1ed77d409ef1cf0726c6d9d7cf16"
var apiSecret = "a06d8cfa8692973c755b3b7321a8af7de448ec56dcfe3739716f5fa11187e4ac"
var baseURL = "http://localhost:8081"

// Creates a local server to serve client tests
func initTesting() {
	if isInit == true {
		return
	}

	router := NewRouter()

	log.Println("Enu Test API server started")
	log.Println(http.ListenAndServe("localhost:8081", router).Error())
}

func computeHmac512(message []byte, secret string) string {
	key := []byte(secret)
	h := hmac.New(sha512.New, key)
	h.Write(message)
	return hex.EncodeToString(h.Sum(nil))
}

func doEnuAPI(method string, url string, postData []byte) ([]byte, int64, error) {
	if method != "POST" && method != "GET" {
		return nil, -1000, errors.New("DoEnuAPI must be called for a POST or GET method only")
	}
	postDataJson := string(postData)

	log.Printf("Posting: %s", postDataJson)

	// Set headers
	req, err := http.NewRequest(method, url, bytes.NewBufferString(postDataJson))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("accessKey", apiKey)
	req.Header.Set("signature", ComputeHmac512(postData, apiSecret))

	// Perform request
	clientPointer := &http.Client{}
	resp, err := clientPointer.Do(req)
	if err != nil {
		panic(err)
	}

	// Did not receive an OK or Accepted
	if resp.StatusCode != 201 && resp.StatusCode != 200 {
		log.Printf("Request failed. Status code: %d\n", resp.StatusCode)

		body, err := ioutil.ReadAll(resp.Body)
		defer resp.Body.Close()

		if err != nil {
			panic(err)
		}

		//		log.Printf("Reply: %s\n", string(body))

		return body, -2000, errors.New(string(body))
	}

	body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	if err != nil {
		panic(err)
	}

	log.Printf("Reply: %s\n", string(body))

	return body, 0, nil
}

func TestWalletCreate(t *testing.T) {
	go initTesting()

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

	responseData, statusCode, err := doEnuAPI("POST", url, assetJsonBytes)

	// deserialise the response if the status is 0
	if err != nil && statusCode != 0 {
		t.Errorf("Error in API call. Error: %s, statusCode: %d\n", err, statusCode)
	}

	if err := json.Unmarshal(responseData, &wallet); err != nil {
		t.Errorf("Error in API call. Unable to unmarshal responseData. Error: %s", err)
	}

	log.FluentfObject("walletHandlers_test.go", wallet, "Received reply")
}
