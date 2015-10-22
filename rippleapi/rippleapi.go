package rippleapi

import (
	"bytes"

	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	//	"math/big"
	"net/http"
	"os"
	"strconv"
	//	"sync"

	"github.com/vennd/enu/consts"
	"github.com/vennd/enu/log"

	"github.com/vennd/enu/internal/github.com/gorilla/securecookie"
	"github.com/vennd/enu/internal/golang.org/x/net/context"
)

type Wallet struct {
	Address string `json:"address"`
	Secret  string `json:"secret"`
}

type NewWallet struct {
	NwWallet Wallet `json:"wallet"`
	Success  bool   `json:"success"`
}

type AccountSettings struct {
	AccSettings Settings `json:"settings"`
	Success     bool     `json:"success"`
}

type Settings struct {
	Account                 string `json:"account"`
	Transfer_rate           string `json:"transfer_rate"`
	Password_spent          bool   `json:"password_spent"`
	Require_destination_tag bool   `json:"require_destination_tag"`
	Require_authorization   bool   `json:"require_authorization"`
	Disallow_xrp            bool   `json:"disallow_xrp"`
	Disable_master          bool   `json:"disable_master"`
	No_freeze               bool   `json:"no_freeze"`
	Global_freeze           bool   `json:"global_freeze"`
	Default_ripple          bool   `json:"default_ripple"`
	Transaction_sequence    string `json:"transaction_sequence"`
	Email_hash              string `json:"email_hash"`
	Wallet_locator          string `json:"wallet_locator"`
	Wallet_size             string `json:"wallet_size"`
	Message_key             string `json:"message_key"`
	Domain                  string `json:"domain"`
	Signers                 string `json:"signers"`
}

type AccountBalance struct {
	Ledger    int64     `json:"ledger"`
	Validated bool      `json:"validated"`
	Balances  []Balance `json:"balances"`
	Success   bool      `json:"success"`
}

type Balance struct {
	Value        string `json:"value"`
	Currency     string `json:"currency"`
	Counterparty string `json:"counterparty"`
}

type Amount struct {
	Value    string `json:"value"`
	Currency string `json:"currency"`
	Issuer   string `json:"issuer"`
}

type PreparePaymentList struct {
	Payments []Payment `json:"payments"`
	Success  bool      `json:"success"`
}

type PaymentList struct {
	Secret             string  `json:"secret"`
	Client_resource_id string  `json:"client_resource_id"`
	Payments           Payment `json:"payment"`
}

type Payment struct {
	Source_account      string `json:"source_account"`
	Source_tag          string `json:"source_tag"`
	Source_amount       Amount `json:"source_amount"`
	Source_slippage     string `json:"source_slippage"`
	Destination_account string `json:"destination_account"`
	Destination_tag     string `json:"destination_tag"`
	Destination_amount  Amount `json:"destination_amount"`
	Invoice_id          string `json:"invoice_id"`
	Paths               string `json:"paths"`
	Partial_payment     bool   `json:"partial_payment"`
	No_direct_ripple    bool   `json:"no_direct_ripple"`
}

type ConfirmPayment struct {
	Destination_balance_changes Amount `json:"destination_balance_changes"`
}

type TrustlineList struct {
	Secret     string    `json:"secret"`
	Trustlines Trustline `json:"trustline"`
}

type Trustline struct {
	Account                  string `json:"account"`
	Limit                    string `json:"limit"`
	Currency                 string `json:"currency"`
	Counterparty             string `json:"counterparty"`
	Account_allows_rippling  bool   `json:"account_allows_rippling"`
	Account_trustline_frozen bool   `json:"account_trustline_frozen"`
	State                    string `json:"state"`
	Legder                   string `json:"ledger"`
	Hash                     string `json:"hash"`
}

type TrustlineResult struct {
	Success    bool      `json:"success"`
	Trustlines Trustline `json:"trustline"`
}

type GetTrustlinesResult struct {
	Ledger        int64          `json:"ledger"`
	Validated     bool           `json:"validated"`
	GetTrustLines []GetTrustline `json:"trustlines"`
	Success       bool           `json:"success"`
}

