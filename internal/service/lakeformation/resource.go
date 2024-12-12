// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lakeformation

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lakeformation"
	awstypes "github.com/aws/aws-sdk-go-v2/service/lakeformation/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_lakeformation_resource", name="Resource")
func ResourceResource() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceResourceCreate,
		ReadWithoutTimeout:   resourceResourceRead,
		DeleteWithoutTimeout: resourceResourceDelete,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"hybrid_access_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"last_modified": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrRoleARN: {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"use_service_linked_role": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
			},
			"with_federation": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
		},
	}
}

func resourceResourceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LakeFormationClient(ctx)

	resourceARN := d.Get(names.AttrARN).(string)
	input := &lakeformation.RegisterResourceInput{
		ResourceArn: aws.String(resourceARN),
	}

	if v, ok := d.GetOk("hybrid_access_enabled"); ok {
		input.HybridAccessEnabled = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk(names.AttrRoleARN); ok {
		input.RoleArn = aws.String(v.(string))
	} else {
		input.UseServiceLinkedRole = aws.Bool(true)
	}

	if v, ok := d.GetOkExists("use_service_linked_role"); ok {
		input.UseServiceLinkedRole = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("with_federation"); ok {
		input.WithFederation = aws.Bool(v.(bool))
	}

	_, err := conn.RegisterResource(ctx, input)

	if errs.IsA[*awstypes.AlreadyExistsException](err) {
		log.Printf("[WARN] Lake Formation Resource (%s) already exists", resourceARN)
	} else if err != nil {
		return sdkdiag.AppendErrorf(diags, "registering Lake Formation Resource (%s): %s", resourceARN, err)
	}

	d.SetId(resourceARN)

	return append(diags, resourceResourceRead(ctx, d, meta)...)
}

func resourceResourceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LakeFormationClient(ctx)

	resource, err := FindResourceByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Resource Lake Formation Resource (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Lake Formation Resource (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, d.Id())
	d.Set("hybrid_access_enabled", resource.HybridAccessEnabled)
	if v := resource.LastModified; v != nil { // output not including last modified currently
		d.Set("last_modified", v.Format(time.RFC3339))
	}
	d.Set(names.AttrRoleARN, resource.RoleArn)
	d.Set("with_federation", resource.WithFederation)

	return diags
}

func resourceResourceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LakeFormationClient(ctx)

	log.Printf("[INFO] Deleting Lake Formation Resource: %s", d.Id())
	_, err := conn.DeregisterResource(ctx, &lakeformation.DeregisterResourceInput{
		ResourceArn: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.EntityNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deregistering Lake Formation Resource (%s): %s", d.Id(), err)
	}

	return diags
}

func FindResourceByARN(ctx context.Context, conn *lakeformation.Client, arn string) (*awstypes.ResourceInfo, error) {
	input := &lakeformation.DescribeResourceInput{
		ResourceArn: aws.String(arn),
	}

	output, err := conn.DescribeResource(ctx, input)

	if errs.IsA[*awstypes.EntityNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.ResourceInfo == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.ResourceInfo, nil
}
