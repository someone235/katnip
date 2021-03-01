package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jessevdk/go-flags"
	"github.com/kaspanet/kaspad/util"
	"github.com/someone235/katnip/server/config"
	"github.com/someone235/katnip/server/version"
)

const (
	logFilename    = "serverd.log"
	errLogFilename = "serverd_err.log"
)

var (
	// Default configuration options
	defaultLogDir     = util.AppDataDir("serverd", false)
	defaultHTTPListen = "0.0.0.0:8080"
	activeConfig      *Config
)

// ActiveConfig returns the active configuration struct
func ActiveConfig() *Config {
	return activeConfig
}

// Config defines the configuration options for the API server.
type Config struct {
	HTTPListen string `long:"listen" description:"HTTP address to listen on (default: 0.0.0.0:8080)"`
	config.CommonConfigFlags
}

// Parse parses the CLI arguments and returns a config struct.
func Parse() error {
	activeConfig = &Config{
		HTTPListen: defaultHTTPListen,
	}
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

	err = activeConfig.ResolveCommonFlags(parser, defaultLogDir, logFilename, errLogFilename, false)
	if err != nil {
		return err
	}

	return nil
}
