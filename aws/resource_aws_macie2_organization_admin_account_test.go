package aws

import (
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/macie2"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAwsMacie2OrganizationAdminAccount_basic(t *testing.T) {
	resourceName := "aws_macie2_organization_admin_account.test"
	adminAccountID := os.Getenv("AWS_ADMIN_ACCOUNT_ID")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsMacie2OrganizationAdminAccountDestroy,
		ErrorCheck:        testAccErrorCheck(t, macie2.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccAwsMacieOrganizationAdminAccountConfigBasic(adminAccountID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMacie2OrganizationAdminAccountExists(resourceName),
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

func TestAccAwsMacie2OrganizationAdminAccount_disappears(t *testing.T) {
	resourceName := "aws_macie2_organization_admin_account.test"
	adminAccountID := os.Getenv("AWS_ADMIN_ACCOUNT_ID")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsMacie2OrganizationAdminAccountDestroy,
		ErrorCheck:        testAccErrorCheck(t, macie2.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccAwsMacieOrganizationAdminAccountConfigBasic(adminAccountID),
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

		exists := false
		err := conn.ListOrganizationAdminAccountsPages(&macie2.ListOrganizationAdminAccountsInput{}, func(page *macie2.ListOrganizationAdminAccountsOutput, lastPage bool) bool {
			for _, account := range page.AdminAccounts {
				if aws.StringValue(account.AccountId) != rs.Primary.Attributes["admin_account_id"] {
					exists = true
					return false
				}
			}

			return true
		})
		if err != nil {
			return err
		}

		if !exists {
			return fmt.Errorf("macie OrganizationAdminAccount %q does not exist", rs.Primary.ID)
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

		deleted := true
		err := conn.ListOrganizationAdminAccountsPages(&macie2.ListOrganizationAdminAccountsInput{}, func(page *macie2.ListOrganizationAdminAccountsOutput, lastPage bool) bool {
			for _, account := range page.AdminAccounts {
				if aws.StringValue(account.AccountId) != rs.Primary.Attributes["admin_account_id"] {
					deleted = false
					return false
				}
			}

			return true
		})

		if tfawserr.ErrCodeEquals(err, macie2.ErrCodeAccessDeniedException) {
			continue
		}

		if err != nil {
			return err
		}

		if !deleted {
			return fmt.Errorf("macie OrganizationAdminAccount %q still exists", rs.Primary.ID)
		}
	}

	return nil

}

func testAccAwsMacieOrganizationAdminAccountConfigBasic(accountID string) string {
	return fmt.Sprintf(`resource "aws_macie2_account" "test" {}

	resource "aws_macie2_organization_admin_account" "test" {
		admin_account_id = "%s"
		depends_on = [aws_macie2_account.test]
	}
`, accountID)
}
