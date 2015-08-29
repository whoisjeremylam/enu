package main

import (
	"log"
	"net/http"
	"time"

	"github.com/vennd/enu/consts"
	"github.com/vennd/enu/enulib"
	"github.com/vennd/enu/database"	

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

	accessKey, nonceInt, err := CheckHeaderGeneric(w, r)
	if err != nil {
		ReturnServerError(w, err)
	}

	log.Printf("Generated accessKey: %s", accessKey)

	database.UpdateNonce(accessKey, nonceInt)
	if err != nil {
		ReturnServerError(w, err)

	}

	// setup the context the way you want
	parent := context.TODO()

	ctx := context.WithValue(parent, consts.RequestIdKey, requestId)
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
