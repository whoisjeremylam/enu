package ripplehandlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"errors"
	"github.com/vennd/enu/consts"
	"github.com/vennd/enu/database"
	"github.com/vennd/enu/enulib"
	"github.com/vennd/enu/internal/golang.org/x/net/context"
	"github.com/vennd/enu/log"
	"github.com/vennd/enu/rippleapi"
	"github.com/vennd/enu/ripplecrypto"
)

func AssetCreate(c context.Context, w http.ResponseWriter, r *http.Request, m map[string]interface{}) *enulib.AppError {

	var assetStruct enulib.Asset
	var distributionAddress string
	var distributionPassphrase string

	requestId := c.Value(consts.RequestIdKey).(string)
	assetStruct.RequestId = requestId
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	// The issuing address
	sourceAddress := m["sourceAddress"].(string)
	passphrase := m["passphrase"].(string)

	// The address which will hold the asset once it is issued
	if m["distributionAddress"] != nil {
		distributionAddress = m["distributionAddress"].(string)
	}
	if m["distributionPassphrase"] != nil {
		distributionPassphrase = m["distributionPassphrase"].(string)
	}

	asset := m["asset"].(string)
	quantity := uint64(m["quantity"].(float64))

	log.FluentfContext(consts.LOGINFO, c, "AssetCreate: received request Address: %s, asset: %s, quantity: %d, distributionAddress: %s from accessKey: %s\n", sourceAddress, asset, quantity, distributionAddress, c.Value(consts.AccessKeyKey).(string))

	// Generate an assetId
	assetId := enulib.GenerateAssetId()
	log.Printf("Generated assetId: %s", assetId)
	assetStruct.AssetId = assetId
	rippleAsset, err := rippleapi.ToCurrency(asset)
	if err != nil {
		log.FluentfContext(consts.LOGINFO, c, "Error in call to rippleapi.ToCurrency(): %s", err.Error())
	}
	assetStruct.Asset = rippleAsset
	assetStruct.Description = asset
	assetStruct.Quantity = quantity
	assetStruct.SourceAddress = sourceAddress
	assetStruct.DistributionAddress = distributionAddress

	// Return to the client the assetId and unblock the client
	w.WriteHeader(http.StatusCreated)
	if err = json.NewEncoder(w).Encode(assetStruct); err != nil {
		panic(err)
	}

	// Start asset creation in async mode
	go delegatedAssetCreate(c, sourceAddress, passphrase, distributionAddress, distributionPassphrase, rippleAsset, asset, quantity, assetId)

	return nil
}

