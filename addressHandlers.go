package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/vennd/enu/bitcoinapi"
	"github.com/vennd/enu/database"
	"github.com/vennd/enu/enulib"
)

func AddressCreate(w http.ResponseWriter, r *http.Request) {
	var address enulib.Address

	_, accessKey, nonce, err := CheckAndParseJson(w, r)
	if err != nil {
		ReturnServerError(w, err)

		return
	}

	database.UpdateNonce(accessKey, nonce)

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
