// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package synthetics_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/synthetics/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfsynthetics "github.com/hashicorp/terraform-provider-aws/internal/service/synthetics"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSyntheticsCanary_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var conf1, conf2 awstypes.Canary
	rName := fmt.Sprintf("tf-acc-test-%s", sdkacctest.RandString(8))
	resourceName := "aws_synthetics_canary.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SyntheticsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCanaryDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCanaryConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCanaryExists(ctx, resourceName, &conf1),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "synthetics", regexache.MustCompile(`canary:.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "runtime_version", "syn-nodejs-puppeteer-6.1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "run_config.0.memory_in_mb", "1000"),
					resource.TestCheckResourceAttr(resourceName, "run_config.0.timeout_in_seconds", "840"),
					resource.TestCheckResourceAttr(resourceName, "failure_retention_period", "31"),
					resource.TestCheckResourceAttr(resourceName, "success_retention_period", "31"),
					resource.TestCheckResourceAttr(resourceName, "handler", "exports.handler"),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.duration_in_seconds", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.expression", "rate(0 hour)"),
					acctest.MatchResourceAttrRegionalARN(resourceName, "engine_arn", "lambda", regexache.MustCompile(fmt.Sprintf(`function:cwsyn-%s.+`, rName))),
					acctest.MatchResourceAttrRegionalARN(resourceName, "source_location_arn", "lambda", regexache.MustCompile(fmt.Sprintf(`layer:cwsyn-%s.+`, rName))),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrExecutionRoleARN, "aws_iam_role.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "artifact_s3_location", fmt.Sprintf("%s/", rName)),
					resource.TestCheckResourceAttr(resourceName, "timeline.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "timeline.0.created"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "READY"),
					resource.TestCheckResourceAttr(resourceName, "artifact_config.#", acctest.Ct0),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"zip_file", "start_canary", "delete_lambda"},
			},
			{
				Config: testAccCanaryConfig_zipUpdated(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCanaryExists(ctx, resourceName, &conf2),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "synthetics", regexache.MustCompile(`canary:.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "runtime_version", "syn-nodejs-puppeteer-6.1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "run_config.0.memory_in_mb", "1000"),
					resource.TestCheckResourceAttr(resourceName, "run_config.0.timeout_in_seconds", "840"),
					resource.TestCheckResourceAttr(resourceName, "failure_retention_period", "31"),
					resource.TestCheckResourceAttr(resourceName, "success_retention_period", "31"),
					resource.TestCheckResourceAttr(resourceName, "handler", "exports.handler"),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.duration_in_seconds", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.expression", "rate(0 hour)"),
					acctest.MatchResourceAttrRegionalARN(resourceName, "engine_arn", "lambda", regexache.MustCompile(fmt.Sprintf(`function:cwsyn-%s.+`, rName))),
					acctest.MatchResourceAttrRegionalARN(resourceName, "source_location_arn", "lambda", regexache.MustCompile(fmt.Sprintf(`layer:cwsyn-%s.+`, rName))),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrExecutionRoleARN, "aws_iam_role.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "artifact_s3_location", fmt.Sprintf("%s/test/", rName)),
					resource.TestCheckResourceAttr(resourceName, "timeline.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "timeline.0.created"),
					resource.TestCheckResourceAttrSet(resourceName, "timeline.0.last_modified"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "READY"),
					testAccCheckCanaryIsUpdated(&conf1, &conf2),
				),
			},
		},
	})
}

func TestAccSyntheticsCanary_artifactEncryption(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.Canary
	rName := fmt.Sprintf("tf-acc-test-%s", sdkacctest.RandString(8))
	resourceName := "aws_synthetics_canary.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SyntheticsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCanaryDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCanaryConfig_artifactEncryption(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCanaryExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "artifact_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "artifact_config.0.s3_encryption.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "artifact_config.0.s3_encryption.0.encryption_mode", "SSE_S3"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"zip_file", "start_canary", "delete_lambda"},
			},
			{
				Config: testAccCanaryConfig_artifactEncryptionKMS(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCanaryExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "artifact_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "artifact_config.0.s3_encryption.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "artifact_config.0.s3_encryption.0.encryption_mode", "SSE_KMS"),
					resource.TestCheckResourceAttrPair(resourceName, "artifact_config.0.s3_encryption.0.kms_key_arn", "aws_kms_key.test", names.AttrARN),
				),
			},
		},
	})
}

func TestAccSyntheticsCanary_runtimeVersion(t *testing.T) {
	ctx := acctest.Context(t)
	var conf1 awstypes.Canary
	rName := fmt.Sprintf("tf-acc-test-%s", sdkacctest.RandString(8))
	resourceName := "aws_synthetics_canary.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SyntheticsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCanaryDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCanaryConfig_runtimeVersion(rName, "syn-nodejs-puppeteer-6.0"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCanaryExists(ctx, resourceName, &conf1),
					resource.TestCheckResourceAttr(resourceName, "runtime_version", "syn-nodejs-puppeteer-6.0"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"zip_file", "start_canary", "delete_lambda"},
			},
			{
				Config: testAccCanaryConfig_runtimeVersion(rName, "syn-nodejs-puppeteer-6.1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCanaryExists(ctx, resourceName, &conf1),
					resource.TestCheckResourceAttr(resourceName, "runtime_version", "syn-nodejs-puppeteer-6.1"),
				),
			},
		},
	})
}

func TestAccSyntheticsCanary_startCanary(t *testing.T) {
	ctx := acctest.Context(t)
	var conf1, conf2, conf3 awstypes.Canary
	rName := fmt.Sprintf("tf-acc-test-%s", sdkacctest.RandString(8))
	resourceName := "aws_synthetics_canary.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SyntheticsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCanaryDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCanaryConfig_start(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCanaryExists(ctx, resourceName, &conf1),
					resource.TestCheckResourceAttr(resourceName, "timeline.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "timeline.0.last_started"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"zip_file", "start_canary", "delete_lambda"},
			},
			{
				Config: testAccCanaryConfig_start(rName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCanaryExists(ctx, resourceName, &conf2),
					resource.TestCheckResourceAttr(resourceName, "timeline.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "timeline.0.last_started"),
					resource.TestCheckResourceAttrSet(resourceName, "timeline.0.last_stopped"),
				),
			},
			{
				Config: testAccCanaryConfig_start(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCanaryExists(ctx, resourceName, &conf3),
					resource.TestCheckResourceAttr(resourceName, "timeline.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "timeline.0.last_started"),
					resource.TestCheckResourceAttrSet(resourceName, "timeline.0.last_stopped"),
					testAccCheckCanaryIsStartedAfter(&conf2, &conf3),
				),
			},
		},
	})
}

func TestAccSyntheticsCanary_StartCanary_codeChanges(t *testing.T) {
	ctx := acctest.Context(t)
	var conf1, conf2 awstypes.Canary
	rName := fmt.Sprintf("tf-acc-test-%s", sdkacctest.RandString(8))
	resourceName := "aws_synthetics_canary.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SyntheticsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCanaryDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCanaryConfig_start(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCanaryExists(ctx, resourceName, &conf1),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "RUNNING"),
					resource.TestCheckResourceAttr(resourceName, "timeline.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "timeline.0.last_started"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"zip_file", "start_canary", "delete_lambda"},
			},
			{
				Config: testAccCanaryConfig_startZipUpdated(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCanaryExists(ctx, resourceName, &conf2),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "RUNNING"),
					resource.TestCheckResourceAttr(resourceName, "timeline.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "timeline.0.last_started"),
					resource.TestCheckResourceAttrSet(resourceName, "timeline.0.last_stopped"),
					testAccCheckCanaryIsStartedAfter(&conf1, &conf2),
				),
			},
		},
	})
}

func TestAccSyntheticsCanary_s3(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.Canary
	rName := fmt.Sprintf("tf-acc-test-%s", sdkacctest.RandString(8))
	resourceName := "aws_synthetics_canary.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SyntheticsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCanaryDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCanaryConfig_basicS3Code(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCanaryExists(ctx, resourceName, &conf),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "synthetics", regexache.MustCompile(`canary:.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "runtime_version", "syn-nodejs-puppeteer-6.1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "run_config.0.memory_in_mb", "1000"),
					resource.TestCheckResourceAttr(resourceName, "run_config.0.timeout_in_seconds", "840"),
					resource.TestCheckResourceAttr(resourceName, "run_config.0.active_tracing", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "failure_retention_period", "31"),
					resource.TestCheckResourceAttr(resourceName, "success_retention_period", "31"),
					resource.TestCheckResourceAttr(resourceName, "handler", "exports.handler"),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.duration_in_seconds", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.expression", "rate(0 hour)"),
					acctest.MatchResourceAttrRegionalARN(resourceName, "engine_arn", "lambda", regexache.MustCompile(fmt.Sprintf(`function:cwsyn-%s.+`, rName))),
					acctest.MatchResourceAttrRegionalARN(resourceName, "source_location_arn", "lambda", regexache.MustCompile(fmt.Sprintf(`layer:cwsyn-%s.+`, rName))),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrExecutionRoleARN, "aws_iam_role.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "artifact_s3_location", fmt.Sprintf("%s/", rName)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrS3Bucket, "s3_key", "s3_version", "start_canary", "delete_lambda"},
			},
		},
	})
}

