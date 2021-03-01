package controllers

import (
	"net/http"

	"github.com/kaspanet/kaspad/util"
	"github.com/pkg/errors"
	"github.com/someone235/katnip/server/httpserverutils"
	"github.com/someone235/katnip/server/serverd/config"
)

func validateAddress(address string) error {
	_, err := util.DecodeAddress(address, config.ActiveConfig().ActiveNetParams.Prefix)
	if err != nil {
		return httpserverutils.NewHandlerErrorWithCustomClientMessage(http.StatusUnprocessableEntity,
			errors.Wrap(err, "error decoding address"),
			"The given address is not a well-formatted P2PKH or P2SH address")
	}

	return nil
}
