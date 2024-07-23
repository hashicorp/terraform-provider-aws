// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ivschat_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ivschat"
	"github.com/aws/aws-sdk-go-v2/service/ivschat/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfivschat "github.com/hashicorp/terraform-provider-aws/internal/service/ivschat"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccIVSChatLoggingConfiguration_basic_cloudwatch(t *testing.T) {
	ctx := acctest.Context(t)
	var loggingconfiguration ivschat.GetLoggingConfigurationOutput
	resourceName := "aws_ivschat_logging_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.IVSChatEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.IVSChatServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLoggingConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLoggingConfigurationConfig_basic_cloudwatch(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLoggingConfigurationExists(ctx, resourceName, &loggingconfiguration),
					resource.TestCheckResourceAttrPair(resourceName, "destination_configuration.0.cloudwatch_logs.0.log_group_name", "aws_cloudwatch_log_group.test", names.AttrName),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, "ACTIVE"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrID),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "ivschat", regexache.MustCompile(`logging-configuration/.+`)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccIVSChatLoggingConfiguration_basic_firehose(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var loggingconfiguration ivschat.GetLoggingConfigurationOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ivschat_logging_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.IVSChatEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.IVSChatServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLoggingConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLoggingConfigurationConfig_basic_firehose(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLoggingConfigurationExists(ctx, resourceName, &loggingconfiguration),
					resource.TestCheckResourceAttrPair(resourceName, "destination_configuration.0.firehose.0.delivery_stream_name", "aws_kinesis_firehose_delivery_stream.test", names.AttrName),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, "ACTIVE"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrID),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "ivschat", regexache.MustCompile(`logging-configuration/.+`)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccIVSChatLoggingConfiguration_basic_s3(t *testing.T) {
	ctx := acctest.Context(t)
	var loggingconfiguration ivschat.GetLoggingConfigurationOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ivschat_logging_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.IVSChatEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.IVSChatServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLoggingConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLoggingConfigurationConfig_basic_s3(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLoggingConfigurationExists(ctx, resourceName, &loggingconfiguration),
					resource.TestCheckResourceAttrPair(resourceName, "destination_configuration.0.s3.0.bucket_name", "aws_s3_bucket.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, "ACTIVE"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrID),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "ivschat", regexache.MustCompile(`logging-configuration/.+`)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccIVSChatLoggingConfiguration_update_s3(t *testing.T) {
	ctx := acctest.Context(t)
	var v1, v2 ivschat.GetLoggingConfigurationOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ivschat_logging_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.IVSChatEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.IVSChatServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLoggingConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLoggingConfigurationConfig_basic_s3(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLoggingConfigurationExists(ctx, resourceName, &v1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccLoggingConfigurationConfig_update_s3(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLoggingConfigurationExists(ctx, resourceName, &v2),
					testAccCheckLoggingConfigurationNotRecreated(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
				),
			},
		},
	})
}

func TestAccIVSChatLoggingConfiguration_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var v1, v2, v3 ivschat.GetLoggingConfigurationOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ivschat_logging_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.IVSChatEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.IVSChatServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLoggingConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLoggingConfigurationConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLoggingConfigurationExists(ctx, resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccLoggingConfigurationConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLoggingConfigurationExists(ctx, resourceName, &v2),
					testAccCheckLoggingConfigurationNotRecreated(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccLoggingConfigurationConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLoggingConfigurationExists(ctx, resourceName, &v3),
					testAccCheckLoggingConfigurationNotRecreated(&v2, &v3),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccIVSChatLoggingConfiguration_failure_invalidDestination(t *testing.T) {
	ctx := acctest.Context(t)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.IVSChatEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.IVSChatServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLoggingConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccLoggingConfigurationConfig_failure_cloudwatch_s3(),
				ExpectError: regexache.MustCompile(`Invalid combination of arguments`),
			},
			{
				Config:      testAccLoggingConfigurationConfig_failure_firehose_s3(),
				ExpectError: regexache.MustCompile(`Invalid combination of arguments`),
			},
			{
				Config:      testAccLoggingConfigurationConfig_failure_cloudwatch_firehose(),
				ExpectError: regexache.MustCompile(`Invalid combination of arguments`),
			},
			{
				Config:      testAccLoggingConfigurationConfig_failure_cloudwatch_firehose_s3(),
				ExpectError: regexache.MustCompile(`Invalid combination of arguments`),
			},
			{
				Config:      testAccLoggingConfigurationConfig_failure_noDestination(),
				ExpectError: regexache.MustCompile(`Invalid combination of arguments`),
			},
		},
	})
}

