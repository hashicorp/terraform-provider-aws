// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package mailmanager_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/mailmanager"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfmailmanager "github.com/hashicorp/terraform-provider-aws/internal/service/mailmanager"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccMailManagerRuleSet_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_mailmanager_rule_set.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccRuleSetPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.MailManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleSetDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleSetConfigBasic(rName, "X-Example"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRuleSetExists(ctx, t, resourceName, nil),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "ses", regexache.MustCompile(`.+`)),
					acctest.CheckResourceAttrRFC3339(resourceName, names.AttrCreatedDate),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrID),
					acctest.CheckResourceAttrRFC3339(resourceName, "last_modification_date"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrRegion, acctest.Region()),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.action.0.add_header.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.action.0.add_header.0.header_name", "X-Example"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.action.0.add_header.0.header_value", "example"),
				),
			},
			{
				Config: testAccRuleSetConfigBasic(rName, "X-Updated"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRuleSetExists(ctx, t, resourceName, nil),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "rule.0.action.0.add_header.0.header_name", "X-Updated"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.action.0.add_header.0.header_value", "example"),
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

func TestAccMailManagerRuleSet_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_mailmanager_rule_set.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccRuleSetPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.MailManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleSetDestroy(ctx, t),
		Steps: []resource.TestStep{{
			Config: testAccRuleSetConfigBasic(rName, "X-Example"),
			Check: resource.ComposeAggregateTestCheckFunc(
				testAccCheckRuleSetExists(ctx, t, resourceName, nil),
				acctest.CheckFrameworkResourceDisappears(ctx, t, tfmailmanager.ResourceRuleSet, resourceName),
			),
			ExpectNonEmptyPlan: true,
			ConfigPlanChecks: resource.ConfigPlanChecks{
				PostApplyPostRefresh: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
				},
			},
		}},
	})
}

func TestAccMailManagerRuleSet_conditionTypes(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_mailmanager_rule_set.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccRuleSetPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.MailManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleSetDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleSetConfigConditionTypes(rName),
				Check:  testAccCheckRuleSetExists(ctx, t, resourceName, nil),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRule).AtSliceIndex(0).AtMapKey(names.AttrCondition), knownvalue.ListExact([]knownvalue.Check{
						testAccRuleSetUnionValue("boolean_expression", map[string]knownvalue.Check{
							"evaluate": testAccRuleSetEvaluateValue(map[string]knownvalue.Check{"attribute": knownvalue.StringExact("TLS")}),
							"operator": knownvalue.StringExact("IS_TRUE"),
						}),
						testAccRuleSetUnionValue("dmarc_expression", map[string]knownvalue.Check{
							"operator":       knownvalue.StringExact("NOT_EQUALS"),
							names.AttrValues: knownvalue.ListExact([]knownvalue.Check{knownvalue.StringExact("REJECT")}),
						}),
						testAccRuleSetUnionValue("ip_expression", map[string]knownvalue.Check{
							"evaluate":       testAccRuleSetEvaluateValue(map[string]knownvalue.Check{"attribute": knownvalue.StringExact("SOURCE_IP")}),
							"operator":       knownvalue.StringExact("CIDR_MATCHES"),
							names.AttrValues: knownvalue.ListExact([]knownvalue.Check{knownvalue.StringExact("192.0.2.0/24")}),
						}),
						testAccRuleSetUnionValue("number_expression", map[string]knownvalue.Check{
							"evaluate":      testAccRuleSetEvaluateValue(map[string]knownvalue.Check{"attribute": knownvalue.StringExact("MESSAGE_SIZE")}),
							"operator":      knownvalue.StringExact("GREATER_THAN"),
							names.AttrValue: knownvalue.Float64Exact(1024),
						}),
						testAccRuleSetUnionValue("string_expression", map[string]knownvalue.Check{
							"evaluate":       testAccRuleSetEvaluateValue(map[string]knownvalue.Check{"mime_header_attribute": knownvalue.StringExact("X-Example")}),
							"operator":       knownvalue.StringExact("CONTAINS"),
							names.AttrValues: knownvalue.ListExact([]knownvalue.Check{knownvalue.StringExact("example")}),
						}),
						testAccRuleSetUnionValue("verdict_expression", map[string]knownvalue.Check{
							"evaluate":       testAccRuleSetEvaluateValue(map[string]knownvalue.Check{"attribute": knownvalue.StringExact("DKIM")}),
							"operator":       knownvalue.StringExact("EQUALS"),
							names.AttrValues: knownvalue.ListExact([]knownvalue.Check{knownvalue.StringExact("PASS")}),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRule).AtSliceIndex(0).AtMapKey("unless"), knownvalue.ListExact([]knownvalue.Check{
						testAccRuleSetUnionValue("string_expression", map[string]knownvalue.Check{
							"evaluate":       testAccRuleSetEvaluateValue(map[string]knownvalue.Check{"attribute": knownvalue.StringExact("MAIL_FROM")}),
							"operator":       knownvalue.StringExact("ENDS_WITH"),
							names.AttrValues: knownvalue.ListExact([]knownvalue.Check{knownvalue.StringExact("@example.com")}),
						}),
					})),
				},
			},
		},
	})
}

