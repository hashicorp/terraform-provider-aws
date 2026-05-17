// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package guardduty_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/guardduty/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	tfguardduty "github.com/hashicorp/terraform-provider-aws/internal/service/guardduty"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccOrganizationConfiguration_basic(t *testing.T) {
	ctx := acctest.Context(t)
	detectorResourceName := "aws_guardduty_detector.test"
	resourceName := "aws_guardduty_organization_configuration.test"

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
				Config: testAccOrganizationConfigurationConfig_autoEnableOrganizationMembers("NONE"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationConfigurationExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "auto_enable_organization_members", "NONE"),
					resource.TestCheckResourceAttrPair(resourceName, "detector_id", detectorResourceName, names.AttrID),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccOrganizationConfiguration_autoEnableOrganizationMembers(t *testing.T) {
	const (
		resourceName = "aws_guardduty_organization_configuration.test"
	)

	for _, from := range enum.Values[types.AutoEnableMembers]() {
		t.Run(from, func(t *testing.T) {
			for _, to := range enum.Values[types.AutoEnableMembers]() {
				if from == to {
					continue
				}

				t.Run(to, func(t *testing.T) {
					ctx := acctest.Context(t)

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
								Config: testAccOrganizationConfigurationConfig_autoEnableOrganizationMembers(types.AutoEnableMembers(from)),
								Check: resource.ComposeTestCheckFunc(
									testAccCheckOrganizationConfigurationExists(ctx, t, resourceName),
									resource.TestCheckResourceAttr(resourceName, "auto_enable_organization_members", from),
								),
							},
							{
								ResourceName:      resourceName,
								ImportState:       true,
								ImportStateVerify: true,
							},
							{
								Config: testAccOrganizationConfigurationConfig_autoEnableOrganizationMembers(types.AutoEnableMembers(to)),
								Check: resource.ComposeTestCheckFunc(
									testAccCheckOrganizationConfigurationExists(ctx, t, resourceName),
									resource.TestCheckResourceAttr(resourceName, "auto_enable_organization_members", to),
								),
							},
							{
								ResourceName:      resourceName,
								ImportState:       true,
								ImportStateVerify: true,
							},
						},
					})
				})
			}
		})
	}
}

func testAccOrganizationConfiguration_s3logs(t *testing.T) {
	ctx := acctest.Context(t)
	detectorResourceName := "aws_guardduty_detector.test"
	resourceName := "aws_guardduty_organization_configuration.test"

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
				Config: testAccOrganizationConfigurationConfig_s3Logs(true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationConfigurationExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "datasources.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.s3_logs.#", "1"),
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
					testAccCheckOrganizationConfigurationExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "datasources.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.s3_logs.#", "1"),
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
				Config: testAccOrganizationConfigurationConfig_kubernetes(true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationConfigurationExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "datasources.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.kubernetes.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.kubernetes.0.audit_logs.#", "1"),
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
					testAccCheckOrganizationConfigurationExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "datasources.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.kubernetes.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.kubernetes.0.audit_logs.#", "1"),
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
				Config: testAccOrganizationConfigurationConfig_malwareprotection(true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationConfigurationExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "datasources.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.malware_protection.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.malware_protection.0.scan_ec2_instance_with_findings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.malware_protection.0.scan_ec2_instance_with_findings.0.ebs_volumes.#", "1"),
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
					testAccCheckOrganizationConfigurationExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "datasources.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.malware_protection.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.malware_protection.0.scan_ec2_instance_with_findings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.malware_protection.0.scan_ec2_instance_with_findings.0.ebs_volumes.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.malware_protection.0.scan_ec2_instance_with_findings.0.ebs_volumes.0.auto_enable", acctest.CtFalse),
					resource.TestCheckResourceAttrPair(resourceName, "detector_id", detectorResourceName, names.AttrID),
				),
			},
		},
	})
}

func testAccCheckOrganizationConfigurationExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).GuardDutyClient(ctx)

		_, err := tfguardduty.FindOrganizationConfigurationByID(ctx, conn, rs.Primary.ID)

		return err
	}
}

const testAccOrganizationConfigurationConfig_base = `
resource "aws_guardduty_detector" "test" {}

resource "aws_guardduty_organization_admin_account" "test" {
  admin_account_id = aws_guardduty_detector.test.account_id
}
`

func testAccOrganizationConfigurationConfig_autoEnableOrganizationMembers(value types.AutoEnableMembers) string {
	return acctest.ConfigCompose(testAccOrganizationConfigurationConfig_base, fmt.Sprintf(`
resource "aws_guardduty_organization_configuration" "test" {
  depends_on = [aws_guardduty_organization_admin_account.test]

  detector_id = aws_guardduty_detector.test.id

  auto_enable_organization_members = %[1]q
}
`, value))
}

func testAccOrganizationConfigurationConfig_s3Logs(autoEnable bool) string {
	return acctest.ConfigCompose(testAccOrganizationConfigurationConfig_base, fmt.Sprintf(`
resource "aws_guardduty_organization_configuration" "test" {
  depends_on = [aws_guardduty_organization_admin_account.test]

  detector_id = aws_guardduty_detector.test.id

  auto_enable_organization_members = "NONE"

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

  detector_id = aws_guardduty_detector.test.id

  auto_enable_organization_members = "NONE"

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

  detector_id = aws_guardduty_detector.test.id

  auto_enable_organization_members = "NONE"

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
