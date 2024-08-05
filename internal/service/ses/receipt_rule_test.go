// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ses_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfses "github.com/hashicorp/terraform-provider-aws/internal/service/ses"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSESReceiptRule_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var rule ses.ReceiptRule
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ses_receipt_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
			testAccPreCheckReceiptRule(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SESServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReceiptRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccReceiptRuleConfig_basic(rName, acctest.DefaultEmailAddress),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReceiptRuleExists(ctx, resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "rule_set_name", rName),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "ses", fmt.Sprintf("receipt-rule-set/%s:receipt-rule/%s", rName, rName)),
					resource.TestCheckResourceAttr(resourceName, "add_header_action.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "bounce_action.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "lambda_action.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "s3_action.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "sns_action.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "stop_action.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "workmail_action.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "recipients.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "recipients.*", acctest.DefaultEmailAddress),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "scan_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "tls_policy", "Require"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccReceiptRuleImportStateIdFunc(resourceName),
			},
		},
	})
}

func TestAccSESReceiptRule_s3Action(t *testing.T) {
	ctx := acctest.Context(t)
	var rule ses.ReceiptRule
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ses_receipt_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
			testAccPreCheckReceiptRule(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SESServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReceiptRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccReceiptRuleConfig_s3Action(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReceiptRuleExists(ctx, resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "s3_action.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "s3_action.*.bucket_name", "aws_s3_bucket.test", names.AttrID),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "s3_action.*", map[string]string{
						"position": acctest.Ct1,
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccReceiptRuleImportStateIdFunc(resourceName),
			},
		},
	})
}

func TestAccSESReceiptRule_snsAction(t *testing.T) {
	ctx := acctest.Context(t)
	var rule ses.ReceiptRule
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ses_receipt_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
			testAccPreCheckReceiptRule(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SESServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReceiptRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccReceiptRuleConfig_snsAction(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReceiptRuleExists(ctx, resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "sns_action.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "sns_action.*.topic_arn", "aws_sns_topic.test", names.AttrARN),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "sns_action.*", map[string]string{
						"encoding": "UTF-8",
						"position": acctest.Ct1,
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccReceiptRuleImportStateIdFunc(resourceName),
			},
		},
	})
}

func TestAccSESReceiptRule_snsActionEncoding(t *testing.T) {
	ctx := acctest.Context(t)
	var rule ses.ReceiptRule
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ses_receipt_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
			testAccPreCheckReceiptRule(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SESServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReceiptRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccReceiptRuleConfig_snsActionEncoding(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReceiptRuleExists(ctx, resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "sns_action.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "sns_action.*.topic_arn", "aws_sns_topic.test", names.AttrARN),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "sns_action.*", map[string]string{
						"encoding": "Base64",
						"position": acctest.Ct1,
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccReceiptRuleImportStateIdFunc(resourceName),
			},
		},
	})
}

func TestAccSESReceiptRule_lambdaAction(t *testing.T) {
	ctx := acctest.Context(t)
	var rule ses.ReceiptRule
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ses_receipt_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
			testAccPreCheckReceiptRule(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SESServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReceiptRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccReceiptRuleConfig_lambdaAction(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReceiptRuleExists(ctx, resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "lambda_action.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "lambda_action.*.function_arn", "aws_lambda_function.test", names.AttrARN),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "lambda_action.*", map[string]string{
						"invocation_type": "Event",
						"position":        acctest.Ct1,
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccReceiptRuleImportStateIdFunc(resourceName),
			},
		},
	})
}

func TestAccSESReceiptRule_stopAction(t *testing.T) {
	ctx := acctest.Context(t)
	var rule ses.ReceiptRule
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ses_receipt_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
			testAccPreCheckReceiptRule(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SESServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReceiptRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccReceiptRuleConfig_stopAction(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReceiptRuleExists(ctx, resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "stop_action.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "stop_action.*", map[string]string{
						names.AttrScope: "RuleSet",
					}),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "stop_action.*.topic_arn", "aws_sns_topic.test", names.AttrARN),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccReceiptRuleImportStateIdFunc(resourceName),
			},
		},
	})
}

func TestAccSESReceiptRule_order(t *testing.T) {
	ctx := acctest.Context(t)
	var rule ses.ReceiptRule
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ses_receipt_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
			testAccPreCheckReceiptRule(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SESServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReceiptRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccReceiptRuleConfig_order(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReceiptRuleExists(ctx, resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, "second"),
					resource.TestCheckResourceAttrPair(resourceName, "after", "aws_ses_receipt_rule.test1", names.AttrName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccReceiptRuleImportStateIdFunc(resourceName),
			},
		},
	})
}

