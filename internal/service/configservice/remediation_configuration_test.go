// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package configservice_test

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/configservice/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfconfig "github.com/hashicorp/terraform-provider-aws/internal/service/configservice"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccRemediationConfiguration_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var rc types.RemediationConfiguration
	resourceName := "aws_config_remediation_configuration.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rAttempts := sdkacctest.RandIntRange(1, 25)
	rSeconds := sdkacctest.RandIntRange(1, 2678000)
	rExecPct := sdkacctest.RandIntRange(1, 100)
	rErrorPct := sdkacctest.RandIntRange(1, 100)
	sseAlgorithm := "AES256"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ConfigServiceServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRemediationConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRemediationConfigurationConfig_basic(rName, sseAlgorithm, rAttempts, rSeconds, rExecPct, rErrorPct, acctest.CtFalse),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRemediationConfigurationExists(ctx, resourceName, &rc),
					resource.TestCheckResourceAttr(resourceName, "config_rule_name", rName),
					resource.TestCheckResourceAttr(resourceName, "target_id", "AWS-EnableS3BucketEncryption"),
					resource.TestCheckResourceAttr(resourceName, "target_type", "SSM_DOCUMENT"),
					resource.TestCheckResourceAttr(resourceName, "parameter.#", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "automatic", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "maximum_automatic_attempts", strconv.Itoa(rAttempts)),
					resource.TestCheckResourceAttr(resourceName, "retry_attempt_seconds", strconv.Itoa(rSeconds)),
					resource.TestCheckResourceAttr(resourceName, "execution_controls.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "execution_controls.*.ssm_controls.*", map[string]string{
						"concurrent_execution_rate_percentage": strconv.Itoa(rExecPct),
						"error_percentage":                     strconv.Itoa(rErrorPct),
					}),
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

func testAccRemediationConfiguration_basicBackwardCompatible(t *testing.T) {
	ctx := acctest.Context(t)
	var rc types.RemediationConfiguration
	resourceName := "aws_config_remediation_configuration.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sseAlgorithm := "AES256"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ConfigServiceServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRemediationConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRemediationConfigurationConfig_olderSchema(rName, sseAlgorithm),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRemediationConfigurationExists(ctx, resourceName, &rc),
					resource.TestCheckResourceAttr(resourceName, "config_rule_name", rName),
					resource.TestCheckResourceAttr(resourceName, "target_id", "AWS-EnableS3BucketEncryption"),
					resource.TestCheckResourceAttr(resourceName, "target_type", "SSM_DOCUMENT"),
					resource.TestCheckResourceAttr(resourceName, "parameter.#", acctest.Ct3),
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

func testAccRemediationConfiguration_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var rc types.RemediationConfiguration
	resourceName := "aws_config_remediation_configuration.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rAttempts := sdkacctest.RandIntRange(1, 25)
	rSeconds := sdkacctest.RandIntRange(1, 2678000)
	rExecPct := sdkacctest.RandIntRange(1, 100)
	rErrorPct := sdkacctest.RandIntRange(1, 100)
	sseAlgorithm := "AES256"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ConfigServiceServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRemediationConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRemediationConfigurationConfig_basic(rName, sseAlgorithm, rAttempts, rSeconds, rExecPct, rErrorPct, acctest.CtFalse),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRemediationConfigurationExists(ctx, resourceName, &rc),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfconfig.ResourceRemediationConfiguration(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccRemediationConfiguration_updates(t *testing.T) {
	ctx := acctest.Context(t)
	var original types.RemediationConfiguration
	var updated types.RemediationConfiguration
	resourceName := "aws_config_remediation_configuration.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rAttempts := sdkacctest.RandIntRange(1, 25)
	rSeconds := sdkacctest.RandIntRange(1, 2678000)
	rExecPct := sdkacctest.RandIntRange(1, 100)
	rErrorPct := sdkacctest.RandIntRange(1, 100)
	uAttempts := sdkacctest.RandIntRange(1, 25)
	uSeconds := sdkacctest.RandIntRange(1, 2678000)
	uExecPct := sdkacctest.RandIntRange(1, 100)
	uErrorPct := sdkacctest.RandIntRange(1, 100)
	originalSseAlgorithm := "AES256"
	updatedSseAlgorithm := "aws:kms"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ConfigServiceServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRemediationConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRemediationConfigurationConfig_basic(rName, originalSseAlgorithm, rAttempts, rSeconds, rExecPct, rErrorPct, acctest.CtFalse),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRemediationConfigurationExists(ctx, resourceName, &original),
					resource.TestCheckResourceAttr(resourceName, "parameter.2.static_value", originalSseAlgorithm),
					resource.TestCheckResourceAttr(resourceName, "automatic", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "maximum_automatic_attempts", strconv.Itoa(rAttempts)),
					resource.TestCheckResourceAttr(resourceName, "retry_attempt_seconds", strconv.Itoa(rSeconds)),
					resource.TestCheckResourceAttr(resourceName, "execution_controls.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "execution_controls.*.ssm_controls.*", map[string]string{
						"concurrent_execution_rate_percentage": strconv.Itoa(rExecPct),
						"error_percentage":                     strconv.Itoa(rErrorPct),
					}),
				),
			},
			{
				Config: testAccRemediationConfigurationConfig_basic(rName, updatedSseAlgorithm, uAttempts, uSeconds, uExecPct, uErrorPct, acctest.CtTrue),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRemediationConfigurationExists(ctx, resourceName, &updated),
					testAccCheckRemediationConfigurationNotRecreated(&original, &updated),
					resource.TestCheckResourceAttr(resourceName, "parameter.2.static_value", updatedSseAlgorithm),
					resource.TestCheckResourceAttr(resourceName, "automatic", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "maximum_automatic_attempts", strconv.Itoa(uAttempts)),
					resource.TestCheckResourceAttr(resourceName, "retry_attempt_seconds", strconv.Itoa(uSeconds)),
					resource.TestCheckResourceAttr(resourceName, "execution_controls.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "execution_controls.*.ssm_controls.*", map[string]string{
						"concurrent_execution_rate_percentage": strconv.Itoa(uExecPct),
						"error_percentage":                     strconv.Itoa(uErrorPct),
					}),
				),
			},
		},
	})
}

