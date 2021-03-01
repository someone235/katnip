package main

import (
	"github.com/kaspanet/kaspad/util/panics"
	"github.com/someone235/katnip/server/logger"
)

var (
	log   = logger.Logger("KVSD")
	spawn = panics.GoroutineWrapperFunc(log)
)
