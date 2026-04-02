// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package directconnect_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/directconnect/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfdirectconnect "github.com/hashicorp/terraform-provider-aws/internal/service/directconnect"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccDirectConnectGatewayAssociationProposal_basicVPNGateway(t *testing.T) {
	ctx := acctest.Context(t)
	var proposal awstypes.DirectConnectGatewayAssociationProposal
	rBgpAsn := acctest.RandIntRange(t, 64512, 65534)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_dx_gateway_association_proposal.test"
	resourceNameDxGw := "aws_dx_gateway.test"
	resourceNameVgw := "aws_vpn_gateway.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAlternateAccount(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DirectConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckGatewayAssociationProposalDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGatewayAssociationProposalConfig_basicVPN(rName, rBgpAsn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGatewayAssociationProposalExists(ctx, t, resourceName, &proposal),
					resource.TestCheckResourceAttr(resourceName, "allowed_prefixes.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "associated_gateway_id", resourceNameVgw, names.AttrID),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, "associated_gateway_owner_account_id"),
					resource.TestCheckResourceAttr(resourceName, "associated_gateway_type", "virtualPrivateGateway"),
					resource.TestCheckResourceAttrPair(resourceName, "dx_gateway_id", resourceNameDxGw, names.AttrID),
				),
			},
			{
				Config:            testAccGatewayAssociationProposalConfig_basicVPN(rName, rBgpAsn),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccDirectConnectGatewayAssociationProposal_basicTransitGateway(t *testing.T) {
	ctx := acctest.Context(t)
	var proposal awstypes.DirectConnectGatewayAssociationProposal
	rBgpAsn := acctest.RandIntRange(t, 64512, 65534)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_dx_gateway_association_proposal.test"
	resourceNameDxGw := "aws_dx_gateway.test"
	resourceNameTgw := "aws_ec2_transit_gateway.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAlternateAccount(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DirectConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckGatewayAssociationProposalDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGatewayAssociationProposalConfig_basicTransit(rName, rBgpAsn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGatewayAssociationProposalExists(ctx, t, resourceName, &proposal),
					resource.TestCheckResourceAttr(resourceName, "allowed_prefixes.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "allowed_prefixes.*", "10.255.255.0/30"),
					resource.TestCheckTypeSetElemAttr(resourceName, "allowed_prefixes.*", "10.255.255.8/30"),
					resource.TestCheckResourceAttrPair(resourceName, "associated_gateway_id", resourceNameTgw, names.AttrID),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, "associated_gateway_owner_account_id"),
					resource.TestCheckResourceAttr(resourceName, "associated_gateway_type", "transitGateway"),
					resource.TestCheckResourceAttrPair(resourceName, "dx_gateway_id", resourceNameDxGw, names.AttrID),
				),
			},
			{
				Config:            testAccGatewayAssociationProposalConfig_basicVPN(rName, rBgpAsn),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccDirectConnectGatewayAssociationProposal_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var proposal awstypes.DirectConnectGatewayAssociationProposal
	rBgpAsn := acctest.RandIntRange(t, 64512, 65534)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_dx_gateway_association_proposal.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAlternateAccount(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DirectConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckGatewayAssociationProposalDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGatewayAssociationProposalConfig_basicVPN(rName, rBgpAsn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGatewayAssociationProposalExists(ctx, t, resourceName, &proposal),
					acctest.CheckSDKResourceDisappears(ctx, t, tfdirectconnect.ResourceGatewayAssociationProposal(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccDirectConnectGatewayAssociationProposal_endOfLifeVPN(t *testing.T) {
	ctx := acctest.Context(t)
	var proposal awstypes.DirectConnectGatewayAssociationProposal
	rBgpAsn := acctest.RandIntRange(t, 64512, 65534)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_dx_gateway_association_proposal.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAlternateAccount(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DirectConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckGatewayAssociationProposalDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGatewayAssociationProposalConfig_endOfLifeVPN(rName, rBgpAsn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGatewayAssociationProposalExists(ctx, t, resourceName, &proposal),
					testAccCheckGatewayAssociationProposalAccepted(ctx, t, resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfdirectconnect.ResourceGatewayAssociationProposal(), resourceName),
				),
			},
			{
				ResourceName: resourceName,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					return strings.Join([]string{
						aws.ToString(proposal.ProposalId),
						aws.ToString(proposal.DirectConnectGatewayId),
						aws.ToString(proposal.AssociatedGateway.Id),
					}, "/"), nil
				},
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccDirectConnectGatewayAssociationProposal_endOfLifeTgw(t *testing.T) {
	ctx := acctest.Context(t)
	var proposal awstypes.DirectConnectGatewayAssociationProposal
	rBgpAsn := acctest.RandIntRange(t, 64512, 65534)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_dx_gateway_association_proposal.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAlternateAccount(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DirectConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckGatewayAssociationProposalDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGatewayAssociationProposalConfig_endOfLifeTgw(rName, rBgpAsn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGatewayAssociationProposalExists(ctx, t, resourceName, &proposal),
					testAccCheckGatewayAssociationProposalAccepted(ctx, t, resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfdirectconnect.ResourceGatewayAssociationProposal(), resourceName),
				),
			},
			{
				ResourceName: resourceName,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					return strings.Join([]string{
						aws.ToString(proposal.ProposalId),
						aws.ToString(proposal.DirectConnectGatewayId),
						aws.ToString(proposal.AssociatedGateway.Id),
					}, "/"), nil
				},
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccDirectConnectGatewayAssociationProposal_allowedPrefixes(t *testing.T) {
	ctx := acctest.Context(t)
	var proposal1, proposal2 awstypes.DirectConnectGatewayAssociationProposal
	rBgpAsn := acctest.RandIntRange(t, 64512, 65534)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_dx_gateway_association_proposal.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAlternateAccount(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DirectConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckGatewayAssociationProposalDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGatewayAssociationProposalConfig_allowedPrefixes1(rName, rBgpAsn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGatewayAssociationProposalExists(ctx, t, resourceName, &proposal1),
					resource.TestCheckResourceAttr(resourceName, "allowed_prefixes.#", "1"),
				),
			},
			{
				Config:            testAccGatewayAssociationProposalConfig_allowedPrefixes1(rName, rBgpAsn),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccGatewayAssociationProposalConfig_allowedPrefixes2(rName, rBgpAsn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGatewayAssociationProposalExists(ctx, t, resourceName, &proposal2),
					testAccCheckGatewayAssociationProposalRecreated(&proposal1, &proposal2),
					resource.TestCheckResourceAttr(resourceName, "allowed_prefixes.#", "2"),
				),
			},
		},
	})
}

func testAccCheckGatewayAssociationProposalDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).DirectConnectClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_dx_gateway_association_proposal" {
				continue
			}

			_, err := tfdirectconnect.FindGatewayAssociationProposalByID(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Direct Connect Gateway Association Proposal %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckGatewayAssociationProposalExists(ctx context.Context, t *testing.T, n string, v *awstypes.DirectConnectGatewayAssociationProposal) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).DirectConnectClient(ctx)

		output, err := tfdirectconnect.FindGatewayAssociationProposalByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckGatewayAssociationProposalRecreated(old, new *awstypes.DirectConnectGatewayAssociationProposal) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if old, new := aws.ToString(old.ProposalId), aws.ToString(new.ProposalId); old == new {
			return fmt.Errorf("Direct Connect Gateway Association Proposal (%s) not recreated", old)
		}

		return nil
	}
}

