package bitcoinapi

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/vennd/enu/internal/github.com/btcsuite/btcrpcclient"
	"github.com/vennd/enu/internal/github.com/btcsuite/btcutil"
)

// Globals
var config btcrpcclient.ConnConfig
var isInit bool = false // set to true only after the init sequence is complete

// Initialises global variables and database connection for all handlers
func Init() {
	var configFilePath string

	if isInit == true {
		return
	}

	if _, err := os.Stat("./enuapi.json"); err == nil {
		log.Println("Found and using configuration file ./enuapi.json")
		configFilePath = "./enuapi.json"
	} else {
		if _, err := os.Stat(os.Getenv("GOPATH") + "/bin/enuapi.json"); err == nil {
			configFilePath = os.Getenv("GOPATH") + "/bin/enuapi.json"
			log.Printf("Found and using configuration file from GOPATH: %s\n", configFilePath)

		} else {
					if _, err := os.Stat(os.Getenv("GOPATH") + "/src/github.com/vennd/enu/enuapi.json"); err == nil {
			configFilePath = os.Getenv("GOPATH") + "/src/github.com/vennd/enu/enuapi.json"
			log.Printf("Found and using configuration file from GOPATH: %s\n", configFilePath)

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
	log.Printf("Reading %s\n", configFilePath)
	file, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		log.Println("Unable to read configuration file enuapi.json")
		log.Fatalln(err)
	}

	err = json.Unmarshal(file, &configuration)

	if err != nil {
		log.Println("Unable to parse enuapi.json")
		log.Fatalln(err)
	}

	m := configuration.(map[string]interface{})

	// Bitcoin API parameters
	config.Host = m["btchost"].(string)     // Hostname:port for Bitcoin Core or BTCD
	config.User = m["btcuser"].(string)     // Basic authentication user name
	config.Pass = m["btcpassword"].(string) // Basic authentication password
	config.HTTPPostMode = true              // Bitcoin core only supports HTTP POST mode
	config.DisableTLS = true                // Bitcoin core does not provide TLS by default

	isInit = true
}

// Thanks to https://raw.githubusercontent.com/btcsuite/btcrpcclient/master/examples/bitcoincorehttp/main.go
func GetBlockCount() (int64, error) {
	if isInit == false {
		Init()
	}

	// Notice the notification parameter is nil since notifications are
	// not supported in HTTP POST mode.
	client, err := btcrpcclient.New(&config, nil)
	if err != nil {
		log.Println(err)
		return 0, err
	}
	defer client.Shutdown()

	// Get the current block count.
	blockCount, err := client.GetBlockCount()
	if err != nil {
		log.Println(err)
		return 0, err
	}

	return blockCount, nil
}

func GetNewAddress() (string, error) {
	if isInit == false {
		Init()
	}

	// Notice the notification parameter is nil since notifications are
	// not supported in HTTP POST mode.
	client, err := btcrpcclient.New(&config, nil)
	if err != nil {
		return "", err
	}
	defer client.Shutdown()

	// Get a new BTC address.
	address, err := client.GetNewAddress("")
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s", address), nil
}

// Transmits to the bitcoin network the raw transaction as provided.
// The transaction should be encoded as a hex string, as per the original Bitcoin RPC JSON API.
// The TxId of the transaction is returned if successfully transmitted.
func SendRawTransaction(txHexString string) (string, error) {
	if isInit == false {
		Init()
	}

	// Convert the hex string to a byte array
	txBytes, err := hex.DecodeString(txHexString)
	if err != nil {
		log.Println(err)
		return "", err
	}

	// Deserialise the transaction
	tx, err := btcutil.NewTxFromBytes(txBytes)
	if err != nil {
		log.Println(err)
		return "", err
	}

	msgTx := tx.MsgTx()

	// Notice the notification parameter is nil since notifications are
	// not supported in HTTP POST mode.
	client, err := btcrpcclient.New(&config, nil)
	if err != nil {
		log.Println(err)
		return "", err
	}
	defer client.Shutdown()

	// Send the tx
	result, err := client.SendRawTransaction(msgTx, true)
	if err != nil {
		log.Println(err)
		return "", err
	}

	return fmt.Sprintf("%s", result.String()), nil
}

func GetBalance(account string) (float64, error) {
	if isInit == false {
		Init()
	}

	// Notice the notification parameter is nil since notifications are
	// not supported in HTTP POST mode.
	client, err := btcrpcclient.New(&config, nil)
	if err != nil {
		log.Println(err)
		return 0, err
	}
	defer client.Shutdown()

	// Get the current balance
	amount, err := client.GetBalance(account)
	if err != nil {
		log.Println(err)
		return 0, err
	}
	var unit btcutil.AmountUnit = 0
	return amount.ToUnit(unit), nil
}
