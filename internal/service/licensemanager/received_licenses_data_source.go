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
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_licensemanager_received_licenses")
func DataSourceReceivedLicenses() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceReceivedLicensesRead,
		Schema: map[string]*schema.Schema{
			names.AttrARNs: {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrFilter: DataSourceFiltersSchema(),
		},
	}
}

func dataSourceReceivedLicensesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LicenseManagerClient(ctx)

	in := &licensemanager.ListReceivedLicensesInput{}

	in.Filters = BuildFiltersDataSource(
		d.Get(names.AttrFilter).(*schema.Set),
	)

	if len(in.Filters) == 0 {
		in.Filters = nil
	}

	out, err := FindReceivedLicenses(ctx, conn, in)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Received Licenses: %s", err)
	}

	var licenseARNs []string

	for _, v := range out {
		licenseARNs = append(licenseARNs, aws.ToString(v.LicenseArn))
	}

	d.SetId(meta.(*conns.AWSClient).Region)
	d.Set(names.AttrARNs, licenseARNs)

	return diags
}

func FindReceivedLicenses(ctx context.Context, conn *licensemanager.Client, in *licensemanager.ListReceivedLicensesInput) ([]awstypes.GrantedLicense, error) {
	var out []awstypes.GrantedLicense

	err := listReceivedLicensesPages(ctx, conn, in, func(page *licensemanager.ListReceivedLicensesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		out = append(out, page.Licenses...)

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

	return out, nil
}
