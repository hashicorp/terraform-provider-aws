SWEEP               ?= us-west-2,us-east-1,us-east-2,us-west-1
TEST                ?= ./...
SWEEP_DIR           ?= ./internal/sweep
PKG_NAME            ?= internal
SVC_DIR             ?= ./internal/service
TEST_COUNT          ?= 1
ACCTEST_TIMEOUT     ?= 180m
ACCTEST_PARALLELISM ?= 20
P                   ?= 20
GO_VER              ?= go
SWEEP_TIMEOUT       ?= 60m

ifneq ($(origin PKG), undefined)
	PKG_NAME = internal/service/$(PKG)
	TEST = ./$(PKG_NAME)/...
endif

ifneq ($(origin K), undefined)
	PKG_NAME = internal/service/$(K)
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

ifeq ($(PKG_NAME), internal/service/ebs)
	PKG_NAME = internal/service/ec2
	TEST = ./$(PKG_NAME)/...
endif

ifeq ($(PKG_NAME), internal/service/ipam)
	PKG_NAME = internal/service/ec2
	TEST = ./$(PKG_NAME)/...
endif

ifeq ($(PKG_NAME), internal/service/transitgateway)
	PKG_NAME = internal/service/ec2
	TEST = ./$(PKG_NAME)/...
endif

ifeq ($(PKG_NAME), internal/service/vpc)
	PKG_NAME = internal/service/ec2
	TEST = ./$(PKG_NAME)/...
endif

ifeq ($(PKG_NAME), internal/service/vpnclient)
	PKG_NAME = internal/service/ec2
	TEST = ./$(PKG_NAME)/...
endif

ifeq ($(PKG_NAME), internal/service/vpnsite)
	PKG_NAME = internal/service/ec2
	TEST = ./$(PKG_NAME)/...
endif

ifeq ($(PKG_NAME), internal/service/wavelength)
	PKG_NAME = internal/service/ec2
	TEST = ./$(PKG_NAME)/...
endif

ifneq ($(P), 20)
	ACCTEST_PARALLELISM = $(P)
endif

default: build

# Please keep targets in alphabetical order

build: fmtcheck
	$(GO_VER) install

cleango:
	@echo "==> Cleaning Go..."
	@echo "WARNING: This will kill gopls and clean Go caches"
	@vscode=`ps -ef | grep Visual\ Studio\ Code | wc -l | xargs` ; \
	if [ $$vscode -gt 1 ] ; then \
		echo "ALERT: vscode is running. Close it and try again." ; \
		exit 1 ; \
	fi
	@for proc in `pgrep gopls` ; do \
		echo "Killing gopls process $$proc" ; \
		kill -9 $$proc ; \
	done ; \
	go clean -modcache -testcache -cache ; \

clean: cleango build tools

copyright:
	@copywrite headers

depscheck:
	@echo "==> Checking source code with go mod tidy..."
	@$(GO_VER) mod tidy
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
		-enable-contents-check \
		-ignore-file-missing-data-sources aws_alb,aws_alb_listener,aws_alb_target_group,aws_albs \
		-ignore-file-missing-resources aws_alb,aws_alb_listener,aws_alb_listener_certificate,aws_alb_listener_rule,aws_alb_target_group,aws_alb_target_group_attachment \
		-provider-name=aws \
		-require-resource-subcategory
	@misspell -error -source text CHANGELOG.md .changelog

