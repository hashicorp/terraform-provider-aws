// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssm

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_ssm_instances, name="Instances")
func dataSourceInstances() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceInstancesRead,

		Schema: map[string]*schema.Schema{
			names.AttrFilter: {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrName: {
							Type:     schema.TypeString,
							Required: true,
						},
						names.AttrValues: {
							Type:     schema.TypeList,
							Required: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			names.AttrIDs: {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceInstancesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSMClient(ctx)

	input := &ssm.DescribeInstanceInformationInput{}

	if v, ok := d.GetOk(names.AttrFilter); ok {
		input.Filters = expandInstanceInformationStringFilters(v.(*schema.Set).List())
	}

	var output []awstypes.InstanceInformation

	pages := ssm.NewDescribeInstanceInformationPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading SSM Instances: %s", err)
		}

		output = append(output, page.InstanceInformationList...)
	}

	d.SetId(meta.(*conns.AWSClient).Region)
	d.Set(names.AttrIDs, tfslices.ApplyToAll(output, func(v awstypes.InstanceInformation) string {
		return aws.ToString(v.InstanceId)
	}))

	return diags
}

func expandInstanceInformationStringFilters(tfList []interface{}) []awstypes.InstanceInformationStringFilter {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.InstanceInformationStringFilter

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandInstanceInformationStringFilter(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func expandInstanceInformationStringFilter(tfMap map[string]interface{}) *awstypes.InstanceInformationStringFilter {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.InstanceInformationStringFilter{}

	if v, ok := tfMap[names.AttrName].(string); ok && v != "" {
		apiObject.Key = aws.String(v)
	}

	if v, ok := tfMap[names.AttrValues].([]interface{}); ok && len(v) > 0 {
		apiObject.Values = flex.ExpandStringValueList(v)
	}

	return apiObject
}
