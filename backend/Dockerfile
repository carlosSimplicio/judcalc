# Build from the repository root:
# docker build -f backend/Dockerfile -t judcalc-backend .

FROM python:3.13-alpine AS database

WORKDIR /src
COPY python/ ./python/
RUN python -m python.scripts.init_database --output /tmp/app.db

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

FROM alpine:3.22 AS runtime

RUN apk add --no-cache ca-certificates \
    && addgroup -S app \
    && adduser -S -G app app \
    && mkdir -p /app/data \
    && chown app:app /app/data

WORKDIR /app
COPY --from=build --chown=app:app /out/judcalc-api ./judcalc-api
COPY --from=database --chown=app:app /tmp/app.db ./data/app.db

USER app

ENV HTTP_ADDR=:8080 \
    DATABASE_PATH=/app/data/app.db

EXPOSE 8080
VOLUME ["/app/data"]

ENTRYPOINT ["/app/judcalc-api"]
