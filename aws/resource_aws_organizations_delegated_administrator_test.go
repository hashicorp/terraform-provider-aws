package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func testAccAwsOrganizationsDelegatedAdministrator_basic(t *testing.T) {
	var organization organizations.DelegatedAdministrator
	resourceName := "aws_organizations_delegated_administrator.test"
	servicePrincipal := ""

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t); testAccOrganizationsAccountPreCheck(t) },
		ErrorCheck:        testAccErrorCheck(t, organizations.EndpointsID),
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsOrganizationsDelegatedAdministratorDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsOrganizationsDelegatedAdministratorConfig(servicePrincipal),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsOrganizationsDelegatedAdministratorExists(resourceName, &organization),
					testAccCheckResourceAttrAccountID(resourceName, "account_id"),
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

func testAccAwsOrganizationsDelegatedAdministrator_disappears(t *testing.T) {
	var organization organizations.DelegatedAdministrator
	resourceName := "aws_organizations_delegated_administrator.test"
	servicePrincipal := ""

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsOrganizationsDelegatedAdministratorDestroy,
		ErrorCheck:        testAccErrorCheck(t, organizations.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccAwsOrganizationsDelegatedAdministratorConfig(servicePrincipal),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsOrganizationsDelegatedAdministratorExists(resourceName, &organization),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsOrganizationsDelegatedAdministrator(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAwsOrganizationsDelegatedAdministratorDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).organizationsconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_organizations_organization" {
			continue
		}

		input := &organizations.ListDelegatedAdministratorsInput{
			ServicePrincipal: aws.String(rs.Primary.Attributes["service_principal"]),
		}

		exists := false
		err := conn.ListDelegatedAdministratorsPages(input, func(page *organizations.ListDelegatedAdministratorsOutput, lastPage bool) bool {
			for _, delegated := range page.DelegatedAdministrators {
				if aws.StringValue(delegated.Id) != rs.Primary.ID {
					exists = true
				}
			}

			return !lastPage
		})

		if err != nil {
			return err
		}

		if exists {
			return fmt.Errorf("organization DelegatedAdministrator still exists: %q", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckAwsOrganizationsDelegatedAdministratorExists(n string, org *organizations.DelegatedAdministrator) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Organization ID not set")
		}

		conn := testAccProvider.Meta().(*AWSClient).organizationsconn
		input := &organizations.ListDelegatedAdministratorsInput{
			ServicePrincipal: aws.String(rs.Primary.Attributes["service_principal"]),
		}

		exists := false
		var resp *organizations.DelegatedAdministrator
		err := conn.ListDelegatedAdministratorsPages(input, func(page *organizations.ListDelegatedAdministratorsOutput, lastPage bool) bool {
			for _, delegated := range page.DelegatedAdministrators {
				if aws.StringValue(delegated.Id) != rs.Primary.ID {
					exists = true
					resp = delegated
				}
			}

			return !lastPage
		})

		if err != nil {
			return err
		}

		if !exists {
			return fmt.Errorf("organization DelegatedAdministrator %q does not exist", rs.Primary.ID)
		}

		*org = *resp

		return nil
	}
}

func testAccAwsOrganizationsDelegatedAdministratorConfig(servicePrincipal string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

resource "aws_organizations_delegated_administrator" "test" {
  account_id        = data.aws_caller_identity.current.account_id
  service_principal = %[1]q
}
`, servicePrincipal)
}
