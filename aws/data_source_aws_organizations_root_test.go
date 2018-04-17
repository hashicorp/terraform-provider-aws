package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccDataSourceAwsOrganizationRoot_basic(t *testing.T) {
	resourceName := "data.aws_organizations_root.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsOrganizationRootConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceAwsOrganizationRootCheck(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
				),
			},
		},
	})
}

func testAccDataSourceAwsOrganizationRootCheck(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("root module has no resource called %s", name)
		}

		return nil
	}
}

const testAccDataSourceAwsOrganizationRootConfig = `
data "aws_organizations_root" "test" {}
`
