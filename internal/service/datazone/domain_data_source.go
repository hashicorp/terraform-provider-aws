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
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_datazone_domain", name="Domain")
func newDataSourceDataZoneDomain(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceDataZoneDomain{}, nil
}

const (
	DSNameDataZoneDomain = "Domain Data Source"
)

type dataSourceDataZoneDomain struct {
	framework.DataSourceWithConfigure
}

func (d *dataSourceDataZoneDomain) Metadata(_ context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) { // nosemgrep:ci.meta-in-func-name
	response.TypeName = "aws_datazone_domain"
}

func (d *dataSourceDataZoneDomain) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"arn": framework.ARNAttributeComputedOnly(),
			"id":  framework.IDAttribute(),
			"name": schema.StringAttribute{
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

	domain, err := findDomainByName(ctx, conn, data.Name.ValueString())
	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DataZone, create.ErrActionReading, DSNameDataZoneDomain, data.Name.String(), err),
			err.Error(),
		)
		return
	}

	response.Diagnostics.Append(flex.Flatten(ctx, domain, &data, flex.WithFieldNamePrefix("DataZoneDomain"))...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func findDomainByName(ctx context.Context, conn *datazone.Client, name string) (*awstypes.DomainSummary, error) {
	return _findDomainByName(ctx, conn, name, nil)
}

func _findDomainByName(ctx context.Context, conn *datazone.Client, name string, nextToken *string) (*awstypes.DomainSummary, error) {
	// GetDomain requires a domain identifier, so we need to list all domains and find the one with the matching name.
	domainsInput := &datazone.ListDomainsInput{}

	if nextToken != nil {
		domainsInput.NextToken = aws.String(*nextToken)
	}

	domains, err := conn.ListDomains(ctx, domainsInput)
	if err != nil {
		return nil, err
	}

	if domains == nil {
		return nil, tfresource.NewEmptyResultError(domainsInput)
	}

	for i := range domains.Items {
		domain := domains.Items[i]
		if name == aws.ToString(domain.Name) {
			return &domain, nil
		}
	}

	if domains.NextToken == nil {
		return nil, tfresource.NewEmptyResultError(domainsInput)
	}

	return _findDomainByName(ctx, conn, name, domains.NextToken)
}

type dataSourceDomainModel struct {
	ARN  types.String `tfsdk:"arn"`
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}
