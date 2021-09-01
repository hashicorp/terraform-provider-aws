package aws

import (
	"errors"
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("aws_ec2_transit_gateway_connect", &resource.Sweeper{
		Name: "aws_ec2_transit_gateway_connect",
		F:    testSweepEc2TransitGatewayConnect,
	})
}

func testSweepEc2TransitGatewayConnect(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).ec2conn
	input := &ec2.DescribeTransitGatewayConnectsInput{}

	for {
		output, err := conn.DescribeTransitGatewayConnects(input)

		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping EC2 Transit Gateway Connect sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error retrieving EC2 Transit Gateway Connects: %s", err)
		}

		for _, attachment := range output.TransitGatewayConnects {
			if aws.StringValue(attachment.State) == ec2.TransitGatewayAttachmentStateDeleted {
				continue
			}

			id := aws.StringValue(attachment.TransitGatewayAttachmentId)

			input := &ec2.DeleteTransitGatewayConnectInput{
				TransitGatewayAttachmentId: aws.String(id),
			}

			log.Printf("[INFO] Deleting EC2 Transit Gateway Connect: %s", id)
			_, err := conn.DeleteTransitGatewayConnect(input)

			if isAWSErr(err, "InvalidTransitGatewayAttachmentID.NotFound", "") {
				continue
			}

			if err != nil {
				return fmt.Errorf("error deleting EC2 Transit Gateway Connect (%s): %s", id, err)
			}

			if err := waitForEc2TransitGatewayConnectDeletion(conn, id); err != nil {
				return fmt.Errorf("error waiting for EC2 Transit Gateway Connect (%s) deletion: %s", id, err)
			}
		}

		if aws.StringValue(output.NextToken) == "" {
			break
		}

		input.NextToken = output.NextToken
	}

	return nil
}

func TestAccAWSEc2TransitGatewayConnect_basic(t *testing.T) {
	var transitGatewayConnect ec2.TransitGatewayConnect
	resourceName := "aws_ec2_transit_gateway_connect.test"
	transportTransitGatewayResourceName := "aws_ec2_transit_gateway_vpc_attachment.test"
	transitGatewayResourceName := "aws_ec2_transit_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEc2TransitGateway(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEc2TransitGatewayConnectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEc2TransitGatewayConnectConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEc2TransitGatewayConnectExists(resourceName, &transitGatewayConnect),
					resource.TestCheckResourceAttr(resourceName, "protocol", ec2.ProtocolValueGre),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_default_route_table_association", "true"),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_default_route_table_propagation", "true"),
					resource.TestCheckResourceAttrPair(resourceName, "transport_transit_gateway_attachment_id", transportTransitGatewayResourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "transit_gateway_id", transitGatewayResourceName, "id"),
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

