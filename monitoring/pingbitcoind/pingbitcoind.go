package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/vennd/enu/bitcoinapi"
)

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

	// First check the internal block height via our API
	c1 := make(chan int64, 1)
	go func() {
		ourBlockHeight, err := bitcoinapi.GetBlockCount()
		
		if err != nil {
			log.Println("Error retrieving our block height")
			os.Exit(-1)
		}
		c1 <- ourBlockHeight
	}()

	select {
	case result1 = <-c1:
		log.Printf("Our block height: %d\n", result1)
	case <-time.After(time.Second * 10):
		log.Println("Timeout when retrieving our block height")
		os.Exit(-1)
	}

	// Then check the block height from blockchain.info
	c2 := make(chan int64, 1)
	go func() {
		request, err2 := http.Get("https://blockchain.info/q/getblockcount")

		defer request.Body.Close()
		response, err := ioutil.ReadAll(request.Body)

		if err != nil {
			log.Println("Error reading from blockchain.info")
			os.Exit(-2)
		}

		result, err2 := strconv.ParseInt(string(response), 10, 64)

		if err2 != nil {
			log.Println("Error reading from blockchain.info")
			os.Exit(-2)
		}

		c2 <- result
	}()

	select {
	case result2 = <-c2:
		log.Printf("Blockchain.info block height: %d\n", result2)
	case <-time.After(time.Second * 10):
		log.Println("Timeout when retrieving blockchain.info block height")
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
