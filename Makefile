help: ## show help message
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m\033[0m\n"} /^[$$()% a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

pre-commit: lint test govuln

govuln: ## run govulncheck to scan for known vulnerabilities in Go or imported packages
	go install golang.org/x/vuln/cmd/govulncheck@latest
	govulncheck ./...

lint: ## run linters using golangci-lint configuration file (default lookup)
	golangci-lint cache clean
	golangci-lint run -v  ./...

test: ## run tests, report code coverage (sans mocks) and check for race conditions
	go test -race ./... -covermode=atomic -coverprofile=coverage.out.tmp
	cat coverage.out.tmp | grep -v "/mocks" > coverage.out
	go tool cover -func=coverage.out

run: ## run the app
	CGO_ENABLED=0 go run cmd/reddit/main.go
