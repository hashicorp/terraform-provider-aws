// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appconfig

import (
	"context"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/appconfig/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_appconfig_environment", name="Environment")
// @Tags(identifierAttribute="arn")
func dataSourceEnvironment() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceEnvironmentRead,

		Schema: map[string]*schema.Schema{
			names.AttrApplicationID: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringMatch(regexache.MustCompile(`[a-z\d]{4,7}`), ""),
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"environment_id": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringMatch(regexache.MustCompile(`[a-z\d]{4,7}`), ""),
			},
			"monitor": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"alarm_arn": {
							Computed: true,
							Type:     schema.TypeString,
						},
						"alarm_role_arn": {
							Computed: true,
							Type:     schema.TypeString,
						},
					},
				},
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrState: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceEnvironmentRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppConfigClient(ctx)

	applicationID := d.Get(names.AttrApplicationID).(string)
	environmentID := d.Get("environment_id").(string)
	id := environmentCreateResourceID(environmentID, applicationID)

	out, err := findEnvironmentByTwoPartKey(ctx, conn, applicationID, environmentID)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading AppConfig Environment (%s): %s", id, err)
	}

	d.SetId(id)
	d.Set(names.AttrApplicationID, applicationID)
	d.Set(names.AttrARN, environmentARN(ctx, meta.(*conns.AWSClient), applicationID, environmentID))
	d.Set(names.AttrDescription, out.Description)
	d.Set("environment_id", environmentID)
	if err := d.Set("monitor", flattenEnvironmentMonitors(out.Monitors)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting monitor: %s", err)
	}
	d.Set(names.AttrName, out.Name)
	d.Set(names.AttrState, out.State)

	return diags
}

func flattenEnvironmentMonitors(monitors []awstypes.Monitor) []any {
	if len(monitors) == 0 {
		return nil
	}

	var tfList []any

	for _, monitor := range monitors {
		tfList = append(tfList, flattenEnvironmentMonitor(monitor))
	}

	return tfList
}

func flattenEnvironmentMonitor(monitor awstypes.Monitor) map[string]any {
	tfMap := map[string]any{}

	if v := monitor.AlarmArn; v != nil {
		tfMap["alarm_arn"] = aws.ToString(v)
	}

	if v := monitor.AlarmRoleArn; v != nil {
		tfMap["alarm_role_arn"] = aws.ToString(v)
	}

	return tfMap
}
