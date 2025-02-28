// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package medialive

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/medialive"
	awstypes "github.com/aws/aws-sdk-go-v2/service/medialive/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_medialive_multiplex_program", name="Multiplex Program")
func newResourceMultiplexProgram(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := multiplexProgram{}
	r.SetDefaultCreateTimeout(30 * time.Second)

	return &r, nil
}

const (
	ResNameMultiplexProgram = "Multiplex Program"
)

type multiplexProgram struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
	framework.WithImportByID
}

func (m *multiplexProgram) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID: framework.IDAttribute(),
			"multiplex_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"program_name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"multiplex_program_settings": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[multiplexProgramSettings](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"program_number": schema.Int64Attribute{
							Required: true,
						},
						"preferred_channel_pipeline": schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.PreferredChannelPipeline](),
							Required:   true,
						},
					},
					Blocks: map[string]schema.Block{
						"service_descriptor": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[serviceDescriptor](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrProviderName: schema.StringAttribute{
										Required: true,
									},
									names.AttrServiceName: schema.StringAttribute{
										Required: true,
									},
								},
							},
						},
						"video_settings": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[videoSettings](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"constant_bitrate": schema.Int64Attribute{
										Optional: true,
										Computed: true,
										PlanModifiers: []planmodifier.Int64{
											int64planmodifier.UseStateForUnknown(),
										},
									},
								},
								Blocks: map[string]schema.Block{
									"statmux_settings": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[statmuxSettings](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"minimum_bitrate": schema.Int64Attribute{
													Optional: true,
													Computed: true,
													PlanModifiers: []planmodifier.Int64{
														int64planmodifier.UseStateForUnknown(),
													},
												},
												"maximum_bitrate": schema.Int64Attribute{
													Optional: true,
													Computed: true,
													PlanModifiers: []planmodifier.Int64{
														int64planmodifier.UseStateForUnknown(),
													},
												},
												names.AttrPriority: schema.Int64Attribute{
													Optional: true,
													Computed: true,
													PlanModifiers: []planmodifier.Int64{
														int64planmodifier.UseStateForUnknown(),
													},
												},
											},
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
			}),
		},
	}
}

func (m *multiplexProgram) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := m.Meta().MediaLiveClient(ctx)

	var plan resourceMultiplexProgramData
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := medialive.CreateMultiplexProgramInput{
		RequestId: aws.String(id.UniqueId()),
	}
	resp.Diagnostics.Append(fwflex.Expand(ctx, plan, &input)...)
	if resp.Diagnostics.HasError() {
		return
	}

	multiplexId := plan.MultiplexID.ValueString()
	programName := plan.ProgramName.ValueString()

	_, err := conn.CreateMultiplexProgram(ctx, &input)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.MediaLive, create.ErrActionCreating, ResNameMultiplexProgram, plan.ProgramName.String(), nil),
			err.Error(),
		)
		return
	}

	state := plan
	state.ID = fwflex.StringValueToFramework(ctx, fmt.Sprintf("%s/%s", programName, multiplexId))

	createTimeout := m.CreateTimeout(ctx, plan.Timeouts)
	outputRaws, err := tfresource.RetryWhenNotFound(ctx, createTimeout, func() (any, error) {
		return findMultiplexProgramByID(ctx, conn, multiplexId, programName)
	})
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.MediaLive, create.ErrActionCreating, ResNameMultiplexProgram, plan.ProgramName.String(), err),
			err.Error(),
		)
		return
	}

	output := outputRaws.(*medialive.DescribeMultiplexProgramOutput)
	resp.Diagnostics.Append(fwflex.Flatten(ctx, output, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (m *multiplexProgram) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := m.Meta().MediaLiveClient(ctx)

	var state resourceMultiplexProgramData
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	programName, multiplexId, err := parseMultiplexProgramID(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.MediaLive, create.ErrActionReading, ResNameMultiplexProgram, state.ProgramName.String(), err),
			err.Error(),
		)
		return
	}

	out, err := findMultiplexProgramByID(ctx, conn, multiplexId, programName)

	if tfresource.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.MediaLive, create.ErrActionReading, ResNameMultiplexProgram, state.ProgramName.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(fwflex.Flatten(ctx, out, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (m *multiplexProgram) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := m.Meta().MediaLiveClient(ctx)

	var plan, state resourceMultiplexProgramData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	programName, multiplexId, err := parseMultiplexProgramID(plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.MediaLive, create.ErrActionReading, ResNameMultiplexProgram, plan.ProgramName.String(), err),
			err.Error(),
		)
		return
	}

	diff, d := fwflex.Diff(ctx, plan, state)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		input := medialive.UpdateMultiplexProgramInput{}
		resp.Diagnostics.Append(fwflex.Expand(ctx, plan, &input)...)
		if resp.Diagnostics.HasError() {
			return
		}
		input.ProgramName = aws.String(programName)
		input.MultiplexId = aws.String(multiplexId)

		_, err = conn.UpdateMultiplexProgram(ctx, &input)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.MediaLive, create.ErrActionUpdating, ResNameMultiplexProgram, plan.ProgramName.String(), err),
				err.Error(),
			)
			return
		}

		//Need to find multiplex program because output from update does not provide state data
		output, err := findMultiplexProgramByID(ctx, conn, multiplexId, programName)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.MediaLive, create.ErrActionUpdating, ResNameMultiplexProgram, plan.ProgramName.String(), err),
				err.Error(),
			)
			return
		}

		resp.Diagnostics.Append(fwflex.Flatten(ctx, output, &plan)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (m *multiplexProgram) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := m.Meta().MediaLiveClient(ctx)

	var state resourceMultiplexProgramData
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	programName, multiplexId, err := parseMultiplexProgramID(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.MediaLive, create.ErrActionDeleting, ResNameMultiplexProgram, state.ProgramName.String(), err),
			err.Error(),
		)
		return
	}

	input := medialive.DeleteMultiplexProgramInput{
		MultiplexId: aws.String(multiplexId),
		ProgramName: aws.String(programName),
	}
	_, err = conn.DeleteMultiplexProgram(ctx, &input)

	if errs.IsA[*awstypes.NotFoundException](err) {
		return
	}

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.MediaLive, create.ErrActionDeleting, ResNameMultiplexProgram, state.ProgramName.String(), err),
			err.Error(),
		)
		return
	}
}

