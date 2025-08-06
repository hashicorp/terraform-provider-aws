// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrockagentcore

import (
	"context"
	"errors"
	"time"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdkretry "github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for resource registration to the Provider. DO NOT EDIT.
// @FrameworkResource("aws_bedrockagentcore_agent_runtime", name="Agent Runtime")
func newResourceAgentRuntime(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceAgentRuntime{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameAgentRuntime = "Agent Runtime"
)

type resourceAgentRuntime struct {
	framework.ResourceWithModel[resourceAgentRuntimeModel]
	framework.WithTimeouts
	framework.WithImportByID
}

func (r *resourceAgentRuntime) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	serverProtocol := fwtypes.StringEnumType[awstypes.ServerProtocol]()
	networkMode := fwtypes.StringEnumType[awstypes.NetworkMode]()

	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrID:  framework.IDAttribute(),
			names.AttrName: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"client_token": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
			},
			"environment_variables": schema.MapAttribute{
				CustomType: fwtypes.MapOfStringType,
				Optional:   true,
			},
			names.AttrNetworkConfiguration: schema.ObjectAttribute{
				CustomType: fwtypes.NewObjectTypeOf[networkConfigurationModel](ctx),
				Optional:   true,
				Computed:   true,
				Default: objectdefault.StaticValue(
					types.ObjectValueMust(

						map[string]attr.Type{
							"network_mode": networkMode,
						},
						map[string]attr.Value{
							"network_mode": fwtypes.StringEnumValue(awstypes.NetworkModePublic),
						},
					),
				),
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrRoleARN: schema.StringAttribute{
				Required: true,
			},
			"workload_identity_details": schema.ListAttribute{
				CustomType: fwtypes.NewListNestedObjectTypeOf[workloadIdentityDetailsModel](ctx),
				Computed:   true,
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
			},
			names.AttrVersion: schema.StringAttribute{
				Computed: true,
			},
		},

		Blocks: map[string]schema.Block{
			"artifact": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[artifactModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"container_configuration": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[ContainerConfigurationModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"container_uri": schema.StringAttribute{
										Required: true,
									},
								},
							},
						},
					},
				},
			},
			"authorizer_configuration": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[authorizerConfigurationModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"custom_jwt_authorizer": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[CustomJWTAuthorizerConfigurationModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"discovery_url": schema.StringAttribute{
										Required: true,
									},
									"allowed_audience": schema.SetAttribute{
										CustomType:  fwtypes.SetOfStringType,
										ElementType: types.StringType,
										Optional:    true,
									},
									"allowed_clients": schema.SetAttribute{
										CustomType:  fwtypes.SetOfStringType,
										ElementType: types.StringType,
										Optional:    true,
									},
								},
							},
						},
					},
				},
			},
			"protocol_configuration": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[protocolConfigurationModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"server_protocol": schema.StringAttribute{
							Optional:   true,
							CustomType: serverProtocol,
						},
					},
				},
			},
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func (r *resourceAgentRuntime) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().BedrockAgentCoreClient(ctx)

	var plan resourceAgentRuntimeModel
	smerr.EnrichAppend(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	var input bedrockagentcorecontrol.CreateAgentRuntimeInput
	smerr.EnrichAppend(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &input, flex.WithFieldNamePrefix("AgentRuntime")))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := conn.CreateAgentRuntime(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.Name.String())
		return
	}
	if out == nil {
		smerr.AddError(ctx, &resp.Diagnostics, errors.New("empty output"), smerr.ID, plan.Name.String())
		return
	}

	smerr.EnrichAppend(ctx, &resp.Diagnostics, flex.Flatten(ctx, out, &plan, flex.WithFieldNamePrefix("AgentRuntime")))
	if resp.Diagnostics.HasError() {
		return
	}

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	_, err = waitAgentRuntimeCreated(ctx, conn, plan.ID.ValueString(), createTimeout)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.Name.String())
		return
	}

	smerr.EnrichAppend(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *resourceAgentRuntime) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().BedrockAgentCoreClient(ctx)

	var state resourceAgentRuntimeModel
	smerr.EnrichAppend(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findAgentRuntimeByID(ctx, conn, state.ID.ValueString())
	if tfresource.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ID.String())
		return
	}

	smerr.EnrichAppend(ctx, &resp.Diagnostics, flex.Flatten(ctx, out, &state, flex.WithFieldNamePrefix("AgentRuntime")))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.EnrichAppend(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *resourceAgentRuntime) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().BedrockAgentCoreClient(ctx)

	var plan, state resourceAgentRuntimeModel
	smerr.EnrichAppend(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	smerr.EnrichAppend(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	diff, d := flex.Diff(ctx, plan, state)
	smerr.EnrichAppend(ctx, &resp.Diagnostics, d)
	if resp.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		var input bedrockagentcorecontrol.UpdateAgentRuntimeInput
		smerr.EnrichAppend(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &input, flex.WithFieldNamePrefix("AgentRuntime")))
		if resp.Diagnostics.HasError() {
			return
		}

		out, err := conn.UpdateAgentRuntime(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.ID.String())
			return
		}
		if out == nil {
			smerr.AddError(ctx, &resp.Diagnostics, errors.New("empty output"), smerr.ID, plan.ID.String())
			return
		}

		smerr.EnrichAppend(ctx, &resp.Diagnostics, flex.Flatten(ctx, out, &plan, flex.WithFieldNamePrefix("AgentRuntime")))
		if resp.Diagnostics.HasError() {
			return
		}
	}

	updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
	_, err := waitAgentRuntimeUpdated(ctx, conn, plan.ID.ValueString(), updateTimeout)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.ID.String())
		return
	}

	smerr.EnrichAppend(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *resourceAgentRuntime) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().BedrockAgentCoreClient(ctx)

	var state resourceAgentRuntimeModel
	smerr.EnrichAppend(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	input := bedrockagentcorecontrol.DeleteAgentRuntimeInput{
		AgentRuntimeId: state.ID.ValueStringPointer(),
	}

	_, err := conn.DeleteAgentRuntime(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}

		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ID.String())
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitAgentRuntimeDeleted(ctx, conn, state.ID.ValueString(), deleteTimeout)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ID.String())
		return
	}
}

