package aws

import (
	"fmt"
	"log"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directconnect"
	multierror "github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/service/directconnect/finder"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/service/directconnect/lister"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/tfresource"
)

func init() {
	resource.AddTestSweepers("aws_dx_gateway_association_proposal", &resource.Sweeper{
		Name: "aws_dx_gateway_association_proposal",
		F:    testSweepDirectConnectGatewayAssociationProposals,
	})
}

func testSweepDirectConnectGatewayAssociationProposals(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.(*AWSClient).dxconn
	input := &directconnect.DescribeDirectConnectGatewayAssociationProposalsInput{}
	var sweeperErrs *multierror.Error
	sweepResources := make([]*testSweepResource, 0)

	err = lister.DescribeDirectConnectGatewayAssociationProposalsPages(conn, input, func(page *directconnect.DescribeDirectConnectGatewayAssociationProposalsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, proposal := range page.DirectConnectGatewayAssociationProposals {
			proposalID := aws.StringValue(proposal.ProposalId)

			if proposalRegion := aws.StringValue(proposal.AssociatedGateway.Region); proposalRegion != region {
				log.Printf("[INFO] Skipping Direct Connect Gateway Association Proposal (%s) in different home region: %s", proposalID, proposalRegion)
				continue
			}

			if state := aws.StringValue(proposal.ProposalState); state != directconnect.GatewayAssociationProposalStateAccepted {
				log.Printf("[INFO] Skipping Direct Connect Gateway Association Proposal (%s) in non-accepted (%s) state", proposalID, state)
				continue
			}

			r := resourceAwsDxGatewayAssociationProposal()
			d := r.Data(nil)
			d.SetId(proposalID)

			sweepResources = append(sweepResources, NewTestSweepResource(r, d, client))
		}

		return !lastPage
	})

	if testSweepSkipSweepError(err) {
		log.Print(fmt.Errorf("[WARN] Skipping Direct Connect Gateway Association Proposal sweep for %s: %w", region, err))
		return sweeperErrs // In case we have completed some pages, but had errors
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing Direct Connect Gateway Association Proposals (%s): %w", region, err))
	}

	err = testSweepResourceOrchestrator(sweepResources)

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error sweeping Direct Connect Gateway Association Proposals (%s): %w", region, err))
	}

	return sweeperErrs.ErrorOrNil()
}

