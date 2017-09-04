# OVPM - OpenVPN Management Server

[![Build Status](https://travis-ci.org/cad/ovpm.svg?branch=master)](https://travis-ci.org/cad/ovpm)
[![GitHub version](https://badge.fury.io/gh/cad%2Fovpm.svg)](https://badge.fury.io/gh/cad%2Fovpm)
[![codecov](https://codecov.io/gh/cad/ovpm/branch/master/graph/badge.svg)](https://codecov.io/gh/cad/ovpm)
[![GoDoc](https://godoc.org/github.com/cad/ovpm?status.svg)](https://godoc.org/github.com/cad/ovpm)

*OVPM* allows you to administrate an **OpenVPN** server on linux easily via command line. 

With OVPM you can create and run an OpenVPN server, add/remove VPN users, generate client .ovpn files for your users etc. 

*This software is not stable yet. We recommend against using it for anything serious until, version 1.0 is released.*

**Roadmap**

- [x] OpenVPN management functionality
- [x] User management functionality
- [x] Network management functionality
- [x] Command Line Interface (CLI)
- [ ] Web User Interface (WebUI)
- [ ] Import/Export/Backup OVPM config
- [ ] Effortless client profile (.ovpn file) delivery over Web
 
## Installation
**from RPM (CentOS/Fedora):**

```bash
# Add YUM Repo
$ sudo yum-config-manager --add-repo https://cad.github.io/ovpm/rpm/ovpm.repo

# Install OVPM
$ sudo yum install ovpm
```

**from DEB (Ubuntu/Debian):**

This is tested only on Ubuntu >=16.04.3 LTS

```bash
# Add APT Repo
$ sudo sh -c 'echo "deb [trusted=yes] https://cad.github.io/ovpm/deb/ ovpm main" >> /etc/apt/sources.list'
$ sudo apt update

# Install OVPM
$ sudo apt install ovpm
```

**from Source (go get):**

Only dependency for ovpm is **OpenVPN>=2.3**.

```bash
$ go get -u github.com/cad/ovpm/...
```

## Start the Server
You need to start the start OVPM server, which is called **ovpmd**, before doing anything.

**CentOS/Fedora/Ubuntu/Debian (RPM or DEB Package)**

Just use systemd to manage ovpmd.

```bash
$ systemctl start ovpmd
$ systemctl enable ovpmd  # enable ovpmd to start on boot
```

**If You've Installed From Source (go get)**

Run in another terminal.

```bash
$ sudo ovpmd

INFO[0000] OVPM is running :9090 ...                    
ERRO[0000] can not launch OpenVPN because system is not initialized 
```

It complains about an error due to server not being initialized, it's completely fine getting this when you first start **ovpmd**.


## Usage

**Demo**
Here is a little demo of what it looks on terminal to init the server,  create a vpn user and generate **.ovpn** file for the created user.

[![asciicast](https://asciinema.org/a/136016.png)](https://asciinema.org/a/136016)


### Init Server
If you just installed the ovpm from scratch you have started the **ovpm server** (ovpmd) then now you need to initialize the server.

You can do so by invoking;

```bash
$ ovpm vpn init -s <your-vpn-server's-fqdn-or-ip-addr>

This operation will cause invalidation of existing user certificates.
After this opeartion, new client config files (.ovpn) should be generated for each existing user.

Are you sure ? (y/N)
y
INFO[0003] ovpm server initialized 
```

Now you have your server initialized, up and running.

### Create a VPN user
If you have initialized your ovpm server now you can add users.

Add a VPN user;

```bash
$ ovpm user create -u john -p 1234            

INFO[0000] user created: john  
```

Please note that user password is taken but it will be used in the future releases. Such as for the Web UI and etc..


### Export the OpenVPN Client Config
After creating a user, you can export the client config for them.

```bash
$ ovpm user genconfig -u john

INFO[0000] exported to john.ovpn
```

This .ovpn file contains all necesarray bits and pieces for the client to connect to your newly created VPN server.
You can copy the OpenVPN client config file (e.g. john.ovpn) to the any OpenVPN client and use it to connect to your VPN server.
