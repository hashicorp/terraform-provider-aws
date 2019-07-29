package aws

import (
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directconnect"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAwsDxHostedTransitVirtualInterface_basic(t *testing.T) {
	key := "DX_CONNECTION_ID"
	connectionId := os.Getenv(key)
	if connectionId == "" {
		t.Skipf("Environment variable %s is not set", key)
	}

	var providers []*schema.Provider
	resourceNameHostedVif := "aws_dx_hosted_transit_virtual_interface.test"
	resourceNameHostedVifAccepter := "aws_dx_hosted_transit_virtual_interface_accepter.test"
	rName := fmt.Sprintf("tf-testacc-transit-vif-%s", acctest.RandString(9))
	amzAsn := randIntRange(64512, 65534)
	bgpAsn := randIntRange(64512, 65534)
	vlan := randIntRange(2049, 4094)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccAlternateAccountPreCheck(t)
		},
		ProviderFactories: testAccProviderFactories(&providers),
		CheckDestroy:      testAccCheckAwsDxHostedTransitVirtualInterfaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDxHostedTransitVirtualInterfaceConfig_basic(connectionId, rName, amzAsn, bgpAsn, vlan),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsDxHostedTransitVirtualInterfaceExists(resourceNameHostedVif),
					testAccCheckAwsDxHostedTransitVirtualInterfaceAccepterExists(resourceNameHostedVifAccepter),
					resource.TestCheckResourceAttr(resourceNameHostedVif, "name", rName),
					resource.TestCheckResourceAttr(resourceNameHostedVif, "mtu", "1500"),
					resource.TestCheckResourceAttr(resourceNameHostedVif, "jumbo_frame_capable", "true"),
					resource.TestCheckResourceAttr(resourceNameHostedVifAccepter, "tags.%", "0"),
				),
			},
			{
				Config: testAccDxHostedTransitVirtualInterfaceConfig_updated(connectionId, rName, amzAsn, bgpAsn, vlan),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsDxHostedTransitVirtualInterfaceExists(resourceNameHostedVif),
					testAccCheckAwsDxHostedTransitVirtualInterfaceAccepterExists(resourceNameHostedVifAccepter),
					resource.TestCheckResourceAttr(resourceNameHostedVif, "name", rName),
					resource.TestCheckResourceAttr(resourceNameHostedVifAccepter, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceNameHostedVifAccepter, "tags.Environment", "test"),
				),
			},
			// Test import.
			{
				Config:            testAccDxHostedTransitVirtualInterfaceConfig_updated(connectionId, rName, amzAsn, bgpAsn, vlan),
				ResourceName:      resourceNameHostedVif,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckAwsDxHostedTransitVirtualInterfaceDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).dxconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_dx_hosted_transit_virtual_interface" {
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

func testAccCheckAwsDxHostedTransitVirtualInterfaceExists(name string) resource.TestCheckFunc {
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

func testAccCheckAwsDxHostedTransitVirtualInterfaceAccepterExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		return nil
	}
}

func testAccDxHostedTransitVirtualInterfaceConfig_base(cid, rName string, amzAsn, bgpAsn, vlan int) string {
	return testAccAlternateAccountProviderConfig() + fmt.Sprintf(`
# Creator
resource "aws_dx_hosted_transit_virtual_interface" "test" {
  connection_id    = %[1]q
  owner_account_id = "${data.aws_caller_identity.accepter.account_id}"

  name           = %[2]q
  vlan           = %[5]d
  address_family = "ipv4"
  bgp_asn        = %[4]d

  # The aws_dx_hosted_transit_virtual_interface
  # must be destroyed before the aws_dx_gateway.
  depends_on = ["aws_dx_gateway.test"]
}

# Accepter
data "aws_caller_identity" "accepter" {
  provider = "aws.alternate"
}

resource "aws_dx_gateway" "test" {
  provider = "aws.alternate"

  name            = %[2]q
  amazon_side_asn = %[3]d
}
`, cid, rName, amzAsn, bgpAsn, vlan)
}

func testAccDxHostedTransitVirtualInterfaceConfig_basic(cid, rName string, amzAsn, bgpAsn, vlan int) string {
	return testAccDxHostedTransitVirtualInterfaceConfig_base(cid, rName, amzAsn, bgpAsn, vlan) + fmt.Sprintf(`
resource "aws_dx_hosted_transit_virtual_interface_accepter" "test" {
  provider             = "aws.alternate"

  virtual_interface_id = "${aws_dx_hosted_transit_virtual_interface.test.id}"
  dx_gateway_id        = "${aws_dx_gateway.test.id}"
}
`)
}

func testAccDxHostedTransitVirtualInterfaceConfig_updated(cid, rName string, amzAsn, bgpAsn, vlan int) string {
	return testAccDxHostedTransitVirtualInterfaceConfig_base(cid, rName, amzAsn, bgpAsn, vlan) + fmt.Sprintf(`
resource "aws_dx_hosted_transit_virtual_interface_accepter" "test" {
  provider             = "aws.alternate"

  virtual_interface_id = "${aws_dx_hosted_transit_virtual_interface.test.id}"
  dx_gateway_id        = "${aws_dx_gateway.test.id}"

  tags = {
    Environment = "test"
  }
}
`)
}
