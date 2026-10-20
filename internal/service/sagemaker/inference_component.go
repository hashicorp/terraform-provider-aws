// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package sagemaker

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sagemaker"
	awstypes "github.com/aws/aws-sdk-go-v2/service/sagemaker/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/float32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/float32planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
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

// @FrameworkResource("aws_sagemaker_inference_component", name="Inference Component")
// @Tags(identifierAttribute="arn")
// @IdentityAttribute("name")
// @Testing(importStateIdAttribute="name")
// @Testing(hasNoPreExistingResource=true)
func newInferenceComponentResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &inferenceComponentResource{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

type inferenceComponentResource struct {
	framework.ResourceWithModel[inferenceComponentResourceModel]
	framework.WithTimeouts
	framework.WithImportByIdentity
}

func (r *inferenceComponentResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrCreationTime: schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			"endpoint_arn": framework.ARNAttributeComputedOnly(),
			"endpoint_name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					inferenceComponentNameValidator,
				},
			},
			"failure_reason": schema.StringAttribute{
				Computed: true,
			},
			"last_modified_time": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					inferenceComponentNameValidator,
				},
			},
			names.AttrStatus: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.InferenceComponentStatus](),
				Computed:   true,
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			// variant_name is conditionally required: the CreateInferenceComponent API
			// rejects a null value for a standard inference component but forbids it for
			// an adapter inference component (one that sets base_inference_component_name),
			// which instead inherits -- and echoes back on read -- the base component's
			// variant. It is therefore Optional (requirement enforced server-side) and
			// Computed (absorbs the inherited value so the adapter case does not force
			// replacement or produce a perpetual diff).
			"variant_name": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					inferenceComponentNameValidator,
				},
			},
		},
		Blocks: map[string]schema.Block{
			"deployment_config": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[inferenceComponentDeploymentConfigModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"auto_rollback_configuration": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[inferenceComponentAutoRollbackConfigModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Blocks: map[string]schema.Block{
									"alarms": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[inferenceComponentAlarmModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(10),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"alarm_name": schema.StringAttribute{
													Required: true,
												},
											},
										},
									},
								},
							},
						},
						"rolling_update_policy": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[inferenceComponentRollingUpdatePolicyModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
								listvalidator.IsRequired(),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"maximum_execution_timeout_in_seconds": schema.Int32Attribute{
										Optional: true,
										Validators: []validator.Int32{
											int32validator.Between(600, 28800),
										},
									},
									"wait_interval_in_seconds": schema.Int32Attribute{
										Required: true,
										Validators: []validator.Int32{
											int32validator.Between(0, 3600),
										},
									},
								},
								Blocks: map[string]schema.Block{
									"maximum_batch_size": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[inferenceComponentCapacitySizeModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
											listvalidator.IsRequired(),
										},
										NestedObject: inferenceComponentCapacitySizeSchema(),
									},
									"rollback_maximum_batch_size": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[inferenceComponentCapacitySizeModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: inferenceComponentCapacitySizeSchema(),
									},
								},
							},
						},
					},
				},
			},
			"runtime_config": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[inferenceComponentRuntimeConfigModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"copy_count": schema.Int32Attribute{
							Required: true,
							Validators: []validator.Int32{
								int32validator.AtLeast(0),
							},
						},
						"current_copy_count": schema.Int32Attribute{
							Computed: true,
						},
						"desired_copy_count": schema.Int32Attribute{
							Computed: true,
						},
						// Modeled as a computed list attribute rather than a nested block.
						// A block's cardinality is controlled by configuration, and nothing
						// ever configures this read-only value, so a block would be planned
						// with a count of zero and then fail the apply with an
						// inconsistent-result error once the service returned entries.
						"placement_status": schema.ListAttribute{
							CustomType:  fwtypes.NewListNestedObjectTypeOf[inferenceComponentPlacementStatusModel](ctx),
							ElementType: fwtypes.NewObjectTypeOf[inferenceComponentPlacementStatusModel](ctx),
							Computed:    true,
						},
					},
				},
			},
			"specification": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[inferenceComponentSpecificationModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
					listvalidator.ExactlyOneOf(
						path.MatchRoot("specification"),
						path.MatchRoot("specifications"),
					),
				},
				NestedObject: inferenceComponentSpecificationSchema(ctx),
			},
			"specifications": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[inferenceComponentSpecificationModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtLeast(2),
				},
				NestedObject: inferenceComponentSpecificationSchema(ctx),
			},
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func inferenceComponentCapacitySizeSchema() schema.NestedBlockObject {
	return schema.NestedBlockObject{
		Attributes: map[string]schema.Attribute{
			names.AttrType: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.InferenceComponentCapacitySizeType](),
				Required:   true,
			},
			names.AttrValue: schema.Int32Attribute{
				Required: true,
				Validators: []validator.Int32{
					int32validator.AtLeast(1),
				},
			},
		},
	}
}

