// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package sagemaker

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sagemaker"
	awstypes "github.com/aws/aws-sdk-go-v2/service/sagemaker/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
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
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_sagemaker_model_card_export_job", name="Model Card Export Job")
func newModelCardExportJobResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &modelCardExportJobResource{}
	r.SetDefaultCreateTimeout(15 * time.Minute)
	return r, nil
}

type modelCardExportJobResource struct {
	framework.ResourceWithModel[modelCardExportJobResourceModel]
	framework.WithNoUpdate
	framework.WithNoOpDelete
	framework.WithTimeouts
}

func (r *modelCardExportJobResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"export_artifacts":          framework.ResourceComputedListOfObjectsAttribute[modelCardExportArtifactsModel](ctx),
			"model_card_export_job_arn": framework.ARNAttributeComputedOnly(),
			"model_card_export_job_name": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexache.MustCompile(`^[a-zA-Z0-9](-*[a-zA-Z0-9]){0,62}$`), ""),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"model_card_name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"model_card_version": schema.Int32Attribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Int32{
					int32planmodifier.RequiresReplace(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"output_config": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[modelCardExportOutputConfigModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtLeast(1),
					listvalidator.SizeAtMost(1),
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"s3_output_path": schema.StringAttribute{
							Required: true,
							Validators: []validator.String{
								stringvalidator.RegexMatches(httpsOrS3URIRegexp, "must be HTTPS or Amazon S3 URI"),
								stringvalidator.LengthBetween(0, 1024),
							},
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
					},
				},
			},
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
			}),
		},
	}
}

func (r *modelCardExportJobResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data modelCardExportJobResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().SageMakerClient(ctx)

	name := fwflex.StringValueFromFramework(ctx, data.ModelCardExportJobName)
	var input sagemaker.CreateModelCardExportJobInput
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	outputCMCEJ, err := conn.CreateModelCardExportJob(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating SageMaker AI Model Card Export Job (%s)", name), err.Error())
		return
	}

	arn := aws.ToString(outputCMCEJ.ModelCardExportJobArn)
	outputDMCEJ, err := waitModelCardExportJobCompleted(ctx, conn, arn, r.CreateTimeout(ctx, data.Timeouts))

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for SageMaker AI Model Card Export Job (%s) complete", arn), err.Error())
	}

	// Set values for unknowns.
	response.Diagnostics.Append(fwflex.Flatten(ctx, outputDMCEJ, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *modelCardExportJobResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data modelCardExportJobResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().SageMakerClient(ctx)

	arn := fwflex.StringValueFromFramework(ctx, data.ModelCardExportJobARN)
	output, err := findModelCardExportJobByARN(ctx, conn, arn)

	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading SageMaker AI Model Card Export Job (%s)", arn), err.Error())
		return
	}

	// Set attributes for import.
	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *modelCardExportJobResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("model_card_export_job_arn"), request, response)
}

func findModelCardExportJobByARN(ctx context.Context, conn *sagemaker.Client, arn string) (*sagemaker.DescribeModelCardExportJobOutput, error) {
	input := sagemaker.DescribeModelCardExportJobInput{
		ModelCardExportJobArn: aws.String(arn),
	}

	return findModelCardExportJob(ctx, conn, &input)
}

func findModelCardExportJob(ctx context.Context, conn *sagemaker.Client, input *sagemaker.DescribeModelCardExportJobInput) (*sagemaker.DescribeModelCardExportJobOutput, error) {
	output, err := conn.DescribeModelCardExportJob(ctx, input)

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

	return output, nil
}

func statusModelCardExportJob(conn *sagemaker.Client, name string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findModelCardExportJobByARN(ctx, conn, name)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func waitModelCardExportJobCompleted(ctx context.Context, conn *sagemaker.Client, arn string, timeout time.Duration) (*sagemaker.DescribeModelCardExportJobOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ModelCardExportJobStatusInProgress),
		Target:  enum.Slice(awstypes.ModelCardExportJobStatusCompleted),
		Refresh: statusModelCardExportJob(conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*sagemaker.DescribeModelCardExportJobOutput); ok {
		retry.SetLastError(err, errors.New(aws.ToString(output.FailureReason)))
		return output, err
	}

	return nil, err
}

type modelCardExportJobResourceModel struct {
	framework.WithRegionModel
	ExportArtifacts        fwtypes.ListNestedObjectValueOf[modelCardExportArtifactsModel]    `tfsdk:"export_artifacts"`
	ModelCardExportJobARN  types.String                                                      `tfsdk:"model_card_export_job_arn"`
	ModelCardExportJobName types.String                                                      `tfsdk:"model_card_export_job_name"`
	ModelCardName          types.String                                                      `tfsdk:"model_card_name"`
	ModelCardVersion       types.Int32                                                       `tfsdk:"model_card_version"`
	OutputConfig           fwtypes.ListNestedObjectValueOf[modelCardExportOutputConfigModel] `tfsdk:"output_config"`
	Timeouts               timeouts.Value                                                    `tfsdk:"timeouts"`
}

type modelCardExportArtifactsModel struct {
	S3ExportArtifacts types.String `tfsdk:"s3_export_artifacts"`
}

type modelCardExportOutputConfigModel struct {
	S3OutputPath types.String `tfsdk:"s3_output_path"`
}