func waitAgentRuntimeCreated(ctx context.Context, conn *bedrockagentcorecontrol.Client, id string, timeout time.Duration) (*bedrockagentcorecontrol.GetAgentRuntimeOutput, error) {
	stateConf := &sdkretry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.AgentStatusCreating),
		Target:                    enum.Slice(awstypes.AgentStatusReady),
		Refresh:                   statusAgentRuntime(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*bedrockagentcorecontrol.GetAgentRuntimeOutput); ok {
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitAgentRuntimeUpdated(ctx context.Context, conn *bedrockagentcorecontrol.Client, id string, timeout time.Duration) (*bedrockagentcorecontrol.GetAgentRuntimeOutput, error) {
	stateConf := &sdkretry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.AgentStatusUpdating),
		Target:                    enum.Slice(awstypes.AgentStatusReady),
		Refresh:                   statusAgentRuntime(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*bedrockagentcorecontrol.GetAgentRuntimeOutput); ok {
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitAgentRuntimeDeleted(ctx context.Context, conn *bedrockagentcorecontrol.Client, id string, timeout time.Duration) (*bedrockagentcorecontrol.GetAgentRuntimeOutput, error) {
	stateConf := &sdkretry.StateChangeConf{
		Pending: enum.Slice(awstypes.AgentStatusDeleting, awstypes.AgentStatusReady),
		Target:  []string{},
		Refresh: statusAgentRuntime(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*bedrockagentcorecontrol.GetAgentRuntimeOutput); ok {
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func statusAgentRuntime(ctx context.Context, conn *bedrockagentcorecontrol.Client, id string) sdkretry.StateRefreshFunc {
	return func() (any, string, error) {
		out, err := findAgentRuntimeByID(ctx, conn, id)
		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", smarterr.NewError(err)
		}

		return out, string(out.Status), nil
	}
}

func findAgentRuntimeByID(ctx context.Context, conn *bedrockagentcorecontrol.Client, id string) (*bedrockagentcorecontrol.GetAgentRuntimeOutput, error) {
	input := bedrockagentcorecontrol.GetAgentRuntimeInput{
		AgentRuntimeId: aws.String(id),
	}

	out, err := conn.GetAgentRuntime(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, smarterr.NewError(&sdkretry.NotFoundError{
				LastError:   err,
				LastRequest: &input,
			})
		}

		return nil, smarterr.NewError(err)
	}

	if out == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError(&input))
	}

	return out, nil
}

type resourceAgentRuntimeModel struct {
	framework.WithRegionModel

	ARN         types.String `tfsdk:"arn"`
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	RoleArn     types.String `tfsdk:"role_arn"`
	ClientToken types.String `tfsdk:"client_token"`
	Version     types.String `tfsdk:"version"`

	EnvironmentVariables    fwtypes.MapOfString                                           `tfsdk:"environment_variables"`
	Artifact                fwtypes.ListNestedObjectValueOf[artifactModel]                `tfsdk:"artifact"`
	AuthorizerConfiguration fwtypes.ListNestedObjectValueOf[authorizerConfigurationModel] `tfsdk:"authorizer_configuration"`
	NetworkConfiguration    fwtypes.ObjectValueOf[networkConfigurationModel]              `tfsdk:"network_configuration"`
	ProtocolConfiguration   fwtypes.ListNestedObjectValueOf[protocolConfigurationModel]   `tfsdk:"protocol_configuration"`
	WorkloadIdentityDetails fwtypes.ListNestedObjectValueOf[workloadIdentityDetailsModel] `tfsdk:"workload_identity_details"`

	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

type artifactModel struct {
	ContainerConfiguration fwtypes.ListNestedObjectValueOf[ContainerConfigurationModel] `tfsdk:"container_configuration"`
}

var (
	_ flex.Expander  = artifactModel{}
	_ flex.Flattener = &artifactModel{}
)

func (m *artifactModel) Flatten(ctx context.Context, v any) (diags diag.Diagnostics) {
	switch t := v.(type) {
	case awstypes.AgentArtifactMemberContainerConfiguration:
		var model ContainerConfigurationModel
		d := flex.Flatten(ctx, t.Value, &model)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}
		m.ContainerConfiguration = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)
		return diags

	default:
		return diags
	}
}

