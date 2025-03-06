// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package datazone

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/datazone"
	awstypes "github.com/aws/aws-sdk-go-v2/service/datazone/types"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_datazone_domain", name="Domain")
func newDataSourceDataZoneDomain(_ context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceDataZoneDomain{}, nil
}

const (
	DSNameDataZoneDomain = "Domain Data Source"
)

type dataSourceDataZoneDomain struct {
	framework.DataSourceWithConfigure
}

func (d *dataSourceDataZoneDomain) Schema(_ context.Context, _ datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrID:  framework.IDAttribute(),
			names.AttrName: schema.StringAttribute{
				Required: true,
			},
		},
	}
}

func (d *dataSourceDataZoneDomain) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	conn := d.Meta().DataZoneClient(ctx)

	var data dataSourceDomainModel
	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	output, err := findDomain(ctx, conn, func(domain *awstypes.DomainSummary) bool {
		return aws.ToString(domain.Name) == data.Name.ValueString()
	})

	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DataZone, create.ErrActionReading, DSNameDataZoneDomain, data.Name.String(), err),
			err.Error(),
		)
		return
	}

	response.Diagnostics.Append(flex.Flatten(ctx, output, &data, flex.WithFieldNamePrefix("DataZoneDomain"))...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func findDomain(ctx context.Context, conn *datazone.Client, filter tfslices.Predicate[*awstypes.DomainSummary]) (*awstypes.DomainSummary, error) {
	domain, err := findDomains(ctx, conn, filter)
	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(domain)
}

func findDomains(ctx context.Context, conn *datazone.Client, filter tfslices.Predicate[*awstypes.DomainSummary]) ([]awstypes.DomainSummary, error) {
	var output []awstypes.DomainSummary

	pages := datazone.NewListDomainsPaginator(conn, &datazone.ListDomainsInput{})
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, domain := range page.Items {
			if filter(&domain) {
				output = append(output, domain)
			}
		}
	}

	return output, nil
}

type dataSourceDomainModel struct {
	ARN  types.String `tfsdk:"arn"`
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}
