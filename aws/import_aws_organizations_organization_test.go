package aws

import (
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccAWSOrganizationsOrganization_importBasic(t *testing.T) {
	resourceName := "aws_organizations_organization.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSOrganizationsOrganizationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSOrganizationsOrganizationConfig,
			},

			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