func TestAccAWSEc2TransitGatewayConnect_disappears(t *testing.T) {
	var transitGatewayConnect ec2.TransitGatewayConnect
	resourceName := "aws_ec2_transit_gateway_connect.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEc2TransitGateway(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEc2TransitGatewayConnectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEc2TransitGatewayConnectConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEc2TransitGatewayConnectExists(resourceName, &transitGatewayConnect),
					testAccCheckAWSEc2TransitGatewayConnectDisappears(&transitGatewayConnect),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSEc2TransitGatewayConnect_Protocol(t *testing.T) {
	var transitGatewayConnect ec2.TransitGatewayConnect
	resourceName := "aws_ec2_transit_gateway_connect.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEc2TransitGateway(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEc2TransitGatewayConnectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEc2TransitGatewayConnectConfigProtocol(ec2.ProtocolValueGre),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEc2TransitGatewayConnectExists(resourceName, &transitGatewayConnect),
					resource.TestCheckResourceAttr(resourceName, "protocol", ec2.ProtocolValueGre),
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

func TestAccAWSEc2TransitGatewayConnect_Tags(t *testing.T) {
	var transitGatewayConnect1, transitGatewayConnect2, transitGatewayConnect3 ec2.TransitGatewayConnect
	resourceName := "aws_ec2_transit_gateway_connect.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEc2TransitGateway(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEc2TransitGatewayConnectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEc2TransitGatewayConnectConfigTags1("key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEc2TransitGatewayConnectExists(resourceName, &transitGatewayConnect1),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSEc2TransitGatewayConnectConfigTags2("key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEc2TransitGatewayConnectExists(resourceName, &transitGatewayConnect2),
					testAccCheckAWSEc2TransitGatewayConnectNotRecreated(&transitGatewayConnect1, &transitGatewayConnect2),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSEc2TransitGatewayConnectConfigTags1("key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEc2TransitGatewayConnectExists(resourceName, &transitGatewayConnect3),
					testAccCheckAWSEc2TransitGatewayConnectNotRecreated(&transitGatewayConnect1, &transitGatewayConnect3),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccAWSEc2TransitGatewayConnect_TransitGatewayDefaultRouteTableAssociationAndPropagationDisabled(t *testing.T) {
	var transitGateway1 ec2.TransitGateway
	var transitGatewayConnect1 ec2.TransitGatewayConnect
	resourceName := "aws_ec2_transit_gateway_connect.test"
	transitGatewayResourceName := "aws_ec2_transit_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEc2TransitGateway(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEc2TransitGatewayConnectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEc2TransitGatewayConnectConfigTransitGatewayDefaultRouteTableAssociationAndPropagationDisabled(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEc2TransitGatewayExists(transitGatewayResourceName, &transitGateway1),
					testAccCheckAWSEc2TransitGatewayConnectExists(resourceName, &transitGatewayConnect1),
					testAccCheckAWSEc2TransitGatewayAssociationDefaultRouteTableConnectNotAssociated(&transitGateway1, &transitGatewayConnect1),
					testAccCheckAWSEc2TransitGatewayPropagationDefaultRouteTableConnectNotPropagated(&transitGateway1, &transitGatewayConnect1),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_default_route_table_association", "false"),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_default_route_table_propagation", "false"),
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

func TestAccAWSEc2TransitGatewayConnect_TransitGatewayDefaultRouteTableAssociation(t *testing.T) {
	var transitGateway1, transitGateway2, transitGateway3 ec2.TransitGateway
	var transitGatewayConnect1, transitGatewayConnect2, transitGatewayConnect3 ec2.TransitGatewayConnect
	resourceName := "aws_ec2_transit_gateway_connect.test"
	transitGatewayResourceName := "aws_ec2_transit_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEc2TransitGateway(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEc2TransitGatewayConnectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEc2TransitGatewayConnectConfigTransitGatewayDefaultRouteTableAssociation(false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEc2TransitGatewayExists(transitGatewayResourceName, &transitGateway1),
					testAccCheckAWSEc2TransitGatewayConnectExists(resourceName, &transitGatewayConnect1),
					testAccCheckAWSEc2TransitGatewayAssociationDefaultRouteTableConnectNotAssociated(&transitGateway1, &transitGatewayConnect1),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_default_route_table_association", "false"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSEc2TransitGatewayConnectConfigTransitGatewayDefaultRouteTableAssociation(true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEc2TransitGatewayExists(transitGatewayResourceName, &transitGateway2),
					testAccCheckAWSEc2TransitGatewayConnectExists(resourceName, &transitGatewayConnect2),
					testAccCheckAWSEc2TransitGatewayConnectNotRecreated(&transitGatewayConnect1, &transitGatewayConnect2),
					testAccCheckAWSEc2TransitGatewayAssociationDefaultRouteTableConnectAssociated(&transitGateway2, &transitGatewayConnect2),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_default_route_table_association", "true"),
				),
			},
			{
				Config: testAccAWSEc2TransitGatewayConnectConfigTransitGatewayDefaultRouteTableAssociation(false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEc2TransitGatewayExists(transitGatewayResourceName, &transitGateway3),
					testAccCheckAWSEc2TransitGatewayConnectExists(resourceName, &transitGatewayConnect3),
					testAccCheckAWSEc2TransitGatewayConnectNotRecreated(&transitGatewayConnect2, &transitGatewayConnect3),
					testAccCheckAWSEc2TransitGatewayAssociationDefaultRouteTableConnectNotAssociated(&transitGateway3, &transitGatewayConnect3),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_default_route_table_association", "false"),
				),
			},
		},
	})
}

