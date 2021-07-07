package aws

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directconnect"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/directconnect/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

func resourceAwsDxGatewayAssociationProposal() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsDxGatewayAssociationProposalCreate,
		Read:   resourceAwsDxGatewayAssociationProposalRead,
		Delete: resourceAwsDxGatewayAssociationProposalDelete,

		Importer: &schema.ResourceImporter{
			State: resourceAwsDxGatewayAssociationProposalImport,
		},

		CustomizeDiff: customdiff.Sequence(
			// Accepting the proposal with overridden prefixes changes the returned RequestedAllowedPrefixesToDirectConnectGateway value (allowed_prefixes attribute).
			// We only want to force a new resource if this value changes and the current proposal state is "requested".
			customdiff.ForceNewIf("allowed_prefixes", func(_ context.Context, d *schema.ResourceDiff, meta interface{}) bool {

				log.Printf("[DEBUG] Checking diff for Direct Connect Gateway Association Proposal (%s) allowed_prefixes", d.Id())

				if len(strings.Join(strings.Fields(d.Id()), "")) < 1 {
					log.Printf("[WARN] Direct Connect Gateway Association Proposal Id not available (%s)", d.Id())
					log.Printf("[DEBUG] Direct Connect Gateway Association Proposal UpdatedKeys (%s)", strings.Join(d.UpdatedKeys(), "/"))
					// assume proposal is end-of-life, rely on Read func to test
					return false
				}

				conn := meta.(*AWSClient).dxconn

				proposal, err := describeDirectConnectGatewayAssociationProposal(conn, d.Id())
				if err != nil {
					log.Printf("[ERROR] Error reading Direct Connect Gateway Association Proposal (%s): %s", d.Id(), err)
					return false
				}

				if proposal == nil {
					// proposal maybe end-of-life and removed by AWS, existence checked in Read func
					return false
				}

				return aws.StringValue(proposal.ProposalState) == directconnect.GatewayAssociationProposalStateRequested
			}),
		),

		Schema: map[string]*schema.Schema{
			"allowed_prefixes": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"associated_gateway_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"associated_gateway_owner_account_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"associated_gateway_type": {
				Type:     schema.TypeString,
				Computed: true,
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
		},
	}
}

func resourceAwsDxGatewayAssociationProposalCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).dxconn

	allowedPrefixes := expandDirectConnectGatewayAssociationProposalAllowedPrefixes(d.Get("allowed_prefixes").(*schema.Set).List())
	input := &directconnect.CreateDirectConnectGatewayAssociationProposalInput{
		AddAllowedPrefixesToDirectConnectGateway: allowedPrefixes,
		DirectConnectGatewayId:                   aws.String(d.Get("dx_gateway_id").(string)),
		DirectConnectGatewayOwnerAccount:         aws.String(d.Get("dx_gateway_owner_account_id").(string)),
		GatewayId:                                aws.String(d.Get("associated_gateway_id").(string)),
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
	log.Printf("[DEBUG] Read Direct Connect Gateway Association Proposal: %s", d.Id())

	var proposal *directconnect.GatewayAssociationProposal

	conn := meta.(*AWSClient).dxconn

	trimmedId := strings.Join(strings.Fields(d.Id()), "")
	if len(trimmedId) > 0 {
		var err error
		proposal, err = describeDirectConnectGatewayAssociationProposal(conn, d.Id())

		if err != nil {
			return fmt.Errorf("error reading Direct Connect Gateway Association Proposal (%s): %s", d.Id(), err)
		}
	} else {
		log.Printf("[WARN] Direct Connect Gateway Association Proposal Id not available (%s)", d.Id())
	}

	if proposal == nil || len(trimmedId) < 1 {
		log.Printf("[WARN] Direct Connect Gateway Association Proposal (%s) not found, checking for associated gateway", d.Id())

		var dxGatewayId string
		if rawDGId, ok := d.GetOk("dx_gateway_id"); ok {
			dxGatewayId = rawDGId.(string)
		} else if rawDGId == nil {
			d.SetId("")
			return fmt.Errorf("error reading dx_gateway_id (%s) from Proposal state", d.Id())
		}

		var associatedGatewayId string
		if rawAGId, ok := d.GetOk("associated_gateway_id"); ok {
			associatedGatewayId = rawAGId.(string)
		} else if rawAGId == nil {
			d.SetId("")
			return fmt.Errorf("error reading associated_gateway_id (%s) from Proposal state", d.Id())
		}

		log.Printf("[DEBUG] looking for Direct Connect Gateway Association using dx_gateway_id (%s) and associated_gateway_id (%s) to validate Proposal state data", dxGatewayId, associatedGatewayId)
		assocRaw, state, err := getDxGatewayAssociation(conn, dxGatewayId, associatedGatewayId)()

		if err != nil {
			d.SetId("")
			return fmt.Errorf("error reading Direct Connect gateway association (%s) from Proposal state: %s", d.Id(), err)
		}

		if state == gatewayAssociationStateDeleted {
			log.Printf("[WARN] Direct Connect gateway association (%s/%s/%s) not found, removing from state", d.Id(), dxGatewayId, associatedGatewayId)
			d.SetId("")
			return nil
		}

		// once accepted, AWS will delete the proposal after after some time (days?)
		// in this case we don't need to create a new proposal, use metadata from the association
		// to artificially populate the missing proposal in state as if it was still there.
		log.Printf("[INFO] Direct Connect Gateway Association Proposal (%s) has reached end-of-life and has been removed by AWS.", d.Id())
		assoc := assocRaw.(*directconnect.GatewayAssociation)

		err = d.Set("allowed_prefixes", flattenDxRouteFilterPrefixes(assoc.AllowedPrefixesToDirectConnectGateway))
		if err != nil {
			return fmt.Errorf("error setting allowed_prefixes: %s", err)
		}

		d.Set("associated_gateway_id", assoc.AssociatedGateway.Id)
		d.Set("dx_gateway_id", assoc.DirectConnectGatewayId)
		d.Set("dx_gateway_owner_account_id", assoc.DirectConnectGatewayOwnerAccount)
	} else {
		log.Printf("[DEBUG] Direct Connect Gateway Association Proposal (%s) found, continuing as normal: %s", d.Id(), proposal.String())

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

		d.Set("associated_gateway_id", proposal.AssociatedGateway.Id)
		d.Set("associated_gateway_owner_account_id", proposal.AssociatedGateway.OwnerAccount)
		d.Set("associated_gateway_type", proposal.AssociatedGateway.Type)
		d.Set("dx_gateway_id", proposal.DirectConnectGatewayId)
		d.Set("dx_gateway_owner_account_id", proposal.DirectConnectGatewayOwnerAccount)

	}
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

func resourceAwsDxGatewayAssociationProposalImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	parts := strings.Split(strings.ToLower(d.Id()), "/")

	var proposalID, directConnectGatewayID, associatedGatewayID string
	switch n := len(parts); n {
	case 1:
		return []*schema.ResourceData{d}, nil

	case 3:
		proposalID = parts[0]
		directConnectGatewayID = parts[1]
		associatedGatewayID = parts[2]

		if directConnectGatewayID == "" || associatedGatewayID == "" {
			return nil, fmt.Errorf("Incorrect resource ID format: %q. DXGATEWAYID and TARGETGATEWAYID must not be empty strings", d.Id())
		}

		break

	default:
		return nil, fmt.Errorf("Incorrect resource ID format: %q. Expected PROPOSALID or PROPOSALID/DXGATEWAYID/TARGETGATEWAYID", d.Id())
	}

	conn := meta.(*AWSClient).dxconn

	if proposalID != "" {
		_, err := finder.GatewayAssociationProposalByID(conn, proposalID)

		if tfresource.NotFound(err) {
			// Proposal not found.
		} else if err != nil {
			return nil, err
		} else {
			// Proposal still exists.
			d.SetId(proposalID)

			return []*schema.ResourceData{d}, nil
		}
	}

	_, err := finder.GatewayAssociationByDirectConnectGatewayIDAndAssociatedGatewayID(conn, directConnectGatewayID, associatedGatewayID)

	if err != nil {
		return nil, err
	}

	d.SetId(proposalID)
	d.Set("associated_gateway_id", associatedGatewayID)
	d.Set("dx_gateway_id", directConnectGatewayID)

	return []*schema.ResourceData{d}, nil
}

func getDxGatewayAssociation(conn *directconnect.DirectConnect, dxGatewayId, associatedGatewayId string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {

		resp, err := conn.DescribeDirectConnectGatewayAssociations(&directconnect.DescribeDirectConnectGatewayAssociationsInput{
			AssociatedGatewayId:    &associatedGatewayId,
			DirectConnectGatewayId: &dxGatewayId,
		})

		if err != nil {
			return nil, "", err
		}

		n := len(resp.DirectConnectGatewayAssociations)
		switch n {
		case 0:
			return "", gatewayAssociationStateDeleted, nil

		case 1:
			assoc := resp.DirectConnectGatewayAssociations[0]

			if stateChangeError := aws.StringValue(assoc.StateChangeError); stateChangeError != "" {
				id := dxGatewayAssociationId(
					aws.StringValue(resp.DirectConnectGatewayAssociations[0].DirectConnectGatewayId),
					aws.StringValue(resp.DirectConnectGatewayAssociations[0].AssociatedGateway.Id))
				log.Printf("[INFO] Direct Connect gateway association (%s) state change error: %s", id, stateChangeError)
			}

			return assoc, aws.StringValue(assoc.AssociationState), nil

		default:
			return nil, "", fmt.Errorf("Found %d Direct Connect gateway associations for %s, expected 1", n, associatedGatewayId)
		}
	}
}
