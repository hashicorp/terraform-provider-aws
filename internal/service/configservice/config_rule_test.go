// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package configservice_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
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

func testAccConfigRule_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var cr types.ConfigRule
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_config_config_rule.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ConfigServiceServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConfigRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConfigRuleConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckConfigRuleExists(ctx, resourceName, &cr),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "source.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "source.0.owner", "AWS"),
					resource.TestCheckResourceAttr(resourceName, "source.0.source_identifier", "S3_BUCKET_VERSIONING_ENABLED"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "evaluation_mode.*", map[string]string{
						names.AttrMode: "DETECTIVE",
					}),
				),
			},
		},
	})
}

func testAccConfigRule_evaluationMode(t *testing.T) {
	ctx := acctest.Context(t)
	var cr types.ConfigRule
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_config_config_rule.test"

	evaluationMode1 := `
  evaluation_mode {
    mode = "DETECTIVE"
  }
`

	evaluationMode2 := `
  evaluation_mode {
    mode = "DETECTIVE"
  }

  evaluation_mode {
    mode = "PROACTIVE"
  }
`
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ConfigServiceServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConfigRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConfigRuleConfig_evaluationMode(rName, evaluationMode1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigRuleExists(ctx, resourceName, &cr),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "evaluation_mode.*", map[string]string{
						names.AttrMode: "DETECTIVE",
					}),
				),
			},
			{
				Config: testAccConfigRuleConfig_evaluationMode(rName, evaluationMode2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigRuleExists(ctx, resourceName, &cr),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "evaluation_mode.*", map[string]string{
						names.AttrMode: "DETECTIVE",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "evaluation_mode.*", map[string]string{
						names.AttrMode: "PROACTIVE",
					}),
				),
			},
		},
	})
}

