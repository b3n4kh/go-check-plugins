VERSION = 0.34.1
CURRENT_REVISION = $(shell git rev-parse --short HEAD)
ifeq ($(OS),Windows_NT)
GOPATH_ROOT:=$(shell cygpath ${GOPATH})
else
GOPATH_ROOT:=${GOPATH}
endif

export GO111MODULE=on

.PHONY: all
all: clean testconvention test build rpm deb

.PHONY: test
test: lint
	go test $(TESTFLAGS) ./...

.PHONY: devel-deps
devel-deps:
	cd && go get golang.org/x/lint/golint  \
	  github.com/pierrre/gotestcover \
	  github.com/mattn/goveralls

.PHONY: check-release-deps
check-release-deps:
	@have_error=0; \
	for command in cpanm hub ghch gobump; do \
	  if ! command -v $$command > /dev/null; then \
	    have_error=1; \
	    echo "\`$$command\` command is required for releasing"; \
	  fi; \
	done; \
	test $$have_error = 0

.PHONY: lint
lint: devel-deps
	go vet ./...
	golint -set_exit_status ./...

.PHONY: testconvention
testconvention:
	prove -r t/
	@go generate ./... && git diff --exit-code -- . ':(exclude)go.*' || (echo 'please `go generate ./...` and commit them' && false)

.PHONY: cover
cover: devel-deps
	gotestcover -v -short -covermode=count -coverprofile=.profile.cov -parallelpackages=4 ./...

.PHONY: build
build:
	mkdir -p build
	for i in $(filter-out check-windows-%, $(wildcard check-*)); do \
	  go build -ldflags "-s -w" -o build/$$i \
	  `pwd | sed -e "s|${GOPATH_ROOT}/src/||"`/$$i; \
	done

build/mackerel-check:
	mkdir -p build
	go build -ldflags="-s -w -X main.gitcommit=$(CURRENT_REVISION)" \
	  -o build/mackerel-check

.PHONY: rpm
rpm: rpm-v1 rpm-v2

.PHONY: rpm-v1
rpm-v1:
	make build GOOS=linux GOARCH=386
	rpmbuild --define "_sourcedir `pwd`"  --define "_version ${VERSION}" --define "buildarch noarch" -bb packaging/rpm/mackerel-check-plugins.spec
	make build GOOS=linux GOARCH=amd64
	rpmbuild --define "_sourcedir `pwd`"  --define "_version ${VERSION}" --define "buildarch x86_64" -bb packaging/rpm/mackerel-check-plugins.spec

.PHONY: rpm-v2
rpm-v2:
	make build/mackerel-check GOOS=linux GOARCH=amd64
	rpmbuild --define "_sourcedir `pwd`"  --define "_version ${VERSION}" \
	  --define "buildarch x86_64" --define "dist .el7.centos" \
	  -bb packaging/rpm/mackerel-check-plugins-v2.spec
	rpmbuild --define "_sourcedir `pwd`"  --define "_version ${VERSION}" \
	  --define "buildarch x86_64" --define "dist .amzn2" \
	  -bb packaging/rpm/mackerel-check-plugins-v2.spec

.PHONY: deb
deb: deb-v1 deb-v2

.PHONY: deb-v1
deb-v1:
	make build GOOS=linux GOARCH=386
	for i in `cat packaging/deb/debian/source/include-binaries`; do \
	  cp build/`basename $$i` packaging/deb/debian/; \
	done
	cd packaging/deb && debuild --no-tgz-check -rfakeroot -uc -us

.PHONY: deb-v2
deb-v2:
	make build/mackerel-check GOOS=linux GOARCH=amd64
	cp build/mackerel-check packaging/deb-v2/debian/
	cd packaging/deb-v2 && debuild --no-tgz-check -rfakeroot -uc -us

.PHONY: release
release: check-release-deps
	(cd tool && cpanm -qn --installdeps .)
	perl tool/create-release-pullrequest

.PHONY: clean
clean:
	if [ -d build ]; then \
	  rm -f build/check-*; \
	  rmdir build; \
	fi
	go clean

.PHONY: update
update:
	go get -u ./...
	go mod tidy
