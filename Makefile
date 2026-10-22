APP_NAME := dmesg-watch
VERSION ?= $(shell git describe --tags --always 2>/dev/null || echo "v0.1.0")

# Go Toolchain (native binary with automatic Docker fallback)
GO_BIN ?= $(shell which go 2>/dev/null || ( [ -x $(HOME)/.local/share/brew/bin/go ] && echo $(HOME)/.local/share/brew/bin/go ))
GOLANG_IMAGE ?= golang:alpine
ifeq ($(GO_BIN),)
GO ?= docker run --rm -v $(CURDIR):$(CURDIR) -w $(CURDIR) -e CGO_ENABLED=0 $(GOLANG_IMAGE) go
GOFMT ?= docker run --rm -v $(CURDIR):$(CURDIR) -w $(CURDIR) $(GOLANG_IMAGE) gofmt
TEST_FLAGS ?= -v
else
GO ?= $(GO_BIN)
GOFMT ?= $(shell which gofmt 2>/dev/null || echo gofmt)
TEST_FLAGS ?= -v -race
endif

PREFIX ?= $(HOME)/.local
SYSTEMD_USER_DIR ?= $(HOME)/.config/systemd/user

LDFLAGS := -ldflags "-X main.version=$(VERSION) -s -w"

.PHONY: all build fmt test vet clean install uninstall service-enable service-disable service-status dist nix-build nix-check nix-fmt

all: build

build:
	@mkdir -p bin
	CGO_ENABLED=0 $(GO) build $(LDFLAGS) -o bin/$(APP_NAME) ./cmd/$(APP_NAME)

fmt:
	$(GOFMT) -w -s .

test:
	$(GO) test $(TEST_FLAGS) ./...

vet:
	$(GO) vet ./...

install: build
	@mkdir -p $(PREFIX)/bin
	@mkdir -p $(SYSTEMD_USER_DIR)
	install -m 755 bin/$(APP_NAME) $(PREFIX)/bin/$(APP_NAME)
	sed 's|ExecStart=/usr/bin/$(APP_NAME)|ExecStart=$(PREFIX)/bin/$(APP_NAME)|' systemd/$(APP_NAME).service > $(SYSTEMD_USER_DIR)/$(APP_NAME).service
	chmod 644 $(SYSTEMD_USER_DIR)/$(APP_NAME).service
	systemctl --user daemon-reload || true
	systemctl --user enable --now $(APP_NAME).service || true
	@echo "Installed $(APP_NAME) to $(PREFIX)/bin/$(APP_NAME)"
	@echo "Enabled and started $(APP_NAME).service in user session"

uninstall:
	systemctl --user stop $(APP_NAME).service 2>/dev/null || true
	systemctl --user disable $(APP_NAME).service 2>/dev/null || true
	rm -f $(PREFIX)/bin/$(APP_NAME)
	rm -f $(SYSTEMD_USER_DIR)/$(APP_NAME).service
	systemctl --user daemon-reload || true
	@echo "Uninstalled $(APP_NAME)"

service-enable:
	systemctl --user daemon-reload
	systemctl --user enable --now $(APP_NAME).service

service-disable:
	systemctl --user disable --now $(APP_NAME).service

service-status:
	systemctl --user status $(APP_NAME).service

dist: build
	@mkdir -p dist/$(APP_NAME)-$(VERSION)-amd64
	cp bin/$(APP_NAME) dist/$(APP_NAME)-$(VERSION)-amd64/
	cp systemd/$(APP_NAME).service dist/$(APP_NAME)-$(VERSION)-amd64/
	cp systemd/50-$(APP_NAME).preset dist/$(APP_NAME)-$(VERSION)-amd64/
	cp scripts/install.sh dist/$(APP_NAME)-$(VERSION)-amd64/
	cp README.md dist/$(APP_NAME)-$(VERSION)-amd64/
	tar -czf dist/$(APP_NAME)-$(VERSION)-amd64.tar.gz -C dist $(APP_NAME)-$(VERSION)-amd64
	@rm -rf dist/$(APP_NAME)-$(VERSION)-amd64
	@echo "Created dist/$(APP_NAME)-$(VERSION)-amd64.tar.gz"

clean:
	rm -rf bin/ dist/

# Nix Toolchain (native binary with automatic Docker fallback)
NIX_BIN ?= $(shell which nix 2>/dev/null)
ifeq ($(NIX_BIN),)
NIX ?= docker run --rm -v $(CURDIR):$(CURDIR) -w $(CURDIR) nixos/nix:latest nix --extra-experimental-features "nix-command flakes"
else
NIX ?= $(NIX_BIN) --extra-experimental-features "nix-command flakes"
endif

nix-build:
	$(NIX) build path:.

nix-check:
	$(NIX) flake check path:. --no-build

nix-fmt:
	$(NIX) run nixpkgs#nixpkgs-fmt -- flake.nix nix/*.nix