func (m artifactModel) Expand(ctx context.Context) (result any, diags diag.Diagnostics) {
	switch {
	case !m.ContainerConfiguration.IsNull():
		model, d := m.ContainerConfiguration.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		var r awstypes.AgentArtifactMemberContainerConfiguration
		diags.Append(flex.Expand(ctx, model, &r.Value)...)
		if diags.HasError() {
			return nil, diags
		}
		return &r, diags
	}
	return nil, diags
}

type ContainerConfigurationModel struct {
	ContainerUri types.String `tfsdk:"container_uri"`
}

type authorizerConfigurationModel struct {
	CustomJWTAuthorizer fwtypes.ListNestedObjectValueOf[CustomJWTAuthorizerConfigurationModel] `tfsdk:"custom_jwt_authorizer"`
}

var (
	_ flex.Expander  = authorizerConfigurationModel{}
	_ flex.Flattener = &authorizerConfigurationModel{}
)

func (m *authorizerConfigurationModel) Flatten(ctx context.Context, v any) (diags diag.Diagnostics) {
	switch t := v.(type) {
	case awstypes.AuthorizerConfigurationMemberCustomJWTAuthorizer:
		var model CustomJWTAuthorizerConfigurationModel
		d := flex.Flatten(ctx, t.Value, &model)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}
		m.CustomJWTAuthorizer = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)
		return diags

	default:
		return diags
	}
}

func (m authorizerConfigurationModel) Expand(ctx context.Context) (result any, diags diag.Diagnostics) {
	switch {
	case !m.CustomJWTAuthorizer.IsNull():
		model, d := m.CustomJWTAuthorizer.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		var r awstypes.AuthorizerConfigurationMemberCustomJWTAuthorizer
		diags.Append(flex.Expand(ctx, model, &r.Value)...)
		if diags.HasError() {
			return nil, diags
		}
		return &r, diags
	}
	return nil, diags
}

type CustomJWTAuthorizerConfigurationModel struct {
	DiscoveryUrl    types.String        `tfsdk:"discovery_url"`
	AllowedAudience fwtypes.SetOfString `tfsdk:"allowed_audience"`
	AllowedClients  fwtypes.SetOfString `tfsdk:"allowed_clients"`
}

type networkConfigurationModel struct {
	NetworkMode fwtypes.StringEnum[awstypes.NetworkMode] `tfsdk:"network_mode"`
}

type protocolConfigurationModel struct {
	ServerProtocol fwtypes.StringEnum[awstypes.ServerProtocol] `tfsdk:"server_protocol"`
}

type workloadIdentityDetailsModel struct {
	WorkloadIdentityArn types.String `tfsdk:"workload_identity_arn"`
}
