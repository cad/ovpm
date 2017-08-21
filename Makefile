.PHONY: build
docker-build:
	docker run --rm -i -t -v `pwd`:/fs/src/github.com/cad/ovpm -w /fs/src/github.com/cad/ovpm fedora ./build.sh
	#docker run --rm -i -t -v `pwd`:/fs/src/github.com/cad/ovpm -w /fs/src/github.com/cad/ovpm fedora /bin/bash
