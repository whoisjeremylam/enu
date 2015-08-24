package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/vennd/enu/counterpartyapi"
	"github.com/vennd/enu/counterpartycrypto"
	"github.com/vennd/enu/database"
	"github.com/vennd/enu/enulib"
	"github.com/vennd/enu/internal/github.com/gorilla/mux"
	"golang.org/x/net/context"

)

func AssetCreate(c context.Context, w http.ResponseWriter, r *http.Request) *appError {

	var assetStruct enulib.Asset	
	requestId := c.Value(requestIdKey).(string)
	assetStruct.RequestId = requestId
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	
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

	//	**** Need to check all the types are as expected and all required parameters received

	log.Printf("AssetCreate: received request sourceAddress: %s, asset: %s, description: %s, quantity: %s, divisible: %b from accessKey: %s\n", sourceAddress, asset, description, quantity, divisible, c.Value(accessKeyKey).(string))

	database.UpdateNonce(c.Value(accessKeyKey).(string), c.Value(nonceIntKey).(int64))
	if err != nil {
		ReturnServerError(w, err)
		return nil
	}

	sourceAddressPubKey, err := counterpartycrypto.GetPublicKey(passphrase, sourceAddress)
	if err != nil {
		log.Printf("Error: %s\n", err)		
		w.WriteHeader(http.StatusBadRequest)

		returnCode := enulib.ReturnCode{RequestId: requestId ,Code: -3, Description: err.Error()}
		if err := json.NewEncoder(w).Encode(returnCode); err != nil {
			panic(err)
		}
		return nil
	}



	log.Printf("retrieved publickey: %s", sourceAddressPubKey)
	
	
	// Generate an assetId
	assetId := enulib.GenerateAssetId()
	log.Printf("Generated assetId: %s", assetId)

	// Return to the client the assetId and unblock the client

	assetStruct.AssetId = assetId
	
//	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusCreated)
	if err = json.NewEncoder(w).Encode(assetStruct); err != nil {
		panic(err)
	}

//	result, err := counterpartyapi.DelegatedCreateIssuance(passphrase, sourceAddress, asset, description, quantity, divisible)
	go counterpartyapi.DelegatedCreateIssuance(c.Value(accessKeyKey).(string), passphrase, sourceAddress, assetId, asset, description, quantity, divisible, requestId)

	return nil
}

func DividendCreate(w http.ResponseWriter, r *http.Request) {
	payload, accessKey, nonce, err := CheckAndParseJson(w, r)

	m := payload.(map[string]interface{})

	passphrase := m["passphrase"].(string)
	sourceAddress := m["sourceAddress"].(string)
	asset := m["asset"].(string)
	dividendAsset := m["dividendAsset"].(string)
	quantityPerUnit := uint64(m["quantityPerUnit"].(float64))

	//	**** Need to check all the types are as expected and all required parameters received

	log.Printf("DividendCreate: received request sourceAddress: %s, asset: %s, dividendAsset: %s, quantityPerUnit: %d from accessKey: %s\n", sourceAddress, asset, dividendAsset, quantityPerUnit, accessKey)

	database.UpdateNonce(accessKey, nonce)

	if err != nil {
		ReturnServerError(w, err)

		return
	}


	// Generate an assetId
	dividendId := enulib.GenerateDividendId()
	log.Printf("Generated dividendId: %s", dividendId)

	// Return to the client the assetId and unblock the client
	var dividendStruct enulib.Dividend
	dividendStruct.DividendId = dividendId
	
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusCreated)
	if err = json.NewEncoder(w).Encode(dividendStruct); err != nil {
		panic(err)
	} else {
		log.Println(err)
	}

//	result, err := counterpartyapi.DelegatedCreateDividend(passphrase, sourceAddress, asset, dividendAsset, quantityPerUnit)
	go counterpartyapi.DelegatedCreateDividend(accessKey, passphrase, dividendId, sourceAddress, asset, dividendAsset, quantityPerUnit)


}

func AssetBalance(w http.ResponseWriter, r *http.Request) {
	type amount struct {
		Address  string `json:"address"`
		Quantity uint64 `json:"quantity"`
	}
	var assetBalances struct {
		Asset     string   `json:"asset"`
		Divisible bool     `json:"divisible"`
		Balances  []amount `json:"balances"`
	}

	vars := mux.Vars(r)
	asset := vars["asset"]

	_, accessKey, nonce, err := CheckAndParseJson(w, r)
	if err != nil {
		ReturnServerError(w, err)

		return
	}

	//	**** Need to check all the types are as expected and all required parameters received

	log.Printf("AssetBalance: received request asset: %s from accessKey: %s\n", asset, accessKey)

	database.UpdateNonce(accessKey, nonce)
	if err != nil {
		ReturnServerError(w, err)

		return
	}

	result, err := counterpartyapi.GetBalancesByAsset(asset)
	if err != nil {
		ReturnServerError(w, err)

		return
	}

	// Iterate and gather the balances to return
	assetBalances.Asset = asset
	for _, item := range result {
		var balance amount

		balance.Address = item.Address
		balance.Quantity = item.Quantity

		assetBalances.Balances = append(assetBalances.Balances, balance)
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	if err = json.NewEncoder(w).Encode(assetBalances); err != nil {
		panic(err)
	}
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
func AssetLedger(w http.ResponseWriter, r *http.Request) {
	type amount struct {
		Address           string  `json:"address"`
		Quantity          uint64  `json:"quantity"`
		PercentageHolding float64 `json:"percentageHolding"`
	}
	var assetBalances struct {
		Asset        string   `json:"asset"`
		Locked       bool     `json:"locked"`
		Divisible    bool     `json:"divisible"`
		Divisibility uint64   `json:"divisibility"`
		Description  string   `json:"description"`
		Supply       uint64   `json:"quantity"`
		Balances     []amount `json:"balances"`
	}

	vars := mux.Vars(r)
	asset := vars["asset"]

	_, accessKey, nonce, err := CheckAndParseJson(w, r)
	if err != nil {
		ReturnServerError(w, err)

		return
	}

	//	**** Need to check all the types are as expected and all required parameters received

	log.Printf("AssetLedger: received request asset: %s from accessKey: %s\n", asset, accessKey)

	database.UpdateNonce(accessKey, nonce)
	if err != nil {
		ReturnServerError(w, err)

		return
	}

	result, err := counterpartyapi.GetBalancesByAsset(asset)
	if err != nil {
		ReturnServerError(w, err)

		return
	}

	resultIssuances, err := counterpartyapi.GetIssuances(asset)
	if err != nil {
		ReturnServerError(w, err)

		return
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
		var balance amount
		var percentage float64

		percentage = float64(item.Quantity) / float64(assetBalances.Supply) * 100

		balance.Address = item.Address
		balance.Quantity = item.Quantity
		balance.PercentageHolding = percentage

		assetBalances.Balances = append(assetBalances.Balances, balance)
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	if err = json.NewEncoder(w).Encode(assetBalances); err != nil {
		panic(err)
	}
}
