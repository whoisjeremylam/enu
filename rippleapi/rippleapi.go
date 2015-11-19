package rippleapi

import (
	"bytes"
	"math/big"
	"strings"

	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/vennd/enu/consts"
	"github.com/vennd/enu/log"

	"github.com/vennd/enu/internal/github.com/gorilla/securecookie"
	"github.com/vennd/enu/internal/golang.org/x/net/context"
)

var defaultFee = "10000"

// Account set flags
const AsfRequireDest = 1
const AsfRequireAuth = 2
const AsfDisallowXRP = 3
const AsfDisableMaster = 4
const AsfAccountTxnID = 5
const AsfNoFreeze = 6
const AsfGlobalFreeze = 7
const AsfDefaultRipple = 8

// Trust set flags (on the transaction)
const TfSetfAuth = 65536
const TfSetNoRipple = 131072
const TfClearNoRipple = 262144
const TfSetFreeze = 1048576
const TfClearFreeze = 2097152

// Structure for payment transactions for custom currencies
type PaymentAssetTx struct {
	// Common fields
	Account            string `json:",omitempty"`
	AccountTxnID       string `json:",omitempty"`
	Fee                string `json:",omitempty"`
	Flags              uint32 `json:",omitempty"`
	LastLedgerSequence uint32 `json:",omitempty"`
	Memos              []Memo
	Sequence           uint32 `json:",omitempty"`
	SigningPubKey      string `json:",omitempty"`
	SourceTag          uint32 `json:",omitempty"`
	TransactionType    string `json:",omitempty"`
	TxnSignature       string `json:",omitempty"`

	// Payment specific fields
	Amount         Amount
	Destination    string
	DestinationTag uint32
	InvoiceID      string
	//	Paths
	//	SendMax Currency
	//	DeliverMin Currency
}

// Structure for payment transactions for xrp
type PaymentXrpTx struct {
	// Common fields
	Account            string `json:",omitempty"`
	AccountTxnID       string `json:",omitempty"`
	Fee                string `json:",omitempty"`
	Flags              uint32 `json:",omitempty"`
	LastLedgerSequence uint32 `json:",omitempty"`
	Memos              []Memo
	Sequence           uint32 `json:",omitempty"`
	SigningPubKey      string `json:",omitempty"`
	SourceTag          uint32 `json:",omitempty"`
	TransactionType    string `json:",omitempty"`
	TxnSignature       string `json:",omitempty"`

	// Payment specific fields
	Amount         string `json:",omitempty"`
	Destination    string `json:",omitempty"`
	DestinationTag uint32 `json:",omitempty"`
	InvoiceID      string `json:",omitempty"`
	//	Paths
	//	SendMax Currency
	//	DeliverMin Currency
}

type Memo struct {
	MemoData   string `json:",omitempty"`
	MemoFormat string `json:",omitempty"`
	MemoType   string `json:",omitempty"`
}