func TestAccMailManagerRuleSet_actionSimpleTypes(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_mailmanager_rule_set.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccRuleSetPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.MailManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleSetDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleSetConfigActionSimpleTypes(rName),
				Check:  testAccCheckRuleSetExists(ctx, t, resourceName, nil),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRule).AtSliceIndex(0), knownvalue.ObjectPartial(map[string]knownvalue.Check{
						names.AttrName: knownvalue.StringExact("primary"),
						names.AttrAction: knownvalue.ListExact([]knownvalue.Check{
							testAccRuleSetUnionValue("add_header", map[string]knownvalue.Check{
								"header_name": knownvalue.StringExact("X-Example"), "header_value": knownvalue.StringExact("example"),
							}),
							testAccRuleSetUnionValue("drop", map[string]knownvalue.Check{}),
							testAccRuleSetUnionValue("replace_recipient", map[string]knownvalue.Check{
								"replace_with": knownvalue.ListExact([]knownvalue.Check{knownvalue.StringExact(acctest.DefaultEmailAddress)}),
							}),
						}),
					})),
				},
			},
		},
	})
}

func TestAccMailManagerRuleSet_actionInvokeLambda(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_mailmanager_rule_set.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccRuleSetPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.MailManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleSetDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleSetConfigActionInvokeLambda(rName),
				Check:  testAccCheckRuleSetExists(ctx, t, resourceName, nil),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRule).AtSliceIndex(0).AtMapKey(names.AttrAction), knownvalue.ListExact([]knownvalue.Check{
						testAccRuleSetUnionValue("invoke_lambda", map[string]knownvalue.Check{
							names.AttrFunctionARN:   knownvalue.NotNull(),
							"invocation_type":       knownvalue.StringExact("EVENT"),
							names.AttrRoleARN:       knownvalue.NotNull(),
							"action_failure_policy": knownvalue.StringExact("CONTINUE"),
							"retry_time_minutes":    knownvalue.Null(),
						}),
					})),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccMailManagerRuleSet_actionPublishToSNS(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_mailmanager_rule_set.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccRuleSetPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.MailManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleSetDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleSetConfigActionPublishToSNS(rName),
				Check:  testAccCheckRuleSetExists(ctx, t, resourceName, nil),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRule).AtSliceIndex(0).AtMapKey(names.AttrAction), knownvalue.ListExact([]knownvalue.Check{
						testAccRuleSetUnionValue("publish_to_sns", map[string]knownvalue.Check{
							names.AttrTopicARN:      knownvalue.NotNull(),
							names.AttrRoleARN:       knownvalue.NotNull(),
							"encoding":              knownvalue.StringExact("UTF-8"),
							"payload_type":          knownvalue.StringExact("HEADERS"),
							"action_failure_policy": knownvalue.StringExact("CONTINUE"),
						}),
					})),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccMailManagerRuleSet_actionWriteToS3(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_mailmanager_rule_set.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccRuleSetPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.MailManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleSetDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleSetConfigActionWriteToS3(rName),
				Check:  testAccCheckRuleSetExists(ctx, t, resourceName, nil),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRule).AtSliceIndex(0).AtMapKey(names.AttrAction), knownvalue.ListExact([]knownvalue.Check{
						testAccRuleSetUnionValue("write_to_s3", map[string]knownvalue.Check{
							names.AttrS3Bucket:      knownvalue.NotNull(),
							names.AttrRoleARN:       knownvalue.NotNull(),
							"s3_prefix":             knownvalue.StringExact("mail/"),
							"s3_sse_kms_key_id":     knownvalue.Null(),
							"action_failure_policy": knownvalue.StringExact("CONTINUE"),
						}),
					})),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccMailManagerRuleSet_evaluateTypes(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_mailmanager_rule_set.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccRuleSetPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.MailManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleSetDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleSetConfigEvaluateTypes(rName),
				Check:  testAccCheckRuleSetExists(ctx, t, resourceName, nil),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRule).AtSliceIndex(0).AtMapKey(names.AttrCondition), knownvalue.ListExact([]knownvalue.Check{
						testAccRuleSetUnionValue("string_expression", map[string]knownvalue.Check{
							"evaluate":       testAccRuleSetEvaluateValue(map[string]knownvalue.Check{"client_certificate_attribute": knownvalue.StringExact("CN")}),
							"operator":       knownvalue.StringExact("STARTS_WITH"),
							names.AttrValues: knownvalue.ListExact([]knownvalue.Check{knownvalue.StringExact("example")}),
						}),
					})),
				},
			},
		},
	})
}

