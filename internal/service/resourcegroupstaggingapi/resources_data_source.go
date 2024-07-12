// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package resourcegroupstaggingapi

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/resourcegroupstaggingapi"
	"github.com/aws/aws-sdk-go-v2/service/resourcegroupstaggingapi/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_resourcegroupstaggingapi_resources")
func dataSourceResources() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceResourcesRead,

		Schema: map[string]*schema.Schema{
			"exclude_compliant_resources": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"include_compliance_details": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"resource_arn_list": {
				Type:          schema.TypeSet,
				Optional:      true,
				Elem:          &schema.Schema{Type: schema.TypeString},
				ConflictsWith: []string{"tag_filter"},
			},
			"resource_type_filters": {
				Type:          schema.TypeSet,
				Optional:      true,
				MaxItems:      100,
				Elem:          &schema.Schema{Type: schema.TypeString},
				ConflictsWith: []string{"resource_arn_list"},
			},
			"tag_filter": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 50,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrKey: {
							Type:     schema.TypeString,
							Required: true,
						},
						names.AttrValues: {
							Type:     schema.TypeSet,
							Optional: true,
							MaxItems: 20,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			"resource_tag_mapping_list": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrResourceARN: {
							Type:     schema.TypeString,
							Computed: true,
						},
						"compliance_details": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"compliance_status": {
										Type:     schema.TypeBool,
										Computed: true,
									},
									"keys_with_noncompliant_values": {
										Type:     schema.TypeSet,
										Computed: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									"non_compliant_keys": {
										Type:     schema.TypeSet,
										Computed: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
								},
							},
						},
						names.AttrTags: tftags.TagsSchemaComputed(),
					},
				},
			},
		},
	}
}

func dataSourceResourcesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ResourceGroupsTaggingAPIClient(ctx)

	input := &resourcegroupstaggingapi.GetResourcesInput{}

	if v, ok := d.GetOk("include_compliance_details"); ok {
		input.IncludeComplianceDetails = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("exclude_compliant_resources"); ok {
		input.ExcludeCompliantResources = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("resource_arn_list"); ok && v.(*schema.Set).Len() > 0 {
		input.ResourceARNList = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("tag_filter"); ok {
		input.TagFilters = expandTagFilters(v.([]interface{}))
	}

	if v, ok := d.GetOk("resource_type_filters"); ok && v.(*schema.Set).Len() > 0 {
		input.ResourceTypeFilters = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	var taggings []types.ResourceTagMapping

	pages := resourcegroupstaggingapi.NewGetResourcesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading Resource Groups Tagging API Resources: %s", err)
		}

		taggings = append(taggings, page.ResourceTagMappingList...)
	}

	d.SetId(meta.(*conns.AWSClient).Partition)

	if err := d.Set("resource_tag_mapping_list", flattenResourceTagMappings(ctx, taggings)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting resource tag mapping list: %s", err)
	}

	return diags
}

func expandTagFilters(filters []interface{}) []types.TagFilter {
	result := make([]types.TagFilter, len(filters))

	for i, filter := range filters {
		m := filter.(map[string]interface{})

		result[i] = types.TagFilter{
			Key: aws.String(m[names.AttrKey].(string)),
		}

		if v, ok := m[names.AttrValues]; ok && v.(*schema.Set).Len() > 0 {
			result[i].Values = flex.ExpandStringValueSet(v.(*schema.Set))
		}
	}

	return result
}

func flattenResourceTagMappings(ctx context.Context, list []types.ResourceTagMapping) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(list))

	for _, i := range list {
		l := map[string]interface{}{
			names.AttrResourceARN: aws.ToString(i.ResourceARN),
			names.AttrTags:        KeyValueTags(ctx, i.Tags).Map(),
		}

		if i.ComplianceDetails != nil {
			l["compliance_details"] = flattenComplianceDetails(i.ComplianceDetails)
		}

		result = append(result, l)
	}

	return result
}

func flattenComplianceDetails(details *types.ComplianceDetails) []map[string]interface{} {
	if details == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"compliance_status":             aws.ToBool(details.ComplianceStatus),
		"keys_with_noncompliant_values": details.KeysWithNoncompliantValues,
		"non_compliant_keys":            details.NoncompliantKeys,
	}

	return []map[string]interface{}{m}
}
