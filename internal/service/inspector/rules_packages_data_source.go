// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package inspector

import (
	"context"
	"sort"

	"github.com/aws/aws-sdk-go-v2/service/inspector"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_inspector_rules_packages")
func DataSourceRulesPackages() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceRulesPackagesRead,

		Schema: map[string]*schema.Schema{
			names.AttrARNs: {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceRulesPackagesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).InspectorClient(ctx)

	output, err := findRulesPackageARNs(ctx, conn)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Inspector Classic Rules Packages: %s", err)
	}
	arns := output
	sort.Strings(arns)

	d.SetId(meta.(*conns.AWSClient).Region)
	d.Set(names.AttrARNs, arns)

	return diags
}

func findRulesPackageARNs(ctx context.Context, conn *inspector.Client) ([]string, error) {
	input := &inspector.ListRulesPackagesInput{}
	var output []string

	pages := inspector.NewListRulesPackagesPaginator(conn, input)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		output = append(output, page.RulesPackageArns...)
	}

	return output, nil
}
