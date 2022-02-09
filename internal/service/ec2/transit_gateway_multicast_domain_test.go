package ec2_test

import (
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
)

func testAccTransitGatewayMulticastDomain_basic(t *testing.T) {
	var domain ec2.TransitGatewayMulticastDomain
	resourceName := "aws_ec2_transit_gateway_multicast_domain.test"
	rName := fmt.Sprintf("tf-testacc-tgwmulticast-%s", sdkacctest.RandString(8))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheckTransitGateway(t)
		},
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckTransitGatewayMulticastDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayMulticastDomainConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayMulticastDomainExists(resourceName, &domain),
					resource.TestCheckResourceAttrSet(resourceName, "transit_gateway_id"),
				),
			},
		},
	})
}

func testAccAWSTransitGatewayMulticastDomain_disappears(t *testing.T) {
	var domain ec2.TransitGatewayMulticastDomain
	resourceName := "aws_ec2_transit_gateway_multicast_domain.test"
	rName := fmt.Sprintf("tf-testacc-tgwmulticast-%s", sdkacctest.RandString(8))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheckTransitGateway(t)
		},
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckTransitGatewayMulticastDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayMulticastDomainConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayMulticastDomainExists(resourceName, &domain),
					testAccCheckTransitGatewayMulticastDomainDisappears(&domain),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccAWSTransitGatewayMulticastDomain_Tags(t *testing.T) {
	// When checking for the number of tags, the Name tag will account for an additional value in the tags attribute count
	var domain1 ec2.TransitGatewayMulticastDomain
	resourceName := "aws_ec2_transit_gateway_multicast_domain.test"
	rName := fmt.Sprintf("tf-testacc-tgwmulticast-%s", sdkacctest.RandString(8))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheckTransitGateway(t)
		},
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckTransitGatewayMulticastDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayMulticastDomainConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayMulticastDomainExists(resourceName, &domain1),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				Config: testAccTransitGatewayMulticastDomainConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "tags.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccTransitGatewayMulticastDomainConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccAWSTransitGatewayMulticastDomain_Associations(t *testing.T) {
	var domain1 ec2.TransitGatewayMulticastDomain
	var attachment1, attachment2 ec2.TransitGatewayVpcAttachment
	var subnet1, subnet2 ec2.Subnet
	rName := fmt.Sprintf("tf-testacc-tgwmulticast-%s", sdkacctest.RandString(8))
	resourceName := "aws_ec2_transit_gateway_multicast_domain.test"
	attachmentName1 := "aws_ec2_transit_gateway_vpc_attachment.test1"
	attachmentName2 := "aws_ec2_transit_gateway_vpc_attachment.test2"
	subnetName1 := "aws_subnet.test1"
	subnetName2 := "aws_subnet.test2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheckTransitGateway(t)
		},
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckTransitGatewayMulticastDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayMulticastDomainConfigAssociation1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayMulticastDomainExists(resourceName, &domain1),
					testAccCheckTransitGatewayVPCAttachmentExists(attachmentName1, &attachment1),
					testAccCheckSubnetExists(subnetName1, &subnet1),
					testAccCheckTransitGatewayMulticastDomainAssociations(&domain1, 1, map[*ec2.TransitGatewayVpcAttachment][]*ec2.Subnet{
						&attachment1: {&subnet1},
					}),
				),
			},
			{
				Config: testAccTransitGatewayMulticastDomainConfigAssociation2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayVPCAttachmentExists(attachmentName2, &attachment2),
					testAccCheckSubnetExists(subnetName2, &subnet2),
					testAccCheckTransitGatewayMulticastDomainAssociations(&domain1, 2, map[*ec2.TransitGatewayVpcAttachment][]*ec2.Subnet{
						&attachment1: {&subnet1},
						&attachment2: {&subnet2},
					}),
				),
			},
		},
	})
}

