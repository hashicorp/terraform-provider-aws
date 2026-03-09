SHELL := /bin/bash

ACCTEST_PARALLELISM          ?= 20
ACCTEST_TIMEOUT              ?= 360m
BASE_REF                     ?= main
GO_VER                       ?= $(shell echo go`cat .go-version | xargs`)
P                            ?= 20
PKG_NAME                     ?= internal
SEMGREP_ARGS                 ?= --error
SEMGREP_ENABLE_VERSION_CHECK ?= false
SEMGREP_SEND_METRICS         ?= off
SEMGREP_TIMEOUT              ?= 900 # 15 minutes, some runs go over 5 minutes
SVC_DIR                      ?= ./internal/service
SWEEP                        ?= us-west-2,us-east-1,us-east-2,us-west-1
SWEEP_DIR                    ?= ./internal/sweep
SWEEP_TIMEOUT                ?= 360m
TEST                         ?= ./...
TEST_COUNT                   ?= 1

# NOTE:
# 1. Keep targets in alphabetical order
# 2. For any changes, also update:
#    - docs/makefile-cheat-sheet.md
#    - docs/continuous-integration.md

# VARIABLE REFERENCE:
# Service-specific variables (interchangeable for user convenience):
#   PKG=<service>     - Service name (e.g., ses, lambda, s3) - traditional usage
#   K=<service>       - Service name (e.g., ses, lambda, s3) - shorter alias
#
# Test-specific variables:
#   T=<pattern>       - Test name pattern (e.g., TestAccLambda) - preferred
#   TESTS=<pattern>   - Test name pattern - legacy alias for T
#
# Derived variables (set automatically based on above):
#   PKG_NAME          - Full package path (e.g., internal/service/ses)
#   SVC_DIR           - Service directory path (e.g., ./internal/service/ses)
#   TEST              - Test path pattern (e.g., ./internal/service/ses/...)
#
# Examples:
#   make quick-fix PKG=ses     # Fix code in SES service
#   make quick-fix K=lambda    # Same as above, but shorter (both work)
#   make t T=TestAccRole PKG=iam  # Run specific test in IAM service

# Variable consolidation for backward compatibility and user convenience:
# - PKG and K both refer to service names (e.g., 'ses', 'lambda')
# - If one is provided, automatically set the other for consistency
# - This allows 'make quick-fix PKG=ses' and 'make quick-fix K=ses' to work identically
ifneq ($(origin PKG), undefined)
	PKG_NAME = internal/service/$(PKG)
	SVC_DIR = ./internal/service/$(PKG)
	TEST = ./$(PKG_NAME)/...
	# Auto-set K for compatibility
	K = $(PKG)
endif

ifneq ($(origin K), undefined)
	PKG_NAME = internal/service/$(K)
	SVC_DIR = ./internal/service/$(K)
	TEST = ./$(PKG_NAME)/...
	# Auto-set PKG for compatibility (only if not already set)
	ifeq ($(origin PKG), undefined)
		PKG = $(K)
	endif
endif

ifneq ($(origin TESTS), undefined)
	RUNARGS = -run='$(TESTS)'
endif

ifneq ($(origin T), undefined)
	RUNARGS = -run='$(T)'
endif

ifneq ($(origin SWEEPERS), undefined)
	SWEEPARGS = -sweep-run='$(SWEEPERS)'
endif

ifeq ($(origin CURDIR), undefined)
	CURDIR = $(PWD)
endif

ifeq ($(PKG_NAME), internal/service/ebs)
	PKG_NAME = internal/service/ec2
	SVC_DIR = ./internal/service/ec2
	TEST = ./$(PKG_NAME)/...
endif

ifeq ($(PKG_NAME), internal/service/ipam)
	PKG_NAME = internal/service/ec2
	SVC_DIR = ./internal/service/ec2
	TEST = ./$(PKG_NAME)/...
endif

ifeq ($(PKG_NAME), internal/service/transitgateway)
	PKG_NAME = internal/service/ec2
	SVC_DIR = ./internal/service/ec2
	TEST = ./$(PKG_NAME)/...
endif

ifeq ($(PKG_NAME), internal/service/vpc)
	PKG_NAME = internal/service/ec2
	SVC_DIR = ./internal/service/ec2
	TEST = ./$(PKG_NAME)/...
endif

ifeq ($(PKG_NAME), internal/service/vpnclient)
	PKG_NAME = internal/service/ec2
	SVC_DIR = ./internal/service/ec2
	TEST = ./$(PKG_NAME)/...
endif

ifeq ($(PKG_NAME), internal/service/vpnsite)
	PKG_NAME = internal/service/ec2
	SVC_DIR = ./internal/service/ec2
	TEST = ./$(PKG_NAME)/...
endif

ifeq ($(PKG_NAME), internal/service/wavelength)
	PKG_NAME = internal/service/ec2
	SVC_DIR = ./internal/service/ec2
	TEST = ./$(PKG_NAME)/...
endif

ifneq ($(P), 20)
	ACCTEST_PARALLELISM = $(P)
endif

default: build ## build

acctest-lint: testacc-lint testacc-tflint ## [CI] Run all CI acceptance test checks

build: prereq-go fmt-check ## Build provider
	@echo "make: Building provider..."
	@$(GO_VER) install

changelog-misspell: ## [CI] CHANGELOG Misspell / misspell
	@echo "make: CHANGELOG Misspell / misspell..."
	@misspell -error -source text CHANGELOG.md .changelog

ci: tools go-build gen-check acctest-lint copyright deps-check docs examples-tflint gh-workflow-lint golangci-lint import-lint provider-lint provider-markdown-lint semgrep skaff-check-compile sweeper-check test tfproviderdocs website yamllint ## [CI] Run all CI checks

ci-quick: tools go-build testacc-lint copyright deps-check docs examples-tflint gh-workflow-lint golangci-lint1 import-lint provider-lint provider-markdown-lint semgrep-code-quality semgrep-naming semgrep-naming-cae website-markdown-lint website-misspell website-terrafmt yamllint ## [CI] Run quicker CI checks

clean: clean-make-tests clean-go clean-tidy build tools ## Clean up Go cache, tidy and re-install tools
	@echo "make: Clean complete"

clean-go: prereq-go ## Clean up Go cache
	@echo "make: Cleaning Go..."
	@echo "make: WARNING: This will kill gopls and clean Go caches"
	@vscode=`ps -ef | grep Visual\ Studio\ Code | wc -l | xargs` ; \
	if [ $$vscode -gt 1 ] ; then \
		echo "make: ALERT: vscode is running. Close it and try again." ; \
		exit 1 ; \
	fi
	@for proc in `pgrep gopls` ; do \
		echo "make: killing gopls process $$proc" ; \
		kill -9 $$proc ; \
	done ; \
	echo "make: cleaning Go caches..." ; \
	$(GO_VER) clean -modcache -testcache -cache -i -r
	go clean -modcache -testcache -cache -i -r
	@echo "make: Go caches cleaned"

clean-go-cache-trim: prereq-go ## Trim Go build cache to manageable size (keeps recent entries)
	@echo "make: Trimming Go build cache..."
	@cache_dir=$$(go env GOCACHE) ; \
	if [ -d "$$cache_dir" ]; then \
		echo "make: Current cache size: $$(du -sh $$cache_dir | cut -f1)" ; \
		echo "make: Removing cache entries older than 7 days..." ; \
		find "$$cache_dir" -type f -atime +7 -delete 2>/dev/null || true ; \
		find "$$cache_dir" -type d -empty -delete 2>/dev/null || true ; \
		echo "make: Cache size after trim: $$(du -sh $$cache_dir | cut -f1)" ; \
		cache_size_mb=$$(du -sm "$$cache_dir" | cut -f1) ; \
		if [ $$cache_size_mb -gt 51200 ]; then \
			echo "make: WARNING: Cache still large ($$cache_size_mb MB). Consider 'make clean-go' for full cleanup." ; \
		fi ; \
	else \
		echo "make: No cache directory found at $$cache_dir" ; \
	fi

