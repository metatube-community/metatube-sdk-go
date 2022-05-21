MODULE := github.com/javtube/javtube-sdk-go

SERVER_NAME := javtube-server
SERVER_CODE := cmd/server/main.go

BUILD_DIR     := build
BUILD_TAGS    :=
BUILD_FLAGS   := -v
BUILD_COMMIT  := $(shell git rev-parse --short HEAD)
BUILD_VERSION := $(shell git describe --abbrev=0 --tags HEAD)

CGO_ENABLED := 0
GO111MODULE := on

LDFLAGS += -w -s -buildid=
LDFLAGS += -X "$(MODULE)/internal/constant.Version=$(BUILD_VERSION)"
LDFLAGS += -X "$(MODULE)/internal/constant.GitCommit=$(BUILD_COMMIT)"

GO_BUILD = GO111MODULE=$(GO111MODULE) CGO_ENABLED=$(CGO_ENABLED) \
	go build $(BUILD_FLAGS) -ldflags '$(LDFLAGS)' -tags '$(BUILD_TAGS)' -trimpath

all: server

server:
	$(GO_BUILD) -o $(BUILD_DIR)/$(SERVER_NAME) $(SERVER_CODE)

lint:
	golangci-lint run --disable-all -E govet -E gofumpt -E megacheck ./...

clean:
	rm -rf $(BUILD_DIR)