fmt:
	@echo "==> Fixing source code with gofmt..."
	gofmt -s -w ./$(PKG_NAME) ./names $(filter-out ./.ci/providerlint/go% ./.ci/providerlint/README.md ./.ci/providerlint/vendor, $(wildcard ./.ci/providerlint/*))

# Currently required by tf-deploy compile
fmtcheck:
	@sh -c "'$(CURDIR)/.ci/scripts/gofmtcheck.sh'"

fumpt:
	@echo "==> Fixing source code with gofumpt..."
	gofumpt -w ./$(PKG_NAME) ./names $(filter-out ./.ci/providerlint/go% ./.ci/providerlint/README.md ./.ci/providerlint/vendor, $(wildcard ./.ci/providerlint/*))

gen:
	rm -f .github/labeler-issue-triage.yml
	rm -f .github/labeler-pr-triage.yml
	rm -f infrastructure/repository/labels-service.tf
	rm -f internal/conns/*_gen.go
	rm -f internal/provider/*_gen.go
	rm -f internal/service/**/*_gen.go
	rm -f internal/sweep/sweep_test.go internal/sweep/service_packages_gen_test.go
	rm -f names/caps.md
	rm -f names/*_gen.go
	rm -f website/docs/guides/custom-service-endpoints.html.md
	rm -f .ci/.semgrep-caps-aws-ec2.yml
	rm -f .ci/.semgrep-configs.yml
	rm -f .ci/.semgrep-service-name*.yml
	$(GO_VER) generate ./...
	# Generate service package lists last as they may depend on output of earlier generators.
	rm -f internal/provider/service_packages_gen.go
	$(GO_VER) generate ./internal/provider
	rm -f internal/sweep/sweep_test.go internal/sweep/service_packages_gen_test.go
	$(GO_VER) generate ./internal/sweep

gencheck:
	@echo "==> Checking generated source code..."
	@$(MAKE) gen
	@git diff --compact-summary --exit-code || \
		(echo; echo "Unexpected difference in directories after code generation. Run 'make gen' command and commit."; exit 1)

generate-changelog:
	@echo "==> Generating changelog..."
	@sh -c "'$(CURDIR)/.ci/scripts/generate-changelog.sh'"

gh-workflows-lint:
	@echo "==> Checking github workflows with actionlint..."
	@actionlint

golangci-lint:
	@echo "==> Checking source code with golangci-lint..."
	@golangci-lint run \
		--config .ci/.golangci.yml \
		--config .ci/.golangci2.yml \
		./$(PKG_NAME)/...

importlint:
	@echo "==> Checking source code with importlint..."
	@impi --local . --scheme stdThirdPartyLocal ./internal/...

lint: golangci-lint providerlint importlint

lint-fix: testacc-lint-fix website-lint-fix docs-lint-fix

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
		./internal/service/... ./internal/provider/...

sane:
	@echo "==> Sane Check (48 tests of Top 30 resources)"
	@echo "==> Like 'sanity' except full output and stops soon after 1st error"
	@echo "==> NOTE: NOT an exhaustive set of tests! Finds big problems only."
	@TF_ACC=1 $(GO_VER) test \
		./internal/service/iam/... \
		-v -count $(TEST_COUNT) -parallel $(ACCTEST_PARALLELISM) -run='TestAccIAMRole_basic|TestAccIAMRole_namePrefix|TestAccIAMRole_disappears|TestAccIAMRole_InlinePolicy_basic|TestAccIAMPolicyDocumentDataSource_basic|TestAccIAMPolicyDocumentDataSource_sourceConflicting|TestAccIAMPolicyDocumentDataSource_sourceJSONValidJSON|TestAccIAMRolePolicyAttachment_basic|TestAccIAMRolePolicyAttachment_disappears|TestAccIAMRolePolicyAttachment_Disappears_role|TestAccIAMPolicy_basic|TestAccIAMPolicy_policy|TestAccIAMPolicy_tags|TestAccIAMRolePolicy_basic|TestAccIAMRolePolicy_unknownsInPolicy|TestAccIAMInstanceProfile_basic|TestAccIAMInstanceProfile_tags' -timeout $(ACCTEST_TIMEOUT)
	@TF_ACC=1 $(GO_VER) test \
		./internal/service/logs/... \
		./internal/service/ec2/... \
		./internal/service/ecs/... \
		./internal/service/elbv2/... \
		./internal/service/kms/... \
		-v -count $(TEST_COUNT) -parallel $(ACCTEST_PARALLELISM) -run='TestAccVPCSecurityGroup_basic|TestAccVPCSecurityGroup_ipRangesWithSameRules|TestAccVPCSecurityGroup_vpcAllEgress|TestAccVPCSecurityGroupRule_race|TestAccVPCSecurityGroupRule_protocolChange|TestAccVPCDataSource_basic|TestAccVPCSubnet_basic|TestAccVPC_tenancy|TestAccVPCRouteTableAssociation_Subnet_basic|TestAccVPCRouteTable_basic|TestAccLogsGroup_basic|TestAccLogsGroup_multiple|TestAccKMSKey_basic|TestAccELBV2TargetGroup_basic|TestAccECSTaskDefinition_basic|TestAccECSService_basic' -timeout $(ACCTEST_TIMEOUT)
	@TF_ACC=1 $(GO_VER) test \
		./internal/service/lambda/... \
		./internal/service/meta/... \
		./internal/service/route53/... \
		./internal/service/s3/... \
		./internal/service/secretsmanager/... \
		./internal/service/sts/... \
		-v -count $(TEST_COUNT) -parallel $(ACCTEST_PARALLELISM) -run='TestAccSTSCallerIdentityDataSource_basic|TestAccMetaRegionDataSource_basic|TestAccMetaRegionDataSource_endpoint|TestAccMetaPartitionDataSource_basic|TestAccS3Bucket_Basic_basic|TestAccS3Bucket_Security_corsUpdate|TestAccS3BucketPublicAccessBlock_basic|TestAccS3BucketPolicy_basic|TestAccS3BucketACL_updateACL|TestAccRoute53Record_basic|TestAccRoute53Record_Latency_basic|TestAccRoute53ZoneDataSource_name|TestAccLambdaFunction_basic|TestAccLambdaPermission_basic|TestAccSecretsManagerSecret_basic' -timeout $(ACCTEST_TIMEOUT)

sanity:
	@echo "==> Sanity Check (48 tests of Top 30 resources)"
	@echo "==> Like 'sane' but less output and runs all tests despite most errors"
	@echo "==> NOTE: NOT an exhaustive set of tests! Finds big problems only."
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
		-v -count $(TEST_COUNT) -parallel $(ACCTEST_PARALLELISM) -run='TestAccVPCSecurityGroup_basic|TestAccVPCSecurityGroup_ipRangesWithSameRules|TestAccVPCSecurityGroup_vpcAllEgress|TestAccVPCSecurityGroupRule_race|TestAccVPCSecurityGroupRule_protocolChange|TestAccVPCDataSource_basic|TestAccVPCSubnet_basic|TestAccVPC_tenancy|TestAccVPCRouteTableAssociation_Subnet_basic|TestAccVPCRouteTable_basic|TestAccLogsGroup_basic|TestAccLogsGroup_multiple|TestAccKMSKey_basic|TestAccELBV2TargetGroup_basic|TestAccECSTaskDefinition_basic|TestAccECSService_basic' -timeout $(ACCTEST_TIMEOUT) || true` ; \
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

semall: semgrep-validate
	@echo "==> Running Semgrep checks locally (must have semgrep installed)..."
	@semgrep --error --metrics=off \
		$(if $(filter-out $(origin PKG), undefined),--include $(PKG_NAME),) \
		--config .ci/.semgrep.yml \
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

semgrep-validate:
	@semgrep --error --validate \
		--config .ci/.semgrep.yml \
		--config .ci/.semgrep-caps-aws-ec2.yml \
		--config .ci/.semgrep-configs.yml \
		--config .ci/.semgrep-service-name0.yml \
		--config .ci/.semgrep-service-name1.yml \
		--config .ci/.semgrep-service-name2.yml \
		--config .ci/.semgrep-service-name3.yml \
		--config .ci/semgrep/

semgrep: semgrep-validate
	@echo "==> Running Semgrep static analysis..."
	@docker run --rm --volume "${PWD}:/src" returntocorp/semgrep semgrep --config .ci/.semgrep.yml

skaff:
	cd skaff && $(GO_VER) install github.com/hashicorp/terraform-provider-aws/skaff

sweep:
	# make sweep SWEEPARGS=-sweep-run=aws_example_thing
	# set SWEEPARGS=-sweep-allow-failures to continue after first failure
	@echo "WARNING: This will destroy infrastructure. Use only in development accounts."
	$(GO_VER) test $(SWEEP_DIR) -v -tags=sweep -sweep=$(SWEEP) $(SWEEPARGS) -timeout $(SWEEP_TIMEOUT)

t: fmtcheck
	TF_ACC=1 $(GO_VER) test ./$(PKG_NAME)/... -v -count $(TEST_COUNT) -parallel $(ACCTEST_PARALLELISM) $(RUNARGS) $(TESTARGS) -timeout $(ACCTEST_TIMEOUT)

test: fmtcheck
	$(GO_VER) test $(TEST) $(TESTARGS) -timeout=5m

test-compile:
	@if [ "$(TEST)" = "./..." ]; then \
		echo "ERROR: Set TEST to a specific package. For example,"; \
		echo "  make test-compile TEST=./$(PKG_NAME)"; \
		exit 1; \
	fi
	$(GO_VER) test -c $(TEST) $(TESTARGS)

testacc: fmtcheck
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

testacc-lint:
	@echo "Checking acceptance tests with terrafmt"
	find $(SVC_DIR) -type f -name '*_test.go' \
    | sort -u \
    | xargs -I {} terrafmt diff --check --fmtcompat {}

testacc-lint-fix:
	@echo "Fixing acceptance tests with terrafmt"
	find $(SVC_DIR) -type f -name '*_test.go' \
	| sort -u \
	| xargs -I {} terrafmt fmt  --fmtcompat {}

testacc-short: fmtcheck
	@echo "Running acceptance tests with -short flag"
	TF_ACC=1 $(GO_VER) test ./$(PKG_NAME)/... -v -short -count $(TEST_COUNT) -parallel $(ACCTEST_PARALLELISM) $(RUNARGS) $(TESTARGS) -timeout $(ACCTEST_TIMEOUT)

tfsdk2fw:
	cd tools/tfsdk2fw && $(GO_VER) install github.com/hashicorp/terraform-provider-aws/tools/tfsdk2fw

tools:
	cd .ci/providerlint && $(GO_VER) install .
	cd .ci/tools && $(GO_VER) install github.com/YakDriver/tfproviderdocs
	cd .ci/tools && $(GO_VER) install github.com/client9/misspell/cmd/misspell
	cd .ci/tools && $(GO_VER) install github.com/golangci/golangci-lint/cmd/golangci-lint
	cd .ci/tools && $(GO_VER) install github.com/katbyte/terrafmt
	cd .ci/tools && $(GO_VER) install github.com/terraform-linters/tflint
	cd .ci/tools && $(GO_VER) install github.com/pavius/impi/cmd/impi
	cd .ci/tools && $(GO_VER) install github.com/hashicorp/go-changelog/cmd/changelog-build
	cd .ci/tools && $(GO_VER) install github.com/hashicorp/copywrite
	cd .ci/tools && $(GO_VER) install github.com/rhysd/actionlint/cmd/actionlint
	cd .ci/tools && $(GO_VER) install mvdan.cc/gofumpt

ts: testacc-short

website-link-check:
	@.ci/scripts/markdown-link-check.sh

website-link-check-ghrc:
	@LINK_CHECK_CONTAINER="ghcr.io/tcort/markdown-link-check:stable" .ci/scripts/markdown-link-check.sh

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

yamllint:
	@yamllint .

# Please keep targets in alphabetical order
.PHONY: \
	build \
	depscheck \
	docs-lint \
	docs-lint-fix \
	docscheck \
	fmt \
	fmtcheck \
	fumpt \
	gen \
	gencheck \
	generate-changelog \
	gh-workflows-lint \
	golangci-lint \
	importlint \
	lint \
	lint-fix \
	providerlint \
	sane \
	sanity \
	semall \
	semgrep \
	skaff \
	sweep \
	t \
	test \
	test-compile \
	testacc \
	testacc-lint \
	testacc-lint-fix \
	testacc-short \
	tfsdk2fw \
	tools \
	ts \
	website-link-check \
	website-link-check-ghrc \
	website-lint \
	website-lint-fix \
	yamllint
