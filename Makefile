.PHONY: deps mocks test build dist clean

PACKAGES := $(shell go list ./... | grep -v /mock)
BUILD_VERSION := $(shell git describe --tags)

build: deps clean
	go build -v

deps:
	go get -insecure gopkg.in/yaml.v2 github.com/aws/aws-sdk-go/aws/session github.com/aws/aws-sdk-go/service/sts github.com/spf13/cobra github.com/hashicorp/go-getter

test: deps
	go test -race -cover $(PACKAGES)

dist: deps
	echo building ${BUILD_VERSION}
	gox -osarch="darwin/amd64" -osarch="linux/386" -osarch="linux/amd64" -osarch="windows/amd64" \
		-ldflags "-X main.version=${BUILD_VERSION}" -output "dist/ncd_{{.OS}}_{{.Arch}}"

ghr:
	go get -u github.com/tcnksm/ghr

prerelease: ghr dist
	ghr --prerelease -u ${CIRCLE_PROJECT_USERNAME} -r ${CIRCLE_PROJECT_REPONAME} --replace `git describe --tags` dist/

release: ghr dist
	ghr -u ${CIRCLE_PROJECT_USERNAME} -r ${CIRCLE_PROJECT_REPONAME} --replace `git describe --tags` dist/

clean:
	rm -f fargate-create
	rm -rf iac
	rm -rf fargate-create-template

install: build
	cp -p fargate-create /usr/local/bin/
