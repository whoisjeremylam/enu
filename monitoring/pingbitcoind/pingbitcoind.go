package main

import (
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/vennd/enu/bitcoinapi"
	"github.com/vennd/enu/log"
)

const sourceFilename = "pingbitcoind.go"

// This application is to be used for monitoring purposes of bitcoind or btcd.
//
// Calls GetBlockCount to check to see if bitcoind is alive
// Compares the blockheight against blockchain.info
// If the difference is < 5 blocks pass, otherwise fail test
//
// Returns -1 if there was a problem reading internal block height
// Returns -2 if there was a problem reading from blockchain.info
// Returns -3 if there is a different > 5 blocks between internal and blockchain.info
func main() {
	var result1, result2 int64

	// Check if path to config file has been specified and file exists
	// then attempt to init with the file
	if len(os.Args) > 1 {
		if _, err := os.Stat(os.Args[1]); err == nil {
			bitcoinapi.InitWithConfigPath(os.Args[1])
		}
	}

	// First check the internal block height via our API
	c1 := make(chan int64, 1)
	go func() {
		ourBlockHeight, err := bitcoinapi.GetBlockCount()

		if err != nil {
			log.Fluentln(sourceFilename, "Error retrieving our block height")
			log.Fluentln(sourceFilename, err.Error())
			os.Exit(-1)
		}
		c1 <- ourBlockHeight
	}()

	select {
	case result1 = <-c1:
		log.Fluentf(sourceFilename, "Our block height: %d\n", result1)
	case <-time.After(time.Second * 10):
		log.Fluentln(sourceFilename, "Timeout when retrieving our block height")
		os.Exit(-1)
	}

	// Then check the block height from blockchain.info
	c2 := make(chan int64, 1)
	go func() {
		request, err2 := http.Get("https://blockchain.info/q/getblockcount")

		defer request.Body.Close()
		response, err := ioutil.ReadAll(request.Body)

		if err != nil {
			log.Fluentln(sourceFilename, "Error reading from blockchain.info")
			log.Fluentln(sourceFilename, err.Error())
			os.Exit(-2)
		}

		result, err2 := strconv.ParseInt(string(response), 10, 64)

		if err2 != nil {
			log.Fluentln(sourceFilename, "Error reading from blockchain.info")
			log.Fluentln(sourceFilename, err2.Error())
			os.Exit(-2)
		}

		c2 <- result
	}()

	select {
	case result2 = <-c2:
		log.Fluentf(sourceFilename, "Blockchain.info block height: %d\n", result2)
	case <-time.After(time.Second * 10):
		log.Fluentf(sourceFilename, "Timeout when retrieving blockchain.info block height")
		os.Exit(-2)
	}

	var difference int64
	if (result1 - result2) < 0 {
		difference = result2 - result1
	} else {
		difference = result1 - result2
	}
	// Check the difference < 5
	if difference > 5 {
		os.Exit(-3)
	}

	return
}
