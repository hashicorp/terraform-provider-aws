package directconnect_test

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directconnect"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccDirectConnectHostedPublicVirtualInterface_basic(t *testing.T) {
	key := "DX_CONNECTION_ID"
	connectionId := os.Getenv(key)
	if connectionId == "" {
		t.Skipf("Environment variable %s is not set", key)
	}

	var providers []*schema.Provider
	var vif directconnect.VirtualInterface
	resourceName := "aws_dx_hosted_public_virtual_interface.test"
	accepterResourceName := "aws_dx_hosted_public_virtual_interface_accepter.test"
	rName := fmt.Sprintf("tf-testacc-public-vif-%s", sdkacctest.RandString(10))
	amazonAddress := "175.45.176.5/28"
	customerAddress := "175.45.176.6/28"
	bgpAsn := sdkacctest.RandIntRange(64512, 65534)
	vlan := sdkacctest.RandIntRange(2049, 4094)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckAlternateAccount(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, directconnect.EndpointsID),
		ProviderFactories: acctest.FactoriesAlternate(&providers),
		CheckDestroy:      testAccCheckHostedPublicVirtualInterfaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccHostedPublicVirtualInterfaceConfig_basic(connectionId, rName, amazonAddress, customerAddress, bgpAsn, vlan),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHostedPublicVirtualInterfaceExists(resourceName, &vif),
					resource.TestCheckResourceAttr(resourceName, "address_family", "ipv4"),
					resource.TestCheckResourceAttrSet(resourceName, "amazon_address"),
					resource.TestCheckResourceAttrSet(resourceName, "amazon_side_asn"),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "directconnect", regexp.MustCompile(fmt.Sprintf("dxvif/%s", aws.StringValue(vif.VirtualInterfaceId)))),
					resource.TestCheckResourceAttrSet(resourceName, "aws_device"),
					resource.TestCheckResourceAttr(resourceName, "bgp_asn", strconv.Itoa(bgpAsn)),
					resource.TestCheckResourceAttrSet(resourceName, "bgp_auth_key"),
					resource.TestCheckResourceAttr(resourceName, "connection_id", connectionId),
					resource.TestCheckResourceAttr(resourceName, "customer_address", customerAddress),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "route_filter_prefixes.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "route_filter_prefixes.*", "210.52.109.0/24"),
					resource.TestCheckTypeSetElemAttr(resourceName, "route_filter_prefixes.*", "175.45.176.0/22"),
					resource.TestCheckResourceAttr(resourceName, "vlan", strconv.Itoa(vlan)),
					// Accepter's attributes:
					resource.TestCheckResourceAttrSet(accepterResourceName, "arn"),
					resource.TestCheckResourceAttr(accepterResourceName, "tags.%", "0"),
					resource.TestCheckResourceAttrPair(accepterResourceName, "virtual_interface_id", resourceName, "id"),
				),
			},
			// Test import.
			{
				Config:            testAccHostedPublicVirtualInterfaceConfig_basic(connectionId, rName, amazonAddress, customerAddress, bgpAsn, vlan),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccDirectConnectHostedPublicVirtualInterface_accepterTags(t *testing.T) {
	key := "DX_CONNECTION_ID"
	connectionId := os.Getenv(key)
	if connectionId == "" {
		t.Skipf("Environment variable %s is not set", key)
	}

	var providers []*schema.Provider
	var vif directconnect.VirtualInterface
	resourceName := "aws_dx_hosted_public_virtual_interface.test"
	accepterResourceName := "aws_dx_hosted_public_virtual_interface_accepter.test"
	rName := fmt.Sprintf("tf-testacc-public-vif-%s", sdkacctest.RandString(10))
	amazonAddress := "175.45.176.7/28"
	customerAddress := "175.45.176.8/28"
	bgpAsn := sdkacctest.RandIntRange(64512, 65534)
	vlan := sdkacctest.RandIntRange(2049, 4094)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckAlternateAccount(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, directconnect.EndpointsID),
		ProviderFactories: acctest.FactoriesAlternate(&providers),
		CheckDestroy:      testAccCheckHostedPublicVirtualInterfaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccHostedPublicVirtualInterfaceConfig_accepterTags(connectionId, rName, amazonAddress, customerAddress, bgpAsn, vlan),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHostedPublicVirtualInterfaceExists(resourceName, &vif),
					resource.TestCheckResourceAttr(resourceName, "address_family", "ipv4"),
					resource.TestCheckResourceAttr(resourceName, "amazon_address", amazonAddress),
					resource.TestCheckResourceAttrSet(resourceName, "amazon_side_asn"),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "directconnect", regexp.MustCompile(fmt.Sprintf("dxvif/%s", aws.StringValue(vif.VirtualInterfaceId)))),
					resource.TestCheckResourceAttrSet(resourceName, "aws_device"),
					resource.TestCheckResourceAttr(resourceName, "bgp_asn", strconv.Itoa(bgpAsn)),
					resource.TestCheckResourceAttrSet(resourceName, "bgp_auth_key"),
					resource.TestCheckResourceAttr(resourceName, "connection_id", connectionId),
					resource.TestCheckResourceAttr(resourceName, "customer_address", customerAddress),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "route_filter_prefixes.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "route_filter_prefixes.*", "210.52.109.0/24"),
					resource.TestCheckTypeSetElemAttr(resourceName, "route_filter_prefixes.*", "175.45.176.0/22"),
					resource.TestCheckResourceAttr(resourceName, "vlan", strconv.Itoa(vlan)),
					// Accepter's attributes:
					resource.TestCheckResourceAttrSet(accepterResourceName, "arn"),
					resource.TestCheckResourceAttr(accepterResourceName, "tags.%", "3"),
					resource.TestCheckResourceAttr(accepterResourceName, "tags.Name", rName),
					resource.TestCheckResourceAttr(accepterResourceName, "tags.Key1", "Value1"),
					resource.TestCheckResourceAttr(accepterResourceName, "tags.Key2", "Value2a"),
					resource.TestCheckResourceAttrPair(accepterResourceName, "virtual_interface_id", resourceName, "id"),
				),
			},
			{
				Config: testAccHostedPublicVirtualInterfaceConfig_accepterTagsUpdated(connectionId, rName, amazonAddress, customerAddress, bgpAsn, vlan),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHostedPublicVirtualInterfaceExists(resourceName, &vif),
					resource.TestCheckResourceAttr(resourceName, "address_family", "ipv4"),
					resource.TestCheckResourceAttr(resourceName, "amazon_address", amazonAddress),
					resource.TestCheckResourceAttrSet(resourceName, "amazon_side_asn"),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "directconnect", regexp.MustCompile(fmt.Sprintf("dxvif/%s", aws.StringValue(vif.VirtualInterfaceId)))),
					resource.TestCheckResourceAttrSet(resourceName, "aws_device"),
					resource.TestCheckResourceAttr(resourceName, "bgp_asn", strconv.Itoa(bgpAsn)),
					resource.TestCheckResourceAttrSet(resourceName, "bgp_auth_key"),
					resource.TestCheckResourceAttr(resourceName, "connection_id", connectionId),
					resource.TestCheckResourceAttr(resourceName, "customer_address", customerAddress),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "route_filter_prefixes.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "route_filter_prefixes.*", "210.52.109.0/24"),
					resource.TestCheckTypeSetElemAttr(resourceName, "route_filter_prefixes.*", "175.45.176.0/22"),
					resource.TestCheckResourceAttr(resourceName, "vlan", strconv.Itoa(vlan)),
					// Accepter's attributes:
					resource.TestCheckResourceAttrSet(accepterResourceName, "arn"),
					resource.TestCheckResourceAttr(accepterResourceName, "tags.%", "3"),
					resource.TestCheckResourceAttr(accepterResourceName, "tags.Name", rName),
					resource.TestCheckResourceAttr(accepterResourceName, "tags.Key2", "Value2b"),
					resource.TestCheckResourceAttr(accepterResourceName, "tags.Key3", "Value3"),
					resource.TestCheckResourceAttrPair(accepterResourceName, "virtual_interface_id", resourceName, "id"),
				),
			},
		},
	})
}

