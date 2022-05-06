package synthetics_test

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/aws/aws-sdk-go/service/synthetics"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfsynthetics "github.com/hashicorp/terraform-provider-aws/internal/service/synthetics"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccSyntheticsCanary_basic(t *testing.T) {
	var conf1, conf2 synthetics.Canary
	rName := fmt.Sprintf("tf-acc-test-%s", sdkacctest.RandString(8))
	resourceName := "aws_synthetics_canary.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, synthetics.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckCanaryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCanaryBasicConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCanaryExists(resourceName, &conf1),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", synthetics.ServiceName, regexp.MustCompile(`canary:.+`)),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "runtime_version", "syn-nodejs-puppeteer-3.2"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "run_config.0.memory_in_mb", "1000"),
					resource.TestCheckResourceAttr(resourceName, "run_config.0.timeout_in_seconds", "840"),
					resource.TestCheckResourceAttr(resourceName, "failure_retention_period", "31"),
					resource.TestCheckResourceAttr(resourceName, "success_retention_period", "31"),
					resource.TestCheckResourceAttr(resourceName, "handler", "exports.handler"),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.duration_in_seconds", "0"),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.expression", "rate(0 hour)"),
					acctest.MatchResourceAttrRegionalARN(resourceName, "engine_arn", "lambda", regexp.MustCompile(fmt.Sprintf(`function:cwsyn-%s.+`, rName))),
					acctest.MatchResourceAttrRegionalARN(resourceName, "source_location_arn", "lambda", regexp.MustCompile(fmt.Sprintf(`layer:cwsyn-%s.+`, rName))),
					resource.TestCheckResourceAttrPair(resourceName, "execution_role_arn", "aws_iam_role.test", "arn"),
					resource.TestCheckResourceAttr(resourceName, "artifact_s3_location", fmt.Sprintf("%s/", rName)),
					resource.TestCheckResourceAttr(resourceName, "timeline.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "timeline.0.created"),
					resource.TestCheckResourceAttr(resourceName, "status", "READY"),
					resource.TestCheckResourceAttr(resourceName, "artifact_config.#", "0"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"zip_file", "start_canary"},
			},
			{
				Config: testAccCanaryZipUpdatedConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCanaryExists(resourceName, &conf2),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", synthetics.ServiceName, regexp.MustCompile(`canary:.+`)),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "runtime_version", "syn-nodejs-puppeteer-3.2"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "run_config.0.memory_in_mb", "1000"),
					resource.TestCheckResourceAttr(resourceName, "run_config.0.timeout_in_seconds", "840"),
					resource.TestCheckResourceAttr(resourceName, "failure_retention_period", "31"),
					resource.TestCheckResourceAttr(resourceName, "success_retention_period", "31"),
					resource.TestCheckResourceAttr(resourceName, "handler", "exports.handler"),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.duration_in_seconds", "0"),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.expression", "rate(0 hour)"),
					acctest.MatchResourceAttrRegionalARN(resourceName, "engine_arn", "lambda", regexp.MustCompile(fmt.Sprintf(`function:cwsyn-%s.+`, rName))),
					acctest.MatchResourceAttrRegionalARN(resourceName, "source_location_arn", "lambda", regexp.MustCompile(fmt.Sprintf(`layer:cwsyn-%s.+`, rName))),
					resource.TestCheckResourceAttrPair(resourceName, "execution_role_arn", "aws_iam_role.test", "arn"),
					resource.TestCheckResourceAttr(resourceName, "artifact_s3_location", fmt.Sprintf("%s/test/", rName)),
					resource.TestCheckResourceAttr(resourceName, "timeline.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "timeline.0.created"),
					resource.TestCheckResourceAttrSet(resourceName, "timeline.0.last_modified"),
					resource.TestCheckResourceAttr(resourceName, "status", "READY"),
					testAccCheckCanaryIsUpdated(&conf1, &conf2),
				),
			},
		},
	})
}

