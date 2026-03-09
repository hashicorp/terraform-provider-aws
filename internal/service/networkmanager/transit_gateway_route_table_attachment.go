// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package networkmanager

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/networkmanager"
	awstypes "github.com/aws/aws-sdk-go-v2/service/networkmanager/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_networkmanager_transit_gateway_route_table_attachment", name="Transit Gateway Route Table Attachment")
// @Tags(identifierAttribute="arn")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/networkmanager/types;awstypes;awstypes.TransitGatewayRouteTableAttachment")
// @Testing(skipEmptyTags=true)
// @Testing(generator=false)
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
			"routing_policy_label": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(0, 256),
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
	input := networkmanager.CreateTransitGatewayRouteTableAttachmentInput{
		PeeringId:                   aws.String(peeringID),
		Tags:                        getTagsIn(ctx),
		TransitGatewayRouteTableArn: aws.String(transitGatewayRouteTableARN),
	}

	if v, ok := d.GetOk("routing_policy_label"); ok {
		input.RoutingPolicyLabel = aws.String(v.(string))
	}

	output, err := conn.CreateTransitGatewayRouteTableAttachment(ctx, &input)

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

	if !d.IsNewResource() && retry.NotFound(err) {
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
	coreNetworkID := aws.ToString(attachment.CoreNetworkId)
	d.Set("core_network_id", coreNetworkID)
	d.Set("edge_location", attachment.EdgeLocation)
	d.Set(names.AttrOwnerAccountID, attachment.OwnerAccountId)
	d.Set("peering_id", transitGatewayRouteTableAttachment.PeeringId)
	d.Set(names.AttrResourceARN, attachment.ResourceArn)
	if routingPolicyLabel, err := findAttachmentRoutingPolicyAssociationLabelByTwoPartKey(ctx, conn, coreNetworkID, d.Id()); err != nil && !retry.NotFound(err) {
		return sdkdiag.AppendErrorf(diags, "reading Network Manager Transit Gateway Route Table Attachment (%s) routing policy label: %s", d.Id(), err)
	} else {
		d.Set("routing_policy_label", routingPolicyLabel)
	}
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
	input := networkmanager.DeleteAttachmentInput{
		AttachmentId: aws.String(d.Id()),
	}
	_, err := conn.DeleteAttachment(ctx, &input)

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
	input := networkmanager.GetTransitGatewayRouteTableAttachmentInput{
		AttachmentId: aws.String(id),
	}

	return findTransitGatewayRouteTableAttachment(ctx, conn, &input)
}

func findTransitGatewayRouteTableAttachment(ctx context.Context, conn *networkmanager.Client, input *networkmanager.GetTransitGatewayRouteTableAttachmentInput) (*awstypes.TransitGatewayRouteTableAttachment, error) {
	output, err := conn.GetTransitGatewayRouteTableAttachment(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.TransitGatewayRouteTableAttachment == nil || output.TransitGatewayRouteTableAttachment.Attachment == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output.TransitGatewayRouteTableAttachment, nil
}

func statusTransitGatewayRouteTableAttachment(conn *networkmanager.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findTransitGatewayRouteTableAttachmentByID(ctx, conn, id)

		if retry.NotFound(err) {
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
		Refresh:                   statusTransitGatewayRouteTableAttachment(conn, id),
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.TransitGatewayRouteTableAttachment); ok {
		retry.SetLastError(err, attachmentsError(output.Attachment.LastModificationErrors))

		return output, err
	}

	return nil, err
}

func waitTransitGatewayRouteTableAttachmentDeleted(ctx context.Context, conn *networkmanager.Client, id string, timeout time.Duration) (*awstypes.TransitGatewayRouteTableAttachment, error) {
	stateConf := &retry.StateChangeConf{
		Pending:        enum.Slice(awstypes.AttachmentStateDeleting, awstypes.AttachmentStatePendingNetworkUpdate),
		Target:         []string{},
		Timeout:        timeout,
		Refresh:        statusTransitGatewayRouteTableAttachment(conn, id),
		NotFoundChecks: 1,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.TransitGatewayRouteTableAttachment); ok {
		retry.SetLastError(err, attachmentsError(output.Attachment.LastModificationErrors))

		return output, err
	}

	return nil, err
}

func waitTransitGatewayRouteTableAttachmentAvailable(ctx context.Context, conn *networkmanager.Client, id string, timeout time.Duration) (*awstypes.TransitGatewayRouteTableAttachment, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.AttachmentStateCreating, awstypes.AttachmentStatePendingAttachmentAcceptance, awstypes.AttachmentStatePendingNetworkUpdate),
		Target:  enum.Slice(awstypes.AttachmentStateAvailable),
		Timeout: timeout,
		Refresh: statusTransitGatewayRouteTableAttachment(conn, id),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.TransitGatewayRouteTableAttachment); ok {
		retry.SetLastError(err, attachmentsError(output.Attachment.LastModificationErrors))

		return output, err
	}

	return nil, err
}
