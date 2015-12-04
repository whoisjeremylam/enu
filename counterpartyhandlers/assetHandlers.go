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

func GetAsset(c context.Context, w http.ResponseWriter, r *http.Request, m map[string]interface{}) *enulib.AppError {
	//	var asset enulib.Asset
	requestId := c.Value(consts.RequestIdKey).(string)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	vars := mux.Vars(r)
	assetId := vars["assetId"]

	if assetId == "" || len(assetId) < 16 {
		handlers.ReturnUnprocessableEntity(c, w, consts.GenericErrors.InvalidAssetId.Code, errors.New(consts.GenericErrors.InvalidAssetId.Description))

		return nil

	}

	log.FluentfContext(consts.LOGINFO, c, "GetAsset called for '%s' by '%s'\n", assetId, c.Value(consts.AccessKeyKey).(string))

	asset, err := database.GetAssetByAssetId(c, c.Value(consts.AccessKeyKey).(string), assetId)
	if err != nil {
		handlers.ReturnServerError(c, w, consts.GenericErrors.GeneralError.Code, errors.New(consts.GenericErrors.GeneralError.Description))

		return nil
	}
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

func DividendCreate(c context.Context, w http.ResponseWriter, r *http.Request, m map[string]interface{}) *enulib.AppError {

	var dividendStruct enulib.Dividend
	requestId := c.Value(consts.RequestIdKey).(string)
	dividendStruct.RequestId = requestId
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	// Add to the context the RequestType
	c = context.WithValue(c, consts.RequestTypeKey, "dividend")

	passphrase := m["passphrase"].(string)
	sourceAddress := m["sourceAddress"].(string)
	asset := m["asset"].(string)
	dividendAsset := m["dividendAsset"].(string)
	quantityPerUnit := uint64(m["quantityPerUnit"].(float64))

	log.FluentfContext(consts.LOGINFO, c, "DividendCreate: received request sourceAddress: %s, asset: %s, dividendAsset: %s, quantityPerUnit: %d from accessKey: %s\n", sourceAddress, asset, dividendAsset, quantityPerUnit, c.Value(consts.AccessKeyKey).(string))

	sourceAddressPubKey, err := counterpartycrypto.GetPublicKey(passphrase, sourceAddress)
	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Error: %s\n", err)
		w.WriteHeader(http.StatusBadRequest)

		returnCode := enulib.ReturnCode{RequestId: requestId, Code: -3, Description: err.Error()}
		if err := json.NewEncoder(w).Encode(returnCode); err != nil {
			log.FluentfContext(consts.LOGERROR, c, "Error in Encode(): %s", err.Error())
			handlers.ReturnServerError(c, w, consts.GenericErrors.GeneralError.Code, errors.New(consts.GenericErrors.GeneralError.Description))

			return nil
		}
		return nil
	}
	log.FluentfContext(consts.LOGINFO, c, "retrieved publickey: %s", sourceAddressPubKey)

	// Generate a dividendId
	dividendId := enulib.GenerateDividendId()
	log.FluentfContext(consts.LOGINFO, c, "Generated dividendId: %s", dividendId)
	dividendStruct.DividendId = dividendId

	dividendStruct.SourceAddress = sourceAddress
	dividendStruct.Asset = asset
	dividendStruct.DividendAsset = dividendAsset
	dividendStruct.QuantityPerUnit = quantityPerUnit

	// Return to the client the assetId and unblock the client
	w.WriteHeader(http.StatusCreated)
	if err = json.NewEncoder(w).Encode(dividendStruct); err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Error in Encode(): %s", err.Error())
		handlers.ReturnServerError(c, w, consts.GenericErrors.GeneralError.Code, errors.New(consts.GenericErrors.GeneralError.Description))

		return nil
	}

	// Start dividend creation in async mode
	go counterpartyapi.DelegatedCreateDividend(c, c.Value(consts.AccessKeyKey).(string), passphrase, dividendId, sourceAddress, asset, dividendAsset, quantityPerUnit)

	return nil
}

