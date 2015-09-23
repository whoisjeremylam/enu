package counterpartyapi

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io/ioutil"
	"math/big"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

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

//   tx_index INTEGER PRIMARY KEY,
//   tx_hash TEXT UNIQUE,
//   block_index INTEGER,
//   asset TEXT,
//   quantity INTEGER,
//   divisible BOOL,
//   source TEXT,
//   issuer TEXT,
//   transfer BOOL,
//   callable BOOL,
//   call_date INTEGER,
//   call_price REAL,
//   description TEXT,
//   fee_paid INTEGER,
//   locked BOOL,
//   status TEXT,
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
	DbVersionMajor       uint64    `json:"db_version_major"`
	DbVersionMinor       uint64    `json:"db_version_minor"`
	LastBlock            LastBlock `json:"last_block"`
}

// Globals
var isInit bool = false // set to true only after the init sequence is complete
var counterpartyHost string
var counterpartyUser string
var counterpartyPassword string
var counterpartyTransactionEncoding string

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

	// Bitcoin API parameters
	counterpartyHost = m["counterpartyhost"].(string)                               // End point for JSON RPC server
	counterpartyUser = m["counterpartyuser"].(string)                               // Basic authentication user name
	counterpartyPassword = m["counterpartypassword"].(string)                       // Basic authentication password
	counterpartyTransactionEncoding = m["counterpartytransactionencoding"].(string) // The encoding that should be used for Counterparty transactions "auto" will let Counterparty select, valid values "multisig", "opreturn"

	isInit = true
}

func postAPI(c context.Context, postData []byte) ([]byte, int64, error) {
	postDataJson := string(postData)

	log.FluentfContext(consts.LOGDEBUG, c, "counterpartyapi postAPI() posting: %s", postDataJson)

	// Set headers
	req, err := http.NewRequest("POST", counterpartyHost, bytes.NewBufferString(postDataJson))
	req.SetBasicAuth(counterpartyUser, counterpartyPassword)
	req.Header.Set("Content-Type", "application/json")

	clientPointer := &http.Client{}
	resp, err := clientPointer.Do(req)

	if err != nil {
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

func GetBalancesByAddress(c context.Context, address string) ([]Balance, error) {
	var payload payloadGetBalances
	var result ResultGetBalances

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
		return nil, err
	}

	responseData, _, err := postAPI(c, payloadJsonBytes)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(responseData, &result); err != nil {
		return nil, errors.New("Unable to unmarshal responseData")
	}

	return result.Result, nil
}

func GetBalancesByAsset(c context.Context, asset string) ([]Balance, error) {
	var payload payloadGetBalances
	var result ResultGetBalances

	if isInit == false {
		Init()
	}

	filterCondition := filter{Field: "asset", Op: "==", Value: asset}

	payload.Method = "get_balances"
	payload.Params.Filters = append(payload.Params.Filters, filterCondition)
	payload.Jsonrpc = "2.0"
	payload.Id = generateId(c)

	payloadJsonBytes, err := json.Marshal(payload)

	//	log.Println(string(payloadJsonBytes))

	if err != nil {
		return nil, err
	}

	responseData, _, err := postAPI(c, payloadJsonBytes)
	if err != nil {
		return nil, err
	}

	//	log.Println(string(responseData))

	if err := json.Unmarshal(responseData, &result); err != nil {
		return nil, errors.New("Unable to unmarshal responseData")
	}

	return result.Result, nil
}

func GetIssuances(c context.Context, asset string) ([]Issuance, error) {
	var payload payloadGetIssuances
	var result ResultGetIssuances

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
		return nil, err
	}

	responseData, _, err := postAPI(c, payloadJsonBytes)
	if err != nil {
		return nil, err
	}

	log.Println(string(responseData))

	if err := json.Unmarshal(responseData, &result); err != nil {
		//		return nil, errors.New("Unable to unmarshal responseData!")
		return nil, err
	}

	return result.Result, nil
}

func CreateSend(c context.Context, sourceAddress string, destinationAddress string, asset string, quantity uint64, pubKeyHexString string) (string, error) {
	var payload payloadCreateSend_Counterparty
	var result ResultCreateSend_Counterparty

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

	// Encode the payload into JSON
	payloadJsonBytes, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	// Post to the counterparty daemon backend
	responseData, _, err := postAPI(c, payloadJsonBytes)
	if err != nil {
		return "", err
	}

	if err := json.Unmarshal(responseData, &result); err != nil {
		//		return "", errors.New("Counterparty_CreateSend(): Unable to unmarshal responseData")
		return "", err
	}

	// Try to see if we got a server error instead
	if result.Result == "" {
		return "", errors.New(string(responseData))
	}

	return result.Result, nil
}

