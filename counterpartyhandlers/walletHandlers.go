package counterpartyhandlers

import (
	"encoding/json"
	"errors"
	"net/http"

	//	"github.com/vennd/enu/bitcoinapi"
	"github.com/vennd/enu/consts"
	//	"github.com/vennd/enu/counterpartyapi"
	"github.com/vennd/enu/counterpartycrypto"
	"github.com/vennd/enu/enulib"
	"github.com/vennd/enu/handlers"

	//	"github.com/vennd/enu/internal/github.com/gorilla/mux"
	"github.com/vennd/enu/internal/golang.org/x/net/context"
	"github.com/vennd/enu/log"
)

func WalletCreate(c context.Context, w http.ResponseWriter, r *http.Request, m map[string]interface{}) *enulib.AppError {

	var wallet counterpartycrypto.CounterpartyWallet
	var err error

	requestId := c.Value(consts.RequestIdKey).(string)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	wallet.RequestId = requestId

	var number int
	if m["numberOfAddresses"] != nil {
		number = int(m["numberOfAddresses"].(float64))
	}

	// Create the wallet
	wallet, err = counterpartycrypto.CreateWallet(number)
	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Error in CreateWallet(): %s", err.Error())
		handlers.ReturnServerError(c, w, consts.GenericErrors.GeneralError.Code, errors.New(consts.GenericErrors.GeneralError.Description))

		return nil
	}
	log.FluentfContext(consts.LOGINFO, c, "Created a new wallet with first address: %s for access key: %s\n (requestID: %s)", wallet.Addresses[0], c.Value(consts.AccessKeyKey).(string), requestId)

	// Return the wallet
	wallet.RequestId = requestId
	w.WriteHeader(http.StatusCreated)
	if err = json.NewEncoder(w).Encode(wallet); err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Error in Encode(): %s", err.Error())
		handlers.ReturnServerError(c, w, consts.GenericErrors.GeneralError.Code, errors.New(consts.GenericErrors.GeneralError.Description))

		return nil
	}

	return nil
}