func testAccTransitGatewayMulticastDomain_Groups(t *testing.T) {
	var domain1 ec2.TransitGatewayMulticastDomain
	var instance1, instance2 ec2.Instance
	rName := fmt.Sprintf("tf-testacc-tgwmulticast-%s", sdkacctest.RandString(8))
	resourceName := "aws_ec2_transit_gateway_multicast_domain.test"
	instanceName1 := "aws_instance.test1"
	instanceName2 := "aws_instance.test2"
	multicastGroup1 := "224.0.0.1"
	multicastGroup2 := "224.0.0.2"
	// Note: Currently only one source per-group is allowed
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckTransitGatewayMulticastDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayMulticastDomainConfigGroup1(rName, multicastGroup1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(instanceName1, &instance1),
					testAccCheckTransitGatewayMulticastDomainExists(resourceName, &domain1),
					testAccCheckTransitGatewayMulticastDomainGroups(&domain1, 1, true, map[string][]*ec2.Instance{
						multicastGroup1: {&instance1},
					}),
					testAccCheckTransitGatewayMulticastDomainGroups(&domain1, 1, false, map[string][]*ec2.Instance{
						multicastGroup1: {&instance1},
					}),
				),
			},
			{
				Config: testAccTransitGatewayMulticastDomainConfigGroup2(rName, multicastGroup1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(instanceName2, &instance2),
					testAccCheckTransitGatewayMulticastDomainGroups(&domain1, 2, true, map[string][]*ec2.Instance{
						multicastGroup1: {&instance1, &instance2},
					}),
					testAccCheckTransitGatewayMulticastDomainGroups(&domain1, 1, false, map[string][]*ec2.Instance{
						multicastGroup1: {&instance1},
					})),
			},
			{
				Config: testAccTransitGatewayMulticastDomainConfigGroup3(rName, multicastGroup1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayMulticastDomainGroups(&domain1, 2, true, map[string][]*ec2.Instance{
						multicastGroup1: {&instance1, &instance2},
					}),
					testAccCheckTransitGatewayMulticastDomainGroups(&domain1, 1, false, map[string][]*ec2.Instance{
						multicastGroup1: {&instance1},
					})),
			},
			{
				Config: testAccTransitGatewayMulticastDomainConfigGroup4(rName, multicastGroup1, multicastGroup2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayMulticastDomainGroups(&domain1, 2, true, map[string][]*ec2.Instance{
						multicastGroup1: {&instance1},
						multicastGroup2: {&instance2},
					}),
					testAccCheckTransitGatewayMulticastDomainGroups(&domain1, 2, false, map[string][]*ec2.Instance{
						multicastGroup1: {&instance1},
						multicastGroup2: {&instance2},
					})),
			},
		},
	})
}

func testAccCheckTransitGatewayMulticastDomainExists(resourceName string, multicastDomain *ec2.TransitGatewayMulticastDomain) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		id := rs.Primary.ID
		if id == "" {
			return fmt.Errorf("no EC2 Transit Gateway Multicast Domain ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn
		domain, err := tfec2.DescribeTransitGatewayMulticastDomain(conn, id)
		if err != nil {
			return err
		}

		if domain == nil {
			return fmt.Errorf("EC2 Transit Gateway Multicast Domain (%s) not found", id)
		}

		state := aws.StringValue(domain.State)
		if state != ec2.TransitGatewayMulticastDomainStateAvailable {
			return fmt.Errorf(
				"EC2 Transit Gateway Multicast Domain (%s) exists in non-available (%s) state", id, state)
		}

		*multicastDomain = *domain

		return nil
	}
}

func testAccCheckTransitGatewayMulticastDomainDisappears(domain *ec2.TransitGatewayMulticastDomain) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn
		id := aws.StringValue(domain.TransitGatewayMulticastDomainId)
		input := &ec2.DeleteTransitGatewayMulticastDomainInput{
			TransitGatewayMulticastDomainId: aws.String(id),
		}

		log.Printf("[DEBUG] Deleting EC2 Transit Gateway Multicast Domain (%s): %s", id, input)
		if _, err := conn.DeleteTransitGatewayMulticastDomain(input); err != nil {
			return err
		}

		return tfec2.WaitForTransitGatewayMulticastDomainDeletion(conn, id)
	}
}

