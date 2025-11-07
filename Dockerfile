# This as of this update will pick up 1.24.9 - consider adding sha for security
FROM golang:1.25 AS build
WORKDIR /build/src

# Splitting this makes it a cached layer of your dependencies - it's optional
COPY go.mod go.sum .
RUN go mod download

# Copy the (rest of the) source tree
COPY . .

# Real stuff - statically linked, stripped, pure go build
# Assumes what you build is in the top level, like it should,
# don't go and add cmd/foo/main.go until you really have more
# than 1 binary and even then there is probably a main one
# having cmd/foo also makes `go install github.com/yourname/yourepo@latest`
# not work (or be longer than necessary) when you don't put your main at the top,
# but if not and if you must replace . (current package) by ./cmd/foo/ next line:
RUN CGO_ENABLED=0 go build -trimpath -ldflags=-s -o kaedama .

# This is the important bit, making for a final image with just your binary:
FROM scratch
COPY --from=build /build/src/kaedama /usr/bin/kaedama
# Not needed anymore, see below why/how
# COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
ENTRYPOINT ["/usr/bin/kaedama"]
