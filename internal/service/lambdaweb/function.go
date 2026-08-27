// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// Pre-GA: this resource targets the not-yet-public aws-sdk-go-v2
// "lambdaweb" service client, currently satisfied by the hand-written
// shim in .pre-ga-sdk/ (see the replace directive in go.mod). Swap to the
// real SDK module when it ships at GA.

package lambdaweb

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambdaweb"
	awstypes "github.com/aws/aws-sdk-go-v2/service/lambdaweb/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
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

// @FrameworkResource("aws_lambdaweb_function", name="Function")
// @IdentityAttribute("function_name")
// @Tags(identifierAttribute="arn")
// @Testing(hasNoPreExistingResource=true)
// @Testing(preCheck="testAccPreCheck")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/lambdaweb;lambdaweb.GetWebFunctionOutput")
// @Testing(importStateIdAttribute="function_name")
func newFunctionResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &functionResource{}

	r.SetDefaultCreateTimeout(15 * time.Minute)
	r.SetDefaultUpdateTimeout(15 * time.Minute)
	r.SetDefaultDeleteTimeout(15 * time.Minute)

	return r, nil
}

const (
	ResNameFunction = "Function"

	iamPropagationTimeout = 5 * time.Minute
)

type functionResource struct {
	framework.ResourceWithModel[functionResourceModel]
	framework.WithTimeouts
	framework.WithImportByIdentity
}

