// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apigateway

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/apigateway"
	awstypes "github.com/aws/aws-sdk-go-v2/service/apigateway/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_api_gateway_vpc_link", name="VPC Link")
// @Tags(identifierAttribute="arn")
func ResourceVPCLink() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceVPCLinkCreate,
		ReadWithoutTimeout:   resourceVPCLinkRead,
		UpdateWithoutTimeout: resourceVPCLinkUpdate,
		DeleteWithoutTimeout: resourceVPCLinkDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"target_arns": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Required: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceVPCLinkCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayClient(ctx)

	input := &apigateway.CreateVpcLinkInput{
		Name:       aws.String(d.Get("name").(string)),
		TargetArns: flex.ExpandStringValueList(d.Get("target_arns").([]interface{})),
		Tags:       getTagsIn(ctx),
	}
	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	resp, err := conn.CreateVpcLink(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating API Gateway VPC Link (%s): %s", d.Get("name").(string), err)
	}

	d.SetId(aws.ToString(resp.Id))

	if err := waitVPCLinkAvailable(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "creating API Gateway VPC Link (%s): waiting for completion: %s", d.Get("name").(string), err)
	}

	return append(diags, resourceVPCLinkRead(ctx, d, meta)...)
}

func resourceVPCLinkRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayClient(ctx)

	input := &apigateway.GetVpcLinkInput{
		VpcLinkId: aws.String(d.Id()),
	}

	resp, err := conn.GetVpcLink(ctx, input)
	if err != nil {
		if !d.IsNewResource() && errs.IsA[*awstypes.NotFoundException](err) {
			log.Printf("[WARN] API Gateway VPC Link %s not found, removing from state", d.Id())
			d.SetId("")
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "reading API Gateway VPC Link (%s): %s", d.Id(), err)
	}

	setTagsOut(ctx, resp.Tags)

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "apigateway",
		Region:    meta.(*conns.AWSClient).Region,
		Resource:  fmt.Sprintf("/vpclinks/%s", d.Id()),
	}.String()
	d.Set("arn", arn)

	d.Set("name", resp.Name)
	d.Set("description", resp.Description)
	if err := d.Set("target_arns", flex.FlattenStringValueList(resp.TargetArns)); err != nil {
		return sdkdiag.AppendErrorf(diags, "reading API Gateway VPC Link (%s): %s", d.Id(), err)
	}

	return diags
}

func resourceVPCLinkUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayClient(ctx)

	operations := make([]awstypes.PatchOperation, 0)

	if d.HasChange("name") {
		operations = append(operations, awstypes.PatchOperation{
			Op:    awstypes.Op("replace"),
			Path:  aws.String("/name"),
			Value: aws.String(d.Get("name").(string)),
		})
	}

	if d.HasChange("description") {
		operations = append(operations, awstypes.PatchOperation{
			Op:    awstypes.Op("replace"),
			Path:  aws.String("/description"),
			Value: aws.String(d.Get("description").(string)),
		})
	}

	input := &apigateway.UpdateVpcLinkInput{
		VpcLinkId:       aws.String(d.Id()),
		PatchOperations: operations,
	}

	_, err := conn.UpdateVpcLink(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating API Gateway VPC Link (%s): %s", d.Id(), err)
	}

	if err := waitVPCLinkAvailable(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "updating API Gateway VPC Link (%s): waiting for completion: %s", d.Id(), err)
	}

	return append(diags, resourceVPCLinkRead(ctx, d, meta)...)
}

func resourceVPCLinkDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayClient(ctx)

	input := &apigateway.DeleteVpcLinkInput{
		VpcLinkId: aws.String(d.Id()),
	}

	_, err := conn.DeleteVpcLink(ctx, input)

	if errs.IsA[*awstypes.NotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting API Gateway VPC Link (%s): %s", d.Id(), err)
	}

	if err := waitVPCLinkDeleted(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for API Gateway VPC Link (%s) deletion: %s", d.Id(), err)
	}

	return diags
}
