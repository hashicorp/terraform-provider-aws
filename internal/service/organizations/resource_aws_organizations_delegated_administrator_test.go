package organizations_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/provider"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func testAccAwsOrganizationsDelegatedAdministrator_basic(t *testing.T) {
	var providers []*schema.Provider
	var organization organizations.DelegatedAdministrator
	resourceName := "aws_organizations_delegated_administrator.test"
	servicePrincipal := "config-multiaccountsetup.amazonaws.com"
	dataSourceIdentity := "data.aws_caller_identity.delegated"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckAlternateAccount(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, organizations.EndpointsID),
		ProviderFactories: acctest.FactoriesAlternate(&providers),
		CheckDestroy:      testAccCheckAwsOrganizationsDelegatedAdministratorDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsOrganizationsDelegatedAdministratorConfig(servicePrincipal),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsOrganizationsDelegatedAdministratorExists(resourceName, &organization),
					resource.TestCheckResourceAttrPair(resourceName, "account_id", dataSourceIdentity, "account_id"),
					resource.TestCheckResourceAttr(resourceName, "service_principal", servicePrincipal),
					acctest.CheckResourceAttrRFC3339(resourceName, "delegation_enabled_date"),
					acctest.CheckResourceAttrRFC3339(resourceName, "joined_timestamp"),
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
	var providers []*schema.Provider
	var organization organizations.DelegatedAdministrator
	resourceName := "aws_organizations_delegated_administrator.test"
	servicePrincipal := "config-multiaccountsetup.amazonaws.com"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckAlternateAccount(t)
		},
		ProviderFactories: acctest.FactoriesAlternate(&providers),
		CheckDestroy:      testAccCheckAwsOrganizationsDelegatedAdministratorDestroy,
		ErrorCheck:        acctest.ErrorCheck(t, organizations.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccAwsOrganizationsDelegatedAdministratorConfig(servicePrincipal),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsOrganizationsDelegatedAdministratorExists(resourceName, &organization),
					acctest.CheckResourceDisappears(acctest.Provider, ResourceDelegatedAdministrator(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAwsOrganizationsDelegatedAdministratorDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).OrganizationsConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_organizations_delegated_administrator" {
			continue
		}

		accountID, servicePrincipal, err := decodeOrganizationDelegatedAdministratorID(rs.Primary.ID)
		if err != nil {
			return err
		}
		input := &organizations.ListDelegatedAdministratorsInput{
			ServicePrincipal: aws.String(servicePrincipal),
		}

		exists := false
		err = conn.ListDelegatedAdministratorsPages(input, func(page *organizations.ListDelegatedAdministratorsOutput, lastPage bool) bool {
			for _, delegated := range page.DelegatedAdministrators {
				if aws.StringValue(delegated.Id) == accountID {
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

		accountID, servicePrincipal, err := decodeOrganizationDelegatedAdministratorID(rs.Primary.ID)
		if err != nil {
			return err
		}
		conn := acctest.Provider.Meta().(*conns.AWSClient).OrganizationsConn
		input := &organizations.ListDelegatedAdministratorsInput{
			ServicePrincipal: aws.String(servicePrincipal),
		}

		exists := false
		var resp *organizations.DelegatedAdministrator
		err = conn.ListDelegatedAdministratorsPages(input, func(page *organizations.ListDelegatedAdministratorsOutput, lastPage bool) bool {
			for _, delegated := range page.DelegatedAdministrators {
				if aws.StringValue(delegated.Id) == accountID {
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
	return acctest.ConfigAlternateAccountProvider() + fmt.Sprintf(`
data "aws_caller_identity" "delegated" {
  provider = "awsalternate"
}

resource "aws_organizations_delegated_administrator" "test" {
  account_id        = data.aws_caller_identity.delegated.account_id
  service_principal = %[1]q
}
`, servicePrincipal)
}
