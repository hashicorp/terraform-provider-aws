// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package datazone

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/datazone"
	awstypes "github.com/aws/aws-sdk-go-v2/service/datazone/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/datasourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_datazone_domain", name="Domain")
func newDataSourceDomain(_ context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceDomain{}, nil
}

const (
	DSNameDomain = "Domain Data Source"
)

type dataSourceDomain struct {
	framework.DataSourceWithConfigure
}

func (d *dataSourceDomain) Schema(_ context.Context, _ datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrCreatedAt: schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			names.AttrDescription: schema.StringAttribute{
				Computed: true,
			},
			"domain_version": schema.StringAttribute{
				Computed: true,
			},
			names.AttrID: schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"last_updated_at": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			"managed_account_id": schema.StringAttribute{
				Computed: true,
			},
			names.AttrName: schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"portal_url": schema.StringAttribute{
				Computed: true,
			},
			names.AttrStatus: schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (d *dataSourceDomain) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	conn := d.Meta().DataZoneClient(ctx)

	var data dataSourceDomainModel
	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	// set filter for input attribute
	var filter tfslices.Predicate[*awstypes.DomainSummary]
	if !data.Name.IsNull() {
		filter = func(domain *awstypes.DomainSummary) bool {
			return aws.ToString(domain.Name) == data.Name.ValueString()
		}
	}

	if !data.ID.IsNull() {
		filter = func(domain *awstypes.DomainSummary) bool {
			return aws.ToString(domain.Id) == data.ID.ValueString()
		}
	}

	output, err := findDomain(ctx, conn, filter)
	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DataZone, create.ErrActionReading, DSNameDomain, data.Name.String(), err),
			err.Error(),
		)
		return
	}

	response.Diagnostics.Append(flex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (d *dataSourceDomain) ConfigValidators(_ context.Context) []datasource.ConfigValidator {
	return []datasource.ConfigValidator{
		datasourcevalidator.AtLeastOneOf(
			path.MatchRoot(names.AttrName),
			path.MatchRoot(names.AttrID),
		),
	}
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
	ARN              types.String      `tfsdk:"arn"`
	CreatedAt        timetypes.RFC3339 `tfsdk:"created_at"`
	Description      types.String      `tfsdk:"description"`
	DomainVersion    types.String      `tfsdk:"domain_version"`
	ID               types.String      `tfsdk:"id"`
	LastUpdatedAt    timetypes.RFC3339 `tfsdk:"last_updated_at"`
	ManagedAccountID types.String      `tfsdk:"managed_account_id"`
	Name             types.String      `tfsdk:"name"`
	PortalURL        types.String      `tfsdk:"portal_url"`
	Status           types.String      `tfsdk:"status"`
}