func (r *functionResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrID:  framework.IDAttribute(),
			"function_name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 64),
				},
			},
			names.AttrState: schema.StringAttribute{
				Computed: true,
			},
			"latest_revision_id": schema.StringAttribute{
				Computed: true,
			},
			names.AttrDomainName: schema.StringAttribute{
				Computed: true,
			},
			// PerRegion endpoints serve an independent domain per region; without
			// this the regional domains of an inline endpoint_config cannot be
			// referenced from a configuration at all.
			"regional_domain_names": schema.MapAttribute{
				CustomType:  fwtypes.MapOfStringType,
				Computed:    true,
				ElementType: types.StringType,
			},
			// Why a function is Pending or Failed.
			"state_reason": schema.StringAttribute{
				Computed: true,
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"revision_config": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[revisionConfigModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrDescription: schema.StringAttribute{
							Optional: true,
						},
						names.AttrKMSKeyARN: schema.StringAttribute{
							CustomType: fwtypes.ARNType,
							Optional:   true,
						},
					},
					Blocks: map[string]schema.Block{
						"build_config": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[buildConfigModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Blocks: map[string]schema.Block{
									"runtime_config": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[runtimeConfigModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"runtime": schema.StringAttribute{
													Required: true,
													Validators: []validator.String{
														stringvalidator.RegexMatches(
															regexache.MustCompile(`^[a-z][a-z0-9]*[0-9]+(\.[a-z0-9]+)*$`),
															"must be a runtime identifier such as nodejs24.x",
														),
													},
												},
											},
										},
									},
									"code_config": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[codeConfigModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Blocks: map[string]schema.Block{
												"s3_object": schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[s3ObjectModel](ctx),
													Validators: []validator.List{
														listvalidator.SizeAtMost(1),
													},
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															names.AttrBucket: schema.StringAttribute{
																Required: true,
															},
															names.AttrKey: schema.StringAttribute{
																Required: true,
															},
															"version_id": schema.StringAttribute{
																Optional: true,
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
						"service_config": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[serviceConfigModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrExecutionRoleARN: schema.StringAttribute{
										CustomType: fwtypes.ARNType,
										Required:   true,
									},
									"timeout_seconds": schema.Int64Attribute{
										Optional: true,
										// Assigned by the service (30) when unset: must be
										// Computed, otherwise the value read back produces a
										// perpetual diff.
										Computed: true,
										PlanModifiers: []planmodifier.Int64{
											int64planmodifier.UseStateForUnknown(),
										},
										Validators: []validator.Int64{
											int64validator.Between(3, 900),
										},
									},
									"max_concurrency_per_environment": schema.Int64Attribute{
										Optional: true,
										// Assigned by the service (64) when unset: see
										// timeout_seconds above.
										Computed: true,
										PlanModifiers: []planmodifier.Int64{
											int64planmodifier.UseStateForUnknown(),
										},
										Validators: []validator.Int64{
											int64validator.Between(1, 128),
										},
									},
									"environment_variables": schema.MapAttribute{
										CustomType:  fwtypes.MapOfStringType,
										Optional:    true,
										Sensitive:   true,
										ElementType: types.StringType,
									},
								},
								Blocks: map[string]schema.Block{
									"telemetry_config": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[telemetryConfigModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Blocks: map[string]schema.Block{
												"logging_config": schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[loggingConfigModel](ctx),
													Validators: []validator.List{
														listvalidator.SizeAtMost(1),
													},
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															"log_group": schema.StringAttribute{
																Optional: true,
															},
															"application_log_level": schema.StringAttribute{
																CustomType: fwtypes.StringEnumType[awstypes.ApplicationLogLevel](),
																Optional:   true,
															},
															"system_log_level": schema.StringAttribute{
																CustomType: fwtypes.StringEnumType[awstypes.SystemLogLevel](),
																Optional:   true,
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
				},
			},
			"endpoint_config": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[endpointConfigModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"endpoint_name": schema.StringAttribute{
							Required: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
							Validators: []validator.String{
								stringvalidator.LengthBetween(1, 64),
							},
						},
						names.AttrEndpointType: schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.EndpointType](),
							Required:   true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
						"auth_type": schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.AuthType](),
							Required:   true,
						},
						"auto_deployment_mode": schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.AutoDeploymentMode](),
							Optional:   true,
							Computed:   true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						names.AttrDescription: schema.StringAttribute{
							Optional: true,
						},
						// Objects rather than blocks: the service assigns
						// account-level defaults when unset, so they must be
						// Computed (see aws_lambdaweb_endpoint).
						"scaling_config": schema.ObjectAttribute{
							CustomType: fwtypes.NewObjectTypeOf[scalingConfigModel](ctx),
							Optional:   true,
							Computed:   true,
							PlanModifiers: []planmodifier.Object{
								objectplanmodifier.UseStateForUnknown(),
							},
						},
						"throttle_config": schema.ObjectAttribute{
							CustomType: fwtypes.NewObjectTypeOf[throttleConfigModel](ctx),
							Optional:   true,
							Computed:   true,
							PlanModifiers: []planmodifier.Object{
								objectplanmodifier.UseStateForUnknown(),
							},
						},
						"regions": schema.SetAttribute{
							CustomType:  fwtypes.SetOfStringType,
							Optional:    true,
							Computed:    true,
							ElementType: types.StringType,
							PlanModifiers: []planmodifier.Set{
								// Regions are immutable on an endpoint and
								// UpdateWebFunctionEndpoint does not accept them.
								setplanmodifier.RequiresReplace(),
								// HomeRegion endpoints default to the function's
								// region server-side when regions is omitted, so
								// keep the known value to avoid post-apply drift.
								setplanmodifier.UseStateForUnknown(),
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
		},
	}
}

// ValidateConfig applies the rules the service enforces on the inline
// endpoint_config block and on the environment variables, so they surface at
// plan time rather than as an API error midway through an apply.
func (r *functionResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var cfg functionResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Config.Get(ctx, &cfg))
	if resp.Diagnostics.HasError() {
		return
	}

	validateEnvironmentVariablesSize(ctx, cfg, resp)

	if cfg.EndpointConfig.IsNull() || cfg.EndpointConfig.IsUnknown() {
		return
	}

	eps, d := cfg.EndpointConfig.ToSlice(ctx)
	smerr.AddEnrich(ctx, &resp.Diagnostics, d)
	if resp.Diagnostics.HasError() || len(eps) == 0 {
		return
	}
	ep := eps[0]
	if ep.EndpointType.IsUnknown() || ep.AutoDeploymentMode.IsUnknown() {
		return
	}

	endpointType := ep.EndpointType.ValueString()
	if endpointType != string(awstypes.EndpointTypeMultiRegion) && endpointType != string(awstypes.EndpointTypePerRegion) {
		return
	}

	mode := awstypes.AutoDeploymentModeLatestRevision
	if !ep.AutoDeploymentMode.IsNull() {
		mode = awstypes.AutoDeploymentMode(ep.AutoDeploymentMode.ValueString())
	}
	if mode != awstypes.AutoDeploymentModeDisabled {
		resp.Diagnostics.AddAttributeError(path.Root("endpoint_config"),
			"Invalid auto_deployment_mode",
			fmt.Sprintf("%s endpoints require `auto_deployment_mode = \"Disabled\"`.", endpointType))
	}

	// The home region is added automatically and at least two distinct regions
	// are then required, so a single configured region is always rejected.
	if !ep.Regions.IsNull() && !ep.Regions.IsUnknown() && len(ep.Regions.Elements()) == 1 {
		resp.Diagnostics.AddAttributeError(path.Root("endpoint_config"),
			"Invalid regions",
			fmt.Sprintf("%s endpoints require at least 2 distinct regions, or no `regions` at all: the home region is added automatically.", endpointType))
	}
}

// ModifyPlan warns when an apply is about to publish a revision that the inline
// endpoint will not serve. An endpoint with `auto_deployment_mode = "Disabled"`
// keeps the traffic weights it already has, and weights are only settable
// through UpdateWebFunctionEndpoint: CreateWebFunction's endpointConfig has no
// revisionWeights member, so `endpoint_config` cannot express them. Without a
// warning the apply reports success, the next plan is empty, and the endpoint
// keeps serving the old revision indefinitely. `MultiRegion` and `PerRegion`
// endpoints are always affected because the service requires `Disabled` there.
// environmentVariablesMaxBytes is the documented 32 KB limit the service
// enforces on a revision's environment variables.
const environmentVariablesMaxBytes = 32 * 1024

// validateEnvironmentVariablesSize reports environment variables that cannot
// fit. The service rejects them, but only up to a point: past roughly 37 KB the
// whole request is too large and the error stops mentioning environment
// variables at all ("Request must be smaller than 37888 bytes"), which sends
// people looking in the wrong place. The keys and values are summed without the
// serialization overhead the service also counts, so this only ever rejects a
// configuration the service would reject too.
func validateEnvironmentVariablesSize(ctx context.Context, cfg functionResourceModel, resp *resource.ValidateConfigResponse) {
	if cfg.RevisionConfig.IsNull() || cfg.RevisionConfig.IsUnknown() {
		return
	}

	revisionConfig, d := cfg.RevisionConfig.ToPtr(ctx)
	if d.HasError() || revisionConfig == nil || revisionConfig.ServiceConfig.IsNull() || revisionConfig.ServiceConfig.IsUnknown() {
		return
	}

	serviceConfig, d := revisionConfig.ServiceConfig.ToPtr(ctx)
	if d.HasError() || serviceConfig == nil || serviceConfig.EnvironmentVariables.IsNull() || serviceConfig.EnvironmentVariables.IsUnknown() {
		return
	}

	var size int
	for name, value := range serviceConfig.EnvironmentVariables.Elements() {
		if value.IsUnknown() {
			return
		}
		size += len(name)
		if v, ok := value.(types.String); ok {
			size += len(v.ValueString())
		}
	}

	if size > environmentVariablesMaxBytes {
		resp.Diagnostics.AddAttributeError(path.Root("revision_config"),
			"Environment variables too large",
			fmt.Sprintf("`environment_variables` must not exceed %d bytes across all names and values; the configuration holds %d.",
				environmentVariablesMaxBytes, size))
	}
}

func (r *functionResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.State.Raw.IsNull() || req.Plan.Raw.IsNull() {
		return
	}

	var state, plan functionResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	// Only a revision_config change publishes a new revision.
	if plan.RevisionConfig.Equal(state.RevisionConfig) || plan.EndpointConfig.IsNull() || plan.EndpointConfig.IsUnknown() {
		return
	}

	ep, d := plan.EndpointConfig.ToPtr(ctx)
	smerr.AddEnrich(ctx, &resp.Diagnostics, d)
	if resp.Diagnostics.HasError() || ep == nil || ep.AutoDeploymentMode.IsUnknown() {
		return
	}
	if ep.AutoDeploymentMode.ValueString() != string(awstypes.AutoDeploymentModeDisabled) {
		return
	}

	resp.Diagnostics.AddAttributeWarning(path.Root("endpoint_config"),
		"New revision will not receive traffic",
		fmt.Sprintf("Endpoint %q uses `auto_deployment_mode = \"Disabled\"`, so it keeps serving the revision its traffic weights already point at. "+
			"This apply publishes a new revision, but traffic weights are only settable through the UpdateWebFunctionEndpoint API, which `endpoint_config` "+
			"does not expose. The apply will report success and the next plan will be empty while the endpoint still serves the previous revision.\n\n"+
			"To shift traffic to new revisions, manage the endpoint with a separate `aws_lambdaweb_endpoint` resource and point its `revision_weights` at "+
			"`latest_revision_id` of this function. `MultiRegion` and `PerRegion` endpoints always require `auto_deployment_mode = \"Disabled\"`.",
			ep.EndpointName.ValueString()))
}

func (r *functionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().LambdaWebClient(ctx)

	var plan functionResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	var input lambdaweb.CreateWebFunctionInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input))
	if resp.Diagnostics.HasError() {
		return
	}
	input.Tags = getTagsIn(ctx)

	name := plan.FunctionName.ValueString()

	// IAM/S3-policy eventual consistency: a freshly created execution role or
	// bucket policy can take tens of seconds to become visible to the service.
	outputRaw, err := tfresource.RetryWhen(ctx, iamPropagationTimeout,
		func(ctx context.Context) (any, error) {
			return conn.CreateWebFunction(ctx, &input)
		},
		func(err error) (bool, error) {
			if errs.IsAErrorMessageContains[*awstypes.ValidationException](err, "cannot be assumed") {
				return true, err
			}
			if errs.IsAErrorMessageContains[*awstypes.ValidationException](err, "does not have s3:GetObject") {
				return true, err
			}
			if errs.IsAErrorMessageContains[*awstypes.ValidationException](err, "S3 versioning enabled") {
				return true, err
			}
			return false, err
		})
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, name)
		return
	}
	createOut := outputRaw.(*lambdaweb.CreateWebFunctionOutput)

	// Set partial state so a failed wait below does not orphan the function.
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root(names.AttrID), name)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("function_name"), name)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := waitFunctionCreated(ctx, conn, name, r.CreateTimeout(ctx, plan.Timeouts))
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, name)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, out, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	plan.ID = fwflex.StringToFramework(ctx, out.FunctionName)
	plan.ARN = fwflex.StringToFramework(ctx, out.FunctionArn)

	plan.LatestRevisionID = types.StringNull()
	if createOut.Revision != nil {
		plan.LatestRevisionID = fwflex.StringToFramework(ctx, createOut.Revision.RevisionId)

		// Read the revision back: the function-level output carries no build or
		// service configuration, so server-assigned values inside
		// revision_config would otherwise remain unknown after apply.
		rev, err := findRevisionByID(ctx, conn, name, aws.ToString(createOut.Revision.RevisionId))
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, name)
			return
		}

		var revModel revisionConfigModel
		smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, rev, &revModel))
		if resp.Diagnostics.HasError() {
			return
		}
		restoreTelemetryConfig(ctx, &revModel, plan.RevisionConfig)
		plan.RevisionConfig = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &revModel)
	}

	plan.DomainName = types.StringNull()
	plan.RegionalDomainNames = fwtypes.NewMapValueOfMust[types.String](ctx, map[string]attr.Value{})
	if !plan.EndpointConfig.IsNull() {
		endpointName, err := endpointNameFromConfig(ctx, plan.EndpointConfig)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, name)
			return
		}

		ep, err := waitEndpointActive(ctx, conn, name, endpointName, r.CreateTimeout(ctx, plan.Timeouts))
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, name)
			return
		}
		plan.DomainName = fwflex.StringToFramework(ctx, ep.DomainName)
		plan.RegionalDomainNames = regionalDomainNames(ctx, ep.RegionalEndpoints)

		var epModel endpointConfigModel
		smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, ep, &epModel))
		if resp.Diagnostics.HasError() {
			return
		}
		epModel.Description = normalizeDescription(epModel.Description)
		plan.EndpointConfig = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &epModel)
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *functionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().LambdaWebClient(ctx)

	var state functionResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	name := state.FunctionName.ValueString()

	out, err := findFunctionByName(ctx, conn, name)
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, name)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, out, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	state.ID = fwflex.StringToFramework(ctx, out.FunctionName)
	state.ARN = fwflex.StringToFramework(ctx, out.FunctionArn)

	revisionID, err := findLatestRevisionID(ctx, conn, name)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, name)
		return
	}
	state.LatestRevisionID = fwflex.StringToFramework(ctx, revisionID)

	// Refresh revision_config from the latest revision. GetWebFunction does not
	// return the build and service configuration, but GetWebFunctionRevision
	// does, so read it explicitly: otherwise revision_config is never
	// reconciled with the service (no drift detection on code, environment
	// variables, description or the KMS key) and cannot be populated on import.
	priorRevisionConfig := state.RevisionConfig
	state.RevisionConfig = fwtypes.NewListNestedObjectValueOfNull[revisionConfigModel](ctx)
	if revisionID != nil {
		rev, err := findRevisionByID(ctx, conn, name, aws.ToString(revisionID))
		switch {
		case retry.NotFound(err):
		case err != nil:
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, name)
			return
		default:
			var revModel revisionConfigModel
			smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, rev, &revModel))
			if resp.Diagnostics.HasError() {
				return
			}
			restoreTelemetryConfig(ctx, &revModel, priorRevisionConfig)
			state.RevisionConfig = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &revModel)
		}
	}

	state.DomainName = types.StringNull()
	state.RegionalDomainNames = fwtypes.NewMapValueOfMust[types.String](ctx, map[string]attr.Value{})
	switch endpointName, err := readEndpointName(ctx, conn, name, state.EndpointConfig); {
	case err != nil:
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, name)
		return
	case endpointName != "":
		ep, err := findEndpointByName(ctx, conn, name, endpointName)
		switch {
		case retry.NotFound(err):
			// The endpoint described by endpoint_config was deleted outside of
			// Terraform. Clear the block so the plan proposes restoring it:
			// keeping it reports no changes while the function serves no
			// traffic, and pairing a stale endpoint_config with the null
			// domain_name set above breaks every configuration that references
			// the domain, leaving plan and apply unable to run at all.
			state.EndpointConfig = fwtypes.NewListNestedObjectValueOfNull[endpointConfigModel](ctx)
		case err != nil:
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, name)
			return
		default:
			state.DomainName = fwflex.StringToFramework(ctx, ep.DomainName)
			state.RegionalDomainNames = regionalDomainNames(ctx, ep.RegionalEndpoints)

			var epModel endpointConfigModel
			smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, ep, &epModel))
			if resp.Diagnostics.HasError() {
				return
			}
			epModel.Description = normalizeDescription(epModel.Description)
			state.EndpointConfig = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &epModel)
		}
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *functionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().LambdaWebClient(ctx)

	var plan, state functionResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	name := plan.FunctionName.ValueString()

	if !plan.RevisionConfig.Equal(state.RevisionConfig) {
		// CreateWebFunctionRevision takes description/kmsKeyArn/buildConfig/
		// serviceConfig at the top level (only CreateWebFunction nests them
		// under revisionConfig), so expand the block, not the resource model.
		revisionConfig, diags := plan.RevisionConfig.ToPtr(ctx)
		smerr.AddEnrich(ctx, &resp.Diagnostics, diags)
		if resp.Diagnostics.HasError() {
			return
		}

		var input lambdaweb.CreateWebFunctionRevisionInput
		smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, revisionConfig, &input))
		if resp.Diagnostics.HasError() {
			return
		}
		input.FunctionName = aws.String(name)

		revOut, err := conn.CreateWebFunctionRevision(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, name)
			return
		}

		revisionID := aws.ToString(revOut.RevisionId)
		if _, err := waitRevisionActive(ctx, conn, name, revisionID, r.UpdateTimeout(ctx, plan.Timeouts)); err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, name)
			return
		}
		plan.LatestRevisionID = fwflex.StringToFramework(ctx, revOut.RevisionId)
	} else {
		plan.LatestRevisionID = state.LatestRevisionID
	}

	// Carried over unless the endpoint is actually updated below: both are
	// computed, so leaving them unknown fails the apply.
	plan.DomainName = state.DomainName
	plan.RegionalDomainNames = state.RegionalDomainNames
	if !plan.EndpointConfig.Equal(state.EndpointConfig) && !plan.EndpointConfig.IsNull() {
		endpointName, err := endpointNameFromConfig(ctx, plan.EndpointConfig)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, name)
			return
		}

		// UpdateWebFunctionEndpoint takes description/authType/
		// autoDeploymentMode at the top level; expand the endpoint_config
		// block, not the resource model.
		endpointConfig, diags := plan.EndpointConfig.ToPtr(ctx)
		smerr.AddEnrich(ctx, &resp.Diagnostics, diags)
		if resp.Diagnostics.HasError() {
			return
		}

		var input lambdaweb.UpdateWebFunctionEndpointInput
		smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, endpointConfig, &input))
		if resp.Diagnostics.HasError() {
			return
		}
		input.FunctionName = aws.String(name)
		input.EndpointName = aws.String(endpointName)
		// An omitted description means "leave unchanged" to the API, so an empty
		// string is what clears one that was removed from the configuration.
		if endpointConfig.Description.IsNull() {
			if prior, d := state.EndpointConfig.ToPtr(ctx); !d.HasError() && prior != nil && !prior.Description.IsNull() {
				input.Description = aws.String("")
			}
		}

		if _, err := conn.UpdateWebFunctionEndpoint(ctx, &input); err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, name)
			return
		}

		ep, err := waitEndpointUpdated(ctx, conn, name, endpointName, r.UpdateTimeout(ctx, plan.Timeouts))
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, name)
			return
		}
		plan.DomainName = fwflex.StringToFramework(ctx, ep.DomainName)
		plan.RegionalDomainNames = regionalDomainNames(ctx, ep.RegionalEndpoints)

		// Mirror Create: flatten the endpoint back so computed sub-attributes
		// (auto_deployment_mode, regions) are known after apply.
		var epModel endpointConfigModel
		smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, ep, &epModel))
		if resp.Diagnostics.HasError() {
			return
		}
		epModel.Description = normalizeDescription(epModel.Description)
		plan.EndpointConfig = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &epModel)
	}

	out, err := waitFunctionUpdated(ctx, conn, name, r.UpdateTimeout(ctx, plan.Timeouts))
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, name)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, out, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *functionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().LambdaWebClient(ctx)

	var state functionResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	name := state.FunctionName.ValueString()

	input := lambdaweb.DeleteWebFunctionInput{
		FunctionName: aws.String(name),
	}
	_, err := conn.DeleteWebFunction(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, name)
		return
	}

	if _, err := waitFunctionDeleted(ctx, conn, name, r.DeleteTimeout(ctx, state.Timeouts)); err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, name)
		return
	}
}