// When given the 12 word passphrase:
// 1) Parses the raw TX to find the address being sent from
// 2) Derives the parent key and the child key for the address found in step 1)
// 3) Signs all the TX inputs
//
// Assumptions
// 1) This is a Counterparty transaction so all inputs need to be signed with the same pubkeyhash
func SignRawTransaction(c context.Context, passphrase string, rawTxHexString string) (string, error) {
	// Convert the hex string to a byte array
	txBytes, err := hex.DecodeString(rawTxHexString)
	if err != nil {
		log.Println(err.Error())
		os.Exit(-103)
	}

	//	log.Printf("Unsigned tx: %s", rawTxHexString)

	// Deserialise the transaction
	tx, err := btcutil.NewTxFromBytes(txBytes)
	if err != nil {
		log.Println(err.Error())
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
			log.Println(err.Error())
			return nil, false, nil
		}

		//		log.Printf("Private key retrieved!\n")

		privateKeyBytes, err := hex.DecodeString(privateKeyString)
		if err != nil {
			log.Println(err.Error())
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
// This function MUST NOT be called directly. The high level function Counterparty_CreateIssuanceAndSend() should be used instead.
func CreateIssuance(c context.Context, sourceAddress string, asset string, description string, quantity uint64, divisible bool, pubKeyHexString string) (string, error) {
	var payload payloadCreateIssuance_Counterparty
	var result ResultCreateIssuance_Counterparty

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

	payloadJsonBytes, err := json.Marshal(payload)

	//	log.Println(string(payloadJsonBytes))

	if err != nil {
		return "", err
	}

	responseData, _, err := postAPI(c, payloadJsonBytes)
	if err != nil {
		return "", err
	}

	//	log.Println(string(responseData))

	if err := json.Unmarshal(responseData, &result); err != nil {
		return "", errors.New("Counterparty_CreateIssuance() unable to unmarshal responseData")
	}

	// Try to see if we got a server error instead
	if result.Result == "" {
		return "", errors.New(string(responseData))
	}

	return result.Result, nil
}

// Automatically generates a numeric asset name and generates unsigned hex encoded transaction to issue an asset on Counterparty
func CreateNumericIssuance(c context.Context, sourceAddress string, asset string, quantity uint64, divisible bool, pubKeyHexString string) (string, string, error) {
	var description string

	if isInit == false {
		Init()
	}

	if len(asset) > 52 {
		description = asset[0:51]
	} else {
		description = asset
	}

	randomAssetName, err := generateRandomAssetName(c)
	if err != nil {
		return "", "", err
	}

	result, err := CreateIssuance(c, sourceAddress, randomAssetName, description, quantity, divisible, pubKeyHexString)
	if err != nil {
		return "", "", err
	}

	return randomAssetName, result, nil
}

// Generates unsigned hex encoded transaction to pay a dividend on an asset on Counterparty
func CreateDividend(c context.Context, sourceAddress string, asset string, dividendAsset string, quantityPerUnit uint64, pubKeyHexString string) (string, error) {
	var payload payloadCreateDividend_Counterparty
	var result ResultCreateDividend_Counterparty

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

	payloadJsonBytes, err := json.Marshal(payload)

	//		log.Println(string(payloadJsonBytes))

	if err != nil {
		return result.Result, err
	}

	responseData, _, err := postAPI(c, payloadJsonBytes)
	if err != nil {
		return result.Result, err
	}

	log.Println(string(responseData))

	if err := json.Unmarshal(responseData, &result); err != nil {
		return result.Result, errors.New("Unable to unmarshal responseData")
	}

	// Try to see if we got a server error instead
	if result.Result == "" {
		return result.Result, errors.New(string(responseData))
	}

	return result.Result, nil
}

// Concurrency safe to create and send transactions from a single address.
func DelegatedSend(c context.Context, accessKey string, passphrase string, sourceAddress string, destinationAddress string, asset string, quantity uint64, paymentId string, paymentTag string) (string, error) {
	if isInit == false {
		Init()
	}

	// Copy same context values to local variables which are often accessed
	env := c.Value(consts.EnvKey).(string)

	// Write the payment with the generated payment id to the database
	go database.InsertPayment(c, accessKey, 0, paymentId, sourceAddress, destinationAddress, asset, quantity, "valid", 0, 1500, paymentTag)

	sourceAddressPubKey, err := counterpartycrypto.GetPublicKey(passphrase, sourceAddress)
	if err != nil {
		return "", err
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
	createResult, err := CreateSend(c, sourceAddress, destinationAddress, asset, quantity, sourceAddressPubKey)
	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Err in CreateSend(): %s", err.Error())
		database.UpdatePaymentWithErrorByPaymentId(c, accessKey, paymentId, err.Error())
		return "", err
	}

	log.FluentfContext(consts.LOGINFO, c, "Created send of %d %s to %s: %s", quantity, asset, destinationAddress, createResult)

	// Sign the transactions
	signed, err := SignRawTransaction(c, passphrase, createResult)
	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Err in SignRawTransaction(): %s\n", err.Error())
		database.UpdatePaymentWithErrorByPaymentId(c, accessKey, paymentId, err.Error())
		return "", err
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
			database.UpdatePaymentWithErrorByPaymentId(c, accessKey, paymentId, err.Error())
			return "", err
		}

		txId = txIdSignedTx
	} else {
		txId = "success"
	}

	database.UpdatePaymentCompleteByPaymentId(c, accessKey, paymentId, txId)

	log.FluentfContext(consts.LOGINFO, c, "Complete.")

	return txId, nil
}

