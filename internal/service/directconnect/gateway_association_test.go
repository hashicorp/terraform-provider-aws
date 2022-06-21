package directconnect_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directconnect"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfdirectconnect "github.com/hashicorp/terraform-provider-aws/internal/service/directconnect"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// V0 state upgrade testing must be done via acceptance testing due to API call
func TestAccDirectConnectGatewayAssociation_v0StateUpgrade(t *testing.T) {
	resourceName := "aws_dx_gateway_association.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rBgpAsn := sdkacctest.RandIntRange(64512, 65534)
	var ga directconnect.GatewayAssociation
	var gap directconnect.GatewayAssociationProposal

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, directconnect.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGatewayAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGatewayAssociationConfig_basicVPNSingleAccount(rName, rBgpAsn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGatewayAssociationExists(resourceName, &ga, &gap),
					testAccCheckGatewayAssociationStateUpgradeV0(resourceName),
				),
			},
		},
	})
}

func TestAccDirectConnectGatewayAssociation_basicVPNGatewaySingleAccount(t *testing.T) {
	resourceName := "aws_dx_gateway_association.test"
	resourceNameDxGw := "aws_dx_gateway.test"
	resourceNameVgw := "aws_vpn_gateway.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rBgpAsn := sdkacctest.RandIntRange(64512, 65534)
	var ga directconnect.GatewayAssociation
	var gap directconnect.GatewayAssociationProposal

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, directconnect.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGatewayAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGatewayAssociationConfig_basicVPNSingleAccount(rName, rBgpAsn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGatewayAssociationExists(resourceName, &ga, &gap),
					resource.TestCheckResourceAttr(resourceName, "allowed_prefixes.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "allowed_prefixes.*", "10.255.255.0/28"),
					resource.TestCheckResourceAttrPair(resourceName, "associated_gateway_id", resourceNameVgw, "id"),
					acctest.CheckResourceAttrAccountID(resourceName, "associated_gateway_owner_account_id"),
					resource.TestCheckResourceAttr(resourceName, "associated_gateway_type", "virtualPrivateGateway"),
					resource.TestCheckResourceAttrSet(resourceName, "dx_gateway_association_id"),
					resource.TestCheckResourceAttrPair(resourceName, "dx_gateway_id", resourceNameDxGw, "id"),
					acctest.CheckResourceAttrAccountID(resourceName, "dx_gateway_owner_account_id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccGatewayAssociationImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccDirectConnectGatewayAssociation_basicVPNGatewayCrossAccount(t *testing.T) {
	var providers []*schema.Provider
	resourceName := "aws_dx_gateway_association.test"
	resourceNameDxGw := "aws_dx_gateway.test"
	resourceNameVgw := "aws_vpn_gateway.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rBgpAsn := sdkacctest.RandIntRange(64512, 65534)
	var ga directconnect.GatewayAssociation
	var gap directconnect.GatewayAssociationProposal

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckAlternateAccount(t) },
		ErrorCheck:        acctest.ErrorCheck(t, directconnect.EndpointsID),
		ProviderFactories: acctest.FactoriesAlternate(&providers),
		CheckDestroy:      testAccCheckGatewayAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGatewayAssociationConfig_basicVPNCrossAccount(rName, rBgpAsn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGatewayAssociationExists(resourceName, &ga, &gap),
					resource.TestCheckResourceAttr(resourceName, "allowed_prefixes.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "allowed_prefixes.*", "10.255.255.0/28"),
					resource.TestCheckResourceAttrPair(resourceName, "associated_gateway_id", resourceNameVgw, "id"),
					acctest.CheckResourceAttrAccountID(resourceName, "associated_gateway_owner_account_id"),
					resource.TestCheckResourceAttr(resourceName, "associated_gateway_type", "virtualPrivateGateway"),
					resource.TestCheckResourceAttrSet(resourceName, "dx_gateway_association_id"),
					resource.TestCheckResourceAttrPair(resourceName, "dx_gateway_id", resourceNameDxGw, "id"),
					// dx_gateway_owner_account_id is the "awsalternate" provider's account ID.
					// acctest.CheckResourceAttrAccountID(resourceName, "dx_gateway_owner_account_id"),
				),
			},
		},
	})
}

