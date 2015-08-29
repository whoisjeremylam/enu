package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/vennd/enu/consts"
	"github.com/vennd/enu/database"
	"github.com/vennd/enu/enulib"

	"github.com/vennd/enu/internal/golang.org/x/net/context"
)

type appError struct {
	Error   error
	Message string
	Code    int
}

/*func Logger(fn http.HandlerFunc, name string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		fn(w, r)    // Call the real handler
		log.Printf( // Log how long it took
			"%s,%s,%s,%s",
			r.Method,
			r.RequestURI,
			name,
			time.Since(start),
		)
	}
}
*/

type ctxHandler func(context.Context, http.ResponseWriter, *http.Request) *appError

func (fn ctxHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	// Generate an requestId
	requestId := enulib.GenerateRequestId()
	log.Printf("Generated requestId: %s", requestId)

	// Create a context with only the requestId in it initially
	parent := context.TODO()
	ctx := context.WithValue(parent, consts.RequestIdKey, requestId)

	accessKey, nonceInt, err := CheckHeaderGeneric(ctx, w, r)
	if err != nil {
		return
	}

	log.Printf("accessKey for request: %s", accessKey)

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
		e := fmt.Sprintf("Unsupported blockchain. Valid values: %s", strings.Join(supportedBlockchains, ", "))
		ReturnServerError(ctx, w, errors.New(e))

		return
	}

	var blockchainId string
	if blockchainValid == true {
		blockchainId = requestBlockchainId
	} else {
		blockchainId = usersDefaultBlockchain
	}

	log.Printf("Operating on blockchain: %s\n", blockchainId)

	// Add accessKey and blockchainId to the context - nonce has been dropped from context since it has already been updated in the DB
	ctx1 := context.WithValue(ctx, consts.AccessKeyKey, accessKey)
	//	ctx = context.WithValue(ctx1, consts.NonceIntKey, nonceInt)
	ctx = context.WithValue(ctx1, consts.BlockchainIdKey, blockchainId)

	// run function
	if e := fn(ctx, w, r); e != nil { // e is *appError, not os.Error.
		http.Error(w, e.Message, e.Code)
	}

	log.Printf( // Log how long it took
		"%s,%s,%s,%s",
		r.Method,
		r.RequestURI,
		time.Since(start),
	)

}
