package aws

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("aws_ec2_transit_gateway_connect_peer", &resource.Sweeper{
		Name: "aws_ec2_transit_gateway_connect_peer",
		F:    testSweepEc2TransitGatewayConnectPeer,
	})
}

func testSweepEc2TransitGatewayConnectPeer(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).ec2conn
	input := &ec2.DescribeTransitGatewayConnectPeersInput{}

	for {
		output, err := conn.DescribeTransitGatewayConnectPeers(input)

		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping EC2 Transit Gateway Connect Peer sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error retrieving EC2 Transit Gateway Connect Peers: %s", err)
		}

		for _, peer := range output.TransitGatewayConnectPeers {
			if aws.StringValue(peer.State) == ec2.TransitGatewayConnectPeerStateDeleted {
				continue
			}

			id := aws.StringValue(peer.TransitGatewayConnectPeerId)

			input := &ec2.DeleteTransitGatewayConnectPeerInput{
				TransitGatewayConnectPeerId: aws.String(id),
			}

			log.Printf("[INFO] Deleting EC2 Transit Gateway Connect Peer: %s", id)
			_, err := conn.DeleteTransitGatewayConnectPeer(input)

			if err != nil {
				return fmt.Errorf("error deleting EC2 Transit Gateway Connect Peer (%s): %s", id, err)
			}

			if err := waitForEc2TransitGatewayConnectPeerDeletion(conn, id); err != nil {
				return fmt.Errorf("error waiting for EC2 Transit Gateway Connect Peer (%s) deletion: %s", id, err)
			}
		}

		if aws.StringValue(output.NextToken) == "" {
			break
		}

		input.NextToken = output.NextToken
	}

	return nil
}

