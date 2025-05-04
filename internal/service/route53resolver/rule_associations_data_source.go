// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53resolver

import (
	"context"

	"github.com/aws/aws-sdk-go/service/route53resolver"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtype "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for datasource registration to the Provider. DO NOT EDIT.
// @FrameworkDataSource(name="Rule Associations")
func newDataSourceRuleAssociations(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceRuleAssociations{}, nil
}

const (
	DSNameRuleAssociations = "Rule Associations Data Source"
)

type dataSourceRuleAssociations struct {
	framework.DataSourceWithConfigure
}

func (d *dataSourceRuleAssociations) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) { // nosemgrep:ci.meta-in-func-name
	resp.TypeName = "aws_route53_resolver_rule_associations"
}

func (d *dataSourceRuleAssociations) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Blocks: map[string]schema.Block{
			"associations": schema.ListNestedBlock{
				CustomType: fwtype.NewListNestedObjectTypeOf[DataSourceRuleAssociation](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed: true,
						},
						"name": schema.StringAttribute{
							Computed: true,
						},
						"resolver_rule_id": schema.StringAttribute{
							Computed: true,
						},
						"status": schema.StringAttribute{
							Computed: true,
						},
						"status_message": schema.StringAttribute{
							Computed: true,
						},
						"vpc_id": schema.StringAttribute{
							Computed: true,
						},
					},
				},
			},
			"filter": schema.SetNestedBlock{
				CustomType: fwtype.NewSetNestedObjectTypeOf[Filter](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Required: true,
						},
						"values": schema.SetAttribute{
							ElementType: types.StringType,
							Required:    true,
						},
					},
				},
			},
		},
	}
}

func (d *dataSourceRuleAssociations) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().Route53ResolverConn(ctx)

	var data DataSourceRuleAssociationsData
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	filters, diags := getFiltersFromData(ctx, &data)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	input := &route53resolver.ListResolverRuleAssociationsInput{Filters: filters}
	associations := []*DataSourceRuleAssociation{}
	err := conn.ListResolverRuleAssociationsPagesWithContext(ctx, input, func(page *route53resolver.ListResolverRuleAssociationsOutput, lastPage bool) bool {
		var ruleAssociation DataSourceRuleAssociation
		for _, out := range page.ResolverRuleAssociations {
			diags := flex.Flatten(ctx, out, &ruleAssociation)
			if diags.HasError() {
				resp.Diagnostics.Append(diags...)
				return true
			}
			associations = append(associations, &ruleAssociation)
		}

		return !lastPage
	})

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Route53Resolver, create.ErrActionReading, DSNameRuleAssociations, DSNameRuleAssociations, err),
			err.Error(),
		)
		return
	}

	if len(associations) != 0 {
		data.ResolverRuleAssociations, diags = fwtype.NewListNestedObjectValueOfSlice[DataSourceRuleAssociation](ctx, associations)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func getFiltersFromData(ctx context.Context, data *DataSourceRuleAssociationsData) ([]*route53resolver.Filter, diag.Diagnostics) {
	filters := []*route53resolver.Filter{}
	df, diags := data.Filters.ToSlice(ctx)
	if diags.HasError() {
		return nil, diags
	}

	for _, f := range df {
		filter := &route53resolver.Filter{
			Name:   f.Name.ValueStringPointer(),
			Values: flex.ExpandFrameworkStringSet(ctx, f.Values),
		}
		filters = append(filters, filter)
	}

	return filters, nil
}

type DataSourceRuleAssociationsData struct {
	Filters                  fwtype.SetNestedObjectValueOf[Filter]                     `tfsdk:"filter"`
	ResolverRuleAssociations fwtype.ListNestedObjectValueOf[DataSourceRuleAssociation] `tfsdk:"associations"`
}

type DataSourceRuleAssociation struct {
	Id             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	ResolverRuleId types.String `tfsdk:"resolver_rule_id"`
	Status         types.String `tfsdk:"status"`
	StatusMessage  types.String `tfsdk:"status_message"`
	VPCId          types.String `tfsdk:"vpc_id"`
}

type Filter struct {
	Name   types.String                    `tfsdk:"name"`
	Values fwtype.SetValueOf[types.String] `tfsdk:"values"`
}
