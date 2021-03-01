package config

import (
	"path/filepath"
	"strconv"

	"github.com/jessevdk/go-flags"
	"github.com/kaspanet/kaspad/infrastructure/config"
	"github.com/pkg/errors"
	"github.com/someone235/katnip/server/logger"
)

// CommonConfigFlags holds configuration common to both the server and the sync daemon.
type CommonConfigFlags struct {
	ShowVersion bool   `short:"V" long:"version" description:"Display version information and exit"`
	LogDir      string `long:"logdir" description:"Directory to log output."`
	DebugLevel  string `short:"d" long:"debuglevel" description:"Set log level {trace, debug, info, warn, error, critical}"`
	DBAddress   string `long:"dbaddress" description:"Database address" default:"localhost:5432"`
	DBSSLMode   string `long:"dbsslmode" description:"Database SSL mode" choice:"disable" choice:"allow" choice:"prefer" choice:"require" choice:"verify-ca" choice:"verify-full" default:"disable"`
	DBUser      string `long:"dbuser" description:"Database user" required:"true"`
	DBPassword  string `long:"dbpass" description:"Database password" required:"true"`
	DBName      string `long:"dbname" description:"Database name" required:"true"`
	RPCServer   string `short:"s" long:"rpcserver" description:"RPC server to connect to"`
	Profile     string `long:"profile" description:"Enable HTTP profiling on the given port"`
	config.NetworkFlags
}

// ResolveCommonFlags parses command line arguments and sets CommonConfigFlags accordingly.
func (commonFlags *CommonConfigFlags) ResolveCommonFlags(parser *flags.Parser,
	defaultLogDir, logFilename, errLogFilename string, isMigrate bool) error {
	if commonFlags.LogDir == "" {
		commonFlags.LogDir = defaultLogDir
	}
	logFile := filepath.Join(commonFlags.LogDir, logFilename)
	errLogFile := filepath.Join(commonFlags.LogDir, errLogFilename)
	logger.InitLog(logFile, errLogFile)

	if commonFlags.DebugLevel != "" {
		err := logger.SetLogLevels(commonFlags.DebugLevel)
		if err != nil {
			return err
		}
	}

	if commonFlags.RPCServer == "" && !isMigrate {
		return errors.New("--rpcserver is required")
	}

	if isMigrate {
		return nil
	}

	if commonFlags.Profile != "" {
		profilePort, err := strconv.Atoi(commonFlags.Profile)
		if err != nil || profilePort < 1024 || profilePort > 65535 {
			return errors.New("The profile port must be between 1024 and 65535")
		}
	}

	return commonFlags.ResolveNetwork(parser)
}
