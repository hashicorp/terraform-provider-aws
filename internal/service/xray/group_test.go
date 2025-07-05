// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package xray_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/xray/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	tfstatecheck "github.com/hashicorp/terraform-provider-aws/internal/acctest/statecheck"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfxray "github.com/hashicorp/terraform-provider-aws/internal/service/xray"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccXRayGroup_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.Group
	resourceName := "aws_xray_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.XRayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_basic(rName, "responsetime > 5"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "xray", regexache.MustCompile(`group/.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrGroupName, rName),
					resource.TestCheckResourceAttr(resourceName, "filter_expression", "responsetime > 5"),
					resource.TestCheckResourceAttr(resourceName, "insights_configuration.#", "1"), // Computed.
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccGroupConfig_basic(rName, "responsetime > 10"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "xray", regexache.MustCompile(`group/.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrGroupName, rName),
					resource.TestCheckResourceAttr(resourceName, "filter_expression", "responsetime > 10"),
					resource.TestCheckResourceAttr(resourceName, "insights_configuration.#", "1"),
				),
			},
		},
	})
}

func TestAccXRayGroup_insights(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.Group
	resourceName := "aws_xray_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.XRayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_basicInsights(rName, "responsetime > 5", true, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "insights_configuration.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "insights_configuration.*", map[string]string{
						"insights_enabled":      acctest.CtTrue,
						"notifications_enabled": acctest.CtTrue,
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccGroupConfig_basicInsights(rName, "responsetime > 10", false, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "insights_configuration.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "insights_configuration.*", map[string]string{
						"insights_enabled":      acctest.CtFalse,
						"notifications_enabled": acctest.CtFalse,
					}),
				),
			},
		},
	})
}

func TestAccXRayGroup_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.Group
	resourceName := "aws_xray_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.XRayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_basic(rName, "responsetime > 5"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfxray.ResourceGroup(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccXRayGroup_Identity_ExistingResource(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.Group
	resourceName := "aws_xray_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_12_0),
		},
		PreCheck:     func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:   acctest.ErrorCheck(t, names.XRayServiceID),
		CheckDestroy: testAccCheckGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {
						Source:            "hashicorp/aws",
						VersionConstraint: "5.100.0",
					},
				},
				Config: testAccGroupConfig_basic(rName, "responsetime > 5"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					tfstatecheck.ExpectNoIdentity(resourceName),
				},
			},
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {
						Source:            "hashicorp/aws",
						VersionConstraint: "6.0.0",
					},
				},
				Config: testAccGroupConfig_basic(rName, "responsetime > 5"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &v),
				),
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
						names.AttrARN: knownvalue.Null(),
					}),
				},
			},
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				Config:                   testAccGroupConfig_basic(rName, "responsetime > 5"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &v),
				),
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
						names.AttrARN: tfknownvalue.RegionalARNRegexp("xray", regexache.MustCompile(`group/.+`)),
					}),
				},
			},
		},
	})
}

func testAccCheckGroupExists(ctx context.Context, n string, v *types.Group) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).XRayClient(ctx)

		output, err := tfxray.FindGroupByARN(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckGroupDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_xray_group" {
				continue
			}

			conn := acctest.Provider.Meta().(*conns.AWSClient).XRayClient(ctx)

			_, err := tfxray.FindGroupByARN(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("XRay Group %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccGroupConfig_basic(rName, expression string) string {
	return fmt.Sprintf(`
resource "aws_xray_group" "test" {
  group_name        = %[1]q
  filter_expression = %[2]q
}
`, rName, expression)
}

func testAccGroupConfig_basicInsights(rName, expression string, insightsEnabled bool, notificationsEnabled bool) string {
	return fmt.Sprintf(`
resource "aws_xray_group" "test" {
  group_name        = %[1]q
  filter_expression = %[2]q

  insights_configuration {
    insights_enabled      = %[3]t
    notifications_enabled = %[4]t
  }
}
`, rName, expression, insightsEnabled, notificationsEnabled)
}
