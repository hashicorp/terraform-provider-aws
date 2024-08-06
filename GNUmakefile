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

ifneq ($(origin PKG), undefined)
	PKG_NAME = internal/service/$(PKG)
	SVC_DIR = ./internal/service/$(PKG)
	TEST = ./$(PKG_NAME)/...
endif

ifneq ($(origin K), undefined)
	PKG_NAME = internal/service/$(K)
	SVC_DIR = ./internal/service/$(PKG)
	TEST = ./$(PKG_NAME)/...
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

awssdkpatch: prereq-go ## Install awssdkpatch
	cd tools/awssdkpatch && $(GO_VER) install github.com/hashicorp/terraform-provider-aws/tools/awssdkpatch

awssdkpatch-apply: awssdkpatch-gen ## Apply a patch generated with awssdkpatch
	@echo "Applying patch for $(PKG)..."
	@gopatch --skip-generated -p awssdk.patch ./$(PKG_NAME)/...

awssdkpatch-gen: awssdkpatch ## Generate a patch file using awssdkpatch
	@if [ "$(PKG)" = "" ]; then \
		echo "PKG must be set. Try again like:" ; \
		echo "PKG=foo make awssdkpatch-gen" ; \
		exit 1 ; \
	fi
	@awssdkpatch $(AWSSDKPATCH_OPTS) -service $(PKG)

build: prereq-go fmt-check ## Build provider
	@echo "make: Building provider..."
	@$(GO_VER) install

changelog-misspell: ## [CI] CHANGELOG Misspell / misspell
	@echo "make: CHANGELOG Misspell / misspell..."
	@misspell -error -source text CHANGELOG.md .changelog

ci: tools go-build gen-check acctest-lint copyright deps-check docs examples-tflint gh-workflow-lint golangci-lint import-lint preferred-lib provider-lint provider-markdown-lint semgrep skaff-check-compile sweeper-check test tfproviderdocs website yamllint ## [CI] Run all CI checks

ci-quick: tools go-build testacc-lint copyright deps-check docs examples-tflint gh-workflow-lint golangci-lint1 import-lint preferred-lib provider-lint provider-markdown-lint semgrep-code-quality semgrep-naming semgrep-naming-cae website-markdown-lint website-misspell website-terrafmt yamllint ## [CI] Run quicker CI checks

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
	cd tools/awssdkpatch && $$gover mod tidy && cd ../.. ; \
	cd tools/tfsdk2fw && $$gover mod tidy && cd ../.. ; \
	cd .ci/tools && $$gover mod tidy && cd ../.. ; \
	cd .ci/providerlint && $$gover mod tidy && cd ../.. ; \
	cd skaff && $$gover mod tidy && cd .. ; \
	$$gover mod tidy
	@echo "make: Go mods tidied"

copyright: ## [CI] Copyright Checks / add headers check
	@echo "make: Copyright Checks / add headers check..."
	@copywrite headers

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
		-ignore-file-missing-data-sources aws_alb,aws_alb_listener,aws_alb_target_group,aws_albs \
		-ignore-file-missing-resources aws_alb,aws_alb_listener,aws_alb_listener_certificate,aws_alb_listener_rule,aws_alb_target_group,aws_alb_target_group_attachment \
		-provider-name=aws \
		-require-resource-subcategory

docs-link-check: ## [CI] Documentation Checks / markdown-link-check
	@echo "make: Documentation Checks / markdown-link-check..."
	@docker run --rm \
		-v "$(PWD):/markdown" \
		ghcr.io/yakdriver/md-check-links:2.1.0 \
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

examples-tflint: ## [CI] Examples Checks / tflint
	@echo "make: Examples Checks / tflint..."
	@tflint --config .ci/.tflint.hcl --init
	@exit_code=0 ; \
	TFLINT_CONFIG="`pwd -P`/.ci/.tflint.hcl" ; \
	for DIR in `find ./examples -type f -name '*.tf' -exec dirname {} \; | sort -u`; do \
		pushd "$$DIR" ; \
		tflint --config="$$TFLINT_CONFIG" \
			--enable-rule=terraform_comment_syntax \
			--enable-rule=terraform_deprecated_index \
			--enable-rule=terraform_deprecated_interpolation \
			--enable-rule=terraform_required_version \
			--disable-rule=terraform_required_providers \
			--disable-rule=terraform_typed_variables \
			|| exit_code=1 ; \
		popd ; \
	done ; \
	exit $$exit_code