type Wallet struct {
	AccountId     string `json:"account_id"`
	KeyType       string `json:"key_type"`
	MasterKey     string `json:"master_key"`
	MasterSeed    string `json:"master_seed"`
	MasterSeedHex string `json:"master_seed_hex"`
	PublicKey     string `json:"public_key"`
	PublicKeyHex  string `json:"public_key_hex"`
	Status        string `json:"status"`
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

//type AccountBalance struct {
//	Ledger    int64     `json:"ledger"`
//	Validated bool      `json:"validated"`
//	Balances  []Balance `json:"balances"`
//	Success   bool      `json:"success"`
//}

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

type PayloadGetAccountlines struct {
	Method string            `json:"method"`
	Params ParmsGetAcctlines `json:"params"`
}

type ParmsGetAcctlines []PayloadGetAccountlinesParms

type PayloadGetAccountlinesParms struct {
	Account string `json:"account"`
	Ledger  string `json:"ledger"`
}

type Accountlines struct {
	Result AccountlinesResult `json:"result"`
}

type AccountlinesResult struct {
	Account              string        `json:"account"`
	Ledger_current_index int64         `json:"ledger_current_index"`
	GetAccountLines      []Accountline `json:"lines"`
	Status               string        `json:"status"`
	Validated            bool          `json:"validated"`
}

type Accountline struct {
	Account     string `json:"account"`
	Balance     string `json:"balance"`
	Currency    string `json:"currency"`
	Limit       string `json:"limit"`
	Limit_peer  string `json:"limit_peer"`
	Quality_in  int64  `json:"quality_in"`
	Quality_out int64  `json:"quality_out"`
}

type ApiResult struct {
	resp *http.Response
	err  error
}

type payloadGetServerInfo struct {
	Method string                     `json:"method"`
	Params payloadGetServerInfoParams `json:"params"`
}

type payloadGetServerInfoParams struct{}

type payloadGetCurrenciesByAccount struct {
	Method string                   `json:"method"`
	Params payloadGetCcyByAcctParms `json:"params"`
}

type payloadGetCcyByAcctParms []PayloadGetCcyByAcct

type PayloadGetCcyByAcct struct {
	Account       string `json:"account"`
	Account_index int64  `json:"account_index"`
	Ledger_index  string `json:"ledger_index"`
	Strict        bool   `json:"strict"`
}

type CurrenciesByAccount struct {
	Result CcyByAccountResult `json:"result"`
}

type CcyByAccountResult struct {
	Ledger_hash       string   `json:"ledger_hash"`
	Ledger_index      int64    `json:"ledger_index"`
	ReceiveCurrencies []string `json:"receive_currencies"`
	SendCurrencies    []string `json:"send_currencies"`
	Status            string   `json:"status"`
	Validated         bool     `json:"validated"`
}

type Currency struct {
	Currency string `json:"currency"`
}

type AccountSet struct {
	// Common fields
	Account            string `json:",omitempty"`
	AccountTxnID       string `json:",omitempty"`
	Fee                string `json:",omitempty"`
	Flags              uint32 `json:",omitempty"`
	LastLedgerSequence uint32 `json:",omitempty"`
	Memos              []Memo
	Sequence           uint32 `json:",omitempty"`
	SigningPubKey      string `json:",omitempty"`
	SourceTag          uint32 `json:",omitempty"`
	TransactionType    string `json:",omitempty"`
	TxnSignature       string `json:",omitempty"`

	ClearFlag    uint32 `json:",omitempty"`
	Domain       string `json:",omitempty"`
	EmailHash    string `json:",omitempty"`
	MessageKey   string `json:",omitempty"`
	SetFlag      uint32 `json:",omitempty"`
	TransferRate uint32 `json:",omitempty"`
}

type LimitAmount struct {
	Value    string `json:"value,omitempty"`
	Currency string `json:"currency,omitempty"`
	Issuer   string `json:"issuer,omitempty"`
}

type TrustSetStruct struct {
	// Common fields
	Account            string `json:",omitempty"`
	AccountTxnID       string `json:",omitempty"`
	Fee                string `json:",omitempty"`
	Flags              uint32 `json:",omitempty"`
	LastLedgerSequence uint32 `json:",omitempty"`
	Memos              []Memo `json:",omitempty"`
	Sequence           uint32 `json:",omitempty"`
	SigningPubKey      string `json:",omitempty"`
	SourceTag          uint32 `json:",omitempty"`
	TransactionType    string `json:",omitempty"`
	TxnSignature       string `json:",omitempty"`

	LimitAmount LimitAmount `json:",omitempty"`
	QualityIn   uint32      `json:",omitempty"`
	QualityOut  uint32      `json:",omitempty"`
}

type Lines struct {
	Account      string `json:"account,omitempty"`
	Balance      string `json:"balance,omitempty"`
	Currency     string `json:"currency,omitempty"`
	Limit        string `json:"limit,omitempty"`
	LimitPeer    string `json:"limit_peer,omitempty"`
	NoRipple     bool   `json:"no_ripple,omitempty"`
	NoRipplePeer bool   `json:"no_ripple_peer,omitempty"`
	QualityIn    uint   `json:"quality_in,omitempty"`
	QualityOut   uint   `json:"quality_out,omitempty"`
}

// Initialises global variables and database connection for all handlers
var isInit bool = false // set to true only after the init sequence is complete
var rippleHost string
var rippleRestHost string

var numericAssetIdMinString = "95428956661682176"
var numericAssetIdMaxString = "18446744073709551616"

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
	rippleRestHost = m["rippleRestHost"].(string) // End point for JSON RESTAPI server
	rippleHost = m["rippleHost"].(string)         // End point for JSON RPC server

	isInit = true
}

