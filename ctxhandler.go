package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/vennd/enu/consts"
	"github.com/vennd/enu/counterpartyhandlers"
	"github.com/vennd/enu/database"
	"github.com/vennd/enu/enulib"
	"github.com/vennd/enu/log"
	"github.com/vennd/enu/ripplehandlers"

	"github.com/vennd/enu/internal/golang.org/x/net/context"
)

/*
type appError struct {
	Error   error
	Message string
	Code    int
}
*/

type blockchainHandler func(context.Context, http.ResponseWriter, *http.Request, map[string]interface{}) *enulib.AppError
type blockchainFunction map[string]blockchainHandler

// Contains the function to call for each respective blockchain and requestType
var blockchainFunctions = map[string]blockchainFunction{
	"counterparty": {
		"address":         counterpartyhandlers.AddressCreate,
		"walletCreate":    counterpartyhandlers.WalletCreate,
		"walletPayment":   counterpartyhandlers.WalletSend,
		"walletBalance":   counterpartyhandlers.WalletBalance,
		"activateaddress": counterpartyhandlers.ActivateAddress,

		"asset":       counterpartyhandlers.AssetCreate,
		"dividend":    counterpartyhandlers.DividendCreate,
		"issuances":   counterpartyhandlers.AssetIssuances,
		"ledger":      counterpartyhandlers.AssetLedger,
		"getdividend": counterpartyhandlers.GetDividend,
		"getasset":    counterpartyhandlers.GetAsset,

		"simplepayment":    counterpartyhandlers.PaymentCreate,
		"paymentretry":     counterpartyhandlers.PaymentRetry,
		"getpayment":       counterpartyhandlers.GetPayment,
		"paymentbyaddress": counterpartyhandlers.GetPaymentsByAddress,
	},
	"ripple": {
		"walletCreate": ripplehandlers.WalletCreate,
		"dividend":     ripplehandlers.Unhandled,
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

type ctxHandler func(context.Context, http.ResponseWriter, *http.Request) *enulib.AppError

func (fn ctxHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	// Generate an requestId
	requestId := enulib.GenerateRequestId()

	// Create a context with only the requestId in it initially
	parent := context.TODO()
	ctx := context.WithValue(parent, consts.RequestIdKey, requestId)

	// Get the env we are running in
	env := os.Getenv("ENV")
	if env == "" {
		env = "dev"
	}
	ctx = context.WithValue(ctx, consts.EnvKey, env)

	log.FluentfContext(consts.LOGINFO, ctx, "%s %s entered.", r.Method, r.URL.Path)

	accessKey, err := CheckHeaderGeneric(ctx, w, r)
	if err != nil {
		return
	}

	// Add to the context the accessKey now that we know it
	ctx1 := context.WithValue(ctx, consts.AccessKeyKey, accessKey)

	// Determine blockchain this request is acting upon
	usersDefaultBlockchain := database.GetBlockchainIdByUserKey(accessKey)
	p := strings.Split(r.URL.Path, "/")
	requestBlockchainId := p[1]

	supportedBlockchains := consts.SupportedBlockchains
	sort.Strings(supportedBlockchains)

	// Search if the first part of the path after the "/" is the blockchain name. ie "/counterparty" or "/ripple"
	i := sort.SearchStrings(supportedBlockchains, requestBlockchainId)
	blockchainValid := i < len(supportedBlockchains) && supportedBlockchains[i] == requestBlockchainId

	// Search if the user has a valid default associated with their access key
	i = sort.SearchStrings(supportedBlockchains, usersDefaultBlockchain)
	userBlockchainIdValid := i < len(supportedBlockchains) && supportedBlockchains[i] == usersDefaultBlockchain

	// If the blockchain specified in the path isn't valid and a default blockchainId isn't set in the userkey then fail
	if blockchainValid == false && userBlockchainIdValid == false {
		log.FluentfContext(consts.LOGINFO, ctx, "Unsupported blockchain. Valid values: %s", strings.Join(supportedBlockchains, ", "))
		e := fmt.Sprintf("Unsupported blockchain. Valid values: %s", strings.Join(supportedBlockchains, ", "))
		ReturnServerError(ctx, w, consts.GenericErrors.UnsupportedBlockchain.Code, errors.New(e))

		return
	}

	// If the blockchain was explicitly given in the resource path, use it. Otherwise use the blockchain from the user's profile
	var blockchainId string
	if blockchainValid == true {
		blockchainId = requestBlockchainId
	} else {
		blockchainId = usersDefaultBlockchain
	}

	// Add blockchain to the context now that we know it
	ctx = context.WithValue(ctx1, consts.BlockchainIdKey, blockchainId)

	log.FluentfContext(consts.LOGINFO, ctx, "Calling context function.")

	// run function
	if e := fn(ctx, w, r); e != nil { // e is *appError, not os.Error.
		http.Error(w, e.Message, e.Code)
	}

	// Log how long it took
	log.FluentfContext(consts.LOGINFO, ctx,
		"%s,%s,%s",
		r.Method,
		r.RequestURI,
		time.Since(start),
	)

}
