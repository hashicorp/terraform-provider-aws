package globalaccelerator_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/globalaccelerator"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfglobalaccelerator "github.com/hashicorp/terraform-provider-aws/internal/service/globalaccelerator"
)

func TestAccGlobalAcceleratorCustomRoutingEndpointGroup_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v globalaccelerator.CustomRoutingEndpointGroup
	resourceName := "aws_globalaccelerator_custom_routing_endpoint_group.test"
	accName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, globalaccelerator.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalAcceleratorCustomRoutingAcceleratorDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalAcceleratorCustomRoutingEndpointGroupConfig(accName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalAcceleratorCustomRoutingAcceleratorExists(accName),
					testAccCheckGlobalAcceleratorCustomRoutingEndpointGroupExists(resourceName, &v),
				),
			},
		},
	})

}

func testAccGlobalAcceleratorCustomRoutingEndpointGroupConfig(accName string) string {
	return fmt.Sprintf(`
resource "aws_globalaccelerator_custom_routing_accelerator" "test_acc" {
  name = %[1]q
}

resource "aws_globalaccelerator_custom_routing_listener" "test_listener" {
	accelerator_arn = aws_globalaccelerator_custom_routing_accelerator.test_acc.id
	port_range = {
		from_port = 443
		to_port = 443
	}
}

resource "aws_globalaccelerator_custom_routing_endpoint_group" "test" {
	listener_arn: aws_globalaccelerator_custom_routing_listener.test_listener.id
}
`, accName)
}

func testAccCheckGlobalAcceleratorCustomRoutingEndpointGroupExists(name string, v *globalaccelerator.CustomRoutingEndpointGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).GlobalAcceleratorConn()

		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Global Accelerator endpoint group ID is set")
		}

		customRoutingEndpointGroup, err := tfglobalaccelerator.FindCustomRoutingEndpointGroupByARN(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *customRoutingEndpointGroup

		return nil
	}
}
