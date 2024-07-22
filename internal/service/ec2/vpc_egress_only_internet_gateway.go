// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_egress_only_internet_gateway", name="Egress-Only Internet Gateway")
// @Tags(identifierAttribute="id")
// @Testing(tagsTest=false)
func resourceEgressOnlyInternetGateway() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceEgressOnlyInternetGatewayCreate,
		ReadWithoutTimeout:   resourceEgressOnlyInternetGatewayRead,
		UpdateWithoutTimeout: resourceEgressOnlyInternetGatewayUpdate,
		DeleteWithoutTimeout: resourceEgressOnlyInternetGatewayDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		CustomizeDiff: verify.SetTagsDiff,

		Schema: map[string]*schema.Schema{
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			names.AttrVPCID: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceEgressOnlyInternetGatewayCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	input := &ec2.CreateEgressOnlyInternetGatewayInput{
		ClientToken:       aws.String(id.UniqueId()),
		TagSpecifications: getTagSpecificationsIn(ctx, awstypes.ResourceTypeEgressOnlyInternetGateway),
		VpcId:             aws.String(d.Get(names.AttrVPCID).(string)),
	}

	output, err := conn.CreateEgressOnlyInternetGateway(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EC2 Egress-only Internet Gateway: %s", err)
	}

	d.SetId(aws.ToString(output.EgressOnlyInternetGateway.EgressOnlyInternetGatewayId))

	return append(diags, resourceEgressOnlyInternetGatewayRead(ctx, d, meta)...)
}

func resourceEgressOnlyInternetGatewayRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	outputRaw, err := tfresource.RetryWhenNewResourceNotFound(ctx, ec2PropagationTimeout, func() (interface{}, error) {
		return findEgressOnlyInternetGatewayByID(ctx, conn, d.Id())
	}, d.IsNewResource())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 Egress-only Internet Gateway %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 Egress-only Internet Gateway (%s): %s", d.Id(), err)
	}

	ig := outputRaw.(*awstypes.EgressOnlyInternetGateway)

	if len(ig.Attachments) == 1 && ig.Attachments[0].State == awstypes.AttachmentStatusAttached {
		d.Set(names.AttrVPCID, ig.Attachments[0].VpcId)
	} else {
		d.Set(names.AttrVPCID, nil)
	}

	setTagsOut(ctx, ig.Tags)

	return diags
}

func resourceEgressOnlyInternetGatewayUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// Tags only.

	return append(diags, resourceEgressOnlyInternetGatewayRead(ctx, d, meta)...)
}

func resourceEgressOnlyInternetGatewayDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	log.Printf("[INFO] Deleting EC2 Egress-only Internet Gateway: %s", d.Id())
	_, err := conn.DeleteEgressOnlyInternetGateway(ctx, &ec2.DeleteEgressOnlyInternetGatewayInput{
		EgressOnlyInternetGatewayId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidGatewayIDNotFound) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting EC2 Egress-only Internet Gateway (%s): %s", d.Id(), err)
	}

	return diags
}
