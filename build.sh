#!/bin/bash
set -ex

echo "travis build no: $TRAVIS_BUILD_NUMBER"
echo "travis tag: $TRAVIS_TAG"
echo "travis go version: $TRAVIS_GO_VERSION"

# deps
rpm --import https://mirror.go-repo.io/fedora/RPM-GPG-KEY-GO-REPO
curl -s https://mirror.go-repo.io/fedora/go-repo.repo | tee /etc/yum.repos.d/go-repo.repo
yum -y install golang ruby ruby-devel gcc make redhat-rpm-config git rpm-build rpmdevtools
gem install fpm

# prep
export DIR="/fs/src/github.com/cad/ovpm"
export UNITDIR="/usr/lib/systemd/system/"
export GOPATH="/fs/"
#export PATH=":$PATH"
mkdir -p $DIR/build/
mkdir -p $DIR/rpm/
rm -rf $DIR/build/*
rm -rf $DIR/rpm/*
mkdir -p $DIR/build/usr/sbin/
mkdir -p $DIR/build/usr/bin/
mkdir -p $DIR/build/var/db/ovpm
mkdir -p $DIR/build/$UNITDIR
go get -v -t -d ./...

#build
#install
GOOS=linux  go build  -o $DIR/build/usr/sbin/ovpmd ./cmd/ovpmd
GOOS=linux  go build  -o $DIR/build/usr/bin/ovpm   ./cmd/ovpm
cp $DIR/contrib/systemd/ovpmd.service $DIR/build/$UNITDIR

#package
fpm -s dir -t rpm -n ovpm --version `git name-rev --tags --name-only $(git rev-parse HEAD) | cut -d 'v' -f 2` --iteration $TRAVIS_BUILD_NUMBER --depends openvpn --description "OVPM makes all aspects of OpenVPN server administration a breeze." --after-install $DIR/contrib/afterinstall.sh --before-remove $DIR/contrib/beforeremove.sh -p $DIR/rpm -C $DIR/build .
