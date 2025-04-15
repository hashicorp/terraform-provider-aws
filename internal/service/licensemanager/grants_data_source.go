// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package licensemanager

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/licensemanager"
	awstypes "github.com/aws/aws-sdk-go-v2/service/licensemanager/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/namevaluesfilters"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_licensemanager_grants", name="Grants")
func dataSourceGrants() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceDistributedGrantsRead,

		Schema: map[string]*schema.Schema{
			names.AttrARNs: {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrFilter: namevaluesfilters.Schema(),
		},
	}
}

func dataSourceDistributedGrantsRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LicenseManagerClient(ctx)

	input := &licensemanager.ListDistributedGrantsInput{}

	if v, ok := d.GetOk(names.AttrFilter); ok && v.(*schema.Set).Len() > 0 {
		input.Filters = namevaluesfilters.New(v.(*schema.Set)).LicenseManagerFilters()
	}

	grants, err := findDistributedGrants(ctx, conn, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading License Manager Distributed Grants: %s", err)
	}

	d.SetId(meta.(*conns.AWSClient).Region(ctx))
	d.Set(names.AttrARNs, tfslices.ApplyToAll(grants, func(v awstypes.Grant) string {
		return aws.ToString(v.GrantArn)
	}))

	return diags
}

func findDistributedGrants(ctx context.Context, conn *licensemanager.Client, input *licensemanager.ListDistributedGrantsInput) ([]awstypes.Grant, error) {
	var output []awstypes.Grant

	err := listDistributedGrantsPages(ctx, conn, input, func(page *licensemanager.ListDistributedGrantsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		output = append(output, page.Grants...)

		return !lastPage
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}