func TestAccSyntheticsCanary_artifactEncryption(t *testing.T) {
	var conf synthetics.Canary
	rName := fmt.Sprintf("tf-acc-test-%s", sdkacctest.RandString(8))
	resourceName := "aws_synthetics_canary.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, synthetics.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckCanaryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCanaryArtifactEncryptionConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCanaryExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "artifact_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "artifact_config.0.s3_encryption.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "artifact_config.0.s3_encryption.0.encryption_mode", "SSE_S3"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"zip_file", "start_canary"},
			},
			{
				Config: testAccCanaryArtifactEncryptionKMSConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCanaryExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "artifact_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "artifact_config.0.s3_encryption.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "artifact_config.0.s3_encryption.0.encryption_mode", "SSE_KMS"),
					resource.TestCheckResourceAttrPair(resourceName, "artifact_config.0.s3_encryption.0.kms_key_arn", "aws_kms_key.test", "arn"),
				),
			},
		},
	})
}

func TestAccSyntheticsCanary_runtimeVersion(t *testing.T) {
	var conf1 synthetics.Canary
	rName := fmt.Sprintf("tf-acc-test-%s", sdkacctest.RandString(8))
	resourceName := "aws_synthetics_canary.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, synthetics.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckCanaryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCanaryRuntimeVersionConfig(rName, "syn-nodejs-puppeteer-3.1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCanaryExists(resourceName, &conf1),
					resource.TestCheckResourceAttr(resourceName, "runtime_version", "syn-nodejs-puppeteer-3.1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"zip_file", "start_canary"},
			},
			{
				Config: testAccCanaryRuntimeVersionConfig(rName, "syn-nodejs-puppeteer-3.2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCanaryExists(resourceName, &conf1),
					resource.TestCheckResourceAttr(resourceName, "runtime_version", "syn-nodejs-puppeteer-3.2"),
				),
			},
		},
	})
}

func TestAccSyntheticsCanary_startCanary(t *testing.T) {
	var conf1, conf2, conf3 synthetics.Canary
	rName := fmt.Sprintf("tf-acc-test-%s", sdkacctest.RandString(8))
	resourceName := "aws_synthetics_canary.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, synthetics.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckCanaryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCanaryStartCanaryConfig(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCanaryExists(resourceName, &conf1),
					resource.TestCheckResourceAttr(resourceName, "timeline.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "timeline.0.last_started"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"zip_file", "start_canary"},
			},
			{
				Config: testAccCanaryStartCanaryConfig(rName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCanaryExists(resourceName, &conf2),
					resource.TestCheckResourceAttr(resourceName, "timeline.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "timeline.0.last_started"),
					resource.TestCheckResourceAttrSet(resourceName, "timeline.0.last_stopped"),
				),
			},
			{
				Config: testAccCanaryStartCanaryConfig(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCanaryExists(resourceName, &conf3),
					resource.TestCheckResourceAttr(resourceName, "timeline.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "timeline.0.last_started"),
					resource.TestCheckResourceAttrSet(resourceName, "timeline.0.last_stopped"),
					testAccCheckCanaryIsStartedAfter(&conf2, &conf3),
				),
			},
		},
	})
}

func TestAccSyntheticsCanary_StartCanary_codeChanges(t *testing.T) {
	var conf1, conf2 synthetics.Canary
	rName := fmt.Sprintf("tf-acc-test-%s", sdkacctest.RandString(8))
	resourceName := "aws_synthetics_canary.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, synthetics.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckCanaryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCanaryStartCanaryConfig(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCanaryExists(resourceName, &conf1),
					resource.TestCheckResourceAttr(resourceName, "status", "RUNNING"),
					resource.TestCheckResourceAttr(resourceName, "timeline.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "timeline.0.last_started"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"zip_file", "start_canary"},
			},
			{
				Config: testAccCanaryStartCanaryZipUpdatedConfig(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCanaryExists(resourceName, &conf2),
					resource.TestCheckResourceAttr(resourceName, "status", "RUNNING"),
					resource.TestCheckResourceAttr(resourceName, "timeline.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "timeline.0.last_started"),
					resource.TestCheckResourceAttrSet(resourceName, "timeline.0.last_stopped"),
					testAccCheckCanaryIsStartedAfter(&conf1, &conf2),
				),
			},
		},
	})
}

