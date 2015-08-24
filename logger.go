package main

import (
	"log"
	"net/http"
	"time"
	"golang.org/x/net/context"
	"github.com/vennd/enu/enulib"	

)


type appError struct {
        Error   error
        Message string
        Code    int
}

type key int

const requestIdKey key = 0
const accessKeyKey key = 1
const nonceIntKey key = 2


func Logger(fn http.HandlerFunc, name string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		fn(w, r)			// Call the real handler
		log.Printf(			// Log how long it took
			"%s,%s,%s,%s",
			r.Method,
			r.RequestURI,
			name,
			time.Since(start),
		)
	}
}

type ctxHandler func(context.Context, http.ResponseWriter, *http.Request) *appError

func (fn ctxHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {


		// Generate an requestId
		requestId := enulib.GenerateRequestId()
		log.Printf("Generated requestId: %s", requestId)

		accessKey, nonceInt, err := CheckHeaderGeneric(w, r)
		if err != nil {
			ReturnServerError(w, err)

		}


		log.Printf("Generated accessKey: %s", accessKey)
        // setup the context the way you want
        parent := context.TODO()

		ctx := context.WithValue(parent, requestIdKey, requestId)
	    ctx1 := context.WithValue(ctx, accessKeyKey, accessKey)
		ctx = context.WithValue(ctx1, nonceIntKey, nonceInt)
		
		// Check the args
//        query := r.FormValue("q")
//        if query == "" {
//                http.Error(w, "no query", http.StatusBadRequest)
//                return
 //       }

		// run function
        if e := fn(ctx, w, r); e != nil { // e is *appError, not os.Error.
                http.Error(w, e.Message, e.Code)
        }

}

