package aws

import (
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccAWSDxConnection_importBasic(t *testing.T) {
	resourceName := "aws_dx_connection.hoge"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsDxConnectionDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccDxConnectionConfig(acctest.RandString(5)),
			},

			resource.TestStep{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