func TestAccAWSEc2TransitGatewayConnectPeer_basic(t *testing.T) {
	var transitGatewayConnectPeer ec2.TransitGatewayConnectPeer
	resourceName := "aws_ec2_transit_gateway_connect_peer.test"
	transitGatewayConnectResourceName := "aws_ec2_transit_gateway_connect.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEc2TransitGateway(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEc2TransitGatewayConnectPeerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEc2TransitGatewayConnectPeerConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEc2TransitGatewayConnectPeerExists(resourceName, &transitGatewayConnectPeer),
					resource.TestCheckResourceAttr(resourceName, "peer_asn", "64512"),
					resource.TestCheckTypeSetElemAttr(resourceName, "inside_cidr_blocks.*", "169.254.10.0/29"),
					resource.TestCheckResourceAttr(resourceName, "peer_address", "10.0.0.4"),
					resource.TestMatchResourceAttr(resourceName, "transit_gateway_address", regexp.MustCompile("^10.100.0")),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttrPair(resourceName, "transit_gateway_attachment_id", transitGatewayConnectResourceName, "id"),
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

func TestAccAWSEc2TransitGatewayConnectPeer_PeerAsn(t *testing.T) {
	var transitGatewayConnectPeer ec2.TransitGatewayConnectPeer
	resourceName := "aws_ec2_transit_gateway_connect_peer.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEc2TransitGateway(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEc2TransitGatewayConnectPeerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEc2TransitGatewayConnectPeerConfigPeerAsn("64515"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEc2TransitGatewayConnectPeerExists(resourceName, &transitGatewayConnectPeer),
					resource.TestCheckResourceAttr(resourceName, "peer_asn", "64515"),
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

func TestAccAWSEc2TransitGatewayConnectPeer_InsideCidrBlocks(t *testing.T) {
	var transitGatewayConnectPeer ec2.TransitGatewayConnectPeer
	resourceName := "aws_ec2_transit_gateway_connect_peer.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEc2TransitGateway(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEc2TransitGatewayConnectPeerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEc2TransitGatewayConnectPeerConfigInsideCidrBlocks("169.254.10.0/29", "fd00::/125"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEc2TransitGatewayConnectPeerExists(resourceName, &transitGatewayConnectPeer),
					resource.TestCheckTypeSetElemAttr(resourceName, "inside_cidr_blocks.*", "169.254.10.0/29"),
					resource.TestCheckTypeSetElemAttr(resourceName, "inside_cidr_blocks.*", "fd00::/125"),
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

func TestAccAWSEc2TransitGatewayConnectPeer_PeerAddress(t *testing.T) {
	var transitGatewayConnectPeer ec2.TransitGatewayConnectPeer
	resourceName := "aws_ec2_transit_gateway_connect_peer.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEc2TransitGateway(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEc2TransitGatewayConnectPeerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEc2TransitGatewayConnectPeerConfigPeerAddress("10.0.0.44"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEc2TransitGatewayConnectPeerExists(resourceName, &transitGatewayConnectPeer),
					resource.TestCheckResourceAttr(resourceName, "peer_address", "10.0.0.44"),
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

func TestAccAWSEc2TransitGatewayConnectPeer_TransitGatewayAddress(t *testing.T) {
	var transitGatewayConnectPeer ec2.TransitGatewayConnectPeer
	resourceName := "aws_ec2_transit_gateway_connect_peer.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEc2TransitGateway(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEc2TransitGatewayConnectPeerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEc2TransitGatewayConnectPeerConfigTransitGatewayAddress("10.100.0.100"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEc2TransitGatewayConnectPeerExists(resourceName, &transitGatewayConnectPeer),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_address", "10.100.0.100"),
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

func TestAccAWSEc2TransitGatewayConnectPeer_disappears(t *testing.T) {
	var transitGatewayConnectPeer ec2.TransitGatewayConnectPeer
	resourceName := "aws_ec2_transit_gateway_connect_peer.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEc2TransitGateway(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEc2TransitGatewayConnectPeerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEc2TransitGatewayConnectPeerConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEc2TransitGatewayConnectPeerExists(resourceName, &transitGatewayConnectPeer),
					testAccCheckAWSEc2TransitGatewayConnectPeerDisappears(&transitGatewayConnectPeer),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSEc2TransitGatewayConnectPeer_Tags(t *testing.T) {
	var transitGatewayConnectPeer1, transitGatewayConnectPeer2, transitGatewayConnectPeer3 ec2.TransitGatewayConnectPeer
	resourceName := "aws_ec2_transit_gateway_connect_peer.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEc2TransitGateway(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEc2TransitGatewayConnectPeerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEc2TransitGatewayConnectPeerConfigTags1("key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEc2TransitGatewayConnectPeerExists(resourceName, &transitGatewayConnectPeer1),
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
				Config: testAccAWSEc2TransitGatewayConnectPeerConfigTags2("key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEc2TransitGatewayConnectPeerExists(resourceName, &transitGatewayConnectPeer2),
					testAccCheckAWSEc2TransitGatewayConnectPeerNotRecreated(&transitGatewayConnectPeer1, &transitGatewayConnectPeer2),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSEc2TransitGatewayConnectPeerConfigTags1("key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEc2TransitGatewayConnectPeerExists(resourceName, &transitGatewayConnectPeer3),
					testAccCheckAWSEc2TransitGatewayConnectPeerNotRecreated(&transitGatewayConnectPeer1, &transitGatewayConnectPeer3),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckAWSEc2TransitGatewayConnectPeerExists(resourceName string, transitGatewayConnectPeer *ec2.TransitGatewayConnectPeer) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EC2 Transit Gateway Connect Peer ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).ec2conn

		peer, err := ec2DescribeTransitGatewayConnectPeer(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		if peer == nil {
			return fmt.Errorf("EC2 Transit Gateway Connect Peer not found")
		}

		if aws.StringValue(peer.State) != ec2.TransitGatewayConnectPeerStateAvailable {
			return fmt.Errorf("EC2 Transit Gateway Connect Peer (%s) exists in non-available (%s) state", rs.Primary.ID, aws.StringValue(peer.State))
		}

		*transitGatewayConnectPeer = *peer

		return nil
	}
}

func testAccCheckAWSEc2TransitGatewayConnectPeerDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ec2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ec2_transit_gateway_route_table" {
			continue
		}

		peer, err := ec2DescribeTransitGatewayConnectPeer(conn, rs.Primary.ID)

		if isAWSErr(err, "InvalidTransitGatewayConnectPeerID.NotFound", "") {
			continue
		}

		if err != nil {
			return err
		}

		if peer == nil {
			continue
		}

		if aws.StringValue(peer.State) != ec2.TransitGatewayConnectPeerStateDeleted {
			return fmt.Errorf("EC2 Transit Gateway Connect (%s) still exists in non-deleted (%s) state", rs.Primary.ID, aws.StringValue(peer.State))
		}
	}

	return nil
}

func testAccCheckAWSEc2TransitGatewayConnectPeerDisappears(transitGatewayConnectPeer *ec2.TransitGatewayConnectPeer) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).ec2conn

		input := &ec2.DeleteTransitGatewayConnectPeerInput{
			TransitGatewayConnectPeerId: transitGatewayConnectPeer.TransitGatewayConnectPeerId,
		}

		if _, err := conn.DeleteTransitGatewayConnectPeer(input); err != nil {
			return err
		}

		return waitForEc2TransitGatewayConnectPeerDeletion(conn, aws.StringValue(transitGatewayConnectPeer.TransitGatewayConnectPeerId))
	}
}