func testAccCheckRuleSetExists(ctx context.Context, t *testing.T, name string, output *mailmanager.GetRuleSetOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.MailManager, create.ErrActionCheckingExistence, tfmailmanager.ResNameRuleSet, name, errors.New("not found"))
		}
		if rs.Primary.ID == "" {
			return create.Error(names.MailManager, create.ErrActionCheckingExistence, tfmailmanager.ResNameRuleSet, name, errors.New("not set"))
		}
		conn := acctest.ProviderMeta(ctx, t).MailManagerClient(ctx)
		out, err := tfmailmanager.FindRuleSetByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return create.Error(names.MailManager, create.ErrActionCheckingExistence, tfmailmanager.ResNameRuleSet, rs.Primary.ID, err)
		}
		if output != nil {
			*output = *out
		}
		return nil
	}
}

func testAccCheckRuleSetDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).MailManagerClient(ctx)
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_mailmanager_rule_set" {
				continue
			}
			_, err := tfmailmanager.FindRuleSetByID(ctx, conn, rs.Primary.ID)
			if retry.NotFound(err) {
				continue
			}
			if err != nil {
				return create.Error(names.MailManager, create.ErrActionCheckingDestroyed, tfmailmanager.ResNameRuleSet, rs.Primary.ID, err)
			}
			return create.Error(names.MailManager, create.ErrActionCheckingDestroyed, tfmailmanager.ResNameRuleSet, rs.Primary.ID, errors.New("not destroyed"))
		}
		return nil
	}
}

func testAccRuleSetPreCheck(ctx context.Context, t *testing.T) {
	var input mailmanager.ListRuleSetsInput
	_, err := acctest.ProviderMeta(ctx, t).MailManagerClient(ctx).ListRuleSets(ctx, &input)
	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccRuleSetUnionValue(name string, attributes map[string]knownvalue.Check) knownvalue.Check {
	return knownvalue.ObjectPartial(map[string]knownvalue.Check{
		name: knownvalue.ListExact([]knownvalue.Check{
			knownvalue.ObjectExact(attributes),
		}),
	})
}

func testAccRuleSetEvaluateValue(attributes map[string]knownvalue.Check) knownvalue.Check {
	return knownvalue.ListExact([]knownvalue.Check{
		knownvalue.ObjectPartial(attributes),
	})
}

func testAccRuleSetConfigBasic(rName, headerName string) string {
	return fmt.Sprintf(`
resource "aws_mailmanager_rule_set" "test" {
  name = %[1]q

  rule {
    action {
      add_header {
        header_name  = %[2]q
        header_value = "example"
      }
    }
  }
}
`, rName, headerName)
}

func testAccRuleSetConfigConditionTypes(rName string) string {
	return fmt.Sprintf(`
resource "aws_mailmanager_rule_set" "test" {
  name = %[1]q

  rule {
    condition {
      boolean_expression {
        operator = "IS_TRUE"

        evaluate {
          attribute = "TLS"
        }
      }
    }

    condition {
      dmarc_expression {
        operator = "NOT_EQUALS"
        values   = ["REJECT"]
      }
    }

    condition {
      ip_expression {
        operator = "CIDR_MATCHES"
        values   = ["192.0.2.0/24"]

        evaluate {
          attribute = "SOURCE_IP"
        }
      }
    }

    condition {
      number_expression {
        operator = "GREATER_THAN"
        value    = 1024

        evaluate {
          attribute = "MESSAGE_SIZE"
        }
      }
    }

    condition {
      string_expression {
        operator = "CONTAINS"
        values   = ["example"]

        evaluate {
          mime_header_attribute = "X-Example"
        }
      }
    }

    condition {
      verdict_expression {
        operator = "EQUALS"
        values   = ["PASS"]

        evaluate {
          attribute = "DKIM"
        }
      }
    }

    unless {
      string_expression {
        operator = "ENDS_WITH"
        values   = ["@example.com"]

        evaluate {
          attribute = "MAIL_FROM"
        }
      }
    }

    action {
      add_header {
        header_name  = "X-Example"
        header_value = "example"
      }
    }
  }
}
`, rName)
}

func testAccRuleSetConfigActionSimpleTypes(rName string) string {
	return fmt.Sprintf(`
resource "aws_mailmanager_rule_set" "test" {
  name = %[1]q

  rule {
    name = "primary"

    action {
      add_header {
        header_name  = "X-Example"
        header_value = "example"
      }
    }

    action {
      drop {}
    }

    action {
      replace_recipient {
        replace_with = [%[2]q]
      }
    }
  }
}
`, rName, acctest.DefaultEmailAddress)
}

func testAccRuleSetConfigEvaluateTypes(rName string) string {
	return fmt.Sprintf(`
resource "aws_mailmanager_rule_set" "test" {
  name = %[1]q

  rule {
    condition {
      string_expression {
        operator = "STARTS_WITH"
        values   = ["example"]

        evaluate {
          client_certificate_attribute = "CN"
        }
      }
    }

    action {
      add_header {
        header_name  = "X-Example"
        header_value = "example"
      }
    }
  }
}
`, rName)
}

func testAccRuleSetConfigSESRoleBase(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect    = "Allow"
      Principal = { Service = "ses.amazonaws.com" }
      Action    = "sts:AssumeRole"
    }]
  })
}
`, rName)
}

func testAccRuleSetConfigActionInvokeLambda(rName string) string {
	return acctest.ConfigCompose(
		testAccRuleSetConfigSESRoleBase(rName),
		fmt.Sprintf(`
