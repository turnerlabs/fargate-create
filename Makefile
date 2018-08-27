.PHONY: deps mocks test build dist clean

PACKAGES := $(shell go list ./... | grep -v /mock)
BUILD_VERSION := $(shell git describe --tags)

build: deps
	make clean
	go build -v

deps:
	go get -insecure gopkg.in/yaml.v2

test: deps
	go test -race -cover $(PACKAGES)

dist: deps
	echo building ${BUILD_VERSION}
	gox -osarch="darwin/amd64" -osarch="linux/386" -osarch="linux/amd64" -osarch="windows/amd64" \
		-ldflags "-X main.version=${BUILD_VERSION}" -output "dist/ncd_{{.OS}}_{{.Arch}}"

prerelease: build
	ghr --prerelease -u ${CIRCLE_PROJECT_USERNAME} -r ${CIRCLE_PROJECT_REPONAME} --replace `git describe --tags` dist/

release: build
	ghr -u ${CIRCLE_PROJECT_USERNAME} -r ${CIRCLE_PROJECT_REPONAME} --replace `git describe --tags` dist/

clean:
	rm fargate-create
	rm -rf iac
	rm -rf fargate-create-template
