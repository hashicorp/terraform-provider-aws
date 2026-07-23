// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package lambdamicrovms

import (
	"context"
	"errors"
	"time"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambdamicrovms"
	awstypes "github.com/aws/aws-sdk-go-v2/service/lambdamicrovms/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	sweepfw "github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for resource registration to the Provider. DO NOT EDIT.
// @FrameworkResource("aws_lambdamicrovms_microvm", name="Microvm")
// @IdentityAttribute("microvm_id")
// @Testing(importStateIdAttribute="microvm_id")
// @Testing(importIgnore="image_identifier;execution_role_arn;logging;run_hook_payload", plannableImportAction="Replace")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/lambdamicrovms;lambdamicrovms.GetMicrovmOutput")
// @Testing(preCheck="testAccPreCheck")
// @Testing(hasNoPreExistingResource=true)
func newMicrovmResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &microvmResource{}
	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameMicrovm = "Microvm"
)

type microvmResource struct {
	framework.ResourceWithModel[microvmResourceModel]
	framework.WithTimeouts
	framework.WithImportByIdentity
}

func (r *microvmResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"egress_network_connectors": schema.ListAttribute{
				CustomType:  fwtypes.ListOfStringType,
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
					listplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrEndpoint: schema.StringAttribute{
				Computed: true,
			},
			names.AttrExecutionRoleARN: schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Optional:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"image_arn": framework.ARNAttributeComputedOnly(),
			"image_identifier": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"image_version": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"ingress_network_connectors": schema.ListAttribute{
				CustomType:  fwtypes.ListOfStringType,
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"maximum_duration_in_seconds": schema.Int32Attribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Int32{
					int32planmodifier.RequiresReplace(),
					int32planmodifier.UseStateForUnknown(),
				},
			},
			"microvm_id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"run_hook_payload": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"started_at": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			names.AttrState: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.MicrovmState](),
				Computed:   true,
			},
		},
		Blocks: map[string]schema.Block{
			"idle_policy": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[idlePolicyModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"auto_resume_enabled": schema.BoolAttribute{
							Required: true,
						},
						"max_idle_duration_seconds": schema.Int32Attribute{
							Required: true,
						},
						"suspended_duration_seconds": schema.Int32Attribute{
							Required: true,
						},
					},
				},
			},
			"logging": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[loggingModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"cloud_watch": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[cloudWatchLoggingModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
								listvalidator.ExactlyOneOf(path.MatchRelative().AtParent().AtName("disabled")),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"log_group": schema.StringAttribute{
										Optional: true,
									},
									"log_stream": schema.StringAttribute{
										Optional: true,
									},
								},
							},
						},
						"disabled": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[loggingDisabledModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
								listvalidator.ExactlyOneOf(path.MatchRelative().AtParent().AtName("cloud_watch")),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{},
							},
						},
					},
				},
			},
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Delete: true,
			}),
		},
	}
}

func (r *microvmResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().LambdaMicrovmsClient(ctx)

	var plan microvmResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	var input lambdamicrovms.RunMicrovmInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &input))
	input.ClientToken = aws.String(create.UniqueId(ctx))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := conn.RunMicrovm(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.ImageIdentifier.String())
		return
	}
	if out == nil {
		smerr.AddError(ctx, &resp.Diagnostics, errors.New("empty output"), smerr.ID, plan.ImageIdentifier.String())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, out, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	outWait, err := waitMicrovmRunning(ctx, conn, plan.MicrovmID.ValueString(), createTimeout)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.MicrovmID.String())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, outWait, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *microvmResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().LambdaMicrovmsClient(ctx)

	var state microvmResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findMicrovmByID(ctx, conn, state.MicrovmID.ValueString())
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.MicrovmID.String())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, out, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

// No Update operation for MicroVMs as all configurations are immutable

func (r *microvmResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().LambdaMicrovmsClient(ctx)

	var state microvmResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	input := lambdamicrovms.TerminateMicrovmInput{
		MicrovmIdentifier: state.MicrovmID.ValueStringPointer(),
	}

	_, err := conn.TerminateMicrovm(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}

		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.MicrovmID.ValueString())
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitMicrovmTerminated(ctx, conn, state.MicrovmID.ValueString(), deleteTimeout)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.MicrovmID.ValueString())
		return
	}
}

func waitMicrovmRunning(ctx context.Context, conn *lambdamicrovms.Client, id string, timeout time.Duration) (*lambdamicrovms.GetMicrovmOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.MicrovmStatePending),
		Target:                    enum.Slice(awstypes.MicrovmStateRunning),
		Refresh:                   statusMicrovm(conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*lambdamicrovms.GetMicrovmOutput); ok {
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitMicrovmTerminated(ctx context.Context, conn *lambdamicrovms.Client, id string, timeout time.Duration) (*lambdamicrovms.GetMicrovmOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.MicrovmStateTerminating, awstypes.MicrovmStateRunning, awstypes.MicrovmStateSuspending, awstypes.MicrovmStateSuspended),
		Target:  []string{},
		Refresh: statusMicrovm(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*lambdamicrovms.GetMicrovmOutput); ok {
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func statusMicrovm(conn *lambdamicrovms.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		out, err := findMicrovmByID(ctx, conn, id)
		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", smarterr.NewError(err)
		}

		return out, string(out.State), nil
	}
}

