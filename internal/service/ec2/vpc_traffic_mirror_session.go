// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_ec2_traffic_mirror_session", name="Traffic Mirror Session")
// @Tags(identifierAttribute="id")
// @Testing(tagsTest=false)
func resourceTrafficMirrorSession() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceTrafficMirrorSessionCreate,
		UpdateWithoutTimeout: resourceTrafficMirrorSessionUpdate,
		ReadWithoutTimeout:   resourceTrafficMirrorSessionRead,
		DeleteWithoutTimeout: resourceTrafficMirrorSessionDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		CustomizeDiff: verify.SetTagsDiff,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Optional: true,
			},
			names.AttrNetworkInterfaceID: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrOwnerID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"packet_length": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"session_number": {
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntBetween(1, 32766),
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"traffic_mirror_filter_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"traffic_mirror_target_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"virtual_network_id": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntBetween(1, 16777216),
			},
		},
	}
}

func resourceTrafficMirrorSessionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	input := &ec2.CreateTrafficMirrorSessionInput{
		ClientToken:           aws.String(id.UniqueId()),
		NetworkInterfaceId:    aws.String(d.Get(names.AttrNetworkInterfaceID).(string)),
		TagSpecifications:     getTagSpecificationsIn(ctx, awstypes.ResourceTypeTrafficMirrorSession),
		TrafficMirrorFilterId: aws.String(d.Get("traffic_mirror_filter_id").(string)),
		TrafficMirrorTargetId: aws.String(d.Get("traffic_mirror_target_id").(string)),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("packet_length"); ok {
		input.PacketLength = aws.Int32(int32(v.(int)))
	}

	if v, ok := d.GetOk("session_number"); ok {
		input.SessionNumber = aws.Int32(int32(v.(int)))
	}

	if v, ok := d.GetOk("virtual_network_id"); ok {
		input.VirtualNetworkId = aws.Int32(int32(v.(int)))
	}

	output, err := conn.CreateTrafficMirrorSession(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EC2 Traffic Mirror Session: %s", err)
	}

	d.SetId(aws.ToString(output.TrafficMirrorSession.TrafficMirrorSessionId))

	return append(diags, resourceTrafficMirrorSessionRead(ctx, d, meta)...)
}

func resourceTrafficMirrorSessionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	session, err := findTrafficMirrorSessionByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 Traffic Mirror Session %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 Traffic Mirror Session (%s): %s", d.Id(), err)
	}

	ownerID := aws.ToString(session.OwnerId)
	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "ec2",
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: ownerID,
		Resource:  "traffic-mirror-session/" + d.Id(),
	}.String()
	d.Set(names.AttrARN, arn)
	d.Set(names.AttrDescription, session.Description)
	d.Set(names.AttrNetworkInterfaceID, session.NetworkInterfaceId)
	d.Set(names.AttrOwnerID, ownerID)
	d.Set("packet_length", session.PacketLength)
	d.Set("session_number", session.SessionNumber)
	d.Set("traffic_mirror_filter_id", session.TrafficMirrorFilterId)
	d.Set("traffic_mirror_target_id", session.TrafficMirrorTargetId)
	d.Set("virtual_network_id", session.VirtualNetworkId)

	setTagsOut(ctx, session.Tags)

	return diags
}

func resourceTrafficMirrorSessionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := &ec2.ModifyTrafficMirrorSessionInput{
			TrafficMirrorSessionId: aws.String(d.Id()),
		}

		if d.HasChange("session_number") {
			input.SessionNumber = aws.Int32(int32(d.Get("session_number").(int)))
		}

		if d.HasChange("traffic_mirror_filter_id") {
			input.TrafficMirrorFilterId = aws.String(d.Get("traffic_mirror_filter_id").(string))
		}

		if d.HasChange("traffic_mirror_target_id") {
			input.TrafficMirrorTargetId = aws.String(d.Get("traffic_mirror_target_id").(string))
		}

		var removeFields []awstypes.TrafficMirrorSessionField

		if d.HasChange(names.AttrDescription) {
			if v := d.Get(names.AttrDescription).(string); v != "" {
				input.Description = aws.String(v)
			} else {
				removeFields = append(removeFields, awstypes.TrafficMirrorSessionFieldDescription)
			}
		}

		if d.HasChange("packet_length") {
			if v := d.Get("packet_length").(int); v != 0 {
				input.PacketLength = aws.Int32(int32(v))
			} else {
				removeFields = append(removeFields, awstypes.TrafficMirrorSessionFieldPacketLength)
			}
		}

		if d.HasChange("virtual_network_id") {
			if v := d.Get("virtual_network_id").(int); v != 0 {
				input.VirtualNetworkId = aws.Int32(int32(v))
			} else {
				removeFields = append(removeFields, awstypes.TrafficMirrorSessionFieldVirtualNetworkId)
			}
		}

		if len(removeFields) > 0 {
			input.RemoveFields = removeFields
		}

		_, err := conn.ModifyTrafficMirrorSession(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating EC2 Traffic Mirror Session (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceTrafficMirrorSessionRead(ctx, d, meta)...)
}

func resourceTrafficMirrorSessionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	log.Printf("[DEBUG] Deleting EC2 Traffic Mirror Session: %s", d.Id())
	_, err := conn.DeleteTrafficMirrorSession(ctx, &ec2.DeleteTrafficMirrorSessionInput{
		TrafficMirrorSessionId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidTrafficMirrorSessionIdNotFound) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting EC2 Traffic Mirror Session (%s): %s", d.Id(), err)
	}

	return diags
}
