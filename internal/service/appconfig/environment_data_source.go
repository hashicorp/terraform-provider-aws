// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appconfig

import (
	"context"
	"fmt"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/appconfig"
	awstypes "github.com/aws/aws-sdk-go-v2/service/appconfig/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_appconfig_environment", name="Environment")
// @Tags(identifierAttribute="arn")
func DataSourceEnvironment() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceEnvironmentRead,
		Schema: map[string]*schema.Schema{
			names.AttrApplicationID: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringMatch(regexache.MustCompile(`[a-z\d]{4,7}`), ""),
			},
			"environment_id": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringMatch(regexache.MustCompile(`[a-z\d]{4,7}`), ""),
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Computed: true,
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
			names.AttrState: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
		},
	}
}

const (
	DSNameEnvironment = "Environment Data Source"
)

func dataSourceEnvironmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).AppConfigClient(ctx)

	appID := d.Get(names.AttrApplicationID).(string)
	envID := d.Get("environment_id").(string)
	ID := fmt.Sprintf("%s:%s", envID, appID)

	out, err := findEnvironmentByApplicationAndEnvironment(ctx, conn, appID, envID)
	if err != nil {
		return create.AppendDiagError(diags, names.AppConfig, create.ErrActionReading, DSNameEnvironment, ID, err)
	}

	d.SetId(ID)

	d.Set(names.AttrApplicationID, appID)
	d.Set("environment_id", envID)
	d.Set(names.AttrDescription, out.Description)
	d.Set(names.AttrName, out.Name)
	d.Set(names.AttrState, out.State)

	if err := d.Set("monitor", flattenEnvironmentMonitors(out.Monitors)); err != nil {
		return create.AppendDiagError(diags, names.AppConfig, create.ErrActionReading, DSNameEnvironment, ID, err)
	}

	arn := environmentARN(meta.(*conns.AWSClient), appID, envID).String()

	d.Set(names.AttrARN, arn)

	return diags
}

func findEnvironmentByApplicationAndEnvironment(ctx context.Context, conn *appconfig.Client, appId string, envId string) (*appconfig.GetEnvironmentOutput, error) {
	res, err := conn.GetEnvironment(ctx, &appconfig.GetEnvironmentInput{
		ApplicationId: aws.String(appId),
		EnvironmentId: aws.String(envId),
	})

	if err != nil {
		return nil, err
	}

	return res, nil
}

func flattenEnvironmentMonitors(monitors []awstypes.Monitor) []interface{} {
	if len(monitors) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, monitor := range monitors {
		tfList = append(tfList, flattenEnvironmentMonitor(monitor))
	}

	return tfList
}

func flattenEnvironmentMonitor(monitor awstypes.Monitor) map[string]interface{} {
	tfMap := map[string]interface{}{}

	if v := monitor.AlarmArn; v != nil {
		tfMap["alarm_arn"] = aws.ToString(v)
	}

	if v := monitor.AlarmRoleArn; v != nil {
		tfMap["alarm_role_arn"] = aws.ToString(v)
	}

	return tfMap
}
