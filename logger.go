package main

import (
	"errors"
	"fmt"
	//	"log"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/vennd/enu/consts"
	"github.com/vennd/enu/database"
	"github.com/vennd/enu/enulib"
	"github.com/vennd/enu/log"

	"github.com/vennd/enu/internal/golang.org/x/net/context"
)

var sourceFile = "logger.go"

type appError struct {
	Error   error
	Message string
	Code    int
}

type ctxHandler func(context.Context, http.ResponseWriter, *http.Request) *appError

func (fn ctxHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	// Generate an requestId
	requestId := enulib.GenerateRequestId()

	// Create a context with only the requestId in it initially
	parent := context.TODO()
	ctx := context.WithValue(parent, consts.RequestIdKey, requestId)

	log.FluentfContext(sourceFile, ctx, "%s %s entered.", r.Method, r.URL.Path)

	accessKey, nonceInt, err := CheckHeaderGeneric(ctx, w, r)
	if err != nil {
		return
	}

	// Add to the context the accessKey now that we know it
	ctx1 := context.WithValue(ctx, consts.AccessKeyKey, accessKey)

	database.UpdateNonce(accessKey, nonceInt)
	if err != nil {
		ReturnServerError(ctx, w, err)

		return
	}
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
		log.FluentfContext(sourceFile, ctx, "Unsupported blockchain. Valid values: %s", strings.Join(supportedBlockchains, ", "))
		e := fmt.Sprintf("Unsupported blockchain. Valid values: %s", strings.Join(supportedBlockchains, ", "))
		ReturnServerError(ctx, w, errors.New(e))

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

	log.FluentfContext(sourceFile, ctx, "Calling context function.")

	// run function
	if e := fn(ctx, w, r); e != nil { // e is *appError, not os.Error.
		http.Error(w, e.Message, e.Code)
	}

	// Log how long it took
	log.FluentfContext(sourceFile, ctx,
		"%s,%s,%s",
		r.Method,
		r.RequestURI,
		time.Since(start),
	)

}
