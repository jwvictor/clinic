.PHONY: build test-build test test-clean test-shell

# Local build
build:
	go build -o cliq .

# Build the test container
test-build:
	docker build -f Dockerfile.test -t cliq-test .

# Run tests interactively in a fresh container (destroyed on exit)
test-shell:
	docker run --rm -it cliq-test /bin/bash

# Run a specific cliq command in a fresh container
test-run:
	docker run --rm cliq-test cliq $(CMD)

# Quick smoke test — version, stacks, list
test-smoke:
	docker run --rm cliq-test sh -c '\
		echo "=== version ===" && cliq version && \
		echo "=== stacks ===" && cliq stacks && \
		echo "=== list --all ===" && cliq list --all'

# Clean up test images
test-clean:
	docker rmi cliq-test 2>/dev/null || true
