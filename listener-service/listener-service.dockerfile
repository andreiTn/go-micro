FROM golang:1.27.3-alpine AS builder
WORKDIR /src
RUN apk add --no-cache git ca-certificates
COPY pkg ./pkg
COPY listener-service ./listener-service
WORKDIR /src/listener-service
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/listener-service .

FROM alpine:3.20
RUN adduser -D -H -u 10001 appuser \
  && apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /out/listener-service /app/listener-service
USER appuser
ENTRYPOINT ["/app/listener-service"]
