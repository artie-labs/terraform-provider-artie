default: testacc

# Run acceptance tests
.PHONY: testacc
testacc:
	TF_ACC=1 go test ./... -v $(TESTARGS) -timeout 120m

.PHONY: generate
generate:
	go generate ./...

.PHONY: upgrade
upgrade:
	go get github.com/artie-labs/transfer
	go mod tidy
	echo "Upgrade complete"
