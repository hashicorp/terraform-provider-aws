// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package sagemaker

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sagemaker"
	awstypes "github.com/aws/aws-sdk-go-v2/service/sagemaker/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	// Default create/update/delete timeouts. These are only the fallback values for the
	// resource's user-overridable timeouts block, not fixed limits.
	inferenceComponentInServiceDefaultTimeout = 30 * time.Minute
	inferenceComponentDeletedDefaultTimeout   = 30 * time.Minute
)

// @SDKResource("aws_sagemaker_inference_component", name="Inference Component")
// @Tags(identifierAttribute="arn")
func resourceInferenceComponent() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceInferenceComponentCreate,
		ReadWithoutTimeout:   resourceInferenceComponentRead,
		UpdateWithoutTimeout: resourceInferenceComponentUpdate,
		DeleteWithoutTimeout: resourceInferenceComponentDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(inferenceComponentInServiceDefaultTimeout),
			Update: schema.DefaultTimeout(inferenceComponentInServiceDefaultTimeout),
			Delete: schema.DefaultTimeout(inferenceComponentDeletedDefaultTimeout),
		},

		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
				names.AttrARN: {
					Type:     schema.TypeString,
					Computed: true,
				},
				names.AttrCreationTime: {
					Type:     schema.TypeString,
					Computed: true,
				},
				"endpoint_arn": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"failure_reason": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"last_modified_time": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"deployment_config": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"auto_rollback_configuration": {
								Type:     schema.TypeList,
								Optional: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"alarms": {
											Type:     schema.TypeList,
											Optional: true,
											MaxItems: 10,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"alarm_name": {
														Type:     schema.TypeString,
														Required: true,
													},
												},
											},
										},
									},
								},
							},
							"rolling_update_policy": {
								Type:     schema.TypeList,
								Required: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"maximum_batch_size": {
											Type:     schema.TypeList,
											Required: true,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													names.AttrType: {
														Type:             schema.TypeString,
														Required:         true,
														ValidateDiagFunc: enum.Validate[awstypes.InferenceComponentCapacitySizeType](),
													},
													names.AttrValue: {
														Type:         schema.TypeInt,
														Required:     true,
														ValidateFunc: validation.IntAtLeast(1),
													},
												},
											},
										},
										"maximum_execution_timeout_in_seconds": {
											Type:         schema.TypeInt,
											Optional:     true,
											ValidateFunc: validation.IntBetween(600, 28800),
										},
										"rollback_maximum_batch_size": {
											Type:     schema.TypeList,
											Optional: true,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													names.AttrType: {
														Type:             schema.TypeString,
														Required:         true,
														ValidateDiagFunc: enum.Validate[awstypes.InferenceComponentCapacitySizeType](),
													},
													names.AttrValue: {
														Type:         schema.TypeInt,
														Required:     true,
														ValidateFunc: validation.IntAtLeast(1),
													},
												},
											},
										},
										"wait_interval_in_seconds": {
											Type:         schema.TypeInt,
											Required:     true,
											ValidateFunc: validation.IntBetween(0, 3600),
										},
									},
								},
							},
						},
					},
				},
				"endpoint_name": {
					Type:         schema.TypeString,
					Required:     true,
					ForceNew:     true,
					ValidateFunc: validName,
				},
				names.AttrName: {
					Type:         schema.TypeString,
					Required:     true,
					ForceNew:     true,
					ValidateFunc: validName,
				},
				"runtime_config": {
					Type:     schema.TypeList,
					Optional: true,
					Computed: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"copy_count": {
								Type:         schema.TypeInt,
								Required:     true,
								ValidateFunc: validation.IntAtLeast(0),
							},
							"current_copy_count": {
								Type:     schema.TypeInt,
								Computed: true,
							},
							"desired_copy_count": {
								Type:     schema.TypeInt,
								Computed: true,
							},
							"placement_status": {
								Type:     schema.TypeList,
								Computed: true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"current_copy_count": {
											Type:     schema.TypeInt,
											Computed: true,
										},
										names.AttrInstanceType: {
											Type:     schema.TypeString,
											Computed: true,
										},
									},
								},
							},
						},
					},
				},
				"specification": {
					Type:         schema.TypeList,
					Optional:     true,
					MaxItems:     1,
					ExactlyOneOf: []string{"specification", "specifications"},
					Elem: &schema.Resource{
						Schema: inferenceComponentSpecificationSchema(),
					},
				},
				"specifications": {
					Type:         schema.TypeList,
					Optional:     true,
					MinItems:     2,
					ExactlyOneOf: []string{"specification", "specifications"},
					Elem: &schema.Resource{
						Schema: inferenceComponentSpecificationSchema(),
					},
				},
				names.AttrStatus: {
					Type:     schema.TypeString,
					Computed: true,
				},
				names.AttrTags:    tftags.TagsSchema(),
				names.AttrTagsAll: tftags.TagsSchemaComputed(),
				// variant_name is conditionally required: the CreateInferenceComponent API
				// rejects a null value for a standard inference component but forbids it for
				// an adapter component (one that sets base_inference_component_name), which
				// instead inherits — and echoes back on read — the base component's variant.
				// It is therefore Optional (requirement enforced server-side) and Computed
				// (absorbs the inherited value so the adapter case does not force replacement).
				"variant_name": {
					Type:         schema.TypeString,
					Optional:     true,
					Computed:     true,
					ForceNew:     true,
					ValidateFunc: validName,
				},
			}
		},
	}
}

func inferenceComponentSpecificationSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"base_inference_component_name": {
			Type:     schema.TypeString,
			Optional: true,
		},
		// An adapter inference component (one that sets base_inference_component_name)
		// omits this block and inherits the base component's compute resources, which the
		// API then echoes back on read. Computed absorbs those inherited values so the
		// adapter case does not produce a perpetual diff.
		"compute_resource_requirements": {
			Type:     schema.TypeList,
			Optional: true,
			Computed: true,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"max_memory_required_in_mb": {
						Type:         schema.TypeInt,
						Optional:     true,
						ValidateFunc: validation.IntAtLeast(128),
					},
					"min_memory_required_in_mb": {
						Type:         schema.TypeInt,
						Required:     true,
						ValidateFunc: validation.IntAtLeast(128),
					},
					"number_of_accelerator_devices_required": {
						Type:         schema.TypeFloat,
						Optional:     true,
						ValidateFunc: validation.FloatAtLeast(1),
					},
					"number_of_cpu_cores_required": {
						Type:         schema.TypeFloat,
						Optional:     true,
						ValidateFunc: validation.FloatAtLeast(0.25),
					},
				},
			},
		},
		"container": {
			Type:     schema.TypeList,
			Optional: true,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"artifact_url": {
						Type:     schema.TypeString,
						Optional: true,
					},
					"container_metrics_config": {
						Type:     schema.TypeList,
						Optional: true,
						MaxItems: 1,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"metrics_endpoint": {
									Type:     schema.TypeList,
									Optional: true,
									MaxItems: 1,
									Elem: &schema.Resource{
										Schema: map[string]*schema.Schema{
											"metric_publish_frequency_in_seconds": {
												Type:         schema.TypeInt,
												Optional:     true,
												Default:      60,
												ValidateFunc: validation.IntInSlice([]int{10, 30, 60, 120, 180, 240, 300}),
											},
											"metrics_endpoint_path": {
												Type:     schema.TypeString,
												Required: true,
												ValidateFunc: validation.All(
													validation.StringLenBetween(1, 256),
													validation.StringMatch(regexache.MustCompile(`^/`), "must start with '/'"),
												),
											},
										},
									},
								},
							},
						},
					},
					names.AttrEnvironment: {
						Type:     schema.TypeMap,
						Optional: true,
						Elem:     &schema.Schema{Type: schema.TypeString},
					},
					"image": {
						Type:     schema.TypeString,
						Optional: true,
					},
					"resolved_image": {
						Type:     schema.TypeString,
						Computed: true,
					},
				},
			},
		},
		// The service may return a data_cache_config (e.g. caching enabled by default on
		// accelerator instances) even when the block is omitted, so it is Computed to
		// absorb that server-populated value without producing a perpetual diff.
		"data_cache_config": {
			Type:     schema.TypeList,
			Optional: true,
			Computed: true,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"enable_caching": {
						Type:     schema.TypeBool,
						Required: true,
					},
				},
			},
		},
		names.AttrInstanceType: {
			Type:             schema.TypeString,
			Optional:         true,
			ValidateDiagFunc: enum.Validate[awstypes.ProductionVariantInstanceType](),
		},
		"model_name": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"scheduling_config": {
			Type:     schema.TypeList,
			Optional: true,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"availability_zone_balance": {
						Type:     schema.TypeList,
						Optional: true,
						MaxItems: 1,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"enforcement_mode": {
									Type:             schema.TypeString,
									Required:         true,
									ValidateDiagFunc: enum.Validate[awstypes.AvailabilityZoneBalanceEnforcementMode](),
								},
								"max_imbalance": {
									Type:     schema.TypeInt,
									Optional: true,
								},
							},
						},
					},
					"placement_strategy": {
						Type:             schema.TypeString,
						Required:         true,
						ValidateDiagFunc: enum.Validate[awstypes.InferenceComponentPlacementStrategy](),
					},
				},
			},
		},
		"startup_parameters": {
			Type:     schema.TypeList,
			Optional: true,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"container_startup_health_check_timeout_in_seconds": {
						Type:         schema.TypeInt,
						Optional:     true,
						ValidateFunc: validation.IntBetween(60, 3600),
					},
					"model_data_download_timeout_in_seconds": {
						Type:         schema.TypeInt,
						Optional:     true,
						ValidateFunc: validation.IntBetween(60, 3600),
					},
				},
			},
		},
	}
}

func resourceInferenceComponentCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &sagemaker.CreateInferenceComponentInput{
		InferenceComponentName: aws.String(name),
		EndpointName:           aws.String(d.Get("endpoint_name").(string)),
		Tags:                   getTagsIn(ctx),
	}

	if v, ok := d.GetOk("variant_name"); ok {
		input.VariantName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("runtime_config"); ok && len(v.([]any)) > 0 {
		input.RuntimeConfig = expandInferenceComponentRuntimeConfig(v.([]any))
	}

	if v, ok := d.GetOk("specification"); ok && len(v.([]any)) > 0 {
		input.Specification = expandInferenceComponentSpecification(v.([]any)[0].(map[string]any))
	}

	if v, ok := d.GetOk("specifications"); ok && len(v.([]any)) > 0 {
		input.Specifications = expandInferenceComponentSpecifications(v.([]any))
	}

	_, err := conn.CreateInferenceComponent(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating SageMaker Inference Component (%s): %s", name, err)
	}

	d.SetId(name)

	if err := waitInferenceComponentInService(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for SageMaker Inference Component (%s) to be in service: %s", d.Id(), err)
	}

	return append(diags, resourceInferenceComponentRead(ctx, d, meta)...)
}

func resourceInferenceComponentRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerClient(ctx)

	output, err := findInferenceComponentByName(ctx, conn, d.Id())
	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] SageMaker Inference Component (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SageMaker Inference Component (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, output.InferenceComponentArn)
	d.Set(names.AttrName, output.InferenceComponentName)
	d.Set("endpoint_arn", output.EndpointArn)
	d.Set("endpoint_name", output.EndpointName)
	d.Set("variant_name", output.VariantName)
	d.Set(names.AttrStatus, string(output.InferenceComponentStatus))
	d.Set("failure_reason", output.FailureReason)

	if output.CreationTime != nil {
		d.Set(names.AttrCreationTime, aws.ToTime(output.CreationTime).Format(time.RFC3339))
	}

	if output.LastModifiedTime != nil {
		d.Set("last_modified_time", aws.ToTime(output.LastModifiedTime).Format(time.RFC3339))
	}

	if err := d.Set("runtime_config", flattenInferenceComponentRuntimeConfigSummary(output.RuntimeConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting runtime_config: %s", err)
	}

	if output.Specification != nil {
		if err := d.Set("specification", []any{flattenInferenceComponentSpecificationSummary(output.Specification)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting specification: %s", err)
		}
		d.Set("specifications", nil)
	} else if output.Specifications != nil {
		d.Set("specification", nil)
		if err := d.Set("specifications", flattenInferenceComponentSpecificationSummaries(output.Specifications)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting specifications: %s", err)
		}
	} else {
		d.Set("specification", nil)
		d.Set("specifications", nil)
	}

	// deployment_config is an update-only rollout directive: the CreateInferenceComponent
	// API does not accept it, and DescribeInferenceComponent only reports it via
	// LastDeploymentConfig after an update has occurred. To avoid perpetual diffs on
	// resources that were never updated, we preserve the configured value in state rather
	// than refreshing it from the API.

	return diags
}

func resourceInferenceComponentUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := &sagemaker.UpdateInferenceComponentInput{
			InferenceComponentName: aws.String(d.Id()),
		}

		// UpdateInferenceComponent requires RuntimeConfig or a Specification; it rejects a
		// request carrying only the name or only a DeploymentConfig. Track whether one of
		// the required fields is populated so we skip a no-op/invalid API call (e.g. when
		// the sole change is clearing an optional block).
		var hasSpecOrRuntime bool

		if d.HasChange("runtime_config") {
			if v, ok := d.GetOk("runtime_config"); ok && len(v.([]any)) > 0 {
				input.RuntimeConfig = expandInferenceComponentRuntimeConfig(v.([]any))
				hasSpecOrRuntime = true
			}
		}

		if d.HasChange("specification") {
			if v, ok := d.GetOk("specification"); ok && len(v.([]any)) > 0 {
				input.Specification = expandInferenceComponentSpecification(v.([]any)[0].(map[string]any))
				hasSpecOrRuntime = true
			}
		}

		if d.HasChange("specifications") {
			if v, ok := d.GetOk("specifications"); ok && len(v.([]any)) > 0 {
				input.Specifications = expandInferenceComponentSpecifications(v.([]any))
				hasSpecOrRuntime = true
			}
		}

		// deployment_config only governs how a spec/runtime change rolls out, so it is
		// attached to the request but never triggers one on its own.
		if d.HasChange("deployment_config") {
			if v, ok := d.GetOk("deployment_config"); ok && len(v.([]any)) > 0 {
				input.DeploymentConfig = expandInferenceComponentDeploymentConfig(v.([]any))
			}
		}

		if hasSpecOrRuntime {
			_, err := conn.UpdateInferenceComponent(ctx, input)
			if err != nil {
				return sdkdiag.AppendErrorf(diags, "updating SageMaker Inference Component (%s): %s", d.Id(), err)
			}

			if err := waitInferenceComponentInService(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
				return sdkdiag.AppendErrorf(diags, "waiting for SageMaker Inference Component (%s) to be in service: %s", d.Id(), err)
			}
		}
	}

	return append(diags, resourceInferenceComponentRead(ctx, d, meta)...)
}

func resourceInferenceComponentDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerClient(ctx)

	log.Printf("[INFO] Deleting SageMaker Inference Component: %s", d.Id())
	// A delete may be rejected with a transient conflict: the component still has an
	// adapter component attached and being torn down ("is a base inference component
	// for one or more associated"), or it is mid-transition (Creating/Updating). Retry
	// only those cases. A component already in DELETE_IN_PROGRESS is treated as success
	// below rather than retried, since the delete has effectively been accepted.
	_, err := tfresource.RetryWhen(ctx, d.Timeout(schema.TimeoutDelete),
		func(ctx context.Context) (any, error) {
			return conn.DeleteInferenceComponent(ctx, &sagemaker.DeleteInferenceComponentInput{
				InferenceComponentName: aws.String(d.Id()),
			})
		},
		func(err error) (bool, error) {
			if tfawserr.ErrMessageContains(err, ErrCodeValidationException, "is a base inference component") {
				return true, err
			}
			return false, err
		})

	// Already gone, or a concurrent/prior delete is already in progress. In both cases
	// the component will be (or already is being) removed, so fall through to the wait.
	if tfawserr.ErrMessageContains(err, ErrCodeValidationException, "Could not find inference component") ||
		tfawserr.ErrMessageContains(err, ErrCodeValidationException, "DELETE_IN_PROGRESS") {
		err = nil
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting SageMaker Inference Component (%s): %s", d.Id(), err)
	}

	if err := waitInferenceComponentDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for SageMaker Inference Component (%s) to be deleted: %s", d.Id(), err)
	}

	return diags
}

func findInferenceComponentByName(ctx context.Context, conn *sagemaker.Client, name string) (*sagemaker.DescribeInferenceComponentOutput, error) {
	input := &sagemaker.DescribeInferenceComponentInput{
		InferenceComponentName: aws.String(name),
	}

	output, err := conn.DescribeInferenceComponent(ctx, input)

	if tfawserr.ErrMessageContains(err, ErrCodeValidationException, "Could not find inference component") {
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

	if output.InferenceComponentStatus == awstypes.InferenceComponentStatusDeleting {
		return nil, &retry.NotFoundError{
			Message: string(output.InferenceComponentStatus),
		}
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
			return nil, "", err
		}

		return output, string(output.InferenceComponentStatus), nil
	}
}

// statusInferenceComponentForDeletion reports the real status, including Deleting,
// rather than collapsing Deleting to not-found the way findInferenceComponentByName
// does for drift detection. This lets the delete waiter block until the component is
// genuinely gone (a bare "Could not find" from Describe), which matters when a base
// component's deletion must wait for its adapter component to be fully removed first.
func statusInferenceComponentForDeletion(conn *sagemaker.Client, name string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		input := &sagemaker.DescribeInferenceComponentInput{
			InferenceComponentName: aws.String(name),
		}

		output, err := conn.DescribeInferenceComponent(ctx, input)

		if tfawserr.ErrMessageContains(err, ErrCodeValidationException, "Could not find inference component") {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.InferenceComponentStatus), nil
	}
}

