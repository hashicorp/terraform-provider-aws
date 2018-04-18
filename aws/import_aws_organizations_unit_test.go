package aws

import (
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func testAccAwsOrganizationsUnit_importBasic(t *testing.T) {
	resourceName := "aws_organizations_unit.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsOrganizationsUnitDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsOrganizationsUnitConfig("foo"),
			},

			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
