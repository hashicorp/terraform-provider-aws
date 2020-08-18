package aws

import (
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("aws_vpn_connection", &resource.Sweeper{
		Name: "aws_vpn_connection",
		F:    testSweepEc2VpnConnections,
	})
}

func testSweepEc2VpnConnections(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).ec2conn
	input := &ec2.DescribeVpnConnectionsInput{}

	// DescribeVpnConnections does not currently have any form of pagination
	output, err := conn.DescribeVpnConnections(input)

	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping EC2 VPN Connection sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error retrieving EC2 VPN Connections: %s", err)
	}

	for _, vpnConnection := range output.VpnConnections {
		if aws.StringValue(vpnConnection.State) == ec2.VpnStateDeleted {
			continue
		}

		id := aws.StringValue(vpnConnection.VpnConnectionId)
		input := &ec2.DeleteVpnConnectionInput{
			VpnConnectionId: vpnConnection.VpnConnectionId,
		}

		log.Printf("[INFO] Deleting EC2 VPN Connection: %s", id)

		_, err := conn.DeleteVpnConnection(input)

		if isAWSErr(err, "InvalidVpnConnectionID.NotFound", "") {
			continue
		}

		if err != nil {
			return fmt.Errorf("error deleting EC2 VPN Connection (%s): %s", id, err)
		}

		if err := waitForEc2VpnConnectionDeletion(conn, id); err != nil {
			return fmt.Errorf("error waiting for VPN connection (%s) to delete: %s", id, err)
		}
	}

	return nil
}

func TestAccAWSVpnConnection_basic(t *testing.T) {
	rInt := acctest.RandInt()
	rBgpAsn := acctest.RandIntRange(64512, 65534)
	resourceName := "aws_vpn_connection.test"
	var vpn ec2.VpnConnection

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccAwsVpnConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsVpnConnectionConfig(rBgpAsn),
				Check: resource.ComposeTestCheckFunc(
					testAccAwsVpnConnectionExists(resourceName, &vpn),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_attachment_id", ""),
					resource.TestCheckResourceAttr(resourceName, "enable_acceleration", "false"),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "ec2", regexp.MustCompile(`vpn-connection/vpn-.+`)),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAwsVpnConnectionConfigUpdate(rInt, rBgpAsn),
				Check: resource.ComposeTestCheckFunc(
					testAccAwsVpnConnectionExists(resourceName, &vpn),
				),
			},
		},
	})
}

