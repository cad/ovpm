# OVPM - OpenVPn Manager

[![Build Status](https://travis-ci.org/cad/ovpm.svg?branch=master)](https://travis-ci.org/cad/ovpm)
[![codecov](https://codecov.io/gh/cad/ovpm/branch/master/graph/badge.svg)](https://codecov.io/gh/cad/ovpm)
[![GoDoc](https://godoc.org/github.com/cad/ovpm?status.svg)](https://godoc.org/github.com/cad/ovpm)

OVPM allows you to manage an OpenVPN server from command line easily. With OVPM you can create and run an OpenVPN server, add/remove vpn users, generate client .ovpn files for your users etc. 

## Usage

Install OVPM:

```bash
$ go get -u github.com/cad/ovpm/...
```

And verify the installation by running ovpmd:

```bash
$ ovpmd --version

ovpmd version 0.1.0
```

And also make sure openvpn is also installed on the host:

```bash
$ openvpn --version

OpenVPN 2.4.3 x86_64-pc-linux-gnu [SSL (OpenSSL)] [LZO] [LZ4] [EPOLL] [PKCS11] [MH/PKTINFO]
...

```

Now you can actually run the ovpmd server:

```bash
# Since ovpmd launches and supervises openvpn binary it needs root privileges.
$ sudo ovpmd

INFO[0000] OVPM is running :9090 ...                    
```

In another terminal you can use ovpm via the command line tool, ovpm:

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
