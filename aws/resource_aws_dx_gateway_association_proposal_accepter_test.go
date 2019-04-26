package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directconnect"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAwsDxGatewayAssociationProposalAccepter_basic(t *testing.T) {
	var proposal1 directconnect.GatewayAssociationProposal
	var providers []*schema.Provider
	rBgpAsn := randIntRange(64512, 65534)
	rName := fmt.Sprintf("terraform-testacc-dxgwassoc-accpt-%d", acctest.RandInt())
	resourceName := "aws_dx_gateway_association_proposal_accepter.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccAlternateAccountPreCheck(t)
		},
		ProviderFactories: testAccProviderFactories(&providers),
		CheckDestroy:      testAccCheckAwsDxGatewayAssociationProposalAccepterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDxGatewayAssociationProposalAccepterConfig(rName, rBgpAsn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsDxGatewayAssociationProposalAccepterExists(resourceName, &proposal1),
				),
			},
			{
				Config:            testAccDxGatewayAssociationProposalAccepterConfig(rName, rBgpAsn),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckAwsDxGatewayAssociationProposalAccepterDestroy(s *terraform.State) error {
	// We don't destroy the underlying Direct Connect Gateway Association Proposal Accepter.
	return nil
}

func testAccCheckAwsDxGatewayAssociationProposalAccepterExists(resourceName string, gatewayAssociationProposal *directconnect.GatewayAssociationProposal) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		conn := testAccProvider.Meta().(*AWSClient).dxconn

		proposal, err := describeDirectConnectGatewayAssociationProposal(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		if proposal == nil {
			return fmt.Errorf("Direct Connect Gateway Association Proposal (%s) not found", rs.Primary.ID)
		}
		if aws.StringValue(proposal.ProposalState) != directconnect.GatewayAssociationProposalStateAccepted {
			return fmt.Errorf("Direct Connect Gateway Association Proposal (%s) in incorrect state (%s)", rs.Primary.ID, aws.StringValue(proposal.ProposalState))
		}

		*gatewayAssociationProposal = *proposal

		return nil
	}
}

func testAccDxGatewayAssociationProposalAccepterConfigBase(rName string, rBgpAsn int) string {
	return testAccAlternateAccountProviderConfig() + fmt.Sprintf(`
# Creator
data "aws_caller_identity" "creator" {}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpn_gateway" "test" {
  vpc_id = "${aws_vpc.test.id}"

  tags = {
    Name = %[1]q
  }
}

# Accepter
resource "aws_dx_gateway" "test" {
  provider = "aws.alternate"

  amazon_side_asn = %[2]d
  name            = %[1]q
}
`, rName, rBgpAsn)
}

func testAccDxGatewayAssociationProposalAccepterConfig(rName string, rBgpAsn int) string {
	return testAccDxGatewayAssociationProposalAccepterConfigBase(rName, rBgpAsn) + fmt.Sprintf(`
# Creator
resource "aws_dx_gateway_association_proposal" "test" {
  dx_gateway_id               = "${aws_dx_gateway.test.id}"
  dx_gateway_owner_account_id = "${aws_dx_gateway.test.owner_account_id}"
  vpn_gateway_id              = "${aws_vpn_gateway.test.id}"
}

# Accepter
resource "aws_dx_gateway_association_proposal_accepter" "test" {
  provider = "aws.alternate"

  proposal_id                  = "${aws_dx_gateway_association_proposal.test.id}"
  dx_gateway_id                = "${aws_dx_gateway.test.id}"
  vpn_gateway_owner_account_id = "${data.aws_caller_identity.creator.account_id}"
}
`)
}
