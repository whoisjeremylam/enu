package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/vennd/enu/consts"
	"github.com/vennd/enu/counterpartyapi"
	"github.com/vennd/enu/counterpartycrypto"
	"github.com/vennd/enu/database"
	"github.com/vennd/enu/enulib"

	"github.com/vennd/enu/internal/github.com/gorilla/mux"
	"github.com/vennd/enu/internal/golang.org/x/net/context"
)

func AssetCreate(c context.Context, w http.ResponseWriter, r *http.Request) *appError {

	var assetStruct enulib.Asset
	requestId := c.Value(consts.RequestIdKey).(string)
	assetStruct.RequestId = requestId
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")



// check generic args and parse
	payload, err := CheckAndParseJsonCTX(c, w, r)
	if err != nil {
		ReturnServerError(w, err)
		return nil
	}
	m := payload.(map[string]interface{})

	passphrase := m["passphrase"].(string)
	sourceAddress := m["sourceAddress"].(string)
	asset := m["asset"].(string)
	description := m["description"].(string)
	quantity := uint64(m["quantity"].(float64))
	divisible := m["divisible"].(bool)

	log.Printf("AssetCreate: received request sourceAddress: %s, asset: %s, description: %s, quantity: %s, divisible: %b from accessKey: %s\n", sourceAddress, asset, description, quantity, divisible, c.Value(consts.AccessKeyKey).(string))


// check function specific args
//	**** Need to check all the types are as expected and all required parameters received


	sourceAddressPubKey, err := counterpartycrypto.GetPublicKey(passphrase, sourceAddress)
	if err != nil {
		log.Printf("Error: %s\n", err)
		w.WriteHeader(http.StatusBadRequest)

		returnCode := enulib.ReturnCode{RequestId: requestId, Code: -3, Description: err.Error()}
		if err := json.NewEncoder(w).Encode(returnCode); err != nil {
			panic(err)
		}
		return nil
	}
	log.Printf("retrieved publickey: %s", sourceAddressPubKey)


// Generate an assetId
	assetId := enulib.GenerateAssetId()
	log.Printf("Generated assetId: %s", assetId)
	assetStruct.AssetId = assetId

	
// Return to the client the assetId and unblock the client
	w.WriteHeader(http.StatusCreated)
	if err = json.NewEncoder(w).Encode(assetStruct); err != nil {
		panic(err)
	}

// Start asset creation in async mode
	go counterpartyapi.DelegatedCreateIssuance(c.Value(consts.AccessKeyKey).(string), passphrase, sourceAddress, assetId, asset, description, quantity, divisible, requestId)

	return nil
}

func DividendCreate(c context.Context, w http.ResponseWriter, r *http.Request) *appError {

	var dividendStruct enulib.Dividend
	requestId := c.Value(consts.RequestIdKey).(string)
	dividendStruct.RequestId = requestId
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")


// check generic args and parse	
	payload, err := CheckAndParseJsonCTX(c, w, r)
	if err != nil {
		ReturnServerError(w, err)
		return nil
	}
	
	m := payload.(map[string]interface{})

	passphrase := m["passphrase"].(string)
	sourceAddress := m["sourceAddress"].(string)
	asset := m["asset"].(string)
	dividendAsset := m["dividendAsset"].(string)
	quantityPerUnit := uint64(m["quantityPerUnit"].(float64))

	log.Printf("DividendCreate: received request sourceAddress: %s, asset: %s, dividendAsset: %s, quantityPerUnit: %d from accessKey: %s\n", sourceAddress, asset, dividendAsset, quantityPerUnit, c.Value(consts.AccessKeyKey).(string))

// check function specific args
	//	**** Need to check all the types are as expected and all required parameters received

	sourceAddressPubKey, err := counterpartycrypto.GetPublicKey(passphrase, sourceAddress)
	if err != nil {
		log.Printf("Error: %s\n", err)
		w.WriteHeader(http.StatusBadRequest)

		returnCode := enulib.ReturnCode{RequestId: requestId, Code: -3, Description: err.Error()}
		if err := json.NewEncoder(w).Encode(returnCode); err != nil {
			panic(err)
		}
		return nil
	}
	log.Printf("retrieved publickey: %s", sourceAddressPubKey)



// Generate a dividendId
	dividendId := enulib.GenerateDividendId()
	log.Printf("Generated dividendId: %s", dividendId)
	dividendStruct.DividendId = dividendId


// Return to the client the assetId and unblock the client
	w.WriteHeader(http.StatusCreated)
	if err = json.NewEncoder(w).Encode(dividendStruct); err != nil {
		panic(err)
	}

// Start dividend creation in async mode
	go counterpartyapi.DelegatedCreateDividend(c.Value(consts.AccessKeyKey).(string), passphrase, dividendId, sourceAddress, asset, dividendAsset, quantityPerUnit, requestId)

	return nil
}


