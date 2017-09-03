#!/bin/bash
set -ex

echo "travis build no: $TRAVIS_BUILD_NUMBER"
echo "travis tag: $TRAVIS_TAG"
echo "travis go version: $TRAVIS_GO_VERSION"

# deps
rpm --import https://mirror.go-repo.io/fedora/RPM-GPG-KEY-GO-REPO
curl -s https://mirror.go-repo.io/fedora/go-repo.repo | tee /etc/yum.repos.d/go-repo.repo
yum -y install golang ruby ruby-devel gcc make redhat-rpm-config git rpm-build rpmdevtools createrepo reprepro
gem install fpm

# prep
export DIR="/fs/src/github.com/cad/ovpm"
export RELEASEDIR=$DIR/release
export UNITDIR="/usr/lib/systemd/system/"
export GOPATH="/fs/"
export RELEASEVER=${TRAVIS_BUILD_NUMBER:-"1"}
echo "releasever: $RELEASEVER"
mkdir -p $RELEASEDIR/
mkdir -p $RELEASEDIR/build/
mkdir -p $RELEASEDIR/rpm/
mkdir -p $RELEASEDIR/deb/
rm -rf $RELEASEDIR/build/*
rm -rf $RELEASEDIR/rpm/*
rm -rf $RELEASEDIR/deb/*
mkdir -p $RELEASEDIR/build/usr/sbin/
mkdir -p $RELEASEDIR/build/usr/bin/
mkdir -p $RELEASEDIR/build/var/db/ovpm
mkdir -p $RELEASEDIR/build/$UNITDIR
mkdir -p $RELEASEDIR/deb/conf
go get -v -t -d ./...

#build
#install
GOOS=linux  go build  -o $RELEASEDIR/build/usr/sbin/ovpmd ./cmd/ovpmd
GOOS=linux  go build  -o $RELEASEDIR/build/usr/bin/ovpm   ./cmd/ovpm
cp $DIR/contrib/systemd/ovpmd.service $RELEASEDIR/build/$UNITDIR
cp $DIR/contrib/yumrepo.repo $RELEASEDIR/rpm/ovpm.repo
cp $DIR/contrib/deb-repo-config $RELEASEDIR/deb/conf/distributions

#package
fpm -s dir -t rpm -n ovpm --version `git name-rev --tags --name-only $(git rev-parse HEAD) | cut -d 'v' -f 2` --iteration $RELEASEVER --depends openvpn --description "OVPM makes all aspects of OpenVPN server administration a breeze." --after-install $DIR/contrib/afterinstall.sh --before-remove $DIR/contrib/beforeremove.sh --after-upgrade $DIR/contrib/afterupgrade.sh -p $RELEASEDIR/rpm -C $RELEASEDIR/build .

fpm -s dir -t deb -n ovpm --version `git name-rev --tags --name-only $(git rev-parse HEAD) | cut -d 'v' -f 2` --iteration $RELEASEVER --depends openvpn --description "OVPM makes all aspects of OpenVPN server administration a breeze." --after-install $DIR/contrib/afterinstall.sh --before-remove $DIR/contrib/beforeremove.sh --after-upgrade $DIR/contrib/afterupgrade.sh -p $RELEASEDIR/deb -C $RELEASEDIR/build .

#create rpm repo
createrepo --database $RELEASEDIR/rpm

#create deb repo
reprepro -b $RELEASEDIR/deb/ includedeb ovpm $RELEASEDIR/deb/*.deb

# clean
rm -rf $RELEASEDIR/build
