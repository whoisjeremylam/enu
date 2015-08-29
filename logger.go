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
	p := strings.Split(r.URL.Path, "/")
	blockchainId := p[1]

	supportedBlockchains := consts.SupportedBlockchains
	sort.Strings(supportedBlockchains)

	i := sort.Search(len(supportedBlockchains), func(i int) bool { return supportedBlockchains[i] == blockchainId })
	if i == 0 {
		e := fmt.Sprintf("Unsupported blockchain. Valid values: %s", strings.Join(supportedBlockchains, ", "))
		ReturnServerError(ctx, w, errors.New(e))

		return
	}

	// Add accessKey and nonce to the context
	ctx1 := context.WithValue(ctx, consts.AccessKeyKey, accessKey)
	ctx = context.WithValue(ctx1, consts.NonceIntKey, nonceInt)

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
