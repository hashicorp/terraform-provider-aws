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
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfresiliencehubv2 "github.com/hashicorp/terraform-provider-aws/internal/service/resiliencehubv2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccResilienceHubV2ServiceFunction_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var sf awstypes.ServiceFunction
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_resiliencehubv2_service_function.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ResilienceHubV2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceFunctionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/ServiceFunction/basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceFunctionExists(ctx, t, resourceName, &sf),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("criticality"), tfknownvalue.StringExact(awstypes.ServiceFunctionCriticalityPrimary)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(rName)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("service_function_id"), knownvalue.NotNull()),
				},
			},
		},
	})
}

func TestAccResilienceHubV2ServiceFunction_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var sf awstypes.ServiceFunction
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_resiliencehubv2_service_function.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ResilienceHubV2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceFunctionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/ServiceFunction/basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceFunctionExists(ctx, t, resourceName, &sf),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfresiliencehubv2.ResourceServiceFunction, resourceName),
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

func TestAccResilienceHubV2ServiceFunction_description(t *testing.T) {
	ctx := acctest.Context(t)
	var sf awstypes.ServiceFunction
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_resiliencehubv2_service_function.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ResilienceHubV2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceFunctionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/ServiceFunction/description/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					"description":   config.StringVariable("desc1"),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceFunctionExists(ctx, t, resourceName, &sf),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.StringExact("desc1")),
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/ServiceFunction/description/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					"description":   config.StringVariable("desc2"),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceFunctionExists(ctx, t, resourceName, &sf),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.StringExact("desc2")),
				},
			},
		},
	})
}

func TestAccResilienceHubV2ServiceFunction_criticality(t *testing.T) {
	ctx := acctest.Context(t)
	var sf awstypes.ServiceFunction
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_resiliencehubv2_service_function.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ResilienceHubV2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceFunctionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/ServiceFunction/criticality/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					"criticality":   config.StringVariable(string(awstypes.ServiceFunctionCriticalitySupplemental)),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceFunctionExists(ctx, t, resourceName, &sf),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("criticality"), tfknownvalue.StringExact(awstypes.ServiceFunctionCriticalitySupplemental)),
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/ServiceFunction/criticality/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					"criticality":   config.StringVariable(string(awstypes.ServiceFunctionCriticalityPrimary)),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceFunctionExists(ctx, t, resourceName, &sf),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("criticality"), tfknownvalue.StringExact(awstypes.ServiceFunctionCriticalityPrimary)),
				},
			},
		},
	})
}

func TestAccResilienceHubV2ServiceFunction_name(t *testing.T) {
	ctx := acctest.Context(t)
	var sf awstypes.ServiceFunction
	rName1 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rName2 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_resiliencehubv2_service_function.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ResilienceHubV2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceFunctionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/ServiceFunction/name/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName1),
					acctest.CtName:  config.StringVariable(rName1),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceFunctionExists(ctx, t, resourceName, &sf),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(rName1)),
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/ServiceFunction/name/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName1),
					acctest.CtName:  config.StringVariable(rName2),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceFunctionExists(ctx, t, resourceName, &sf),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(rName2)),
				},
			},
		},
	})
}

func testAccCheckServiceFunctionDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).ResilienceHubV2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_resiliencehubv2_service_function" {
				continue
			}

			_, err := tfresiliencehubv2.FindServiceFunctionByTwoPartKey(ctx, conn, rs.Primary.Attributes["service_arn"], rs.Primary.Attributes["service_function_id"])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Resilience Hub V2 Service Function %s still exists", rs.Primary.Attributes["service_function_id"])
		}

		return nil
	}
}

func testAccCheckServiceFunctionExists(ctx context.Context, t *testing.T, n string, v *awstypes.ServiceFunction) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Service Function not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).ResilienceHubV2Client(ctx)

		output, err := tfresiliencehubv2.FindServiceFunctionByTwoPartKey(ctx, conn, rs.Primary.Attributes["service_arn"], rs.Primary.Attributes["service_function_id"])
		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}
