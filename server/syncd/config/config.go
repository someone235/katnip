package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jessevdk/go-flags"
	"github.com/kaspanet/kaspad/util"
	"github.com/pkg/errors"
	"github.com/someone235/katnip/server/config"
	"github.com/someone235/katnip/server/version"
)

const (
	logFilename    = "katnip_syncd.log"
	errLogFilename = "katnip_syncd_err.log"
)

var (
	// Default configuration options
	defaultLogDir = util.AppDir("katnip_syncd", false)
	activeConfig  *Config
)

// ActiveConfig returns the active configuration struct
func ActiveConfig() *Config {
	return activeConfig
}

// Config defines the configuration options for the sync daemon.
type Config struct {
	Migrate           bool   `long:"migrate" description:"Migrate the database to the latest version. The daemon will not start when using this flag."`
	MQTTBrokerAddress string `long:"mqttaddress" description:"MQTT broker address" required:"false"`
	MQTTUser          string `long:"mqttuser" description:"MQTT server user" required:"false"`
	MQTTPassword      string `long:"mqttpass" description:"MQTT server password" required:"false"`
	config.CommonConfigFlags
}

// Parse parses the CLI arguments and returns a config struct.
func Parse() error {
	activeConfig = &Config{}
	parser := flags.NewParser(activeConfig, flags.HelpFlag)
	_, err := parser.Parse()
	// Show the version and exit if the version flag was specified.

	if activeConfig.ShowVersion {
		appName := filepath.Base(os.Args[0])
		appName = strings.TrimSuffix(appName, filepath.Ext(appName))
		fmt.Println(appName, "version", version.Version())
		os.Exit(0)
	}

	if err != nil {
		return err
	}

	err = activeConfig.ResolveCommonFlags(parser, defaultLogDir, logFilename, errLogFilename, activeConfig.Migrate)
	if err != nil {
		return err
	}

	if (activeConfig.MQTTBrokerAddress != "" || activeConfig.MQTTUser != "" || activeConfig.MQTTPassword != "") &&
		(activeConfig.MQTTBrokerAddress == "" || activeConfig.MQTTUser == "" || activeConfig.MQTTPassword == "") {
		return errors.New("--mqttaddress, --mqttuser, and --mqttpass must be passed all together")
	}

	return nil
}
