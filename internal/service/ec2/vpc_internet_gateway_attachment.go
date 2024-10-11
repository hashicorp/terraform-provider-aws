// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_internet_gateway_attachment", name="Internet Gateway Attachment")
func resourceInternetGatewayAttachment() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceInternetGatewayAttachmentCreate,
		ReadWithoutTimeout:   resourceInternetGatewayAttachmentRead,
		DeleteWithoutTimeout: resourceInternetGatewayAttachmentDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(20 * time.Minute),
			Delete: schema.DefaultTimeout(20 * time.Minute),
		},

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"internet_gateway_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrVPCID: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceInternetGatewayAttachmentCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	igwID := d.Get("internet_gateway_id").(string)
	vpcID := d.Get(names.AttrVPCID).(string)

	if err := attachInternetGateway(ctx, conn, igwID, vpcID, d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EC2 Internet Gateway Attachment: %s", err)
	}

	d.SetId(internetGatewayAttachmentCreateResourceID(igwID, vpcID))

	return append(diags, resourceInternetGatewayAttachmentRead(ctx, d, meta)...)
}

func resourceInternetGatewayAttachmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	igwID, vpcID, err := internetGatewayAttachmentParseResourceID(d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 Internet Gateway Attachment (%s): %s", d.Id(), err)
	}

	outputRaw, err := tfresource.RetryWhenNewResourceNotFound(ctx, ec2PropagationTimeout, func() (interface{}, error) {
		return findInternetGatewayAttachment(ctx, conn, igwID, vpcID)
	}, d.IsNewResource())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 Internet Gateway Attachment %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 Internet Gateway Attachment (%s): %s", d.Id(), err)
	}

	igw := outputRaw.(*awstypes.InternetGatewayAttachment)

	d.Set("internet_gateway_id", igwID)
	d.Set(names.AttrVPCID, igw.VpcId)

	return diags
}

func resourceInternetGatewayAttachmentDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	igwID, vpcID, err := internetGatewayAttachmentParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting EC2 Internet Gateway Attachment (%s): %s", d.Id(), err)
	}

	if err := detachInternetGateway(ctx, conn, igwID, vpcID, d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting EC2 Internet Gateway Attachment (%s): %s", d.Id(), err)
	}

	return diags
}

const internetGatewayAttachmentIDSeparator = ":"

func internetGatewayAttachmentCreateResourceID(igwID, vpcID string) string {
	parts := []string{igwID, vpcID}
	id := strings.Join(parts, internetGatewayAttachmentIDSeparator)

	return id
}

func internetGatewayAttachmentParseResourceID(id string) (string, string, error) {
	parts := strings.Split(id, internetGatewayAttachmentIDSeparator)

	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected INTERNET-GATEWAY-ID%[2]sVPC-ID", id, internetGatewayAttachmentIDSeparator)
}
