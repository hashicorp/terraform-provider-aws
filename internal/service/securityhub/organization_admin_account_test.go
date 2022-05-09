package securityhub_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/securityhub"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfsecurityhub "github.com/hashicorp/terraform-provider-aws/internal/service/securityhub"
)

func testAccOrganizationAdminAccount_basic(t *testing.T) {
	resourceName := "aws_securityhub_organization_admin_account.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckOrganizationsAccount(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, securityhub.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckOrganizationAdminAccountDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationAdminAccountConfig_self(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationAdminAccountExists(resourceName),
					acctest.CheckResourceAttrAccountID(resourceName, "admin_account_id"),
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

func testAccOrganizationAdminAccount_disappears(t *testing.T) {
	resourceName := "aws_securityhub_organization_admin_account.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckOrganizationsAccount(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, securityhub.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckOrganizationAdminAccountDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationAdminAccountConfig_self(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationAdminAccountExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfsecurityhub.ResourceOrganizationAdminAccount(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccOrganizationAdminAccount_MultiRegion(t *testing.T) {
	var providers []*schema.Provider

	resourceName := "aws_securityhub_organization_admin_account.test"
	altResourceName := "aws_securityhub_organization_admin_account.alternate"
	thirdResourceName := "aws_securityhub_organization_admin_account.third"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckOrganizationsAccount(t)
			acctest.PreCheckMultipleRegion(t, 3)
		},
		ErrorCheck:        acctest.ErrorCheck(t, securityhub.EndpointsID),
		ProviderFactories: acctest.FactoriesMultipleRegion(&providers, 3),
		CheckDestroy:      testAccCheckOrganizationAdminAccountDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationAdminAccountConfig_multiRegion(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationAdminAccountExists(resourceName),
					testAccCheckOrganizationAdminAccountExists(altResourceName),
					testAccCheckOrganizationAdminAccountExists(thirdResourceName),
				),
			},
		},
	})
}

func testAccCheckOrganizationAdminAccountDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).SecurityHubConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_securityhub_organization_admin_account" {
			continue
		}

		adminAccount, err := tfsecurityhub.FindAdminAccount(conn, rs.Primary.ID)

		// Because of this resource's dependency, the Organizations organization
		// will be deleted first, resulting in the following valid error
		if tfawserr.ErrMessageContains(err, securityhub.ErrCodeAccessDeniedException, "account is not a member of an organization") {
			continue
		}

		if err != nil {
			return err
		}

		if adminAccount == nil {
			continue
		}

		return fmt.Errorf("expected Security Hub Organization Admin Account (%s) to be removed", rs.Primary.ID)
	}

	return nil
}

func testAccCheckOrganizationAdminAccountExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SecurityHubConn

		adminAccount, err := tfsecurityhub.FindAdminAccount(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		if adminAccount == nil {
			return fmt.Errorf("Security Hub Organization Admin Account (%s) not found", rs.Primary.ID)
		}

		return nil
	}
}

func testAccOrganizationAdminAccountConfig_self() string {
	return `
data "aws_caller_identity" "current" {}

data "aws_partition" "current" {}

resource "aws_organizations_organization" "test" {
  aws_service_access_principals = ["securityhub.${data.aws_partition.current.dns_suffix}"]
  feature_set                   = "ALL"
}

resource "aws_securityhub_account" "test" {}

resource "aws_securityhub_organization_admin_account" "test" {
  depends_on = [aws_organizations_organization.test]

  admin_account_id = data.aws_caller_identity.current.account_id
}
`
}

func testAccOrganizationAdminAccountConfig_multiRegion() string {
	return acctest.ConfigCompose(
		acctest.ConfigMultipleRegionProvider(3),
		`
data "aws_caller_identity" "current" {}

data "aws_partition" "current" {}

resource "aws_organizations_organization" "test" {
  aws_service_access_principals = ["securityhub.${data.aws_partition.current.dns_suffix}"]
  feature_set                   = "ALL"
}

resource "aws_securityhub_account" "test" {}

resource "aws_securityhub_organization_admin_account" "test" {
  depends_on = [aws_organizations_organization.test]

  admin_account_id = data.aws_caller_identity.current.account_id
}

resource "aws_securityhub_organization_admin_account" "alternate" {
  provider = awsalternate

  depends_on = [aws_organizations_organization.test]

  admin_account_id = data.aws_caller_identity.current.account_id
}

resource "aws_securityhub_organization_admin_account" "third" {
  provider = awsthird

  depends_on = [aws_organizations_organization.test]

  admin_account_id = data.aws_caller_identity.current.account_id
}
`)
}
