package generalhandlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/whoisjeremylam/enu/bitcoinapi"
	"github.com/whoisjeremylam/enu/consts"
	"github.com/whoisjeremylam/enu/database"
	"github.com/whoisjeremylam/enu/enulib"
	"github.com/whoisjeremylam/enu/handlers"
	"github.com/whoisjeremylam/enu/internal/github.com/gorilla/mux"
	"github.com/whoisjeremylam/enu/internal/golang.org/x/net/context"
	"github.com/whoisjeremylam/enu/log"
)

func GetAsset(c context.Context, w http.ResponseWriter, r *http.Request, m map[string]interface{}) *enulib.AppError {
	//	var asset enulib.Asset
	requestId := c.Value(consts.RequestIdKey).(string)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	vars := mux.Vars(r)
	assetId := vars["assetId"]

	if assetId == "" || len(assetId) < 16 {
		log.FluentfContext(consts.LOGERROR, c, "Invalid assetId")
		handlers.ReturnUnprocessableEntity(c, w, consts.GenericErrors.InvalidAssetId.Code, errors.New(consts.GenericErrors.InvalidAssetId.Description))

		return nil

	}

	log.FluentfContext(consts.LOGINFO, c, "GetAsset called for '%s' by '%s'\n", assetId, c.Value(consts.AccessKeyKey).(string))

	asset, err := database.GetAssetByAssetId(c, c.Value(consts.AccessKeyKey).(string), assetId)
	if err != nil {
		handlers.ReturnServerError(c, w)

		return nil
	}
	asset.RequestId = requestId
	
		if asset.BlockchainId == consts.RippleBlockchainId {
		asset.Issuer = asset.SourceAddress
	}

	// Add the blockchain status
	if asset.BroadcastTxId != "" && asset.BlockchainId == consts.CounterpartyBlockchainId {
		confirmations, err := bitcoinapi.GetConfirmations(asset.BroadcastTxId)
		if err == nil || confirmations == 0 {
			asset.BlockchainStatus = "unconfimed"
			asset.BlockchainConfirmations = 0
		}

		asset.BlockchainStatus = "confirmed"
		asset.BlockchainConfirmations = confirmations
	}

	if err := json.NewEncoder(w).Encode(asset); err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Error in Encode(): %s", err.Error())
		handlers.ReturnServerError(c, w)

		return nil
	}

	w.WriteHeader(http.StatusOK)
	return nil
}
