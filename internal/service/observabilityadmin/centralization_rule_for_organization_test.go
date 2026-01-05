// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package observabilityadmin_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/observabilityadmin"
	awstypes "github.com/aws/aws-sdk-go-v2/service/observabilityadmin/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/endpoints"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfobservabilityadmin "github.com/hashicorp/terraform-provider-aws/internal/service/observabilityadmin"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccObservabilityAdminCentralizationRuleForOrganization_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var rule observabilityadmin.GetCentralizationRuleForOrganizationOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_observabilityadmin_centralization_rule_for_organization.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
			// https://docs.aws.amazon.com/organizations/latest/userguide/services-that-can-integrate-cloudwatch.html.
			acctest.PreCheckIAMServiceLinkedRole(ctx, t, "/aws-service-role/observabilityadmin.amazonaws.com")
			acctest.PreCheckOrganizationsEnabledServicePrincipal(ctx, t, "observabilityadmin.amazonaws.com")
			acctest.PreCheckIAMServiceLinkedRole(ctx, t, "/aws-service-role/logs-centralization.observabilityadmin.amazonaws.com")
			acctest.PreCheckPartition(t, endpoints.AwsPartitionID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ObservabilityAdminServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCentralizationRuleForOrganizationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCentralizationRuleForOrganizationConfig_basic(rName, endpoints.EuWest1RegionID, endpoints.ApSoutheast1RegionID),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCentralizationRuleForOrganizationExists(ctx, resourceName, &rule),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("rule_arn"), tfknownvalue.RegionalARNExact("observabilityadmin", `organization-centralization-rule/`+rName)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "rule_name"),
				ImportStateVerifyIdentifierAttribute: "rule_name",
			},
		},
	})
}

func TestAccObservabilityAdminCentralizationRuleForOrganization_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var rule observabilityadmin.GetCentralizationRuleForOrganizationOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_observabilityadmin_centralization_rule_for_organization.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
			acctest.PreCheckIAMServiceLinkedRole(ctx, t, "/aws-service-role/observabilityadmin.amazonaws.com")
			acctest.PreCheckOrganizationsEnabledServicePrincipal(ctx, t, "observabilityadmin.amazonaws.com")
			acctest.PreCheckIAMServiceLinkedRole(ctx, t, "/aws-service-role/logs-centralization.observabilityadmin.amazonaws.com")
			acctest.PreCheckPartition(t, endpoints.AwsPartitionID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ObservabilityAdminServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCentralizationRuleForOrganizationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCentralizationRuleForOrganizationConfig_basic(rName, endpoints.EuWest1RegionID, endpoints.ApSoutheast1RegionID),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCentralizationRuleForOrganizationExists(ctx, resourceName, &rule),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfobservabilityadmin.ResourceCentralizationRuleForOrganization, resourceName),
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

