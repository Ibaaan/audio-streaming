# The checks CI runs. `make check` runs them all; integration needs Docker,
# ffmpeg, the compose services and the .env variables loaded.

GOVULNCHECK := go run golang.org/x/vuln/cmd/govulncheck@latest

.PHONY: check tidy lint test vuln integration

check: tidy lint test vuln

tidy:
	go mod tidy -diff

lint:
	golangci-lint run --build-tags integration ./...

test:
	go test -race ./...

vuln:
	$(GOVULNCHECK) ./...

integration:
	go test -race -tags integration ./...
