SWEEP               ?= us-west-2,us-east-1,us-east-2
TEST                ?= ./...
SWEEP_DIR           ?= ./internal/sweep
PKG_NAME            ?= internal
TEST_COUNT          ?= 1
ACCTEST_TIMEOUT     ?= 180m
ACCTEST_PARALLELISM ?= 20

ifneq ($(origin PKG), undefined)
	PKG_NAME = internal/service/$(PKG)
endif

ifneq ($(origin TESTS), undefined)
	RUNARGS = -run='$(TESTS)'
endif

default: build

build: fmtcheck
	go install

gen:
	rm -f .github/labeler-issue-triage.yml
	rm -f .github/labeler-pr-triage.yml
	rm -f infrastructure/repository/labels-service.tf
	rm -f internal/conns/*_gen.go
	rm -f internal/service/**/*_gen.go
	rm -f internal/sweep/sweep_test.go
	rm -f names/*_gen.go
	rm -f website/allowed-subcategories.txt
	rm -f website/docs/guides/custom-service-endpoints.html.md
	go generate ./...

sweep:
	# make sweep SWEEPARGS=-sweep-run=aws_example_thing
	@echo "WARNING: This will destroy infrastructure. Use only in development accounts."
	go test $(SWEEP_DIR) -v -tags=sweep -sweep=$(SWEEP) $(SWEEPARGS) -timeout 60m

test: fmtcheck
	go test $(TEST) $(TESTARGS) -timeout=5m

testacc: fmtcheck
	@if [ "$(TESTARGS)" = "-run=TestAccXXX" ]; then \
		echo ""; \
		echo "Error: Skipping example acceptance testing pattern. Update PKG and TESTS for the relevant *_test.go file."; \
		echo ""; \
		echo "For example if updating internal/service/acm/certificate.go, use the test names in internal/service/acm/certificate_test.go starting with TestAcc and up to the underscore:"; \
		echo "make testacc TESTS=TestAccACMCertificate_ PKG=acm"; \
		echo ""; \
		echo "See the contributing guide for more information: https://github.com/hashicorp/terraform-provider-aws/blob/main/docs/contributing/running-and-writing-acceptance-tests.md"; \
		exit 1; \
	fi
	TF_ACC=1 go test ./$(PKG_NAME)/... -v -count $(TEST_COUNT) -parallel $(ACCTEST_PARALLELISM) $(RUNARGS) $(TESTARGS) -timeout $(ACCTEST_TIMEOUT)

fmt:
	@echo "==> Fixing source code with gofmt..."
	gofmt -s -w ./$(PKG_NAME) ./names $(filter-out ./providerlint/go% ./providerlint/README.md ./providerlint/vendor, $(wildcard ./providerlint/*))

# Currently required by tf-deploy compile
fmtcheck:
	@sh -c "'$(CURDIR)/scripts/gofmtcheck.sh'"

gencheck:
	@echo "==> Checking generated source code..."
	@$(MAKE) gen
	@git diff --compact-summary --exit-code || \
		(echo; echo "Unexpected difference in directories after code generation. Run 'make gen' command and commit."; exit 1)

generate-changelog:
	@echo "==> Generating changelog..."
	@sh -c "'$(CURDIR)/scripts/generate-changelog.sh'"

depscheck:
	@echo "==> Checking source code with go mod tidy..."
	@go mod tidy
	@git diff --exit-code -- go.mod go.sum || \
		(echo; echo "Unexpected difference in go.mod/go.sum files. Run 'go mod tidy' command or revert any go.mod/go.sum changes and commit."; exit 1)

docs-lint:
	@echo "==> Checking docs against linters..."
	@misspell -error -source=text docs/ || (echo; \
		echo "Unexpected misspelling found in docs files."; \
		echo "To automatically fix the misspelling, run 'make docs-lint-fix' and commit the changes."; \
		exit 1)
	@docker run --rm -v $(PWD):/markdown 06kellyjac/markdownlint-cli docs/ || (echo; \
		echo "Unexpected issues found in docs Markdown files."; \
		echo "To apply any automatic fixes, run 'make docs-lint-fix' and commit the changes."; \
		exit 1)

docs-lint-fix:
	@echo "==> Applying automatic docs linter fixes..."
	@misspell -w -source=text docs/
	@docker run --rm -v $(PWD):/markdown 06kellyjac/markdownlint-cli --fix docs/

docscheck:
	@tfproviderdocs check \
		-allowed-resource-subcategories-file website/allowed-subcategories.txt \
		-ignore-side-navigation-data-sources aws_alb,aws_alb_listener,aws_alb_target_group,aws_kms_secret \
		-require-resource-subcategory
	@misspell -error -source text CHANGELOG.md .changelog

lint: golangci-lint providerlint importlint

gh-workflows-lint:
	@echo "==> Checking github workflows with actionlint..."
	@actionlint

golangci-lint:
	@echo "==> Checking source code with golangci-lint..."
	@golangci-lint run ./$(PKG_NAME)/...

providerlint:
	@echo "==> Checking source code with providerlint..."
	@providerlint \
		-c 1 \
		-AT001.ignored-filename-suffixes=_data_source_test.go \
		-AWSAT006=false \
		-AWSR002=false \
		-AWSV001=false \
		-R001=false \
		-R010=false \
		-R018=false \
		-R019=false \
		-V001=false \
		-V009=false \
		-V011=false \
		-V012=false \
		-V013=false \
		-V014=false \
		-XR001=false \
		-XR002=false \
		-XR003=false \
		-XR004=false \
		-XR005=false \
		-XS001=false \
		-XS002=false \
		./$(PKG_NAME)/service/... ./$(PKG_NAME)/provider/...

importlint:
	@echo "==> Checking source code with importlint..."
	@impi --local . --scheme stdThirdPartyLocal ./$(PKG_NAME)/...

tools:
	cd providerlint && go install .
	cd tools && go install github.com/bflad/tfproviderdocs
	cd tools && go install github.com/client9/misspell/cmd/misspell
	cd tools && go install github.com/golangci/golangci-lint/cmd/golangci-lint
	cd tools && go install github.com/katbyte/terrafmt
	cd tools && go install github.com/terraform-linters/tflint
	cd tools && go install github.com/pavius/impi/cmd/impi
	cd tools && go install github.com/hashicorp/go-changelog/cmd/changelog-build
	cd tools && go install github.com/rhysd/actionlint/cmd/actionlint

test-compile:
	@if [ "$(TEST)" = "./..." ]; then \
		echo "ERROR: Set TEST to a specific package. For example,"; \
		echo "  make test-compile TEST=./$(PKG_NAME)"; \
		exit 1; \
	fi
	go test -c $(TEST) $(TESTARGS)

website-link-check:
	@scripts/markdown-link-check.sh

website-link-check-ghrc:
	@LINK_CHECK_CONTAINER="ghcr.io/tcort/markdown-link-check:stable" scripts/markdown-link-check.sh	

website-lint:
	@echo "==> Checking website against linters..."
	@misspell -error -source=text website/ || (echo; \
		echo "Unexpected mispelling found in website files."; \
		echo "To automatically fix the misspelling, run 'make website-lint-fix' and commit the changes."; \
		exit 1)
	@docker run --rm -v $(PWD):/markdown 06kellyjac/markdownlint-cli website/docs/ || (echo; \
		echo "Unexpected issues found in website Markdown files."; \
		echo "To apply any automatic fixes, run 'make website-lint-fix' and commit the changes."; \
		exit 1)
	@terrafmt diff ./website --check --pattern '*.markdown' --quiet || (echo; \
		echo "Unexpected differences in website HCL formatting."; \
		echo "To see the full differences, run: terrafmt diff ./website --pattern '*.markdown'"; \
		echo "To automatically fix the formatting, run 'make website-lint-fix' and commit the changes."; \
		exit 1)

website-lint-fix:
	@echo "==> Applying automatic website linter fixes..."
	@misspell -w -source=text website/
	@docker run --rm -v $(PWD):/markdown 06kellyjac/markdownlint-cli --fix website/docs/
	@terrafmt fmt ./website --pattern '*.markdown'

semgrep:
	@echo "==> Running Semgrep static analysis..."
	@docker run --rm --volume "${PWD}:/src" returntocorp/semgrep --config .semgrep.yml

.PHONY: providerlint build gen generate-changelog gh-workflows-lint golangci-lint sweep test testacc fmt fmtcheck lint tools test-compile website-link-check website-lint website-lint-fix depscheck docscheck semgrep