func httpGet(c context.Context, url string) ([]byte, int64, error) {
	// Set headers

	req, err := http.NewRequest("GET", rippleRestHost+url, nil)

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
	req, err := http.NewRequest("POST", rippleRestHost+request, bytes.NewBufferString(postDataJson))
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

func postRPCAPI(c context.Context, postData []byte) (map[string]interface{}, int64, error) {

	var result map[string]interface{}
	var apiResp ApiResult

	postDataJson := string(postData)
	//postDataJson := `{"method":"account_lines","params":[{"account":"rE1Lec75PEmeDFwAuumto2Nbo8ZwG3aT9V","ledger":"current"}]}`
	log.FluentfContext(consts.LOGDEBUG, c, "rippleapi postRPCAPI() posting: %s", postDataJson)

	// Set headers
	req, err := http.NewRequest("POST", rippleHost, bytes.NewBufferString(postDataJson))
	req.Header.Set("Content-Type", "application/json")

	clientPointer := &http.Client{}

	// Call ripple JSON RPC service with 10 second timeout
	c1 := make(chan ApiResult, 1)
	go func() {
		var result ApiResult // Wrap the response into a struct so we can return both the error and response

		resp, err := clientPointer.Do(req)
		result.resp = resp
		result.err = err

		c1 <- result
	}()

	select {
	case apiResp = <-c1:
	case <-time.After(time.Second * 10):
		return result, consts.RippleErrors.Timeout.Code, errors.New(consts.RippleErrors.Timeout.Description)
	}

	if apiResp.err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Error in Do(req): %s", apiResp.err.Error())
		return result, consts.RippleErrors.MiscError.Code, errors.New(consts.RippleErrors.MiscError.Description)
	}

	// Success, read body and return
	body, err := ioutil.ReadAll(apiResp.resp.Body)
	log.FluentfContext(consts.LOGDEBUG, c, "rippleapi postRPCAPI() body returned: %s", string(body))

	defer apiResp.resp.Body.Close()

	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Error in ReadAll(): %s", err.Error())
		return nil, consts.RippleErrors.MiscError.Code, errors.New(consts.RippleErrors.MiscError.Description)
	}

	// Unmarshall body
	if unmarshallErr := json.Unmarshal(body, &result); unmarshallErr != nil {
		// If we couldn't parse the error properly, log error to fluent and return unhandled error
		log.FluentfContext(consts.LOGERROR, c, "Error in Unmarshal(): %s", unmarshallErr.Error())

		return result, 0, nil
	}

	return result, 0, nil
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

func GetServerStatusRest(c context.Context) ([]byte, error) {
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

// Submits a transaction to the Ripple network
func Submit(c context.Context, txHexString string) (string, int64, error) {
	if isInit == false {
		Init()
	}

	var payload = make(map[string]interface{})
	var params = make(map[string]interface{})
	var paramsArray []map[string]interface{}
	var result string

	// Build parameters
	params["tx_blob"] = txHexString
	paramsArray = append(paramsArray, params)

	// Build payload
	payload["method"] = "submit"
	payload["params"] = paramsArray
	payloadJsonBytes, err := json.Marshal(payload)
	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Error in Marshal(): %s", err.Error())
		return "", consts.RippleErrors.MiscError.Code, errors.New(consts.RippleErrors.MiscError.Description)
	}

	responseData, errorCode, err := postRPCAPI(c, payloadJsonBytes)
	if err != nil {
		return "", errorCode, err
	}

	log.Printf("%#v", responseData)

	if responseData["result"] != nil {
		if responseData["result"].(map[string]interface{})["status"] != nil && responseData["result"].(map[string]interface{})["status"] == "success" {
			result = responseData["result"].(map[string]interface{})["tx_json"].(map[string]interface{})["hash"].(string)
		} else {
			// do some errorhandling here
			log.Printf("sign returned some kind of error")
		}
	}

	return result, 0, nil
}

