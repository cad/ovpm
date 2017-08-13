# OVPM - OpenVPn Manager

[![Build Status](https://travis-ci.org/cad/ovpm.svg?branch=master)](https://travis-ci.org/cad/ovpm)
[![codecov](https://codecov.io/gh/cad/ovpm/branch/master/graph/badge.svg)](https://codecov.io/gh/cad/ovpm)

OVPM allows you to manage an OpenVPN server from command line easily. With OVPM you can create and run an OpenVPN server, add/remove vpn users, generate client .ovpn files for your users etc. 

## Usage

Install OVPM:

```bash
$ go get github.com/cad/ovpm/...
```

And run the server ovpmd:

```bash
$ sudo ovpmd --version

ovpmd version 0.1.0
```

In another terminal ovpm command line tool, ovpm:

```bash
$ ovpm 

NAME:
   ovpm - OpenVPN Manager

USAGE:
   main [global options] command [command options] [arguments...]

VERSION:
   0.1.0

COMMANDS:
     user     User Operations
     vpn      VPN Operations
     help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --verbose            verbose output
   --daemon-port value  port number for OVPM daemon to call
   --help, -h           show help
   --version, -v        print the version

```
