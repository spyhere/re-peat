MODULE := $(shell go list -m)
APP := repeat
BIN := bin

TAG := $(shell git describe --tags --abbrev=0 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE := $(shell date -u +%Y-%m-%dT%H:%m:%s%Z)

GOFLAGS := -trimpath
LDFLAGS := -X main.tag=$(TAG) \
					 -X main.commit=$(COMMIT) \
					 -X main.date=$(DATE) \
					 -s -w

APP_BUNDLE := $(BIN)/$(APP).app
CONTENTS := Contents
MACOS := MacOS
RESOURCES := Resources

run:
	go run .
debug:
	go run -tags=debug .
check-heap:
	go run -gcflags="-m -m" . 2>&1 | grep $(MODULE)
build:
	@echo "Building MacOS GUI"
	@mkdir -p $(APP_BUNDLE)/$(CONTENTS)/$(MACOS)
	@mkdir -p $(APP_BUNDLE)/$(CONTENTS)/$(RESOURCES)
	GOOD=darwin GOARCH=arm64 CGO_ENABLED=1 \
	go build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(APP_BUNDLE)/$(CONTENTS)/$(MACOS)/$(APP) .
	cp Info.plist $(APP_BUNDLE)/$(CONTENTS)/
	@echo "Successfully created"
build-darwin:
	@echo "Building MacOS GUI"
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=1 \
	go build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o /tmp/$(APP)-amd64 .
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=1 \
	go build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o /tmp/$(APP)-arm64 .

	@mkdir -p $(APP_BUNDLE)/$(CONTENTS)/$(MACOS)
	@mkdir -p $(APP_BUNDLE)/$(CONTENTS)/$(RESOURCES)
	lipo -create \
		/tmp/$(APP)-arm64 \
		/tmp/$(APP)-amd64 \
		-output $(APP_BUNDLE)/$(CONTENTS)/$(MACOS)/$(APP)

	cp Info.plist $(APP_BUNDLE)/$(CONTENTS)/
	@echo "Successfully created"
build-windows:
	@echo "Building Windows GUI"
	@mkdir -p $(BIN)
	GOOS=windows GOARCH=amd64 CGO_ENABLED=1 \
	go build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BIN)/$(APP)-windows-amd64.exe .
	@echo "Successfully created"
clean:
	rm -rf $(BIN)/
clean-all: clean
	go clean -cache -modcache -testcache

