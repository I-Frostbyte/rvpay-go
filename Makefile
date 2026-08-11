.PHONY: test bench-test

test:
	@echo "Running tests in all subdirectories..."
	@find . -name "*_test.go" -exec dirname {} \; | sort | uniq | xargs -I {} go test {}

bench-test:
	@echo "Running benchmarks in all subdirectories..."
	@find . -name "*_test.go" -exec dirname {} \; | sort | uniq | xargs -I {} go test -bench=. {}

generate:
	go generate ./...