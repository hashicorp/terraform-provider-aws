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
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource(name="AWS Log Source")
func newAWSLogSourceResource(context.Context) (resource.ResourceWithConfigure, error) {
	r := &awsLogSourceResource{}

	return r, nil
}

type awsLogSourceResource struct {
	framework.ResourceWithConfigure
	framework.WithNoUpdate
	framework.WithImportByID
}

func (r *awsLogSourceResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_securitylake_aws_log_source"
}

func (r *awsLogSourceResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID: framework.IDAttribute(),
		},
		Blocks: map[string]schema.Block{
			names.AttrSource: schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[awsLogSourceSourceModel](ctx),
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				Validators: []validator.List{
					listvalidator.IsRequired(),
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

func (r *awsLogSourceResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data awsLogSourceResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().SecurityLakeClient(ctx)

	input := &securitylake.CreateAwsLogSourceInput{}
	response.Diagnostics.Append(fwflex.Expand(ctx, data, input)...)
	if response.Diagnostics.HasError() {
		return
	}

	_, err := retryDataLakeConflictWithMutex(ctx, func() (*securitylake.CreateAwsLogSourceOutput, error) {
		return conn.CreateAwsLogSource(ctx, input)
	})

	if err != nil {
		response.Diagnostics.AddError("creating Security Lake AWS Log Source", err.Error())

		return
	}

	// Set values for unknowns.
	data.ID = fwflex.StringValueToFramework(ctx, input.Sources[0].SourceName)

	logSource, err := findAWSLogSourceBySourceName(ctx, conn, awstypes.AwsLogSourceName(data.ID.ValueString()))

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Security Lake AWS Log Source (%s)", data.ID.ValueString()), err.Error())

		return
	}

	sourceData, diags := data.Source.ToPtr(ctx)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	sourceData.Accounts.SetValue = fwflex.FlattenFrameworkStringValueSet(ctx, logSource.Accounts)
	sourceData.SourceVersion = fwflex.StringToFramework(ctx, logSource.SourceVersion)
	data.Source = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, sourceData)

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *awsLogSourceResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data awsLogSourceResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().SecurityLakeClient(ctx)

	logSource, err := findAWSLogSourceBySourceName(ctx, conn, awstypes.AwsLogSourceName(data.ID.ValueString()))

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Security Lake AWS Log Source (%s)", data.ID.ValueString()), err.Error())

		return
	}

	// We can't use AutoFlEx with the top-level resource model because the API structure uses Go interfaces.
	var sourceData awsLogSourceSourceModel
	response.Diagnostics.Append(fwflex.Flatten(ctx, logSource, &sourceData)...)
	if response.Diagnostics.HasError() {
		return
	}

	data.Source = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &sourceData)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *awsLogSourceResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data awsLogSourceResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().SecurityLakeClient(ctx)

	input := &securitylake.DeleteAwsLogSourceInput{}
	response.Diagnostics.Append(fwflex.Expand(ctx, data, input)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Workaround for acceptance tests deletion.
	if len(input.Sources) == 0 {
		logSource, err := findAWSLogSourceBySourceName(ctx, conn, awstypes.AwsLogSourceName(data.ID.ValueString()))

		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("reading Security Lake AWS Log Source (%s)", data.ID.ValueString()), err.Error())

			return
		}

		input.Sources = []awstypes.AwsLogSourceConfiguration{*logSource}
	}

	_, err := retryDataLakeConflictWithMutex(ctx, func() (*securitylake.DeleteAwsLogSourceOutput, error) {
		return conn.DeleteAwsLogSource(ctx, input)
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting Security Lake AWS Log Source (%s)", data.ID.ValueString()), err.Error())

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
