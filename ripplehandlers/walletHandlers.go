package ripplehandlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/vennd/enu/internal/github.com/gorilla/mux"
	"github.com/vennd/enu/internal/golang.org/x/net/context"

	"github.com/vennd/enu/consts"
	"github.com/vennd/enu/database"
	"github.com/vennd/enu/enulib"
	"github.com/vennd/enu/handlers"
	"github.com/vennd/enu/internal/github.com/vennd/mneumonic"
	"github.com/vennd/enu/log"
	"github.com/vennd/enu/rippleapi"
	"github.com/vennd/enu/ripplecrypto"
)

var ripple_BackEndPollRate = 3000

var ripple_Mutexes = struct {
	sync.RWMutex
	m map[string]*sync.Mutex
}{m: make(map[string]*sync.Mutex)}

func WalletCreate(c context.Context, w http.ResponseWriter, r *http.Request, m map[string]interface{}) *enulib.AppError {
	var walletModel enulib.Wallet
	requestId := c.Value(consts.RequestIdKey).(string)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	// Create the wallet
	wallet, errCode, err := rippleapi.CreateWallet(c)
	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Error in CreateWallet(): %s", err.Error())
		handlers.ReturnServerError(c, w, errCode, err)

		return nil
	}
	log.FluentfContext(consts.LOGINFO, c, "Created a new wallet with address: %s for access key: %s\n (requestID: %s)", wallet.AccountId, c.Value(consts.AccessKeyKey).(string), requestId)

	// Return the wallet
	walletModel.RequestId = requestId
	walletModel.Addresses = append(walletModel.Addresses, wallet.AccountId) // The address is what Ripple calls the account Id
	walletModel.BlockchainId = consts.RippleBlockchainId
	walletModel.HexSeed = wallet.MasterSeedHex
	walletModel.KeyType = wallet.KeyType
	walletModel.PublicKey = wallet.PublicKey
	walletModel.PublicKeyHex = wallet.PublicKeyHex
	walletModel.MasterSeed = wallet.MasterSeed

	mn := mneumonic.FromHexstring(wallet.MasterSeedHex)
	walletModel.Passphrase = strings.Join(mn.ToWords(), " ") // The hex seed for Ripple wallets can be translated to the same mneumonic that generates counterparty wallets

	w.WriteHeader(http.StatusCreated)
	if err = json.NewEncoder(w).Encode(walletModel); err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Error in Encode(): %s", err.Error())
		handlers.ReturnServerError(c, w, consts.GenericErrors.GeneralError.Code, errors.New(consts.GenericErrors.GeneralError.Description))

		return nil
	}

	return nil
}

func WalletSend(c context.Context, w http.ResponseWriter, r *http.Request, m map[string]interface{}) *enulib.AppError {

	var walletPayment enulib.WalletPayment
	var paymentTag string
	var issuer string

	requestId := c.Value(consts.RequestIdKey).(string)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	walletPayment.RequestId = requestId

	// Add to the context the RequestType
	c = context.WithValue(c, consts.RequestTypeKey, "walletPayment")

	passphrase := m["passphrase"].(string)
	sourceAddress := m["sourceAddress"].(string)
	destinationAddress := m["destinationAddress"].(string)
	asset := m["asset"].(string)
	quantity := uint64(m["quantity"].(float64))

	if m["paymentTag"] != nil {
		paymentTag = m["paymentTag"].(string)
	}

	if m["issuer"] != nil {
		issuer = m["issuer"].(string)
	}

	// If a custom asset is specified, then an issuer must be provided
	if strings.ToUpper(asset) != "XRP" && issuer == "" {
		log.FluentfContext(consts.LOGERROR, c, consts.RippleErrors.IssuerMustBeGiven.Description)
		handlers.ReturnBadRequest(c, w, consts.RippleErrors.IssuerMustBeGiven.Code, errors.New(consts.RippleErrors.IssuerMustBeGiven.Description))
		return nil
	}

	log.FluentfContext(consts.LOGINFO, c, "WalletSend: received request sourceAddress: %s, destinationAddress: %s, asset: %s, issuer: %s, quantity: %d, paymentTag: %s from accessKey: %s\n", sourceAddress, destinationAddress, asset, issuer, quantity, c.Value(consts.AccessKeyKey).(string), paymentTag)
	// Generate a paymentId
	paymentId := enulib.GeneratePaymentId()

	log.FluentfContext(consts.LOGINFO, c, "Generated paymentId: %s", paymentId)

	// Return to the client the walletPayment containing requestId and paymentId and unblock the client
	walletPayment.PaymentId = paymentId
	walletPayment.Asset = asset
	walletPayment.SourceAddress = sourceAddress
	walletPayment.DestinationAddress = destinationAddress
	walletPayment.Quantity = quantity
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(walletPayment); err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Error in Encode(): %s", err.Error())
		handlers.ReturnServerError(c, w, consts.GenericErrors.GeneralError.Code, errors.New(consts.GenericErrors.GeneralError.Description))

		return nil
	}

	//	txHash, errCode, err := rippleapi.SendPayment(c, sourceAddress, destinationAddress, amount, asset, issuer, secret)
	go delegatedSend(c, c.Value(consts.AccessKeyKey).(string), passphrase, sourceAddress, destinationAddress, asset, issuer, quantity, paymentId, paymentTag)

	return nil
}

