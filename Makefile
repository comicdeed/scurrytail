APP_NAME := dmesg-watch
VERSION ?= $(shell git describe --tags --always 2>/dev/null || echo "v0.1.0")
GO ?= $(shell which go 2>/dev/null || echo /home/linuxbrew/.linuxbrew/opt/go/bin/go)

PREFIX ?= $(HOME)/.local
SYSTEMD_USER_DIR ?= $(HOME)/.config/systemd/user

LDFLAGS := -ldflags "-X main.version=$(VERSION) -s -w"

.PHONY: all build test vet clean install uninstall service-enable service-disable service-status dist

all: build

build:
	@mkdir -p bin
	CGO_ENABLED=0 $(GO) build $(LDFLAGS) -o bin/$(APP_NAME) ./cmd/$(APP_NAME)

test:
	$(GO) test -v -race ./...

vet:
	$(GO) vet ./...

install: build
	@mkdir -p $(PREFIX)/bin
	@mkdir -p $(SYSTEMD_USER_DIR)
	install -m 755 bin/$(APP_NAME) $(PREFIX)/bin/$(APP_NAME)
	sed 's|ExecStart=/usr/bin/$(APP_NAME)|ExecStart=$(PREFIX)/bin/$(APP_NAME)|' systemd/$(APP_NAME).service > $(SYSTEMD_USER_DIR)/$(APP_NAME).service
	chmod 644 $(SYSTEMD_USER_DIR)/$(APP_NAME).service
	systemctl --user daemon-reload || true
	@echo "Installed $(APP_NAME) to $(PREFIX)/bin/$(APP_NAME)"
	@echo "Installed systemd service to $(SYSTEMD_USER_DIR)/$(APP_NAME).service"

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
	@mkdir -p dist/$(APP_NAME)-$(VERSION)-linux-amd64
	cp bin/$(APP_NAME) dist/$(APP_NAME)-$(VERSION)-linux-amd64/
	cp systemd/$(APP_NAME).service dist/$(APP_NAME)-$(VERSION)-linux-amd64/
	cp scripts/install.sh dist/$(APP_NAME)-$(VERSION)-linux-amd64/
	cp README.md dist/$(APP_NAME)-$(VERSION)-linux-amd64/
	tar -czf dist/$(APP_NAME)-$(VERSION)-linux-amd64.tar.gz -C dist $(APP_NAME)-$(VERSION)-linux-amd64
	@rm -rf dist/$(APP_NAME)-$(VERSION)-linux-amd64
	@echo "Created dist/$(APP_NAME)-$(VERSION)-linux-amd64.tar.gz"

clean:
	rm -rf bin/ dist/
