package aws

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestAccDataSourceAwsOrganizationsDelegatedServices_basic(t *testing.T) {
	var providers []*schema.Provider
	dataSourceName := "data.aws_organizations_delegated_services.test"
	dataSourceIdentity := "data.aws_caller_identity.delegated"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccAlternateAccountPreCheck(t)
		},
		ErrorCheck:        testAccErrorCheck(t, organizations.EndpointsID),
		ProviderFactories: testAccProviderFactoriesAlternate(&providers),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsOrganizationsDelegatedServicesConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "account_id", dataSourceIdentity, "account_id"),
				),
			},
		},
	})
}

func testAccDataSourceAwsOrganizationsDelegatedServicesConfig() string {
	return testAccAlternateAccountProviderConfig() + `
data "aws_caller_identity" "delegated" {
  provider = "awsalternate"
}

data "aws_organizations_delegated_services" "test" {
  account_id        = data.aws_caller_identity.delegated.account_id
}
`
}