func readEndpointName(ctx context.Context, conn *lambdaweb.Client, functionName string, cfg fwtypes.ListNestedObjectValueOf[endpointConfigModel]) (string, error) {
	if !cfg.IsNull() {
		return endpointNameFromConfig(ctx, cfg)
	}

	// No configuration to match against, i.e. import. CreateWebFunction always
	// creates an endpoint with the function, so the oldest endpoint is the one
	// endpoint_config describes. Giving up when a function has several
	// endpoints would leave the Required endpoint_config block empty in state,
	// and populating it from null triggers the RequiresReplace plan modifiers
	// on endpoint_name and endpoint_type: the first apply after import would
	// then replace the function, silently deleting every other endpoint it
	// owns along with their domain names.
	input := lambdaweb.ListWebFunctionEndpointsInput{
		FunctionName: aws.String(functionName),
	}

	var oldest *awstypes.FunctionEndpointSummary
	for {
		out, err := conn.ListWebFunctionEndpoints(ctx, &input)
		if err != nil {
			return "", err
		}
		for i := range out.Endpoints {
			if oldest == nil || createdBefore(aws.ToString(out.Endpoints[i].CreatedAt), aws.ToString(oldest.CreatedAt)) {
				oldest = &out.Endpoints[i]
			}
		}
		if out.NextToken == nil {
			break
		}
		input.NextToken = out.NextToken
	}

	if oldest == nil {
		return "", nil
	}
	return aws.ToString(oldest.EndpointName), nil
}

