FROM golang:1.22 AS builder
WORKDIR /build

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/root/.cache/go-build go mod download
COPY . .
RUN --mount=type=cache,target=/root/.cache/go-build CGO_ENABLED=0 go build -o ./bin/fake ./entrypoint/fake/main.go

FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

COPY --from=builder ./build/bin/fake /fake
COPY --from=builder /var/run /var/run
