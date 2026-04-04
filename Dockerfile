FROM --platform=$BUILDPLATFORM golang:1.26-alpine AS builder

WORKDIR /app

RUN apk add --no-cache gcc musl-dev libc-dev

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . .

RUN --mount=type=cache,target=/go/pkg/mod \
    CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o dataset-cli .

FROM alpine:3.19

RUN apk add --no-cache ca-certificates postgresql-client

WORKDIR /app

COPY --from=builder /app/dataset-cli .

RUN mkdir -p /root/.dataset-cli

ENV PATH="/app:${PATH}"

ENTRYPOINT ["/app/dataset-cli"]