func endpointNameFromConfig(ctx context.Context, l fwtypes.ListNestedObjectValueOf[endpointConfigModel]) (string, error) {
	models, d := l.ToSlice(ctx)
	if d.HasError() {
		return "", errors.New("reading endpoint_config")
	}
	if len(models) == 0 {
		return "", errors.New("empty endpoint_config")
	}
	return models[0].EndpointName.ValueString(), nil
}

// restoreTelemetryConfig keeps telemetry_config as configuration-only state.
// The service assigns default telemetry (a log group and INFO log levels) to
// every revision and reports it back, but revisions are immutable, so
// telemetry the practitioner never wrote is a server default, not drift:
// reading it into state breaks every apply with "block count changed from 0 to
// 1", and a partially configured block (say, log_group only) would surface the
// filled-in log levels as a perpetual diff. prior is the configured
// revision_config (the plan on create, prior state on refresh); its
// telemetry_config replaces the server's wholesale. On import there is no
// prior, so telemetry_config is left unpopulated rather than importing the
// server defaults.
func restoreTelemetryConfig(ctx context.Context, revModel *revisionConfigModel, prior fwtypes.ListNestedObjectValueOf[revisionConfigModel]) {
	desired := fwtypes.NewListNestedObjectValueOfNull[telemetryConfigModel](ctx)
	if !prior.IsNull() && !prior.IsUnknown() {
		if priorRev, d := prior.ToPtr(ctx); !d.HasError() && priorRev != nil && !priorRev.ServiceConfig.IsNull() && !priorRev.ServiceConfig.IsUnknown() {
			if priorSvc, d := priorRev.ServiceConfig.ToPtr(ctx); !d.HasError() && priorSvc != nil && !priorSvc.TelemetryConfig.IsUnknown() {
				desired = priorSvc.TelemetryConfig
			}
		}
	}

	if revModel.ServiceConfig.IsNull() || revModel.ServiceConfig.IsUnknown() {
		return
	}
	svc, d := revModel.ServiceConfig.ToPtr(ctx)
	if d.HasError() || svc == nil {
		return
	}
	svc.TelemetryConfig = desired
	revModel.ServiceConfig = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, svc)
}

