package ec2_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
)

func testAccTransitGatewayConnectPeer_basic(t *testing.T) {
	var transitGatewayConnectPeer1 ec2.TransitGatewayConnectPeer
	resourceName := "aws_ec2_transit_gateway_connect_peer.test"
	transitGatewayConnectResourceName := "aws_ec2_transit_gateway_connect.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckTransitGatewayConnect(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckTransitGatewayConnectPeerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayConnectPeerConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayConnectPeerExists(resourceName, &transitGatewayConnectPeer1),
					resource.TestCheckResourceAttr(resourceName, "bgp_asn", "64512"),
					resource.TestCheckResourceAttr(resourceName, "inside_cidr_blocks.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "peer_address", "1.1.1.1"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "transit_gateway_address"),
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

func testAccTransitGatewayConnectPeer_disappears(t *testing.T) {
	var transitGatewayConnectPeer1 ec2.TransitGatewayConnectPeer
	resourceName := "aws_ec2_transit_gateway_connect_peer.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckTransitGatewayConnect(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckTransitGatewayConnectPeerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayConnectPeerConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayConnectPeerExists(resourceName, &transitGatewayConnectPeer1),
					testAccCheckTransitGatewayConnectPeerDisappears(&transitGatewayConnectPeer1),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccTransitGatewayConnectPeer_BgpAsn(t *testing.T) {
	var transitGatewayConnectPeer1 ec2.TransitGatewayConnectPeer
	resourceName := "aws_ec2_transit_gateway_connect_peer.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckTransitGatewayConnect(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckTransitGatewayConnectPeerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayConnectPeerBgpAsnConfig(12345),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayConnectPeerExists(resourceName, &transitGatewayConnectPeer1),
					resource.TestCheckResourceAttr(resourceName, "bgp_asn", "12345"),
				),
			},
		},
	})
}

func testAccTransitGatewayConnectPeer_InsideCidrBlocks(t *testing.T) {
	var transitGatewayConnectPeer1, transitGatewayConnectPeer2, transitGatewayConnectPeer3 ec2.TransitGatewayConnectPeer
	resourceName := "aws_ec2_transit_gateway_connect_peer.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckTransitGatewayConnect(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckTransitGatewayConnectPeerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayConnectPeerInsideCidrBlocks2Config(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayConnectPeerExists(resourceName, &transitGatewayConnectPeer1),
					resource.TestCheckResourceAttr(resourceName, "inside_cidr_blocks.#", "2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTransitGatewayConnectPeerConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayConnectPeerExists(resourceName, &transitGatewayConnectPeer2),
					testAccCheckTransitGatewayConnectPeerNotRecreated(&transitGatewayConnectPeer1, &transitGatewayConnectPeer2),
					resource.TestCheckResourceAttr(resourceName, "inside_cidr_blocks.#", "1"),
				),
			},
			{
				Config: testAccTransitGatewayConnectPeerInsideCidrBlocks2Config(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayConnectPeerExists(resourceName, &transitGatewayConnectPeer3),
					testAccCheckTransitGatewayConnectPeerNotRecreated(&transitGatewayConnectPeer2, &transitGatewayConnectPeer3),
					resource.TestCheckResourceAttr(resourceName, "inside_cidr_blocks.#", "2"),
				),
			},
		},
	})
}

func testAccTransitGatewayConnectPeer_Tags(t *testing.T) {
	var transitGatewayConnectPeer1, transitGatewayConnectPeer2, transitGatewayConnectPeer3 ec2.TransitGatewayConnectPeer
	resourceName := "aws_ec2_transit_gateway_connect_peer.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckTransitGatewayConnect(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckTransitGatewayConnectPeerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayConnectPeerTags1Config("key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayConnectPeerExists(resourceName, &transitGatewayConnectPeer1),
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
				Config: testAccTransitGatewayConnectPeerTags2Config("key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayConnectPeerExists(resourceName, &transitGatewayConnectPeer2),
					testAccCheckTransitGatewayConnectPeerNotRecreated(&transitGatewayConnectPeer1, &transitGatewayConnectPeer2),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccTransitGatewayConnectPeerTags1Config("key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayConnectPeerExists(resourceName, &transitGatewayConnectPeer3),
					testAccCheckTransitGatewayConnectPeerNotRecreated(&transitGatewayConnectPeer2, &transitGatewayConnectPeer3),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccTransitGatewayConnectPeer_TransitGatewayAddress(t *testing.T) {
	var transitGatewayConnectPeer1 ec2.TransitGatewayConnectPeer
	resourceName := "aws_ec2_transit_gateway_connect_peer.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckTransitGatewayConnect(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckTransitGatewayConnectPeerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayConnectPeerTransitGatewayAddressConfig("10.20.30.200"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayConnectPeerExists(resourceName, &transitGatewayConnectPeer1),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_address", "10.20.30.200"),
				),
			},
		},
	})
}