func testAccConfigRule_ownerAWS(t *testing.T) { // nosemgrep:ci.aws-in-func-name
	ctx := acctest.Context(t)
	var cr types.ConfigRule
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_config_config_rule.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ConfigServiceServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConfigRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConfigRuleConfig_ownerAWS(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigRuleExists(ctx, resourceName, &cr),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "config", regexache.MustCompile("config-rule/config-rule-[0-9a-z]+$")),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestMatchResourceAttr(resourceName, "rule_id", regexache.MustCompile("config-rule-[0-9a-z]+$")),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Terraform Acceptance tests"),
					resource.TestCheckResourceAttr(resourceName, "source.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "source.0.owner", "AWS"),
					resource.TestCheckResourceAttr(resourceName, "source.0.source_identifier", "REQUIRED_TAGS"),
					resource.TestCheckResourceAttr(resourceName, "source.0.source_detail.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "scope.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "scope.0.compliance_resource_id", "blablah"),
					resource.TestCheckResourceAttr(resourceName, "scope.0.compliance_resource_types.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "scope.0.compliance_resource_types.*", "AWS::EC2::Instance"),
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

func testAccConfigRule_customlambda(t *testing.T) {
	ctx := acctest.Context(t)
	var cr types.ConfigRule
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_config_config_rule.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ConfigServiceServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConfigRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConfigRuleConfig_customLambda(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigRuleExists(ctx, resourceName, &cr),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "config", regexache.MustCompile("config-rule/config-rule-[0-9a-z]+$")),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestMatchResourceAttr(resourceName, "rule_id", regexache.MustCompile("config-rule-[0-9a-z]+$")),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Terraform Acceptance tests"),
					resource.TestCheckResourceAttr(resourceName, "maximum_execution_frequency", "Six_Hours"),
					resource.TestCheckResourceAttr(resourceName, "source.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "source.0.owner", "CUSTOM_LAMBDA"),
					resource.TestCheckResourceAttrPair(resourceName, "source.0.source_identifier", "aws_lambda_function.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "source.0.source_detail.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "source.0.source_detail.*", map[string]string{
						"event_source":                "aws.config",
						"message_type":                "ConfigurationSnapshotDeliveryCompleted",
						"maximum_execution_frequency": "",
					}),
					resource.TestCheckResourceAttr(resourceName, "scope.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "scope.0.tag_key", "IsTemporary"),
					resource.TestCheckResourceAttr(resourceName, "scope.0.tag_value", "yes"),
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

func testAccConfigRule_ownerPolicy(t *testing.T) {
	ctx := acctest.Context(t)
	var cr types.ConfigRule
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_config_config_rule.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ConfigServiceServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConfigRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConfigRuleConfig_ownerPolicy(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigRuleExists(ctx, resourceName, &cr),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "config", regexache.MustCompile("config-rule/config-rule-[0-9a-z]+$")),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestMatchResourceAttr(resourceName, "rule_id", regexache.MustCompile("config-rule-[0-9a-z]+$")),
					resource.TestCheckResourceAttr(resourceName, "scope.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "source.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "source.0.owner", "CUSTOM_POLICY"),
					resource.TestCheckResourceAttr(resourceName, "source.0.source_detail.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "source.0.source_detail.0.message_type", "ConfigurationItemChangeNotification"),
					resource.TestCheckResourceAttr(resourceName, "source.0.custom_policy_details.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "source.0.custom_policy_details.0.enable_debug_log_delivery", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "source.0.custom_policy_details.0.policy_runtime", "guard-2.x.x"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"source.0.custom_policy_details.0.policy_text"},
			},
			{
				Config: testAccConfigRuleConfig_ownerPolicy(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigRuleExists(ctx, resourceName, &cr),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "config", regexache.MustCompile("config-rule/config-rule-[0-9a-z]+$")),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestMatchResourceAttr(resourceName, "rule_id", regexache.MustCompile("config-rule-[0-9a-z]+$")),
					resource.TestCheckResourceAttr(resourceName, "scope.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "source.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "source.0.owner", "CUSTOM_POLICY"),
					resource.TestCheckResourceAttr(resourceName, "source.0.source_detail.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "source.0.source_detail.0.message_type", "ConfigurationItemChangeNotification"),
					resource.TestCheckResourceAttr(resourceName, "source.0.custom_policy_details.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "source.0.custom_policy_details.0.enable_debug_log_delivery", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "source.0.custom_policy_details.0.policy_runtime", "guard-2.x.x"),
				),
			},
		},
	})
}

func testAccConfigRule_Scope_TagKey(t *testing.T) {
	ctx := acctest.Context(t)
	var configRule types.ConfigRule
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_config_config_rule.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ConfigServiceServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConfigRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConfigRuleConfig_scopeTagKey(rName, acctest.CtKey1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigRuleExists(ctx, resourceName, &configRule),
					resource.TestCheckResourceAttr(resourceName, "scope.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "scope.0.tag_key", acctest.CtKey1),
				),
			},
			{
				Config: testAccConfigRuleConfig_scopeTagKey(rName, acctest.CtKey2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigRuleExists(ctx, resourceName, &configRule),
					resource.TestCheckResourceAttr(resourceName, "scope.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "scope.0.tag_key", acctest.CtKey2),
				),
			},
		},
	})
}

func testAccConfigRule_Scope_TagKey_Empty(t *testing.T) {
	ctx := acctest.Context(t)
	var configRule types.ConfigRule
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_config_config_rule.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ConfigServiceServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConfigRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConfigRuleConfig_scopeTagKey(rName, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigRuleExists(ctx, resourceName, &configRule),
				),
			},
		},
	})
}

func testAccConfigRule_Scope_TagValue(t *testing.T) {
	ctx := acctest.Context(t)
	var configRule types.ConfigRule
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_config_config_rule.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ConfigServiceServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConfigRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConfigRuleConfig_scopeTagValue(rName, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigRuleExists(ctx, resourceName, &configRule),
					resource.TestCheckResourceAttr(resourceName, "scope.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "scope.0.tag_value", acctest.CtValue1),
				),
			},
			{
				Config: testAccConfigRuleConfig_scopeTagValue(rName, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigRuleExists(ctx, resourceName, &configRule),
					resource.TestCheckResourceAttr(resourceName, "scope.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "scope.0.tag_value", acctest.CtValue2),
				),
			},
		},
	})
}

