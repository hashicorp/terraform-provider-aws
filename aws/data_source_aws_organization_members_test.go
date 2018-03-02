package aws

import (
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestDataSourceAwsOrganizationMembers(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testDataSourceAwsOrganizationMembersConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_organization_members.main", "", ""),
				),
			},
		},
	})
}

const testDataSourceAwsOrganizationMembersConfig = `
data "aws_organization_members" "main" {}
`