type functionResourceModel struct {
	framework.WithRegionModel
	ARN                 types.String                                         `tfsdk:"arn"`
	DomainName          types.String                                         `tfsdk:"domain_name"`
	EndpointConfig      fwtypes.ListNestedObjectValueOf[endpointConfigModel] `tfsdk:"endpoint_config"`
	FunctionName        types.String                                         `tfsdk:"function_name"`
	ID                  types.String                                         `tfsdk:"id"`
	LatestRevisionID    types.String                                         `tfsdk:"latest_revision_id"`
	RevisionConfig      fwtypes.ListNestedObjectValueOf[revisionConfigModel] `tfsdk:"revision_config"`
	RegionalDomainNames fwtypes.MapOfString                                  `tfsdk:"regional_domain_names"`
	State               types.String                                         `tfsdk:"state"`
	StateReason         types.String                                         `tfsdk:"state_reason"`
	Tags                tftags.Map                                           `tfsdk:"tags"`
	TagsAll             tftags.Map                                           `tfsdk:"tags_all"`
	Timeouts            timeouts.Value                                       `tfsdk:"timeouts"`
}

type revisionConfigModel struct {
	Description   types.String                                        `tfsdk:"description"`
	KMSKeyARN     fwtypes.ARN                                         `tfsdk:"kms_key_arn"`
	BuildConfig   fwtypes.ListNestedObjectValueOf[buildConfigModel]   `tfsdk:"build_config"`
	ServiceConfig fwtypes.ListNestedObjectValueOf[serviceConfigModel] `tfsdk:"service_config"`
}