func AssetIssuances(c context.Context, w http.ResponseWriter, r *http.Request, m map[string]interface{}) *enulib.AppError {

	var issuanceForAsset enulib.AssetIssuances

	requestId := c.Value(consts.RequestIdKey).(string)
	issuanceForAsset.RequestId = requestId
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	vars := mux.Vars(r)
	asset := vars["asset"]

	if asset == "" || len(asset) < 5 {
		w.WriteHeader(http.StatusBadRequest)
		returnCode := enulib.ReturnCode{RequestId: c.Value(consts.RequestIdKey).(string), Code: -3, Description: "Incorrect value of asset code received in the request"}
		if err := json.NewEncoder(w).Encode(returnCode); err != nil {
			log.FluentfContext(consts.LOGERROR, c, "Error in Encode(): %s", err.Error())
			handlers.ReturnServerError(c, w, consts.GenericErrors.GeneralError.Code, errors.New(consts.GenericErrors.GeneralError.Description))

			return nil
		}
		return nil
	}

	log.FluentfContext(consts.LOGINFO, c, "AssetIssuances: received request asset: %s from accessKey: %s\n", asset, c.Value(consts.AccessKeyKey).(string))
	result, errorCode, err := counterpartyapi.GetIssuances(c, asset)
	if err != nil {
		handlers.ReturnServerError(c, w, errorCode, err)

		return nil
	}

	// Iterate and gather the balances to return
	issuanceForAsset.Asset = asset

	if len(result) > 0 {
		if result[0].Divisible == 1 { // the first valid issuance always defines divisibility
			issuanceForAsset.Divisible = true
			issuanceForAsset.Divisibility = 100000000 // always divisible to 8 decimal places for counterparty divisible assets
		} else {
			issuanceForAsset.Divisible = false
			issuanceForAsset.Divisibility = 0
		}

		// If any issuances has locked the asset, then the supply of the asset is locked
		var isLocked = false
		for _, i := range result {
			if i.Locked == 1 {
				isLocked = true
			}
		}

		issuanceForAsset.Description = result[len(result)-1].Description // get the last description on the asset
		issuanceForAsset.Locked = isLocked
	}

	for _, item := range result {
		var issuance enulib.Issuance

		issuance.BlockIndex = item.BlockIndex
		issuance.Issuer = item.Issuer
		issuance.Quantity = item.Quantity
		//		issuance.Transfer = item.Transfer

		issuanceForAsset.Issuances = append(issuanceForAsset.Issuances, issuance)
	}

	w.WriteHeader(http.StatusOK)
	if err = json.NewEncoder(w).Encode(issuanceForAsset); err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Error in Encode(): %s", err.Error())
		handlers.ReturnServerError(c, w, consts.GenericErrors.GeneralError.Code, errors.New(consts.GenericErrors.GeneralError.Description))

		return nil
	}

	return nil
}