func testAccCheckHostedPublicVirtualInterfaceDestroy(s *terraform.State) error {
	return testAccCheckVirtualInterfaceDestroy(s, "aws_dx_hosted_public_virtual_interface")
}

func testAccCheckHostedPublicVirtualInterfaceExists(name string, vif *directconnect.VirtualInterface) resource.TestCheckFunc {
	return testAccCheckVirtualInterfaceExists(name, vif)
}

func testAccHostedPublicVirtualInterfaceConfig_base(cid, rName, amzAddr, custAddr string, bgpAsn, vlan int) string {
	return acctest.ConfigAlternateAccountProvider() + fmt.Sprintf(`
# Creator
resource "aws_dx_hosted_public_virtual_interface" "test" {
  address_family   = "ipv4"
  amazon_address   = %[3]q
  bgp_asn          = %[5]d
  connection_id    = %[1]q
  customer_address = %[4]q
  name             = %[2]q
  owner_account_id = data.aws_caller_identity.accepter.account_id
  vlan             = %[6]d

  route_filter_prefixes = [
    "175.45.176.0/22",
    "210.52.109.0/24",
  ]
}

# Accepter
data "aws_caller_identity" "accepter" {
  provider = "awsalternate"
}
`, cid, rName, amzAddr, custAddr, bgpAsn, vlan)
}

func testAccHostedPublicVirtualInterfaceConfig_basic(cid, rName, amzAddr, custAddr string, bgpAsn, vlan int) string {
	return testAccHostedPublicVirtualInterfaceConfig_base(cid, rName, amzAddr, custAddr, bgpAsn, vlan) + `
resource "aws_dx_hosted_public_virtual_interface_accepter" "test" {
  provider = "awsalternate"

  virtual_interface_id = aws_dx_hosted_public_virtual_interface.test.id
}
`
}

func testAccHostedPublicVirtualInterfaceConfig_accepterTags(cid, rName, amzAddr, custAddr string, bgpAsn, vlan int) string {
	return testAccHostedPublicVirtualInterfaceConfig_base(cid, rName, amzAddr, custAddr, bgpAsn, vlan) + fmt.Sprintf(`
resource "aws_dx_hosted_public_virtual_interface_accepter" "test" {
  provider = "awsalternate"

  virtual_interface_id = aws_dx_hosted_public_virtual_interface.test.id

  tags = {
    Name = %[1]q
    Key1 = "Value1"
    Key2 = "Value2a"
  }
}
`, rName)
}

func testAccHostedPublicVirtualInterfaceConfig_accepterTagsUpdated(cid, rName, amzAddr, custAddr string, bgpAsn, vlan int) string {
	return testAccHostedPublicVirtualInterfaceConfig_base(cid, rName, amzAddr, custAddr, bgpAsn, vlan) + fmt.Sprintf(`
resource "aws_dx_hosted_public_virtual_interface_accepter" "test" {
  provider = "awsalternate"

  virtual_interface_id = aws_dx_hosted_public_virtual_interface.test.id

  tags = {
    Name = %[1]q
    Key2 = "Value2b"
    Key3 = "Value3"
  }
}
`, rName)
}
