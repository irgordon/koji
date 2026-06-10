# Packaging and Deployment

[Home](../Home.md) | Related: [Installation](../Operations/Installation.md), [Configuration](../Operations/Configuration.md), [Release Architecture](Release-Architecture.md)

## What Problem This Solves

Packaging makes Koji runnable outside the source tree with stable Linux runtime paths.

## How It Works

Release artifacts include Linux binaries, a rootfs tarball, example configs, systemd units, and checksums. Runtime paths use `/usr/bin`, `/etc/koji`, `/var/lib/koji`, `/run/koji`, and `/usr/share/koji`.

## What Protects It

The install layout avoids developer-local paths. Systemd units point at installed paths. Static assets are served from configured absolute paths in production.

## What Can Fail

Ownership can be wrong, config files can be missing, static assets can be absent, or the DB directory can be unwritable.

## How To Diagnose It

Run `make verify-release`, inspect systemd unit paths, check `/readyz`, and verify `/usr/share/koji/dist`.

## How To Recover

Reinstall from the rootfs artifact, correct directory ownership, restore example configs, and restart services.
