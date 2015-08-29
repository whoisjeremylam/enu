package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/vennd/enu/consts"
	"github.com/vennd/enu/database"
	"github.com/vennd/enu/enulib"

	"github.com/vennd/enu/internal/golang.org/x/net/context"
)

func ReturnUnauthorised(c context.Context, w http.ResponseWriter, e error) {
	var returnCode enulib.ReturnCode

	if e == nil {
		returnCode = enulib.ReturnCode{Code: -1, Description: "Forbidden", RequestId: c.Value(consts.RequestIdKey).(string)}
	} else {
		returnCode = enulib.ReturnCode{Code: -1, Description: e.Error(), RequestId: c.Value(consts.RequestIdKey).(string)}
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusForbidden)
	if err := json.NewEncoder(w).Encode(returnCode); err != nil {
		panic(err)
	}
}

func ReturnUnprocessableEntity(c context.Context, w http.ResponseWriter, e error) {
	var returnCode enulib.ReturnCode

	if e == nil {
		returnCode = enulib.ReturnCode{Code: -2, Description: "Unprocessable entity", RequestId: c.Value(consts.RequestIdKey).(string)}
	} else {
		returnCode = enulib.ReturnCode{Code: -2, Description: e.Error(), RequestId: c.Value(consts.RequestIdKey).(string)}
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(422)
	if err := json.NewEncoder(w).Encode(returnCode); err != nil {
		panic(err)
	}
}

func ReturnCreated(c context.Context, w http.ResponseWriter) {
	returnCode := enulib.ReturnCode{Code: 0, Description: "Success", RequestId: c.Value(consts.RequestIdKey).(string)}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(returnCode); err != nil {
		panic(err)
	}
}

func ReturnOK(c context.Context, w http.ResponseWriter) {
	returnCode := enulib.ReturnCode{Code: 0, Description: "Success", RequestId: c.Value(consts.RequestIdKey).(string)}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(returnCode); err != nil {
		panic(err)
	}
}

func ReturnNotFound(c context.Context, w http.ResponseWriter) {
	returnCode := enulib.ReturnCode{Code: -3, Description: "Not found", RequestId: c.Value(consts.RequestIdKey).(string)}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusNotFound)
	if err := json.NewEncoder(w).Encode(returnCode); err != nil {
		panic(err)
	}
}

func ReturnNotFoundWithCustomError(c context.Context, w http.ResponseWriter, errorString string) {
	returnCode := enulib.ReturnCode{Code: -3, Description: errorString, RequestId: c.Value(consts.RequestIdKey).(string)}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusNotFound)
	if err := json.NewEncoder(w).Encode(returnCode); err != nil {
		panic(err)
	}
}

func ReturnServerError(c context.Context, w http.ResponseWriter, e error) {
	var returnCode enulib.ReturnCode

	if e == nil {
		log.Printf("Unspecified server error.\n")
		returnCode = enulib.ReturnCode{Code: -10000, Description: "Unspecified server error. Please contact Vennd.io support.", RequestId: c.Value(consts.RequestIdKey).(string)}
	} else {
		log.Printf("Server error: %s\n", e.Error())
		returnCode = enulib.ReturnCode{Code: -10000, Description: e.Error(), RequestId: c.Value(consts.RequestIdKey).(string)}
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusInternalServerError)
	if err := json.NewEncoder(w).Encode(returnCode); err != nil {
		panic(err)
	}
}

func Index(c context.Context, w http.ResponseWriter, r *http.Request) *appError {
	rand.Seed(time.Now().UnixNano())
	number := rand.Intn(len(quotes))

	fmt.Fprintf(w, "%s\n", quotes[number])
	return nil
}

func CheckHeaderGeneric(c context.Context, w http.ResponseWriter, r *http.Request) (string, int64, error) {
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
		ReturnUnauthorised(c, w, err)

		return accessKey, nonceInt, err
	} else if convertNonceErr != nil {
		err = errors.New("Invalid nonce value")
		// Unable to convert the value of nonce in the header to an integer
		log.Println(convertNonceErr)
		ReturnUnauthorised(c, w, err)

		return accessKey, nonceInt, err
	} else if nonceInt <= database.GetNonceByAccessKey(accessKey) {
		err = errors.New("Invalid nonce value")

		//Nonce is not greater than the nonce in the DB
		log.Printf("Nonce for accessKey %s provided is <= nonce in db. %s <= %d\n", accessKey, nonce, nonceDB)
		ReturnUnauthorised(c, w, err)

		return accessKey, nonceInt, err
	} else if database.UserKeyExists(accessKey) == false {
		returnErr := errors.New("Attempt to access API with unknown user key")
		// User key doesn't exist
		log.Printf("Attempt to access API with unknown user key: %s", accessKey)
		ReturnUnauthorised(c, w, returnErr)

		return accessKey, nonceInt, returnErr
	}

	return accessKey, nonceInt, nil
}

//func CheckAndParseJson(w http.ResponseWriter, r *http.Request) (interface{}, string, int64, error) {
//	//	var blockchainId string
//	var payload interface{}

//	// Pull headers that are necessary
//	accessKey := r.Header.Get("AccessKey")
//	nonce := r.Header.Get("Nonce")
//	nonceInt, _ := strconv.ParseInt(nonce, 10, 64)
//	//	signatureVersion := r.Header.Get("SignatureVersion")
//	//	signatureMethod := r.Header.Get("SignatureMethod")
//	signature := r.Header.Get("Signature")

//	accessKey, nonceInt, err := CheckHeaderGeneric(w, r)
//	if err != nil {
//		return nil, accessKey, nonceInt, err
//	}

//	// Limit amount read to 512,000 bytes and parse body
//	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 512000))
//	if err != nil {
//		panic(err)
//	}
//	if err := r.Body.Close(); err != nil {
//		panic(err)
//	}
//	if err := json.Unmarshal(body, &payload); err != nil {
//		ReturnUnprocessableEntity(w, errors.New("Unable to unmarshal body"))
//	}
//	log.Printf("Request received: %s", body)

//	// Then look up secret and calculate digest
//	calculatedSignature := enulib.ComputeHmac512(body, database.GetSecretByAccessKey(accessKey))

//	// If we didn't receive the expected signature then raise a forbidden
//	if calculatedSignature != signature {
//		errorString := fmt.Sprintf("Could not verify HMAC signature. Expected: %s, received: %s", calculatedSignature, signature)
//		err := errors.New(errorString)

//		return nil, accessKey, nonceInt, err
//	}

//	database.UpdateNonce(accessKey, nonceInt)

//	return payload, accessKey, nonceInt, nil
//}

func CheckAndParseJsonCTX(c context.Context, w http.ResponseWriter, r *http.Request) (interface{}, error) {
	//	var blockchainId string
	var payload interface{}

	signature := r.Header.Get("Signature")

	// Limit amount read to 512,000 bytes and parse body
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 512000))
	if err != nil {
		panic(err)
	}
	if err := r.Body.Close(); err != nil {
		panic(err)
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		ReturnUnprocessableEntity(c, w, errors.New("Unable to unmarshal body"))
	}
	log.Printf("Request received: %s", body)

	// Then look up secret and calculate digest
	accessKey := c.Value(consts.AccessKeyKey).(string)
	calculatedSignature := enulib.ComputeHmac512(body, database.GetSecretByAccessKey(accessKey))

	// If we didn't receive the expected signature then raise a forbidden
	if calculatedSignature != signature {
		errorString := fmt.Sprintf("Could not verify HMAC signature. Expected: %s, received: %s", calculatedSignature, signature)
		err := errors.New(errorString)

		return nil, err
	}

	return payload, nil
}
