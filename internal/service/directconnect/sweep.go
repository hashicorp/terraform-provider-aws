// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package directconnect

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/directconnect"
	awstypes "github.com/aws/aws-sdk-go-v2/service/directconnect/types"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
)

func RegisterSweepers() {
	resource.AddTestSweepers("aws_dx_connection", &resource.Sweeper{
		Name: "aws_dx_connection",
		F:    sweepConnections,
	})

	resource.AddTestSweepers("aws_dx_gateway_association_proposal", &resource.Sweeper{
		Name: "aws_dx_gateway_association_proposal",
		F:    sweepGatewayAssociationProposals,
	})

	resource.AddTestSweepers("aws_dx_gateway_association", &resource.Sweeper{
		Name: "aws_dx_gateway_association",
		F:    sweepGatewayAssociations,
		Dependencies: []string{
			"aws_dx_gateway_association_proposal",
		},
	})

	resource.AddTestSweepers("aws_dx_gateway", &resource.Sweeper{
		Name: "aws_dx_gateway",
		F:    sweepGateways,
		Dependencies: []string{
			"aws_dx_gateway_association",
			"aws_networkmanager_dx_gateway_attachment",
		},
	})

	resource.AddTestSweepers("aws_dx_lag", &resource.Sweeper{
		Name:         "aws_dx_lag",
		F:            sweepLags,
		Dependencies: []string{"aws_dx_connection"},
	})

	resource.AddTestSweepers("aws_dx_macsec_key", &resource.Sweeper{
		Name:         "aws_dx_macsec_key",
		F:            sweepMacSecKeys,
		Dependencies: []string{},
	})
}

func sweepConnections(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	input := &directconnect.DescribeConnectionsInput{}
	conn := client.DirectConnectClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	output, err := conn.DescribeConnections(ctx, input)

	if awsv2.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Direct Connect Connection sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing Direct Connect Connections (%s): %w", region, err)
	}

	for _, v := range output.Connections {
		r := resourceConnection()
		d := r.Data(nil)
		d.SetId(aws.ToString(v.ConnectionId))

		sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Direct Connect Connections (%s): %w", region, err)
	}

	return nil
}