cache-info: prereq-go ## Display Go cache and GitHub Actions cache information
	@echo "=== Go Cache Information ==="
	@gocache=$$(go env GOCACHE) ; \
	gomodcache=$$(go env GOMODCACHE) ; \
	echo "GOCACHE:     $$gocache" ; \
	echo "GOMODCACHE:  $$gomodcache" ; \
	echo "" ; \
	if [ -d "$$gocache" ]; then \
		size=$$(du -sh "$$gocache" 2>/dev/null | cut -f1) ; \
		files=$$(find "$$gocache" -type f 2>/dev/null | wc -l | xargs) ; \
		echo "Build cache size:  $$size ($$files files)" ; \
		recent=$$(find "$$gocache" -type f -mtime -1 2>/dev/null | wc -l | xargs) ; \
		week=$$(find "$$gocache" -type f -mtime -7 2>/dev/null | wc -l | xargs) ; \
		old=$$(find "$$gocache" -type f -mtime +7 2>/dev/null | wc -l | xargs) ; \
		echo "  < 1 day old:  $$recent files" ; \
		echo "  < 7 days old: $$week files" ; \
		echo "  > 7 days old: $$old files" ; \
	else \
		echo "Build cache: not found" ; \
	fi ; \
	echo "" ; \
	if [ -d "$$gomodcache" ]; then \
		size=$$(du -sh "$$gomodcache" 2>/dev/null | cut -f1) ; \
		echo "Module cache size: $$size" ; \
	else \
		echo "Module cache: not found" ; \
	fi ; \
	echo "" ; \
	echo "=== GitHub Actions Cache (if in CI) ===" ; \
	if [ -n "$$GITHUB_ACTIONS" ]; then \
		echo "Running in GitHub Actions" ; \
		echo "CACHE_DATE: $${CACHE_DATE:-not set}" ; \
		echo "Runner OS:  $${RUNNER_OS:-not set}" ; \
		if [ -n "$$ACTIONS_CACHE_URL" ]; then \
			echo "Cache service: available" ; \
		else \
			echo "Cache service: not available" ; \
		fi ; \
	else \
		echo "Not running in GitHub Actions" ; \
		echo "To check GitHub cache usage, visit:" ; \
		echo "  https://github.com/hashicorp/terraform-provider-aws/actions/caches" ; \
	fi

clean-make-tests: ## Clean up artifacts from make tests
	@echo "make: Cleaning up artifacts from make tests..."
	@rm -rf sweeper-bin
	@rm -rf terraform-plugin-dir
	@rm -rf .terraform/providers
	@rm -rf terraform-providers-schema
	@rm -rf example.tf
	@rm -rf skaff/skaff

clean-tidy: prereq-go ## Clean up tidy
	@echo "make: Tidying Go mods..."
	@gover="$(GO_VER)" ; \
	echo "make: tidying with $$gover" ; \
	if [ "$$gover" = "go" ] ; then \
		gover=go`cat .go-version | xargs` ; \
		echo "make: WARNING: no version provided so tidying with $$gover" ; \
		echo "make: tidying with newer versions can make go.mod incompatible" ; \
		echo "make: to use a different version, use 'GO_VER=go1.16 make clean-tidy'" ; \
		echo "make: to use the version in .go-version, use 'make clean-tidy'" ; \
		echo "make: if you get an error, see https://go.dev/doc/manage-install to locally install various Go versions" ; \
	fi ; \
	cd .ci/providerlint && $$gover mod tidy && cd ../.. ; \
	cd tools/tfsdk2fw && $$gover mod tidy && cd ../.. ; \
	cd .ci/tools && $$gover mod tidy && cd ../.. ; \
	cd .ci/providerlint && $$gover mod tidy && cd ../.. ; \
	cd skaff && $$gover mod tidy && cd .. ; \
	$$gover mod tidy
	@echo "make: Go mods tidied"

copyright: ## [CI] Copyright Checks / headers check
	@echo "make: Copyright Checks / headers check..."
	@which copyplop > /dev/null || go install github.com/YakDriver/copyplop
	@copyplop check

copyright-fix: ## Fix copyright headers
	@echo "make: Fixing copyright headers..."
	@which copyplop > /dev/null || go install github.com/YakDriver/copyplop
	@copyplop fix

deps-check: clean-tidy ## [CI] Dependency Checks / go_mod
	@echo "make: Dependency Checks / go_mod..."
	@git diff --exit-code -- go.mod go.sum || \
		(echo; echo "Unexpected difference in go.mod/go.sum files. Run 'go mod tidy' command or revert any go.mod/go.sum changes and commit."; exit 1)

docs: docs-link-check docs-markdown-lint docs-misspell ## [CI] Run all CI documentation checks

docs-check: ## Check provider documentation (Legacy, use caution)
	@echo "make: Legacy target, use caution..."
	@tfproviderdocs check \
		-allowed-resource-subcategories-file website/allowed-subcategories.txt \
		-enable-contents-check \
		-ignore-contents-check-data-sources aws_kms_secrets,aws_kms_secret \
		-ignore-file-missing-data-sources aws_alb,aws_alb_listener,aws_alb_target_group,aws_albs \
		-ignore-file-missing-resources aws_alb,aws_alb_listener,aws_alb_listener_certificate,aws_alb_listener_rule,aws_alb_target_group,aws_alb_target_group_attachment \
		-provider-name=aws \
		-require-resource-subcategory

docs-link-check: ## [CI] Documentation Checks / markdown-link-check
	@echo "make: Documentation Checks / markdown-link-check..."
	@docker run --rm \
		-v "$(PWD):/markdown" \
		ghcr.io/yakdriver/md-check-links:2.2.0 \
		--config /markdown/.ci/.markdownlinkcheck.json \
		--verbose yes \
		--quiet yes \
		--directory /markdown/docs \
		--extension '.md' \
		--branch main \
		--modified no

docs-lint: ## Lint documentation (Legacy, use caution)
	@echo "make: Legacy target, use caution..."
	@misspell -error -source text docs/ || (echo; \
		echo "Unexpected misspelling found in docs files."; \
		echo "To automatically fix the misspelling, run 'make docs-lint-fix' and commit the changes."; \
		exit 1)
	@docker run --rm -v $(PWD):/markdown 06kellyjac/markdownlint-cli docs/ || (echo; \
		echo "Unexpected issues found in docs Markdown files."; \
		echo "To apply any automatic fixes, run 'make docs-lint-fix' and commit the changes."; \
		exit 1)

docs-lint-fix: ## Fix documentation linter findings (Legacy, use caution)
	@echo "make: Legacy target, use caution..."
	@misspell -w -source=text docs/
	@docker run --rm -v $(PWD):/markdown 06kellyjac/markdownlint-cli --fix docs/

docs-markdown-lint: ## [CI] Documentation Checks / markdown-lint
	@echo "make: Documentation Checks / markdown-lint..."
	@docker run --rm \
		-v "$(PWD):/markdown" \
		avtodev/markdown-lint:v1.5.0 \
		--config markdown/.markdownlint.yml \
		/markdown/docs/**/*.md

docs-misspell: ## [CI] Documentation Checks / misspell
	@echo "make: Documentation Checks / misspell..."
	@misspell -error -source text docs/

examples-tflint: tflint-init ## [CI] Examples Checks / tflint
	@echo "make: Examples Checks / tflint..."
	TFLINT_CONFIG="$(PWD)/.ci/.tflint.hcl" ; \
	tflint --config="$$TFLINT_CONFIG" --chdir=./examples --recursive \
		--disable-rule=terraform_typed_variables

fix-constants: semgrep-constants fmt ## Use Semgrep to fix constants

