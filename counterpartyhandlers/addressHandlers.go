package counterpartyhandlers

import (
	"encoding/json"
	"net/http"

	"github.com/whoisjeremylam/enu/bitcoinapi"
	"github.com/whoisjeremylam/enu/consts"
	"github.com/whoisjeremylam/enu/database"
	"github.com/whoisjeremylam/enu/enulib"
	"github.com/whoisjeremylam/enu/handlers"
	"github.com/whoisjeremylam/enu/internal/golang.org/x/net/context"
	"github.com/whoisjeremylam/enu/log"
)

func AddressCreate(c context.Context, w http.ResponseWriter, r *http.Request, m map[string]interface{}) *enulib.AppError {
	var address enulib.Address
	requestId := c.Value(consts.RequestIdKey).(string)
	address.RequestId = requestId
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	// Create the address
	newAddress, err := bitcoinapi.GetNewAddress()
	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Unable to create a new bitcoin address. Error: %s", err.Error())
		handlers.ReturnServerError(c, w)
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
			handlers.ReturnServerError(c, w)

			return nil
		}

		//		ReturnCreated(w)
	}

	return nil

}
