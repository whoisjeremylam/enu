package ripplehandlers

import (
	"encoding/json"
	"errors"
	//	"fmt"
	"net/http"

	"github.com/vennd/enu/internal/github.com/gorilla/mux"
	"github.com/vennd/enu/internal/golang.org/x/net/context"

	"github.com/vennd/enu/consts"
	"github.com/vennd/enu/database"
	"github.com/vennd/enu/enulib"
	"github.com/vennd/enu/handlers"
	"github.com/vennd/enu/log"
)

func GetPayment(c context.Context, w http.ResponseWriter, r *http.Request, m map[string]interface{}) *enulib.AppError {

	var payment enulib.SimplePayment
	requestId := c.Value(consts.RequestIdKey).(string)
	payment.RequestId = requestId
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	vars := mux.Vars(r)
	paymentId := vars["paymentId"]

	if paymentId == "" || len(paymentId) < 16 {
		w.WriteHeader(http.StatusBadRequest)
		returnCode := enulib.ReturnCode{RequestId: c.Value(consts.RequestIdKey).(string), Code: -3, Description: "Incorrect paymentId"}
		if err := json.NewEncoder(w).Encode(returnCode); err != nil {
			log.FluentfContext(consts.LOGERROR, c, "Error in Encode(): %s", err.Error())
			handlers.ReturnServerError(c, w, consts.GenericErrors.GeneralError.Code, errors.New(consts.GenericErrors.GeneralError.Description))

			return nil
		}
		return nil

	}

	log.FluentfContext(consts.LOGINFO, c, "GetPayment called for '%s' by '%s'\n", paymentId, c.Value(consts.AccessKeyKey).(string))

	payment = database.GetPaymentByPaymentId(c, c.Value(consts.AccessKeyKey).(string), paymentId)
	// errorhandling here!!

	// Add the blockchain status
	//	if payment.BroadcastTxId != "" {
	//		confirmations, err := bitcoinapi.GetConfirmations(payment.BroadcastTxId)
	//		if err == nil || confirmations == 0 {
	//			payment.BlockchainStatus = "unconfimed"
	//			payment.BlockchainConfirmations = 0
	//		}

	//		payment.BlockchainStatus = "confirmed"
	//		payment.BlockchainConfirmations = confirmations
	//	}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(payment); err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Error in Encode(): %s", err.Error())
		handlers.ReturnServerError(c, w, consts.GenericErrors.GeneralError.Code, errors.New(consts.GenericErrors.GeneralError.Description))

		return nil
	}

	return nil
}