fix-imports: ## Fixing source code imports with goimports
	@echo "make: Fixing source code imports with goimports..."
	@if [ -d "./$(PKG_NAME)" ] && ls ./$(PKG_NAME)/*.go >/dev/null 2>&1; then \
		echo "make: Processing ./$(PKG_NAME)..."; \
		goimports -w ./$(PKG_NAME)/*.go; \
	fi
	@for dir in $$(find ./$(PKG_NAME) -mindepth 1 -type d | sort); do \
		if ls $$dir/*.go >/dev/null 2>&1; then \
			echo "make: Processing $$dir..."; \
			goimports -w $$dir/*.go; \
		fi; \
	done

fix-imports-core: ## Fixing core directory imports with goimports
	@echo "make: Fixing core directory imports with goimports..."
	@go list ./... 2>/dev/null | grep -v '/internal/service/' | sed 's|github.com/hashicorp/terraform-provider-aws|.|' | while read pkg; do \
		if [ -d "$$pkg" ] && ls $$pkg/*.go >/dev/null 2>&1; then \
			echo "make: Processing $$pkg..."; \
			goimports -w $$pkg/*.go; \
		fi; \
	done

fmt: ## Fix Go source formatting
	@echo "make: Fixing source code with gofmt..."
	gofmt -s -w ./$(PKG_NAME) ./names $(filter-out ./.ci/providerlint/go% ./.ci/providerlint/README.md ./.ci/providerlint/vendor, $(wildcard ./.ci/providerlint/*))

fmt-core: ## Fix Go source formatting in core directories
	@echo "make: Fixing core directory source code with gofmt..."
	@core_pkgs=$$(go list ./... 2>/dev/null | grep -v '/internal/service/' | sed 's|github.com/hashicorp/terraform-provider-aws|.|'); \
	gofmt -s -w $$core_pkgs

# Currently required by tf-deploy compile
fmt-check: ## Verify Go source is formatted
	@echo "make: Verifying source code with gofmt..."
	@sh -c "'$(CURDIR)/.ci/scripts/gofmtcheck.sh'"

fumpt: ## Run gofumpt
	@echo "make: Fixing source code with gofumpt..."
	gofumpt -w ./$(PKG_NAME) ./names $(filter-out ./.ci/providerlint/go% ./.ci/providerlint/README.md ./.ci/providerlint/vendor, $(wildcard ./.ci/providerlint/*))

gen: prereq-go gen-raw ## Run all Go generators (with Go version check)

gen-raw: ## Run all Go generators
	@echo "make: Running Go generators..."
	$(GO_VER) generate ./...
	# Generate service package lists last as they may depend on output of earlier generators.
	$(GO_VER) generate ./internal/provider/...
	$(GO_VER) generate ./internal/sweep

gen-check: gen ## [CI] Provider Checks / go_generate
	@echo "make: Provider Checks / go_generate..."
	@echo "make: NOTE: commit any changes before running this check"
	@git diff origin/$(BASE_REF) --compact-summary --exit-code || \
		(echo; echo "Unexpected difference in directories after code generation. Run 'make gen' command and commit."; exit 1)

generate-changelog: ## Generate changelog
	@echo "make: Generating changelog..."
	@sh -c "'$(CURDIR)/.ci/scripts/generate-changelog.sh'"

gh-workflow-lint: ## [CI] Workflow Linting / actionlint
	@echo "make: Workflow Linting / actionlint..."
	@actionlint -shellcheck=

go-build: ## [CI] Provider Checks / go-build
	@os_arch=`go env GOOS`_`go env GOARCH` ; \
	echo "make: Provider Checks / go-build ($$os_arch)..." ; \
	go build -o terraform-plugin-dir/registry.terraform.io/hashicorp/aws/99.99.99/$$os_arch/terraform-provider-aws .

go-misspell: ## [CI] Provider Checks / misspell
	@echo "make: Provider Checks / misspell..."
	@misspell -error -source auto -i "littel,ceasar,ect" internal/

golangci-lint: golangci-lint1 golangci-lint2 golangci-lint3 golangci-lint4 golangci-lint5 ## [CI] All golangci-lint Checks

golangci-lint1: ## [CI] golangci-lint Checks / 1 of 5
	@echo "make: golangci-lint Checks / 1 of 5..."
	@golangci-lint run \
		--config .ci/.golangci.yml \
		$(TEST)

golangci-lint2: ## [CI] golangci-lint Checks / 2 of 5
	@echo "make: golangci-lint Checks / 2 of 5..."
	@golangci-lint run \
		--config .ci/.golangci2.yml \
		$(TEST)

golangci-lint3: ## [CI] golangci-lint Checks / 3 of 5
	@echo "make: golangci-lint Checks / 3 of 5..."
	@golangci-lint run \
		--config .ci/.golangci3.yml \
		$(TEST)

golangci-lint4: ## [CI] golangci-lint Checks / 4 of 5
	@echo "make: golangci-lint Checks / 4 of 5..."
	@golangci-lint run \
		--config .ci/.golangci4.yml \
		$(TEST)

golangci-lint5: ## [CI] golangci-lint Checks / 5 of 5
	@echo "make: golangci-lint Checks / 5 of 5..."
	@golangci-lint run \
		--config .ci/.golangci5.yml \
		$(TEST)

help: ## Display this help
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-27s\033[0m %s\n", $$1, $$2}'

import-lint: ## [CI] Provider Checks / import-lint
	@echo "make: Provider Checks / import-lint..."
	@impi --local . --scheme stdThirdPartyLocal $(TEST)

install: build ## build

lint: golangci-lint provider-lint import-lint ## Legacy target, use caution

lint-fix: testacc-lint-fix website-lint-fix docs-lint-fix ## Fix acceptance test, website, and docs linter findings

misspell: changelog-misspell docs-misspell website-misspell go-misspell ## [CI] Run all CI misspell checks

modern-check: prereq-go ## [CI] Check for modern Go code (best run in individual services)
	@echo "make: Checking for modern Go code..."
	@$(GO_VER) run golang.org/x/tools/gopls/internal/analysis/modernize/cmd/modernize@v0.21.0 -test $(TEST)

modern-fix: prereq-go ## [CI] Fix checks for modern Go code (best run in individual services)
	@echo "make: Fixing checks for modern Go code..."
	@$(GO_VER) run golang.org/x/tools/gopls/internal/analysis/modernize/cmd/modernize@v0.21.0 -fix -test $(TEST)

modern-fix-core: prereq-go ## Fix checks for modern Go code in core directories
	@echo "make: Fixing checks for modern Go code in core directories..."
	@core_pkgs=$$(go list ./... 2>/dev/null | grep -v '/internal/service/'); \
	$(GO_VER) run golang.org/x/tools/gopls/internal/analysis/modernize/cmd/modernize@v0.21.0 -fix -test $$core_pkgs

pr-target-check: ## [CI] Check for pull request target
	@echo "make: Checking for pull request target..."
	@disallowed_files=$$(grep -rl 'pull_request_target' ./.github/workflows/*.yml | grep -vE './.github/workflows/(maintainer_helpers|triage|closed_items|community_note|readiness_comment).yml' || true); \
	if [ -n "$$disallowed_files" ]; then \
		echo "Error: 'pull_request_target' found in disallowed files:"; \
		echo "$$disallowed_files"; \
		exit 1; \
	fi
	@echo "make: pr-target-check passed."

prereq-go: ## If $(GO_VER) is not installed, install it
	@if ! type "$(GO_VER)" > /dev/null 2>&1 ; then \
		echo "make: $(GO_VER) not found" ; \
		echo "make: installing $(GO_VER)..." ; \
		echo "make: if you get an error, see https://go.dev/doc/manage-install to locally install various Go versions" ; \
		go install golang.org/dl/$(GO_VER)@latest ; \
		$(GO_VER) download ; \
		echo "make: $(GO_VER) ready" ; \
	fi

provider-lint: ## [CI] ProviderLint Checks / providerlint
	@echo "make: ProviderLint Checks / providerlint..."
	@cd .ci/providerlint && go install -buildvcs=false .
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
		$(SVC_DIR)/... ./internal/provider/...

quick-fix-core-heading: ## Just a heading for quick-fix-core
	@echo "make: Quick fixes for core (non-service) directories..."
	@echo "make: Multiple runs are needed if it finds errors (later targets not reached)"

quick-fix-core: quick-fix-core-heading copyright-fix fmt-core testacc-lint-fix-core fix-imports-core modern-fix-core semgrep-fix-core website-terrafmt-fix ## Quick fixes for core directories (non-internal/service)

quick-fix-heading: ## Just a heading for quick-fix
	@echo "make: Quick fixes..."
	@echo "make: Multiple runs are needed if it finds errors (later targets not reached)"

quick-fix: quick-fix-heading copyright-fix fmt testacc-lint-fix fix-imports modern-fix semgrep-fix terraform-fmt website-terrafmt-fix ## Some quick fixes

provider-markdown-lint: ## [CI] Provider Check / markdown-lint
	@echo "make: Provider Check / markdown-lint..."
	@docker run --rm \
		-v "$(PWD):/markdown" \
		avtodev/markdown-lint:v1.5.0 \
		--config markdown/.markdownlint.yml \
		--ignore markdown/docs \
		--ignore markdown/website/docs \
		--ignore markdown/CHANGELOG.md \
		--ignore markdown/internal/service/cloudformation/test-fixtures/examplecompany-exampleservice-exampleresource/docs \
		/markdown/**/*.md

# The 2 smoke test targets run exactly the same set of acceptance tests.
# The tests must pass in the AWS Commercial and AWS GovCloud (US) partitions.
# The tests must pass on the earliest supported Terraform version (0.12.31).

