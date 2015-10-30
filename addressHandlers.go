package main

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/vennd/enu/bitcoinapi"
	"github.com/vennd/enu/consts"
	"github.com/vennd/enu/database"
	"github.com/vennd/enu/enulib"
	"github.com/vennd/enu/internal/golang.org/x/net/context"
	"github.com/vennd/enu/log"
)

func AddressCreate(c context.Context, w http.ResponseWriter, r *http.Request) *enulib.AppError {
	var address enulib.Address
	requestId := c.Value(consts.RequestIdKey).(string)
	address.RequestId = requestId
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	_, err := CheckAndParseJsonCTX(c, w, r)
	if err != nil {
		// Status errors are handled inside CheckAndParseJsonCTX, so we just exit gracefully
		return nil
	}

	// Create the address
	newAddress, err := bitcoinapi.GetNewAddress()
	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Unable to create a new bitcoin address. Error: %s", err.Error())
		ReturnServerError(c, w, consts.GenericErrors.GeneralError.Code, errors.New(consts.GenericErrors.GeneralError.Description))
		return nil
	}
	address.Value = newAddress

	err2 := database.CreateSecondaryAddress(c, c.Value(consts.AccessKeyKey).(string), newAddress)
	if err2 != nil {
		log.FluentfContext(consts.LOGERROR, c, "Unable to persist new address to database. Error: %s\n", err2.Error())
	} else {
		log.FluentfContext(consts.LOGINFO, c, "Created secondary address: %s for access key: %s\n", newAddress, c.Value(consts.AccessKeyKey).(string))
		w.WriteHeader(http.StatusCreated)
		if err = json.NewEncoder(w).Encode(address); err != nil {
			log.FluentfContext(consts.LOGERROR, c, "Error in Encode(): %s", err.Error())
			ReturnServerError(c, w, consts.GenericErrors.GeneralError.Code, errors.New(consts.GenericErrors.GeneralError.Description))

			return nil
		}

		//		ReturnCreated(w)
	}

	return nil

}
