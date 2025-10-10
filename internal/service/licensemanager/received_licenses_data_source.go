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

// @SDKDataSource("aws_licensemanager_received_licenses", name="Received Licenses")
func dataSourceReceivedLicenses() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceReceivedLicensesRead,

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

func dataSourceReceivedLicensesRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LicenseManagerClient(ctx)

	input := &licensemanager.ListReceivedLicensesInput{}

	if v, ok := d.GetOk(names.AttrFilter); ok && v.(*schema.Set).Len() > 0 {
		input.Filters = namevaluesfilters.New(v.(*schema.Set)).LicenseManagerFilters()
	}

	licenses, err := findReceivedLicenses(ctx, conn, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading License Manager Received Licenses: %s", err)
	}

	d.SetId(meta.(*conns.AWSClient).Region(ctx))
	d.Set(names.AttrARNs, tfslices.ApplyToAll(licenses, func(v awstypes.GrantedLicense) string {
		return aws.ToString(v.LicenseArn)
	}))

	return diags
}

func findReceivedLicenses(ctx context.Context, conn *licensemanager.Client, in *licensemanager.ListReceivedLicensesInput) ([]awstypes.GrantedLicense, error) {
	var output []awstypes.GrantedLicense

	err := listReceivedLicensesPages(ctx, conn, in, func(page *licensemanager.ListReceivedLicensesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		output = append(output, page.Licenses...)

		return !lastPage
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}
