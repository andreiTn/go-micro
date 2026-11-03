# Go Micro

[![CI](https://github.com/andreiTn/go-micro/actions/workflows/ci.yml/badge.svg)](https://github.com/andreiTn/go-micro/actions/workflows/ci.yml)

I started this from a Udemy Go microservices course.

This repo is that project after a cleanup pass: shared packages, config from env, proper Docker setup, tests, and CI. The interesting part of the course is still here — one broker talking to other services over **HTTP**, **net/rpc**, **gRPC**, and **RabbitMQ**.

## How it fits together

Everything goes through the **broker** on port `8080`. The little **API Console** (`:8088`) is just a front-end that hits the broker so you can click around without Postman.

- **Auth** checks email/password against Postgres, then tells the logger someone logged in.
- **Logger** writes to Mongo. You can reach it three ways from the broker (plain HTTP, old-school RPC, or gRPC), or publish to RabbitMQ and let the **listener** forward the message.
- **Mail** sends through Mailhog so you can peek at messages in the browser (`:8025`) instead of a real inbox.

Shared stuff lives in [`pkg/`](pkg/) (JSON helpers, config, Rabbit dial, HTTP server). The gRPC contract is a single proto under [`contracts/`](contracts/).

| Service | What it does |
|---------|----------------|
| broker-service | Gateway — `/handle` and `/log-grpc` |
| authentication-service | Login against Postgres |
| logger-service | Store logs (HTTP + RPC `:5001` + gRPC `:50001`) |
| mail-service | SMTP via Mailhog |
| listener-service | RabbitMQ consumer → logger |
| front-end | API Console |

## Run it

Needs Docker Compose. Go 1.27+ if you want to build/test outside containers.

```bash
cp project/.env.example project/.env
docker compose -f project/docker-compose.yml --env-file project/.env up --build -d
```

(If you have Make: `make up`.)

Then open:

- Console → http://localhost:8088
- Broker → http://localhost:8080
- Mailhog → http://localhost:8025

Demo login: `admin@example.com` / `verysecret`

Postgres only seeds on a fresh volume. If login fails after you've already run things once, wipe `project/db-data/postgres` and bring the stack up again.

## Trying the log transports

On `POST /handle` with `"action": "log"`, set `"transport"` to one of:

| `transport` | What happens |
|-------------|--------------|
| `rpc` (default) | broker → logger over `net/rpc` |
| `http` | broker → logger `/log` |
| `grpc` | broker → logger gRPC |
| `rabbit` | broker publishes to `logs_topic`, listener picks it up |

Same flows are in the console and in [`docs/postman/`](docs/postman/).

## Config

Env vars only — see [`project/.env.example`](project/.env.example). The usual ones: `DSN`, `RABBITMQ_URL`, `MONGO_*`, service URLs, `LOG_LEVEL`, `SHUTDOWN_TIMEOUT`.

## Dev commands

```bash
make tidy
make test
make lint
make build
```

On Windows without Make, run the `go test` / `go vet` bits per module, or the compose command above.

## CI

Pushes and PRs run vet + tests + a Docker build for each service (`.github/workflows/ci.yml`). Tags / `main` can push images to GHCR via the CD workflow.

## What's intentionally not here

No JWT rewrite, no mTLS, no full observability stack. It's still a local demo — just one I'd actually show someone, with graceful shutdown, health checks, structured logs, and CI instead of course leftovers.
