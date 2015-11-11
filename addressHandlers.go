package main

import (

	//	"errors"
	"net/http"

	"github.com/vennd/enu/consts"
	"github.com/vennd/enu/counterpartyhandlers"
	"github.com/vennd/enu/enulib"
	"github.com/vennd/enu/internal/golang.org/x/net/context"
	//	"github.com/vennd/enu/log"
	"github.com/vennd/enu/ripplehandlers"
)

func AddressCreate(c context.Context, w http.ResponseWriter, r *http.Request) *enulib.AppError {
	// Add to the context the RequestType
	c = context.WithValue(c, consts.RequestTypeKey, "address")

	// check generic args and parse
	m, err := CheckAndParseJsonCTX(c, w, r)
	if err != nil {
		// Status errors are handled inside CheckAndParseJsonCTX, so we just exit gracefully
		return nil
	}

	blockchainId := m["blockchainId"].(string)
	if m["blockchainId"] != nil {
		// check if blockchainId is valid
		c = context.WithValue(c, consts.BlockchainIdKey, blockchainId)
	} else {
		// set error no blockchain specified
	}

	if blockchainId == consts.CounterpartyBlockchainId {
		err := counterpartyhandlers.AddressCreate(c, w, r, m)

		return err
	} else if blockchainId == consts.RippleBlockchainId {
		err := ripplehandlers.AddressCreate(c, w, r, m)
		return err
	}

	return nil

}
