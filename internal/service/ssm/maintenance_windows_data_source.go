// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssm

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
)

// @SDKDataSource("aws_ssm_maintenance_windows")
func DataSourceMaintenanceWindows() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataMaintenanceWindowsRead,
		Schema: map[string]*schema.Schema{
			"filter": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
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
			"ids": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataMaintenanceWindowsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSMConn(ctx)

	input := &ssm.DescribeMaintenanceWindowsInput{}

	if v, ok := d.GetOk("filter"); ok {
		input.Filters = expandMaintenanceWindowFilters(v.(*schema.Set).List())
	}

	var results []*ssm.MaintenanceWindowIdentity

	err := conn.DescribeMaintenanceWindowsPagesWithContext(ctx, input, func(page *ssm.DescribeMaintenanceWindowsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, windowIdentities := range page.WindowIdentities {
			if windowIdentities == nil {
				continue
			}

			results = append(results, windowIdentities)
		}

		return !lastPage
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SSM Maintenance Windows: %s", err)
	}

	var windowIDs []string

	for _, r := range results {
		windowIDs = append(windowIDs, aws.StringValue(r.WindowId))
	}

	d.SetId(meta.(*conns.AWSClient).Region)
	d.Set("ids", windowIDs)

	return diags
}

func expandMaintenanceWindowFilters(tfList []interface{}) []*ssm.MaintenanceWindowFilter {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*ssm.MaintenanceWindowFilter

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandMaintenanceWindowFilter(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandMaintenanceWindowFilter(tfMap map[string]interface{}) *ssm.MaintenanceWindowFilter {
	if tfMap == nil {
		return nil
	}

	apiObject := &ssm.MaintenanceWindowFilter{}

	if v, ok := tfMap["name"].(string); ok && v != "" {
		apiObject.Key = aws.String(v)
	}

	if v, ok := tfMap["values"].([]interface{}); ok && len(v) > 0 {
		apiObject.Values = flex.ExpandStringList(v)
	}

	return apiObject
}
