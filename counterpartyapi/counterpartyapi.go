// Contains API to counterparty functions
// Regarding errorhandling, if a lower level function returns an errorCode, propagate the error back upwards
// If the function handling the error is not exposed directly to the HTTP handlers, it's better that the original error is propagated to preserve the error

package counterpartyapi

import (
	"bytes"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/big"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	_ "code.google.com/p/go-sqlite/go1/sqlite3"

	"github.com/vennd/enu/bitcoinapi"
	"github.com/vennd/enu/consts"
	"github.com/vennd/enu/counterpartycrypto"
	"github.com/vennd/enu/database"
	"github.com/vennd/enu/log"

	"github.com/vennd/enu/internal/github.com/btcsuite/btcd/btcec"
	"github.com/vennd/enu/internal/github.com/btcsuite/btcd/chaincfg"
	"github.com/vennd/enu/internal/github.com/btcsuite/btcd/txscript"
	"github.com/vennd/enu/internal/github.com/btcsuite/btcd/wire"
	"github.com/vennd/enu/internal/github.com/btcsuite/btcutil"
	"github.com/vennd/enu/internal/github.com/gorilla/securecookie"
	"github.com/vennd/enu/internal/golang.org/x/net/context"
)

var Counterparty_DefaultDustSize uint64 = 5430
var Counterparty_DefaultTxFee uint64 = 10000       // in satoshis
var Counterparty_DefaultTestingTxFee uint64 = 1500 // in satoshis
var numericAssetIdMinString = "95428956661682176"
var numericAssetIdMaxString = "18446744073709551616"
var counterparty_BackEndPollRate = 2000 // milliseconds

var counterparty_Mutexes = struct {
	sync.RWMutex
	m map[string]*sync.Mutex
}{m: make(map[string]*sync.Mutex)}

type payloadGetBalances struct {
	Method  string                   `json:"method"`
	Params  payloadGetBalancesParams `json:"params"`
	Jsonrpc string                   `json:"jsonrpc"`
	Id      uint32                   `json:"id"`
}

type payloadGetBalancesParams struct {
	Filters  filters `json:"filters"`
	FilterOp string  `json:"filterop"`
}

type filters []filter

type filter struct {
	Field string `json:"field"`
	Op    string `json:"op"`
	Value string `json:"value"`
}

type ResultGetBalances struct {
	Jsonrpc string    `json:"jsonrpc"`
	Id      uint32    `json:"id"`
	Result  []Balance `json:"result"`
}

type Balance struct {
	Quantity uint64 `json:"quantity"`
	Asset    string `json:"asset"`
	Address  string `json:"address"`
}

// Struct definitions for creating a send Counterparty transaction
type payloadCreateSend_Counterparty struct {
	Method  string                               `json:"method"`
	Params  payloadCreateSendParams_Counterparty `json:"params"`
	Jsonrpc string                               `json:"jsonrpc"`
	Id      uint32                               `json:"id"`
}

//  myParams = ["source":sourceAddress,"destination":destinationAddress,"asset":asset,"quantity":amount,"allow_unconfirmed_inputs":true,"encoding":counterpartyTransactionEncoding,"pubkey":pubkey]
type payloadCreateSendParams_Counterparty struct {
	Source                 string `json:"source"`
	Destination            string `json:"destination"`
	Asset                  string `json:"asset"`
	Quantity               uint64 `json:"quantity"`
	AllowUnconfirmedInputs string `json:"allow_unconfirmed_inputs"`
	Encoding               string `json:"encoding"`
	PubKey                 string `json:"pubkey"`
	Fee                    uint64 `json:"fee"`
	DustSize               uint64 `json:"regular_dust_size"`
}

type ResultCreateSend_Counterparty struct {
	Jsonrpc string `json:"jsonrpc"`
	Id      uint32 `json:"id"`
	Result  string `json:"result"`
}

type payloadCreateIssuance_Counterparty struct {
	Method  string                                   `json:"method"`
	Params  payloadCreateIssuanceParams_Counterparty `json:"params"`
	Jsonrpc string                                   `json:"jsonrpc"`
	Id      uint32                                   `json:"id"`
}

type payloadCreateIssuanceParams_Counterparty struct {
	Source      string `json:"source"`
	Quantity    uint64 `json:"quantity"`
	Asset       string `json:"asset"`
	Divisible   bool   `json:"divisible"`
	Description string `json:"description"`
	//	TransferDestination    string `json:"transfer_destination"`
	Encoding               string `json:"encoding"`
	PubKey                 string `json:"pubkey"`
	AllowUnconfirmedInputs string `json:"allow_unconfirmed_inputs"`
	Fee                    uint64 `json:"fee"`
	DustSize               uint64 `json:"regular_dust_size"`
}

type ResultCreateIssuance_Counterparty struct {
	Jsonrpc string `json:"jsonrpc"`
	Id      uint32 `json:"id"`
	Result  string `json:"result"`
}

type payloadCreateDividend_Counterparty struct {
	Method  string                                   `json:"method"`
	Params  payloadCreateDividendParams_Counterparty `json:"params"`
	Jsonrpc string                                   `json:"jsonrpc"`
	Id      uint32                                   `json:"id"`
}

type payloadCreateDividendParams_Counterparty struct {
	Source                 string `json:"source"`
	Asset                  string `json:"asset"`
	DividendAsset          string `json:"dividend_asset"`
	QuantityPerUnit        uint64 `json:"quantity_per_unit"`
	Encoding               string `json:"encoding"`
	PubKey                 string `json:"pubkey"`
	AllowUnconfirmedInputs string `json:"allow_unconfirmed_inputs"`
	Fee                    uint64 `json:"fee"`
	DustSize               uint64 `json:"regular_dust_size"`
}

type ResultCreateDividend_Counterparty struct {
	Jsonrpc string `json:"jsonrpc"`
	Id      uint32 `json:"id"`
	Result  string `json:"result"`
}

type ResultError_Counterparty struct {
	Jsonrpc string `json:"jsonrpc"`
	Id      uint32 `json:"id"`
	Error   string `json:"error"`
}

type payloadGetIssuances struct {
	Method  string                    `json:"method"`
	Params  payloadGetIssuancesParams `json:"params"`
	Jsonrpc string                    `json:"jsonrpc"`
	Id      uint32                    `json:"id"`
}

type payloadGetIssuancesParams struct {
	OrderBy  string  `json:"order_by"`
	OrderDir string  `json:"order_dir"`
	Filters  filters `json:"filters"`
}

type ResultGetIssuances struct {
	Jsonrpc string     `json:"jsonrpc"`
	Id      uint32     `json:"id"`
	Result  []Issuance `json:"result"`
}

// Create wrapper for http response and error
type ApiResult struct {
	resp *http.Response
	err  error
}

type Issuance struct {
	TxIndex     uint64 `json:"tx_index"`
	TxHash      string `json:"tx_hash"`
	BlockIndex  uint64 `json:"block_index"`
	Asset       string `json:"asset"`
	Quantity    uint64 `json:"quantity"`
	Divisible   uint64 `json:"divisible"`
	Source      string `json:"source"`
	Issuer      string `json:"issuer"`
	Transfer    uint64 `json:"transfer"`
	Description string `json:"description"`
	FeePaid     uint64 `json:"fee_paid"`
	Locked      uint64 `json:"locked"`
	Status      string `json:"status"`
}

type payloadGetRunningInfo struct {
	Method  string `json:"method"`
	Jsonrpc string `json:"jsonrpc"`
	Id      uint32 `json:"id"`
}

type ResultGetRunningInfo struct {
	Jsonrpc string      `json:"jsonrpc"`
	Id      uint32      `json:"id"`
	Result  RunningInfo `json:"result"`
}

type LastBlock struct {
	BlockIndex uint64 `json:"block_index"`
	BlockHash  string `json:"block_hash"`
}

