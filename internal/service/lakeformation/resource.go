// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lakeformation

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lakeformation"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKResource("aws_lakeformation_resource")
func ResourceResource() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceResourceCreate,
		ReadWithoutTimeout:   resourceResourceRead,
		DeleteWithoutTimeout: resourceResourceDelete,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"last_modified": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"role_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
		},
	}
}

func resourceResourceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LakeFormationConn(ctx)
	resourceArn := d.Get("arn").(string)

	input := &lakeformation.RegisterResourceInput{
		ResourceArn: aws.String(resourceArn),
	}

	if v, ok := d.GetOk("role_arn"); ok {
		input.RoleArn = aws.String(v.(string))
	} else {
		input.UseServiceLinkedRole = aws.Bool(true)
	}

	_, err := conn.RegisterResourceWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, lakeformation.ErrCodeAlreadyExistsException) {
		log.Printf("[WARN] Lake Formation Resource (%s) already exists", resourceArn)
	} else if err != nil {
		return sdkdiag.AppendErrorf(diags, "registering Lake Formation Resource (%s): %s", resourceArn, err)
	}

	d.SetId(resourceArn)
	return append(diags, resourceResourceRead(ctx, d, meta)...)
}

func resourceResourceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LakeFormationConn(ctx)
	resourceArn := d.Get("arn").(string)

	input := &lakeformation.DescribeResourceInput{
		ResourceArn: aws.String(resourceArn),
	}

	output, err := conn.DescribeResourceWithContext(ctx, input)

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, lakeformation.ErrCodeEntityNotFoundException) {
		log.Printf("[WARN] Resource Lake Formation Resource (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading resource Lake Formation Resource (%s): %s", d.Id(), err)
	}

	if output == nil || output.ResourceInfo == nil {
		return sdkdiag.AppendErrorf(diags, "reading resource Lake Formation Resource (%s): empty response", d.Id())
	}

	// d.Set("arn", output.ResourceInfo.ResourceArn) // output not including resource arn currently
	d.Set("role_arn", output.ResourceInfo.RoleArn)
	if output.ResourceInfo.LastModified != nil { // output not including last modified currently
		d.Set("last_modified", output.ResourceInfo.LastModified.Format(time.RFC3339))
	}

	return diags
}

func resourceResourceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LakeFormationConn(ctx)
	resourceArn := d.Get("arn").(string)

	input := &lakeformation.DeregisterResourceInput{
		ResourceArn: aws.String(resourceArn),
	}

	_, err := conn.DeregisterResourceWithContext(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deregistering Lake Formation Resource (%s): %s", d.Id(), err)
	}

	return diags
}