func TestAccSyntheticsCanary_run(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.Canary
	rName := fmt.Sprintf("tf-acc-test-%s", sdkacctest.RandString(8))
	resourceName := "aws_synthetics_canary.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SyntheticsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCanaryDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCanaryConfig_run1(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCanaryExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "run_config.0.memory_in_mb", "1000"),
					resource.TestCheckResourceAttr(resourceName, "run_config.0.timeout_in_seconds", "60"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"zip_file", "start_canary", "delete_lambda"},
			},
			{
				Config: testAccCanaryConfig_run2(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCanaryExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "run_config.0.memory_in_mb", "960"),
					resource.TestCheckResourceAttr(resourceName, "run_config.0.timeout_in_seconds", "120"),
				),
			},
			{
				Config: testAccCanaryConfig_run1(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCanaryExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "run_config.0.memory_in_mb", "960"),
					resource.TestCheckResourceAttr(resourceName, "run_config.0.timeout_in_seconds", "60"),
				),
			},
		},
	})
}

func TestAccSyntheticsCanary_runTracing(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.Canary
	rName := fmt.Sprintf("tf-acc-test-%s", sdkacctest.RandString(8))
	resourceName := "aws_synthetics_canary.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SyntheticsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCanaryDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCanaryConfig_runTracing(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCanaryExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "run_config.0.active_tracing", acctest.CtTrue),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"zip_file", "start_canary", "delete_lambda"},
			},
			{
				Config: testAccCanaryConfig_runTracing(rName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCanaryExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "run_config.0.active_tracing", acctest.CtFalse),
				),
			},
			{
				Config: testAccCanaryConfig_runTracing(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCanaryExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "run_config.0.active_tracing", acctest.CtTrue),
				),
			},
		},
	})
}

