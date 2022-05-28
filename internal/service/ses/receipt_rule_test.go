package ses_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfses "github.com/hashicorp/terraform-provider-aws/internal/service/ses"
)

func TestAccSESReceiptRule_basic(t *testing.T) {
	var rule ses.ReceiptRule

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ses_receipt_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheck(t)
			testAccPreCheckReceiptRule(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, ses.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckReceiptRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccReceiptRuleConfig_basic(rName, acctest.DefaultEmailAddress),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReceiptRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "rule_set_name", rName),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "ses", fmt.Sprintf("receipt-rule-set/%s:receipt-rule/%s", rName, rName)),
					resource.TestCheckResourceAttr(resourceName, "add_header_action.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "bounce_action.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "lambda_action.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "s3_action.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "sns_action.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "stop_action.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "workmail_action.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "recipients.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "recipients.*", acctest.DefaultEmailAddress),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "scan_enabled", "true"),
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
	var rule ses.ReceiptRule

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ses_receipt_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheck(t)
			testAccPreCheckReceiptRule(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, ses.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckReceiptRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccReceiptRuleConfig_s3Action(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReceiptRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "s3_action.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "s3_action.*.bucket_name", "aws_s3_bucket.test", "id"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "s3_action.*", map[string]string{
						"position": "1",
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
	var rule ses.ReceiptRule

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ses_receipt_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheck(t)
			testAccPreCheckReceiptRule(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, ses.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckReceiptRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccReceiptRuleConfig_snsAction(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReceiptRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "sns_action.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "sns_action.*.topic_arn", "aws_sns_topic.test", "arn"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "sns_action.*", map[string]string{
						"encoding": "UTF-8",
						"position": "1",
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
	var rule ses.ReceiptRule

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ses_receipt_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheck(t)
			testAccPreCheckReceiptRule(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, ses.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckReceiptRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccReceiptRuleConfig_snsActionEncoding(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReceiptRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "sns_action.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "sns_action.*.topic_arn", "aws_sns_topic.test", "arn"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "sns_action.*", map[string]string{
						"encoding": "Base64",
						"position": "1",
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
	var rule ses.ReceiptRule

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ses_receipt_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheck(t)
			testAccPreCheckReceiptRule(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, ses.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckReceiptRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccReceiptRuleConfig_lambdaAction(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReceiptRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "lambda_action.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "lambda_action.*.function_arn", "aws_lambda_function.test", "arn"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "lambda_action.*", map[string]string{
						"invocation_type": "Event",
						"position":        "1",
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
	var rule ses.ReceiptRule

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ses_receipt_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheck(t)
			testAccPreCheckReceiptRule(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, ses.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckReceiptRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccReceiptRuleConfig_stopAction(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReceiptRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "stop_action.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "stop_action.*", map[string]string{
						"scope": "RuleSet",
					}),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "stop_action.*.topic_arn", "aws_sns_topic.test", "arn"),
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
	var rule ses.ReceiptRule

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ses_receipt_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheck(t)
			testAccPreCheckReceiptRule(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, ses.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckReceiptRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccReceiptRuleConfig_order(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReceiptRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "name", "second"),
					resource.TestCheckResourceAttrPair(resourceName, "after", "aws_ses_receipt_rule.test1", "name"),
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
	var rule ses.ReceiptRule

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ses_receipt_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheck(t)
			testAccPreCheckReceiptRule(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, ses.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckReceiptRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccReceiptRuleConfig_actions(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReceiptRuleExists(resourceName, &rule),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "add_header_action.*", map[string]string{
						"header_name":  "Added-By",
						"header_value": "Terraform",
						"position":     "2",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "add_header_action.*", map[string]string{
						"header_name":  "Another-Header",
						"header_value": "First",
						"position":     "1",
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
	var rule ses.ReceiptRule

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ses_receipt_rule.test"

	ruleSetResourceName := "aws_ses_receipt_rule_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheck(t)
			testAccPreCheckReceiptRule(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, ses.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckReceiptRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccReceiptRuleConfig_basic(rName, acctest.DefaultEmailAddress),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReceiptRuleExists(resourceName, &rule),
					acctest.CheckResourceDisappears(acctest.Provider, tfses.ResourceReceiptRuleSet(), ruleSetResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccReceiptRuleConfig_basic(rName, acctest.DefaultEmailAddress),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReceiptRuleExists(resourceName, &rule),
					acctest.CheckResourceDisappears(acctest.Provider, tfses.ResourceReceiptRule(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckReceiptRuleDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).SESConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ses_receipt_rule" {
			continue
		}

		params := &ses.DescribeReceiptRuleInput{
			RuleName:    aws.String(rs.Primary.Attributes["name"]),
			RuleSetName: aws.String(rs.Primary.Attributes["rule_set_name"]),
		}

		_, err := conn.DescribeReceiptRule(params)
		if err == nil {
			return fmt.Errorf("Receipt rule %s still exists. Failing!", rs.Primary.ID)
		}

		// Verify the error is what we want
		_, ok := err.(awserr.Error)
		if !ok {
			return err
		}

	}

	return nil

}

func testAccCheckReceiptRuleExists(n string, rule *ses.ReceiptRule) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("SES Receipt Rule not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("SES Receipt Rule name not set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SESConn

		params := &ses.DescribeReceiptRuleInput{
			RuleName:    aws.String(rs.Primary.Attributes["name"]),
			RuleSetName: aws.String(rs.Primary.Attributes["rule_set_name"]),
		}

		resp, err := conn.DescribeReceiptRule(params)
		if err != nil {
			return err
		}

		*rule = *resp.Rule

		return nil
	}
}

func testAccReceiptRuleImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not Found: %s", resourceName)
		}

		return fmt.Sprintf("%s:%s", rs.Primary.Attributes["rule_set_name"], rs.Primary.Attributes["name"]), nil
	}
}

func testAccPreCheckReceiptRule(t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).SESConn

	input := &ses.DescribeReceiptRuleInput{
		RuleName:    aws.String("MyRule"),
		RuleSetName: aws.String("MyRuleSet"),
	}

	_, err := conn.DescribeReceiptRule(input)

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

resource "aws_s3_bucket_acl" "test" {
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
  runtime          = "nodejs12.x"
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
