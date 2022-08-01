APPNAME = opennoise_daemon
RELEASE_DIR = release

VERSION = $(shell git describe --tags --dirty)
DATE = $(shell date +%Y%m%d)
BUILD_PATH = $(RELEASE_DIR)/$(APPNAME)-$(VERSION)

LDFLAGS = "-X main.version=${VERSION}_${DATE}"
GO_FILES = $(shell find -L . -type f -name '*.go')
MAIN_FILE = cmd/main.go
# Right now, no cross compile because sqlite has native bindings

.PHONY: build clean

build: $(GO_FILES)
	mkdir -p $(BUILD_PATH)
	go build -ldflags $(LDFLAGS) -o $(BUILD_PATH) $(MAIN_FILE)

clean:
	rm -rf $(RELEASE_DIR)

