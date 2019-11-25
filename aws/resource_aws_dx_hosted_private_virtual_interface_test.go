package aws

import (
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directconnect"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccAwsDxHostedPrivateVirtualInterface_basic(t *testing.T) {
	key := "DX_CONNECTION_ID"
	connectionId := os.Getenv(key)
	if connectionId == "" {
		t.Skipf("Environment variable %s is not set", key)
	}

	var providers []*schema.Provider
	resourceNameHostedVif := "aws_dx_hosted_private_virtual_interface.test"
	resourceNameHostedVifAccepter := "aws_dx_hosted_private_virtual_interface_accepter.test"
	rName := fmt.Sprintf("tf-testacc-private-vif-%s", acctest.RandString(9))
	bgpAsn := randIntRange(64512, 65534)
	vlan := randIntRange(2049, 4094)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccAlternateAccountPreCheck(t)
		},
		ProviderFactories: testAccProviderFactories(&providers),
		CheckDestroy:      testAccCheckAwsDxHostedPrivateVirtualInterfaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDxHostedPrivateVirtualInterfaceConfig_basic(connectionId, rName, bgpAsn, vlan),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsDxHostedPrivateVirtualInterfaceExists(resourceNameHostedVif),
					testAccCheckAwsDxHostedPrivateVirtualInterfaceAccepterExists(resourceNameHostedVifAccepter),
					resource.TestCheckResourceAttr(resourceNameHostedVif, "name", rName),
					resource.TestCheckResourceAttr(resourceNameHostedVif, "mtu", "1500"),
					resource.TestCheckResourceAttr(resourceNameHostedVif, "jumbo_frame_capable", "true"),
					resource.TestCheckResourceAttr(resourceNameHostedVifAccepter, "tags.%", "0"),
				),
			},
			{
				Config: testAccDxHostedPrivateVirtualInterfaceConfig_updated(connectionId, rName, bgpAsn, vlan),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsDxHostedPrivateVirtualInterfaceExists(resourceNameHostedVif),
					testAccCheckAwsDxHostedPrivateVirtualInterfaceAccepterExists(resourceNameHostedVifAccepter),
					resource.TestCheckResourceAttr(resourceNameHostedVif, "name", rName),
					resource.TestCheckResourceAttr(resourceNameHostedVifAccepter, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceNameHostedVifAccepter, "tags.Environment", "test"),
				),
			},
			// Test import.
			{
				Config:            testAccDxHostedPrivateVirtualInterfaceConfig_updated(connectionId, rName, bgpAsn, vlan),
				ResourceName:      resourceNameHostedVif,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckAwsDxHostedPrivateVirtualInterfaceDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).dxconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_dx_hosted_private_virtual_interface" {
			continue
		}

		input := &directconnect.DescribeVirtualInterfacesInput{
			VirtualInterfaceId: aws.String(rs.Primary.ID),
		}

		resp, err := conn.DescribeVirtualInterfaces(input)
		if err != nil {
			return err
		}
		for _, v := range resp.VirtualInterfaces {
			if *v.VirtualInterfaceId == rs.Primary.ID && !(*v.VirtualInterfaceState == directconnect.VirtualInterfaceStateDeleted) {
				return fmt.Errorf("[DESTROY ERROR] Dx Private VIF (%s) not deleted", rs.Primary.ID)
			}
		}
	}
	return nil
}

func testAccCheckAwsDxHostedPrivateVirtualInterfaceExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		return nil
	}
}

func testAccCheckAwsDxHostedPrivateVirtualInterfaceAccepterExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		return nil
	}
}

func testAccDxHostedPrivateVirtualInterfaceConfig_base(cid, rName string, bgpAsn, vlan int) string {
	return testAccAlternateAccountProviderConfig() + fmt.Sprintf(`
# Creator
resource "aws_dx_hosted_private_virtual_interface" "test" {
  connection_id    = %[1]q
  owner_account_id = "${data.aws_caller_identity.accepter.account_id}"

  name           = %[2]q
  vlan           = %[4]d
  address_family = "ipv4"
  bgp_asn        = %[3]d

  # The aws_dx_hosted_private_virtual_interface
  # must be destroyed before the aws_vpn_gateway.
  depends_on = ["aws_vpn_gateway.test"]
}

# Accepter
data "aws_caller_identity" "accepter" {
  provider = "aws.alternate"
}

resource "aws_vpn_gateway" "test" {
  provider = "aws.alternate"

  tags = {
    Name = %[2]q
  }
}
`, cid, rName, bgpAsn, vlan)
}

func testAccDxHostedPrivateVirtualInterfaceConfig_basic(cid, rName string, bgpAsn, vlan int) string {
	return testAccDxHostedPrivateVirtualInterfaceConfig_base(cid, rName, bgpAsn, vlan) + fmt.Sprintf(`
resource "aws_dx_hosted_private_virtual_interface_accepter" "test" {
  provider             = "aws.alternate"

  virtual_interface_id = "${aws_dx_hosted_private_virtual_interface.test.id}"
  vpn_gateway_id       = "${aws_vpn_gateway.test.id}"
}
`)
}

func testAccDxHostedPrivateVirtualInterfaceConfig_updated(cid, rName string, bgpAsn, vlan int) string {
	return testAccDxHostedPrivateVirtualInterfaceConfig_base(cid, rName, bgpAsn, vlan) + fmt.Sprintf(`
resource "aws_dx_hosted_private_virtual_interface_accepter" "test" {
  provider             = "aws.alternate"

  virtual_interface_id = "${aws_dx_hosted_private_virtual_interface.test.id}"
  vpn_gateway_id       = "${aws_vpn_gateway.test.id}"

  tags = {
    Environment = "test"
  }
}
`)
}
