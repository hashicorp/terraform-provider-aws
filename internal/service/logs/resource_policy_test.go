// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package logs_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	tfstatecheck "github.com/hashicorp/terraform-provider-aws/internal/acctest/statecheck"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tflogs "github.com/hashicorp/terraform-provider-aws/internal/service/logs"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccLogsResourcePolicy_basic_accountScope(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_log_resource_policy.test"
	var resourcePolicy awstypes.ResourcePolicy

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResourcePolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccResourcePolicyConfig_basic1(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourcePolicyExists(ctx, t, resourceName, &resourcePolicy),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrID, resourceName, "policy_name"),
					resource.TestCheckResourceAttr(resourceName, "policy_name", rName),
					resource.TestCheckResourceAttr(resourceName, "policy_document", fmt.Sprintf("{\"Statement\":[{\"Action\":[\"logs:PutLogEvents\",\"logs:CreateLogStream\"],\"Effect\":\"Allow\",\"Principal\":{\"Service\":\"rds.%s\"},\"Resource\":\"arn:%s:logs:*:*:log-group:/aws/rds/*\"}],\"Version\":\"2012-10-17\"}", acctest.PartitionDNSSuffix(), acctest.Partition())),
					resource.TestCheckResourceAttr(resourceName, "policy_scope", string(awstypes.PolicyScopeAccount)),
					resource.TestCheckResourceAttr(resourceName, names.AttrResourceARN, ""),
					resource.TestCheckResourceAttr(resourceName, "revision_id", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccResourcePolicyConfig_basic2(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourcePolicyExists(ctx, t, resourceName, &resourcePolicy),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrID, resourceName, "policy_name"),
					resource.TestCheckResourceAttr(resourceName, "policy_name", rName),
					resource.TestCheckResourceAttr(resourceName, "policy_document", fmt.Sprintf("{\"Statement\":[{\"Action\":[\"logs:PutLogEvents\",\"logs:CreateLogStream\"],\"Effect\":\"Allow\",\"Principal\":{\"Service\":\"rds.%s\"},\"Resource\":\"arn:%s:logs:*:*:log-group:/aws/rds/example.com\"}],\"Version\":\"2012-10-17\"}", acctest.PartitionDNSSuffix(), acctest.Partition())),
					resource.TestCheckResourceAttr(resourceName, "policy_scope", string(awstypes.PolicyScopeAccount)),
					resource.TestCheckResourceAttr(resourceName, names.AttrResourceARN, ""),
					resource.TestCheckResourceAttr(resourceName, "revision_id", ""),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
		},
	})
}

func TestAccLogsResourcePolicy_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_log_resource_policy.test"
	var resourcePolicy awstypes.ResourcePolicy

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResourcePolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccResourcePolicyConfig_basic1(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourcePolicyExists(ctx, t, resourceName, &resourcePolicy),
					acctest.CheckSDKResourceDisappears(ctx, t, tflogs.ResourceResourcePolicy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func TestAccLogsResourcePolicy_ignoreEquivalent(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_log_resource_policy.test"
	var resourcePolicy awstypes.ResourcePolicy

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResourcePolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccResourcePolicyConfig_order(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourcePolicyExists(ctx, t, resourceName, &resourcePolicy),
					resource.TestCheckResourceAttr(resourceName, "policy_name", rName),
					resource.TestCheckResourceAttr(resourceName, "policy_document", fmt.Sprintf("{\"Statement\":[{\"Action\":[\"logs:CreateLogStream\",\"logs:PutLogEvents\"],\"Effect\":\"Allow\",\"Principal\":{\"Service\":[\"rds.%s\"]},\"Resource\":[\"arn:%s:logs:*:*:log-group:/aws/rds/example.com\"]}],\"Version\":\"2012-10-17\"}", acctest.PartitionDNSSuffix(), acctest.Partition())),
				),
			},
			{
				Config: testAccResourcePolicyConfig_newOrder(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourcePolicyExists(ctx, t, resourceName, &resourcePolicy),
					resource.TestCheckResourceAttr(resourceName, "policy_name", rName),
					resource.TestCheckResourceAttr(resourceName, "policy_document", fmt.Sprintf("{\"Statement\":[{\"Action\":[\"logs:CreateLogStream\",\"logs:PutLogEvents\"],\"Effect\":\"Allow\",\"Principal\":{\"Service\":[\"rds.%s\"]},\"Resource\":[\"arn:%s:logs:*:*:log-group:/aws/rds/example.com\"]}],\"Version\":\"2012-10-17\"}", acctest.PartitionDNSSuffix(), acctest.Partition())),
				),
			},
		},
	})
}

