#!/bin/bash
set -ex

## After Docker
echo "travis build no: $TRAVIS_BUILD_NUMBER"
echo "travis tag: $TRAVIS_TAG"
echo "travis go version: $TRAVIS_GO_VERSION"

export RELEASEVER=${TRAVIS_BUILD_NUMBER:-"1"}
echo "releasever: $RELEASEVER"

export VERSION="0.0"
export LOCAL_GIT_TAG=`git name-rev --tags --name-only $(git rev-parse HEAD) | cut -d 'v' -f 2`
if [ "$LOCAL_GIT_TAG" != "undefined" ]; then
    export VERSION=$LOCAL_GIT_TAG
fi
echo "Version is $VERSION"

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
fpm -s dir -t rpm -n ovpm --version $VERSION  --iteration $RELEASEVER --depends openvpn --description "OVPM makes all aspects of OpenVPN server administration a breeze." --before-install $DIR/contrib/beforeinstall.sh --after-install $DIR/contrib/afterinstall.sh --before-remove $DIR/contrib/beforeremove.sh --after-upgrade $DIR/contrib/afterupgrade.sh -p $RELEASEDIR/rpm -C $RELEASEDIR/build .

fpm -s dir -t deb -n ovpm --version $VERSION --iteration $RELEASEVER --depends openvpn --description "OVPM makes all aspects of OpenVPN server administration a breeze." --before-install $DIR/contrib/beforeinstall.sh --after-install $DIR/contrib/afterinstall.sh --before-remove $DIR/contrib/beforeremove.sh --after-upgrade $DIR/contrib/afterupgrade.sh -p $RELEASEDIR/deb -C $RELEASEDIR/build .

#create rpm repo
createrepo --database $RELEASEDIR/rpm

#create deb repo
reprepro -b $RELEASEDIR/deb/ includedeb ovpm $RELEASEDIR/deb/*.deb

# clean
rm -rf $RELEASEDIR/build
echo "packages are ready at ./deb/ and ./rpm/"
