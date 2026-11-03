FROM golang:1.27.3-alpine AS builder
WORKDIR /src
RUN apk add --no-cache git ca-certificates
COPY pkg ./pkg
COPY contracts ./contracts
COPY logger-service ./logger-service
WORKDIR /src/logger-service
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/logger-service ./cmd/api

FROM alpine:3.20
RUN adduser -D -H -u 10001 appuser \
  && apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /out/logger-service /app/logger-service
USER appuser
EXPOSE 80 5001 50001
ENTRYPOINT ["/app/logger-service"]
