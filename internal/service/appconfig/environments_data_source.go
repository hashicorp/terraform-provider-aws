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
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_appconfig_environments", name="Environments")
func dataSourceEnvironments() *schema.Resource {
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

func dataSourceEnvironmentsRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppConfigClient(ctx)

	applicationID := d.Get(names.AttrApplicationID).(string)
	input := appconfig.ListEnvironmentsInput{
		ApplicationId: aws.String(applicationID),
	}

	output, err := findEnvironments(ctx, conn, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading AppConfig Environments: %s", err)
	}

	d.SetId(applicationID)
	d.Set("environment_ids", tfslices.ApplyToAll(output, func(v awstypes.Environment) string {
		return aws.ToString(v.Id)
	}))

	return diags
}

func findEnvironments(ctx context.Context, conn *appconfig.Client, input *appconfig.ListEnvironmentsInput) ([]awstypes.Environment, error) {
	var output []awstypes.Environment

	pages := appconfig.NewListEnvironmentsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		output = append(output, page.Items...)
	}

	return output, nil
}