func TestAccIVSChatLoggingConfiguration_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var loggingconfiguration ivschat.GetLoggingConfigurationOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ivschat_logging_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.IVSChatEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.IVSChatServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLoggingConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLoggingConfigurationConfig_basic_s3(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLoggingConfigurationExists(ctx, resourceName, &loggingconfiguration),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfivschat.ResourceLoggingConfiguration(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckLoggingConfigurationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).IVSChatClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ivschat_logging_configuration" {
				continue
			}

			_, err := conn.GetLoggingConfiguration(ctx, &ivschat.GetLoggingConfigurationInput{
				Identifier: aws.String(rs.Primary.ID),
			})
			if err != nil {
				var nfe *types.ResourceNotFoundException
				if errors.As(err, &nfe) {
					return nil
				}
				return err
			}

			return create.Error(names.IVSChat, create.ErrActionCheckingDestroyed, tfivschat.ResNameLoggingConfiguration, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckLoggingConfigurationExists(ctx context.Context, name string, loggingconfiguration *ivschat.GetLoggingConfigurationOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.IVSChat, create.ErrActionCheckingExistence, tfivschat.ResNameLoggingConfiguration, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.IVSChat, create.ErrActionCheckingExistence, tfivschat.ResNameLoggingConfiguration, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).IVSChatClient(ctx)

		resp, err := conn.GetLoggingConfiguration(ctx, &ivschat.GetLoggingConfigurationInput{
			Identifier: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return create.Error(names.IVSChat, create.ErrActionCheckingExistence, tfivschat.ResNameLoggingConfiguration, rs.Primary.ID, err)
		}

		*loggingconfiguration = *resp

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).IVSChatClient(ctx)

	input := &ivschat.ListLoggingConfigurationsInput{}
	_, err := conn.ListLoggingConfigurations(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccCheckLoggingConfigurationNotRecreated(before, after *ivschat.GetLoggingConfigurationOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if before, after := aws.ToString(before.Arn), aws.ToString(after.Arn); before != after {
			return create.Error(names.IVSChat, create.ErrActionCheckingNotRecreated, tfivschat.ResNameLoggingConfiguration, before, errors.New("recreated"))
		}

		return nil
	}
}

func testAccLoggingConfigurationConfig_s3(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}
`, rName)
}

func testAccLoggingConfigurationConfig_cloudwatch() string {
	return `
resource "aws_cloudwatch_log_group" "test" {}
`
}

func testAccLoggingConfigurationConfig_firehose(rName string) string {
	return fmt.Sprintf(`
resource "aws_kinesis_firehose_delivery_stream" "test" {
  name        = %[1]q
  destination = "extended_s3"

  extended_s3_configuration {
    role_arn   = aws_iam_role.test.arn
    bucket_arn = aws_s3_bucket.test.arn
  }

  tags = {
    "LogDeliveryEnabled" = "true"
  }
}

resource "aws_s3_bucket" "test" {
  bucket_prefix = %[1]q
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
        "Service": "firehose.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}
`, rName)
}

func testAccLoggingConfigurationConfig_basic_s3(rName string) string {
	return acctest.ConfigCompose(
		testAccLoggingConfigurationConfig_s3(rName),
		`
resource "aws_ivschat_logging_configuration" "test" {
  destination_configuration {
    s3 {
      bucket_name = aws_s3_bucket.test.id
    }
  }
}
`)
}

func testAccLoggingConfigurationConfig_update_s3(rName string) string {
	return acctest.ConfigCompose(
		testAccLoggingConfigurationConfig_s3(rName),
		fmt.Sprintf(`
resource "aws_ivschat_logging_configuration" "test" {
  name = %[1]q
  destination_configuration {
    s3 {
      bucket_name = aws_s3_bucket.test.id
    }
  }
}
`, rName))
}

func testAccLoggingConfigurationConfig_basic_cloudwatch() string {
	return acctest.ConfigCompose(
		testAccLoggingConfigurationConfig_cloudwatch(),
		`
resource "aws_ivschat_logging_configuration" "test" {
  destination_configuration {
    cloudwatch_logs {
      log_group_name = aws_cloudwatch_log_group.test.name
    }
  }
}
`)
}

func testAccLoggingConfigurationConfig_basic_firehose(rName string) string {
	return acctest.ConfigCompose(
		testAccLoggingConfigurationConfig_firehose(rName),
		`
resource "aws_ivschat_logging_configuration" "test" {
  destination_configuration {
    firehose {
      delivery_stream_name = aws_kinesis_firehose_delivery_stream.test.name
    }
  }
}
`)
}

func testAccLoggingConfigurationConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(
		testAccLoggingConfigurationConfig_s3(rName),
		fmt.Sprintf(`
resource "aws_ivschat_logging_configuration" "test" {
  destination_configuration {
    s3 {
      bucket_name = aws_s3_bucket.test.id
    }
  }
  tags = {
    %[1]q = %[2]q
  }
}
`, tagKey1, tagValue1))
}

func testAccLoggingConfigurationConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(
		testAccLoggingConfigurationConfig_s3(rName),
		fmt.Sprintf(`
resource "aws_ivschat_logging_configuration" "test" {
  destination_configuration {
    s3 {
      bucket_name = aws_s3_bucket.test.id
    }
  }
  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccLoggingConfigurationConfig_failure_cloudwatch_s3() string {
	return `
resource "aws_ivschat_logging_configuration" "test" {
  destination_configuration {
    cloudwatch_logs {
      log_group_name = "log_group_name"
    }
    s3 {
      bucket_name = "bucket_name"
    }
  }
}
`
}

func testAccLoggingConfigurationConfig_failure_firehose_s3() string {
	return `
resource "aws_ivschat_logging_configuration" "test" {
  destination_configuration {
    firehose {
      delivery_stream_name = "delivery_stream_name"
    }
    s3 {
      bucket_name = "bucket_name"
    }
  }
}
`
}

func testAccLoggingConfigurationConfig_failure_cloudwatch_firehose() string {
	return `
resource "aws_ivschat_logging_configuration" "test" {
  destination_configuration {
    cloudwatch_logs {
      log_group_name = "log_group_name"
    }
    firehose {
      delivery_stream_name = "delivery_stream_name"
    }
  }
}
`
}

func testAccLoggingConfigurationConfig_failure_cloudwatch_firehose_s3() string {
	return `
resource "aws_ivschat_logging_configuration" "test" {
  destination_configuration {
    cloudwatch_logs {
      log_group_name = "log_group_name"
    }
    firehose {
      delivery_stream_name = "delivery_stream_name"
    }
    s3 {
      bucket_name = "bucket_name"
    }
  }
}
`
}

func testAccLoggingConfigurationConfig_failure_noDestination() string {
	return `
resource "aws_ivschat_logging_configuration" "test" {
  destination_configuration {}
}
`
}