// Concurrency safe to create and send transactions from a single address.
func delegatedSend(c context.Context, accessKey string, passphrase string, sourceAddress string, destinationAddress string, asset string, issuer string, quantity uint64, paymentId string, paymentTag string) (string, int64, error) {
	// Copy same context values to local variables which are often accessed
	env := c.Value(consts.EnvKey).(string)

	// Write the payment with the generated payment id to the database
	defaultFee, err := strconv.ParseUint(rippleapi.DefaultFee, 10, 64)
	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Error in converting ripple fee: %s", err.Error())
	}
	go database.InsertPayment(c, accessKey, 0, c.Value(consts.BlockchainIdKey).(string), paymentId, sourceAddress, destinationAddress, asset, issuer, quantity, "valid", 0, defaultFee, paymentTag)

	// Mutex lock this address
	ripple_Mutexes.Lock()
	log.FluentfContext(consts.LOGINFO, c, "Locked the map") // The map of mutexes must be locked before we modify the mutexes stored in the map

	// If an entry doesn't currently exist in the map for that address
	if ripple_Mutexes.m[sourceAddress] == nil {
		log.FluentfContext(consts.LOGINFO, c, "Created new entry in map for %s", sourceAddress)
		ripple_Mutexes.m[sourceAddress] = new(sync.Mutex)
	}

	ripple_Mutexes.m[sourceAddress].Lock()
	log.FluentfContext(consts.LOGINFO, c, "Locked: %s\n", sourceAddress)

	defer ripple_Mutexes.Unlock()
	defer ripple_Mutexes.m[sourceAddress].Unlock()

	// We must sleep for at least the time it takes for most transactions to enter a ledger
	log.FluentfContext(consts.LOGINFO, c, "Sleeping %d milliseconds", ripple_BackEndPollRate+5000)
	time.Sleep(time.Duration(ripple_BackEndPollRate+5000) * time.Millisecond)

	log.FluentfContext(consts.LOGINFO, c, "Sleep complete")

	// Convert int to the ripple amount
	amount, err := rippleapi.Uint64ToAmount(quantity)
	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Error in Uint64ToAmount(): %s", err.Error())
		database.UpdatePaymentWithErrorByPaymentId(c, accessKey, paymentId, consts.GenericErrors.GeneralError.Code, consts.GenericErrors.GeneralError.Description)

		return "", consts.GenericErrors.GeneralError.Code, errors.New(consts.GenericErrors.GeneralError.Description)
	}

	// Convert asset name to ripple currency name
	currency, err := rippleapi.ToCurrency(asset)
	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Error in Uint64ToAmount(): %s", err.Error())
		database.UpdatePaymentWithErrorByPaymentId(c, accessKey, paymentId, consts.GenericErrors.GeneralError.Code, consts.GenericErrors.GeneralError.Description)

		return "", consts.GenericErrors.GeneralError.Code, errors.New(consts.GenericErrors.GeneralError.Description)
	}

	// Convert passphrase to ripple secret
	seed := mneumonic.FromWords(strings.Split(passphrase, " "))
	secret, err := ripplecrypto.ToSecret(seed.ToHex())
	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Error in ripplecrypto.ToSecret(): %s", err.Error())
		database.UpdatePaymentWithErrorByPaymentId(c, accessKey, paymentId, consts.GenericErrors.InvalidPassphrase.Code, consts.GenericErrors.InvalidPassphrase.Description)

		return "", consts.GenericErrors.InvalidPassphrase.Code, errors.New(consts.GenericErrors.InvalidPassphrase.Description)
	}

	// Create and sign the transaction
	signedTx, errCode, err := rippleapi.CreatePayment(c, sourceAddress, destinationAddress, amount, currency, issuer, secret)
	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Error in rippleapi.CreatePayment(): %s", err.Error())
		database.UpdatePaymentWithErrorByPaymentId(c, accessKey, paymentId, errCode, err.Error())

		return "", errCode, err
	}

	//	 Submit the transaction if not in dev, otherwise stub out the return
	var txId string
	if env != "dev" {
		txHash, errCode, err := rippleapi.Submit(c, signedTx)
		if err != nil {
			log.FluentfContext(consts.LOGERROR, c, "Error in Submit(): %s", err.Error())
			database.UpdatePaymentWithErrorByPaymentId(c, accessKey, paymentId, consts.RippleErrors.SubmitError.Code, consts.RippleErrors.SubmitError.Description)

			return "", errCode, err
		}

		txId = txHash
	} else {
		log.FluentfContext(consts.LOGINFO, c, "In dev mode, not submitting tx to Ripple network.")
		txId = "success"
	}

	database.UpdatePaymentCompleteByPaymentId(c, accessKey, paymentId, txId)

	log.FluentfContext(consts.LOGINFO, c, "Complete.")

	return txId, 0, nil
}

