// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appconfig

import (
	"context"
	"fmt"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appconfig"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_appconfig_environment")
func DataSourceEnvironment() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceEnvironmentRead,
		Schema: map[string]*schema.Schema{
			"application_id": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringMatch(regexache.MustCompile(`[a-z\d]{4,7}`), ""),
			},
			"environment_id": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringMatch(regexache.MustCompile(`[a-z\d]{4,7}`), ""),
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
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
			"state": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tftags.TagsSchemaComputed(),
		},
	}
}

const (
	DSNameEnvironment = "Environment Data Source"
)

func dataSourceEnvironmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).AppConfigConn(ctx)

	appID := d.Get("application_id").(string)
	envID := d.Get("environment_id").(string)
	ID := fmt.Sprintf("%s:%s", envID, appID)

	out, err := findEnvironmentByApplicationAndEnvironment(ctx, conn, appID, envID)
	if err != nil {
		return create.AppendDiagError(diags, names.AppConfig, create.ErrActionReading, DSNameEnvironment, ID, err)
	}

	d.SetId(ID)

	d.Set("application_id", appID)
	d.Set("environment_id", envID)
	d.Set("description", out.Description)
	d.Set("name", out.Name)
	d.Set("state", out.State)

	if err := d.Set("monitor", flattenEnvironmentMonitors(out.Monitors)); err != nil {
		return create.AppendDiagError(diags, names.AppConfig, create.ErrActionReading, DSNameEnvironment, ID, err)
	}

	arn := environmentARN(meta.(*conns.AWSClient), appID, envID).String()

	d.Set("arn", arn)

	tags, err := listTags(ctx, conn, arn)

	if err != nil {
		return create.AppendDiagError(diags, names.AppConfig, create.ErrActionReading, DSNameEnvironment, ID, err)
	}

	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig
	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.Map()); err != nil {
		return create.AppendDiagError(diags, names.AppConfig, create.ErrActionSetting, DSNameEnvironment, ID, err)
	}

	return diags
}

func findEnvironmentByApplicationAndEnvironment(ctx context.Context, conn *appconfig.AppConfig, appId string, envId string) (*appconfig.GetEnvironmentOutput, error) {
	res, err := conn.GetEnvironmentWithContext(ctx, &appconfig.GetEnvironmentInput{
		ApplicationId: aws.String(appId),
		EnvironmentId: aws.String(envId),
	})

	if err != nil {
		return nil, err
	}

	return res, nil
}

func flattenEnvironmentMonitors(monitors []*appconfig.Monitor) []interface{} {
	if len(monitors) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, monitor := range monitors {
		if monitor == nil {
			continue
		}

		tfList = append(tfList, flattenEnvironmentMonitor(monitor))
	}

	return tfList
}

func flattenEnvironmentMonitor(monitor *appconfig.Monitor) map[string]interface{} {
	if monitor == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := monitor.AlarmArn; v != nil {
		tfMap["alarm_arn"] = aws.StringValue(v)
	}

	if v := monitor.AlarmRoleArn; v != nil {
		tfMap["alarm_role_arn"] = aws.StringValue(v)
	}

	return tfMap
}