func TestAccSyntheticsCanary_s3(t *testing.T) {
	var conf synthetics.Canary
	rName := fmt.Sprintf("tf-acc-test-%s", sdkacctest.RandString(8))
	resourceName := "aws_synthetics_canary.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, synthetics.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckCanaryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCanaryBasicS3CodeConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCanaryExists(resourceName, &conf),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", synthetics.ServiceName, regexp.MustCompile(`canary:.+`)),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "runtime_version", "syn-nodejs-puppeteer-3.2"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "run_config.0.memory_in_mb", "1000"),
					resource.TestCheckResourceAttr(resourceName, "run_config.0.timeout_in_seconds", "840"),
					resource.TestCheckResourceAttr(resourceName, "run_config.0.active_tracing", "false"),
					resource.TestCheckResourceAttr(resourceName, "failure_retention_period", "31"),
					resource.TestCheckResourceAttr(resourceName, "success_retention_period", "31"),
					resource.TestCheckResourceAttr(resourceName, "handler", "exports.handler"),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.duration_in_seconds", "0"),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.expression", "rate(0 hour)"),
					acctest.MatchResourceAttrRegionalARN(resourceName, "engine_arn", "lambda", regexp.MustCompile(fmt.Sprintf(`function:cwsyn-%s.+`, rName))),
					acctest.MatchResourceAttrRegionalARN(resourceName, "source_location_arn", "lambda", regexp.MustCompile(fmt.Sprintf(`layer:cwsyn-%s.+`, rName))),
					resource.TestCheckResourceAttrPair(resourceName, "execution_role_arn", "aws_iam_role.test", "arn"),
					resource.TestCheckResourceAttr(resourceName, "artifact_s3_location", fmt.Sprintf("%s/", rName)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"s3_bucket", "s3_key", "s3_version", "start_canary"},
			},
		},
	})
}

func TestAccSyntheticsCanary_run(t *testing.T) {
	var conf synthetics.Canary
	rName := fmt.Sprintf("tf-acc-test-%s", sdkacctest.RandString(8))
	resourceName := "aws_synthetics_canary.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, synthetics.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckCanaryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCanaryRun1Config(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCanaryExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "run_config.0.memory_in_mb", "1000"),
					resource.TestCheckResourceAttr(resourceName, "run_config.0.timeout_in_seconds", "60"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"zip_file", "start_canary"},
			},
			{
				Config: testAccCanaryRun2Config(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCanaryExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "run_config.0.memory_in_mb", "960"),
					resource.TestCheckResourceAttr(resourceName, "run_config.0.timeout_in_seconds", "120"),
				),
			},
			{
				Config: testAccCanaryRun1Config(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCanaryExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "run_config.0.memory_in_mb", "960"),
					resource.TestCheckResourceAttr(resourceName, "run_config.0.timeout_in_seconds", "60"),
				),
			},
		},
	})
}

func TestAccSyntheticsCanary_runTracing(t *testing.T) {
	var conf synthetics.Canary
	rName := fmt.Sprintf("tf-acc-test-%s", sdkacctest.RandString(8))
	resourceName := "aws_synthetics_canary.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, synthetics.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckCanaryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCanaryRunTracingConfig(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCanaryExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "run_config.0.active_tracing", "true"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"zip_file", "start_canary"},
			},
			{
				Config: testAccCanaryRunTracingConfig(rName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCanaryExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "run_config.0.active_tracing", "false"),
				),
			},
			{
				Config: testAccCanaryRunTracingConfig(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCanaryExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "run_config.0.active_tracing", "true"),
				),
			},
		},
	})
}

func TestAccSyntheticsCanary_runEnvironmentVariables(t *testing.T) {
	var conf synthetics.Canary
	rName := fmt.Sprintf("tf-acc-test-%s", sdkacctest.RandString(8))
	resourceName := "aws_synthetics_canary.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, synthetics.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckCanaryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCanaryRunEnvVariables1Config(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCanaryExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "run_config.0.environment_variables.test1", "result1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"zip_file", "start_canary", "run_config.0.environment_variables"},
			},
			{
				Config: testAccCanaryRunEnvVariables2Config(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCanaryExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "run_config.0.environment_variables.test1", "result1"),
					resource.TestCheckResourceAttr(resourceName, "run_config.0.environment_variables.test2", "result2"),
				),
			},
			{
				Config: testAccCanaryBasicConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCanaryExists(resourceName, &conf),
					resource.TestCheckNoResourceAttr(resourceName, "run_config.0.environment_variables"),
				),
			},
		},
	})
}