type GetTrustline struct {
	Account                       string `json:"account"`
	Counterparty                  string `json:"counterparty"`
	Currency                      string `json:"currency"`
	Limit                         string `json:"trustlimit"`
	Reciprocated_limit            string `json:"reciprocated_trust_limit"`
	Account_allows_rippling       bool   `json:"account_allows_rippling"`
	Counterparty_allows_rippling  bool   `json:"counterparty_allows_rippling"`
	Account_trustline_frozen      bool   `json:"account_trustline_frozen"`
	Counterparty_trustline_frozen bool   `json:"counterparty_trustline_frozen"`
}

// Initialises global variables and database connection for all handlers
var isInit bool = false // set to true only after the init sequence is complete
var rippleHost string

func Init() {
	var configFilePath string

	if isInit == true {
		return
	}

	if _, err := os.Stat("./enuapi.json"); err == nil {
		//		log.Println("Found and using configuration file ./rippleapi.json")
		configFilePath = "./enuapi.json"
	} else {
		if _, err := os.Stat(os.Getenv("GOPATH") + "/bin/enuapi.json"); err == nil {
			configFilePath = os.Getenv("GOPATH") + "/bin/enuapi.json"
			//			log.Printf("Found and using configuration file from GOPATH: %s\n", configFilePath)

		} else {
			if _, err := os.Stat(os.Getenv("GOPATH") + "/src/github.com/vennd/enu/enuapi.json"); err == nil {
				configFilePath = os.Getenv("GOPATH") + "/src/github.com/vennd/enu/enuapi.json"
				//				log.Printf("Found and using configuration file from GOPATH: %s\n", configFilePath)
			} else {
				log.Println("Cannot find enuapi.json")
				os.Exit(-100)
			}
		}
	}

	InitWithConfigPath(configFilePath)
}

func InitWithConfigPath(configFilePath string) {
	var configuration interface{}

	if isInit == true {
		return
	}

	// Read configuration from file
	//	log.Printf("Reading %s\n", configFilePath)
	file, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		log.Println("Unable to read configuration file enuapi.json")
		log.Println(err.Error())
		os.Exit(-101)
	}

	err = json.Unmarshal(file, &configuration)

	if err != nil {
		log.Println("Unable to parse enuapi.json")
		log.Println(err.Error())
		os.Exit(-101)
	}

	m := configuration.(map[string]interface{})

	// Ripple API parameters
	rippleHost = m["rippleHost"].(string) // End point for JSON RPC server

	isInit = true
}

func httpGet(c context.Context, url string) ([]byte, int64, error) {
	// Set headers

	req, err := http.NewRequest("GET", rippleHost+url, nil)

	clientPointer := &http.Client{}
	resp, err := clientPointer.Do(req)

	if err != nil {
		log.FluentfContext(consts.LOGDEBUG, c, "Request failed. %s", err.Error())
		return nil, 0, err
	}

	if resp.StatusCode != 200 {
		log.FluentfContext(consts.LOGDEBUG, c, "Request failed. Status code: %d\n", resp.StatusCode)

		body, err := ioutil.ReadAll(resp.Body)
		defer resp.Body.Close()

		if err != nil {
			return nil, 0, err
		}

		log.FluentfContext(consts.LOGDEBUG, c, "Reply: %s", string(body))

		return body, -1000, errors.New(string(body))

	}

	body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	if err != nil {
		return nil, 0, err
	}

	return body, 0, nil
}

func postAPI(c context.Context, request string, postData []byte) ([]byte, int64, error) {
	postDataJson := string(postData)

	log.FluentfContext(consts.LOGDEBUG, c, "rippleapi postAPI() posting: %s", postDataJson)

	// Set headers
	req, err := http.NewRequest("POST", rippleHost+request, bytes.NewBufferString(postDataJson))
	req.Header.Set("Content-Type", "application/json")

	clientPointer := &http.Client{}
	resp, err := clientPointer.Do(req)

	if err != nil {
		return nil, 0, err
	}

	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		log.FluentfContext(consts.LOGDEBUG, c, "Request failed. Status code: %d\n", resp.StatusCode)

		body, err := ioutil.ReadAll(resp.Body)
		defer resp.Body.Close()

		if err != nil {
			return nil, 0, err
		}

		log.FluentfContext(consts.LOGDEBUG, c, "Reply: %s", string(body))

		return nil, -1000, errors.New(string(body))
	}

	body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	if err != nil {
		return nil, 0, err
	}

	return body, 0, nil
}

