package ec2_test

import (
	"fmt"
	// "log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	// "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
)

// func init() {
// 	// TODO: Remove once multicast support is extended beyond us-east-1
// 	os.Setenv("AWS_DEFAULT_REGION", "us-east-1")
// }

func TestAccAWSTransitGatewayMulticastDomain_basic(t *testing.T) {
	var domain *ec2.TransitGatewayMulticastDomain
	resourceName := "aws_ec2_transit_gateway_multicast_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheckTransitGateway(t)
		},
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckTransitGatewayMulticastDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSTransitGatewayMulticastDomainConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayMulticastDomainExists(resourceName, domain),
					resource.TestCheckResourceAttrSet(resourceName, "transit_gateway_id"),
				),
			},
		},
	})
}

// func testAccAWSTransitGatewayMulticastDomain_disappears(t *testing.T) {
// 	var domain tfec2.TransitGatewayMulticastDomain
// 	resourceName := "aws_ec2_transit_gateway_multicast_domain.test"

// 	resource.ParallelTest(t, resource.TestCase{
// 		PreCheck: func() {
// 			acctest.PreCheck(t)
// 			testAccPreCheckTransitGateway(t)
// 		},
// 		Providers:    acctest.Providers,
// 		CheckDestroy: testAccCheckTransitGatewayMulticastDomainDestroy,
// 		Steps: []resource.TestStep{
// 			{
// 				Config: testAccAWSTransitGatewayMulticastDomainConfig(),
// 				Check: resource.ComposeTestCheckFunc(
// 					testAccCheckTransitGatewayMulticastDomainExists(resourceName, &domain),
// 					testAccCheckTransitGatewayMulticastDomainDisappears(&domain),
// 				),
// 				ExpectNonEmptyPlan: true,
// 			},
// 		},
// 	})
// }

// func testAccAWSTransitGatewayMulticastDomain_Tags(t *testing.T) {
// 	var domain1 tfec2.TransitGatewayMulticastDomain
// 	resourceName := "aws_ec2_transit_gateway_multicast_domain.test"

// 	resource.ParallelTest(t, resource.TestCase{
// 		PreCheck: func() {
// 			acctest.PreCheck(t)
// 			testAccPreCheckTransitGateway(t)
// 		},
// 		Providers:    acctest.Providers,
// 		CheckDestroy: testAccCheckTransitGatewayMulticastDomainDestroy,
// 		Steps: []resource.TestStep{
// 			{
// 				Config: testAccAWSTransitGatewayMulticastDomainConfigTags1("key1", "value1"),
// 				Check: resource.ComposeTestCheckFunc(
// 					testAccCheckTransitGatewayMulticastDomainExists(resourceName, &domain1),
// 					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
// 					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
// 				),
// 			},
// 			{
// 				Config: testAccAWSTransitGatewayMulticastDomainConfigTags2("key1", "value1updated", "key2", "value2"),
// 				Check: resource.ComposeTestCheckFunc(
// 					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
// 					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
// 					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
// 				),
// 			},
// 			{
// 				Config: testAccAWSTransitGatewayMulticastDomainConfigTags1("key2", "value2"),
// 				Check: resource.ComposeTestCheckFunc(
// 					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
// 					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
// 				),
// 			},
// 		},
// 	})
// }

// func testAccAWSTransitGatewayMulticastDomain_Associations(t *testing.T) {
// 	var domain1 tfec2.TransitGatewayMulticastDomain
// 	var attachment1, attachment2 tfec2.TransitGatewayVpcAttachment
// 	var subnet1, subnet2 ec2.Subnet
// 	resourceName := "aws_ec2_transit_gateway_multicast_domain.test"
// 	attachmentName1 := "aws_ec2_transit_gateway_vpc_attachment.test1"
// 	attachmentName2 := "aws_ec2_transit_gateway_vpc_attachment.test2"
// 	subnetName1 := "aws_subnet.test1"
// 	subnetName2 := "aws_subnet.test2"