func testAccRemediationConfiguration_values(t *testing.T) {
	ctx := acctest.Context(t)
	var rc types.RemediationConfiguration
	resourceName := "aws_config_remediation_configuration.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rAttempts := sdkacctest.RandIntRange(1, 25)
	rSeconds := sdkacctest.RandIntRange(1, 2678000)
	rExecPct := sdkacctest.RandIntRange(1, 100)
	rErrorPct := sdkacctest.RandIntRange(1, 100)
	sseAlgorithm := "AES256"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ConfigServiceServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRemediationConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRemediationConfigurationConfig_values(rName, sseAlgorithm, rAttempts, rSeconds, rExecPct, rErrorPct, acctest.CtFalse),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRemediationConfigurationExists(ctx, resourceName, &rc),
					resource.TestCheckResourceAttr(resourceName, "target_id", "AWS-EnableS3BucketEncryption"),
					resource.TestCheckResourceAttr(resourceName, "target_type", "SSM_DOCUMENT"),
					resource.TestCheckResourceAttr(resourceName, "parameter.#", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "automatic", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "maximum_automatic_attempts", strconv.Itoa(rAttempts)),
					resource.TestCheckResourceAttr(resourceName, "retry_attempt_seconds", strconv.Itoa(rSeconds)),
					resource.TestCheckResourceAttr(resourceName, "execution_controls.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "execution_controls.*.ssm_controls.*", map[string]string{
						"concurrent_execution_rate_percentage": strconv.Itoa(rExecPct),
						"error_percentage":                     strconv.Itoa(rErrorPct),
					}),
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

func testAccCheckRemediationConfigurationExists(ctx context.Context, n string, v *types.RemediationConfiguration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ConfigServiceClient(ctx)

		output, err := tfconfig.FindRemediationConfigurationByConfigRuleName(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckRemediationConfigurationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ConfigServiceClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_config_remediation_configuration" {
				continue
			}

			_, err := tfconfig.FindRemediationConfigurationByConfigRuleName(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("ConfigService Remediation Configuration %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckRemediationConfigurationNotRecreated(before, after *types.RemediationConfiguration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.ToString(before.Arn) != aws.ToString(after.Arn) {
			return errors.New("ConfigService Remediation Configuration was recreated")
		}
		return nil
	}
}

