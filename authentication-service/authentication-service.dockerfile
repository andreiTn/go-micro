FROM golang:1.27.3-alpine AS builder
WORKDIR /src
RUN apk add --no-cache git ca-certificates
COPY pkg ./pkg
COPY authentication-service ./authentication-service
WORKDIR /src/authentication-service
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/authentication-service ./cmd/api

FROM alpine:3.20
RUN adduser -D -H -u 10001 appuser \
  && apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /out/authentication-service /app/authentication-service
USER appuser
EXPOSE 80
ENTRYPOINT ["/app/authentication-service"]
