package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAwsDxGateway_importBasic(t *testing.T) {
	resourceName := "aws_dx_gateway.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsDxGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDxGatewayConfig(acctest.RandString(5), randIntRange(64512, 65534)),
			},

			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAwsDxGateway_importComplex(t *testing.T) {
	checkFn := func(s []*terraform.InstanceState) error {
		if len(s) != 3 {
			return fmt.Errorf("Got %d resources, expected 3. State: %#v", len(s), s)
		}
		return nil
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsDxGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDxGatewayAssociationConfig_multiVgws(acctest.RandString(5), randIntRange(64512, 65534)),
			},

			{
				ResourceName:      "aws_dx_gateway.test",
				ImportState:       true,
				ImportStateCheck:  checkFn,
				ImportStateVerify: true,
			},
		},
	})
}
