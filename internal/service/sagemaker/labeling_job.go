// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package sagemaker

import (
	"context"
	"fmt"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sagemaker"
	awstypes "github.com/aws/aws-sdk-go-v2/service/sagemaker/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdkretry "github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
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

// @FrameworkResource("aws_sagemaker_labeling_job", name="Labeling Job")
// @Tags(identifierAttribute="labeling_job_arn")
// @Testing(tagsTest=false)
func newLabelingJobResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &labelingJobResource{}
	return r, nil
}

type labelingJobResource struct {
	framework.ResourceWithModel[labelingJobResourceModel]
}

func (r *labelingJobResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"labeling_job_arn": framework.ARNAttributeComputedOnly(),
			"labeling_job_name": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexache.MustCompile(`^[a-zA-Z0-9](-*[a-zA-Z0-9]){0,62}$`), ""),
				},
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
		},
		Blocks: map[string]schema.Block{
			"stopping_conditions": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[labelingJobStoppingConditionsModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"max_human_labeled_object_count": schema.Int32Attribute{
							Optional: true,
							Validators: []validator.Int32{
								int32validator.AtLeast(1),
							},
							PlanModifiers: []planmodifier.Int32{
								int32planmodifier.RequiresReplace(),
							},
						},
						"max_percentage_of_input_dataset_labeled": schema.Int32Attribute{
							Optional: true,
							Validators: []validator.Int32{
								int32validator.Between(1, 100),
							},
							PlanModifiers: []planmodifier.Int32{
								int32planmodifier.RequiresReplace(),
							},
						},
					},
				},
			},
		},
	}
}

func (r *labelingJobResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data labelingJobResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().SageMakerClient(ctx)

	name := fwflex.StringValueFromFramework(ctx, data.LabelingJobName)
	var input sagemaker.CreateLabelingJobInput
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.Tags = getTagsIn(ctx)

	_, err := conn.CreateLabelingJob(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating SageMaker AI Labeling Job (%s)", name), err.Error())
		return
	}

	// Set values for unknowns.

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *labelingJobResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data labelingJobResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().SageMakerClient(ctx)

	name := fwflex.StringValueFromFramework(ctx, data.LabelingJobName)
	output, err := findLabelingJobByName(ctx, conn, name)

	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading SageMaker AI Labeling Job (%s)", name), err.Error())
		return
	}

	// Set attributes for import.
	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *labelingJobResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data labelingJobResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().SageMakerClient(ctx)

	name := fwflex.StringValueFromFramework(ctx, data.LabelingJobName)
	input := sagemaker.StopLabelingJobInput{
		LabelingJobName: aws.String(name),
	}
	_, err := conn.StopLabelingJob(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFound](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("stopping SageMaker AI Labeling Job (%s)", name), err.Error())
		return
	}
}

func (r *labelingJobResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("labeling_job_name"), request, response)
}

func findLabelingJobByName(ctx context.Context, conn *sagemaker.Client, name string) (*sagemaker.DescribeLabelingJobOutput, error) {
	input := sagemaker.DescribeLabelingJobInput{
		LabelingJobName: aws.String(name),
	}

	output, err := findLabelingJob(ctx, conn, &input)

	if err != nil {
		return nil, err
	}

	if status := output.LabelingJobStatus; status == awstypes.LabelingJobStatusStopping || status == awstypes.LabelingJobStatusStopped {
		return nil, &retry.NotFoundError{
			Message: string(status),
		}
	}

	return output, nil
}

func findLabelingJob(ctx context.Context, conn *sagemaker.Client, input *sagemaker.DescribeLabelingJobInput) (*sagemaker.DescribeLabelingJobOutput, error) {
	output, err := conn.DescribeLabelingJob(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFound](err) {
		return nil, &sdkretry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output, nil
}

type labelingJobResourceModel struct {
	framework.WithRegionModel
	LabelingJobARN     types.String                                                        `tfsdk:"labeling_job_arn"`
	LabelingJobName    types.String                                                        `tfsdk:"labeling_job_name"`
	RoleARN            fwtypes.ARN                                                         `tfsdk:"role_arn"`
	StoppingConditions fwtypes.ListNestedObjectValueOf[labelingJobStoppingConditionsModel] `tfsdk:"stopping_conditions"`
	Tags               tftags.Map                                                          `tfsdk:"tags"`
	TagsAll            tftags.Map                                                          `tfsdk:"tags_all"`
}

type labelingJobStoppingConditionsModel struct {
	MaxHumanLabeledObjectCount         types.Int32 `tfsdk:"max_human_labeled_object_count"`
	MaxPercentageOfInputDatasetLabeled types.Int32 `tfsdk:"max_percentage_of_input_dataset_labeled"`
}
