// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sagemaker

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sagemaker"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_sagemaker_image_version")
func ResourceImageVersion() *schema.Resource {
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

func resourceImageVersionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerConn(ctx)

	name := d.Get("image_name").(string)
	input := &sagemaker.CreateImageVersionInput{
		ImageName: aws.String(name),
		BaseImage: aws.String(d.Get("base_image").(string)),
	}

	_, err := conn.CreateImageVersionWithContext(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating SageMaker Image Version %s: %s", name, err)
	}

	d.SetId(name)

	if _, err := WaitImageVersionCreated(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for SageMaker Image Version (%s) to be created: %s", d.Id(), err)
	}

	return append(diags, resourceImageVersionRead(ctx, d, meta)...)
}

func resourceImageVersionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerConn(ctx)

	image, err := FindImageVersionByName(ctx, conn, d.Id())
	if err != nil {
		if tfawserr.ErrMessageContains(err, sagemaker.ErrCodeResourceNotFound, "does not exist") {
			d.SetId("")
			log.Printf("[WARN] Unable to find SageMaker Image Version (%s); removing from state", d.Id())
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "reading SageMaker Image Version (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, image.ImageVersionArn)
	d.Set("base_image", image.BaseImage)
	d.Set("image_arn", image.ImageArn)
	d.Set("container_image", image.ContainerImage)
	d.Set(names.AttrVersion, image.Version)
	d.Set("image_name", d.Id())

	return diags
}

func resourceImageVersionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerConn(ctx)

	input := &sagemaker.DeleteImageVersionInput{
		ImageName: aws.String(d.Id()),
		Version:   aws.Int64(int64(d.Get(names.AttrVersion).(int))),
	}

	if _, err := conn.DeleteImageVersionWithContext(ctx, input); err != nil {
		if tfawserr.ErrMessageContains(err, sagemaker.ErrCodeResourceNotFound, "does not exist") {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "deleting SageMaker Image Version (%s): %s", d.Id(), err)
	}

	if _, err := WaitImageVersionDeleted(ctx, conn, d.Id()); err != nil {
		if tfawserr.ErrMessageContains(err, sagemaker.ErrCodeResourceNotFound, "does not exist") {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "waiting for SageMaker Image Version (%s) to delete: %s", d.Id(), err)
	}

	return diags
}
