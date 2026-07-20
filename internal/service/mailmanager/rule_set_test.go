// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package mailmanager_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/mailmanager"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
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
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
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
					acctest.CheckResourceAttrRFC3339(resourceName, "created_date"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrID),
					acctest.CheckResourceAttrRFC3339(resourceName, "last_modification_date"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrRegion, acctest.Region()),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
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
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
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
		}},
	})
}

func TestAccMailManagerRuleSet_conditionTypes(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
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
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("rule").AtSliceIndex(0).AtMapKey("condition"), knownvalue.ListExact([]knownvalue.Check{
						testAccRuleSetUnionValue("boolean_expression", map[string]knownvalue.Check{
							"evaluate": testAccRuleSetEvaluateValue(map[string]knownvalue.Check{"attribute": knownvalue.StringExact("TLS")}),
							"operator": knownvalue.StringExact("IS_TRUE"),
						}),
						testAccRuleSetUnionValue("dmarc_expression", map[string]knownvalue.Check{
							"operator": knownvalue.StringExact("NOT_EQUALS"),
							"values":   knownvalue.ListExact([]knownvalue.Check{knownvalue.StringExact("REJECT")}),
						}),
						testAccRuleSetUnionValue("ip_expression", map[string]knownvalue.Check{
							"evaluate": testAccRuleSetEvaluateValue(map[string]knownvalue.Check{"attribute": knownvalue.StringExact("SOURCE_IP")}),
							"operator": knownvalue.StringExact("CIDR_MATCHES"),
							"values":   knownvalue.ListExact([]knownvalue.Check{knownvalue.StringExact("192.0.2.0/24")}),
						}),
						testAccRuleSetUnionValue("number_expression", map[string]knownvalue.Check{
							"evaluate": testAccRuleSetEvaluateValue(map[string]knownvalue.Check{"attribute": knownvalue.StringExact("MESSAGE_SIZE")}),
							"operator": knownvalue.StringExact("GREATER_THAN"),
							"value":    knownvalue.Float64Exact(1024),
						}),
						testAccRuleSetUnionValue("string_expression", map[string]knownvalue.Check{
							"evaluate": testAccRuleSetEvaluateValue(map[string]knownvalue.Check{"mime_header_attribute": knownvalue.StringExact("X-Example")}),
							"operator": knownvalue.StringExact("CONTAINS"),
							"values":   knownvalue.ListExact([]knownvalue.Check{knownvalue.StringExact("example")}),
						}),
						testAccRuleSetUnionValue("verdict_expression", map[string]knownvalue.Check{
							"evaluate": testAccRuleSetEvaluateValue(map[string]knownvalue.Check{"attribute": knownvalue.StringExact("DKIM")}),
							"operator": knownvalue.StringExact("EQUALS"),
							"values":   knownvalue.ListExact([]knownvalue.Check{knownvalue.StringExact("PASS")}),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("rule").AtSliceIndex(0).AtMapKey("unless"), knownvalue.ListExact([]knownvalue.Check{
						testAccRuleSetUnionValue("string_expression", map[string]knownvalue.Check{
							"evaluate": testAccRuleSetEvaluateValue(map[string]knownvalue.Check{"attribute": knownvalue.StringExact("MAIL_FROM")}),
							"operator": knownvalue.StringExact("ENDS_WITH"),
							"values":   knownvalue.ListExact([]knownvalue.Check{knownvalue.StringExact("@example.com")}),
						}),
					})),
				},
			},
		},
	})
}

