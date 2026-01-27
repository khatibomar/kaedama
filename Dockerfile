FROM golang:1.25 AS build
WORKDIR /build/src

COPY go.mod go.sum .
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -trimpath -ldflags=-s -o kaedama .

FROM alpine:latest
RUN apk add --no-cache ca-certificates wget
COPY --from=build /build/src/kaedama /usr/bin/kaedama

ENTRYPOINT ["/usr/bin/kaedama"]
