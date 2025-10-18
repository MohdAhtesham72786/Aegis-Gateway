# Aegis Gateway - Coding Test

This is a small Go service that implements a reverse-proxy gateway enforcing policy-as-code for agent->tool calls, with mock adapters and telemetry.

Quick start

1. Install Go 1.20+
2. From project root run:

   go run ./cmd/aegis

3) Run demo scripts in PowerShell:

   ./scripts/run.ps1

Docker (if you don't have Go locally)

1. Build and run the gateway image:

   docker build -t aegis:local .
   docker run --rm -p 8080:8080 aegis:local

Or use docker-compose from project root:

   docker-compose up --build

Deploy compose (collector + Jaeger)

There is a `deploy/docker-compose.yml` that brings up the gateway image, an OpenTelemetry Collector, and Jaeger UI. From the `deploy/` folder run:

   docker-compose up --build

Then open Jaeger UI at http://localhost:16686 to view traces.

Adapters in deploy compose

The `deploy/docker-compose.yml` builds and runs the payments and files adapters as separate containers (binaries are built from `./cmd/payments` and `./cmd/files`).

Tests

If you have Go locally you can run unit tests for policy evaluation:

  go test ./internal/policy

Development shortcuts

If you have Make and Docker installed you can:

   make docker-build
   make docker-run

Or run the full deploy stack and demo from PowerShell:

   ./scripts/run-container.ps1

If you edit dependencies

   go mod tidy

Logging

Logs are written to stdout and rotated files under `./logs/aegis.log` (rotation via lumberjack).

Layout
- `cmd/aegis` — main binary
- `internal/gateway` — HTTP gateway and routing
- `internal/policy` — policy loader and evaluator (hot-reloads policies/)
- `internal/adapters/payments` — mock payments service (in-process)
- `internal/adapters/files` — mock files service (in-process)
- `pkg/telemetry` — simple OTel console exporter
- `policies/` — sample policies
- `scripts/` — demo scripts

Policies
Sample policies are in `policies/finance.yaml` and `policies/hr.yaml`.

Demo
See `scripts/demo.ps1` for the four demo cases.

Notes
- Telemetry uses console OTel exporter. Replace with OTLP exporter for real collector.
- Policies hot-reload: the loader watches `./policies` and logs errors for invalid files without crashing.
