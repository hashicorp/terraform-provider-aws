// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package servicecatalog

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/servicecatalog/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_servicecatalog_launch_paths", name="Launch Paths")
func dataSourceLaunchPaths() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceLaunchPathsRead,

		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(LaunchPathsReadyTimeout),
		},

		Schema: map[string]*schema.Schema{
			"accept_language": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      acceptLanguageEnglish,
				ValidateFunc: validation.StringInSlice(acceptLanguage_Values(), false),
			},
			"product_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"summaries": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"constraint_summaries": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrDescription: {
										Type:     schema.TypeString,
										Computed: true,
									},
									names.AttrType: {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
						"path_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrName: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrTags: tftags.TagsSchemaComputed(),
					},
				},
			},
		},
	}
}

func dataSourceLaunchPathsRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ServiceCatalogClient(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig(ctx)

	summaries, err := waitLaunchPathsReady(ctx, conn, d.Get("accept_language").(string), d.Get("product_id").(string), d.Timeout(schema.TimeoutRead))

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "describing Service Catalog Launch Paths: %s", err)
	}

	if err := d.Set("summaries", flattenLaunchPathSummaries(ctx, summaries, ignoreTagsConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting summaries: %s", err)
	}

	d.SetId(d.Get("product_id").(string))

	return diags
}

func flattenLaunchPathSummary(ctx context.Context, apiObject awstypes.LaunchPathSummary, ignoreTagsConfig *tftags.IgnoreConfig) map[string]any {
	tfMap := map[string]any{}

	if len(apiObject.ConstraintSummaries) > 0 {
		tfMap["constraint_summaries"] = flattenConstraintSummaries(apiObject.ConstraintSummaries)
	}

	if apiObject.Id != nil {
		tfMap["path_id"] = aws.ToString(apiObject.Id)
	}

	if apiObject.Name != nil {
		tfMap[names.AttrName] = aws.ToString(apiObject.Name)
	}

	tags := keyValueTags(ctx, apiObject.Tags)

	tfMap[names.AttrTags] = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()

	return tfMap
}

func flattenLaunchPathSummaries(ctx context.Context, apiObjects []awstypes.LaunchPathSummary, ignoreTagsConfig *tftags.IgnoreConfig) []any {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []any

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenLaunchPathSummary(ctx, apiObject, ignoreTagsConfig))
	}

	return tfList
}

func flattenConstraintSummary(apiObject awstypes.ConstraintSummary) map[string]any {
	tfMap := map[string]any{}

	if apiObject.Description != nil {
		tfMap[names.AttrDescription] = aws.ToString(apiObject.Description)
	}

	if apiObject.Type != nil {
		tfMap[names.AttrType] = aws.ToString(apiObject.Type)
	}

	return tfMap
}

func flattenConstraintSummaries(apiObjects []awstypes.ConstraintSummary) []any {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []any

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenConstraintSummary(apiObject))
	}

	return tfList
}
