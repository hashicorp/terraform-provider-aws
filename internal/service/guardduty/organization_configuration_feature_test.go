// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package guardduty_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/guardduty/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	tfguardduty "github.com/hashicorp/terraform-provider-aws/internal/service/guardduty"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccOrganizationConfigurationFeature_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_guardduty_organization_configuration_feature.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
			testAccPreCheckDetectorNotExists(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.GuardDutyServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationConfigurationFeatureConfig_basic("RDS_LOGIN_EVENTS", "ALL"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccOrganizationConfigurationFeatureExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "auto_enable", "ALL"),
					resource.TestCheckResourceAttrSet(resourceName, "detector_id"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, "RDS_LOGIN_EVENTS"),
				),
			},
		},
	})
}

func testAccOrganizationConfigurationFeature_additionalConfiguration(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_guardduty_organization_configuration_feature.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
			testAccPreCheckDetectorNotExists(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.GuardDutyServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationConfigurationFeatureConfig_additionalConfiguration("NEW", "NONE"),
				Check: resource.ComposeTestCheckFunc(
					testAccOrganizationConfigurationFeatureExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "auto_enable", "NEW"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, "EKS_RUNTIME_MONITORING"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("additional_configuration"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrName: tfknownvalue.StringExact(awstypes.OrgFeatureAdditionalConfigurationEksAddonManagement),
							"auto_enable":  tfknownvalue.StringExact(awstypes.OrgFeatureStatusNone),
						}),
					})),
				},
			},
			{
				Config: testAccOrganizationConfigurationFeatureConfig_additionalConfiguration("ALL", "ALL"),
				Check: resource.ComposeTestCheckFunc(
					testAccOrganizationConfigurationFeatureExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, "EKS_RUNTIME_MONITORING"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("additional_configuration"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrName: tfknownvalue.StringExact(awstypes.OrgFeatureAdditionalConfigurationEksAddonManagement),
							"auto_enable":  tfknownvalue.StringExact(awstypes.OrgFeatureStatusAll),
						}),
					})),
				},
			},
		},
	})
}

func testAccOrganizationConfigurationFeature_additionalConfiguration_rdsDataRisk(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_guardduty_organization_configuration_feature.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
			testAccPreCheckDetectorNotExists(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.GuardDutyServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationConfigurationFeatureConfig_additionalConfiguration_rdsDataRisk("NEW", "NEW"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					testAccOrganizationConfigurationFeatureExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "auto_enable", "NEW"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, "RDS_LOGIN_EVENTS"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("additional_configuration"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrName: tfknownvalue.StringExact(awstypes.OrgFeatureAdditionalConfigurationRdsDataRisk),
							"auto_enable":  tfknownvalue.StringExact(awstypes.OrgFeatureStatusNew),
						}),
					})),
				},
			},
			{
				Config: testAccOrganizationConfigurationFeatureConfig_additionalConfiguration_rdsDataRisk("NEW", "NONE"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					testAccOrganizationConfigurationFeatureExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "auto_enable", "NEW"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, "RDS_LOGIN_EVENTS"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("additional_configuration"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrName: tfknownvalue.StringExact(awstypes.OrgFeatureAdditionalConfigurationRdsDataRisk),
							"auto_enable":  tfknownvalue.StringExact(awstypes.OrgFeatureStatusNone),
						}),
					})),
				},
			},
		},
	})
}

func testAccOrganizationConfigurationFeature_additionalConfiguration_rdsDataRisk_remove(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_guardduty_organization_configuration_feature.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
			testAccPreCheckDetectorNotExists(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.GuardDutyServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationConfigurationFeatureConfig_additionalConfiguration_rdsDataRisk("NEW", "NEW"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					testAccOrganizationConfigurationFeatureExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "auto_enable", "NEW"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, "RDS_LOGIN_EVENTS"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("additional_configuration"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrName: tfknownvalue.StringExact(awstypes.OrgFeatureAdditionalConfigurationRdsDataRisk),
							"auto_enable":  tfknownvalue.StringExact(awstypes.OrgFeatureStatusNew),
						}),
					})),
				},
			},
			{
				Config: testAccOrganizationConfigurationFeatureConfig_basic("RDS_LOGIN_EVENTS", "NEW"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					testAccOrganizationConfigurationFeatureExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "auto_enable", "NEW"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, "RDS_LOGIN_EVENTS"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("additional_configuration"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrName: tfknownvalue.StringExact(awstypes.OrgFeatureAdditionalConfigurationRdsDataRisk),
							"auto_enable":  tfknownvalue.StringExact(awstypes.OrgFeatureStatusNew),
						}),
					})),
				},
			},
		},
	})
}

