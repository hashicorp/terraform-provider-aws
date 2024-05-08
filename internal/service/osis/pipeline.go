// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package osis

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/osis"
	awstypes "github.com/aws/aws-sdk-go-v2/service/osis/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
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

// @FrameworkResource(name="Pipeline")
// @Tags(identifierAttribute="pipeline_arn")
func newPipelineResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &pipelineResource{}

	r.SetDefaultCreateTimeout(45 * time.Minute)
	r.SetDefaultUpdateTimeout(45 * time.Minute)
	r.SetDefaultDeleteTimeout(45 * time.Minute)

	return r, nil
}

type pipelineResource struct {
	framework.ResourceWithConfigure
	framework.WithImportByID
	framework.WithTimeouts
}

func (r *pipelineResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_osis_pipeline"
}

func (r *pipelineResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID: framework.IDAttribute(),
			"ingest_endpoint_urls": schema.SetAttribute{
				CustomType:  fwtypes.SetOfStringType,
				Computed:    true,
				ElementType: types.StringType,
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
			},
			"max_units": schema.Int64Attribute{
				Required: true,
				Validators: []validator.Int64{
					int64validator.AtLeast(1),
				},
			},
			"min_units": schema.Int64Attribute{
				Required: true,
				Validators: []validator.Int64{
					int64validator.AtLeast(1),
				},
			},
			"pipeline_arn": framework.ARNAttributeComputedOnly(),
			"pipeline_configuration_body": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 24000),
				},
			},
			"pipeline_name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(3, 28),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"buffer_options": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[bufferOptionsModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"persistent_buffer_enabled": schema.BoolAttribute{
							Required: true,
						},
					},
				},
			},
			"encryption_at_rest_options": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[encryptionAtRestOptionsModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrKMSKeyARN: schema.StringAttribute{
							CustomType: fwtypes.ARNType,
							Required:   true,
						},
					},
				},
			},
			"log_publishing_options": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[logPublishingOptionsModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"is_logging_enabled": schema.BoolAttribute{
							Optional: true,
						},
					},
					Blocks: map[string]schema.Block{
						"cloudwatch_log_destination": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[cloudWatchLogDestinationModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"log_group": schema.StringAttribute{
										Required: true,
										Validators: []validator.String{
											stringvalidator.LengthBetween(1, 512),
										},
									},
								},
							},
						},
					},
				},
			},
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
			"vpc_options": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[vpcOptionsModel](ctx),
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrSecurityGroupIDs: schema.SetAttribute{
							CustomType:  fwtypes.SetOfStringType,
							Optional:    true,
							ElementType: types.StringType,
							PlanModifiers: []planmodifier.Set{
								setplanmodifier.RequiresReplace(),
							},
							Validators: []validator.Set{
								setvalidator.SizeBetween(1, 12),
							},
						},
						names.AttrSubnetIDs: schema.SetAttribute{
							CustomType:  fwtypes.SetOfStringType,
							Required:    true,
							ElementType: types.StringType,
							PlanModifiers: []planmodifier.Set{
								setplanmodifier.RequiresReplace(),
							},
							Validators: []validator.Set{
								setvalidator.SizeBetween(1, 12),
							},
						},
					},
				},
			},
		},
	}
}

func (r *pipelineResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data pipelineResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().OpenSearchIngestionClient(ctx)

	name := data.PipelineName.ValueString()
	input := &osis.CreatePipelineInput{}
	response.Diagnostics.Append(fwflex.Expand(ctx, data, input)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.Tags = getTagsIn(ctx)

	// Retry for IAM eventual consistency.
	_, err := tfresource.RetryWhenIsA[*awstypes.ValidationException](ctx, propagationTimeout, func() (interface{}, error) {
		return conn.CreatePipeline(ctx, input)
	})

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating OpenSearch Ingestion Pipeline (%s)", name), err.Error())

		return
	}

	data.setID()

	pipeline, err := waitPipelineCreated(ctx, conn, name, r.CreateTimeout(ctx, data.Timeouts))

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for OpenSearch Ingestion Pipeline (%s) create", name), err.Error())

		return
	}

	// Set values for unknowns.
	data.IngestEndpointUrls.SetValue = fwflex.FlattenFrameworkStringValueSet(ctx, pipeline.IngestEndpointUrls)
	data.PipelineARN = fwflex.StringToFramework(ctx, pipeline.PipelineArn)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *pipelineResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data pipelineResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	if err := data.InitFromID(); err != nil {
		response.Diagnostics.AddError("parsing resource ID", err.Error())

		return
	}

	conn := r.Meta().OpenSearchIngestionClient(ctx)

	name := data.PipelineName.ValueString()
	pipeline, err := findPipelineByName(ctx, conn, name)

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading OpenSearch Ingestion Pipeline (%s)", name), err.Error())

		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, pipeline, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *pipelineResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var old, new pipelineResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().OpenSearchIngestionClient(ctx)

	if !new.BufferOptions.Equal(old.BufferOptions) ||
		!new.EncryptionAtRestOptions.Equal(old.EncryptionAtRestOptions) ||
		!new.LogPublishingOptions.Equal(old.LogPublishingOptions) ||
		!new.MaxUnits.Equal(old.MaxUnits) ||
		!new.MinUnits.Equal(old.MinUnits) ||
		!new.PipelineConfigurationBody.Equal(old.PipelineConfigurationBody) {
		input := &osis.UpdatePipelineInput{}
		response.Diagnostics.Append(fwflex.Expand(ctx, new, input)...)
		if response.Diagnostics.HasError() {
			return
		}

		name := new.PipelineName.ValueString()
		_, err := conn.UpdatePipeline(ctx, input)

		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("updating OpenSearch Ingestion Pipeline (%s)", name), err.Error())

			return
		}

		if _, err := waitPipelineUpdated(ctx, conn, name, r.UpdateTimeout(ctx, new.Timeouts)); err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("waiting for OpenSearch Ingestion Pipeline (%s) update", name), err.Error())

			return
		}
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *pipelineResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data pipelineResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().OpenSearchIngestionClient(ctx)

	name := data.PipelineName.ValueString()
	input := &osis.DeletePipelineInput{
		PipelineName: aws.String(name),
	}

	_, err := conn.DeletePipeline(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting OpenSearch Ingestion Pipeline (%s)", name), err.Error())

		return
	}

	if _, err := waitPipelineDeleted(ctx, conn, name, r.DeleteTimeout(ctx, data.Timeouts)); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for OpenSearch Ingestion Pipeline (%s) delete", name), err.Error())

		return
	}
}

