PACKAGE_NAME := ysf/raftsample

# compress golang binary size: https://blog.filippo.io/shrink-your-go-binaries-with-this-one-weird-trick/
build: install
	ls -al
	rm -rf artifacts
	GO111MODULE=on CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -mod=mod -a -installsuffix cgo -o artifacts/raftsample $(PACKAGE_NAME)/cmd/api
	GO111MODULE=on CGO_ENABLED=0 GOOS=darwin go build -ldflags="-s -w" -mod=mod -a -installsuffix cgo -o artifacts/raftsample.osx $(PACKAGE_NAME)/cmd/api
	GO111MODULE=on CGO_ENABLED=0 GOOS=windows go build -ldflags="-s -w" -mod=mod -a -installsuffix cgo -o artifacts/raftsample.exe $(PACKAGE_NAME)/cmd/api

install:
	go mod tidy
