package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/service/ec2/finder"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/provider"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

func TestAccAWSVpnGatewayAttachment_basic(t *testing.T) {
	var v ec2.VpcAttachment
	resourceName := "aws_vpn_gateway_attachment.test"
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckVpnGatewayAttachmentDestroy,
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

func TestAccAWSVpnGatewayAttachment_disappears(t *testing.T) {
	var v ec2.VpcAttachment
	resourceName := "aws_vpn_gateway_attachment.test"
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckVpnGatewayAttachmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVpnGatewayAttachmentConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpnGatewayAttachmentExists(resourceName, &v),
					acctest.CheckResourceDisappears(acctest.Provider, ResourceVPNGatewayAttachment(), resourceName),
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
			return fmt.Errorf("No ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn
		out, err := finder.VpnGatewayVpcAttachment(conn, rs.Primary.Attributes["vpn_gateway_id"], rs.Primary.Attributes["vpc_id"])
		if err != nil {
			return err
		}
		if out == nil {
			return fmt.Errorf("VPN Gateway Attachment not found")
		}
		if state := aws.StringValue(out.State); state != ec2.AttachmentStatusAttached {
			return fmt.Errorf("VPN Gateway Attachment in incorrect state. Expected: %s, got: %s", ec2.AttachmentStatusAttached, state)
		}

		*v = *out

		return nil
	}
}

func testAccCheckVpnGatewayAttachmentDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_vpn_gateway_attachment" {
			continue
		}

		out, err := finder.VpnGatewayVpcAttachment(conn, rs.Primary.Attributes["vpn_gateway_id"], rs.Primary.Attributes["vpc_id"])
		if err != nil {
			return err
		}
		if out == nil {
			continue
		}
		if state := aws.StringValue(out.State); state != ec2.AttachmentStatusDetached {
			return fmt.Errorf("VPN Gateway Attachment in incorrect state. Expected: %s, got: %s", ec2.AttachmentStatusDetached, state)
		}
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
