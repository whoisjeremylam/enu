package main

import (
	"github.com/vennd/enu/internal/bitbucket.org/dchapes/ripple/crypto/rkey"
	"fmt"
	"log"
	"math/big"

	"github.com/vennd/enu/ripplecrypto"
)

func main() {
	secret := "ssDjVhSj8P3GguMyBdyYmgnsBwzMb"
	//	seedInt := "50060557627507529689996917628571969848"

	// Generate hex seed from ripple secret
	s, _ := rkey.NewFamilySeed(secret)
	seedHex := fmt.Sprintf("%x", s.Seed)
	log.Printf("From secret: %s, hex seed: %x", secret, s.Seed)

	// And go back again
	ints2 := new(big.Int)
	ints2.SetString(seedHex, 16)
	s2, _ := rkey.NewSeed(ints2)
	address, _ := s2.MarshalText()
	log.Printf("From seedHex: %s, address: %s", seedHex, string(address))

	wallet, _ := ripplecrypto.CreateWallet(20)
	log.Printf("%+#v", wallet)
}
