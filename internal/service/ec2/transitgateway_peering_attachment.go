// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package ec2

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_ec2_transit_gateway_peering_attachment", name="Transit Gateway Peering Attachment")
// @Tags(identifierAttribute="id")
// @Testing(tagsTest=false)
func resourceTransitGatewayPeeringAttachment() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceTransitGatewayPeeringAttachmentCreate,
		ReadWithoutTimeout:   resourceTransitGatewayPeeringAttachmentRead,
		UpdateWithoutTimeout: resourceTransitGatewayPeeringAttachmentUpdate,
		DeleteWithoutTimeout: resourceTransitGatewayPeeringAttachmentDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
				names.AttrARN: {
					Type:     schema.TypeString,
					Computed: true,
				},
				"multipath_resilience_status": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"peer_account_id": {
					Type:         schema.TypeString,
					Optional:     true,
					ForceNew:     true,
					Computed:     true,
					ValidateFunc: verify.ValidAccountID,
				},
				"peer_region": {
					Type:     schema.TypeString,
					Required: true,
					ForceNew: true,
				},
				"peer_transit_gateway_id": {
					Type:     schema.TypeString,
					Required: true,
					ForceNew: true,
				},
				names.AttrState: {
					Type:     schema.TypeString,
					Computed: true,
				},
				names.AttrTags:    tftags.TagsSchema(),
				names.AttrTagsAll: tftags.TagsSchemaOptionalComputedWarnShouldBeComputedOnly(),
				names.AttrTransitGatewayID: {
					Type:     schema.TypeString,
					Required: true,
					ForceNew: true,
				},
				"options": {
					Type:     schema.TypeList,
					Optional: true,
					Computed: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"allow_multipath_resilience": {
								Type:     schema.TypeBool,
								Optional: true,
								Computed: true,
							},
							"dynamic_routing": {
								Type:             schema.TypeString,
								Optional:         true,
								Computed:         true,
								ForceNew:         true,
								ValidateDiagFunc: enum.Validate[awstypes.DynamicRoutingValue](),
							},
							"multipath_resilience_mode": {
								Type:             schema.TypeString,
								Optional:         true,
								Computed:         true,
								ValidateDiagFunc: enum.Validate[awstypes.TransitGatewayPeeringAttachmentMultipathResilienceMode](),
							},
						},
					},
				},
			}
		},
	}
}

func resourceTransitGatewayPeeringAttachmentCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	peerAccountID := meta.(*conns.AWSClient).AccountID(ctx)
	if v, ok := d.GetOk("peer_account_id"); ok {
		peerAccountID = v.(string)
	}
	input := &ec2.CreateTransitGatewayPeeringAttachmentInput{
		PeerAccountId:        aws.String(peerAccountID),
		PeerRegion:           aws.String(d.Get("peer_region").(string)),
		PeerTransitGatewayId: aws.String(d.Get("peer_transit_gateway_id").(string)),
		TagSpecifications:    getTagSpecificationsIn(ctx, awstypes.ResourceTypeTransitGatewayAttachment),
		TransitGatewayId:     aws.String(d.Get(names.AttrTransitGatewayID).(string)),
	}

	if v, ok := d.GetOk("options"); ok {
		input.Options = expandCreateTransitGatewayPeeringAttachmentRequestOptions(d, v.([]any))
	}

	output, err := conn.CreateTransitGatewayPeeringAttachment(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EC2 Transit Gateway Peering Attachment: %s", err)
	}

	d.SetId(aws.ToString(output.TransitGatewayPeeringAttachment.TransitGatewayAttachmentId))

	if _, err := waitTransitGatewayPeeringAttachmentCreated(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EC2 Transit Gateway Peering Attachment (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceTransitGatewayPeeringAttachmentRead(ctx, d, meta)...)
}

func resourceTransitGatewayPeeringAttachmentRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	c := meta.(*conns.AWSClient)
	conn := c.EC2Client(ctx)

	transitGatewayPeeringAttachment, err := findTransitGatewayPeeringAttachmentByID(ctx, conn, d.Id())

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] EC2 Transit Gateway Peering Attachment (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 Transit Gateway Peering Attachment (%s): %s", d.Id(), err)
	}

	resourceOwnerID := aws.ToString(transitGatewayPeeringAttachment.RequesterTgwInfo.OwnerId)
	d.Set(names.AttrARN, transitGatewayAttachmentARN(ctx, c, resourceOwnerID, d.Id()))
	if v := transitGatewayPeeringAttachment.Options; v != nil {
		d.Set("multipath_resilience_status", v.MultipathResilienceStatus)
	} else {
		d.Set("multipath_resilience_status", nil)
	}
	d.Set("peer_account_id", transitGatewayPeeringAttachment.AccepterTgwInfo.OwnerId)
	d.Set("peer_region", transitGatewayPeeringAttachment.AccepterTgwInfo.Region)
	d.Set("peer_transit_gateway_id", transitGatewayPeeringAttachment.AccepterTgwInfo.TransitGatewayId)
	d.Set(names.AttrState, transitGatewayPeeringAttachment.State)
	d.Set(names.AttrTransitGatewayID, transitGatewayPeeringAttachment.RequesterTgwInfo.TransitGatewayId)

	if err := d.Set("options", flattenTransitGatewayPeeringAttachmentOptions(transitGatewayPeeringAttachment.Options)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting options: %s", err)
	}

	setTagsOut(ctx, transitGatewayPeeringAttachment.Tags)

	return diags
}

func resourceTransitGatewayPeeringAttachmentUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	if d.HasChanges("options.0.allow_multipath_resilience", "options.0.multipath_resilience_mode") {
		if err := modifyTransitGatewayPeeringAttachmentOptions(ctx, conn, d); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	return append(diags, resourceTransitGatewayPeeringAttachmentRead(ctx, d, meta)...)
}

func resourceTransitGatewayPeeringAttachmentDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	log.Printf("[DEBUG] Deleting EC2 Transit Gateway Peering Attachment: %s", d.Id())
	input := ec2.DeleteTransitGatewayPeeringAttachmentInput{
		TransitGatewayAttachmentId: aws.String(d.Id()),
	}
	_, err := conn.DeleteTransitGatewayPeeringAttachment(ctx, &input)

	if tfawserr.ErrCodeEquals(err, errCodeInvalidTransitGatewayAttachmentIDNotFound) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting EC2 Transit Gateway Peering Attachment (%s): %s", d.Id(), err)
	}

	if _, err := waitTransitGatewayPeeringAttachmentDeleted(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EC2 Transit Gateway Peering Attachment (%s) delete: %s", d.Id(), err)
	}

	return diags
}

// modifyTransitGatewayPeeringAttachmentOptions modifies the multipath resilience settings of the
// side of the attachment in the Region (and account) that the resource is managed in.
// Each side can only be modified once the attachment is available (IncorrectState otherwise).
func modifyTransitGatewayPeeringAttachmentOptions(ctx context.Context, conn *ec2.Client, d *schema.ResourceData) error {
	options := expandModifyTransitGatewayPeeringAttachmentRequestOptions(d)
	if options == nil {
		return nil
	}

	input := ec2.ModifyTransitGatewayPeeringAttachmentInput{
		Options:                    options,
		TransitGatewayAttachmentId: aws.String(d.Id()),
	}
	output, err := conn.ModifyTransitGatewayPeeringAttachment(ctx, &input)

	if err != nil {
		return fmt.Errorf("modifying EC2 Transit Gateway Peering Attachment (%s): %w", d.Id(), err)
	}

	return waitTransitGatewayPeeringAttachmentMultipathResilienceEnabledIfAllowed(ctx, conn, d.Id(), output.TransitGatewayPeeringAttachment.Options)
}

// waitTransitGatewayPeeringAttachmentMultipathResilienceEnabledIfAllowed waits for multipath resilience to be enabled when both sides allow it.
// Disallowing takes effect immediately; enabling takes effect asynchronously once both sides allow it.
func waitTransitGatewayPeeringAttachmentMultipathResilienceEnabledIfAllowed(ctx context.Context, conn *ec2.Client, id string, options *awstypes.TransitGatewayPeeringAttachmentOptions) error {
	if options == nil || !aws.ToBool(options.AccepterAllowMultipathResilience) || !aws.ToBool(options.RequesterAllowMultipathResilience) {
		return nil
	}

	if _, err := waitTransitGatewayPeeringAttachmentMultipathResilienceEnabled(ctx, conn, id); err != nil {
		return fmt.Errorf("waiting for EC2 Transit Gateway Peering Attachment (%s) multipath resilience enable: %w", id, err)
	}

	return nil
}

