// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package licensemanager

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/licensemanager"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_licensemanager_grants")
func DataSourceDistributedGrants() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceDistributedGrantsRead,
		Schema: map[string]*schema.Schema{
			names.AttrFilter: DataSourceFiltersSchema(),
			names.AttrARNs: {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceDistributedGrantsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LicenseManagerConn(ctx)

	in := &licensemanager.ListDistributedGrantsInput{}

	in.Filters = BuildFiltersDataSource(
		d.Get(names.AttrFilter).(*schema.Set),
	)

	if len(in.Filters) == 0 {
		in.Filters = nil
	}

	out, err := FindDistributedDistributedGrants(ctx, conn, in)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Distributes Grants: %s", err)
	}

	var grantARNs []string

	for _, v := range out {
		grantARNs = append(grantARNs, aws.StringValue(v.GrantArn))
	}

	d.SetId(meta.(*conns.AWSClient).Region)
	d.Set(names.AttrARNs, grantARNs)

	return diags
}

func FindDistributedDistributedGrants(ctx context.Context, conn *licensemanager.LicenseManager, in *licensemanager.ListDistributedGrantsInput) ([]*licensemanager.Grant, error) {
	var out []*licensemanager.Grant

	err := listDistributedGrantsPages(ctx, conn, in, func(page *licensemanager.ListDistributedGrantsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Grants {
			if v != nil {
				out = append(out, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, licensemanager.ErrCodeResourceNotFoundException) {
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