func TestAccSyntheticsCanary_vpc(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var conf synthetics.Canary
	rName := fmt.Sprintf("tf-acc-test-%s", sdkacctest.RandString(8))
	resourceName := "aws_synthetics_canary.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, synthetics.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckCanaryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCanaryVPC1Config(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCanaryExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.subnet_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.security_group_ids.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_config.0.vpc_id", "aws_vpc.test", "id"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"zip_file", "start_canary"},
			},
			{
				Config: testAccCanaryVPC2Config(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCanaryExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.subnet_ids.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.security_group_ids.#", "2"),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_config.0.vpc_id", "aws_vpc.test", "id"),
				),
			},
			{
				Config: testAccCanaryVPC3Config(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCanaryExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.subnet_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.security_group_ids.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_config.0.vpc_id", "aws_vpc.test", "id"),
					testAccCheckCanaryDeleteImplicitResources(resourceName),
				),
			},
		},
	})
}

func TestAccSyntheticsCanary_tags(t *testing.T) {
	var conf synthetics.Canary
	rName := fmt.Sprintf("tf-acc-test-%s", sdkacctest.RandString(8))
	resourceName := "aws_synthetics_canary.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, synthetics.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckCanaryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCanaryTags1Config(rName, "key1", "value1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCanaryExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"zip_file", "start_canary"},
			},
			{
				Config: testAccCanaryTags2Config(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCanaryExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccCanaryTags1Config(rName, "key2", "value2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCanaryExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccSyntheticsCanary_disappears(t *testing.T) {
	var conf synthetics.Canary
	rName := fmt.Sprintf("tf-acc-test-%s", sdkacctest.RandString(8))
	resourceName := "aws_synthetics_canary.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, synthetics.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckCanaryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCanaryBasicConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCanaryExists(resourceName, &conf),
					acctest.CheckResourceDisappears(acctest.Provider, tfsynthetics.ResourceCanary(), resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfsynthetics.ResourceCanary(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckCanaryDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).SyntheticsConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_synthetics_canary" {
			continue
		}

		_, err := tfsynthetics.FindCanaryByName(conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("Synthetics Canary %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckCanaryExists(n string, canary *synthetics.Canary) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Synthetics Canary ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SyntheticsConn

		output, err := tfsynthetics.FindCanaryByName(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*canary = *output

		return nil
	}
}

func testAccCheckCanaryDeleteImplicitResources(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("synthetics Canary not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("synthetics Canary name not set")
		}

		lambdaConn := acctest.Provider.Meta().(*conns.AWSClient).LambdaConn

		layerArn := rs.Primary.Attributes["source_location_arn"]
		layerArnObj, err := arn.Parse(layerArn)
		if err != nil {
			return fmt.Errorf("synthetics Canary Lambda Layer %s incorrect arn format: %w", layerArn, err)
		}

		layerName := strings.Split(layerArnObj.Resource, ":")

		deleteLayerVersionInput := &lambda.DeleteLayerVersionInput{
			LayerName:     aws.String(layerName[1]),
			VersionNumber: aws.Int64(1),
		}

		_, err = lambdaConn.DeleteLayerVersion(deleteLayerVersionInput)
		if err != nil {
			return fmt.Errorf("synthetics Canary Lambda Layer %s could not be deleted: %w", layerArn, err)
		}

		lambdaArn := rs.Primary.Attributes["engine_arn"]
		lambdaArnObj, err := arn.Parse(layerArn)
		if err != nil {
			return fmt.Errorf("synthetics Canary Lambda %s incorrect arn format: %w", lambdaArn, err)
		}
		lambdaArnParts := strings.Split(lambdaArnObj.Resource, ":")

		deleteLambdaInput := &lambda.DeleteFunctionInput{
			FunctionName: aws.String(lambdaArnParts[1]),
		}

		_, err = lambdaConn.DeleteFunction(deleteLambdaInput)
		if err != nil {
			return fmt.Errorf("synthetics Canary Lambda %s could not be deleted: %w", lambdaArn, err)
		}

		return nil
	}
}

func testAccCheckCanaryIsUpdated(first, second *synthetics.Canary) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if !second.Timeline.LastModified.After(*first.Timeline.LastModified) {
			return fmt.Errorf("synthetics Canary not updated")

		}

		return nil
	}
}

func testAccCheckCanaryIsStartedAfter(first, second *synthetics.Canary) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if !second.Timeline.LastStarted.After(*first.Timeline.LastStarted) {
			return fmt.Errorf("synthetics Canary not updated")

		}

		return nil
	}
}

func testAccCanaryBaseConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_s3_bucket_acl" "test" {
  bucket = aws_s3_bucket.test.id
  acl    = "private"
}

resource "aws_s3_bucket_versioning" "test" {
  bucket = aws_s3_bucket.test.id
  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "lambda.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF

  tags = {
    Name = %[1]q
  }
}

data "aws_partition" "current" {}

resource "aws_iam_role_policy" "test" {
  name = %[1]q
  role = aws_iam_role.test.id

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
        "Effect": "Allow",
        "Action": [
            "logs:CreateLogGroup",
            "logs:CreateLogStream",
            "logs:PutLogEvents"
        ],
        "Resource": "arn:${data.aws_partition.current.partition}:logs:*:*:*"
    },
    {
      "Effect": "Allow",
      "Action": [
        "s3:PutObject"
      ],
      "Resource": [
        "${aws_s3_bucket.test.arn}/*"
      ]
    },
    {
      "Effect": "Allow",
      "Action": [
        "s3:GetBucketLocation",
        "s3:ListAllMyBuckets"
      ],
      "Resource": [
        "*"
      ]
    },
    {
      "Effect": "Allow",
      "Action": [
        "cloudwatch:PutMetricData"
      ],
      "Resource": [
        "*"
      ]
    },
    {
      "Effect": "Allow",
      "Action": [
        "ec2:CreateNetworkInterface",
        "ec2:DescribeNetworkInterfaces",
        "ec2:DeleteNetworkInterface"
      ],
      "Resource": [
        "*"
      ]
    }
  ]
}
EOF
}
`, rName)
}

func testAccCanaryRun1Config(rName string) string {
	return acctest.ConfigCompose(testAccCanaryBaseConfig(rName), fmt.Sprintf(`
resource "aws_synthetics_canary" "test" {
  name                 = %[1]q
  artifact_s3_location = "s3://${aws_s3_bucket.test.bucket}/"
  execution_role_arn   = aws_iam_role.test.arn
  handler              = "exports.handler"
  zip_file             = "test-fixtures/lambdatest.zip"
  runtime_version      = "syn-nodejs-puppeteer-3.2"

  schedule {
    expression = "rate(0 minute)"
  }

  run_config {
    timeout_in_seconds = 60
  }

  depends_on = [aws_iam_role.test, aws_iam_role_policy.test]
}
`, rName))
}

func testAccCanaryRun2Config(rName string) string {
	return acctest.ConfigCompose(testAccCanaryBaseConfig(rName), fmt.Sprintf(`
resource "aws_synthetics_canary" "test" {
  name                 = %[1]q
  artifact_s3_location = "s3://${aws_s3_bucket.test.bucket}/"
  execution_role_arn   = aws_iam_role.test.arn
  handler              = "exports.handler"
  zip_file             = "test-fixtures/lambdatest.zip"
  runtime_version      = "syn-nodejs-puppeteer-3.2"

  schedule {
    expression = "rate(0 minute)"
  }

  run_config {
    timeout_in_seconds = 120
    memory_in_mb       = 960
  }

  depends_on = [aws_iam_role.test, aws_iam_role_policy.test]
}
`, rName))
}

func testAccCanaryRunTracingConfig(rName string, tracing bool) string {
	return acctest.ConfigCompose(testAccCanaryBaseConfig(rName), fmt.Sprintf(`