func waitInferenceComponentInService(ctx context.Context, conn *sagemaker.Client, name string, timeout time.Duration) error {
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
		return err
	}

	return err
}

func waitInferenceComponentDeleted(ctx context.Context, conn *sagemaker.Client, name string, timeout time.Duration) error {
	// Use the deletion-aware status refresh so the wait blocks through the Deleting
	// state until the component is actually gone, rather than returning as soon as
	// deletion begins.
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.InferenceComponentStatusInService, awstypes.InferenceComponentStatusCreating, awstypes.InferenceComponentStatusUpdating, awstypes.InferenceComponentStatusFailed, awstypes.InferenceComponentStatusDeleting),
		Target:  []string{},
		Refresh: statusInferenceComponentForDeletion(conn, name),
		Timeout: timeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)
	return err
}

// Expand functions

func expandInferenceComponentRuntimeConfig(tfList []any) *awstypes.InferenceComponentRuntimeConfig {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]any)
	return &awstypes.InferenceComponentRuntimeConfig{
		CopyCount: aws.Int32(int32(tfMap["copy_count"].(int))),
	}
}

func expandInferenceComponentSpecification(tfMap map[string]any) *awstypes.InferenceComponentSpecification {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.InferenceComponentSpecification{}

	if v, ok := tfMap["base_inference_component_name"].(string); ok && v != "" {
		apiObject.BaseInferenceComponentName = aws.String(v)
	}

	if v, ok := tfMap["model_name"].(string); ok && v != "" {
		apiObject.ModelName = aws.String(v)
	}

	if v, ok := tfMap[names.AttrInstanceType].(string); ok && v != "" {
		apiObject.InstanceType = awstypes.ProductionVariantInstanceType(v)
	}

	if v, ok := tfMap["compute_resource_requirements"].([]any); ok && len(v) > 0 && v[0] != nil {
		apiObject.ComputeResourceRequirements = expandInferenceComponentComputeResourceRequirements(v[0].(map[string]any))
	}

	if v, ok := tfMap["container"].([]any); ok && len(v) > 0 && v[0] != nil {
		apiObject.Container = expandInferenceComponentContainerSpecification(v[0].(map[string]any))
	}

	if v, ok := tfMap["data_cache_config"].([]any); ok && len(v) > 0 && v[0] != nil {
		apiObject.DataCacheConfig = expandInferenceComponentDataCacheConfig(v[0].(map[string]any))
	}

	if v, ok := tfMap["scheduling_config"].([]any); ok && len(v) > 0 && v[0] != nil {
		apiObject.SchedulingConfig = expandInferenceComponentSchedulingConfig(v[0].(map[string]any))
	}

	if v, ok := tfMap["startup_parameters"].([]any); ok && len(v) > 0 && v[0] != nil {
		apiObject.StartupParameters = expandInferenceComponentStartupParameters(v[0].(map[string]any))
	}

	return apiObject
}

func expandInferenceComponentSpecifications(tfList []any) []awstypes.InferenceComponentSpecification {
	if len(tfList) == 0 {
		return nil
	}

	apiObjects := make([]awstypes.InferenceComponentSpecification, 0, len(tfList))
	for _, tfMapRaw := range tfList {
		if tfMapRaw == nil {
			continue
		}
		tfMap := tfMapRaw.(map[string]any)
		apiObject := expandInferenceComponentSpecification(tfMap)
		if apiObject != nil {
			apiObjects = append(apiObjects, *apiObject)
		}
	}

	return apiObjects
}

func expandInferenceComponentComputeResourceRequirements(tfMap map[string]any) *awstypes.InferenceComponentComputeResourceRequirements {
	apiObject := &awstypes.InferenceComponentComputeResourceRequirements{
		MinMemoryRequiredInMb: aws.Int32(int32(tfMap["min_memory_required_in_mb"].(int))),
	}

	if v, ok := tfMap["max_memory_required_in_mb"].(int); ok && v > 0 {
		apiObject.MaxMemoryRequiredInMb = aws.Int32(int32(v))
	}

	if v, ok := tfMap["number_of_accelerator_devices_required"].(float64); ok && v > 0 {
		apiObject.NumberOfAcceleratorDevicesRequired = aws.Float32(float32(v))
	}

	if v, ok := tfMap["number_of_cpu_cores_required"].(float64); ok && v > 0 {
		apiObject.NumberOfCpuCoresRequired = aws.Float32(float32(v))
	}

	return apiObject
}

