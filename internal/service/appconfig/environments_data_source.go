// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appconfig

import (
	"context"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/appconfig"
	awstypes "github.com/aws/aws-sdk-go-v2/service/appconfig/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_appconfig_environments", name="Environments")
func DataSourceEnvironments() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceEnvironmentsRead,
		Schema: map[string]*schema.Schema{
			names.AttrApplicationID: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringMatch(regexache.MustCompile(`[a-z\d]{4,7}`), ""),
			},
			"environment_ids": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

const (
	DSNameEnvironments = "Environments Data Source"
)

func dataSourceEnvironmentsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).AppConfigClient(ctx)
	appID := d.Get(names.AttrApplicationID).(string)

	out, err := findEnvironmentsByApplication(ctx, conn, appID)
	if err != nil {
		return create.AppendDiagError(diags, names.AppConfig, create.ErrActionReading, DSNameEnvironments, appID, err)
	}

	d.SetId(appID)

	var environmentIds []*string
	for _, v := range out {
		environmentIds = append(environmentIds, v.Id)
	}
	d.Set("environment_ids", aws.ToStringSlice(environmentIds))

	return diags
}

func findEnvironmentsByApplication(ctx context.Context, conn *appconfig.Client, appId string) ([]awstypes.Environment, error) {
	var outputs []awstypes.Environment

	pages := appconfig.NewListEnvironmentsPaginator(conn, &appconfig.ListEnvironmentsInput{
		ApplicationId: aws.String(appId),
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
