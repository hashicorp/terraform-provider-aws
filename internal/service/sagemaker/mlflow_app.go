// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package sagemaker

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sagemaker"
	awstypes "github.com/aws/aws-sdk-go-v2/service/sagemaker/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_sagemaker_mlflow_app", name="Mlflow App")
// @Tags(identifierAttribute="arn")
func resourceMlflowApp(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &mlflowAppResource{}, nil
}

type mlflowAppResource struct {
	framework.ResourceWithModel[mlflowAppResourceModel]
}

func (r *mlflowAppResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"account_default_status": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.AccountDefaultStatus](),
				Optional:   true,
			},
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"artifact_store_uri": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"default_domain_id_list": schema.SetAttribute{
				CustomType:  fwtypes.SetOfStringType,
				ElementType: types.StringType,
				Optional:    true,
			},
			"model_registration_mode": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.ModelRegistrationMode](),
				Optional:   true,
				Computed:   true,
				Default:    stringdefault.StaticString("AutoModelRegistrationDisabled"),
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrRoleARN: schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			"weekly_maintenance_window_start": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *mlflowAppResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data mlflowAppResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().SageMakerClient(ctx)

	var input sagemaker.CreateMlflowAppInput
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	input.Tags = getTagsIn(ctx)

	output, err := conn.CreateMlflowApp(ctx, &input)
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating SageMaker Mlflow App (%s)", data.Name.ValueString()), err.Error())
		return
	}

	data.ARN = fwflex.StringToFramework(ctx, output.Arn)
	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	if _, err := waitMlflowAppCreated(ctx, conn, data.ARN.ValueString()); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for SageMaker Mlflow App (%s) create", data.ARN.ValueString()), err.Error())
		return
	}

	// Read the resource to get all computed values
	output2, err := findMlflowAppByARN(ctx, conn, data.ARN.ValueString())
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading SageMaker Mlflow App (%s) after create", data.ARN.ValueString()), err.Error())
		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output2, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *mlflowAppResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data mlflowAppResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().SageMakerClient(ctx)

	output, err := findMlflowAppByARN(ctx, conn, data.ARN.ValueString())
	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading SageMaker Mlflow App (%s)", data.ARN.ValueString()), err.Error())
		return
	}

	if output.Status == "Deleted" {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(fmt.Errorf("status: %s", output.Status)))
		response.State.RemoveResource(ctx)
		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *mlflowAppResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state, plan mlflowAppResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().SageMakerClient(ctx)

	if !plan.AccountDefaultStatus.Equal(state.AccountDefaultStatus) ||
		!plan.ArtifactStoreUri.Equal(state.ArtifactStoreUri) ||
		!plan.DefaultDomainIdList.Equal(state.DefaultDomainIdList) ||
		!plan.ModelRegistrationMode.Equal(state.ModelRegistrationMode) ||
		!plan.WeeklyMaintenanceWindowStart.Equal(state.WeeklyMaintenanceWindowStart) {
		var input sagemaker.UpdateMlflowAppInput
		response.Diagnostics.Append(fwflex.Expand(ctx, plan, &input)...)
		if response.Diagnostics.HasError() {
			return
		}

		_, err := conn.UpdateMlflowApp(ctx, &input)
		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("updating SageMaker Mlflow App (%s)", plan.ARN.ValueString()), err.Error())
			return
		}

		if _, err := waitMlflowAppUpdated(ctx, conn, plan.ARN.ValueString()); err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("waiting for SageMaker Mlflow App (%s) update", plan.ARN.ValueString()), err.Error())
			return
		}
	}

	response.Diagnostics.Append(response.State.Set(ctx, &plan)...)
}

func (r *mlflowAppResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data mlflowAppResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().SageMakerClient(ctx)

	_, err := conn.DeleteMlflowApp(ctx, &sagemaker.DeleteMlflowAppInput{
		Arn: data.ARN.ValueStringPointer(),
	})

	if errs.IsA[*awstypes.ResourceNotFound](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting SageMaker Mlflow App (%s)", data.ARN.ValueString()), err.Error())
		return
	}

	if err := waitMlflowAppDeleted(ctx, conn, data.ARN.ValueString()); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for SageMaker Mlflow App (%s) delete", data.ARN.ValueString()), err.Error())
	}
}

func (r *mlflowAppResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrARN), request, response)
}

type mlflowAppResourceModel struct {
	framework.WithRegionModel
	AccountDefaultStatus         fwtypes.StringEnum[awstypes.AccountDefaultStatus]  `tfsdk:"account_default_status"`
	ARN                          types.String                                       `tfsdk:"arn"`
	ArtifactStoreUri             types.String                                       `tfsdk:"artifact_store_uri"`
	DefaultDomainIdList          fwtypes.SetValueOf[types.String]                   `tfsdk:"default_domain_id_list"`
	ModelRegistrationMode        fwtypes.StringEnum[awstypes.ModelRegistrationMode] `tfsdk:"model_registration_mode"`
	Name                         types.String                                       `tfsdk:"name"`
	RoleArn                      fwtypes.ARN                                        `tfsdk:"role_arn"`
	Tags                         tftags.Map                                         `tfsdk:"tags"`
	TagsAll                      tftags.Map                                         `tfsdk:"tags_all"`
	WeeklyMaintenanceWindowStart types.String                                       `tfsdk:"weekly_maintenance_window_start"`
}

func findMlflowAppByARN(ctx context.Context, conn *sagemaker.Client, arn string) (*sagemaker.DescribeMlflowAppOutput, error) {
	input := &sagemaker.DescribeMlflowAppInput{
		Arn: aws.String(arn),
	}

	output, err := conn.DescribeMlflowApp(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFound](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	if output.Status == awstypes.MlflowAppStatusDeleted {
		return nil, &retry.NotFoundError{
			Message: "resource is deleted",
		}
	}

	return output, nil
}
