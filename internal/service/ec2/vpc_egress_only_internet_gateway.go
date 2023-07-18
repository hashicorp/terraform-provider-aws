// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
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

// @SDKResource("aws_egress_only_internet_gateway", name="Egress-only Internet Gateway")
// @Tags(identifierAttribute="id")
func ResourceEgressOnlyInternetGateway() *schema.Resource {
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
			"vpc_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceEgressOnlyInternetGatewayCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn(ctx)

	input := &ec2.CreateEgressOnlyInternetGatewayInput{
		ClientToken:       aws.String(id.UniqueId()),
		TagSpecifications: getTagSpecificationsIn(ctx, ec2.ResourceTypeEgressOnlyInternetGateway),
		VpcId:             aws.String(d.Get("vpc_id").(string)),
	}

	output, err := conn.CreateEgressOnlyInternetGatewayWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EC2 Egress-only Internet Gateway: %s", err)
	}

	d.SetId(aws.StringValue(output.EgressOnlyInternetGateway.EgressOnlyInternetGatewayId))

	return append(diags, resourceEgressOnlyInternetGatewayRead(ctx, d, meta)...)
}

func resourceEgressOnlyInternetGatewayRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn(ctx)

	outputRaw, err := tfresource.RetryWhenNewResourceNotFound(ctx, ec2PropagationTimeout, func() (interface{}, error) {
		return FindEgressOnlyInternetGatewayByID(ctx, conn, d.Id())
	}, d.IsNewResource())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 Egress-only Internet Gateway %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 Egress-only Internet Gateway (%s): %s", d.Id(), err)
	}

	ig := outputRaw.(*ec2.EgressOnlyInternetGateway)

	if len(ig.Attachments) == 1 && aws.StringValue(ig.Attachments[0].State) == ec2.AttachmentStatusAttached {
		d.Set("vpc_id", ig.Attachments[0].VpcId)
	} else {
		d.Set("vpc_id", nil)
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
	conn := meta.(*conns.AWSClient).EC2Conn(ctx)

	log.Printf("[INFO] Deleting EC2 Egress-only Internet Gateway: %s", d.Id())
	_, err := conn.DeleteEgressOnlyInternetGatewayWithContext(ctx, &ec2.DeleteEgressOnlyInternetGatewayInput{
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
