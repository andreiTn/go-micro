FROM golang:1.27.3-alpine AS builder
WORKDIR /src
RUN apk add --no-cache git ca-certificates
COPY pkg ./pkg
COPY mail-service ./mail-service
WORKDIR /src/mail-service
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/mail-service ./cmd/api

FROM alpine:3.20
RUN adduser -D -H -u 10001 appuser \
  && apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /out/mail-service /app/mail-service
COPY --from=builder /src/mail-service/templates /templates
ENV TEMPLATES_DIR=/templates
USER appuser
EXPOSE 80
ENTRYPOINT ["/app/mail-service"]
