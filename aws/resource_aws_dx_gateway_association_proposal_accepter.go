package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directconnect"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsDxGatewayAssociationProposalAccepter() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsDxGatewayAssociationProposalAccepterCreate,
		Read:   resourceAwsDxGatewayAssociationProposalAccepterRead,
		Delete: resourceAwsDxGatewayAssociationProposalAccepterDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"dx_gateway_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"override_allowed_prefixes": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"proposal_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"vpn_gateway_owner_account_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateAwsAccountId,
			},
		},
	}
}

func resourceAwsDxGatewayAssociationProposalAccepterCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).dxconn

	input := &directconnect.AcceptDirectConnectGatewayAssociationProposalInput{
		AssociatedGatewayOwnerAccount:                 aws.String(d.Get("vpn_gateway_owner_account_id").(string)),
		DirectConnectGatewayId:                        aws.String(d.Get("dx_gateway_id").(string)),
		OverrideAllowedPrefixesToDirectConnectGateway: expandDxRouteFilterPrefixes(d.Get("override_allowed_prefixes").(*schema.Set)),
		ProposalId: aws.String(d.Get("proposal_id").(string)),
	}

	log.Printf("[DEBUG] Accepting Direct Connect Gateway Association Proposal: %s", input)
	_, err := conn.AcceptDirectConnectGatewayAssociationProposal(input)

	if err != nil {
		return fmt.Errorf("error accepting Direct Connect Gateway Association Proposal: %s", err)
	}

	d.SetId(aws.StringValue(input.ProposalId))

	return resourceAwsDxGatewayAssociationProposalAccepterRead(d, meta)
}

func resourceAwsDxGatewayAssociationProposalAccepterRead(d *schema.ResourceData, meta interface{}) error {
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

	d.Set("dx_gateway_id", aws.StringValue(proposal.DirectConnectGatewayId))
	if err := d.Set("override_allowed_prefixes", flattenDxRouteFilterPrefixes(proposal.RequestedAllowedPrefixesToDirectConnectGateway)); err != nil {
		return fmt.Errorf("error setting override_allowed_prefixes: %s", err)
	}
	d.Set("proposal_id", aws.StringValue(proposal.ProposalId))
	d.Set("vpn_gateway_owner_account_id", aws.StringValue(proposal.AssociatedGateway.OwnerAccount))

	return nil
}

func resourceAwsDxGatewayAssociationProposalAccepterDelete(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[WARN] Will not delete Direct Connect Gateway Association Proposal Accepter. Terraform will remove this resource from the state file, however resources may remain.")
	return nil
}
