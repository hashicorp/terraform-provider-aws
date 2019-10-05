package aws

import (
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccDataSourceAwsOrganizationsOrganizationalUnit_basic(t *testing.T) {
	resourceName := "aws_organizations_organizational_unit.test_ou2"
	dataSourceName := "data.aws_organizations_organizational_unit.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsOrganizationsOrganizationalUnitConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "name", dataSourceName, "organizational_units.0.name"),
					resource.TestCheckResourceAttrPair(resourceName, "id", dataSourceName, "organizational_units.0.id"),
					resource.TestCheckResourceAttrPair(resourceName, "arn", dataSourceName, "organizational_units.0.arn"),
				),
				/* We ExpectNonEmptyPlan due to the explicit datasource dependency on the test_ou2 resource.
				 * See Terraform config comments for more details. */
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

const testAccDataSourceAwsOrganizationsOrganizationalUnitConfig = `
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

# We add an explicit dependency here because if we don't, Terraform will happily
# fetch all other existing OUs, if any, which probably don't include our test OU
# since it won't have been created yet.

data "aws_organizations_organizational_unit" "test" {
    parent_id = "${aws_organizations_organizational_unit.test_ou1.id}"
    depends_on = [aws_organizations_organizational_unit.test_ou2]
}
`
