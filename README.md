# Go Micro

A production-minded Go microservices portfolio project: an API gateway (broker) fans out to authentication, logging, and mail services over **HTTP**, **net/rpc**, **gRPC**, and **RabbitMQ**.

Originally based on a Udemy course stack; this repo has been refactored for consistent style, shared packages, env-driven config, tests, and CI/CD.

## Architecture

```text
Browser / Postman
        │
        ▼
   front-end (:8088)  ──►  broker-service (:8080)
                              │
          ┌───────────────────┼───────────────────┐
          ▼                   ▼                   ▼
 authentication (:8081)   logger-service      mail-service (:8082)
          │              HTTP / RPC / gRPC          │
          ▼                   ▲                   ▼
       Postgres            MongoDB              Mailhog
                              ▲
                     listener-service ◄── RabbitMQ (logs_topic)
```

| Service | Role |
|---------|------|
| **broker-service** | Edge API (`/handle`, `/log-grpc`) |
| **authentication-service** | Email/password auth against Postgres |
| **logger-service** | Persist logs (HTTP + RPC `:5001` + gRPC `:50001`) |
| **mail-service** | SMTP via Mailhog |
| **listener-service** | Consumes RabbitMQ topic → logger HTTP |
| **front-end** | API Console playground |

Shared modules live under [`pkg/`](pkg/) and [`contracts/`](contracts/) (single `logs.proto`).

## Quick start

**Requirements:** Docker + Docker Compose, Go 1.27+ (for local builds/tests).

```bash
cp project/.env.example project/.env
make up
```

- API Console: http://localhost:8088  
- Broker: http://localhost:8080  
- Mailhog UI: http://localhost:8025

Demo credentials: `admin@example.com` / `verysecret`

> Postgres init scripts only run on a **fresh** volume. If auth fails after an upgrade, delete `project/db-data/postgres` and `make up` again.

## Log transports

`POST /handle` with `"action":"log"` accepts optional `"transport"`:

| Value | Path |
|-------|------|
| `rpc` (default) | `net/rpc` → logger `:5001` |
| `http` | HTTP → logger `/log` |
| `grpc` | gRPC → logger `:50001` |
| `rabbit` | Publish to `logs_topic` → listener → logger |

The API Console and [Postman collection](docs/postman/go-micro.postman_collection.json) expose every path.

## Configuration

All runtime knobs are environment variables. See [`project/.env.example`](project/.env.example).

Notable variables: `DSN`, `RABBITMQ_URL`, `MONGO_*`, `AUTH_URL`, `LOGGER_*`, `MAIL_*`, `BROKER_URL`, `TEMPLATES_DIR`, `LOG_LEVEL`, `SHUTDOWN_TIMEOUT`.

## Local development

```bash
make tidy    # go mod tidy across modules
make test    # unit tests
make lint    # go vet
make build   # binaries → bin/
```

Workspace entrypoint: [`go.work`](go.work).

## CI/CD

- **CI** ([`.github/workflows/ci.yml`](.github/workflows/ci.yml)): vet, test, Docker build matrix  
- **CD** ([`.github/workflows/cd.yml`](.github/workflows/cd.yml)): push images to `ghcr.io/<owner>/go-micro-<service>`

## Postman

Import:

1. [`docs/postman/go-micro.postman_collection.json`](docs/postman/go-micro.postman_collection.json)
2. [`docs/postman/go-micro.postman_environment.json`](docs/postman/go-micro.postman_environment.json)

## Intentional local-demo choices

No JWT/OAuth rewrite, mTLS, or full OpenTelemetry - the goal is a clean, showcaseable multi-protocol microservices demo that still looks production-minded (graceful shutdown, health probes, structured logs, shared packages, CI).

## License

Use freely for portfolio / learning purposes.