sane: prereq-go ## Run sane check
	@echo "make: Sane Smoke Tests (x tests of Top y resources)"
	@echo "make: Like 'sanity' except full output and stops soon after 1st error"
	@echo "make: NOTE: NOT an exhaustive set of tests! Finds big problems only."
	@TF_ACC=1 $(GO_VER) test \
		./internal/service/iam/... \
		-v -count $(TEST_COUNT) -parallel $(ACCTEST_PARALLELISM) -run='^TestAccIAMRole_basic$$|^TestAccIAMRole_namePrefix$$|^TestAccIAMRole_disappears$$|^TestAccIAMRole_InlinePolicy_basic$$|^TestAccIAMPolicyDocumentDataSource_basic$$|^TestAccIAMPolicyDocumentDataSource_sourceConflicting$$|^TestAccIAMPolicyDocumentDataSource_sourceJSONValidJSON$$|^TestAccIAMRolePolicyAttachment_basic$$|^TestAccIAMRolePolicyAttachment_disappears$$|^TestAccIAMRolePolicyAttachment_Disappears_role$$|^TestAccIAMPolicy_basic$$|^TestAccIAMPolicy_policy$$|^TestAccIAMPolicy_tags$$|^TestAccIAMRolePolicy_basic$$|^TestAccIAMRolePolicy_unknownsInPolicy$$|^TestAccIAMInstanceProfile_basic$$|^TestAccIAMInstanceProfile_tags$$|^TestAccIAMPolicy_List_Basic$$|^TestAccIAMRole_Identity_Basic$$' -timeout $(ACCTEST_TIMEOUT) -vet=off
	@TF_ACC=1 $(GO_VER) test \
		./internal/service/logs/... \
		./internal/service/ec2/... \
		./internal/service/ecs/... \
		./internal/service/elbv2/... \
		./internal/service/events/... \
		./internal/service/kms/... \
		-v -count $(TEST_COUNT) -parallel $(ACCTEST_PARALLELISM) -run='^TestAccVPCSecurityGroup_basic$$|^TestAccVPCSecurityGroup_egressMode$$|^TestAccVPCSecurityGroup_vpcAllEgress$$|^TestAccVPCSecurityGroupRule_race$$|^TestAccVPCSecurityGroupRule_protocolChange$$|^TestAccVPCDataSource_basic$$|^TestAccVPCSubnet_basic$$|^TestAccVPC_tenancy$$|^TestAccVPCRouteTableAssociation_Subnet_basic$$|^TestAccVPCRouteTable_basic$$|^TestAccLogsLogGroup_basic$$|^TestAccLogsLogGroup_multiple$$|^TestAccKMSKey_basic$$|^TestAccELBV2TargetGroup_basic$$|^TestAccECSTaskDefinition_basic$$|^TestAccECSService_basic$$|^TestAccEventsPutEventsAction_basic$$' -timeout $(ACCTEST_TIMEOUT) -vet=off
	@TF_ACC=1 $(GO_VER) test \
		./internal/service/lambda/... \
		./internal/service/meta/... \
		./internal/service/route53/... \
		./internal/service/s3/... \
		./internal/service/ssm/... \
		./internal/service/secretsmanager/... \
		./internal/service/sts/... \
		./internal/function/... \
		-v -count $(TEST_COUNT) -parallel $(ACCTEST_PARALLELISM) -run='^TestAccSTSCallerIdentityDataSource_basic$$|^TestAccMetaRegionDataSource_basic$$|^TestAccMetaRegionDataSource_endpoint$$|^TestAccMetaPartitionDataSource_basic$$|^TestAccS3Bucket_Basic_basic$$|^TestAccS3Bucket_Security_corsUpdate$$|^TestAccS3BucketPublicAccessBlock_basic$$|^TestAccS3BucketPolicy_basic$$|^TestAccS3BucketACL_updateACL$$|^TestAccS3Object_basic$$|^TestAccRoute53Record_basic$$|^TestAccRoute53Record_Latency_basic$$|^TestAccRoute53ZoneDataSource_name$$|^TestAccLambdaFunction_basic$$|^TestAccLambdaPermission_basic$$|^TestAccSecretsManagerSecret_basic$$|^TestAccSSMParameterEphemeral_basic$$|^TestAccLambdaCapacityProvider_List_Basic$$|^TestARNParseFunction_known$$' -timeout $(ACCTEST_TIMEOUT) -vet=off

sanity: prereq-go ## Run sanity check (failures allowed)
	@echo "make: Sanity Smoke Tests (x tests of Top y resources)"
	@echo "make: Like 'sane' but less output and runs all tests despite most errors"
	@echo "make: NOTE: NOT an exhaustive set of tests! Finds big problems only."
	@iam=`TF_ACC=1 $(GO_VER) test \
		./internal/service/iam/... \
		-v -count $(TEST_COUNT) -parallel $(ACCTEST_PARALLELISM) -run='^TestAccIAMRole_basic$$|^TestAccIAMRole_namePrefix$$|^TestAccIAMRole_disappears$$|^TestAccIAMRole_InlinePolicy_basic$$|^TestAccIAMPolicyDocumentDataSource_basic$$|^TestAccIAMPolicyDocumentDataSource_sourceConflicting$$|^TestAccIAMPolicyDocumentDataSource_sourceJSONValidJSON$$|^TestAccIAMRolePolicyAttachment_basic$$|^TestAccIAMRolePolicyAttachment_disappears$$|^TestAccIAMRolePolicyAttachment_Disappears_role$$|^TestAccIAMPolicy_basic$$|^TestAccIAMPolicy_policy$$|^TestAccIAMPolicy_tags$$|^TestAccIAMRolePolicy_basic$$|^TestAccIAMRolePolicy_unknownsInPolicy$$|^TestAccIAMInstanceProfile_basic$$|^TestAccIAMInstanceProfile_tags$$|^TestAccIAMPolicy_List_Basic$$|^TestAccIAMRole_Identity_Basic$$' -timeout $(ACCTEST_TIMEOUT) -vet=off || true` ; \
	fails1=`echo -n $$iam | grep -Fo FAIL: | wc -l | xargs` ; \
	passes=$$(( 18-$$fails1 )) ; \
	echo "18 of 54 complete: $$passes passed, $$fails1 failed" ; \
	logs=`TF_ACC=1 $(GO_VER) test \
		./internal/service/logs/... \
		./internal/service/ec2/... \
		./internal/service/ecs/... \
		./internal/service/elbv2/... \
		./internal/service/events/... \
		./internal/service/kms/... \
		-v -count $(TEST_COUNT) -parallel $(ACCTEST_PARALLELISM) -run='^TestAccVPCSecurityGroup_basic$$|^TestAccVPCSecurityGroup_egressMode$$|^TestAccVPCSecurityGroup_vpcAllEgress$$|^TestAccVPCSecurityGroupRule_race$$|^TestAccVPCSecurityGroupRule_protocolChange$$|^TestAccVPCDataSource_basic$$|^TestAccVPCSubnet_basic$$|^TestAccVPC_tenancy$$|^TestAccVPCRouteTableAssociation_Subnet_basic$$|^TestAccVPCRouteTable_basic$$|^TestAccLogsLogGroup_basic$$|^TestAccLogsLogGroup_multiple$$|^TestAccKMSKey_basic$$|^TestAccELBV2TargetGroup_basic$$|^TestAccECSTaskDefinition_basic$$|^TestAccECSService_basic$$|^TestAccEventsPutEventsAction_basic$$' -timeout $(ACCTEST_TIMEOUT) -vet=off || true` ; \
	fails2=`echo -n $$logs | grep -Fo FAIL: | wc -l | xargs` ; \
	tot_fails=$$(( $$fails1+$$fails2 )) ; \
	passes=$$(( 35-$$tot_fails )) ; \
	echo "35 of 54 complete: $$passes passed, $$tot_fails failed" ; \
	lambda=`TF_ACC=1 $(GO_VER) test \
		./internal/service/lambda/... \
		./internal/service/meta/... \
		./internal/service/route53/... \
		./internal/service/s3/... \
		./internal/service/secretsmanager/... \
		./internal/service/sts/... \
		./internal/function/... \
		-v -count $(TEST_COUNT) -parallel $(ACCTEST_PARALLELISM) -run='^TestAccSTSCallerIdentityDataSource_basic$$|^TestAccMetaRegionDataSource_basic$$|^TestAccMetaRegionDataSource_endpoint$$|^TestAccMetaPartitionDataSource_basic$$|^TestAccS3Bucket_Basic_basic$$|^TestAccS3Bucket_Security_corsUpdate$$|^TestAccS3BucketPublicAccessBlock_basic$$|^TestAccS3BucketPolicy_basic$$|^TestAccS3BucketACL_updateACL$$|^TestAccS3Object_basic$$|^TestAccRoute53Record_basic$$|^TestAccRoute53Record_Latency_basic$$|^TestAccRoute53ZoneDataSource_name$$|^TestAccLambdaFunction_basic$$|^TestAccLambdaPermission_basic$$|^TestAccSecretsManagerSecret_basic$$|^TestAccSSMParameterEphemeral_basic$$|^TestAccLambdaCapacityProvider_List_Basic$$|^TestARNParseFunction_known$$' -timeout $(ACCTEST_TIMEOUT) -vet=off || true` ; \
	fails3=`echo -n $$lambda | grep -Fo FAIL: | wc -l | xargs` ; \
	tot_fails=$$(( $$fails1+$$fails2+$$fails3 )) ; \
	passes=$$(( 54-$$tot_fails )) ; \
	echo "54 of 54 complete: $$passes passed, $$tot_fails failed" ; \
	if [ $$tot_fails -gt 0 ] ; then \
		echo "Sanity tests failed"; \
		exit 1; \
	fi

semgrep: semgrep-code-quality semgrep-naming semgrep-naming-cae semgrep-service-naming ## [CI] Run all CI Semgrep checks