func TestAccDirectConnectGatewayAssociation_basicTransitGatewaySingleAccount(t *testing.T) {
	resourceName := "aws_dx_gateway_association.test"
	resourceNameDxGw := "aws_dx_gateway.test"
	resourceNameTgw := "aws_ec2_transit_gateway.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rBgpAsn := sdkacctest.RandIntRange(64512, 65534)
	var ga directconnect.GatewayAssociation
	var gap directconnect.GatewayAssociationProposal

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, directconnect.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGatewayAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGatewayAssociationConfig_basicTransitSingleAccount(rName, rBgpAsn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGatewayAssociationExists(resourceName, &ga, &gap),
					resource.TestCheckResourceAttr(resourceName, "allowed_prefixes.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "allowed_prefixes.*", "10.255.255.0/30"),
					resource.TestCheckTypeSetElemAttr(resourceName, "allowed_prefixes.*", "10.255.255.8/30"),
					resource.TestCheckResourceAttrPair(resourceName, "associated_gateway_id", resourceNameTgw, "id"),
					acctest.CheckResourceAttrAccountID(resourceName, "associated_gateway_owner_account_id"),
					resource.TestCheckResourceAttr(resourceName, "associated_gateway_type", "transitGateway"),
					resource.TestCheckResourceAttrSet(resourceName, "dx_gateway_association_id"),
					resource.TestCheckResourceAttrPair(resourceName, "dx_gateway_id", resourceNameDxGw, "id"),
					acctest.CheckResourceAttrAccountID(resourceName, "dx_gateway_owner_account_id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccGatewayAssociationImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccDirectConnectGatewayAssociation_basicTransitGatewayCrossAccount(t *testing.T) {
	var providers []*schema.Provider
	resourceName := "aws_dx_gateway_association.test"
	resourceNameDxGw := "aws_dx_gateway.test"
	resourceNameTgw := "aws_ec2_transit_gateway.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rBgpAsn := sdkacctest.RandIntRange(64512, 65534)
	var ga directconnect.GatewayAssociation
	var gap directconnect.GatewayAssociationProposal

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckAlternateAccount(t) },
		ErrorCheck:        acctest.ErrorCheck(t, directconnect.EndpointsID),
		ProviderFactories: acctest.FactoriesAlternate(&providers),
		CheckDestroy:      testAccCheckGatewayAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGatewayAssociationConfig_basicTransitCrossAccount(rName, rBgpAsn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGatewayAssociationExists(resourceName, &ga, &gap),
					resource.TestCheckResourceAttr(resourceName, "allowed_prefixes.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "allowed_prefixes.*", "10.255.255.0/30"),
					resource.TestCheckTypeSetElemAttr(resourceName, "allowed_prefixes.*", "10.255.255.8/30"),
					resource.TestCheckResourceAttrPair(resourceName, "associated_gateway_id", resourceNameTgw, "id"),
					acctest.CheckResourceAttrAccountID(resourceName, "associated_gateway_owner_account_id"),
					resource.TestCheckResourceAttr(resourceName, "associated_gateway_type", "transitGateway"),
					resource.TestCheckResourceAttrSet(resourceName, "dx_gateway_association_id"),
					resource.TestCheckResourceAttrPair(resourceName, "dx_gateway_id", resourceNameDxGw, "id"),
					// dx_gateway_owner_account_id is the "awsalternate" provider's account ID.
					// acctest.CheckResourceAttrAccountID(resourceName, "dx_gateway_owner_account_id"),
				),
			},
		},
	})
}

func TestAccDirectConnectGatewayAssociation_multiVPNGatewaysSingleAccount(t *testing.T) {
	resourceName1 := "aws_dx_gateway_association.test.0"
	resourceName2 := "aws_dx_gateway_association.test.1"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rBgpAsn := sdkacctest.RandIntRange(64512, 65534)
	var ga directconnect.GatewayAssociation
	var gap directconnect.GatewayAssociationProposal

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, directconnect.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGatewayAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGatewayConfig_associationMultiVPNSingleAccount(rName, rBgpAsn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGatewayAssociationExists(resourceName1, &ga, &gap),
					testAccCheckGatewayAssociationExists(resourceName2, &ga, &gap),
					resource.TestCheckResourceAttr(resourceName1, "allowed_prefixes.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName1, "allowed_prefixes.*", "10.255.255.0/28"),
					resource.TestCheckResourceAttrSet(resourceName1, "dx_gateway_association_id"),
					resource.TestCheckResourceAttr(resourceName2, "allowed_prefixes.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName2, "allowed_prefixes.*", "10.255.255.16/28"),
					resource.TestCheckResourceAttrSet(resourceName2, "dx_gateway_association_id"),
				),
			},
		},
	})
}

