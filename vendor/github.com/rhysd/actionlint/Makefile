SRCS := $(filter-out %_test.go, $(wildcard *.go cmd/actionlint/*.go)) go.mod go.sum
TESTS := $(filter %_test.go, $(wildcard *.go))
TOOL := $(wildcard scripts/*/*.go)
TESTDATA := $(wildcard testdata/examples/*.yaml testdata/examples/*.out)
GOTEST := $(shell command -v gotest 2>/dev/null)

all: clean build test

.testtimestamp: $(TESTS) $(SRCS) $(TESTDATA) $(TOOL)
ifdef GOTEST
	gotest ./ ./scripts/... # https://github.com/rhysd/gotest
else
	go test -v ./ ./scripts/...
endif
	touch .testtimestamp

test: .testtimestamp

.staticchecktimestamp: $(TESTS) $(SRCS) $(TOOL)
	staticcheck ./ ./cmd/... ./scripts/...
	GOOS=js GOARCH=wasm staticcheck ./playground
	touch .staticchecktimestamp

lint: .staticchecktimestamp

popular_actions.go all_webhooks.go availability.go: scripts/generate-popular-actions/main.go scripts/generate-webhook-events/main.go scripts/generate-availability/main.go
ifdef SKIP_GO_GENERATE
	touch popular_actions.go all_webhooks.go availability.go
else
	go generate
endif

actionlint: $(SRCS) popular_actions.go all_webhooks.go
	CGO_ENABLED=0 go build ./cmd/actionlint

build: actionlint

actionlint_fuzz-fuzz.zip:
	go-fuzz-build ./fuzz

fuzz: actionlint_fuzz-fuzz.zip
	go-fuzz -bin ./actionlint_fuzz-fuzz.zip -func $(FUZZ_FUNC)

man/actionlint.1 man/actionlint.1.html: man/actionlint.1.ronn
	ronn man/actionlint.1.ronn

man: man/actionlint.1

bench:
	go test -bench Lint -benchmem

.github/actionlint-matcher.json: scripts/generate-actionlint-matcher/object.js
	node ./scripts/generate-actionlint-matcher/main.js .github/actionlint-matcher.json

scripts/generate-actionlint-matcher/test/escape.txt: actionlint
	./actionlint -color ./testdata/err/one_error.yaml > ./scripts/generate-actionlint-matcher/test/escape.txt || true 
scripts/generate-actionlint-matcher/test/no_escape.txt: actionlint
	./actionlint -no-color ./testdata/err/one_error.yaml > ./scripts/generate-actionlint-matcher/test/no_escape.txt || true
scripts/generate-actionlint-matcher/test/want.json: actionlint
	./actionlint -format '{{json .}}' ./testdata/err/one_error.yaml > scripts/generate-actionlint-matcher/test/want.json || true

clean:
	rm -f ./actionlint ./.testtimestamp ./.staticchecktimestamp ./actionlint_fuzz-fuzz.zip ./man/actionlint.1 ./man/actionlint.1.html ./actionlint-workflow-ast
	rm -rf ./corpus ./crashers

b: build
t: test
c: clean
l: lint

.PHONY: all test clean build lint fuzz man bench b t c l
