FROM --platform=$BUILDPLATFORM golang:1.22.7-alpine3.20 AS builder

ARG VERSION=v0.0.0

RUN apk add --no-cache make gcc musl-dev linux-headers git jq bash

# build op-node with the shared go.mod & go.sum files
COPY ./tw-node /app/tw-node
COPY ./tw-service /app/tw-service
COPY ./common /app/common
COPY ./tw-supervisor /app/tw-supervisor
COPY ./tw-deployer /app/tw-deployer

COPY ./go.mod /app/go.mod
COPY ./go.sum /app/go.sum

WORKDIR /app/tw-node

RUN go mod tidy

ARG TARGETOS TARGETARCH

RUN go build -o bin/tw-node ./cmd/main.go

FROM alpine:3.21

COPY --from=builder /app/tw-node/bin/tw-node /usr/local/bin

CMD ["tw-node"]