type RunningInfo struct {
	DbCaughtUp           bool      `json:"db_caught_up"`
	BitCoinBlockCount    uint64    `json:"bitcoin_block_count"`
	CounterpartydVersion string    `json:"counterpartyd_version"`
	LastMessageIndex     uint64    `json:"last_message_index"`
	RunningTestnet       bool      `json:"running_testnet"`
	LastBlock            LastBlock `json:"last_block"`
}

// Globals
var isInit bool = false // set to true only after the init sequence is complete
var counterpartyHost string
var counterpartyUser string
var counterpartyPassword string
var counterpartyTransactionEncoding string
var counterpartyDBLocation string

// Initialises global variables and database connection for all handlers
func Init() {
	var configFilePath string

	if isInit == true {
		return
	}

	if _, err := os.Stat("./enuapi.json"); err == nil {
		//		log.Println("Found and using configuration file ./enuapi.json")
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

	// Counterparty API parameters
	counterpartyHost = m["counterpartyhost"].(string)                               // End point for JSON RPC server
	counterpartyUser = m["counterpartyuser"].(string)                               // Basic authentication user name
	counterpartyPassword = m["counterpartypassword"].(string)                       // Basic authentication password
	counterpartyTransactionEncoding = m["counterpartytransactionencoding"].(string) // The encoding that should be used for Counterparty transactions "auto" will let Counterparty select, valid values "multisig", "opreturn"
	counterpartyDBLocation = m["counterpartydblocation"].(string)                   // Direct location of counterpartydb if we can't reach the API

	isInit = true
}

// Posts to the given counterparty JSON RPC call. Returns a map[string]interface{} which has already unmarshalled the JSON result
// Attempts to interpret the counterparty errors such that the caller doesn't need to work out what is going on
func postAPI(c context.Context, postData []byte) (map[string]interface{}, int64, error) {
	var result map[string]interface{}
	var apiResp ApiResult

	postDataJson := string(postData)
	//		log.FluentfContext(consts.LOGDEBUG, c, "counterpartyapi postAPI() posting: %s", postDataJson)

	// Set headers
	req, err := http.NewRequest("POST", counterpartyHost, bytes.NewBufferString(postDataJson))
	req.SetBasicAuth(counterpartyUser, counterpartyPassword)
	req.Header.Set("Content-Type", "application/json")

	clientPointer := &http.Client{}

	// Call counterparty JSON service with 10 second timeout
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
		return result, consts.CounterpartyErrors.Timeout.Code, errors.New(consts.CounterpartyErrors.Timeout.Description)
	}

	if apiResp.err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Error in Do(req): %s", apiResp.err.Error())
		return result, consts.CounterpartyErrors.MiscError.Code, errors.New(consts.CounterpartyErrors.MiscError.Description)
	}

	// Unsuccessful - ie didn't return HTTP status 200
	if apiResp.resp.StatusCode != 200 {
		log.FluentfContext(consts.LOGDEBUG, c, "Request didn't return a 200. Status code: %d\n", apiResp.resp.StatusCode)

		// Even though we got an error, counterparty often sends back errors inside the body
		body, err := ioutil.ReadAll(apiResp.resp.Body)
		defer apiResp.resp.Body.Close()
		if err != nil {
			log.FluentfContext(consts.LOGERROR, c, err.Error())
			return nil, consts.CounterpartyErrors.MiscError.Code, errors.New(consts.CounterpartyErrors.MiscError.Description)
		}

		//		log.FluentfContext(consts.LOGDEBUG, c, "Reply: %s", string(body))

		// Attempt to parse body if not empty in case it is something json-like from counterpartyd
		var errResult map[string]interface{}
		if unmarshallErr := json.Unmarshal(body, &errResult); unmarshallErr != nil {
			// If we couldn't parse the error properly, log error to fluent and return unhandled error
			log.FluentfContext(consts.LOGERROR, c, unmarshallErr.Error())

			return result, consts.CounterpartyErrors.MiscError.Code, errors.New(consts.CounterpartyErrors.MiscError.Description)
		}

		if apiResp.resp.StatusCode == 503 {
			log.FluentfContext(consts.LOGDEBUG, c, "Reply: %s", string(body))
		}

		// Counterparty DB is behind backend / reparsing or timed out, read directly from DB
		if errResult["code"].(float64) == -32000 || errResult["code"].(float64) == -10000 {
			return result, consts.CounterpartyErrors.ReparsingOrUnavailable.Code, errors.New(consts.CounterpartyErrors.ReparsingOrUnavailable.Description)
		}
	}

	// Success, read body and return
	body, err := ioutil.ReadAll(apiResp.resp.Body)
	defer apiResp.resp.Body.Close()

	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Error in ReadAll(): %s", err.Error())
		return nil, consts.CounterpartyErrors.MiscError.Code, errors.New(consts.CounterpartyErrors.MiscError.Description)
	}

	// Unmarshall body
	if unmarshallErr := json.Unmarshal(body, &result); unmarshallErr != nil {
		// If we couldn't parse the error properly, log error to fluent and return unhandled error
		log.FluentfContext(consts.LOGERROR, c, "Error in Unmarshal(): %s", err.Error())

		return result, consts.CounterpartyErrors.MiscError.Code, errors.New(consts.CounterpartyErrors.MiscError.Description)
	}

	// If the body doesn't contain a result then the call must have failed. Attempt to read the payload to work out what happened
	if result["result"] == nil {
		// Uncomment this to work out what the hell counterparty sent back
		//		log.FluentfContext(consts.LOGDEBUG, c, "Result not returned from counterpartyd. Got: %s", fmt.Sprintf("%#v", result))

		// Got an error
		if result["error"] != nil {
			var dataMap map[string]interface{}
			errorMap := result["error"].(map[string]interface{})

			if errorMap["data"] != nil {
				dataMap = errorMap["data"].(map[string]interface{})
			}

			// Only issuer can pay dividends
			if dataMap["message"] != nil && strings.Contains(dataMap["message"].(string), consts.CounterpartylibOnlyIssuerCanPayDividends) {
				return result, consts.CounterpartyErrors.OnlyIssuerCanPayDividends.Code, errors.New(consts.CounterpartyErrors.OnlyIssuerCanPayDividends.Description)
			}

			// Insufficient asset in address
			if dataMap["message"] != nil && strings.Contains(dataMap["message"].(string), consts.CounterpartylibInsufficientFunds) {
				return result, consts.CounterpartyErrors.InsufficientFunds.Code, errors.New(consts.CounterpartyErrors.InsufficientFunds.Description)
			}

			// Bad address
			if dataMap["message"] != nil && strings.Contains(dataMap["message"].(string), consts.CounterpartylibMalformedAddress) {
				return result, consts.CounterpartyErrors.MalformedAddress.Code, errors.New(consts.CounterpartyErrors.MalformedAddress.Description)
			}

			// No such asset
			if dataMap["message"] != nil && strings.Contains(dataMap["message"].(string), consts.CountpartylibNoSuchAsset) {
				return result, consts.CounterpartyErrors.NoSuchAsset.Code, errors.New(consts.CounterpartyErrors.NoSuchAsset.Description)
			}

			//Insufficient BTC at address
			if dataMap["message"] != nil && strings.Contains(dataMap["message"].(string), consts.CounterpartylibInsufficientBTC) {
				return result, consts.CounterpartyErrors.InsufficientFees.Code, errors.New(consts.CounterpartyErrors.InsufficientFees.Description)
			}

			//Counterparty is just restarting now
			if dataMap["message"] != nil && strings.Contains(dataMap["message"].(string), consts.CountpartylibMempoolIsNotReady) {
				return result, consts.CounterpartyErrors.ReparsingOrUnavailable.Code, errors.New(consts.CounterpartyErrors.ReparsingOrUnavailable.Description)
			}
		}

		log.FluentfContext(consts.LOGDEBUG, c, "Counterparty returned an error in the body but returned a HTTP status of 200. Got: %s", fmt.Sprintf("%#v", result))

		return result, consts.CounterpartyErrors.MiscError.Code, errors.New(consts.CounterpartyErrors.MiscError.Description)
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

func GetBalancesByAddress(c context.Context, address string) ([]Balance, int64, error) {
	var payload payloadGetBalances
	var result []Balance

	if isInit == false {
		Init()
	}

	filterCondition := filter{Field: "address", Op: "==", Value: address}

	payload.Method = "get_balances"
	payload.Params.Filters = append(payload.Params.Filters, filterCondition)
	payload.Jsonrpc = "2.0"
	payload.Id = generateId(c)

	payloadJsonBytes, err := json.Marshal(payload)
	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Error in Marshal(): %s", err.Error())
		return result, consts.CounterpartyErrors.MiscError.Code, errors.New(consts.CounterpartyErrors.MiscError.Description)
	}

	responseData, errorCode, err := postAPI(c, payloadJsonBytes)
	if err != nil {
		// Counterparty DB is behind backend / reparsing or timed out, read directly from DB
		if errorCode == consts.CounterpartyErrors.ReparsingOrUnavailable.Code || errorCode == consts.CounterpartyErrors.Timeout.Code {
			return GetBalancesByAddressDB(c, address)
		}

		return result, errorCode, err
	}

	// Range over the result from api and create the reply
	if responseData["result"] != nil {
		for _, b := range responseData["result"].([]interface{}) {
			c := b.(map[string]interface{})
			result = append(result,
				Balance{Address: c["address"].(string),
					Asset:    c["asset"].(string),
					Quantity: uint64(c["quantity"].(float64))})
		}
	}

	return result, 0, nil
}

