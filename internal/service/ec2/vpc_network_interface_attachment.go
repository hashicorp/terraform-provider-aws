// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
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

func resourceNetworkInterfaceAttachmentCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	attachmentID, err := attachNetworkInterface(ctx, conn,
		d.Get(names.AttrNetworkInterfaceID).(string),
		d.Get(names.AttrInstanceID).(string),
		d.Get("device_index").(int),
		networkInterfaceAttachedTimeout,
	)

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	if attachmentID != "" {
		d.SetId(attachmentID)
	}

	return append(diags, resourceNetworkInterfaceAttachmentRead(ctx, d, meta)...)
}

func resourceNetworkInterfaceAttachmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	network_interface, err := findNetworkInterfaceByAttachmentID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 Network Interface Attachment (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 Network Interface Attachment (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrNetworkInterfaceID, network_interface.NetworkInterfaceId)
	d.Set("attachment_id", network_interface.Attachment.AttachmentId)
	d.Set("device_index", network_interface.Attachment.DeviceIndex)
	d.Set(names.AttrInstanceID, network_interface.Attachment.InstanceId)
	d.Set(names.AttrStatus, network_interface.Attachment.Status)

	return diags
}

func resourceNetworkInterfaceAttachmentDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	if err := detachNetworkInterface(ctx, conn, d.Get(names.AttrNetworkInterfaceID).(string), d.Id(), networkInterfaceDetachedTimeout); err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	return diags
}
