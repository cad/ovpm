# Notice
This is my fork of [https://github.com/cad/ovpm](OVPM v0.2.7) for [https://targetpractice.network](TargetPractice).
I'll be adding features that suit my use case, and it may become backwards incompatible with the original unless my changes are found useful by cad and accepted upstream.
Since the original is GPLv3, I'll be publishing all my changes to the code.

# OVPM - OpenVPN Management Server

[![Build Status](https://travis-ci.org/cad/ovpm.svg?branch=master)](https://travis-ci.org/cad/ovpm)
[![GitHub version](https://badge.fury.io/gh/cad%2Fovpm.svg)](https://badge.fury.io/gh/cad%2Fovpm)
[![codecov](https://codecov.io/gh/cad/ovpm/branch/master/graph/badge.svg)](https://codecov.io/gh/cad/ovpm)
[![GoDoc](https://godoc.org/github.com/cad/ovpm?status.svg)](https://godoc.org/github.com/cad/ovpm)

*OVPM* allows you to administrate an **OpenVPN** server on linux easily via command line and web interface. 

With OVPM you can create and run an OpenVPN server, add/remove VPN users, generate client .ovpn files for your users etc. 

*This software is not stable yet. We recommend against using it for anything serious until, version 1.0 is released.*

**Roadmap**

- [x] OpenVPN management functionality
- [x] User management functionality
- [x] Network management functionality
- [x] Command Line Interface (CLI)
- [x] API (REST and gRPC)
- [x] Web User Interface (WebUI)
- [ ] Import/Export/Backup OVPM config
- [ ] Effortless client profile (.ovpn file) delivery over Web
- [ ] Monitoring and Quota functionality

**Demo**
Here is a little demo of what it looks on terminal to init the server, create a vpn user and generate **.ovpn** file for the created user.

[![asciicast](https://asciinema.org/a/136016.png)](https://asciinema.org/a/136016)

 
## Installation
**from RPM (CentOS/Fedora):**

```bash
# Add YUM Repo
$ sudo yum-config-manager --add-repo https://cad.github.io/ovpm/rpm/ovpm.repo

# Install OVPM
$ sudo yum install ovpm

# Enable and start ovpmd service
$ systemctl start ovpmd
$ systemctl enable ovpmd
```

**from DEB (Ubuntu/Debian):**

This is tested only on Ubuntu >=16.04.3 LTS

```bash
# Add APT Repo
$ sudo sh -c 'echo "deb [trusted=yes] https://cad.github.io/ovpm/deb/ ovpm main" >> /etc/apt/sources.list'
$ sudo apt update

# Install OVPM
$ sudo apt install ovpm

# Enable and start ovpmd service
$ systemctl start ovpmd
$ systemctl enable ovpmd  
```

**from Source (go get):**

Only dependency for ovpm is **OpenVPN>=2.3.3**.

```bash
$ go get -u github.com/cad/ovpm/...

# Make sure user nobody and group nogroup is available
# on the system
$ sudo useradd nobody
$ sudo groupadd nogroup

# Start ovpmd on a seperate terminal
$ sudo ovpmd
```

Now ovpmd should be running.

## Quickstart
Create a vpn user and export vpn profile for the created user.

```bash
# We should init the server after fresh install
$ ovpm vpn init --hostname <vpn.example.com>
INFO[0004] ovpm server initialized

# Now, lets create a new vpn user
$ ovpm user create -u joe -p verySecretPassword
INFO[0000] user created: joe

# Finally export the vpn profile for, the created user, joe
$ ovpm user genconfig -u joe
INFO[0000] exported to joe.ovpn
```

OpenVPN profile for user joe is exported to joe.ovpn file.
You can simply use this file with OpenVPN to connect to the vpn server from 
another computer.


# Next Steps

* [User Management](https://github.com/cad/ovpm/wiki/User-Management)
* [Network Management](https://github.com/cad/ovpm/wiki/Network-Management)
* [Web Interface](https://github.com/cad/ovpm/wiki/Web-Interface)