func TestAccAwsDxGatewayAssociationProposal_basicVpnGateway(t *testing.T) {
	var proposal directconnect.GatewayAssociationProposal
	var providers []*schema.Provider
	rBgpAsn := acctest.RandIntRange(64512, 65534)
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_dx_gateway_association_proposal.test"
	resourceNameDxGw := "aws_dx_gateway.test"
	resourceNameVgw := "aws_vpn_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t); testAccAlternateAccountPreCheck(t) },
		ErrorCheck:        testAccErrorCheck(t, directconnect.EndpointsID),
		ProviderFactories: testAccProviderFactoriesAlternate(&providers),
		CheckDestroy:      testAccCheckAwsDxGatewayAssociationProposalDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDxGatewayAssociationProposalConfig_basicVpnGateway(rName, rBgpAsn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsDxGatewayAssociationProposalExists(resourceName, &proposal),
					resource.TestCheckResourceAttr(resourceName, "allowed_prefixes.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "associated_gateway_id", resourceNameVgw, "id"),
					testAccCheckResourceAttrAccountID(resourceName, "associated_gateway_owner_account_id"),
					resource.TestCheckResourceAttr(resourceName, "associated_gateway_type", "virtualPrivateGateway"),
					resource.TestCheckResourceAttrPair(resourceName, "dx_gateway_id", resourceNameDxGw, "id"),
				),
			},
			{
				Config:            testAccDxGatewayAssociationProposalConfig_basicVpnGateway(rName, rBgpAsn),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAwsDxGatewayAssociationProposal_basicTransitGateway(t *testing.T) {
	var proposal directconnect.GatewayAssociationProposal
	var providers []*schema.Provider
	rBgpAsn := acctest.RandIntRange(64512, 65534)
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_dx_gateway_association_proposal.test"
	resourceNameDxGw := "aws_dx_gateway.test"
	resourceNameTgw := "aws_ec2_transit_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t); testAccAlternateAccountPreCheck(t) },
		ErrorCheck:        testAccErrorCheck(t, directconnect.EndpointsID),
		ProviderFactories: testAccProviderFactoriesAlternate(&providers),
		CheckDestroy:      testAccCheckAwsDxGatewayAssociationProposalDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDxGatewayAssociationProposalConfig_basicTransitGateway(rName, rBgpAsn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsDxGatewayAssociationProposalExists(resourceName, &proposal),
					resource.TestCheckResourceAttr(resourceName, "allowed_prefixes.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "allowed_prefixes.*", "10.255.255.0/30"),
					resource.TestCheckTypeSetElemAttr(resourceName, "allowed_prefixes.*", "10.255.255.8/30"),
					resource.TestCheckResourceAttrPair(resourceName, "associated_gateway_id", resourceNameTgw, "id"),
					testAccCheckResourceAttrAccountID(resourceName, "associated_gateway_owner_account_id"),
					resource.TestCheckResourceAttr(resourceName, "associated_gateway_type", "transitGateway"),
					resource.TestCheckResourceAttrPair(resourceName, "dx_gateway_id", resourceNameDxGw, "id"),
				),
			},
			{
				Config:            testAccDxGatewayAssociationProposalConfig_basicVpnGateway(rName, rBgpAsn),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAwsDxGatewayAssociationProposal_disappears(t *testing.T) {
	var proposal directconnect.GatewayAssociationProposal
	var providers []*schema.Provider
	rBgpAsn := acctest.RandIntRange(64512, 65534)
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_dx_gateway_association_proposal.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t); testAccAlternateAccountPreCheck(t) },
		ErrorCheck:        testAccErrorCheck(t, directconnect.EndpointsID),
		ProviderFactories: testAccProviderFactoriesAlternate(&providers),
		CheckDestroy:      testAccCheckAwsDxGatewayAssociationProposalDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDxGatewayAssociationProposalConfig_basicVpnGateway(rName, rBgpAsn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsDxGatewayAssociationProposalExists(resourceName, &proposal),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsDxGatewayAssociationProposal(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAwsDxGatewayAssociationProposal_endOfLifeVpn(t *testing.T) {
	var proposal directconnect.GatewayAssociationProposal
	var providers []*schema.Provider
	rBgpAsn := acctest.RandIntRange(64512, 65534)
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_dx_gateway_association_proposal.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t); testAccAlternateAccountPreCheck(t) },
		ErrorCheck:        testAccErrorCheck(t, directconnect.EndpointsID),
		ProviderFactories: testAccProviderFactoriesAlternate(&providers),
		CheckDestroy:      testAccCheckAwsDxGatewayAssociationProposalDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDxGatewayAssociationProposalConfig_endOfLifeVpn(rName, rBgpAsn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsDxGatewayAssociationProposalExists(resourceName, &proposal),
					testAccCheckAwsDxGatewayAssociationProposalAccepted(resourceName),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsDxGatewayAssociationProposal(), resourceName),
				),
			},
			{
				ResourceName: resourceName,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					return strings.Join([]string{
						aws.StringValue(proposal.ProposalId),
						aws.StringValue(proposal.DirectConnectGatewayId),
						aws.StringValue(proposal.AssociatedGateway.Id),
					}, "/"), nil
				},
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAwsDxGatewayAssociationProposal_endOfLifeTgw(t *testing.T) {
	var proposal directconnect.GatewayAssociationProposal
	var providers []*schema.Provider
	rBgpAsn := acctest.RandIntRange(64512, 65534)
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_dx_gateway_association_proposal.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t); testAccAlternateAccountPreCheck(t) },
		ErrorCheck:        testAccErrorCheck(t, directconnect.EndpointsID),
		ProviderFactories: testAccProviderFactoriesAlternate(&providers),
		CheckDestroy:      testAccCheckAwsDxGatewayAssociationProposalDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDxGatewayAssociationProposalConfig_endOfLifeTgw(rName, rBgpAsn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsDxGatewayAssociationProposalExists(resourceName, &proposal),
					testAccCheckAwsDxGatewayAssociationProposalAccepted(resourceName),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsDxGatewayAssociationProposal(), resourceName),
				),
			},
			{
				ResourceName: resourceName,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					return strings.Join([]string{
						aws.StringValue(proposal.ProposalId),
						aws.StringValue(proposal.DirectConnectGatewayId),
						aws.StringValue(proposal.AssociatedGateway.Id),
					}, "/"), nil
				},
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAwsDxGatewayAssociationProposal_AllowedPrefixes(t *testing.T) {
	var proposal1, proposal2 directconnect.GatewayAssociationProposal
	var providers []*schema.Provider
	rBgpAsn := acctest.RandIntRange(64512, 65534)
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_dx_gateway_association_proposal.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t); testAccAlternateAccountPreCheck(t) },
		ErrorCheck:        testAccErrorCheck(t, directconnect.EndpointsID),
		ProviderFactories: testAccProviderFactoriesAlternate(&providers),
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

		_, err := finder.GatewayAssociationProposalByID(conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("Direct Connect Gateway Association Proposal %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckAwsDxGatewayAssociationProposalExists(resourceName string, gatewayAssociationProposal *directconnect.GatewayAssociationProposal) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).dxconn

		output, err := finder.GatewayAssociationProposalByID(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*gatewayAssociationProposal = *output

		return nil
	}
}

func testAccCheckAwsDxGatewayAssociationProposalRecreated(old, new *directconnect.GatewayAssociationProposal) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if old, new := aws.StringValue(old.ProposalId), aws.StringValue(new.ProposalId); old == new {
			return fmt.Errorf("Direct Connect Gateway Association Proposal (%s) not recreated", old)
		}

		return nil
	}
}