// Recommended call which summarises the ledger for a particular asset
func AssetLedger(c context.Context, w http.ResponseWriter, r *http.Request, m map[string]interface{}) *enulib.AppError {

	var assetBalances enulib.AssetBalances

	requestId := c.Value(consts.RequestIdKey).(string)
	assetBalances.RequestId = requestId
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	vars := mux.Vars(r)
	asset := vars["asset"]

	if asset == "" || len(asset) < 5 {
		w.WriteHeader(http.StatusBadRequest)
		returnCode := enulib.ReturnCode{RequestId: c.Value(consts.RequestIdKey).(string), Code: -3, Description: "Incorrect value of asset code received in the request"}
		if err := json.NewEncoder(w).Encode(returnCode); err != nil {
			panic(err)
		}
		return nil
	}

	log.FluentfContext(consts.LOGINFO, c, "AssetLedger: received request asset: %s from accessKey: %s\n", asset, c.Value(consts.AccessKeyKey).(string))

	result, errorCode, err := counterpartyapi.GetBalancesByAsset(c, asset)
	if err != nil {
		handlers.ReturnServerError(c, w, errorCode, err)
		return nil
	}

	resultIssuances, errorCode, err := counterpartyapi.GetIssuances(c, asset)
	if err != nil {
		handlers.ReturnServerError(c, w, errorCode, err)
		return nil
	}

	// Summarise asset information
	// Calculate supply
	for _, issuanceItem := range resultIssuances {
		assetBalances.Supply += issuanceItem.Quantity
	}

	if len(resultIssuances) > 0 {
		if resultIssuances[0].Divisible == 1 { // the first valid issuance always defines divisibility
			assetBalances.Divisible = true
			assetBalances.Divisibility = 100000000 // always divisible to 8 decimal places for counterparty divisible assets
		} else {
			assetBalances.Divisible = false
			assetBalances.Divisibility = 1
		}

		// If any issuances has locked the asset, then the supply of the asset is locked
		assetBalances.Locked = false
		for _, i := range resultIssuances {
			if i.Locked == 1 {
				assetBalances.Locked = true
			}
		}

		assetBalances.Description = resultIssuances[len(resultIssuances)-1].Description // get the last description on the asset
	}

	// Iterate and gather the balances to return
	assetBalances.Asset = asset
	//	assetBalances.Supply = supply
	//	assetBalances.Divisible = divisible
	//	assetBalances.Divisibility = divisibility
	//	assetBalances.Locked = locked
	//	assetBalances.Description = description
	for _, item := range result {
		var balance enulib.AddressAmount
		var percentage float64

		percentage = float64(item.Quantity) / float64(assetBalances.Supply) * 100

		balance.Address = item.Address
		balance.Quantity = item.Quantity
		balance.PercentageHolding = percentage

		assetBalances.Balances = append(assetBalances.Balances, balance)
	}

	if err = json.NewEncoder(w).Encode(assetBalances); err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Error in Encode(): %s", err.Error())
		handlers.ReturnServerError(c, w, consts.GenericErrors.GeneralError.Code, errors.New(consts.GenericErrors.GeneralError.Description))

		return nil
	}

	w.WriteHeader(http.StatusOK)
	return nil
}

func GetDividend(c context.Context, w http.ResponseWriter, r *http.Request, m map[string]interface{}) *enulib.AppError {
	var dividend enulib.Dividend
	requestId := c.Value(consts.RequestIdKey).(string)
	dividend.RequestId = requestId
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	vars := mux.Vars(r)
	dividendId := vars["dividendId"]

	if dividendId == "" || len(dividendId) < 16 {
		handlers.ReturnUnprocessableEntity(c, w, consts.GenericErrors.InvalidDividendId.Code, errors.New(consts.GenericErrors.InvalidDividendId.Description))

		return nil

	}

	log.FluentfContext(consts.LOGINFO, c, "GetDividend called for '%s' by '%s'\n", dividendId, c.Value(consts.AccessKeyKey).(string))

	dividend, err := database.GetDividendByDividendId(c, c.Value(consts.AccessKeyKey).(string), dividendId)
	if err != nil {

	}
	dividend.RequestId = requestId

	// Add the blockchain status
	if dividend.BroadcastTxId == "" {
		dividend.BlockchainStatus = "unconfimed"
		dividend.BlockchainConfirmations = 0
	}
	if dividend.BroadcastTxId != "" {
		confirmations, err := bitcoinapi.GetConfirmations(dividend.BroadcastTxId)
		if err == nil || confirmations == 0 {
			dividend.BlockchainStatus = "unconfimed"
			dividend.BlockchainConfirmations = 0
		}

		dividend.BlockchainStatus = "confirmed"
		dividend.BlockchainConfirmations = confirmations
	}

	if err := json.NewEncoder(w).Encode(dividend); err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Error in Encode(): %s", err.Error())
		handlers.ReturnServerError(c, w, consts.GenericErrors.GeneralError.Code, errors.New(consts.GenericErrors.GeneralError.Description))

		return nil
	}

	w.WriteHeader(http.StatusOK)
	return nil
}