func testAccRemediationConfigurationConfig_olderSchema(rName, sseAlgorithm string) string {
	return fmt.Sprintf(`
resource "aws_config_remediation_configuration" "test" {
  config_rule_name = aws_config_config_rule.test.name

  resource_type  = "AWS::S3::Bucket"
  target_id      = "AWS-EnableS3BucketEncryption"
  target_type    = "SSM_DOCUMENT"
  target_version = "1"

  parameter {
    name         = "AutomationAssumeRole"
    static_value = aws_iam_role.test.arn
  }
  parameter {
    name           = "BucketName"
    resource_value = "RESOURCE_ID"
  }
  parameter {
    name         = "SSEAlgorithm"
    static_value = %[2]q
  }
}

resource "aws_sns_topic" "test" {
  name = %[1]q
}

resource "aws_config_config_rule" "test" {
  name = %[1]q

  source {
    owner             = "AWS"
    source_identifier = "S3_BUCKET_VERSIONING_ENABLED"
  }

  depends_on = [aws_config_configuration_recorder.test]
}

resource "aws_config_configuration_recorder" "test" {
  name     = %[1]q
  role_arn = aws_iam_role.test.arn
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
        "Service": "config.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "test" {
  name = %[1]q
  role = aws_iam_role.test.id

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
        "Action": "config:Put*",
        "Effect": "Allow",
        "Resource": "*"

    }
  ]
}
EOF
}
`, rName, sseAlgorithm)
}

func testAccRemediationConfigurationConfig_basic(rName, sseAlgorithm string, randAttempts, randSeconds, randExecPct, randErrorPct int, automatic string) string {
	return fmt.Sprintf(`
resource "aws_config_remediation_configuration" "test" {
  config_rule_name = aws_config_config_rule.test.name

  resource_type  = "AWS::S3::Bucket"
  target_id      = "AWS-EnableS3BucketEncryption"
  target_type    = "SSM_DOCUMENT"
  target_version = "1"

  parameter {
    name         = "AutomationAssumeRole"
    static_value = aws_iam_role.test.arn
  }
  parameter {
    name           = "BucketName"
    resource_value = "RESOURCE_ID"
  }
  parameter {
    name         = "SSEAlgorithm"
    static_value = %[2]q
  }
  automatic                  = %[7]s
  maximum_automatic_attempts = %[3]d
  retry_attempt_seconds      = %[4]d
  execution_controls {
    ssm_controls {
      concurrent_execution_rate_percentage = %[5]d
      error_percentage                     = %[6]d
    }
  }
}

resource "aws_sns_topic" "test" {
  name = %[1]q
}

resource "aws_config_config_rule" "test" {
  name = %[1]q

  source {
    owner             = "AWS"
    source_identifier = "S3_BUCKET_VERSIONING_ENABLED"
  }

  depends_on = [aws_config_configuration_recorder.test]
}

resource "aws_config_configuration_recorder" "test" {
  name     = %[1]q
  role_arn = aws_iam_role.test.arn
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
        "Service": "config.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "test" {
  name = %[1]q
  role = aws_iam_role.test.id

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
        "Action": "config:Put*",
        "Effect": "Allow",
        "Resource": "*"

    }
  ]
}
EOF
}
`, rName, sseAlgorithm, randAttempts, randSeconds, randExecPct, randErrorPct, automatic)
}

func testAccRemediationConfigurationConfig_values(rName, sseAlgorithm string, randAttempts int, randSeconds int, randExecPct int, randErrorPct int, automatic string) string {
	return fmt.Sprintf(`
resource "aws_config_remediation_configuration" "test" {
  config_rule_name = aws_config_config_rule.test.name

  resource_type  = "AWS::S3::Bucket"
  target_id      = "AWS-EnableS3BucketEncryption"
  target_type    = "SSM_DOCUMENT"
  target_version = "1"

  parameter {
    name          = "AutomationAssumeRole"
    static_values = [aws_iam_role.test.arn, aws_iam_role.test2.arn]
  }

  parameter {
    name           = "BucketName"
    resource_value = "RESOURCE_ID"
  }

  parameter {
    name         = "SSEAlgorithm"
    static_value = %[2]q
  }

  automatic                  = %[7]s
  maximum_automatic_attempts = %[3]d
  retry_attempt_seconds      = %[4]d

  execution_controls {
    ssm_controls {
      concurrent_execution_rate_percentage = %[5]d
      error_percentage                     = %[6]d
    }
  }
}

resource "aws_sns_topic" "test" {
  name = %[1]q
}

resource "aws_config_config_rule" "test" {
  name = %[1]q

  source {
    owner             = "AWS"
    source_identifier = "S3_BUCKET_VERSIONING_ENABLED"
  }

  depends_on = [aws_config_configuration_recorder.test]
}

resource "aws_config_configuration_recorder" "test" {
  name     = %[1]q
  role_arn = aws_iam_role.test.arn
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
        "Service": "config.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_role" "test2" {
  name = "%[1]s-2"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "config.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "test" {
  name = %[1]q
  role = aws_iam_role.test.id

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
        "Action": "config:Put*",
        "Effect": "Allow",
        "Resource": "*"

    }
  ]
}
EOF
}
`, rName, sseAlgorithm, randAttempts, randSeconds, randExecPct, randErrorPct, automatic)
}
