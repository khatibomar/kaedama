FROM golang:1.27.3 AS build
WORKDIR /build/src

COPY go.mod go.sum .
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -trimpath -ldflags=-s -o kaedama .

FROM scratch
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /build/src/kaedama /usr/bin/kaedama

HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
  CMD ["/usr/bin/kaedama", "healthcheck"]

ENTRYPOINT ["/usr/bin/kaedama"]
