// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package servicecatalog

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicecatalog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_servicecatalog_launch_paths")
func DataSourceLaunchPaths() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceLaunchPathsRead,

		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(LaunchPathsReadyTimeout),
		},

		Schema: map[string]*schema.Schema{
			"accept_language": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "en",
				ValidateFunc: validation.StringInSlice(AcceptLanguage_Values(), false),
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

func dataSourceLaunchPathsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ServiceCatalogConn(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	summaries, err := WaitLaunchPathsReady(ctx, conn, d.Get("accept_language").(string), d.Get("product_id").(string), d.Timeout(schema.TimeoutRead))

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "describing Service Catalog Launch Paths: %s", err)
	}

	if err := d.Set("summaries", flattenLaunchPathSummaries(ctx, summaries, ignoreTagsConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting summaries: %s", err)
	}

	d.SetId(d.Get("product_id").(string))

	return diags
}

func flattenLaunchPathSummary(ctx context.Context, apiObject *servicecatalog.LaunchPathSummary, ignoreTagsConfig *tftags.IgnoreConfig) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if len(apiObject.ConstraintSummaries) > 0 {
		tfMap["constraint_summaries"] = flattenConstraintSummaries(apiObject.ConstraintSummaries)
	}

	if apiObject.Id != nil {
		tfMap["path_id"] = aws.StringValue(apiObject.Id)
	}

	if apiObject.Name != nil {
		tfMap[names.AttrName] = aws.StringValue(apiObject.Name)
	}

	tags := KeyValueTags(ctx, apiObject.Tags)

	tfMap[names.AttrTags] = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()

	return tfMap
}

func flattenLaunchPathSummaries(ctx context.Context, apiObjects []*servicecatalog.LaunchPathSummary, ignoreTagsConfig *tftags.IgnoreConfig) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenLaunchPathSummary(ctx, apiObject, ignoreTagsConfig))
	}

	return tfList
}

func flattenConstraintSummary(apiObject *servicecatalog.ConstraintSummary) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject.Description != nil {
		tfMap[names.AttrDescription] = aws.StringValue(apiObject.Description)
	}

	if apiObject.Type != nil {
		tfMap[names.AttrType] = aws.StringValue(apiObject.Type)
	}

	return tfMap
}

func flattenConstraintSummaries(apiObjects []*servicecatalog.ConstraintSummary) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenConstraintSummary(apiObject))
	}

	return tfList
}
