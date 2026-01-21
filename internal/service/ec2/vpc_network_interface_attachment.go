// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package ec2

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_network_interface_attachment", name="Network Interface Attachment")
func resourceNetworkInterfaceAttachment() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceNetworkInterfaceAttachmentCreate,
		ReadWithoutTimeout:   resourceNetworkInterfaceAttachmentRead,
		DeleteWithoutTimeout: resourceNetworkInterfaceAttachmentDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"attachment_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"device_index": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},
			names.AttrInstanceID: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"network_card_index": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			names.AttrNetworkInterfaceID: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrStatus: {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceNetworkInterfaceAttachmentCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	input := ec2.AttachNetworkInterfaceInput{
		NetworkInterfaceId: aws.String(d.Get(names.AttrNetworkInterfaceID).(string)),
		InstanceId:         aws.String(d.Get(names.AttrInstanceID).(string)),
		DeviceIndex:        aws.Int32(int32(d.Get("device_index").(int))),
	}

	if v, ok := d.GetOk("network_card_index"); ok {
		if v, ok := v.(int); ok {
			input.NetworkCardIndex = aws.Int32(int32(v))
		}
	}

	attachmentID, err := attachNetworkInterface(ctx, conn, &input)

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	if attachmentID != "" {
		d.SetId(attachmentID)
	}

	return append(diags, resourceNetworkInterfaceAttachmentRead(ctx, d, meta)...)
}

func resourceNetworkInterfaceAttachmentRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	eni, err := findNetworkInterfaceByAttachmentID(ctx, conn, d.Id())

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] EC2 Network Interface Attachment (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 Network Interface Attachment (%s): %s", d.Id(), err)
	}

	attachment := eni.Attachment
	d.Set("attachment_id", attachment.AttachmentId)
	d.Set("device_index", attachment.DeviceIndex)
	d.Set(names.AttrInstanceID, attachment.InstanceId)
	d.Set("network_card_index", attachment.NetworkCardIndex)
	d.Set(names.AttrNetworkInterfaceID, eni.NetworkInterfaceId)
	d.Set(names.AttrStatus, eni.Attachment.Status)

	return diags
}

func resourceNetworkInterfaceAttachmentDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	if err := detachNetworkInterface(ctx, conn, d.Get(names.AttrNetworkInterfaceID).(string), d.Id(), networkInterfaceDetachedTimeout); err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	return diags
}
