package globalaccelerator_test

import (
	"context"
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
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, globalaccelerator.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCustomRoutingAcceleratorDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCustomRoutingEndpointGroupConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCustomRoutingEndpointGroupExists(ctx, resourceName, &v),
				),
			},
		},
	})

}

func testAccCustomRoutingEndpointGroupConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_globalaccelerator_custom_routing_accelerator" "test" {
  name = %[1]q
}

resource "aws_globalaccelerator_custom_routing_listener" "test" {
  accelerator_arn = aws_globalaccelerator_custom_routing_accelerator.test.id

  port_range = {
    from_port = 443
    to_port   = 443
  }
}

resource "aws_globalaccelerator_custom_routing_endpoint_group" "test" {
  listener_arn = aws_globalaccelerator_custom_routing_listener.test.id
}
`, rName)
}

func testAccCheckCustomRoutingEndpointGroupExists(ctx context.Context, n string, v *globalaccelerator.CustomRoutingEndpointGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).GlobalAcceleratorConn()

		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Global Accelerator Custom Routing Endpoint Group ID is set")
		}

		output, err := tfglobalaccelerator.FindCustomRoutingEndpointGroupByARN(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}
