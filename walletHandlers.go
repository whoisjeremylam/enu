package main

import (
	"net/http"

	"github.com/vennd/enu/consts"
	"github.com/vennd/enu/counterpartyhandlers"
	"github.com/vennd/enu/enulib"

	"github.com/vennd/enu/internal/golang.org/x/net/context"
)

type blockchainHandler func(context.Context, http.ResponseWriter, *http.Request, map[string]interface{}) *enulib.AppError
type blockchainFunction map[string]blockchainHandler

// Contains the function to call for each respective blockchain and requestType
var blockchainFunctions = map[string]blockchainFunction{
	"counterparty": {
		"walletCreate":    counterpartyhandlers.WalletCreate,
		"walletPayment":   counterpartyhandlers.WalletSend,
		"walletBalance":   counterpartyhandlers.WalletBalance,
		"activateaddress": counterpartyhandlers.ActivateAddress,
	},
	"ripple": {
		// For example
		"walletCreate":    counterpartyhandlers.WalletCreate,
		"walletPayment":   counterpartyhandlers.WalletSend,
		"walletBalance":   counterpartyhandlers.WalletBalance,
		"activateaddress": counterpartyhandlers.ActivateAddress,
	},
}

func handle(c context.Context, w http.ResponseWriter, r *http.Request) *enulib.AppError {
	// check generic args and parse
	m, err := CheckAndParseJsonCTX(c, w, r)
	if err != nil {
		// Status errors are handled inside CheckAndParseJsonCTX, so we just exit gracefully
		return nil
	}

	// Look up the required function in the map and execute
	blockchainId := c.Value(consts.BlockchainIdKey).(string)
	requestType := c.Value(consts.RequestTypeKey).(string)
	blockchainFunctions[blockchainId][requestType](c, w, r, m)

	return nil
}

func WalletCreate(c context.Context, w http.ResponseWriter, r *http.Request) *enulib.AppError {
	// Add to the context the RequestType
	c = context.WithValue(c, consts.RequestTypeKey, "walletCreate")

	return handle(c, w, r)
}

func WalletBalance(c context.Context, w http.ResponseWriter, r *http.Request) *enulib.AppError {
	// Add to the context the RequestType
	c = context.WithValue(c, consts.RequestTypeKey, "walletBalance")

	return handle(c, w, r)
}

func WalletSend(c context.Context, w http.ResponseWriter, r *http.Request) *enulib.AppError {
	// Add to the context the RequestType
	c = context.WithValue(c, consts.RequestTypeKey, "walletPayment")

	return handle(c, w, r)
}

func ActivateAddress(c context.Context, w http.ResponseWriter, r *http.Request) *enulib.AppError {
	// Add to the context the RequestType
	c = context.WithValue(c, consts.RequestTypeKey, "activateaddress")

	return handle(c, w, r)
}
