.PHONY: mocks test build dist clean

PACKAGES := $(shell go list ./... | grep -v /mock)
BUILD_VERSION := $(shell git describe --tags)

test:
	go test -race -cover $(PACKAGES)

build:
	make clean
	git checkout .gitignore .dockerignore
	go build -v

dist:
	echo building ${BUILD_VERSION}
	gox -osarch="darwin/amd64" -osarch="linux/386" -osarch="linux/amd64" -osarch="windows/amd64" \
		-ldflags "-X main.version=${BUILD_VERSION}" -output "dist/ncd_{{.OS}}_{{.Arch}}"

prerelease:
	ghr --prerelease -u ${CIRCLE_PROJECT_USERNAME} -r ${CIRCLE_PROJECT_REPONAME} --replace `git describe --tags` dist/

release:
	ghr -u ${CIRCLE_PROJECT_USERNAME} -r ${CIRCLE_PROJECT_REPONAME} --replace `git describe --tags` dist/

clean:
	rm -rf iac
	rm -rf fargate-create-template