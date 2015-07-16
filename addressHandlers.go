package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	"github.com/vennd/enu/bitcoinapi"
	"github.com/vennd/enu/database"
	"github.com/vennd/enu/enulib"
)

func CheckAndParseAddress(w http.ResponseWriter, r *http.Request) (int64, string, error) {
	// Pull headers that are necessary
	accessKey := r.Header.Get("AccessKey")
	nonce := r.Header.Get("Nonce")
	nonceInt, _ := strconv.ParseInt(nonce, 10, 64)
	//	signatureVersion := r.Header.Get("SignatureVersion")
	//	signatureMethod := r.Header.Get("SignatureMethod")
	signature := r.Header.Get("Signature")

	if headerError := CheckHeader(w, r); headerError != nil {
		return 0, "", headerError
	}

	// Limit amount read to 1MB and parse body
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1024))
	if err != nil {
		panic(err)
	}
	if err := r.Body.Close(); err != nil {
		panic(err)
	}

	// Then look up secret and sourceAddress and calculate digest
	calculatedSignature := enulib.ComputeHmac512(body, database.GetSecretByAccessKey(accessKey))

	// If we didn't receive the expected signature then raise a forbidden
	if calculatedSignature != signature {
		errorString := fmt.Sprintf("Could not verify HMAC signature. Expected: %s, received: %s\n", calculatedSignature, signature)
		myError := errors.New(errorString)

		ReturnUnauthorised(w, myError)
		return 0, "", myError
	}

	return nonceInt, accessKey, nil
}

func AddressCreate(w http.ResponseWriter, r *http.Request) {
	var address enulib.Address

	_, accessKey, err := CheckAndParseAddress(w, r)

	// Create the address
	newAddress, err := bitcoinapi.GetNewAddress()
	if err != nil {
		log.Printf("Unable to create a new bitcoin address. Error: %s\n", err.Error())

		ReturnServerError(w, err)
		return
	}
	address.Value = newAddress

	err2 := database.CreateSecondaryAddress(accessKey, newAddress)
	if err2 != nil {
		log.Printf("Unable to persist new address to database. Error: %s\n", err2.Error())
	} else {
		log.Printf("Created secondary address: %s for access key: %s\n", newAddress, accessKey)

		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusCreated)
		if err = json.NewEncoder(w).Encode(address); err != nil {
			panic(err)
		}

		//		ReturnCreated(w)
	}

}