// 	resource.ParallelTest(t, resource.TestCase{
// 		PreCheck: func() {
// 			acctest.PreCheck(t)
// 			testAccPreCheckTransitGateway(t)
// 		},
// 		Providers:    acctest.Providers,
// 		CheckDestroy: testAccCheckTransitGatewayMulticastDomainDestroy,
// 		Steps: []resource.TestStep{
// 			{
// 				Config: testAccAWSTransitGatewayMulticastDomainConfigAssociation1(),
// 				Check: resource.ComposeTestCheckFunc(
// 					testAccCheckTransitGatewayMulticastDomainExists(resourceName, &domain1),
// 					// testAccCheckTransitGatewayVPCAttachmentExists(attachmentName1, &attachment1),
// 					// testAccCheckSubnetExists(subnetName1, &subnet1),
// 					// testAccCheckSubnetExists(subnetName2, &subnet2),
// 					testAccCheckTransitGatewayMulticastDomainAssociations(&domain1, 1, map[*tfec2.TransitGatewayVpcAttachment][]*ec2.Subnet{
// 						&attachment1: {&subnet1},
// 					}),
// 				),
// 			},
// 			{
// 				Config: testAccAWSTransitGatewayMulticastDomainConfigAssociation2(),
// 				Check: testAccCheckTransitGatewayMulticastDomainAssociations(&domain1, 2, map[*tfec2.TransitGatewayVpcAttachment][]*ec2.Subnet{
// 					&attachment1: {
// 						&subnet1,
// 						&subnet2,
// 					},
// 				}),
// 			},
// 			{
// 				Config: testAccAWSTransitGatewayMulticastDomainConfigAssociation3(),
// 				Check: testAccCheckTransitGatewayMulticastDomainAssociations(&domain1, 2, map[*tfec2.TransitGatewayVpcAttachment][]*ec2.Subnet{
// 					&attachment1: {
// 						&subnet1,
// 						&subnet2,
// 					},
// 				}),
// 			},
// 			{
// 				Config: testAccAWSTransitGatewayMulticastDomainConfigAssociation4(),
// 				Check: testAccCheckTransitGatewayMulticastDomainAssociations(&domain1, 1, map[*tfec2.TransitGatewayVpcAttachment][]*ec2.Subnet{
// 					&attachment1: {&subnet1},
// 				}),
// 			},
// 			{
// 				Config: testAccAWSTransitGatewayMulticastDomainConfigAssociation5(),
// 				Check: resource.ComposeTestCheckFunc(
// 					// testAccCheckTransitGatewayVpcAttachmentExists(attachmentName2, &attachment2),
// 					// testAccCheckSubnetExists(subnetName2, &subnet2),
// 					testAccCheckTransitGatewayMulticastDomainAssociations(&domain1, 2, map[*tfec2.TransitGatewayVpcAttachment][]*ec2.Subnet{
// 						&attachment1: {&subnet1},
// 						&attachment2: {&subnet2},
// 					}),
// 				),
// 			},
// 		},
// 	})
// }

// func testAccAWSTransitGatewayMulticastDomain_Groups(t *testing.T) {
// 	var domain1 tfec2.TransitGatewayMulticastDomain
// 	var instance1, instance2 ec2.Instance
// 	resourceName := "aws_ec2_transit_gateway_multicast_domain.test"
// 	instanceName1 := "aws_instance.test1"
// 	instanceName2 := "aws_instance.test2"
// 	// Note: Currently only one source per-group is allowed

// 	resource.ParallelTest(t, resource.TestCase{
// 		PreCheck: func() {
// 			acctest.PreCheck(t)
// 			testAccPreCheckTransitGateway(t)
// 		},
// 		Providers:    acctest.Providers,
// 		CheckDestroy: testAccCheckTransitGatewayMulticastDomainDestroy,
// 		Steps: []resource.TestStep{
// 			{
// 				Config: testAccAWSTransitGatewayMulticastDomainConfigGroup1(),
// 				Check: resource.ComposeTestCheckFunc(
// 					testAccCheckInstanceExists(instanceName1, &instance1),
// 					testAccCheckTransitGatewayMulticastDomainExists(resourceName, &domain1),
// 					testAccCheckTransitGatewayMulticastDomainGroups(&domain1, 1, true, map[string][]*ec2.Instance{
// 						"224.0.0.1": {&instance1},
// 					}),
// 					testAccCheckTransitGatewayMulticastDomainGroups(&domain1, 1, false, map[string][]*ec2.Instance{
// 						"224.0.0.1": {&instance1},
// 					}),
// 				),
// 			},
// 			{
// 				Config: testAccAWSTransitGatewayMulticastDomainConfigGroup2(),
// 				Check: resource.ComposeTestCheckFunc(
// 					testAccCheckInstanceExists(instanceName2, &instance2),
// 					testAccCheckTransitGatewayMulticastDomainGroups(&domain1, 2, true, map[string][]*ec2.Instance{
// 						"224.0.0.1": {&instance1, &instance2},
// 					}),
// 					testAccCheckTransitGatewayMulticastDomainGroups(&domain1, 1, false, map[string][]*ec2.Instance{
// 						"224.0.0.1": {&instance1},
// 					})),
// 			},
// 			{
// 				Config: testAccAWSTransitGatewayMulticastDomainConfigGroup3(),
// 				Check: resource.ComposeTestCheckFunc(
// 					testAccCheckTransitGatewayMulticastDomainGroups(&domain1, 2, true, map[string][]*ec2.Instance{
// 						"224.0.0.1": {&instance1, &instance2},
// 					}),
// 					testAccCheckTransitGatewayMulticastDomainGroups(&domain1, 1, false, map[string][]*ec2.Instance{
// 						"224.0.0.1": {&instance1},
// 					})),
// 			},
// 			{
// 				Config: testAccAWSTransitGatewayMulticastDomainConfigGroup4(),
// 				Check: resource.ComposeTestCheckFunc(
// 					testAccCheckTransitGatewayMulticastDomainGroups(&domain1, 2, true, map[string][]*ec2.Instance{
// 						"224.0.0.1": {&instance1},
// 						"224.0.0.2": {&instance2},
// 					}),
// 					testAccCheckTransitGatewayMulticastDomainGroups(&domain1, 2, false, map[string][]*ec2.Instance{
// 						"224.0.0.1": {&instance1},
// 						"224.0.0.2": {&instance2},
// 					})),
// 			},
// 		},
// 	})
// }

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

