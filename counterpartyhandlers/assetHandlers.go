package counterpartyhandlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/vennd/enu/bitcoinapi"
	"github.com/vennd/enu/consts"
	"github.com/vennd/enu/counterpartyapi"
	"github.com/vennd/enu/counterpartycrypto"
	"github.com/vennd/enu/database"
	"github.com/vennd/enu/enulib"
	"github.com/vennd/enu/handlers"
	"github.com/vennd/enu/internal/github.com/gorilla/mux"
	"github.com/vennd/enu/internal/golang.org/x/net/context"
	"github.com/vennd/enu/log"
)

func AssetCreate(c context.Context, w http.ResponseWriter, r *http.Request, m map[string]interface{}) *enulib.AppError {

	var assetStruct enulib.Asset
	requestId := c.Value(consts.RequestIdKey).(string)
	assetStruct.RequestId = requestId
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	passphrase := m["passphrase"].(string)
	sourceAddress := m["sourceAddress"].(string)
	asset := m["asset"].(string)
	//	description := m["description"].(string)
	quantity := uint64(m["quantity"].(float64))
	divisible := m["divisible"].(bool)

	log.FluentfContext(consts.LOGINFO, c, "AssetCreate: received request sourceAddress: %s, asset: %s, quantity: %s, divisible: %b from accessKey: %s\n", sourceAddress, asset, quantity, divisible, c.Value(consts.AccessKeyKey).(string))

	sourceAddressPubKey, err := counterpartycrypto.GetPublicKey(passphrase, sourceAddress)
	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Error: %s\n", err)

		handlers.ReturnServerError(c, w, consts.CounterpartyErrors.MiscError.Code, errors.New(consts.CounterpartyErrors.MiscError.Description))
		return nil
	}

	log.FluentfContext(consts.LOGINFO, c, "retrieved publickey: %s", sourceAddressPubKey)

	// Generate random asset name
	randomAssetName, errorCode, err := counterpartyapi.GenerateRandomAssetName(c)
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
	assetStruct.Divisible = divisible
	assetStruct.SourceAddress = sourceAddress

	// Return to the client the assetId and unblock the client
	w.WriteHeader(http.StatusCreated)
	if err = json.NewEncoder(w).Encode(assetStruct); err != nil {
		panic(err)
	}

	// Start asset creation in async mode
	go counterpartyapi.DelegatedCreateIssuance(c, c.Value(consts.AccessKeyKey).(string), passphrase, sourceAddress, assetId, randomAssetName, asset, quantity, divisible)

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
	if asset.BroadcastTxId != "" {
		confirmations, err := bitcoinapi.GetConfirmations(asset.BroadcastTxId)
		if err == nil || confirmations == 0 {
			asset.BlockchainStatus = "unconfimed"
			asset.BlockchainConfirmations = 0
		}

		asset.BlockchainStatus = "confirmed"
		asset.BlockchainConfirmations = confirmations
	}

	if err := json.NewEncoder(w).Encode(asset); err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Error in Encode(): %s", err.Error())
		handlers.ReturnServerError(c, w, consts.GenericErrors.GeneralError.Code, errors.New(consts.GenericErrors.GeneralError.Description))

		return nil
	}

	w.WriteHeader(http.StatusOK)
	return nil
}
