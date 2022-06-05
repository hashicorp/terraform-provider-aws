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

func TestAccVPCInternetGatewayAttachment_basic(t *testing.T) {
	var v ec2.InternetGatewayAttachment
	resourceName := "aws_internet_gateway_attachment.test"
	igwResourceName := "aws_internet_gateway.test"
	vpcResourceName := "aws_vpc.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInternetGatewayAttachmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCInternetGatewayAttachmentConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInternetGatewayAttachmentExists(resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "internet_gateway_id", igwResourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_id", vpcResourceName, "id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCInternetGatewayAttachment_disappears(t *testing.T) {
	var v ec2.InternetGatewayAttachment
	resourceName := "aws_internet_gateway_attachment.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInternetGatewayAttachmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCInternetGatewayAttachmentConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInternetGatewayAttachmentExists(resourceName, &v),
					acctest.CheckResourceDisappears(acctest.Provider, tfec2.ResourceInternetGatewayAttachment(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckInternetGatewayAttachmentDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_internet_gateway_attachment" {
			continue
		}

		igwID, vpcID, err := tfec2.InternetGatewayAttachmentParseResourceID(rs.Primary.ID)

		if err != nil {
			return err
		}

		_, err = tfec2.FindInternetGatewayAttachment(conn, igwID, vpcID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("EC2 Internet Gateway Attachment %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckInternetGatewayAttachmentExists(n string, v *ec2.InternetGatewayAttachment) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EC2 Internet Gateway Attachment ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

		igwID, vpcID, err := tfec2.InternetGatewayAttachmentParseResourceID(rs.Primary.ID)

		if err != nil {
			return err
		}

		output, err := tfec2.FindInternetGatewayAttachment(conn, igwID, vpcID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccVPCInternetGatewayAttachmentConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_internet_gateway" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_internet_gateway_attachment" "test" {
  internet_gateway_id = aws_internet_gateway.test.id
  vpc_id              = aws_vpc.test.id
}
`, rName)
}