func findMicrovmByID(ctx context.Context, conn *lambdamicrovms.Client, id string) (*lambdamicrovms.GetMicrovmOutput, error) {
	input := lambdamicrovms.GetMicrovmInput{
		MicrovmIdentifier: aws.String(id),
	}

	out, err := conn.GetMicrovm(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, smarterr.NewError(&retry.NotFoundError{
				LastError: err,
			})
		}

		return nil, smarterr.NewError(err)
	}

	if out == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	// A terminated MicroVM remains queryable via GetMicrovm for some time, but
	// is logically gone. Treat it as not found so Read, the delete waiter, and
	// destroy checks all agree.
	if out.State == awstypes.MicrovmStateTerminated {
		return nil, smarterr.NewError(&retry.NotFoundError{
			LastError: errors.New("MicroVM is terminated"),
		})
	}

	return out, nil
}

type microvmResourceModel struct {
	framework.WithRegionModel
	EgressNetworkConnectors  fwtypes.ListOfString                             `tfsdk:"egress_network_connectors"`
	Endpoint                 types.String                                     `tfsdk:"endpoint"`
	ExecutionRoleARN         fwtypes.ARN                                      `tfsdk:"execution_role_arn"`
	IdlePolicy               fwtypes.ListNestedObjectValueOf[idlePolicyModel] `tfsdk:"idle_policy"`
	ImageARN                 types.String                                     `tfsdk:"image_arn"`
	ImageIdentifier          types.String                                     `tfsdk:"image_identifier"`
	ImageVersion             types.String                                     `tfsdk:"image_version"`
	IngressNetworkConnectors fwtypes.ListOfString                             `tfsdk:"ingress_network_connectors"`
	Logging                  fwtypes.ListNestedObjectValueOf[loggingModel]    `tfsdk:"logging"`
	MaximumDurationInSeconds types.Int32                                      `tfsdk:"maximum_duration_in_seconds"`
	MicrovmID                types.String                                     `tfsdk:"microvm_id"`
	RunHookPayload           types.String                                     `tfsdk:"run_hook_payload"`
	StartedAt                timetypes.RFC3339                                `tfsdk:"started_at"`
	State                    fwtypes.StringEnum[awstypes.MicrovmState]        `tfsdk:"state"`
	Timeouts                 timeouts.Value                                   `tfsdk:"timeouts"`
}

type idlePolicyModel struct {
	AutoResumeEnabled        types.Bool  `tfsdk:"auto_resume_enabled"`
	MaxIdleDurationSeconds   types.Int32 `tfsdk:"max_idle_duration_seconds"`
	SuspendedDurationSeconds types.Int32 `tfsdk:"suspended_duration_seconds"`
}

type loggingModel struct {
	CloudWatch fwtypes.ListNestedObjectValueOf[cloudWatchLoggingModel] `tfsdk:"cloud_watch"`
	Disabled   fwtypes.ListNestedObjectValueOf[loggingDisabledModel]   `tfsdk:"disabled"`
}

type cloudWatchLoggingModel struct {
	LogGroup  types.String `tfsdk:"log_group"`
	LogStream types.String `tfsdk:"log_stream"`
}

type loggingDisabledModel struct{}

var _ flex.Expander = loggingModel{}

func (m loggingModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics

	switch {
	case !m.CloudWatch.IsNull():
		data, d := m.CloudWatch.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		var r awstypes.LoggingMemberCloudWatch
		diags.Append(flex.Expand(ctx, data, &r.Value)...)
		if diags.HasError() {
			return nil, diags
		}

		return &r, diags
	case !m.Disabled.IsNull():
		return &awstypes.LoggingMemberDisabled{
			Value: awstypes.LoggingDisabled{},
		}, diags
	}

	return nil, diags
}

func sweepMicrovms(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	input := lambdamicrovms.ListMicrovmsInput{}
	conn := client.LambdaMicrovmsClient(ctx)
	var sweepResources []sweep.Sweepable

	pages := lambdamicrovms.NewListMicrovmsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, smarterr.NewError(err)
		}

		for _, v := range page.Items {
			// Terminated MicroVMs remain listable until they age out; skip them
			// so the sweeper does not churn on resources that are already gone.
			if v.State == awstypes.MicrovmStateTerminated {
				continue
			}

			sweepResources = append(sweepResources, sweepfw.NewSweepResource(newMicrovmResource, client,
				sweepfw.NewAttribute("microvm_id", aws.ToString(v.MicrovmId))),
			)
		}
	}

	return sweepResources, nil
}
