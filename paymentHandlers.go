package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/vennd/enu/internal/github.com/gorilla/mux"

	"github.com/vennd/enu/consts"
	"github.com/vennd/enu/database"
	"github.com/vennd/enu/enulib"
	"github.com/vennd/enu/internal/golang.org/x/net/context"	
)

func PaymentCreate(c context.Context, w http.ResponseWriter, r *http.Request) *appError {


	var simplePayment enulib.SimplePayment
	requestId := c.Value(consts.RequestIdKey).(string)
	simplePayment.RequestId = requestId
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	
// check generic args and parse
	payload, err := CheckAndParseJsonCTX(c, w, r)
	if err != nil {
		ReturnServerError(w, err)
		return nil
	}

	m := payload.(map[string]interface{})

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
		log.Printf("Generated paymentId: %s", simplePayment.PaymentId)
	}


	database.InsertPayment(c.Value(consts.AccessKeyKey).(string), 0, paymentId, sourceAddress, destinationAddress, asset, amount, "Authorized", 0, txFee, paymentTag, requestId)
// errorhandling here!!


	
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
		ReturnServerError(w, err)
		return nil
	}

	//	var simplePayment enulib.SimplePayment

	vars := mux.Vars(r)
	paymentId := vars["paymentId"]

	log.Printf("PaymentRetry called for paymentId %s\n", paymentId)

	payment = database.GetPaymentByPaymentId(c.Value(consts.AccessKeyKey).(string), paymentId)

	// Payment not found
	if payment.Status == "Not found" || payment.Status == "" {
		log.Printf("PaymentId: %s not found", paymentId)

		ReturnNotFound(w)
		return nil
	}

	// Payment isn't in an error state or manual state
	if payment.Status != "error" && payment.Status != "manual" {
		errorString := fmt.Sprintf("PaymentId: %s is not in an 'error' or 'manual' state. It is in '%s' state.", paymentId, payment.Status)
		log.Println(errorString)

		ReturnNotFoundWithCustomError(w, errorString)
		return nil
	}

	err = database.UpdatePaymentStatusByPaymentId(c.Value(consts.AccessKeyKey).(string), paymentId, "authorized")

	if err != nil {
		log.Println(err.Error())
		ReturnUnprocessableEntity(w, err)
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
		ReturnServerError(w, err)
		return nil
	}

	vars := mux.Vars(r)
	paymentId := vars["paymentId"]

	log.Printf("GetPayment called for '%s' by '%s'\n", paymentId, c.Value(consts.AccessKeyKey).(string))

	payment = database.GetPaymentByPaymentId(c.Value(consts.AccessKeyKey).(string), paymentId)
// errorhandling here!!

	w.WriteHeader(http.StatusOK)
	if err = json.NewEncoder(w).Encode(payment); err != nil {
		panic(err)
	}
	
	return nil
}