func testAccOrganizationConfigurationFeature_multiple(t *testing.T) {
	ctx := acctest.Context(t)
	resource1Name := "aws_guardduty_organization_configuration_feature.test1"
	resource2Name := "aws_guardduty_organization_configuration_feature.test2"
	resource3Name := "aws_guardduty_organization_configuration_feature.test3"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
			testAccPreCheckDetectorNotExists(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.GuardDutyServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationConfigurationFeatureConfig_multiple("ALL", "NEW", "NONE"),
				Check: resource.ComposeTestCheckFunc(
					testAccOrganizationConfigurationFeatureExists(ctx, t, resource1Name),
					testAccOrganizationConfigurationFeatureExists(ctx, t, resource2Name),
					testAccOrganizationConfigurationFeatureExists(ctx, t, resource3Name),
					resource.TestCheckResourceAttr(resource1Name, "additional_configuration.#", "0"),
					resource.TestCheckResourceAttr(resource1Name, "auto_enable", "ALL"),
					resource.TestCheckResourceAttr(resource1Name, names.AttrName, "EBS_MALWARE_PROTECTION"),
					resource.TestCheckResourceAttr(resource2Name, "additional_configuration.#", "0"),
					resource.TestCheckResourceAttr(resource2Name, "auto_enable", "NEW"),
					resource.TestCheckResourceAttr(resource2Name, names.AttrName, "LAMBDA_NETWORK_LOGS"),
					resource.TestCheckResourceAttr(resource3Name, "additional_configuration.#", "0"),
					resource.TestCheckResourceAttr(resource3Name, "auto_enable", "NONE"),
					resource.TestCheckResourceAttr(resource3Name, names.AttrName, "S3_DATA_EVENTS"),
				),
			},
			{
				Config: testAccOrganizationConfigurationFeatureConfig_multiple("NEW", "ALL", "ALL"),
				Check: resource.ComposeTestCheckFunc(
					testAccOrganizationConfigurationFeatureExists(ctx, t, resource1Name),
					testAccOrganizationConfigurationFeatureExists(ctx, t, resource2Name),
					testAccOrganizationConfigurationFeatureExists(ctx, t, resource3Name),
					resource.TestCheckResourceAttr(resource1Name, "additional_configuration.#", "0"),
					resource.TestCheckResourceAttr(resource1Name, "auto_enable", "NEW"),
					resource.TestCheckResourceAttr(resource1Name, names.AttrName, "EBS_MALWARE_PROTECTION"),
					resource.TestCheckResourceAttr(resource2Name, "additional_configuration.#", "0"),
					resource.TestCheckResourceAttr(resource2Name, "auto_enable", "ALL"),
					resource.TestCheckResourceAttr(resource2Name, names.AttrName, "LAMBDA_NETWORK_LOGS"),
					resource.TestCheckResourceAttr(resource3Name, "additional_configuration.#", "0"),
					resource.TestCheckResourceAttr(resource3Name, "auto_enable", "ALL"),
					resource.TestCheckResourceAttr(resource3Name, names.AttrName, "S3_DATA_EVENTS"),
				),
			},
			{
				Config: testAccOrganizationConfigurationFeatureConfig_multiple("NONE", "NONE", "NONE"),
				Check: resource.ComposeTestCheckFunc(
					testAccOrganizationConfigurationFeatureExists(ctx, t, resource1Name),
					testAccOrganizationConfigurationFeatureExists(ctx, t, resource2Name),
					testAccOrganizationConfigurationFeatureExists(ctx, t, resource3Name),
					resource.TestCheckResourceAttr(resource1Name, "additional_configuration.#", "0"),
					resource.TestCheckResourceAttr(resource1Name, "auto_enable", "NONE"),
					resource.TestCheckResourceAttr(resource1Name, names.AttrName, "EBS_MALWARE_PROTECTION"),
					resource.TestCheckResourceAttr(resource2Name, "additional_configuration.#", "0"),
					resource.TestCheckResourceAttr(resource2Name, "auto_enable", "NONE"),
					resource.TestCheckResourceAttr(resource2Name, names.AttrName, "LAMBDA_NETWORK_LOGS"),
					resource.TestCheckResourceAttr(resource3Name, "additional_configuration.#", "0"),
					resource.TestCheckResourceAttr(resource3Name, "auto_enable", "NONE"),
					resource.TestCheckResourceAttr(resource3Name, names.AttrName, "S3_DATA_EVENTS"),
				),
			},
		},
	})
}

func testAccOrganizationConfigurationFeatureExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).GuardDutyClient(ctx)

		_, err := tfguardduty.FindOrganizationConfigurationFeatureByTwoPartKey(ctx, conn, rs.Primary.Attributes["detector_id"], rs.Primary.Attributes[names.AttrName])

		return err
	}
}

var testAccOrganizationConfigurationFeatureConfig_base = testAccOrganizationConfigurationConfig_autoEnableOrganizationMembers(awstypes.AutoEnableMembersNone)

func testAccOrganizationConfigurationFeatureConfig_basic(name, autoEnable string) string {
	return acctest.ConfigCompose(
		testAccOrganizationConfigurationFeatureConfig_base,
		fmt.Sprintf(`
resource "aws_guardduty_organization_configuration_feature" "test" {
  depends_on = [aws_guardduty_organization_configuration.test]

  detector_id = aws_guardduty_detector.test.id
  name        = %[1]q
  auto_enable = %[2]q
}
`, name, autoEnable))
}

func testAccOrganizationConfigurationFeatureConfig_additionalConfiguration(featureAutoEnable, additionalConfigurationAutoEnable string) string {
	return acctest.ConfigCompose(
		testAccOrganizationConfigurationFeatureConfig_base,
		fmt.Sprintf(`
resource "aws_guardduty_organization_configuration_feature" "test" {
  depends_on = [aws_guardduty_organization_configuration.test]

  detector_id = aws_guardduty_detector.test.id
  name        = "EKS_RUNTIME_MONITORING"
  auto_enable = %[1]q

  additional_configuration {
    name        = "EKS_ADDON_MANAGEMENT"
    auto_enable = %[2]q
  }
}
`, featureAutoEnable, additionalConfigurationAutoEnable))
}

func testAccOrganizationConfigurationFeatureConfig_additionalConfiguration_rdsDataRisk(featureAutoEnable, additionalConfigurationAutoEnable string) string {
	return acctest.ConfigCompose(
		testAccOrganizationConfigurationFeatureConfig_base,
		fmt.Sprintf(`
resource "aws_guardduty_organization_configuration_feature" "test" {
  depends_on = [aws_guardduty_organization_configuration.test]

  detector_id = aws_guardduty_detector.test.id
  name        = "RDS_LOGIN_EVENTS"
  auto_enable = %[1]q

  additional_configuration {
    name        = "RDS_DATA_RISK"
    auto_enable = %[2]q
  }
}
`, featureAutoEnable, additionalConfigurationAutoEnable))
}

func testAccOrganizationConfigurationFeatureConfig_multiple(autoEnable1, autoEnable2, autoEnable3 string) string {
	return acctest.ConfigCompose(
		testAccOrganizationConfigurationFeatureConfig_base,
		fmt.Sprintf(`
resource "aws_guardduty_organization_configuration_feature" "test1" {
  depends_on = [aws_guardduty_organization_configuration.test]

  detector_id = aws_guardduty_detector.test.id
  name        = "EBS_MALWARE_PROTECTION"
  auto_enable = %[1]q
}

resource "aws_guardduty_organization_configuration_feature" "test2" {
  depends_on = [aws_guardduty_organization_configuration.test]

  detector_id = aws_guardduty_detector.test.id
  name        = "LAMBDA_NETWORK_LOGS"
  auto_enable = %[2]q
}

resource "aws_guardduty_organization_configuration_feature" "test3" {
  depends_on = [aws_guardduty_organization_configuration.test]

  detector_id = aws_guardduty_detector.test.id
  name        = "S3_DATA_EVENTS"
  auto_enable = %[3]q
}
`, autoEnable1, autoEnable2, autoEnable3))
}
