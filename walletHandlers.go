package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/vennd/enu/consts"
	"github.com/vennd/enu/counterpartyapi"
	"github.com/vennd/enu/counterpartycrypto"
	"github.com/vennd/enu/database"
	"github.com/vennd/enu/enulib"

	"github.com/vennd/enu/internal/github.com/gorilla/mux"
	"github.com/vennd/enu/internal/golang.org/x/net/context"	
)

func WalletCreate(c context.Context, w http.ResponseWriter, r *http.Request) *appError {
	
	var wallet counterpartycrypto.CounterpartyWallet
	requestId := c.Value(consts.RequestIdKey).(string)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	wallet.RequestId = requestId

// check generic args and parse
	_, err := CheckAndParseJsonCTX(c, w, r)	
	if err != nil {
		ReturnServerError(w, err)
		return nil
	}

	// Create the wallet
	wallet, err = counterpartycrypto.CreateWallet()
	if err != nil {
		log.Printf("Unable to create a Counterparty wallet. Error: %s\n", err.Error())
		ReturnServerError(w, err)
		return nil
	}
	log.Printf("Created a new wallet with first address: %s for access key: %s\n (requestID: %s)", wallet.Addresses[0], c.Value(consts.AccessKeyKey).(string), requestId)

	// Return the wallet
	wallet.RequestId = requestId
	w.WriteHeader(http.StatusCreated)
	if err = json.NewEncoder(w).Encode(wallet); err != nil {
		panic(err)
	} 
	
	return nil
}

func WalletSend(w http.ResponseWriter, r *http.Request) {
	var paymentTag string

	requestId := ""

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

	go counterpartyapi.DelegatedSend(accessKey, passphrase, sourceAddress, destinationAddress, asset, quantity, paymentId, paymentTag, requestId)
}


func WalletBalance(c context.Context, w http.ResponseWriter, r *http.Request) *appError {

	var walletbalance enulib.AddressBalances
	
	requestId := c.Value(consts.RequestIdKey).(string)
	walletbalance.RequestId = requestId
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

// check generic args and parse
	_, err := CheckAndParseJsonCTX(c, w, r)
	if err != nil {
		ReturnServerError(w, err)
		return nil
	}

	vars := mux.Vars(r)
	address := vars["address"]


	//	**** Need to check all the types are as expected and all required parameters received

	log.Printf("WalletBalance: received request address: %s from accessKey: %s\n", address, c.Value(consts.AccessKeyKey).(string))


	result, err := counterpartyapi.GetBalancesByAddress(address)
	if err != nil {
		ReturnServerError(w, err)
		return nil
	}

	// Iterate and gather the balances to return
	walletbalance.Address = address
	for _, item := range result {
		var balance enulib.Amount

		balance.Asset = item.Asset
		balance.Quantity = item.Quantity

		walletbalance.Balances = append(walletbalance.Balances, balance)
	}


	w.WriteHeader(http.StatusOK)
	if err = json.NewEncoder(w).Encode(walletbalance); err != nil {
		panic(err)
	}
	
	return nil
}