func TestAccMailManagerRuleSet_actionTypes(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
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
				Config: testAccRuleSetConfigActionTypes(rName),
				Check:  testAccCheckRuleSetExists(ctx, t, resourceName, nil),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("rule").AtSliceIndex(0), knownvalue.ObjectPartial(map[string]knownvalue.Check{
						names.AttrName: knownvalue.StringExact("primary"),
						"action": knownvalue.ListExact([]knownvalue.Check{
							testAccRuleSetUnionValue("add_header", map[string]knownvalue.Check{
								"header_name": knownvalue.StringExact("X-Example"), "header_value": knownvalue.StringExact("example"),
							}),
							testAccRuleSetUnionValue("archive", map[string]knownvalue.Check{
								"action_failure_policy": knownvalue.StringExact("CONTINUE"), "target_archive": knownvalue.StringExact("example-archive"),
							}),
							testAccRuleSetUnionValue("bounce", map[string]knownvalue.Check{
								"action_failure_policy": knownvalue.StringExact("DROP"), "diagnostic_message": knownvalue.StringExact("Message rejected"), "message": knownvalue.StringExact("Rejected by rule set"), "role_arn": knownvalue.NotNull(), "sender": knownvalue.StringExact("mailer-daemon@example.com"), "smtp_reply_code": knownvalue.StringExact("550"), "status_code": knownvalue.StringExact("5.1.1"),
							}),
							testAccRuleSetUnionValue("deliver_to_mailbox", map[string]knownvalue.Check{
								"action_failure_policy": knownvalue.StringExact("CONTINUE"), "mailbox_arn": knownvalue.NotNull(), "role_arn": knownvalue.NotNull(),
							}),
							testAccRuleSetUnionValue("deliver_to_q_business", map[string]knownvalue.Check{
								"action_failure_policy": knownvalue.StringExact("DROP"), "application_id": knownvalue.StringExact("00000000-0000-0000-0000-000000000001"), "index_id": knownvalue.StringExact("00000000-0000-0000-0000-000000000002"), "role_arn": knownvalue.NotNull(),
							}),
							testAccRuleSetUnionValue("drop", map[string]knownvalue.Check{}),
							testAccRuleSetUnionValue("invoke_lambda", map[string]knownvalue.Check{
								"action_failure_policy": knownvalue.StringExact("CONTINUE"), "function_arn": knownvalue.NotNull(), "invocation_type": knownvalue.StringExact("EVENT"), "retry_time_minutes": knownvalue.Int64Exact(60), "role_arn": knownvalue.NotNull(),
							}),
							testAccRuleSetUnionValue("publish_to_sns", map[string]knownvalue.Check{
								"action_failure_policy": knownvalue.StringExact("DROP"), "encoding": knownvalue.StringExact("BASE64"), "payload_type": knownvalue.StringExact("HEADERS"), "role_arn": knownvalue.NotNull(), "topic_arn": knownvalue.NotNull(),
							}),
							testAccRuleSetUnionValue("replace_recipient", map[string]knownvalue.Check{
								"replace_with": knownvalue.ListExact([]knownvalue.Check{knownvalue.StringExact("one@example.com"), knownvalue.StringExact("two@example.com")}),
							}),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("rule").AtSliceIndex(1), knownvalue.ObjectPartial(map[string]knownvalue.Check{
						names.AttrName: knownvalue.StringExact("secondary"),
						"action": knownvalue.ListExact([]knownvalue.Check{
							testAccRuleSetUnionValue("send", map[string]knownvalue.Check{
								"action_failure_policy": knownvalue.StringExact("DROP"), "role_arn": knownvalue.NotNull(),
							}),
							testAccRuleSetUnionValue("write_to_s3", map[string]knownvalue.Check{
								"action_failure_policy": knownvalue.StringExact("CONTINUE"), "role_arn": knownvalue.NotNull(), "s3_bucket": knownvalue.StringExact("example-bucket"), "s3_prefix": knownvalue.StringExact("mail"), "s3_sse_kms_key_id": knownvalue.NotNull(),
							}),
						}),
					})),
				},
			},
		},
	})
}