semgrep-all: semgrep-test semgrep-validate ## Run semgrep on all files
	@echo "make: Running Semgrep checks locally (must have semgrep installed)..."
	@semgrep $(SEMGREP_ARGS) \
		$(if $(filter-out $(origin PKG), undefined),--include $(PKG_NAME),) \
		--config .ci/.semgrep.yml \
		--config .ci/.semgrep-constants.yml \
		--config .ci/.semgrep-test-constants.yml \
		--config .ci/.semgrep-caps-aws-ec2.yml \
		--config .ci/.semgrep-configs.yml \
		--config .ci/.semgrep-service-name0.yml \
		--config .ci/.semgrep-service-name1.yml \
		--config .ci/.semgrep-service-name2.yml \
		--config .ci/.semgrep-service-name3.yml \
		--config .ci/semgrep/ \
		--config 'r/dgryski.semgrep-go.badnilguard' \
		--config 'r/dgryski.semgrep-go.errnilcheck' \
		--config 'r/dgryski.semgrep-go.marshaljson' \
		--config 'r/dgryski.semgrep-go.nilerr' \
		--config 'r/dgryski.semgrep-go.oddifsequence' \
		--config 'r/dgryski.semgrep-go.oserrors'

semgrep-code-quality: semgrep-test semgrep-validate ## [CI] Semgrep Checks / Code Quality Scan
	@echo "make: Semgrep Checks / Code Quality Scan..."
	@echo "make: Running Semgrep checks locally (must have semgrep installed)"
	@semgrep $(SEMGREP_ARGS) \
		$(if $(filter-out $(origin PKG), undefined),--include $(PKG_NAME),) \
		--config .ci/.semgrep.yml \
		--config .ci/.semgrep-constants.yml \
		--config .ci/.semgrep-test-constants.yml \
		--config .ci/semgrep/ \
		--config 'r/dgryski.semgrep-go.badnilguard' \
		--config 'r/dgryski.semgrep-go.errnilcheck' \
		--config 'r/dgryski.semgrep-go.marshaljson' \
		--config 'r/dgryski.semgrep-go.nilerr' \
		--config 'r/dgryski.semgrep-go.oddifsequence' \
		--config 'r/dgryski.semgrep-go.oserrors'

semgrep-constants: semgrep-validate ## Fix constants with Semgrep --autofix
	@echo "make: Fix constants with Semgrep --autofix"
	@semgrep $(SEMGREP_ARGS) --autofix \
		$(if $(filter-out $(origin PKG), undefined),--include $(PKG_NAME),) \
		--config .ci/.semgrep-constants.yml \
		--config .ci/.semgrep-test-constants.yml

semgrep-docker: semgrep-validate ## Run Semgrep (Legacy, use caution)
	@echo "make: Legacy target, use caution..."
	@docker run --rm --volume "${PWD}:/src" returntocorp/semgrep semgrep --config .ci/.semgrep.yml --config .ci/.semgrep-constants.yml --config .ci/.semgrep-test-constants.yml

semgrep-fix: semgrep-validate ## Fix Semgrep issues that have fixes
	@echo "make: Running Semgrep checks locally (must have semgrep installed)..."
	@echo "make: Applying fixes with --autofix"
	@echo "make: WARNING: This will not fix rules that don't have autofixes"
	@semgrep $(SEMGREP_ARGS) --autofix \
		$(if $(filter-out $(origin PKG), undefined),--include $(PKG_NAME),) \
		--config .ci/.semgrep.yml \
		--config .ci/.semgrep-constants.yml \
		--config .ci/.semgrep-test-constants.yml \
		--config .ci/.semgrep-caps-aws-ec2.yml \
		--config .ci/.semgrep-configs.yml \
		--config .ci/.semgrep-service-name0.yml \
		--config .ci/.semgrep-service-name1.yml \
		--config .ci/.semgrep-service-name2.yml \
		--config .ci/.semgrep-service-name3.yml \
		--config .ci/semgrep/ \
		--config 'r/dgryski.semgrep-go.badnilguard' \
		--config 'r/dgryski.semgrep-go.errnilcheck' \
		--config 'r/dgryski.semgrep-go.marshaljson' \
		--config 'r/dgryski.semgrep-go.nilerr' \
		--config 'r/dgryski.semgrep-go.oddifsequence' \
		--config 'r/dgryski.semgrep-go.oserrors'

semgrep-fix-core: semgrep-validate ## Fix Semgrep issues in core directories
	@echo "make: Running Semgrep checks on core directories (must have semgrep installed)..."
	@echo "make: Applying fixes with --autofix"
	@semgrep $(SEMGREP_ARGS) --autofix \
		--exclude 'internal/service/**/*.go' \
		--config .ci/.semgrep.yml \
		--config .ci/.semgrep-constants.yml \
		--config .ci/.semgrep-test-constants.yml \
		--config .ci/.semgrep-caps-aws-ec2.yml \
		--config .ci/.semgrep-configs.yml \
		--config .ci/.semgrep-service-name0.yml \
		--config .ci/.semgrep-service-name1.yml \
		--config .ci/.semgrep-service-name2.yml \
		--config .ci/.semgrep-service-name3.yml \
		--config .ci/semgrep/ \
		--config 'r/dgryski.semgrep-go.badnilguard' \
		--config 'r/dgryski.semgrep-go.errnilcheck' \
		--config 'r/dgryski.semgrep-go.marshaljson' \
		--config 'r/dgryski.semgrep-go.nilerr' \
		--config 'r/dgryski.semgrep-go.oddifsequence' \
		--config 'r/dgryski.semgrep-go.oserrors'

semgrep-naming: semgrep-validate ## [CI] Semgrep Checks / Test Configs Scan
	@echo "make: Semgrep Checks / Test Configs Scan..."
	@echo "make: Running Semgrep checks locally (must have semgrep installed)"
	@semgrep $(SEMGREP_ARGS) \
		$(if $(filter-out $(origin PKG), undefined),--include $(PKG_NAME),) \
		--config .ci/.semgrep-configs.yml

semgrep-naming-cae: semgrep-validate ## [CI] Semgrep Checks / Naming Scan Caps/AWS/EC2
	@echo "make: Semgrep Checks / Naming Scan Caps/AWS/EC2..."
	@echo "make: Running Semgrep checks locally (must have semgrep installed)"
	@semgrep $(SEMGREP_ARGS) \
		$(if $(filter-out $(origin PKG), undefined),--include $(PKG_NAME),) \
		--config .ci/.semgrep-caps-aws-ec2.yml

semgrep-test: semgrep-validate ## Test Semgrep configuration files
	@echo "make: Running Semgrep rule tests..."
	@semgrep --quiet \
		--test .ci/semgrep/

semgrep-service-naming: semgrep-validate ## [CI] Semgrep Checks / Service Name Scan A-Z
	@echo "make: Semgrep Checks / Service Name Scan A-Z..."
	@echo "make: Running Semgrep checks locally (must have semgrep installed)"
	@semgrep $(SEMGREP_ARGS) \
		$(if $(filter-out $(origin PKG), undefined),--include $(PKG_NAME),) \
		--config .ci/.semgrep-service-name0.yml \
		--config .ci/.semgrep-service-name1.yml \
		--config .ci/.semgrep-service-name2.yml \
		--config .ci/.semgrep-service-name3.yml

semgrep-validate: ## Validate Semgrep configuration files
	@echo "make: Validating Semgrep configuration files..."
	@SEMGREP_TIMEOUT=300 semgrep --error --validate \
		--config .ci/.semgrep.yml \
		--config .ci/.semgrep-constants.yml \
		--config .ci/.semgrep-test-constants.yml \
		--config .ci/.semgrep-caps-aws-ec2.yml \
		--config .ci/.semgrep-configs.yml \
		--config .ci/.semgrep-service-name0.yml \
		--config .ci/.semgrep-service-name1.yml \
		--config .ci/.semgrep-service-name2.yml \
		--config .ci/.semgrep-service-name3.yml \
		--config .ci/semgrep/

semgrep-vcr: ## Enable VCR support with Semgrep --autofix
	@echo "make: Enable VCR support with Semgrep --autofix"
	@echo "WARNING: Because some autofixes are inside code blocks replaced by other rules,"
	@echo "this target may need to be run twice."
	@semgrep $(SEMGREP_ARGS) --autofix \
		$(if $(filter-out $(origin PKG), undefined),--include $(PKG_NAME),) \
		--config internal/vcr/.semgrep-vcr.yml

skaff: prereq-go ## Install skaff
	@echo "make: Installing skaff..."
	cd skaff && $(GO_VER) install github.com/hashicorp/terraform-provider-aws/skaff

skaff-check-compile: ## [CI] Skaff Checks / Compile skaff
	@echo "make: Skaff Checks / Compile skaff..."
	@cd skaff ; \
	go build