// Concurrency safe to create and send transactions from a single address.
func delegatedAssetCreate(c context.Context, issuingAddress string, issuingPassphrase string, distributionAddress string, distributionPassphrase string, asset string, assetDescription string, quantity uint64, assetId string) (int64, error) {
	//	var complete bool = false
	//	var numLinesRequired = 0
	//	var retries int = 0

	// Copy same context values to local variables which are often accessed
	accessKey := c.Value(consts.AccessKeyKey).(string)
	//	blockchainId := c.Value(consts.BlockchainIdKey).(string)

	// Write the asset with the generated asset id to the database
	go database.InsertAsset(accessKey, assetId, issuingAddress, distributionAddress, asset, assetDescription, quantity, true, "valid")

	// Set issuer up as a gateway https://ripple.com/build/gateway-guide/
	// set DefaultRipple on the issuer https://ripple.com/build/gateway-guide/#defaultripple
	//
	// First check if the defaultRipple flag is already set
	accountInfo, _, err := rippleapi.GetAccountInfo(c, issuingAddress)
	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Error in rippleapi.GetAccountInfo(): %s", err.Error())

		database.UpdateAssetWithErrorByAssetId(c, accessKey, assetId, consts.RippleErrors.MiscError.Code, consts.RippleErrors.MiscError.Description)
		return consts.RippleErrors.MiscError.Code, errors.New(consts.RippleErrors.MiscError.Description)
	}

	// If defaultRipple isn't set, set it
	defaultRipple := accountInfo.Flags & rippleapi.LsfDefaultRipple
	log.FluentfContext(consts.LOGINFO, c, "flags: %d", defaultRipple)
	log.FluentfContext(consts.LOGINFO, c, "defaultripple flag: %d", rippleapi.LsfDefaultRipple)

	if defaultRipple == rippleapi.LsfDefaultRipple {
		log.FluentfContext(consts.LOGINFO, c, "defaultRipple is NOT set for account %s. Setting the flag...", issuingAddress)
		txHash, _, err := rippleapi.AccountSetFlag(c, issuingAddress, 8, ripplecrypto.PassphraseToSecret(c, issuingPassphrase))
		if err != nil {
			log.FluentfContext(consts.LOGERROR, c, "Error in rippleapi.AccountSetFlag(): %s", err.Error())

			database.UpdateAssetWithErrorByAssetId(c, accessKey, assetId, consts.RippleErrors.MiscError.Code, consts.RippleErrors.MiscError.Description)
			return consts.RippleErrors.MiscError.Code, errors.New(consts.RippleErrors.MiscError.Description)
		}

		log.FluentfContext(consts.LOGINFO, c, "defaultRipple set on %s. TxId: %s", issuingAddress, txHash)
	} else {
		log.FluentfContext(consts.LOGINFO, c, "defaultRipple already set on %s", issuingAddress)
	}

	//   create a trust line between the distribution address and issuer
	if distributionAddress != "" && distributionPassphrase == "" {
		log.FluentfContext(consts.LOGERROR, c, "If a distribution address is specified, the passphrase for the distribution address must be given.")
		database.UpdateAssetWithErrorByAssetId(c, accessKey, assetId, consts.RippleErrors.DistributionPassphraseMissing.Code, consts.RippleErrors.DistributionPassphraseMissing.Description)

		return consts.RippleErrors.DistributionPassphraseMissing.Code, errors.New(consts.RippleErrors.DistributionPassphraseMissing.Description)
	}

	// If the distribution address has been specified
	if distributionAddress != "" {
		// Check if a trust line already exists between the issuing account and the distribution account
		lines, _, err := rippleapi.GetAccountLines(c, distributionAddress)
		if err != nil {
			log.FluentfContext(consts.LOGERROR, c, "Error in rippleapi.GetAccountLines: %s", err.Error())

			database.UpdateAssetWithErrorByAssetId(c, accessKey, assetId, consts.RippleErrors.MiscError.Code, consts.RippleErrors.MiscError.Description)
			return consts.RippleErrors.MiscError.Code, errors.New(consts.RippleErrors.MiscError.Description)
		}

		rippleAsset, err := rippleapi.ToCurrency(asset)
		if err != nil {
			log.FluentfContext(consts.LOGERROR, c, "Error in rippleapi.ToCurrency: %s", err.Error())

			database.UpdateAssetWithErrorByAssetId(c, accessKey, assetId, consts.RippleErrors.MiscError.Code, consts.RippleErrors.MiscError.Description)
			return consts.RippleErrors.MiscError.Code, errors.New(consts.RippleErrors.MiscError.Description)
		}

		// Create trustline if one doesn't exist
		if lines.Contains(issuingAddress, rippleAsset) == false {
			log.FluentfContext(consts.LOGINFO, c, "Trust line from distribution %s to issuer %s does not exist", distributionAddress, issuingAddress)

			// Check if the distribution address is funded sufficiently
			accountInfo, _, err := rippleapi.GetAccountInfo(c, distributionAddress)
			if err != nil {
				log.FluentfContext(consts.LOGERROR, c, "Error in rippleapi.GetAccountInfo: %s", err.Error())

				database.UpdateAssetWithErrorByAssetId(c, accessKey, assetId, consts.RippleErrors.MiscError.Code, consts.RippleErrors.MiscError.Description)
				return consts.RippleErrors.MiscError.Code, errors.New(consts.RippleErrors.MiscError.Description)
			}

			accountBalance, err := strconv.ParseUint(accountInfo.Balance, 10, 64)
			if err != nil {
				log.FluentfContext(consts.LOGERROR, c, "Error in ParseUint(): %s", err.Error())

				database.UpdateAssetWithErrorByAssetId(c, accessKey, assetId, consts.RippleErrors.MiscError.Code, consts.RippleErrors.MiscError.Description)
				return consts.RippleErrors.MiscError.Code, errors.New(consts.RippleErrors.MiscError.Description)
			}

			// Retrieve the current number of trust lines
			accountLines, _, err := rippleapi.GetAccountLines(c, distributionAddress)
			if err != nil {
				log.FluentfContext(consts.LOGERROR, c, "Error in rippleapi.GetAccountLines: %s", err.Error())

				database.UpdateAssetWithErrorByAssetId(c, accessKey, assetId, consts.RippleErrors.MiscError.Code, consts.RippleErrors.MiscError.Description)
				return consts.RippleErrors.MiscError.Code, errors.New(consts.RippleErrors.MiscError.Description)
			}

			// Calculate what the reserve should be including the new trust line that we need to create to the issuer
			reserveRequired := rippleapi.CalculateReserve(c, uint64(accountLines.Len())+1) + rippleapi.DefaultFeeI

			log.FluentfContext(consts.LOGERROR, c, "Distribution address: %s holds %d XRP, requires %d XRP", distributionAddress, accountBalance, reserveRequired)

			// If there isn't enough XRP in the address, error out
			if accountBalance < reserveRequired {
				log.FluentfContext(consts.LOGERROR, c, consts.RippleErrors.DistributionInsufficientFunds.Description)

				database.UpdateAssetWithErrorByAssetId(c, accessKey, assetId, consts.RippleErrors.DistributionInsufficientFunds.Code, consts.RippleErrors.DistributionInsufficientFunds.Description)
				return consts.RippleErrors.DistributionInsufficientFunds.Code, errors.New(consts.RippleErrors.DistributionInsufficientFunds.Description)
			}

			defaultTrustAmount, err := rippleapi.Uint64ToAmount(rippleapi.DefaultAmountToTrust)
			if err != nil {
				log.FluentfContext(consts.LOGERROR, c, "Error in rippleapi.ToCurrency (for defaultAmountToTrust): %s", err.Error())

				database.UpdateAssetWithErrorByAssetId(c, accessKey, assetId, consts.RippleErrors.MiscError.Code, consts.RippleErrors.MiscError.Description)
				return consts.RippleErrors.MiscError.Code, errors.New(consts.RippleErrors.MiscError.Description)
			}

			log.FluentfContext(consts.LOGINFO, c, "Creating trust line from distribution %s to issuer %s does not exist for %s", distributionAddress, issuingAddress, asset)

			txhash, _, err := rippleapi.TrustSet(c, distributionAddress, asset, defaultTrustAmount, issuingAddress, 0, distributionPassphrase)
			if err != nil {
				log.FluentfContext(consts.LOGERROR, c, "Error in rippleapi.TrustSet: %s", err.Error())

				database.UpdateAssetWithErrorByAssetId(c, accessKey, assetId, consts.RippleErrors.MiscError.Code, consts.RippleErrors.MiscError.Description)
				return consts.RippleErrors.MiscError.Code, errors.New(consts.RippleErrors.MiscError.Description)
			}

			log.FluentfContext(consts.LOGERROR, c, "Trust line created from distribution account: %s to issuer: %s for %s. Txhash: %s", issuingAddress, distributionAddress, asset, txhash)
		}
	}

	// If no distribution wallet was specified
	//  create the wallet
	//  activate the wallet specifying a trust line for the asset from the issuing address

	//

	//	log.FluentfContext(consts.LOGINFO, c, "Number of trust lines requested: %d", len(assets))

	//	// Copy same context values to local variables which are often accessed
	//	accessKey := c.Value(consts.AccessKeyKey).(string)
	//	blockchainId := c.Value(consts.BlockchainIdKey).(string)
	//	//	env := c.Value(consts.EnvKey).(string)

	//	// Need a better way to secure internal wallets
	//	// Array of internal wallets that can be round robined to activate addresses
	//	var wallets = []struct {
	//		Address      string
	//		Passphrase   string
	//		BlockchainId string
	//	}{
	//		{"rpu8gxvRzQ2JLQMN7Goxs6x9zffH3sjQBd", "laid circle drag adore rainbow color peaceful other huge breathe release pen", "ripple"},
	//	}

	//	var currentBalance uint64 // The current amount of ripples in the account
	//	var targetReserve uint64  // The amount we need to reach in this account to fulful the reserve and trustlines we want to create

	//	for complete == false {
	//		// Check the address to activate to see how much XRP it already holds
	//		accountInfo, _, err := rippleapi.GetAccountInfo(c, addressToActivate)
	//		if err != nil {
	//			log.FluentfContext(consts.LOGERROR, c, "Error in rippleapi.GetAccountInfo(): %s", err.Error())
	//			database.UpdatePaymentWithErrorByPaymentId(c, accessKey, activationId, consts.RippleErrors.MiscError.Code, consts.RippleErrors.MiscError.Description)
	//			return consts.RippleErrors.MiscError.Code, errors.New(consts.RippleErrors.MiscError.Description)
	//		}

	//		if accountInfo.Balance != "" {
	//			currentBalance, err = strconv.ParseUint(accountInfo.Balance, 10, 64)
	//		} else {
	//			currentBalance = 0
	//		}
	//		if err != nil {
	//			log.FluentfContext(consts.LOGERROR, c, "Error in ParseUint(): %s", err.Error())
	//			database.UpdatePaymentWithErrorByPaymentId(c, accessKey, activationId, consts.RippleErrors.MiscError.Code, consts.RippleErrors.MiscError.Description)
	//			return consts.RippleErrors.MiscError.Code, errors.New(consts.RippleErrors.MiscError.Description)
	//		}

	//		log.FluentfContext(consts.LOGINFO, c, "Wallet currently contains %d XRP", currentBalance)

	//		// Get trust lines for destination account
	//		lines, _, err := rippleapi.GetAccountLines(c, addressToActivate)
	//		if err != nil {
	//			log.FluentfContext(consts.LOGERROR, c, "Error in GetAccountLines(): %s", err.Error())
	//			database.UpdatePaymentWithErrorByPaymentId(c, accessKey, activationId, consts.RippleErrors.MiscError.Code, consts.RippleErrors.MiscError.Description)
	//			return consts.RippleErrors.MiscError.Code, errors.New(consts.RippleErrors.MiscError.Description)
	//		}

	//		//		 Calculate how much XRP is required as reserve for address to activate right now
	//		//		requiredReserve := rippleapi.CalculateReserve(c, uint64(len(lines)))

	//		// Find the trust lines which were requested but don't exist
	//		var linesUsed int = 0
	//		var assetNamesReqired []string
	//		for _, asset := range assets {
	//			if lines.Contains(asset.Issuer, asset.Currency) {
	//				linesUsed++
	//			} else {
	//				linesRequired = append(linesRequired, asset)
	//				assetNamesReqired = append(assetNamesReqired, asset.Currency+"("+asset.Issuer+")")
	//			}
	//		}
	//		numLinesRequired = len(linesRequired)

	//		log.FluentfContext(consts.LOGINFO, c, "Number of trust lines to be added: %d, %s", numLinesRequired, strings.Join(assetNamesReqired, ", "))

	//		// Calculate the target reserve which is reserve + enough XRP for all the trust lines which haven't been established + 1 spare
	//		// If the account hasn't been created then the target reserve is based upon an empty wallet + requested trust lines + 1
	//		targetReserve = rippleapi.CalculateReserve(c, uint64(len(lines))+uint64(numLinesRequired))

	//		// If the current balance is higher than the target reserve, then we don't need to send any XRP to meet the reserve
	//		if currentBalance >= targetReserve {
	//			targetReserve = currentBalance
	//		}

	//		// We need to send xrp to cover the difference from the amount of xrp we want to reach vs what is already in the wallet
	//		var amountXRPToSend = targetReserve - currentBalance

	//		log.FluentfContext(consts.LOGINFO, c, "XRP required to cover reserve + lines requested: %d", amountXRPToSend)

	//		// Add on the amount required for the number of transactions the client wishes to be able to perform
	//		txXRPAmount, _, err := rippleapi.CalculateFeeAmount(c, amount)
	//		if err != nil {
	//			log.FluentfContext(consts.LOGERROR, c, "Error in CalculateFeeAmount(): %s", err.Error())
	//			return consts.RippleErrors.MiscError.Code, errors.New(consts.RippleErrors.MiscError.Description)
	//		}
	//		amountXRPToSend += txXRPAmount

	//		log.FluentfContext(consts.LOGINFO, c, "XRP for %d transactions txXRPAmount: %d", amount, txXRPAmount)
	//		log.FluentfContext(consts.LOGINFO, c, "XRP that we need to send from our master wallet: %d", amountXRPToSend)

	//		// Pick an internal address to send from
	//		var randomNumber int = 0
	//		var sourceAddress = wallets[randomNumber].Address

	//		// Write the activation with the generated activation id to the database
	//		database.InsertActivation(c, accessKey, activationId, blockchainId, sourceAddress, amountXRPToSend)

	//		// Send the xrp
	//		//		_, _, err = delegatedSend(c, accessKey, wallets[randomNumber].Passphrase, wallets[randomNumber].Address, addressToActivate, "XRP", "", amountXRPToSend, activationId, "")
	//		//		if err != nil {
	//		//			log.FluentfContext(consts.LOGERROR, c, "Error in delegatedSend(): %s", err.Error())
	//		//			database.UpdatePaymentWithErrorByPaymentId(c, accessKey, activationId, consts.RippleErrors.MiscError.Code, consts.RippleErrors.MiscError.Description)
	//		//			return consts.RippleErrors.MiscError.Code, errors.New(consts.RippleErrors.MiscError.Description)
	//		//		}

	//		complete = true

	//		// Throttle
	//		if complete == false {
	//			log.FluentfContext(consts.LOGERROR, c, "Unable to send XRP waiting 1 minute and retrying...")

	//			time.Sleep(time.Duration(1) * time.Minute)
	//		}
	//	}

	//	// If reserve already is sufficient, proceed to create the trust line straight away, otherwise wait
	//	if currentBalance < targetReserve {
	//		log.FluentfContext(consts.LOGERROR, c, "Waiting for send of XRP to complete...")

	//		time.Sleep(time.Duration(10000) * time.Millisecond)

	//		log.FluentfContext(consts.LOGERROR, c, "Wait complete")
	//	}

	//	// For each trustline which doesn't already exist, create it
	//	for _, line := range linesRequired {
	//		log.FluentfContext(consts.LOGERROR, c, "Creating trust line: %s, %s, %d", line.Currency, line.Issuer, rippleapi.DefaultAmountToTrust)
	//		database.InsertTrustAsset(c, accessKey, activationId, blockchainId, line.Currency, line.Issuer, rippleapi.DefaultAmountToTrust)

	//		// Convert int to the ripple amount
	//		rippleAmount, err := rippleapi.Uint64ToAmount(rippleapi.DefaultAmountToTrust)
	//		if err != nil {
	//			log.FluentfContext(consts.LOGERROR, c, "Error in Uint64ToAmount(): %s", err.Error())
	//			return consts.RippleErrors.MiscError.Code, errors.New(consts.RippleErrors.MiscError.Description)
	//		}

	//		// Convert asset name to ripple currency name
	//		currency, err := rippleapi.ToCurrency(line.Currency)
	//		if err != nil {
	//			log.FluentfContext(consts.LOGERROR, c, "Error in Uint64ToAmount(): %s", err.Error())
	//			return consts.RippleErrors.MiscError.Code, errors.New(consts.RippleErrors.MiscError.Description)
	//		}

	//		// Convert passphrase to ripple secret
	//		seed := mneumonic.FromWords(strings.Split(passphrase, " "))
	//		secret, err := ripplecrypto.ToSecret(seed.ToHex())
	//		if err != nil {
	//			log.FluentfContext(consts.LOGERROR, c, "Error in ripplecrypto.ToSecret(): %s", err.Error())
	//			return consts.RippleErrors.MiscError.Code, errors.New(consts.RippleErrors.MiscError.Description)
	//		}

	//		_, _, err = rippleapi.TrustSet(c, addressToActivate, currency, rippleAmount, line.Issuer, 0, secret)
	//	}

	//	log.FluentfContext(consts.LOGERROR, c, "delegatedActivateAddress() complete")

	return 0, nil

}
