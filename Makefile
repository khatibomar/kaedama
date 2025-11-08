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

.PHONY: docker-down
docker-down:
	docker compose down

#-- Curl
.PHONY: call-health
call-health:
	curl -v localhost:4140/health

.PHONY: call-proxy
call-proxy:
	curl "http://0.0.0.0:4140/proxy?url=https://lightningspark77.pro/_v7/1ab5d45273a9183bebb58eb74d5722d8ea6384f350caf008f08cf018f1f0566d0cb82a2a799830d1af97cd3f4b6a9a81ef3aed2fb783292b1abcf1b8560a1d1aa308008b88420298522a9f761e5aa1024fbe74e5aa853cfc933cd1219327d1232e91847a185021b184c027f97ae732b36d3983eb284b42a76caee7186508aade/master.m3u8"
