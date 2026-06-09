PREFIX ?= /usr
SYSCONFDIR ?= /etc
LOCALSTATEDIR ?= /var/lib
SYSTEMDUNITDIR ?= $(PREFIX)/lib/systemd/system
DESTDIR ?= build/rootfs
GO ?= go
NPM ?= npm

.PHONY: fmt test build build-web build-kojid build-agent lint install package clean

fmt:
	$(GO) fmt ./...

test:
	$(GO) test ./...

lint:
	$(GO) vet ./...

build: build-web build-kojid build-agent

build-web:
	$(NPM) --prefix web run build

build-kojid:
	install -d build/bin
	$(GO) build -o build/bin/kojid ./cmd/kojid

build-agent:
	install -d build/bin
	$(GO) build -o build/bin/koji-agent ./cmd/koji-agent

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

clean:
	rm -rf build/bin build/rootfs
