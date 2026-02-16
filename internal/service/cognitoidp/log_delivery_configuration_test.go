// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package cognitoidp_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tfcognitoidp "github.com/hashicorp/terraform-provider-aws/internal/service/cognitoidp"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccCognitoIDPLogDeliveryConfiguration_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var logDeliveryConfiguration awstypes.LogDeliveryConfigurationType
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cognito_log_delivery_configuration.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLogDeliveryConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccLogDeliveryConfigurationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLogDeliveryConfigurationExists(ctx, t, resourceName, &logDeliveryConfiguration),
					resource.TestCheckResourceAttrPair("aws_cognito_user_pool.test", names.AttrID, resourceName, names.AttrUserPoolID),
					resource.TestCheckResourceAttr(resourceName, "log_configurations.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "log_configurations.0.event_source", "userNotification"),
					resource.TestCheckResourceAttr(resourceName, "log_configurations.0.log_level", "ERROR"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccLogDeliveryConfigurationImportStateIdFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrUserPoolID,
			},
		},
	})
}

func TestAccCognitoIDPLogDeliveryConfiguration_update(t *testing.T) {
	ctx := acctest.Context(t)
	var logDeliveryConfiguration awstypes.LogDeliveryConfigurationType
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cognito_log_delivery_configuration.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLogDeliveryConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccLogDeliveryConfigurationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLogDeliveryConfigurationExists(ctx, t, resourceName, &logDeliveryConfiguration),
					resource.TestCheckResourceAttr(resourceName, "log_configurations.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "log_configurations.0.event_source", "userNotification"),
					resource.TestCheckResourceAttr(resourceName, "log_configurations.0.log_level", "ERROR"),
				),
			},
			{
				Config: testAccLogDeliveryConfigurationConfig_firehose(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLogDeliveryConfigurationExists(ctx, t, resourceName, &logDeliveryConfiguration),
					resource.TestCheckResourceAttr(resourceName, "log_configurations.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "log_configurations.0.event_source", "userNotification"),
					resource.TestCheckResourceAttr(resourceName, "log_configurations.0.log_level", "INFO"),
					resource.TestCheckResourceAttr(resourceName, "log_configurations.1.event_source", "userAuthEvents"),
					resource.TestCheckResourceAttr(resourceName, "log_configurations.1.log_level", "ERROR"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccLogDeliveryConfigurationImportStateIdFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrUserPoolID,
			},
		},
	})
}

func TestAccCognitoIDPLogDeliveryConfiguration_logLevelUpdate(t *testing.T) {
	ctx := acctest.Context(t)
	var logDeliveryConfiguration awstypes.LogDeliveryConfigurationType
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cognito_log_delivery_configuration.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLogDeliveryConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccLogDeliveryConfigurationConfig_logLevel(rName, "ERROR"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLogDeliveryConfigurationExists(ctx, t, resourceName, &logDeliveryConfiguration),
					resource.TestCheckResourceAttr(resourceName, "log_configurations.0.log_level", "ERROR"),
				),
			},
			{
				Config: testAccLogDeliveryConfigurationConfig_logLevel(rName, "INFO"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLogDeliveryConfigurationExists(ctx, t, resourceName, &logDeliveryConfiguration),
					resource.TestCheckResourceAttr(resourceName, "log_configurations.0.log_level", "INFO"),
				),
			},
		},
	})
}

func TestAccCognitoIDPLogDeliveryConfiguration_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var logDeliveryConfiguration awstypes.LogDeliveryConfigurationType
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cognito_log_delivery_configuration.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLogDeliveryConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccLogDeliveryConfigurationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLogDeliveryConfigurationExists(ctx, t, resourceName, &logDeliveryConfiguration),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfcognitoidp.ResourceLogDeliveryConfiguration, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccCognitoIDPLogDeliveryConfiguration_firehose(t *testing.T) {
	ctx := acctest.Context(t)
	var logDeliveryConfiguration awstypes.LogDeliveryConfigurationType
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cognito_log_delivery_configuration.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLogDeliveryConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccLogDeliveryConfigurationConfig_firehose(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLogDeliveryConfigurationExists(ctx, t, resourceName, &logDeliveryConfiguration),
					resource.TestCheckResourceAttr(resourceName, "log_configurations.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "log_configurations.0.event_source", "userNotification"),
					resource.TestCheckResourceAttr(resourceName, "log_configurations.0.log_level", "INFO"),
					resource.TestCheckResourceAttr(resourceName, "log_configurations.1.event_source", "userAuthEvents"),
					resource.TestCheckResourceAttr(resourceName, "log_configurations.1.log_level", "ERROR"),
					resource.TestCheckResourceAttrPair(resourceName, "log_configurations.1.firehose_configuration.0.stream_arn", "aws_kinesis_firehose_delivery_stream.test", names.AttrARN),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccLogDeliveryConfigurationImportStateIdFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrUserPoolID,
			},
		},
	})
}

func testAccCheckLogDeliveryConfigurationDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).CognitoIDPClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_cognito_log_delivery_configuration" {
				continue
			}

			out, err := tfcognitoidp.FindLogDeliveryConfigurationByUserPoolID(ctx, conn, rs.Primary.Attributes[names.AttrUserPoolID])

			if errs.IsA[*awstypes.ResourceNotFoundException](err) || (out != nil && out.LogConfigurations == nil) {
				continue
			}

			if err != nil {
				return err
			}

			return create.Error(names.CognitoIDP, create.ErrActionCheckingDestroyed, "Log Delivery Configuration", rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckLogDeliveryConfigurationExists(ctx context.Context, t *testing.T, name string, logDeliveryConfiguration *awstypes.LogDeliveryConfigurationType) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.CognitoIDP, create.ErrActionCheckingExistence, "Log Delivery Configuration", name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.CognitoIDP, create.ErrActionCheckingExistence, "Log Delivery Configuration", name, errors.New("not set"))
		}

		conn := acctest.ProviderMeta(ctx, t).CognitoIDPClient(ctx)

		resp, err := tfcognitoidp.FindLogDeliveryConfigurationByUserPoolID(ctx, conn, rs.Primary.Attributes[names.AttrUserPoolID])

		if err != nil {
			return create.Error(names.CognitoIDP, create.ErrActionCheckingExistence, "Log Delivery Configuration", rs.Primary.ID, err)
		}

		*logDeliveryConfiguration = *resp

		return nil
	}
}

func testAccLogDeliveryConfigurationImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return rs.Primary.Attributes[names.AttrUserPoolID], nil
	}
}

func testAccLogDeliveryConfigurationConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name = %[1]q
}

resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}

resource "aws_cognito_log_delivery_configuration" "test" {
  user_pool_id = aws_cognito_user_pool.test.id

  log_configurations {
    event_source = "userNotification"
    log_level    = "ERROR"

    cloud_watch_logs_configuration {
      log_group_arn = aws_cloudwatch_log_group.test.arn
    }
  }
}
`, rName)
}

func testAccLogDeliveryConfigurationConfig_firehose(rName string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name = %[1]q
}

resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}

resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_iam_role" "firehose" {
  name = "%[1]s-firehose"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "firehose.amazonaws.com"
        }
      }
    ]
  })
}

resource "aws_iam_role_policy" "firehose" {
  name = "%[1]s-firehose"
  role = aws_iam_role.firehose.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "s3:AbortMultipartUpload",
          "s3:GetBucketLocation",
          "s3:GetObject",
          "s3:ListBucket",
          "s3:ListBucketMultipartUploads",
          "s3:PutObject"
        ]
        Resource = [
          aws_s3_bucket.test.arn,
          "${aws_s3_bucket.test.arn}/*"
        ]
      }
    ]
  })
}

resource "aws_kinesis_firehose_delivery_stream" "test" {
  name        = %[1]q
  destination = "extended_s3"

  extended_s3_configuration {
    role_arn   = aws_iam_role.firehose.arn
    bucket_arn = aws_s3_bucket.test.arn
  }
}

resource "aws_cognito_log_delivery_configuration" "test" {
  user_pool_id = aws_cognito_user_pool.test.id

  log_configurations {
    event_source = "userNotification"
    log_level    = "INFO"

    cloud_watch_logs_configuration {
      log_group_arn = aws_cloudwatch_log_group.test.arn
    }
  }

  log_configurations {
    event_source = "userAuthEvents"
    log_level    = "ERROR"

    firehose_configuration {
      stream_arn = aws_kinesis_firehose_delivery_stream.test.arn
    }
  }
}
`, rName)
}

func testAccLogDeliveryConfigurationConfig_logLevel(rName, logLevel string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name = %[1]q
}

resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}

resource "aws_cognito_log_delivery_configuration" "test" {
  user_pool_id = aws_cognito_user_pool.test.id

  log_configurations {
    event_source = "userNotification"
    log_level    = %[2]q

    cloud_watch_logs_configuration {
      log_group_arn = aws_cloudwatch_log_group.test.arn
    }
  }
}
`, rName, logLevel)
}