fix-constants: semgrep-constants fmt ## Use Semgrep to fix constants

fix-imports: ## Fixing source code imports with goimports
	@echo "make: Fixing source code imports with goimports..."
	@find internal -name "*.go" -type f -exec goimports -w {} \;

fmt: ## Fix Go source formatting
	@echo "make: Fixing source code with gofmt..."
	gofmt -s -w ./$(PKG_NAME) ./names $(filter-out ./.ci/providerlint/go% ./.ci/providerlint/README.md ./.ci/providerlint/vendor, $(wildcard ./.ci/providerlint/*))

# Currently required by tf-deploy compile
fmt-check: ## Verify Go source is formatted
	@echo "make: Verifying source code with gofmt..."
	@sh -c "'$(CURDIR)/.ci/scripts/gofmtcheck.sh'"

fumpt: ## Run gofumpt
	@echo "make: Fixing source code with gofumpt..."
	gofumpt -w ./$(PKG_NAME) ./names $(filter-out ./.ci/providerlint/go% ./.ci/providerlint/README.md ./.ci/providerlint/vendor, $(wildcard ./.ci/providerlint/*))

gen: prereq-go ## Run all Go generators
	@echo "make: Running Go generators..."
	$(GO_VER) generate ./...
	# Generate service package lists last as they may depend on output of earlier generators.
	$(GO_VER) generate ./internal/provider
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
	@misspell -error -source auto -i "littel,ceasar" internal/

golangci-lint: golangci-lint1 golangci-lint2 ## [CI] All golangci-lint Checks

golangci-lint1: ## [CI] golangci-lint Checks / 1 of 2
	@echo "make: golangci-lint Checks / 1 of 2..."
	@golangci-lint run \
		--config .ci/.golangci.yml \
		$(TEST)

golangci-lint2: ## [CI] golangci-lint Checks / 2 of 2
	@echo "make: golangci-lint Checks / 2 of 2..."
	@golangci-lint run \
		--config .ci/.golangci2.yml \
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

preferred-lib: ## [CI] Preferred Library Version Check / diffgrep
	@echo "make: Preferred Library Version Check / diffgrep..."
	@found=`git diff origin/$(BASE_REF) internal/ | grep '^\+\s*"github.com/aws/aws-sdk-go/'` ; \
	if [ "$$found" != "" ] ; then \
		echo "Found a new reference to github.com/aws/aws-sdk-go in the codebase. Please use the preferred library github.com/aws/aws-sdk-go-v2 instead." ; \
		exit 1 ; \
	fi

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

sane: prereq-go ## Run sane check
	@echo "make: Sane Check (48 tests of Top 30 resources)"
	@echo "make: Like 'sanity' except full output and stops soon after 1st error"
	@echo "make: NOTE: NOT an exhaustive set of tests! Finds big problems only."
	@TF_ACC=1 $(GO_VER) test \
		./internal/service/iam/... \
		-v -count $(TEST_COUNT) -parallel $(ACCTEST_PARALLELISM) -run='TestAccIAMRole_basic|TestAccIAMRole_namePrefix|TestAccIAMRole_disappears|TestAccIAMRole_InlinePolicy_basic|TestAccIAMPolicyDocumentDataSource_basic|TestAccIAMPolicyDocumentDataSource_sourceConflicting|TestAccIAMPolicyDocumentDataSource_sourceJSONValidJSON|TestAccIAMRolePolicyAttachment_basic|TestAccIAMRolePolicyAttachment_disappears|TestAccIAMRolePolicyAttachment_Disappears_role|TestAccIAMPolicy_basic|TestAccIAMPolicy_policy|TestAccIAMPolicy_tags|TestAccIAMRolePolicy_basic|TestAccIAMRolePolicy_unknownsInPolicy|TestAccIAMInstanceProfile_basic|TestAccIAMInstanceProfile_tags' -timeout $(ACCTEST_TIMEOUT)
	@TF_ACC=1 $(GO_VER) test \
		./internal/service/logs/... \
		./internal/service/ec2/... \
		./internal/service/ecs/... \
		./internal/service/elbv2/... \
		./internal/service/kms/... \
		-v -count $(TEST_COUNT) -parallel $(ACCTEST_PARALLELISM) -run='TestAccVPCSecurityGroup_basic|TestAccVPCSecurityGroup_egressMode|TestAccVPCSecurityGroup_vpcAllEgress|TestAccVPCSecurityGroupRule_race|TestAccVPCSecurityGroupRule_protocolChange|TestAccVPCDataSource_basic|TestAccVPCSubnet_basic|TestAccVPC_tenancy|TestAccVPCRouteTableAssociation_Subnet_basic|TestAccVPCRouteTable_basic|TestAccLogsGroup_basic|TestAccLogsGroup_multiple|TestAccKMSKey_basic|TestAccELBV2TargetGroup_basic|TestAccECSTaskDefinition_basic|TestAccECSService_basic' -timeout $(ACCTEST_TIMEOUT)
	@TF_ACC=1 $(GO_VER) test \
		./internal/service/lambda/... \
		./internal/service/meta/... \
		./internal/service/route53/... \
		./internal/service/s3/... \
		./internal/service/secretsmanager/... \
		./internal/service/sts/... \
		-v -count $(TEST_COUNT) -parallel $(ACCTEST_PARALLELISM) -run='TestAccSTSCallerIdentityDataSource_basic|TestAccMetaRegionDataSource_basic|TestAccMetaRegionDataSource_endpoint|TestAccMetaPartitionDataSource_basic|TestAccS3Bucket_Basic_basic|TestAccS3Bucket_Security_corsUpdate|TestAccS3BucketPublicAccessBlock_basic|TestAccS3BucketPolicy_basic|TestAccS3BucketACL_updateACL|TestAccRoute53Record_basic|TestAccRoute53Record_Latency_basic|TestAccRoute53ZoneDataSource_name|TestAccLambdaFunction_basic|TestAccLambdaPermission_basic|TestAccSecretsManagerSecret_basic' -timeout $(ACCTEST_TIMEOUT)

sanity: prereq-go ## Run sanity check (failures allowed)
	@echo "make: Sanity Check (48 tests of Top 30 resources)"
	@echo "make: Like 'sane' but less output and runs all tests despite most errors"
	@echo "make: NOTE: NOT an exhaustive set of tests! Finds big problems only."
	@iam=`TF_ACC=1 $(GO_VER) test \
		./internal/service/iam/... \
		-v -count $(TEST_COUNT) -parallel $(ACCTEST_PARALLELISM) -run='TestAccIAMRole_basic|TestAccIAMRole_namePrefix|TestAccIAMRole_disappears|TestAccIAMRole_InlinePolicy_basic|TestAccIAMPolicyDocumentDataSource_basic|TestAccIAMPolicyDocumentDataSource_sourceConflicting|TestAccIAMPolicyDocumentDataSource_sourceJSONValidJSON|TestAccIAMRolePolicyAttachment_basic|TestAccIAMRolePolicyAttachment_disappears|TestAccIAMRolePolicyAttachment_Disappears_role|TestAccIAMPolicy_basic|TestAccIAMPolicy_policy|TestAccIAMPolicy_tags|TestAccIAMRolePolicy_basic|TestAccIAMRolePolicy_unknownsInPolicy|TestAccIAMInstanceProfile_basic|TestAccIAMInstanceProfile_tags' -timeout $(ACCTEST_TIMEOUT) || true` ; \
	fails1=`echo -n $$iam | grep -Fo FAIL: | wc -l | xargs` ; \
	passes=$$(( 17-$$fails1 )) ; \
	echo "17 of 48 complete: $$passes passed, $$fails1 failed" ; \
	logs=`TF_ACC=1 $(GO_VER) test \
		./internal/service/logs/... \
		./internal/service/ec2/... \
		./internal/service/ecs/... \
		./internal/service/elbv2/... \
		./internal/service/kms/... \
		-v -count $(TEST_COUNT) -parallel $(ACCTEST_PARALLELISM) -run='TestAccVPCSecurityGroup_basic|TestAccVPCSecurityGroup_egressMode|TestAccVPCSecurityGroup_vpcAllEgress|TestAccVPCSecurityGroupRule_race|TestAccVPCSecurityGroupRule_protocolChange|TestAccVPCDataSource_basic|TestAccVPCSubnet_basic|TestAccVPC_tenancy|TestAccVPCRouteTableAssociation_Subnet_basic|TestAccVPCRouteTable_basic|TestAccLogsGroup_basic|TestAccLogsGroup_multiple|TestAccKMSKey_basic|TestAccELBV2TargetGroup_basic|TestAccECSTaskDefinition_basic|TestAccECSService_basic' -timeout $(ACCTEST_TIMEOUT) || true` ; \
	fails2=`echo -n $$logs | grep -Fo FAIL: | wc -l | xargs` ; \
	tot_fails=$$(( $$fails1+$$fails2 )) ; \
	passes=$$(( 33-$$tot_fails )) ; \
	echo "33 of 48 complete: $$passes passed, $$tot_fails failed" ; \
	lambda=`TF_ACC=1 $(GO_VER) test \
		./internal/service/lambda/... \
		./internal/service/meta/... \
		./internal/service/route53/... \
		./internal/service/s3/... \
		./internal/service/secretsmanager/... \
		./internal/service/sts/... \
		-v -count $(TEST_COUNT) -parallel $(ACCTEST_PARALLELISM) -run='TestAccSTSCallerIdentityDataSource_basic|TestAccMetaRegionDataSource_basic|TestAccMetaRegionDataSource_endpoint|TestAccMetaPartitionDataSource_basic|TestAccS3Bucket_Basic_basic|TestAccS3Bucket_Security_corsUpdate|TestAccS3BucketPublicAccessBlock_basic|TestAccS3BucketPolicy_basic|TestAccS3BucketACL_updateACL|TestAccRoute53Record_basic|TestAccRoute53Record_Latency_basic|TestAccRoute53ZoneDataSource_name|TestAccLambdaFunction_basic|TestAccLambdaPermission_basic|TestAccSecretsManagerSecret_basic' -timeout $(ACCTEST_TIMEOUT) || true` ; \
	fails3=`echo -n $$lambda | grep -Fo FAIL: | wc -l | xargs` ; \
	tot_fails=$$(( $$fails1+$$fails2+$$fails3 )) ; \
	passes=$$(( 48-$$tot_fails )) ; \
	echo "48 of 48 complete: $$passes passed, $$tot_fails failed" ; \
	if [ $$tot_fails -gt 0 ] ; then \
		echo "Sanity tests failed"; \
		exit 1; \
	fi

semgrep: semgrep-code-quality semgrep-naming semgrep-naming-cae semgrep-service-naming ## [CI] Run all CI Semgrep checks

semgrep-all: semgrep-test semgrep-validate ## Run semgrep on all files
	@echo "make: Running Semgrep checks locally (must have semgrep installed)..."
	@semgrep $(SEMGREP_ARGS) \
		$(if $(filter-out $(origin PKG), undefined),--include $(PKG_NAME),) \
		--exclude .ci/semgrep/**/*.go \
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
		--exclude .ci/semgrep/**/*.go \
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
		--exclude .ci/semgrep/**/*.go \
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

skaff: prereq-go ## Install skaff
	@echo "make: Installing skaff..."
	cd skaff && $(GO_VER) install github.com/hashicorp/terraform-provider-aws/skaff

skaff-check-compile: ## [CI] Skaff Checks / Compile skaff
	@echo "make: Skaff Checks / Compile skaff..."
	@cd skaff ; \
	go build

sweep: prereq-go ## Run sweepers
	# make sweep SWEEPARGS=-sweep-run=aws_example_thing
	# set SWEEPARGS=-sweep-allow-failures to continue after first failure
	@echo "WARNING: This will destroy infrastructure. Use only in development accounts."
	$(GO_VER) test $(SWEEP_DIR) -v -sweep=$(SWEEP) $(SWEEPARGS) -timeout $(SWEEP_TIMEOUT)

sweeper: prereq-go ## Run sweepers with failures allowed
	@echo "WARNING: This will destroy infrastructure. Use only in development accounts."
	$(GO_VER) test $(SWEEP_DIR) -v -sweep=$(SWEEP) -sweep-allow-failures -timeout $(SWEEP_TIMEOUT)

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
	TF_ACC=1 $(GO_VER) test ./$(PKG_NAME)/... -v -count $(TEST_COUNT) -parallel $(ACCTEST_PARALLELISM) $(RUNARGS) $(TESTARGS) -timeout $(ACCTEST_TIMEOUT)

test: prereq-go fmt-check ## Run unit tests
	@echo "make: Running unit tests..."
	$(GO_VER) test -count $(TEST_COUNT) $(TEST) $(TESTARGS) -timeout=5m

test-compile: prereq-go ## Test package compilation
	@if [ "$(TEST)" = "./..." ]; then \
		echo "ERROR: Set TEST to a specific package. For example,"; \
		echo "  make test-compile TEST=./$(PKG_NAME)"; \
		exit 1; \
	fi
	$(GO_VER) test -c $(TEST) $(TESTARGS)

testacc: prereq-go fmt-check ## Run acceptance tests
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
	TF_ACC=1 $(GO_VER) test ./$(PKG_NAME)/... -v -count $(TEST_COUNT) -parallel $(ACCTEST_PARALLELISM) $(RUNARGS) $(TESTARGS) -timeout $(ACCTEST_TIMEOUT)

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

testacc-short: prereq-go fmt-check ## Run acceptace tests with the -short flag
	@echo "Running acceptance tests with -short flag"
	TF_ACC=1 $(GO_VER) test ./$(PKG_NAME)/... -v -short -count $(TEST_COUNT) -parallel $(ACCTEST_PARALLELISM) $(RUNARGS) $(TESTARGS) -timeout $(ACCTEST_TIMEOUT)

testacc-tflint: ## [CI] Acceptance Test Linting / tflint
	@echo "make: Acceptance Test Linting / tflint..."
	@tflint --config .ci/.tflint.hcl --init
	@find $(SVC_DIR) -type f -name '*_test.go' \
		| .ci/scripts/validate-terraform.sh

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
		-ignore-file-missing-data-sources aws_alb,aws_alb_listener,aws_alb_target_group,aws_alb_trust_store,aws_alb_trust_store_revocation,aws_albs \
		-ignore-file-missing-resources aws_alb,aws_alb_listener,aws_alb_listener_certificate,aws_alb_listener_rule,aws_alb_target_group,aws_alb_target_group_attachment,aws_alb_trust_store,aws_alb_trust_store_revocation \
		-provider-source registry.terraform.io/hashicorp/aws \
		-providers-schema-json terraform-providers-schema/schema.json \
		-require-resource-subcategory \
		-ignore-cdktf-missing-files

tfsdk2fw: prereq-go ## Install tfsdk2fw
	@echo "make: Installing tfsdk2fw..."
	cd tools/tfsdk2fw && $(GO_VER) install github.com/hashicorp/terraform-provider-aws/tools/tfsdk2fw

tools: prereq-go ## Install tools
	@echo "make: Installing tools..."
	cd .ci/providerlint && $(GO_VER) install .
	cd .ci/tools && $(GO_VER) install github.com/YakDriver/tfproviderdocs
	cd .ci/tools && $(GO_VER) install github.com/client9/misspell/cmd/misspell
	cd .ci/tools && $(GO_VER) install github.com/golangci/golangci-lint/cmd/golangci-lint
	cd .ci/tools && $(GO_VER) install github.com/hashicorp/copywrite
	cd .ci/tools && $(GO_VER) install github.com/hashicorp/go-changelog/cmd/changelog-build
	cd .ci/tools && $(GO_VER) install github.com/katbyte/terrafmt
	cd .ci/tools && $(GO_VER) install github.com/pavius/impi/cmd/impi
	cd .ci/tools && $(GO_VER) install github.com/rhysd/actionlint/cmd/actionlint
	cd .ci/tools && $(GO_VER) install github.com/terraform-linters/tflint
	cd .ci/tools && $(GO_VER) install github.com/uber-go/gopatch
	cd .ci/tools && $(GO_VER) install mvdan.cc/gofumpt

ts: testacc-short ## Alias to testacc-short

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
		ghcr.io/yakdriver/md-check-links:2.1.0 \
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
		ghcr.io/yakdriver/md-check-links:2.1.0 \
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

website-tflint: ## [CI] Website Checks / tflint
	@echo "make: Website Checks / tflint..."
	@tflint --config .ci/.tflint.hcl --init
	@exit_code=0 ; \
	shared_rules=( \
		"--enable-rule=terraform_comment_syntax" \
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
		"--disable-rule=aws_worklink_website_certificate_authority_association_invalid_certificate" \
		"--disable-rule=terraform_required_providers" \
		"--disable-rule=terraform_unused_declarations" \
		"--disable-rule=terraform_typed_variables" \
	) ; \
	while read -r filename; do \
		rules=("$${shared_rules[@]}") ; \
		if [[ "$$filename" == "./website/docs/guides/version-2-upgrade.html.md" ]] ; then \
			rules+=( \
			"--disable-rule=terraform_deprecated_index" \
			"--disable-rule=terraform_deprecated_interpolation" \
			) ; \
		elif [[ "$$filename" == "./website/docs/guides/version-3-upgrade.html.md" ]]; then \
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
	done < <(find ./website/docs -type f \( -name '*.md' -o -name '*.markdown' \) | sort -u) ; \
	exit $$exit_code

yamllint: ## [CI] YAML Linting / yamllint
	@echo "make: YAML Linting / yamllint..."
	@yamllint .

# Please keep targets in alphabetical order
.PHONY: \
	acctest-lint \
	awssdkpatch-apply \
	awssdkpatch-gen \
	awssdkpatch \
	build \
	changelog-misspell \
	ci-quick \
	ci \
	clean-go \
	clean-make-tests \
	clean-tidy \
	clean \
	copyright \
	default \
	deps-check \
	docs-check \
	docs-link-check \
	docs-lint-fix \
	docs-lint \
	docs-markdown-lint \
	docs-misspell \
	docs \
	examples-tflint \
	fix-constants \
	fix-imports \
	fmt-check \
	fmt \
	fumpt \
	gen-check \
	gen \
	generate-changelog \
	gh-workflows-lint \
	go-build \
	go-misspell \
	golangci-lint1 \
	golangci-lint2 \
	golangci-lint \
	help \
	import-lint \
	install \
	lint-fix \
	lint \
	misspell \
	preferred-lib \
	prereq-go \
	provider-lint \
	provider-markdown-lint \
	sane \
	sanity \
	semgrep-all \
	semgrep-code-quality \
	semgrep-constants \
	semgrep-docker \
	semgrep-fix \
	semgrep-naming-cae \
	semgrep-naming \
	semgrep-service-naming \
	semgrep-validate \
	semgrep \
	skaff-check-compile \
	skaff \
	sweep \
	sweeper-check \
	sweeper-linked \
	sweeper-unlinked \
	sweeper \
	t \
	test-compile \
	test \
	testacc-lint-fix \
	testacc-lint \
	testacc-short \
	testacc-tflint \
	testacc \
	tfproviderdocs \
	tfsdk2fw \
	tools \
	ts \
	website-link-check-ghrc \
	website-link-check-markdown \
	website-link-check-md \
	website-link-check \
	website-lint-fix \
	website-lint \
	website-markdown-lint \
	website-misspell \
	website-terrafmt \
	website-tflint \
	website \
	yamllint
