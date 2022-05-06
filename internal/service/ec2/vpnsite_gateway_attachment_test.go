package ec2_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccVPNSiteGatewayAttachment_basic(t *testing.T) {
	var v ec2.VpcAttachment
	resourceName := "aws_vpn_gateway_attachment.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckVpnGatewayAttachmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVpnGatewayAttachmentConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpnGatewayAttachmentExists(resourceName, &v),
				),
			},
		},
	})
}

func TestAccVPNSiteGatewayAttachment_disappears(t *testing.T) {
	var v ec2.VpcAttachment
	resourceName := "aws_vpn_gateway_attachment.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckVpnGatewayAttachmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVpnGatewayAttachmentConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpnGatewayAttachmentExists(resourceName, &v),
					acctest.CheckResourceDisappears(acctest.Provider, tfec2.ResourceVPNGatewayAttachment(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckVpnGatewayAttachmentExists(n string, v *ec2.VpcAttachment) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EC2 VPN Gateway Attachment ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

		output, err := tfec2.FindVPNGatewayVPCAttachment(conn, rs.Primary.Attributes["vpn_gateway_id"], rs.Primary.Attributes["vpc_id"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckVpnGatewayAttachmentDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_vpn_gateway_attachment" {
			continue
		}

		_, err := tfec2.FindVPNGatewayVPCAttachment(conn, rs.Primary.Attributes["vpn_gateway_id"], rs.Primary.Attributes["vpc_id"])

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("EC2 VPN Gateway Attachment %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccVpnGatewayAttachmentConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpn_gateway" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_vpn_gateway_attachment" "test" {
  vpc_id         = aws_vpc.test.id
  vpn_gateway_id = aws_vpn_gateway.test.id
}
`, rName)
}
