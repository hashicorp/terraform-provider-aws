// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package bedrockagentcore

import (
	"context"
	"fmt"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdkid "github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_bedrockagentcore_agent_runtime", name="Agent Runtime")
// @Tags(identifierAttribute="agent_runtime_arn")
// @Testing(tagsTest=false)
func newAgentRuntimeResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &agentRuntimeResource{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

type agentRuntimeResource struct {
	framework.ResourceWithModel[agentRuntimeResourceModel]
	framework.WithImportByIdentity
	framework.WithTimeouts
}

func (r *agentRuntimeResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"agent_runtime_arn": framework.ARNAttributeComputedOnly(),
			"agent_runtime_id":  framework.IDAttribute(),
			"agent_runtime_name": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexache.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]{0,47}$`), ""),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"agent_runtime_version": schema.StringAttribute{
				Computed: true,
			},
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 4096),
				},
			},
			"environment_variables": schema.MapAttribute{
				CustomType: fwtypes.MapOfStringType,
				Optional:   true,
			},
			"lifecycle_configuration": framework.ResourceOptionalComputedListOfObjectsAttribute[lifecycleConfigurationModel](ctx, 1, nil, listplanmodifier.UseStateForUnknown()),
			names.AttrRoleARN: schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
			},
			names.AttrTags:              tftags.TagsAttribute(),
			names.AttrTagsAll:           tftags.TagsAttributeComputedOnly(),
			"workload_identity_details": framework.ResourceComputedListOfObjectsAttribute[workloadIdentityDetailsModel](ctx, listplanmodifier.UseStateForUnknown()),
		},

		Blocks: map[string]schema.Block{
			"agent_runtime_artifact": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[agentRuntimeArtifactModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtMost(1),
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplaceIf(
						func(ctx context.Context, request planmodifier.ListRequest, response *listplanmodifier.RequiresReplaceIfFuncResponse) {
							// If code_configuration was set in the previous configuration and container_configuration is set in the planned configuration, a replacement is required — and vice versa.
							var prev, plan agentRuntimeArtifactModel
							smerr.AddEnrich(ctx, &response.Diagnostics, request.State.GetAttribute(ctx, path.Root("agent_runtime_artifact").AtListIndex(0), &prev))
							smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.GetAttribute(ctx, path.Root("agent_runtime_artifact").AtListIndex(0), &plan))
							if response.Diagnostics.HasError() {
								return
							}
							if (!prev.ContainerConfiguration.IsNull() && !plan.CodeConfiguration.IsNull()) ||
								(!prev.CodeConfiguration.IsNull() && !plan.ContainerConfiguration.IsNull()) {
								response.RequiresReplace = true
							}
						},
						"Artifact type change between code_configuration and container_configuration requires replacement",
						"",
					),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"code_configuration": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[codeConfigurationModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
								listvalidator.ExactlyOneOf(
									path.MatchRelative().AtParent().AtName("container_configuration"),
									path.MatchRelative().AtParent().AtName("code_configuration"),
								),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"entry_point": schema.ListAttribute{
										CustomType: fwtypes.ListOfStringType,
										Required:   true,
										Validators: []validator.List{
											listvalidator.SizeAtLeast(1),
											listvalidator.SizeAtMost(2),
											listvalidator.ValueStringsAre(stringvalidator.LengthBetween(1, 128)),
										},
										ElementType: types.StringType,
									},
									"runtime": schema.StringAttribute{
										Required:   true,
										CustomType: fwtypes.StringEnumType[awstypes.AgentManagedRuntimeType](),
									},
								},
								Blocks: map[string]schema.Block{
									"code": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[codeConfigurationCodeModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Blocks: map[string]schema.Block{
												"s3": schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[s3CodeConfigurationModel](ctx),
													Validators: []validator.List{
														listvalidator.SizeAtMost(1),
														listvalidator.ExactlyOneOf(
															// If another member is added to the union, this will need to be updated.
															path.MatchRelative().AtParent().AtName("s3"),
														),
													},
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															names.AttrBucket: schema.StringAttribute{
																Required: true,
																Validators: []validator.String{
																	stringvalidator.RegexMatches(regexache.MustCompile(`^[a-z0-9][a-z0-9.-]{1,61}[a-z0-9]$`), "must be a valid S3 bucket name"),
																},
															},
															names.AttrPrefix: schema.StringAttribute{
																Required: true,
																Validators: []validator.String{
																	stringvalidator.LengthBetween(1, 1024),
																},
															},
															"version_id": schema.StringAttribute{
																Optional: true,
																Validators: []validator.String{
																	stringvalidator.LengthBetween(3, 1024),
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
						"container_configuration": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[containerConfigurationModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
								listvalidator.ExactlyOneOf(
									path.MatchRelative().AtParent().AtName("container_configuration"),
									path.MatchRelative().AtParent().AtName("code_configuration"),
								),
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
							CustomType: fwtypes.NewListNestedObjectTypeOf[customJWTAuthorizerConfigurationModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"allowed_audience": schema.SetAttribute{
										CustomType: fwtypes.SetOfStringType,
										Optional:   true,
									},
									"allowed_clients": schema.SetAttribute{
										CustomType: fwtypes.SetOfStringType,
										Optional:   true,
									},
									"discovery_url": schema.StringAttribute{
										Required: true,
									},
								},
							},
						},
					},
				},
			},
			names.AttrNetworkConfiguration: schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[networkConfigurationModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"network_mode": schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.NetworkMode](),
							Required:   true,
						},
					},
					Blocks: map[string]schema.Block{
						"network_mode_config": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[vpcConfigModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrSecurityGroups: schema.SetAttribute{
										CustomType: fwtypes.SetOfStringType,
										Required:   true,
									},
									names.AttrSubnets: schema.SetAttribute{
										CustomType: fwtypes.SetOfStringType,
										Required:   true,
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
							CustomType: fwtypes.StringEnumType[awstypes.ServerProtocol](),
						},
					},
				},
			},
			"request_header_configuration": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[requestHeaderConfigurationModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"request_header_allowlist": schema.SetAttribute{
							CustomType: fwtypes.SetOfStringType,
							Optional:   true,
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

func (r *agentRuntimeResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data agentRuntimeResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentCoreClient(ctx)

	var input bedrockagentcorecontrol.CreateAgentRuntimeInput
	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, data, &input, fwflex.WithFieldNamePrefix("AgentRuntime")))
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.ClientToken = aws.String(sdkid.UniqueId())
	input.Tags = getTagsIn(ctx)

	var (
		out *bedrockagentcorecontrol.CreateAgentRuntimeOutput
		err error
	)
	err = tfresource.Retry(ctx, propagationTimeout, func(ctx context.Context) *tfresource.RetryError {
		out, err = conn.CreateAgentRuntime(ctx, &input)

		// IAM propagation.
		if tfawserr.ErrMessageContains(err, errCodeValidationException, "Role validation failed") {
			return tfresource.RetryableError(err)
		}
		if tfawserr.ErrMessageContains(err, errCodeValidationException, "Access denied while validating ECR URI") {
			return tfresource.RetryableError(err)
		}

		if err != nil {
			return tfresource.NonRetryableError(err)
		}

		return nil
	})
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, data.AgentRuntimeName.String())
		return
	}

	agentRuntimeID := aws.ToString(out.AgentRuntimeId)

	if _, err := waitAgentRuntimeCreated(ctx, conn, agentRuntimeID, r.CreateTimeout(ctx, data.Timeouts)); err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, agentRuntimeID)
		return
	}

	runtime, err := findAgentRuntimeByID(ctx, conn, agentRuntimeID)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, agentRuntimeID)
		return
	}

	// Set values for unknowns.
	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Flatten(ctx, runtime, &data, fwflex.WithFieldNamePrefix("AgentRuntime")))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, data))
}

func (r *agentRuntimeResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data agentRuntimeResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentCoreClient(ctx)

	agentRuntimeID := fwflex.StringValueFromFramework(ctx, data.AgentRuntimeID)
	out, err := findAgentRuntimeByID(ctx, conn, agentRuntimeID)
	if retry.NotFound(err) {
		smerr.AddOne(ctx, &response.Diagnostics, fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, agentRuntimeID)
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Flatten(ctx, out, &data, fwflex.WithFieldNamePrefix("AgentRuntime")))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data))
}

func (r *agentRuntimeResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var new, old agentRuntimeResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &new))
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &old))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentCoreClient(ctx)

	diff, d := fwflex.Diff(ctx, new, old)
	smerr.AddEnrich(ctx, &response.Diagnostics, d)
	if response.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		agentRuntimeID := fwflex.StringValueFromFramework(ctx, new.AgentRuntimeID)
		var input bedrockagentcorecontrol.UpdateAgentRuntimeInput
		smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, new, &input, fwflex.WithFieldNamePrefix("AgentRuntime")))
		if response.Diagnostics.HasError() {
			return
		}

		// Additional fields.
		input.ClientToken = aws.String(sdkid.UniqueId())

		out, err := conn.UpdateAgentRuntime(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, agentRuntimeID)
			return
		}

		smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Flatten(ctx, out, &new, fwflex.WithFieldNamePrefix("AgentRuntime")))
		if response.Diagnostics.HasError() {
			return
		}

		if _, err := waitAgentRuntimeUpdated(ctx, conn, agentRuntimeID, r.UpdateTimeout(ctx, new.Timeouts)); err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, agentRuntimeID)
			return
		}
	} else {
		new.AgentRuntimeVersion = old.AgentRuntimeVersion
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &new))
}

func (r *agentRuntimeResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data agentRuntimeResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentCoreClient(ctx)

	agentRuntimeID := fwflex.StringValueFromFramework(ctx, data.AgentRuntimeID)
	input := bedrockagentcorecontrol.DeleteAgentRuntimeInput{
		AgentRuntimeId: aws.String(agentRuntimeID),
	}

	_, err := conn.DeleteAgentRuntime(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, agentRuntimeID)
		return
	}

	if _, err := waitAgentRuntimeDeleted(ctx, conn, agentRuntimeID, r.DeleteTimeout(ctx, data.Timeouts)); err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, agentRuntimeID)
		return
	}
}

func (r *agentRuntimeResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("agent_runtime_id"), request, response)
}

func waitAgentRuntimeCreated(ctx context.Context, conn *bedrockagentcorecontrol.Client, id string, timeout time.Duration) (*bedrockagentcorecontrol.GetAgentRuntimeOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.AgentRuntimeStatusCreating),
		Target:                    enum.Slice(awstypes.AgentRuntimeStatusReady),
		Refresh:                   statusAgentRuntime(conn, id),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*bedrockagentcorecontrol.GetAgentRuntimeOutput); ok {
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitAgentRuntimeUpdated(ctx context.Context, conn *bedrockagentcorecontrol.Client, id string, timeout time.Duration) (*bedrockagentcorecontrol.GetAgentRuntimeOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.AgentRuntimeStatusUpdating),
		Target:                    enum.Slice(awstypes.AgentRuntimeStatusReady),
		Refresh:                   statusAgentRuntime(conn, id),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*bedrockagentcorecontrol.GetAgentRuntimeOutput); ok {
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitAgentRuntimeDeleted(ctx context.Context, conn *bedrockagentcorecontrol.Client, id string, timeout time.Duration) (*bedrockagentcorecontrol.GetAgentRuntimeOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.AgentRuntimeStatusDeleting, awstypes.AgentRuntimeStatusReady),
		Target:  []string{},
		Refresh: statusAgentRuntime(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*bedrockagentcorecontrol.GetAgentRuntimeOutput); ok {
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func statusAgentRuntime(conn *bedrockagentcorecontrol.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
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

	return findAgentRuntime(ctx, conn, &input)
}

func findAgentRuntime(ctx context.Context, conn *bedrockagentcorecontrol.Client, input *bedrockagentcorecontrol.GetAgentRuntimeInput) (*bedrockagentcorecontrol.GetAgentRuntimeOutput, error) {
	out, err := conn.GetAgentRuntime(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, smarterr.NewError(&retry.NotFoundError{
			LastError: err,
		})
	}

	if err != nil {
		return nil, smarterr.NewError(err)
	}

	if out == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	return out, nil
}

type agentRuntimeResourceModel struct {
	framework.WithRegionModel
	AgentRuntimeARN            types.String                                                     `tfsdk:"agent_runtime_arn"`
	AgentRuntimeArtifact       fwtypes.ListNestedObjectValueOf[agentRuntimeArtifactModel]       `tfsdk:"agent_runtime_artifact"`
	AgentRuntimeID             types.String                                                     `tfsdk:"agent_runtime_id"`
	AgentRuntimeName           types.String                                                     `tfsdk:"agent_runtime_name"`
	AgentRuntimeVersion        types.String                                                     `tfsdk:"agent_runtime_version"`
	AuthorizerConfiguration    fwtypes.ListNestedObjectValueOf[authorizerConfigurationModel]    `tfsdk:"authorizer_configuration"`
	Description                types.String                                                     `tfsdk:"description"`
	EnvironmentVariables       fwtypes.MapOfString                                              `tfsdk:"environment_variables"`
	LifecycleConfiguration     fwtypes.ListNestedObjectValueOf[lifecycleConfigurationModel]     `tfsdk:"lifecycle_configuration"`
	NetworkConfiguration       fwtypes.ListNestedObjectValueOf[networkConfigurationModel]       `tfsdk:"network_configuration"`
	ProtocolConfiguration      fwtypes.ListNestedObjectValueOf[protocolConfigurationModel]      `tfsdk:"protocol_configuration"`
	RequestHeaderConfiguration fwtypes.ListNestedObjectValueOf[requestHeaderConfigurationModel] `tfsdk:"request_header_configuration"`
	RoleARN                    fwtypes.ARN                                                      `tfsdk:"role_arn"`
	Tags                       tftags.Map                                                       `tfsdk:"tags"`
	TagsAll                    tftags.Map                                                       `tfsdk:"tags_all"`
	Timeouts                   timeouts.Value                                                   `tfsdk:"timeouts"`
	WorkloadIdentityDetails    fwtypes.ListNestedObjectValueOf[workloadIdentityDetailsModel]    `tfsdk:"workload_identity_details"`
}

type agentRuntimeArtifactModel struct {
	CodeConfiguration      fwtypes.ListNestedObjectValueOf[codeConfigurationModel]      `tfsdk:"code_configuration"`
	ContainerConfiguration fwtypes.ListNestedObjectValueOf[containerConfigurationModel] `tfsdk:"container_configuration"`
}

var (
	_ fwflex.Expander  = agentRuntimeArtifactModel{}
	_ fwflex.Flattener = &agentRuntimeArtifactModel{}
)

func (m *agentRuntimeArtifactModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics
	switch t := v.(type) {
	case awstypes.AgentRuntimeArtifactMemberCodeConfiguration:
		var data codeConfigurationModel
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t.Value, &data))
		if diags.HasError() {
			return diags
		}
		m.CodeConfiguration = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &data)
	case awstypes.AgentRuntimeArtifactMemberContainerConfiguration:
		var data containerConfigurationModel
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t.Value, &data))
		if diags.HasError() {
			return diags
		}
		m.ContainerConfiguration = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &data)

	default:
		diags.AddError(
			"Unsupported Type",
			fmt.Sprintf("artifact flatten: %T", v),
		)
	}
	return diags
}

func (m agentRuntimeArtifactModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	switch {
	case !m.CodeConfiguration.IsNull():
		data, d := m.CodeConfiguration.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		var r awstypes.AgentRuntimeArtifactMemberCodeConfiguration
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, data, &r.Value))
		if diags.HasError() {
			return nil, diags
		}
		return &r, diags
	case !m.ContainerConfiguration.IsNull():
		data, d := m.ContainerConfiguration.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		var r awstypes.AgentRuntimeArtifactMemberContainerConfiguration
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, data, &r.Value))
		if diags.HasError() {
			return nil, diags
		}
		return &r, diags
	}
	return nil, diags
}

type codeConfigurationModel struct {
	Code       fwtypes.ListNestedObjectValueOf[codeConfigurationCodeModel] `tfsdk:"code"`
	EntryPoint fwtypes.ListOfString                                        `tfsdk:"entry_point"`
	Runtime    fwtypes.StringEnum[awstypes.AgentManagedRuntimeType]        `tfsdk:"runtime"`
}

type codeConfigurationCodeModel struct {
	S3 fwtypes.ListNestedObjectValueOf[s3CodeConfigurationModel] `tfsdk:"s3"`
}

type s3CodeConfigurationModel struct {
	Bucket    types.String `tfsdk:"bucket"`
	Prefix    types.String `tfsdk:"prefix"`
	VersionID types.String `tfsdk:"version_id"`
}

var (
	_ fwflex.Expander  = codeConfigurationCodeModel{}
	_ fwflex.Flattener = &codeConfigurationCodeModel{}
)

func (m *codeConfigurationCodeModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics
	switch t := v.(type) {
	case awstypes.CodeMemberS3:
		var data s3CodeConfigurationModel
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t.Value, &data))
		if diags.HasError() {
			return diags
		}
		m.S3 = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &data)

	default:
		diags.AddError(
			"Unsupported Type",
			fmt.Sprintf("code configuration code flatten: %T", v),
		)
	}
	return diags
}

func (m codeConfigurationCodeModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	switch {
	case !m.S3.IsNull():
		data, d := m.S3.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		var r awstypes.CodeMemberS3
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, data, &r.Value))
		if diags.HasError() {
			return nil, diags
		}
		return &r, diags
	}
	return nil, diags
}

type containerConfigurationModel struct {
	ContainerURI types.String `tfsdk:"container_uri"`
}

type authorizerConfigurationModel struct {
	CustomJWTAuthorizer fwtypes.ListNestedObjectValueOf[customJWTAuthorizerConfigurationModel] `tfsdk:"custom_jwt_authorizer"`
}

var (
	_ fwflex.Expander  = authorizerConfigurationModel{}
	_ fwflex.Flattener = &authorizerConfigurationModel{}
)

func (m *authorizerConfigurationModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics
	switch t := v.(type) {
	case awstypes.AuthorizerConfigurationMemberCustomJWTAuthorizer:
		var data customJWTAuthorizerConfigurationModel
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t.Value, &data))
		if diags.HasError() {
			return diags
		}
		m.CustomJWTAuthorizer = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &data)

	default:
		diags.AddError(
			"Unsupported Type",
			fmt.Sprintf("authorization configuration flatten: %T", v),
		)
	}
	return diags
}

func (m authorizerConfigurationModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	switch {
	case !m.CustomJWTAuthorizer.IsNull():
		data, d := m.CustomJWTAuthorizer.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		var r awstypes.AuthorizerConfigurationMemberCustomJWTAuthorizer
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, data, &r.Value))
		if diags.HasError() {
			return nil, diags
		}
		return &r, diags
	}
	return nil, diags
}

type customJWTAuthorizerConfigurationModel struct {
	AllowedAudience fwtypes.SetOfString `tfsdk:"allowed_audience"`
	AllowedClients  fwtypes.SetOfString `tfsdk:"allowed_clients"`
	DiscoveryURL    types.String        `tfsdk:"discovery_url"`
}

type lifecycleConfigurationModel struct {
	IdleRuntimeSessionTimeout types.Int32 `tfsdk:"idle_runtime_session_timeout"`
	MaxLifetime               types.Int32 `tfsdk:"max_lifetime"`
}

type networkConfigurationModel struct {
	NetworkMode       fwtypes.StringEnum[awstypes.NetworkMode]        `tfsdk:"network_mode"`
	NetworkModeConfig fwtypes.ListNestedObjectValueOf[vpcConfigModel] `tfsdk:"network_mode_config"`
}

type vpcConfigModel struct {
	SecurityGroups fwtypes.SetOfString `tfsdk:"security_groups"`
	Subnets        fwtypes.SetOfString `tfsdk:"subnets"`
}

type protocolConfigurationModel struct {
	ServerProtocol fwtypes.StringEnum[awstypes.ServerProtocol] `tfsdk:"server_protocol"`
}

type requestHeaderConfigurationModel struct {
	RequestHeaderAllowlist fwtypes.SetOfString `tfsdk:"request_header_allowlist"`
}

var (
	_ fwflex.Expander  = requestHeaderConfigurationModel{}
	_ fwflex.Flattener = &requestHeaderConfigurationModel{}
)

func (m *requestHeaderConfigurationModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics
	switch t := v.(type) {
	case awstypes.RequestHeaderConfigurationMemberRequestHeaderAllowlist:
		m.RequestHeaderAllowlist = fwflex.FlattenFrameworkStringValueSetOfString(ctx, t.Value)

	default:
		diags.AddError(
			"Unsupported Type",
			fmt.Sprintf("artifact flatten: %T", v),
		)
	}
	return diags
}

func (m requestHeaderConfigurationModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	switch {
	case !m.RequestHeaderAllowlist.IsNull():
		var diags diag.Diagnostics
		var r awstypes.RequestHeaderConfigurationMemberRequestHeaderAllowlist
		r.Value = fwflex.ExpandFrameworkStringValueSet(ctx, m.RequestHeaderAllowlist)
		return &r, diags
	}
	return nil, diags
}

type workloadIdentityDetailsModel struct {
	WorkloadIdentityARN types.String `tfsdk:"workload_identity_arn"`
}
