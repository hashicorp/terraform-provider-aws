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

func TestAccAwsDxHostedPublicVirtualInterface_basic(t *testing.T) {
	key := "DX_CONNECTION_ID"
	connectionId := os.Getenv(key)
	if connectionId == "" {
		t.Skipf("Environment variable %s is not set", key)
	}

	var providers []*schema.Provider
	resourceNameHostedVif := "aws_dx_hosted_public_virtual_interface.test"
	resourceNameHostedVifAccepter := "aws_dx_hosted_public_virtual_interface_accepter.test"
	rName := fmt.Sprintf("tf-testacc-public-vif-%s", acctest.RandString(10))
	bgpAsn := randIntRange(64512, 65534)
	vlan := randIntRange(2049, 4094)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccAlternateAccountPreCheck(t)
		},
		ProviderFactories: testAccProviderFactories(&providers),
		CheckDestroy:      testAccCheckAwsDxHostedPublicVirtualInterfaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDxHostedPublicVirtualInterfaceConfig_basic(connectionId, rName, bgpAsn, vlan),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsDxHostedPublicVirtualInterfaceExists(resourceNameHostedVif),
					testAccCheckAwsDxHostedPublicVirtualInterfaceAccepterExists(resourceNameHostedVifAccepter),
					resource.TestCheckResourceAttr(resourceNameHostedVif, "name", rName),
					resource.TestCheckResourceAttr(resourceNameHostedVifAccepter, "tags.%", "0"),
				),
			},
			{
				Config: testAccDxHostedPublicVirtualInterfaceConfig_updated(connectionId, rName, bgpAsn, vlan),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsDxHostedPublicVirtualInterfaceExists(resourceNameHostedVif),
					testAccCheckAwsDxHostedPublicVirtualInterfaceAccepterExists(resourceNameHostedVifAccepter),
					resource.TestCheckResourceAttr(resourceNameHostedVif, "name", rName),
					resource.TestCheckResourceAttr(resourceNameHostedVifAccepter, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceNameHostedVifAccepter, "tags.Environment", "test"),
				),
			},
			// Test import.
			{
				Config:            testAccDxHostedPublicVirtualInterfaceConfig_updated(connectionId, rName, bgpAsn, vlan),
				ResourceName:      resourceNameHostedVif,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckAwsDxHostedPublicVirtualInterfaceDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).dxconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_dx_hosted_public_virtual_interface" {
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
				return fmt.Errorf("[DESTROY ERROR] Dx Public VIF (%s) not deleted", rs.Primary.ID)
			}
		}
	}
	return nil
}

func testAccCheckAwsDxHostedPublicVirtualInterfaceExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		return nil
	}
}

func testAccCheckAwsDxHostedPublicVirtualInterfaceAccepterExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		return nil
	}
}

func testAccDxHostedPublicVirtualInterfaceConfig_base(cid, rName string, bgpAsn, vlan int) string {
	return testAccAlternateAccountProviderConfig() + fmt.Sprintf(`
# Creator
resource "aws_dx_hosted_public_virtual_interface" "test" {
  connection_id    = %[1]q
  owner_account_id = "${data.aws_caller_identity.accepter.account_id}"

  name           = %[2]q
  vlan           = %[4]d
  address_family = "ipv4"
  bgp_asn        = %[3]d

  customer_address = "175.45.176.1/30"
  amazon_address   = "175.45.176.2/30"

  route_filter_prefixes = [
    "210.52.109.0/24",
    "175.45.176.0/22",
  ]
}

# Accepter
data "aws_caller_identity" "accepter" {
  provider = "aws.alternate"
}
`, cid, rName, bgpAsn, vlan)
}

func testAccDxHostedPublicVirtualInterfaceConfig_basic(cid, rName string, bgpAsn, vlan int) string {
	return testAccDxHostedPublicVirtualInterfaceConfig_base(cid, rName, bgpAsn, vlan) + fmt.Sprintf(`
resource "aws_dx_hosted_public_virtual_interface_accepter" "test" {
  provider             = "aws.alternate"

  virtual_interface_id = "${aws_dx_hosted_public_virtual_interface.test.id}"
}
`)
}

func testAccDxHostedPublicVirtualInterfaceConfig_updated(cid, rName string, bgpAsn, vlan int) string {
	return testAccDxHostedPublicVirtualInterfaceConfig_base(cid, rName, bgpAsn, vlan) + fmt.Sprintf(`
resource "aws_dx_hosted_public_virtual_interface_accepter" "test" {
  provider             = "aws.alternate"

  virtual_interface_id = "${aws_dx_hosted_public_virtual_interface.test.id}"

  tags = {
    Environment = "test"
  }
}
`)
}