func inferenceComponentSpecificationSchema(ctx context.Context) schema.NestedBlockObject {
	return schema.NestedBlockObject{
		Attributes: map[string]schema.Attribute{
			"base_inference_component_name": schema.StringAttribute{
				Optional: true,
			},
			names.AttrInstanceType: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.ProductionVariantInstanceType](),
				Optional:   true,
			},
			"model_name": schema.StringAttribute{
				Optional: true,
			},
		},
		Blocks: map[string]schema.Block{
			// An adapter inference component (one that sets base_inference_component_name)
			// omits compute_resource_requirements and inherits the base component's compute
			// resources, which the API echoes back on read. A returned value is discarded
			// when the configuration omitted the block (see
			// preserveOmittedSpecificationBlocks), because a block's cardinality comes from
			// configuration. (The provider is served over Terraform protocol v5, which does
			// not allow nested attributes, so this must be a block.)
			"compute_resource_requirements": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[inferenceComponentComputeResourceRequirementsModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"max_memory_required_in_mb": schema.Int32Attribute{
							Optional: true,
							Computed: true,
							PlanModifiers: []planmodifier.Int32{
								int32planmodifier.UseStateForUnknown(),
							},
							Validators: []validator.Int32{
								int32validator.AtLeast(128),
							},
						},
						"min_memory_required_in_mb": schema.Int32Attribute{
							Required: true,
							Validators: []validator.Int32{
								int32validator.AtLeast(128),
							},
						},
						"number_of_accelerator_devices_required": schema.Float32Attribute{
							Optional: true,
							Computed: true,
							PlanModifiers: []planmodifier.Float32{
								float32planmodifier.UseStateForUnknown(),
							},
							Validators: []validator.Float32{
								float32validator.AtLeast(1),
							},
						},
						"number_of_cpu_cores_required": schema.Float32Attribute{
							Optional: true,
							Computed: true,
							PlanModifiers: []planmodifier.Float32{
								float32planmodifier.UseStateForUnknown(),
							},
							Validators: []validator.Float32{
								float32validator.AtLeast(0.25),
							},
						},
					},
				},
			},
			// The service may return a data_cache_config (e.g. caching enabled by default on
			// accelerator instances) even when it is omitted.
			"data_cache_config": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[inferenceComponentDataCacheConfigModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"enable_caching": schema.BoolAttribute{
							Required: true,
						},
					},
				},
			},
			"container": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[inferenceComponentContainerModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"artifact_url": schema.StringAttribute{
							Optional: true,
						},
						names.AttrEnvironment: schema.MapAttribute{
							CustomType:  fwtypes.MapOfStringType,
							ElementType: types.StringType,
							Optional:    true,
						},
						"image": schema.StringAttribute{
							Optional: true,
							Computed: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"resolved_image": schema.StringAttribute{
							Computed: true,
						},
					},
					Blocks: map[string]schema.Block{
						"container_metrics_config": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[inferenceComponentContainerMetricsConfigModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Blocks: map[string]schema.Block{
									"metrics_endpoint": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[inferenceComponentMetricsEndpointModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"metric_publish_frequency_in_seconds": schema.Int32Attribute{
													Optional: true,
													Computed: true,
													PlanModifiers: []planmodifier.Int32{
														int32planmodifier.UseStateForUnknown(),
													},
													Validators: []validator.Int32{
														int32validator.OneOf(10, 30, 60, 120, 180, 240, 300),
													},
												},
												"metrics_endpoint_path": schema.StringAttribute{
													Required: true,
													Validators: []validator.String{
														stringvalidator.LengthBetween(1, 256),
														stringvalidator.RegexMatches(regexache.MustCompile(`^/`), "must start with '/'"),
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
			"scheduling_config": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[inferenceComponentSchedulingConfigModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"placement_strategy": schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.InferenceComponentPlacementStrategy](),
							Required:   true,
						},
					},
					Blocks: map[string]schema.Block{
						"availability_zone_balance": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[inferenceComponentAvailabilityZoneBalanceModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"enforcement_mode": schema.StringAttribute{
										CustomType: fwtypes.StringEnumType[awstypes.AvailabilityZoneBalanceEnforcementMode](),
										Required:   true,
									},
									"max_imbalance": schema.Int32Attribute{
										Optional: true,
										Computed: true,
										PlanModifiers: []planmodifier.Int32{
											int32planmodifier.UseStateForUnknown(),
										},
									},
								},
							},
						},
					},
				},
			},
			"startup_parameters": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[inferenceComponentStartupParametersModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"container_startup_health_check_timeout_in_seconds": schema.Int32Attribute{
							Optional: true,
							Validators: []validator.Int32{
								int32validator.Between(60, 3600),
							},
						},
						"model_data_download_timeout_in_seconds": schema.Int32Attribute{
							Optional: true,
							Validators: []validator.Int32{
								int32validator.Between(60, 3600),
							},
						},
					},
				},
			},
		},
	}
}