func TestAccSyntheticsCanary_runEnvironmentVariables(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.Canary
	rName := fmt.Sprintf("tf-acc-test-%s", sdkacctest.RandString(8))
	resourceName := "aws_synthetics_canary.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SyntheticsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCanaryDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCanaryConfig_runEnvVariables1(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCanaryExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "run_config.0.environment_variables.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "run_config.0.environment_variables.test1", "result1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"zip_file", "start_canary", "delete_lambda", "run_config.0.environment_variables"},
			},
			{
				Config: testAccCanaryConfig_runEnvVariables2(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCanaryExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "run_config.0.environment_variables.%", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "run_config.0.environment_variables.test1", "result1"),
					resource.TestCheckResourceAttr(resourceName, "run_config.0.environment_variables.test2", "result2"),
				),
			},
		},
	})
}

func TestAccSyntheticsCanary_vpc(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var conf awstypes.Canary
	rName := fmt.Sprintf("tf-acc-test-%s", sdkacctest.RandString(8))
	resourceName := "aws_synthetics_canary.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SyntheticsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCanaryDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCanaryConfig_vpc1(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCanaryExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.subnet_ids.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.security_group_ids.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_config.0.vpc_id", "aws_vpc.test", names.AttrID),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"zip_file", "start_canary", "delete_lambda"},
			},
			{
				Config: testAccCanaryConfig_vpc2(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCanaryExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.subnet_ids.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.security_group_ids.#", acctest.Ct2),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_config.0.vpc_id", "aws_vpc.test", names.AttrID),
				),
			},
			{
				Config: testAccCanaryConfig_vpc3(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCanaryExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.subnet_ids.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.security_group_ids.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_config.0.vpc_id", "aws_vpc.test", names.AttrID),
				),
			},
		},
	})
}

