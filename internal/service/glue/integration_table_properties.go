// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package glue

import (
	"context"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/glue"
	awstypes "github.com/aws/aws-sdk-go-v2/service/glue/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_glue_integration_table_properties", name="Integration Table Properties")
func newIntegrationTablePropertiesResource(context.Context) (resource.ResourceWithConfigure, error) {
	r := &integrationTablePropertiesResource{}

	r.SetDefaultCreateTimeout(10 * time.Minute)
	r.SetDefaultUpdateTimeout(10 * time.Minute)
	r.SetDefaultDeleteTimeout(10 * time.Minute)

	return r, nil
}

type integrationTablePropertiesResource struct {
	framework.ResourceWithModel[integrationTablePropertiesModel]
	framework.WithTimeouts
}

type sourceTableConfigModel struct {
	Fields            fwtypes.ListOfString `tfsdk:"fields"`
	FilterPredicate   types.String         `tfsdk:"filter_predicate"`
	PrimaryKey        fwtypes.ListOfString `tfsdk:"primary_key"`
	RecordUpdateField types.String         `tfsdk:"record_update_field"`
}

type partitionSpecModel struct {
	ConversionSpec types.String `tfsdk:"conversion_spec"`
	FieldName      types.String `tfsdk:"field_name"`
	FunctionSpec   types.String `tfsdk:"function_spec"`
}

type targetTableConfigModel struct {
	PartitionSpec   []partitionSpecModel `tfsdk:"partition_spec"`
	TargetTableName types.String         `tfsdk:"target_table_name"`
	UnnestSpec      types.String         `tfsdk:"unnest_spec"`
}

type integrationTablePropertiesModel struct {
	framework.WithRegionModel
	ResourceARN fwtypes.ARN             `tfsdk:"resource_arn"`
	TableName   types.String            `tfsdk:"table_name"`
	Source      *sourceTableConfigModel `tfsdk:"source_table_config"`
	Target      *targetTableConfigModel `tfsdk:"target_table_config"`
	Timeouts    timeouts.Value          `tfsdk:"timeouts"`
}

func (r *integrationTablePropertiesResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrResourceARN: schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
			},
			names.AttrTableName: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"source_table_config": schema.SingleNestedBlock{ // nosemgrep:ci.avoid-SingleNestedBlock pre-existing, will be converted
				Attributes: map[string]schema.Attribute{
					"fields": schema.ListAttribute{
						CustomType: fwtypes.ListOfStringType,
						Optional:   true,
						PlanModifiers: []planmodifier.List{
							listplanmodifier.RequiresReplace(),
						},
					},
					"filter_predicate": schema.StringAttribute{Optional: true},
					"primary_key": schema.ListAttribute{
						CustomType: fwtypes.ListOfStringType,
						Optional:   true,
						PlanModifiers: []planmodifier.List{
							listplanmodifier.RequiresReplace(),
						},
					},
					"record_update_field": schema.StringAttribute{Optional: true},
				},
			},
			"target_table_config": schema.SingleNestedBlock{ // nosemgrep:ci.avoid-SingleNestedBlock pre-existing, will be converted
				Attributes: map[string]schema.Attribute{
					"target_table_name": schema.StringAttribute{Optional: true},
					"unnest_spec":       schema.StringAttribute{Optional: true},
				},
				Blocks: map[string]schema.Block{
					"partition_spec": schema.ListNestedBlock{
						NestedObject: schema.NestedBlockObject{
							Attributes: map[string]schema.Attribute{
								"conversion_spec": schema.StringAttribute{Optional: true},
								"field_name":      schema.StringAttribute{Required: true},
								"function_spec":   schema.StringAttribute{Optional: true},
							},
						},
					},
				},
			},
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{Create: true, Update: true, Delete: true}),
		},
	}
}