func sweepGatewayAssociationProposals(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	input := &directconnect.DescribeDirectConnectGatewayAssociationProposalsInput{}
	conn := client.DirectConnectClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	err = describeDirectConnectGatewayAssociationProposalsPages(ctx, conn, input, func(page *directconnect.DescribeDirectConnectGatewayAssociationProposalsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.DirectConnectGatewayAssociationProposals {
			id := aws.ToString(v.ProposalId)

			if proposalRegion := aws.ToString(v.AssociatedGateway.Region); proposalRegion != region {
				log.Printf("[INFO] Skipping Direct Connect Gateway Association Proposal %s: AssociatedGateway.Region=%s", id, proposalRegion)
				continue
			}

			if state := v.ProposalState; state != awstypes.DirectConnectGatewayAssociationProposalStateAccepted {
				log.Printf("[INFO] Skipping Direct Connect Gateway Association Proposal %s: ProposalState=%s", id, state)
				continue
			}

			r := resourceGatewayAssociationProposal()
			d := r.Data(nil)
			d.SetId(id)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if awsv2.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Direct Connect Gateway Association Proposal sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("listing Direct Connect Gateway Association Proposals (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("sweeping Direct Connect Gateway Association Proposals (%s): %w", region, err)
	}

	return nil
}

func sweepGatewayAssociations(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	input := &directconnect.DescribeDirectConnectGatewaysInput{}
	conn := client.DirectConnectClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	err = describeDirectConnectGatewaysPages(ctx, conn, input, func(page *directconnect.DescribeDirectConnectGatewaysOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.DirectConnectGateways {
			directConnectGatewayID := aws.ToString(v.DirectConnectGatewayId)

			input := &directconnect.DescribeDirectConnectGatewayAssociationsInput{
				DirectConnectGatewayId: aws.String(directConnectGatewayID),
			}

			err := describeDirectConnectGatewayAssociationsPages(ctx, conn, input, func(page *directconnect.DescribeDirectConnectGatewayAssociationsOutput, lastPage bool) bool {
				if page == nil {
					return !lastPage
				}

				for _, v := range page.DirectConnectGatewayAssociations {
					if v.AssociatedGateway == nil {
						continue
					}

					gatewayID := aws.ToString(v.AssociatedGateway.Id)

					if gatewayRegion := aws.ToString(v.AssociatedGateway.Region); gatewayRegion != region {
						log.Printf("[INFO] Skipping Direct Connect Gateway (%s) Association (%s): AssociatedGateway.Region=%s", directConnectGatewayID, gatewayID, gatewayRegion)
						continue
					}

					if state := v.AssociationState; state != awstypes.DirectConnectGatewayAssociationStateAssociated {
						log.Printf("[INFO] Skipping Direct Connect Gateway (%s) Association (%s): AssociationState=%s", directConnectGatewayID, gatewayID, state)
						continue
					}

					r := resourceGatewayAssociation()
					d := r.Data(nil)
					d.SetId(gatewayAssociationCreateResourceID(directConnectGatewayID, gatewayID))
					d.Set("dx_gateway_association_id", v.AssociationId)

					sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
				}

				return !lastPage
			})

			if err != nil {
				continue
			}
		}

		return !lastPage
	})

	if awsv2.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Direct Connect Gateway Association sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing Direct Connect Gateways (%s): %w", region, err)
	}

	// Handle cross-account EC2 Transit Gateway associations.
	// Direct Connect does not provide an easy lookup method for
	// these within the service itself so they can only be found
	// via AssociatedGatewayId of the EC2 Transit Gateway since the
	// DirectConnectGatewayId lives in the other account.
	ec2conn := client.EC2Client(ctx)

	pages := ec2.NewDescribeTransitGatewaysPaginator(ec2conn, &ec2.DescribeTransitGatewaysInput{})
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return fmt.Errorf("error listing EC2 Transit Gateways (%s): %w", region, err)
		}

		for _, v := range page.TransitGateways {
			if v.State == ec2types.TransitGatewayStateDeleted {
				continue
			}

			transitGatewayID := aws.ToString(v.TransitGatewayId)

			input := &directconnect.DescribeDirectConnectGatewayAssociationsInput{
				AssociatedGatewayId: aws.String(transitGatewayID),
			}

			err := describeDirectConnectGatewayAssociationsPages(ctx, conn, input, func(page *directconnect.DescribeDirectConnectGatewayAssociationsOutput, lastPage bool) bool {
				if page == nil {
					return !lastPage
				}

				for _, v := range page.DirectConnectGatewayAssociations {
					directConnectGatewayID := aws.ToString(v.DirectConnectGatewayId)

					if state := v.AssociationState; state != awstypes.DirectConnectGatewayAssociationStateAssociated {
						log.Printf("[INFO] Skipping Direct Connect Gateway (%s) Association (%s): %s", directConnectGatewayID, transitGatewayID, state)
						continue
					}

					r := resourceGatewayAssociation()
					d := r.Data(nil)
					d.SetId(gatewayAssociationCreateResourceID(directConnectGatewayID, transitGatewayID))
					d.Set("dx_gateway_association_id", v.AssociationId)

					sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
				}

				return !lastPage
			})

			if err != nil {
				continue
			}
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Direct Connect Gateway Associations (%s): %w", region, err)
	}

	return nil
}

func sweepGateways(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	input := &directconnect.DescribeDirectConnectGatewaysInput{}
	conn := client.DirectConnectClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	err = describeDirectConnectGatewaysPages(ctx, conn, input, func(page *directconnect.DescribeDirectConnectGatewaysOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.DirectConnectGateways {
			directConnectGatewayID := aws.ToString(v.DirectConnectGatewayId)

			if state := v.DirectConnectGatewayState; state != awstypes.DirectConnectGatewayStateAvailable {
				log.Printf("[INFO] Skipping Direct Connect Gateway (%s): DirectConnectGatewayState=%s", directConnectGatewayID, state)
				continue
			}

			input := &directconnect.DescribeDirectConnectGatewayAssociationsInput{
				DirectConnectGatewayId: aws.String(directConnectGatewayID),
			}

			var associations bool

			err := describeDirectConnectGatewayAssociationsPages(ctx, conn, input, func(page *directconnect.DescribeDirectConnectGatewayAssociationsOutput, lastPage bool) bool {
				if page == nil {
					return !lastPage
				}

				// If associations still remain, its likely that our region is not the home
				// region of those associations and the previous sweepers skipped them.
				// When we hit this condition, we skip trying to delete the gateway as it
				// will go from deleting -> available after a few minutes and timeout.
				if len(page.DirectConnectGatewayAssociations) > 0 {
					associations = true

					return false
				}

				return !lastPage
			})

			if err != nil {
				continue
			}

			if associations {
				log.Printf("[INFO] Skipping Direct Connect Gateway with remaining associations: %s", directConnectGatewayID)
				continue
			}

			r := resourceGateway()
			d := r.Data(nil)
			d.SetId(directConnectGatewayID)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if awsv2.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Direct Connect Gateway sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing Direct Connect Gateways (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Direct Connect Gateways (%s): %w", region, err)
	}

	return nil
}

func sweepLags(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	input := &directconnect.DescribeLagsInput{}
	conn := client.DirectConnectClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	output, err := conn.DescribeLags(ctx, input)

	if awsv2.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Direct Connect LAG sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing Direct Connect LAGs for %s: %w", region, err)
	}

	for _, v := range output.Lags {
		r := resourceLag()
		d := r.Data(nil)
		d.SetId(aws.ToString(v.LagId))

		sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Direct Connect LAGs (%s): %w", region, err)
	}

	return nil
}

func sweepMacSecKeys(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	input := &directconnect.DescribeConnectionsInput{}
	dxConn := client.DirectConnectClient(ctx)
	// Clean up leaked Secrets Manager resources created by Direct Connect.
	// Direct Connect does not remove the corresponding Secrets Manager
	// key when deleting the MACsec key association. The only option to
	// clean up the dangling resource is to use Secrets Manager to delete
	// the MACsec key secret.
	smConn := client.SecretsManagerClient(ctx)

	output, err := dxConn.DescribeConnections(ctx, input)

	if awsv2.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Direct Connect MACsec Keys sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing Direct Connect Connections (%s): %w", region, err)
	}

	for _, v := range output.Connections {
		for _, v := range v.MacSecKeys {
			arn := aws.ToString(v.SecretARN)

			input := &secretsmanager.DeleteSecretInput{
				SecretId: aws.String(arn),
			}

			log.Printf("[DEBUG] Deleting MACSec secret key: %s", arn)
			_, err := smConn.DeleteSecret(ctx, input)

			if err != nil {
				continue
			}
		}
	}

	return nil
}
