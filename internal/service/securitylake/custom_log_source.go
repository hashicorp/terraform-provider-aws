// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package securitylake

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/securitylake"
	awstypes "github.com/aws/aws-sdk-go-v2/service/securitylake/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for resource registration to the Provider. DO NOT EDIT.
// @FrameworkResource(name="Custom Log Source")
func newCustomLogSourceResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceCustomLogSource{}

	return r, nil
}

const (
	ResNameCustomLogSource = "Custom Log Source"
)

type resourceCustomLogSource struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
}

func (r *resourceCustomLogSource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_securitylake_custom_log_source"
}

func (r *resourceCustomLogSource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"event_classes": schema.SetAttribute{
				CustomType:  fwtypes.SetOfStringType,
				ElementType: types.StringType,
				Optional:    true,
			},
			names.AttrID: framework.IDAttribute(),
			"source_name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"source_version": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"source": schema.ListAttribute{
				Optional:   true,
				Computed:   true,
				CustomType: fwtypes.NewListNestedObjectTypeOf[customLogSourceModel](ctx),
				ElementType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"source_name":    types.StringType,
						"source_version": types.StringType,
						"attributes": types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"crawler_arn":  types.StringType,
								"database_arn": types.StringType,
								"table_arn":    types.StringType,
							},
						},
						"provider": types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"role_arn": types.StringType,
								"location": types.StringType,
							},
						},
					},
				},
			},
		},
		Blocks: map[string]schema.Block{
			"configuration": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[customLogSourceConfigurationModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
					listvalidator.SizeAtLeast(1),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"crawler_configuration": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[customLogCrawlerConfigurationModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
								listvalidator.SizeAtLeast(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"role_arn": schema.StringAttribute{
										CustomType: fwtypes.ARNType,
										Required:   true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplace(),
										},
									},
								},
							},
						},
						"provider_identity": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[customLogProviderIdentityModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
								listvalidator.SizeAtLeast(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"external_id": schema.StringAttribute{
										Required: true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplace(),
										},
									},
									"principal": schema.StringAttribute{
										Required: true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplace(),
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func (r *resourceCustomLogSource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().SecurityLakeClient(ctx)

	var data customLogSourceResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := &securitylake.CreateCustomLogSourceInput{}
	resp.Diagnostics.Append(flex.Expand(ctx, data, input)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := conn.CreateCustomLogSource(ctx, input)

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.SecurityLake, create.ErrActionCreating, ResNameAWSLogSource, data.ID.ValueString(), err),
			err.Error(),
		)
		return
	}

	// Set values for unknowns.
	data.ID = flex.StringValueToFramework(ctx, aws.ToString(out.Source.SourceName))
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.SecurityLake, create.ErrActionReading, ResNameAWSLogSource, data.ID.String(), err),
			err.Error(),
		)
		return
	}

	var source customLogSourceModel
	resp.Diagnostics.Append(flex.Flatten(ctx, out.Source, &source)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set values for unknowns after creation is complete.
	data.Source = fwtypes.NewListNestedObjectValueOfPtr(ctx, &source)
	data.SourceName = flex.StringToFramework(ctx, out.Source.SourceName)
	data.SourceVersion = flex.StringToFramework(ctx, out.Source.SourceVersion)

	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

func (r *resourceCustomLogSource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().SecurityLakeClient(ctx)

	var data customLogSourceResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	customLogSource, err := findCustomLogSourceBySourceName(ctx, conn, data.ID.ValueString())

	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.SecurityLake, create.ErrActionReading, ResNameAWSLogSource, data.ID.String(), err),
			err.Error(),
		)
		return
	}

	var source customLogSourceModel
	resp.Diagnostics.Append(flex.Flatten(ctx, customLogSource, &source)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set values for unknowns after creation is complete.
	data.Source = fwtypes.NewListNestedObjectValueOfPtr(ctx, &source)
	data.SourceName = flex.StringToFramework(ctx, customLogSource.SourceName)
	data.SourceVersion = flex.StringToFramework(ctx, customLogSource.SourceVersion)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *resourceCustomLogSource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// NoOP.
}

func (r *resourceCustomLogSource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().SecurityLakeClient(ctx)

	var data customLogSourceResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := &securitylake.DeleteCustomLogSourceInput{}
	resp.Diagnostics.Append(flex.Expand(ctx, data, input)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := conn.DeleteCustomLogSource(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.SecurityLake, create.ErrActionDeleting, ResNameCustomLogSource, data.ID.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceCustomLogSource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func findCustomLogSourceBySourceName(ctx context.Context, conn *securitylake.Client, sourceName string) (*awstypes.CustomLogSourceResource, error) {
	input := &securitylake.ListLogSourcesInput{}
	var output *awstypes.CustomLogSourceResource

	pages := securitylake.NewListLogSourcesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		for _, v := range page.Sources {
			for _, v := range v.Sources {
				if v, ok := v.(*awstypes.LogSourceResourceMemberCustomLogSource); ok {
					if v := v.Value; aws.ToString(v.SourceName) == sourceName {
						if output == nil {
							output = &awstypes.CustomLogSourceResource{
								Attributes:    v.Attributes,
								Provider:      v.Provider,
								SourceName:    v.SourceName,
								SourceVersion: v.SourceVersion,
							}
						}
					}
				}
			}
		}
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(sourceName)
	}

	return output, nil
}

type customLogSourceResourceModel struct {
	ID            types.String                                                       `tfsdk:"id"`
	Configuration fwtypes.ListNestedObjectValueOf[customLogSourceConfigurationModel] `tfsdk:"configuration"`
	Source        fwtypes.ListNestedObjectValueOf[customLogSourceModel]              `tfsdk:"source"`
	SourceName    types.String                                                       `tfsdk:"source_name"`
	SourceVersion types.String                                                       `tfsdk:"source_version"`
	EventClasses  fwtypes.SetValueOf[types.String]                                   `tfsdk:"event_classes"`
}

type customLogSourceConfigurationModel struct {
	CrawlerConfiguration fwtypes.ListNestedObjectValueOf[customLogCrawlerConfigurationModel] `tfsdk:"crawler_configuration"`
	ProviderIdentity     fwtypes.ListNestedObjectValueOf[customLogProviderIdentityModel]     `tfsdk:"provider_identity"`
}

type customLogCrawlerConfigurationModel struct {
	RoleArn types.String `tfsdk:"role_arn"`
}

type customLogProviderIdentityModel struct {
	ExternalID types.String `tfsdk:"external_id"`
	Principal  types.String `tfsdk:"principal"`
}

type customLogSourceModel struct {
	Attributes    fwtypes.ListNestedObjectValueOf[customLogAttributesModel] `tfsdk:"attributes"`
	Provider      fwtypes.ListNestedObjectValueOf[customLogProviderModel]   `tfsdk:"provider"`
	SourceName    types.String                                              `tfsdk:"source_name"`
	SourceVersion types.String                                              `tfsdk:"source_version"`
}

type customLogAttributesModel struct {
	CrawlerArn  types.String `tfsdk:"crawler_arn"`
	DatabaseArn types.String `tfsdk:"database_arn"`
	TableArn    types.String `tfsdk:"table_arn"`
}

type customLogProviderModel struct {
	RoleArn  types.String `tfsdk:"role_arn"`
	Location types.String `tfsdk:"location"`
}
