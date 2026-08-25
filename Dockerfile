# -----------------------------------------------------------------------------
# Stage 1: Go Backend Build
# -----------------------------------------------------------------------------
FROM golang:1.26-alpine AS backend

RUN apk add --no-cache gcc musl-dev sqlite-dev

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 go build -o /bin/server ./cmd/server

# -----------------------------------------------------------------------------
# Stage 2: Final Runtime Image
# -----------------------------------------------------------------------------
FROM alpine:3.20

# tzdata: IANA-таймзоны для time.LoadLocation (Европа/Москва и др.)
RUN apk add --no-cache ca-certificates sqlite-libs tzdata

WORKDIR /app

# Бинарник, конфиг и статика
COPY --from=backend /bin/server /app/server
COPY --from=backend /app/config.yaml /app/config.yaml
COPY --from=backend /app/web /app/web

# Рабочая директория для SQLite
RUN mkdir -p /app/data

# Порт HTTP-сервера (переопределяется через docker run -e PORT=9000)
ENV PORT=8080

EXPOSE ${PORT}

HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:${PORT}/healthz || exit 1

CMD ["/app/server"]
