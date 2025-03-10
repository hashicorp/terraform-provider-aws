// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccEC2EIPAssociation_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var a types.Address
	resourceName := "aws_eip_association.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEIPAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEIPAssociationConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEIPAssociationExists(ctx, resourceName, &a),
					resource.TestMatchResourceAttr(resourceName, names.AttrID, regexache.MustCompile(`^eipassoc-\w+$`)),
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
	ctx := acctest.Context(t)
	var a types.Address
	resourceName := "aws_eip_association.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEIPAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEIPAssociationConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEIPAssociationExists(ctx, resourceName, &a),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfec2.ResourceEIPAssociation(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccEC2EIPAssociation_instance_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var a types.Address
	resourceName := "aws_eip_association.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEIPAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEIPAssociationConfig_instance(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEIPAssociationExists(ctx, resourceName, &a),
					resource.TestMatchResourceAttr(resourceName, names.AttrID, regexache.MustCompile(`^eipassoc-\w+$`)),
					resource.TestCheckResourceAttrPair(resourceName, "allocation_id", "aws_eip.test", names.AttrID),
					resource.TestCheckNoResourceAttr(resourceName, "allow_reassociation"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrInstanceID, "aws_instance.test", names.AttrID),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrNetworkInterfaceID),
					resource.TestCheckResourceAttrPair(resourceName, "private_ip_address", "aws_instance.test", "private_ip"),
					resource.TestCheckResourceAttrPair(resourceName, "public_ip", "aws_eip.test", "public_ip"),
				),
			},
		},
	})
}

func TestAccEC2EIPAssociation_instance_publicIP(t *testing.T) {
	ctx := acctest.Context(t)
	var a types.Address
	resourceName := "aws_eip_association.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEIPAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEIPAssociationConfig_instance_publicIP(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEIPAssociationExists(ctx, resourceName, &a),
					resource.TestMatchResourceAttr(resourceName, names.AttrID, regexache.MustCompile(`^eipassoc-\w+$`)),
					resource.TestCheckResourceAttrPair(resourceName, "allocation_id", "aws_eip.test", names.AttrID),
					resource.TestCheckNoResourceAttr(resourceName, "allow_reassociation"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrInstanceID, "aws_instance.test", names.AttrID),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrNetworkInterfaceID),
					resource.TestCheckResourceAttrPair(resourceName, "private_ip_address", "aws_instance.test", "private_ip"),
					resource.TestCheckResourceAttrPair(resourceName, "public_ip", "aws_eip.test", "public_ip"),
				),
			},
		},
	})
}

func TestAccEC2EIPAssociation_networkInterface(t *testing.T) {
	ctx := acctest.Context(t)
	var a types.Address
	resourceName := "aws_eip_association.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEIPAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEIPAssociationConfig_networkInterface(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEIPAssociationExists(ctx, resourceName, &a),
					resource.TestMatchResourceAttr(resourceName, names.AttrID, regexache.MustCompile(`^eipassoc-\w+$`)),
					resource.TestCheckResourceAttrPair(resourceName, "allocation_id", "aws_eip.test", names.AttrID),
					resource.TestCheckNoResourceAttr(resourceName, "allow_reassociation"),
					resource.TestCheckResourceAttr(resourceName, names.AttrInstanceID, ""),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrNetworkInterfaceID, "aws_network_interface.test", names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "private_ip_address", "aws_network_interface.test", "private_ip"),
					resource.TestCheckResourceAttrPair(resourceName, "public_ip", "aws_eip.test", "public_ip"),
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
	ctx := acctest.Context(t)
	var a types.Address
	resourceName := "aws_eip_association.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	publicKey, _, err := sdkacctest.RandSSHKeyPair(acctest.DefaultEmailAddress)
	if err != nil {
		t.Fatalf("error generating random SSH key: %s", err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEIPAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEIPAssociationConfig_spotInstance(rName, publicKey),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEIPAssociationExists(ctx, resourceName, &a),
					resource.TestMatchResourceAttr(resourceName, names.AttrID, regexache.MustCompile(`^eipassoc-\w+$`)),
					resource.TestCheckResourceAttrPair(resourceName, "allocation_id", "aws_eip.test", names.AttrID),
					resource.TestCheckNoResourceAttr(resourceName, "allow_reassociation"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrInstanceID, "aws_spot_instance_request.test", "spot_instance_id"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrNetworkInterfaceID),
					resource.TestCheckResourceAttrPair(resourceName, "private_ip_address", "aws_spot_instance_request.test", "private_ip"),
					resource.TestCheckResourceAttrPair(resourceName, "public_ip", "aws_eip.test", "public_ip"),
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

func testAccCheckEIPAssociationExists(ctx context.Context, n string, v *types.Address) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		output, err := tfec2.FindEIPByAssociationID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckEIPAssociationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_eip_association" {
				continue
			}

			_, err := tfec2.FindEIPByAssociationID(ctx, conn, rs.Primary.ID)

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
}

func testAccEIPAssociationConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSARM64AMI(),
		acctest.ConfigVPCWithSubnets(rName, 1),
		acctest.AvailableEC2InstanceTypeForAvailabilityZone("data.aws_availability_zones.available.names[0]", "t4g.micro", "t3a.micro"),
		fmt.Sprintf(`
resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn2-ami-minimal-hvm-ebs-arm64.id
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type
  subnet_id     = aws_subnet.test[0].id

  tags = {
    Name = %[1]q
  }
}

resource "aws_eip" "test" {
  domain = "vpc"

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
		acctest.ConfigLatestAmazonLinux2HVMEBSARM64AMI(),
		acctest.ConfigVPCWithSubnets(rName, 1),
		acctest.AvailableEC2InstanceTypeForAvailabilityZone("data.aws_availability_zones.available.names[0]", "t4g.micro", "t3a.micro"),
		fmt.Sprintf(`
resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn2-ami-minimal-hvm-ebs-arm64.id
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type
  subnet_id     = aws_subnet.test[0].id

  tags = {
    Name = %[1]q
  }
}

resource "aws_eip" "test" {
  domain = "vpc"

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

func testAccEIPAssociationConfig_instance_publicIP(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSARM64AMI(),
		acctest.ConfigVPCWithSubnets(rName, 1),
		acctest.AvailableEC2InstanceTypeForAvailabilityZone("data.aws_availability_zones.available.names[0]", "t4g.micro", "t3a.micro"),
		fmt.Sprintf(`
resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn2-ami-minimal-hvm-ebs-arm64.id
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type
  subnet_id     = aws_subnet.test[0].id

  tags = {
    Name = %[1]q
  }
}

resource "aws_eip" "test" {
  domain = "vpc"

  tags = {
    Name = %[1]q
  }
}

resource "aws_eip_association" "test" {
  public_ip   = aws_eip.test.public_ip
  instance_id = aws_instance.test.id
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
  domain = "vpc"

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
		acctest.ConfigLatestAmazonLinux2HVMEBSARM64AMI(),
		acctest.ConfigVPCWithSubnets(rName, 1),
		acctest.AvailableEC2InstanceTypeForAvailabilityZone("data.aws_availability_zones.available.names[0]", "t4g.micro", "t3a.micro"),
		fmt.Sprintf(`
resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_key_pair" "test" {
  key_name   = %[1]q
  public_key = %[2]q

  tags = {
    Name = %[1]q
  }
}

resource "aws_spot_instance_request" "test" {
  ami                  = data.aws_ami.amzn2-ami-minimal-hvm-ebs-arm64.id
  instance_type        = data.aws_ec2_instance_type_offering.available.instance_type
  key_name             = aws_key_pair.test.key_name
  spot_price           = "0.10"
  wait_for_fulfillment = true
  subnet_id            = aws_subnet.test[0].id

  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_tag" "test" {
  resource_id = aws_spot_instance_request.test.spot_instance_id
  key         = "Name"
  value       = %[1]q
}

resource "aws_eip" "test" {
  domain = "vpc"

  tags = {
    Name = %[1]q
  }
}

resource "aws_eip_association" "test" {
  allocation_id = aws_eip.test.id
  instance_id   = aws_spot_instance_request.test.spot_instance_id
}
`, rName, publicKey))
}