func generateId(c context.Context) uint32 {
	buf := securecookie.GenerateRandomKey(4)
	randomUint64, err := strconv.ParseUint(hex.EncodeToString(buf), 16, 32)

	if err != nil {
		panic(err)
	}

	randomUint32 := uint32(randomUint64)

	return randomUint32
}

func GetServerStatus(c context.Context) ([]byte, error) {
	if isInit == false {
		Init()
	}

	var requestString string = "/v1/server"

	result, status, err := httpGet(c, requestString)
	if result == nil {
		return result, errors.New("Ripple unavailable")
	}

	if status != 0 {
		log.FluentfContext(consts.LOGERROR, c, string(result))
		return result, err
	}

	println(string(result))

	return result, nil
}

func CreateWallet(c context.Context) (NewWallet, error) {
	if isInit == false {
		Init()
	}

	var r NewWallet

	result, status, err := httpGet(c, "/v1/wallet/new")

	println(string(result))

	if result == nil {
		return r, errors.New("Ripple unavailable")
	}

	if status != 0 {
		log.FluentfContext(consts.LOGERROR, c, string(result))
		return r, err
	}

	if err := json.Unmarshal(result, &r); err != nil {
		log.FluentfContext(consts.LOGERROR, c, err.Error())
		return r, err
	}

	if !r.Success {
		return r, err
	}

	return r, nil
}

func GetAccountSettings(c context.Context, account string) (AccountSettings, error) {
	if isInit == false {
		Init()
	}

	var r AccountSettings

	result, status, err := httpGet(c, "/v1/accounts/"+account+"/settings")

	if result == nil {
		return r, errors.New("Ripple unavailable")
	}

	if status != 0 {
		log.FluentfContext(consts.LOGERROR, c, string(result))
		return r, err
	}

	//	var r interface{}
	println(string(result))

	if err := json.Unmarshal(result, &r); err != nil {
		log.FluentfContext(consts.LOGERROR, c, err.Error())
		return r, err
	}

	//	m := r.(map[string]interface{})

	if !r.Success {
		return r, err
	}

	return r, nil
}

func GetAccountBalances(c context.Context, account string) (AccountBalance, error) {
	if isInit == false {
		Init()
	}

	var r AccountBalance

	result, status, err := httpGet(c, "/v1/accounts/"+account+"/balances")

	if result == nil {
		return r, errors.New("Ripple unavailable")
	}

	if status != 0 {
		log.FluentfContext(consts.LOGERROR, c, string(result))
		return r, err
	}

	//	var r interface{}
	//	println (string(result))

	if err := json.Unmarshal(result, &r); err != nil {
		log.FluentfContext(consts.LOGERROR, c, err.Error())
		println("Marshall error")
		return r, err
	}

	//	m := r.(map[string]interface{})

	if !r.Success {
		return r, err
	}

	return r, nil
}

func PreparePayment(c context.Context, source_address string, destination_address string, amount int64, currency string, issuer string) (PreparePaymentList, error) {
	if isInit == false {
		Init()
	}

	var r PreparePaymentList

	var requestString string = "/v1/accounts/" + source_address + "/payments/paths/" + destination_address + "/" + strconv.FormatInt(amount, 20) + "+" + currency + "+" + issuer

	result, status, err := httpGet(c, requestString)
	if result == nil {
		return r, errors.New("Ripple unavailable")
	}

	if status != 0 {
		log.FluentfContext(consts.LOGERROR, c, string(result))
		return r, err
	}

	//	var r interface{}
	//	println (string(result))

	if err := json.Unmarshal(result, &r); err != nil {
		log.FluentfContext(consts.LOGERROR, c, err.Error())
		println("Marshall error")
		return r, err
	}

	//	m := r.(map[string]interface{})

	if !r.Success {
		return r, err
	}

	return r, nil
}

