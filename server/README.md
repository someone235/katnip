Katnip Server
====

[![ISC License](http://img.shields.io/badge/license-ISC-blue.svg)](https://choosealicense.com/licenses/isc/)

Katnip server is an API server that is used as the backend for the Katnip block explorer.

This project contains the following executables:
- serverd - the server. Handles user requests.
- syncd - the sync daemon. Maintains sync with a full Kaspa node.

## Requirements

Latest version of [Go](http://golang.org) (currently 1.16).

## Installation

#### Build from Source

- Install Go according to the installation instructions here:
  http://golang.org/doc/install

- Ensure Go was installed properly and is a supported version:

```bash
$ go version
$ go env GOROOT GOPATH
```

NOTE: The `GOROOT` and `GOPATH` above must not be the same path. It is
recommended that `GOPATH` is set to a directory in your home directory such as
`~/dev/go` to avoid write permission issues. It is also recommended to add
`$GOPATH/bin` to your `PATH` at this point.

- Run the following commands to obtain and install server, syncd, and the wallet including all dependencies:

```bash
$ git clone https://github.com/someone235/katnip/server $GOPATH/src/github.com/someone235/katnip/server
$ cd $GOPATH/src/github.com/someone235/katnip/server
$ go install ./...
```

- serverd, syncd, and the wallet should now be installed in `$GOPATH/bin`. If you did
  not already add the bin directory to your system path during Go installation,
  you are encouraged to do so now.


## Getting Started

The Katnip server expects to have access to the following systems:
- A Kaspa RPC server (usually [kaspad](https://github.com/kaspanet/kaspad) with RPC turned on)
- A Postgres database
- An optional MQTT broker

### Linux/BSD/POSIX/Source

#### serverd

```bash
$ ./serverd --rpcserver=localhost:16210 --rpccert=path/to/rpc.cert --rpcuser=user --rpcpass=pass --dbuser=user --dbpass=pass --dbaddress=localhost:3306 --dbname=katnip --testnet
```

#### syncd

```bash
$ ./syncd --rpcserver=localhost:16210 --rpccert=path/to/rpc.cert --rpcuser=user --rpcpass=pass --dbuser=user --dbpass=pass --dbaddress=localhost:3306 --dbname=katnip --migrate --testnet
$ ./syncd --rpcserver=localhost:16210 --rpccert=path/to/rpc.cert --rpcuser=user --rpcpass=pass --dbuser=user --dbpass=pass --dbaddress=localhost:3306 --dbname=katnip --mqttaddress=localhost:1883 --mqttuser=user --mqttpass=pass --testnet
```

## Discord
Join our discord server using the following link: https://discord.gg/WmGhhzk

## Issue Tracker

The [integrated github issue tracker](https://github.com/someone235/katnip/issues)
is used for this project.

## License

Katnip Server is licensed under the copyfree [ISC License](https://choosealicense.com/licenses/isc/).

