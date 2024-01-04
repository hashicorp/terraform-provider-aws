// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package securitylake

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/securitylake"
	awstypes "github.com/aws/aws-sdk-go-v2/service/securitylake/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource(name="AWS Log Source")
func newAWSLogSourceResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &awsLogSourceResource{}

	return r, nil
}

const (
	ResNameAWSLogSource = "AWS Log Source"
)

type awsLogSourceResource struct {
	framework.ResourceWithConfigure
	framework.WithImportByID
}

func (r *awsLogSourceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_securitylake_aws_log_source"
}

func (r *awsLogSourceResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID: framework.IDAttribute(),
		},
		Blocks: map[string]schema.Block{
			"source": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[awsLogSourceSourceModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"accounts": schema.SetAttribute{
							CustomType:  fwtypes.SetOfStringType,
							ElementType: types.StringType,
							Optional:    true,
							Computed:    true,
							PlanModifiers: []planmodifier.Set{
								setplanmodifier.RequiresReplace(),
							},
						},
						"regions": schema.SetAttribute{
							CustomType:  fwtypes.SetOfStringType,
							ElementType: types.StringType,
							Required:    true,
							PlanModifiers: []planmodifier.Set{
								setplanmodifier.RequiresReplace(),
							},
						},
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
					},
				},
			},
		},
	}
}

func (r *awsLogSourceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().SecurityLakeClient(ctx)

	var data awsLogSourceResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := &securitylake.CreateAwsLogSourceInput{}
	resp.Diagnostics.Append(flex.Expand(ctx, data, input)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := conn.CreateAwsLogSource(ctx, input)

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.SecurityLake, create.ErrActionCreating, ResNameAWSLogSource, data.ID.ValueString(), err),
			err.Error(),
		)
		return
	}

	// Set values for unknowns.
	data.ID = flex.StringValueToFramework(ctx, input.Sources[0].SourceName)

	logSource, err := findAWSLogSourceBySourceName(ctx, conn, awstypes.AwsLogSourceName(data.ID.ValueString()))

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.SecurityLake, create.ErrActionReading, ResNameAWSLogSource, data.ID.String(), err),
			err.Error(),
		)
		return
	}

	sourceData, diags := data.Source.ToPtr(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	sourceData.Accounts.SetValue = flex.FlattenFrameworkStringValueSet(ctx, logSource.Accounts)
	sourceData.SourceVersion = flex.StringToFramework(ctx, logSource.SourceVersion)

	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

func (r *awsLogSourceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().SecurityLakeClient(ctx)

	var data awsLogSourceResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	logSource, err := findAWSLogSourceBySourceName(ctx, conn, awstypes.AwsLogSourceName(data.ID.ValueString()))

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

	// We can't use AutoFlEx with the top-level resource model because the API structure uses Go interfaces.
	var sourceData awsLogSourceSourceModel
	resp.Diagnostics.Append(flex.Flatten(ctx, logSource, &sourceData)...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.Source = fwtypes.NewListNestedObjectValueOfPtr(ctx, &sourceData)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *awsLogSourceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// NoOP.
}

func (r *awsLogSourceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().SecurityLakeClient(ctx)

	var data awsLogSourceResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := &securitylake.DeleteAwsLogSourceInput{}
	resp.Diagnostics.Append(flex.Expand(ctx, data, input)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Workaround for acceptance tests deletion.
	if len(input.Sources) == 0 {
		logSource, err := findAWSLogSourceBySourceName(ctx, conn, awstypes.AwsLogSourceName(data.ID.ValueString()))

		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.SecurityLake, create.ErrActionReading, ResNameAWSLogSource, data.ID.String(), err),
				err.Error(),
			)
			return
		}

		input.Sources = tfslices.Of(*logSource)
	}

	_, err := conn.DeleteAwsLogSource(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.SecurityLake, create.ErrActionDeleting, ResNameAWSLogSource, data.ID.String(), err),
			err.Error(),
		)
		return
	}
}

func findAWSLogSourceBySourceName(ctx context.Context, conn *securitylake.Client, sourceName awstypes.AwsLogSourceName) (*awstypes.AwsLogSourceConfiguration, error) {
	input := &securitylake.ListLogSourcesInput{}
	var output *awstypes.AwsLogSourceConfiguration

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
			account, region := aws.ToString(v.Account), aws.ToString(v.Region)
			for _, v := range v.Sources {
				if v, ok := v.(*awstypes.LogSourceResourceMemberAwsLogSource); ok {
					if v := v.Value; v.SourceName == sourceName {
						if output == nil {
							output = &awstypes.AwsLogSourceConfiguration{
								SourceName:    v.SourceName,
								SourceVersion: v.SourceVersion,
							}
						}
						output.Accounts = tfslices.AppendUnique(output.Accounts, account)
						output.Regions = tfslices.AppendUnique(output.Regions, region)
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

type awsLogSourceResourceModel struct {
	ID     types.String                                             `tfsdk:"id"`
	Source fwtypes.ListNestedObjectValueOf[awsLogSourceSourceModel] `tfsdk:"source"`
}

type awsLogSourceSourceModel struct {
	Accounts      fwtypes.SetValueOf[types.String] `tfsdk:"accounts"`
	Regions       fwtypes.SetValueOf[types.String] `tfsdk:"regions"`
	SourceName    types.String                     `tfsdk:"source_name"`
	SourceVersion types.String                     `tfsdk:"source_version"`
}
