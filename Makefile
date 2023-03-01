GO               = go
M                = $(shell printf "\033[34;1m>>\033[0m")

# Check richgo does exist.
ifeq (, $(shell which richgo))
$(warning "could not find richgo in $(PATH), run: go get github.com/kyoh86/richgo")
endif

.PHONY: test sync codecov test-app

.PHONY: default
default: all

.PHONY: all
all: build test

.PHONY: test
test: sync
	$(info $(M) running tests)
	 go test -race ./...

.PHONY: coverage
coverage: sync
	$(info $(M) running tests coverage)
	 go test -coverprofile=c.out ./...;\
  	 go tool cover -func=c.out;\
  	 rm c.out



.PHONY: build
build:
	$(info $(M) build and compile)
	 go build ./...

.PHONY: sync
sync: 
	$(info $(M) downloading dependencies)
	go get -v ./...

.PHONY: fmt
fmt:
	$(info $(M) format code)
	@ret=0 && for d in $$($(GO) list -f '{{.Dir}}' ./... | grep -v /vendor/); do \
		$(GO) fmt $$d/*.go || ret=$$? ; \
		done ; exit $$ret

.PHONY: lint
lint: ## Run linters
	$(info $(M) running golangci linter)
	golangci-lint run --timeout 5m0s ./...

