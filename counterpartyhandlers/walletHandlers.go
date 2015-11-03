package counterpartyhandlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/vennd/enu/bitcoinapi"
	"github.com/vennd/enu/consts"
	"github.com/vennd/enu/counterpartyapi"
	"github.com/vennd/enu/counterpartycrypto"

	"github.com/vennd/enu/enulib"

	"github.com/vennd/enu/internal/github.com/gorilla/mux"
	"github.com/vennd/enu/internal/golang.org/x/net/context"
	"github.com/vennd/enu/log"
)

func WalletCreate(c context.Context, w http.ResponseWriter, r *http.Request, m map[string]interface{}) *enulib.AppError {

	var wallet counterpartycrypto.CounterpartyWallet
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
		ReturnServerError(c, w, consts.GenericErrors.GeneralError.Code, errors.New(consts.GenericErrors.GeneralError.Description))

		return nil
	}
	log.FluentfContext(consts.LOGINFO, c, "Created a new wallet with first address: %s for access key: %s\n (requestID: %s)", wallet.Addresses[0], c.Value(consts.AccessKeyKey).(string), requestId)

	// Return the wallet
	wallet.RequestId = requestId
	w.WriteHeader(http.StatusCreated)
	if err = json.NewEncoder(w).Encode(wallet); err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Error in Encode(): %s", err.Error())
		ReturnServerError(c, w, consts.GenericErrors.GeneralError.Code, errors.New(consts.GenericErrors.GeneralError.Description))

		return nil
	}

	return nil
}

func WalletSend(c context.Context, w http.ResponseWriter, r *http.Request) *enulib.AppError {

	var walletPayment enulib.WalletPayment
	var paymentTag string

	requestId := c.Value(consts.RequestIdKey).(string)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	walletPayment.RequestId = requestId

	// Add to the context the RequestType
	c = context.WithValue(c, consts.RequestTypeKey, "walletPayment")

	// check generic args and parse
	m, err := CheckAndParseJsonCTX(c, w, r)
	if err != nil {
		// Status errors are handled inside CheckAndParseJsonCTX, so we just exit gracefully
		return nil
	}

	passphrase := m["passphrase"].(string)
	sourceAddress := m["sourceAddress"].(string)
	destinationAddress := m["destinationAddress"].(string)
	asset := m["asset"].(string)
	quantity := uint64(m["quantity"].(float64))

	if m["paymentTag"] != nil {
		paymentTag = m["paymentTag"].(string)
	}

	log.FluentfContext(consts.LOGINFO, c, "WalletSend: received request sourceAddress: %s, destinationAddress: %s, asset: %s, quantity: %d, paymentTag: %s from accessKey: %s\n", sourceAddress, destinationAddress, asset, quantity, c.Value(consts.AccessKeyKey).(string), paymentTag)
	// Generate a paymentId
	paymentId := enulib.GeneratePaymentId()

	log.FluentfContext(consts.LOGINFO, c, "Generated paymentId: %s", paymentId)

	// Return to the client the walletPayment containing requestId and paymentId and unblock the client
	walletPayment.PaymentId = paymentId
	walletPayment.Asset = asset
	walletPayment.SourceAddress = sourceAddress
	walletPayment.DestinationAddress = destinationAddress
	walletPayment.Quantity = quantity
	w.WriteHeader(http.StatusCreated)

	if err = json.NewEncoder(w).Encode(walletPayment); err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Error in Encode(): %s", err.Error())
		ReturnServerError(c, w, consts.GenericErrors.GeneralError.Code, errors.New(consts.GenericErrors.GeneralError.Description))

		return nil
	}

	go counterpartyapi.DelegatedSend(c, c.Value(consts.AccessKeyKey).(string), passphrase, sourceAddress, destinationAddress, asset, quantity, paymentId, paymentTag)

	return nil
}

