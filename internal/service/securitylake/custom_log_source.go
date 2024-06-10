// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package securitylake

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/securitylake"
	awstypes "github.com/aws/aws-sdk-go-v2/service/securitylake/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource(name="Custom Log Source")
func newCustomLogSourceResource(context.Context) (resource.ResourceWithConfigure, error) {
	r := &customLogSourceResource{}

	return r, nil
}

type customLogSourceResource struct {
	framework.ResourceWithConfigure
	framework.WithNoUpdate
	framework.WithImportByID
}

func (r *customLogSourceResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_securitylake_custom_log_source"
}

func (r *customLogSourceResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrAttributes: schema.ListAttribute{
				CustomType: fwtypes.NewListNestedObjectTypeOf[customLogSourceAttributesModel](ctx),
				Computed:   true,
				ElementType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"crawler_arn":  types.StringType,
						"database_arn": types.StringType,
						"table_arn":    types.StringType,
					},
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"event_classes": schema.SetAttribute{
				CustomType:  fwtypes.SetOfStringType,
				Optional:    true,
				ElementType: types.StringType,
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.RequiresReplace(),
				},
			},
			names.AttrID: framework.IDAttribute(),
			"provider_details": schema.ListAttribute{
				CustomType: fwtypes.NewListNestedObjectTypeOf[customLogSourceProviderModel](ctx),
				Computed:   true,
				ElementType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						names.AttrLocation: types.StringType,
						names.AttrRoleARN:  types.StringType,
					},
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"source_name": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthAtMost(20),
				},
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
		},
		Blocks: map[string]schema.Block{
			names.AttrConfiguration: schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[customLogSourceConfigurationModel](ctx),
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtMost(1),
					listvalidator.SizeAtLeast(1),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"crawler_configuration": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[customLogSourceCrawlerConfigurationModel](ctx),
							PlanModifiers: []planmodifier.List{
								listplanmodifier.RequiresReplace(),
							},
							Validators: []validator.List{
								listvalidator.IsRequired(),
								listvalidator.SizeAtMost(1),
								listvalidator.SizeAtLeast(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrRoleARN: schema.StringAttribute{
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
							CustomType: fwtypes.NewListNestedObjectTypeOf[customLogSourceProviderIdentityModel](ctx),
							PlanModifiers: []planmodifier.List{
								listplanmodifier.RequiresReplace(),
							},
							Validators: []validator.List{
								listvalidator.IsRequired(),
								listvalidator.SizeAtMost(1),
								listvalidator.SizeAtLeast(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrExternalID: schema.StringAttribute{
										Required: true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplace(),
										},
									},
									names.AttrPrincipal: schema.StringAttribute{
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

func (r *customLogSourceResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data customLogSourceResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().SecurityLakeClient(ctx)

	input := &securitylake.CreateCustomLogSourceInput{}
	response.Diagnostics.Append(fwflex.Expand(ctx, data, input)...)
	if response.Diagnostics.HasError() {
		return
	}

	output, err := retryDataLakeConflictWithMutex(ctx, func() (*securitylake.CreateCustomLogSourceOutput, error) {
		return conn.CreateCustomLogSource(ctx, input)
	})

	if err != nil {
		response.Diagnostics.AddError("creating Security Lake Custom Log Source", err.Error())

		return
	}

	var dataFromCreate customLogSourceSourceModel
	response.Diagnostics.Append(fwflex.Flatten(ctx, output.Source, &dataFromCreate)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Set values for unknowns.
	data.Attributes = dataFromCreate.Attributes
	data.ProviderDetails = dataFromCreate.Provider
	data.SourceVersion = dataFromCreate.SourceVersion
	data.setID()

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *customLogSourceResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data customLogSourceResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	if err := data.InitFromID(); err != nil {
		response.Diagnostics.AddError("parsing resource ID", err.Error())

		return
	}

	conn := r.Meta().SecurityLakeClient(ctx)

	sourceName := data.SourceName.ValueString()
	customLogSource, err := findCustomLogSourceBySourceName(ctx, conn, sourceName)

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Security Lake Custom Log Source (%s)", sourceName), err.Error())

		return
	}

	var dataFromRead customLogSourceSourceModel
	response.Diagnostics.Append(fwflex.Flatten(ctx, customLogSource, &dataFromRead)...)
	if response.Diagnostics.HasError() {
		return
	}

	data.Attributes = dataFromRead.Attributes
	data.ProviderDetails = dataFromRead.Provider
	data.SourceVersion = dataFromRead.SourceVersion

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *customLogSourceResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data customLogSourceResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().SecurityLakeClient(ctx)

	input := &securitylake.DeleteCustomLogSourceInput{}
	response.Diagnostics.Append(fwflex.Expand(ctx, data, input)...)
	if response.Diagnostics.HasError() {
		return
	}

	_, err := retryDataLakeConflictWithMutex(ctx, func() (*securitylake.DeleteCustomLogSourceOutput, error) {
		return conn.DeleteCustomLogSource(ctx, input)
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting Security Lake Custom Log Source (%s)", data.ID.ValueString()), err.Error())

		return
	}
}

func findCustomLogSourceBySourceName(ctx context.Context, conn *securitylake.Client, sourceName string) (*awstypes.CustomLogSourceResource, error) {
	input := &securitylake.ListLogSourcesInput{}

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
						return &v, nil
					}
				}
			}
		}
	}

	return nil, tfresource.NewEmptyResultError(sourceName)
}

type customLogSourceResourceModel struct {
	Attributes      fwtypes.ListNestedObjectValueOf[customLogSourceAttributesModel]    `tfsdk:"attributes"`
	Configuration   fwtypes.ListNestedObjectValueOf[customLogSourceConfigurationModel] `tfsdk:"configuration"`
	EventClasses    fwtypes.SetValueOf[types.String]                                   `tfsdk:"event_classes"`
	ID              types.String                                                       `tfsdk:"id"`
	ProviderDetails fwtypes.ListNestedObjectValueOf[customLogSourceProviderModel]      `tfsdk:"provider_details"`
	SourceName      types.String                                                       `tfsdk:"source_name"`
	SourceVersion   types.String                                                       `tfsdk:"source_version"`
}

func (data *customLogSourceResourceModel) InitFromID() error {
	data.SourceName = data.ID

	return nil
}

func (data *customLogSourceResourceModel) setID() {
	data.ID = data.SourceName
}

type customLogSourceConfigurationModel struct {
	CrawlerConfiguration fwtypes.ListNestedObjectValueOf[customLogSourceCrawlerConfigurationModel] `tfsdk:"crawler_configuration"`
	ProviderIdentity     fwtypes.ListNestedObjectValueOf[customLogSourceProviderIdentityModel]     `tfsdk:"provider_identity"`
}

type customLogSourceCrawlerConfigurationModel struct {
	RoleArn types.String `tfsdk:"role_arn"`
}

type customLogSourceProviderIdentityModel struct {
	ExternalID types.String `tfsdk:"external_id"`
	Principal  types.String `tfsdk:"principal"`
}

type customLogSourceSourceModel struct {
	Attributes    fwtypes.ListNestedObjectValueOf[customLogSourceAttributesModel] `tfsdk:"attributes"`
	Provider      fwtypes.ListNestedObjectValueOf[customLogSourceProviderModel]   `tfsdk:"provider"`
	SourceName    types.String                                                    `tfsdk:"source_name"`
	SourceVersion types.String                                                    `tfsdk:"source_version"`
}

type customLogSourceAttributesModel struct {
	CrawlerARN  types.String `tfsdk:"crawler_arn"`
	DatabaseARN types.String `tfsdk:"database_arn"`
	TableARN    types.String `tfsdk:"table_arn"`
}

type customLogSourceProviderModel struct {
	Location types.String `tfsdk:"location"`
	RoleARN  types.String `tfsdk:"role_arn"`
}