func ActivateAddress(c context.Context, w http.ResponseWriter, r *http.Request, m map[string]interface{}) *enulib.AppError {
	requestId := c.Value(consts.RequestIdKey).(string)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	// Add to the context the RequestType
	c = context.WithValue(c, consts.RequestTypeKey, "activateaddress")

	vars := mux.Vars(r)
	address := vars["address"]

	if address == "" || len(address) != 34 {
		w.WriteHeader(http.StatusBadRequest)
		returnCode := enulib.ReturnCode{RequestId: c.Value(consts.RequestIdKey).(string), Code: -3, Description: "Incorrect value of address received in the request"}
		if err := json.NewEncoder(w).Encode(returnCode); err != nil {
			log.FluentfContext(consts.LOGERROR, c, "Error in Encode(): %s", err.Error())
			handlers.ReturnServerError(c, w, consts.GenericErrors.GeneralError.Code, errors.New(consts.GenericErrors.GeneralError.Description))

			return nil
		}

		return nil
	}

	// Get the amount from the parameters
	var amount uint64
	if m["amount"] == nil {
		amount = consts.RippleAddressActivationAmount
	} else {
		amount = uint64(m["amount"].(float64))
	}

	// Get the assets to create trust lines for
	var assets []rippleapi.Amount
	if m["assets"] != nil {
		a, ok := m["assets"].([]map[string]interface{})
		if ok {
			for _, b := range a {
				var amount rippleapi.Amount

				amount.Currency = b["currency"].(string)
				amount.Issuer = b["issuer"].(string)

				assets = append(assets, amount)
			}
		}
	}

	var passphrase string
	if m["passphrase"] != nil {
		passphrase = m["passphrase"].(string)
	}

	log.FluentfContext(consts.LOGINFO, c, "ActivateAddress: received request address to activate: %s, number of transactions to activate: %d", address, amount)
	// Generate an activationId
	activationId := enulib.GenerateActivationId()

	log.FluentfContext(consts.LOGINFO, c, "Generated activationId: %s", activationId)

	// Return to the client the activationId and requestId and unblock the client
	var result = map[string]interface{}{
		"address":       address,
		"amount":        amount,
		"assets":        assets,
		"activationId":  activationId,
		"broadcastTxId": "",
		"status":        "valid",
		"errorMessage":  "",
		"requestId":     requestId,
	}

	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(result); err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Error in Encode(): %s", err.Error())
		handlers.ReturnServerError(c, w, consts.GenericErrors.GeneralError.Code, errors.New(consts.GenericErrors.GeneralError.Description))

		return nil
	}

	go delegatedActivateAddress(c, address, passphrase, amount, assets, activationId)

	return nil
}