func testAccCheckTransitGatewayConnectPeerExists(resourceName string, transitGatewayConnectPeer *ec2.TransitGatewayConnectPeer) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EC2 Transit Gateway Connect Peer ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

		attachment, err := tfec2.DescribeTransitGatewayConnectPeer(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		if attachment == nil {
			return fmt.Errorf("EC2 Transit Gateway Connect Peer not found")
		}

		if aws.StringValue(attachment.State) != ec2.TransitGatewayAttachmentStateAvailable && aws.StringValue(attachment.State) != ec2.TransitGatewayAttachmentStatePendingAcceptance {
			return fmt.Errorf("EC2 Transit Gateway Connect Peer (%s) exists in non-available/pending acceptance (%s) state", rs.Primary.ID, aws.StringValue(attachment.State))
		}

		*transitGatewayConnectPeer = *attachment

		return nil
	}
}

func testAccCheckTransitGatewayConnectPeerDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ec2_transit_gateway_route_table" {
			continue
		}

		vpcAttachment, err := tfec2.DescribeTransitGatewayConnectPeer(conn, rs.Primary.ID)

		if tfawserr.ErrMessageContains(err, "InvalidTransitGatewayAttachmentID.NotFound", "") {
			continue
		}

		if err != nil {
			return err
		}

		if vpcAttachment == nil {
			continue
		}

		if aws.StringValue(vpcAttachment.State) != ec2.TransitGatewayAttachmentStateDeleted {
			return fmt.Errorf("EC2 Transit Gateway Connect Peer (%s) still exists in non-deleted (%s) state", rs.Primary.ID, aws.StringValue(vpcAttachment.State))
		}
	}

	return nil
}

func testAccCheckTransitGatewayConnectPeerDisappears(transitGatewayConnectPeer *ec2.TransitGatewayConnectPeer) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

		input := &ec2.DeleteTransitGatewayConnectPeerInput{
			TransitGatewayConnectPeerId: transitGatewayConnectPeer.TransitGatewayConnectPeerId,
		}

		if _, err := conn.DeleteTransitGatewayConnectPeer(input); err != nil {
			return err
		}

		return tfec2.WaitForTransitGatewayConnectPeerDeletion(conn, aws.StringValue(transitGatewayConnectPeer.TransitGatewayConnectPeerId))
	}
}

func testAccCheckTransitGatewayConnectPeerNotRecreated(i, j *ec2.TransitGatewayConnectPeer) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(i.TransitGatewayAttachmentId) != aws.StringValue(j.TransitGatewayAttachmentId) {
			return errors.New("EC2 Transit Gateway Connect Peer was recreated")
		}

		return nil
	}
}

func testAccTransitGatewayConnectPeerConfig() string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptInDefaultExclude(), `
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
  cidr_blocks = ["10.20.30.0/24"]
}

resource "aws_ec2_transit_gateway_vpc_attachment" "test" {
  subnet_ids         = [aws_subnet.test.id]
  transit_gateway_id = aws_ec2_transit_gateway.test.id
  vpc_id             = aws_vpc.test.id
}

resource "aws_ec2_transit_gateway_connect" "test" {
  transit_gateway_id      = aws_ec2_transit_gateway.test.id
  transport_attachment_id = aws_ec2_transit_gateway_vpc_attachment.test.id
}

resource "aws_ec2_transit_gateway_connect_peer" "test" {
  inside_cidr_blocks            = ["169.254.200.0/29"]
  peer_address                  = "1.1.1.1"
  transit_gateway_attachment_id = aws_ec2_transit_gateway_connect.test.id
}
`)
}

func testAccTransitGatewayConnectPeerBgpAsnConfig(bgpAsn int) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptInDefaultExclude(), fmt.Sprintf(`
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
  cidr_blocks = ["10.20.30.0/24"]
}

resource "aws_ec2_transit_gateway_vpc_attachment" "test" {
  subnet_ids         = [aws_subnet.test.id]
  transit_gateway_id = aws_ec2_transit_gateway.test.id
  vpc_id             = aws_vpc.test.id
}

resource "aws_ec2_transit_gateway_connect" "test" {
  transit_gateway_id      = aws_ec2_transit_gateway.test.id
  transport_attachment_id = aws_ec2_transit_gateway_vpc_attachment.test.id
}

resource "aws_ec2_transit_gateway_connect_peer" "test" {
  bgp_asn                       = %d
  inside_cidr_blocks            = ["169.254.200.0/29"]
  peer_address                  = "1.1.1.1"
  transit_gateway_attachment_id = aws_ec2_transit_gateway_connect.test.id
}
`, bgpAsn))
}