// Concurrency safe to create and send transactions from a single address.
func DelegatedCreateIssuance(c context.Context, accessKey string, passphrase string, sourceAddress string, assetId string, asset string, quantity uint64, divisible bool) (string, error) {
	if isInit == false {
		Init()
	}

	// Copy same context values to local variables which are often accessed
	env := c.Value(consts.EnvKey).(string)

	// Write the asset with the generated asset id to the database
	go database.InsertAsset(accessKey, assetId, sourceAddress, "TBA", asset, quantity, divisible, "valid")

	sourceAddressPubKey, err := counterpartycrypto.GetPublicKey(passphrase, sourceAddress)
	if err != nil {
		log.Printf("Error: %s\n", err)
		return "", err
	}

	// Mutex lock this address
	counterparty_Mutexes.Lock()
	log.Println("Locked the map") // The map of mutexes must be locked before we modify the mutexes stored in the map

	// If an entry doesn't currently exist in the map for that address
	if counterparty_Mutexes.m[sourceAddress] == nil {
		log.Printf("Created new entry in map for %s\n", sourceAddress)
		counterparty_Mutexes.m[sourceAddress] = new(sync.Mutex)
	}

	counterparty_Mutexes.m[sourceAddress].Lock()
	log.Printf("Locked: %s\n", sourceAddress)

	// We must sleep for at least the time it takes for any transactions to propagate through to the counterparty mempool
	log.Println("Sleeping")
	time.Sleep(time.Duration(counterparty_BackEndPollRate+3000) * time.Millisecond)

	defer counterparty_Mutexes.Unlock()
	defer counterparty_Mutexes.m[sourceAddress].Unlock()

	log.Println("Composing the CreateNumericIssuance transaction")
	// Create the issuance
	randomAssetName, createResult, err := CreateNumericIssuance(c, sourceAddress, asset, quantity, divisible, sourceAddressPubKey)
	if err != nil {
		log.Printf("Err in CreateIssuance(): %s\n", err.Error())
		database.UpdateAssetWithErrorByAssetId(c, accessKey, assetId, err.Error())
		return "", err
	}

	log.Printf("Created issuance of %d %s (%s) at %s: %s\n", quantity, asset, randomAssetName, sourceAddress, createResult)
	database.UpdateAssetNameByAssetId(c, accessKey, assetId, randomAssetName)

	// Sign the transactions
	signed, err := SignRawTransaction(c, passphrase, createResult)
	if err != nil {
		log.Printf("Err in SignRawTransaction(): %s\n", err.Error())
		database.UpdateAssetWithErrorByAssetId(c, accessKey, assetId, err.Error())
		return "", err
	}

	log.Printf("Signed tx: %s\n", signed)

	//	 Transmit the transaction if not in dev, otherwise stub out the return
	var txId string
	if env != "dev" {
		txIdSignedTx, err := bitcoinapi.SendRawTransaction(signed)
		if err != nil {
			log.FluentfContext(consts.LOGERROR, c, err.Error())
			database.UpdateAssetWithErrorByAssetId(c, accessKey, assetId, err.Error())
			return "", err
		}

		txId = txIdSignedTx
	} else {
		txId = "success"
	}

	database.UpdateAssetCompleteByAssetId(c, accessKey, assetId, txId)

	return txId, nil
}