func testAccCheckTransitGatewayMulticastDomainDisappears(domain *tfec2.TransitGatewayMulticastDomain) resource.TestCheckFunc {
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

		return waitForTransitGatewayMulticastDomainDeletion(conn, id)
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

func testAccCheckTransitGatewayMulticastDomainAssociations(domain *tfec2.TransitGatewayMulticastDomain, count int, expected map[*tfec2.TransitGatewayVpcAttachment][]*ec2.Subnet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn
		id := aws.StringValue(domain.TransitGatewayMulticastDomainId)

		assocSet, err := GetTransitGatewayMulticastDomainAssociations(conn, id)
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

func testAccCheckTransitGatewayMulticastDomainGroups(domain *tfec2.TransitGatewayMulticastDomain, count int, member bool, expected map[string][]*ec2.Instance) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn
		id := aws.StringValue(domain.TransitGatewayMulticastDomainId)

		groups, err := SearchTransitGatewayMulticastDomainGroupsByType(conn, id, member)
		if err != nil {
			return err
		}

		groupLen := len(groups)
		groupType := resourceTransitGatewayMulticastDomainGroupType(member)
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

func testAccAWSTransitGatewayMulticastDomainConfig() string {
	return fmt.Sprintf(`
resource "aws_ec2_transit_gateway" "test" {
  multicast_support = "enable"
}
resource "aws_ec2_transit_gateway_multicast_domain" "test" {
  transit_gateway_id = aws_ec2_transit_gateway.test.id
}
`)
}

func aws_ec2_transit_gateway_multicast_domain() string {
	return fmt.Sprintf(`
resource "aws_vpc" "test1" {
  cidr_block = "10.0.0.0/16"
}
resource "aws_subnet" "test1" {
  vpc_id            = aws_vpc.test1.id
  cidr_block        = "10.0.1.0/24"
  availability_zone = "us-east-1a"
}
resource "aws_subnet" "test2" {
  vpc_id            = aws_vpc.test1.id
  cidr_block        = "10.0.2.0/24"
  availability_zone = "us-east-1b"
}
resource "aws_ec2_transit_gateway" "test" {
  multicast_support = "enable"
}
resource "aws_ec2_transit_gateway_vpc_attachment" "test1" {
  subnet_ids         = [aws_subnet.test1.id, aws_subnet.test2.id]
  transit_gateway_id = aws_ec2_transit_gateway.test.id
  vpc_id             = aws_vpc.test1.id
}
resource "aws_ec2_transit_gateway_multicast_domain" "test" {
  transit_gateway_id = aws_ec2_transit_gateway.test.id
  association {
    transit_gateway_attachment_id = aws_ec2_transit_gateway_vpc_attachment.test1.id
    subnet_ids                    = [aws_subnet.test1.id]
  }
}
`)
}

func testAccAWSTransitGatewayMulticastDomainConfigAssociation2() string {
	return fmt.Sprintf(`
resource "aws_vpc" "test1" {
  cidr_block = "10.0.0.0/16"
}
resource "aws_subnet" "test1" {
  vpc_id            = aws_vpc.test1.id
  cidr_block        = "10.0.1.0/24"
  availability_zone = "us-east-1a"
}
resource "aws_subnet" "test2" {
  vpc_id            = aws_vpc.test1.id
  cidr_block        = "10.0.2.0/24"
  availability_zone = "us-east-1b"
}
resource "aws_ec2_transit_gateway" "test" {
  multicast_support = "enable"
}
resource "aws_ec2_transit_gateway_vpc_attachment" "test1" {
  subnet_ids         = [aws_subnet.test1.id, aws_subnet.test2.id]
  transit_gateway_id = aws_ec2_transit_gateway.test.id
  vpc_id             = aws_vpc.test1.id
}
resource "aws_ec2_transit_gateway_multicast_domain" "test" {
  transit_gateway_id = aws_ec2_transit_gateway.test.id
  association {
    transit_gateway_attachment_id = aws_ec2_transit_gateway_vpc_attachment.test1.id
    subnet_ids                    = [aws_subnet.test1.id, aws_subnet.test2.id]
  }
}
`)
}

func testAccAWSTransitGatewayMulticastDomainConfigAssociation3() string {
	return fmt.Sprintf(`
resource "aws_vpc" "test1" {
  cidr_block = "10.0.0.0/16"
}
resource "aws_subnet" "test1" {
  vpc_id            = aws_vpc.test1.id
  cidr_block        = "10.0.1.0/24"
  availability_zone = "us-east-1a"
}
resource "aws_subnet" "test2" {
  vpc_id            = aws_vpc.test1.id
  cidr_block        = "10.0.2.0/24"
  availability_zone = "us-east-1b"
}
resource "aws_ec2_transit_gateway" "test" {
  multicast_support = "enable"
}
resource "aws_ec2_transit_gateway_vpc_attachment" "test1" {
  subnet_ids         = [aws_subnet.test1.id, aws_subnet.test2.id]
  transit_gateway_id = aws_ec2_transit_gateway.test.id
  vpc_id             = aws_vpc.test1.id
}
resource "aws_ec2_transit_gateway_multicast_domain" "test" {
  transit_gateway_id = aws_ec2_transit_gateway.test.id
  association {
    transit_gateway_attachment_id = aws_ec2_transit_gateway_vpc_attachment.test1.id
    subnet_ids                    = [aws_subnet.test1.id]
  }
  association {
    transit_gateway_attachment_id = aws_ec2_transit_gateway_vpc_attachment.test1.id
    subnet_ids                    = [aws_subnet.test2.id]
  }
}
`)
}

func testAccAWSTransitGatewayMulticastDomainConfigAssociation4() string {
	return fmt.Sprintf(`
resource "aws_vpc" "test1" {
  cidr_block = "10.0.0.0/16"
}
resource "aws_subnet" "test1" {
  vpc_id            = aws_vpc.test1.id
  cidr_block        = "10.0.1.0/24"
  availability_zone = "us-east-1a"
}
resource "aws_subnet" "test2" {
  vpc_id            = aws_vpc.test1.id
  cidr_block        = "10.0.2.0/24"
  availability_zone = "us-east-1b"
}
resource "aws_ec2_transit_gateway" "test" {
  multicast_support = "enable"
}
resource "aws_ec2_transit_gateway_vpc_attachment" "test1" {
  subnet_ids         = [aws_subnet.test1.id, aws_subnet.test2.id]
  transit_gateway_id = aws_ec2_transit_gateway.test.id
  vpc_id             = aws_vpc.test1.id
}
resource "aws_ec2_transit_gateway_multicast_domain" "test" {
  transit_gateway_id = aws_ec2_transit_gateway.test.id
  association {
    transit_gateway_attachment_id = aws_ec2_transit_gateway_vpc_attachment.test1.id
    subnet_ids                    = [aws_subnet.test1.id]
  }
  association {
    transit_gateway_attachment_id = aws_ec2_transit_gateway_vpc_attachment.test1.id
    subnet_ids                    = [aws_subnet.test1.id]
  }
}
`)
}

func testAccAWSTransitGatewayMulticastDomainConfigAssociation5() string {
	return fmt.Sprintf(`
resource "aws_vpc" "test1" {
  cidr_block = "10.0.0.0/16"
}
resource "aws_vpc" "test2" {
  cidr_block = "11.0.0.0/16"
}
resource "aws_subnet" "test1" {
  vpc_id            = aws_vpc.test1.id
  cidr_block        = "10.0.1.0/24"
  availability_zone = "us-east-1a"
}
resource "aws_subnet" "test2" {
  vpc_id            = aws_vpc.test2.id
  cidr_block        = "11.0.1.0/24"
  availability_zone = "us-east-1b"
}
resource "aws_ec2_transit_gateway" "test" {
  multicast_support = "enable"
}
resource "aws_ec2_transit_gateway_vpc_attachment" "test1" {
  subnet_ids         = [aws_subnet.test1.id]
  transit_gateway_id = aws_ec2_transit_gateway.test.id
  vpc_id             = aws_vpc.test1.id
}
resource "aws_ec2_transit_gateway_vpc_attachment" "test2" {
  subnet_ids         = [aws_subnet.test2.id]
  transit_gateway_id = aws_ec2_transit_gateway.test.id
  vpc_id             = aws_vpc.test2.id
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
}
`)
}

func testAccAWSTransitGatewayMulticastDomainConfigGroup1() string {
	return fmt.Sprintf(`
resource "aws_vpc" "test1" {
  cidr_block = "10.0.0.0/16"
}
resource "aws_subnet" "test1" {
  vpc_id            = aws_vpc.test1.id
  cidr_block        = "10.0.1.0/24"
  availability_zone = "us-east-1a"
}
resource "aws_instance" "test1" {
  ami           = "ami-04b9e92b5572fa0d1"
  instance_type = "t2.micro"
  subnet_id     = aws_subnet.test1.id
}
resource "aws_ec2_transit_gateway" "test" {
  multicast_support = "enable"
}
resource "aws_ec2_transit_gateway_vpc_attachment" "test1" {
  subnet_ids         = [aws_subnet.test1.id]
  transit_gateway_id = aws_ec2_transit_gateway.test.id
  vpc_id             = aws_vpc.test1.id
}
resource "aws_ec2_transit_gateway_multicast_domain" "test" {
  transit_gateway_id = aws_ec2_transit_gateway.test.id
  association {
    transit_gateway_attachment_id = aws_ec2_transit_gateway_vpc_attachment.test1.id
    subnet_ids                    = [aws_subnet.test1.id]
  }
  members {
    group_ip_address = "224.0.0.1"
    network_interface_ids = [aws_instance.test1.primary_network_interface_id]
  }
  sources {
    group_ip_address = "224.0.0.1"
    network_interface_ids = [aws_instance.test1.primary_network_interface_id]
  }
}
`)
}

func testAccAWSTransitGatewayMulticastDomainConfigGroup2() string {
	return fmt.Sprintf(`
resource "aws_vpc" "test1" {
  cidr_block = "10.0.0.0/16"
}
resource "aws_subnet" "test1" {
  vpc_id            = aws_vpc.test1.id
  cidr_block        = "10.0.1.0/24"
  availability_zone = "us-east-1a"
}
resource "aws_instance" "test1" {
  ami           = "ami-04b9e92b5572fa0d1"
  instance_type = "t2.micro"
  subnet_id     = aws_subnet.test1.id
}
resource "aws_instance" "test2" {
  ami           = "ami-04b9e92b5572fa0d1"
  instance_type = "t2.micro"
  subnet_id     = aws_subnet.test1.id
}
resource "aws_ec2_transit_gateway" "test" {
  multicast_support = "enable"
}
resource "aws_ec2_transit_gateway_vpc_attachment" "test1" {
  subnet_ids         = [aws_subnet.test1.id]
  transit_gateway_id = aws_ec2_transit_gateway.test.id
  vpc_id             = aws_vpc.test1.id
}
resource "aws_ec2_transit_gateway_multicast_domain" "test" {
  transit_gateway_id = aws_ec2_transit_gateway.test.id
  association {
    transit_gateway_attachment_id = aws_ec2_transit_gateway_vpc_attachment.test1.id
    subnet_ids                    = [aws_subnet.test1.id]
  }
  members {
    group_ip_address = "224.0.0.1"
    network_interface_ids = [aws_instance.test1.primary_network_interface_id, aws_instance.test2.primary_network_interface_id]
  }
  sources {
    group_ip_address = "224.0.0.1"
    network_interface_ids = [aws_instance.test1.primary_network_interface_id]
  }
}
`)
}

func testAccAWSTransitGatewayMulticastDomainConfigGroup3() string {
	return fmt.Sprintf(`
resource "aws_vpc" "test1" {
  cidr_block = "10.0.0.0/16"
}
resource "aws_subnet" "test1" {
  vpc_id            = aws_vpc.test1.id
  cidr_block        = "10.0.1.0/24"
  availability_zone = "us-east-1a"
}
resource "aws_instance" "test1" {
  ami           = "ami-04b9e92b5572fa0d1"
  instance_type = "t2.micro"
  subnet_id     = aws_subnet.test1.id
}
resource "aws_instance" "test2" {
  ami           = "ami-04b9e92b5572fa0d1"
  instance_type = "t2.micro"
  subnet_id     = aws_subnet.test1.id
}
resource "aws_ec2_transit_gateway" "test" {
  multicast_support = "enable"
}
resource "aws_ec2_transit_gateway_vpc_attachment" "test1" {
  subnet_ids         = [aws_subnet.test1.id]
  transit_gateway_id = aws_ec2_transit_gateway.test.id
  vpc_id             = aws_vpc.test1.id
}
resource "aws_ec2_transit_gateway_multicast_domain" "test" {
  transit_gateway_id = aws_ec2_transit_gateway.test.id
  association {
    transit_gateway_attachment_id = aws_ec2_transit_gateway_vpc_attachment.test1.id
    subnet_ids                    = [aws_subnet.test1.id]
  }
  members {
    group_ip_address = "224.0.0.1"
    network_interface_ids = [aws_instance.test1.primary_network_interface_id]
  }
  members {
    group_ip_address = "224.0.0.1"
    network_interface_ids = [aws_instance.test2.primary_network_interface_id]
  }
  sources {
    group_ip_address = "224.0.0.1"
    network_interface_ids = [aws_instance.test1.primary_network_interface_id]
  }
}
`)
}

func testAccAWSTransitGatewayMulticastDomainConfigGroup4() string {
	return fmt.Sprintf(`
resource "aws_vpc" "test1" {
  cidr_block = "10.0.0.0/16"
}
resource "aws_subnet" "test1" {
  vpc_id            = aws_vpc.test1.id
  cidr_block        = "10.0.1.0/24"
  availability_zone = "us-east-1a"
}
resource "aws_instance" "test1" {
  ami           = "ami-04b9e92b5572fa0d1"
  instance_type = "t2.micro"
  subnet_id     = aws_subnet.test1.id
}
resource "aws_instance" "test2" {
  ami           = "ami-04b9e92b5572fa0d1"
  instance_type = "t2.micro"
  subnet_id     = aws_subnet.test1.id
}
resource "aws_ec2_transit_gateway" "test" {
  multicast_support = "enable"
}
resource "aws_ec2_transit_gateway_vpc_attachment" "test1" {
  subnet_ids         = [aws_subnet.test1.id]
  transit_gateway_id = aws_ec2_transit_gateway.test.id
  vpc_id             = aws_vpc.test1.id
}
resource "aws_ec2_transit_gateway_multicast_domain" "test" {
  transit_gateway_id = aws_ec2_transit_gateway.test.id
  association {
    transit_gateway_attachment_id = aws_ec2_transit_gateway_vpc_attachment.test1.id
    subnet_ids                    = [aws_subnet.test1.id]
  }
  members {
    group_ip_address = "224.0.0.1"
    network_interface_ids = [aws_instance.test1.primary_network_interface_id]
  }
  members {
    group_ip_address = "224.0.0.2"
    network_interface_ids = [aws_instance.test2.primary_network_interface_id]
  }
  sources {
    group_ip_address = "224.0.0.1"
    network_interface_ids = [aws_instance.test1.primary_network_interface_id]
  }
  sources {
    group_ip_address = "224.0.0.2"
    network_interface_ids = [aws_instance.test2.primary_network_interface_id]
  }
}
`)
}

func testAccAWSTransitGatewayMulticastDomainConfigTags1(tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_ec2_transit_gateway" "test" {
  multicast_support = "enable"
}
resource "aws_ec2_transit_gateway_multicast_domain" "test" {
  transit_gateway_id = aws_ec2_transit_gateway.test.id
  tags = {
    %q = %q
  }
}
`, tagKey1, tagValue1)
}

func testAccAWSTransitGatewayMulticastDomainConfigTags2(tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_ec2_transit_gateway" "test" {
  multicast_support = "enable"
}
resource "aws_ec2_transit_gateway_multicast_domain" "test" {
  transit_gateway_id = aws_ec2_transit_gateway.test.id
  tags = {
    %q = %q
    %q = %q
  }
}
`, tagKey1, tagValue1, tagKey2, tagValue2)
}