func (r *inferenceComponentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().SageMakerClient(ctx)

	var plan inferenceComponentResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	var input sagemaker.CreateInferenceComponentInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input))
	if resp.Diagnostics.HasError() {
		return
	}

	input.Tags = getTagsIn(ctx)

	name := plan.InferenceComponentName.ValueString()
	_, err := conn.CreateInferenceComponent(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, name)
		return
	}

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	out, err := waitInferenceComponentInService(ctx, conn, name, createTimeout)
	if err != nil {
		// The component was created, so record its name to taint the resource rather
		// than leaving it untracked for a subsequent apply to recreate. Best effort:
		// a failure to write partial state must not mask the original error.
		resp.State.SetAttribute(ctx, path.Root(names.AttrName), name) //nolint:errcheck // Set 'name' so as to taint the resource.
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, name)
		return
	}

	r.flatten(ctx, out, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		resp.State.SetAttribute(ctx, path.Root(names.AttrName), name) //nolint:errcheck // Set 'name' so as to taint the resource.
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *inferenceComponentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().SageMakerClient(ctx)

	var state inferenceComponentResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	name := state.InferenceComponentName.ValueString()
	out, err := findInferenceComponentByName(ctx, conn, name)
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, name)
		return
	}

	r.flatten(ctx, out, &state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

// flatten maps the DescribeInferenceComponent output into the resource model.
//
// deployment_config is intentionally preserved from the prior model value rather
// than refreshed: it is an update-only rollout directive that the API does not
// return via DescribeInferenceComponent (only as LastDeploymentConfig after an
// update), so refreshing it would produce perpetual diffs.
func (r *inferenceComponentResource) flatten(ctx context.Context, out *sagemaker.DescribeInferenceComponentOutput, data *inferenceComponentResourceModel, diags *diag.Diagnostics) {
	deploymentConfig := data.DeploymentConfig
	priorSpecification, priorSpecifications := data.Specification, data.Specifications

	diags.Append(fwflex.Flatten(ctx, out, data)...)
	if diags.HasError() {
		return
	}

	data.DeploymentConfig = deploymentConfig

	preserveOmittedSpecificationBlocks(ctx, priorSpecification, &data.Specification, diags)
	if diags.HasError() {
		return
	}

	preserveOmittedSpecificationBlocks(ctx, priorSpecifications, &data.Specifications, diags)
}

// preserveOmittedSpecificationBlocks keeps compute_resource_requirements and
// data_cache_config absent when the configuration omitted them.
//
// Both are optional blocks that the service may return anyway: an adapter component
// inherits the base component's compute resources, and data caching can be enabled by
// default. A block's cardinality is controlled by configuration, so absorbing a returned
// value would change the planned count from zero to one and fail the apply with an
// inconsistent-result error, or produce a recurring removal diff on refresh. Marking the
// child attributes Computed does not help, because it is the block count that changes.
//
// Preservation only applies when a prior value exists to correlate against, so import
// (where there is no prior state) still populates the blocks from the API.
func preserveOmittedSpecificationBlocks(ctx context.Context, prior fwtypes.ListNestedObjectValueOf[inferenceComponentSpecificationModel], current *fwtypes.ListNestedObjectValueOf[inferenceComponentSpecificationModel], diags *diag.Diagnostics) {
	if prior.IsNull() || prior.IsUnknown() || current.IsNull() || current.IsUnknown() {
		return
	}

	priorSpecs, d := prior.ToSlice(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return
	}

	currentSpecs, d := current.ToSlice(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return
	}

	// Without a one-to-one correspondence the elements cannot be correlated.
	if len(priorSpecs) != len(currentSpecs) {
		return
	}

	for i, priorSpec := range priorSpecs {
		if priorSpec.ComputeResourceRequirements.IsNull() {
			currentSpecs[i].ComputeResourceRequirements = priorSpec.ComputeResourceRequirements
		}
		if priorSpec.DataCacheConfig.IsNull() {
			currentSpecs[i].DataCacheConfig = priorSpec.DataCacheConfig
		}

		preserveConfiguredEmptyEnvironment(ctx, priorSpec.Container, &currentSpecs[i].Container, diags)
		if diags.HasError() {
			return
		}
	}

	v, d := fwtypes.NewListNestedObjectValueOfSlice(ctx, currentSpecs, nil)
	diags.Append(d...)
	if diags.HasError() {
		return
	}

	*current = v
}

// preserveConfiguredEmptyEnvironment keeps an explicitly configured empty container
// environment from being replaced by null.
//
// environment is Optional without Computed, so its planned value is exactly what the
// configuration holds, and Terraform distinguishes an empty map from a null one. The
// service reports no entries for an empty environment, which flattens to null and would
// fail the apply with an inconsistent-result error. An omitted environment is null in the
// configuration and stays null.
func preserveConfiguredEmptyEnvironment(ctx context.Context, prior fwtypes.ListNestedObjectValueOf[inferenceComponentContainerModel], current *fwtypes.ListNestedObjectValueOf[inferenceComponentContainerModel], diags *diag.Diagnostics) {
	if prior.IsNull() || prior.IsUnknown() || current.IsNull() || current.IsUnknown() {
		return
	}

	priorContainer, d := prior.ToPtr(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return
	}

	currentContainer, d := current.ToPtr(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return
	}

	if priorContainer == nil || currentContainer == nil {
		return
	}

	// Only an explicitly configured empty map is restored. A prior value that holds
	// entries is left alone so genuine drift is still reported.
	priorEnvironment := priorContainer.Environment
	if priorEnvironment.IsNull() || priorEnvironment.IsUnknown() || priorEnvironment.Length(fwtypes.CollectionLengthUnhandledAsZero) > 0 {
		return
	}

	if !currentContainer.Environment.IsNull() {
		return
	}

	currentContainer.Environment = priorEnvironment

	v, d := fwtypes.NewListNestedObjectValueOfPtr(ctx, currentContainer)
	diags.Append(d...)
	if diags.HasError() {
		return
	}

	*current = v
}

func (r *inferenceComponentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().SageMakerClient(ctx)

	var plan, state inferenceComponentResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	name := plan.InferenceComponentName.ValueString()

	// UpdateInferenceComponent requires RuntimeConfig or a Specification; it rejects a
	// request that carries only a DeploymentConfig. A deployment_config-only change
	// therefore does not trigger an API call -- the directive only governs how a future
	// spec/runtime change rolls out, so the new value is simply persisted to state.
	//
	// The comparison is made on the expanded request structures rather than on the models.
	// Read-only attributes such as placement_status, the copy count summaries and
	// container.resolved_image are Computed without a plan modifier, so the Framework
	// marks them unknown in every plan. Comparing models would then report a change for
	// any update at all, including a tags-only one, and needlessly redeploy the model. The
	// request types carry only configurable fields, so read-only values are excluded
	// structurally rather than by a list that has to be maintained by hand.
	var input, stateInput sagemaker.UpdateInferenceComponentInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input), smerr.ID, name)
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, state, &stateInput), smerr.ID, name)
	if resp.Diagnostics.HasError() {
		return
	}

	specOrRuntimeChanged := !reflect.DeepEqual(input.RuntimeConfig, stateInput.RuntimeConfig) ||
		!reflect.DeepEqual(input.Specification, stateInput.Specification) ||
		!reflect.DeepEqual(input.Specifications, stateInput.Specifications)

	if specOrRuntimeChanged {
		_, err := conn.UpdateInferenceComponent(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, name)
			return
		}

		updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
		out, err := waitInferenceComponentInService(ctx, conn, name, updateTimeout)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, name)
			return
		}

		r.flatten(ctx, out, &plan, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}
	} else {
		// Nothing is sent to the API for a deployment_config-only change, but the
		// Framework marks every Computed attribute with a null configuration value as
		// unknown in the plan, and unknown values cannot be written to state. Refresh to
		// resolve them; otherwise Terraform rejects the result object after apply.
		out, err := findInferenceComponentByName(ctx, conn, name)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, name)
			return
		}

		r.flatten(ctx, out, &plan, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *inferenceComponentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().SageMakerClient(ctx)

	var state inferenceComponentResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	name := state.InferenceComponentName.ValueString()
	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)

	// A delete may be rejected with a transient conflict: the component is still a base
	// inference component for one or more associated adapter components that are being
	// torn down. Retry only that case. A component already in DELETE_IN_PROGRESS, or
	// already gone, is treated as success below rather than retried.
	_, err := tfresource.RetryWhen(ctx, deleteTimeout,
		func(ctx context.Context) (any, error) {
			return conn.DeleteInferenceComponent(ctx, &sagemaker.DeleteInferenceComponentInput{
				InferenceComponentName: aws.String(name),
			})
		},
		func(err error) (bool, error) {
			if tfawserr.ErrMessageContains(err, ErrCodeValidationException, "is a base inference component") {
				return true, err
			}
			return false, err
		})

	if tfawserr.ErrMessageContains(err, ErrCodeValidationException, "Could not find inference component") ||
		tfawserr.ErrMessageContains(err, ErrCodeValidationException, "DELETE_IN_PROGRESS") {
		err = nil
	}

	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, name)
		return
	}

	if _, err := waitInferenceComponentDeleted(ctx, conn, name, deleteTimeout); err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, name)
		return
	}
}

