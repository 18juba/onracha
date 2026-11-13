FROM golang:1.27.3-alpine AS builder

WORKDIR /src

COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend/ ./
COPY docs/openapi.yaml /docs/openapi.yaml

RUN mkdir -p /out \
    && CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/onracha .

FROM alpine:3.24

RUN apk add --no-cache ca-certificates \
    && addgroup -S -g 10001 onracha \
    && adduser -S -D -H -u 10001 -G onracha onracha \
    && mkdir -p /uploads /docs \
    && chown -R 10001:10001 /uploads /docs

COPY --from=builder --chown=10001:10001 /out/onracha /usr/local/bin/onracha
COPY --from=builder --chown=10001:10001 /docs/openapi.yaml /docs/openapi.yaml

ENV OPENAPI_SPEC_PATH=/docs/openapi.yaml
ENV PHOTO_STORAGE_DIR=/uploads

USER 10001:10001

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
    CMD wget -q -O /dev/null http://127.0.0.1:8080/health || exit 1

ENTRYPOINT ["/usr/local/bin/onracha"]