func TestAccLogsResourcePolicy_basic_resourceScope(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_log_resource_policy.test"
	var resourcePolicy awstypes.ResourcePolicy

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResourcePolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccResourcePolicyConfig_resourceARN(rName, "test1", "test1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourcePolicyExists(ctx, t, resourceName, &resourcePolicy),
					resource.TestCheckResourceAttrSet(resourceName, "policy_document"),
					resource.TestCheckNoResourceAttr(resourceName, "policy_name"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrResourceARN, "aws_cloudwatch_log_group.test1", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "policy_scope", string(awstypes.PolicyScopeResource)),
					resource.TestCheckResourceAttr(resourceName, "revision_id", "1"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				// Update the resource policy
				Config: testAccResourcePolicyConfig_resourceARN(rName, "test1", "test2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourcePolicyExists(ctx, t, resourceName, &resourcePolicy),
					resource.TestCheckResourceAttrSet(resourceName, "policy_document"),
					resource.TestCheckNoResourceAttr(resourceName, "policy_name"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrResourceARN, "aws_cloudwatch_log_group.test1", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "policy_scope", string(awstypes.PolicyScopeResource)),
					resource.TestCheckResourceAttr(resourceName, "revision_id", "2"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
			{
				// Update the resource ARN to a different log group, which requires resource replacement
				Config: testAccResourcePolicyConfig_resourceARN(rName, "test2", "test2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourcePolicyExists(ctx, t, resourceName, &resourcePolicy),
					resource.TestCheckResourceAttrSet(resourceName, "policy_document"),
					resource.TestCheckNoResourceAttr(resourceName, "policy_name"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrResourceARN, "aws_cloudwatch_log_group.test2", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "policy_scope", string(awstypes.PolicyScopeResource)),
					resource.TestCheckResourceAttr(resourceName, "revision_id", "1"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionDestroyBeforeCreate),
					},
				},
			},
		},
	})
}

func TestAccLogsResourcePolicy_Identity_Resourcescope_basic(t *testing.T) {
	ctx := acctest.Context(t)

	var v awstypes.ResourcePolicy
	resourceName := "aws_cloudwatch_log_resource_policy.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_12_0),
		},
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		CheckDestroy:             testAccCheckResourcePolicyDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ConfigDirectory: config.StaticDirectory("testdata/ResourcePolicy/resourcescope_basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourcePolicyExists(ctx, t, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.Region())),
					statecheck.ExpectIdentity(resourceName, map[string]knownvalue.Check{
						names.AttrAccountID:   tfknownvalue.AccountID(),
						names.AttrRegion:      knownvalue.StringExact(acctest.Region()),
						"policy_name":         knownvalue.Null(),
						names.AttrResourceARN: knownvalue.NotNull(),
					}),
					statecheck.ExpectIdentityValueMatchesState(resourceName, tfjsonpath.New(names.AttrResourceARN)),
				},
			},

			// Step 2: Import command
			{
				ConfigDirectory: config.StaticDirectory("testdata/ResourcePolicy/resourcescope_basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				ImportStateKind:   resource.ImportCommandWithID,
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},

			// Step 3: Import block with Import ID
			{
				ConfigDirectory: config.StaticDirectory("testdata/ResourcePolicy/resourcescope_basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				ResourceName:    resourceName,
				ImportState:     true,
				ImportStateKind: resource.ImportBlockWithID,
				ImportPlanChecks: resource.ImportPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.Region())),
					},
				},
			},

			// Step 4: Import block with Resource Identity
			{
				ConfigDirectory: config.StaticDirectory("testdata/ResourcePolicy/resourcescope_basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				ResourceName:    resourceName,
				ImportState:     true,
				ImportStateKind: resource.ImportBlockWithResourceIdentity,
				ImportPlanChecks: resource.ImportPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.Region())),
					},
				},
			},
		},
	})
}

