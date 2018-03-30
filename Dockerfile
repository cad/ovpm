FROM fedora:latest
LABEL maintainer="Mustafa Arici (mustafa@arici.io)"

# Deps
RUN rpm --import https://mirror.go-repo.io/fedora/RPM-GPG-KEY-GO-REPO
RUN curl -s https://mirror.go-repo.io/fedora/go-repo.repo | tee /etc/yum.repos.d/go-repo.repo
RUN yum -y install golang ruby ruby-devel gcc make redhat-rpm-config git rpm-build rpmdevtools createrepo reprepro npm
RUN gem install fpm

VOLUME /fs/src/github.com/cad/ovpm

ENV DIR="/fs/src/github.com/cad/ovpm"
ENV RELEASEDIR=$DIR/release
ENV UNITDIR="/usr/lib/systemd/system/"
ENV GOPATH="/fs/"

WORKDIR /fs/src/github.com/cad/ovpm

CMD ["./build.sh"]