type buildConfigModel struct {
	CodeConfig    fwtypes.ListNestedObjectValueOf[codeConfigModel]    `tfsdk:"code_config"`
	RuntimeConfig fwtypes.ListNestedObjectValueOf[runtimeConfigModel] `tfsdk:"runtime_config"`
}

type codeConfigModel struct {
	S3Object fwtypes.ListNestedObjectValueOf[s3ObjectModel] `tfsdk:"s3_object"`
}

type s3ObjectModel struct {
	Bucket    types.String `tfsdk:"bucket"`
	Key       types.String `tfsdk:"key"`
	VersionID types.String `tfsdk:"version_id"`
}

type runtimeConfigModel struct {
	Runtime types.String `tfsdk:"runtime"`
}

type serviceConfigModel struct {
	ExecutionRoleARN             fwtypes.ARN                                           `tfsdk:"execution_role_arn"`
	TimeoutSeconds               types.Int64                                           `tfsdk:"timeout_seconds"`
	MaxConcurrencyPerEnvironment types.Int64                                           `tfsdk:"max_concurrency_per_environment"`
	EnvironmentVariables         fwtypes.MapOfString                                   `tfsdk:"environment_variables"`
	TelemetryConfig              fwtypes.ListNestedObjectValueOf[telemetryConfigModel] `tfsdk:"telemetry_config"`
}

