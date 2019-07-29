package aws

import (
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directconnect"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAwsDxTransitVirtualInterface_basic(t *testing.T) {
	key := "DX_CONNECTION_ID"
	connectionId := os.Getenv(key)
	if connectionId == "" {
		t.Skipf("Environment variable %s is not set", key)
	}

	resourceName := "aws_dx_transit_virtual_interface.test"
	rName := fmt.Sprintf("tf-testacc-transit-vif-%s", acctest.RandString(9))
	amzAsn := randIntRange(64512, 65534)
	bgpAsn := randIntRange(64512, 65534)
	vlan := randIntRange(2049, 4094)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsDxTransitVirtualInterfaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDxTransitVirtualInterfaceConfig_basic(connectionId, rName, amzAsn, bgpAsn, vlan),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsDxTransitVirtualInterfaceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "mtu", "1500"),
					resource.TestCheckResourceAttr(resourceName, "jumbo_frame_capable", "true"),
				),
			},
			{
				Config: testAccDxTransitVirtualInterfaceConfig_updated(connectionId, rName, amzAsn, bgpAsn, vlan),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsDxTransitVirtualInterfaceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Environment", "test"),
					resource.TestCheckResourceAttr(resourceName, "mtu", "8500"),
				),
			},
			// Test import.
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckAwsDxTransitVirtualInterfaceDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).dxconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_dx_transit_virtual_interface" {
			continue
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		resp, err := conn.DescribeVirtualInterfaces(&directconnect.DescribeVirtualInterfacesInput{
			VirtualInterfaceId: aws.String(rs.Primary.ID),
		})
		if isAWSErr(err, directconnect.ErrCodeClientException, "does not exist") {
			continue
		}
		if err != nil {
			return err
		}

		n := len(resp.VirtualInterfaces)
		switch n {
		case 0:
			continue
		case 1:
			if aws.StringValue(resp.VirtualInterfaces[0].VirtualInterfaceState) == directconnect.VirtualInterfaceStateDeleted {
				continue
			}
			return fmt.Errorf("still exist.")
		default:
			return fmt.Errorf("Found %d Direct Connect virtual interfaces for %s, expected 1", n, rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckAwsDxTransitVirtualInterfaceExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).dxconn

		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		_, err := conn.DescribeVirtualInterfaces(&directconnect.DescribeVirtualInterfacesInput{
			VirtualInterfaceId: aws.String(rs.Primary.ID),
		})
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccDxTransitVirtualInterfaceConfig_basic(cid, rName string, amzAsn, bgpAsn, vlan int) string {
	return fmt.Sprintf(`
resource "aws_dx_gateway" "test" {
  name            = %[2]q
  amazon_side_asn = %[3]d
}

resource "aws_dx_transit_virtual_interface" "test" {
  connection_id    = %[1]q

  dx_gateway_id  = "${aws_dx_gateway.test.id}"
  name           = %[2]q
  vlan           = %[5]d
  address_family = "ipv4"
  bgp_asn        = %[4]d
}
`, cid, rName, amzAsn, bgpAsn, vlan)
}

func testAccDxTransitVirtualInterfaceConfig_updated(cid, rName string, amzAsn, bgpAsn, vlan int) string {
	return fmt.Sprintf(`
resource "aws_dx_gateway" "test" {
  name            = %[2]q
  amazon_side_asn = %[3]d
}

resource "aws_dx_transit_virtual_interface" "test" {
  connection_id    = %[1]q

  dx_gateway_id  = "${aws_dx_gateway.test.id}"
  name           = %[2]q
  vlan           = %[5]d
  address_family = "ipv4"
  bgp_asn        = %[4]d
  mtu            = 8500

  tags = {
    Environment = "test"
  }
}
`, cid, rName, amzAsn, bgpAsn, vlan)
}
