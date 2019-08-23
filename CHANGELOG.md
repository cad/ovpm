# Change Log

## [Unreleased](https://github.com/cad/ovpm/tree/HEAD)

[Full Changelog](https://github.com/cad/ovpm/compare/v0.2.6...HEAD)

**Implemented enhancements:**

- webui:  make sure webui fetches all assets from the ovpm http server [\#65](https://github.com/cad/ovpm/issues/65)
- Information about certificates expiration.... [\#59](https://github.com/cad/ovpm/issues/59)
- add '\*' wildcard option to cli for bulk operations [\#46](https://github.com/cad/ovpm/issues/46)
- show connected users [\#40](https://github.com/cad/ovpm/issues/40)
- add cli tests [\#38](https://github.com/cad/ovpm/issues/38)

**Fixed bugs:**

- web ui: when user list is long, some usernames are not visible in dropdown [\#64](https://github.com/cad/ovpm/issues/64)
- on centos there is no nogroup group defined [\#62](https://github.com/cad/ovpm/issues/62)

**Closed issues:**

- install fresh TUNTAP error [\#61](https://github.com/cad/ovpm/issues/61)
- Not work depot [\#60](https://github.com/cad/ovpm/issues/60)
- All Set, but from VPN Client i can´t reach other network... [\#58](https://github.com/cad/ovpm/issues/58)

## [v0.2.6](https://github.com/cad/ovpm/tree/v0.2.6) (2017-10-16)
[Full Changelog](https://github.com/cad/ovpm/compare/v0.2.5...v0.2.6)

**Fixed bugs:**

- ccd emit doesn't work correctly [\#57](https://github.com/cad/ovpm/issues/57)

## [v0.2.5](https://github.com/cad/ovpm/tree/v0.2.5) (2017-10-13)
[Full Changelog](https://github.com/cad/ovpm/compare/v0.2.4...v0.2.5)

**Implemented enhancements:**

- WARNING: INSECURE cipher with block size less than 128 bit \(64 bit\).  [\#56](https://github.com/cad/ovpm/issues/56)
- expose vpn restart functionality over api [\#50](https://github.com/cad/ovpm/issues/50)

**Fixed bugs:**

- WEBUI \> NETWORKS [\#55](https://github.com/cad/ovpm/issues/55)
- block giving vpn server's internal ip to some user as static ip [\#54](https://github.com/cad/ovpm/issues/54)
- when starting ensure emit [\#51](https://github.com/cad/ovpm/issues/51)

## [v0.2.4](https://github.com/cad/ovpm/tree/v0.2.4) (2017-10-06)
[Full Changelog](https://github.com/cad/ovpm/compare/v0.2.3...v0.2.4)

**Fixed bugs:**

- allow update if the target ip is user's already statically allocated ip [\#49](https://github.com/cad/ovpm/issues/49)

## [v0.2.3](https://github.com/cad/ovpm/tree/v0.2.3) (2017-10-06)
[Full Changelog](https://github.com/cad/ovpm/compare/v0.2.2...v0.2.3)

## [v0.2.2](https://github.com/cad/ovpm/tree/v0.2.2) (2017-10-05)
[Full Changelog](https://github.com/cad/ovpm/compare/v0.2.1...v0.2.2)

## [v0.2.1](https://github.com/cad/ovpm/tree/v0.2.1) (2017-10-04)
[Full Changelog](https://github.com/cad/ovpm/compare/v0.2.0...v0.2.1)

## [v0.2.0](https://github.com/cad/ovpm/tree/v0.2.0) (2017-10-03)
[Full Changelog](https://github.com/cad/ovpm/compare/v0.1.18...v0.2.0)

**Fixed bugs:**

- Ubuntu Group needs to be "nogroup" [\#48](https://github.com/cad/ovpm/issues/48)

## [v0.1.18](https://github.com/cad/ovpm/tree/v0.1.18) (2017-09-19)
[Full Changelog](https://github.com/cad/ovpm/compare/v0.1.17...v0.1.18)

**Implemented enhancements:**

- web ui [\#11](https://github.com/cad/ovpm/issues/11)

## [v0.1.17](https://github.com/cad/ovpm/tree/v0.1.17) (2017-09-19)
[Full Changelog](https://github.com/cad/ovpm/compare/v0.1.16...v0.1.17)

## [v0.1.16](https://github.com/cad/ovpm/tree/v0.1.16) (2017-09-19)
[Full Changelog](https://github.com/cad/ovpm/compare/v0.1.15...v0.1.16)

**Fixed bugs:**

- ovpmd.service wrong exe path... [\#47](https://github.com/cad/ovpm/issues/47)

## [v0.1.15](https://github.com/cad/ovpm/tree/v0.1.15) (2017-09-12)
[Full Changelog](https://github.com/cad/ovpm/compare/v0.1.14...v0.1.15)

**Implemented enhancements:**

- rest api authentication [\#45](https://github.com/cad/ovpm/issues/45)

## [v0.1.14](https://github.com/cad/ovpm/tree/v0.1.14) (2017-09-03)
[Full Changelog](https://github.com/cad/ovpm/compare/v0.1.13...v0.1.14)

## [v0.1.13](https://github.com/cad/ovpm/tree/v0.1.13) (2017-09-03)
[Full Changelog](https://github.com/cad/ovpm/compare/v0.1.12...v0.1.13)

**Implemented enhancements:**

- change dns to push to clients [\#41](https://github.com/cad/ovpm/issues/41)

## [v0.1.12](https://github.com/cad/ovpm/tree/v0.1.12) (2017-09-01)
[Full Changelog](https://github.com/cad/ovpm/compare/v0.1.11...v0.1.12)

**Implemented enhancements:**

- be able to change initial ip block [\#29](https://github.com/cad/ovpm/issues/29)

## [v0.1.11](https://github.com/cad/ovpm/tree/v0.1.11) (2017-08-31)
[Full Changelog](https://github.com/cad/ovpm/compare/v0.1.10...v0.1.11)

**Fixed bugs:**

- can add duplicate static ip [\#37](https://github.com/cad/ovpm/issues/37)
- net def --via flag doesn't work as documented [\#36](https://github.com/cad/ovpm/issues/36)
- Error when group 'nobody' doesn't exist [\#32](https://github.com/cad/ovpm/issues/32)
- --static option doesn't work when user update [\#28](https://github.com/cad/ovpm/issues/28)

**Merged pull requests:**

- openvpn user created by openvpn package, so use openvpn user instead. [\#35](https://github.com/cad/ovpm/pull/35) ([ilkerdagli](https://github.com/ilkerdagli))

## [v0.1.10](https://github.com/cad/ovpm/tree/v0.1.10) (2017-08-29)
[Full Changelog](https://github.com/cad/ovpm/compare/v0.1.9...v0.1.10)

**Implemented enhancements:**

- command line flags for tcp or udp at initialize [\#30](https://github.com/cad/ovpm/issues/30)
- show network types in cli [\#27](https://github.com/cad/ovpm/issues/27)

## [v0.1.9](https://github.com/cad/ovpm/tree/v0.1.9) (2017-08-27)
[Full Changelog](https://github.com/cad/ovpm/compare/v0.1.8...v0.1.9)

**Implemented enhancements:**

- static route support  [\#21](https://github.com/cad/ovpm/issues/21)

## [v0.1.8](https://github.com/cad/ovpm/tree/v0.1.8) (2017-08-26)
[Full Changelog](https://github.com/cad/ovpm/compare/v0.1.7...v0.1.8)

**Implemented enhancements:**

- access control to existing networks on the machine [\#1](https://github.com/cad/ovpm/issues/1)

## [v0.1.7](https://github.com/cad/ovpm/tree/v0.1.7) (2017-08-23)
[Full Changelog](https://github.com/cad/ovpm/compare/v0.1.6...v0.1.7)

## [v0.1.6](https://github.com/cad/ovpm/tree/v0.1.6) (2017-08-23)
[Full Changelog](https://github.com/cad/ovpm/compare/v0.1.5...v0.1.6)

## [v0.1.5](https://github.com/cad/ovpm/tree/v0.1.5) (2017-08-23)
[Full Changelog](https://github.com/cad/ovpm/compare/v0.1.4...v0.1.5)

## [v0.1.4](https://github.com/cad/ovpm/tree/v0.1.4) (2017-08-23)
[Full Changelog](https://github.com/cad/ovpm/compare/v0.1.3...v0.1.4)

**Implemented enhancements:**

- make rpm package [\#24](https://github.com/cad/ovpm/issues/24)

**Fixed bugs:**

- stop ovpmd systemd unit upon removal [\#26](https://github.com/cad/ovpm/issues/26)
- ensure nat after openvpn process restart [\#25](https://github.com/cad/ovpm/issues/25)

## [v0.1.3](https://github.com/cad/ovpm/tree/v0.1.3) (2017-08-22)
[Full Changelog](https://github.com/cad/ovpm/compare/v0.1.2...v0.1.3)

**Implemented enhancements:**

- add edit user command  [\#23](https://github.com/cad/ovpm/issues/23)
- give user's a static vpn ip addr [\#17](https://github.com/cad/ovpm/issues/17)
- show user's vpn ip addr in the cli output [\#16](https://github.com/cad/ovpm/issues/16)
- don't push vpn server as the default gateway for some users [\#15](https://github.com/cad/ovpm/issues/15)
- fix user password storage [\#2](https://github.com/cad/ovpm/issues/2)

**Fixed bugs:**

- when ovpm is freshly installed and initialized \(and applied\)OpenVPN process is not started [\#19](https://github.com/cad/ovpm/issues/19)
- OpenVPN clients whose version is 2.3 and below is complaining about certificate verification [\#14](https://github.com/cad/ovpm/issues/14)

## [v0.1.2](https://github.com/cad/ovpm/tree/v0.1.2) (2017-08-15)
[Full Changelog](https://github.com/cad/ovpm/compare/v0.1.1...v0.1.2)

## [v0.1.1](https://github.com/cad/ovpm/tree/v0.1.1) (2017-08-15)
[Full Changelog](https://github.com/cad/ovpm/compare/v0.1.0...v0.1.1)

**Implemented enhancements:**

- hook up iptables to give nat masquerading [\#13](https://github.com/cad/ovpm/issues/13)

## [v0.1.0](https://github.com/cad/ovpm/tree/v0.1.0) (2017-08-13)
**Implemented enhancements:**

- handle sigint \(Ctrl-C\) [\#10](https://github.com/cad/ovpm/issues/10)
- supervise openvpn process [\#9](https://github.com/cad/ovpm/issues/9)
- implement remote control proto [\#8](https://github.com/cad/ovpm/issues/8)
- write docs [\#4](https://github.com/cad/ovpm/issues/4)
- write unit tests [\#3](https://github.com/cad/ovpm/issues/3)

