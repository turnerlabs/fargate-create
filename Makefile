.PHONY: mocks test build dist clean

PACKAGES := $(shell go list ./... | grep -v /mock)
BUILD_VERSION := $(shell git describe --tags)
AWS_DEFAULT_REGION := us-east-1

test:
	go test -race -cover $(PACKAGES)

build:
	make clean
	go build -v

dist:
	echo building ${BUILD_VERSION}
	GOOS=linux GOARCH=386 go build -ldflags "-X main.version=${BUILD_VERSION}" -o dist/ncd_linux_386
	GOOS=linux GOARCH=amd64 go build -ldflags "-X main.version=${BUILD_VERSION}" -o dist/ncd_linux_amd64
	GOOS=linux GOARCH=arm64 go build -ldflags "-X main.version=${BUILD_VERSION}" -o dist/ncd_linux_arm64
	GOOS=darwin GOARCH=amd64 go build -ldflags "-X main.version=${BUILD_VERSION}" -o dist/ncd_darwin_amd64
	GOOS=darwin GOARCH=arm64 go build -ldflags "-X main.version=${BUILD_VERSION}" -o dist/ncd_darwin_arm64
	GOOS=windows GOARCH=amd64 go build -ldflags "-X main.version=${BUILD_VERSION}" -o dist/ncd_windows_amd64.exe

prerelease:
	gh release create ${BUILD_VERSION} --generate-notes --prerelease dist/*
	aws s3 cp dist/ s3://get-fargate-create.turnerlabs.io/${BUILD_VERSION}/ --recursive --region ${AWS_DEFAULT_REGION}
	echo ${BUILD_VERSION} > develop && aws s3 cp ./develop s3://get-fargate-create.turnerlabs.io/ --region ${AWS_DEFAULT_REGION}

release:
	gh release create ${BUILD_VERSION} --generate-notes dist/*
	aws s3 cp dist/ s3://get-fargate-create.turnerlabs.io/${BUILD_VERSION}/ --recursive --region ${AWS_DEFAULT_REGION}
	echo ${BUILD_VERSION} > master && aws s3 cp ./master s3://get-fargate-create.turnerlabs.io/ --region ${AWS_DEFAULT_REGION}

clean:
	rm -f fargate-create
	rm -rf iac
	rm -rf fargate-create-template
	rm -rf dist