smoke: sane ## Smoke tests (alias of sane)

sweep: prereq-go ## Run sweepers
	# make sweep SWEEPARGS=-sweep-run=aws_example_thing
	# set SWEEPARGS=-sweep-allow-failures to continue after first failure
	@echo "WARNING: This will destroy infrastructure. Use only in development accounts."
	$(GO_VER) test $(SWEEP_DIR) -v -sweep=$(SWEEP) $(SWEEPARGS) -timeout $(SWEEP_TIMEOUT) -vet=off

sweeper: prereq-go ## Run sweepers with failures allowed
	@echo "WARNING: This will destroy infrastructure. Use only in development accounts."
	$(GO_VER) test $(SWEEP_DIR) -v -sweep=$(SWEEP) -sweep-allow-failures -timeout $(SWEEP_TIMEOUT) -vet=off

sweeper-check: sweeper-linked sweeper-unlinked ## [CI] Provider Checks / Sweeper Linked, Unlinked

sweeper-linked: ## [CI] Provider Checks / Sweeper Functions Linked
	@echo "make: Provider Checks / Sweeper Functions Linked..." ; \
	go test -c -o ./sweeper-bin ./internal/sweep/ ; \
	count=`strings ./sweeper-bin | \
		grep --count --extended-regexp 'internal/service/[a-zA-Z0-9]+\.sweep[a-zA-Z0-9]+$$'` ; \
	echo "make: sweeper-linked: found $$count, expected more than 0" ; \
	[ $$count -gt 0 ] || \
		(echo; echo "Expected `strings` to detect sweeper function names in sweeper binary."; exit 1)

sweeper-unlinked: go-build ## [CI] Provider Checks / Sweeper Functions Not Linked
	@os_arch=`go env GOOS`_`go env GOARCH` ; \
	echo "make: Provider Checks / Sweeper Functions Not Linked ($$os_arch)..." ; \
	count=`strings "terraform-plugin-dir/registry.terraform.io/hashicorp/aws/99.99.99/$$os_arch/terraform-provider-aws" | \
		grep --count --extended-regexp 'internal/service/[a-zA-Z0-9]+\.sweep[a-zA-Z0-9]+$$'` ; \
	echo "make: sweeper-unlinked: found $$count, expected 0" ; \
	[ $$count -eq 0 ] || \
		(echo "Expected `strings` to detect no sweeper function names in provider binary."; exit 1)

t: prereq-go fmt-check ## Run acceptance tests (similar to testacc)
	@branch=$$(git rev-parse --abbrev-ref HEAD); \
	printf "make: Running acceptance tests on branch: \033[1m%s\033[0m...\n" "🌿 $$branch 🌿"
	TF_ACC=1 $(GO_VER) test ./$(PKG_NAME)/... -v -count $(TEST_COUNT) -parallel $(ACCTEST_PARALLELISM) $(RUNARGS) $(TESTARGS) -timeout $(ACCTEST_TIMEOUT) -vet=off

test-compile: prereq-go ## Test package compilation
	@if [ "$(TEST)" = "./..." ]; then \
		echo "ERROR: Set TEST to a specific package. For example,"; \
		echo "  make test-compile TEST=./$(PKG_NAME)"; \
		exit 1; \
	fi
	$(GO_VER) test -c $(TEST) $(TESTARGS) -vet=off

test: prereq-go ## Run unit tests (auto-detects environment and scope)
	@branch=$$(git rev-parse --abbrev-ref HEAD); \
	printf "make: Running unit tests on branch: \033[1m%s\033[0m...\n" "🌿 $$branch 🌿"
	@# Auto-detect: single service or full codebase
	@if [ -n "$(PKG)$(K)" ]; then \
		echo "Testing single service: $(or $(PKG),$(K))"; \
		$(MAKE) test-single-service; \
	else \
		echo "Testing full codebase"; \
		$(MAKE) test-full; \
	fi

test-single-service: ## Internal: test single service
	@# macOS: use temp cache to avoid CrowdStrike scanning
	@if [ "$$(uname)" = "Darwin" ]; then \
		build_dir="/tmp/terraform-$(or $(PKG),$(K))-$$$$"; \
		mkdir -p "$$build_dir/cache"; \
		export GOCACHE="$$build_dir/cache"; \
		export GOTMPDIR="$$build_dir"; \
	fi; \
	cores=$$(getconf _NPROCESSORS_ONLN 2>/dev/null || nproc 2>/dev/null || echo 8); \
	test_parallel=$${TEST_PARALLEL:-$$cores}; \
	printf "Testing with -parallel %s\n" "$$test_parallel"; \
	$(GO_VER) test $(TEST) \
		-parallel $$test_parallel \
		-run '^Test[^A]|^TestA[^c]|^TestAc[^c]' \
		$(RUNARGS) $(TESTARGS) \
		-timeout 30m \
		-vet=off \
		-buildvcs=false \
		-count=1; \
	if [ "$$(uname)" = "Darwin" ] && [ -n "$$build_dir" ]; then rm -rf "$$build_dir"; fi

test-full: ## Internal: test full codebase
	@# macOS: use temp cache to avoid CrowdStrike scanning
	@if [ "$$(uname)" = "Darwin" ]; then \
		build_dir="/tmp/terraform-aws-build-$$$$"; \
		mkdir -p "$$build_dir/cache" "$$build_dir/tmp"; \
		export GOCACHE="$$build_dir/cache"; \
		export GOTMPDIR="$$build_dir/tmp"; \
	fi; \
	cores=$$(getconf _NPROCESSORS_ONLN 2>/dev/null || nproc 2>/dev/null || sysctl -n hw.ncpu 2>/dev/null || echo 8); \
	test_p=$${TEST_P:-$$cores}; \
	test_parallel=$${TEST_PARALLEL:-$$((cores * 2))}; \
	printf "Testing with -p %s -parallel %s\n" "$$test_p" "$$test_parallel"; \
	$(GO_VER) test $(TEST) \
		-p $$test_p \
		-parallel $$test_parallel \
		-run '^Test[^A]|^TestA[^c]|^TestAc[^c]' \
		$(RUNARGS) $(TESTARGS) \
		-timeout 60m \
		-vet=off \
		-buildvcs=false; \
	if [ "$$(uname)" = "Darwin" ] && [ -n "$$build_dir" ]; then rm -rf "$$build_dir"; fi

test-shard: prereq-go ## Run unit tests for a specific shard (CI only: SHARD=0 TOTAL_SHARDS=4)
	@if [ -z "$(SHARD)" ] || [ -z "$(TOTAL_SHARDS)" ]; then \
		echo "Error: SHARD and TOTAL_SHARDS must be set"; \
		echo "Example: make test-shard SHARD=0 TOTAL_SHARDS=4"; \
		exit 1; \
	fi
	@echo "Running shard $(SHARD) of $(TOTAL_SHARDS)..."
	@packages=$$($(GO_VER) list ./... | grep -v '/vendor/' | sort); \
	count=0; \
	shard_packages=""; \
	for pkg in $$packages; do \
		if [ $$((count % $(TOTAL_SHARDS))) -eq $(SHARD) ]; then \
			shard_packages="$$shard_packages $$pkg"; \
		fi; \
		count=$$((count + 1)); \
	done; \
	if [ -z "$$shard_packages" ]; then \
		echo "No packages assigned to shard $(SHARD)"; \
		exit 0; \
	fi; \
	cores=$$(getconf _NPROCESSORS_ONLN 2>/dev/null || nproc 2>/dev/null || sysctl -n hw.ncpu 2>/dev/null || echo 8); \
	test_p=$${TEST_P:-$$cores}; \
	test_parallel=$${TEST_PARALLEL:-$$((cores * 2))}; \
	echo "Testing $$( echo $$shard_packages | wc -w | xargs ) packages with -p $$test_p -parallel $$test_parallel"; \
	$(GO_VER) test $$shard_packages \
		-p $$test_p \
		-parallel $$test_parallel \
		-run '^Test[^A]|^TestA[^c]|^TestAc[^c]' \
		$(RUNARGS) $(TESTARGS) \
		-timeout 60m \
		-vet=off \
		-buildvcs=false

test-naming: ## Check test naming conventions
	@.ci/scripts/check-test-naming.sh

testacc: prereq-go fmt-check ## Run acceptance tests
	@branch=$$(git rev-parse --abbrev-ref HEAD); \
	printf "make: Running acceptance tests on branch: \033[1m%s\033[0m...\n" "🌿 $$branch 🌿"
	@if [ "$(TESTARGS)" = "-run=TestAccXXX" ]; then \
		echo ""; \
		echo "Error: Skipping example acceptance testing pattern. Update PKG and TESTS for the relevant *_test.go file."; \
		echo ""; \
		echo "For example if updating internal/service/acm/certificate.go, use the test names in internal/service/acm/certificate_test.go starting with TestAcc and up to the underscore:"; \
		echo "make testacc TESTS=TestAccACMCertificate_ PKG=acm"; \
		echo ""; \
		echo "See the contributing guide for more information: https://hashicorp.github.io/terraform-provider-aws/running-and-writing-acceptance-tests"; \
		exit 1; \
	fi
	TF_ACC=1 $(GO_VER) test ./$(PKG_NAME)/... -v -count $(TEST_COUNT) -parallel $(ACCTEST_PARALLELISM) $(RUNARGS) $(TESTARGS) -timeout $(ACCTEST_TIMEOUT) -vet=off

