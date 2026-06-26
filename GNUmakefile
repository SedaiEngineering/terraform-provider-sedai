default: fmt lint install generate

build:
	go build -v ./...

install: build
	go install -v ./...

lint:
	golangci-lint run

generate:
	cd tools; go generate ./...

fmt:
	gofmt -s -w -e .

test:
	go test -v -cover -timeout=120s -parallel=10 ./...

testacc:
	TF_ACC=1 go test -v -cover -timeout 120m ./...

# Run unit tests only (no TF_ACC, no real backend needed)
test-unit:
	go test -v -count=1 -timeout=120s ./internal/provider/... -run "^Test[^A][^c][^c]"

# Run all acceptance tests against a live Sedai backend (requires SEDAI_BASE_URL + SEDAI_API_TOKEN)
test-acc:
	TF_ACC=1 go test -v -count=1 -timeout=90m -parallel=4 \
	  ./internal/provider/... -run "^TestAcc" 2>&1 | tee test-results.jsonl; \
	  exit $${PIPESTATUS[0]}

# Run acceptance tests and emit go test -json output for the report generator
test-acc-json:
	TF_ACC=1 go test -v -count=1 -timeout=90m -parallel=4 -json \
	  ./internal/provider/... -run "^TestAcc" > test-results.json 2>&1

# Run system-level tests (scale + dependency tests) — requires TF_SYSTEM_TESTS=1
test-system:
	TF_ACC=1 TF_SYSTEM_TESTS=1 go test -v -count=1 -timeout=120m -parallel=1 \
	  ./internal/provider/... -run "^TestAcc(SCALE|DEP)" 2>&1 | tee system-results.jsonl; \
	  exit $${PIPESTATUS[0]}

# Generate HTML test report from test-results.json
test-report:
	go run ./scripts/report/main.go test-results.json > test-report.html
	@echo "Report written to test-report.html"

# Run both unit and acceptance tests
test-all: test-unit test-acc

# Run a specific manifest test by ID, e.g.: make test-id ID=GSET-006
test-id:
	TF_ACC=1 go test -v -count=1 -timeout=30m ./internal/provider/... \
	  -run "TestAcc.*/$(ID)$$"

.PHONY: fmt lint test testacc build install generate \
        test-unit test-acc test-acc-json test-system test-report test-all test-id