func testAccCheckTransitGatewayMulticastDomainDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ec2_transit_gateway_multicast_domain" {
			continue
		}

		id := rs.Primary.ID
		domain, err := tfec2.DescribeTransitGatewayMulticastDomain(conn, id)
		if tfawserr.ErrMessageContains(err, "InvalidTransitGatewayMulticastDomainId.NotFound", "") {
			continue
		}

		if err != nil {
			return err
		}

		if domain == nil {
			continue
		}

		state := aws.StringValue(domain.State)
		if state != ec2.TransitGatewayMulticastDomainStateDeleted {
			return fmt.Errorf(
				"EC2 Transit Gateway Multicast Domain (%s) still exists in a non-deleted (%s) state",
				id, state)
		}
	}

	return nil
}

func testAccCheckTransitGatewayMulticastDomainAssociations(domain *ec2.TransitGatewayMulticastDomain, count int, expected map[*ec2.TransitGatewayVpcAttachment][]*ec2.Subnet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn
		id := aws.StringValue(domain.TransitGatewayMulticastDomainId)

		assocSet, err := tfec2.GetTransitGatewayMulticastDomainAssociations(conn, id)
		if err != nil {
			return err
		}

		assocLen := len(assocSet)
		if assocLen != count {
			return fmt.Errorf(
				"expected %d EC2 Transit Gateway Multicast Domain assoctiations; got %d", count, assocLen)
		}

		expectedIDs := make(map[string][]string)
		for attachment, subnets := range expected {
			var subnetIDs []string
			for _, subnet := range subnets {
				subnetIDs = append(subnetIDs, aws.StringValue(subnet.SubnetId))
			}
			expectedIDs[aws.StringValue(attachment.TransitGatewayAttachmentId)] = subnetIDs
		}

		for _, assoc := range assocSet {
			attachmentID := aws.StringValue(assoc.TransitGatewayAttachmentId)
			actualSubnetID := aws.StringValue(assoc.Subnet.SubnetId)
			subnetIDs := expectedIDs[attachmentID]
			log.Printf("[DEBUG] Subnet IDS: %s", subnetIDs)
			found := false
			for _, subnetID := range subnetIDs {
				if subnetID == actualSubnetID {
					found = true
					break
				}
			}

			if !found {
				return fmt.Errorf(
					"subnet (%s) not found for expected EC2 Transit Gateway VPC Attachment (%s)",
					actualSubnetID, attachmentID)
			}
		}

		return nil
	}
}

func testAccCheckTransitGatewayMulticastDomainGroups(domain *ec2.TransitGatewayMulticastDomain, count int, member bool, expected map[string][]*ec2.Instance) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn
		id := aws.StringValue(domain.TransitGatewayMulticastDomainId)

		groups, err := tfec2.SearchTransitGatewayMulticastDomainGroupsByType(conn, id, member)
		if err != nil {
			return err
		}

		groupLen := len(groups)
		groupType := tfec2.ResourceTransitGatewayMulticastDomainGroupType(member)
		if groupLen != count {
			return fmt.Errorf(
				"expected %d EC2 Transit Gateway Multicast Domain groups of type %s; got %d",
				count, groupType, groupLen)
		}

		expectedIDs := make(map[string][]string)
		for groupIP, instances := range expected {
			var netIDs []string
			for _, instance := range instances {
				netIDs = append(netIDs, aws.StringValue(instance.NetworkInterfaces[0].NetworkInterfaceId))
			}
			expectedIDs[groupIP] = netIDs
		}

		for _, group := range groups {
			groupIP := aws.StringValue(group.GroupIpAddress)
			actualNetID := aws.StringValue(group.NetworkInterfaceId)
			netIDs := expectedIDs[groupIP]
			log.Printf("[DEBUG] Network Interface IDs: %s", netIDs)
			found := false
			for _, netID := range netIDs {
				if netID == actualNetID {
					found = true
					break
				}
			}

			if !found {
				return fmt.Errorf(
					"network interface ID (%s) not found for expected group IP (%s)", actualNetID, groupIP)
			}
		}

		return nil
	}
}

func testAccTransitGatewayMulticastDomainConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_ec2_transit_gateway" "test" {
  multicast_support = "enable"
  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_transit_gateway_multicast_domain" "test" {
  transit_gateway_id = aws_ec2_transit_gateway.test.id
  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccTransitGatewayMulticastDomainConfigAssociation1(rName string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"
}