func TestAccAWSVpnConnection_TransitGatewayID(t *testing.T) {
	var vpn ec2.VpnConnection
	rBgpAsn := acctest.RandIntRange(64512, 65534)
	transitGatewayResourceName := "aws_ec2_transit_gateway.test"
	resourceName := "aws_vpn_connection.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckAWSEc2TransitGateway(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccAwsVpnConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsVpnConnectionConfigTransitGatewayID(rBgpAsn),
				Check: resource.ComposeTestCheckFunc(
					testAccAwsVpnConnectionExists(resourceName, &vpn),
					resource.TestMatchResourceAttr(resourceName, "transit_gateway_attachment_id", regexp.MustCompile(`tgw-attach-.+`)),
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

func TestAccAWSVpnConnection_tunnelOptions(t *testing.T) {
	badCidrRangeErr := regexp.MustCompile(`expected \w+ to not be any of \[[\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}\/30\s?]+\]`)
	rBgpAsn := acctest.RandIntRange(64512, 65534)
	resourceName := "aws_vpn_connection.test"
	var vpn ec2.VpnConnection

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccAwsVpnConnectionDestroy,
		Steps: []resource.TestStep{
			// Checking CIDR blocks
			{
				Config:      testAccAwsVpnConnectionConfigSingleTunnelOptions(rBgpAsn, "12345678", "not-a-cidr"),
				ExpectError: regexp.MustCompile(`invalid CIDR address: not-a-cidr`),
			},
			{
				Config:      testAccAwsVpnConnectionConfigSingleTunnelOptions(rBgpAsn, "12345678", "169.254.254.0/31"),
				ExpectError: regexp.MustCompile(`expected "\w+" to contain a network Value with between 30 and 30 significant bits`),
			},
			{
				Config:      testAccAwsVpnConnectionConfigSingleTunnelOptions(rBgpAsn, "12345678", "172.16.0.0/30"),
				ExpectError: regexp.MustCompile(`must be within 169.254.0.0/16`),
			},
			{
				Config:      testAccAwsVpnConnectionConfigSingleTunnelOptions(rBgpAsn, "12345678", "169.254.0.0/30"),
				ExpectError: badCidrRangeErr,
			},
			{
				Config:      testAccAwsVpnConnectionConfigSingleTunnelOptions(rBgpAsn, "12345678", "169.254.1.0/30"),
				ExpectError: badCidrRangeErr,
			},
			{
				Config:      testAccAwsVpnConnectionConfigSingleTunnelOptions(rBgpAsn, "12345678", "169.254.2.0/30"),
				ExpectError: badCidrRangeErr,
			},
			{
				Config:      testAccAwsVpnConnectionConfigSingleTunnelOptions(rBgpAsn, "12345678", "169.254.3.0/30"),
				ExpectError: badCidrRangeErr,
			},
			{
				Config:      testAccAwsVpnConnectionConfigSingleTunnelOptions(rBgpAsn, "12345678", "169.254.4.0/30"),
				ExpectError: badCidrRangeErr,
			},
			{
				Config:      testAccAwsVpnConnectionConfigSingleTunnelOptions(rBgpAsn, "12345678", "169.254.5.0/30"),
				ExpectError: badCidrRangeErr,
			},
			{
				Config:      testAccAwsVpnConnectionConfigSingleTunnelOptions(rBgpAsn, "12345678", "169.254.169.252/30"),
				ExpectError: badCidrRangeErr,
			},

			// Checking PreShared Key
			{
				Config:      testAccAwsVpnConnectionConfigSingleTunnelOptions(rBgpAsn, "1234567", "169.254.254.0/30"),
				ExpectError: regexp.MustCompile(`expected length of \w+ to be in the range \(8 - 64\)`),
			},
			{
				Config:      testAccAwsVpnConnectionConfigSingleTunnelOptions(rBgpAsn, acctest.RandStringFromCharSet(65, acctest.CharSetAlpha), "169.254.254.0/30"),
				ExpectError: regexp.MustCompile(`expected length of \w+ to be in the range \(8 - 64\)`),
			},
			{
				Config:      testAccAwsVpnConnectionConfigSingleTunnelOptions(rBgpAsn, "01234567", "169.254.254.0/30"),
				ExpectError: regexp.MustCompile(`cannot start with zero character`),
			},
			{
				Config:      testAccAwsVpnConnectionConfigSingleTunnelOptions(rBgpAsn, "1234567!", "169.254.254.0/30"),
				ExpectError: regexp.MustCompile(`can only contain alphanumeric, period and underscore characters`),
			},

			// Should pre-check:
			// - local_ipv4_network_cidr
			// - local_ipv6_network_cidr
			// - remote_ipv4_network_cidr
			// - remote_ipv6_network_cidr
			// - tunnel_inside_ip_version
			// - tunnel1_dpd_timeout_action
			// - tunnel1_dpd_timeout_seconds
			// - tunnel1_phase1_lifetime_seconds
			// - tunnel1_phase2_lifetime_seconds
			// - tunnel1_rekey_fuzz_percentage
			// - tunnel1_rekey_margin_time_seconds
			// - tunnel1_replay_window_size
			// - tunnel1_startup_action
			// - tunnel1_inside_cidr
			// - tunnel1_inside_ipv6_cidr

			//Try actual building
			{
				Config: testAccAwsVpnConnectionConfigTunnelOptions(
					rBgpAsn,
					"192.168.1.1/32",
					"192.168.1.2/32",
					"12345678",
					"169.254.8.0/30",
					"clear",
					30,
					"\"ikev1\", \"ikev2\"",
					"2, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24",
					"\"AES128\", \"AES256\", \"AES128-GCM-16\", \"AES256-GCM-16\"",
					"\"SHA1\", \"SHA2-256\", \"SHA2-384\", \"SHA2-512\"",
					28800,
					"2, 5, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24",
					"\"AES128\", \"AES256\", \"AES128-GCM-16\", \"AES256-GCM-16\"",
					"\"SHA1\", \"SHA2-256\", \"SHA2-384\", \"SHA2-512\"",
					3600,
					100,
					540,
					1024,
					"add",
					"abcdefgh",
					"169.254.9.0/30",
					"clear",
					30,
					"\"ikev1\", \"ikev2\"",
					"2, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24",
					"\"AES128\", \"AES256\", \"AES128-GCM-16\", \"AES256-GCM-16\"",
					"\"SHA1\", \"SHA2-256\", \"SHA2-384\", \"SHA2-512\"",
					28800,
					"2, 5, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24",
					"\"AES128\", \"AES256\", \"AES128-GCM-16\", \"AES256-GCM-16\"",
					"\"SHA1\", \"SHA2-256\", \"SHA2-384\", \"SHA2-512\"",
					3600,
					100,
					540,
					1024,
					"add"),
				Check: resource.ComposeTestCheckFunc(
					testAccAwsVpnConnectionExists(resourceName, &vpn),
					resource.TestCheckResourceAttr(resourceName, "static_routes_only", "false"),

					resource.TestCheckResourceAttr(resourceName, "tunnel1_inside_cidr", "169.254.8.0/30"),
					resource.TestCheckResourceAttr(resourceName, "tunnel1_preshared_key", "12345678"),

					resource.TestCheckResourceAttr(resourceName, "tunnel2_inside_cidr", "169.254.9.0/30"),
					resource.TestCheckResourceAttr(resourceName, "tunnel2_preshared_key", "abcdefgh"),
				),
			},
			// TODO: Once #396, #3359, #5809 are fixed, an import test step should be added here
		},
	})
}