func TestAccObservabilityAdminCentralizationRuleForOrganization_update(t *testing.T) {
	ctx := acctest.Context(t)
	var rule observabilityadmin.GetCentralizationRuleForOrganizationOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_observabilityadmin_centralization_rule_for_organization.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
			acctest.PreCheckIAMServiceLinkedRole(ctx, t, "/aws-service-role/observabilityadmin.amazonaws.com")
			acctest.PreCheckOrganizationsEnabledServicePrincipal(ctx, t, "observabilityadmin.amazonaws.com")
			acctest.PreCheckIAMServiceLinkedRole(ctx, t, "/aws-service-role/logs-centralization.observabilityadmin.amazonaws.com")
			acctest.PreCheckPartition(t, endpoints.AwsPartitionID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ObservabilityAdminServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCentralizationRuleForOrganizationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCentralizationRuleForOrganizationConfig_basic(rName, endpoints.EuWest1RegionID, endpoints.ApSoutheast1RegionID),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCentralizationRuleForOrganizationExists(ctx, resourceName, &rule),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRule), knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectPartial(map[string]knownvalue.Check{
						names.AttrDestination: knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectPartial(map[string]knownvalue.Check{
							"destination_logs_configuration": knownvalue.ListSizeExact(0),
							names.AttrRegion:                 knownvalue.StringExact(endpoints.EuWest1RegionID),
						})}),
						names.AttrSource: knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectPartial(map[string]knownvalue.Check{
							"regions": knownvalue.SetExact([]knownvalue.Check{
								knownvalue.StringExact(endpoints.ApSoutheast1RegionID),
							}),
							"source_logs_configuration": knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectPartial(map[string]knownvalue.Check{
								"log_group_selection_criteria": knownvalue.StringExact("*"),
							})}),
						})}),
					})})),
				},
			},
			{
				Config: testAccCentralizationRuleForOrganizationConfig_basic(rName, endpoints.EuWest1RegionID, endpoints.ApSoutheast1RegionID, endpoints.UsEast1RegionID),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCentralizationRuleForOrganizationExists(ctx, resourceName, &rule),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRule), knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectPartial(map[string]knownvalue.Check{
						names.AttrDestination: knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectPartial(map[string]knownvalue.Check{
							"destination_logs_configuration": knownvalue.ListSizeExact(0),
							names.AttrRegion:                 knownvalue.StringExact(endpoints.EuWest1RegionID),
						})}),
						names.AttrSource: knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectPartial(map[string]knownvalue.Check{
							"regions": knownvalue.SetExact([]knownvalue.Check{
								knownvalue.StringExact(endpoints.ApSoutheast1RegionID),
								knownvalue.StringExact(endpoints.UsEast1RegionID),
							}),
							"source_logs_configuration": knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectPartial(map[string]knownvalue.Check{
								"log_group_selection_criteria": knownvalue.StringExact("*"),
							})}),
						})}),
					})})),
				},
			},
			{
				Config: testAccCentralizationRuleForOrganizationConfig_updated(rName, endpoints.EuWest1RegionID, endpoints.UsWest1RegionID, endpoints.ApSoutheast1RegionID, endpoints.UsEast1RegionID),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCentralizationRuleForOrganizationExists(ctx, resourceName, &rule),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRule), knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectPartial(map[string]knownvalue.Check{
						names.AttrDestination: knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectPartial(map[string]knownvalue.Check{
							"destination_logs_configuration": knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectExact(map[string]knownvalue.Check{
								"backup_configuration": knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectPartial(map[string]knownvalue.Check{
									names.AttrRegion: knownvalue.StringExact(endpoints.UsWest1RegionID),
								})}),
								"logs_encryption_configuration": knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectPartial(map[string]knownvalue.Check{
									"encryption_strategy": tfknownvalue.StringExact(awstypes.EncryptionStrategyAwsOwned),
								})}),
							})}),
							names.AttrRegion: knownvalue.StringExact(endpoints.EuWest1RegionID),
						})}),
						names.AttrSource: knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectPartial(map[string]knownvalue.Check{
							"regions": knownvalue.SetExact([]knownvalue.Check{
								knownvalue.StringExact(endpoints.ApSoutheast1RegionID),
								knownvalue.StringExact(endpoints.UsEast1RegionID),
							}),
							"source_logs_configuration": knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectPartial(map[string]knownvalue.Check{
								"log_group_selection_criteria": knownvalue.StringExact("LogGroupName LIKE '/aws/lambda%'"),
							})}),
						})}),
					})})),
				},
			},
		},
	})
}

func TestAccObservabilityAdminCentralizationRuleForOrganization_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var rule observabilityadmin.GetCentralizationRuleForOrganizationOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_observabilityadmin_centralization_rule_for_organization.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
			acctest.PreCheckIAMServiceLinkedRole(ctx, t, "/aws-service-role/observabilityadmin.amazonaws.com")
			acctest.PreCheckOrganizationsEnabledServicePrincipal(ctx, t, "observabilityadmin.amazonaws.com")
			acctest.PreCheckIAMServiceLinkedRole(ctx, t, "/aws-service-role/logs-centralization.observabilityadmin.amazonaws.com")
			acctest.PreCheckPartition(t, endpoints.AwsPartitionID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ObservabilityAdminServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCentralizationRuleForOrganizationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCentralizationRuleForOrganizationConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCentralizationRuleForOrganizationExists(ctx, resourceName, &rule),
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
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "rule_name"),
				ImportStateVerifyIdentifierAttribute: "rule_name",
			},
			{
				Config: testAccCentralizationRuleForOrganizationConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCentralizationRuleForOrganizationExists(ctx, resourceName, &rule),
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
				Config: testAccCentralizationRuleForOrganizationConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCentralizationRuleForOrganizationExists(ctx, resourceName, &rule),
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

func testAccCheckCentralizationRuleForOrganizationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ObservabilityAdminClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_observabilityadmin_centralization_rule_for_organization" {
				continue
			}

			_, err := tfobservabilityadmin.FindCentralizationRuleForOrganizationByID(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Observability Admin Centralization Rule For Organization %s still exists", rs.Primary.Attributes["rule_name"])
		}

		return nil
	}
}

