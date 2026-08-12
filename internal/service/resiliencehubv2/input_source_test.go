// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package resiliencehubv2_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/resiliencehubv2/types"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfresiliencehubv2 "github.com/hashicorp/terraform-provider-aws/internal/service/resiliencehubv2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccCheckInputSourceImportStateIDFunc(resourceName string) resource.ImportStateIdFunc {
	return acctest.AttrsImportStateIdFunc(resourceName, ",", "service_arn", "input_source_id")
}

func TestAccResilienceHubV2InputSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var is awstypes.InputSourceSummary
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_resiliencehubv2_input_source.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ResilienceHubV2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInputSourceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/InputSource/basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInputSourceExists(ctx, t, resourceName, &is),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("input_source_id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("resource_configuration"), knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectExact(map[string]knownvalue.Check{
						"cfn_stack_arn":      knownvalue.NotNull(),
						"design_file_s3_url": knownvalue.Null(),
						"eks":                knownvalue.ListSizeExact(0),
						"resource_tag":       knownvalue.ListSizeExact(0),
						"tf_state_file_url":  knownvalue.Null(),
					})})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("service_arn"), knownvalue.NotNull()),
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/InputSource/basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				ImportStateIdFunc:                    testAccCheckInputSourceImportStateIDFunc(resourceName),
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "input_source_id",
			},
		},
	})
}

func TestAccResilienceHubV2InputSource_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var is awstypes.InputSourceSummary
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_resiliencehubv2_input_source.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ResilienceHubV2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInputSourceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/InputSource/basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInputSourceExists(ctx, t, resourceName, &is),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfresiliencehubv2.ResourceInputSource, resourceName),
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

func TestAccResilienceHubV2InputSource_tfStateFileURL(t *testing.T) {
	ctx := acctest.Context(t)
	var is awstypes.InputSourceSummary
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_resiliencehubv2_input_source.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ResilienceHubV2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInputSourceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/InputSource/tf_state_file_url/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInputSourceExists(ctx, t, resourceName, &is),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("input_source_id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("resource_configuration"), knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectExact(map[string]knownvalue.Check{
						"cfn_stack_arn":      knownvalue.Null(),
						"design_file_s3_url": knownvalue.Null(),
						"eks":                knownvalue.ListSizeExact(0),
						"resource_tag":       knownvalue.ListSizeExact(0),
						"tf_state_file_url":  knownvalue.NotNull(),
					})})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("service_arn"), knownvalue.NotNull()),
				},
			},
		},
	})
}

func TestAccResilienceHubV2InputSource_designFileS3URL(t *testing.T) {
	ctx := acctest.Context(t)
	var is awstypes.InputSourceSummary
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_resiliencehubv2_input_source.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ResilienceHubV2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInputSourceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/InputSource/design_file_s3_url/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInputSourceExists(ctx, t, resourceName, &is),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("input_source_id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("resource_configuration"), knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectExact(map[string]knownvalue.Check{
						"cfn_stack_arn":      knownvalue.Null(),
						"design_file_s3_url": knownvalue.NotNull(),
						"eks":                knownvalue.ListSizeExact(0),
						"resource_tag":       knownvalue.ListSizeExact(0),
						"tf_state_file_url":  knownvalue.Null(),
					})})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("service_arn"), knownvalue.NotNull()),
				},
			},
		},
	})
}

func TestAccResilienceHubV2InputSource_resourceTags(t *testing.T) {
	ctx := acctest.Context(t)
	var is awstypes.InputSourceSummary
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_resiliencehubv2_input_source.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ResilienceHubV2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInputSourceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/InputSource/resource_tag/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInputSourceExists(ctx, t, resourceName, &is),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("input_source_id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("resource_configuration"), knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectExact(map[string]knownvalue.Check{
						"cfn_stack_arn":      knownvalue.Null(),
						"design_file_s3_url": knownvalue.Null(),
						"eks":                knownvalue.ListSizeExact(0),
						"resource_tag": knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrKey:    knownvalue.StringExact("Name"),
							names.AttrValues: knownvalue.SetExact([]knownvalue.Check{knownvalue.StringExact(rName)}),
						})}),
						"tf_state_file_url": knownvalue.Null(),
					})})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("service_arn"), knownvalue.NotNull()),
				},
			},
		},
	})
}

func testAccCheckInputSourceDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).ResilienceHubV2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_resiliencehubv2_input_source" {
				continue
			}

			_, err := tfresiliencehubv2.FindInputSourceByTwoPartKey(ctx, conn, rs.Primary.Attributes["service_arn"], rs.Primary.Attributes["input_source_id"])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Resilience Hub V2 Input Source %s still exists", rs.Primary.Attributes["input_source_id"])
		}

		return nil
	}
}

func testAccCheckInputSourceExists(ctx context.Context, t *testing.T, n string, v *awstypes.InputSourceSummary) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).ResilienceHubV2Client(ctx)

		output, err := tfresiliencehubv2.FindInputSourceByTwoPartKey(ctx, conn, rs.Primary.Attributes["service_arn"], rs.Primary.Attributes["input_source_id"])
		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}