func GetBalancesByAddressDB(c context.Context, address string) ([]Balance, int64, error) {
	var result []Balance

	// sqlite drivers are not concurrency safe, so must create a connection each time
	db, err := sql.Open("sqlite3", counterpartyDBLocation)
	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Failed to open DB. Reason: %s", err.Error())
		return result, consts.CounterpartyErrors.MiscError.Code, errors.New(consts.CounterpartyErrors.MiscError.Description)
	}

	err = db.Ping()
	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Failed to ping DB. Reason: %s", err.Error())
		return result, consts.CounterpartyErrors.MiscError.Code, errors.New(consts.CounterpartyErrors.MiscError.Description)
	}

	//	 Query DB
	//	log.Fluentf(consts.LOGDEBUG, "select address, asset, quantity from balances where address = %s", address)
	stmt, err := db.Prepare("select address, asset, quantity from balances where address = ?")
	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Failed to prepare statement. Reason: %s", err.Error())
		return result, consts.CounterpartyErrors.MiscError.Code, errors.New(consts.CounterpartyErrors.MiscError.Description)
	}
	defer stmt.Close()

	//	 Get row
	rows, err := stmt.Query(address)
	defer rows.Close()
	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Failed to query. Reason: %s", err.Error())
		return result, consts.CounterpartyErrors.MiscError.Code, errors.New(consts.CounterpartyErrors.MiscError.Description)
	}

	for rows.Next() {
		var balance = Balance{}
		var address []byte
		var asset []byte
		var quantity uint64

		if err := rows.Scan(&address, &asset, &quantity); err == sql.ErrNoRows {
			if err.Error() == "sql: no rows in result set" {
			}
		} else if err != nil {
			log.FluentfContext(consts.LOGERROR, c, "Failed to Scan. Reason: %s", err.Error())
		} else {
			balance = Balance{Address: string(address), Asset: string(asset), Quantity: quantity}
		}

		result = append(result, balance)
	}

	return result, 0, nil
}

func GetBalancesByAsset(c context.Context, asset string) ([]Balance, int64, error) {
	var payload payloadGetBalances
	var result []Balance

	if isInit == false {
		Init()
	}

	filterCondition := filter{Field: "asset", Op: "==", Value: asset}

	payload.Method = "get_balances"
	payload.Params.Filters = append(payload.Params.Filters, filterCondition)
	payload.Jsonrpc = "2.0"
	payload.Id = generateId(c)

	payloadJsonBytes, err := json.Marshal(payload)
	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Error in Marshal(): %s", err.Error())
		return result, consts.CounterpartyErrors.MiscError.Code, errors.New(consts.CounterpartyErrors.MiscError.Description)
	}

	responseData, errorCode, err := postAPI(c, payloadJsonBytes)
	if err != nil {
		// Counterparty DB is behind backend / reparsing or timed out, read directly from DB
		if errorCode == consts.CounterpartyErrors.ReparsingOrUnavailable.Code || errorCode == consts.CounterpartyErrors.Timeout.Code {
			return GetBalancesByAssetDB(c, asset)
		}

		return result, errorCode, err
	}

	// Range over the result from api and create the reply
	if responseData["result"] != nil {
		for _, b := range responseData["result"].([]interface{}) {
			c := b.(map[string]interface{})
			result = append(result,
				Balance{Address: c["address"].(string),
					Asset:    c["asset"].(string),
					Quantity: uint64(c["quantity"].(float64))})
		}
	}

	return result, 0, nil
}

func GetBalancesByAssetDB(c context.Context, asset string) ([]Balance, int64, error) {
	var result []Balance

	// sqlite drivers are not concurrency safe, so must create a connection each time
	db, err := sql.Open("sqlite3", counterpartyDBLocation)
	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Failed to open DB. Reason: %s", err.Error())
		return result, consts.CounterpartyErrors.MiscError.Code, errors.New(consts.CounterpartyErrors.MiscError.Description)
	}

	err = db.Ping()
	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Failed to ping DB. Reason: %s", err.Error())
		return result, consts.CounterpartyErrors.MiscError.Code, errors.New(consts.CounterpartyErrors.MiscError.Description)
	}

	//	 Query DB
	//	log.Fluentf(consts.LOGDEBUG, "select address, asset, quantity from balances where asset = %s", asset)
	stmt, err := db.Prepare("select address, asset, quantity from balances where asset = ?")
	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Failed to prepare statement. Reason: %s", err.Error())
		return result, consts.CounterpartyErrors.MiscError.Code, errors.New(consts.CounterpartyErrors.MiscError.Description)
	}
	defer stmt.Close()

	//	 Get row
	rows, err := stmt.Query(asset)
	defer rows.Close()
	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Failed to query. Reason: %s", err.Error())
		return result, consts.CounterpartyErrors.MiscError.Code, errors.New(consts.CounterpartyErrors.MiscError.Description)
	}

	for rows.Next() {
		var balance = Balance{}
		var address []byte
		var asset []byte
		var quantity uint64

		if err := rows.Scan(&address, &asset, &quantity); err == sql.ErrNoRows {
			if err.Error() == "sql: no rows in result set" {
			}
		} else if err != nil {
			log.FluentfContext(consts.LOGERROR, c, "Failed to Scan. Reason: %s", err.Error())
		} else {
			balance = Balance{Address: string(address), Asset: string(asset), Quantity: quantity}
		}

		result = append(result, balance)
	}

	return result, 0, nil
}

func GetIssuances(c context.Context, asset string) ([]Issuance, int64, error) {
	var payload payloadGetIssuances
	var result []Issuance

	if isInit == false {
		Init()
	}
	filterCondition := filter{Field: "asset", Op: "==", Value: asset}
	filterCondition2 := filter{Field: "status", Op: "==", Value: "valid"}

	payload.Method = "get_issuances"
	payload.Params.OrderBy = "tx_index"
	payload.Params.OrderDir = "asc"
	payload.Params.Filters = append(payload.Params.Filters, filterCondition)
	payload.Params.Filters = append(payload.Params.Filters, filterCondition2)
	payload.Jsonrpc = "2.0"
	payload.Id = generateId(c)

	payloadJsonBytes, err := json.Marshal(payload)
	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Error in Marshal(): %s", err.Error())
		return result, consts.CounterpartyErrors.MiscError.Code, errors.New(consts.CounterpartyErrors.MiscError.Description)
	}

	responseData, errorCode, err := postAPI(c, payloadJsonBytes)
	if err != nil {
		// Counterparty DB is behind backend / reparsing or timed out, read directly from DB
		if errorCode == consts.CounterpartyErrors.ReparsingOrUnavailable.Code || errorCode == consts.CounterpartyErrors.Timeout.Code {
			return GetIssuancesDB(c, asset)
		}

		return result, errorCode, err
	}

	// Range over the result from api and create the reply
	if responseData["result"] != nil {
		for _, b := range responseData["result"].([]interface{}) {
			c := b.(map[string]interface{})
			result = append(result,
				Issuance{TxIndex: uint64(c["tx_index"].(float64)),
					TxHash:      c["tx_hash"].(string),
					BlockIndex:  uint64(c["block_index"].(float64)),
					Asset:       c["asset"].(string),
					Quantity:    uint64(c["quantity"].(float64)),
					Divisible:   uint64(c["divisible"].(float64)),
					Source:      c["source"].(string),
					Issuer:      c["issuer"].(string),
					Transfer:    uint64(c["transfer"].(float64)),
					Description: c["description"].(string),
					FeePaid:     uint64(c["fee_paid"].(float64)),
					Locked:      uint64(c["locked"].(float64)),
					Status:      c["status"].(string)})
		}
	}

	return result, 0, nil
}