func testAccCheckAWSEc2TransitGatewayConnectPeerNotRecreated(i, j *ec2.TransitGatewayConnectPeer) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(i.TransitGatewayConnectPeerId) != aws.StringValue(j.TransitGatewayConnectPeerId) {
			return errors.New("EC2 Transit Gateway Connect Peer was recreated")
		}

		return nil
	}
}

func testAccAWSEc2TransitGatewayConnectPeerConfig() string {
	return composeConfig(testAccAvailableAZsNoOptInDefaultExcludeConfig(), `
resource "aws_vpc" "test" {
	cidr_block = "10.0.0.0/16"
	
	tags = {
		Name = "tf-acc-test-ec2-transit-gateway-connect-peer"
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
		transit_gateway_cidr_blocks = ["10.100.0.0/24"]
	}
	
	resource "aws_ec2_transit_gateway_vpc_attachment" "test" {
	subnet_ids         = [aws_subnet.test.id]
	transit_gateway_id = aws_ec2_transit_gateway.test.id
	vpc_id             = aws_vpc.test.id
	}

	resource "aws_ec2_transit_gateway_connect" "test" {
		transport_transit_gateway_attachment_id = aws_ec2_transit_gateway_vpc_attachment.test.id
	}

	resource "aws_ec2_transit_gateway_connect_peer" "test" {
		transit_gateway_attachment_id = aws_ec2_transit_gateway_connect.test.id
		inside_cidr_blocks = ["169.254.10.0/29"]
		peer_address       = "10.0.0.4"
	}

`)
}

func testAccAWSEc2TransitGatewayConnectPeerConfigPeerAsn(peerAsn string) string {
	return composeConfig(testAccAvailableAZsNoOptInDefaultExcludeConfig(), fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "tf-acc-test-ec2-transit-gateway-connect-peer"
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
	transit_gateway_cidr_blocks = ["10.100.0.0/24"]
}

resource "aws_ec2_transit_gateway_vpc_attachment" "test" {
  subnet_ids         = [aws_subnet.test.id]
  transit_gateway_id = aws_ec2_transit_gateway.test.id
  vpc_id             = aws_vpc.test.id
}

resource "aws_ec2_transit_gateway_connect" "test" {
	transport_transit_gateway_attachment_id = aws_ec2_transit_gateway_vpc_attachment.test.id
}

resource "aws_ec2_transit_gateway_connect_peer" "test" {
	transit_gateway_attachment_id = aws_ec2_transit_gateway_connect.test.id
	peer_asn = %[1]q
	inside_cidr_blocks = ["169.254.10.0/29"]
	peer_address       = "10.0.0.4"
}

`, peerAsn))
}

func testAccAWSEc2TransitGatewayConnectPeerConfigInsideCidrBlocks(insideCidrBlock1, insideCidrBlock2 string) string {
	return composeConfig(testAccAvailableAZsNoOptInDefaultExcludeConfig(), fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "tf-acc-test-ec2-transit-gateway-connect-peer"
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
	transit_gateway_cidr_blocks = ["10.100.0.0/24"]
}

resource "aws_ec2_transit_gateway_vpc_attachment" "test" {
  subnet_ids         = [aws_subnet.test.id]
  transit_gateway_id = aws_ec2_transit_gateway.test.id
  vpc_id             = aws_vpc.test.id
}

resource "aws_ec2_transit_gateway_connect" "test" {
	transport_transit_gateway_attachment_id = aws_ec2_transit_gateway_vpc_attachment.test.id
}

resource "aws_ec2_transit_gateway_connect_peer" "test" {
	transit_gateway_attachment_id = aws_ec2_transit_gateway_connect.test.id
	peer_address       = "10.0.0.4"
	inside_cidr_blocks = [%[1]q, %[2]q]
}

`, insideCidrBlock1, insideCidrBlock2))
}