func TestAccAWSEc2TransitGatewayConnect_TransitGatewayDefaultRouteTablePropagation(t *testing.T) {
	var transitGateway1, transitGateway2, transitGateway3 ec2.TransitGateway
	var transitGatewayConnect1, transitGatewayConnect2, transitGatewayConnect3 ec2.TransitGatewayConnect
	resourceName := "aws_ec2_transit_gateway_connect.test"
	transitGatewayResourceName := "aws_ec2_transit_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEc2TransitGateway(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEc2TransitGatewayConnectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEc2TransitGatewayConnectConfigTransitGatewayDefaultRouteTablePropagation(false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEc2TransitGatewayExists(transitGatewayResourceName, &transitGateway1),
					testAccCheckAWSEc2TransitGatewayConnectExists(resourceName, &transitGatewayConnect1),
					testAccCheckAWSEc2TransitGatewayPropagationDefaultRouteTableConnectNotPropagated(&transitGateway1, &transitGatewayConnect1),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_default_route_table_propagation", "false"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSEc2TransitGatewayConnectConfigTransitGatewayDefaultRouteTablePropagation(true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEc2TransitGatewayExists(transitGatewayResourceName, &transitGateway2),
					testAccCheckAWSEc2TransitGatewayConnectExists(resourceName, &transitGatewayConnect2),
					testAccCheckAWSEc2TransitGatewayConnectNotRecreated(&transitGatewayConnect1, &transitGatewayConnect2),
					testAccCheckAWSEc2TransitGatewayPropagationDefaultRouteTableConnectPropagated(&transitGateway2, &transitGatewayConnect2),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_default_route_table_propagation", "true"),
				),
			},
			{
				Config: testAccAWSEc2TransitGatewayConnectConfigTransitGatewayDefaultRouteTablePropagation(false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEc2TransitGatewayExists(transitGatewayResourceName, &transitGateway3),
					testAccCheckAWSEc2TransitGatewayConnectExists(resourceName, &transitGatewayConnect3),
					testAccCheckAWSEc2TransitGatewayConnectNotRecreated(&transitGatewayConnect2, &transitGatewayConnect3),
					testAccCheckAWSEc2TransitGatewayPropagationDefaultRouteTableConnectNotPropagated(&transitGateway3, &transitGatewayConnect3),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_default_route_table_propagation", "false"),
				),
			},
		},
	})
}

func testAccCheckAWSEc2TransitGatewayConnectExists(resourceName string, transitGatewayConnect *ec2.TransitGatewayConnect) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EC2 Transit Gateway Connect ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).ec2conn

		attachment, err := ec2DescribeTransitGatewayConnect(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		if attachment == nil {
			return fmt.Errorf("EC2 Transit Gateway Connect not found")
		}

		if aws.StringValue(attachment.State) != ec2.TransitGatewayAttachmentStateAvailable {
			return fmt.Errorf("EC2 Transit Gateway Connect (%s) exists in non-available (%s) state", rs.Primary.ID, aws.StringValue(attachment.State))
		}

		*transitGatewayConnect = *attachment

		return nil
	}
}

func testAccCheckAWSEc2TransitGatewayConnectDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ec2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ec2_transit_gateway_route_table" {
			continue
		}

		connectAttachment, err := ec2DescribeTransitGatewayConnect(conn, rs.Primary.ID)

		if isAWSErr(err, "InvalidTransitGatewayAttachmentID.NotFound", "") {
			continue
		}

		if err != nil {
			return err
		}

		if connectAttachment == nil {
			continue
		}

		if aws.StringValue(connectAttachment.State) != ec2.TransitGatewayAttachmentStateDeleted {
			return fmt.Errorf("EC2 Transit Gateway Connect (%s) still exists in non-deleted (%s) state", rs.Primary.ID, aws.StringValue(connectAttachment.State))
		}
	}

	return nil
}

