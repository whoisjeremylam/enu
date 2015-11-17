package ripplehandlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/vennd/enu/consts"
	"github.com/vennd/enu/enulib"
	"github.com/vennd/enu/handlers"
	"github.com/vennd/enu/internal/github.com/vennd/mneumonic"
	"github.com/vennd/enu/log"
	"github.com/vennd/enu/rippleapi"

	"github.com/vennd/enu/internal/golang.org/x/net/context"
)

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