resource "aws_vpc" "test1" {
  cidr_block = "10.0.0.0/16"
  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test1" {
  vpc_id            = aws_vpc.test1.id
  cidr_block        = "10.0.1.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]
  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_transit_gateway" "test" {
  multicast_support = "enable"
  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_transit_gateway_vpc_attachment" "test1" {
  subnet_ids         = [aws_subnet.test1.id]
  transit_gateway_id = aws_ec2_transit_gateway.test.id
  vpc_id             = aws_vpc.test1.id
  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_transit_gateway_multicast_domain" "test" {
  transit_gateway_id = aws_ec2_transit_gateway.test.id
  association {
    transit_gateway_attachment_id = aws_ec2_transit_gateway_vpc_attachment.test1.id
    subnet_ids                    = [aws_subnet.test1.id]
  }
  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccTransitGatewayMulticastDomainConfigAssociation2(rName string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"
}

resource "aws_vpc" "test1" {
  cidr_block = "10.0.0.0/16"
  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc" "test2" {
  cidr_block = "11.0.0.0/16"
  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test1" {
  vpc_id            = aws_vpc.test1.id
  cidr_block        = "10.0.1.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]
  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test2" {
  vpc_id            = aws_vpc.test2.id
  cidr_block        = "11.0.1.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]
  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_transit_gateway" "test" {
  multicast_support = "enable"
  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_transit_gateway_vpc_attachment" "test1" {
  subnet_ids         = [aws_subnet.test1.id]
  transit_gateway_id = aws_ec2_transit_gateway.test.id
  vpc_id             = aws_vpc.test1.id
  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_transit_gateway_vpc_attachment" "test2" {
  subnet_ids         = [aws_subnet.test2.id]
  transit_gateway_id = aws_ec2_transit_gateway.test.id
  vpc_id             = aws_vpc.test2.id
  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_transit_gateway_multicast_domain" "test" {
  transit_gateway_id = aws_ec2_transit_gateway.test.id
  association {
    transit_gateway_attachment_id = aws_ec2_transit_gateway_vpc_attachment.test1.id
    subnet_ids                    = [aws_subnet.test1.id]
  }
  association {
    transit_gateway_attachment_id = aws_ec2_transit_gateway_vpc_attachment.test2.id
    subnet_ids                    = [aws_subnet.test2.id]
  }
  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccTransitGatewayMulticastDomainConfigGroup1(rName, multicastGroup1 string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"
}

data "aws_ami" "amazon_linux" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name = "name"
    values = [
      "amzn-ami-hvm-*-x86_64-gp2",
    ]
  }

  filter {
    name = "owner-alias"
    values = [
      "amazon",
    ]
  }
}

resource "aws_vpc" "test1" {
  cidr_block = "10.0.0.0/16"
  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test1" {
  vpc_id            = aws_vpc.test1.id
  cidr_block        = "10.0.1.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]
  tags = {
    Name = %[1]q
  }
}

resource "aws_instance" "test1" {
  ami           = data.aws_ami.amazon_linux.id
  instance_type = "t2.micro"
  subnet_id     = aws_subnet.test1.id
  tags = {
    Name = %[1]q
  }
  lifecycle {
    ignore_changes = [
      iam_instance_profile,
      tags,
      tags_all,
    ]
  }
}

