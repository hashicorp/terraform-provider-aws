// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sagemaker

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sagemaker"
	awstypes "github.com/aws/aws-sdk-go-v2/service/sagemaker/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_sagemaker_image_version", name="Image Version")
func resourceImageVersion() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceImageVersionCreate,
		ReadWithoutTimeout:   resourceImageVersionRead,
		DeleteWithoutTimeout: resourceImageVersionDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"base_image": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"container_image": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"image_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"image_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrVersion: {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func resourceImageVersionCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerClient(ctx)

	name := d.Get("image_name").(string)
	input := &sagemaker.CreateImageVersionInput{
		ImageName: aws.String(name),
		BaseImage: aws.String(d.Get("base_image").(string)),
	}

	_, err := conn.CreateImageVersion(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating SageMaker AI Image Version %s: %s", name, err)
	}

	d.SetId(name)

	if _, err := waitImageVersionCreated(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for SageMaker AI Image Version (%s) to be created: %s", d.Id(), err)
	}

	return append(diags, resourceImageVersionRead(ctx, d, meta)...)
}

func resourceImageVersionRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerClient(ctx)

	image, err := findImageVersionByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		d.SetId("")
		log.Printf("[WARN] Unable to find SageMaker AI Image Version (%s); removing from state", d.Id())
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SageMaker AI Image Version (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, image.ImageVersionArn)
	d.Set("base_image", image.BaseImage)
	d.Set("image_arn", image.ImageArn)
	d.Set("container_image", image.ContainerImage)
	d.Set(names.AttrVersion, image.Version)
	d.Set("image_name", d.Id())

	return diags
}

func resourceImageVersionDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerClient(ctx)

	input := &sagemaker.DeleteImageVersionInput{
		ImageName: aws.String(d.Id()),
		Version:   aws.Int32(int32(d.Get(names.AttrVersion).(int))),
	}

	if _, err := conn.DeleteImageVersion(ctx, input); err != nil {
		if errs.IsAErrorMessageContains[*awstypes.ResourceNotFound](err, "does not exist") {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "deleting SageMaker AI Image Version (%s): %s", d.Id(), err)
	}

	if _, err := waitImageVersionDeleted(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for SageMaker AI Image Version (%s) to delete: %s", d.Id(), err)
	}

	return diags
}

func findImageVersionByName(ctx context.Context, conn *sagemaker.Client, name string) (*sagemaker.DescribeImageVersionOutput, error) {
	input := &sagemaker.DescribeImageVersionInput{
		ImageName: aws.String(name),
	}

	output, err := conn.DescribeImageVersion(ctx, input)

	if errs.IsAErrorMessageContains[*awstypes.ResourceNotFound](err, "does not exist") {
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