func AssetBalance(c context.Context, w http.ResponseWriter, r *http.Request) *appError {

	var assetBalances enulib.AssetBalances

	requestId := c.Value(consts.RequestIdKey).(string)
	assetBalances.RequestId = requestId
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

// check generic args and parse
	_, err := CheckAndParseJsonCTX(c, w, r)
	if err != nil {
		ReturnServerError(w, err)
		return nil
	}

	vars := mux.Vars(r)
	asset := vars["asset"]

	//	**** Need to check all the types are as expected and all required parameters received

	log.Printf("AssetBalance: received request asset: %s from accessKey: %s\n", asset, c.Value(consts.AccessKeyKey).(string))

	result, err := counterpartyapi.GetBalancesByAsset(asset)
	if err != nil {
		ReturnServerError(w, err)
		return nil
	}

	// Iterate and gather the balances to return
	assetBalances.Asset = asset
	for _, item := range result {
		var balance enulib.AddressAmount

		balance.Address = item.Address
		balance.Quantity = item.Quantity

		assetBalances.Balances = append(assetBalances.Balances, balance)
	}

	w.WriteHeader(http.StatusOK)
	if err = json.NewEncoder(w).Encode(assetBalances); err != nil {
		panic(err)
	}
	return nil
}

func AssetIssuances(w http.ResponseWriter, r *http.Request) {
	type issuance struct {
		BlockIndex uint64 `json:"block_index"`
		Quantity   uint64 `json:"quantity"`
		Issuer     string `json:"issuer"`
		Transfer   bool   `json:"transfer"`
	}
	var issuanceForAsset struct {
		Asset        string     `json:"asset"`
		Divisible    bool       `json:"divisible"`
		Divisibility uint64     `json:"divisibility"`
		Description  string     `json:"description"`
		Locked       bool       `json:"locked"`
		Issuances    []issuance `json:"issuances"`
	}

	vars := mux.Vars(r)
	asset := vars["asset"]

	_, accessKey, nonce, err := CheckAndParseJson(w, r)
	if err != nil {
		ReturnServerError(w, err)

		return
	}

	//	**** Need to check all the types are as expected and all required parameters received

	log.Printf("AssetIssuances: received request asset: %s from accessKey: %s\n", asset, accessKey)

	database.UpdateNonce(accessKey, nonce)
	if err != nil {
		ReturnServerError(w, err)

		return
	}

	result, err := counterpartyapi.GetIssuances(asset)
	if err != nil {
		ReturnServerError(w, err)

		return
	}

	//	log.Println(result)

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
		var issuance issuance

		issuance.BlockIndex = item.BlockIndex
		issuance.Issuer = item.Issuer
		issuance.Quantity = item.Quantity
		//		issuance.Transfer = item.Transfer

		issuanceForAsset.Issuances = append(issuanceForAsset.Issuances, issuance)
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	if err = json.NewEncoder(w).Encode(issuanceForAsset); err != nil {
		panic(err)
	}
}

// Recommended call which summarises the ledger for a particular asset
func AssetLedger(c context.Context, w http.ResponseWriter, r *http.Request) *appError {

	var assetBalances enulib.AssetBalances

	requestId := c.Value(consts.RequestIdKey).(string)
	assetBalances.RequestId = requestId
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

// check generic args and parse
	_, err := CheckAndParseJsonCTX(c, w, r)
	if err != nil {
		ReturnServerError(w, err)
		return nil
	}

	vars := mux.Vars(r)
	asset := vars["asset"]

	//	**** Need to check all the types are as expected and all required parameters received

	log.Printf("AssetLedger: received request asset: %s from accessKey: %s\n", asset, c.Value(consts.AccessKeyKey).(string))

	result, err := counterpartyapi.GetBalancesByAsset(asset)
	if err != nil {
		ReturnServerError(w, err)
		return nil
	}

	resultIssuances, err := counterpartyapi.GetIssuances(asset)
	if err != nil {
		ReturnServerError(w, err)
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


	w.WriteHeader(http.StatusOK)
	if err = json.NewEncoder(w).Encode(assetBalances); err != nil {
		panic(err)
	}
	return nil
}
