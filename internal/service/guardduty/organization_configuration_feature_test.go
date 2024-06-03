// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package guardduty_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfguardduty "github.com/hashicorp/terraform-provider-aws/internal/service/guardduty"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccOrganizationConfigurationFeature_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_guardduty_organization_configuration_feature.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationsAccount(ctx, t)
			testAccPreCheckDetectorNotExists(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.GuardDutyServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationConfigurationFeatureConfig_basic("RDS_LOGIN_EVENTS", "ALL"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccOrganizationConfigurationFeatureExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "additional_configuration.#", acctest.Ct0),
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

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationsAccount(ctx, t)
			testAccPreCheckDetectorNotExists(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.GuardDutyServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationConfigurationFeatureConfig_additionalConfiguration("NEW", "NONE"),
				Check: resource.ComposeTestCheckFunc(
					testAccOrganizationConfigurationFeatureExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "auto_enable", "NEW"),
					resource.TestCheckResourceAttr(resourceName, "additional_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "additional_configuration.0.auto_enable", "NONE"),
					resource.TestCheckResourceAttr(resourceName, "additional_configuration.0.name", "EKS_ADDON_MANAGEMENT"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, "EKS_RUNTIME_MONITORING"),
				),
			},
			{
				Config: testAccOrganizationConfigurationFeatureConfig_additionalConfiguration("ALL", "ALL"),
				Check: resource.ComposeTestCheckFunc(
					testAccOrganizationConfigurationFeatureExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "auto_enable", "ALL"),
					resource.TestCheckResourceAttr(resourceName, "additional_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "additional_configuration.0.auto_enable", "ALL"),
					resource.TestCheckResourceAttr(resourceName, "additional_configuration.0.name", "EKS_ADDON_MANAGEMENT"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, "EKS_RUNTIME_MONITORING"),
				),
			},
		},
	})
}

func testAccOrganizationConfigurationFeature_multiple(t *testing.T) {
	ctx := acctest.Context(t)
	resource1Name := "aws_guardduty_organization_configuration_feature.test1"
	resource2Name := "aws_guardduty_organization_configuration_feature.test2"
	resource3Name := "aws_guardduty_organization_configuration_feature.test3"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationsAccount(ctx, t)
			testAccPreCheckDetectorNotExists(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.GuardDutyServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationConfigurationFeatureConfig_multiple("ALL", "NEW", "NONE"),
				Check: resource.ComposeTestCheckFunc(
					testAccOrganizationConfigurationFeatureExists(ctx, resource1Name),
					testAccOrganizationConfigurationFeatureExists(ctx, resource2Name),
					testAccOrganizationConfigurationFeatureExists(ctx, resource3Name),
					resource.TestCheckResourceAttr(resource1Name, "additional_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resource1Name, "auto_enable", "ALL"),
					resource.TestCheckResourceAttr(resource1Name, names.AttrName, "EBS_MALWARE_PROTECTION"),
					resource.TestCheckResourceAttr(resource2Name, "additional_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resource2Name, "auto_enable", "NEW"),
					resource.TestCheckResourceAttr(resource2Name, names.AttrName, "LAMBDA_NETWORK_LOGS"),
					resource.TestCheckResourceAttr(resource3Name, "additional_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resource3Name, "auto_enable", "NONE"),
					resource.TestCheckResourceAttr(resource3Name, names.AttrName, "S3_DATA_EVENTS"),
				),
			},
			{
				Config: testAccOrganizationConfigurationFeatureConfig_multiple("NEW", "ALL", "ALL"),
				Check: resource.ComposeTestCheckFunc(
					testAccOrganizationConfigurationFeatureExists(ctx, resource1Name),
					testAccOrganizationConfigurationFeatureExists(ctx, resource2Name),
					testAccOrganizationConfigurationFeatureExists(ctx, resource3Name),
					resource.TestCheckResourceAttr(resource1Name, "additional_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resource1Name, "auto_enable", "NEW"),
					resource.TestCheckResourceAttr(resource1Name, names.AttrName, "EBS_MALWARE_PROTECTION"),
					resource.TestCheckResourceAttr(resource2Name, "additional_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resource2Name, "auto_enable", "ALL"),
					resource.TestCheckResourceAttr(resource2Name, names.AttrName, "LAMBDA_NETWORK_LOGS"),
					resource.TestCheckResourceAttr(resource3Name, "additional_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resource3Name, "auto_enable", "ALL"),
					resource.TestCheckResourceAttr(resource3Name, names.AttrName, "S3_DATA_EVENTS"),
				),
			},
			{
				Config: testAccOrganizationConfigurationFeatureConfig_multiple("NONE", "NONE", "NONE"),
				Check: resource.ComposeTestCheckFunc(
					testAccOrganizationConfigurationFeatureExists(ctx, resource1Name),
					testAccOrganizationConfigurationFeatureExists(ctx, resource2Name),
					testAccOrganizationConfigurationFeatureExists(ctx, resource3Name),
					resource.TestCheckResourceAttr(resource1Name, "additional_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resource1Name, "auto_enable", "NONE"),
					resource.TestCheckResourceAttr(resource1Name, names.AttrName, "EBS_MALWARE_PROTECTION"),
					resource.TestCheckResourceAttr(resource2Name, "additional_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resource2Name, "auto_enable", "NONE"),
					resource.TestCheckResourceAttr(resource2Name, names.AttrName, "LAMBDA_NETWORK_LOGS"),
					resource.TestCheckResourceAttr(resource3Name, "additional_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resource3Name, "auto_enable", "NONE"),
					resource.TestCheckResourceAttr(resource3Name, names.AttrName, "S3_DATA_EVENTS"),
				),
			},
		},
	})
}

func testAccOrganizationConfigurationFeatureExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).GuardDutyConn(ctx)

		_, err := tfguardduty.FindOrganizationConfigurationFeatureByTwoPartKey(ctx, conn, rs.Primary.Attributes["detector_id"], rs.Primary.Attributes[names.AttrName])

		return err
	}
}

var testAccOrganizationConfigurationFeatureConfig_base = acctest.ConfigCompose(testAccOrganizationConfigurationConfig_base, `
resource "aws_guardduty_organization_configuration" "test" {
  depends_on = [aws_guardduty_organization_admin_account.test]

  auto_enable_organization_members = "ALL"
  detector_id                      = aws_guardduty_detector.test.id
}
`)

func testAccOrganizationConfigurationFeatureConfig_basic(name, autoEnable string) string {
	return acctest.ConfigCompose(testAccOrganizationConfigurationFeatureConfig_base, fmt.Sprintf(`
resource "aws_guardduty_organization_configuration_feature" "test" {
  depends_on = [aws_guardduty_organization_configuration.test]

  detector_id = aws_guardduty_detector.test.id
  name        = %[1]q
  auto_enable = %[2]q
}
`, name, autoEnable))
}

func testAccOrganizationConfigurationFeatureConfig_additionalConfiguration(featureAutoEnable, additionalConfigurationAutoEnable string) string {
	return acctest.ConfigCompose(testAccOrganizationConfigurationFeatureConfig_base, fmt.Sprintf(`
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

func testAccOrganizationConfigurationFeatureConfig_multiple(autoEnable1, autoEnable2, autoEnable3 string) string {
	return acctest.ConfigCompose(testAccOrganizationConfigurationFeatureConfig_base, fmt.Sprintf(`
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
