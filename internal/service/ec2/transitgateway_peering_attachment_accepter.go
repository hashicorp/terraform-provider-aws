// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package ec2

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_ec2_transit_gateway_peering_attachment_accepter", name="Transit Gateway Peering Attachment Accepter")
// @Tags(identifierAttribute="id")
// @Testing(tagsTest=false)
func resourceTransitGatewayPeeringAttachmentAccepter() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceTransitGatewayPeeringAttachmentAccepterCreate,
		ReadWithoutTimeout:   resourceTransitGatewayPeeringAttachmentAccepterRead,
		UpdateWithoutTimeout: resourceTransitGatewayPeeringAttachmentAccepterUpdate,
		DeleteWithoutTimeout: resourceTransitGatewayPeeringAttachmentAccepterDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
				"multipath_resilience_status": {
					Type:     schema.TypeString,
					Computed: true,
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
							"multipath_resilience_mode": {
								Type:             schema.TypeString,
								Optional:         true,
								Computed:         true,
								ValidateDiagFunc: enum.Validate[awstypes.TransitGatewayPeeringAttachmentMultipathResilienceMode](),
							},
						},
					},
				},
				"peer_account_id": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"peer_region": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"peer_transit_gateway_id": {
					Type:     schema.TypeString,
					Computed: true,
				},
				names.AttrTags:    tftags.TagsSchema(),
				names.AttrTagsAll: tftags.TagsSchemaComputed(),
				names.AttrTransitGatewayAttachmentID: {
					Type:     schema.TypeString,
					Required: true,
					ForceNew: true,
				},
				names.AttrTransitGatewayID: {
					Type:     schema.TypeString,
					Computed: true,
				},
			}
		},
	}
}

func resourceTransitGatewayPeeringAttachmentAccepterCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	transitGatewayAttachmentID := d.Get(names.AttrTransitGatewayAttachmentID).(string)
	input := &ec2.AcceptTransitGatewayPeeringAttachmentInput{
		TransitGatewayAttachmentId: aws.String(transitGatewayAttachmentID),
	}

	log.Printf("[DEBUG] Accepting EC2 Transit Gateway Peering Attachment: %s", transitGatewayAttachmentID)
	output, err := conn.AcceptTransitGatewayPeeringAttachment(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "accepting EC2 Transit Gateway Peering Attachment (%s): %s", transitGatewayAttachmentID, err)
	}

	d.SetId(aws.ToString(output.TransitGatewayPeeringAttachment.TransitGatewayAttachmentId))

	transitGatewayPeeringAttachment, err := waitTransitGatewayPeeringAttachmentAccepted(ctx, conn, d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EC2 Transit Gateway Peering Attachment (%s) update: %s", d.Id(), err)
	}

	// Accepting an attachment whose requester allows multipath resilience also allows it for the accepter.
	// The accepter's side can only be modified once the attachment is available.
	if transitGatewayPeeringAttachmentAccepterOptionsChanged(d, transitGatewayPeeringAttachment.Options) {
		if err := modifyTransitGatewayPeeringAttachmentOptions(ctx, conn, d); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	} else if err := waitTransitGatewayPeeringAttachmentMultipathResilienceEnabledIfAllowed(ctx, conn, d.Id(), transitGatewayPeeringAttachment.Options); err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	if err := createTags(ctx, conn, d.Id(), getTagsIn(ctx)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting EC2 Transit Gateway Peering Attachment (%s) tags: %s", d.Id(), err)
	}

	return append(diags, resourceTransitGatewayPeeringAttachmentAccepterRead(ctx, d, meta)...)
}

func resourceTransitGatewayPeeringAttachmentAccepterRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	transitGatewayPeeringAttachment, err := findTransitGatewayPeeringAttachmentByID(ctx, conn, d.Id())

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] EC2 Transit Gateway Peering Attachment (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 Transit Gateway Peering Attachment (%s): %s", d.Id(), err)
	}

	transitGatewayID := aws.ToString(transitGatewayPeeringAttachment.AccepterTgwInfo.TransitGatewayId)
	_, err = findTransitGatewayByID(ctx, conn, transitGatewayID)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 Transit Gateway (%s): %s", transitGatewayID, err)
	}

	if v := transitGatewayPeeringAttachment.Options; v != nil {
		d.Set("multipath_resilience_status", v.MultipathResilienceStatus)
	} else {
		d.Set("multipath_resilience_status", nil)
	}
	if err := d.Set("options", flattenTransitGatewayPeeringAttachmentAccepterOptions(transitGatewayPeeringAttachment.Options)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting options: %s", err)
	}
	d.Set("peer_account_id", transitGatewayPeeringAttachment.RequesterTgwInfo.OwnerId)
	d.Set("peer_region", transitGatewayPeeringAttachment.RequesterTgwInfo.Region)
	d.Set("peer_transit_gateway_id", transitGatewayPeeringAttachment.RequesterTgwInfo.TransitGatewayId)
	d.Set(names.AttrTransitGatewayAttachmentID, transitGatewayPeeringAttachment.TransitGatewayAttachmentId)
	d.Set(names.AttrTransitGatewayID, transitGatewayPeeringAttachment.AccepterTgwInfo.TransitGatewayId)

	setTagsOut(ctx, transitGatewayPeeringAttachment.Tags)

	return diags
}

func resourceTransitGatewayPeeringAttachmentAccepterUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	if d.HasChanges("options.0.allow_multipath_resilience", "options.0.multipath_resilience_mode") {
		if err := modifyTransitGatewayPeeringAttachmentOptions(ctx, conn, d); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	return append(diags, resourceTransitGatewayPeeringAttachmentAccepterRead(ctx, d, meta)...)
}

func resourceTransitGatewayPeeringAttachmentAccepterDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
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

// transitGatewayPeeringAttachmentAccepterOptionsChanged reports whether the configured multipath resilience
// settings differ from the accepter's current settings.
func transitGatewayPeeringAttachmentAccepterOptionsChanged(d *schema.ResourceData, apiObject *awstypes.TransitGatewayPeeringAttachmentOptions) bool {
	options := expandModifyTransitGatewayPeeringAttachmentRequestOptions(d)
	if options == nil {
		return false
	}

	if apiObject == nil {
		return true
	}

	if aws.ToBool(options.AllowMultipathResilience) != aws.ToBool(apiObject.AccepterAllowMultipathResilience) {
		return true
	}

	return options.MultipathResilienceMode != "" && options.MultipathResilienceMode != apiObject.AccepterMultipathResilienceMode
}

// flattenTransitGatewayPeeringAttachmentAccepterOptions flattens the options of the accepter's side of the attachment.
func flattenTransitGatewayPeeringAttachmentAccepterOptions(apiObject *awstypes.TransitGatewayPeeringAttachmentOptions) []any {
	if apiObject == nil || (apiObject.AccepterAllowMultipathResilience == nil && apiObject.AccepterMultipathResilienceMode == "") {
		return nil
	}

	tfMap := map[string]any{}

	if v := apiObject.AccepterAllowMultipathResilience; v != nil {
		tfMap["allow_multipath_resilience"] = aws.ToBool(v)
	}

	if v := apiObject.AccepterMultipathResilienceMode; v != "" {
		tfMap["multipath_resilience_mode"] = v
	}

	return []any{tfMap}
}
