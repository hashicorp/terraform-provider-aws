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

func testAccOrganizationConfiguration_basic(t *testing.T) {
	ctx := acctest.Context(t)
	detectorResourceName := "aws_guardduty_detector.test"
	resourceName := "aws_guardduty_organization_configuration.test"

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
				Config: testAccOrganizationConfigurationConfig_autoEnable(true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "auto_enable", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "auto_enable_organization_members", "NEW"),
					resource.TestCheckResourceAttrPair(resourceName, "detector_id", detectorResourceName, names.AttrID),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccOrganizationConfigurationConfig_autoEnable(false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "auto_enable", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "auto_enable_organization_members", "NONE"),
					resource.TestCheckResourceAttrPair(resourceName, "detector_id", detectorResourceName, names.AttrID),
				),
			},
		},
	})
}

func testAccOrganizationConfiguration_autoEnableOrganizationMembers(t *testing.T) {
	ctx := acctest.Context(t)
	detectorResourceName := "aws_guardduty_detector.test"
	resourceName := "aws_guardduty_organization_configuration.test"

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
				Config: testAccOrganizationConfigurationConfig_autoEnableOrganizationMembers("ALL"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "auto_enable_organization_members", "ALL"),
					resource.TestCheckResourceAttr(resourceName, "auto_enable", acctest.CtTrue),
					resource.TestCheckResourceAttrPair(resourceName, "detector_id", detectorResourceName, names.AttrID),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccOrganizationConfigurationConfig_autoEnableOrganizationMembers("NONE"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "auto_enable_organization_members", "NONE"),
					resource.TestCheckResourceAttr(resourceName, "auto_enable", acctest.CtFalse),
					resource.TestCheckResourceAttrPair(resourceName, "detector_id", detectorResourceName, names.AttrID),
				),
			},
			{
				Config: testAccOrganizationConfigurationConfig_autoEnableOrganizationMembers("ALL"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "auto_enable_organization_members", "ALL"),
					resource.TestCheckResourceAttr(resourceName, "auto_enable", acctest.CtTrue),
					resource.TestCheckResourceAttrPair(resourceName, "detector_id", detectorResourceName, names.AttrID),
				),
			},
			{
				Config: testAccOrganizationConfigurationConfig_autoEnable(true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "auto_enable_organization_members", "NEW"),
					resource.TestCheckResourceAttr(resourceName, "auto_enable", acctest.CtTrue),
					resource.TestCheckResourceAttrPair(resourceName, "detector_id", detectorResourceName, names.AttrID),
				),
			},
			{
				Config: testAccOrganizationConfigurationConfig_autoEnableOrganizationMembers("NONE"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "auto_enable_organization_members", "NONE"),
					resource.TestCheckResourceAttr(resourceName, "auto_enable", acctest.CtFalse),
					resource.TestCheckResourceAttrPair(resourceName, "detector_id", detectorResourceName, names.AttrID),
				),
			},
			{
				Config: testAccOrganizationConfigurationConfig_autoEnable(false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "auto_enable_organization_members", "NONE"),
					resource.TestCheckResourceAttr(resourceName, "auto_enable", acctest.CtFalse),
					resource.TestCheckResourceAttrPair(resourceName, "detector_id", detectorResourceName, names.AttrID),
				),
			},
		},
	})
}

func testAccOrganizationConfiguration_s3logs(t *testing.T) {
	ctx := acctest.Context(t)
	detectorResourceName := "aws_guardduty_detector.test"
	resourceName := "aws_guardduty_organization_configuration.test"

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
				Config: testAccOrganizationConfigurationConfig_s3Logs(true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "auto_enable", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "datasources.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.s3_logs.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.s3_logs.0.auto_enable", acctest.CtTrue),
					resource.TestCheckResourceAttrPair(resourceName, "detector_id", detectorResourceName, names.AttrID),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccOrganizationConfigurationConfig_s3Logs(false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "auto_enable", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "datasources.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.s3_logs.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.s3_logs.0.auto_enable", acctest.CtFalse),
					resource.TestCheckResourceAttrPair(resourceName, "detector_id", detectorResourceName, names.AttrID),
				),
			},
		},
	})
}

func testAccOrganizationConfiguration_kubernetes(t *testing.T) {
	ctx := acctest.Context(t)
	detectorResourceName := "aws_guardduty_detector.test"
	resourceName := "aws_guardduty_organization_configuration.test"

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
				Config: testAccOrganizationConfigurationConfig_kubernetes(true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "auto_enable", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "datasources.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.kubernetes.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.kubernetes.0.audit_logs.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.kubernetes.0.audit_logs.0.enable", acctest.CtTrue),
					resource.TestCheckResourceAttrPair(resourceName, "detector_id", detectorResourceName, names.AttrID),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccOrganizationConfigurationConfig_kubernetes(false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "auto_enable", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "datasources.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.kubernetes.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.kubernetes.0.audit_logs.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.kubernetes.0.audit_logs.0.enable", acctest.CtFalse),
					resource.TestCheckResourceAttrPair(resourceName, "detector_id", detectorResourceName, names.AttrID),
				),
			},
		},
	})
}