func TestAccDirectConnectGatewayAssociation_allowedPrefixesVPNGatewaySingleAccount(t *testing.T) {
	resourceName := "aws_dx_gateway_association.test"
	resourceNameDxGw := "aws_dx_gateway.test"
	resourceNameVgw := "aws_vpn_gateway.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rBgpAsn := sdkacctest.RandIntRange(64512, 65534)
	var ga directconnect.GatewayAssociation
	var gap directconnect.GatewayAssociationProposal

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, directconnect.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGatewayAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGatewayAssociationConfig_allowedPrefixesVPNSingleAccount(rName, rBgpAsn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGatewayAssociationExists(resourceName, &ga, &gap),
					resource.TestCheckResourceAttr(resourceName, "allowed_prefixes.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "allowed_prefixes.*", "10.255.255.0/30"),
					resource.TestCheckTypeSetElemAttr(resourceName, "allowed_prefixes.*", "10.255.255.8/30"),
					resource.TestCheckResourceAttrPair(resourceName, "associated_gateway_id", resourceNameVgw, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "dx_gateway_association_id"),
					resource.TestCheckResourceAttrPair(resourceName, "dx_gateway_id", resourceNameDxGw, "id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccGatewayAssociationImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccGatewayAssociationConfig_allowedPrefixesVPNSingleAccountUpdated(rName, rBgpAsn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGatewayAssociationExists(resourceName, &ga, &gap),
					resource.TestCheckResourceAttr(resourceName, "allowed_prefixes.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "allowed_prefixes.*", "10.255.255.8/29"),
				),
			},
		},
	})
}

func TestAccDirectConnectGatewayAssociation_allowedPrefixesVPNGatewayCrossAccount(t *testing.T) {
	var providers []*schema.Provider
	resourceName := "aws_dx_gateway_association.test"
	resourceNameDxGw := "aws_dx_gateway.test"
	resourceNameVgw := "aws_vpn_gateway.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rBgpAsn := sdkacctest.RandIntRange(64512, 65534)
	var ga directconnect.GatewayAssociation
	var gap directconnect.GatewayAssociationProposal

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckAlternateAccount(t) },
		ErrorCheck:        acctest.ErrorCheck(t, directconnect.EndpointsID),
		ProviderFactories: acctest.FactoriesAlternate(&providers),
		CheckDestroy:      testAccCheckGatewayAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGatewayAssociationConfig_allowedPrefixesVPNCrossAccount(rName, rBgpAsn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGatewayAssociationExists(resourceName, &ga, &gap),
					resource.TestCheckResourceAttr(resourceName, "allowed_prefixes.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "allowed_prefixes.*", "10.255.255.8/29"),
					resource.TestCheckResourceAttrPair(resourceName, "associated_gateway_id", resourceNameVgw, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "dx_gateway_association_id"),
					resource.TestCheckResourceAttrPair(resourceName, "dx_gateway_id", resourceNameDxGw, "id"),
				),
				// Accepting the proposal with overridden prefixes changes the returned RequestedAllowedPrefixesToDirectConnectGateway value (allowed_prefixes attribute).
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccGatewayAssociationConfig_allowedPrefixesVPNCrossAccountUpdated(rName, rBgpAsn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGatewayAssociationExists(resourceName, &ga, &gap),
					resource.TestCheckResourceAttr(resourceName, "allowed_prefixes.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "allowed_prefixes.*", "10.255.255.0/30"),
					resource.TestCheckTypeSetElemAttr(resourceName, "allowed_prefixes.*", "10.255.255.8/30"),
					resource.TestCheckResourceAttrPair(resourceName, "associated_gateway_id", resourceNameVgw, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "dx_gateway_association_id"),
					resource.TestCheckResourceAttrPair(resourceName, "dx_gateway_id", resourceNameDxGw, "id"),
				),
			},
		},
	})
}

