// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrock

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrock"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrock/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource(name="Provisioned Model Throughput")
// @Tags(identifierAttribute="provisioned_model_arn")
func newProvisionedModelThroughputResource(context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceProvisionedModelThroughput{}

	r.SetDefaultCreateTimeout(10 * time.Minute)

	return r, nil
}

type resourceProvisionedModelThroughput struct {
	framework.ResourceWithConfigure
	framework.WithNoOpUpdate[provisionedModelThroughputResourceModel]
	framework.WithImportByID
	framework.WithTimeouts
}

func (r *resourceProvisionedModelThroughput) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_bedrock_provisioned_model_throughput"
}

func (r *resourceProvisionedModelThroughput) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"commitment_duration": schema.StringAttribute{
				Optional:   true,
				CustomType: fwtypes.StringEnumType[awstypes.CommitmentDuration](),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrID: framework.IDAttribute(),
			"model_arn": schema.StringAttribute{
				Required:   true,
				CustomType: fwtypes.ARNType,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"model_units": schema.Int64Attribute{
				Required: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"provisioned_model_arn": framework.ARNAttributeComputedOnly(),
			"provisioned_model_name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
			}),
		},
	}
}

func (r *resourceProvisionedModelThroughput) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data provisionedModelThroughputResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockClient(ctx)

	input := &bedrock.CreateProvisionedModelThroughputInput{}
	response.Diagnostics.Append(fwflex.Expand(ctx, data, input)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.ClientRequestToken = aws.String(id.UniqueId())
	input.ModelId = fwflex.StringFromFramework(ctx, data.ModelARN) // Different field name on Create.
	input.Tags = getTagsIn(ctx)

	name := data.ProvisionedModelName.ValueString()
	output, err := conn.CreateProvisionedModelThroughput(ctx, input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating Bedrock Provisioned Model Throughput (%s)", name), err.Error())

		return
	}

	// Set values for unknowns.
	data.ProvisionedModelARN = fwflex.StringToFramework(ctx, output.ProvisionedModelArn)
	data.setID()

	if _, err := waitProvisionedModelThroughputCreated(ctx, conn, data.ID.ValueString(), r.CreateTimeout(ctx, data.Timeouts)); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for Bedrock Provisioned Model Throughput (%s) create", data.ID.ValueString()), err.Error())

		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *resourceProvisionedModelThroughput) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data provisionedModelThroughputResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	if err := data.InitFromID(); err != nil {
		response.Diagnostics.AddError("parsing resource ID", err.Error())

		return
	}

	conn := r.Meta().BedrockClient(ctx)

	output, err := findProvisionedModelThroughputByID(ctx, conn, data.ID.ValueString())

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Bedrock Provisioned Model Throughput (%s)", data.ID.ValueString()), err.Error())

		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *resourceProvisionedModelThroughput) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data provisionedModelThroughputResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockClient(ctx)

	_, err := conn.DeleteProvisionedModelThroughput(ctx, &bedrock.DeleteProvisionedModelThroughputInput{
		ProvisionedModelId: fwflex.StringFromFramework(ctx, data.ID),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting Bedrock Provisioned Model Throughput (%s)", data.ID.ValueString()), err.Error())

		return
	}
}

func findProvisionedModelThroughputByID(ctx context.Context, conn *bedrock.Client, id string) (*bedrock.GetProvisionedModelThroughputOutput, error) {
	input := &bedrock.GetProvisionedModelThroughputInput{
		ProvisionedModelId: aws.String(id),
	}

	output, err := conn.GetProvisionedModelThroughput(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func statusProvisionedModelThroughput(ctx context.Context, conn *bedrock.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findProvisionedModelThroughputByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func waitProvisionedModelThroughputCreated(ctx context.Context, conn *bedrock.Client, id string, timeout time.Duration) (*bedrock.GetProvisionedModelThroughputOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ProvisionedModelStatusCreating),
		Target:  enum.Slice(awstypes.ProvisionedModelStatusInService),
		Refresh: statusProvisionedModelThroughput(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*bedrock.GetProvisionedModelThroughputOutput); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.FailureMessage)))

		return output, err
	}

	return nil, err
}

type provisionedModelThroughputResourceModel struct {
	CommitmentDuration   fwtypes.StringEnum[awstypes.CommitmentDuration] `tfsdk:"commitment_duration"`
	ID                   types.String                                    `tfsdk:"id"`
	ModelARN             fwtypes.ARN                                     `tfsdk:"model_arn"`
	ModelUnits           types.Int64                                     `tfsdk:"model_units"`
	ProvisionedModelARN  types.String                                    `tfsdk:"provisioned_model_arn"`
	ProvisionedModelName types.String                                    `tfsdk:"provisioned_model_name"`
	Tags                 types.Map                                       `tfsdk:"tags"`
	TagsAll              types.Map                                       `tfsdk:"tags_all"`
	Timeouts             timeouts.Value                                  `tfsdk:"timeouts"`
}

func (data *provisionedModelThroughputResourceModel) InitFromID() error {
	data.ProvisionedModelARN = data.ID

	return nil
}

func (data *provisionedModelThroughputResourceModel) setID() {
	data.ID = data.ProvisionedModelARN
}
