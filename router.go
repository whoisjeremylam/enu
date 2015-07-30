package main

import (
	//	"net/http"

	"github.com/vennd/enu/internal/github.com/gorilla/mux"
)

func NewRouter() *mux.Router {
	router := mux.NewRouter().StrictSlash(true)

	router.HandleFunc("/", Logger(Index, "Index")).Methods("GET")
	router.HandleFunc("/payment", Logger(PaymentCreate, "PaymentCreate")).Methods("POST")
	router.HandleFunc("/payment/address", Logger(AddressCreate, "AddressCreate")).Methods("POST")
	router.HandleFunc("/payment/{paymentId}", Logger(GetPayment, "GetPayment")).Methods("GET")
	router.HandleFunc("/payment/status/{paymentId}", Logger(PaymentRetry, "PaymentRetry")).Methods("POST")

	router.HandleFunc("/asset", AssetCreate).Methods("POST")
	router.HandleFunc("/asset/balances/{asset}", Logger(AssetBalance, "AssetBalance")).Methods("GET")
	router.HandleFunc("/asset/dividend", Logger(DividendCreate, "DividendCreate")).Methods("POST")
	router.HandleFunc("/asset/issuances/{asset}", Logger(AssetIssuances, "AssetIssuances")).Methods("GET")
	router.HandleFunc("/asset/ledger/{asset}", Logger(AssetLedger, "AssetLedger")).Methods("GET")

	router.HandleFunc("/wallet", Logger(WalletCreate, "WalletCreate")).Methods("POST")
	router.HandleFunc("/wallet/balances/{address}", Logger(WalletBalance, "WalletBalance")).Methods("GET")
	router.HandleFunc("/wallet/payment", Logger(WalletSend, "WalletSend")).Methods("POST")

	router.HandleFunc("/blocks", Logger(GetBlocks, "GetBlocks")).Methods("GET")

	return router
}
