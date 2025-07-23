// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sagemaker

import (
	"context"
	"log"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sagemaker"
	awstypes "github.com/aws/aws-sdk-go-v2/service/sagemaker/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_sagemaker_hub", name="Hub")
// @Tags(identifierAttribute="arn")
func resourceHub() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceHubCreate,
		ReadWithoutTimeout:   resourceHubRead,
		UpdateWithoutTimeout: resourceHubUpdate,
		DeleteWithoutTimeout: resourceHubDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"hub_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 63),
					validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z](-*[0-9A-Za-z]){0,62}$`),
						"Valid characters are a-z, A-Z, 0-9, and - (hyphen)."),
				),
			},
			"hub_description": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 1024),
			},
			"hub_display_name": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
			},
			"hub_search_keywords": {
				Type:     schema.TypeSet,
				MaxItems: 50,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Optional: true,
			},
			"s3_storage_config": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"s3_output_path": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
							ValidateFunc: validation.All(
								validation.StringMatch(regexache.MustCompile(`^(https|s3)://([^/])/?(.*)$`), ""),
								validation.StringLenBetween(1, 1024),
							),
						},
					},
				},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourceHubCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerClient(ctx)

	name := d.Get("hub_name").(string)
	input := &sagemaker.CreateHubInput{
		HubName:        aws.String(name),
		HubDescription: aws.String(d.Get("hub_description").(string)),
		Tags:           getTagsIn(ctx),
	}

	if v, ok := d.GetOk("hub_display_name"); ok {
		input.HubDisplayName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("hub_search_keywords"); ok {
		input.HubSearchKeywords = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("s3_storage_config"); ok {
		input.S3StorageConfig = expandS3StorageConfig(v.([]any))
	}

	_, err := conn.CreateHub(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating SageMaker AI Hub %s: %s", name, err)
	}

	d.SetId(name)

	if _, err := waitHubInService(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for SageMaker AI Hub (%s) to be InService: %s", d.Id(), err)
	}

	return append(diags, resourceHubRead(ctx, d, meta)...)
}

func resourceHubRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerClient(ctx)

	hub, err := findHubByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		d.SetId("")
		log.Printf("[WARN] Unable to find SageMaker AI Hub (%s); removing from state", d.Id())
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SageMaker AI Hub (%s): %s", d.Id(), err)
	}

	d.Set("hub_name", hub.HubName)
	d.Set(names.AttrARN, hub.HubArn)
	d.Set("hub_description", hub.HubDescription)
	d.Set("hub_display_name", hub.HubDisplayName)
	d.Set("hub_search_keywords", flex.FlattenStringValueSet(hub.HubSearchKeywords))

	if err := d.Set("s3_storage_config", flattenS3StorageConfig(hub.S3StorageConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting s3_storage_config for SageMaker AI Hub (%s): %s", d.Id(), err)
	}

	return diags
}

func resourceHubUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		modifyOpts := &sagemaker.UpdateHubInput{
			HubName: aws.String(d.Id()),
		}

		if d.HasChange("hub_description") {
			modifyOpts.HubDescription = aws.String(d.Get("hub_description").(string))
		}

		if d.HasChange("hub_display_name") {
			modifyOpts.HubDisplayName = aws.String(d.Get("hub_display_name").(string))
		}

		if d.HasChange("hub_search_keywords") {
			modifyOpts.HubSearchKeywords = flex.ExpandStringValueSet(d.Get("hub_search_keywords").(*schema.Set))
		}

		if _, err := conn.UpdateHub(ctx, modifyOpts); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating SageMaker AI Hub (%s): %s", d.Id(), err)
		}

		if _, err := waitHubUpdated(ctx, conn, d.Id()); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for SageMaker AI Hub (%s) to be updated: %s", d.Id(), err)
		}
	}

	return append(diags, resourceHubRead(ctx, d, meta)...)
}

func resourceHubDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerClient(ctx)

	input := &sagemaker.DeleteHubInput{
		HubName: aws.String(d.Id()),
	}

	if _, err := conn.DeleteHub(ctx, input); err != nil {
		if errs.IsA[*awstypes.ResourceNotFound](err) {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "deleting SageMaker AI Hub (%s): %s", d.Id(), err)
	}

	if _, err := waitHubDeleted(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for SageMaker AI Hub (%s) to delete: %s", d.Id(), err)
	}

	return diags
}

func findHubByName(ctx context.Context, conn *sagemaker.Client, name string) (*sagemaker.DescribeHubOutput, error) {
	input := &sagemaker.DescribeHubInput{
		HubName: aws.String(name),
	}

	output, err := conn.DescribeHub(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFound](err) {
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

func expandS3StorageConfig(configured []any) *awstypes.HubS3StorageConfig {
	if len(configured) == 0 {
		return nil
	}

	m := configured[0].(map[string]any)

	c := &awstypes.HubS3StorageConfig{}

	if v, ok := m["s3_output_path"].(string); ok && v != "" {
		c.S3OutputPath = aws.String(v)
	}

	return c
}

func flattenS3StorageConfig(config *awstypes.HubS3StorageConfig) []map[string]any {
	if config == nil {
		return []map[string]any{}
	}

	m := map[string]any{}

	if config.S3OutputPath != nil {
		m["s3_output_path"] = aws.ToString(config.S3OutputPath)
	}

	return []map[string]any{m}
}