func expandInferenceComponentContainerSpecification(tfMap map[string]any) *awstypes.InferenceComponentContainerSpecification {
	apiObject := &awstypes.InferenceComponentContainerSpecification{}

	if v, ok := tfMap["artifact_url"].(string); ok && v != "" {
		apiObject.ArtifactUrl = aws.String(v)
	}

	if v, ok := tfMap["image"].(string); ok && v != "" {
		apiObject.Image = aws.String(v)
	}

	if v, ok := tfMap[names.AttrEnvironment].(map[string]any); ok && len(v) > 0 {
		apiObject.Environment = flex.ExpandStringValueMap(v)
	}

	if v, ok := tfMap["container_metrics_config"].([]any); ok && len(v) > 0 && v[0] != nil {
		apiObject.ContainerMetricsConfig = expandInferenceComponentContainerMetricsConfig(v[0].(map[string]any))
	}

	return apiObject
}

func expandInferenceComponentContainerMetricsConfig(tfMap map[string]any) *awstypes.ContainerMetricsConfig {
	apiObject := &awstypes.ContainerMetricsConfig{}

	if v, ok := tfMap["metrics_endpoint"].([]any); ok && len(v) > 0 && v[0] != nil {
		endpointMap := v[0].(map[string]any)
		metricsEndpoint := awstypes.MetricsEndpoint{
			MetricsEndpointPath: aws.String(endpointMap["metrics_endpoint_path"].(string)),
		}
		if freq, ok := endpointMap["metric_publish_frequency_in_seconds"].(int); ok && freq > 0 {
			metricsEndpoint.MetricPublishFrequencyInSeconds = int32(freq)
		}
		apiObject.MetricsEndpoints = []awstypes.MetricsEndpoint{metricsEndpoint}
	}

	return apiObject
}

func expandInferenceComponentDataCacheConfig(tfMap map[string]any) *awstypes.InferenceComponentDataCacheConfig {
	return &awstypes.InferenceComponentDataCacheConfig{
		EnableCaching: aws.Bool(tfMap["enable_caching"].(bool)),
	}
}

func expandInferenceComponentSchedulingConfig(tfMap map[string]any) *awstypes.InferenceComponentSchedulingConfig {
	apiObject := &awstypes.InferenceComponentSchedulingConfig{
		PlacementStrategy: awstypes.InferenceComponentPlacementStrategy(tfMap["placement_strategy"].(string)),
	}

	if v, ok := tfMap["availability_zone_balance"].([]any); ok && len(v) > 0 && v[0] != nil {
		azMap := v[0].(map[string]any)
		azBalance := &awstypes.InferenceComponentAvailabilityZoneBalance{
			EnforcementMode: awstypes.AvailabilityZoneBalanceEnforcementMode(azMap["enforcement_mode"].(string)),
		}
		if maxImbalance, ok := azMap["max_imbalance"].(int); ok && maxImbalance > 0 {
			azBalance.MaxImbalance = aws.Int32(int32(maxImbalance))
		}
		apiObject.AvailabilityZoneBalance = azBalance
	}

	return apiObject
}

func expandInferenceComponentStartupParameters(tfMap map[string]any) *awstypes.InferenceComponentStartupParameters {
	apiObject := &awstypes.InferenceComponentStartupParameters{}

	if v, ok := tfMap["container_startup_health_check_timeout_in_seconds"].(int); ok && v > 0 {
		apiObject.ContainerStartupHealthCheckTimeoutInSeconds = aws.Int32(int32(v))
	}

	if v, ok := tfMap["model_data_download_timeout_in_seconds"].(int); ok && v > 0 {
		apiObject.ModelDataDownloadTimeoutInSeconds = aws.Int32(int32(v))
	}

	return apiObject
}

func expandInferenceComponentDeploymentConfig(tfList []any) *awstypes.InferenceComponentDeploymentConfig {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]any)
	apiObject := &awstypes.InferenceComponentDeploymentConfig{}

	if v, ok := tfMap["rolling_update_policy"].([]any); ok && len(v) > 0 && v[0] != nil {
		apiObject.RollingUpdatePolicy = expandInferenceComponentRollingUpdatePolicy(v[0].(map[string]any))
	}

	if v, ok := tfMap["auto_rollback_configuration"].([]any); ok && len(v) > 0 && v[0] != nil {
		apiObject.AutoRollbackConfiguration = expandInferenceComponentAutoRollbackConfig(v[0].(map[string]any))
	}

	return apiObject
}