func testAccConfigRule_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var cr types.ConfigRule
	resourceName := "aws_config_config_rule.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ConfigServiceServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConfigRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConfigRuleConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigRuleExists(ctx, resourceName, &cr),
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
				Config: testAccConfigRuleConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigRuleExists(ctx, resourceName, &cr),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccConfigRuleConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigRuleExists(ctx, resourceName, &cr),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccConfigRule_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var cr types.ConfigRule
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_config_config_rule.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ConfigServiceServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConfigRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConfigRuleConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigRuleExists(ctx, resourceName, &cr),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfconfig.ResourceConfigRule(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckConfigRuleExists(ctx context.Context, n string, v *types.ConfigRule) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ConfigServiceClient(ctx)

		output, err := tfconfig.FindConfigRuleByName(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckConfigRuleDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ConfigServiceClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_config_config_rule" {
				continue
			}

			_, err := tfconfig.FindConfigRuleByName(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("ConfigService Config Rule %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccConfigRuleConfig_base(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

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
        "Service": "config.${data.aws_partition.current.dns_suffix}"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_role_policy_attachment" "test" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AWS_ConfigRole"
  role       = aws_iam_role.test.name
}
`, rName)
}

func testAccConfigRuleConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccConfigRuleConfig_base(rName), fmt.Sprintf(`
resource "aws_config_config_rule" "test" {
  name = %[1]q

  source {
    owner             = "AWS"
    source_identifier = "S3_BUCKET_VERSIONING_ENABLED"
  }

  depends_on = [aws_config_configuration_recorder.test]
}
`, rName))
}

func testAccConfigRuleConfig_ownerAWS(rName string) string { // nosemgrep:ci.aws-in-func-name
	return acctest.ConfigCompose(testAccConfigRuleConfig_base(rName), fmt.Sprintf(`
resource "aws_config_config_rule" "test" {
  name        = %[1]q
  description = "Terraform Acceptance tests"

  source {
    owner             = "AWS"
    source_identifier = "REQUIRED_TAGS"
  }

  scope {
    compliance_resource_id    = "blablah"
    compliance_resource_types = ["AWS::EC2::Instance"]
  }

  input_parameters = <<PARAMS
{
  "tag1Key": "CostCenter",
  "tag2Key": "Owner"
}
PARAMS

  depends_on = [aws_config_configuration_recorder.test]
}
`, rName))
}

func testAccConfigRuleConfig_ownerPolicy(rName string, enableDebugLogDelivery bool) string {
	return acctest.ConfigCompose(testAccConfigRuleConfig_base(rName), fmt.Sprintf(`
resource "aws_config_config_rule" "test" {
  name = %[1]q

  source {
    owner = "CUSTOM_POLICY"

    source_detail {
      message_type = "ConfigurationItemChangeNotification"
    }

    custom_policy_details {
      policy_runtime = "guard-2.x.x"
      policy_text    = <<EOF
	  rule tableisactive when
		  resourceType == "AWS::DynamoDB::Table" {
		  configuration.tableStatus == ['ACTIVE']
	  }

	  rule checkcompliance when
		  resourceType == "AWS::DynamoDB::Table"
		  tableisactive {
			  supplementaryConfiguration.ContinuousBackupsDescription.pointInTimeRecoveryDescription.pointInTimeRecoveryStatus == "ENABLED"
	  }
EOF

      enable_debug_log_delivery = %[2]t
    }
  }

  depends_on = [aws_config_configuration_recorder.test]
}
`, rName, enableDebugLogDelivery))
}

func testAccConfigRuleConfig_customLambda(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigLambdaBase(rName, rName, rName), fmt.Sprintf(`
resource "aws_config_config_rule" "test" {
  name                        = %[1]q
  description                 = "Terraform Acceptance tests"
  maximum_execution_frequency = "Six_Hours"

  source {
    owner             = "CUSTOM_LAMBDA"
    source_identifier = aws_lambda_function.test.arn

    source_detail {
      event_source = "aws.config"
      message_type = "ConfigurationSnapshotDeliveryCompleted"
    }
  }

  scope {
    tag_key   = "IsTemporary"
    tag_value = "yes"
  }

  depends_on = [
    aws_config_configuration_recorder.test,
    aws_config_delivery_channel.test,
  ]
}

resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "index.handler"
  runtime       = "nodejs20.x"
}

resource "aws_lambda_permission" "test" {
  statement_id  = "AllowExecutionFromConfig"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.test.arn
  principal     = "config.${data.aws_partition.current.dns_suffix}"
}

resource "aws_iam_role_policy_attachment" "test" {
  role       = aws_iam_role.iam_for_lambda.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AWSConfigRulesExecutionRole"
}

resource "aws_config_delivery_channel" "test" {
  name           = %[1]q
  s3_bucket_name = aws_s3_bucket.test.bucket

  snapshot_delivery_properties {
    delivery_frequency = "Six_Hours"
  }

  depends_on = [aws_config_configuration_recorder.test]
}

resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_config_configuration_recorder" "test" {
  name     = %[1]q
  role_arn = aws_iam_role.test.arn
}

