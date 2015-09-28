package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/vennd/enu/internal/github.com/gorilla/mux"
	"github.com/vennd/enu/internal/golang.org/x/net/context"

	"github.com/vennd/enu/bitcoinapi"
	"github.com/vennd/enu/consts"
	"github.com/vennd/enu/database"
	"github.com/vennd/enu/enulib"
	"github.com/vennd/enu/log"
)

func PaymentCreate(c context.Context, w http.ResponseWriter, r *http.Request) *appError {

	var simplePayment enulib.SimplePayment
	requestId := c.Value(consts.RequestIdKey).(string)
	simplePayment.RequestId = requestId
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	// Add to the context the RequestType
	c = context.WithValue(c, consts.RequestTypeKey, "simplePayment")

	// check generic args and parse
	m, err := CheckAndParseJsonCTX(c, w, r)
	if err != nil {
		// Status errors are handled inside CheckAndParseJsonCTX, so we just exit gracefully
		return nil
	}

	//	m := payload.(map[string]interface{})

	paymentId := m["paymentId"].(string)
	sourceAddress := m["sourceAddress"].(string)
	destinationAddress := m["destinationAddress"].(string)
	asset := m["asset"].(string)
	amount := uint64(m["amount"].(float64))
	txFee := uint64(m["txFee"].(float64))
	paymentTag := m["paymentTag"].(string)

	if m["paymentTag"] != nil {
		paymentTag = m["paymentTag"].(string)
	}

	// If a paymentId is not specified, generate one
	if paymentId == "" {
		paymentId = enulib.GeneratePaymentId()
		simplePayment.PaymentId = paymentId
		log.FluentfContext(consts.LOGINFO, c, "Generated paymentId: %s", simplePayment.PaymentId)
	}

	database.InsertPayment(c, c.Value(consts.AccessKeyKey).(string), 0, paymentId, sourceAddress, destinationAddress, asset, amount, "Authorized", 0, txFee, paymentTag)
	// errorhandling here!!

	simplePayment.SourceAddress = sourceAddress
	simplePayment.DestinationAddress = destinationAddress
	simplePayment.Asset = asset
	simplePayment.Amount = amount
	simplePayment.TxFee = int64(txFee)
	simplePayment.PaymentTag = paymentTag

	// Return to the client the paymentId
	w.WriteHeader(http.StatusCreated)
	if err = json.NewEncoder(w).Encode(simplePayment); err != nil {
		panic(err)
	}

	return nil
}

func PaymentRetry(c context.Context, w http.ResponseWriter, r *http.Request) *appError {

	var payment enulib.SimplePayment
	requestId := c.Value(consts.RequestIdKey).(string)
	payment.RequestId = requestId
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	// check generic args and parse
	_, err := CheckAndParseJsonCTX(c, w, r)
	if err != nil {
		// Status errors are handled inside CheckAndParseJsonCTX, so we just exit gracefully
		return nil
	}

	//	var simplePayment enulib.SimplePayment

	vars := mux.Vars(r)
	paymentId := vars["paymentId"]

	log.FluentfContext(consts.LOGINFO, c, "PaymentRetry called for paymentId %s\n", paymentId)
	payment = database.GetPaymentByPaymentId(c, c.Value(consts.AccessKeyKey).(string), paymentId)

	// Payment not found
	if payment.Status == "Not found" || payment.Status == "" {
		log.FluentfContext(consts.LOGERROR, c, "PaymentId: %s not found", paymentId)
		ReturnNotFound(c, w)
		return nil
	}

	// Payment isn't in an error state or manual state
	if payment.Status != "error" && payment.Status != "manual" {
		errorString := fmt.Sprintf("PaymentId: %s is not in an 'error' or 'manual' state. It is in '%s' state.", paymentId, payment.Status)
		log.FluentfContext(consts.LOGINFO, c, errorString)
		ReturnNotFoundWithCustomError(c, w, errorString)
		return nil
	}

	err = database.UpdatePaymentStatusByPaymentId(c, c.Value(consts.AccessKeyKey).(string), paymentId, "authorized")

	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, err.Error())
		ReturnUnprocessableEntity(c, w, err)
	}

	w.WriteHeader(http.StatusOK)
	if err = json.NewEncoder(w).Encode(payment); err != nil {
		panic(err)
	}

	return nil
}

