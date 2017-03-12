package main

import (
	"net/http"

	"github.com/whoisjeremylam/enu/consts"
	"github.com/whoisjeremylam/enu/enulib"
	"github.com/whoisjeremylam/enu/internal/golang.org/x/net/context"
)

func AddressCreate(c context.Context, w http.ResponseWriter, r *http.Request) *enulib.AppError {
	// Add to the context the RequestType
	c = context.WithValue(c, consts.RequestTypeKey, "address")

	return handle(c, w, r)
}
