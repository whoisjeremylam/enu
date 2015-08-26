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


	if err == nil {
		database.InsertPayment(c.Value(consts.AccessKeyKey).(string), 0, paymentId, sourceAddress, destinationAddress, asset, amount, "Authorized", 0, txFee, paymentTag, requestId)
//		ReturnCreated(w)
	} else {
		log.Println(err)
	}
	
// Return to the client the paymentId 
	w.WriteHeader(http.StatusCreated)
	if err = json.NewEncoder(w).Encode(simplePayment); err != nil {
		panic(err)
	}	
	
	return nil
}

func PaymentRetry(w http.ResponseWriter, r *http.Request) {
	//	var simplePayment enulib.SimplePayment

	vars := mux.Vars(r)
	paymentId := vars["paymentId"]

	_, accessKey, nonce, err := CheckAndParseJson(w, r)
	if err != nil {
		ReturnServerError(w, err)

		return
	}

	log.Printf("PaymentRetry called for paymentId %s\n", paymentId)

	if err == nil {
		payment := database.GetPaymentByPaymentId(accessKey, paymentId)

		// Payment not found
		if payment.Status == "Not found" || payment.Status == "" {
			log.Printf("PaymentId: %s not found", paymentId)

			ReturnNotFound(w)
			return
		}

		// Payment isn't in an error state or manual state
		if payment.Status != "error" && payment.Status != "manual" {
			errorString := fmt.Sprintf("PaymentId: %s is not in an 'error' or 'manual' state. It is in '%s' state.", paymentId, payment.Status)
			log.Println(errorString)

			ReturnNotFoundWithCustomError(w, errorString)
			return
		}

		err = database.UpdatePaymentStatusByPaymentId(accessKey, paymentId, "authorized")

		if err != nil {
			log.Println(err.Error())
			ReturnUnprocessableEntity(w, err)
		}

		database.UpdateNonce(accessKey, nonce)
		ReturnOK(w)
	} else {
		log.Println(err)
	}
}

func GetPayment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	paymentId := vars["paymentId"]

	_, accessKey, nonce, err := CheckAndParseJson(w, r)
	if err != nil {
		ReturnServerError(w, err)

		return
	}

	log.Printf("GetPayment called for '%s' by '%s'\n", paymentId, accessKey)

	database.UpdateNonce(accessKey, nonce)

	if err == nil {
		payment := database.GetPaymentByPaymentId(accessKey, paymentId)

		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusOK)
		if err = json.NewEncoder(w).Encode(payment); err != nil {
			panic(err)
		}
	} else {
		log.Println(err)
	}
}