func TestAccLogsResourcePolicy_Identity_Resourcescope_regionOverride(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName := "aws_cloudwatch_log_resource_policy.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_12_0),
		},
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		CheckDestroy:             acctest.CheckDestroyNoop,
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ConfigDirectory: config.StaticDirectory("testdata/ResourcePolicy/resourcescope_region_override/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					"region":        config.StringVariable(acctest.AlternateRegion()),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.AlternateRegion())),
					statecheck.ExpectIdentity(resourceName, map[string]knownvalue.Check{
						names.AttrAccountID:   tfknownvalue.AccountID(),
						names.AttrRegion:      knownvalue.StringExact(acctest.AlternateRegion()),
						"policy_name":         knownvalue.Null(),
						names.AttrResourceARN: knownvalue.NotNull(),
					}),
					statecheck.ExpectIdentityValueMatchesState(resourceName, tfjsonpath.New(names.AttrResourceARN)),
				},
			},

			// Step 2: Import command
			{
				ConfigDirectory: config.StaticDirectory("testdata/ResourcePolicy/resourcescope_region_override/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					"region":        config.StringVariable(acctest.AlternateRegion()),
				},
				ImportStateKind:   resource.ImportCommandWithID,
				ImportStateIdFunc: acctest.CrossRegionImportStateIdFunc(resourceName),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},

			// Step 3: Import block with Import ID
			{
				ConfigDirectory: config.StaticDirectory("testdata/ResourcePolicy/resourcescope_region_override/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					"region":        config.StringVariable(acctest.AlternateRegion()),
				},
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateKind:   resource.ImportBlockWithID,
				ImportStateIdFunc: acctest.CrossRegionImportStateIdFunc(resourceName),
				ImportPlanChecks: resource.ImportPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.AlternateRegion())),
					},
				},
			},

			// Step 4: Import block with Resource Identity
			{
				ConfigDirectory: config.StaticDirectory("testdata/ResourcePolicy/resourcescope_region_override/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					"region":        config.StringVariable(acctest.AlternateRegion()),
				},
				ResourceName:    resourceName,
				ImportState:     true,
				ImportStateKind: resource.ImportBlockWithResourceIdentity,
				ImportPlanChecks: resource.ImportPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.AlternateRegion())),
					},
				},
			},
		},
	})
}

// Resource Identity was added after v6.51.0
func TestAccLogsResourcePolicy_Identity_Resourcescope_ExistingResource_basic(t *testing.T) {
	ctx := acctest.Context(t)

	var v awstypes.ResourcePolicy
	resourceName := "aws_cloudwatch_log_resource_policy.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_12_0),
		},
		PreCheck:     func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:   acctest.ErrorCheck(t, names.LogsServiceID),
		CheckDestroy: testAccCheckResourcePolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			// Step 1: Create pre-Identity
			{
				ConfigDirectory: config.StaticDirectory("testdata/ResourcePolicy/resourcescope_v6.51.0/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourcePolicyExists(ctx, t, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					tfstatecheck.ExpectNoIdentity(resourceName),
				},
			},

			// Step 2: Current version
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				ConfigDirectory:          config.StaticDirectory("testdata/ResourcePolicy/resourcescope_basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectIdentity(resourceName, map[string]knownvalue.Check{
						names.AttrAccountID:   tfknownvalue.AccountID(),
						names.AttrRegion:      knownvalue.StringExact(acctest.Region()),
						"policy_name":         knownvalue.Null(),
						names.AttrResourceARN: knownvalue.NotNull(),
					}),
					statecheck.ExpectIdentityValueMatchesState(resourceName, tfjsonpath.New(names.AttrResourceARN)),
				},
			},
		},
	})
}

// Resource Identity was added after v6.51.0
func TestAccLogsResourcePolicy_Identity_Resourcescope_ExistingResource_noRefreshNoChange(t *testing.T) {
	ctx := acctest.Context(t)

	var v awstypes.ResourcePolicy
	resourceName := "aws_cloudwatch_log_resource_policy.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_12_0),
		},
		PreCheck:     func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:   acctest.ErrorCheck(t, names.LogsServiceID),
		CheckDestroy: testAccCheckResourcePolicyDestroy(ctx, t),
		AdditionalCLIOptions: &resource.AdditionalCLIOptions{
			Plan: resource.PlanOptions{
				NoRefresh: true,
			},
		},
		Steps: []resource.TestStep{
			// Step 1: Create pre-Identity
			{
				ConfigDirectory: config.StaticDirectory("testdata/ResourcePolicy/resourcescope_v6.51.0/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourcePolicyExists(ctx, t, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					tfstatecheck.ExpectNoIdentity(resourceName),
				},
			},

			// Step 2: Current version
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				ConfigDirectory:          config.StaticDirectory("testdata/ResourcePolicy/resourcescope_basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					tfstatecheck.ExpectNoIdentity(resourceName),
				},
			},
		},
	})
}

