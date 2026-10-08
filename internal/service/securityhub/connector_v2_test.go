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

func testAccConnectorV2_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var connector securityhub.GetConnectorV2Output
	resourceName := "aws_securityhub_connector_v2.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityHubServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConnectorV2Destroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccConnectorV2Config_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectorV2Exists(ctx, t, resourceName, &connector),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrARN), tfknownvalue.RegionalARNRegexp("securityhub", regexache.MustCompile(`connectorv2/.+`))),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("connector_id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("health"), knownvalue.ListSizeExact(1)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(rName)),
				},
			},
			{
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "connector_id"),
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "connector_id",
			},
		},
	})
}

func testAccConnectorV2_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var connector securityhub.GetConnectorV2Output
	resourceName := "aws_securityhub_connector_v2.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityHubServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConnectorV2Destroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccConnectorV2Config_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectorV2Exists(ctx, t, resourceName, &connector),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfsecurityhub.ResourceConnectorV2, resourceName),
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

func testAccConnectorV2_description(t *testing.T) {
	ctx := acctest.Context(t)
	var connector securityhub.GetConnectorV2Output
	resourceName := "aws_securityhub_connector_v2.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityHubServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConnectorV2Destroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccConnectorV2Config_description(rName, "initial description"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectorV2Exists(ctx, t, resourceName, &connector),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.StringExact("initial description")),
				},
			},
			{
				Config: testAccConnectorV2Config_description(rName, "updated description"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectorV2Exists(ctx, t, resourceName, &connector),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.StringExact("updated description")),
				},
			},
		},
	})
}

func testAccConnectorV2_kmsKeyARN(t *testing.T) {
	ctx := acctest.Context(t)
	var connector securityhub.GetConnectorV2Output
	resourceName := "aws_securityhub_connector_v2.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityHubServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConnectorV2Destroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccConnectorV2Config_kmsKeyARN(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectorV2Exists(ctx, t, resourceName, &connector),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "connector_id"),
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "connector_id",
			},
		},
	})
}

func testAccConnectorV2_connectorProviderUpdate(t *testing.T) {
	ctx := acctest.Context(t)
	var connector securityhub.GetConnectorV2Output
	resourceName := "aws_securityhub_connector_v2.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityHubServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConnectorV2Destroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccConnectorV2Config_jiraProjectKey(rName, "TEST1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectorV2Exists(ctx, t, resourceName, &connector),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				Config: testAccConnectorV2Config_jiraProjectKey(rName, "TEST2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectorV2Exists(ctx, t, resourceName, &connector),
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

func testAccConnectorV2_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var connector securityhub.GetConnectorV2Output
	resourceName := "aws_securityhub_connector_v2.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityHubServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConnectorV2Destroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccConnectorV2Config_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectorV2Exists(ctx, t, resourceName, &connector),
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
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "connector_id"),
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "connector_id",
			},
			{
				Config: testAccConnectorV2Config_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectorV2Exists(ctx, t, resourceName, &connector),
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
				Config: testAccConnectorV2Config_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectorV2Exists(ctx, t, resourceName, &connector),
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

func testAccCheckConnectorV2Exists(ctx context.Context, t *testing.T, n string, v *securityhub.GetConnectorV2Output) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).SecurityHubClient(ctx)

		output, err := tfsecurityhub.FindConnectorV2ByID(ctx, conn, rs.Primary.Attributes["connector_id"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckConnectorV2Destroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).SecurityHubClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_securityhub_connector_v2" {
				continue
			}

			_, err := tfsecurityhub.FindConnectorV2ByID(ctx, conn, rs.Primary.Attributes["connector_id"])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Security Hub V2 Connector %s still exists", rs.Primary.Attributes["connector_id"])
		}

		return nil
	}
}

func testAccConnectorV2Config_base() string {
	return fmt.Sprintf(`
resource "aws_securityhub_account_v2" "test" {}

resource "aws_securityhub_aggregator_v2" "test" {
  region_linking_mode = "SPECIFIED_REGIONS"
  linked_regions      = [%[1]q]

  depends_on = [aws_securityhub_account_v2.test]
}
`, endpoints.EuWest1RegionID)
}

func testAccConnectorV2Config_basic(rName string) string {
	return acctest.ConfigCompose(testAccConnectorV2Config_base(), fmt.Sprintf(`
resource "aws_securityhub_connector_v2" "test" {
  name = %[1]q

  connector_provider {
    jira_cloud {
      project_key = "TEST"
    }
  }

  depends_on = [aws_securityhub_aggregator_v2.test]
}
`, rName))
}

func testAccConnectorV2Config_description(rName, description string) string {
	return acctest.ConfigCompose(testAccConnectorV2Config_base(), fmt.Sprintf(`
resource "aws_securityhub_connector_v2" "test" {
  name        = %[1]q
  description = %[2]q

  connector_provider {
    jira_cloud {
      project_key = "TEST"
    }
  }

  depends_on = [aws_securityhub_aggregator_v2.test]
}
`, rName, description))
}

func testAccConnectorV2Config_kmsKeyARN(rName string) string {
	return acctest.ConfigCompose(testAccConnectorV2Config_base(), fmt.Sprintf(`
resource "aws_securityhub_connector_v2" "test" {
  name        = %[1]q
  kms_key_arn = aws_kms_key.test.arn

  connector_provider {
    jira_cloud {
      project_key = "TEST"
    }
  }

  depends_on = [aws_securityhub_aggregator_v2.test]
}

resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
}
`, rName))
}

func testAccConnectorV2Config_jiraProjectKey(rName, projectKey string) string {
	return acctest.ConfigCompose(testAccConnectorV2Config_base(), fmt.Sprintf(`
resource "aws_securityhub_connector_v2" "test" {
  name = %[1]q

  connector_provider {
    jira_cloud {
      project_key = %[2]q
    }
  }

  depends_on = [aws_securityhub_aggregator_v2.test]
}
`, rName, projectKey))
}

func testAccConnectorV2Config_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccConnectorV2Config_base(), fmt.Sprintf(`
resource "aws_securityhub_connector_v2" "test" {
  name = %[1]q

  connector_provider {
    jira_cloud {
      project_key = "TEST"
    }
  }

  tags = {
    %[2]q = %[3]q
  }

  depends_on = [aws_securityhub_aggregator_v2.test]
}
`, rName, tagKey1, tagValue1))
}

func testAccConnectorV2Config_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccConnectorV2Config_base(), fmt.Sprintf(`
resource "aws_securityhub_connector_v2" "test" {
  name = %[1]q

  connector_provider {
    jira_cloud {
      project_key = "TEST"
    }
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }

  depends_on = [aws_securityhub_aggregator_v2.test]
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}
