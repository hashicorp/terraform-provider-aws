// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package autoscaling

import (
	"context"
	"sort"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/autoscaling"
	awstypes "github.com/aws/aws-sdk-go-v2/service/autoscaling/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_autoscaling_groups", name="Groups")
func dataSourceGroups() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceGroupsRead,

		Schema: map[string]*schema.Schema{
			"arns": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"filter": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrName: {
							Type:     schema.TypeString,
							Required: true,
						},
						"values": {
							Type:     schema.TypeList,
							Required: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			"names": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func buildFiltersDataSource(set *schema.Set) []awstypes.Filter {
	var filters []awstypes.Filter
	for _, v := range set.List() {
		m := v.(map[string]interface{})
		var filterValues []string
		for _, e := range m["values"].([]interface{}) {
			filterValues = append(filterValues, e.(string))
		}

		// In previous iterations, users were expected to provide "key" and "value" tag names.
		// With the addition of asgs filters, the signature is "tag-key" and "tag-value", so these conditions prevent breaking changes.
		// https://docs.aws.amazon.com/sdk-for-go/api/service/autoscaling/#Filter
		name := m[names.AttrName].(string)
		if name == names.AttrKey {
			name = "tag-key"
		}
		if name == names.AttrValue {
			name = "tag-value"
		}
		filters = append(filters, awstypes.Filter{
			Name:   aws.String(name),
			Values: filterValues,
		})
	}
	return filters
}

func dataSourceGroupsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AutoScalingClient(ctx)

	input := &autoscaling.DescribeAutoScalingGroupsInput{}

	if v, ok := d.GetOk("names"); ok && len(v.([]interface{})) > 0 {
		input.AutoScalingGroupNames = flex.ExpandStringValueList(v.([]interface{}))
	}

	if v, ok := d.GetOk("filter"); ok {
		input.Filters = buildFiltersDataSource(v.(*schema.Set))
	}

	groups, err := findGroups(ctx, conn, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Auto Scaling Groups: %s", err)
	}

	var arns, names []string

	for _, group := range groups {
		arns = append(arns, aws.ToString(group.AutoScalingGroupARN))
		names = append(names, aws.ToString(group.AutoScalingGroupName))
	}

	sort.Strings(arns)
	sort.Strings(names)

	d.SetId(meta.(*conns.AWSClient).Region)
	d.Set("arns", arns)
	d.Set("names", names)

	return diags
}
