// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package cloudwatchomni_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudwatchomni/types"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfcloudwatchomni "github.com/hashicorp/terraform-provider-aws/internal/service/cloudwatchomni"
	"github.com/hashicorp/terraform-provider-aws/names"
)

var (
	checkSpaceARN                = tfknownvalue.RegionalARNRegexp("cloudwatch", regexache.MustCompile(`space/[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`))
	checkSpaceARNAlternateRegion = tfknownvalue.RegionalARNAlternateRegionRegexp("cloudwatch", regexache.MustCompile(`space/[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`))
	checkSpaceID                 = knownvalue.StringRegexp(regexache.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`))
)

func testAccSpace_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var space awstypes.Space
	resourceName := "aws_cloudwatchomni_space.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudWatchOmniServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSpaceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/Space/basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSpaceExists(ctx, t, resourceName, &space),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("agent_core_evaluation_role_arn"), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("data_access_role_arn"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("domain_arn"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("domain_id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrEncryptionConfiguration), knownvalue.ListSizeExact(0)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(rName)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("space_arn"), checkSpaceARN),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("space_id"), checkSpaceID),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrStatus), tfknownvalue.StringExact(awstypes.SpaceStatusActive)),
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/Space/basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "space_id"),
				ImportStateVerifyIdentifierAttribute: "space_id",
			},
		},
	})
}

func testAccSpace_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var space awstypes.Space
	resourceName := "aws_cloudwatchomni_space.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudWatchOmniServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSpaceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/Space/basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSpaceExists(ctx, t, resourceName, &space),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfcloudwatchomni.ResourceSpace, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccSpace_agentCore(t *testing.T) {
	ctx := acctest.Context(t)
	var space awstypes.Space
	resourceName := "aws_cloudwatchomni_space.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudWatchOmniServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSpaceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/Space/agent_core_evaluation_role_arn/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSpaceExists(ctx, t, resourceName, &space),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("agent_core_evaluation_role_arn"), knownvalue.NotNull()),
				},
			},
		},
	})
}

func testAccSpace_name(t *testing.T) {
	ctx := acctest.Context(t)
	var space awstypes.Space
	resourceName := "aws_cloudwatchomni_space.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	name1, name2 := rName, acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudWatchOmniServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSpaceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/Space/name/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					acctest.CtName:  config.StringVariable(name1),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSpaceExists(ctx, t, resourceName, &space),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(name1)),
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/Space/name/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					acctest.CtName:  config.StringVariable(name2),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSpaceExists(ctx, t, resourceName, &space),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(name2)),
				},
			},
		},
	})
}

func testAccSpace_encryptionCreateAWS_OWNED(t *testing.T) { // nosemgrep:ci.aws-in-func-name
	ctx := acctest.Context(t)
	var space awstypes.Space
	resourceName := "aws_cloudwatchomni_space.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudWatchOmniServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSpaceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/Space/encryption_strategy.AWS_OWNED/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSpaceExists(ctx, t, resourceName, &space),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrEncryptionConfiguration), knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectExact(map[string]knownvalue.Check{
						"encryption_strategy": tfknownvalue.StringExact(awstypes.EncryptionStrategyAwsOwned),
						names.AttrKMSKeyARN:   knownvalue.Null(),
					})})),
				},
			},
			// Because of the encryption_strategy normalization, on import encryption_strategy will be empty.
			// {
			// 	ConfigDirectory: config.StaticDirectory("testdata/Space/encryption_strategy.AWS_OWNED/"),
			// 	ConfigVariables: config.Variables{
			// 		acctest.CtRName: config.StringVariable(rName),
			// 	},
			// 	ResourceName:                         resourceName,
			// 	ImportState:                          true,
			// 	ImportStateVerify:                    true,
			// 	ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "space_id"),
			// 	ImportStateVerifyIdentifierAttribute: "space_id",
			// },
			{
				ConfigDirectory: config.StaticDirectory("testdata/Space/encryption_strategy.CUSTOMER_MANAGED/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSpaceExists(ctx, t, resourceName, &space),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrEncryptionConfiguration), knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectExact(map[string]knownvalue.Check{
						"encryption_strategy": tfknownvalue.StringExact(awstypes.EncryptionStrategyCustomerManaged),
						names.AttrKMSKeyARN:   knownvalue.NotNull(),
					})})),
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/Space/basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSpaceExists(ctx, t, resourceName, &space),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrEncryptionConfiguration), knownvalue.ListSizeExact(0)),
				},
			},
		},
	})
}

func testAccSpace_encryptionCreateCUSTOMER_MANAGED(t *testing.T) {
	ctx := acctest.Context(t)
	var space awstypes.Space
	resourceName := "aws_cloudwatchomni_space.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudWatchOmniServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSpaceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/Space/encryption_strategy.CUSTOMER_MANAGED/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSpaceExists(ctx, t, resourceName, &space),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrEncryptionConfiguration), knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectExact(map[string]knownvalue.Check{
						"encryption_strategy": tfknownvalue.StringExact(awstypes.EncryptionStrategyCustomerManaged),
						names.AttrKMSKeyARN:   knownvalue.NotNull(),
					})})),
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/Space/encryption_strategy.CUSTOMER_MANAGED/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "space_id"),
				ImportStateVerifyIdentifierAttribute: "space_id",
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/Space/encryption_strategy.AWS_OWNED/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSpaceExists(ctx, t, resourceName, &space),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrEncryptionConfiguration), knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectExact(map[string]knownvalue.Check{
						"encryption_strategy": tfknownvalue.StringExact(awstypes.EncryptionStrategyAwsOwned),
						names.AttrKMSKeyARN:   knownvalue.Null(),
					})})),
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/Space/basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSpaceExists(ctx, t, resourceName, &space),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrEncryptionConfiguration), knownvalue.ListSizeExact(0)),
				},
			},
		},
	})
}

func testAccCheckSpaceDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).CloudWatchOmniClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_cloudwatchomni_space" {
				continue
			}

			_, err := tfcloudwatchomni.FindSpaceByID(ctx, conn, rs.Primary.Attributes["space_id"])
			if retry.NotFound(err) {
				continue
			}
			if err != nil {
				return err
			}

			return fmt.Errorf("CloudWatch Omni Space %s still exists", rs.Primary.Attributes["space_id"])
		}

		return nil
	}
}

func testAccCheckSpaceExists(ctx context.Context, t *testing.T, n string, v *awstypes.Space) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).CloudWatchOmniClient(ctx)

		output, err := tfcloudwatchomni.FindSpaceByID(ctx, conn, rs.Primary.Attributes["space_id"])
		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}