testacc-lint: ## [CI] Acceptance Test Linting / terrafmt
	@echo "make: Acceptance Test Linting / terrafmt..."
	@find $(SVC_DIR) -type f -name '*_test.go' \
		| sort -u \
		| xargs -I {} terrafmt diff --check --fmtcompat {}

testacc-lint-fix: ## Fix acceptance test linter findings
	@echo "make: Fixing Acceptance Test Linting / terrafmt..."
	@find $(SVC_DIR) -type f -name '*_test.go' \
		| sort -u \
		| xargs -I {} terrafmt fmt  --fmtcompat {}

testacc-lint-fix-core: ## Fix acceptance test linter findings in core directories
	@echo "make: Fixing Acceptance Test Linting / terrafmt in core directories..."
	@find . -name '*_test.go' -type f ! -path './internal/service/*' ! -path './.git/*' ! -path './vendor/*' ! -path './tools/*' \
		| sort -u \
		| xargs -I {} terrafmt fmt --fmtcompat {}

terraform-fmt: ## Format all .tf, .tfvars, .tftest.hcl, and .tfquery.hcl files
	@echo "make: Formatting .tf, .tfvars, .tftest.hcl, and .tfquery.hcl files..."
	@find . -name "*.tfquery.hcl" -type f -exec sh -c 'mv "$$1" "$${1%.tfquery.hcl}.BEGIANT.tf"' _ {} \;
	@terraform fmt -recursive .
	@find . -name "*.BEGIANT.tf" -type f -exec sh -c 'mv "$$1" "$${1%.BEGIANT.tf}.tfquery.hcl"' _ {} \;

testacc-short: prereq-go fmt-check ## Run acceptace tests with the -short flag
	@echo "Running acceptance tests with -short flag"
	TF_ACC=1 $(GO_VER) test ./$(PKG_NAME)/... -v -short -count $(TEST_COUNT) -parallel $(ACCTEST_PARALLELISM) $(RUNARGS) $(TESTARGS) -timeout $(ACCTEST_TIMEOUT) -vet=off

testacc-tflint: testacc-tflint-dir testacc-tflint-embedded ## [CI] Acceptance Test Linting / tflint

testacc-tflint-dir: tflint-init ## Run tflint on Terraform directories
	@echo "make: Acceptance Test Linting (standalone) / tflint..."
	@# tflint always resolves config flies relative to the working directory when using --recursive
	@tflint_config="$(PWD)/.ci/.tflint.hcl" ; \
	tflint --config  "$$tflint_config" --chdir=./internal/service --recursive

testacc-tflint-dir-fix: tflint-init ## fix Terraform directory linter findings
	@echo "make: Acceptance Test Linting (standalone) / tflint..."
	@# tflint always resolves config flies relative to the working directory when using --recursive
	@tflint_config="$(PWD)/.ci/.tflint.hcl" ; \
	tflint --config  "$$tflint_config" --chdir=./internal/service --recursive --fix

testacc-tflint-embedded: tflint-init ## Run tflint on embedded Terraform configs
	@echo "make: Acceptance Test Linting (embedded) / tflint..."
	@find $(SVC_DIR) -type f -name '*_test.go' \
		| .ci/scripts/validate-terraform.sh

tflint-init: ## Initialize tflint
	@tflint --config .ci/.tflint.hcl --init

tfproviderdocs: go-build ## [CI] Provider Checks / tfproviderdocs
	@echo "make: Provider Checks / tfproviderdocs..."
	@trap 'rm -rf terraform-providers-schema example.tf .terraform.lock.hcl' EXIT ; \
	rm -rf terraform-providers-schema example.tf .terraform.lock.hcl ; \
	echo 'data "aws_partition" "example" {}' > example.tf ; \
	terraform init -plugin-dir terraform-plugin-dir ; \
	mkdir -p terraform-providers-schema ; \
	terraform providers schema -json > terraform-providers-schema/schema.json ; \
	tfproviderdocs check \
		-allowed-resource-subcategories-file website/allowed-subcategories.txt \
		-enable-contents-check \
		-ignore-contents-check-data-sources aws_kms_secrets,aws_kms_secret \
		-ignore-file-missing-data-sources aws_alb,aws_alb_listener,aws_alb_target_group,aws_alb_trust_store,aws_alb_trust_store_revocation,aws_albs \
		-ignore-file-missing-resources aws_alb,aws_alb_listener,aws_alb_listener_certificate,aws_alb_listener_rule,aws_alb_target_group,aws_alb_target_group_attachment,aws_alb_trust_store,aws_alb_trust_store_revocation \
		-provider-source registry.terraform.io/hashicorp/aws \
		-providers-schema-json terraform-providers-schema/schema.json \
		-require-resource-subcategory \
		-ignore-cdktf-missing-files \
		-ignore-enhanced-region-check-subcategories-file website/ignore-enhanced-region-check-subcategories.txt \
		-ignore-enhanced-region-check-data-sources-file website/ignore-enhanced-region-check-data-sources.txt \
		-ignore-enhanced-region-check-resources-file website/ignore-enhanced-region-check-resources.txt \
		-enable-enhanced-region-check

tfsdk2fw: prereq-go ## Install tfsdk2fw
	@echo "make: Installing tfsdk2fw..."
	cd tools/tfsdk2fw && $(GO_VER) install github.com/hashicorp/terraform-provider-aws/tools/tfsdk2fw

tools: prereq-go ## Install tools
	@echo "make: Installing tools..."
	cd .ci/providerlint && $(GO_VER) install .
	cd .ci/tools && $(GO_VER) install github.com/YakDriver/tfproviderdocs
	cd .ci/tools && $(GO_VER) install github.com/client9/misspell/cmd/misspell
	cd .ci/tools && $(GO_VER) install github.com/golangci/golangci-lint/v2/cmd/golangci-lint
	cd .ci/tools && $(GO_VER) install github.com/YakDriver/copyplop
	cd .ci/tools && $(GO_VER) install github.com/hashicorp/go-changelog/cmd/changelog-build
	cd .ci/tools && $(GO_VER) install github.com/katbyte/terrafmt
	cd .ci/tools && $(GO_VER) install github.com/pavius/impi/cmd/impi
	cd .ci/tools && $(GO_VER) install github.com/rhysd/actionlint/cmd/actionlint
	cd .ci/tools && $(GO_VER) install github.com/terraform-linters/tflint
	cd .ci/tools && $(GO_VER) install golang.org/x/tools/cmd/stringer
	cd .ci/tools && $(GO_VER) install mvdan.cc/gofumpt

ts: testacc-short ## Alias to testacc-short

update: prereq-go ## Update dependencies
	@echo "make: Updating dependencies..."
	$(GO_VER) get -u ./...
	$(GO_VER) mod tidy
	cd ./tools/literally && $(GO_VER) get -u ./... && $(GO_VER) mod tidy
	cd ./tools/tfsdk2fw && $(GO_VER) get -u ./... && $(GO_VER) mod tidy
	cd .ci/tools && $(GO_VER) get -u && $(GO_VER) mod tidy
	cd .ci/providerlint && $(GO_VER) get -u && $(GO_VER) mod tidy
	cd .ci/providerlint/passes/AWSAT005/testdata && $(GO_VER) get -u ./... && $(GO_VER) mod tidy
	cd .ci/providerlint/passes/AWSAT002/testdata && $(GO_VER) get -u ./... && $(GO_VER) mod tidy
	cd .ci/providerlint/passes/AWSAT003/testdata && $(GO_VER) get -u ./... && $(GO_VER) mod tidy
	cd .ci/providerlint/passes/AWSAT004/testdata && $(GO_VER) get -u ./... && $(GO_VER) mod tidy
	cd .ci/providerlint/passes/AWSV001/testdata && $(GO_VER) get -u ./... && $(GO_VER) mod tidy
	cd .ci/providerlint/passes/AWSR001/testdata && $(GO_VER) get -u ./... && $(GO_VER) mod tidy
	cd .ci/providerlint/passes/AWSAT001/testdata && $(GO_VER) get -u ./... && $(GO_VER) mod tidy
	cd .ci/providerlint/passes/AWSAT006/testdata && $(GO_VER) get -u ./... && $(GO_VER) mod tidy
	cd ./skaff && $(GO_VER) get -u ./... && $(GO_VER) mod tidy

