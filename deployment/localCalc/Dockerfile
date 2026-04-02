# Build from repository root (order-pack-calculator):
#   docker compose -f deployment/localCalc/docker-compose.yml build
FROM golang:1.24-bookworm AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /server ./cmd/calculator

# Alpine + wget so Docker HEALTHCHECK can hit /ready (distroless has no shell/curl).
FROM alpine:3.21
RUN apk add --no-cache ca-certificates wget \
	&& addgroup -g 65532 -S nonroot \
	&& adduser -S -u 65532 -G nonroot nonroot
COPY --from=build /server /server
EXPOSE 8080
USER nonroot
ENTRYPOINT ["/server"]