func testAccOrganizationConfiguration_malwareprotection(t *testing.T) {
	ctx := acctest.Context(t)
	detectorResourceName := "aws_guardduty_detector.test"
	resourceName := "aws_guardduty_organization_configuration.test"

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
				Config: testAccOrganizationConfigurationConfig_malwareprotection(true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "auto_enable", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "datasources.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.malware_protection.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.malware_protection.0.scan_ec2_instance_with_findings.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.malware_protection.0.scan_ec2_instance_with_findings.0.ebs_volumes.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.malware_protection.0.scan_ec2_instance_with_findings.0.ebs_volumes.0.auto_enable", acctest.CtTrue),
					resource.TestCheckResourceAttrPair(resourceName, "detector_id", detectorResourceName, names.AttrID),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccOrganizationConfigurationConfig_malwareprotection(false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "auto_enable", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "datasources.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.malware_protection.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.malware_protection.0.scan_ec2_instance_with_findings.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.malware_protection.0.scan_ec2_instance_with_findings.0.ebs_volumes.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.malware_protection.0.scan_ec2_instance_with_findings.0.ebs_volumes.0.auto_enable", acctest.CtFalse),
					resource.TestCheckResourceAttrPair(resourceName, "detector_id", detectorResourceName, names.AttrID),
				),
			},
		},
	})
}

func testAccCheckOrganizationConfigurationExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).GuardDutyConn(ctx)

		_, err := tfguardduty.FindOrganizationConfigurationByID(ctx, conn, rs.Primary.ID)

		return err
	}
}

const testAccOrganizationConfigurationConfig_base = `
data "aws_caller_identity" "current" {}

data "aws_partition" "current" {}

resource "aws_organizations_organization" "test" {
  aws_service_access_principals = [
    "guardduty.${data.aws_partition.current.dns_suffix}",
    "malware-protection.guardduty.${data.aws_partition.current.dns_suffix}",
  ]

  feature_set = "ALL"
}

resource "aws_guardduty_detector" "test" {}

resource "aws_guardduty_organization_admin_account" "test" {
  depends_on = [aws_organizations_organization.test, aws_guardduty_detector.test]

  admin_account_id = data.aws_caller_identity.current.account_id
}
`

func testAccOrganizationConfigurationConfig_autoEnable(autoEnable bool) string {
	return acctest.ConfigCompose(testAccOrganizationConfigurationConfig_base, fmt.Sprintf(`
resource "aws_guardduty_organization_configuration" "test" {
  depends_on = [aws_guardduty_organization_admin_account.test]

  auto_enable = %[1]t
  detector_id = aws_guardduty_detector.test.id
}
`, autoEnable))
}

func testAccOrganizationConfigurationConfig_autoEnableOrganizationMembers(value string) string {
	return acctest.ConfigCompose(testAccOrganizationConfigurationConfig_base, fmt.Sprintf(`
resource "aws_guardduty_organization_configuration" "test" {
  depends_on = [aws_guardduty_organization_admin_account.test]

  auto_enable_organization_members = %[1]q
  detector_id                      = aws_guardduty_detector.test.id
}
`, value))
}

func testAccOrganizationConfigurationConfig_s3Logs(autoEnable bool) string {
	return acctest.ConfigCompose(testAccOrganizationConfigurationConfig_base, fmt.Sprintf(`
resource "aws_guardduty_organization_configuration" "test" {
  depends_on = [aws_guardduty_organization_admin_account.test]

  auto_enable = true
  detector_id = aws_guardduty_detector.test.id

  datasources {
    s3_logs {
      auto_enable = %[1]t
    }
  }
}
`, autoEnable))
}

func testAccOrganizationConfigurationConfig_kubernetes(autoEnable bool) string {
	return acctest.ConfigCompose(testAccOrganizationConfigurationConfig_base, fmt.Sprintf(`
resource "aws_guardduty_organization_configuration" "test" {
  depends_on = [aws_guardduty_organization_admin_account.test]

  auto_enable = true
  detector_id = aws_guardduty_detector.test.id

  datasources {
    kubernetes {
      audit_logs {
        enable = %[1]t
      }
    }
  }
}
`, autoEnable))
}

func testAccOrganizationConfigurationConfig_malwareprotection(autoEnable bool) string {
	return acctest.ConfigCompose(testAccOrganizationConfigurationConfig_base, fmt.Sprintf(`
resource "aws_guardduty_organization_configuration" "test" {
  depends_on = [aws_guardduty_organization_admin_account.test]

  auto_enable = true
  detector_id = aws_guardduty_detector.test.id

  datasources {
    malware_protection {
      scan_ec2_instance_with_findings {
        ebs_volumes {
          auto_enable = %[1]t
        }
      }
    }
  }
}
`, autoEnable))
}
