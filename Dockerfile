FROM golang:1.26.5 AS build
WORKDIR /build/src

COPY go.mod go.sum .
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -trimpath -ldflags=-s -o kaedama .

FROM scratch
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /build/src/kaedama /usr/bin/kaedama

ENTRYPOINT ["/usr/bin/kaedama"]
