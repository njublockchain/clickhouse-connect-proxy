# builder stage
FROM golang:1.21 AS build-stage

WORKDIR /build

COPY go.mod go.sum ./
COPY . ./

RUN GOPROXY=https://goproxy.io,direct CGO_ENABLED=0 GOOS=linux go build ./cmd/proxy

# release stage
FROM debian:12 AS build-release-stage

WORKDIR /app

COPY --from=build-stage /build/proxy /app/proxy

ENTRYPOINT ["./proxy"]