// transitGatewayPeeringAttachmentOptionConfigured reports whether attribute attr is set in the configuration of the options block.
func transitGatewayPeeringAttachmentOptionConfigured(d *schema.ResourceData, attr string) bool {
	v := d.GetRawConfig().GetAttr("options")
	if !v.IsKnown() || v.IsNull() || v.LengthInt() == 0 {
		return false
	}

	v = v.Index(cty.NumberIntVal(0))
	if !v.IsKnown() || v.IsNull() {
		return false
	}

	v = v.GetAttr(attr)

	return v.IsKnown() && !v.IsNull()
}

func expandCreateTransitGatewayPeeringAttachmentRequestOptions(d *schema.ResourceData, tfList []any) *awstypes.CreateTransitGatewayPeeringAttachmentRequestOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	apiObject := &awstypes.CreateTransitGatewayPeeringAttachmentRequestOptions{}

	tfMap := tfList[0].(map[string]any)

	if v, ok := tfMap["dynamic_routing"].(string); ok {
		apiObject.DynamicRouting = awstypes.DynamicRoutingValue(v)
	}

	if v, ok := tfMap["allow_multipath_resilience"].(bool); ok && v && transitGatewayPeeringAttachmentOptionConfigured(d, "allow_multipath_resilience") {
		apiObject.AllowMultipathResilience = aws.Bool(v)
	}

	// Sent whenever configured so that an invalid combination (a mode without multipath resilience allowed) is rejected by the API.
	if v, ok := tfMap["multipath_resilience_mode"].(string); ok && v != "" && transitGatewayPeeringAttachmentOptionConfigured(d, "multipath_resilience_mode") {
		apiObject.MultipathResilienceMode = awstypes.TransitGatewayPeeringAttachmentMultipathResilienceMode(v)
	}

	return apiObject
}

// expandModifyTransitGatewayPeeringAttachmentRequestOptions returns nil when multipath resilience isn't configured:
// the arguments are Optional+Computed, so the current setting is left unchanged.
// A configured mode is always sent so that an invalid combination is rejected by the API rather than silently ignored.
func expandModifyTransitGatewayPeeringAttachmentRequestOptions(d *schema.ResourceData) *awstypes.ModifyTransitGatewayPeeringAttachmentRequestOptions {
	allowConfigured := transitGatewayPeeringAttachmentOptionConfigured(d, "allow_multipath_resilience")
	modeConfigured := transitGatewayPeeringAttachmentOptionConfigured(d, "multipath_resilience_mode")

	if !allowConfigured && !modeConfigured {
		return nil
	}

	apiObject := &awstypes.ModifyTransitGatewayPeeringAttachmentRequestOptions{}

	if allowConfigured {
		apiObject.AllowMultipathResilience = aws.Bool(d.Get("options.0.allow_multipath_resilience").(bool))
	}

	if v := d.Get("options.0.multipath_resilience_mode").(string); modeConfigured && v != "" {
		apiObject.MultipathResilienceMode = awstypes.TransitGatewayPeeringAttachmentMultipathResilienceMode(v)
	}

	return apiObject
}

// flattenTransitGatewayPeeringAttachmentOptions flattens the options of the requester's side of the attachment.
func flattenTransitGatewayPeeringAttachmentOptions(apiObject *awstypes.TransitGatewayPeeringAttachmentOptions) []any {
	// Treat a nil or empty Options object as "no options block".
	// The API may return an empty Options struct (for example when a field is present in the wire response but not yet modeled in the SDK).
	// Without this guard the provider would flatten that into a block containing an empty "dynamic_routing", diverging from a configuration with no options block.
	if apiObject == nil || (apiObject.DynamicRouting == "" && apiObject.RequesterAllowMultipathResilience == nil && apiObject.RequesterMultipathResilienceMode == "") {
		return nil
	}

	tfMap := map[string]any{
		"dynamic_routing": apiObject.DynamicRouting,
	}

	if v := apiObject.RequesterAllowMultipathResilience; v != nil {
		tfMap["allow_multipath_resilience"] = aws.ToBool(v)
	}

	if v := apiObject.RequesterMultipathResilienceMode; v != "" {
		tfMap["multipath_resilience_mode"] = v
	}

	return []any{tfMap}
}