func TestAccSyntheticsCanary_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.Canary
	rName := fmt.Sprintf("tf-acc-test-%s", sdkacctest.RandString(8))
	resourceName := "aws_synthetics_canary.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SyntheticsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCanaryDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCanaryConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCanaryExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"zip_file", "start_canary", "delete_lambda"},
			},
			{
				Config: testAccCanaryConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCanaryExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccCanaryConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCanaryExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccSyntheticsCanary_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.Canary
	rName := fmt.Sprintf("tf-acc-test-%s", sdkacctest.RandString(8))
	resourceName := "aws_synthetics_canary.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SyntheticsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCanaryDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCanaryConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCanaryExists(ctx, resourceName, &conf),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfsynthetics.ResourceCanary(), resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfsynthetics.ResourceCanary(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckCanaryDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SyntheticsClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_synthetics_canary" {
				continue
			}

			_, err := tfsynthetics.FindCanaryByName(ctx, conn, rs.Primary.ID)

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
}

func testAccCheckCanaryExists(ctx context.Context, n string, canary *awstypes.Canary) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Synthetics Canary ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SyntheticsClient(ctx)

		output, err := tfsynthetics.FindCanaryByName(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*canary = *output

		return nil
	}
}

func testAccCheckCanaryIsUpdated(first, second *awstypes.Canary) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if !second.Timeline.LastModified.After(*first.Timeline.LastModified) {
			return fmt.Errorf("synthetics Canary not updated")
		}

		return nil
	}
}

func testAccCheckCanaryIsStartedAfter(first, second *awstypes.Canary) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if !second.Timeline.LastStarted.After(*first.Timeline.LastStarted) {
			return fmt.Errorf("synthetics Canary not updated")
		}

		return nil
	}
}

