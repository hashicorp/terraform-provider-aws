package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/macie2"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func testAccAwsMacie2OrganizationAdminAccount_basic(t *testing.T) {
	resourceName := "aws_macie2_organization_admin_account.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsMacie2OrganizationAdminAccountDestroy,
		ErrorCheck:        testAccErrorCheck(t, macie2.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccAwsMacieOrganizationAdminAccountConfigBasic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMacie2OrganizationAdminAccountExists(resourceName),
					testAccCheckResourceAttrAccountID(resourceName, "admin_account_id"),
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
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsMacie2OrganizationAdminAccountDestroy,
		ErrorCheck:        testAccErrorCheck(t, macie2.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccAwsMacieOrganizationAdminAccountConfigBasic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMacie2OrganizationAdminAccountExists(resourceName),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsMacie2Account(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAwsMacie2OrganizationAdminAccountExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		conn := testAccProvider.Meta().(*AWSClient).macie2conn

		adminAccount, err := getMacie2OrganizationAdminAccount(conn, rs.Primary.ID)

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
	conn := testAccProvider.Meta().(*AWSClient).macie2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_macie2_organization_admin_account" {
			continue
		}

		adminAccount, err := getMacie2OrganizationAdminAccount(conn, rs.Primary.ID)

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

resource "aws_macie2_organization_admin_account" "test" {
  admin_account_id = data.aws_caller_identity.current.account_id
  depends_on       = [aws_macie2_account.test]
}
`
}