type telemetryConfigModel struct {
	LoggingConfig fwtypes.ListNestedObjectValueOf[loggingConfigModel] `tfsdk:"logging_config"`
}

type loggingConfigModel struct {
	ApplicationLogLevel fwtypes.StringEnum[awstypes.ApplicationLogLevel] `tfsdk:"application_log_level"`
	LogGroup            types.String                                     `tfsdk:"log_group"`
	SystemLogLevel      fwtypes.StringEnum[awstypes.SystemLogLevel]      `tfsdk:"system_log_level"`
}

type endpointConfigModel struct {
	EndpointName       types.String                                    `tfsdk:"endpoint_name"`
	Description        types.String                                    `tfsdk:"description"`
	EndpointType       fwtypes.StringEnum[awstypes.EndpointType]       `tfsdk:"endpoint_type"`
	AuthType           fwtypes.StringEnum[awstypes.AuthType]           `tfsdk:"auth_type"`
	AutoDeploymentMode fwtypes.StringEnum[awstypes.AutoDeploymentMode] `tfsdk:"auto_deployment_mode"`
	Regions            fwtypes.SetOfString                             `tfsdk:"regions"`
	ScalingConfig      fwtypes.ObjectValueOf[scalingConfigModel]       `tfsdk:"scaling_config"`
	ThrottleConfig     fwtypes.ObjectValueOf[throttleConfigModel]      `tfsdk:"throttle_config"`
}