resource "aws_iam_role_policy" "test" {
  name = %[1]q
  role = aws_iam_role.test.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect   = "Allow"
      Action   = "lambda:InvokeFunction"
      Resource = "*"
    }]
  })
}

resource "aws_iam_role" "lambda" {
  name = "%[1]s-lambda"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect    = "Allow"
      Principal = { Service = "lambda.amazonaws.com" }
      Action    = "sts:AssumeRole"
    }]
  })
}

resource "aws_lambda_function" "test" {
  filename         = "test-fixtures/lambdatest.zip"
  source_code_hash = filebase64sha256("test-fixtures/lambdatest.zip")
  function_name    = %[1]q
  role             = aws_iam_role.lambda.arn
  handler          = "exports.example"
  runtime          = "nodejs22.x"
}

resource "aws_mailmanager_rule_set" "test" {
  name = %[1]q

  rule {
    action {
      invoke_lambda {
        function_arn          = aws_lambda_function.test.arn
        invocation_type       = "EVENT"
        role_arn              = aws_iam_role.test.arn
        action_failure_policy = "CONTINUE"
      }
    }
  }
}
`, rName))
}

func testAccRuleSetConfigActionPublishToSNS(rName string) string {
	return acctest.ConfigCompose(
		testAccRuleSetConfigSESRoleBase(rName),
		fmt.Sprintf(`
resource "aws_iam_role_policy" "test" {
  name = %[1]q
  role = aws_iam_role.test.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect   = "Allow"
      Action   = "sns:Publish"
      Resource = aws_sns_topic.test.arn
    }]
  })
}

resource "aws_sns_topic" "test" {
  name = %[1]q
}

resource "aws_mailmanager_rule_set" "test" {
  name = %[1]q

  rule {
    action {
      publish_to_sns {
        topic_arn             = aws_sns_topic.test.arn
        role_arn              = aws_iam_role.test.arn
        encoding              = "UTF-8"
        payload_type          = "HEADERS"
        action_failure_policy = "CONTINUE"
      }
    }
  }
}
`, rName))
}

func testAccRuleSetConfigActionWriteToS3(rName string) string {
	return acctest.ConfigCompose(
		testAccRuleSetConfigSESRoleBase(rName),
		fmt.Sprintf(`
resource "aws_iam_role_policy" "test" {
  name = %[1]q
  role = aws_iam_role.test.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect   = "Allow"
      Action   = ["s3:PutObject", "s3:GetBucketLocation"]
      Resource = [aws_s3_bucket.test.arn, "${aws_s3_bucket.test.arn}/*"]
    }]
  })
}

resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_mailmanager_rule_set" "test" {
  name = %[1]q

  rule {
    action {
      write_to_s3 {
        s3_bucket             = aws_s3_bucket.test.id
        role_arn              = aws_iam_role.test.arn
        s3_prefix             = "mail/"
        action_failure_policy = "CONTINUE"
      }
    }
  }
}
`, rName))
}