vcr-enable: ## Enable VCR testing
	$(MAKE) semgrep-vcr || true
	$(MAKE) semgrep-vcr || true
	$(MAKE) fmt
	goimports -w ./$(PKG_NAME)/

website: website-link-check-markdown website-link-check-md website-markdown-lint website-misspell website-terrafmt website-tflint ## [CI] Run all CI website checks

website-link-check: ## Check website links (Legacy, use caution)
	@echo "make: Legacy target, use caution..."
	@.ci/scripts/markdown-link-check.sh

website-link-check-ghrc: ## Check website links with ghrc (Legacy, use caution)
	@echo "make: Legacy target, use caution..."
	@LINK_CHECK_CONTAINER="ghcr.io/tcort/markdown-link-check:stable" .ci/scripts/markdown-link-check.sh

website-link-check-markdown: ## [CI] Website Checks / markdown-link-check-a-z-markdown
	@echo "make: Website Checks / markdown-link-check-a-z-markdown..."
	@docker run --rm \
		-v "$(PWD):/markdown" \
		ghcr.io/yakdriver/md-check-links:2.2.0 \
		--config /markdown/.ci/.markdownlinkcheck.json \
		--verbose yes \
		--quiet yes \
		--directory "/markdown/website/docs/r, /markdown/website/docs/d" \
		--extension ".markdown" \
		--modified no

website-link-check-md: ## [CI] Website Checks / markdown-link-check-md
	@echo "make: Website Checks / markdown-link-check-md..."
	@docker run --rm \
		-v "$(PWD):/markdown" \
		ghcr.io/yakdriver/md-check-links:2.2.0 \
		--config /markdown/.ci/.markdownlinkcheck.json \
		--verbose yes \
		--quiet yes \
		--directory /markdown/website/docs \
		--extension '.md' \
		--depth 2

website-lint: ## Lint website files (Legacy, use caution)
	@echo "make: Legacy target, use caution..."
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

website-lint-fix: ## Fix website linter findings (Legacy, use caution)
	@echo "make: Legacy target, use caution..."
	@misspell -w -source=text website/
	@docker run --rm -v $(PWD):/markdown 06kellyjac/markdownlint-cli --fix website/docs/
	@terrafmt fmt ./website --pattern '*.markdown'

website-markdown-lint: ## [CI] Website Checks / markdown-lint
	@echo "make: Website Checks / markdown-lint..."
	@docker run --rm \
		-v "$(PWD):/markdown" \
		avtodev/markdown-lint:v1.5.0 \
		--config /markdown/.markdownlint.yml \
		--ignore /markdown/website/docs/cdktf/python/guides \
		--ignore /markdown/website/docs/cdktf/typescript/guides \
		/markdown/website/docs

website-misspell: ## [CI] Website Checks / misspell
	@echo "make: Website Checks / misspell..."
	@misspell -error -source text website/

website-terrafmt: ## [CI] Website Checks / terrafmt
	@echo "make: Website Checks / terrafmt..."
	@terrafmt diff ./website --check --pattern '*.markdown'

website-terrafmt-fix: ## [CI] Fix Website / terrafmt
	@echo "make: Fix Website / terrafmt..."
	@echo "make: Fixing website/docs root files with terrafmt..."
	@find ./website/docs -maxdepth 1 -type f -name '*.markdown' -exec terrafmt fmt {} \;
	@for dir in $$(find ./website/docs -maxdepth 1 -type d ! -name docs ! -name cdktf | sort); do \
		echo "make: Fixing $$dir with terrafmt..."; \
		terrafmt fmt $$dir --pattern '*.markdown'; \
	done

website-tflint: tflint-init ## [CI] Website Checks / tflint
	@echo "make: Website Checks / tflint..."
	@exit_code=0 ; \
	shared_rules=( \
		"--disable-rule=aws_cloudwatch_event_target_invalid_arn" \
		"--disable-rule=aws_db_instance_default_parameter_group" \
		"--disable-rule=aws_elasticache_cluster_default_parameter_group" \
		"--disable-rule=aws_elasticache_replication_group_default_parameter_group" \
		"--disable-rule=aws_iam_policy_sid_invalid_characters" \
		"--disable-rule=aws_iam_saml_provider_invalid_saml_metadata_document" \
		"--disable-rule=aws_iam_server_certificate_invalid_certificate_body" \
		"--disable-rule=aws_iam_server_certificate_invalid_private_key" \
		"--disable-rule=aws_iot_certificate_invalid_csr" \
		"--disable-rule=aws_lb_invalid_load_balancer_type" \
		"--disable-rule=aws_lb_target_group_invalid_protocol" \
		"--disable-rule=aws_networkfirewall_rule_group_invalid_rules" \
		"--disable-rule=aws_s3_object_copy_invalid_source" \
		"--disable-rule=aws_servicecatalog_portfolio_share_invalid_type" \
		"--disable-rule=aws_transfer_ssh_key_invalid_body" \
		"--disable-rule=terraform_unused_declarations" \
		"--disable-rule=terraform_typed_variables" \
	) ; \
	while read -r filename; do \
		rules=("$${shared_rules[@]}") ; \
		if [[ "$$filename" == "./website/docs/guides/version-2-upgrade.html.markdown" ]] ; then \
			rules+=( \
			"--disable-rule=terraform_deprecated_index" \
			"--disable-rule=terraform_deprecated_interpolation" \
			) ; \
		elif [[ "$$filename" == "./website/docs/guides/version-3-upgrade.html.markdown" ]]; then \
			rules+=( \
			"--enable-rule=terraform_deprecated_index" \
			"--disable-rule=terraform_deprecated_interpolation" \
			) ; \
		else \
			rules+=( \
			"--enable-rule=terraform_deprecated_index" \
			"--enable-rule=terraform_deprecated_interpolation" \
			) ; \
		fi ; \
		set +e ; \
		./.ci/scripts/validate-terraform-file.sh "$$filename" "$${rules[@]}" || exit_code=1 ; \
		set -e ; \
	done < <(find ./website/docs -not \( -path ./website/docs/cdktf -prune \) -type f -name '*.markdown' | sort -u) ; \
	exit $$exit_code

yamllint: ## [CI] YAML Linting / yamllint
	@echo "make: YAML Linting / yamllint..."
	@yamllint .

# Please keep targets in alphabetical order
.PHONY: \
	acctest-lint \
	build \
	cache-info \
	changelog-misspell \
	ci \
	ci-quick \
	clean \
	clean-go \
	clean-go-cache-trim \
	clean-make-tests \
	clean-tidy \
	copyright \
	default \
	deps-check \
	docs \
	docs-check \
	docs-link-check \
	docs-lint \
	docs-lint-fix \
	docs-markdown-lint \
	docs-misspell \
	examples-tflint \
	fix-constants \
	fix-imports \
	fix-imports-core \
	fmt \
	fmt-check \
	fmt-core \
	fumpt \
	gen \
	gen-check \
	gen-raw \
	generate-changelog \
	gh-workflows-lint \
	go-build \
	go-misspell \
	golangci-lint \
	golangci-lint1 \
	golangci-lint2 \
	golangci-lint3 \
	golangci-lint4 \
	golangci-lint5 \
	help \
	import-lint \
	install \
	lint \
	lint-fix \
	misspell \
	modern-check \
	modern-fix \
	modern-fix-core \
	pr-target-check \
	prereq-go \
	provider-lint \
	provider-markdown-lint \
	quick-fix \
	quick-fix-core \
	quick-fix-core-heading \
	quick-fix-heading \
	sane \
	sanity \
	semgrep \
	semgrep-all \
	semgrep-code-quality \
	semgrep-constants \
	semgrep-docker \
	semgrep-fix \
	semgrep-fix-core \
	semgrep-naming \
	semgrep-naming-cae \
	semgrep-service-naming \
	semgrep-validate \
	semgrep-vcr \
	skaff \
	skaff-check-compile \
	smoke \
	sweep \
	sweeper \
	sweeper-check \
	sweeper-linked \
	sweeper-unlinked \
	t \
	test \
	test-compile \
	test-full \
	test-shard \
	test-single-service \
	testacc \
	testacc-lint \
	testacc-lint-fix \
	testacc-lint-fix-core \
	testacc-short \
	testacc-tflint \
	testacc-tflint-dir \
	testacc-tflint-embedded \
	terraform-fmt \
	tflint-init \
	tfproviderdocs \
	tfsdk2fw \
	tools \
	ts \
	update \
	vcr-enable \
	website \
	website-link-check \
	website-link-check-ghrc \
	website-link-check-markdown \
	website-link-check-md \
	website-lint \
	website-lint-fix \
	website-markdown-lint \
	website-misspell \
	website-terrafmt \
	website-terrafmt-fix \
	website-tflint \
	yamllint
