// Signs a raw transaction with debugging
package main

import (
	"strconv"
	//	"fmt"
	//	"os"
	//	"reflect"
	"bytes"
	"encoding/hex"
	"errors"

	"github.com/whoisjeremylam/enu/consts"
	"github.com/whoisjeremylam/enu/counterpartycrypto"
	"github.com/whoisjeremylam/enu/enulib"
	"github.com/whoisjeremylam/enu/log"

	"github.com/whoisjeremylam/enu/internal/github.com/btcsuite/btcd/btcec"
	"github.com/whoisjeremylam/enu/internal/github.com/btcsuite/btcd/chaincfg"
	"github.com/whoisjeremylam/enu/internal/github.com/btcsuite/btcd/txscript"
	"github.com/whoisjeremylam/enu/internal/github.com/btcsuite/btcd/wire"
	"github.com/whoisjeremylam/enu/internal/github.com/btcsuite/btcutil"

	//	"github.com/btcsuite/btcd/btcec"
	//	"github.com/btcsuite/btcd/chaincfg"
	//	"github.com/btcsuite/btcd/txscript"
	//	"github.com/btcsuite/btcd/wire"
	//	"github.com/btcsuite/btcutil"

	"github.com/whoisjeremylam/enu/internal/golang.org/x/net/context"
)

var passphrase string = "confuse patient join toss stolen hurry pencil grew toward handle remember mirror"
var unsignedRawTransaction string = "0100000002e4679db827c70493aad18d33077669f6e3f5b1e494b59423f5c77fef0cdf9b85000000001976a914f881641aa74d4c2d0953cc332b6d9264b033abea88acffffffff03679d7f80b91a6a46916056711fa05da9b9dc8e5751d56aa38beba57c95fd1f000000001976a914f881641aa74d4c2d0953cc332b6d9264b033abea88acffffffff0336150000000000001976a914599a578f3650398cce1156c39bbf073161dd737b88ac00000000000000001e6a1cf01b4b5792903450e169daa92c385a6385d218802e74fa74991750cb48641700000000001976a914f881641aa74d4c2d0953cc332b6d9264b033abea88ac00000000"

