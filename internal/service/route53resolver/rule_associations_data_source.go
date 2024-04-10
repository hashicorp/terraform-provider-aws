// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53resolver

// **PLEASE DELETE THIS AND ALL TIP COMMENTS BEFORE SUBMITTING A PR FOR REVIEW!**
//
// TIP: ==== INTRODUCTION ====
// Thank you for trying the skaff tool!
//
// You have opted to include these helpful comments. They all include "TIP:"
// to help you find and remove them when you're done with them.
//
// While some aspects of this file are customized to your input, the
// scaffold tool does *not* look at the AWS API and ensure it has correct
// function, structure, and variable names. It makes guesses based on
// commonalities. You will need to make significant adjustments.
//
// In other words, as generated, this is a rough outline of the work you will
// need to do. If something doesn't make sense for your situation, get rid of
// it.

import (
	// TIP: ==== IMPORTS ====
	// This is a common set of imports but not customized to your code since
	// your code hasn't been written yet. Make sure you, your IDE, or
	// goimports -w <file> fixes these imports.
	//
	// The provider linter wants your imports to be in two groups: first,
	// standard library (i.e., "fmt" or "strings"), second, everything else.
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53resolver"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
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
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"name": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"resolver_rule_id": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"status": schema.StringAttribute{
				Computed:   true,
				Optional:   true,
				Validators: []validator.String{stringvalidator.OneOf(route53resolver.ResolverRuleAssociationStatus_Values()...)},
			},
			"status_message": schema.StringAttribute{
				Computed: true,
			},
			"vpc_id": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
		},
	}
}

// TIP: ==== ASSIGN CRUD METHODS ====
// Data sources only have a read method.
func (d *dataSourceRuleAssociations) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	// TIP: ==== DATA SOURCE READ ====
	// Generally, the Read function should do the following things. Make
	// sure there is a good reason if you don't do one of these.
	//
	// 1. Get a client connection to the relevant service
	// 2. Fetch the config
	// 3. Get information about a resource from AWS
	// 4. Set the ID, arguments, and attributes
	// 5. Set the tags
	// 6. Set the state
	// TIP: -- 1. Get a client connection to the relevant service
	conn := d.Meta().Route53ResolverConn(ctx)

	// TIP: -- 2. Fetch the config
	var in dataSourceRuleAssociationsData
	resp.Diagnostics.Append(req.Config.Get(ctx, &in)...)
	if resp.Diagnostics.HasError() {
		return
	}

	filters := getFiltersFromData(&in)

	// TIP: -- 3. Get information about a resource from AWS
	input := &route53resolver.ListResolverRuleAssociationsInput{Filters: filters}
	data := []dataSourceRuleAssociationsData{}
	err := conn.ListResolverRuleAssociationsPagesWithContext(ctx, input, func(page *route53resolver.ListResolverRuleAssociationsOutput, lastPage bool) bool {
		var ruleAssociation dataSourceRuleAssociationsData
		for _, out := range page.ResolverRuleAssociations {
			diags := flex.Flatten(ctx, out, ruleAssociation)
			if diags.HasError() {
				resp.Diagnostics.Append(diags...)
				return true
			}

			data = append(data, ruleAssociation)
		}

		return !lastPage
	})
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Route53Resolver, create.ErrActionReading, DSNameRuleAssociations, data.Name.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func getFiltersFromData(data *dataSourceRuleAssociationsData) []*route53resolver.Filter {
	filters := []*route53resolver.Filter{}
	if !data.Name.IsNull() {
		filters = append(filters, &route53resolver.Filter{Name: aws.String("Name"), Values: []*string{data.Name.ValueStringPointer()}})
	}

	if !data.RuleResolverId.IsNull() {
		filters = append(filters, &route53resolver.Filter{Name: aws.String("ResolverRuleId"), Values: []*string{data.RuleResolverId.ValueStringPointer()}})
	}

	if !data.Status.IsNull() {
		filters = append(filters, &route53resolver.Filter{Name: aws.String("Status"), Values: []*string{data.Status.ValueStringPointer()}})
	}

	if !data.VpcId.IsNull() {
		filters = append(filters, &route53resolver.Filter{Name: aws.String("VPCId"), Values: []*string{data.VpcId.ValueStringPointer()}})
	}

	return filters
}

// TIP: ==== DATA STRUCTURES ====
// With Terraform Plugin-Framework configurations are deserialized into
// Go types, providing type safety without the need for type assertions.
// These structs should match the schema definition exactly, and the `tfsdk`
// tag value should match the attribute name.
//
// Nested objects are represented in their own data struct. These will
// also have a corresponding attribute type mapping for use inside flex
// functions.
//
// See more:
// https://developer.hashicorp.com/terraform/plugin/framework/handling-data/accessing-values
type dataSourceRuleAssociationsData struct {
	Id             types.String `tfsdk:"arn"`
	Name           types.String `tfsdk:"complex_argument"`
	RuleResolverId types.String `tfsdk:"description"`
	Status         types.String `tfsdk:"id"`
	StatusMessage  types.String `tfsdk:"name"`
	VpcId          types.String `tfsdk:"type"`
}