func GetIssuancesDB(c context.Context, asset string) ([]Issuance, int64, error) {
	var result []Issuance

	// sqlite drivers are not concurrency safe, so must create a connection each time
	db, err := sql.Open("sqlite3", counterpartyDBLocation)
	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Failed to open DB. Reason: %s", err.Error())
		return result, consts.CounterpartyErrors.MiscError.Code, errors.New(consts.CounterpartyErrors.MiscError.Description)
	}

	err = db.Ping()
	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Failed to ping DB. Reason: %s", err.Error())
		return result, consts.CounterpartyErrors.MiscError.Code, errors.New(consts.CounterpartyErrors.MiscError.Description)
	}

	//	 Query DB
	//	log.Fluentf(consts.LOGDEBUG, "select tx_index, tx_hash, block_index, asset, quantity, divisible, source, issuer, transfer, description, fee_paid, locked, status from issuances where status='valid' and asset=%s", asset)
	stmt, err := db.Prepare("select tx_index, tx_hash, block_index, asset, quantity, divisible, source, issuer, transfer, description, fee_paid, locked, status from issuances where status='valid' and asset=?")
	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Failed to prepare statement. Reason: %s", err.Error())
		return result, consts.CounterpartyErrors.MiscError.Code, errors.New(consts.CounterpartyErrors.MiscError.Description)
	}
	defer stmt.Close()

	//	 Get row
	rows, err := stmt.Query(asset)
	defer rows.Close()
	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Failed to query. Reason: %s", err.Error())
		return result, consts.CounterpartyErrors.MiscError.Code, errors.New(consts.CounterpartyErrors.MiscError.Description)
	}

	for rows.Next() {
		var issuance = Issuance{}
		var tx_index uint64
		var tx_hash []byte
		var block_index uint64
		var asset []byte
		var quantity uint64
		var divisible []byte // returned as a string from the DB driver, we need to return as an int
		var source []byte
		var issuer []byte
		var transfer []byte // returned as a string from the DB driver, we need to return as an int
		var description []byte
		var fee_paid uint64
		var locked []byte // returned as a string from the DB driver, we need to return as an int
		var status []byte

		if err := rows.Scan(&tx_index, &tx_hash, &block_index, &asset, &quantity, &divisible, &source, &issuer, &transfer, &description, &fee_paid, &locked, &status); err == sql.ErrNoRows {
			if err.Error() == "sql: no rows in result set" {
			}
		} else if err != nil {
			log.FluentfContext(consts.LOGERROR, c, "Failed to Scan. Reason: %s", err.Error())
		} else {
			var divisibleResult uint64
			if string(divisible) == "true" {
				divisibleResult = 1
			} else {
				divisibleResult = 0
			}

			var transferResult uint64
			if string(transfer) == "true" {
				transferResult = 1
			} else {
				transferResult = 0
			}

			var lockedResult uint64
			if string(locked) == "true" {
				lockedResult = 1
			} else {
				lockedResult = 0
			}

			issuance = Issuance{TxIndex: tx_index, TxHash: string(tx_hash), BlockIndex: block_index, Asset: string(asset), Quantity: quantity, Divisible: divisibleResult, Source: string(source), Issuer: string(issuer), Transfer: transferResult, Description: string(description), FeePaid: fee_paid, Locked: lockedResult, Status: string(status)}
		}

		result = append(result, issuance)
	}

	return result, 0, nil
}

// Generates a hex string serialed tx which contains the bitcoin transaction to send an asset from sourceAddress to destinationAddress
// Not exposed to the public
func CreateSend(c context.Context, sourceAddress string, destinationAddress string, asset string, quantity uint64, pubKeyHexString string) (string, int64, error) {
	var payload payloadCreateSend_Counterparty
	var result string

	if isInit == false {
		Init()
	}

	//	log.Println("In counterpartyapi.CreateSend()")

	// ["source":sourceAddress,"destination":destinationAddress,"asset":asset,"quantity":amount,"allow_unconfirmed_inputs":true,"encoding":counterpartyTransactionEncoding,"pubkey":pubkey]
	payload.Method = "create_send"
	payload.Jsonrpc = "2.0"
	payload.Id = generateId(c)
	payload.Params.Source = sourceAddress
	payload.Params.Destination = destinationAddress
	payload.Params.Asset = asset
	payload.Params.Quantity = quantity
	payload.Params.AllowUnconfirmedInputs = "true"
	payload.Params.Encoding = counterpartyTransactionEncoding
	payload.Params.PubKey = pubKeyHexString
	payload.Params.Fee = Counterparty_DefaultTxFee
	payload.Params.DustSize = Counterparty_DefaultDustSize

	// Marshal into json
	payloadJsonBytes, err := json.Marshal(payload)
	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Error in Marshal(): %s", err.Error())
		return "", consts.CounterpartyErrors.MiscError.Code, errors.New(consts.CounterpartyErrors.MiscError.Description)
	}

	// Post the request to counterpartyd
	responseData, errorCode, err := postAPI(c, payloadJsonBytes)
	if err != nil {
		return "", errorCode, err
	}

	// Get the result
	if responseData["result"] != nil {
		result = responseData["result"].(string)
	}

	return result, 0, nil
}

