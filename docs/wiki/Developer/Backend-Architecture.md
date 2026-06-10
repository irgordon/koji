# Backend Architecture

[Home](../Home.md) | Related: [Request Flow](../Architecture/Request-Flow.md), [Agent Architecture](../Architecture/Agent-Architecture.md), [Database Schema](Database-Schema.md)

## What Problem This Solves

The backend enforces governance before exposing host data or accepting operational intent.

## How It Works

HTTP handlers coordinate auth, capabilities, audit, policy checks, and store calls. Jobs persist intent. The worker advances approved jobs. The agent owns future privileged mutation. The platform command runner owns direct command execution.

## What Protects It

Package boundaries prevent direct `systemctl` execution from HTTP and keep mutation agent-owned.

## What Can Fail

New code can accidentally bypass capability checks, audit, jobs, or command ownership.

## How To Diagnose It

Run Go tests and ownership scans for direct command execution.

## How To Recover

Refactor the behavior into the correct package and add regression tests around the failed boundary.