func PostPayment(c context.Context, secret string, client_resource_id string, source_address string, destination_address string, amount int64, currency string, issuer string) (PaymentList, error) {
	if isInit == false {
		Init()
	}

	var payload PaymentList
	var result PaymentList

	var request string = "/v1/accounts/" + source_address + "/payments?validated=true"

	payload.Secret = secret
	payload.Client_resource_id = client_resource_id
	payload.Payments.Source_account = source_address
	payload.Payments.Source_amount.Value = fmt.Sprintf("%d", amount)
	payload.Payments.Source_amount.Currency = currency
	payload.Payments.Source_amount.Issuer = issuer
	payload.Payments.Destination_account = destination_address
	payload.Payments.Destination_amount.Value = fmt.Sprintf("%d", amount)
	payload.Payments.Destination_amount.Currency = currency
	payload.Payments.Destination_amount.Issuer = issuer
	payload.Payments.Paths = "[]"

	payloadJsonBytes, err := json.Marshal(payload)

	//		log.Println(string(payloadJsonBytes))

	if err != nil {
		return result, err
	}

	responseData, _, err := postAPI(c, request, payloadJsonBytes)
	if err != nil {
		return result, err
	}

	log.Println(string(responseData))

	if err := json.Unmarshal(responseData, &result); err != nil {
		return result, errors.New("Unable to unmarshal responseData")
	}

	//	if !result.Success {
	// 	return result, err
	//	}

	return result, nil
}

func GetConfirmPayment(c context.Context, source_address string, client_resource_id string) (ConfirmPayment, error) {
	if isInit == false {
		Init()
	}

	var r ConfirmPayment

	var requestString string = "/v1/accounts/" + source_address + "/payments/" + client_resource_id

	result, status, err := httpGet(c, requestString)
	if result == nil {
		return r, errors.New("Ripple unavailable")
	}

	if status != 0 {
		log.FluentfContext(consts.LOGERROR, c, string(result))
		return r, err
	}

	//	var r interface{}
	println(string(result))

	if err := json.Unmarshal(result, &r); err != nil {
		log.FluentfContext(consts.LOGERROR, c, err.Error())
		println("Marshall error")
		return r, err
	}

	//	m := r.(map[string]interface{})

	//	if !r.Success {
	// 	return r, err
	//	}

	return r, nil
}

func PostTrustline(c context.Context, secret string, source_address string, destination_address string, limit int64, currency string) (TrustlineResult, error) {
	if isInit == false {
		Init()
	}

	var payload TrustlineList
	var result TrustlineResult

	var request string = "/v1/accounts/" + destination_address + "/trustlines?validated=true"

	payload.Secret = secret
	payload.Trustlines.Limit = fmt.Sprintf("%d", limit)
	payload.Trustlines.Currency = currency
	payload.Trustlines.Counterparty = source_address
	payload.Trustlines.Account_allows_rippling = false

	payloadJsonBytes, err := json.Marshal(payload)

	//		log.Println(string(payloadJsonBytes))

	if err != nil {
		return result, err
	}

	responseData, _, err := postAPI(c, request, payloadJsonBytes)
	if err != nil {
		return result, err
	}

	log.Println(string(responseData))

	if err := json.Unmarshal(responseData, &result); err != nil {
		return result, errors.New("Unable to unmarshal responseData")
	}

	//	if !result.Success {
	// 	return result, err
	//	}

	return result, nil
}

func GetTrustLines(c context.Context, account string) (GetTrustlinesResult, error) {
	if isInit == false {
		Init()
	}

	var r GetTrustlinesResult

	result, status, err := httpGet(c, "/v1/accounts/"+account+"/trustlines")

	if result == nil {
		return r, errors.New("Ripple unavailable")
	}

	if status != 0 {
		log.FluentfContext(consts.LOGERROR, c, string(result))
		return r, err
	}

	//	var r interface{}
	println(string(result))

	if err := json.Unmarshal(result, &r); err != nil {
		log.FluentfContext(consts.LOGERROR, c, err.Error())
		println("Marshall error")
		return r, err
	}

	//	m := r.(map[string]interface{})

	if !r.Success {
		return r, err
	}

	return r, nil
}