// When given the 12 word passphrase:
// 1) Parses the raw TX to find the address being sent from
// 2) Derives the parent key and the child key for the address found in step 1)
// 3) Signs all the TX inputs
//
// Assumptions
// 1) This is a Counterparty transaction so all inputs need to be signed with the same pubkeyhash
func signRawTransaction(c context.Context, passphrase string, rawTxHexString string) (string, error) {
	// Convert the hex string to a byte array
	txBytes, err := hex.DecodeString(rawTxHexString)
	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Error in DecodeString(): %s", err.Error())
		return "", err
	}

	//	log.Printf("Unsigned tx: %s", rawTxHexString)

	// Deserialise the transaction
	tx, err := btcutil.NewTxFromBytes(txBytes)
	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Error in NewTxFromBytes(): %s", err.Error())
		return "", err
	}
	//	log.Printf("Deserialised ok!: %+v", tx)

	msgTx := tx.MsgTx()
	//	log.Printf("MsgTx: %+v", msgTx)
	//	log.Printf("Number of txes in: %d\n", len(msgTx.TxIn))
	for i := 0; i <= len(msgTx.TxIn)-1; i++ {
		//		log.Printf("MsgTx.TxIn[%d]:\n", i)
		//		log.Printf("TxIn[%d].PreviousOutPoint.Hash: %s\n", i, msgTx.TxIn[i].PreviousOutPoint.Hash)
		//		log.Printf("TxIn[%d].PreviousOutPoint.Index: %d\n", i, msgTx.TxIn[i].PreviousOutPoint.Index)
		//		log.Printf("TxIn[%d].SignatureScript: %s\n", i, hex.EncodeToString(msgTx.TxIn[i].SignatureScript))
		//		scriptHex := "76a914128004ff2fcaf13b2b91eb654b1dc2b674f7ec6188ac"
		script := msgTx.TxIn[i].SignatureScript

		//		disasm, err := txscript.DisasmString(script)
		//		if err != nil {
		//			return "", err
		//		}
		//		log.Printf("TxIn[%d] Script Disassembly: %s", i, disasm)

		// Extract and print details from the script.
		//		scriptClass, addresses, reqSigs, err := txscript.ExtractPkScriptAddrs(script, &chaincfg.MainNetParams)
		scriptClass, _, _, err := txscript.ExtractPkScriptAddrs(script, &chaincfg.MainNetParams)
		if err != nil {
			log.FluentfContext(consts.LOGERROR, c, "Error in ExtractPkScriptAddrs(): %s", err.Error())
			return "", err
		}

		// This function only supports pubkeyhash signing at this time (ie not multisig or P2SH)
		//				log.Printf("TxIn[%d] Script Class: %s\n", i, scriptClass)
		if scriptClass.String() != "pubkeyhash" {
			return "", errors.New("Counterparty_SignRawTransaction() currently only supports pubkeyhash script signing. However, the script type in the TX to sign was: " + scriptClass.String())
		}

		//		log.Printf("TxIn[%d] Addresses: %s\n", i, addresses)
		//		log.Printf("TxIn[%d] Required Signatures: %d\n", i, reqSigs)
	}

	msgScript := msgTx.TxIn[0].SignatureScript

	// Callback to look up the signing key
	lookupKey := func(a btcutil.Address) (*btcec.PrivateKey, bool, error) {
		address := a.String()

		//		log.Printf("Looking up the private key for: %s\n", address)

		privateKeyString, err := counterpartycrypto.GetPrivateKey(passphrase, address)
		if err != nil {
			log.FluentfContext(consts.LOGERROR, c, "Error in counterpartycrypto.GetPrivateKey(): %s", err.Error())
			return nil, false, nil
		}

		//		log.Printf("Private key retrieved!\n")

		privateKeyBytes, err := hex.DecodeString(privateKeyString)
		if err != nil {
			log.FluentfContext(consts.LOGERROR, c, "Error in DecodeString(): %s", err.Error())
			return nil, false, nil
		}

		privKey, _ := btcec.PrivKeyFromBytes(btcec.S256(), privateKeyBytes)

		return privKey, true, nil
	}

	// Range over TxIns and sign
	for i, txIn := range msgTx.TxIn {
		// Get the sigscript
		// Notice that the script database parameter is nil here since it isn't
		// used.  It must be specified when pay-to-script-hash transactions are
		// being signed.
		sigScript, err := txscript.SignTxOutput(&chaincfg.MainNetParams, msgTx, 0, txIn.SignatureScript, txscript.SigHashAll, txscript.KeyClosure(lookupKey), nil, nil)
		if err != nil {
			return "", err
		}

		// Copy the signed sigscript into the tx
		msgTx.TxIn[i].SignatureScript = sigScript
	}

	// Prove that the transaction has been validly signed by executing the
	// script pair.
	flags := txscript.ScriptBip16 | txscript.ScriptVerifyDERSignatures | txscript.ScriptStrictMultiSig | txscript.ScriptDiscourageUpgradableNops
	vm, err := txscript.NewEngine(msgScript, msgTx, 0, flags)
	if err != nil {
		return "", err
	}
	if err := vm.Execute(); err != nil {
		return "", err
	}
	//	log.Println("Transaction successfully signed")

	var byteBuffer bytes.Buffer
	encodeError := msgTx.BtcEncode(&byteBuffer, wire.ProtocolVersion)

	if encodeError != nil {
		return "", err
	}

	payloadBytes := byteBuffer.Bytes()
	payloadHexString := hex.EncodeToString(payloadBytes)

	//	log.Printf("Signed and encoded transaction: %s\n", payloadHexString)

	return payloadHexString, nil
}

// Reproduces counterwallet function to generate a random asset name
// Original JS:
//self.generateRandomId = function() {
//    var r = bigInt.randBetween(NUMERIC_ASSET_ID_MIN, NUMERIC_ASSET_ID_MAX);
//    self.name('A' + r);
//}
func generateRandomAssetName(c context.Context) (string, error) {
	numericAssetIdMin := new(big.Int)
	numericAssetIdMax := new(big.Int)
	//	var err error

	numericAssetIdMin.SetString(numericAssetIdMinString, 10)
	numericAssetIdMax.SetString(numericAssetIdMaxString, 10)

	//	log.Printf("numericAssetIdMax: %s", numericAssetIdMin.String())
	//	log.Printf("numericAssetIdMin: %s", numericAssetIdMax.String())

	numericAssetIdMax = numericAssetIdMax.Add(numericAssetIdMax, numericAssetIdMin)

	x, err := rand.Int(rand.Reader, numericAssetIdMax)
	xFinal := x.Sub(x, numericAssetIdMin)

	if err != nil {
		return "", err
	}

	return "A" + string(xFinal.String()), nil
}

// Generates unsigned hex encoded transaction to issue an asset on Counterparty
// This function MUST NOT be accessed by the client directly. The high level function Counterparty_CreateIssuanceAndSend() should be used instead.
func createIssuance(c context.Context, sourceAddress string, asset string, description string, quantity uint64, divisible bool, pubKeyHexString string) (string, int64, error) {
	var payload payloadCreateIssuance_Counterparty
	var result string

	if isInit == false {
		Init()
	}

	payload.Method = "create_issuance"
	payload.Jsonrpc = "2.0"
	payload.Id = generateId(c)
	payload.Params.Source = sourceAddress
	payload.Params.Asset = asset
	payload.Params.Description = description
	payload.Params.Quantity = quantity
	payload.Params.AllowUnconfirmedInputs = "true"
	payload.Params.Encoding = counterpartyTransactionEncoding
	payload.Params.PubKey = pubKeyHexString
	payload.Params.Fee = Counterparty_DefaultTxFee
	payload.Params.DustSize = Counterparty_DefaultDustSize

	// Marshal into json
	payloadJsonBytes, err := json.Marshal(payload)
	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Error in Marshal(): %s", err.Error())
		return "", consts.CounterpartyErrors.MiscError.Code, errors.New(consts.CounterpartyErrors.MiscError.Description)
	}

	// Post the request to counterpartyd
	responseData, errorCode, err := postAPI(c, payloadJsonBytes)
	if err != nil {
		return "", errorCode, err
	}

	// Get the result
	if responseData["result"] != nil {
		result = responseData["result"].(string)
	}

	return result, 0, nil
}

// Automatically generates a numeric asset name and generates unsigned hex encoded transaction to issue an asset on Counterparty
// Returns:
// randomAssetName that was generated
// hex string encoded transaction
// errorCode
// error
func CreateNumericIssuance(c context.Context, sourceAddress string, asset string, quantity uint64, divisible bool, pubKeyHexString string) (string, string, int64, error) {
	var description string

	if isInit == false {
		Init()
	}

	if len(asset) > 52 {
		description = asset[0:51]
	} else {
		description = asset
	}

	// Generate random asset name
	var err error
	var randomAssetName string
	randomAssetName, err = generateRandomAssetName(c)

	// If random asset name already exists, keep trying until we find a spare one
	for balance, errorCode, err := GetBalancesByAsset(c, randomAssetName); len(balance) != 0; {
		if err != nil {
			log.FluentfContext(consts.LOGERROR, c, "Error in GetBalancesByAsset(): %s, errorCode: %d", err.Error(), errorCode)
			return "", "", errorCode, err
		}
		randomAssetName, err = generateRandomAssetName(c)
		if err != nil {
			log.FluentfContext(consts.LOGERROR, c, "Error in generateRandomAssetName(): %s", err.Error())
			return "", "", consts.CounterpartyErrors.MiscError.Code, errors.New(consts.CounterpartyErrors.MiscError.Description)
		}

		balance, errorCode, err = GetBalancesByAsset(c, randomAssetName)
	}

	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Error in after trying to check if asset exists: %s", err.Error())
		return "", "", consts.CounterpartyErrors.MiscError.Code, errors.New(consts.CounterpartyErrors.MiscError.Description)
	}

	// Call counterparty to create the issuance
	result, errorCode, err := createIssuance(c, sourceAddress, randomAssetName, description, quantity, divisible, pubKeyHexString)
	if err != nil {
		return "", "", errorCode, err
	}

	return randomAssetName, result, 0, nil
}