func expandInferenceComponentRollingUpdatePolicy(tfMap map[string]any) *awstypes.InferenceComponentRollingUpdatePolicy {
	apiObject := &awstypes.InferenceComponentRollingUpdatePolicy{
		WaitIntervalInSeconds: aws.Int32(int32(tfMap["wait_interval_in_seconds"].(int))),
	}

	if v, ok := tfMap["maximum_batch_size"].([]any); ok && len(v) > 0 && v[0] != nil {
		apiObject.MaximumBatchSize = expandInferenceComponentCapacitySize(v[0].(map[string]any))
	}

	if v, ok := tfMap["maximum_execution_timeout_in_seconds"].(int); ok && v > 0 {
		apiObject.MaximumExecutionTimeoutInSeconds = aws.Int32(int32(v))
	}

	if v, ok := tfMap["rollback_maximum_batch_size"].([]any); ok && len(v) > 0 && v[0] != nil {
		apiObject.RollbackMaximumBatchSize = expandInferenceComponentCapacitySize(v[0].(map[string]any))
	}

	return apiObject
}

func expandInferenceComponentCapacitySize(tfMap map[string]any) *awstypes.InferenceComponentCapacitySize {
	return &awstypes.InferenceComponentCapacitySize{
		Type:  awstypes.InferenceComponentCapacitySizeType(tfMap[names.AttrType].(string)),
		Value: aws.Int32(int32(tfMap[names.AttrValue].(int))),
	}
}

func expandInferenceComponentAutoRollbackConfig(tfMap map[string]any) *awstypes.AutoRollbackConfig {
	apiObject := &awstypes.AutoRollbackConfig{}

	if v, ok := tfMap["alarms"].([]any); ok && len(v) > 0 {
		apiObject.Alarms = expandAlarms(v)
	}

	return apiObject
}

// Flatten functions

func flattenInferenceComponentRuntimeConfigSummary(apiObject *awstypes.InferenceComponentRuntimeConfigSummary) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{
		"current_copy_count": aws.ToInt32(apiObject.CurrentCopyCount),
		"desired_copy_count": aws.ToInt32(apiObject.DesiredCopyCount),
		"placement_status":   flattenInferenceComponentPlacementStatus(apiObject.PlacementStatus),
	}

	if apiObject.DesiredCopyCount != nil {
		tfMap["copy_count"] = aws.ToInt32(apiObject.DesiredCopyCount)
	} else if apiObject.CurrentCopyCount != nil {
		tfMap["copy_count"] = aws.ToInt32(apiObject.CurrentCopyCount)
	}

	return []any{tfMap}
}

func flattenInferenceComponentPlacementStatus(apiObjects []awstypes.InferenceComponentPlacementStatus) []any {
	if len(apiObjects) == 0 {
		return nil
	}

	tfList := make([]any, 0, len(apiObjects))
	for _, apiObject := range apiObjects {
		tfList = append(tfList, map[string]any{
			"current_copy_count":   aws.ToInt32(apiObject.CurrentCopyCount),
			names.AttrInstanceType: string(apiObject.InstanceType),
		})
	}

	return tfList
}

func flattenInferenceComponentSpecificationSummary(apiObject *awstypes.InferenceComponentSpecificationSummary) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.BaseInferenceComponentName != nil {
		tfMap["base_inference_component_name"] = aws.ToString(apiObject.BaseInferenceComponentName)
	}

	if apiObject.ModelName != nil {
		tfMap["model_name"] = aws.ToString(apiObject.ModelName)
	}

	if apiObject.InstanceType != "" {
		tfMap[names.AttrInstanceType] = string(apiObject.InstanceType)
	}

	if apiObject.ComputeResourceRequirements != nil {
		tfMap["compute_resource_requirements"] = flattenInferenceComponentComputeResourceRequirements(apiObject.ComputeResourceRequirements)
	}

	if apiObject.Container != nil {
		tfMap["container"] = flattenInferenceComponentContainerSummary(apiObject.Container)
	}

	if apiObject.DataCacheConfig != nil {
		tfMap["data_cache_config"] = flattenInferenceComponentDataCacheConfigSummary(apiObject.DataCacheConfig)
	}

	if apiObject.SchedulingConfig != nil {
		tfMap["scheduling_config"] = flattenInferenceComponentSchedulingConfig(apiObject.SchedulingConfig)
	}

	if apiObject.StartupParameters != nil {
		tfMap["startup_parameters"] = flattenInferenceComponentStartupParameters(apiObject.StartupParameters)
	}

	return tfMap
}