resource "aws_iam_role" "test" {
  name = "%[1]s_config"

  assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "config.${data.aws_partition.current.dns_suffix}"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
POLICY
}

resource "aws_iam_role_policy" "test" {
  name = "%[1]s_config"
  role = aws_iam_role.test.id

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "config:Put*",
      "Effect": "Allow",
      "Resource": "*"
    },
    {
      "Action": "s3:*",
      "Effect": "Allow",
      "Resource": [
        "${aws_s3_bucket.test.arn}",
        "${aws_s3_bucket.test.arn}/*"
      ]
    },
    {
      "Action": "lambda:*",
      "Effect": "Allow",
      "Resource": "${aws_lambda_function.test.arn}"
    }
  ]
}
POLICY
}
`, rName))
}

func testAccConfigRuleConfig_scopeTagKey(rName, tagKey string) string {
	return acctest.ConfigCompose(testAccConfigRuleConfig_base(rName), fmt.Sprintf(`
resource "aws_config_config_rule" "test" {
  name = %[1]q

  scope {
    tag_key = %[2]q
  }

  source {
    owner             = "AWS"
    source_identifier = "S3_BUCKET_VERSIONING_ENABLED"
  }

  depends_on = [aws_config_configuration_recorder.test]
}
`, rName, tagKey))
}

func testAccConfigRuleConfig_scopeTagValue(rName, tagValue string) string {
	return acctest.ConfigCompose(testAccConfigRuleConfig_base(rName), fmt.Sprintf(`
resource "aws_config_config_rule" "test" {
  name = %[1]q

  scope {
    tag_key   = "key"
    tag_value = %[2]q
  }

  source {
    owner             = "AWS"
    source_identifier = "S3_BUCKET_VERSIONING_ENABLED"
  }

  depends_on = [aws_config_configuration_recorder.test]
}
`, rName, tagValue))
}

func testAccConfigRuleConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccConfigRuleConfig_base(rName), fmt.Sprintf(`
resource "aws_config_config_rule" "test" {
  name = %[1]q

  source {
    owner             = "AWS"
    source_identifier = "S3_BUCKET_VERSIONING_ENABLED"
  }

  tags = {
    %[2]s = %[3]q
  }

  depends_on = [aws_config_configuration_recorder.test]
}
`, rName, tagKey1, tagValue1))
}

func testAccConfigRuleConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccConfigRuleConfig_base(rName), fmt.Sprintf(`
resource "aws_config_config_rule" "test" {
  name = %[1]q

  source {
    owner             = "AWS"
    source_identifier = "S3_BUCKET_VERSIONING_ENABLED"
  }

  tags = {
    %[2]s = %[3]q
    %[4]s = %[5]q
  }

  depends_on = [aws_config_configuration_recorder.test]
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccConfigRuleConfig_evaluationMode(rName, evaluationMode string) string {
	return acctest.ConfigCompose(testAccConfigRuleConfig_base(rName), fmt.Sprintf(`
resource "aws_config_config_rule" "test" {
  name = %[1]q

  source {
    owner             = "AWS"
    source_identifier = "EIP_ATTACHED"
  }

%[2]s

  depends_on = [aws_config_configuration_recorder.test]
}
`, rName, evaluationMode))
}
