// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bedrockagentcore_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfbedrockagentcore "github.com/hashicorp/terraform-provider-aws/internal/service/bedrockagentcore"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccBedrockAgentCoreBrowserProfile_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var browserProfile bedrockagentcorecontrol.GetBrowserProfileOutput
	rName := strings.ReplaceAll(acctest.RandomWithPrefix(t, acctest.ResourcePrefix), "-", "_")
	resourceName := "aws_bedrockagentcore_browser_profile.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckBrowserProfiles(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBrowserProfileDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccBrowserProfileConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBrowserProfileExists(ctx, t, resourceName, &browserProfile),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("profile_arn"), tfknownvalue.RegionalARNRegexp("bedrock-agentcore", regexache.MustCompile(`browser-profile/.+`))),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("profile_id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "profile_id"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "profile_id",
			},
		},
	})
}

func TestAccBedrockAgentCoreBrowserProfile_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var browserProfile bedrockagentcorecontrol.GetBrowserProfileOutput
	rName := strings.ReplaceAll(acctest.RandomWithPrefix(t, acctest.ResourcePrefix), "-", "_")
	resourceName := "aws_bedrockagentcore_browser_profile.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckBrowserProfiles(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBrowserProfileDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccBrowserProfileConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBrowserProfileExists(ctx, t, resourceName, &browserProfile),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfbedrockagentcore.ResourceBrowserProfile, resourceName),
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

func TestAccBedrockAgentCoreBrowserProfile_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var browserProfile bedrockagentcorecontrol.GetBrowserProfileOutput
	rName := strings.ReplaceAll(acctest.RandomWithPrefix(t, acctest.ResourcePrefix), "-", "_")
	resourceName := "aws_bedrockagentcore_browser_profile.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckBrowserProfiles(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBrowserProfileDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccBrowserProfileConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBrowserProfileExists(ctx, t, resourceName, &browserProfile),
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
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "profile_id"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "profile_id",
			},
			{
				Config: testAccBrowserProfileConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBrowserProfileExists(ctx, t, resourceName, &browserProfile),
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
				Config: testAccBrowserProfileConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBrowserProfileExists(ctx, t, resourceName, &browserProfile),
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

func TestAccBedrockAgentCoreBrowserProfile_description(t *testing.T) {
	ctx := acctest.Context(t)
	var browserProfile bedrockagentcorecontrol.GetBrowserProfileOutput
	rName := strings.ReplaceAll(acctest.RandomWithPrefix(t, acctest.ResourcePrefix), "-", "_")
	resourceName := "aws_bedrockagentcore_browser_profile.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckBrowserProfiles(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBrowserProfileDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccBrowserProfileConfig_description(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBrowserProfileExists(ctx, t, resourceName, &browserProfile),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("profile_arn"), tfknownvalue.RegionalARNRegexp("bedrock-agentcore", regexache.MustCompile(`browser-profile/.+`))),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("profile_id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.StringExact("test description")),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "profile_id"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "profile_id",
			},
		},
	})
}

func testAccCheckBrowserProfileDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).BedrockAgentCoreClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_bedrockagentcore_browser_profile" {
				continue
			}

			_, err := tfbedrockagentcore.FindBrowserProfileByID(ctx, conn, rs.Primary.Attributes["profile_id"])
			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Bedrock Agent Core Browser Profile %s still exists", rs.Primary.Attributes["profile_id"])
		}

		return nil
	}
}

func testAccCheckBrowserProfileExists(ctx context.Context, t *testing.T, n string, v *bedrockagentcorecontrol.GetBrowserProfileOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).BedrockAgentCoreClient(ctx)

		resp, err := tfbedrockagentcore.FindBrowserProfileByID(ctx, conn, rs.Primary.Attributes["profile_id"])
		if err != nil {
			return err
		}

		*v = *resp

		return nil
	}
}

func testAccPreCheckBrowserProfiles(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).BedrockAgentCoreClient(ctx)

	input := bedrockagentcorecontrol.ListBrowserProfilesInput{}

	_, err := conn.ListBrowserProfiles(ctx, &input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccBrowserProfileConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_bedrockagentcore_browser_profile" "test" {
  name = %[1]q
}
`, rName)
}

func testAccBrowserProfileConfig_description(rName string) string {
	return fmt.Sprintf(`
resource "aws_bedrockagentcore_browser_profile" "test" {
  name        = %[1]q
  description = "test description"
}
`, rName)
}

func testAccBrowserProfileConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_bedrockagentcore_browser_profile" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccBrowserProfileConfig_tags2(rName, tagKey1, tagValue1, tag2Key, tag2Value string) string {
	return fmt.Sprintf(`
resource "aws_bedrockagentcore_browser_profile" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tag2Key, tag2Value)
}
