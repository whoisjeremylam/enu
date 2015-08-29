package main

import (
	//	"net/http"

	"github.com/vennd/enu/internal/github.com/gorilla/mux"
)

func NewRouter() *mux.Router {
	router := mux.NewRouter().StrictSlash(true)

	router.HandleFunc("/", Logger(Index, "Index")).Methods("GET")
	router.Handle("/payment", ctxHandler(PaymentCreate)).Methods("POST")
	router.Handle("/payment/address", ctxHandler(AddressCreate)).Methods("POST")
	router.HandleFunc("/payment/{paymentId}", Logger(GetPayment, "GetPayment")).Methods("GET")
	router.HandleFunc("/payment/status/{paymentId}", Logger(PaymentRetry, "PaymentRetry")).Methods("POST")

	router.Handle("/asset", ctxHandler(AssetCreate)).Methods("POST")
	router.Handle("/asset/balances/{asset}", ctxHandler(AssetBalance)).Methods("GET")
	router.Handle("/asset/dividend", ctxHandler(DividendCreate)).Methods("POST")
	router.HandleFunc("/asset/issuances/{asset}", Logger(AssetIssuances, "AssetIssuances")).Methods("GET")
	router.Handle("/asset/ledger/{asset}", ctxHandler(AssetLedger)).Methods("GET")

	router.Handle("/wallet", ctxHandler(WalletCreate)).Methods("POST")
	router.Handle("/wallet/balances/{address}", ctxHandler(WalletBalance)).Methods("GET")
	router.HandleFunc("/wallet/payment", Logger(WalletSend, "WalletSend")).Methods("POST")

	router.HandleFunc("/blocks", Logger(GetBlocks, "GetBlocks")).Methods("GET")

	return router
}