func WalletBalance(c context.Context, w http.ResponseWriter, r *http.Request) *enulib.AppError {

	var walletbalance enulib.AddressBalances

	requestId := c.Value(consts.RequestIdKey).(string)
	walletbalance.RequestId = requestId
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	// check generic args and parse
	_, err := CheckAndParseJsonCTX(c, w, r)
	if err != nil {
		// Status errors are handled inside CheckAndParseJsonCTX, so we just exit gracefully
		return nil
	}

	vars := mux.Vars(r)
	address := vars["address"]

	if address == "" || len(address) != 34 {
		w.WriteHeader(http.StatusBadRequest)
		returnCode := enulib.ReturnCode{RequestId: c.Value(consts.RequestIdKey).(string), Code: -3, Description: "Incorrect value of address received in the request"}
		if err := json.NewEncoder(w).Encode(returnCode); err != nil {
			log.FluentfContext(consts.LOGERROR, c, "Error in Encode(): %s", err.Error())
			ReturnServerError(c, w, consts.GenericErrors.GeneralError.Code, errors.New(consts.GenericErrors.GeneralError.Description))

			return nil
		}

		return nil
	}

	log.FluentfContext(consts.LOGINFO, c, "WalletBalance: received request address: %s from accessKey: %s\n", address, c.Value(consts.AccessKeyKey).(string))

	// Get counterparty balances
	result, errorCode, err := counterpartyapi.GetBalancesByAddress(c, address)
	if err != nil {
		ReturnServerError(c, w, errorCode, err)
		return nil
	}

	// Iterate and gather the balances to return
	walletbalance.Address = address
	walletbalance.BlockchainId = consts.CounterpartyBlockchainId
	for _, item := range result {
		var balance enulib.Amount

		balance.Asset = item.Asset
		balance.Quantity = item.Quantity

		walletbalance.Balances = append(walletbalance.Balances, balance)
	}

	// Add BTC balances
	btcbalance, err := bitcoinapi.GetBalance(c, address)
	walletbalance.Balances = append(walletbalance.Balances, enulib.Amount{Asset: "BTC", Quantity: btcbalance})

	// Calculate number of transactions possible
	numberOfTransactions, err := counterpartyapi.CalculateNumberOfTransactions(c, btcbalance)
	if err != nil {
		numberOfTransactions = 0
		log.FluentfContext(consts.LOGERROR, c, "Unable to calculate number of transactions: %s", err.Error())
	}
	walletbalance.NumberOfTransactions = numberOfTransactions

	w.WriteHeader(http.StatusOK)
	if err = json.NewEncoder(w).Encode(walletbalance); err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Error in Encode(): %s", err.Error())
		ReturnServerError(c, w, consts.GenericErrors.GeneralError.Code, errors.New(consts.GenericErrors.GeneralError.Description))

		return nil
	}

	return nil
}

func ActivateAddress(c context.Context, w http.ResponseWriter, r *http.Request) *enulib.AppError {
	requestId := c.Value(consts.RequestIdKey).(string)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	// Add to the context the RequestType
	c = context.WithValue(c, consts.RequestTypeKey, "activateaddress")

	// check generic args and parse
	m, err := CheckAndParseJsonCTX(c, w, r)
	if err != nil {
		// Status errors are handled inside CheckAndParseJsonCTX, so we just exit gracefully
		return nil
	}

	vars := mux.Vars(r)
	address := vars["address"]

	if address == "" || len(address) != 34 {
		w.WriteHeader(http.StatusBadRequest)
		returnCode := enulib.ReturnCode{RequestId: c.Value(consts.RequestIdKey).(string), Code: -3, Description: "Incorrect value of address received in the request"}
		if err := json.NewEncoder(w).Encode(returnCode); err != nil {
			log.FluentfContext(consts.LOGERROR, c, "Error in Encode(): %s", err.Error())
			ReturnServerError(c, w, consts.GenericErrors.GeneralError.Code, errors.New(consts.GenericErrors.GeneralError.Description))

			return nil
		}
		return nil

	}

	// Get the amount from the URL
	var amount uint64
	if m["amount"] == nil {
		amount = consts.CounterpartyAddressActivationAmount
	} else {
		amount = uint64(m["amount"].(float64))
	}

	log.FluentfContext(consts.LOGINFO, c, "ActivateAddress: received request address to activate: %s, number of transactions to activate: %d", address, amount)
	// Generate an activationId
	activationId := enulib.GenerateActivationId()

	log.FluentfContext(consts.LOGINFO, c, "Generated activationId: %s", activationId)

	// Return to the client the activationId and requestId and unblock the client
	var result = map[string]interface{}{
		"address":       address,
		"amount":        amount,
		"activationId":  activationId,
		"broadcastTxId": "",
		"status":        "valid",
		"errorMessage":  "",
		"requestId":     requestId,
	}

	w.WriteHeader(http.StatusCreated)
	if err = json.NewEncoder(w).Encode(result); err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Error in Encode(): %s", err.Error())
		ReturnServerError(c, w, consts.GenericErrors.GeneralError.Code, errors.New(consts.GenericErrors.GeneralError.Description))

		return nil
	}

	go counterpartyapi.DelegatedActivateAddress(c, address, amount, activationId)

	return nil
}
