FROM golang:1.24-alpine AS build

WORKDIR /src
COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend/cmd/ ./cmd/
COPY backend/internal/ ./internal/
RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -ldflags="-s -w" \
    -o /out/judcalc-api \
    ./cmd/api

FROM python:3.13-alpine AS runtime

RUN addgroup -S app \
    && adduser -S -G app app \
    && mkdir -p /app/data \
    && chown app:app /app/data

WORKDIR /app
COPY --from=build --chown=app:app /out/judcalc-api ./judcalc-api
COPY --chown=app:app python/__init__.py ./python/__init__.py
COPY --chown=app:app python/database/ ./python/database/
COPY --chown=app:app python/data/oab-sp.json ./python/data/oab-sp.json
COPY --chown=app:app python/scripts/ ./python/scripts/
COPY --chown=app:app backend/docker-entrypoint.sh ./docker-entrypoint.sh

USER app

ENV HTTP_ADDR=:8080 \
    DATABASE_PATH=/app/data/app.db \
    GIN_MODE=release

EXPOSE 8080
VOLUME ["/app/data"]

ENTRYPOINT ["/app/docker-entrypoint.sh"]
