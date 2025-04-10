// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"fmt"
	"github.com/YakDriver/regexache"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccVPCNetworkInterfaceAttachment_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.NetworkInterface
	resourceName := "aws_network_interface_attachment.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckENIDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNetworkInterfaceAttachmentConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(ctx, "aws_network_interface.test", &conf),
					resource.TestCheckResourceAttrSet(resourceName, "attachment_id"),
					resource.TestCheckResourceAttr(resourceName, "device_index", "1"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrInstanceID),
					resource.TestCheckResourceAttr(resourceName, "network_card_index", "0"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrNetworkInterfaceID),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrStatus),
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

// https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/using-eni.html#network-cards.
// Only specialized (and expensive) instance types support multiple network cards (and hence network_card_index > 0).
// This test verifies that the resource is not created when a non-zero network_card_index is specified on an instance that does not support multiple network cards.
// This ensures that network_card_index is passed when calling the AttachNetworkInterface API.
func TestAccVPCNetworkInterfaceAttachment_networkCardIndex(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckENIDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccVPCNetworkInterfaceAttachmentConfig_networkCardIndex(rName),
				ExpectError: regexache.MustCompile("NetworkCard index 1 exceeds the limit for"),
			},
		},
	})
}

func testAccVPCNetworkInterfaceAttachmentConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
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
  vpc_id            = aws_vpc.test.id
  cidr_block        = "172.16.10.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  vpc_id = aws_vpc.test.id
  name   = %[1]q

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "tcp"
    cidr_blocks = ["10.0.0.0/16"]
  }
}

resource "aws_network_interface" "test" {
  subnet_id       = aws_subnet.test.id
  private_ips     = ["172.16.10.100"]
  security_groups = [aws_security_group.test.id]

  tags = {
    Name = %[1]q
  }
}

resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type
  subnet_id     = aws_subnet.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_network_interface_attachment" "test" {
  device_index         = 1
  instance_id          = aws_instance.test.id
  network_interface_id = aws_network_interface.test.id
}
`, rName))
}

func testAccVPCNetworkInterfaceAttachmentConfig_networkCardIndex(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
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
  vpc_id            = aws_vpc.test.id
  cidr_block        = "172.16.10.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  vpc_id = aws_vpc.test.id
  name   = %[1]q

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "tcp"
    cidr_blocks = ["10.0.0.0/16"]
  }
}

resource "aws_network_interface" "test" {
  subnet_id       = aws_subnet.test.id
  private_ips     = ["172.16.10.100"]
  security_groups = [aws_security_group.test.id]

  tags = {
    Name = %[1]q
  }
}

resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type
  subnet_id     = aws_subnet.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_network_interface_attachment" "test" {
  device_index         = 1
  network_card_index   = 1
  instance_id          = aws_instance.test.id
  network_interface_id = aws_network_interface.test.id
}
`, rName))
}
