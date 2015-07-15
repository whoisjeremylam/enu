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

	"github.com/vennd/enulib"
	"github.com/vennd/enulib/counterpartyapi"
	"github.com/vennd/enulib/counterpartycrypto"
	"github.com/vennd/enulib/database"

	"github.com/gorilla/mux"
)

func CheckHeaderGeneric(w http.ResponseWriter, r *http.Request) (string, int64, error) {
	// Pull headers that are necessary
	accessKey := r.Header.Get("AccessKey")
	nonce := r.Header.Get("Nonce")
	nonceInt, convertNonceErr := strconv.ParseInt(nonce, 10, 64)
	//	signatureVersion := r.Header.Get("SignatureVersion")
	//	signatureMethod := r.Header.Get("SignatureMethod")
	signature := r.Header.Get("Signature")
	var nonceDB int64
	var err error

	// Headers weren't set properly, return forbidden
	if accessKey == "" || nonce == "" || signature == "" {
		err = errors.New("Request headers were not set correctly, ensure the following headers are set: accessKey, none, signature")

		log.Printf("Headers set incorrectly: accessKey=%s, nonce=%s, signature=%s\n", accessKey, nonce, signature)
		ReturnUnauthorised(w, err)

		return accessKey, nonceInt, err
	} else if convertNonceErr != nil {
		err = errors.New("Invalid nonce value")
		// Unable to convert the value of nonce in the header to an integer
		log.Println(convertNonceErr)
		ReturnUnauthorised(w, err)

		return accessKey, nonceInt, err
	} else if nonceInt <= database.GetNonceByAccessKey(accessKey) {
		err = errors.New("Invalid nonce value")

		//Nonce is not greater than the nonce in the DB
		log.Printf("Nonce for accessKey %s provided is <= nonce in db. %s <= %d\n", accessKey, nonce, nonceDB)
		ReturnUnauthorised(w, err)

		return accessKey, nonceInt, err
	} else if database.UserKeyExists(accessKey) == false {
		// User key doesn't exist
		log.Printf("Attempt to access API with unknown user key: %s", accessKey)
		ReturnUnauthorised(w, errors.New("Attempt to access API with unknown user key"))

		return accessKey, nonceInt, err
	}

	return accessKey, nonceInt, nil
}

func CheckAndParseJson(w http.ResponseWriter, r *http.Request) (interface{}, string, int64, error) {
	//	var blockchainId string
	var payload interface{}

	// Pull headers that are necessary
	accessKey := r.Header.Get("AccessKey")
	nonce := r.Header.Get("Nonce")
	nonceInt, _ := strconv.ParseInt(nonce, 10, 64)
	//	signatureVersion := r.Header.Get("SignatureVersion")
	//	signatureMethod := r.Header.Get("SignatureMethod")
	signature := r.Header.Get("Signature")

	accessKey, nonceInt, err := CheckHeaderGeneric(w, r)
	if err != nil {
		return nil, accessKey, nonceInt, err
	}

	// Limit amount read to 512,000 bytes and parse body
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 512000))
	if err != nil {
		panic(err)
	}
	if err := r.Body.Close(); err != nil {
		panic(err)
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		ReturnUnprocessableEntity(w, errors.New("Unable to unmarshal body"))
	}
	log.Printf("Received: %s", body)

	// Then look up secret and calculate digest
	calculatedSignature := enulib.ComputeHmac512(body, database.GetSecretByAccessKey(accessKey))

	// If we didn't receive the expected signature then raise a forbidden
	if calculatedSignature != signature {
		errorString := fmt.Sprintf("Could not verify HMAC signature. Expected: %s, received: %s\n", calculatedSignature, signature)
		err := errors.New(errorString)

		ReturnUnauthorised(w, err)
		return nil, accessKey, nonceInt, err
	}

	database.UpdateNonce(accessKey, nonceInt)

	return payload, accessKey, nonceInt, nil
}

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
