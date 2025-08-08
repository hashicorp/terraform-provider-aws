// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sagemaker

import (
	"context"
	"log"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sagemaker"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_sagemaker_model_package_group", name="Model Package Group")
// @Tags(identifierAttribute="arn")
func resourceModelPackageGroup() *schema.Resource {
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
	}
}

func resourceModelPackageGroupCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerClient(ctx)

	name := d.Get("model_package_group_name").(string)
	input := &sagemaker.CreateModelPackageGroupInput{
		ModelPackageGroupName: aws.String(name),
		Tags:                  getTagsIn(ctx),
	}

	if v, ok := d.GetOk("model_package_group_description"); ok {
		input.ModelPackageGroupDescription = aws.String(v.(string))
	}

	_, err := conn.CreateModelPackageGroup(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating SageMaker AI Model Package Group %s: %s", name, err)
	}

	d.SetId(name)

	if _, err := waitModelPackageGroupCompleted(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for SageMaker AI Model Package Group (%s) to be created: %s", d.Id(), err)
	}

	return append(diags, resourceModelPackageGroupRead(ctx, d, meta)...)
}

func resourceModelPackageGroupRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerClient(ctx)

	mpg, err := findModelPackageGroupByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		d.SetId("")
		log.Printf("[WARN] Unable to find SageMaker AI Model Package Group (%s); removing from state", d.Id())
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SageMaker AI Model Package Group (%s): %s", d.Id(), err)
	}

	d.Set("model_package_group_name", mpg.ModelPackageGroupName)
	d.Set(names.AttrARN, mpg.ModelPackageGroupArn)
	d.Set("model_package_group_description", mpg.ModelPackageGroupDescription)

	return diags
}

func resourceModelPackageGroupUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	// Tags only.

	return append(diags, resourceModelPackageGroupRead(ctx, d, meta)...)
}

func resourceModelPackageGroupDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerClient(ctx)

	input := &sagemaker.DeleteModelPackageGroupInput{
		ModelPackageGroupName: aws.String(d.Id()),
	}

	if _, err := conn.DeleteModelPackageGroup(ctx, input); err != nil {
		if tfawserr.ErrMessageContains(err, ErrCodeValidationException, "does not exist") {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "deleting SageMaker AI Model Package Group (%s): %s", d.Id(), err)
	}

	if _, err := waitModelPackageGroupDeleted(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for SageMaker AI Model Package Group (%s) to delete: %s", d.Id(), err)
	}

	return diags
}

func findModelPackageGroupByName(ctx context.Context, conn *sagemaker.Client, name string) (*sagemaker.DescribeModelPackageGroupOutput, error) {
	input := &sagemaker.DescribeModelPackageGroupInput{
		ModelPackageGroupName: aws.String(name),
	}

	output, err := conn.DescribeModelPackageGroup(ctx, input)

	if tfawserr.ErrMessageContains(err, ErrCodeValidationException, "does not exist") {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}
