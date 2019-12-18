package aws

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccDataSourceAwsOrganizationsOrganizationalUnits_basic(t *testing.T) {
	resourceName := "aws_organizations_organizational_unit.test_ou2"
	dataSourceName := "data.aws_organizations_organizational_units.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsOrganizationsOrganizationalUnitsConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "name", dataSourceName, "children.0.name"),
					resource.TestCheckResourceAttrPair(resourceName, "id", dataSourceName, "children.0.id"),
					resource.TestCheckResourceAttrPair(resourceName, "arn", dataSourceName, "children.0.arn"),
				),
			},
		},
	})
}

const testAccDataSourceAwsOrganizationsOrganizationalUnitsConfig = `
data "aws_organizations_organization" "test" {}

resource "aws_organizations_organizational_unit" "test_ou1" {
    name = "test_ou1"
    parent_id = "${data.aws_organizations_organization.test.roots.0.id}"
}

# We create two OUs so the output of the data source is predictable. Otherwise,
# we'll get a longer list of OUs if others are present in the account at the
# root level.

resource "aws_organizations_organizational_unit" "test_ou2" {
    name = "test_ou2"
    parent_id = "${aws_organizations_organizational_unit.test_ou1.id}"
}

data "aws_organizations_organizational_units" "test" {
    parent_id = aws_organizations_organizational_unit.test_ou2.parent_id
}
`
