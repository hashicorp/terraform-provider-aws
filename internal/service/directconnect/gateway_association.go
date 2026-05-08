// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package directconnect

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/directconnect"
	awstypes "github.com/aws/aws-sdk-go-v2/service/directconnect/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_dx_gateway_association", name="Gateway Association")
func resourceGatewayAssociation() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceGatewayAssociationCreate,
		ReadWithoutTimeout:   resourceGatewayAssociationRead,
		UpdateWithoutTimeout: resourceGatewayAssociationUpdate,
		DeleteWithoutTimeout: resourceGatewayAssociationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceGatewayAssociationImport,
		},

		SchemaVersion: 2,
		StateUpgraders: []schema.StateUpgrader{
			{
				Type:    resourceGatewayAssociationResourceV0().CoreConfigSchema().ImpliedType(),
				Upgrade: gatewayAssociationStateUpgradeV0,
				Version: 0,
			},
			{
				Type:    resourceGatewayAssociationResourceV1().CoreConfigSchema().ImpliedType(),
				Upgrade: gatewayAssociationStateUpgradeV1,
				Version: 1,
			},
		},

		Schema: map[string]*schema.Schema{
			"allowed_prefixes": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"associated_gateway_id": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"associated_gateway_owner_account_id", "proposal_id"},
				AtLeastOneOf:  []string{"associated_gateway_id", "associated_gateway_owner_account_id", "proposal_id"},
			},
			"associated_gateway_owner_account_id": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ValidateFunc:  verify.ValidAccountID,
				ConflictsWith: []string{"associated_gateway_id"},
				RequiredWith:  []string{"proposal_id"},
				AtLeastOneOf:  []string{"associated_gateway_id", "associated_gateway_owner_account_id", "proposal_id"},
			},
			"associated_gateway_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"dx_gateway_association_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"dx_gateway_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"dx_gateway_owner_account_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"proposal_id": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"associated_gateway_id"},
				AtLeastOneOf:  []string{"associated_gateway_id", "associated_gateway_owner_account_id", "proposal_id"},
			},
			names.AttrTransitGatewayAttachmentID: {
				Type:     schema.TypeString,
				Computed: true,
			},
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},
	}
}

func resourceGatewayAssociationCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DirectConnectClient(ctx)

	var associationID string
	directConnectGatewayID := d.Get("dx_gateway_id").(string)

	if associatedGatewayOwnerAccount := d.Get("associated_gateway_owner_account_id").(string); associatedGatewayOwnerAccount != "" {
		proposalID := d.Get("proposal_id").(string)
		input := directconnect.AcceptDirectConnectGatewayAssociationProposalInput{
			AssociatedGatewayOwnerAccount: aws.String(associatedGatewayOwnerAccount),
			DirectConnectGatewayId:        aws.String(directConnectGatewayID),
			ProposalId:                    aws.String(proposalID),
		}

		if v, ok := d.GetOk("allowed_prefixes"); ok && v.(*schema.Set).Len() > 0 {
			input.OverrideAllowedPrefixesToDirectConnectGateway = expandRouteFilterPrefixes(v.(*schema.Set).List())
		}

		output, err := conn.AcceptDirectConnectGatewayAssociationProposal(ctx, &input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "accepting Direct Connect Gateway Association Proposal (%s): %s", proposalID, err)
		}

		// For historical reasons the resource ID isn't set to the association ID returned from the API.
		associationID = aws.ToString(output.DirectConnectGatewayAssociation.AssociationId)
		d.SetId(gatewayAssociationCreateResourceID(directConnectGatewayID, aws.ToString(output.DirectConnectGatewayAssociation.AssociatedGateway.Id)))
	} else {
		associatedGatewayID := d.Get("associated_gateway_id").(string)
		input := directconnect.CreateDirectConnectGatewayAssociationInput{
			DirectConnectGatewayId: aws.String(directConnectGatewayID),
			GatewayId:              aws.String(associatedGatewayID),
		}

		if v, ok := d.GetOk("allowed_prefixes"); ok && v.(*schema.Set).Len() > 0 {
			input.AddAllowedPrefixesToDirectConnectGateway = expandRouteFilterPrefixes(v.(*schema.Set).List())
		}

		output, err := conn.CreateDirectConnectGatewayAssociation(ctx, &input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "creating Direct Connect Gateway Association (%s/%s): %s", directConnectGatewayID, associatedGatewayID, err)
		}

		// For historical reasons the resource ID isn't set to the association ID returned from the API.
		associationID = aws.ToString(output.DirectConnectGatewayAssociation.AssociationId)
		d.SetId(gatewayAssociationCreateResourceID(directConnectGatewayID, associatedGatewayID))
	}

	d.Set("dx_gateway_association_id", associationID)

	if _, err := waitGatewayAssociationCreated(ctx, conn, associationID, d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Direct Connect Gateway Association (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceGatewayAssociationRead(ctx, d, meta)...)
}

func resourceGatewayAssociationRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	c := meta.(*conns.AWSClient)
	conn := c.DirectConnectClient(ctx)

	associationID := d.Get("dx_gateway_association_id").(string)
	output, err := findGatewayAssociationByID(ctx, conn, associationID)

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] Direct Connect Gateway Association (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Direct Connect Gateway Association (%s): %s", d.Id(), err)
	}

	associatedGatewayID, dxGatewayID := aws.ToString(output.AssociatedGateway.Id), aws.ToString(output.DirectConnectGatewayId)
	if err := d.Set("allowed_prefixes", flattenRouteFilterPrefixes(output.AllowedPrefixesToDirectConnectGateway)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting allowed_prefixes: %s", err)
	}
	d.Set("associated_gateway_id", associatedGatewayID)
	d.Set("associated_gateway_owner_account_id", output.AssociatedGateway.OwnerAccount)
	d.Set("associated_gateway_type", output.AssociatedGateway.Type)
	d.Set("dx_gateway_association_id", output.AssociationId)
	d.Set("dx_gateway_id", dxGatewayID)
	d.Set("dx_gateway_owner_account_id", output.DirectConnectGatewayOwnerAccount)
	if output.AssociatedGateway.Type == awstypes.GatewayTypeTransitGateway {
		transitGatewayAttachment, err := tfec2.FindTransitGatewayAttachmentByTransitGatewayIDAndDirectConnectGatewayID(ctx, c.EC2Client(ctx), associatedGatewayID, dxGatewayID)

		switch {
		case tfawserr.ErrCodeEquals(err, "UnauthorizedOperation"):
			d.Set(names.AttrTransitGatewayAttachmentID, nil)
		case err != nil:
			return sdkdiag.AppendErrorf(diags, "reading EC2 Transit Gateway (%s) Attachment (%s): %s", associatedGatewayID, dxGatewayID, err)
		default:
			d.Set(names.AttrTransitGatewayAttachmentID, transitGatewayAttachment.TransitGatewayAttachmentId)
		}
	} else {
		d.Set(names.AttrTransitGatewayAttachmentID, nil)
	}

	return diags
}

func resourceGatewayAssociationUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DirectConnectClient(ctx)

	associationID := d.Get("dx_gateway_association_id").(string)
	input := directconnect.UpdateDirectConnectGatewayAssociationInput{
		AssociationId: aws.String(associationID),
	}

	o, n := d.GetChange("allowed_prefixes")
	os, ns := o.(*schema.Set), n.(*schema.Set)

	if add := ns.Difference(os); add.Len() > 0 {
		input.AddAllowedPrefixesToDirectConnectGateway = expandRouteFilterPrefixes(add.List())
	}

	if del := os.Difference(ns); del.Len() > 0 {
		input.RemoveAllowedPrefixesToDirectConnectGateway = expandRouteFilterPrefixes(del.List())
	}

	_, err := conn.UpdateDirectConnectGatewayAssociation(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Direct Connect Gateway Association (%s): %s", d.Id(), err)
	}

	if _, err := waitGatewayAssociationUpdated(ctx, conn, associationID, d.Timeout(schema.TimeoutUpdate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Direct Connect Gateway Association (%s) update: %s", d.Id(), err)
	}

	return append(diags, resourceGatewayAssociationRead(ctx, d, meta)...)
}

func resourceGatewayAssociationDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DirectConnectClient(ctx)

	associationID := d.Get("dx_gateway_association_id").(string)

	log.Printf("[DEBUG] Deleting Direct Connect Gateway Association: %s", d.Id())
	const (
		timeout = 1 * time.Minute
	)
	input := directconnect.DeleteDirectConnectGatewayAssociationInput{
		AssociationId: aws.String(associationID),
	}

	_, err := tfresource.RetryWhenIsAErrorMessageContains[any, *awstypes.DirectConnectClientException](ctx, timeout, func(ctx context.Context) (any, error) {
		return conn.DeleteDirectConnectGatewayAssociation(ctx, &input)
	}, "has non-deleted Private IP VPN")

	if errs.IsAErrorMessageContains[*awstypes.DirectConnectClientException](err, "does not exist") {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Direct Connect Gateway Association (%s): %s", d.Id(), err)
	}

	if _, err := waitGatewayAssociationDeleted(ctx, conn, associationID, d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Direct Connect Gateway Association (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func resourceGatewayAssociationImport(ctx context.Context, d *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
	conn := meta.(*conns.AWSClient).DirectConnectClient(ctx)

	parts := strings.Split(d.Id(), "/")

	if len(parts) != 2 {
		return nil, fmt.Errorf("Incorrect resource ID format: %q. Expected DXGATEWAYID/ASSOCIATEDGATEWAYID", d.Id())
	}

	directConnectGatewayID := parts[0]
	associatedGatewayID := parts[1]

	output, err := findGatewayAssociationByGatewayIDAndAssociatedGatewayID(ctx, conn, directConnectGatewayID, associatedGatewayID)

	if err != nil {
		return nil, err
	}

	d.SetId(gatewayAssociationCreateResourceID(directConnectGatewayID, associatedGatewayID))
	d.Set("dx_gateway_id", output.DirectConnectGatewayId)
	d.Set("dx_gateway_association_id", output.AssociationId)

	return []*schema.ResourceData{d}, nil
}

func gatewayAssociationCreateResourceID(directConnectGatewayID, associatedGatewayID string) string {
	return fmt.Sprintf("ga-%s%s", directConnectGatewayID, associatedGatewayID)
}

func findGatewayAssociationByID(ctx context.Context, conn *directconnect.Client, id string) (*awstypes.DirectConnectGatewayAssociation, error) {
	input := directconnect.DescribeDirectConnectGatewayAssociationsInput{
		AssociationId: aws.String(id),
	}

	return findNonDisassociatedGatewayAssociation(ctx, conn, &input, tfslices.PredicateTrue[*awstypes.DirectConnectGatewayAssociation]())
}

func findGatewayAssociationByGatewayIDAndAssociatedGatewayID(ctx context.Context, conn *directconnect.Client, directConnectGatewayID, associatedGatewayID string) (*awstypes.DirectConnectGatewayAssociation, error) {
	input := directconnect.DescribeDirectConnectGatewayAssociationsInput{
		AssociatedGatewayId:    aws.String(associatedGatewayID),
		DirectConnectGatewayId: aws.String(directConnectGatewayID),
	}

	return findNonDisassociatedGatewayAssociation(ctx, conn, &input, tfslices.PredicateTrue[*awstypes.DirectConnectGatewayAssociation]())
}

func findGatewayAssociationByGatewayIDAndVirtualGatewayID(ctx context.Context, conn *directconnect.Client, directConnectGatewayID, virtualGatewayID string) (*awstypes.DirectConnectGatewayAssociation, error) {
	input := directconnect.DescribeDirectConnectGatewayAssociationsInput{
		DirectConnectGatewayId: aws.String(directConnectGatewayID),
		VirtualGatewayId:       aws.String(virtualGatewayID),
	}

	return findNonDisassociatedGatewayAssociation(ctx, conn, &input, tfslices.PredicateTrue[*awstypes.DirectConnectGatewayAssociation]())
}

func findNonDisassociatedGatewayAssociation(ctx context.Context, conn *directconnect.Client, input *directconnect.DescribeDirectConnectGatewayAssociationsInput, filter tfslices.Predicate[*awstypes.DirectConnectGatewayAssociation]) (*awstypes.DirectConnectGatewayAssociation, error) {
	output, err := findGatewayAssociation(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	if state := output.AssociationState; state == awstypes.DirectConnectGatewayAssociationStateDisassociated {
		return nil, &retry.NotFoundError{
			Message: string(state),
		}
	}

	return output, nil
}

func findGatewayAssociation(ctx context.Context, conn *directconnect.Client, input *directconnect.DescribeDirectConnectGatewayAssociationsInput, filter tfslices.Predicate[*awstypes.DirectConnectGatewayAssociation]) (*awstypes.DirectConnectGatewayAssociation, error) {
	output, err := findGatewayAssociations(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findGatewayAssociations(ctx context.Context, conn *directconnect.Client, input *directconnect.DescribeDirectConnectGatewayAssociationsInput, filter tfslices.Predicate[*awstypes.DirectConnectGatewayAssociation]) ([]awstypes.DirectConnectGatewayAssociation, error) {
	var output []awstypes.DirectConnectGatewayAssociation

	err := describeDirectConnectGatewayAssociationsPages(ctx, conn, input, func(page *directconnect.DescribeDirectConnectGatewayAssociationsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.DirectConnectGatewayAssociations {
			if filter(&v) {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	return output, nil
}

func statusGatewayAssociation(conn *directconnect.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findGatewayAssociationByID(ctx, conn, id)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.AssociationState), nil
	}
}

func waitGatewayAssociationCreated(ctx context.Context, conn *directconnect.Client, id string, timeout time.Duration) (*awstypes.DirectConnectGatewayAssociation, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.DirectConnectGatewayAssociationStateAssociating),
		Target:  enum.Slice(awstypes.DirectConnectGatewayAssociationStateAssociated),
		Refresh: statusGatewayAssociation(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.DirectConnectGatewayAssociation); ok {
		retry.SetLastError(err, errors.New(aws.ToString(output.StateChangeError)))

		return output, err
	}

	return nil, err
}

func waitGatewayAssociationUpdated(ctx context.Context, conn *directconnect.Client, id string, timeout time.Duration) (*awstypes.DirectConnectGatewayAssociation, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.DirectConnectGatewayAssociationStateUpdating),
		Target:  enum.Slice(awstypes.DirectConnectGatewayAssociationStateAssociated),
		Refresh: statusGatewayAssociation(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.DirectConnectGatewayAssociation); ok {
		retry.SetLastError(err, errors.New(aws.ToString(output.StateChangeError)))

		return output, err
	}

	return nil, err
}

func waitGatewayAssociationDeleted(ctx context.Context, conn *directconnect.Client, id string, timeout time.Duration) (*awstypes.DirectConnectGatewayAssociation, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.DirectConnectGatewayAssociationStateDisassociating),
		Target:  []string{},
		Refresh: statusGatewayAssociation(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.DirectConnectGatewayAssociation); ok {
		retry.SetLastError(err, errors.New(aws.ToString(output.StateChangeError)))

		return output, err
	}

	return nil, err
}