func GetPayment(c context.Context, w http.ResponseWriter, r *http.Request) *appError {

	var payment enulib.SimplePayment
	requestId := c.Value(consts.RequestIdKey).(string)
	payment.RequestId = requestId
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	// check generic args and parse
	_, err := CheckAndParseJsonCTX(c, w, r)
	if err != nil {
		// Status errors are handled inside CheckAndParseJsonCTX, so we just exit gracefully
		return nil
	}

	vars := mux.Vars(r)
	paymentId := vars["paymentId"]

	if paymentId == "" || len(paymentId) < 16 {
		w.WriteHeader(http.StatusBadRequest)
		returnCode := enulib.ReturnCode{RequestId: c.Value(consts.RequestIdKey).(string), Code: -3, Description: "Incorrect paymentId"}
		if err := json.NewEncoder(w).Encode(returnCode); err != nil {
			panic(err)
		}
		return nil

	}

	log.FluentfContext(consts.LOGINFO, c, "GetPayment called for '%s' by '%s'\n", paymentId, c.Value(consts.AccessKeyKey).(string))

	payment = database.GetPaymentByPaymentId(c, c.Value(consts.AccessKeyKey).(string), paymentId)
	// errorhandling here!!

	// Add the blockchain status
	if payment.BroadcastTxId != "" {
		confirmations, err := bitcoinapi.GetConfirmations(payment.BroadcastTxId)
		if err == nil || confirmations == 0 {
			payment.BlockchainStatus = "unconfimed"
			payment.BlockchainConfirmations = 0
		}

		payment.BlockchainStatus = "confirmed"
		payment.BlockchainConfirmations = confirmations
	}

	w.WriteHeader(http.StatusOK)
	if err = json.NewEncoder(w).Encode(payment); err != nil {
		panic(err)
	}

	return nil
}

func GetPaymentsByAddress(c context.Context, w http.ResponseWriter, r *http.Request) *appError {

	var payment enulib.SimplePayment
	requestId := c.Value(consts.RequestIdKey).(string)
	payment.RequestId = requestId
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	// check generic args and parse
	_, err := CheckAndParseJsonCTX(c, w, r)
	if err != nil {
		// Status errors are handled inside CheckAndParseJsonCTX, so we just exit gracefully
		return nil
	}

	vars := mux.Vars(r)
	address := vars["address"]

	if address == "" {
		w.WriteHeader(http.StatusBadRequest)
		returnCode := enulib.ReturnCode{RequestId: c.Value(consts.RequestIdKey).(string), Code: -3, Description: "Incorrect address"}
		if err := json.NewEncoder(w).Encode(returnCode); err != nil {
			panic(err)
		}
		return nil
	}

	log.FluentfContext(consts.LOGINFO, c, "GetPaymentByAddress called for '%s' by '%s'\n", address, c.Value(consts.AccessKeyKey).(string))

	payments := database.GetPaymentsByAddress(c, c.Value(consts.AccessKeyKey).(string), address)
	// errorhandling here!!

	// Add the blockchain status

	for i, p := range payments {
		if p.BroadcastTxId != "" {
			confirmations, err := bitcoinapi.GetConfirmations(p.BroadcastTxId)
			if err == nil || confirmations == 0 {
				payments[i].BlockchainStatus = "unconfimed"
				payments[i].BlockchainConfirmations = 0
			}

			payments[i].BlockchainStatus = "confirmed"
			payments[i].BlockchainConfirmations = confirmations
		}
	}

	w.WriteHeader(http.StatusOK)
	if err = json.NewEncoder(w).Encode(payments); err != nil {
		panic(err)
	}

	return nil
}