func TestAccMailManagerRuleSet_evaluateTypes(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
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
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("rule").AtSliceIndex(0).AtMapKey("condition"), knownvalue.ListExact([]knownvalue.Check{
						testAccRuleSetUnionValue("boolean_expression", map[string]knownvalue.Check{
							"evaluate": testAccRuleSetEvaluateValue(map[string]knownvalue.Check{
								"is_in_address_list": knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectExact(map[string]knownvalue.Check{
									"address_lists": knownvalue.ListExact([]knownvalue.Check{knownvalue.StringExact("example-list")}), "attribute": knownvalue.StringExact("MAIL_FROM"),
								})}),
							}),
							"operator": knownvalue.StringExact("IS_FALSE"),
						}),
						testAccRuleSetUnionValue("string_expression", map[string]knownvalue.Check{
							"evaluate": testAccRuleSetEvaluateValue(map[string]knownvalue.Check{"client_certificate_attribute": knownvalue.StringExact("CN")}),
							"operator": knownvalue.StringExact("STARTS_WITH"),
							"values":   knownvalue.ListExact([]knownvalue.Check{knownvalue.StringExact("example")}),
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
		out, err := tfmailmanager.FindRuleSetByID(ctx, acctest.ProviderMeta(ctx, t).MailManagerClient(ctx), rs.Primary.ID)
		if err == nil && output != nil {
			*output = *out
		}
		return create.Error(names.MailManager, create.ErrActionCheckingExistence, tfmailmanager.ResNameRuleSet, rs.Primary.ID, err)
	}
}

func testAccCheckRuleSetDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_mailmanager_rule_set" {
				continue
			}
			_, err := tfmailmanager.FindRuleSetByID(ctx, acctest.ProviderMeta(ctx, t).MailManagerClient(ctx), rs.Primary.ID)
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
	_, err := acctest.ProviderMeta(ctx, t).MailManagerClient(ctx).ListRuleSets(ctx, &mailmanager.ListRuleSetsInput{})
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

func testAccRuleSetConfigActionTypes(rName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}
data "aws_partition" "current" {}
data "aws_region" "current" {}

resource "aws_iam_role" "test" {
	name = %[1]q

	assume_role_policy = jsonencode({
		Version = "2012-10-17"
		Statement = [{
			Action = "sts:AssumeRole"
			Effect = "Allow"
			Principal = {
				Service = "ses.amazonaws.com"
			}
		}]
	})
}

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
			archive {
				action_failure_policy = "CONTINUE"
				target_archive        = "example-archive"
			}
		}

		action {
			bounce {
				action_failure_policy = "DROP"
				diagnostic_message    = "Message rejected"
				message               = "Rejected by rule set"
				role_arn              = aws_iam_role.test.arn
				sender                = "mailer-daemon@example.com"
				smtp_reply_code       = "550"
				status_code           = "5.1.1"
			}
		}

		action {
			deliver_to_mailbox {
				action_failure_policy = "CONTINUE"
				mailbox_arn           = "arn:${data.aws_partition.current.partition}:workmail:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:organization/m-00000000000000000000000000000000"
				role_arn              = aws_iam_role.test.arn
			}
		}

		action {
			deliver_to_q_business {
				action_failure_policy = "DROP"
				application_id        = "00000000-0000-0000-0000-000000000001"
				index_id              = "00000000-0000-0000-0000-000000000002"
				role_arn              = aws_iam_role.test.arn
			}
		}

		action {
			drop {}
		}

		action {
			invoke_lambda {
				action_failure_policy = "CONTINUE"
				function_arn          = "arn:${data.aws_partition.current.partition}:lambda:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:function:example"
				invocation_type       = "EVENT"
				retry_time_minutes    = 60
				role_arn              = aws_iam_role.test.arn
			}
		}

		action {
			publish_to_sns {
				action_failure_policy = "DROP"
				encoding              = "BASE64"
				payload_type          = "HEADERS"
				role_arn              = aws_iam_role.test.arn
				topic_arn             = "arn:${data.aws_partition.current.partition}:sns:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:example"
			}
		}

		action {
			replace_recipient {
				replace_with = ["one@example.com", "two@example.com"]
			}
		}
	}

	rule {
		name = "secondary"

		action {
			send {
				action_failure_policy = "DROP"
				role_arn              = aws_iam_role.test.arn
			}
		}

		action {
			write_to_s3 {
				action_failure_policy = "CONTINUE"
				role_arn              = aws_iam_role.test.arn
				s3_bucket             = "example-bucket"
				s3_prefix             = "mail"
				s3_sse_kms_key_id     = "arn:${data.aws_partition.current.partition}:kms:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:key/00000000-0000-0000-0000-000000000000"
			}
		}
	}
}
`, rName)
}

func testAccRuleSetConfigEvaluateTypes(rName string) string {
	return fmt.Sprintf(`
resource "aws_mailmanager_rule_set" "test" {
	name = %[1]q

	rule {
		condition {
			boolean_expression {
				operator = "IS_FALSE"

				evaluate {
					is_in_address_list {
						address_lists = ["example-list"]
						attribute     = "MAIL_FROM"
					}
				}
			}
		}

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