func flattenInferenceComponentSpecificationSummaries(apiObjects []awstypes.InferenceComponentSpecificationSummary) []any {
	if len(apiObjects) == 0 {
		return nil
	}

	tfList := make([]any, 0, len(apiObjects))
	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenInferenceComponentSpecificationSummary(&apiObject))
	}

	return tfList
}

func flattenInferenceComponentComputeResourceRequirements(apiObject *awstypes.InferenceComponentComputeResourceRequirements) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{
		"min_memory_required_in_mb": aws.ToInt32(apiObject.MinMemoryRequiredInMb),
	}

	if apiObject.MaxMemoryRequiredInMb != nil {
		tfMap["max_memory_required_in_mb"] = aws.ToInt32(apiObject.MaxMemoryRequiredInMb)
	}

	if apiObject.NumberOfAcceleratorDevicesRequired != nil {
		tfMap["number_of_accelerator_devices_required"] = aws.ToFloat32(apiObject.NumberOfAcceleratorDevicesRequired)
	}

	if apiObject.NumberOfCpuCoresRequired != nil {
		tfMap["number_of_cpu_cores_required"] = aws.ToFloat32(apiObject.NumberOfCpuCoresRequired)
	}

	return []any{tfMap}
}

func flattenInferenceComponentContainerSummary(apiObject *awstypes.InferenceComponentContainerSpecificationSummary) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.DeployedImage != nil {
		if apiObject.DeployedImage.SpecifiedImage != nil {
			tfMap["image"] = aws.ToString(apiObject.DeployedImage.SpecifiedImage)
		}
		if apiObject.DeployedImage.ResolvedImage != nil {
			tfMap["resolved_image"] = aws.ToString(apiObject.DeployedImage.ResolvedImage)
		}
	}

	if apiObject.ArtifactUrl != nil {
		tfMap["artifact_url"] = aws.ToString(apiObject.ArtifactUrl)
	}

	if len(apiObject.Environment) > 0 {
		tfMap[names.AttrEnvironment] = apiObject.Environment
	}

	if apiObject.ContainerMetricsConfig != nil {
		tfMap["container_metrics_config"] = flattenInferenceComponentContainerMetricsConfig(apiObject.ContainerMetricsConfig)
	}

	return []any{tfMap}
}

func flattenInferenceComponentContainerMetricsConfig(apiObject *awstypes.ContainerMetricsConfig) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if len(apiObject.MetricsEndpoints) > 0 {
		endpoint := apiObject.MetricsEndpoints[0]
		endpointMap := map[string]any{
			"metrics_endpoint_path":               aws.ToString(endpoint.MetricsEndpointPath),
			"metric_publish_frequency_in_seconds": int(endpoint.MetricPublishFrequencyInSeconds),
		}
		tfMap["metrics_endpoint"] = []any{endpointMap}
	}

	return []any{tfMap}
}

func flattenInferenceComponentDataCacheConfigSummary(apiObject *awstypes.InferenceComponentDataCacheConfigSummary) []any {
	if apiObject == nil {
		return nil
	}

	return []any{map[string]any{
		"enable_caching": aws.ToBool(apiObject.EnableCaching),
	}}
}

func flattenInferenceComponentSchedulingConfig(apiObject *awstypes.InferenceComponentSchedulingConfig) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{
		"placement_strategy": string(apiObject.PlacementStrategy),
	}

	if apiObject.AvailabilityZoneBalance != nil {
		azMap := map[string]any{
			"enforcement_mode": string(apiObject.AvailabilityZoneBalance.EnforcementMode),
		}
		if apiObject.AvailabilityZoneBalance.MaxImbalance != nil {
			azMap["max_imbalance"] = aws.ToInt32(apiObject.AvailabilityZoneBalance.MaxImbalance)
		}
		tfMap["availability_zone_balance"] = []any{azMap}
	}

	return []any{tfMap}
}

func flattenInferenceComponentStartupParameters(apiObject *awstypes.InferenceComponentStartupParameters) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.ContainerStartupHealthCheckTimeoutInSeconds != nil {
		tfMap["container_startup_health_check_timeout_in_seconds"] = aws.ToInt32(apiObject.ContainerStartupHealthCheckTimeoutInSeconds)
	}

	if apiObject.ModelDataDownloadTimeoutInSeconds != nil {
		tfMap["model_data_download_timeout_in_seconds"] = aws.ToInt32(apiObject.ModelDataDownloadTimeoutInSeconds)
	}

	return []any{tfMap}
}