func TestAccSESReceiptRule_actions(t *testing.T) {
	ctx := acctest.Context(t)
	var rule ses.ReceiptRule
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ses_receipt_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
			testAccPreCheckReceiptRule(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SESServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReceiptRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccReceiptRuleConfig_actions(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReceiptRuleExists(ctx, resourceName, &rule),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "add_header_action.*", map[string]string{
						"header_name":  "Added-By",
						"header_value": "Terraform",
						"position":     acctest.Ct2,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "add_header_action.*", map[string]string{
						"header_name":  "Another-Header",
						"header_value": "First",
						"position":     acctest.Ct1,
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccReceiptRuleImportStateIdFunc(resourceName),
			},
		},
	})
}

func TestAccSESReceiptRule_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var rule ses.ReceiptRule
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ses_receipt_rule.test"
	ruleSetResourceName := "aws_ses_receipt_rule_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
			testAccPreCheckReceiptRule(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SESServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReceiptRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccReceiptRuleConfig_basic(rName, acctest.DefaultEmailAddress),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReceiptRuleExists(ctx, resourceName, &rule),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfses.ResourceReceiptRuleSet(), ruleSetResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccReceiptRuleConfig_basic(rName, acctest.DefaultEmailAddress),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReceiptRuleExists(ctx, resourceName, &rule),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfses.ResourceReceiptRule(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckReceiptRuleDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SESConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ses_receipt_rule" {
				continue
			}

			_, err := tfses.FindReceiptRuleByTwoPartKey(ctx, conn, rs.Primary.ID, rs.Primary.Attributes["rule_set_name"])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("SES Receipt Rule %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckReceiptRuleExists(ctx context.Context, n string, v *ses.ReceiptRule) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No SES Receipt Rule ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SESConn(ctx)

		output, err := tfses.FindReceiptRuleByTwoPartKey(ctx, conn, rs.Primary.ID, rs.Primary.Attributes["rule_set_name"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccReceiptRuleImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not Found: %s", resourceName)
		}

		return fmt.Sprintf("%s:%s", rs.Primary.Attributes["rule_set_name"], rs.Primary.Attributes[names.AttrName]), nil
	}
}

func testAccPreCheckReceiptRule(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).SESConn(ctx)

	input := &ses.DescribeReceiptRuleInput{
		RuleName:    aws.String("MyRule"),
		RuleSetName: aws.String("MyRuleSet"),
	}

	_, err := conn.DescribeReceiptRuleWithContext(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if tfawserr.ErrCodeEquals(err, "RuleSetDoesNotExist") {
		return
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccReceiptRuleConfig_basic(rName, email string) string {
	return fmt.Sprintf(`
resource "aws_ses_receipt_rule_set" "test" {
  rule_set_name = %[1]q
}

resource "aws_ses_receipt_rule" "test" {
  name          = %[1]q
  rule_set_name = aws_ses_receipt_rule_set.test.rule_set_name
  recipients    = [%[2]q]
  enabled       = true
  scan_enabled  = true
  tls_policy    = "Require"
}
`, rName, email)
}

func testAccReceiptRuleConfig_s3Action(rName string) string {
	return fmt.Sprintf(`
resource "aws_ses_receipt_rule_set" "test" {
  rule_set_name = %[1]q
}

resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = "true"
}

resource "aws_s3_bucket_public_access_block" "test" {
  bucket = aws_s3_bucket.test.id

  block_public_acls       = false
  block_public_policy     = false
  ignore_public_acls      = false
  restrict_public_buckets = false
}

resource "aws_s3_bucket_ownership_controls" "test" {
  bucket = aws_s3_bucket.test.id
  rule {
    object_ownership = "BucketOwnerPreferred"
  }
}

resource "aws_s3_bucket_acl" "test" {
  depends_on = [
    aws_s3_bucket_public_access_block.test,
    aws_s3_bucket_ownership_controls.test,
  ]

  bucket = aws_s3_bucket.test.id
  acl    = "public-read-write"
}

resource "aws_ses_receipt_rule" "test" {
  name          = %[1]q
  rule_set_name = aws_ses_receipt_rule_set.test.rule_set_name
  recipients    = [%[2]q]
  enabled       = true
  scan_enabled  = true
  tls_policy    = "Require"

  s3_action {
    bucket_name = aws_s3_bucket_acl.test.bucket
    position    = 1
  }
}
`, rName, acctest.DefaultEmailAddress)
}

func testAccReceiptRuleConfig_snsAction(rName string) string {
	return fmt.Sprintf(`
resource "aws_ses_receipt_rule_set" "test" {
  rule_set_name = %[1]q
}

resource "aws_sns_topic" "test" {
  name = %[1]q
}

resource "aws_ses_receipt_rule" "test" {
  name          = %[1]q
  rule_set_name = aws_ses_receipt_rule_set.test.rule_set_name
  recipients    = [%[2]q]
  enabled       = true
  scan_enabled  = true
  tls_policy    = "Require"

  sns_action {
    topic_arn = aws_sns_topic.test.arn
    position  = 1
  }
}
`, rName, acctest.DefaultEmailAddress)
}

func testAccReceiptRuleConfig_snsActionEncoding(rName string) string {
	return fmt.Sprintf(`
resource "aws_ses_receipt_rule_set" "test" {
  rule_set_name = %[1]q
}

resource "aws_sns_topic" "test" {
  name = %[1]q
}

resource "aws_ses_receipt_rule" "test" {
  name          = %[1]q
  rule_set_name = aws_ses_receipt_rule_set.test.rule_set_name
  recipients    = [%[2]q]
  enabled       = true
  scan_enabled  = true
  tls_policy    = "Require"

  sns_action {
    topic_arn = aws_sns_topic.test.arn
    encoding  = "Base64"
    position  = 1
  }
}
`, rName, acctest.DefaultEmailAddress)
}

func testAccReceiptRuleConfig_lambdaAction(rName string) string {
	return fmt.Sprintf(`
resource "aws_ses_receipt_rule_set" "test" {
  rule_set_name = %[1]q
}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = jsonencode({
    "Version" : "2012-10-17",
    "Statement" : [
      {
        "Action" : "sts:AssumeRole",
        "Principal" : {
          "Service" : "lambda.amazonaws.com"
        },
        "Effect" : "Allow",
        "Sid" : ""
      }
    ]
  })
}

resource "aws_lambda_function" "test" {
  filename         = "test-fixtures/lambdatest.zip"
  source_code_hash = filebase64sha256("test-fixtures/lambdatest.zip")
  function_name    = %[1]q
  role             = aws_iam_role.test.arn
  handler          = "exports.example"
  runtime          = "nodejs16.x"
}

resource "aws_lambda_permission" "test" {
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.test.arn
  principal     = "ses.amazonaws.com"
}

resource "aws_ses_receipt_rule" "test" {
  name          = %[1]q
  rule_set_name = aws_ses_receipt_rule_set.test.rule_set_name
  recipients    = [%[2]q]
  enabled       = true
  scan_enabled  = true
  tls_policy    = "Require"

  lambda_action {
    function_arn = aws_lambda_function.test.arn
    position     = 1
  }

  depends_on = [aws_lambda_permission.test]
}
`, rName, acctest.DefaultEmailAddress)
}

func testAccReceiptRuleConfig_stopAction(rName string) string {
	return fmt.Sprintf(`
resource "aws_ses_receipt_rule_set" "test" {
  rule_set_name = %[1]q
}

resource "aws_sns_topic" "test" {
  name = %[1]q
}

resource "aws_ses_receipt_rule" "test" {
  name          = %[1]q
  rule_set_name = aws_ses_receipt_rule_set.test.rule_set_name
  recipients    = [%[2]q]
  enabled       = true
  scan_enabled  = true
  tls_policy    = "Require"

  stop_action {
    topic_arn = aws_sns_topic.test.arn
    scope     = "RuleSet"
    position  = 1
  }
}
`, rName, acctest.DefaultEmailAddress)
}

func testAccReceiptRuleConfig_order(rName string) string {
	return fmt.Sprintf(`
resource "aws_ses_receipt_rule_set" "test" {
  rule_set_name = %[1]q
}

resource "aws_ses_receipt_rule" "test" {
  name          = "second"
  rule_set_name = aws_ses_receipt_rule_set.test.rule_set_name
  after         = aws_ses_receipt_rule.test1.name
}

resource "aws_ses_receipt_rule" "test1" {
  name          = "first"
  rule_set_name = aws_ses_receipt_rule_set.test.rule_set_name
}
`, rName)
}

func testAccReceiptRuleConfig_actions(rName string) string {
	return fmt.Sprintf(`
resource "aws_ses_receipt_rule_set" "test" {
  rule_set_name = %[1]q
}

resource "aws_ses_receipt_rule" "test" {
  name          = %[1]q
  rule_set_name = aws_ses_receipt_rule_set.test.rule_set_name

  add_header_action {
    header_name  = "Added-By"
    header_value = "Terraform"
    position     = 2
  }

  add_header_action {
    header_name  = "Another-Header"
    header_value = "First"
    position     = 1
  }

  stop_action {
    scope    = "RuleSet"
    position = 3
  }
}
`, rName)
}
