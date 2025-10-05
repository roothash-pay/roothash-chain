FROM --platform=$BUILDPLATFORM golang:1.22.7-alpine3.20 AS builder

ARG VERSION=v0.0.0

RUN apk add --no-cache make gcc musl-dev linux-headers git jq bash

# build op-node with the shared go.mod & go.sum files
COPY ./rhs-node /app/rhs-node
COPY ./rhs-service /app/rhs-service
COPY ./common /app/common
COPY ./rhs-supervisor /app/rhs-supervisor
COPY ./rhs-deployer /app/rhs-deployer

COPY ./go.mod /app/go.mod
COPY ./go.sum /app/go.sum

WORKDIR /app/rhs-node

RUN go mod tidy

ARG TARGETOS TARGETARCH

RUN go build -o bin/rhs-node ./cmd/main.go

FROM alpine:3.21

COPY --from=builder /app/rhs-node/bin/rhs-node /usr/local/bin

CMD ["rhs-node"]
