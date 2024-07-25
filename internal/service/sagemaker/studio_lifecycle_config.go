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
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_sagemaker_studio_lifecycle_config", name="Studio Lifecycle Config")
// @Tags(identifierAttribute="arn")
func ResourceStudioLifecycleConfig() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceStudioLifecycleConfigCreate,
		ReadWithoutTimeout:   resourceStudioLifecycleConfigRead,
		UpdateWithoutTimeout: resourceStudioLifecycleConfigUpdate,
		DeleteWithoutTimeout: resourceStudioLifecycleConfigDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"studio_lifecycle_config_app_type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(sagemaker.StudioLifecycleConfigAppType_Values(), false),
			},
			"studio_lifecycle_config_content": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 16384),
			},
			"studio_lifecycle_config_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 63),
					validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z](-*[0-9A-Za-z])*$`), "Valid characters are a-z, A-Z, 0-9, and - (hyphen)."),
				),
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceStudioLifecycleConfigCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerConn(ctx)

	name := d.Get("studio_lifecycle_config_name").(string)
	input := &sagemaker.CreateStudioLifecycleConfigInput{
		StudioLifecycleConfigName:    aws.String(name),
		StudioLifecycleConfigAppType: aws.String(d.Get("studio_lifecycle_config_app_type").(string)),
		StudioLifecycleConfigContent: aws.String(d.Get("studio_lifecycle_config_content").(string)),
		Tags:                         getTagsIn(ctx),
	}

	log.Printf("[DEBUG] Creating SageMaker Studio Lifecycle Config : %s", input)
	_, err := conn.CreateStudioLifecycleConfigWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating SageMaker Studio Lifecycle Config (%s): %s", name, err)
	}

	d.SetId(name)

	return append(diags, resourceStudioLifecycleConfigRead(ctx, d, meta)...)
}

func resourceStudioLifecycleConfigRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerConn(ctx)

	image, err := FindStudioLifecycleConfigByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] SageMaker Studio Lifecycle Config (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SageMaker Studio Lifecycle Config (%s): %s", d.Id(), err)
	}

	arn := aws.StringValue(image.StudioLifecycleConfigArn)
	d.Set("studio_lifecycle_config_name", image.StudioLifecycleConfigName)
	d.Set("studio_lifecycle_config_app_type", image.StudioLifecycleConfigAppType)
	d.Set("studio_lifecycle_config_content", image.StudioLifecycleConfigContent)
	d.Set(names.AttrARN, arn)

	return diags
}

func resourceStudioLifecycleConfigUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// Tags only.

	return append(diags, resourceStudioLifecycleConfigRead(ctx, d, meta)...)
}

func resourceStudioLifecycleConfigDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerConn(ctx)

	input := &sagemaker.DeleteStudioLifecycleConfigInput{
		StudioLifecycleConfigName: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Deleting SageMaker Studio Lifecycle Config: (%s)", d.Id())
	if _, err := conn.DeleteStudioLifecycleConfigWithContext(ctx, input); err != nil {
		if tfawserr.ErrMessageContains(err, sagemaker.ErrCodeResourceNotFound, "does not exist") {
			return diags
		}

		return sdkdiag.AppendErrorf(diags, "deleting SageMaker Studio Lifecycle Config (%s): %s", d.Id(), err)
	}

	return diags
}
