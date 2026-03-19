.PHONY: build test-build test test-clean test-shell

# Local build
build:
	go build -o clinic .

# Build the test container
test-build:
	docker build -f Dockerfile.test -t clinic-test .

# Run tests interactively in a fresh container (destroyed on exit)
test-shell:
	docker run --rm -it clinic-test /bin/bash

# Run a specific clinic command in a fresh container
test-run:
	docker run --rm clinic-test clinic $(CMD)

# Quick smoke test — version, stacks, list
test-smoke:
	docker run --rm clinic-test sh -c '\
		echo "=== version ===" && clinic version && \
		echo "=== stacks ===" && clinic stacks && \
		echo "=== list --all ===" && clinic list --all'

# Test registry — install every tool, verify version commands and auth commands
# Usage: make test-registry
#        make test-registry TOOLS="gh stripe jq"
test-registry:
	docker run --rm clinic-test /build/scripts/test-registry.sh $(TOOLS)

# Clean up test images
test-clean:
	docker rmi clinic-test 2>/dev/null || true
