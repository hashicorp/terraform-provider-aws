// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appconfig

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/appconfig"
	awstypes "github.com/aws/aws-sdk-go-v2/service/appconfig/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_appconfig_configuration_profiles", name="Configuration Profiles")
func DataSourceConfigurationProfiles() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceConfigurationProfilesRead,
		Schema: map[string]*schema.Schema{
			names.AttrApplicationID: {
				Type:     schema.TypeString,
				Required: true,
			},
			"configuration_profile_ids": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

const (
	DSNameConfigurationProfiles = "Configuration Profiles Data Source"
)

func dataSourceConfigurationProfilesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).AppConfigClient(ctx)
	appId := d.Get(names.AttrApplicationID).(string)

	out, err := findConfigurationProfileSummariesByApplication(ctx, conn, appId)
	if err != nil {
		return create.AppendDiagError(diags, names.AppConfig, create.ErrActionReading, DSNameConfigurationProfiles, appId, err)
	}

	d.SetId(appId)

	var configIds []*string
	for _, v := range out {
		configIds = append(configIds, v.Id)
	}

	d.Set("configuration_profile_ids", aws.ToStringSlice(configIds))

	return diags
}

func findConfigurationProfileSummariesByApplication(ctx context.Context, conn *appconfig.Client, applicationId string) ([]awstypes.ConfigurationProfileSummary, error) {
	var outputs []awstypes.ConfigurationProfileSummary
	pages := appconfig.NewListConfigurationProfilesPaginator(conn, &appconfig.ListConfigurationProfilesInput{
		ApplicationId: aws.String(applicationId),
	})

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		outputs = append(outputs, page.Items...)
	}

	return outputs, nil
}
