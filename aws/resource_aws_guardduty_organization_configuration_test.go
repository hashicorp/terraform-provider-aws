package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/guardduty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/provider"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

func testAccAwsGuardDutyOrganizationConfiguration_basic(t *testing.T) {
	detectorResourceName := "aws_guardduty_detector.test"
	resourceName := "aws_guardduty_organization_configuration.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckOrganizationsAccount(t)
		},
		ErrorCheck: acctest.ErrorCheck(t, guardduty.EndpointsID),
		Providers:  acctest.Providers,
		// GuardDuty Organization Configuration cannot be deleted separately.
		// Ensure parent resource is destroyed instead.
		CheckDestroy: testAccCheckAwsGuardDutyDetectorDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGuardDutyOrganizationConfigurationConfigAutoEnable(true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "auto_enable", "true"),
					resource.TestCheckResourceAttrPair(resourceName, "detector_id", detectorResourceName, "id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccGuardDutyOrganizationConfigurationConfigAutoEnable(false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "auto_enable", "false"),
					resource.TestCheckResourceAttrPair(resourceName, "detector_id", detectorResourceName, "id"),
				),
			},
		},
	})
}

func testAccAwsGuardDutyOrganizationConfiguration_s3logs(t *testing.T) {
	detectorResourceName := "aws_guardduty_detector.test"
	resourceName := "aws_guardduty_organization_configuration.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckOrganizationsAccount(t)
		},
		ErrorCheck:   acctest.ErrorCheck(t, guardduty.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAwsGuardDutyDetectorDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGuardDutyOrganizationConfigurationConfigS3Logs(true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "auto_enable", "true"),
					resource.TestCheckResourceAttrPair(resourceName, "detector_id", detectorResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.s3_logs.0.auto_enable", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccGuardDutyOrganizationConfigurationConfigS3Logs(false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "auto_enable", "true"),
					resource.TestCheckResourceAttrPair(resourceName, "detector_id", detectorResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.s3_logs.0.auto_enable", "false"),
				),
			},
		},
	})
}

func testAccGuardDutyOrganizationConfigurationConfigAutoEnable(autoEnable bool) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

data "aws_partition" "current" {}

resource "aws_organizations_organization" "test" {
  aws_service_access_principals = ["guardduty.${data.aws_partition.current.dns_suffix}"]
  feature_set                   = "ALL"
}

resource "aws_guardduty_detector" "test" {}

resource "aws_guardduty_organization_admin_account" "test" {
  depends_on = [aws_organizations_organization.test]

  admin_account_id = data.aws_caller_identity.current.account_id
}

resource "aws_guardduty_organization_configuration" "test" {
  depends_on = [aws_guardduty_organization_admin_account.test]

  auto_enable = %[1]t
  detector_id = aws_guardduty_detector.test.id
}
`, autoEnable)
}

func testAccGuardDutyOrganizationConfigurationConfigS3Logs(autoEnable bool) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

data "aws_partition" "current" {}

resource "aws_organizations_organization" "test" {
  aws_service_access_principals = ["guardduty.${data.aws_partition.current.dns_suffix}"]
  feature_set                   = "ALL"
}

resource "aws_guardduty_detector" "test" {}

resource "aws_guardduty_organization_admin_account" "test" {
  depends_on = [aws_organizations_organization.test]

  admin_account_id = data.aws_caller_identity.current.account_id
}

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
`, autoEnable)
}
