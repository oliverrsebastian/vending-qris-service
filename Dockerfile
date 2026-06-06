# Build
FROM golang:1.23-alpine AS build
RUN apk add --no-cache git ca-certificates
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server

# Run (works on VPS / any Linux amd64 host with container runtime)
FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=build /out/server .
COPY configs ./configs
EXPOSE 8080
ENV APP_ENV=prod
ENV PORT=8080
USER nobody
ENTRYPOINT ["/app/server"]
