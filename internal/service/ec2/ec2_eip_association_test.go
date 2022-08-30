package ec2_test

import (
	"fmt"
	"strings"
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

func TestAccEC2EIPAssociation_basic(t *testing.T) {
	var a ec2.Address
	resourceName := "aws_eip_association.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEIPAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEIPAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEIPAssociationExists(resourceName, &a),
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

func TestAccEC2EIPAssociation_disappears(t *testing.T) {
	var a ec2.Address
	resourceName := "aws_eip_association.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEIPAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEIPAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEIPAssociationExists(resourceName, &a),
					acctest.CheckResourceDisappears(acctest.Provider, tfec2.ResourceEIPAssociation(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccEC2EIPAssociation_instance(t *testing.T) {
	var a ec2.Address
	resourceName := "aws_eip_association.by_allocation_id"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEIPAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEIPAssociationConfig_instance(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEIPExists("aws_eip.test.0", &a),
					testAccCheckEIPAssociationExists("aws_eip_association.by_allocation_id", &a),
					testAccCheckEIPExists("aws_eip.test.1", &a),
					testAccCheckEIPAssociationExists("aws_eip_association.by_public_ip", &a),
					testAccCheckEIPExists("aws_eip.test.2", &a),
					testAccCheckEIPAssociationExists("aws_eip_association.to_eni", &a),
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

func TestAccEC2EIPAssociation_networkInterface(t *testing.T) {
	var a ec2.Address
	resourceName := "aws_eip_association.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEIPAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEIPAssociationConfig_networkInterface(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEIPAssociationExists(resourceName, &a),
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

func TestAccEC2EIPAssociation_spotInstance(t *testing.T) {
	var a ec2.Address
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_eip_association.test"

	publicKey, _, err := sdkacctest.RandSSHKeyPair(acctest.DefaultEmailAddress)
	if err != nil {
		t.Fatalf("error generating random SSH key: %s", err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEIPAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEIPAssociationConfig_spotInstance(rName, publicKey),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEIPExists("aws_eip.test", &a),
					testAccCheckEIPAssociationExists(resourceName, &a),
					resource.TestCheckResourceAttrSet(resourceName, "allocation_id"),
					resource.TestCheckResourceAttrSet(resourceName, "instance_id"),
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

func testAccCheckEIPAssociationExists(n string, v *ec2.Address) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EC2 EIP Association ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

		var err error
		var output *ec2.Address

		if strings.HasPrefix(rs.Primary.ID, "eipassoc-") {
			output, err = tfec2.FindEIPByAssociationID(conn, rs.Primary.ID)
		} else {
			output, err = tfec2.FindEIPByPublicIP(conn, rs.Primary.ID)
		}

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckEIPAssociationDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_eip_association" {
			continue
		}

		var err error

		if strings.HasPrefix(rs.Primary.ID, "eipassoc-") {
			_, err = tfec2.FindEIPByAssociationID(conn, rs.Primary.ID)
		} else {
			_, err = tfec2.FindEIPByPublicIP(conn, rs.Primary.ID)
		}

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("EC2 EIP %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccEIPAssociationConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
		acctest.ConfigVPCWithSubnets(rName, 1),
		acctest.AvailableEC2InstanceTypeForAvailabilityZone("data.aws_availability_zones.available.names[0]", "t3.micro", "t2.micro"),
		fmt.Sprintf(`
resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type
  subnet_id     = aws_subnet.test[0].id

  tags = {
    Name = %[1]q
  }
}

resource "aws_eip" "test" {
  vpc = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_eip_association" "test" {
  allocation_id = aws_eip.test.id
  instance_id   = aws_instance.test.id
}
`, rName))
}

func testAccEIPAssociationConfig_instance(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptIn(),
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
		fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "192.168.0.0/24"
  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "192.168.0.0/25"
  availability_zone = data.aws_availability_zones.available.names[0]
  tags = {
    Name = %[1]q
  }
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id
}

resource "aws_instance" "test" {
  count             = 2
  ami               = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  availability_zone = data.aws_availability_zones.available.names[0]
  instance_type     = "t2.small"
  subnet_id         = aws_subnet.test.id
  private_ip        = "192.168.0.${count.index + 10}"
}

resource "aws_eip" "test" {
  count = 3
  vpc   = true
}

resource "aws_eip_association" "by_allocation_id" {
  allocation_id = aws_eip.test[0].id
  instance_id   = aws_instance.test[0].id
  depends_on    = [aws_instance.test]
}

resource "aws_eip_association" "by_public_ip" {
  public_ip   = aws_eip.test[1].public_ip
  instance_id = aws_instance.test[1].id
  depends_on  = [aws_instance.test]
}

resource "aws_eip_association" "to_eni" {
  allocation_id        = aws_eip.test[2].id
  network_interface_id = aws_network_interface.test.id
}

resource "aws_network_interface" "test" {
  subnet_id   = aws_subnet.test.id
  private_ips = ["192.168.0.50"]
  depends_on  = [aws_instance.test]

  attachment {
    instance     = aws_instance.test[0].id
    device_index = 1
  }
}
`, rName))
}

func testAccEIPAssociationConfig_networkInterface(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 1), fmt.Sprintf(`
resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_network_interface" "test" {
  subnet_id = aws_subnet.test[0].id

  tags = {
    Name = %[1]q
  }
}

resource "aws_eip" "test" {
  vpc = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_eip_association" "test" {
  allocation_id        = aws_eip.test.id
  network_interface_id = aws_network_interface.test.id
}
`, rName))
}

func testAccEIPAssociationConfig_spotInstance(rName, publicKey string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
		acctest.AvailableEC2InstanceTypeForAvailabilityZone("aws_subnet.test.availability_zone", "t3.micro", "t2.micro"),
		acctest.ConfigAvailableAZsNoOptIn(),
		fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_subnet" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, 0)
  vpc_id            = aws_vpc.test.id
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id
}

resource "aws_key_pair" "test" {
  key_name   = %[1]q
  public_key = %[2]q
}

resource "aws_spot_instance_request" "test" {
  ami                  = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type        = data.aws_ec2_instance_type_offering.available.instance_type
  key_name             = aws_key_pair.test.key_name
  spot_price           = "0.10"
  wait_for_fulfillment = true
  subnet_id            = aws_subnet.test.id
}

resource "aws_eip" "test" {
  vpc = true
}

resource "aws_eip_association" "test" {
  allocation_id = aws_eip.test.id
  instance_id   = aws_spot_instance_request.test.spot_instance_id
}
`, rName, publicKey))
}