// Concurrency safe to create and send transactions from a single address.
func delegatedActivateAddress(c context.Context, addressToActivate string, passphrase string, amount uint64, assets []rippleapi.Amount, activationId string) (int64, error) {
	var complete bool = false
	var linesRequired []rippleapi.Amount
	var numLinesRequired = 0
	//	var retries int = 0

	// Copy same context values to local variables which are often accessed
	accessKey := c.Value(consts.AccessKeyKey).(string)
	blockchainId := c.Value(consts.BlockchainIdKey).(string)
	//	env := c.Value(consts.EnvKey).(string)

	// Need a better way to secure internal wallets
	// Array of internal wallets that can be round robined to activate addresses
	var wallets = []struct {
		Address      string
		Passphrase   string
		BlockchainId string
	}{
		{"rpu8gxvRzQ2JLQMN7Goxs6x9zffH3sjQBd", "laid circle drag adore rainbow color peaceful other huge breathe release pen", "ripple"},
	}

	for complete == false {
		// Check the address to activate to see how much XRP it already holds
		accountInfo, _, err := rippleapi.GetAccountInfo(c, addressToActivate)
		if err != nil {
			log.FluentfContext(consts.LOGERROR, c, "Error in rippleapi.GetAccountInfo(): %s", err.Error())
			database.UpdatePaymentWithErrorByPaymentId(c, accessKey, activationId, consts.RippleErrors.MiscError.Code, consts.RippleErrors.MiscError.Description)
			return consts.RippleErrors.MiscError.Code, errors.New(consts.RippleErrors.MiscError.Description)
		}

		currentBalance, err := strconv.ParseUint(accountInfo.Balance, 10, 64)
		if err != nil {
			log.FluentfContext(consts.LOGERROR, c, "Error in ParseUint(): %s", err.Error())
			database.UpdatePaymentWithErrorByPaymentId(c, accessKey, activationId, consts.RippleErrors.MiscError.Code, consts.RippleErrors.MiscError.Description)
			return consts.RippleErrors.MiscError.Code, errors.New(consts.RippleErrors.MiscError.Description)
		}

		// Get trust lines for destination account
		lines, _, err := rippleapi.GetAccountLines(c, addressToActivate)
		if err != nil {
			log.FluentfContext(consts.LOGERROR, c, "Error in GetAccountLines(): %s", err.Error())
			database.UpdatePaymentWithErrorByPaymentId(c, accessKey, activationId, consts.RippleErrors.MiscError.Code, consts.RippleErrors.MiscError.Description)
			return consts.RippleErrors.MiscError.Code, errors.New(consts.RippleErrors.MiscError.Description)
		}

		//		 Calculate how much XRP is required as reserve for address to activate right now
		//		requiredReserve := rippleapi.CalculateReserve(c, uint64(len(lines)))

		// Find the trust lines which were requested but don't exist
		var linesUsed int = 0
		for _, asset := range assets {
			if lines.Contains(asset.Issuer, asset.Currency) {
				linesUsed++

				linesRequired = append(linesRequired, asset)
			}
		}
		numLinesRequired = len(lines) - linesUsed + 1

		// Calculate the target reserve which is reserve + enough XRP for all the trust lines which haven't been established + 1 spare
		var targetReserve uint64
		// If the account hasn't been created then the target reserve is based upon an empty wallet + requested trust lines + 1
		if accountInfo.Balance == "0" {
			targetReserve = rippleapi.CalculateReserve(c, uint64(len(assets))+1)
		} else {
			// The targetReserve is the current wallet + requested trust lines which don't already exist
			targetReserve = rippleapi.CalculateReserve(c, uint64(numLinesRequired))
		}

		// We need to send xrp to cover the difference from the amount of xrp we want to reach vs what is already in the wallet
		var amountXRPToSend = targetReserve - currentBalance

		// Add on the amount required for the number of transactions the client wishes to be able to perform
		txXRPAmount, _, err := rippleapi.CalculateFeeAmount(c, amount)
		if err != nil {
			log.FluentfContext(consts.LOGERROR, c, "Error in CalculateFeeAmount(): %s", err.Error())
			return consts.RippleErrors.MiscError.Code, errors.New(consts.RippleErrors.MiscError.Description)
		}
		amountXRPToSend += txXRPAmount

		// Pick an internal address to send from
		var randomNumber int = 0
		var sourceAddress = wallets[randomNumber].Address

		// Write the activation with the generated activation id to the database
		database.InsertActivation(c, accessKey, activationId, blockchainId, sourceAddress, amountXRPToSend)

		// Send the xrp
		_, _, err = delegatedSend(c, accessKey, wallets[randomNumber].Passphrase, wallets[randomNumber].Address, addressToActivate, "XRP", "", amountXRPToSend, activationId, "")
		if err != nil {
			log.FluentfContext(consts.LOGERROR, c, "Error in delegatedSend(): %s", err.Error())
			database.UpdatePaymentWithErrorByPaymentId(c, accessKey, activationId, consts.RippleErrors.MiscError.Code, consts.RippleErrors.MiscError.Description)
			return consts.RippleErrors.MiscError.Code, errors.New(consts.RippleErrors.MiscError.Description)
		}

		complete = true
	}

	// For each trustline which doesn't already exist, create it
	for _, line := range linesRequired {
		database.InsertTrustAsset(c, accessKey, activationId, blockchainId, line.Currency, line.Issuer, rippleapi.DefaultAmountToTrust)

		rippleAmount, err := rippleapi.Uint64ToAmount(rippleapi.DefaultAmountToTrust)
		if err != nil {
			log.FluentfContext(consts.LOGERROR, c, "Error in Uint64ToAmount(): %s", err.Error())
			return consts.RippleErrors.MiscError.Code, errors.New(consts.RippleErrors.MiscError.Description)
		}

		_, _, err = rippleapi.TrustSet(c, addressToActivate, line.Currency, rippleAmount, line.Issuer, 0, passphrase)
	}

	return 0, nil

}
