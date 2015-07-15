package main

import (
	//	"net/http"

	"github.com/gorilla/mux"
)

func NewRouter() *mux.Router {
	router := mux.NewRouter().StrictSlash(true)

	router.HandleFunc("/", Logger(Index, "Index")).Methods("GET")
	router.HandleFunc("/payment", PaymentCreate).Methods("POST")
	router.HandleFunc("/payment/address", AddressCreate).Methods("POST")
	router.HandleFunc("/payment/{paymentId}", GetPayment).Methods("GET")
	router.HandleFunc("/payment/status/{paymentId}", PaymentRetry).Methods("POST")

	router.HandleFunc("/asset", AssetCreate).Methods("POST")
	router.HandleFunc("/asset/balances/{asset}", AssetBalance).Methods("GET")
	router.HandleFunc("/asset/dividend", DividendCreate).Methods("POST")
	router.HandleFunc("/asset/issuances/{asset}", AssetIssuances).Methods("GET")
	router.HandleFunc("/asset/ledger/{asset}", AssetLedger).Methods("GET")

	router.HandleFunc("/wallet", Logger(WalletCreate, "WalletCreate")).Methods("POST")
	router.HandleFunc("/wallet/balances/{address}", WalletBalance).Methods("GET")
	router.HandleFunc("/wallet/payment", Logger(WalletSend, "WalletSend")).Methods("POST")

	router.HandleFunc("/blocks", GetBlocks).Methods("GET")

	return router
}
