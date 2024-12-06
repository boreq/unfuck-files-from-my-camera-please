BUILD_DIRECTORY=_build
PROGRAM_NAME=unfuck-names-of-files-from-my-camera-please

.PHONY: all
all: test lint build

.PHONY: ci
ci: tools dependencies generate fmt check-repository-unchanged test lint build

.PHONY: fmt
fmt:
	gosimports -l -w ./

.PHONY: build-directory
build-directory:
	mkdir -p ./${BUILD_DIRECTORY}

.PHONY: build
build: build-directory
	go build -o ./${BUILD_DIRECTORY}/${PROGRAM_NAME} ./cmd/${PROGRAM_NAME}

.PHONY: check-repository-unchanged
check-repository-unchanged:
	./_tools/check_repository_unchanged.sh

.PHONY: tools
tools:
	 go install honnef.co/go/tools/cmd/staticcheck@latest
	 go install github.com/google/wire/cmd/wire@latest
	 go install github.com/rinchsan/gosimports/cmd/gosimports@v0.3.5 # https://github.com/golang/go/issues/20818

.PHONY: lint
lint: 
	go vet ./...
	staticcheck ./...

.PHONY: test
test:
	go test ./...

.PHONY: test-verbose
test-verbose:
	go test -v ./...

.PHONY: clean
clean:
	rm -rf ./${BUILD_DIRECTORY}
