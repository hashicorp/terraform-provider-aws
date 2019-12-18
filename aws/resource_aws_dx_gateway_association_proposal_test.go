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

func TestAccAwsDxGatewayAssociationProposal_basic(t *testing.T) {
	var proposal1 directconnect.GatewayAssociationProposal
	var providers []*schema.Provider
	rBgpAsn := randIntRange(64512, 65534)
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_dx_gateway_association_proposal.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccAlternateAccountPreCheck(t)
		},
		ProviderFactories: testAccProviderFactories(&providers),
		CheckDestroy:      testAccCheckAwsDxGatewayAssociationProposalDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDxGatewayAssociationProposalConfig(rName, rBgpAsn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsDxGatewayAssociationProposalExists(resourceName, &proposal1),
					resource.TestCheckResourceAttr(resourceName, "allowed_prefixes.#", "1"),
				),
			},
			{
				Config:            testAccDxGatewayAssociationProposalConfig(rName, rBgpAsn),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAwsDxGatewayAssociationProposal_disappears(t *testing.T) {
	var proposal1 directconnect.GatewayAssociationProposal
	var providers []*schema.Provider
	rBgpAsn := randIntRange(64512, 65534)
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_dx_gateway_association_proposal.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccAlternateAccountPreCheck(t)
		},
		ProviderFactories: testAccProviderFactories(&providers),
		CheckDestroy:      testAccCheckAwsDxGatewayAssociationProposalDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDxGatewayAssociationProposalConfig(rName, rBgpAsn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsDxGatewayAssociationProposalExists(resourceName, &proposal1),
					testAccCheckAwsDxGatewayAssociationProposalDisappears(&proposal1),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAwsDxGatewayAssociationProposal_AllowedPrefixes(t *testing.T) {
	var proposal1, proposal2 directconnect.GatewayAssociationProposal
	var providers []*schema.Provider
	rBgpAsn := randIntRange(64512, 65534)
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_dx_gateway_association_proposal.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccAlternateAccountPreCheck(t)
		},
		ProviderFactories: testAccProviderFactories(&providers),
		CheckDestroy:      testAccCheckAwsDxGatewayAssociationProposalDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDxGatewayAssociationProposalConfigAllowedPrefixes1(rName, rBgpAsn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsDxGatewayAssociationProposalExists(resourceName, &proposal1),
					resource.TestCheckResourceAttr(resourceName, "allowed_prefixes.#", "1"),
				),
			},
			{
				Config:            testAccDxGatewayAssociationProposalConfigAllowedPrefixes1(rName, rBgpAsn),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDxGatewayAssociationProposalConfigAllowedPrefixes2(rName, rBgpAsn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsDxGatewayAssociationProposalExists(resourceName, &proposal2),
					testAccCheckAwsDxGatewayAssociationProposalRecreated(&proposal1, &proposal2),
					resource.TestCheckResourceAttr(resourceName, "allowed_prefixes.#", "2"),
				),
			},
		},
	})
}

func testAccCheckAwsDxGatewayAssociationProposalDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).dxconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_dx_gateway_association_proposal" {
			continue
		}

		proposal, err := describeDirectConnectGatewayAssociationProposal(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		if proposal == nil {
			continue
		}

		return fmt.Errorf("Direct Connect Gateway Association Proposal (%s) still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckAwsDxGatewayAssociationProposalExists(resourceName string, gatewayAssociationProposal *directconnect.GatewayAssociationProposal) resource.TestCheckFunc {
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

		*gatewayAssociationProposal = *proposal

		return nil
	}
}

func testAccCheckAwsDxGatewayAssociationProposalDisappears(proposal *directconnect.GatewayAssociationProposal) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).dxconn

		input := &directconnect.DeleteDirectConnectGatewayAssociationProposalInput{
			ProposalId: proposal.ProposalId,
		}

		_, err := conn.DeleteDirectConnectGatewayAssociationProposal(input)

		return err
	}
}

func testAccCheckAwsDxGatewayAssociationProposalRecreated(i, j *directconnect.GatewayAssociationProposal) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(i.ProposalId) == aws.StringValue(j.ProposalId) {
			return fmt.Errorf("Direct Connect Gateway Association Proposal not recreated")
		}

		return nil
	}
}

func testAccDxGatewayAssociationProposalConfigBase(rName string, rBgpAsn int) string {
	return testAccAlternateAccountProviderConfig() + fmt.Sprintf(`
resource "aws_dx_gateway" "test" {
  provider = "aws.alternate"

  amazon_side_asn = %[2]d
  name            = %[1]q
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "tf-acc-test-dx-gateway-association-proposal"
  }
}

resource "aws_vpn_gateway" "test" {
  vpc_id = "${aws_vpc.test.id}"

  tags = {
    Name = "tf-acc-test-dx-gateway-association-proposal"
  }
}
`, rName, rBgpAsn)
}

func testAccDxGatewayAssociationProposalConfig(rName string, rBgpAsn int) string {
	return testAccDxGatewayAssociationProposalConfigBase(rName, rBgpAsn) + fmt.Sprintf(`
resource "aws_dx_gateway_association_proposal" "test" {
  dx_gateway_id               = "${aws_dx_gateway.test.id}"
  dx_gateway_owner_account_id = "${aws_dx_gateway.test.owner_account_id}"
  vpn_gateway_id              = "${aws_vpn_gateway.test.id}"
}
`)
}

func testAccDxGatewayAssociationProposalConfigAllowedPrefixes1(rName string, rBgpAsn int) string {
	return testAccDxGatewayAssociationProposalConfigBase(rName, rBgpAsn) + fmt.Sprintf(`
resource "aws_dx_gateway_association_proposal" "test" {
  allowed_prefixes            = ["10.0.0.0/16"]
  dx_gateway_id               = "${aws_dx_gateway.test.id}"
  dx_gateway_owner_account_id = "${aws_dx_gateway.test.owner_account_id}"
  vpn_gateway_id              = "${aws_vpn_gateway.test.id}"
}
`)
}

func testAccDxGatewayAssociationProposalConfigAllowedPrefixes2(rName string, rBgpAsn int) string {
	return testAccDxGatewayAssociationProposalConfigBase(rName, rBgpAsn) + fmt.Sprintf(`
resource "aws_dx_gateway_association_proposal" "test" {
  allowed_prefixes            = ["10.0.0.0/24", "10.0.1.0/24"]
  dx_gateway_id               = "${aws_dx_gateway.test.id}"
  dx_gateway_owner_account_id = "${aws_dx_gateway.test.owner_account_id}"
  vpn_gateway_id              = "${aws_vpn_gateway.test.id}"
}
`)
}
