.PHONY: lint
lint:
	golangci-lint run -c .golangci.yml

.PHONY: install-tools
install-tools:
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sudo sh -s -- -b $(go env GOPATH)/bin

#-- Go
.PHONY: run
run:
	go run ./...

.PHONY: test
test:
	go test -v ./...

#-- Docker
.PHONY: docker-up
docker-up:
	docker compose up --build -d

docker-down:
	docker compose down