resource "aws_ec2_transit_gateway" "test" {
  multicast_support = "enable"
  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_transit_gateway_vpc_attachment" "test1" {
  subnet_ids         = [aws_subnet.test1.id]
  transit_gateway_id = aws_ec2_transit_gateway.test.id
  vpc_id             = aws_vpc.test1.id
  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_transit_gateway_multicast_domain" "test" {
  transit_gateway_id = aws_ec2_transit_gateway.test.id

  static_source_support = "enable"

  association {
    transit_gateway_attachment_id = aws_ec2_transit_gateway_vpc_attachment.test1.id
    subnet_ids                    = [aws_subnet.test1.id]
  }
  members {
    group_ip_address      = "224.0.0.1"
    network_interface_ids = [aws_instance.test1.primary_network_interface_id]
  }
  sources {
    group_ip_address      = "224.0.0.1"
    network_interface_ids = [aws_instance.test1.primary_network_interface_id]
  }
  tags = {
    Name = %[1]q
  }
}
`, rName, multicastGroup1)
}

func testAccTransitGatewayMulticastDomainConfigGroup2(rName, multicastGroup1 string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"
}

data "aws_ami" "amazon_linux" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name = "name"
    values = [
      "amzn-ami-hvm-*-x86_64-gp2",
    ]
  }

  filter {
    name = "owner-alias"
    values = [
      "amazon",
    ]
  }
}

resource "aws_vpc" "test1" {
  cidr_block = "10.0.0.0/16"
  tags = {
    Name = %[1]q
  }
}
resource "aws_subnet" "test1" {
  vpc_id            = aws_vpc.test1.id
  cidr_block        = "10.0.1.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]
  tags = {
    Name = %[1]q
  }
}
resource "aws_instance" "test1" {
  ami           = data.aws_ami.amazon_linux.id
  instance_type = "t2.micro"
  subnet_id     = aws_subnet.test1.id
  tags = {
    Name = %[1]q
  }
  lifecycle {
    ignore_changes = [
      iam_instance_profile,
      tags,
      tags_all,
    ]
  }
}
resource "aws_instance" "test2" {
  ami           = data.aws_ami.amazon_linux.id
  instance_type = "t2.micro"
  subnet_id     = aws_subnet.test1.id
  tags = {
    Name = %[1]q
  }
  lifecycle {
    ignore_changes = [
      iam_instance_profile,
      tags,
      tags_all,
    ]
  }
}
resource "aws_ec2_transit_gateway" "test" {
  multicast_support = "enable"
  tags = {
    Name = %[1]q
  }
}
resource "aws_ec2_transit_gateway_vpc_attachment" "test1" {
  subnet_ids         = [aws_subnet.test1.id]
  transit_gateway_id = aws_ec2_transit_gateway.test.id
  vpc_id             = aws_vpc.test1.id
  tags = {
    Name = %[1]q
  }
}
resource "aws_ec2_transit_gateway_multicast_domain" "test" {
  transit_gateway_id = aws_ec2_transit_gateway.test.id

  static_source_support = "enable"

  association {
    transit_gateway_attachment_id = aws_ec2_transit_gateway_vpc_attachment.test1.id
    subnet_ids                    = [aws_subnet.test1.id]
  }
  members {
    group_ip_address      = %[2]q
    network_interface_ids = [aws_instance.test1.primary_network_interface_id, aws_instance.test2.primary_network_interface_id]
  }
  sources {
    group_ip_address      = %[2]q
    network_interface_ids = [aws_instance.test1.primary_network_interface_id]
  }
  tags = {
    Name = %[1]q
  }
}
`, rName, multicastGroup1)
}

func testAccTransitGatewayMulticastDomainConfigGroup3(rName, multicastGroup1 string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"
}