func testAccCheckResourcePolicyExists(ctx context.Context, t *testing.T, n string, v *awstypes.ResourcePolicy) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).LogsClient(ctx)

		var output *awstypes.ResourcePolicy
		var err error
		if arn.IsARN(rs.Primary.ID) {
			output, err = tflogs.FindResourcePolicyByResourceARN(ctx, conn, rs.Primary.ID)
		} else {
			output, err = tflogs.FindResourcePolicyByName(ctx, conn, rs.Primary.ID)
		}

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckResourcePolicyDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).LogsClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_cloudwatch_log_resource_policy" {
				continue
			}

			_, err := tflogs.FindResourcePolicyByName(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("CloudWatch Logs Resource Policy still exists: %s", rs.Primary.ID)
		}

		return nil
	}
}

func testAccResourcePolicyConfig_basic1(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_iam_policy_document" "test" {
  statement {
    actions = [
      "logs:CreateLogStream",
      "logs:PutLogEvents",
    ]

    resources = ["arn:${data.aws_partition.current.partition}:logs:*:*:log-group:/aws/rds/*"]

    principals {
      identifiers = ["rds.${data.aws_partition.current.dns_suffix}"]
      type        = "Service"
    }
  }
}

resource "aws_cloudwatch_log_resource_policy" "test" {
  policy_name     = %[1]q
  policy_document = data.aws_iam_policy_document.test.json
}
`, rName)
}

func testAccResourcePolicyConfig_basic2(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_iam_policy_document" "test" {
  statement {
    actions = [
      "logs:CreateLogStream",
      "logs:PutLogEvents",
    ]

    resources = ["arn:${data.aws_partition.current.partition}:logs:*:*:log-group:/aws/rds/example.com"]

    principals {
      identifiers = ["rds.${data.aws_partition.current.dns_suffix}"]
      type        = "Service"
    }
  }
}

resource "aws_cloudwatch_log_resource_policy" "test" {
  policy_name     = %[1]q
  policy_document = data.aws_iam_policy_document.test.json
}
`, rName)
}

func testAccResourcePolicyConfig_order(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_cloudwatch_log_resource_policy" "test" {
  policy_name = %[1]q
  policy_document = jsonencode({
    Statement = [{
      Action = [
        "logs:CreateLogStream",
        "logs:PutLogEvents",
      ]
      Effect = "Allow"
      Resource = [
        "arn:${data.aws_partition.current.partition}:logs:*:*:log-group:/aws/rds/example.com",
      ]
      Principal = {
        Service = [
          "rds.${data.aws_partition.current.dns_suffix}",
        ]
      }
    }]
    Version = "2012-10-17"
  })
}
`, rName)
}

func testAccResourcePolicyConfig_newOrder(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_cloudwatch_log_resource_policy" "test" {
  policy_name = %[1]q
  policy_document = jsonencode({
    Statement = [{
      Action = [
        "logs:PutLogEvents",
        "logs:CreateLogStream",
      ]
      Effect = "Allow"
      Resource = [
        "arn:${data.aws_partition.current.partition}:logs:*:*:log-group:/aws/rds/example.com",
      ]
      Principal = {
        Service = [
          "rds.${data.aws_partition.current.dns_suffix}",
        ]
      }
    }]
    Version = "2012-10-17"
  })
}
`, rName)
}

func testAccResourcePolicyConfig_resourceARN(rName, logGroupIdentifier, policyIdentifier string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_cloudwatch_log_group" "test1" {
  name = "/aws/rds/%[1]s-1"
}

resource "aws_cloudwatch_log_group" "test2" {
  name = "/aws/rds/%[1]s-2"
}

data "aws_iam_policy_document" "test1" {
  statement {
    actions = [
      "logs:CreateLogStream",
      "logs:PutLogEvents",
    ]

    resources = [
      aws_cloudwatch_log_group.%[2]s.arn,
    ]

    principals {
      identifiers = ["rds.${data.aws_partition.current.dns_suffix}"]
      type        = "Service"
    }
  }
}

data "aws_iam_policy_document" "test2" {
  statement {
    actions = [
      "logs:CreateLogStream",
      "logs:PutLogEvents",
      "logs:DescribeLogStreams",
    ]

    resources = [
      aws_cloudwatch_log_group.%[2]s.arn,
    ]

    principals {
      identifiers = ["rds.${data.aws_partition.current.dns_suffix}"]
      type        = "Service"
    }
  }
}

resource "aws_cloudwatch_log_resource_policy" "test" {
  policy_document = data.aws_iam_policy_document.%[3]s.json
  resource_arn    = aws_cloudwatch_log_group.%[2]s.arn
}
`, rName, logGroupIdentifier, policyIdentifier)
}
