// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package directconnect

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directconnect"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv1"
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

	conn := client.DirectConnectConn(ctx)

	sweepResources := make([]sweep.Sweepable, 0)
	var sweeperErrs *multierror.Error

	input := &directconnect.DescribeConnectionsInput{}

	// DescribeConnections has no pagination support
	output, err := conn.DescribeConnectionsWithContext(ctx, input)

	if awsv1.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Direct Connect Connection sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil()
	}

	if err != nil {
		sweeperErr := fmt.Errorf("error listing Direct Connect Connections for %s: %w", region, err)
		log.Printf("[ERROR] %s", sweeperErr)
		sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
		return sweeperErrs.ErrorOrNil()
	}

	if output == nil {
		log.Printf("[WARN] Skipping Direct Connect Connection sweep for %s: empty response", region)
		return sweeperErrs.ErrorOrNil()
	}

	for _, connection := range output.Connections {
		if connection == nil {
			continue
		}

		id := aws.StringValue(connection.ConnectionId)

		r := ResourceConnection()
		d := r.Data(nil)
		d.SetId(id)

		sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error sweeping Direct Connect Connection: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepGatewayAssociationProposals(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.DirectConnectConn(ctx)
	input := &directconnect.DescribeDirectConnectGatewayAssociationProposalsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = describeGatewayAssociationProposalsPages(ctx, conn, input, func(page *directconnect.DescribeDirectConnectGatewayAssociationProposalsOutput, lastPage bool) bool {
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

			r := ResourceGatewayAssociationProposal()
			d := r.Data(nil)
			d.SetId(proposalID)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if awsv1.SkipSweepError(err) {
		log.Print(fmt.Errorf("[WARN] Skipping Direct Connect Gateway Association Proposal sweep for %s: %w", region, err))
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
	conn := client.DirectConnectConn(ctx)
	input := &directconnect.DescribeDirectConnectGatewaysInput{}
	var sweeperErrs *multierror.Error
	sweepResources := make([]sweep.Sweepable, 0)

	err = describeGatewaysPages(ctx, conn, input, func(page *directconnect.DescribeDirectConnectGatewaysOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, gateway := range page.DirectConnectGateways {
			directConnectGatewayID := aws.StringValue(gateway.DirectConnectGatewayId)

			input := &directconnect.DescribeDirectConnectGatewayAssociationsInput{
				DirectConnectGatewayId: aws.String(directConnectGatewayID),
			}

			err := describeGatewayAssociationsPages(ctx, conn, input, func(page *directconnect.DescribeDirectConnectGatewayAssociationsOutput, lastPage bool) bool {
				if page == nil {
					return !lastPage
				}

				for _, association := range page.DirectConnectGatewayAssociations {
					gatewayID := aws.StringValue(association.AssociatedGateway.Id)

					if aws.StringValue(association.AssociatedGateway.Region) != region {
						log.Printf("[INFO] Skipping Direct Connect Gateway (%s) Association (%s) in different home region: %s", directConnectGatewayID, gatewayID, aws.StringValue(association.AssociatedGateway.Region))
						continue
					}

					if state := aws.StringValue(association.AssociationState); state != directconnect.GatewayAssociationStateAssociated {
						log.Printf("[INFO] Skipping Direct Connect Gateway (%s) Association in non-available (%s) state: %s", directConnectGatewayID, state, gatewayID)
						continue
					}

					r := ResourceGatewayAssociation()
					d := r.Data(nil)
					d.SetId(GatewayAssociationCreateResourceID(directConnectGatewayID, gatewayID))
					d.Set("dx_gateway_association_id", association.AssociationId)

					sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
				}

				return !lastPage
			})

			if err != nil {
				sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing Direct Connect Gateway Associations (%s): %w", region, err))
			}
		}

		return !lastPage
	})

	if awsv1.SkipSweepError(err) {
		log.Print(fmt.Errorf("[WARN] Skipping Direct Connect Gateway Association sweep for %s: %w", region, err))
		return sweeperErrs // In case we have completed some pages, but had errors
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing Direct Connect Gateways (%s): %w", region, err))
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
			sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing EC2 Transit Gateways (%s): %w", region, err))
		}

		for _, transitGateway := range page.TransitGateways {
			if transitGateway.State == ec2types.TransitGatewayStateDeleted {
				continue
			}

			transitGatewayID := aws.StringValue(transitGateway.TransitGatewayId)

			input := &directconnect.DescribeDirectConnectGatewayAssociationsInput{
				AssociatedGatewayId: aws.String(transitGatewayID),
			}

			err := describeGatewayAssociationsPages(ctx, conn, input, func(page *directconnect.DescribeDirectConnectGatewayAssociationsOutput, lastPage bool) bool {
				if page == nil {
					return !lastPage
				}

				for _, association := range page.DirectConnectGatewayAssociations {
					directConnectGatewayID := aws.StringValue(association.DirectConnectGatewayId)

					if state := aws.StringValue(association.AssociationState); state != directconnect.GatewayAssociationStateAssociated {
						log.Printf("[INFO] Skipping Direct Connect Gateway (%s) Association in non-available (%s) state: %s", directConnectGatewayID, state, transitGatewayID)
						continue
					}

					r := ResourceGatewayAssociation()
					d := r.Data(nil)
					d.SetId(GatewayAssociationCreateResourceID(directConnectGatewayID, transitGatewayID))
					d.Set("dx_gateway_association_id", association.AssociationId)

					sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
				}

				return !lastPage
			})

			if err != nil {
				sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing Direct Connect Gateway Associations (%s): %w", region, err))
			}
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error sweeping Direct Connect Gateway Associations (%s): %w", region, err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepGateways(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.DirectConnectConn(ctx)
	input := &directconnect.DescribeDirectConnectGatewaysInput{}
	var sweeperErrs *multierror.Error
	sweepResources := make([]sweep.Sweepable, 0)

	err = describeGatewaysPages(ctx, conn, input, func(page *directconnect.DescribeDirectConnectGatewaysOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, gateway := range page.DirectConnectGateways {
			directConnectGatewayID := aws.StringValue(gateway.DirectConnectGatewayId)

			if state := aws.StringValue(gateway.DirectConnectGatewayState); state != directconnect.GatewayStateAvailable {
				log.Printf("[INFO] Skipping Direct Connect Gateway in non-available (%s) state: %s", state, directConnectGatewayID)
				continue
			}

			input := &directconnect.DescribeDirectConnectGatewayAssociationsInput{
				DirectConnectGatewayId: aws.String(directConnectGatewayID),
			}

			var associations bool

			err := describeGatewayAssociationsPages(ctx, conn, input, func(page *directconnect.DescribeDirectConnectGatewayAssociationsOutput, lastPage bool) bool {
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
				sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing Direct Connect Gateway Associations (%s): %w", region, err))
			}

			if associations {
				log.Printf("[INFO] Skipping Direct Connect Gateway with remaining associations: %s", directConnectGatewayID)
				continue
			}

			r := ResourceGateway()
			d := r.Data(nil)
			d.SetId(directConnectGatewayID)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if awsv1.SkipSweepError(err) {
		log.Print(fmt.Errorf("[WARN] Skipping Direct Connect Gateway sweep for %s: %w", region, err))
		return sweeperErrs // In case we have completed some pages, but had errors
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing Direct Connect Gateways (%s): %w", region, err))
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error sweeping Direct Connect Gateways (%s): %w", region, err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepLags(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.DirectConnectConn(ctx)

	sweepResources := make([]sweep.Sweepable, 0)
	var sweeperErrs *multierror.Error

	input := &directconnect.DescribeLagsInput{}

	// DescribeLags has no pagination support
	output, err := conn.DescribeLagsWithContext(ctx, input)

	if awsv1.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Direct Connect LAG sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil()
	}

	if err != nil {
		sweeperErr := fmt.Errorf("error listing Direct Connect LAGs for %s: %w", region, err)
		log.Printf("[ERROR] %s", sweeperErr)
		sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
		return sweeperErrs.ErrorOrNil()
	}

	if output == nil {
		log.Printf("[WARN] Skipping Direct Connect LAG sweep for %s: empty response", region)
		return sweeperErrs.ErrorOrNil()
	}

	for _, lag := range output.Lags {
		if lag == nil {
			continue
		}

		id := aws.StringValue(lag.LagId)

		r := ResourceLag()
		d := r.Data(nil)
		d.SetId(id)

		sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error sweeping Direct Connect LAG: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepMacSecKeys(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	dxConn := client.DirectConnectConn(ctx)

	// Clean up leaked Secrets Manager resources created by Direct Connect.
	// Direct Connect does not remove the corresponding Secrets Manager
	// key when deleting the MACsec key association. The only option to
	// clean up the dangling resource is to use Secrets Manager to delete
	// the MACsec key secret.
	smConn := client.SecretsManagerClient(ctx)
	dxInput := &directconnect.DescribeConnectionsInput{}
	var sweeperErrs *multierror.Error

	output, err := dxConn.DescribeConnectionsWithContext(ctx, dxInput)

	if awsv1.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Direct Connect MACsec Keys sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil()
	}

	if err != nil {
		sweeperErr := fmt.Errorf(" listing Direct Connect Connections for %s: %w", region, err)
		log.Printf("[ERROR] %s", sweeperErr)
		sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
		return sweeperErrs.ErrorOrNil()
	}

	if output == nil {
		log.Printf("[WARN] Skipping Direct Connect MACsec Keys sweep for %s: empty response", region)
		return sweeperErrs.ErrorOrNil()
	}

	for _, connection := range output.Connections {
		if connection.MacSecKeys == nil {
			continue
		}

		for _, key := range connection.MacSecKeys {
			arn := aws.StringValue(key.SecretARN)

			input := &secretsmanager.DeleteSecretInput{
				SecretId: aws.String(arn),
			}

			log.Printf("[DEBUG] Deleting MACSec secret key: %s", *input.SecretId)
			_, err := smConn.DeleteSecret(ctx, input)

			if err != nil {
				sweeperErr := fmt.Errorf("error deleting MACsec Secret (%s): %w", arn, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}
		}
	}

	return sweeperErrs.ErrorOrNil()
}
