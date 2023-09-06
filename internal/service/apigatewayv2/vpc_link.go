// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apigatewayv2

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/apigatewayv2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_apigatewayv2_vpc_link", name="VPC Link")
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
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 128),
			},
			"security_group_ids": {
				Type:     schema.TypeSet,
				Required: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"subnet_ids": {
				Type:     schema.TypeSet,
				Required: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceVPCLinkCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayV2Conn(ctx)

	input := &apigatewayv2.CreateVpcLinkInput{
		Name:             aws.String(d.Get("name").(string)),
		SecurityGroupIds: flex.ExpandStringSet(d.Get("security_group_ids").(*schema.Set)),
		SubnetIds:        flex.ExpandStringSet(d.Get("subnet_ids").(*schema.Set)),
		Tags:             getTagsIn(ctx),
	}

	log.Printf("[DEBUG] Creating API Gateway v2 VPC Link: %s", input)
	resp, err := conn.CreateVpcLinkWithContext(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating API Gateway v2 VPC Link: %s", err)
	}

	d.SetId(aws.StringValue(resp.VpcLinkId))

	if _, err := WaitVPCLinkAvailable(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for API Gateway v2 deployment (%s) availability: %s", d.Id(), err)
	}

	return append(diags, resourceVPCLinkRead(ctx, d, meta)...)
}

func resourceVPCLinkRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayV2Conn(ctx)

	outputRaw, _, err := StatusVPCLink(ctx, conn, d.Id())()
	if tfawserr.ErrCodeEquals(err, apigatewayv2.ErrCodeNotFoundException) && !d.IsNewResource() {
		log.Printf("[WARN] API Gateway v2 VPC Link (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading API Gateway v2 VPC Link (%s): %s", d.Id(), err)
	}

	output := outputRaw.(*apigatewayv2.GetVpcLinkOutput)
	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "apigateway",
		Region:    meta.(*conns.AWSClient).Region,
		Resource:  fmt.Sprintf("/vpclinks/%s", d.Id()),
	}.String()
	d.Set("arn", arn)
	d.Set("name", output.Name)
	if err := d.Set("security_group_ids", flex.FlattenStringSet(output.SecurityGroupIds)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting security_group_ids: %s", err)
	}
	if err := d.Set("subnet_ids", flex.FlattenStringSet(output.SubnetIds)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting subnet_ids: %s", err)
	}

	setTagsOut(ctx, output.Tags)

	return diags
}

func resourceVPCLinkUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayV2Conn(ctx)

	if d.HasChange("name") {
		req := &apigatewayv2.UpdateVpcLinkInput{
			VpcLinkId: aws.String(d.Id()),
			Name:      aws.String(d.Get("name").(string)),
		}

		log.Printf("[DEBUG] Updating API Gateway v2 VPC Link: %s", req)
		_, err := conn.UpdateVpcLinkWithContext(ctx, req)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating API Gateway v2 VPC Link (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceVPCLinkRead(ctx, d, meta)...)
}

func resourceVPCLinkDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayV2Conn(ctx)

	log.Printf("[DEBUG] Deleting API Gateway v2 VPC Link: %s", d.Id())
	_, err := conn.DeleteVpcLinkWithContext(ctx, &apigatewayv2.DeleteVpcLinkInput{
		VpcLinkId: aws.String(d.Id()),
	})
	if tfawserr.ErrCodeEquals(err, apigatewayv2.ErrCodeNotFoundException) {
		return diags
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting API Gateway v2 VPC Link (%s): %s", d.Id(), err)
	}

	_, err = WaitVPCLinkDeleted(ctx, conn, d.Id())
	if tfawserr.ErrCodeEquals(err, apigatewayv2.ErrCodeNotFoundException) {
		return diags
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for API Gateway v2 VPC Link (%s) deletion: %s", d.Id(), err)
	}

	return diags
}