func testAccTransitGatewayConnectPeerTags1Config(tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptInDefaultExclude(), fmt.Sprintf(`
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
  cidr_blocks = ["10.20.30.0/24"]
}

resource "aws_ec2_transit_gateway_vpc_attachment" "test" {
  subnet_ids         = [aws_subnet.test.id]
  transit_gateway_id = aws_ec2_transit_gateway.test.id
  vpc_id             = aws_vpc.test.id
}

resource "aws_ec2_transit_gateway_connect" "test" {
  transit_gateway_id      = aws_ec2_transit_gateway.test.id
  transport_attachment_id = aws_ec2_transit_gateway_vpc_attachment.test.id
}

resource "aws_ec2_transit_gateway_connect_peer" "test" {
  inside_cidr_blocks            = ["169.254.200.0/29"]
  peer_address                  = "1.1.1.1"
  transit_gateway_attachment_id = aws_ec2_transit_gateway_connect.test.id

  tags = {
    %[1]q = %[2]q
  }
}
`, tagKey1, tagValue1))
}

func testAccTransitGatewayConnectPeerTags2Config(tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptInDefaultExclude(), fmt.Sprintf(`
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
  cidr_blocks = ["10.20.30.0/24"]
}

resource "aws_ec2_transit_gateway_vpc_attachment" "test" {
  subnet_ids         = [aws_subnet.test.id]
  transit_gateway_id = aws_ec2_transit_gateway.test.id
  vpc_id             = aws_vpc.test.id
}

resource "aws_ec2_transit_gateway_connect" "test" {
  transit_gateway_id      = aws_ec2_transit_gateway.test.id
  transport_attachment_id = aws_ec2_transit_gateway_vpc_attachment.test.id
}

resource "aws_ec2_transit_gateway_connect_peer" "test" {
  inside_cidr_blocks            = ["169.254.200.0/29"]
  peer_address                  = "1.1.1.1"
  transit_gateway_attachment_id = aws_ec2_transit_gateway_connect.test.id

  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccTransitGatewayConnectPeerInsideCidrBlocks2Config() string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptInDefaultExclude(), `
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
  cidr_blocks = ["10.20.30.0/24"]
}

resource "aws_ec2_transit_gateway_vpc_attachment" "test" {
  subnet_ids         = [aws_subnet.test.id]
  transit_gateway_id = aws_ec2_transit_gateway.test.id
  vpc_id             = aws_vpc.test.id
}

resource "aws_ec2_transit_gateway_connect" "test" {
  transit_gateway_id      = aws_ec2_transit_gateway.test.id
  transport_attachment_id = aws_ec2_transit_gateway_vpc_attachment.test.id
}

resource "aws_ec2_transit_gateway_connect_peer" "test" {
  inside_cidr_blocks            = ["169.254.200.0/29", "fd00::/125"]
  peer_address                  = "1.1.1.1"
  transit_gateway_attachment_id = aws_ec2_transit_gateway_connect.test.id
}
`)
}

func testAccTransitGatewayConnectPeerTransitGatewayAddressConfig(transitGatewayAddress string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptInDefaultExclude(), fmt.Sprintf(`
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
  cidr_blocks = ["10.20.30.0/24"]
}

resource "aws_ec2_transit_gateway_vpc_attachment" "test" {
  subnet_ids         = [aws_subnet.test.id]
  transit_gateway_id = aws_ec2_transit_gateway.test.id
  vpc_id             = aws_vpc.test.id
}

resource "aws_ec2_transit_gateway_connect" "test" {
  transit_gateway_id      = aws_ec2_transit_gateway.test.id
  transport_attachment_id = aws_ec2_transit_gateway_vpc_attachment.test.id
}

resource "aws_ec2_transit_gateway_connect_peer" "test" {
  transit_gateway_address       = %[1]q
  inside_cidr_blocks            = ["169.254.200.0/29"]
  peer_address                  = "1.1.1.1"
  transit_gateway_attachment_id = aws_ec2_transit_gateway_connect.test.id
}
`, transitGatewayAddress))
}
