.PHONY: mocks test build dist clean

PACKAGES := $(shell go list ./... | grep -v /mock)
BUILD_VERSION := $(shell git describe --tags)

test:
	go test -race -cover $(PACKAGES)

build:
	make clean
	go build -v

dist:
	echo building ${BUILD_VERSION}
	gox -osarch="darwin/amd64" -osarch="darwin/arm64" -osarch="linux/386" -osarch="linux/amd64" -osarch="windows/amd64" \
		-ldflags "-X main.version=${BUILD_VERSION}" -output "dist/ncd_{{.OS}}_{{.Arch}}"

prerelease:
	ghr --prerelease -u ${CIRCLE_PROJECT_USERNAME} -r ${CIRCLE_PROJECT_REPONAME} --replace `git describe --tags` dist/
	aws s3 cp dist/ s3://get-fargate-create.turnerlabs.io/${BUILD_VERSION}/ --recursive
	echo ${BUILD_VERSION} > develop && aws s3 cp ./develop s3://get-fargate-create.turnerlabs.io/

release:
	ghr -u ${CIRCLE_PROJECT_USERNAME} -r ${CIRCLE_PROJECT_REPONAME} --replace `git describe --tags` dist/
	aws s3 cp dist/ s3://get-fargate-create.turnerlabs.io/${BUILD_VERSION}/ --recursive
	echo ${BUILD_VERSION} > master && aws s3 cp ./master s3://get-fargate-create.turnerlabs.io/

clean:
	rm fargate-create
	rm -rf iac
	rm -rf fargate-create-template