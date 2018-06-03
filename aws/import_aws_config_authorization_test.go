package aws

import (
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccConfigAuthorization_import(t *testing.T) {
	resourceName := "aws_config_authorization.example"
	rString := acctest.RandStringFromCharSet(12, "0123456789")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckConfigAuthorizationDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccConfigAuthorizationConfig_basic(rString),
			},

			resource.TestStep{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