func TestAccDirectConnectGatewayAssociation_recreateProposal(t *testing.T) {
	var providers []*schema.Provider
	resourceName := "aws_dx_gateway_association.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rBgpAsn := sdkacctest.RandIntRange(64512, 65534)
	var ga1, ga2 directconnect.GatewayAssociation
	var gap1, gap2 directconnect.GatewayAssociationProposal

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckAlternateAccount(t) },
		ErrorCheck:        acctest.ErrorCheck(t, directconnect.EndpointsID),
		ProviderFactories: acctest.FactoriesAlternate(&providers),
		CheckDestroy:      testAccCheckGatewayAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGatewayAssociationConfig_basicVPNCrossAccount(rName, rBgpAsn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGatewayAssociationExists(resourceName, &ga1, &gap1),
				),
			},
			{
				Config: testAccGatewayAssociationConfig_basicVPNCrossAccountUpdatedProposal(rName, rBgpAsn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGatewayAssociationExists(resourceName, &ga2, &gap2),
					testAccCheckGatewayAssociationNotRecreated(&ga1, &ga2),
					testAccCheckGatewayAssociationProposalRecreated(&gap1, &gap2),
				),
			},
		},
	})
}

func testAccGatewayAssociationImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not Found: %s", resourceName)
		}

		return fmt.Sprintf("%s/%s", rs.Primary.Attributes["dx_gateway_id"], rs.Primary.Attributes["associated_gateway_id"]), nil
	}
}

func testAccCheckGatewayAssociationDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).DirectConnectConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_dx_gateway_association" {
			continue
		}

		_, err := tfdirectconnect.FindGatewayAssociationByID(conn, rs.Primary.Attributes["dx_gateway_association_id"])

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("Direct Connect Gateway Association %s still exists", rs.Primary.ID)
	}
	return nil
}

func testAccCheckGatewayAssociationExists(name string, ga *directconnect.GatewayAssociation, gap *directconnect.GatewayAssociationProposal) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.Attributes["dx_gateway_association_id"] == "" {
			return fmt.Errorf("No Direct Connect Gateway Association ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).DirectConnectConn

		output, err := tfdirectconnect.FindGatewayAssociationByID(conn, rs.Primary.Attributes["dx_gateway_association_id"])

		if err != nil {
			return err
		}

		if proposalID := rs.Primary.Attributes["proposal_id"]; proposalID != "" {
			output, err := tfdirectconnect.FindGatewayAssociationProposalByID(conn, proposalID)

			if err != nil {
				return err
			}

			*gap = *output
		}

		*ga = *output

		return nil
	}
}

func testAccCheckGatewayAssociationNotRecreated(old, new *directconnect.GatewayAssociation) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if old, new := aws.StringValue(old.AssociationId), aws.StringValue(new.AssociationId); old != new {
			return fmt.Errorf("Direct Connect Gateway Association (%s) recreated (%s)", old, new)
		}

		return nil
	}
}

// Perform check in acceptance testing as this StateUpgrader requires an API call
func testAccCheckGatewayAssociationStateUpgradeV0(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		rawState := map[string]interface{}{
			"dx_gateway_id":  rs.Primary.Attributes["dx_gateway_id"],
			"vpn_gateway_id": rs.Primary.Attributes["associated_gateway_id"], // vpn_gateway_id was removed in 3.0, but older state still has it
		}

		updatedRawState, err := tfdirectconnect.GatewayAssociationStateUpgradeV0(context.Background(), rawState, acctest.Provider.Meta())

		if err != nil {
			return err
		}

		if got, want := updatedRawState["dx_gateway_association_id"], rs.Primary.Attributes["dx_gateway_association_id"]; got != want {
			return fmt.Errorf("Invalid dx_gateway_association_id attribute in migrated state. Expected %s, got %s", want, got)
		}

		return nil
	}
}

func testAccGatewayAssociationConfigBase_vpnGatewaySingleAccount(rName string, rBgpAsn int) string {
	return fmt.Sprintf(`
resource "aws_dx_gateway" "test" {
  name            = %[1]q
  amazon_side_asn = "%[2]d"
}

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
  vpc_id         = aws_vpc.test.id
  vpn_gateway_id = aws_vpn_gateway.test.id
}
`, rName, rBgpAsn)
}

func testAccGatewayAssociationConfigBase_vpnGatewayCrossAccount(rName string, rBgpAsn int) string {
	return acctest.ConfigCompose(
		acctest.ConfigAlternateAccountProvider(),
		fmt.Sprintf(`
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
  vpc_id         = aws_vpc.test.id
  vpn_gateway_id = aws_vpn_gateway.test.id
}

# Accepter
resource "aws_dx_gateway" "test" {
  provider = "awsalternate"

  amazon_side_asn = %[2]d
  name            = %[1]q
}
`, rName, rBgpAsn))
}