func testAccCheckAwsDxGatewayAssociationProposalAccepted(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		conn := testAccProvider.Meta().(*AWSClient).dxconn

		output, err := finder.GatewayAssociationProposalByID(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		if state := aws.StringValue(output.ProposalState); state != directconnect.GatewayAssociationProposalStateAccepted {
			return fmt.Errorf("Direct Connect Gateway Association Proposal (%s) not accepted (%s)", rs.Primary.ID, state)
		}

		return nil
	}
}

func testAccDxGatewayAssociationProposalConfigBase_vpnGateway(rName string, rBgpAsn int) string {
	return composeConfig(testAccAlternateAccountProviderConfig(), fmt.Sprintf(`
resource "aws_dx_gateway" "test" {
  provider = "awsalternate"

  amazon_side_asn = %[2]d
  name            = %[1]q
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpn_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName, rBgpAsn))
}

func testAccDxGatewayAssociationProposalConfig_basicVpnGateway(rName string, rBgpAsn int) string {
	return composeConfig(testAccDxGatewayAssociationProposalConfigBase_vpnGateway(rName, rBgpAsn), `
resource "aws_dx_gateway_association_proposal" "test" {
  dx_gateway_id               = aws_dx_gateway.test.id
  dx_gateway_owner_account_id = aws_dx_gateway.test.owner_account_id
  associated_gateway_id       = aws_vpn_gateway.test.id
}
`)
}

func testAccDxGatewayAssociationProposalConfig_endOfLifeVpn(rName string, rBgpAsn int) string {
	return composeConfig(testAccDxGatewayAssociationProposalConfig_basicVpnGateway(rName, rBgpAsn), `
data "aws_caller_identity" "current" {}

resource "aws_dx_gateway_association" "test" {
  provider = "awsalternate"

  proposal_id                         = aws_dx_gateway_association_proposal.test.id
  dx_gateway_id                       = aws_dx_gateway.test.id
  associated_gateway_owner_account_id = data.aws_caller_identity.current.account_id
}
`)
}

func testAccDxGatewayAssociationProposalConfig_endOfLifeTgw(rName string, rBgpAsn int) string {
	return composeConfig(testAccDxGatewayAssociationProposalConfig_basicTransitGateway(rName, rBgpAsn), `
data "aws_caller_identity" "current" {}

resource "aws_dx_gateway_association" "test" {
  provider = "awsalternate"

  proposal_id                         = aws_dx_gateway_association_proposal.test.id
  dx_gateway_id                       = aws_dx_gateway.test.id
  associated_gateway_owner_account_id = data.aws_caller_identity.current.account_id
}
`)
}

func testAccDxGatewayAssociationProposalConfig_basicTransitGateway(rName string, rBgpAsn int) string {
	return composeConfig(testAccAlternateAccountProviderConfig(), fmt.Sprintf(`
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

resource "aws_dx_gateway_association_proposal" "test" {
  dx_gateway_id               = aws_dx_gateway.test.id
  dx_gateway_owner_account_id = aws_dx_gateway.test.owner_account_id
  associated_gateway_id       = aws_ec2_transit_gateway.test.id

  allowed_prefixes = [
    "10.255.255.0/30",
    "10.255.255.8/30",
  ]
}
`, rName, rBgpAsn))
}

func testAccDxGatewayAssociationProposalConfigAllowedPrefixes1(rName string, rBgpAsn int) string {
	return composeConfig(testAccDxGatewayAssociationProposalConfigBase_vpnGateway(rName, rBgpAsn), `
resource "aws_dx_gateway_association_proposal" "test" {
  allowed_prefixes            = ["10.0.0.0/16"]
  dx_gateway_id               = aws_dx_gateway.test.id
  dx_gateway_owner_account_id = aws_dx_gateway.test.owner_account_id
  associated_gateway_id       = aws_vpn_gateway.test.id
}
`)
}

func testAccDxGatewayAssociationProposalConfigAllowedPrefixes2(rName string, rBgpAsn int) string {
	return composeConfig(testAccDxGatewayAssociationProposalConfigBase_vpnGateway(rName, rBgpAsn), `
resource "aws_dx_gateway_association_proposal" "test" {
  allowed_prefixes            = ["10.0.0.0/24", "10.0.1.0/24"]
  dx_gateway_id               = aws_dx_gateway.test.id
  dx_gateway_owner_account_id = aws_dx_gateway.test.owner_account_id
  associated_gateway_id       = aws_vpn_gateway.test.id
}
`)
}
