package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
)

func TestAccAwsDxCrossAccountGatewayAssociation_basic(t *testing.T) {
	var providers []*schema.Provider
	resourceName := "aws_dx_cross_account_gateway_association.test"
	resourceNameDxGw := "aws_dx_gateway.test"
	resourceNameVgw := "aws_vpn_gateway.test"
	rName := fmt.Sprintf("terraform-testacc-dxxacctgwassoc-%d", acctest.RandInt())
	rBgpAsn := randIntRange(64512, 65534)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccAlternateAccountPreCheck(t)
		},
		ProviderFactories: testAccProviderFactories(&providers),
		CheckDestroy:      testAccCheckAwsDxGatewayAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDxCrossAccountGatewayAssociationConfig_basic(rName, rBgpAsn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsDxGatewayAssociationExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "dx_gateway_id", resourceNameDxGw, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "vpn_gateway_id", resourceNameVgw, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "dx_gateway_association_id"),
					resource.TestCheckResourceAttrSet(resourceName, "dx_gateway_owner_account_id"),
					resource.TestCheckResourceAttr(resourceName, "allowed_prefixes.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "allowed_prefixes.1216997074", "10.255.255.0/28"),
				),
			},
		},
	})
}

func TestAccAwsDxCrossAccountGatewayAssociation_allowedPrefixes(t *testing.T) {
	var providers []*schema.Provider
	resourceName := "aws_dx_cross_account_gateway_association.test"
	resourceNameDxGw := "aws_dx_gateway.test"
	resourceNameVgw := "aws_vpn_gateway.test"
	rName := fmt.Sprintf("terraform-testacc-dxxacctgwassoc-%d", acctest.RandInt())
	rBgpAsn := randIntRange(64512, 65534)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccAlternateAccountPreCheck(t)
		},
		ProviderFactories: testAccProviderFactories(&providers),
		CheckDestroy:      testAccCheckAwsDxGatewayAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDxCrossAccountGatewayAssociationConfig_allowedPrefixes(rName, rBgpAsn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsDxGatewayAssociationExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "dx_gateway_id", resourceNameDxGw, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "vpn_gateway_id", resourceNameVgw, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "dx_gateway_association_id"),
					resource.TestCheckResourceAttrSet(resourceName, "dx_gateway_owner_account_id"),
					resource.TestCheckResourceAttr(resourceName, "allowed_prefixes.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "allowed_prefixes.1642241106", "10.255.255.8/29"),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccDxCrossAccountGatewayAssociationConfig_allowedPrefixesUpdated(rName, rBgpAsn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsDxGatewayAssociationExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "dx_gateway_id", resourceNameDxGw, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "vpn_gateway_id", resourceNameVgw, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "dx_gateway_association_id"),
					resource.TestCheckResourceAttrSet(resourceName, "dx_gateway_owner_account_id"),
					resource.TestCheckResourceAttr(resourceName, "allowed_prefixes.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "allowed_prefixes.2173830893", "10.255.255.0/30"),
					resource.TestCheckResourceAttr(resourceName, "allowed_prefixes.2984398124", "10.255.255.8/30"),
				),
			},
		},
	})
}

func testAccDxCrossAccountGatewayAssociationConfig_base(rName string, rBgpAsn int) string {
	return testAccAlternateAccountProviderConfig() + fmt.Sprintf(`
# Creator
data "aws_caller_identity" "creator" {}

resource "aws_vpc" "test" {
  cidr_block = "10.255.255.0/28"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpn_gateway" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_vpn_gateway_attachment" "test" {
  vpc_id         = "${aws_vpc.test.id}"
  vpn_gateway_id = "${aws_vpn_gateway.test.id}"
}

# Accepter
resource "aws_dx_gateway" "test" {
  provider = "aws.alternate"

  amazon_side_asn = %[2]d
  name            = %[1]q
}
`, rName, rBgpAsn)
}

func testAccDxCrossAccountGatewayAssociationConfig_basic(rName string, rBgpAsn int) string {
	return testAccDxCrossAccountGatewayAssociationConfig_base(rName, rBgpAsn) + fmt.Sprintf(`
# Creator
resource "aws_dx_gateway_association_proposal" "test" {
  dx_gateway_id               = "${aws_dx_gateway.test.id}"
  dx_gateway_owner_account_id = "${aws_dx_gateway.test.owner_account_id}"
  vpn_gateway_id              = "${aws_vpn_gateway.test.id}"
}

# Accepter
resource "aws_dx_cross_account_gateway_association" "test" {
  provider = "aws.alternate"

  proposal_id                  = "${aws_dx_gateway_association_proposal.test.id}"
  dx_gateway_id                = "${aws_dx_gateway.test.id}"
  vpn_gateway_owner_account_id = "${data.aws_caller_identity.creator.account_id}"
}
`)
}

func testAccDxCrossAccountGatewayAssociationConfig_allowedPrefixes(rName string, rBgpAsn int) string {
	return testAccDxCrossAccountGatewayAssociationConfig_base(rName, rBgpAsn) + fmt.Sprintf(`
# Creator
resource "aws_dx_gateway_association_proposal" "test" {
  dx_gateway_id               = "${aws_dx_gateway.test.id}"
  dx_gateway_owner_account_id = "${aws_dx_gateway.test.owner_account_id}"
  vpn_gateway_id              = "${aws_vpn_gateway.test.id}"

  allowed_prefixes = [
    "10.255.255.0/30",
    "10.255.255.8/30",
  ]
}

# Accepter
resource "aws_dx_cross_account_gateway_association" "test" {
  provider = "aws.alternate"

  proposal_id                  = "${aws_dx_gateway_association_proposal.test.id}"
  dx_gateway_id                = "${aws_dx_gateway.test.id}"
  vpn_gateway_owner_account_id = "${data.aws_caller_identity.creator.account_id}"

  allowed_prefixes = [
    "10.255.255.8/29",
  ]
}
`)
}

func testAccDxCrossAccountGatewayAssociationConfig_allowedPrefixesUpdated(rName string, rBgpAsn int) string {
	return testAccDxCrossAccountGatewayAssociationConfig_base(rName, rBgpAsn) + fmt.Sprintf(`
# Creator
resource "aws_dx_gateway_association_proposal" "test" {
  dx_gateway_id               = "${aws_dx_gateway.test.id}"
  dx_gateway_owner_account_id = "${aws_dx_gateway.test.owner_account_id}"
  vpn_gateway_id              = "${aws_vpn_gateway.test.id}"
}

# Accepter
resource "aws_dx_cross_account_gateway_association" "test" {
  provider = "aws.alternate"

  proposal_id                  = "${aws_dx_gateway_association_proposal.test.id}"
  dx_gateway_id                = "${aws_dx_gateway.test.id}"
  vpn_gateway_owner_account_id = "${data.aws_caller_identity.creator.account_id}"

  allowed_prefixes = [
    "10.255.255.0/30",
    "10.255.255.8/30",
  ]
}
`)
}