// Generates unsigned hex encoded transaction to pay a dividend on an asset on Counterparty
func createDividend(c context.Context, sourceAddress string, asset string, dividendAsset string, quantityPerUnit uint64, pubKeyHexString string) (string, int64, error) {
	var payload payloadCreateDividend_Counterparty
	var result string

	if isInit == false {
		Init()
	}

	payload.Method = "create_dividend"
	payload.Jsonrpc = "2.0"
	payload.Id = generateId(c)
	payload.Params.Source = sourceAddress
	payload.Params.Asset = asset
	payload.Params.DividendAsset = dividendAsset
	payload.Params.QuantityPerUnit = quantityPerUnit
	payload.Params.Encoding = counterpartyTransactionEncoding
	payload.Params.PubKey = pubKeyHexString
	payload.Params.Fee = Counterparty_DefaultTxFee
	payload.Params.DustSize = Counterparty_DefaultDustSize

	// Marshal into json
	payloadJsonBytes, err := json.Marshal(payload)
	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Error in Marshal(): %s", err.Error())
		return result, consts.CounterpartyErrors.MiscError.Code, errors.New(consts.CounterpartyErrors.MiscError.Description)
	}

	// Post the request to counterpartyd
	responseData, errorCode, err := postAPI(c, payloadJsonBytes)
	if err != nil {
		return "", errorCode, err
	}

	// Get the result
	if responseData["result"] != nil {
		result = responseData["result"].(string)
	}

	return result, 0, nil
}

// Concurrency safe to create and send transactions from a single address.
func DelegatedSend(c context.Context, accessKey string, passphrase string, sourceAddress string, destinationAddress string, asset string, quantity uint64, paymentId string, paymentTag string) (string, int64, error) {
	if isInit == false {
		Init()
	}

	// Copy same context values to local variables which are often accessed
	env := c.Value(consts.EnvKey).(string)

	// Write the payment with the generated payment id to the database
	go database.InsertPayment(c, accessKey, 0, paymentId, sourceAddress, destinationAddress, asset, quantity, "valid", 0, 1500, paymentTag)

	sourceAddressPubKey, err := counterpartycrypto.GetPublicKey(passphrase, sourceAddress)
	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Err in GetPublicKey(): %s\n", err.Error())
		database.UpdatePaymentWithErrorByPaymentId(c, accessKey, paymentId, consts.CounterpartyErrors.InvalidPassphrase.Code, consts.CounterpartyErrors.InvalidPassphrase.Description)
		return "", consts.CounterpartyErrors.SigningError.Code, errors.New(consts.CounterpartyErrors.SigningError.Description)
	}

	// Mutex lock this address
	counterparty_Mutexes.Lock()
	log.FluentfContext(consts.LOGINFO, c, "Locked the map") // The map of mutexes must be locked before we modify the mutexes stored in the map

	// If an entry doesn't currently exist in the map for that address
	if counterparty_Mutexes.m[sourceAddress] == nil {
		log.FluentfContext(consts.LOGINFO, c, "Created new entry in map for %s", sourceAddress)
		counterparty_Mutexes.m[sourceAddress] = new(sync.Mutex)
	}

	counterparty_Mutexes.m[sourceAddress].Lock()
	log.FluentfContext(consts.LOGINFO, c, "Locked: %s\n", sourceAddress)

	defer counterparty_Mutexes.Unlock()
	defer counterparty_Mutexes.m[sourceAddress].Unlock()

	// We must sleep for at least the time it takes for any transactions to propagate through to the counterparty mempool
	log.FluentfContext(consts.LOGINFO, c, "Sleeping")
	time.Sleep(time.Duration(counterparty_BackEndPollRate+3000) * time.Millisecond)

	log.FluentfContext(consts.LOGINFO, c, "Sleep complete")

	// Create the send
	createResult, errorCode, err := CreateSend(c, sourceAddress, destinationAddress, asset, quantity, sourceAddressPubKey)
	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Err in CreateSend(): %s", err.Error())
		database.UpdatePaymentWithErrorByPaymentId(c, accessKey, paymentId, errorCode, err.Error())
		return "", errorCode, err
	}

	log.FluentfContext(consts.LOGINFO, c, "Created send of %d %s to %s: %s", quantity, asset, destinationAddress, createResult)

	// Sign the transactions
	signed, err := signRawTransaction(c, passphrase, createResult)
	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Err in SignRawTransaction(): %s\n", err.Error())
		database.UpdatePaymentWithErrorByPaymentId(c, accessKey, paymentId, consts.CounterpartyErrors.SigningError.Code, consts.CounterpartyErrors.SigningError.Description)
		return "", consts.CounterpartyErrors.SigningError.Code, errors.New(consts.CounterpartyErrors.SigningError.Description)
	}

	log.FluentfContext(consts.LOGINFO, c, "Signed tx: %s", signed)

	// Update the DB with the raw signed TX. This will allow re-transmissions if something went wrong with sending on the network
	database.UpdatePaymentSignedRawTxByPaymentId(c, accessKey, paymentId, signed)

	//	 Transmit the transaction if not in dev, otherwise stub out the return
	var txId string
	if env != "dev" {
		txIdSignedTx, err := bitcoinapi.SendRawTransaction(signed)
		if err != nil {
			log.FluentfContext(consts.LOGERROR, c, err.Error())
			database.UpdatePaymentWithErrorByPaymentId(c, accessKey, paymentId, consts.CounterpartyErrors.BroadcastError.Code, consts.CounterpartyErrors.BroadcastError.Description)
			return "", consts.CounterpartyErrors.BroadcastError.Code, errors.New(consts.CounterpartyErrors.BroadcastError.Description)
		}

		txId = txIdSignedTx
	} else {
		txId = "success"
	}

	database.UpdatePaymentCompleteByPaymentId(c, accessKey, paymentId, txId)

	log.FluentfContext(consts.LOGINFO, c, "Complete.")

	return txId, 0, nil
}

