// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package networkmanager

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/networkmanager"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_networkmanager_transit_gateway_route_table_attachment", name="Transit Gateway Route Table Attachment")
// @Tags(identifierAttribute="arn")
func ResourceTransitGatewayRouteTableAttachment() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceTransitGatewayRouteTableAttachmentCreate,
		ReadWithoutTimeout:   resourceTransitGatewayRouteTableAttachmentRead,
		UpdateWithoutTimeout: resourceTransitGatewayRouteTableAttachmentUpdate,
		DeleteWithoutTimeout: resourceTransitGatewayRouteTableAttachmentDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		CustomizeDiff: verify.SetTagsDiff,

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

func resourceTransitGatewayRouteTableAttachmentCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).NetworkManagerConn(ctx)

	peeringID := d.Get("peering_id").(string)
	transitGatewayRouteTableARN := d.Get("transit_gateway_route_table_arn").(string)
	input := &networkmanager.CreateTransitGatewayRouteTableAttachmentInput{
		PeeringId:                   aws.String(peeringID),
		Tags:                        getTagsIn(ctx),
		TransitGatewayRouteTableArn: aws.String(transitGatewayRouteTableARN),
	}

	log.Printf("[DEBUG] Creating Network Manager Transit Gateway Route Table Attachment: %s", input)
	output, err := conn.CreateTransitGatewayRouteTableAttachmentWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Network Manager Transit Gateway (%s) Route Table (%s) Attachment: %s", peeringID, transitGatewayRouteTableARN, err)
	}

	d.SetId(aws.StringValue(output.TransitGatewayRouteTableAttachment.Attachment.AttachmentId))

	if _, err := waitTransitGatewayRouteTableAttachmentCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Network Manager Transit Gateway Route Table Attachment (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceTransitGatewayRouteTableAttachmentRead(ctx, d, meta)...)
}

func resourceTransitGatewayRouteTableAttachmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).NetworkManagerConn(ctx)

	transitGatewayRouteTableAttachment, err := FindTransitGatewayRouteTableAttachmentByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Network Manager Transit Gateway Route Table Attachment %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Network Manager Transit Gateway Route Table Attachment (%s): %s", d.Id(), err)
	}

	a := transitGatewayRouteTableAttachment.Attachment
	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "networkmanager",
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("attachment/%s", d.Id()),
	}.String()
	d.Set(names.AttrARN, arn)
	d.Set("attachment_policy_rule_number", a.AttachmentPolicyRuleNumber)
	d.Set("attachment_type", a.AttachmentType)
	d.Set("core_network_arn", a.CoreNetworkArn)
	d.Set("core_network_id", a.CoreNetworkId)
	d.Set("edge_location", a.EdgeLocation)
	d.Set(names.AttrOwnerAccountID, a.OwnerAccountId)
	d.Set("peering_id", transitGatewayRouteTableAttachment.PeeringId)
	d.Set(names.AttrResourceARN, a.ResourceArn)
	d.Set("segment_name", a.SegmentName)
	d.Set(names.AttrState, a.State)
	d.Set("transit_gateway_route_table_arn", transitGatewayRouteTableAttachment.TransitGatewayRouteTableArn)

	setTagsOut(ctx, a.Tags)

	return diags
}

func resourceTransitGatewayRouteTableAttachmentUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Tags only.
	return resourceTransitGatewayRouteTableAttachmentRead(ctx, d, meta)
}

func resourceTransitGatewayRouteTableAttachmentDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).NetworkManagerConn(ctx)

	log.Printf("[DEBUG] Deleting Network Manager Transit Gateway Route Table Attachment: %s", d.Id())
	_, err := conn.DeleteAttachmentWithContext(ctx, &networkmanager.DeleteAttachmentInput{
		AttachmentId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, networkmanager.ErrCodeResourceNotFoundException) {
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

func FindTransitGatewayRouteTableAttachmentByID(ctx context.Context, conn *networkmanager.NetworkManager, id string) (*networkmanager.TransitGatewayRouteTableAttachment, error) {
	input := &networkmanager.GetTransitGatewayRouteTableAttachmentInput{
		AttachmentId: aws.String(id),
	}

	output, err := conn.GetTransitGatewayRouteTableAttachmentWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, networkmanager.ErrCodeResourceNotFoundException) {
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

func statusTransitGatewayRouteTableAttachmentState(ctx context.Context, conn *networkmanager.NetworkManager, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindTransitGatewayRouteTableAttachmentByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Attachment.State), nil
	}
}

func waitTransitGatewayRouteTableAttachmentCreated(ctx context.Context, conn *networkmanager.NetworkManager, id string, timeout time.Duration) (*networkmanager.TransitGatewayRouteTableAttachment, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{networkmanager.AttachmentStateCreating, networkmanager.AttachmentStatePendingNetworkUpdate},
		Target:  []string{networkmanager.AttachmentStateAvailable, networkmanager.AttachmentStatePendingAttachmentAcceptance},
		Timeout: timeout,
		Refresh: statusTransitGatewayRouteTableAttachmentState(ctx, conn, id),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*networkmanager.TransitGatewayRouteTableAttachment); ok {
		return output, err
	}

	return nil, err
}

func waitTransitGatewayRouteTableAttachmentDeleted(ctx context.Context, conn *networkmanager.NetworkManager, id string, timeout time.Duration) (*networkmanager.TransitGatewayRouteTableAttachment, error) {
	stateConf := &retry.StateChangeConf{
		Pending:        []string{networkmanager.AttachmentStateDeleting},
		Target:         []string{},
		Timeout:        timeout,
		Refresh:        statusTransitGatewayRouteTableAttachmentState(ctx, conn, id),
		NotFoundChecks: 1,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*networkmanager.TransitGatewayRouteTableAttachment); ok {
		return output, err
	}

	return nil, err
}

func waitTransitGatewayRouteTableAttachmentAvailable(ctx context.Context, conn *networkmanager.NetworkManager, id string, timeout time.Duration) (*networkmanager.TransitGatewayRouteTableAttachment, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{networkmanager.AttachmentStateCreating, networkmanager.AttachmentStatePendingAttachmentAcceptance, networkmanager.AttachmentStatePendingNetworkUpdate},
		Target:  []string{networkmanager.AttachmentStateAvailable},
		Timeout: timeout,
		Refresh: statusTransitGatewayRouteTableAttachmentState(ctx, conn, id),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*networkmanager.TransitGatewayRouteTableAttachment); ok {
		return output, err
	}

	return nil, err
}