func testAccGatewayAssociationConfig_basicVPNSingleAccount(rName string, rBgpAsn int) string {
	return acctest.ConfigCompose(
		testAccGatewayAssociationConfigBase_vpnGatewaySingleAccount(rName, rBgpAsn),
		`
resource "aws_dx_gateway_association" "test" {
  dx_gateway_id         = aws_dx_gateway.test.id
  associated_gateway_id = aws_vpn_gateway_attachment.test.vpn_gateway_id
}
`)
}

func testAccGatewayAssociationConfig_basicVPNCrossAccount(rName string, rBgpAsn int) string {
	return acctest.ConfigCompose(
		testAccGatewayAssociationConfigBase_vpnGatewayCrossAccount(rName, rBgpAsn),
		`
# Creator
resource "aws_dx_gateway_association_proposal" "test" {
  dx_gateway_id               = aws_dx_gateway.test.id
  dx_gateway_owner_account_id = aws_dx_gateway.test.owner_account_id
  associated_gateway_id       = aws_vpn_gateway_attachment.test.vpn_gateway_id
}

# Accepter
resource "aws_dx_gateway_association" "test" {
  provider = "awsalternate"

  proposal_id                         = aws_dx_gateway_association_proposal.test.id
  dx_gateway_id                       = aws_dx_gateway.test.id
  associated_gateway_owner_account_id = data.aws_caller_identity.creator.account_id
}
`)
}

func testAccGatewayAssociationConfig_basicVPNCrossAccountUpdatedProposal(rName string, rBgpAsn int) string {
	return acctest.ConfigCompose(
		testAccGatewayAssociationConfigBase_vpnGatewayCrossAccount(rName, rBgpAsn),
		`
# Creator
resource "aws_dx_gateway_association_proposal" "test" {
  dx_gateway_id               = aws_dx_gateway.test.id
  dx_gateway_owner_account_id = aws_dx_gateway.test.owner_account_id
  associated_gateway_id       = aws_vpn_gateway_attachment.test.vpn_gateway_id
}

resource "aws_dx_gateway_association_proposal" "test2" {
  dx_gateway_id               = aws_dx_gateway.test.id
  dx_gateway_owner_account_id = aws_dx_gateway.test.owner_account_id
  associated_gateway_id       = aws_vpn_gateway_attachment.test.vpn_gateway_id
}

# Accepter
resource "aws_dx_gateway_association" "test" {
  provider = "awsalternate"

  proposal_id                         = aws_dx_gateway_association_proposal.test2.id
  dx_gateway_id                       = aws_dx_gateway.test.id
  associated_gateway_owner_account_id = data.aws_caller_identity.creator.account_id
}
`)
}

func testAccGatewayAssociationConfig_basicTransitSingleAccount(rName string, rBgpAsn int) string {
	return fmt.Sprintf(`
resource "aws_dx_gateway" "test" {
  name            = %[1]q
  amazon_side_asn = "%[2]d"
}

resource "aws_ec2_transit_gateway" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_dx_gateway_association" "test" {
  dx_gateway_id         = aws_dx_gateway.test.id
  associated_gateway_id = aws_ec2_transit_gateway.test.id

  allowed_prefixes = [
    "10.255.255.0/30",
    "10.255.255.8/30",
  ]
}
`, rName, rBgpAsn)
}

func testAccGatewayAssociationConfig_basicTransitCrossAccount(rName string, rBgpAsn int) string {
	return acctest.ConfigCompose(
		acctest.ConfigAlternateAccountProvider(),
		fmt.Sprintf(`
# Creator
data "aws_caller_identity" "creator" {}

# Accepter
resource "aws_dx_gateway" "test" {
  provider = "awsalternate"

  amazon_side_asn = %[2]d
  name            = %[1]q
}

resource "aws_ec2_transit_gateway" "test" {
  tags = {
    Name = %[1]q
  }
}

# Creator
resource "aws_dx_gateway_association_proposal" "test" {
  dx_gateway_id               = aws_dx_gateway.test.id
  dx_gateway_owner_account_id = aws_dx_gateway.test.owner_account_id
  associated_gateway_id       = aws_ec2_transit_gateway.test.id

  allowed_prefixes = [
    "10.255.255.0/30",
    "10.255.255.8/30",
  ]
}

# Accepter
resource "aws_dx_gateway_association" "test" {
  provider = "awsalternate"

  proposal_id                         = aws_dx_gateway_association_proposal.test.id
  dx_gateway_id                       = aws_dx_gateway.test.id
  associated_gateway_owner_account_id = data.aws_caller_identity.creator.account_id
}
`, rName, rBgpAsn))
}

