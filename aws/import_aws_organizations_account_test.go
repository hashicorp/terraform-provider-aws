package aws

import (
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func testAccAwsOrganizationsAccount_importBasic(t *testing.T) {
	resourceName := "aws_organizations_account.test"
	name := "my_new_account"
	email := "foo@bar.org"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsOrganizationsAccountDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsOrganizationsAccountConfig(name, email),
			},

			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