func testAccCheckAWSEc2TransitGatewayConnectDisappears(transitGatewayConnect *ec2.TransitGatewayConnect) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).ec2conn

		input := &ec2.DeleteTransitGatewayConnectInput{
			TransitGatewayAttachmentId: transitGatewayConnect.TransitGatewayAttachmentId,
		}

		if _, err := conn.DeleteTransitGatewayConnect(input); err != nil {
			return err
		}

		return waitForEc2TransitGatewayConnectDeletion(conn, aws.StringValue(transitGatewayConnect.TransitGatewayAttachmentId))
	}
}

func testAccCheckAWSEc2TransitGatewayConnectNotRecreated(i, j *ec2.TransitGatewayConnect) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(i.TransitGatewayAttachmentId) != aws.StringValue(j.TransitGatewayAttachmentId) {
			return errors.New("EC2 Transit Gateway Connect was recreated")
		}

		return nil
	}
}

func testAccAWSEc2TransitGatewayConnectConfig() string {
	return composeConfig(testAccAvailableAZsNoOptInDefaultExcludeConfig(), `
resource "aws_vpc" "test" {
	cidr_block = "10.0.0.0/16"
	
	tags = {
		Name = "tf-acc-test-ec2-transit-gateway-connect"
	}
	}
	
	resource "aws_subnet" "test" {
	availability_zone = data.aws_availability_zones.available.names[0]
	cidr_block        = "10.0.0.0/24"
	vpc_id            = aws_vpc.test.id
	
	tags = {
		Name = "tf-acc-test-ec2-transit-gateway-connect"
	}
	}
	
	resource "aws_ec2_transit_gateway" "test" {}
	
	resource "aws_ec2_transit_gateway_vpc_attachment" "test" {
	subnet_ids         = [aws_subnet.test.id]
	transit_gateway_id = aws_ec2_transit_gateway.test.id
	vpc_id             = aws_vpc.test.id
	}

	resource "aws_ec2_transit_gateway_connect" "test" {
		transport_transit_gateway_attachment_id = aws_ec2_transit_gateway_vpc_attachment.test.id
	}
`)
}

func testAccAWSEc2TransitGatewayConnectConfigProtocol(protocol string) string {
	return composeConfig(testAccAvailableAZsNoOptInDefaultExcludeConfig(), fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "tf-acc-test-ec2-transit-gateway-connect"
  }
}

resource "aws_subnet" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "10.0.0.0/24"
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = "tf-acc-test-ec2-transit-gateway-vpc-attachment"
  }
}

resource "aws_ec2_transit_gateway" "test" {
}

resource "aws_ec2_transit_gateway_vpc_attachment" "test" {
  subnet_ids         = [aws_subnet.test.id]
  transit_gateway_id = aws_ec2_transit_gateway.test.id
  vpc_id             = aws_vpc.test.id
}

resource "aws_ec2_transit_gateway_connect" "test" {
	transport_transit_gateway_attachment_id = aws_ec2_transit_gateway_vpc_attachment.test.id
	protocol = %[1]q
}

`, protocol))
}

func testAccAWSEc2TransitGatewayConnectConfigTags1(tagKey1, tagValue1 string) string {
	return composeConfig(testAccAvailableAZsNoOptInDefaultExcludeConfig(), fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "tf-acc-test-ec2-transit-gateway-connect"
  }
}

resource "aws_subnet" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "10.0.0.0/24"
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = "tf-acc-test-ec2-transit-gateway-connect"
  }
}

resource "aws_ec2_transit_gateway" "test" {}

resource "aws_ec2_transit_gateway_vpc_attachment" "test" {
  subnet_ids         = [aws_subnet.test.id]
  transit_gateway_id = aws_ec2_transit_gateway.test.id
  vpc_id             = aws_vpc.test.id
}
resource "aws_ec2_transit_gateway_connect" "test" {
	transport_transit_gateway_attachment_id = aws_ec2_transit_gateway_vpc_attachment.test.id
	tags = {
		%[1]q = %[2]q
	}
}
`, tagKey1, tagValue1))
}

func testAccAWSEc2TransitGatewayConnectConfigTags2(tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return composeConfig(testAccAvailableAZsNoOptInDefaultExcludeConfig(), fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "tf-acc-test-ec2-transit-gateway-connect"
  }
}

resource "aws_subnet" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "10.0.0.0/24"
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = "tf-acc-test-ec2-transit-gateway-connect"
  }
}