// Concurrency safe to create and send transactions from a single address.
func DelegatedCreateIssuance(c context.Context, accessKey string, passphrase string, sourceAddress string, assetId string, asset string, quantity uint64, divisible bool) (string, int64, error) {
	if isInit == false {
		Init()
	}

	// Copy same context values to local variables which are often accessed
	env := c.Value(consts.EnvKey).(string)

	// Write the asset with the generated asset id to the database
	go database.InsertAsset(accessKey, assetId, sourceAddress, "TBA", asset, quantity, divisible, "valid")

	sourceAddressPubKey, err := counterpartycrypto.GetPublicKey(passphrase, sourceAddress)
	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Error with GetPublicKey(): %s", err)
		return "", consts.CounterpartyErrors.InvalidPassphrase.Code, errors.New(consts.CounterpartyErrors.InvalidPassphrase.Description)
	}

	// Mutex lock this address
	counterparty_Mutexes.Lock()
	log.FluentfContext(consts.LOGINFO, c, "Locked the map") // The map of mutexes must be locked before we modify the mutexes stored in the map

	// If an entry doesn't currently exist in the map for that address
	if counterparty_Mutexes.m[sourceAddress] == nil {
		log.Printf("Created new entry in map for %s\n", sourceAddress)
		counterparty_Mutexes.m[sourceAddress] = new(sync.Mutex)
	}

	counterparty_Mutexes.m[sourceAddress].Lock()
	log.FluentfContext(consts.LOGINFO, c, "Locked: %s\n", sourceAddress)

	// We must sleep for at least the time it takes for any transactions to propagate through to the counterparty mempool
	log.FluentfContext(consts.LOGINFO, c, "Sleeping")
	time.Sleep(time.Duration(counterparty_BackEndPollRate+3000) * time.Millisecond)

	defer counterparty_Mutexes.Unlock()
	defer counterparty_Mutexes.m[sourceAddress].Unlock()

	log.FluentfContext(consts.LOGINFO, c, "Composing the CreateNumericIssuance transaction")
	// Create the issuance
	randomAssetName, createResult, errCode, err := CreateNumericIssuance(c, sourceAddress, asset, quantity, divisible, sourceAddressPubKey)
	if err != nil {
		return "", errCode, err
	}

	log.FluentfContext(consts.LOGINFO, c, "Created issuance of %d %s (%s) at %s: %s\n", quantity, asset, randomAssetName, sourceAddress, createResult)
	database.UpdateAssetNameByAssetId(c, accessKey, assetId, randomAssetName)

	// Sign the transactions
	signed, err := signRawTransaction(c, passphrase, createResult)
	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Error in SignRawTransaction(): %s", err.Error())

		database.UpdateAssetWithErrorByAssetId(c, accessKey, assetId, consts.CounterpartyErrors.SigningError.Code, consts.CounterpartyErrors.SigningError.Description)
		return "", consts.CounterpartyErrors.SigningError.Code, errors.New(consts.CounterpartyErrors.SigningError.Description)
	}

	log.FluentfContext(consts.LOGINFO, c, "Signed tx: %s\n", signed)

	//	 Transmit the transaction if not in dev, otherwise stub out the return
	var txId string
	if env != "dev" {
		txIdSignedTx, err := bitcoinapi.SendRawTransaction(signed)
		if err != nil {
			log.FluentfContext(consts.LOGERROR, c, "Error in SendRawTransaction(): %s", err.Error())
			database.UpdateAssetWithErrorByAssetId(c, accessKey, assetId, consts.CounterpartyErrors.BroadcastError.Code, consts.CounterpartyErrors.BroadcastError.Description)
			return "", consts.CounterpartyErrors.BroadcastError.Code, errors.New(consts.CounterpartyErrors.BroadcastError.Description)
		}

		txId = txIdSignedTx
	} else {
		txId = "success"
	}

	database.UpdateAssetCompleteByAssetId(c, accessKey, assetId, txId)

	return txId, 0, nil
}

// Concurrency safe to create and send transactions from a single address.
func DelegatedCreateDividend(c context.Context, accessKey string, passphrase string, dividendId string, sourceAddress string, asset string, dividendAsset string, quantityPerUnit uint64) (string, int64, error) {
	if isInit == false {
		Init()
	}

	// Copy same context values to local variables which are often accessed
	env := c.Value(consts.EnvKey).(string)

	// Write the dividend with the generated dividend id to the database
	go database.InsertDividend(accessKey, dividendId, sourceAddress, asset, dividendAsset, quantityPerUnit, "valid")

	sourceAddressPubKey, err := counterpartycrypto.GetPublicKey(passphrase, sourceAddress)
	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Err in GetPublicKey(): %s\n", err.Error())
		database.UpdateDividendWithErrorByDividendId(c, accessKey, dividendId, consts.CounterpartyErrors.InvalidPassphrase.Code, consts.CounterpartyErrors.InvalidPassphrase.Description)
		return "", consts.CounterpartyErrors.InvalidPassphrase.Code, errors.New(consts.CounterpartyErrors.InvalidPassphrase.Description)
	}

	// Mutex lock this address
	counterparty_Mutexes.Lock()
	log.FluentfContext(consts.LOGINFO, c, "Locked the map") // The map of mutexes must be locked before we modify the mutexes stored in the map

	// If an entry doesn't currently exist in the map for that address
	if counterparty_Mutexes.m[sourceAddress] == nil {
		log.FluentfContext(consts.LOGINFO, c, "Created new entry in map for %s\n", sourceAddress)
		counterparty_Mutexes.m[sourceAddress] = new(sync.Mutex)
	}

	counterparty_Mutexes.m[sourceAddress].Lock()
	log.FluentfContext(consts.LOGINFO, c, "Locked: %s", sourceAddress)

	// We must sleep for at least the time it takes for any transactions to propagate through to the counterparty mempool
	log.FluentfContext(consts.LOGINFO, c, "Sleeping")
	time.Sleep(time.Duration(counterparty_BackEndPollRate+3000) * time.Millisecond)

	defer counterparty_Mutexes.Unlock()
	defer counterparty_Mutexes.m[sourceAddress].Unlock()

	// Create the dividend
	createResult, errorCode, err := createDividend(c, sourceAddress, asset, dividendAsset, quantityPerUnit, sourceAddressPubKey)
	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Error in CreateDividend(): %s errorCode: %d", err.Error(), errorCode)
		database.UpdateDividendWithErrorByDividendId(c, accessKey, dividendId, consts.CounterpartyErrors.ComposeError.Code, consts.CounterpartyErrors.ComposeError.Description)
		return "", errorCode, err
	}

	log.FluentfContext(consts.LOGINFO, c, "Created dividend of %d %s for each %s from address %s: %s\n", quantityPerUnit, dividendAsset, asset, sourceAddress, createResult)

	// Sign the transactions
	signed, err := signRawTransaction(c, passphrase, createResult)
	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Error in SignRawTransaction: %s", err.Error())
		database.UpdateDividendWithErrorByDividendId(c, accessKey, dividendId, consts.CounterpartyErrors.SigningError.Code, consts.CounterpartyErrors.SigningError.Description)
		return "", consts.CounterpartyErrors.SigningError.Code, errors.New(consts.CounterpartyErrors.SigningError.Description)
	}

	log.FluentfContext(consts.LOGINFO, c, "Signed tx: %s", signed)

	//	 Transmit the transaction if not in dev, otherwise stub out the return
	var txId string
	if env != "dev" {
		txIdSignedTx, err := bitcoinapi.SendRawTransaction(signed)
		if err != nil {
			log.FluentfContext(consts.LOGERROR, c, "Error in SendRawTransaction(): %s", err.Error())
			database.UpdateDividendWithErrorByDividendId(c, accessKey, dividendId, consts.CounterpartyErrors.BroadcastError.Code, consts.CounterpartyErrors.BroadcastError.Description)
			return "", consts.CounterpartyErrors.BroadcastError.Code, errors.New(consts.CounterpartyErrors.BroadcastError.Description)
		}

		txId = txIdSignedTx
	} else {
		txId = "success"
	}

	database.UpdateDividendCompleteByDividendId(c, accessKey, dividendId, txId)

	return txId, 0, nil
}

// For internal use only - don't expose to customers
func GetRunningInfo(c context.Context) (RunningInfo, int64, error) {
	var payload payloadGetRunningInfo
	var result RunningInfo

	if isInit == false {
		Init()
	}

	payload.Method = "get_running_info"
	payload.Jsonrpc = "2.0"
	payload.Id = generateId(c)

	payloadJsonBytes, err := json.Marshal(payload)
	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Error in Marshal(): %s", err.Error())
		return result, consts.CounterpartyErrors.MiscError.Code, errors.New(consts.CounterpartyErrors.MiscError.Description)
	}

	responseData, errorCode, err := postAPI(c, payloadJsonBytes)
	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Error in postAPI(): %s", err.Error())
		return result, errorCode, err
	}

	// Get result from api and create the reply
	if responseData["result"] != nil {
		resultMap := responseData["result"].(map[string]interface{})
		lastBlockMap := resultMap["last_block"].(map[string]interface{})
		//		log.Printf("%#v\n", resultMap)
		//		log.Printf("%#v\n", lastBlockMap)
		result = RunningInfo{
			DbCaughtUp:           resultMap["db_caught_up"].(bool),
			BitCoinBlockCount:    uint64(resultMap["bitcoin_block_count"].(float64)),
			CounterpartydVersion: string(uint64(resultMap["version_major"].(float64))) + "." + string(uint64(resultMap["version_minor"].(float64))) + "." + string(uint64(resultMap["version_revision"].(float64))),
			LastMessageIndex:     uint64(resultMap["last_message_index"].(float64)),
			RunningTestnet:       resultMap["running_testnet"].(bool),
			LastBlock: LastBlock{
				BlockIndex: uint64(lastBlockMap["block_index"].(float64)),
				BlockHash:  lastBlockMap["block_hash"].(string),
			},
		}
	}

	return result, 0, nil
}