func findMultiplexProgramByID(ctx context.Context, conn *medialive.Client, multiplexId, programName string) (*medialive.DescribeMultiplexProgramOutput, error) {
	input := medialive.DescribeMultiplexProgramInput{
		MultiplexId: aws.String(multiplexId),
		ProgramName: aws.String(programName),
	}
	out, err := conn.DescribeMultiplexProgram(ctx, &input)

	if errs.IsA[*awstypes.NotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if out == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return out, nil
}

func parseMultiplexProgramID(id string) (programName string, multiplexId string, err error) {
	idParts := strings.Split(id, "/")

	if len(idParts) < 2 || (idParts[0] == "" || idParts[1] == "") {
		err = errors.New("invalid id")
		return programName, multiplexId, err
	}

	programName = idParts[0]
	multiplexId = idParts[1]

	return programName, multiplexId, err
}

type resourceMultiplexProgramData struct {
	ID                       types.String                                              `tfsdk:"id"`
	MultiplexID              types.String                                              `tfsdk:"multiplex_id"`
	MultiplexProgramSettings fwtypes.ListNestedObjectValueOf[multiplexProgramSettings] `tfsdk:"multiplex_program_settings"`
	ProgramName              types.String                                              `tfsdk:"program_name"`
	Timeouts                 timeouts.Value                                            `tfsdk:"timeouts"`
}

type multiplexProgramSettings struct {
	ProgramNumber            types.Int64                                           `tfsdk:"program_number"`
	PreferredChannelPipeline fwtypes.StringEnum[awstypes.PreferredChannelPipeline] `tfsdk:"preferred_channel_pipeline"`
	ServiceDescriptor        fwtypes.ListNestedObjectValueOf[serviceDescriptor]    `tfsdk:"service_descriptor"`
	VideoSettings            fwtypes.ListNestedObjectValueOf[videoSettings]        `tfsdk:"video_settings"`
}

type serviceDescriptor struct {
	ProviderName types.String `tfsdk:"provider_name"`
	ServiceName  types.String `tfsdk:"service_name"`
}

type videoSettings struct {
	ConstantBitrate types.Int64                                      `tfsdk:"constant_bitrate"`
	StatmuxSettings fwtypes.ListNestedObjectValueOf[statmuxSettings] `tfsdk:"statmux_settings"`
}

type statmuxSettings struct {
	MaximumBitrate types.Int64 `tfsdk:"maximum_bitrate"`
	MinimumBitrate types.Int64 `tfsdk:"minimum_bitrate"`
	Priority       types.Int64 `tfsdk:"priority"`
}
