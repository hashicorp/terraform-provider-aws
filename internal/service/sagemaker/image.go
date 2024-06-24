// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sagemaker

import (
	"context"
	"log"
	"time"

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

// @SDKResource("aws_sagemaker_image", name="Image")
// @Tags(identifierAttribute="arn")
func ResourceImage() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceImageCreate,
		ReadWithoutTimeout:   resourceImageRead,
		UpdateWithoutTimeout: resourceImageUpdate,
		DeleteWithoutTimeout: resourceImageDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},

			"image_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 63),
					validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z](-*[0-9A-Za-z])*$`), "Valid characters are a-z, A-Z, 0-9, and - (hyphen)."),
				),
			},
			names.AttrRoleARN: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			names.AttrDisplayName: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 128),
			},
			names.AttrDescription: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 512),
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceImageCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerConn(ctx)

	name := d.Get("image_name").(string)
	input := &sagemaker.CreateImageInput{
		ImageName: aws.String(name),
		RoleArn:   aws.String(d.Get(names.AttrRoleARN).(string)),
		Tags:      getTagsIn(ctx),
	}

	if v, ok := d.GetOk(names.AttrDisplayName); ok {
		input.DisplayName = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	// for some reason even if the operation is retried the same error response is given even though the role is valid. a short sleep before creation solves it.
	time.Sleep(1 * time.Minute)
	_, err := conn.CreateImageWithContext(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating SageMaker Image %s: %s", name, err)
	}

	d.SetId(name)

	if _, err := WaitImageCreated(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for SageMaker Image (%s) to be created: %s", d.Id(), err)
	}

	return append(diags, resourceImageRead(ctx, d, meta)...)
}

func resourceImageRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerConn(ctx)

	image, err := FindImageByName(ctx, conn, d.Id())
	if err != nil {
		if tfawserr.ErrMessageContains(err, sagemaker.ErrCodeResourceNotFound, "does not exist") {
			d.SetId("")
			log.Printf("[WARN] Unable to find SageMaker Image (%s); removing from state", d.Id())
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "reading SageMaker Image (%s): %s", d.Id(), err)
	}

	arn := aws.StringValue(image.ImageArn)
	d.Set("image_name", image.ImageName)
	d.Set(names.AttrARN, arn)
	d.Set(names.AttrRoleARN, image.RoleArn)
	d.Set(names.AttrDisplayName, image.DisplayName)
	d.Set(names.AttrDescription, image.Description)

	return diags
}

func resourceImageUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerConn(ctx)
	needsUpdate := false

	input := &sagemaker.UpdateImageInput{
		ImageName: aws.String(d.Id()),
	}

	var deleteProperties []*string

	if d.HasChange(names.AttrDescription) {
		if v, ok := d.GetOk(names.AttrDescription); ok {
			input.Description = aws.String(v.(string))
		} else {
			deleteProperties = append(deleteProperties, aws.String("Description"))
			input.DeleteProperties = deleteProperties
		}
		needsUpdate = true
	}

	if d.HasChange(names.AttrDisplayName) {
		if v, ok := d.GetOk(names.AttrDisplayName); ok {
			input.DisplayName = aws.String(v.(string))
		} else {
			deleteProperties = append(deleteProperties, aws.String("DisplayName"))
			input.DeleteProperties = deleteProperties
		}
		needsUpdate = true
	}

	if needsUpdate {
		log.Printf("[DEBUG] sagemaker Image update config: %#v", *input)
		_, err := conn.UpdateImageWithContext(ctx, input)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating SageMaker Image: %s", err)
		}

		if _, err := WaitImageCreated(ctx, conn, d.Id()); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for SageMaker Image (%s) to update: %s", d.Id(), err)
		}
	}

	return append(diags, resourceImageRead(ctx, d, meta)...)
}

func resourceImageDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerConn(ctx)

	input := &sagemaker.DeleteImageInput{
		ImageName: aws.String(d.Id()),
	}

	if _, err := conn.DeleteImageWithContext(ctx, input); err != nil {
		if tfawserr.ErrMessageContains(err, sagemaker.ErrCodeResourceNotFound, "No Image with the name") {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "deleting SageMaker Image (%s): %s", d.Id(), err)
	}

	if _, err := WaitImageDeleted(ctx, conn, d.Id()); err != nil {
		if tfawserr.ErrMessageContains(err, sagemaker.ErrCodeResourceNotFound, "does not exist") {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "waiting for SageMaker Image (%s) to delete: %s", d.Id(), err)
	}

	return diags
}
