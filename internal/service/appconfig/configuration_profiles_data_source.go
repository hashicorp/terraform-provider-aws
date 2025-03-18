// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appconfig

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/appconfig"
	awstypes "github.com/aws/aws-sdk-go-v2/service/appconfig/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_appconfig_configuration_profiles", name="Configuration Profiles")
func dataSourceConfigurationProfiles() *schema.Resource {
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

func dataSourceConfigurationProfilesRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppConfigClient(ctx)

	applicationID := d.Get(names.AttrApplicationID).(string)
	input := appconfig.ListConfigurationProfilesInput{
		ApplicationId: aws.String(applicationID),
	}
	output, err := findConfigurationProfileSummaries(ctx, conn, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading AppConfig Configuration Profiles: %s", err)
	}

	d.SetId(applicationID)
	d.Set("configuration_profile_ids", tfslices.ApplyToAll(output, func(v awstypes.ConfigurationProfileSummary) string {
		return aws.ToString(v.Id)
	}))

	return diags
}

func findConfigurationProfileSummaries(ctx context.Context, conn *appconfig.Client, input *appconfig.ListConfigurationProfilesInput) ([]awstypes.ConfigurationProfileSummary, error) {
	var output []awstypes.ConfigurationProfileSummary

	pages := appconfig.NewListConfigurationProfilesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.Items...)
	}

	return output, nil
}
