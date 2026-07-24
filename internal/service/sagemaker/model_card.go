// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package sagemaker

import (
	"context"
	"fmt"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sagemaker"
	awstypes "github.com/aws/aws-sdk-go-v2/service/sagemaker/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
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
	tfjson "github.com/hashicorp/terraform-provider-aws/internal/json"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_sagemaker_model_card", name="Model Card")
// @Tags(identifierAttribute="model_card_arn")
// @Testing(tagsTest=false)
func newModelCardResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &modelCardResource{}
	r.SetDefaultDeleteTimeout(15 * time.Minute)
	return r, nil
}

type modelCardResource struct {
	framework.ResourceWithModel[modelCardResourceModel]
	framework.WithTimeouts
}

func (r *modelCardResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrContent: schema.StringAttribute{
				CustomType: jsontypes.NormalizedType{},
				Required:   true,
			},
			"model_card_arn": framework.ARNAttributeComputedOnly(),
			"model_card_name": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexache.MustCompile(`^[a-zA-Z0-9](-*[a-zA-Z0-9]){0,62}$`), ""),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"model_card_status": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.ModelCardStatus](),
				Required:   true,
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"security_config": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[modelCardSecurityConfigModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrKMSKeyID: schema.StringAttribute{
							CustomType: fwtypes.ARNType,
							Required:   true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
					},
				},
			},
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Delete: true,
			}),
		},
	}
}

func (r *modelCardResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data modelCardResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().SageMakerClient(ctx)

	name := fwflex.StringValueFromFramework(ctx, data.ModelCardName)
	var input sagemaker.CreateModelCardInput
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	content, _ := tfjson.CompactString(fwflex.StringValueFromFramework(ctx, data.Content))
	input.Content = aws.String(content)
	input.Tags = getTagsIn(ctx)

	output, err := conn.CreateModelCard(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating SageMaker AI Model Card (%s)", name), err.Error())
		return
	}

	// Set values for unknowns.
	data.ModelCardARN = fwflex.StringToFramework(ctx, output.ModelCardArn)

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *modelCardResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data modelCardResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().SageMakerClient(ctx)

	name := fwflex.StringValueFromFramework(ctx, data.ModelCardName)
	output, err := findModelCardByName(ctx, conn, name)

	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading SageMaker AI Model Card (%s)", name), err.Error())
		return
	}

	// Set attributes for import.
	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *modelCardResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var old, new modelCardResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().SageMakerClient(ctx)

	diff, d := fwflex.Diff(ctx, new, old)
	response.Diagnostics.Append(d...)
	if response.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		name := fwflex.StringValueFromFramework(ctx, new.ModelCardName)
		var input sagemaker.UpdateModelCardInput
		response.Diagnostics.Append(fwflex.Expand(ctx, new, &input)...)
		if response.Diagnostics.HasError() {
			return
		}

		// Additional fields.
		content, _ := tfjson.CompactString(fwflex.StringValueFromFramework(ctx, new.Content))
		input.Content = aws.String(content)

		_, err := conn.UpdateModelCard(ctx, &input)

		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("updating SageMaker AI Model Card (%s)", name), err.Error())
			return
		}
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *modelCardResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data modelCardResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().SageMakerClient(ctx)

	name := fwflex.StringValueFromFramework(ctx, data.ModelCardName)
	input := sagemaker.DeleteModelCardInput{
		ModelCardName: aws.String(name),
	}

	_, err := conn.DeleteModelCard(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFound](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting SageMaker AI Model Card (%s)", name), err.Error())
		return
	}

	if _, err := waitModelCardDeleted(ctx, conn, name, r.DeleteTimeout(ctx, data.Timeouts)); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for SageMaker AI Model Card (%s) delete", name), err.Error())
	}
}

func (r *modelCardResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("model_card_name"), request, response)
}

func findModelCardByName(ctx context.Context, conn *sagemaker.Client, name string) (*sagemaker.DescribeModelCardOutput, error) {
	input := sagemaker.DescribeModelCardInput{
		ModelCardName: aws.String(name),
	}

	output, err := findModelCard(ctx, conn, &input)

	if err != nil {
		return nil, err
	}

	if status := output.ModelCardProcessingStatus; status == awstypes.ModelCardProcessingStatusDeleteCompleted {
		return nil, &retry.NotFoundError{
			Message: string(status),
		}
	}

	return output, nil
}

func findModelCard(ctx context.Context, conn *sagemaker.Client, input *sagemaker.DescribeModelCardInput) (*sagemaker.DescribeModelCardOutput, error) {
	output, err := conn.DescribeModelCard(ctx, input)

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

func statusModelCard(conn *sagemaker.Client, name string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findModelCardByName(ctx, conn, name)

		if tfawserr.ErrMessageContains(err, ErrCodeValidationException, "Model card is being deleted") {
			return new(sagemaker.DescribeModelCardOutput), string(awstypes.ModelCardProcessingStatusDeleteInprogress), nil
		}

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.ModelCardProcessingStatus), nil
	}
}

func waitModelCardDeleted(ctx context.Context, conn *sagemaker.Client, name string, timeout time.Duration) (*sagemaker.DescribeModelCardOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ModelCardProcessingStatusDeletePending, awstypes.ModelCardProcessingStatusDeleteInprogress, awstypes.ModelCardProcessingStatusContentDeleted, awstypes.ModelCardProcessingStatusExportjobsDeleted),
		Target:  []string{},
		Refresh: statusModelCard(conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*sagemaker.DescribeModelCardOutput); ok {
		return output, err
	}

	return nil, err
}

type modelCardResourceModel struct {
	framework.WithRegionModel
	Content         jsontypes.Normalized                                          `tfsdk:"content" autoflex:",noexpand"`
	ModelCardARN    types.String                                                  `tfsdk:"model_card_arn"`
	ModelCardName   types.String                                                  `tfsdk:"model_card_name"`
	ModelCardStatus fwtypes.StringEnum[awstypes.ModelCardStatus]                  `tfsdk:"model_card_status"`
	SecurityConfig  fwtypes.ListNestedObjectValueOf[modelCardSecurityConfigModel] `tfsdk:"security_config"`
	Tags            tftags.Map                                                    `tfsdk:"tags"`
	TagsAll         tftags.Map                                                    `tfsdk:"tags_all"`
	Timeouts        timeouts.Value                                                `tfsdk:"timeouts"`
}

type modelCardSecurityConfigModel struct {
	KMSKeyID fwtypes.ARN `tfsdk:"kms_key_id"`
}
