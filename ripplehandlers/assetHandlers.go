package ripplehandlers

import (
	"encoding/json"
	"net/http"

	"errors"
	"github.com/vennd/enu/consts"
	"github.com/vennd/enu/database"
	"github.com/vennd/enu/enulib"
	"github.com/vennd/enu/handlers"
	"github.com/vennd/enu/internal/github.com/gorilla/mux"
	"github.com/vennd/enu/internal/golang.org/x/net/context"
	"github.com/vennd/enu/log"
	"github.com/vennd/enu/rippleapi"
)

func AssetCreate(c context.Context, w http.ResponseWriter, r *http.Request, m map[string]interface{}) *enulib.AppError {

	var assetStruct enulib.Asset
	requestId := c.Value(consts.RequestIdKey).(string)
	assetStruct.RequestId = requestId
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	passphrase := m["passphrase"].(string)

	// sourceAddress when creating Assets is the Vennd Customer master account
	sourceAddress := "Vennd Customer Master Account"
	asset := m["asset"].(string)
	quantity := uint64(m["quantity"].(float64))
	//divisible := m["divisible"].(bool)

	// destination address is the Second master account of the customer in this case which will hold the asset
	destinationAddress := m["sourceAddress"].(string)

	log.FluentfContext(consts.LOGINFO, c, "AssetCreate: received request Address: %s, asset: %s, quantity: %s, divisible: %b from accessKey: %s\n", destinationAddress, asset, quantity, c.Value(consts.AccessKeyKey).(string))

	// to be added ripple equivalent
	/*
		sourceAddressPubKey, err := counterpartycrypto.GetPublicKey(passphrase, sourceAddress)
		if err != nil {
			log.FluentfContext(consts.LOGERROR, c, "Error: %s\n", err)

			handlers.ReturnServerError(c, w, consts.CounterpartyErrors.MiscError.Code, errors.New(consts.CounterpartyErrors.MiscError.Description))
			return nil
		}

		log.FluentfContext(consts.LOGINFO, c, "retrieved publickey: %s", sourceAddressPubKey)
	*/

	// Generate random asset name
	randomAssetName, errorCode, err := rippleapi.GenerateRandomAssetName(c)
	if err != nil {
		handlers.ReturnServerError(c, w, errorCode, err)

		return nil
	}

	// Generate an assetId
	assetId := enulib.GenerateAssetId()
	log.Printf("Generated assetId: %s", assetId)
	assetStruct.AssetId = assetId
	assetStruct.Asset = randomAssetName
	assetStruct.Description = asset
	assetStruct.Quantity = quantity
	//	assetStruct.Divisible = divisible
	assetStruct.SourceAddress = sourceAddress

	// Return to the client the assetId and unblock the client
	w.WriteHeader(http.StatusCreated)
	if err = json.NewEncoder(w).Encode(assetStruct); err != nil {
		panic(err)
	}

	// definitions for secret is passphrase and limit is quantity and asset is currency
	// in addition there is the masteraccount Vennd that holds the limits
	_, err = rippleapi.PostTrustline(c, passphrase, sourceAddress, destinationAddress, int64(quantity), randomAssetName)
	if err != nil {
		handlers.ReturnServerError(c, w, errorCode, err)

		return nil
	}

	// Start asset creation in async mode
	//	go counterpartyapi.DelegatedCreateIssuance(c, c.Value(consts.AccessKeyKey).(string), passphrase, sourceAddress, assetId, randomAssetName, asset, quantity, divisible)

	return nil
}

func GetAsset(c context.Context, w http.ResponseWriter, r *http.Request) *enulib.AppError {
	var asset enulib.Asset
	requestId := c.Value(consts.RequestIdKey).(string)
	asset.RequestId = requestId
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	vars := mux.Vars(r)
	assetId := vars["assetId"]

	if assetId == "" || len(assetId) < 16 {
		handlers.ReturnUnprocessableEntity(c, w, consts.GenericErrors.InvalidDividendId.Code, errors.New(consts.GenericErrors.InvalidDividendId.Description))

		return nil

	}

	log.FluentfContext(consts.LOGINFO, c, "GetAsset called for '%s' by '%s'\n", assetId, c.Value(consts.AccessKeyKey).(string))

	asset = database.GetAssetByAssetId(c, c.Value(consts.AccessKeyKey).(string), assetId)
	asset.RequestId = requestId

	// Add the blockchain status
	/*
		if asset.BroadcastTxId != "" {
			confirmations, err := rippleapi.GetConfirmPayment(????)
			if err == nil || confirmations == 0 {
				asset.BlockchainStatus = "unconfimed"
				asset.BlockchainConfirmations = 0
			}

			asset.BlockchainStatus = "confirmed"
			asset.BlockchainConfirmations = confirmations
		}
	*/
	if err := json.NewEncoder(w).Encode(asset); err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Error in Encode(): %s", err.Error())
		handlers.ReturnServerError(c, w, consts.GenericErrors.GeneralError.Code, errors.New(consts.GenericErrors.GeneralError.Description))

		return nil
	}

	w.WriteHeader(http.StatusOK)
	return nil
}
