package main

import (
	"io/ioutil"
	//	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/vennd/enu/consts"
	"github.com/vennd/enu/counterpartyapi"
	"github.com/vennd/enu/counterpartycrypto"
	"github.com/vennd/enu/log"
)

var sourceFile = "pingcounterpartyd.go"

var passphrase string = "attention stranger fate plain huge poetry view precious drug world try age"
var sendAddress string = "1CipmbDRHn89cgqs6XbjswkrDxvCKA8Tfb"
var destinationAddress string = "1HpkZBjNFRFagyj6Q2adRSagkfNDERZhg1"

// This application is to be used for monitoring purposes of counterpartyd
//
// Calls get_running_info() to retrieve the last block that counterpartyd processed
// Compares the last block counterpartyd processed against blockchain.info
// If the difference is < 5 blocks pass, otherwise fail test
// Attempt to construct a send with counterparty. The send construction should complete within 10 seconds, or consider bitcoind and counterpartyd stalled
//
// Returns -1 if there was a problem reading the counterpartyd last processed block
// Returns -2 if there was a problem reading from blockchain.info
// Returns -3 if there is a different > 5 blocks between counterpartyd last processed block and blockchain.info
// Returns -4 if the construction of a send with counterpartyd didn't complete successfully (or within 10 seconds)
func main() {
	var result1, result2 uint64
	var result3 string

	// First check the internal block height via our API
	c1 := make(chan uint64, 1)
	go func() {
		result, err := counterpartyapi.GetRunningInfo()

		if err != nil {
			log.Fluentf(consts.LOGINFO, "Error retrieving our block height")
			log.Fluentf(consts.LOGINFO, err.Error())
			os.Exit(-1)
		}
		c1 <- result.LastBlock.BlockIndex
	}()

	select {
	case result1 = <-c1:
		log.Fluentf(consts.LOGINFO, "Counterpartyd last processed block: %d\n", result1)
	case <-time.After(time.Second * 10):
		log.Fluentf(consts.LOGINFO, "Timeout when retrieving last processed counterpartyd block")
		os.Exit(-1)
	}

	// Then check the block height from blockchain.info
	c2 := make(chan uint64, 1)
	go func() {
		request, err2 := http.Get("https://blockchain.info/q/getblockcount")

		defer request.Body.Close()
		response, err := ioutil.ReadAll(request.Body)

		if err != nil {
			log.Fluentf(consts.LOGINFO, "Error reading from blockchain.info")
			log.Fluentf(consts.LOGINFO, err.Error())
			os.Exit(-2)
		}

		result, err2 := strconv.ParseUint(string(response), 10, 64)

		if err2 != nil {
			log.Fluentf(consts.LOGINFO, "Error reading from blockchain.info")
			log.Fluentf(consts.LOGINFO, err2.Error())
			os.Exit(-2)
		}

		c2 <- result
	}()

	select {
	case result2 = <-c2:
		log.Fluentf(consts.LOGINFO, "Blockchain.info block height: %d\n", result2)
	case <-time.After(time.Second * 10):
		log.Fluentf(consts.LOGINFO, "Timeout when retrieving blockchain.info block height")
		os.Exit(-2)
	}

	var difference uint64
	if (result1 - result2) < 0 {
		difference = result2 - result1
	} else {
		difference = result1 - result2
	}
	// Check the difference < 5
	if difference > 5 {
		os.Exit(-3)
	}

	// Attempt to create a send
	c3 := make(chan string, 1)
	go func() {
		pubKey, err := counterpartycrypto.GetPublicKey(passphrase, sendAddress)

		if err != nil {
			log.Println("Error getting pubkey")
			log.Println(err.Error())
		}

		resultCreateSend, err2 := counterpartyapi.CreateSend(sendAddress, destinationAddress, "SHIMA", 1000, pubKey)

		if err2 != nil {
			log.Println("Error creating counterparty send")
			log.Println(err2.Error())
		}

		c3 <- resultCreateSend
	}()

	select {
	case result3 = <-c3:
		log.Fluentf(sourceFile, "Successfully created transaction: %s\n", result3)
	case <-time.After(time.Second * 30):
		log.Fluentf(sourceFile, "Timeout when creating counterparty send")
		os.Exit(-4)
	}

	return
}
