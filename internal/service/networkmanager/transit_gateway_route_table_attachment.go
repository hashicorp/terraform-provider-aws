// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package networkmanager

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/networkmanager"
	awstypes "github.com/aws/aws-sdk-go-v2/service/networkmanager/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_networkmanager_transit_gateway_route_table_attachment", name="Transit Gateway Route Table Attachment")
// @Tags(identifierAttribute="arn")
func resourceTransitGatewayRouteTableAttachment() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceTransitGatewayRouteTableAttachmentCreate,
		ReadWithoutTimeout:   resourceTransitGatewayRouteTableAttachmentRead,
		UpdateWithoutTimeout: resourceTransitGatewayRouteTableAttachmentUpdate,
		DeleteWithoutTimeout: resourceTransitGatewayRouteTableAttachmentDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"attachment_policy_rule_number": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"attachment_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"core_network_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"core_network_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"edge_location": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrOwnerAccountID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"peering_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrResourceARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"segment_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrState: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"transit_gateway_route_table_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
		},
	}
}

func resourceTransitGatewayRouteTableAttachmentCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).NetworkManagerClient(ctx)

	peeringID := d.Get("peering_id").(string)
	transitGatewayRouteTableARN := d.Get("transit_gateway_route_table_arn").(string)
	input := &networkmanager.CreateTransitGatewayRouteTableAttachmentInput{
		PeeringId:                   aws.String(peeringID),
		Tags:                        getTagsIn(ctx),
		TransitGatewayRouteTableArn: aws.String(transitGatewayRouteTableARN),
	}

	output, err := conn.CreateTransitGatewayRouteTableAttachment(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Network Manager Transit Gateway (%s) Route Table (%s) Attachment: %s", peeringID, transitGatewayRouteTableARN, err)
	}

	d.SetId(aws.ToString(output.TransitGatewayRouteTableAttachment.Attachment.AttachmentId))

	if _, err := waitTransitGatewayRouteTableAttachmentCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Network Manager Transit Gateway Route Table Attachment (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceTransitGatewayRouteTableAttachmentRead(ctx, d, meta)...)
}

func resourceTransitGatewayRouteTableAttachmentRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).NetworkManagerClient(ctx)

	transitGatewayRouteTableAttachment, err := findTransitGatewayRouteTableAttachmentByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Network Manager Transit Gateway Route Table Attachment %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Network Manager Transit Gateway Route Table Attachment (%s): %s", d.Id(), err)
	}

	attachment := transitGatewayRouteTableAttachment.Attachment
	d.Set(names.AttrARN, attachmentARN(ctx, meta.(*conns.AWSClient), d.Id()))
	d.Set("attachment_policy_rule_number", attachment.AttachmentPolicyRuleNumber)
	d.Set("attachment_type", attachment.AttachmentType)
	d.Set("core_network_arn", attachment.CoreNetworkArn)
	d.Set("core_network_id", attachment.CoreNetworkId)
	d.Set("edge_location", attachment.EdgeLocation)
	d.Set(names.AttrOwnerAccountID, attachment.OwnerAccountId)
	d.Set("peering_id", transitGatewayRouteTableAttachment.PeeringId)
	d.Set(names.AttrResourceARN, attachment.ResourceArn)
	d.Set("segment_name", attachment.SegmentName)
	d.Set(names.AttrState, attachment.State)
	d.Set("transit_gateway_route_table_arn", transitGatewayRouteTableAttachment.TransitGatewayRouteTableArn)

	setTagsOut(ctx, attachment.Tags)

	return diags
}

func resourceTransitGatewayRouteTableAttachmentUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	// Tags only.
	return resourceTransitGatewayRouteTableAttachmentRead(ctx, d, meta)
}

func resourceTransitGatewayRouteTableAttachmentDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).NetworkManagerClient(ctx)

	log.Printf("[DEBUG] Deleting Network Manager Transit Gateway Route Table Attachment: %s", d.Id())
	_, err := conn.DeleteAttachment(ctx, &networkmanager.DeleteAttachmentInput{
		AttachmentId: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Network Manager Transit Gateway Route Table Attachment (%s): %s", d.Id(), err)
	}

	if _, err := waitTransitGatewayRouteTableAttachmentDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Network Manager Transit Gateway Route Table Attachment (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func findTransitGatewayRouteTableAttachmentByID(ctx context.Context, conn *networkmanager.Client, id string) (*awstypes.TransitGatewayRouteTableAttachment, error) {
	input := &networkmanager.GetTransitGatewayRouteTableAttachmentInput{
		AttachmentId: aws.String(id),
	}

	output, err := conn.GetTransitGatewayRouteTableAttachment(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.TransitGatewayRouteTableAttachment == nil || output.TransitGatewayRouteTableAttachment.Attachment == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.TransitGatewayRouteTableAttachment, nil
}

func statusTransitGatewayRouteTableAttachment(ctx context.Context, conn *networkmanager.Client, id string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findTransitGatewayRouteTableAttachmentByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Attachment.State), nil
	}
}

func waitTransitGatewayRouteTableAttachmentCreated(ctx context.Context, conn *networkmanager.Client, id string, timeout time.Duration) (*awstypes.TransitGatewayRouteTableAttachment, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.AttachmentStateCreating, awstypes.AttachmentStatePendingNetworkUpdate),
		Target:                    enum.Slice(awstypes.AttachmentStateAvailable, awstypes.AttachmentStatePendingAttachmentAcceptance),
		Timeout:                   timeout,
		Refresh:                   statusTransitGatewayRouteTableAttachment(ctx, conn, id),
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.TransitGatewayRouteTableAttachment); ok {
		tfresource.SetLastError(err, attachmentsError(output.Attachment.LastModificationErrors))

		return output, err
	}

	return nil, err
}

func waitTransitGatewayRouteTableAttachmentDeleted(ctx context.Context, conn *networkmanager.Client, id string, timeout time.Duration) (*awstypes.TransitGatewayRouteTableAttachment, error) {
	stateConf := &retry.StateChangeConf{
		Pending:        enum.Slice(awstypes.AttachmentStateDeleting),
		Target:         []string{},
		Timeout:        timeout,
		Refresh:        statusTransitGatewayRouteTableAttachment(ctx, conn, id),
		NotFoundChecks: 1,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.TransitGatewayRouteTableAttachment); ok {
		tfresource.SetLastError(err, attachmentsError(output.Attachment.LastModificationErrors))

		return output, err
	}

	return nil, err
}

func waitTransitGatewayRouteTableAttachmentAvailable(ctx context.Context, conn *networkmanager.Client, id string, timeout time.Duration) (*awstypes.TransitGatewayRouteTableAttachment, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.AttachmentStateCreating, awstypes.AttachmentStatePendingAttachmentAcceptance, awstypes.AttachmentStatePendingNetworkUpdate),
		Target:  enum.Slice(awstypes.AttachmentStateAvailable),
		Timeout: timeout,
		Refresh: statusTransitGatewayRouteTableAttachment(ctx, conn, id),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.TransitGatewayRouteTableAttachment); ok {
		tfresource.SetLastError(err, attachmentsError(output.Attachment.LastModificationErrors))

		return output, err
	}

	return nil, err
}
