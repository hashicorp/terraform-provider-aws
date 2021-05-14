package aws

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceAwsOrganizationsDelegatedServices_basic(t *testing.T) {
	dataSourceName := "data.aws_organizations_delegated_services.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccOrganizationsAccountPreCheck(t)
		},
		ErrorCheck:        testAccErrorCheck(t, organizations.EndpointsID),
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsOrganizationsDelegatedServicesConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceAttrAccountID(dataSourceName, "account_id"),
				),
			},
		},
	})
}

const testAccDataSourceAwsOrganizationsDelegatedServicesConfig = `
data "aws_caller_identity" "current" {}

data "aws_organizations_delegated_services" "test" {
  account_id = data.aws_caller_identity.current.account_id
}
`