resource "aws_ec2_transit_gateway" "test" {}

resource "aws_ec2_transit_gateway_vpc_attachment" "test" {
  subnet_ids         = [aws_subnet.test.id]
  transit_gateway_id = aws_ec2_transit_gateway.test.id
  vpc_id             = aws_vpc.test.id
}

resource "aws_ec2_transit_gateway_connect" "test" {
	transport_transit_gateway_attachment_id = aws_ec2_transit_gateway_vpc_attachment.test.id
	tags = {
		%[1]q = %[2]q
		%[3]q = %[4]q
	}
}

`, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccAWSEc2TransitGatewayConnectConfigTransitGatewayDefaultRouteTableAssociationAndPropagationDisabled() string {
	return composeConfig(testAccAvailableAZsNoOptInDefaultExcludeConfig(), `
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "tf-acc-test-ec2-transit-gateway-connect"
  }
}

resource "aws_subnet" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "10.0.0.0/24"
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = "tf-acc-test-ec2-transit-gateway-connect"
  }
}

resource "aws_ec2_transit_gateway" "test" {
  default_route_table_association = "disable"
  default_route_table_propagation = "disable"
}

resource "aws_ec2_transit_gateway_vpc_attachment" "test" {
  subnet_ids                                      = [aws_subnet.test.id]
  transit_gateway_id                              = aws_ec2_transit_gateway.test.id
  vpc_id                                          = aws_vpc.test.id
  transit_gateway_default_route_table_association = false
  transit_gateway_default_route_table_propagation = false
}

resource "aws_ec2_transit_gateway_connect" "test" {
	transport_transit_gateway_attachment_id = aws_ec2_transit_gateway_vpc_attachment.test.id
	transit_gateway_default_route_table_association = false
  	transit_gateway_default_route_table_propagation = false
}
`)
}

func testAccAWSEc2TransitGatewayConnectConfigTransitGatewayDefaultRouteTableAssociation(transitGatewayDefaultRouteTableAssociation bool) string {
	return composeConfig(testAccAvailableAZsNoOptInDefaultExcludeConfig(), fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "tf-acc-test-ec2-transit-gateway-connect"
  }
}

resource "aws_subnet" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "10.0.0.0/24"
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = "tf-acc-test-ec2-transit-gateway-connect"
  }
}

resource "aws_ec2_transit_gateway" "test" {
}

resource "aws_ec2_transit_gateway_vpc_attachment" "test" {
  subnet_ids                                      = [aws_subnet.test.id]
  transit_gateway_id                              = aws_ec2_transit_gateway.test.id
  vpc_id                                          = aws_vpc.test.id
}

resource "aws_ec2_transit_gateway_connect" "test" {
	transport_transit_gateway_attachment_id = aws_ec2_transit_gateway_vpc_attachment.test.id
	transit_gateway_default_route_table_association = %[1]t
	tags = {
		Name = "tf-acc-test-ec2-transit-gateway-connect-assoc"
	}
}
`, transitGatewayDefaultRouteTableAssociation))
}

func testAccAWSEc2TransitGatewayConnectConfigTransitGatewayDefaultRouteTablePropagation(transitGatewayDefaultRouteTablePropagation bool) string {
	return composeConfig(testAccAvailableAZsNoOptInDefaultExcludeConfig(), fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "tf-acc-test-ec2-transit-gateway-connect"
  }
}

resource "aws_subnet" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "10.0.0.0/24"
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = "tf-acc-test-ec2-transit-gateway-connect"
  }
}

resource "aws_ec2_transit_gateway" "test" {
}

resource "aws_ec2_transit_gateway_vpc_attachment" "test" {
  subnet_ids                                      = [aws_subnet.test.id]
  transit_gateway_id                              = aws_ec2_transit_gateway.test.id
  vpc_id                                          = aws_vpc.test.id
}

resource "aws_ec2_transit_gateway_connect" "test" {
	transport_transit_gateway_attachment_id = aws_ec2_transit_gateway_vpc_attachment.test.id
	transit_gateway_default_route_table_propagation = %[1]t
}
`, transitGatewayDefaultRouteTablePropagation))
}
