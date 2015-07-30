package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/vennd/enu/counterpartyapi"
	"github.com/vennd/enu/counterpartycrypto"
	"github.com/vennd/enu/database"
	"github.com/vennd/enu/enulib"

	"github.com/vennd/enu/internal/github.com/gorilla/mux"
)

func WalletCreate(w http.ResponseWriter, r *http.Request) {
	var wallet counterpartycrypto.CounterpartyWallet

	_, accessKey, nonce, err := CheckAndParseJson(w, r)
	if err != nil {
		ReturnServerError(w, err)

		return
	}

	err2 := database.UpdateNonce(accessKey, nonce)
	if err2 != nil {
		ReturnServerError(w, err)

		return
	}

	// Create the wallet
	wallet, err = counterpartycrypto.CreateWallet()
	if err != nil {
		log.Printf("Unable to create a Counterparty wallet. Error: %s\n", err.Error())

		ReturnServerError(w, err)

		return
	}
	log.Printf("Created a new wallet with first address: %s for access key: %s\n", wallet.Addresses[0], accessKey)

	// Return the wallet
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusCreated)
	if err = json.NewEncoder(w).Encode(wallet); err != nil {
		panic(err)
	} else {
		log.Println(err)
	}
}

func WalletSend(w http.ResponseWriter, r *http.Request) {
	var paymentTag string

	payload, accessKey, nonce, err := CheckAndParseJson(w, r)

	m := payload.(map[string]interface{})

	passphrase := m["passphrase"].(string)
	sourceAddress := m["sourceAddress"].(string)
	destinationAddress := m["destinationAddress"].(string)
	asset := m["asset"].(string)
	quantity := uint64(m["quantity"].(float64))

	if m["paymentTag"] != nil {
		paymentTag = m["paymentTag"].(string)
	}

	//	**** Need to check all the types are as expected and all required parameters received

	log.Printf("WalletSend: received request sourceAddress: %s, destinationAddress: %s, asset: %s, quantity: %d, paymentTag: %s from accessKey: %s\n", sourceAddress, destinationAddress, asset, quantity, accessKey, paymentTag)

	err = database.UpdateNonce(accessKey, nonce)
	if err != nil {
		ReturnServerError(w, err)

		return
	}

	// Generate a paymentId
	paymentId := enulib.GeneratePaymentId()
	log.Printf("Generated paymentId: %s", paymentId)

	// Return to the client the paymentId and unblock the client
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusCreated)
	if err = json.NewEncoder(w).Encode(paymentId); err != nil {
		panic(err)
	} else {
		log.Println(err)
	}

	go counterpartyapi.DelegatedSend(accessKey, passphrase, sourceAddress, destinationAddress, asset, quantity, paymentId, paymentTag)
}

func WalletBalance(w http.ResponseWriter, r *http.Request) {
	type amount struct {
		Asset    string `json:"asset"`
		Quantity uint64 `json:"quantity"`
	}
	var walletBalances struct {
		Address  string   `json:"address"`
		Balances []amount `json:"balances"`
	}

	vars := mux.Vars(r)
	address := vars["address"]

	_, accessKey, nonce, err := CheckAndParseJson(w, r)
	if err != nil {
		ReturnServerError(w, err)

		return
	}

	//	**** Need to check all the types are as expected and all required parameters received

	log.Printf("WalletBalance: received request address: %s from accessKey: %s\n", address, accessKey)

	err = database.UpdateNonce(accessKey, nonce)
	if err != nil {
		ReturnServerError(w, err)

		return
	}

	result, err := counterpartyapi.GetBalancesByAddress(address)
	if err != nil {
		ReturnServerError(w, err)

		return
	}

	// Iterate and gather the balances to return
	walletBalances.Address = address
	for _, item := range result {
		var balance amount

		balance.Asset = item.Asset
		balance.Quantity = item.Quantity

		walletBalances.Balances = append(walletBalances.Balances, balance)
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	if err = json.NewEncoder(w).Encode(walletBalances); err != nil {
		panic(err)
	} else {
		log.Println(err)
	}
}
