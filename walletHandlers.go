package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/vennd/enu/consts"
	"github.com/vennd/enu/counterpartyapi"
	"github.com/vennd/enu/counterpartycrypto"

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
		ReturnServerError(c, w, err)
		return nil
	}

	// Create the wallet
	wallet, err = counterpartycrypto.CreateWallet()
	if err != nil {
		log.Printf("Unable to create a Counterparty wallet. Error: %s\n", err.Error())
		ReturnServerError(c, w, err)
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

func WalletSend(c context.Context, w http.ResponseWriter, r *http.Request) *appError {

	var walletPayment enulib.WalletPayment
	var paymentTag string

	requestId := c.Value(consts.RequestIdKey).(string)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	walletPayment.RequestId = requestId

	// Add to the context the RequestType
	c = context.WithValue(c, consts.RequestTypeKey, "walletPayment")

	// check generic args and parse
	payload, err := CheckAndParseJsonCTX(c, w, r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		returnCode := enulib.ReturnCode{RequestId: c.Value(consts.RequestIdKey).(string), Code: -3, Description: err.Error()}
		if err := json.NewEncoder(w).Encode(returnCode); err != nil {
			panic(err)
		}

		//		ReturnServerError(c, w, err)
		return nil
	}

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

	log.Printf("WalletSend: received request sourceAddress: %s, destinationAddress: %s, asset: %s, quantity: %d, paymentTag: %s from accessKey: %s\n", sourceAddress, destinationAddress, asset, quantity, c.Value(consts.AccessKeyKey).(string), paymentTag)

	// Generate a paymentId
	paymentId := enulib.GeneratePaymentId()
	log.Printf("Generated paymentId: %s", paymentId)

	// Return to the client the paymentId and unblock the client
	w.WriteHeader(http.StatusCreated)
	if err = json.NewEncoder(w).Encode(paymentId); err != nil {
		panic(err)
	}

	go counterpartyapi.DelegatedSend(c.Value(consts.AccessKeyKey).(string), passphrase, sourceAddress, destinationAddress, asset, quantity, paymentId, paymentTag, requestId)

	return nil
}

func WalletBalance(c context.Context, w http.ResponseWriter, r *http.Request) *appError {

	var walletbalance enulib.AddressBalances

	requestId := c.Value(consts.RequestIdKey).(string)
	walletbalance.RequestId = requestId
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	// Add to the context the RequestType
	//	c = context.WithValue(c, consts.RequestTypeKey, "addressBalances")

	// check generic args and parse
	_, err := CheckAndParseJsonCTX(c, w, r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		returnCode := enulib.ReturnCode{RequestId: c.Value(consts.RequestIdKey).(string), Code: -3, Description: err.Error()}
		if err := json.NewEncoder(w).Encode(returnCode); err != nil {
			panic(err)
		}

		//		ReturnServerError(c, w, err)
		return nil
	}

	vars := mux.Vars(r)
	address := vars["address"]

	//	**** Need to check all the types are as expected and all required parameters received

	log.Printf("WalletBalance: received request address: %s from accessKey: %s\n", address, c.Value(consts.AccessKeyKey).(string))

	result, err := counterpartyapi.GetBalancesByAddress(address)
	if err != nil {
		ReturnServerError(c, w, err)
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