func TestAccAWSVpnConnection_withoutStaticRoutes(t *testing.T) {
	rInt := acctest.RandInt()
	rBgpAsn := acctest.RandIntRange(64512, 65534)
	resourceName := "aws_vpn_connection.test"
	var vpn ec2.VpnConnection

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccAwsVpnConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsVpnConnectionConfigUpdate(rInt, rBgpAsn),
				Check: resource.ComposeTestCheckFunc(
					testAccAwsVpnConnectionExists(resourceName, &vpn),
					resource.TestCheckResourceAttr(resourceName, "static_routes_only", "false"),
					resource.TestCheckResourceAttr(resourceName, "enable_acceleration", "false"),
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

func TestAccAWSVpnConnection_withEnableAcceleration(t *testing.T) {
	rBgpAsn := acctest.RandIntRange(64512, 65534)
	resourceName := "aws_vpn_connection.test"
	var vpn ec2.VpnConnection

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccAwsVpnConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsVpnConnectionConfigEnableAcceleration(rBgpAsn),
				Check: resource.ComposeTestCheckFunc(
					testAccAwsVpnConnectionExists(resourceName, &vpn),
					resource.TestCheckResourceAttr(resourceName, "enable_acceleration", "true"),
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

func TestAccAWSVpnConnection_withIpv6(t *testing.T) {
	rBgpAsn := acctest.RandIntRange(64512, 65534)
	resourceName := "aws_vpn_connection.test"
	var vpn ec2.VpnConnection

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccAwsVpnConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsVpnConnectionConfigIpv6(rBgpAsn, "fd00:2001:db8:2:2d1:81ff:fe41:d201/128", "fd00:2001:db8:2:2d1:81ff:fe41:d202/128", "fd00:2001:db8:2:2d1:81ff:fe41:d200/126", "fd00:2001:db8:2:2d1:81ff:fe41:d204/126"),
				Check: resource.ComposeTestCheckFunc(
					testAccAwsVpnConnectionExists(resourceName, &vpn),
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

func TestAccAWSVpnConnection_tags(t *testing.T) {
	rBgpAsn := acctest.RandIntRange(64512, 65534)
	resourceName := "aws_vpn_connection.test"
	var vpn ec2.VpnConnection

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccAwsVpnConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsVpnConnectionConfigTags1(rBgpAsn, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccAwsVpnConnectionExists(resourceName, &vpn),
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
				Config: testAccAwsVpnConnectionConfigTags2(rBgpAsn, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccAwsVpnConnectionExists(resourceName, &vpn),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAwsVpnConnectionConfigTags1(rBgpAsn, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccAwsVpnConnectionExists(resourceName, &vpn),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccAWSVpnConnection_disappears(t *testing.T) {
	rBgpAsn := acctest.RandIntRange(64512, 65534)
	resourceName := "aws_vpn_connection.test"
	var vpn ec2.VpnConnection

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccAwsVpnConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsVpnConnectionConfig(rBgpAsn),
				Check: resource.ComposeTestCheckFunc(
					testAccAwsVpnConnectionExists(resourceName, &vpn),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsVpnConnection(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccAwsVpnConnectionDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ec2conn
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_vpn_connection" {
			continue
		}

		resp, err := conn.DescribeVpnConnections(&ec2.DescribeVpnConnectionsInput{
			VpnConnectionIds: []*string{aws.String(rs.Primary.ID)},
		})

		if err != nil {
			if ec2err, ok := err.(awserr.Error); ok && ec2err.Code() == "InvalidVpnConnectionID.NotFound" {
				// not found
				return nil
			}
			return err
		}

		var vpn *ec2.VpnConnection
		for _, v := range resp.VpnConnections {
			if v.VpnConnectionId != nil && *v.VpnConnectionId == rs.Primary.ID {
				vpn = v
			}
		}

		if vpn == nil {
			// vpn connection not found
			return nil
		}

		if vpn.State != nil && *vpn.State == "deleted" {
			return nil
		}

	}

	return nil
}

func testAccAwsVpnConnectionExists(vpnConnectionResource string, vpnConnection *ec2.VpnConnection) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[vpnConnectionResource]
		if !ok {
			return fmt.Errorf("Not found: %s", vpnConnectionResource)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}
		connection, ok := s.RootModule().Resources[vpnConnectionResource]
		if !ok {
			return fmt.Errorf("Not found: %s", vpnConnectionResource)
		}

		ec2conn := testAccProvider.Meta().(*AWSClient).ec2conn

		resp, err := ec2conn.DescribeVpnConnections(&ec2.DescribeVpnConnectionsInput{
			VpnConnectionIds: []*string{aws.String(connection.Primary.ID)},
		})

		if err != nil {
			return err
		}

		*vpnConnection = *resp.VpnConnections[0]

		return nil
	}
}

func TestAWSVpnConnection_xmlconfig(t *testing.T) {
	tunnelInfo, err := xmlConfigToTunnelInfo(testAccAwsVpnTunnelInfoXML)
	if err != nil {
		t.Fatalf("Error unmarshalling XML: %s", err)
	}
	if tunnelInfo.Tunnel1Address != "FIRST_ADDRESS" {
		t.Fatalf("First address from tunnel XML was incorrect.")
	}
	if tunnelInfo.Tunnel1CgwInsideAddress != "FIRST_CGW_INSIDE_ADDRESS" {
		t.Fatalf("First Customer Gateway inside address from tunnel" +
			" XML was incorrect.")
	}
	if tunnelInfo.Tunnel1VgwInsideAddress != "FIRST_VGW_INSIDE_ADDRESS" {
		t.Fatalf("First VPN Gateway inside address from tunnel " +
			" XML was incorrect.")
	}
	if tunnelInfo.Tunnel1PreSharedKey != "FIRST_KEY" {
		t.Fatalf("First key from tunnel XML was incorrect.")
	}
	if tunnelInfo.Tunnel1BGPASN != "FIRST_BGP_ASN" {
		t.Fatalf("First bgp asn from tunnel XML was incorrect.")
	}
	if tunnelInfo.Tunnel1BGPHoldTime != 31 {
		t.Fatalf("First bgp holdtime from tunnel XML was incorrect.")
	}
	if tunnelInfo.Tunnel2Address != "SECOND_ADDRESS" {
		t.Fatalf("Second address from tunnel XML was incorrect.")
	}
	if tunnelInfo.Tunnel2CgwInsideAddress != "SECOND_CGW_INSIDE_ADDRESS" {
		t.Fatalf("Second Customer Gateway inside address from tunnel" +
			" XML was incorrect.")
	}
	if tunnelInfo.Tunnel2VgwInsideAddress != "SECOND_VGW_INSIDE_ADDRESS" {
		t.Fatalf("Second VPN Gateway inside address from tunnel " +
			" XML was incorrect.")
	}
	if tunnelInfo.Tunnel2PreSharedKey != "SECOND_KEY" {
		t.Fatalf("Second key from tunnel XML was incorrect.")
	}
	if tunnelInfo.Tunnel2BGPASN != "SECOND_BGP_ASN" {
		t.Fatalf("Second bgp asn from tunnel XML was incorrect.")
	}
	if tunnelInfo.Tunnel2BGPHoldTime != 32 {
		t.Fatalf("Second bgp holdtime from tunnel XML was incorrect.")
	}
}

func testAccAwsVpnConnectionConfig(rBgpAsn int) string {
	return fmt.Sprintf(`
resource "aws_vpn_gateway" "vpn_gateway" {
  tags = {
    Name = "vpn_gateway"
  }
}

resource "aws_customer_gateway" "customer_gateway" {
  bgp_asn    = %d
  ip_address = "178.0.0.1"
  type       = "ipsec.1"

  tags = {
    Name = "main-customer-gateway"
  }
}

resource "aws_vpn_connection" "test" {
  vpn_gateway_id      = aws_vpn_gateway.vpn_gateway.id
  customer_gateway_id = aws_customer_gateway.customer_gateway.id
  type                = "ipsec.1"
  static_routes_only  = true
}
`, rBgpAsn)
}

// Change static_routes_only to be false, forcing a refresh.
func testAccAwsVpnConnectionConfigUpdate(rInt, rBgpAsn int) string {
	return fmt.Sprintf(`
resource "aws_vpn_gateway" "vpn_gateway" {
  tags = {
    Name = "vpn_gateway"
  }
}

resource "aws_customer_gateway" "customer_gateway" {
  bgp_asn    = %d
  ip_address = "178.0.0.1"
  type       = "ipsec.1"

  tags = {
    Name = "main-customer-gateway-%d"
  }
}

resource "aws_vpn_connection" "test" {
  vpn_gateway_id      = aws_vpn_gateway.vpn_gateway.id
  customer_gateway_id = aws_customer_gateway.customer_gateway.id
  type                = "ipsec.1"
  static_routes_only  = false
  enable_acceleration = false
}
`, rBgpAsn, rInt)
}

func testAccAwsVpnConnectionConfigEnableAcceleration(rBgpAsn int) string {
	return fmt.Sprintf(`
resource "aws_ec2_transit_gateway" "test" {}
resource "aws_customer_gateway" "customer_gateway" {
  bgp_asn    = %d
  ip_address = "178.0.0.1"
  type       = "ipsec.1"
  tags = {
    Name = "tf-acc-test-ec2-vpn-connection-enable-acceleration"
  }
}
resource "aws_vpn_connection" "test" {
  customer_gateway_id = aws_customer_gateway.customer_gateway.id
  transit_gateway_id  = aws_ec2_transit_gateway.test.id
  type                = "ipsec.1"
  static_routes_only  = false
  enable_acceleration = true
}
`, rBgpAsn)
}

func testAccAwsVpnConnectionConfigIpv6(rBgpAsn int, localIpv6NetworkCidr string, remoteIpv6NetworkCidr string, tunnel1InsideIpv6Cidr string, tunnel2InsideIpv6Cidr string) string {
	return fmt.Sprintf(`
resource "aws_ec2_transit_gateway" "test" {}
resource "aws_customer_gateway" "customer_gateway" {
  bgp_asn    = %d
  ip_address = "178.0.0.1"
  type       = "ipsec.1"
  tags = {
    Name = "tf-acc-test-ec2-vpn-connection-enable-acceleration"
  }
}
resource "aws_vpn_connection" "test" {
  customer_gateway_id = aws_customer_gateway.customer_gateway.id
  transit_gateway_id  = aws_ec2_transit_gateway.test.id
  type                = "ipsec.1"
  static_routes_only  = false
  enable_acceleration = false

  local_ipv6_network_cidr  = "%s"
  remote_ipv6_network_cidr = "%s"
  tunnel_inside_ip_version = "ipv6"

  tunnel1_inside_ipv6_cidr = "%s"
  tunnel2_inside_ipv6_cidr = "%s"
}
`, rBgpAsn, localIpv6NetworkCidr, remoteIpv6NetworkCidr, tunnel1InsideIpv6Cidr, tunnel2InsideIpv6Cidr)
}

func testAccAwsVpnConnectionConfigSingleTunnelOptions(rBgpAsn int, psk string, tunnelCidr string) string {
	return fmt.Sprintf(`
resource "aws_vpn_gateway" "vpn_gateway" {
  tags = {
    Name = "vpn_gateway"
  }
}

resource "aws_customer_gateway" "customer_gateway" {
  bgp_asn    = %d
  ip_address = "178.0.0.1"
  type       = "ipsec.1"

  tags = {
    Name = "main-customer-gateway"
  }
}

resource "aws_vpn_connection" "test" {
  vpn_gateway_id      = aws_vpn_gateway.vpn_gateway.id
  customer_gateway_id = aws_customer_gateway.customer_gateway.id
  type                = "ipsec.1"
  static_routes_only  = false

  tunnel1_inside_cidr   = "%s"
  tunnel1_preshared_key = "%s"
}
`, rBgpAsn, tunnelCidr, psk)
}

func testAccAwsVpnConnectionConfigTransitGatewayID(rBgpAsn int) string {
	return fmt.Sprintf(`
resource "aws_ec2_transit_gateway" "test" {}

resource "aws_customer_gateway" "test" {
  bgp_asn    = %d
  ip_address = "178.0.0.1"
  type       = "ipsec.1"

  tags = {
    Name = "tf-acc-test-ec2-vpn-connection-transit-gateway-id"
  }
}

resource "aws_vpn_connection" "test" {
  customer_gateway_id = aws_customer_gateway.test.id
  transit_gateway_id  = aws_ec2_transit_gateway.test.id
  type                = aws_customer_gateway.test.type
}
`, rBgpAsn)
}

func testAccAwsVpnConnectionConfigTunnelOptions(
	rBgpAsn int,
	localIpv4NetworkCidr string,
	remoteIpv4NetworkCidr string,
	psk string,
	tunnelCidr string,
	dpdTimeoutAction string,
	dpdTimeoutSeconds int,
	ikeVersions string,
	phase1DhGroupNumbers string,
	phase1EncryptionAlgorithms string,
	phase1IntegrityAlgorithms string,
	phase1LifetimeSeconds int,
	phase2DhGroupNumbers string,
	phase2EncryptionAlgorithms string,
	phase2IntegrityAlgorithms string,
	phase2LifetimeSeconds int,
	rekeyFuzzPercentage int,
	rekeyMarginTimeSeconds int,
	replayWindowSize int,
	startupAction string,
	psk2 string,
	tunnelCidr2 string,
	dpdTimeoutAction2 string,
	dpdTimeoutSeconds2 int,
	ikeVersions2 string,
	phase1DhGroupNumbers2 string,
	phase1EncryptionAlgorithms2 string,
	phase1IntegrityAlgorithms2 string,
	phase1LifetimeSeconds2 int,
	phase2DhGroupNumbers2 string,
	phase2EncryptionAlgorithms2 string,
	phase2IntegrityAlgorithms2 string,
	phase2LifetimeSeconds2 int,
	rekeyFuzzPercentage2 int,
	rekeyMarginTimeSeconds2 int,
	replayWindowSize2 int,
	startupAction2 string,
) string {
	return fmt.Sprintf(`
resource "aws_vpn_gateway" "vpn_gateway" {
  tags = {
    Name = "vpn_gateway"
  }
}

resource "aws_customer_gateway" "customer_gateway" {
  bgp_asn    = %d
  ip_address = "178.0.0.1"
  type       = "ipsec.1"

  tags = {
    Name = "main-customer-gateway"
  }
}

resource "aws_vpn_connection" "test" {
  vpn_gateway_id      = aws_vpn_gateway.vpn_gateway.id
  customer_gateway_id = aws_customer_gateway.customer_gateway.id
  type                = "ipsec.1"
  static_routes_only  = false

  local_ipv4_network_cidr  = "%s"
  remote_ipv4_network_cidr = "%s"

  tunnel1_inside_cidr                  = "%s"
  tunnel1_preshared_key                = "%s"
  tunnel1_dpd_timeout_action           = "%s"
  tunnel1_dpd_timeout_seconds          = %d
  tunnel1_ike_versions                 = [%s]
  tunnel1_phase1_dh_group_numbers      = [%s]
  tunnel1_phase1_encryption_algorithms = [%s]
  tunnel1_phase1_integrity_algorithms  = [%s]
  tunnel1_phase1_lifetime_seconds      = %d
  tunnel1_phase2_dh_group_numbers      = [%s]
  tunnel1_phase2_encryption_algorithms = [%s]
  tunnel1_phase2_integrity_algorithms  = [%s]
  tunnel1_phase2_lifetime_seconds      = %d
  tunnel1_rekey_fuzz_percentage        = %d
  tunnel1_rekey_margin_time_seconds    = %d
  tunnel1_replay_window_size           = %d
  tunnel1_startup_action               = "%s"

  tunnel2_inside_cidr                  = "%s"
  tunnel2_preshared_key                = "%s"
  tunnel2_dpd_timeout_action           = "%s"
  tunnel2_dpd_timeout_seconds          = %d
  tunnel2_ike_versions                 = [%s]
  tunnel2_phase1_dh_group_numbers      = [%s]
  tunnel2_phase1_encryption_algorithms = [%s]
  tunnel2_phase1_integrity_algorithms  = [%s]
  tunnel2_phase1_lifetime_seconds      = %d
  tunnel2_phase2_dh_group_numbers      = [%s]
  tunnel2_phase2_encryption_algorithms = [%s]
  tunnel2_phase2_integrity_algorithms  = [%s]
  tunnel2_phase2_lifetime_seconds      = %d
  tunnel2_rekey_fuzz_percentage        = %d
  tunnel2_rekey_margin_time_seconds    = %d
  tunnel2_replay_window_size           = %d
  tunnel2_startup_action               = "%s"
}
`,
		rBgpAsn,
		localIpv4NetworkCidr,
		remoteIpv4NetworkCidr,
		tunnelCidr,
		psk,
		dpdTimeoutAction,
		dpdTimeoutSeconds,
		ikeVersions,
		phase1DhGroupNumbers,
		phase1EncryptionAlgorithms,
		phase1IntegrityAlgorithms,
		phase1LifetimeSeconds,
		phase2DhGroupNumbers,
		phase2EncryptionAlgorithms,
		phase2IntegrityAlgorithms,
		phase2LifetimeSeconds,
		rekeyFuzzPercentage,
		rekeyMarginTimeSeconds,
		replayWindowSize,
		startupAction,
		tunnelCidr2,
		psk2,
		dpdTimeoutAction2,
		dpdTimeoutSeconds2,
		ikeVersions2,
		phase1DhGroupNumbers2,
		phase1EncryptionAlgorithms2,
		phase1IntegrityAlgorithms2,
		phase1LifetimeSeconds2,
		phase2DhGroupNumbers2,
		phase2EncryptionAlgorithms2,
		phase2IntegrityAlgorithms2,
		phase2LifetimeSeconds2,
		rekeyFuzzPercentage2,
		rekeyMarginTimeSeconds2,
		replayWindowSize2,
		startupAction2)
}

func testAccAwsVpnConnectionConfigTags1(rBgpAsn int, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_vpn_gateway" "vpn_gateway" {
  tags = {
    Name = "vpn_gateway"
  }
}

resource "aws_customer_gateway" "customer_gateway" {
  bgp_asn    = %d
  ip_address = "178.0.0.1"
  type       = "ipsec.1"

  tags = {
    Name = "main-customer-gateway"
  }
}

resource "aws_vpn_connection" "test" {
  vpn_gateway_id      = aws_vpn_gateway.vpn_gateway.id
  customer_gateway_id = aws_customer_gateway.customer_gateway.id
  type                = "ipsec.1"
  static_routes_only  = true

  tags = {
    %[2]q = %[3]q
  }
}
`, rBgpAsn, tagKey1, tagValue1)
}

func testAccAwsVpnConnectionConfigTags2(rBgpAsn int, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_vpn_gateway" "vpn_gateway" {
  tags = {
    Name = "vpn_gateway"
  }
}

resource "aws_customer_gateway" "customer_gateway" {
  bgp_asn    = %d
  ip_address = "178.0.0.1"
  type       = "ipsec.1"

  tags = {
    Name = "main-customer-gateway"
  }
}

resource "aws_vpn_connection" "test" {
  vpn_gateway_id      = aws_vpn_gateway.vpn_gateway.id
  customer_gateway_id = aws_customer_gateway.customer_gateway.id
  type                = "ipsec.1"
  static_routes_only  = true

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rBgpAsn, tagKey1, tagValue1, tagKey2, tagValue2)
}

// Test our VPN tunnel config XML parsing
const testAccAwsVpnTunnelInfoXML = `
<vpn_connection id="vpn-abc123">
  <ipsec_tunnel>
    <customer_gateway>
      <tunnel_outside_address>
        <ip_address>123.123.123.123</ip_address>
      </tunnel_outside_address>
      <tunnel_inside_address>
        <ip_address>SECOND_CGW_INSIDE_ADDRESS</ip_address>
        <network_mask>255.255.255.252</network_mask>
        <network_cidr>30</network_cidr>
      </tunnel_inside_address>
    </customer_gateway>
    <vpn_gateway>
      <tunnel_outside_address>
        <ip_address>SECOND_ADDRESS</ip_address>
      </tunnel_outside_address>
      <tunnel_inside_address>
        <ip_address>SECOND_VGW_INSIDE_ADDRESS</ip_address>
        <network_mask>255.255.255.252</network_mask>
        <network_cidr>30</network_cidr>
      </tunnel_inside_address>
      <bgp>
        <asn>SECOND_BGP_ASN</asn>
        <hold_time>32</hold_time>
      </bgp>
    </vpn_gateway>
    <ike>
      <pre_shared_key>SECOND_KEY</pre_shared_key>
    </ike>
  </ipsec_tunnel>
  <ipsec_tunnel>
    <customer_gateway>
      <tunnel_outside_address>
        <ip_address>123.123.123.123</ip_address>
      </tunnel_outside_address>
      <tunnel_inside_address>
        <ip_address>FIRST_CGW_INSIDE_ADDRESS</ip_address>
        <network_mask>255.255.255.252</network_mask>
        <network_cidr>30</network_cidr>
      </tunnel_inside_address>
    </customer_gateway>
    <vpn_gateway>
      <tunnel_outside_address>
        <ip_address>FIRST_ADDRESS</ip_address>
      </tunnel_outside_address>
      <tunnel_inside_address>
        <ip_address>FIRST_VGW_INSIDE_ADDRESS</ip_address>
        <network_mask>255.255.255.252</network_mask>
        <network_cidr>30</network_cidr>
      </tunnel_inside_address>
      <bgp>
        <asn>FIRST_BGP_ASN</asn>
        <hold_time>31</hold_time>
      </bgp>
    </vpn_gateway>
    <ike>
      <pre_shared_key>FIRST_KEY</pre_shared_key>
    </ike>
  </ipsec_tunnel>
</vpn_connection>
`
