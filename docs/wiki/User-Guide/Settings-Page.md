# Settings Page

[Home](../Home.md) | Related: [Configuration](../Operations/Configuration.md), [Capabilities](../Security/Capabilities.md), [Administration Page](Administration-Page.md), [Agent Mutation Controls](../Security/Agent-Mutation-Controls.md)

## What Problem This Solves

The Settings page summarizes policy boundaries for operators without exposing editable privileged configuration.

## How It Works

It presents read-only summaries for session policy, magic token lifetime, process visibility, service allowlists, capabilities, identity administration, and the agent boundary.

## What Protects It

Configuration changes remain outside the UI and require operator access to system files and service restart.

## What Can Fail

Displayed policy can be misunderstood if it is not compared with actual configuration files.

## How To Diagnose It

Compare the page with `/etc/koji/koji.yaml` and `/etc/koji/agent.yaml`.

## How To Recover

Update configuration through normal operational change control, then restart affected services.