func findInferenceComponentByName(ctx context.Context, conn *sagemaker.Client, name string) (*sagemaker.DescribeInferenceComponentOutput, error) {
	input := sagemaker.DescribeInferenceComponentInput{
		InferenceComponentName: aws.String(name),
	}

	output, err := conn.DescribeInferenceComponent(ctx, &input)

	if tfawserr.ErrMessageContains(err, ErrCodeValidationException, "Could not find inference component") {
		return nil, smarterr.NewError(&retry.NotFoundError{
			LastError: err,
		})
	}

	if err != nil {
		return nil, smarterr.NewError(err)
	}

	if output == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	// Treat a component that is being deleted as gone for drift-detection purposes.
	if output.InferenceComponentStatus == awstypes.InferenceComponentStatusDeleting {
		return nil, smarterr.NewError(&retry.NotFoundError{
			Message: string(output.InferenceComponentStatus),
		})
	}

	return output, nil
}

func statusInferenceComponent(conn *sagemaker.Client, name string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findInferenceComponentByName(ctx, conn, name)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", smarterr.NewError(err)
		}

		return output, string(output.InferenceComponentStatus), nil
	}
}

// statusInferenceComponentForDeletion reports the real status, including Deleting,
// rather than collapsing Deleting to not-found the way findInferenceComponentByName
// does for drift detection. This lets the delete waiter block until the component is
// genuinely gone, which matters when a base component's deletion must wait for its
// adapter component to be fully removed first.
func statusInferenceComponentForDeletion(conn *sagemaker.Client, name string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := conn.DescribeInferenceComponent(ctx, &sagemaker.DescribeInferenceComponentInput{
			InferenceComponentName: aws.String(name),
		})

		if tfawserr.ErrMessageContains(err, ErrCodeValidationException, "Could not find inference component") {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", smarterr.NewError(err)
		}

		return output, string(output.InferenceComponentStatus), nil
	}
}

