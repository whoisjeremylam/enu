package main

import (
	//	"net/http"

	"github.com/vennd/enu/internal/github.com/gorilla/mux"
)

func NewRouter() *mux.Router {
	router := mux.NewRouter().StrictSlash(true)

	router.HandleFunc("/", Index).Methods("GET")
	router.Handle("/payment", ctxHandler(PaymentCreate)).Methods("POST")
	router.Handle("/payment/address", ctxHandler(AddressCreate)).Methods("POST")
	router.Handle("/payment/{paymentId}", ctxHandler(GetPayment)).Methods("GET")
	router.Handle("/payment/status/{paymentId}", ctxHandler(PaymentRetry)).Methods("POST")

	router.Handle("/asset", ctxHandler(AssetCreate)).Methods("POST")
	router.Handle("/asset/balances/{asset}", ctxHandler(AssetBalance)).Methods("GET")
	router.Handle("/asset/dividend", ctxHandler(DividendCreate)).Methods("POST")
	router.Handle("/asset/issuances/{asset}", ctxHandler(AssetIssuances)).Methods("GET")
	router.Handle("/asset/ledger/{asset}", ctxHandler(AssetLedger)).Methods("GET")

	router.Handle("/counterparty/asset", ctxHandler(AssetCreate)).Methods("POST")

	router.Handle("/wallet", ctxHandler(WalletCreate)).Methods("POST")
	router.Handle("/wallet/balances/{address}", ctxHandler(WalletBalance)).Methods("GET")
	router.Handle("/wallet/payment", ctxHandler(WalletSend)).Methods("POST")
	router.Handle("/wallet/payment/{paymentId}", ctxHandler(GetPayment)).Methods("GET")

	router.Handle("/counterparty/wallet", ctxHandler(WalletCreate)).Methods("POST")

	router.Handle("/blocks", ctxHandler(GetBlocks)).Methods("GET")

	return router
}
