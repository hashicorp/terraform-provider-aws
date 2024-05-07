// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sagemaker

import (
	"context"
	"log"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sagemaker"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_sagemaker_model_package_group", name="Model Package Group")
// @Tags(identifierAttribute="arn")
func ResourceModelPackageGroup() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceModelPackageGroupCreate,
		ReadWithoutTimeout:   resourceModelPackageGroupRead,
		UpdateWithoutTimeout: resourceModelPackageGroupUpdate,
		DeleteWithoutTimeout: resourceModelPackageGroupDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"model_package_group_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 63),
					validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z](-*[0-9A-Za-z]){0,62}$`),
						"Valid characters are a-z, A-Z, 0-9, and - (hyphen)."),
				),
			},
			"model_package_group_description": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 1024),
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceModelPackageGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerConn(ctx)

	name := d.Get("model_package_group_name").(string)
	input := &sagemaker.CreateModelPackageGroupInput{
		ModelPackageGroupName: aws.String(name),
		Tags:                  getTagsIn(ctx),
	}

	if v, ok := d.GetOk("model_package_group_description"); ok {
		input.ModelPackageGroupDescription = aws.String(v.(string))
	}

	_, err := conn.CreateModelPackageGroupWithContext(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating SageMaker Model Package Group %s: %s", name, err)
	}

	d.SetId(name)

	if _, err := WaitModelPackageGroupCompleted(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for SageMaker Model Package Group (%s) to be created: %s", d.Id(), err)
	}

	return append(diags, resourceModelPackageGroupRead(ctx, d, meta)...)
}

func resourceModelPackageGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerConn(ctx)

	mpg, err := FindModelPackageGroupByName(ctx, conn, d.Id())
	if err != nil {
		if tfawserr.ErrMessageContains(err, "ValidationException", "does not exist") {
			d.SetId("")
			log.Printf("[WARN] Unable to find SageMaker Model Package Group (%s); removing from state", d.Id())
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "reading SageMaker Model Package Group (%s): %s", d.Id(), err)
	}

	arn := aws.StringValue(mpg.ModelPackageGroupArn)
	d.Set("model_package_group_name", mpg.ModelPackageGroupName)
	d.Set(names.AttrARN, arn)
	d.Set("model_package_group_description", mpg.ModelPackageGroupDescription)

	return diags
}

func resourceModelPackageGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// Tags only.

	return append(diags, resourceModelPackageGroupRead(ctx, d, meta)...)
}

func resourceModelPackageGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerConn(ctx)

	input := &sagemaker.DeleteModelPackageGroupInput{
		ModelPackageGroupName: aws.String(d.Id()),
	}

	if _, err := conn.DeleteModelPackageGroupWithContext(ctx, input); err != nil {
		if tfawserr.ErrMessageContains(err, "ValidationException", "does not exist") {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "deleting SageMaker Model Package Group (%s): %s", d.Id(), err)
	}

	if _, err := WaitModelPackageGroupDeleted(ctx, conn, d.Id()); err != nil {
		if tfawserr.ErrMessageContains(err, "ValidationException", "does not exist") {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "waiting for SageMaker Model Package Group (%s) to delete: %s", d.Id(), err)
	}

	return diags
}