resource "aws_synthetics_canary" "test" {
  name                 = %[1]q
  artifact_s3_location = "s3://${aws_s3_bucket.test.bucket}/"
  execution_role_arn   = aws_iam_role.test.arn
  handler              = "exports.handler"
  zip_file             = "test-fixtures/lambdatest.zip"
  runtime_version      = "syn-nodejs-puppeteer-3.2"

  schedule {
    expression = "rate(0 minute)"
  }

  run_config {
    active_tracing     = %[2]t
    timeout_in_seconds = 60
  }

  depends_on = [aws_iam_role.test, aws_iam_role_policy.test]
}
`, rName, tracing))
}

func testAccCanaryRunEnvVariables1Config(rName string) string {
	return acctest.ConfigCompose(testAccCanaryBaseConfig(rName), fmt.Sprintf(`
resource "aws_synthetics_canary" "test" {
  name                 = %[1]q
  artifact_s3_location = "s3://${aws_s3_bucket.test.bucket}/"
  execution_role_arn   = aws_iam_role.test.arn
  handler              = "exports.handler"
  zip_file             = "test-fixtures/lambdatest.zip"
  runtime_version      = "syn-nodejs-puppeteer-3.2"

  schedule {
    expression = "rate(0 minute)"
  }

  run_config {
    environment_variables = {
      test1 = "result1"
    }
  }

  depends_on = [aws_iam_role.test, aws_iam_role_policy.test]
}
`, rName))
}

func testAccCanaryRunEnvVariables2Config(rName string) string {
	return acctest.ConfigCompose(testAccCanaryBaseConfig(rName), fmt.Sprintf(`
resource "aws_synthetics_canary" "test" {
  name                 = %[1]q
  artifact_s3_location = "s3://${aws_s3_bucket.test.bucket}/"
  execution_role_arn   = aws_iam_role.test.arn
  handler              = "exports.handler"
  zip_file             = "test-fixtures/lambdatest.zip"
  runtime_version      = "syn-nodejs-puppeteer-3.2"

  schedule {
    expression = "rate(0 minute)"
  }

  run_config {
    environment_variables = {
      test1 = "result1"
      test2 = "result2"
    }
  }

  depends_on = [aws_iam_role.test, aws_iam_role_policy.test]
}
`, rName))
}

func testAccCanaryBasicConfig(rName string) string {
	return acctest.ConfigCompose(testAccCanaryBaseConfig(rName), fmt.Sprintf(`
resource "aws_synthetics_canary" "test" {
  # Must have bucket versioning enabled first
  depends_on = [aws_s3_bucket_versioning.test, aws_iam_role.test, aws_iam_role_policy.test]

  name                 = %[1]q
  artifact_s3_location = "s3://${aws_s3_bucket.test.bucket}/"
  execution_role_arn   = aws_iam_role.test.arn
  handler              = "exports.handler"
  zip_file             = "test-fixtures/lambdatest.zip"
  runtime_version      = "syn-nodejs-puppeteer-3.2"

  schedule {
    expression = "rate(0 minute)"
  }
}
`, rName))
}

func testAccCanaryArtifactEncryptionConfig(rName string) string {
	return acctest.ConfigCompose(testAccCanaryBaseConfig(rName), fmt.Sprintf(`
resource "aws_synthetics_canary" "test" {
  name                 = %[1]q
  artifact_s3_location = "s3://${aws_s3_bucket.test.bucket}/"
  execution_role_arn   = aws_iam_role.test.arn
  handler              = "exports.handler"
  zip_file             = "test-fixtures/lambdatest.zip"
  runtime_version      = "syn-nodejs-puppeteer-3.3"

  artifact_config {
    s3_encryption {
      encryption_mode = "SSE_S3"
    }
  }

  schedule {
    expression = "rate(0 minute)"
  }

  depends_on = [aws_iam_role.test, aws_iam_role_policy.test]
}
`, rName))
}

func testAccCanaryArtifactEncryptionKMSConfig(rName string) string {
	return acctest.ConfigCompose(testAccCanaryBaseConfig(rName), fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
}

resource "aws_synthetics_canary" "test" {
  name                 = %[1]q
  artifact_s3_location = "s3://${aws_s3_bucket.test.bucket}/"
  execution_role_arn   = aws_iam_role.test.arn
  handler              = "exports.handler"
  zip_file             = "test-fixtures/lambdatest.zip"
  runtime_version      = "syn-nodejs-puppeteer-3.3"

  artifact_config {
    s3_encryption {
      encryption_mode = "SSE_KMS"
      kms_key_arn     = aws_kms_key.test.arn
    }
  }

  schedule {
    expression = "rate(0 minute)"
  }

  depends_on = [aws_iam_role.test, aws_iam_role_policy.test]
}
`, rName))
}