// Signs a tx with the given secret. The tx should be a struct containing the tx to be marshalled into JSON and then signed
func Sign(c context.Context, tx interface{}, secret string) (string, int64, error) {
	if isInit == false {
		Init()
	}

	var payload = make(map[string]interface{})
	var params = make(map[string]interface{})
	var paramsArray []map[string]interface{}
	var result string

	// Build parameters
	params["offline"] = false
	params["secret"] = secret
	params["tx_json"] = tx
	paramsArray = append(paramsArray, params)

	// Build payload
	payload["method"] = "sign"
	payload["params"] = paramsArray

	payloadJsonBytes, err := json.Marshal(payload)
	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Error in Marshal(): %s", err.Error())
		return "", consts.RippleErrors.MiscError.Code, errors.New(consts.RippleErrors.MiscError.Description)
	}

	responseData, errorCode, err := postRPCAPI(c, payloadJsonBytes)
	if err != nil {
		return "", errorCode, err
	}

	log.Printf("%#v", responseData)

	if responseData["result"] != nil {
		if responseData["result"].(map[string]interface{})["status"] != nil && responseData["result"].(map[string]interface{})["status"] == "success" {
			result = responseData["result"].(map[string]interface{})["tx_blob"].(string)
		} else {
			// do some errorhandling here
			log.Printf("sign returned some kind of error")
		}
	}

	return result, 0, nil
}

// Creates a Ripple account offline. ie doesn't use the REST or RPC
func CreateWallet(c context.Context) (Wallet, int64, error) {
	if isInit == false {
		Init()
	}

	var payload = make(map[string]interface{})
	var result Wallet

	payload["method"] = "wallet_propose"
	payloadJsonBytes, err := json.Marshal(payload)
	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Error in Marshal(): %s", err.Error())
		return result, consts.RippleErrors.MiscError.Code, errors.New(consts.RippleErrors.MiscError.Description)
	}

	responseData, errorCode, err := postRPCAPI(c, payloadJsonBytes)
	if err != nil {
		return result, errorCode, err
	}

	if responseData["result"] != nil {
		log.Printf("%#v", responseData["result"])
	}

	responseResult := responseData["result"].(map[string]interface{})
	result.AccountId = responseResult["account_id"].(string)
	result.KeyType = responseResult["key_type"].(string)
	result.MasterKey = responseResult["master_key"].(string)
	result.MasterSeed = responseResult["master_seed"].(string)
	result.MasterSeedHex = responseResult["master_seed_hex"].(string)
	result.PublicKey = responseResult["public_key"].(string)
	result.PublicKeyHex = responseResult["public_key_hex"].(string)
	result.Status = responseResult["status"].(string)

	return result, 0, nil
}

