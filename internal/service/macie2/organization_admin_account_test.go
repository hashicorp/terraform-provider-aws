package macie2_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/macie2"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfmacie2 "github.com/hashicorp/terraform-provider-aws/internal/service/macie2"
)

func testAccAwsMacie2OrganizationAdminAccount_basic(t *testing.T) {
	resourceName := "aws_macie2_organization_admin_account.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckOrganizationsAccount(t)
		},
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAwsMacie2OrganizationAdminAccountDestroy,
		ErrorCheck:        testAccErrorCheckSkipMacie2OrganizationAdminAccount(t),
		Steps: []resource.TestStep{
			{
				Config: testAccAwsMacieOrganizationAdminAccountConfigBasic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMacie2OrganizationAdminAccountExists(resourceName),
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

func testAccAwsMacie2OrganizationAdminAccount_disappears(t *testing.T) {
	resourceName := "aws_macie2_organization_admin_account.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckOrganizationsAccount(t)
		},
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAwsMacie2OrganizationAdminAccountDestroy,
		ErrorCheck:        testAccErrorCheckSkipMacie2OrganizationAdminAccount(t),
		Steps: []resource.TestStep{
			{
				Config: testAccAwsMacieOrganizationAdminAccountConfigBasic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMacie2OrganizationAdminAccountExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfmacie2.ResourceAccount(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccErrorCheckSkipMacie2OrganizationAdminAccount(t *testing.T) resource.ErrorCheckFunc {
	return acctest.ErrorCheckSkipMessagesContaining(t,
		"AccessDeniedException: The request failed because you must be a user of the management account for your AWS organization to perform this operation",
	)
}

func testAccCheckAwsMacie2OrganizationAdminAccountExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).Macie2Conn

		adminAccount, err := tfmacie2.GetOrganizationAdminAccount(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		if adminAccount == nil {
			return fmt.Errorf("macie OrganizationAdminAccount (%s) not found", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckAwsMacie2OrganizationAdminAccountDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).Macie2Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_macie2_organization_admin_account" {
			continue
		}

		adminAccount, err := tfmacie2.GetOrganizationAdminAccount(conn, rs.Primary.ID)

		if tfawserr.ErrCodeEquals(err, macie2.ErrCodeResourceNotFoundException) ||
			tfawserr.ErrMessageContains(err, macie2.ErrCodeAccessDeniedException, "Macie is not enabled") {
			continue
		}

		if adminAccount == nil {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("macie OrganizationAdminAccount %q still exists", rs.Primary.ID)
	}

	return nil

}

func testAccAwsMacieOrganizationAdminAccountConfigBasic() string {
	return `
data "aws_caller_identity" "current" {}

resource "aws_macie2_account" "test" {}

data "aws_partition" "current" {}

resource "aws_organizations_organization" "test" {
  aws_service_access_principals = ["macie.${data.aws_partition.current.dns_suffix}"]
  feature_set                   = "ALL"
}

resource "aws_macie2_organization_admin_account" "test" {
  admin_account_id = data.aws_caller_identity.current.account_id
  depends_on       = [aws_macie2_account.test, aws_organizations_organization.test]
}
`
}
