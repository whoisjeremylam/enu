package main

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"

	"github.com/vennd/enu/log"
)

var isInit = false
var apiKey = "71625888dc50d8915b871912aa6bbdce67fd1ed77d409ef1cf0726c6d9d7cf16"
var apiSecret = "a06d8cfa8692973c755b3b7321a8af7de448ec56dcfe3739716f5fa11187e4ac"
var baseURL = "http://localhost:8081"

// Creates a local server to serve client tests
func InitTesting() {
	if isInit == true {
		return
	}

	isInit = true

	router := NewRouter()

	log.Println("Enu Test API server started")
	log.Println(http.ListenAndServe("localhost:8081", router).Error())
}

func DoEnuAPITesting(method string, url string, postData []byte) ([]byte, int, error) {
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
		//		log.Printf("Request failed. Status code: %d\n", resp.StatusCode)

		body, err := ioutil.ReadAll(resp.Body)
		defer resp.Body.Close()

		if err != nil {
			panic(err)
		}

		//		log.Printf("Reply: %s\n", string(body))

		return body, resp.StatusCode, errors.New(string(body))
	}

	body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	if err != nil {
		panic(err)
	}

	log.Printf("Reply: %#v\n", string(body))

	return body, 0, nil
}
