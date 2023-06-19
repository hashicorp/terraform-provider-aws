package ec2_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
)

func TestAccEC2InstanceConnectEndpoint_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ec2_instance_connect_endpoint.test"
	vpcResourceName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConnectEndpointConfig_basic(vpcResourceName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "instance_id"),
					resource.TestCheckResourceAttr(resourceName, "state", state),
				),
			},
		},
	})
}

func testAccInstanceConnectEndpointConfig_basic(vpcResourceName) string {
	return acctest.ConfigCompose(
		acctest.ConfigVPCWithSubnets(vpcResourceName, 1),
		fmt.Sprintf(`
resource "aws_ec2_instance_connect_endpoint" "test" {
  subnet_id  = aws_subnet.test[0].id
}
`))
}