func testAccCanaryConfig_base(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true

  tags = {
    Name = %[1]q
  }
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

func testAccCanaryConfig_run1(rName string) string {
	return acctest.ConfigCompose(testAccCanaryConfig_base(rName), fmt.Sprintf(`
resource "aws_synthetics_canary" "test" {
  name                 = %[1]q
  artifact_s3_location = "s3://${aws_s3_bucket.test.bucket}/"
  execution_role_arn   = aws_iam_role.test.arn
  handler              = "exports.handler"
  zip_file             = "test-fixtures/lambdatest.zip"
  runtime_version      = "syn-nodejs-puppeteer-6.1"
  delete_lambda        = true

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

func testAccCanaryConfig_run2(rName string) string {
	return acctest.ConfigCompose(testAccCanaryConfig_base(rName), fmt.Sprintf(`
resource "aws_synthetics_canary" "test" {
  name                 = %[1]q
  artifact_s3_location = "s3://${aws_s3_bucket.test.bucket}/"
  execution_role_arn   = aws_iam_role.test.arn
  handler              = "exports.handler"
  zip_file             = "test-fixtures/lambdatest.zip"
  runtime_version      = "syn-nodejs-puppeteer-6.1"
  delete_lambda        = true

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

func testAccCanaryConfig_runTracing(rName string, tracing bool) string {
	return acctest.ConfigCompose(testAccCanaryConfig_base(rName), fmt.Sprintf(`
resource "aws_synthetics_canary" "test" {
  name                 = %[1]q
  artifact_s3_location = "s3://${aws_s3_bucket.test.bucket}/"
  execution_role_arn   = aws_iam_role.test.arn
  handler              = "exports.handler"
  zip_file             = "test-fixtures/lambdatest.zip"
  runtime_version      = "syn-nodejs-puppeteer-6.1"
  delete_lambda        = true

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

func testAccCanaryConfig_runEnvVariables1(rName string) string {
	return acctest.ConfigCompose(testAccCanaryConfig_base(rName), fmt.Sprintf(`
resource "aws_synthetics_canary" "test" {
  name                 = %[1]q
  artifact_s3_location = "s3://${aws_s3_bucket.test.bucket}/"
  execution_role_arn   = aws_iam_role.test.arn
  handler              = "exports.handler"
  zip_file             = "test-fixtures/lambdatest.zip"
  runtime_version      = "syn-nodejs-puppeteer-6.1"
  delete_lambda        = true

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

func testAccCanaryConfig_runEnvVariables2(rName string) string {
	return acctest.ConfigCompose(testAccCanaryConfig_base(rName), fmt.Sprintf(`
resource "aws_synthetics_canary" "test" {
  name                 = %[1]q
  artifact_s3_location = "s3://${aws_s3_bucket.test.bucket}/"
  execution_role_arn   = aws_iam_role.test.arn
  handler              = "exports.handler"
  zip_file             = "test-fixtures/lambdatest.zip"
  runtime_version      = "syn-nodejs-puppeteer-6.1"
  delete_lambda        = true

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

func testAccCanaryConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccCanaryConfig_base(rName), fmt.Sprintf(`
resource "aws_synthetics_canary" "test" {
  # Must have bucket versioning enabled first
  depends_on = [aws_s3_bucket_versioning.test, aws_iam_role.test, aws_iam_role_policy.test]

  name                 = %[1]q
  artifact_s3_location = "s3://${aws_s3_bucket.test.bucket}/"
  execution_role_arn   = aws_iam_role.test.arn
  handler              = "exports.handler"
  zip_file             = "test-fixtures/lambdatest.zip"
  runtime_version      = "syn-nodejs-puppeteer-6.1"
  delete_lambda        = true

  schedule {
    expression = "rate(0 minute)"
  }
}
`, rName))
}

func testAccCanaryConfig_artifactEncryption(rName string) string {
	return acctest.ConfigCompose(testAccCanaryConfig_base(rName), fmt.Sprintf(`
resource "aws_synthetics_canary" "test" {
  name                 = %[1]q
  artifact_s3_location = "s3://${aws_s3_bucket.test.bucket}/"
  execution_role_arn   = aws_iam_role.test.arn
  handler              = "exports.handler"
  zip_file             = "test-fixtures/lambdatest.zip"
  runtime_version      = "syn-nodejs-puppeteer-6.1"
  delete_lambda        = true

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

func testAccCanaryConfig_artifactEncryptionKMS(rName string) string {
	return acctest.ConfigCompose(testAccCanaryConfig_base(rName), fmt.Sprintf(`
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
  runtime_version      = "syn-nodejs-puppeteer-6.1"
  delete_lambda        = true

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

func testAccCanaryConfig_runtimeVersion(rName, version string) string {
	return acctest.ConfigCompose(testAccCanaryConfig_base(rName), fmt.Sprintf(`
resource "aws_synthetics_canary" "test" {
  name                 = %[1]q
  artifact_s3_location = "s3://${aws_s3_bucket.test.bucket}/"
  execution_role_arn   = aws_iam_role.test.arn
  handler              = "exports.handler"
  zip_file             = "test-fixtures/lambdatest.zip"
  runtime_version      = %[2]q
  delete_lambda        = true

  schedule {
    expression = "rate(0 minute)"
  }

  depends_on = [aws_iam_role.test, aws_iam_role_policy.test]
}
`, rName, version))
}

func testAccCanaryConfig_zipUpdated(rName string) string {
	return acctest.ConfigCompose(testAccCanaryConfig_base(rName), fmt.Sprintf(`
resource "aws_synthetics_canary" "test" {
  name                 = %[1]q
  artifact_s3_location = "s3://${aws_s3_bucket.test.bucket}/test/"
  execution_role_arn   = aws_iam_role.test.arn
  handler              = "exports.handler"
  zip_file             = "test-fixtures/lambdatest_modified.zip"
  runtime_version      = "syn-nodejs-puppeteer-6.1"
  delete_lambda        = true

  schedule {
    expression = "rate(0 minute)"
  }

  depends_on = [aws_iam_role.test, aws_iam_role_policy.test]
}
`, rName))
}

func testAccCanaryConfig_start(rName string, state bool) string {
	return acctest.ConfigCompose(testAccCanaryConfig_base(rName), fmt.Sprintf(`
resource "aws_synthetics_canary" "test" {
  name                 = %[1]q
  artifact_s3_location = "s3://${aws_s3_bucket.test.bucket}/"
  execution_role_arn   = aws_iam_role.test.arn
  handler              = "exports.handler"
  zip_file             = "test-fixtures/lambdatest.zip"
  start_canary         = %[2]t
  runtime_version      = "syn-nodejs-puppeteer-6.1"
  delete_lambda        = true

  schedule {
    expression = "rate(0 minute)"
  }

  depends_on = [aws_iam_role.test, aws_iam_role_policy.test]
}
`, rName, state))
}

func testAccCanaryConfig_startZipUpdated(rName string, state bool) string {
	return acctest.ConfigCompose(testAccCanaryConfig_base(rName), fmt.Sprintf(`
resource "aws_synthetics_canary" "test" {
  name                 = %[1]q
  artifact_s3_location = "s3://${aws_s3_bucket.test.bucket}/"
  execution_role_arn   = aws_iam_role.test.arn
  handler              = "exports.handler"
  zip_file             = "test-fixtures/lambdatest_modified.zip"
  start_canary         = %[2]t
  runtime_version      = "syn-nodejs-puppeteer-6.1"
  delete_lambda        = true

  schedule {
    expression = "rate(0 minute)"
  }

  depends_on = [aws_iam_role.test, aws_iam_role_policy.test]
}
`, rName, state))
}

func testAccCanaryConfig_basicS3Code(rName string) string {
	return acctest.ConfigCompose(testAccCanaryConfig_base(rName), fmt.Sprintf(`
resource "aws_synthetics_canary" "test" {
  name                 = %[1]q
  artifact_s3_location = "s3://${aws_s3_bucket.test.bucket}/"
  execution_role_arn   = aws_iam_role.test.arn
  handler              = "exports.handler"
  s3_bucket            = aws_s3_object.test.bucket
  s3_key               = aws_s3_object.test.key
  s3_version           = aws_s3_object.test.version_id
  runtime_version      = "syn-nodejs-puppeteer-6.1"
  delete_lambda        = true

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

func testAccCanarySecurityGroupBaseConfig(rName string, count int) string {
	return acctest.ConfigCompose(fmt.Sprintf(`
resource "aws_security_group" "test" {
  count  = %[2]d
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_iam_role_policy_attachment" "test" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AWSLambdaVPCAccessExecutionRole"
  role       = aws_iam_role.test.name
}
`, rName, count))
}

func testAccCanaryConfig_vpc1(rName string) string {
	return acctest.ConfigCompose(
		testAccCanaryConfig_base(rName),
		acctest.ConfigVPCWithSubnets(rName, 2),
		testAccCanarySecurityGroupBaseConfig(rName, 2),
		fmt.Sprintf(`
resource "aws_synthetics_canary" "test" {
  name                 = %[1]q
  artifact_s3_location = "s3://${aws_s3_bucket.test.bucket}/"
  execution_role_arn   = aws_iam_role.test.arn
  handler              = "exports.handler"
  zip_file             = "test-fixtures/lambdatest.zip"
  runtime_version      = "syn-nodejs-puppeteer-6.1"
  delete_lambda        = true

  schedule {
    expression = "rate(0 minute)"
  }

  vpc_config {
    subnet_ids         = [aws_subnet.test[0].id]
    security_group_ids = [aws_security_group.test[0].id]
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName))
}

func testAccCanaryConfig_vpc2(rName string) string {
	return acctest.ConfigCompose(
		testAccCanaryConfig_base(rName),
		acctest.ConfigVPCWithSubnets(rName, 2),
		testAccCanarySecurityGroupBaseConfig(rName, 2),
		fmt.Sprintf(`
resource "aws_synthetics_canary" "test" {
  name                 = %[1]q
  artifact_s3_location = "s3://${aws_s3_bucket.test.bucket}/"
  execution_role_arn   = aws_iam_role.test.arn
  handler              = "exports.handler"
  zip_file             = "test-fixtures/lambdatest.zip"
  runtime_version      = "syn-nodejs-puppeteer-6.1"
  delete_lambda        = true

  schedule {
    expression = "rate(0 minute)"
  }

  vpc_config {
    subnet_ids         = aws_subnet.test[*].id
    security_group_ids = aws_security_group.test[*].id
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName))
}

func testAccCanaryConfig_vpc3(rName string) string {
	return acctest.ConfigCompose(
		testAccCanaryConfig_base(rName),
		acctest.ConfigVPCWithSubnets(rName, 2),
		testAccCanarySecurityGroupBaseConfig(rName, 2),
		fmt.Sprintf(`
resource "aws_synthetics_canary" "test" {
  name                 = %[1]q
  artifact_s3_location = "s3://${aws_s3_bucket.test.bucket}/"
  execution_role_arn   = aws_iam_role.test.arn
  handler              = "exports.handler"
  zip_file             = "test-fixtures/lambdatest.zip"
  runtime_version      = "syn-nodejs-puppeteer-6.1"
  delete_lambda        = true

  schedule {
    expression = "rate(0 minute)"
  }

  vpc_config {
    subnet_ids         = [aws_subnet.test[1].id]
    security_group_ids = [aws_security_group.test[1].id]
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName))
}

func testAccCanaryConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccCanaryConfig_base(rName), fmt.Sprintf(`
resource "aws_synthetics_canary" "test" {
  name                 = %[1]q
  artifact_s3_location = "s3://${aws_s3_bucket.test.bucket}/"
  execution_role_arn   = aws_iam_role.test.arn
  handler              = "exports.handler"
  zip_file             = "test-fixtures/lambdatest.zip"
  runtime_version      = "syn-nodejs-puppeteer-6.1"
  delete_lambda        = true

  schedule {
    expression = "rate(0 minute)"
  }

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccCanaryConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccCanaryConfig_base(rName), fmt.Sprintf(`
resource "aws_synthetics_canary" "test" {
  name                 = %[1]q
  artifact_s3_location = "s3://${aws_s3_bucket.test.bucket}/"
  execution_role_arn   = aws_iam_role.test.arn
  handler              = "exports.handler"
  zip_file             = "test-fixtures/lambdatest.zip"
  runtime_version      = "syn-nodejs-puppeteer-6.1"
  delete_lambda        = true

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