// Concurrency safe to create and send transactions from a single address.
func DelegatedCreateDividend(c context.Context, accessKey string, passphrase string, dividendId string, sourceAddress string, asset string, dividendAsset string, quantityPerUnit uint64) (string, error) {
	if isInit == false {
		Init()
	}

	// Copy same context values to local variables which are often accessed
	env := c.Value(consts.EnvKey).(string)

	// Write the dividend with the generated dividend id to the database
	go database.InsertDividend(accessKey, dividendId, sourceAddress, asset, dividendAsset, quantityPerUnit, "valid")

	sourceAddressPubKey, err := counterpartycrypto.GetPublicKey(passphrase, sourceAddress)
	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Error: %s\n", err)
		return "", err
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

	// Create the dividend
	createResult, err := CreateDividend(c, sourceAddress, asset, dividendAsset, quantityPerUnit, sourceAddressPubKey)
	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, err.Error())
		database.UpdateDividendWithErrorByDividendId(c, accessKey, dividendId, err.Error())
		return "", err
	}

	log.FluentfContext(consts.LOGINFO, c, "Created dividend of %d %s for each %s from address %s: %s\n", quantityPerUnit, dividendAsset, asset, sourceAddress, createResult)

	// Sign the transactions
	signed, err := SignRawTransaction(c, passphrase, createResult)
	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, err.Error())
		database.UpdateDividendWithErrorByDividendId(c, accessKey, dividendId, err.Error())
		return "", err
	}

	log.FluentfContext(consts.LOGINFO, c, "Signed tx: %s\n", signed)

	//	 Transmit the transaction if not in dev, otherwise stub out the return
	var txId string
	if env != "dev" {
		txIdSignedTx, err := bitcoinapi.SendRawTransaction(signed)
		if err != nil {
			log.FluentfContext(consts.LOGERROR, c, err.Error())
			database.UpdateDividendWithErrorByDividendId(c, accessKey, dividendId, err.Error())
			return "", err
		}

		txId = txIdSignedTx
	} else {
		txId = "success"
	}

	database.UpdateDividendCompleteByDividendId(c, accessKey, dividendId, txId)

	return txId, nil
}

func GetRunningInfo(c context.Context) (RunningInfo, error) {
	var payload payloadGetRunningInfo
	var result ResultGetRunningInfo

	if isInit == false {
		Init()
	}

	payload.Method = "get_running_info"
	payload.Jsonrpc = "2.0"
	payload.Id = generateId(c)

	payloadJsonBytes, err := json.Marshal(payload)

	//	log.Println(string(payloadJsonBytes))

	if err != nil {
		return result.Result, err
	}

	responseData, _, err := postAPI(c, payloadJsonBytes)
	if err != nil {
		return result.Result, err
	}

	//	log.Println(string(responseData))

	if err := json.Unmarshal(responseData, &result); err != nil {
		log.Println(string(responseData))
		return result.Result, errors.New("Unable to unmarshal responseData")
	}

	return result.Result, nil
}

// Concurrency safe to create and send transactions from a single address.
func DelegatedActivateAddress(c context.Context, addressToActivate string, amount uint64, activationId string) (string, error) {
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

	// Calculate the quantity of BTC to send by the amount specified
	// For Counterparty: each transaction = dust_size + miners_fee
	var quantity uint64
	var asset string
	if blockchainId == consts.CounterpartyBlockchainId && env != "dev" {
		quantity = (Counterparty_DefaultDustSize + Counterparty_DefaultTxFee) * amount
		asset = "BTC"
	} else if blockchainId == consts.CounterpartyBlockchainId && env == "dev" {
		quantity = (Counterparty_DefaultDustSize + Counterparty_DefaultTestingTxFee) * amount
		asset = "BTC"
	} else {
		log.FluentfContext(consts.LOGERROR, c, "Unsupported blockchain: %s\n", blockchainId)
		return "", errors.New("Unsupported blockchain")
	}

	// Write the dividend with the generated dividend id to the database
	go database.InsertActivation(c, accessKey, activationId, blockchainId, sourceAddress, amount)

	sourceAddressPubKey, err := counterpartycrypto.GetPublicKey(wallets[randomNumber].Passphrase, sourceAddress)
	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Error: %s\n", err)
		return "", err
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
	createResult, err := CreateSend(c, sourceAddress, addressToActivate, asset, quantity, sourceAddressPubKey)
	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, err.Error())
		database.UpdatePaymentWithErrorByPaymentId(c, accessKey, activationId, err.Error())
		return "", err
	}

	log.FluentfContext(consts.LOGINFO, c, "Created send of %d %s to %s: %s", quantity, asset, addressToActivate, createResult)

	// Sign the transactions
	signed, err := SignRawTransaction(c, wallets[randomNumber].Passphrase, createResult)
	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, err.Error())
		database.UpdatePaymentWithErrorByPaymentId(c, accessKey, activationId, err.Error())
		return "", err
	}

	log.FluentfContext(consts.LOGINFO, c, "Signed tx: %s\n", signed)

	//	 Transmit the transaction if not in dev, otherwise stub out the return
	var txId string
	if env != "dev" {
		txIdSignedTx, err := bitcoinapi.SendRawTransaction(signed)
		if err != nil {
			log.FluentfContext(consts.LOGERROR, c, err.Error())
			database.UpdatePaymentWithErrorByPaymentId(c, accessKey, activationId, err.Error())
			return "", err
		}

		txId = txIdSignedTx
	} else {
		txId = "success"
	}

	database.UpdatePaymentCompleteByPaymentId(c, accessKey, activationId, txId)

	return txId, nil
}