func waitInferenceComponentInService(ctx context.Context, conn *sagemaker.Client, name string, timeout time.Duration) (*sagemaker.DescribeInferenceComponentOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.InferenceComponentStatusCreating, awstypes.InferenceComponentStatusUpdating),
		Target:  enum.Slice(awstypes.InferenceComponentStatusInService),
		Refresh: statusInferenceComponent(conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if output, ok := outputRaw.(*sagemaker.DescribeInferenceComponentOutput); ok {
		if status, reason := output.InferenceComponentStatus, aws.ToString(output.FailureReason); status == awstypes.InferenceComponentStatusFailed && reason != "" {
			retry.SetLastError(err, errors.New(reason))
		}
		return output, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitInferenceComponentDeleted(ctx context.Context, conn *sagemaker.Client, name string, timeout time.Duration) (*sagemaker.DescribeInferenceComponentOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(
			awstypes.InferenceComponentStatusInService,
			awstypes.InferenceComponentStatusCreating,
			awstypes.InferenceComponentStatusUpdating,
			awstypes.InferenceComponentStatusFailed,
			awstypes.InferenceComponentStatusDeleting,
		),
		Target:  []string{},
		Refresh: statusInferenceComponentForDeletion(conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if output, ok := outputRaw.(*sagemaker.DescribeInferenceComponentOutput); ok {
		return output, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

var inferenceComponentNameValidator = stringvalidator.All(
	stringvalidator.LengthBetween(1, 63),
	stringvalidator.RegexMatches(
		regexache.MustCompile(`^[0-9A-Za-z](-*[0-9A-Za-z])*$`),
		"must contain only alphanumeric characters and hyphens, and must not begin or end with a hyphen",
	),
)

type inferenceComponentResourceModel struct {
	framework.WithRegionModel
	CreationTime             timetypes.RFC3339                                                        `tfsdk:"creation_time"`
	DeploymentConfig         fwtypes.ListNestedObjectValueOf[inferenceComponentDeploymentConfigModel] `tfsdk:"deployment_config"`
	EndpointARN              types.String                                                             `tfsdk:"endpoint_arn"`
	EndpointName             types.String                                                             `tfsdk:"endpoint_name"`
	FailureReason            types.String                                                             `tfsdk:"failure_reason"`
	InferenceComponentARN    types.String                                                             `tfsdk:"arn"`
	InferenceComponentName   types.String                                                             `tfsdk:"name"`
	InferenceComponentStatus fwtypes.StringEnum[awstypes.InferenceComponentStatus]                    `tfsdk:"status"`
	LastModifiedTime         timetypes.RFC3339                                                        `tfsdk:"last_modified_time"`
	RuntimeConfig            fwtypes.ListNestedObjectValueOf[inferenceComponentRuntimeConfigModel]    `tfsdk:"runtime_config"`
	Specification            fwtypes.ListNestedObjectValueOf[inferenceComponentSpecificationModel]    `tfsdk:"specification"`
	Specifications           fwtypes.ListNestedObjectValueOf[inferenceComponentSpecificationModel]    `tfsdk:"specifications"`
	Tags                     tftags.Map                                                               `tfsdk:"tags"`
	TagsAll                  tftags.Map                                                               `tfsdk:"tags_all"`
	Timeouts                 timeouts.Value                                                           `tfsdk:"timeouts"`
	VariantName              types.String                                                             `tfsdk:"variant_name"`
}

// inferenceComponentRuntimeConfigModel spans the API's asymmetric runtime types: the
// request carries only CopyCount, while the Describe response returns an
// InferenceComponentRuntimeConfigSummary that also reports CurrentCopyCount,
// DesiredCopyCount and PlacementStatus. The read-only values are exposed as computed
// attributes so practitioners can consume them.
type inferenceComponentRuntimeConfigModel struct {
	CopyCount        types.Int32                                                             `tfsdk:"copy_count"`
	CurrentCopyCount types.Int32                                                             `tfsdk:"current_copy_count"`
	DesiredCopyCount types.Int32                                                             `tfsdk:"desired_copy_count"`
	PlacementStatus  fwtypes.ListNestedObjectValueOf[inferenceComponentPlacementStatusModel] `tfsdk:"placement_status"`
}

var _ fwflex.Flattener = &inferenceComponentRuntimeConfigModel{}

// Flatten maps the InferenceComponentRuntimeConfigSummary (returned by Describe) onto
// the model. copy_count is mapped from DesiredCopyCount, the requested value, so the
// configured value is refreshed consistently (including import).
func (m *inferenceComponentRuntimeConfigModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics

	switch t := v.(type) {
	case awstypes.InferenceComponentRuntimeConfigSummary:
		m.CopyCount = types.Int32PointerValue(t.DesiredCopyCount)
		m.CurrentCopyCount = types.Int32PointerValue(t.CurrentCopyCount)
		m.DesiredCopyCount = types.Int32PointerValue(t.DesiredCopyCount)
		diags.Append(fwflex.Flatten(ctx, t.PlacementStatus, &m.PlacementStatus)...)
	default:
		diags.AddError("Unexpected Type", fmt.Sprintf("runtime_config flatten: %T", v))
	}

	return diags
}

type inferenceComponentPlacementStatusModel struct {
	CurrentCopyCount types.Int32                                                `tfsdk:"current_copy_count"`
	InstanceType     fwtypes.StringEnum[awstypes.ProductionVariantInstanceType] `tfsdk:"instance_type"`
}

type inferenceComponentSpecificationModel struct {
	BaseInferenceComponentName  types.String                                                                        `tfsdk:"base_inference_component_name"`
	ComputeResourceRequirements fwtypes.ListNestedObjectValueOf[inferenceComponentComputeResourceRequirementsModel] `tfsdk:"compute_resource_requirements"`
	Container                   fwtypes.ListNestedObjectValueOf[inferenceComponentContainerModel]                   `tfsdk:"container"`
	DataCacheConfig             fwtypes.ListNestedObjectValueOf[inferenceComponentDataCacheConfigModel]             `tfsdk:"data_cache_config"`
	InstanceType                fwtypes.StringEnum[awstypes.ProductionVariantInstanceType]                          `tfsdk:"instance_type"`
	ModelName                   types.String                                                                        `tfsdk:"model_name"`
	SchedulingConfig            fwtypes.ListNestedObjectValueOf[inferenceComponentSchedulingConfigModel]            `tfsdk:"scheduling_config"`
	StartupParameters           fwtypes.ListNestedObjectValueOf[inferenceComponentStartupParametersModel]           `tfsdk:"startup_parameters"`
}

type inferenceComponentComputeResourceRequirementsModel struct {
	MaxMemoryRequiredInMb              types.Int32   `tfsdk:"max_memory_required_in_mb"`
	MinMemoryRequiredInMb              types.Int32   `tfsdk:"min_memory_required_in_mb"`
	NumberOfAcceleratorDevicesRequired types.Float32 `tfsdk:"number_of_accelerator_devices_required"`
	NumberOfCpuCoresRequired           types.Float32 `tfsdk:"number_of_cpu_cores_required"`
}

type inferenceComponentContainerModel struct {
	ArtifactURL            types.String                                                                   `tfsdk:"artifact_url"`
	ContainerMetricsConfig fwtypes.ListNestedObjectValueOf[inferenceComponentContainerMetricsConfigModel] `tfsdk:"container_metrics_config"`
	Environment            fwtypes.MapOfString                                                            `tfsdk:"environment"`
	Image                  types.String                                                                   `tfsdk:"image"`
	ResolvedImage          types.String                                                                   `tfsdk:"resolved_image"`
}

var _ fwflex.Flattener = &inferenceComponentContainerModel{}

// Flatten maps the InferenceComponentContainerSpecificationSummary (returned by
// Describe) onto the model. The request type carries a flat Image, while the summary
// reports a nested DeployedImage{SpecifiedImage, ResolvedImage}; image is mapped from
// SpecifiedImage (the value the caller provided) and resolved_image from ResolvedImage.
// The symmetric fields are expanded via plain AutoFlex (no custom Expander is needed).
func (m *inferenceComponentContainerModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics

	switch t := v.(type) {
	case awstypes.InferenceComponentContainerSpecificationSummary:
		m.ArtifactURL = types.StringPointerValue(t.ArtifactUrl)

		if t.DeployedImage != nil {
			m.Image = types.StringPointerValue(t.DeployedImage.SpecifiedImage)
			m.ResolvedImage = types.StringPointerValue(t.DeployedImage.ResolvedImage)
		} else {
			m.Image = types.StringNull()
			m.ResolvedImage = types.StringNull()
		}

		if len(t.Environment) > 0 {
			diags.Append(fwflex.Flatten(ctx, t.Environment, &m.Environment)...)
		} else {
			m.Environment = fwtypes.NewMapValueOfNull[types.String](ctx)
		}

		if t.ContainerMetricsConfig != nil {
			diags.Append(fwflex.Flatten(ctx, t.ContainerMetricsConfig, &m.ContainerMetricsConfig)...)
		} else {
			m.ContainerMetricsConfig = fwtypes.NewListNestedObjectValueOfNull[inferenceComponentContainerMetricsConfigModel](ctx)
		}
	default:
		diags.AddError("Unexpected Type", fmt.Sprintf("container flatten: %T", v))
	}

	return diags
}

type inferenceComponentContainerMetricsConfigModel struct {
	MetricsEndpoints fwtypes.ListNestedObjectValueOf[inferenceComponentMetricsEndpointModel] `tfsdk:"metrics_endpoint"`
}

type inferenceComponentMetricsEndpointModel struct {
	MetricPublishFrequencyInSeconds types.Int32  `tfsdk:"metric_publish_frequency_in_seconds"`
	MetricsEndpointPath             types.String `tfsdk:"metrics_endpoint_path"`
}

type inferenceComponentDataCacheConfigModel struct {
	EnableCaching types.Bool `tfsdk:"enable_caching"`
}

type inferenceComponentSchedulingConfigModel struct {
	AvailabilityZoneBalance fwtypes.ListNestedObjectValueOf[inferenceComponentAvailabilityZoneBalanceModel] `tfsdk:"availability_zone_balance"`
	PlacementStrategy       fwtypes.StringEnum[awstypes.InferenceComponentPlacementStrategy]                `tfsdk:"placement_strategy"`
}

type inferenceComponentAvailabilityZoneBalanceModel struct {
	EnforcementMode fwtypes.StringEnum[awstypes.AvailabilityZoneBalanceEnforcementMode] `tfsdk:"enforcement_mode"`
	MaxImbalance    types.Int32                                                         `tfsdk:"max_imbalance"`
}

type inferenceComponentStartupParametersModel struct {
	ContainerStartupHealthCheckTimeoutInSeconds types.Int32 `tfsdk:"container_startup_health_check_timeout_in_seconds"`
	ModelDataDownloadTimeoutInSeconds           types.Int32 `tfsdk:"model_data_download_timeout_in_seconds"`
}

type inferenceComponentDeploymentConfigModel struct {
	AutoRollbackConfiguration fwtypes.ListNestedObjectValueOf[inferenceComponentAutoRollbackConfigModel]  `tfsdk:"auto_rollback_configuration"`
	RollingUpdatePolicy       fwtypes.ListNestedObjectValueOf[inferenceComponentRollingUpdatePolicyModel] `tfsdk:"rolling_update_policy"`
}

type inferenceComponentAutoRollbackConfigModel struct {
	Alarms fwtypes.ListNestedObjectValueOf[inferenceComponentAlarmModel] `tfsdk:"alarms"`
}

type inferenceComponentAlarmModel struct {
	AlarmName types.String `tfsdk:"alarm_name"`
}

type inferenceComponentRollingUpdatePolicyModel struct {
	MaximumBatchSize                 fwtypes.ListNestedObjectValueOf[inferenceComponentCapacitySizeModel] `tfsdk:"maximum_batch_size"`
	MaximumExecutionTimeoutInSeconds types.Int32                                                          `tfsdk:"maximum_execution_timeout_in_seconds"`
	RollbackMaximumBatchSize         fwtypes.ListNestedObjectValueOf[inferenceComponentCapacitySizeModel] `tfsdk:"rollback_maximum_batch_size"`
	WaitIntervalInSeconds            types.Int32                                                          `tfsdk:"wait_interval_in_seconds"`
}

type inferenceComponentCapacitySizeModel struct {
	Type  fwtypes.StringEnum[awstypes.InferenceComponentCapacitySizeType] `tfsdk:"type"`
	Value types.Int32                                                     `tfsdk:"value"`
}
