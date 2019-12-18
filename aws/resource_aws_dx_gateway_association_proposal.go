package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directconnect"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsDxGatewayAssociationProposal() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsDxGatewayAssociationProposalCreate,
		Read:   resourceAwsDxGatewayAssociationProposalRead,
		Delete: resourceAwsDxGatewayAssociationProposalDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"allowed_prefixes": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"dx_gateway_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"dx_gateway_owner_account_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateAwsAccountId,
			},
			"vpn_gateway_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceAwsDxGatewayAssociationProposalCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).dxconn

	input := &directconnect.CreateDirectConnectGatewayAssociationProposalInput{
		AddAllowedPrefixesToDirectConnectGateway: expandDirectConnectGatewayAssociationProposalAllowedPrefixes(d.Get("allowed_prefixes").(*schema.Set).List()),
		DirectConnectGatewayId:                   aws.String(d.Get("dx_gateway_id").(string)),
		DirectConnectGatewayOwnerAccount:         aws.String(d.Get("dx_gateway_owner_account_id").(string)),
		GatewayId:                                aws.String(d.Get("vpn_gateway_id").(string)),
	}

	log.Printf("[DEBUG] Creating Direct Connect Gateway Association Proposal: %s", input)
	output, err := conn.CreateDirectConnectGatewayAssociationProposal(input)

	if err != nil {
		return fmt.Errorf("error creating Direct Connect Gateway Association Proposal: %s", err)
	}

	d.SetId(aws.StringValue(output.DirectConnectGatewayAssociationProposal.ProposalId))

	return resourceAwsDxGatewayAssociationProposalRead(d, meta)
}

func resourceAwsDxGatewayAssociationProposalRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).dxconn

	proposal, err := describeDirectConnectGatewayAssociationProposal(conn, d.Id())

	if err != nil {
		return fmt.Errorf("error reading Direct Connect Gateway Association Proposal (%s): %s", d.Id(), err)
	}

	if proposal == nil {
		log.Printf("[WARN] Direct Connect Gateway Association Proposal (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if aws.StringValue(proposal.ProposalState) == directconnect.GatewayAssociationProposalStateDeleted {
		log.Printf("[WARN] Direct Connect Gateway Association Proposal (%s) deleted, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if proposal.AssociatedGateway == nil {
		return fmt.Errorf("error reading Direct Connect Gateway Association Proposal (%s): missing associated gateway information", d.Id())
	}

	if err := d.Set("allowed_prefixes", flattenDirectConnectGatewayAssociationProposalAllowedPrefixes(proposal.RequestedAllowedPrefixesToDirectConnectGateway)); err != nil {
		return fmt.Errorf("error setting allowed_prefixes: %s", err)
	}

	d.Set("dx_gateway_id", aws.StringValue(proposal.DirectConnectGatewayId))
	d.Set("dx_gateway_owner_account_id", aws.StringValue(proposal.DirectConnectGatewayOwnerAccount))
	d.Set("vpn_gateway_id", aws.StringValue(proposal.AssociatedGateway.Id))

	return nil
}

func resourceAwsDxGatewayAssociationProposalDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).dxconn

	input := &directconnect.DeleteDirectConnectGatewayAssociationProposalInput{
		ProposalId: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Deleting Direct Connect Gateway Association Proposal: %s", d.Id())

	_, err := conn.DeleteDirectConnectGatewayAssociationProposal(input)

	if err != nil {
		return fmt.Errorf("error deleting Direct Connect Gateway Association Proposal (%s): %s", d.Id(), err)
	}

	return nil
}

func describeDirectConnectGatewayAssociationProposal(conn *directconnect.DirectConnect, proposalID string) (*directconnect.GatewayAssociationProposal, error) {
	input := &directconnect.DescribeDirectConnectGatewayAssociationProposalsInput{
		ProposalId: aws.String(proposalID),
	}

	for {
		output, err := conn.DescribeDirectConnectGatewayAssociationProposals(input)

		if err != nil {
			return nil, err
		}

		if output == nil {
			continue
		}

		for _, proposal := range output.DirectConnectGatewayAssociationProposals {
			if aws.StringValue(proposal.ProposalId) == proposalID {
				return proposal, nil
			}
		}

		if aws.StringValue(output.NextToken) == "" {
			break
		}

		input.NextToken = output.NextToken
	}

	return nil, nil
}

func expandDirectConnectGatewayAssociationProposalAllowedPrefixes(allowedPrefixes []interface{}) []*directconnect.RouteFilterPrefix {
	if len(allowedPrefixes) == 0 {
		return nil
	}

	var routeFilterPrefixes []*directconnect.RouteFilterPrefix

	for _, allowedPrefixRaw := range allowedPrefixes {
		if allowedPrefixRaw == nil {
			continue
		}

		routeFilterPrefix := &directconnect.RouteFilterPrefix{
			Cidr: aws.String(allowedPrefixRaw.(string)),
		}

		routeFilterPrefixes = append(routeFilterPrefixes, routeFilterPrefix)
	}

	return routeFilterPrefixes
}

func flattenDirectConnectGatewayAssociationProposalAllowedPrefixes(routeFilterPrefixes []*directconnect.RouteFilterPrefix) []interface{} {
	if len(routeFilterPrefixes) == 0 {
		return []interface{}{}
	}

	var allowedPrefixes []interface{}

	for _, routeFilterPrefix := range routeFilterPrefixes {
		if routeFilterPrefix == nil {
			continue
		}

		allowedPrefix := aws.StringValue(routeFilterPrefix.Cidr)

		allowedPrefixes = append(allowedPrefixes, allowedPrefix)
	}

	return allowedPrefixes
}
