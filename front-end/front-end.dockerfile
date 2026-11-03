FROM golang:1.27.3-alpine AS builder
WORKDIR /src
RUN apk add --no-cache git ca-certificates
COPY pkg ./pkg
COPY front-end ./front-end
WORKDIR /src/front-end
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/front-end ./cmd/web

FROM alpine:3.20
RUN adduser -D -H -u 10001 appuser \
  && apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /out/front-end /app/front-end
COPY --from=builder /src/front-end/cmd/web/templates /app/templates
COPY --from=builder /src/front-end/cmd/web/static /app/static
ENV TEMPLATES_DIR=/app/templates
ENV STATIC_DIR=/app/static
USER appuser
EXPOSE 80
ENTRYPOINT ["/app/front-end"]
