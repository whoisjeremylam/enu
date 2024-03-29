package main

import (
	"net/http"

	"github.com/whoisjeremylam/enu/internal/golang.org/x/net/context"

	"github.com/whoisjeremylam/enu/consts"
	"github.com/whoisjeremylam/enu/enulib"
)

func PaymentCreate(c context.Context, w http.ResponseWriter, r *http.Request) *enulib.AppError {
	// Add to the context the RequestType
	c = context.WithValue(c, consts.RequestTypeKey, "simplepayment")

	return handle(c, w, r)
}

func PaymentRetry(c context.Context, w http.ResponseWriter, r *http.Request) *enulib.AppError {
	// Add to the context the RequestType
	c = context.WithValue(c, consts.RequestTypeKey, "paymentretry")

	return handle(c, w, r)
}

func GetPayment(c context.Context, w http.ResponseWriter, r *http.Request) *enulib.AppError {
	// Add to the context the RequestType
	c = context.WithValue(c, consts.RequestTypeKey, "getpayment")

	return handle(c, w, r)
}

func GetPaymentsByAddress(c context.Context, w http.ResponseWriter, r *http.Request) *enulib.AppError {
	// Add to the context the RequestType
	c = context.WithValue(c, consts.RequestTypeKey, "paymentbyaddress")

	return handle(c, w, r)
}