func testAccCanaryRuntimeVersionConfig(rName, version string) string {
	return acctest.ConfigCompose(testAccCanaryBaseConfig(rName), fmt.Sprintf(`
resource "aws_synthetics_canary" "test" {
  name                 = %[1]q
  artifact_s3_location = "s3://${aws_s3_bucket.test.bucket}/"
  execution_role_arn   = aws_iam_role.test.arn
  handler              = "exports.handler"
  zip_file             = "test-fixtures/lambdatest.zip"
  runtime_version      = %[2]q

  schedule {
    expression = "rate(0 minute)"
  }

  depends_on = [aws_iam_role.test, aws_iam_role_policy.test]
}
`, rName, version))
}

func testAccCanaryZipUpdatedConfig(rName string) string {
	return acctest.ConfigCompose(testAccCanaryBaseConfig(rName), fmt.Sprintf(`
resource "aws_synthetics_canary" "test" {
  name                 = %[1]q
  artifact_s3_location = "s3://${aws_s3_bucket.test.bucket}/test/"
  execution_role_arn   = aws_iam_role.test.arn
  handler              = "exports.handler"
  zip_file             = "test-fixtures/lambdatest_modified.zip"
  runtime_version      = "syn-nodejs-puppeteer-3.2"

  schedule {
    expression = "rate(0 minute)"
  }

  depends_on = [aws_iam_role.test, aws_iam_role_policy.test]
}
`, rName))
}

func testAccCanaryStartCanaryConfig(rName string, state bool) string {
	return acctest.ConfigCompose(testAccCanaryBaseConfig(rName), fmt.Sprintf(`
resource "aws_synthetics_canary" "test" {
  name                 = %[1]q
  artifact_s3_location = "s3://${aws_s3_bucket.test.bucket}/"
  execution_role_arn   = aws_iam_role.test.arn
  handler              = "exports.handler"
  zip_file             = "test-fixtures/lambdatest.zip"
  start_canary         = %[2]t
  runtime_version      = "syn-nodejs-puppeteer-3.2"

  schedule {
    expression = "rate(0 minute)"
  }

  depends_on = [aws_iam_role.test, aws_iam_role_policy.test]
}
`, rName, state))
}

func testAccCanaryStartCanaryZipUpdatedConfig(rName string, state bool) string {
	return acctest.ConfigCompose(testAccCanaryBaseConfig(rName), fmt.Sprintf(`
resource "aws_synthetics_canary" "test" {
  name                 = %[1]q
  artifact_s3_location = "s3://${aws_s3_bucket.test.bucket}/"
  execution_role_arn   = aws_iam_role.test.arn
  handler              = "exports.handler"
  zip_file             = "test-fixtures/lambdatest_modified.zip"
  start_canary         = %[2]t
  runtime_version      = "syn-nodejs-puppeteer-3.2"

  schedule {
    expression = "rate(0 minute)"
  }

  depends_on = [aws_iam_role.test, aws_iam_role_policy.test]
}
`, rName, state))
}

func testAccCanaryBasicS3CodeConfig(rName string) string {
	return acctest.ConfigCompose(testAccCanaryBaseConfig(rName), fmt.Sprintf(`
resource "aws_synthetics_canary" "test" {
  name                 = %[1]q
  artifact_s3_location = "s3://${aws_s3_bucket.test.bucket}/"
  execution_role_arn   = aws_iam_role.test.arn
  handler              = "exports.handler"
  s3_bucket            = aws_s3_object.test.bucket
  s3_key               = aws_s3_object.test.key
  s3_version           = aws_s3_object.test.version_id
  runtime_version      = "syn-nodejs-puppeteer-3.2"

  schedule {
    expression = "rate(0 minute)"
  }

  depends_on = [aws_iam_role.test, aws_iam_role_policy.test]
}

resource "aws_s3_object" "test" {
  # Must have bucket versioning enabled first
  depends_on = [aws_s3_bucket_versioning.test]

  bucket = aws_s3_bucket.test.bucket
  key    = %[1]q
  source = "test-fixtures/lambdatest.zip"
  etag   = filemd5("test-fixtures/lambdatest.zip")
}

`, rName))
}

