default: testacc

# Run acceptance tests
.PHONY: testacc
testacc:
	TF_ACC=1 go test ./... -v $(TESTARGS) -timeout 120m

.PHONY: generate
generate:
	go generate ./...

install:
	go install ./...

.PHONY: lint-fix
lint-fix:
	golangci-lint run --fix