func testAccCheckGatewayAssociationProposalAccepted(ctx context.Context, t *testing.T, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		conn := acctest.ProviderMeta(ctx, t).DirectConnectClient(ctx)

		output, err := tfdirectconnect.FindGatewayAssociationProposalByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		if output.ProposalState != awstypes.DirectConnectGatewayAssociationProposalStateAccepted {
			return fmt.Errorf("Direct Connect Gateway Association Proposal (%s) not accepted (%s)", rs.Primary.ID, string(output.ProposalState))
		}

		return nil
	}
}

func testAccGatewayAssociationProposalConfigBase_vpnGateway(rName string, rBgpAsn int) string {
	return acctest.ConfigCompose(acctest.ConfigAlternateAccountProvider(), fmt.Sprintf(`
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

func testAccGatewayAssociationProposalConfig_basicVPN(rName string, rBgpAsn int) string {
	return acctest.ConfigCompose(testAccGatewayAssociationProposalConfigBase_vpnGateway(rName, rBgpAsn), `
resource "aws_dx_gateway_association_proposal" "test" {
  dx_gateway_id               = aws_dx_gateway.test.id
  dx_gateway_owner_account_id = aws_dx_gateway.test.owner_account_id
  associated_gateway_id       = aws_vpn_gateway.test.id
}
`)
}

func testAccGatewayAssociationProposalConfig_endOfLifeVPN(rName string, rBgpAsn int) string {
	return acctest.ConfigCompose(testAccGatewayAssociationProposalConfig_basicVPN(rName, rBgpAsn), `
data "aws_caller_identity" "current" {}

resource "aws_dx_gateway_association" "test" {
  provider = "awsalternate"

  proposal_id                         = aws_dx_gateway_association_proposal.test.id
  dx_gateway_id                       = aws_dx_gateway.test.id
  associated_gateway_owner_account_id = data.aws_caller_identity.current.account_id
}
`)
}

func testAccGatewayAssociationProposalConfig_endOfLifeTgw(rName string, rBgpAsn int) string {
	return acctest.ConfigCompose(testAccGatewayAssociationProposalConfig_basicTransit(rName, rBgpAsn), `
data "aws_caller_identity" "current" {}

resource "aws_dx_gateway_association" "test" {
  provider = "awsalternate"

  proposal_id                         = aws_dx_gateway_association_proposal.test.id
  dx_gateway_id                       = aws_dx_gateway.test.id
  associated_gateway_owner_account_id = data.aws_caller_identity.current.account_id
}
`)
}

func testAccGatewayAssociationProposalConfig_basicTransit(rName string, rBgpAsn int) string {
	return acctest.ConfigCompose(acctest.ConfigAlternateAccountProvider(), fmt.Sprintf(`
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

func testAccGatewayAssociationProposalConfig_allowedPrefixes1(rName string, rBgpAsn int) string {
	return acctest.ConfigCompose(testAccGatewayAssociationProposalConfigBase_vpnGateway(rName, rBgpAsn), `
resource "aws_dx_gateway_association_proposal" "test" {
  allowed_prefixes            = ["10.0.0.0/16"]
  dx_gateway_id               = aws_dx_gateway.test.id
  dx_gateway_owner_account_id = aws_dx_gateway.test.owner_account_id
  associated_gateway_id       = aws_vpn_gateway.test.id
}
`)
}

func testAccGatewayAssociationProposalConfig_allowedPrefixes2(rName string, rBgpAsn int) string {
	return acctest.ConfigCompose(testAccGatewayAssociationProposalConfigBase_vpnGateway(rName, rBgpAsn), `
resource "aws_dx_gateway_association_proposal" "test" {
  allowed_prefixes            = ["10.0.0.0/24", "10.0.1.0/24"]
  dx_gateway_id               = aws_dx_gateway.test.id
  dx_gateway_owner_account_id = aws_dx_gateway.test.owner_account_id
  associated_gateway_id       = aws_vpn_gateway.test.id
}
`)
}