func testAccCanaryVPCBaseConfig(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test1" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 2, 0)
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test2" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 2, 1)
  availability_zone = data.aws_availability_zones.available.names[1]

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test1" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test2" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_iam_role_policy_attachment" "test" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AWSLambdaVPCAccessExecutionRole"
  role       = aws_iam_role.test.name
}
`, rName))
}

func testAccCanaryVPC1Config(rName string) string {
	return acctest.ConfigCompose(
		testAccCanaryBaseConfig(rName),
		testAccCanaryVPCBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_synthetics_canary" "test" {
  name                 = %[1]q
  artifact_s3_location = "s3://${aws_s3_bucket.test.bucket}/"
  execution_role_arn   = aws_iam_role.test.arn
  handler              = "exports.handler"
  zip_file             = "test-fixtures/lambdatest.zip"
  runtime_version      = "syn-nodejs-puppeteer-3.2"

  schedule {
    expression = "rate(0 minute)"
  }

  vpc_config {
    subnet_ids         = [aws_subnet.test1.id]
    security_group_ids = [aws_security_group.test1.id]
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName))
}

func testAccCanaryVPC2Config(rName string) string {
	return acctest.ConfigCompose(
		testAccCanaryBaseConfig(rName),
		testAccCanaryVPCBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_synthetics_canary" "test" {
  name                 = %[1]q
  artifact_s3_location = "s3://${aws_s3_bucket.test.bucket}/"
  execution_role_arn   = aws_iam_role.test.arn
  handler              = "exports.handler"
  zip_file             = "test-fixtures/lambdatest.zip"
  runtime_version      = "syn-nodejs-puppeteer-3.2"

  schedule {
    expression = "rate(0 minute)"
  }

  vpc_config {
    subnet_ids         = [aws_subnet.test1.id, aws_subnet.test2.id]
    security_group_ids = [aws_security_group.test1.id, aws_security_group.test2.id]
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName))
}

func testAccCanaryVPC3Config(rName string) string {
	return acctest.ConfigCompose(
		testAccCanaryBaseConfig(rName),
		testAccCanaryVPCBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_synthetics_canary" "test" {
  name                 = %[1]q
  artifact_s3_location = "s3://${aws_s3_bucket.test.bucket}/"
  execution_role_arn   = aws_iam_role.test.arn
  handler              = "exports.handler"
  zip_file             = "test-fixtures/lambdatest.zip"
  runtime_version      = "syn-nodejs-puppeteer-3.2"

  schedule {
    expression = "rate(0 minute)"
  }

  vpc_config {
    subnet_ids         = [aws_subnet.test2.id]
    security_group_ids = [aws_security_group.test2.id]
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName))
}

func testAccCanaryTags1Config(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccCanaryBaseConfig(rName), fmt.Sprintf(`
resource "aws_synthetics_canary" "test" {
  name                 = %[1]q
  artifact_s3_location = "s3://${aws_s3_bucket.test.bucket}/"
  execution_role_arn   = aws_iam_role.test.arn
  handler              = "exports.handler"
  zip_file             = "test-fixtures/lambdatest.zip"
  runtime_version      = "syn-nodejs-puppeteer-3.2"

  schedule {
    expression = "rate(0 minute)"
  }

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccCanaryTags2Config(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccCanaryBaseConfig(rName), fmt.Sprintf(`
resource "aws_synthetics_canary" "test" {
  name                 = %[1]q
  artifact_s3_location = "s3://${aws_s3_bucket.test.bucket}/"
  execution_role_arn   = aws_iam_role.test.arn
  handler              = "exports.handler"
  zip_file             = "test-fixtures/lambdatest.zip"
  runtime_version      = "syn-nodejs-puppeteer-3.2"

  schedule {
    expression = "rate(0 minute)"
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }

  depends_on = [aws_iam_role.test, aws_iam_role_policy.test]
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}
