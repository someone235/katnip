package main

import (
	"fmt"
	"os"

	"github.com/kaspanet/kaspad/util/profiling"

	"github.com/pkg/errors"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/kaspanet/kaspad/infrastructure/os/signal"
	"github.com/kaspanet/kaspad/util/panics"
	"github.com/someone235/katnip/server/database"
	"github.com/someone235/katnip/server/kaspadrpc"
	"github.com/someone235/katnip/server/serverd/config"
	"github.com/someone235/katnip/server/serverd/server"
	"github.com/someone235/katnip/server/version"
)

func main() {
	defer panics.HandlePanic(log, "main", nil)
	interrupt := signal.InterruptListener()

	err := config.Parse()
	if err != nil {
		errString := fmt.Sprintf("Error parsing command-line arguments: %s\n", err)
		_, fErr := fmt.Fprint(os.Stderr, errString)
		if fErr != nil {
			panic(errString)
		}
		return
	}

	// Show version at startup.
	log.Infof("Version %s", version.Version())

	// Start the profiling server if required
	if config.ActiveConfig().Profile != "" {
		profiling.Start(config.ActiveConfig().Profile, log)
	}

	err = database.Connect(&config.ActiveConfig().CommonConfigFlags)
	if err != nil {
		panic(errors.Errorf("Error connecting to database: %s", err))
	}
	defer func() {
		err := database.Close()
		if err != nil {
			panic(errors.Errorf("Error closing the database: %s", err))
		}
	}()

	client, err := kaspadrpc.NewClient(&config.ActiveConfig().CommonConfigFlags, false)
	if err != nil {
		panic(errors.Errorf("Error connecting to servers: %s", err))
	}
	defer client.Close()

	shutdownServer := server.Start(config.ActiveConfig().HTTPListen)
	defer shutdownServer()

	<-interrupt
}