data "aws_ami" "amazon_linux" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name = "name"
    values = [
      "amzn-ami-hvm-*-x86_64-gp2",
    ]
  }

  filter {
    name = "owner-alias"
    values = [
      "amazon",
    ]
  }
}
resource "aws_vpc" "test1" {
  cidr_block = "10.0.0.0/16"
  tags = {
    Name = %[1]q
  }
}
resource "aws_subnet" "test1" {
  vpc_id            = aws_vpc.test1.id
  cidr_block        = "10.0.1.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]
  tags = {
    Name = %[1]q
  }
}
resource "aws_instance" "test1" {
  ami           = data.aws_ami.amazon_linux.id
  instance_type = "t2.micro"
  subnet_id     = aws_subnet.test1.id
  tags = {
    Name = %[1]q
  }
  lifecycle {
    ignore_changes = [
      iam_instance_profile,
      tags,
      tags_all,
    ]
  }
}
resource "aws_instance" "test2" {
  ami           = data.aws_ami.amazon_linux.id
  instance_type = "t2.micro"
  subnet_id     = aws_subnet.test1.id
  tags = {
    Name = %[1]q
  }
  lifecycle {
    ignore_changes = [
      iam_instance_profile,
      tags,
      tags_all,
    ]
  }
}
resource "aws_ec2_transit_gateway" "test" {
  multicast_support = "enable"
  tags = {
    Name = %[1]q
  }
}
resource "aws_ec2_transit_gateway_vpc_attachment" "test1" {
  subnet_ids         = [aws_subnet.test1.id]
  transit_gateway_id = aws_ec2_transit_gateway.test.id
  vpc_id             = aws_vpc.test1.id
  tags = {
    Name = %[1]q
  }
}
resource "aws_ec2_transit_gateway_multicast_domain" "test" {
  transit_gateway_id = aws_ec2_transit_gateway.test.id

  static_source_support = "enable"

  association {
    transit_gateway_attachment_id = aws_ec2_transit_gateway_vpc_attachment.test1.id
    subnet_ids                    = [aws_subnet.test1.id]
  }
  members {
    group_ip_address      = %[2]q
    network_interface_ids = [aws_instance.test1.primary_network_interface_id]
  }
  members {
    group_ip_address      = %[2]q
    network_interface_ids = [aws_instance.test2.primary_network_interface_id]
  }
  sources {
    group_ip_address      = %[2]q
    network_interface_ids = [aws_instance.test1.primary_network_interface_id]
  }
  tags = {
    Name = %[1]q
  }
}
`, rName, multicastGroup1)
}

func testAccTransitGatewayMulticastDomainConfigGroup4(rName, multicastGroup1, multicastGroup2 string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"
}

data "aws_ami" "amazon_linux" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name = "name"
    values = [
      "amzn-ami-hvm-*-x86_64-gp2",
    ]
  }

  filter {
    name = "owner-alias"
    values = [
      "amazon",
    ]
  }
}

resource "aws_vpc" "test1" {
  cidr_block = "10.0.0.0/16"
  tags = {
    Name = %[1]q
  }
}
resource "aws_subnet" "test1" {
  vpc_id            = aws_vpc.test1.id
  cidr_block        = "10.0.1.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]
  tags = {
    Name = %[1]q
  }
}
resource "aws_instance" "test1" {
  ami           = data.aws_ami.amazon_linux.id
  instance_type = "t2.micro"
  subnet_id     = aws_subnet.test1.id
  tags = {
    Name = %[1]q
  }
  lifecycle {
    ignore_changes = [
      iam_instance_profile,
      tags,
      tags_all,
    ]
  }
}
resource "aws_instance" "test2" {
  ami           = data.aws_ami.amazon_linux.id
  instance_type = "t2.micro"
  subnet_id     = aws_subnet.test1.id
  tags = {
    Name = %[1]q
  }
}
resource "aws_ec2_transit_gateway" "test" {
  multicast_support = "enable"
  tags = {
    Name = %[1]q
  }
}
resource "aws_ec2_transit_gateway_vpc_attachment" "test1" {
  subnet_ids         = [aws_subnet.test1.id]
  transit_gateway_id = aws_ec2_transit_gateway.test.id
  vpc_id             = aws_vpc.test1.id
  tags = {
    Name = %[1]q
  }
}
resource "aws_ec2_transit_gateway_multicast_domain" "test" {
  transit_gateway_id = aws_ec2_transit_gateway.test.id

  static_source_support = "enable"

  association {
    transit_gateway_attachment_id = aws_ec2_transit_gateway_vpc_attachment.test1.id
    subnet_ids                    = [aws_subnet.test1.id]
  }
  members {
    group_ip_address      = %[2]q
    network_interface_ids = [aws_instance.test1.primary_network_interface_id]
  }
  members {
    group_ip_address      = %[3]q
    network_interface_ids = [aws_instance.test2.primary_network_interface_id]
  }
  sources {
    group_ip_address      = %[2]q
    network_interface_ids = [aws_instance.test1.primary_network_interface_id]
  }
  sources {
    group_ip_address      = %[3]q
    network_interface_ids = [aws_instance.test2.primary_network_interface_id]
  }
  tags = {
    Name = %[1]q
  }
}
`, rName, multicastGroup1, multicastGroup2)
}

func testAccTransitGatewayMulticastDomainConfigTags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_ec2_transit_gateway" "test" {
  multicast_support = "enable"
  tags = {
    Name = %[1]q
  }
}
resource "aws_ec2_transit_gateway_multicast_domain" "test" {
  transit_gateway_id = aws_ec2_transit_gateway.test.id
  tags = {
    Name                          = %[1]q
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccTransitGatewayMulticastDomainConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_ec2_transit_gateway" "test" {
  multicast_support = "enable"
}
resource "aws_ec2_transit_gateway_multicast_domain" "test" {
  transit_gateway_id = aws_ec2_transit_gateway.test.id
  tags = {
    Name                          = %[1]q
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