func (r *integrationTablePropertiesResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data integrationTablePropertiesModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Per AWS API, Source and Target configs cannot both be set in a single request.
	if data.Source != nil && data.Target != nil {
		resp.Diagnostics.AddError("invalid configuration", "source_table_config and target_table_config cannot both be set for aws_glue_integration_table_properties. Set only one per resource instance.")
		return
	}

	conn := r.Meta().GlueClient(ctx)
	input := glue.CreateIntegrationTablePropertiesInput{
		ResourceArn: flex.StringFromFramework(ctx, data.ResourceARN),
		TableName:   data.TableName.ValueStringPointer(),
	}
	if data.Source != nil {
		var stc awstypes.SourceTableConfig
		resp.Diagnostics.Append(flex.Expand(ctx, data.Source, &stc)...)
		if resp.Diagnostics.HasError() {
			return
		}
		input.SourceTableConfig = &stc
	}
	if data.Target != nil {
		var ttc awstypes.TargetTableConfig
		resp.Diagnostics.Append(flex.Expand(ctx, data.Target, &ttc)...)
		if resp.Diagnostics.HasError() {
			return
		}
		input.TargetTableConfig = &ttc
	}

	_, err := conn.CreateIntegrationTableProperties(ctx, &input)
	if err != nil {
		resp.Diagnostics.AddError("creating Glue Integration Table Properties", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *integrationTablePropertiesResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data integrationTablePropertiesModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().GlueClient(ctx)
	input := glue.GetIntegrationTablePropertiesInput{
		ResourceArn: flex.StringFromFramework(ctx, data.ResourceARN),
		TableName:   data.TableName.ValueStringPointer(),
	}
	out, err := conn.GetIntegrationTableProperties(ctx, &input)
	if err != nil {
		resp.Diagnostics.AddError("reading Glue Integration Table Properties", err.Error())
		return
	}

	// Flatten
	if out.SourceTableConfig != nil {
		var m sourceTableConfigModel
		resp.Diagnostics.Append(flex.Flatten(ctx, out.SourceTableConfig, &m)...)
		if resp.Diagnostics.HasError() {
			return
		}
		data.Source = &m
	}
	if out.TargetTableConfig != nil {
		var m targetTableConfigModel
		resp.Diagnostics.Append(flex.Flatten(ctx, out.TargetTableConfig, &m)...)
		if resp.Diagnostics.HasError() {
			return
		}
		data.Target = &m
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *integrationTablePropertiesResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var new, old integrationTablePropertiesModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &new)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(req.State.Get(ctx, &old)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Per AWS API, Source and Target configs cannot both be set in a single request.
	if new.Source != nil && new.Target != nil {
		resp.Diagnostics.AddError("invalid configuration", "source_table_config and target_table_config cannot both be set for aws_glue_integration_table_properties. Set only one per resource instance.")
		return
	}

	conn := r.Meta().GlueClient(ctx)
	input := glue.UpdateIntegrationTablePropertiesInput{
		ResourceArn: flex.StringFromFramework(ctx, new.ResourceARN),
		TableName:   new.TableName.ValueStringPointer(),
	}
	if new.Source != nil {
		var stc awstypes.SourceTableConfig
		resp.Diagnostics.Append(flex.Expand(ctx, new.Source, &stc)...)
		if resp.Diagnostics.HasError() {
			return
		}
		input.SourceTableConfig = &stc
	}
	if new.Target != nil {
		var ttc awstypes.TargetTableConfig
		resp.Diagnostics.Append(flex.Expand(ctx, new.Target, &ttc)...)
		if resp.Diagnostics.HasError() {
			return
		}
		input.TargetTableConfig = &ttc
	}

	_, err := conn.UpdateIntegrationTableProperties(ctx, &input)
	if err != nil {
		resp.Diagnostics.AddError("updating Glue Integration Table Properties", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &new)...)
}

func (r *integrationTablePropertiesResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data integrationTablePropertiesModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().GlueClient(ctx)
	input := glue.DeleteIntegrationTablePropertiesInput{
		ResourceArn: flex.StringFromFramework(ctx, data.ResourceARN),
		TableName:   data.TableName.ValueStringPointer(),
	}
	_, err := conn.DeleteIntegrationTableProperties(ctx, &input)
	if err != nil {
		resp.Diagnostics.AddError("deleting Glue Integration Table Properties", err.Error())
		return
	}
}

func (r *integrationTablePropertiesResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Expect "resource_arn,table_name"
	parts := strings.Split(req.ID, ",")
	if len(parts) != 2 {
		resp.Diagnostics.AddError("invalid import ID", "expected 'resource_arn,table_name'")
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root(names.AttrResourceARN), types.StringValue(parts[0]))...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root(names.AttrTableName), types.StringValue(parts[1]))...)
}

// NOTE: manual expand/flatten removed in favor of flex.Expand/flex.Flatten