func testAccAWSEc2TransitGatewayConnectPeerConfigPeerAddress(peerAddress string) string {
	return composeConfig(testAccAvailableAZsNoOptInDefaultExcludeConfig(), fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "tf-acc-test-ec2-transit-gateway-connect-peer"
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
	transit_gateway_cidr_blocks = ["10.100.0.0/24"]
}

resource "aws_ec2_transit_gateway_vpc_attachment" "test" {
  subnet_ids         = [aws_subnet.test.id]
  transit_gateway_id = aws_ec2_transit_gateway.test.id
  vpc_id             = aws_vpc.test.id
}

resource "aws_ec2_transit_gateway_connect" "test" {
	transport_transit_gateway_attachment_id = aws_ec2_transit_gateway_vpc_attachment.test.id
}

resource "aws_ec2_transit_gateway_connect_peer" "test" {
	transit_gateway_attachment_id = aws_ec2_transit_gateway_connect.test.id
	inside_cidr_blocks = ["169.254.10.0/29"]
	peer_address = %[1]q
}

`, peerAddress))
}

func testAccAWSEc2TransitGatewayConnectPeerConfigTransitGatewayAddress(transitGatewayAddress string) string {
	return composeConfig(testAccAvailableAZsNoOptInDefaultExcludeConfig(), fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "tf-acc-test-ec2-transit-gateway-connect-peer"
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
	transit_gateway_cidr_blocks = ["10.100.0.0/24"]
}

resource "aws_ec2_transit_gateway_vpc_attachment" "test" {
  subnet_ids         = [aws_subnet.test.id]
  transit_gateway_id = aws_ec2_transit_gateway.test.id
  vpc_id             = aws_vpc.test.id
}

resource "aws_ec2_transit_gateway_connect" "test" {
	transport_transit_gateway_attachment_id = aws_ec2_transit_gateway_vpc_attachment.test.id
}

resource "aws_ec2_transit_gateway_connect_peer" "test" {
	transit_gateway_attachment_id = aws_ec2_transit_gateway_connect.test.id
	inside_cidr_blocks = ["169.254.10.0/29"]
	peer_address       = "10.0.0.4"
	transit_gateway_address = %[1]q
}

`, transitGatewayAddress))
}

func testAccAWSEc2TransitGatewayConnectPeerConfigTags1(tagKey1, tagValue1 string) string {
	return composeConfig(testAccAvailableAZsNoOptInDefaultExcludeConfig(), fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "tf-acc-test-ec2-transit-gateway-connect-peer"
  }
}

resource "aws_subnet" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "10.0.0.0/24"
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = "tf-acc-test-ec2-transit-gateway-connect-peer"
  }
}

resource "aws_ec2_transit_gateway" "test" {
	transit_gateway_cidr_blocks = ["10.100.0.0/24"]
}

resource "aws_ec2_transit_gateway_vpc_attachment" "test" {
  subnet_ids         = [aws_subnet.test.id]
  transit_gateway_id = aws_ec2_transit_gateway.test.id
  vpc_id             = aws_vpc.test.id
}
resource "aws_ec2_transit_gateway_connect" "test" {
	transport_transit_gateway_attachment_id = aws_ec2_transit_gateway_vpc_attachment.test.id
}

resource "aws_ec2_transit_gateway_connect_peer" "test" {
	transit_gateway_attachment_id = aws_ec2_transit_gateway_connect.test.id
	inside_cidr_blocks = ["169.254.10.0/29"]
	peer_address       = "10.0.0.4"
	tags = {
		%[1]q = %[2]q
	}
}

`, tagKey1, tagValue1))
}

func testAccAWSEc2TransitGatewayConnectPeerConfigTags2(tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return composeConfig(testAccAvailableAZsNoOptInDefaultExcludeConfig(), fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "tf-acc-test-ec2-transit-gateway-connect-peer"
  }
}

resource "aws_subnet" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "10.0.0.0/24"
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = "tf-acc-test-ec2-transit-gateway-connect-peer"
  }
}

resource "aws_ec2_transit_gateway" "test" {
	transit_gateway_cidr_blocks = ["10.100.0.0/24"]
}

resource "aws_ec2_transit_gateway_vpc_attachment" "test" {
  subnet_ids         = [aws_subnet.test.id]
  transit_gateway_id = aws_ec2_transit_gateway.test.id
  vpc_id             = aws_vpc.test.id
}

resource "aws_ec2_transit_gateway_connect" "test" {
	transport_transit_gateway_attachment_id = aws_ec2_transit_gateway_vpc_attachment.test.id
}

resource "aws_ec2_transit_gateway_connect_peer" "test" {
	transit_gateway_attachment_id = aws_ec2_transit_gateway_connect.test.id
	inside_cidr_blocks = ["169.254.10.0/29"]
	peer_address       = "10.0.0.4"
	tags = {
		%[1]q = %[2]q
		%[3]q = %[4]q
	}
}

`, tagKey1, tagValue1, tagKey2, tagValue2))
}
