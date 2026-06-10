PREFIX ?= /usr
SYSCONFDIR ?= /etc
LOCALSTATEDIR ?= /var/lib
SYSTEMDUNITDIR ?= $(PREFIX)/lib/systemd/system
DESTDIR ?= build/rootfs
GO ?= go
NPM ?= npm
RELEASE_DIR ?= build/release
GOOS ?= linux
GOARCH ?= amd64
GO_BUILD_FLAGS ?= -trimpath

.PHONY: fmt test build build-web build-kojid build-agent lint openapi verify-openapi backup restore verify-restore pre-upgrade-check verify-upgrade install package release checksums verify-release clean

fmt:
	$(GO) fmt ./...

test:
	$(GO) test ./...

lint:
	$(GO) vet ./...

openapi:
	node packaging/scripts/generate_openapi_docs.mjs

verify-openapi:
	packaging/scripts/verify_openapi.sh

backup:
	packaging/scripts/backup.sh

restore:
	packaging/scripts/restore.sh $(BACKUP)

verify-restore:
	packaging/scripts/verify_restore.sh

pre-upgrade-check:
	packaging/scripts/pre_upgrade_check.sh

verify-upgrade:
	packaging/scripts/verify_upgrade.sh

build: build-web build-kojid build-agent

build-web:
	$(NPM) --prefix web run build

build-kojid:
	install -d build/bin
	$(GO) build $(GO_BUILD_FLAGS) -o build/bin/kojid ./cmd/kojid

build-agent:
	install -d build/bin
	$(GO) build $(GO_BUILD_FLAGS) -o build/bin/koji-agent ./cmd/koji-agent

install: build
	install -d $(DESTDIR)$(PREFIX)/bin
	install -m 0755 build/bin/kojid $(DESTDIR)$(PREFIX)/bin/kojid
	install -m 0755 build/bin/koji-agent $(DESTDIR)$(PREFIX)/bin/koji-agent
	install -d $(DESTDIR)$(PREFIX)/share/koji
	rm -rf $(DESTDIR)$(PREFIX)/share/koji/dist
	cp -R dist $(DESTDIR)$(PREFIX)/share/koji/dist
	install -d $(DESTDIR)$(SYSCONFDIR)/koji
	install -m 0640 packaging/examples/koji.yaml $(DESTDIR)$(SYSCONFDIR)/koji/koji.yaml
	install -m 0640 packaging/examples/agent.yaml $(DESTDIR)$(SYSCONFDIR)/koji/agent.yaml
	install -d $(DESTDIR)$(SYSTEMDUNITDIR)
	install -m 0644 packaging/systemd/kojid.service $(DESTDIR)$(SYSTEMDUNITDIR)/kojid.service
	install -m 0644 packaging/systemd/koji-agent.service $(DESTDIR)$(SYSTEMDUNITDIR)/koji-agent.service
	install -d -m 0750 $(DESTDIR)$(LOCALSTATEDIR)/koji

package: install

release: install
	rm -rf $(RELEASE_DIR)
	install -d $(RELEASE_DIR)
	install -m 0755 build/bin/kojid $(RELEASE_DIR)/kojid-$(GOOS)-$(GOARCH)
	install -m 0755 build/bin/koji-agent $(RELEASE_DIR)/koji-agent-$(GOOS)-$(GOARCH)
	tar -czf $(RELEASE_DIR)/koji-rootfs-$(GOOS)-$(GOARCH).tar.gz -C build rootfs
	$(MAKE) checksums

checksums:
	packaging/scripts/checksums.sh $(RELEASE_DIR)

verify-release:
	packaging/scripts/verify_release.sh $(RELEASE_DIR) $(DESTDIR)

clean:
	rm -rf build/bin build/rootfs $(RELEASE_DIR)
