package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccDataSourceAwsOrganizationUnit_empty(t *testing.T) {
	resourceName := "data.aws_organizations_unit.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsOrganizationUnitConfig_empty,
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceAwsOrganizationUnitCheck(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
				),
			},
		},
	})
}

func TestAccDataSourceAwsOrganizationUnit_rootTrue(t *testing.T) {
	resourceName := "data.aws_organizations_unit.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsOrganizationUnitConfig_rootTrue,
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceAwsOrganizationUnitCheck(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
				),
			},
		},
	})
}

func testAccDataSourceAwsOrganizationUnitCheck(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("root module has no resource called %s", name)
		}

		return nil
	}
}

const testAccDataSourceAwsOrganizationUnitConfig_empty = `
data "aws_organizations_unit" "test" {}
`

const testAccDataSourceAwsOrganizationUnitConfig_rootTrue = `
data "aws_organizations_unit" "test" {
	root = true
}
`
