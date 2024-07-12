// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package servicequotas

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/servicequotas"
	awstypes "github.com/aws/aws-sdk-go-v2/service/servicequotas/types"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource(name="Templates")
func newDataSourceTemplates(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceTemplates{}, nil
}

const (
	DSNameTemplates = "Templates Data Source"
)

type dataSourceTemplates struct {
	framework.DataSourceWithConfigure
}

func (d *dataSourceTemplates) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) { // nosemgrep:ci.meta-in-func-name
	resp.TypeName = "aws_servicequotas_templates"
}

func (d *dataSourceTemplates) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID: framework.IDAttribute(),
			names.AttrRegion: schema.StringAttribute{
				Required: true,
			},
		},
		Blocks: map[string]schema.Block{
			"templates": schema.ListNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"global_quota": schema.BoolAttribute{
							Computed: true,
						},
						"quota_code": schema.StringAttribute{
							Computed: true,
						},
						"quota_name": schema.StringAttribute{
							Computed: true,
						},
						names.AttrRegion: schema.StringAttribute{
							Computed: true,
						},
						"service_code": schema.StringAttribute{
							Computed: true,
						},
						names.AttrServiceName: schema.StringAttribute{
							Computed: true,
						},
						names.AttrUnit: schema.StringAttribute{
							Computed: true,
						},
						names.AttrValue: schema.Float64Attribute{
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func (d *dataSourceTemplates) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().ServiceQuotasClient(ctx)

	var data dataSourceTemplatesData
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := servicequotas.ListServiceQuotaIncreaseRequestsInTemplateInput{
		AwsRegion: aws.String(data.Region.ValueString()),
	}
	out, err := conn.ListServiceQuotaIncreaseRequestsInTemplate(ctx, &input)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ServiceQuotas, create.ErrActionReading, DSNameTemplates, data.Region.String(), err),
			err.Error(),
		)
		return
	}

	data.ID = types.StringValue(data.Region.ValueString())

	templates, diags := flattenTemplates(ctx, out.ServiceQuotaIncreaseRequestInTemplateList)
	resp.Diagnostics.Append(diags...)
	data.Templates = templates

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

var templatesSourceAttrTypes = map[string]attr.Type{
	"global_quota":        types.BoolType,
	"quota_code":          types.StringType,
	"quota_name":          types.StringType,
	names.AttrRegion:      types.StringType,
	"service_code":        types.StringType,
	names.AttrServiceName: types.StringType,
	names.AttrUnit:        types.StringType,
	names.AttrValue:       types.Float64Type,
}

type dataSourceTemplatesData struct {
	Region    types.String `tfsdk:"region"`
	ID        types.String `tfsdk:"id"`
	Templates types.List   `tfsdk:"templates"`
}

func flattenTemplates(ctx context.Context, apiObject []awstypes.ServiceQuotaIncreaseRequestInTemplate) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	elemType := types.ObjectType{AttrTypes: templatesSourceAttrTypes}

	elems := []attr.Value{}
	for _, t := range apiObject {
		obj := map[string]attr.Value{
			"global_quota":        types.BoolValue(t.GlobalQuota),
			"quota_code":          flex.StringToFramework(ctx, t.QuotaCode),
			"quota_name":          flex.StringToFramework(ctx, t.QuotaName),
			names.AttrRegion:      flex.StringToFramework(ctx, t.AwsRegion),
			"service_code":        flex.StringToFramework(ctx, t.ServiceCode),
			names.AttrServiceName: flex.StringToFramework(ctx, t.ServiceName),
			names.AttrUnit:        flex.StringToFramework(ctx, t.Unit),
			names.AttrValue:       flex.Float64ToFramework(ctx, t.DesiredValue),
		}
		objVal, d := types.ObjectValue(templatesSourceAttrTypes, obj)
		diags.Append(d...)

		elems = append(elems, objVal)
	}
	listVal, d := types.ListValue(elemType, elems)
	diags.Append(d...)

	return listVal, diags
}