func (r *pipelineResource) ModifyPlan(ctx context.Context, request resource.ModifyPlanRequest, response *resource.ModifyPlanResponse) {
	r.SetTagsAll(ctx, request, response)
}

func findPipelineByName(ctx context.Context, conn *osis.Client, name string) (*awstypes.Pipeline, error) {
	input := &osis.GetPipelineInput{
		PipelineName: aws.String(name),
	}

	output, err := conn.GetPipeline(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Pipeline == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Pipeline, nil
}

func statusPipeline(ctx context.Context, conn *osis.Client, name string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findPipelineByName(ctx, conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func waitPipelineCreated(ctx context.Context, conn *osis.Client, name string, timeout time.Duration) (*awstypes.Pipeline, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(awstypes.PipelineStatusCreating, awstypes.PipelineStatusStarting),
		Target:     enum.Slice(awstypes.PipelineStatusActive),
		Refresh:    statusPipeline(ctx, conn, name),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Pipeline); ok {
		if reason := output.StatusReason; reason != nil {
			tfresource.SetLastError(err, errors.New(aws.ToString(reason.Description)))
		}

		return output, err
	}

	return nil, err
}

func waitPipelineUpdated(ctx context.Context, conn *osis.Client, name string, timeout time.Duration) (*awstypes.Pipeline, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(awstypes.PipelineStatusUpdating),
		Target:     enum.Slice(awstypes.PipelineStatusActive),
		Refresh:    statusPipeline(ctx, conn, name),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Pipeline); ok {
		if reason := output.StatusReason; reason != nil {
			tfresource.SetLastError(err, errors.New(aws.ToString(reason.Description)))
		}

		return output, err
	}

	return nil, err
}

func waitPipelineDeleted(ctx context.Context, conn *osis.Client, name string, timeout time.Duration) (*awstypes.Pipeline, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(awstypes.PipelineStatusDeleting),
		Target:     []string{},
		Refresh:    statusPipeline(ctx, conn, name),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Pipeline); ok {
		if reason := output.StatusReason; reason != nil {
			tfresource.SetLastError(err, errors.New(aws.ToString(reason.Description)))
		}

		return output, err
	}

	return nil, err
}

type pipelineResourceModel struct {
	BufferOptions             fwtypes.ListNestedObjectValueOf[bufferOptionsModel]           `tfsdk:"buffer_options"`
	EncryptionAtRestOptions   fwtypes.ListNestedObjectValueOf[encryptionAtRestOptionsModel] `tfsdk:"encryption_at_rest_options"`
	ID                        types.String                                                  `tfsdk:"id"`
	IngestEndpointUrls        fwtypes.SetValueOf[types.String]                              `tfsdk:"ingest_endpoint_urls"`
	LogPublishingOptions      fwtypes.ListNestedObjectValueOf[logPublishingOptionsModel]    `tfsdk:"log_publishing_options"`
	MaxUnits                  types.Int64                                                   `tfsdk:"max_units"`
	MinUnits                  types.Int64                                                   `tfsdk:"min_units"`
	PipelineARN               types.String                                                  `tfsdk:"pipeline_arn"`
	PipelineConfigurationBody types.String                                                  `tfsdk:"pipeline_configuration_body"`
	PipelineName              types.String                                                  `tfsdk:"pipeline_name"`
	Tags                      types.Map                                                     `tfsdk:"tags"`
	TagsAll                   types.Map                                                     `tfsdk:"tags_all"`
	Timeouts                  timeouts.Value                                                `tfsdk:"timeouts"`
	VPCOptions                fwtypes.ListNestedObjectValueOf[vpcOptionsModel]              `tfsdk:"vpc_options"`
}

func (data *pipelineResourceModel) InitFromID() error {
	data.PipelineName = data.ID

	return nil
}

func (data *pipelineResourceModel) setID() {
	data.ID = data.PipelineName
}

type bufferOptionsModel struct {
	PersistentBufferEnabled types.Bool `tfsdk:"persistent_buffer_enabled"`
}

type encryptionAtRestOptionsModel struct {
	KmsKeyArn fwtypes.ARN `tfsdk:"kms_key_arn"`
}

type logPublishingOptionsModel struct {
	CloudWatchLogDestination fwtypes.ListNestedObjectValueOf[cloudWatchLogDestinationModel] `tfsdk:"cloudwatch_log_destination"`
	IsLoggingEnabled         types.Bool                                                     `tfsdk:"is_logging_enabled"`
}

type cloudWatchLogDestinationModel struct {
	LogGroup types.String `tfsdk:"log_group"`
}

type vpcOptionsModel struct {
	SecurityGroupIDs fwtypes.SetValueOf[types.String] `tfsdk:"security_group_ids"`
	SubnetIDs        fwtypes.SetValueOf[types.String] `tfsdk:"subnet_ids"`
}
