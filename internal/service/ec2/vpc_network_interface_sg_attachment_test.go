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

func TestAccVPCNetworkInterfaceSgAttachment_basic(t *testing.T) {
	networkInterfaceResourceName := "aws_network_interface.test"
	securityGroupResourceName := "aws_security_group.test"
	resourceName := "aws_network_interface_sg_attachment.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckNetworkInterfaceSGAttachmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNetworkInterfaceSGAttachmentConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkInterfaceSGAttachmentExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "network_interface_id", networkInterfaceResourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "security_group_id", securityGroupResourceName, "id"),
				),
			},
		},
	})
}

func TestAccVPCNetworkInterfaceSgAttachment_disappears(t *testing.T) {
	resourceName := "aws_network_interface_sg_attachment.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckNetworkInterfaceSGAttachmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNetworkInterfaceSGAttachmentConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkInterfaceSGAttachmentExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfec2.ResourceNetworkInterfaceSGAttachment(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccVPCNetworkInterfaceSgAttachment_instance(t *testing.T) {
	instanceResourceName := "aws_instance.test"
	securityGroupResourceName := "aws_security_group.test"
	resourceName := "aws_network_interface_sg_attachment.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckNetworkInterfaceSGAttachmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNetworkInterfaceSGAttachmentConfig_viaInstance(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkInterfaceSGAttachmentExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "network_interface_id", instanceResourceName, "primary_network_interface_id"),
					resource.TestCheckResourceAttrPair(resourceName, "security_group_id", securityGroupResourceName, "id"),
				),
			},
		},
	})
}

func TestAccVPCNetworkInterfaceSgAttachment_multiple(t *testing.T) {
	networkInterfaceResourceName := "aws_network_interface.test"
	securityGroupResourceName1 := "aws_security_group.test.0"
	securityGroupResourceName2 := "aws_security_group.test.1"
	securityGroupResourceName3 := "aws_security_group.test.2"
	securityGroupResourceName4 := "aws_security_group.test.3"
	resourceName1 := "aws_network_interface_sg_attachment.test.0"
	resourceName2 := "aws_network_interface_sg_attachment.test.1"
	resourceName3 := "aws_network_interface_sg_attachment.test.2"
	resourceName4 := "aws_network_interface_sg_attachment.test.3"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckNetworkInterfaceSGAttachmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNetworkInterfaceSGAttachmentConfig_multiple(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkInterfaceSGAttachmentExists(resourceName1),
					resource.TestCheckResourceAttrPair(resourceName1, "network_interface_id", networkInterfaceResourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName1, "security_group_id", securityGroupResourceName1, "id"),
					testAccCheckNetworkInterfaceSGAttachmentExists(resourceName2),
					resource.TestCheckResourceAttrPair(resourceName2, "network_interface_id", networkInterfaceResourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName2, "security_group_id", securityGroupResourceName2, "id"),
					testAccCheckNetworkInterfaceSGAttachmentExists(resourceName3),
					resource.TestCheckResourceAttrPair(resourceName3, "network_interface_id", networkInterfaceResourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName3, "security_group_id", securityGroupResourceName3, "id"),
					testAccCheckNetworkInterfaceSGAttachmentExists(resourceName4),
					resource.TestCheckResourceAttrPair(resourceName4, "network_interface_id", networkInterfaceResourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName4, "security_group_id", securityGroupResourceName4, "id"),
				),
			},
		},
	})
}

func testAccCheckNetworkInterfaceSGAttachmentExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EC2 Network Interface Security Group Attachment ID is set: %s", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

		_, err := tfec2.FindNetworkInterfaceSecurityGroup(conn, rs.Primary.Attributes["network_interface_id"], rs.Primary.Attributes["security_group_id"])

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckNetworkInterfaceSGAttachmentDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_network_interface_sg_attachment" {
			continue
		}

		_, err := tfec2.FindNetworkInterfaceSecurityGroup(conn, rs.Primary.Attributes["network_interface_id"], rs.Primary.Attributes["security_group_id"])

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("EC2 Network Interface Security Group Attachment %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccVPCNetworkInterfaceSGAttachmentConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "172.16.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  cidr_block = "172.16.10.0/24"
  vpc_id     = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_network_interface" "test" {
  subnet_id = aws_subnet.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_network_interface_sg_attachment" "test" {
  network_interface_id = aws_network_interface.test.id
  security_group_id    = aws_security_group.test.id
}
`, rName)
}

func testAccVPCNetworkInterfaceSGAttachmentConfig_viaInstance(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
		acctest.AvailableEC2InstanceTypeForRegion("t3.micro", "t2.micro"),
		acctest.ConfigAvailableAZsNoOptIn(),
		fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "172.16.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  cidr_block = "172.16.10.0/24"
  vpc_id     = aws_vpc.test.id

  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type
  subnet_id     = aws_subnet.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_network_interface_sg_attachment" "test" {
  network_interface_id = aws_instance.test.primary_network_interface_id
  security_group_id    = aws_security_group.test.id
}
`, rName))
}

func testAccVPCNetworkInterfaceSGAttachmentConfig_multiple(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "172.16.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  cidr_block = "172.16.10.0/24"
  vpc_id     = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_network_interface" "test" {
  subnet_id = aws_subnet.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  count = 4

  name   = "%[1]s-${count.index}"
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_network_interface_sg_attachment" "test" {
  count                = 4
  network_interface_id = aws_network_interface.test.id
  security_group_id    = aws_security_group.test.*.id[count.index]
}
`, rName)
}
