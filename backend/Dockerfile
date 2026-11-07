# ---- build ----
FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/creme .

# ---- run ----
FROM alpine:3.21
# tzdata : fuseau horaire des logs ; su-exec : l'entrypoint passe de root à PUID:PGID
RUN apk add --no-cache tzdata su-exec && mkdir -p /data
COPY --from=build /out/creme /usr/local/bin/creme
COPY docker-entrypoint.sh /usr/local/bin/docker-entrypoint.sh
RUN chmod +x /usr/local/bin/docker-entrypoint.sh
VOLUME ["/data"]
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s \
  CMD wget -qO- http://127.0.0.1:8080/api/health || exit 1
ENTRYPOINT ["docker-entrypoint.sh"]