func GetAccountSettingsRest(c context.Context, account string) (AccountSettings, error) {
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

func GetAccountBalances(c context.Context, account string) ([]Balance, error) {
	if isInit == false {
		Init()
	}

	var result []Balance

	lines, errCode, err := GetAccountLines(c, account)
	if err != nil {
		return result, err
	}

	for _, line := range lines {
		var balance Balance

		var value big.Float
		if line.Balance > 0 {
			balance.Value = line.Balance
			balance.Currency = line.Currency
			balance.Counterparty = line.Account
		}

		result = append(result, balance)
	}

	return result, nil
}

//func PreparePayment(c context.Context, source_address string, destination_address string, amount int64, currency string, issuer string) (PreparePaymentList, error) {
//	if isInit == false {
//		Init()
//	}

//	var r PreparePaymentList

//	var requestString string = "/v1/accounts/" + source_address + "/payments/paths/" + destination_address + "/" + strconv.FormatInt(amount, 20) + "+" + currency + "+" + issuer

//	result, status, err := httpGet(c, requestString)
//	if result == nil {
//		return r, errors.New("Ripple unavailable")
//	}

//	if status != 0 {
//		log.FluentfContext(consts.LOGERROR, c, string(result))
//		return r, err
//	}

//	//	var r interface{}
//	//	println (string(result))

//	if err := json.Unmarshal(result, &r); err != nil {
//		log.FluentfContext(consts.LOGERROR, c, err.Error())
//		println("Marshall error")
//		return r, err
//	}

//	//	m := r.(map[string]interface{})

//	if !r.Success {
//		return r, err
//	}

//	return r, nil
//}

//func PostPayment(c context.Context, secret string, client_resource_id string, source_address string, destination_address string, amount int64, currency string, issuer string) (PaymentList, error) {
//	if isInit == false {
//		Init()
//	}

//	var payload PaymentList
//	var result PaymentList

//	var request string = "/v1/accounts/" + source_address + "/payments?validated=true"

//	payload.Secret = secret
//	payload.Client_resource_id = client_resource_id
//	payload.Payments.Source_account = source_address
//	payload.Payments.Source_amount.Value = fmt.Sprintf("%d", amount)
//	payload.Payments.Source_amount.Currency = currency
//	payload.Payments.Source_amount.Issuer = issuer
//	payload.Payments.Destination_account = destination_address
//	payload.Payments.Destination_amount.Value = fmt.Sprintf("%d", amount)
//	payload.Payments.Destination_amount.Currency = currency
//	payload.Payments.Destination_amount.Issuer = issuer
//	payload.Payments.Paths = "[]"

//	payloadJsonBytes, err := json.Marshal(payload)

//	//		log.Println(string(payloadJsonBytes))

//	if err != nil {
//		return result, err
//	}

//	responseData, _, err := postAPI(c, request, payloadJsonBytes)
//	if err != nil {
//		return result, err
//	}

//	log.Println(string(responseData))

//	if err := json.Unmarshal(responseData, &result); err != nil {
//		return result, errors.New("Unable to unmarshal responseData")
//	}

//	//	if !result.Success {
//	// 	return result, err
//	//	}

//	return result, nil
//}

//func GetConfirmPayment(c context.Context, source_address string, client_resource_id string) (ConfirmPayment, error) {
//	if isInit == false {
//		Init()
//	}

//	var r ConfirmPayment

//	var requestString string = "/v1/accounts/" + source_address + "/payments/" + client_resource_id

//	result, status, err := httpGet(c, requestString)
//	if result == nil {
//		return r, errors.New("Ripple unavailable")
//	}

//	if status != 0 {
//		log.FluentfContext(consts.LOGERROR, c, string(result))
//		return r, err
//	}

//	//	var r interface{}
//	println(string(result))

//	if err := json.Unmarshal(result, &r); err != nil {
//		log.FluentfContext(consts.LOGERROR, c, err.Error())
//		println("Marshall error")
//		return r, err
//	}

//	//	m := r.(map[string]interface{})

//	//	if !r.Success {
//	// 	return r, err
//	//	}

//	return r, nil
//}

//func PostTrustline(c context.Context, secret string, source_address string, destination_address string, limit int64, currency string) (TrustlineResult, error) {
//	if isInit == false {
//		Init()
//	}

//	var payload TrustlineList
//	var result TrustlineResult

//	var request string = "/v1/accounts/" + destination_address + "/trustlines?validated=true"

//	payload.Secret = secret
//	payload.Trustlines.Account = destination_address
//	payload.Trustlines.Limit = fmt.Sprintf("%d", limit)
//	payload.Trustlines.Currency = currency
//	payload.Trustlines.Counterparty = source_address
//	payload.Trustlines.Account_allows_rippling = false

//	payloadJsonBytes, err := json.Marshal(payload)

//	//		log.Println(string(payloadJsonBytes))

//	if err != nil {
//		return result, err
//	}

//	responseData, _, err := postAPI(c, request, payloadJsonBytes)
//	if err != nil {
//		return result, err
//	}

//	log.Println(string(responseData))

//	if err := json.Unmarshal(responseData, &result); err != nil {
//		return result, errors.New("Unable to unmarshal responseData")
//	}

//	//	if !result.Success {
//	// 	return result, err
//	//	}

//	return result, nil
//}

//func GetTrustLinesRest(c context.Context, account string) (GetTrustlinesResult, error) {
//	if isInit == false {
//		Init()
//	}

//	var r GetTrustlinesResult

//	result, status, err := httpGet(c, "/v1/accounts/"+account+"/trustlines")

//	if result == nil {
//		return r, errors.New("Ripple unavailable")
//	}

//	if status != 0 {
//		log.FluentfContext(consts.LOGERROR, c, string(result))
//		return r, err
//	}

//	//	var r interface{}
//	println(string(result))

//	if err := json.Unmarshal(result, &r); err != nil {
//		log.FluentfContext(consts.LOGERROR, c, err.Error())
//		println("Marshall error")
//		return r, err
//	}

//	//	m := r.(map[string]interface{})

//	if !r.Success {
//		return r, err
//	}

//	return r, nil
//}

func ServerInfo(c context.Context) ([]byte, int64, error) {
	var payload payloadGetServerInfo
	//	var result []Balance
	var result []byte

	if isInit == false {
		Init()
	}

	payload.Method = "server_info"

	payloadJsonBytes, err := json.Marshal(payload)
	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Error in Marshal(): %s", err.Error())
		return result, consts.CounterpartyErrors.MiscError.Code, errors.New(consts.CounterpartyErrors.MiscError.Description)
	}

	responseData, errorCode, err := postRPCAPI(c, payloadJsonBytes)
	if err != nil {
		return result, errorCode, err
	}

	if responseData["result"] != nil {
		log.Printf("%#v", responseData["result"])
	}

	return result, errorCode, nil
}

