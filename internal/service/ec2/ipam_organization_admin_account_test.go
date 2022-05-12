package ec2_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
)

func TestAccIPAMOrganizationAdminAccount_basic(t *testing.T) {
	var providers []*schema.Provider
	var organization organizations.DelegatedAdministrator
	resourceName := "aws_vpc_ipam_organization_admin_account.test"
	dataSourceIdentity := "data.aws_caller_identity.delegated"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckAlternateAccount(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, organizations.EndpointsID),
		ProviderFactories: acctest.FactoriesAlternate(&providers),
		CheckDestroy:      testAccCheckVPCIpamOrganizationAdminAccountDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCIpamOrganizationAdminAccountConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCIpamOrganizationAdminAccountExists(resourceName, &organization),
					resource.TestCheckResourceAttrPair(resourceName, "id", dataSourceIdentity, "account_id"),
					resource.TestCheckResourceAttr(resourceName, "service_principal", tfec2.Ipam_service_principal),
					acctest.MatchResourceAttrGlobalARN(resourceName, "arn", "organizations", regexp.MustCompile("account/.+")),
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

func testAccCheckVPCIpamOrganizationAdminAccountDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).OrganizationsConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_vpc_ipam_organization_admin_account" {
			continue
		}
		id := rs.Primary.ID

		input := &organizations.ListDelegatedAdministratorsInput{
			ServicePrincipal: aws.String(tfec2.Ipam_service_principal),
		}

		output, err := conn.ListDelegatedAdministrators(input)

		if err != nil {
			return fmt.Errorf("error finding IPAM organization delegated account: (%s): %w", id, err)
		}

		if output == nil || len(output.DelegatedAdministrators) == 0 || output.DelegatedAdministrators[0] == nil {
			return nil
		}
		return fmt.Errorf("organization DelegatedAdministrator still exists: %q", id)
	}
	return nil
}

func testAccCheckVPCIpamOrganizationAdminAccountExists(n string, org *organizations.DelegatedAdministrator) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Organization ID not set")
		}

		accountID := rs.Primary.ID

		conn := acctest.Provider.Meta().(*conns.AWSClient).OrganizationsConn
		input := &organizations.ListDelegatedAdministratorsInput{
			ServicePrincipal: aws.String(tfec2.Ipam_service_principal),
		}

		output, err := conn.ListDelegatedAdministrators(input)

		if err != nil {
			return fmt.Errorf("error finding IPAM organization delegated account: (%s): %w", accountID, err)
		}

		if output == nil || len(output.DelegatedAdministrators) == 0 || output.DelegatedAdministrators[0] == nil {
			return fmt.Errorf("organization DelegatedAdministrator %q does not exist", rs.Primary.ID)

		}

		output_account := output.DelegatedAdministrators[0]

		if aws.StringValue(output_account.Id) != accountID {
			return fmt.Errorf("organization DelegatedAdministrator %q does not match expected %s", accountID, aws.StringValue(output_account.Id))
		}
		*org = *output_account
		return nil
	}
}

func testAccVPCIpamOrganizationAdminAccountConfig() string {
	return acctest.ConfigCompose(acctest.ConfigAlternateAccountProvider() + `
data "aws_caller_identity" "delegated" {
  provider = "awsalternate"
}

resource "aws_vpc_ipam_organization_admin_account" "test" {
  delegated_admin_account_id = data.aws_caller_identity.delegated.account_id
}
`)
}
