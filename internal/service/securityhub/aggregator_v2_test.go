// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package securityhub_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/securityhub"
	"github.com/hashicorp/aws-sdk-go-base/v2/endpoints"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfsecurityhub "github.com/hashicorp/terraform-provider-aws/internal/service/securityhub"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccAggregatorV2_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var aggregator securityhub.GetAggregatorV2Output
	resourceName := "aws_securityhub_aggregator_v2.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityHubServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAggregatorV2Destroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAggregatorV2Config_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAggregatorV2Exists(ctx, t, resourceName, &aggregator),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrARN), tfknownvalue.RegionalARNRegexp("securityhub", regexache.MustCompile(`aggregatorv2/.+`))),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("region_linking_mode"), knownvalue.StringExact("SPECIFIED_REGIONS")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("aggregation_region"), knownvalue.NotNull()),
				},
			},
			{
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
		},
	})
}

func testAccAggregatorV2_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var aggregator securityhub.GetAggregatorV2Output
	resourceName := "aws_securityhub_aggregator_v2.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityHubServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAggregatorV2Destroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAggregatorV2Config_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAggregatorV2Exists(ctx, t, resourceName, &aggregator),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfsecurityhub.ResourceAggregatorV2, resourceName),
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

func testAccAggregatorV2_specifiedRegions(t *testing.T) {
	ctx := acctest.Context(t)
	var aggregator securityhub.GetAggregatorV2Output
	resourceName := "aws_securityhub_aggregator_v2.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityHubServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAggregatorV2Destroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAggregatorV2Config_specifiedRegions(endpoints.UsEast1RegionID, endpoints.EuWest1RegionID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAggregatorV2Exists(ctx, t, resourceName, &aggregator),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("region_linking_mode"), knownvalue.StringExact("SPECIFIED_REGIONS")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("linked_regions"), knownvalue.SetSizeExact(2)),
				},
			},
			{
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
			{
				Config: testAccAggregatorV2Config_specifiedRegions(endpoints.UsEast1RegionID, endpoints.EuWest1RegionID, endpoints.ApSoutheast1RegionID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAggregatorV2Exists(ctx, t, resourceName, &aggregator),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("region_linking_mode"), knownvalue.StringExact("SPECIFIED_REGIONS")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("linked_regions"), knownvalue.SetSizeExact(3)),
				},
			},
		},
	})
}

func testAccAggregatorV2_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var aggregator securityhub.GetAggregatorV2Output
	resourceName := "aws_securityhub_aggregator_v2.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityHubServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAggregatorV2Destroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAggregatorV2Config_tags1(acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAggregatorV2Exists(ctx, t, resourceName, &aggregator),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1),
					})),
				},
			},
			{
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
			{
				Config: testAccAggregatorV2Config_tags2(acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAggregatorV2Exists(ctx, t, resourceName, &aggregator),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1Updated),
						acctest.CtKey2: knownvalue.StringExact(acctest.CtValue2),
					})),
				},
			},
			{
				Config: testAccAggregatorV2Config_tags1(acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAggregatorV2Exists(ctx, t, resourceName, &aggregator),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey2: knownvalue.StringExact(acctest.CtValue2),
					})),
				},
			},
		},
	})
}

func testAccCheckAggregatorV2Exists(ctx context.Context, t *testing.T, n string, v *securityhub.GetAggregatorV2Output) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).SecurityHubClient(ctx)

		output, err := tfsecurityhub.FindAggregatorV2ByARN(ctx, conn, rs.Primary.Attributes[names.AttrARN])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckAggregatorV2Destroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).SecurityHubClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_securityhub_aggregator_v2" {
				continue
			}

			_, err := tfsecurityhub.FindAggregatorV2ByARN(ctx, conn, rs.Primary.Attributes[names.AttrARN])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Security Hub V2 Aggregator %s still exists", rs.Primary.Attributes[names.AttrARN])
		}

		return nil
	}
}

func testAccAggregatorV2Config_base() string {
	return `
resource "aws_securityhub_account_v2" "test" {}
`
}

func testAccAggregatorV2Config_basic() string {
	return acctest.ConfigCompose(testAccAggregatorV2Config_base(), fmt.Sprintf(`
resource "aws_securityhub_aggregator_v2" "test" {
  region_linking_mode = "SPECIFIED_REGIONS"
  linked_regions      = [%[1]q]

  depends_on = [aws_securityhub_account_v2.test]
}
`, endpoints.UsEast1RegionID))
}

func testAccAggregatorV2Config_specifiedRegions(regions ...string) string {
	return acctest.ConfigCompose(testAccAggregatorV2Config_base(), fmt.Sprintf(`
resource "aws_securityhub_aggregator_v2" "test" {
  region_linking_mode = "SPECIFIED_REGIONS"
  linked_regions      = [%[1]s]

  depends_on = [aws_securityhub_account_v2.test]
}
`, acctest.ListOfStrings(regions...)))
}

func testAccAggregatorV2Config_tags1(tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccAggregatorV2Config_base(), fmt.Sprintf(`
resource "aws_securityhub_aggregator_v2" "test" {
  region_linking_mode = "SPECIFIED_REGIONS"
  linked_regions      = [%[3]q]

  tags = {
    %[1]q = %[2]q
  }

  depends_on = [aws_securityhub_account_v2.test]
}
`, tagKey1, tagValue1, endpoints.UsEast1RegionID))
}

func testAccAggregatorV2Config_tags2(tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccAggregatorV2Config_base(), fmt.Sprintf(`
resource "aws_securityhub_aggregator_v2" "test" {
  region_linking_mode = "SPECIFIED_REGIONS"
  linked_regions      = [%[5]q]

  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }

  depends_on = [aws_securityhub_account_v2.test]
}
`, tagKey1, tagValue1, tagKey2, tagValue2, endpoints.UsEast1RegionID))
}