func GetCurrenciesByAccount(c context.Context, account string) (CurrenciesByAccount, int64, error) {
	var payload payloadGetCurrenciesByAccount
	var result CurrenciesByAccount
	var result2 []string
	var result3 []string

	if isInit == false {
		Init()
	}

	payload.Method = "account_currencies"
	parms := PayloadGetCcyByAcct{Account: account, Account_index: 0, Ledger_index: "validated", Strict: true}
	payload.Params = append(payload.Params, parms)

	payloadJsonBytes, err := json.Marshal(payload)
	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Error in Marshal(): %s", err.Error())
		return result, consts.CounterpartyErrors.MiscError.Code, errors.New(consts.CounterpartyErrors.MiscError.Description)
	}

	responseData, errorCode, err := postRPCAPI(c, payloadJsonBytes)
	if err != nil {
		return result, errorCode, err
	}

	// Get result from api and create the reply
	if responseData["result"] != nil {
		resultMap := responseData["result"].(map[string]interface{})
		recCcys := resultMap["receive_currencies"].([]interface{})
		sendCcys := resultMap["send_currencies"].([]interface{})

		log.Println("Mapped:")
		log.Printf("%#v\n", resultMap)
		log.Printf("%#v\n", recCcys)
		log.Printf("%#v\n", sendCcys)

		for _, b := range sendCcys {
			c := b.(string)
			result2 = append(result2, c)
		}
		for _, b := range recCcys {
			d := b.(string)
			result3 = append(result3, d)
		}

		result = CurrenciesByAccount{CcyByAccountResult{
			Ledger_hash:       resultMap["ledger_hash"].(string),
			Ledger_index:      int64(resultMap["ledger_index"].(float64)),
			ReceiveCurrencies: result2,
			SendCurrencies:    result3,
			Status:            resultMap["status"].(string),
			Validated:         resultMap["validated"].(bool),
		}}
	}

	return result, 0, nil
}

