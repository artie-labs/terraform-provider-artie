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

.PHONY: prune
prune:
	@current=$$(git branch --show-current) && \
	if [ -z "$$current" ]; then \
		echo "Error: in detached HEAD state — refusing to prune (no branch is protected)."; \
		exit 1; \
	fi && \
	branches=$$(git branch | grep -v "^\* " | sed 's/^  //') && \
	if [ -z "$$branches" ]; then \
		echo "No other local branches to delete."; \
	else \
		echo "This will delete the following local branches:" && \
		echo "$$branches" && \
		echo "" && \
		echo "Current branch '$$current' will be kept." && \
		echo "" && \
		printf "Are you sure? (y/N) " && \
		read confirm && \
		if [ "$$confirm" = "y" ] || [ "$$confirm" = "Y" ]; then \
			echo "$$branches" | xargs git branch -d; \
		else \
			echo "Aborted."; \
		fi \
	fi