func signRawTransactionOld(c context.Context, passphrase string, rawTxHexString string) (string, error) {
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
	redeemTx := wire.NewMsgTx() // Create a new transaction and copy the details from the tx that was serialised. For some reason BTCD can't sign in place transactions
	//	log.Printf("MsgTx: %+v", msgTx)
	//	log.Printf("Number of txes in: %d\n", len(msgTx.TxIn))
	for i := 0; i <= len(msgTx.TxIn)-1; i++ {
		//		log.Printf("MsgTx.TxIn[%d]:\n", i)
		//		log.Printf("TxIn[%d].PreviousOutPoint.Hash: %s\n", i, msgTx.TxIn[i].PreviousOutPoint.Hash)
		//		log.Printf("TxIn[%d].PreviousOutPoint.Index: %d\n", i, msgTx.TxIn[i].PreviousOutPoint.Index)
		//		log.Printf("TxIn[%d].SignatureScript: %s\n", i, hex.EncodeToString(msgTx.TxIn[i].SignatureScript))
		script := msgTx.TxIn[i].SignatureScript

		// Following block is for debugging only
		//		disasm, err := txscript.DisasmString(script)
		//		if err != nil {
		//			return "", err
		//		}
		//		log.Printf("TxIn[%d] Script Disassembly: %s", i, disasm)

		// Extract and print details from the script.
		// next line is for debugging only
		//		scriptClass, addresses, reqSigs, err := txscript.ExtractPkScriptAddrs(script, &chaincfg.MainNetParams)
		scriptClass, _, _, err := txscript.ExtractPkScriptAddrs(script, &chaincfg.MainNetParams)
		if err != nil {
			log.FluentfContext(consts.LOGERROR, c, "Error in ExtractPkScriptAddrs(): %s", err.Error())
			return "", err
		}

		// This function only supports pubkeyhash signing at this time (ie not multisig or P2SH)
		//		log.Printf("TxIn[%d] Script Class: %s\n", i, scriptClass)
		if scriptClass.String() != "pubkeyhash" {
			return "", errors.New("Counterparty_SignRawTransaction() currently only supports pubkeyhash script signing. However, the script type in the TX to sign was: " + scriptClass.String())
		}

		//		log.Printf("TxIn[%d] Addresses: %s\n", i, addresses)
		//		log.Printf("TxIn[%d] Required Signatures: %d\n", i, reqSigs)

		// Build txIn for new redeeming transaction
		prevOut := wire.NewOutPoint(&msgTx.TxIn[i].PreviousOutPoint.Hash, msgTx.TxIn[i].PreviousOutPoint.Index)
		txIn := wire.NewTxIn(prevOut, nil)
		redeemTx.AddTxIn(txIn)
	}

	// Copy TxOuts from serialised tx
	for _, txOut := range msgTx.TxOut {
		out := txOut
		redeemTx.AddTxOut(out)
	}

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
	for i, _ := range redeemTx.TxIn {
		// Get the sigscript
		// Notice that the script database parameter is nil here since it isn't
		// used.  It must be specified when pay-to-script-hash transactions are
		// being signed.
		sigScript, err := txscript.SignTxOutput(&chaincfg.MainNetParams, redeemTx, i, msgTx.TxIn[i].SignatureScript, txscript.SigHashAll, txscript.KeyClosure(lookupKey), nil, nil)

		if err != nil {
			return "", err
		}

		// Copy the signed sigscript into the redeeming tx
		redeemTx.TxIn[i].SignatureScript = sigScript
		//		log.Println(hex.EncodeToString(sigScript))
	}

	// Prove that the transaction has been validly signed by executing the
	// script pair.
	//	log.Println("Checking signature(s)")
	flags := txscript.ScriptBip16 | txscript.ScriptVerifyDERSignatures | txscript.ScriptStrictMultiSig | txscript.ScriptDiscourageUpgradableNops | txscript.ScriptVerifyLowS | txscript.ScriptVerifyCleanStack | txscript.ScriptVerifyMinimalData | txscript.ScriptVerifySigPushOnly | txscript.ScriptVerifyStrictEncoding
	var buildError string
	for i, _ := range redeemTx.TxIn {
		vm, err := txscript.NewEngine(msgTx.TxIn[i].SignatureScript, redeemTx, i, flags)
		if err != nil {
			buildError += "NewEngine() error: " + err.Error() + ","
		}

		if err := vm.Execute(); err != nil {
			buildError += "TxIn[" + strconv.Itoa(i) + "]: " + err.Error() + ", "
		} else {
			// Signature verified
			//			log.Printf("TxIn[%d] ok!\n", i)
		}
	}
	if len(buildError) > 0 {
		return "", errors.New(buildError)
	}
	//	log.Println("Transaction successfully signed")

	// Encode the struct into BTC bytes wire format
	var byteBuffer bytes.Buffer
	encodeError := redeemTx.BtcEncode(&byteBuffer, wire.ProtocolVersion)
	if encodeError != nil {
		return "", err
	}

	// Encode bytes to hex string
	payloadBytes := byteBuffer.Bytes()
	payloadHexString := hex.EncodeToString(payloadBytes)
	//	log.Printf("Signed and encoded transaction: %s\n", payloadHexString)

	return payloadHexString, nil
}

func main() {
	c := context.TODO()
	c = context.WithValue(c, consts.RequestIdKey, "test"+enulib.GenerateRequestId())

	newSignedTx, err := signRawTransaction(c, passphrase, unsignedRawTransaction)
	if err != nil {
		log.Println(err.Error())
	}

	log.Println(newSignedTx)

	oldSignedTx, err := signRawTransactionOld(c, passphrase, unsignedRawTransaction)
	log.Println(oldSignedTx)
}