// Concurrency safe to create and send transactions from a single address.
func DelegatedActivateAddress(c context.Context, addressToActivate string, amount uint64, activationId string) (string, int64, error) {
	if isInit == false {
		Init()
	}

	// Copy same context values to local variables which are often accessed
	accessKey := c.Value(consts.AccessKeyKey).(string)
	blockchainId := c.Value(consts.BlockchainIdKey).(string)
	env := c.Value(consts.EnvKey).(string)

	// Need a better way to secure internal wallets
	// Array of internal wallets that can be round robined to activate addresses
	var wallets = []struct {
		Address      string
		Passphrase   string
		BlockchainId string
	}{
		{"1E5YgFkC4HNHwWTF5iUdDbKpzry1SRLv8e", "bound social cookie wrong yet story cigarette descend metal drug waste candle", "counterparty"},
	}

	// Pick an internal address to send from
	var randomNumber int = 0
	var sourceAddress = wallets[randomNumber].Address

	// Write the dividend with the generated dividend id to the database
	database.InsertActivation(c, accessKey, activationId, blockchainId, sourceAddress, amount)

	// Calculate the quantity of BTC to send by the amount specified
	// For Counterparty: each transaction = dust_size + miners_fee
	quantity, asset, err := CalculateFeeAmount(c, amount)
	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Could not calculate fee: %s", err.Error())
		database.UpdatePaymentWithErrorByPaymentId(c, accessKey, activationId, consts.CounterpartyErrors.MiscError.Code, consts.CounterpartyErrors.MiscError.Description)
		return "", consts.CounterpartyErrors.MiscError.Code, errors.New(consts.CounterpartyErrors.MiscError.Description)
	}

	sourceAddressPubKey, err := counterpartycrypto.GetPublicKey(wallets[randomNumber].Passphrase, sourceAddress)
	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Err in GetPublicKey(): %s\n", err.Error())
		database.UpdatePaymentWithErrorByPaymentId(c, accessKey, activationId, consts.CounterpartyErrors.InvalidPassphrase.Code, consts.CounterpartyErrors.InvalidPassphrase.Description)
		return "", consts.CounterpartyErrors.InvalidPassphrase.Code, errors.New(consts.CounterpartyErrors.InvalidPassphrase.Description)
	}

	// Mutex lock this address
	counterparty_Mutexes.Lock()
	log.FluentfContext(consts.LOGINFO, c, "Locked the map") // The map of mutexes must be locked before we modify the mutexes stored in the map

	// If an entry doesn't currently exist in the map for that address
	if counterparty_Mutexes.m[sourceAddress] == nil {
		log.FluentfContext(consts.LOGINFO, c, "Created new entry in map for %s\n", sourceAddress)
		counterparty_Mutexes.m[sourceAddress] = new(sync.Mutex)
	}

	counterparty_Mutexes.m[sourceAddress].Lock()
	log.FluentfContext(consts.LOGINFO, c, "Locked: %s\n", sourceAddress)

	// We must sleep for at least the time it takes for any transactions to propagate through to the counterparty mempool
	log.FluentfContext(consts.LOGINFO, c, "Sleeping")
	time.Sleep(time.Duration(counterparty_BackEndPollRate+3000) * time.Millisecond)

	defer counterparty_Mutexes.Unlock()
	defer counterparty_Mutexes.m[sourceAddress].Unlock()

	// Write the payment using the activationId as the paymentId to the db
	go database.InsertPayment(c, accessKey, 0, activationId, sourceAddress, addressToActivate, asset, quantity, "valid", 0, 1500, "")

	// Create the send from the internal wallet to the destination address
	createResult, errorCode, err := CreateSend(c, sourceAddress, addressToActivate, asset, quantity, sourceAddressPubKey)
	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Error in createSend(): %s", err.Error())
		database.UpdatePaymentWithErrorByPaymentId(c, accessKey, activationId, errorCode, err.Error())
		return "", errorCode, err
	}

	log.FluentfContext(consts.LOGINFO, c, "Created send of %d %s to %s: %s", quantity, asset, addressToActivate, createResult)

	// Sign the transactions
	signed, err := signRawTransaction(c, wallets[randomNumber].Passphrase, createResult)
	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Error in signRawTransaction(): %s", err.Error())
		database.UpdatePaymentWithErrorByPaymentId(c, accessKey, activationId, consts.CounterpartyErrors.SigningError.Code, consts.CounterpartyErrors.SigningError.Description)
		return "", consts.CounterpartyErrors.SigningError.Code, errors.New(consts.CounterpartyErrors.SigningError.Description)
	}

	log.FluentfContext(consts.LOGINFO, c, "Signed tx: %s\n", signed)

	//	 Transmit the transaction if not in dev, otherwise stub out the return
	var txId string
	if env != "dev" {
		txIdSignedTx, err := bitcoinapi.SendRawTransaction(signed)
		if err != nil {
			log.FluentfContext(consts.LOGERROR, c, "Error in SendRawTransaction(): %s", err.Error())
			database.UpdatePaymentWithErrorByPaymentId(c, accessKey, activationId, consts.CounterpartyErrors.BroadcastError.Code, consts.CounterpartyErrors.BroadcastError.Description)
			return "", consts.CounterpartyErrors.BroadcastError.Code, errors.New(consts.CounterpartyErrors.BroadcastError.Description)
		}

		txId = txIdSignedTx
	} else {
		txId = "success"
	}

	database.UpdatePaymentCompleteByPaymentId(c, accessKey, activationId, txId)

	return txId, 0, nil
}

// Returns the total BTC that is required for the given number of transactions
func CalculateFeeAmount(c context.Context, amount uint64) (uint64, string, error) {
	// Get env and blockchain from context
	env := c.Value(consts.EnvKey).(string)
	blockchainId := c.Value(consts.BlockchainIdKey).(string)

	// Set some maximum and minimums
	var thisAmount = amount
	if thisAmount > 1000 {
		thisAmount = 1000
	}
	if thisAmount < 20 {
		thisAmount = 20
	}

	if blockchainId != consts.CounterpartyBlockchainId {
		errorString := fmt.Sprintf("Blockchain must be %s, got %s", consts.CounterpartyBlockchainId, blockchainId)
		log.FluentfContext(consts.LOGERROR, c, errorString)

		return 0, "", errors.New(errorString)
	}

	var quantity uint64
	var asset string

	if env == "dev" {
		quantity = (Counterparty_DefaultDustSize + Counterparty_DefaultTestingTxFee) * thisAmount
		asset = "BTC"
	} else {
		quantity = (Counterparty_DefaultDustSize + Counterparty_DefaultTxFee) * thisAmount
		asset = "BTC"
	}

	return quantity, asset, nil
}

// Returns the number of transactions that can be performed with the given amount of BTC
// If the env value is not found in the context, calculations are defaulted to production
func CalculateNumberOfTransactions(c context.Context, amount uint64) (uint64, error) {
	// Get env and blockchain from context
	env := c.Value(consts.EnvKey).(string)
	blockchainId := c.Value(consts.BlockchainIdKey).(string)

	if blockchainId != consts.CounterpartyBlockchainId {
		errorString := fmt.Sprintf("Blockchain must be %s, got %s", consts.CounterpartyBlockchainId, blockchainId)
		log.FluentfContext(consts.LOGERROR, c, errorString)

		return 0, errors.New(errorString)
	}

	if env == "dev" {
		return amount / (Counterparty_DefaultDustSize + Counterparty_DefaultTestingTxFee), nil
	} else {
		return amount / (Counterparty_DefaultDustSize + Counterparty_DefaultTxFee), nil
	}
}