func testAccGatewayConfig_associationMultiVPNSingleAccount(rName string, rBgpAsn int) string {
	return fmt.Sprintf(`
resource "aws_dx_gateway" "test" {
  name            = %[1]q
  amazon_side_asn = "%[2]d"
}

resource "aws_vpc" "test" {
  count = 2

  cidr_block = cidrsubnet("10.255.255.0/26", 2, count.index)

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpn_gateway" "test" {
  count = 2

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpn_gateway_attachment" "test" {
  count = 2

  vpc_id         = aws_vpc.test[count.index].id
  vpn_gateway_id = aws_vpn_gateway.test[count.index].id
}

resource "aws_dx_gateway_association" "test" {
  count = 2

  dx_gateway_id         = aws_dx_gateway.test.id
  associated_gateway_id = aws_vpn_gateway_attachment.test[count.index].vpn_gateway_id
}
`, rName, rBgpAsn)
}

func testAccGatewayAssociationConfig_allowedPrefixesVPNSingleAccount(rName string, rBgpAsn int) string {
	return acctest.ConfigCompose(
		testAccGatewayAssociationConfigBase_vpnGatewaySingleAccount(rName, rBgpAsn),
		`
resource "aws_dx_gateway_association" "test" {
  dx_gateway_id         = aws_dx_gateway.test.id
  associated_gateway_id = aws_vpn_gateway_attachment.test.vpn_gateway_id

  allowed_prefixes = [
    "10.255.255.0/30",
    "10.255.255.8/30",
  ]
}
`)
}

func testAccGatewayAssociationConfig_allowedPrefixesVPNSingleAccountUpdated(rName string, rBgpAsn int) string {
	return acctest.ConfigCompose(
		testAccGatewayAssociationConfigBase_vpnGatewaySingleAccount(rName, rBgpAsn),
		`
resource "aws_dx_gateway_association" "test" {
  dx_gateway_id         = aws_dx_gateway.test.id
  associated_gateway_id = aws_vpn_gateway_attachment.test.vpn_gateway_id

  allowed_prefixes = [
    "10.255.255.8/29",
  ]
}
`)
}

func testAccGatewayAssociationConfig_allowedPrefixesVPNCrossAccount(rName string, rBgpAsn int) string {
	return acctest.ConfigCompose(
		testAccGatewayAssociationConfigBase_vpnGatewayCrossAccount(rName, rBgpAsn),
		`
# Creator
resource "aws_dx_gateway_association_proposal" "test" {
  dx_gateway_id               = aws_dx_gateway.test.id
  dx_gateway_owner_account_id = aws_dx_gateway.test.owner_account_id
  associated_gateway_id       = aws_vpn_gateway_attachment.test.vpn_gateway_id

  allowed_prefixes = [
    "10.255.255.0/30",
    "10.255.255.8/30",
  ]
}

# Accepter
resource "aws_dx_gateway_association" "test" {
  provider = "awsalternate"

  proposal_id                         = aws_dx_gateway_association_proposal.test.id
  dx_gateway_id                       = aws_dx_gateway.test.id
  associated_gateway_owner_account_id = data.aws_caller_identity.creator.account_id

  allowed_prefixes = [
    "10.255.255.8/29",
  ]
}
`)
}

func testAccGatewayAssociationConfig_allowedPrefixesVPNCrossAccountUpdated(rName string, rBgpAsn int) string {
	return acctest.ConfigCompose(
		testAccGatewayAssociationConfigBase_vpnGatewayCrossAccount(rName, rBgpAsn),
		`
# Creator
resource "aws_dx_gateway_association_proposal" "test" {
  dx_gateway_id               = aws_dx_gateway.test.id
  dx_gateway_owner_account_id = aws_dx_gateway.test.owner_account_id
  associated_gateway_id       = aws_vpn_gateway_attachment.test.vpn_gateway_id
}

# Accepter
resource "aws_dx_gateway_association" "test" {
  provider = "awsalternate"

  proposal_id                         = aws_dx_gateway_association_proposal.test.id
  dx_gateway_id                       = aws_dx_gateway.test.id
  associated_gateway_owner_account_id = data.aws_caller_identity.creator.account_id

  allowed_prefixes = [
    "10.255.255.0/30",
    "10.255.255.8/30",
  ]
}
`)
}
