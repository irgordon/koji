# Processes Page

[Home](../Home.md) | Related: [Configuration](../Operations/Configuration.md), [Capabilities](../Security/Capabilities.md)

## What Problem This Solves

The Processes page shows a redaction-aware process view without exposing full host process metadata by default.

## How It Works

Koji applies process visibility policy before returning rows. Summary mode hides sensitive fields such as command line or memory details unless configured otherwise.

## What Protects It

Process reads require `host.processes.read`. The backend enforces `process_visibility_mode`, `include_command_line`, and `max_processes`.

## What Can Fail

Fields can show "Hidden by policy" or the view can be denied by missing capability.

## How To Diagnose It

Read the policy notice and compare configuration with [Configuration Reference](../Reference/Configuration-Reference.md).

## How To Recover

Adjust policy only when the operational need justifies more exposure.