func testAccCheckCentralizationRuleForOrganizationExists(ctx context.Context, n string, v *observabilityadmin.GetCentralizationRuleForOrganizationOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ObservabilityAdminClient(ctx)

		output, err := tfobservabilityadmin.FindCentralizationRuleForOrganizationByID(ctx, conn, rs.Primary.Attributes["rule_name"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).ObservabilityAdminClient(ctx)

	input := observabilityadmin.ListCentralizationRulesForOrganizationInput{}
	_, err := conn.ListCentralizationRulesForOrganization(ctx, &input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccCentralizationRuleForOrganizationConfig_basic(rName, dstRegion string, srcRegions ...string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}
data "aws_organizations_organization" "current" {}

resource "aws_observabilityadmin_centralization_rule_for_organization" "test" {
  rule_name = %[1]q

  rule {
    destination {
      region  = %[2]q
      account = data.aws_caller_identity.current.account_id
    }

    source {
      regions = [%[3]s]
      scope   = "OrganizationId = '${data.aws_organizations_organization.current.id}'"

      source_logs_configuration {
        encrypted_log_group_strategy = "SKIP"
        log_group_selection_criteria = "*"
      }
    }
  }
}
`, rName, dstRegion, acctest.ListOfStrings(srcRegions...))
}

func testAccCentralizationRuleForOrganizationConfig_updated(rName, dstRegion, bkupRegion string, srcRegions ...string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}
data "aws_organizations_organization" "current" {}

resource "aws_observabilityadmin_centralization_rule_for_organization" "test" {
  rule_name = %[1]q

  rule {
    destination {
      region  = %[2]q
      account = data.aws_caller_identity.current.account_id

      destination_logs_configuration {
        logs_encryption_configuration {
          encryption_strategy = "AWS_OWNED"
        }

        backup_configuration {
          region = %[3]q
        }
      }
    }

    source {
      regions = [%[4]s]
      scope   = "OrganizationId = '${data.aws_organizations_organization.current.id}'"

      source_logs_configuration {
        encrypted_log_group_strategy = "ALLOW"
        log_group_selection_criteria = "LogGroupName LIKE '/aws/lambda%%'"
      }
    }
  }
}
`, rName, dstRegion, bkupRegion, acctest.ListOfStrings(srcRegions...))
}

func testAccCentralizationRuleForOrganizationConfig_tags1(rName, tag1Key, tag1Value string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}
data "aws_organizations_organization" "current" {}

resource "aws_observabilityadmin_centralization_rule_for_organization" "test" {
  rule_name = %[1]q

  rule {
    destination {
      region  = %[4]q
      account = data.aws_caller_identity.current.account_id
    }

    source {
      regions = [%[5]q]
      scope   = "OrganizationId = '${data.aws_organizations_organization.current.id}'"

      source_logs_configuration {
        encrypted_log_group_strategy = "SKIP"
        log_group_selection_criteria = "*"
      }
    }
  }

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tag1Key, tag1Value, endpoints.EuWest1RegionID, endpoints.ApSoutheast1RegionID)
}

func testAccCentralizationRuleForOrganizationConfig_tags2(rName, tag1Key, tag1Value, tag2Key, tag2Value string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}
data "aws_organizations_organization" "current" {}

resource "aws_observabilityadmin_centralization_rule_for_organization" "test" {
  rule_name = %[1]q

  rule {
    destination {
      region  = %[6]q
      account = data.aws_caller_identity.current.account_id
    }

    source {
      regions = [%[7]q]
      scope   = "OrganizationId = '${data.aws_organizations_organization.current.id}'"

      source_logs_configuration {
        encrypted_log_group_strategy = "SKIP"
        log_group_selection_criteria = "*"
      }
    }
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tag1Key, tag1Value, tag2Key, tag2Value, endpoints.EuWest1RegionID, endpoints.ApSoutheast1RegionID)
}
