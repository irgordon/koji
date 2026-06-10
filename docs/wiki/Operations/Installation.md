# Installation

[Home](../Home.md) | Related: [Packaging and Deployment](../Architecture/Packaging-and-Deployment.md), [Configuration](Configuration.md), [Release Operations](Release-Operations.md)

## What Problem This Solves

Installation places Koji into stable runtime paths so it does not depend on a source checkout.

## How It Works

Use release artifacts or `make install` into a staging root. Runtime files are installed under `/usr/bin`, `/etc/koji`, `/var/lib/koji`, `/run/koji`, and `/usr/share/koji`.

## What Protects It

The packaged rootfs includes systemd units, example configs, static assets, and binaries built by CI. Checksums verify artifact integrity.

## What Can Fail

Files can be missing, executable bits can be lost, ownership can be wrong, or systemd paths can point outside the runtime layout.

## How To Diagnose It

Run `make verify-release`, verify checksums, inspect `/usr/share/koji/dist`, and check systemd unit paths.

## How To Recover

Reinstall from the verified artifact, restore file modes, and correct `/var/lib/koji` and `/run/koji` ownership.
