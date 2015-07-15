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

	"github.com/gorilla/mux"

	"github.com/vennd/enulib"
	"github.com/vennd/enulib/database"
)

func CheckAndParsePayment(w http.ResponseWriter, r *http.Request) (enulib.SimplePayment, int64, string, string, error) {
	var paymentId string

	vars := mux.Vars(r)
	paymentId = vars["paymentId"]

	var simplePayment enulib.SimplePayment

	// Pull headers that are necessary
	accessKey := r.Header.Get("AccessKey")
	nonce := r.Header.Get("Nonce")
	nonceInt, _ := strconv.ParseInt(nonce, 10, 64)
	//	signatureVersion := r.Header.Get("SignatureVersion")
	//	signatureMethod := r.Header.Get("SignatureMethod")
	signature := r.Header.Get("Signature")

	if headerError := CheckHeader(w, r); headerError != nil {
		return simplePayment, 0, "", paymentId, headerError
	}

	// Limit amount read to 1024 bytes and parse body
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1024))
	if err != nil {
		panic(err)
	}
	if err := r.Body.Close(); err != nil {
		panic(err)
	}
	if err := json.Unmarshal(body, &simplePayment); err != nil {
		ReturnUnprocessableEntity(w, errors.New("Unable to unmarshal body"))
	}
	log.Printf("Received: %s", body)

	// Then look up secret and calculate digest
	calculatedSignature := enulib.ComputeHmac512(body, database.GetSecretByAccessKey(accessKey))

	// If we didn't receive the expected signature then raise a forbidden
	if calculatedSignature != signature {
		errorString := fmt.Sprintf("Could not verify HMAC signature. Expected: %s, received: %s\n", calculatedSignature, signature)
		myError := errors.New(errorString)

		ReturnUnauthorised(w, myError)
		return simplePayment, 0, "", paymentId, myError
	}

	// Attempt to get payment details for payment which user has no access to (assumes that a new address is generated for each asset per user)
	if database.GetAssetByAccessKey(accessKey) != simplePayment.Asset {
		errorString := fmt.Sprintf("Attempt to send asset without permission Expected: %s, received: %s\n", database.GetAssetByAccessKey(accessKey), simplePayment.Asset)
		myError := errors.New(errorString)

		ReturnUnauthorised(w, myError)
		return simplePayment, 0, "", paymentId, myError
	}

	// Check if required parameters have been set
	if simplePayment.Amount == 0 || simplePayment.Asset == "" || simplePayment.DestinationAddress == "" {
		errorString := fmt.Sprintf("All of the following parameters are mandatory: amount, asset, destinationAddress. However, not all were specified.")
		myError := errors.New(errorString)

		ReturnUnauthorised(w, myError)
		return simplePayment, 0, "", paymentId, myError
	}

	// Check txFee has been set. If not, set a sane default
	if simplePayment.TxFee == 0 {
		simplePayment.TxFee = 7800
	}

	// Check if sourceAddress was specified, if it wasn't then default to the latest generated source address
	if simplePayment.SourceAddress == "" {
		simplePayment.SourceAddress = database.GetSourceAddressByAccessKey(accessKey)
	}

	return simplePayment, nonceInt, accessKey, paymentId, nil
}

func CheckAndParsePaymentId(w http.ResponseWriter, r *http.Request) (string, int64, string, error) {
	var paymentId string

	vars := mux.Vars(r)
	paymentId = vars["paymentId"]

	// Pull headers that are necessary
	accessKey := r.Header.Get("AccessKey")
	nonce := r.Header.Get("Nonce")
	nonceInt, _ := strconv.ParseInt(nonce, 10, 64)
	signature := r.Header.Get("Signature")

	// Check headers
	if headerError := CheckHeader(w, r); headerError != nil {
		return paymentId, 0, "", headerError
	}

	// Limit amount read to 1024 bytes and parse body
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1024))
	if err != nil {
		panic(err)
	}
	if err := r.Body.Close(); err != nil {
		panic(err)
	}

	// Then look up secret and sourceAddress and calculate digest
	var calculatedSignature string
	if len(body) > 0 {
		calculatedSignature = enulib.ComputeHmac512(body, database.GetSecretByAccessKey(accessKey))
	} else {
		// Workaround for php as it can't seem to calculate the hash on an empty array correctly...
		calculatedSignature = enulib.ComputeHmac512([]byte("[]"), database.GetSecretByAccessKey(accessKey))
	}

	// If we didn't receive the expected signature then raise a forbidden
	if calculatedSignature != signature {
		errorString := fmt.Sprintf("Could not verify HMAC signature. Expected: %s, received: %s\n", calculatedSignature, signature)

		ReturnUnauthorised(w, nil)
		return paymentId, 0, "", errors.New(errorString)
	}

	return paymentId, nonceInt, accessKey, nil
}

func PaymentCreate(w http.ResponseWriter, r *http.Request) {
	var simplePayment enulib.SimplePayment

	simplePayment, newNonce, accessKey, _, err := CheckAndParsePayment(w, r)

	// If a paymentId is not specified, generate one
	if simplePayment.PaymentId == "" {
		simplePayment.PaymentId = enulib.GeneratePaymentId()
	}

	if err == nil {
		database.InsertPayment(accessKey, 0, simplePayment.PaymentId, simplePayment.SourceAddress, simplePayment.DestinationAddress, simplePayment.Asset, simplePayment.Amount, "Authorized", 0, simplePayment.TxFee, simplePayment.PaymentTag)
		database.UpdateNonce(accessKey, newNonce)
		ReturnCreated(w)
	} else {
		log.Println(err)
	}
}

func PaymentRetry(w http.ResponseWriter, r *http.Request) {
	//	var simplePayment enulib.SimplePayment

	paymentId, newNonce, accessKey, err := CheckAndParsePaymentId(w, r)

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

		database.UpdateNonce(accessKey, newNonce)
		ReturnOK(w)
	} else {
		log.Println(err)
	}
}

func GetPayment(w http.ResponseWriter, r *http.Request) {
	paymentId, _, accessKey, err := CheckAndParsePaymentId(w, r)

	log.Printf("GetPayment called for '%s' by '%s'\n", paymentId, accessKey)

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