// Creates and sends the payment for the custom currency that is specified.
// Returns the tx hash if successful
func SendPayment(c context.Context, sourceAddress string, destinationAddress string, quantity uint64, asset string, issuer string, secret string) (string, int64, error) {
	if isInit == false {
		Init()
	}

	var signedTx string
	var errCode int64
	var err error
	var txHash string

	if strings.ToUpper(asset) == "XRP" {
		tx := PaymentXrpTx{
			TransactionType: "Payment",
			Account:         sourceAddress,
			Destination:     destinationAddress,
			Amount:          fmt.Sprintf("%d", quantity),
			Flags:           2147483648, // require canonical signature
			Fee:             defaultFee,
		}

		signedTx, errCode, err = Sign(c, tx, secret)
	} else {
		tx := PaymentAssetTx{
			TransactionType: "Payment",
			Account:         sourceAddress,
			Destination:     destinationAddress,
			Amount: Amount{
				Value:    fmt.Sprintf("%d", quantity),
				Currency: asset,
				Issuer:   issuer,
			},
			Flags: 2147483648, // require canonical signature
			Fee:   defaultFee,
		}

		signedTx, errCode, err = Sign(c, tx, secret)
	}

	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Error in Sign(): %s", err.Error())
		return "", errCode, err
	}

	log.Printf("signed! tx_blob: %s", signedTx)

	txHash, errCode, err = Submit(c, signedTx)
	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Error in Submit(): %s", err.Error())
	}

	return txHash, errCode, err
}

// Sets a specific flag on an account
func AccountSetFlag(c context.Context, account string, flag uint32, secret string) (string, int64, error) {
	if isInit == false {
		Init()
	}

	var signedTx string
	var errCode int64
	var err error
	var txHash string

	tx := AccountSet{
		// Common fields
		TransactionType: "AccountSet",
		Account:         account,
		Flags:           2147483648, // require canonical signature
		Fee:             defaultFee,

		SetFlag: flag,
	}

	signedTx, errCode, err = Sign(c, tx, secret)
	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Error in Sign(): %s", err.Error())
		return "", errCode, err
	}

	log.Printf("signed! tx_blob: %s", signedTx)

	txHash, errCode, err = Submit(c, signedTx)
	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Error in Submit(): %s", err.Error())
	}

	return txHash, errCode, err
}

// Modifies a trust line between two accounts
// The trust line is directional - the given account trusts the issuer account for value amount of currency
// A trust line occupies space in the Ripple ledger and therefore requires a fee to be paid and consequently the secret of the source account
func TrustSet(c context.Context, account string, currency string, value string, issuerAccount string, flag uint32, secret string) (string, int64, error) {
	if isInit == false {
		Init()
	}

	var signedTx string
	var errCode int64
	var err error
	var txHash string

	tx := TrustSetStruct{
		// Common fields
		TransactionType: "TrustSet",
		Account:         account,
		Flags:           2147483648 & flag, // require canonical signature
		Fee:             defaultFee,

		// Set the limit
		LimitAmount: LimitAmount{
			Value:    value,
			Currency: currency,
			Issuer:   issuerAccount,
		},
	}

	signedTx, errCode, err = Sign(c, tx, secret)
	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Error in Sign(): %s", err.Error())
		return "", errCode, err
	}

	log.Printf("signed! tx_blob: %s", signedTx)

	txHash, errCode, err = Submit(c, signedTx)
	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Error in Submit(): %s", err.Error())
	}

	return txHash, errCode, err
}

func GetAccountLines(c context.Context, account string) ([]Lines, int64, error) {
	var payload = make(map[string]interface{})
	var params = make(map[string]interface{})
	var paramsArray []map[string]interface{}
	var result []Lines
	var responseData map[string]interface{}

	if isInit == false {
		Init()
	}

	// Build parameters
	params["account"] = account
	params["ledger"] = "validated"
	paramsArray = append(paramsArray, params)

	// Build payload
	payload["method"] = "account_lines"
	payload["params"] = paramsArray

	payloadJsonBytes, err := json.Marshal(payload)
	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Error in Marshal(): %s", err.Error())
		return result, consts.RippleErrors.MiscError.Code, errors.New(consts.RippleErrors.MiscError.Description)
	}

	responseData, errorCode, err := postRPCAPI(c, payloadJsonBytes)
	if err != nil {
		return result, errorCode, err
	}

	if responseData["result"] == nil {
		log.FluentfContext(consts.LOGERROR, c, "Didn't receive a result from RPC server")
		log.FluentfContext(consts.LOGERROR, c, "Got: %#v", responseData["result"])
		return result, consts.RippleErrors.MiscError.Code, errors.New(consts.RippleErrors.MiscError.Description)
	}

	return result, errorCode, nil
}
