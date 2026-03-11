// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package ecs

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_ecs_capacity_provider", name="Capacity Provider")
// @Tags(identifierAttribute="arn")
// @ArnIdentity
// @V60SDKv2Fix
// @ArnFormat("capacity-provider/{name}")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/ecs/types;awstypes;awstypes.CapacityProvider")
func resourceCapacityProvider() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceCapacityProviderCreate,
		ReadWithoutTimeout:   resourceCapacityProviderRead,
		UpdateWithoutTimeout: resourceCapacityProviderUpdate,
		DeleteWithoutTimeout: resourceCapacityProviderDelete,

		CustomizeDiff: func(ctx context.Context, diff *schema.ResourceDiff, meta any) error {
			// Ensure exactly one of auto_scaling_group_provider or managed_instances_provider is specified
			asgProvider := diff.Get("auto_scaling_group_provider").([]any)
			managedProvider := diff.Get("managed_instances_provider").([]any)
			clusterName := diff.Get("cluster").(string)

			if len(asgProvider) == 0 && len(managedProvider) == 0 {
				return errors.New("exactly one of auto_scaling_group_provider or managed_instances_provider must be specified")
			} else if len(asgProvider) > 0 && len(managedProvider) > 0 {
				return errors.New("only one of auto_scaling_group_provider or managed_instances_provider must be specified")
			}

			// Validate cluster field requirements
			if len(managedProvider) > 0 {
				// cluster is required for Managed Instances CP
				if clusterName == "" {
					return errors.New("cluster is required when using managed_instances_provider")
				}
			} else if len(asgProvider) > 0 {
				// cluster must not be set for ASG CP
				if clusterName != "" {
					return errors.New("cluster must not be set when using auto_scaling_group_provider")
				}
			}

			return nil
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"auto_scaling_group_provider": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"auto_scaling_group_arn": {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: verify.ValidARN,
						},
						"managed_draining": {
							Type:             schema.TypeString,
							Optional:         true,
							Computed:         true,
							ValidateDiagFunc: enum.Validate[awstypes.ManagedDraining](),
						},
						"managed_scaling": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Optional: true,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"instance_warmup_period": {
										Type:         schema.TypeInt,
										Optional:     true,
										Computed:     true,
										ValidateFunc: validation.IntBetween(0, 10000),
									},
									"maximum_scaling_step_size": {
										Type:         schema.TypeInt,
										Optional:     true,
										Computed:     true,
										ValidateFunc: validation.IntBetween(1, 10000),
									},
									"minimum_scaling_step_size": {
										Type:         schema.TypeInt,
										Optional:     true,
										Computed:     true,
										ValidateFunc: validation.IntBetween(1, 10000),
									},
									names.AttrStatus: {
										Type:             schema.TypeString,
										Optional:         true,
										Computed:         true,
										ValidateDiagFunc: enum.Validate[awstypes.ManagedScalingStatus]()},
									"target_capacity": {
										Type:         schema.TypeInt,
										Optional:     true,
										Computed:     true,
										ValidateFunc: validation.IntBetween(1, 100),
									},
								},
							},
						},
						"managed_termination_protection": {
							Type:             schema.TypeString,
							Optional:         true,
							Computed:         true,
							ValidateDiagFunc: enum.Validate[awstypes.ManagedTerminationProtection](),
						},
					},
				},
			},
			"cluster": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validateClusterName,
			},
			"managed_instances_provider": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"infrastructure_optimization": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"scale_in_after": {
										Type:         schema.TypeInt,
										Optional:     true,
										ValidateFunc: validation.IntBetween(-1, 3600),
									},
								},
							},
						},
						"infrastructure_role_arn": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: verify.ValidARN,
						},
						"instance_launch_template": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Required: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"capacity_option_type": {
										Type:             schema.TypeString,
										Optional:         true,
										ForceNew:         true,
										Computed:         true,
										ValidateDiagFunc: enum.Validate[awstypes.CapacityOptionType](),
									},
									"ec2_instance_profile_arn": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: verify.ValidARN,
									},
									"instance_requirements": {
										Type:     schema.TypeList,
										MaxItems: 1,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"accelerator_count": {
													Type:     schema.TypeList,
													MaxItems: 1,
													Optional: true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															names.AttrMax: {
																Type:         schema.TypeInt,
																Optional:     true,
																ValidateFunc: validation.IntAtLeast(0),
															},
															names.AttrMin: {
																Type:         schema.TypeInt,
																Optional:     true,
																ValidateFunc: validation.IntAtLeast(0),
															},
														},
													},
												},
												"accelerator_manufacturers": {
													Type:     schema.TypeSet,
													Optional: true,
													Elem: &schema.Schema{
														Type:             schema.TypeString,
														ValidateDiagFunc: enum.Validate[awstypes.AcceleratorManufacturer](),
													},
												},
												"accelerator_names": {
													Type:     schema.TypeSet,
													Optional: true,
													Elem: &schema.Schema{
														Type:             schema.TypeString,
														ValidateDiagFunc: enum.Validate[awstypes.AcceleratorName](),
													},
												},
												"accelerator_total_memory_mib": {
													Type:     schema.TypeList,
													MaxItems: 1,
													Optional: true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															names.AttrMax: {
																Type:         schema.TypeInt,
																Optional:     true,
																ValidateFunc: validation.IntAtLeast(0),
															},
															names.AttrMin: {
																Type:         schema.TypeInt,
																Optional:     true,
																ValidateFunc: validation.IntAtLeast(0),
															},
														},
													},
												},
												"accelerator_types": {
													Type:     schema.TypeSet,
													Optional: true,
													Elem: &schema.Schema{
														Type:             schema.TypeString,
														ValidateDiagFunc: enum.Validate[awstypes.AcceleratorType](),
													},
												},
												"allowed_instance_types": {
													Type:     schema.TypeSet,
													Optional: true,
													MaxItems: 400,
													Elem: &schema.Schema{
														Type: schema.TypeString,
														ValidateFunc: validation.All(
															validation.StringLenBetween(1, 30),
															validation.StringMatch(regexache.MustCompile(`^[a-zA-Z0-9\.\*\-]+$`), "must contain only alphanumeric characters, dots, asterisks, and hyphens"),
														),
													},
												},
												"bare_metal": {
													Type:             schema.TypeString,
													Optional:         true,
													ValidateDiagFunc: enum.Validate[awstypes.BareMetal](),
												},
												"baseline_ebs_bandwidth_mbps": {
													Type:     schema.TypeList,
													MaxItems: 1,
													Optional: true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															names.AttrMax: {
																Type:         schema.TypeInt,
																Optional:     true,
																ValidateFunc: validation.IntAtLeast(0),
															},
															names.AttrMin: {
																Type:         schema.TypeInt,
																Optional:     true,
																ValidateFunc: validation.IntAtLeast(0),
															},
														},
													},
												},
												"burstable_performance": {
													Type:             schema.TypeString,
													Optional:         true,
													ValidateDiagFunc: enum.Validate[awstypes.BurstablePerformance](),
												},
												"cpu_manufacturers": {
													Type:     schema.TypeSet,
													Optional: true,
													Elem: &schema.Schema{
														Type:             schema.TypeString,
														ValidateDiagFunc: enum.Validate[awstypes.CpuManufacturer](),
													},
												},
												"excluded_instance_types": {
													Type:     schema.TypeSet,
													Optional: true,
													MaxItems: 400,
													Elem: &schema.Schema{
														Type: schema.TypeString,
														ValidateFunc: validation.All(
															validation.StringLenBetween(1, 30),
															validation.StringMatch(regexache.MustCompile(`^[a-zA-Z0-9\.\*\-]+$`), "must contain only alphanumeric characters, dots, asterisks, and hyphens"),
														),
													},
												},
												"instance_generations": {
													Type:     schema.TypeSet,
													Optional: true,
													Elem: &schema.Schema{
														Type:             schema.TypeString,
														ValidateDiagFunc: enum.Validate[awstypes.InstanceGeneration](),
													},
												},
												"local_storage": {
													Type:             schema.TypeString,
													Optional:         true,
													ValidateDiagFunc: enum.Validate[awstypes.LocalStorage](),
												},
												"local_storage_types": {
													Type:     schema.TypeSet,
													Optional: true,
													Elem: &schema.Schema{
														Type:             schema.TypeString,
														ValidateDiagFunc: enum.Validate[awstypes.LocalStorageType](),
													},
												},
												"max_spot_price_as_percentage_of_optimal_on_demand_price": {
													Type:         schema.TypeInt,
													Optional:     true,
													ValidateFunc: validation.IntAtLeast(0),
												},
												"memory_gib_per_vcpu": {
													Type:     schema.TypeList,
													MaxItems: 1,
													Optional: true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															names.AttrMax: {
																Type:         schema.TypeFloat,
																Optional:     true,
																ValidateFunc: validation.FloatAtLeast(0),
															},
															names.AttrMin: {
																Type:         schema.TypeFloat,
																Optional:     true,
																ValidateFunc: validation.FloatAtLeast(0),
															},
														},
													},
												},
												"memory_mib": {
													Type:     schema.TypeList,
													MaxItems: 1,
													Required: true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															names.AttrMax: {
																Type:         schema.TypeInt,
																Optional:     true,
																ValidateFunc: validation.IntAtLeast(1),
															},
															names.AttrMin: {
																Type:         schema.TypeInt,
																Required:     true,
																ValidateFunc: validation.IntAtLeast(1),
															},
														},
													},
												},
												"network_bandwidth_gbps": {
													Type:     schema.TypeList,
													MaxItems: 1,
													Optional: true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															names.AttrMax: {
																Type:         schema.TypeFloat,
																Optional:     true,
																ValidateFunc: validation.FloatAtLeast(0),
															},
															names.AttrMin: {
																Type:         schema.TypeFloat,
																Optional:     true,
																ValidateFunc: validation.FloatAtLeast(0),
															},
														},
													},
												},
												"network_interface_count": {
													Type:     schema.TypeList,
													MaxItems: 1,
													Optional: true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															names.AttrMax: {
																Type:         schema.TypeInt,
																Optional:     true,
																ValidateFunc: validation.IntAtLeast(1),
															},
															names.AttrMin: {
																Type:         schema.TypeInt,
																Optional:     true,
																ValidateFunc: validation.IntAtLeast(1),
															},
														},
													},
												},
												"on_demand_max_price_percentage_over_lowest_price": {
													Type:         schema.TypeInt,
													Optional:     true,
													ValidateFunc: validation.IntAtLeast(0),
												},
												"require_hibernate_support": {
													Type:     schema.TypeBool,
													Optional: true,
												},
												"spot_max_price_percentage_over_lowest_price": {
													Type:         schema.TypeInt,
													Optional:     true,
													ValidateFunc: validation.IntAtLeast(0),
												},
												"total_local_storage_gb": {
													Type:     schema.TypeList,
													MaxItems: 1,
													Optional: true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															names.AttrMax: {
																Type:         schema.TypeFloat,
																Optional:     true,
																ValidateFunc: validation.FloatAtLeast(0),
															},
															names.AttrMin: {
																Type:         schema.TypeFloat,
																Optional:     true,
																ValidateFunc: validation.FloatAtLeast(0),
															},
														},
													},
												},
												"vcpu_count": {
													Type:     schema.TypeList,
													MaxItems: 1,
													Required: true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															names.AttrMax: {
																Type:         schema.TypeInt,
																Optional:     true,
																ValidateFunc: validation.IntAtLeast(1),
															},
															names.AttrMin: {
																Type:         schema.TypeInt,
																Required:     true,
																ValidateFunc: validation.IntAtLeast(1),
															},
														},
													},
												},
											},
										},
									},
									"monitoring": {
										Type:             schema.TypeString,
										Optional:         true,
										ValidateDiagFunc: enum.Validate[awstypes.ManagedInstancesMonitoringOptions](),
									},
									names.AttrNetworkConfiguration: {
										Type:     schema.TypeList,
										MaxItems: 1,
										Required: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												names.AttrSecurityGroups: {
													Type:     schema.TypeSet,
													Optional: true,
													Elem: &schema.Schema{
														Type: schema.TypeString,
													},
												},
												names.AttrSubnets: {
													Type:     schema.TypeSet,
													Required: true,
													Elem: &schema.Schema{
														Type: schema.TypeString,
													},
												},
											},
										},
									},
									"storage_configuration": {
										Type:     schema.TypeList,
										MaxItems: 1,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"storage_size_gib": {
													Type:         schema.TypeInt,
													Required:     true,
													ValidateFunc: validation.IntAtLeast(1),
												},
											},
										},
									},
								},
							},
						},
						names.AttrPropagateTags: {
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: enum.Validate[awstypes.PropagateMITags](),
						},
					},
				},
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourceCapacityProviderCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ECSClient(ctx)
	partition := meta.(*conns.AWSClient).Partition(ctx)

	name := d.Get(names.AttrName).(string)
	input := ecs.CreateCapacityProviderInput{
		AutoScalingGroupProvider: expandAutoScalingGroupProviderCreate(d.Get("auto_scaling_group_provider")),
		ManagedInstancesProvider: expandManagedInstancesProviderCreate(d.Get("managed_instances_provider")),
		Name:                     aws.String(name),
		Tags:                     getTagsIn(ctx),
	}

	if v, ok := d.GetOk("cluster"); ok {
		input.Cluster = aws.String(v.(string))
	}

	output, err := conn.CreateCapacityProvider(ctx, &input)

	// Some partitions (e.g. ISO) may not support tag-on-create.
	if input.Tags != nil && errs.IsUnsupportedOperationInPartitionError(partition, err) {
		input.Tags = nil

		output, err = conn.CreateCapacityProvider(ctx, &input)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating ECS Capacity Provider (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.CapacityProvider.CapacityProviderArn))

	// For partitions not supporting tag-on-create, attempt tag after create.
	if tags := getTagsIn(ctx); input.Tags == nil && len(tags) > 0 {
		err := createTags(ctx, conn, d.Id(), tags)

		// If default tags only, continue. Otherwise, error.
		if v, ok := d.GetOk(names.AttrTags); (!ok || len(v.(map[string]any)) == 0) && errs.IsUnsupportedOperationInPartitionError(partition, err) {
			return append(diags, resourceCapacityProviderRead(ctx, d, meta)...)
		}

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "setting ECS Capacity Provider (%s) tags: %s", d.Id(), err)
		}
	}

	return append(diags, resourceCapacityProviderRead(ctx, d, meta)...)
}

func resourceCapacityProviderRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ECSClient(ctx)

	output, err := findCapacityProviderByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] ECS Capacity Provider (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ECS Capacity Provider (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, output.CapacityProviderArn)
	if err := d.Set("auto_scaling_group_provider", flattenAutoScalingGroupProvider(output.AutoScalingGroupProvider)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting auto_scaling_group_provider: %s", err)
	}
	d.Set("cluster", output.Cluster)
	if err := d.Set("managed_instances_provider", flattenManagedInstancesProvider(output.ManagedInstancesProvider)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting managed_instances_provider: %s", err)
	}
	d.Set(names.AttrName, output.Name)

	setTagsOut(ctx, output.Tags)

	return diags
}

func resourceCapacityProviderUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ECSClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := ecs.UpdateCapacityProviderInput{
			AutoScalingGroupProvider: expandAutoScalingGroupProviderUpdate(d.Get("auto_scaling_group_provider")),
			ManagedInstancesProvider: expandManagedInstancesProviderUpdate(d.Get("managed_instances_provider")),
			Name:                     aws.String(d.Get(names.AttrName).(string)),
		}

		if v, ok := d.GetOk("cluster"); ok {
			input.Cluster = aws.String(v.(string))
		}

		const (
			timeout = 10 * time.Minute
		)
		_, err := tfresource.RetryWhenIsA[any, *awstypes.UpdateInProgressException](ctx, timeout, func(ctx context.Context) (any, error) {
			return conn.UpdateCapacityProvider(ctx, &input)
		})

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating ECS Capacity Provider (%s): %s", d.Id(), err)
		}

		if _, err = waitCapacityProviderUpdated(ctx, conn, d.Id(), timeout); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for ECS Capacity Provider (%s) update: %s", d.Id(), err)
		}
	}

	return append(diags, resourceCapacityProviderRead(ctx, d, meta)...)
}

func resourceCapacityProviderDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ECSClient(ctx)

	log.Printf("[DEBUG] Deleting ECS Capacity Provider: %s", d.Id())
	input := ecs.DeleteCapacityProviderInput{
		CapacityProvider: aws.String(d.Id()),
	}
	_, err := conn.DeleteCapacityProvider(ctx, &input)

	// "An error occurred (ClientException) when calling the DeleteCapacityProvider operation: The specified capacity provider does not exist. Specify a valid name or ARN and try again."
	if errs.IsAErrorMessageContains[*awstypes.ClientException](err, "capacity provider does not exist") {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting ECS Capacity Provider (%s): %s", d.Id(), err)
	}

	const (
		timeout = 20 * time.Minute
	)
	if _, err := waitCapacityProviderDeleted(ctx, conn, d.Id(), timeout); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for ECS Capacity Provider (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func partitionFromConn(conn *ecs.Client) string {
	return names.PartitionForRegion(conn.Options().Region).ID()
}

func findCapacityProvider(ctx context.Context, conn *ecs.Client, input *ecs.DescribeCapacityProvidersInput) (*awstypes.CapacityProvider, error) {
	output, err := findCapacityProviders(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findCapacityProviders(ctx context.Context, conn *ecs.Client, input *ecs.DescribeCapacityProvidersInput) ([]awstypes.CapacityProvider, error) {
	var output []awstypes.CapacityProvider

	err := describeCapacityProvidersPages(ctx, conn, input, func(page *ecs.DescribeCapacityProvidersOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		output = append(output, page.CapacityProviders...)

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	return output, nil
}

func findCapacityProviderByARN(ctx context.Context, conn *ecs.Client, arn string) (*awstypes.CapacityProvider, error) {
	input := ecs.DescribeCapacityProvidersInput{
		CapacityProviders: []string{arn},
		Include:           []awstypes.CapacityProviderField{awstypes.CapacityProviderFieldTags},
	}

	output, err := findCapacityProvider(ctx, conn, &input)

	// Some partitions (i.e., ISO) may not support tagging, giving error.
	if errs.IsUnsupportedOperationInPartitionError(partitionFromConn(conn), err) {
		input.Include = nil

		output, err = findCapacityProvider(ctx, conn, &input)
	}

	if err != nil {
		return nil, err
	}

	if status := output.Status; status == awstypes.CapacityProviderStatusInactive {
		return nil, &retry.NotFoundError{
			Message: string(status),
		}
	}

	return output, nil
}

func statusCapacityProvider(conn *ecs.Client, arn string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findCapacityProviderByARN(ctx, conn, arn)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func statusCapacityProviderUpdate(conn *ecs.Client, arn string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findCapacityProviderByARN(ctx, conn, arn)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.UpdateStatus), nil
	}
}

func waitCapacityProviderUpdated(ctx context.Context, conn *ecs.Client, arn string, timeout time.Duration) (*awstypes.CapacityProvider, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.CapacityProviderUpdateStatusUpdateInProgress),
		Target:  enum.Slice(awstypes.CapacityProviderUpdateStatusUpdateComplete),
		Refresh: statusCapacityProviderUpdate(conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.CapacityProvider); ok {
		retry.SetLastError(err, errors.New(aws.ToString(output.UpdateStatusReason)))

		return output, err
	}

	return nil, err
}

func waitCapacityProviderDeleted(ctx context.Context, conn *ecs.Client, arn string, timeout time.Duration) (*awstypes.CapacityProvider, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.CapacityProviderStatusActive, awstypes.CapacityProviderStatusDeprovisioning),
		Target:  []string{},
		Refresh: statusCapacityProvider(conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.CapacityProvider); ok {
		return output, err
	}

	return nil, err
}

func expandAutoScalingGroupProviderCreate(configured any) *awstypes.AutoScalingGroupProvider {
	if configured == nil {
		return nil
	}

	if len(configured.([]any)) == 0 {
		return nil
	}

	prov := awstypes.AutoScalingGroupProvider{}
	p := configured.([]any)[0].(map[string]any)
	arn := p["auto_scaling_group_arn"].(string)
	prov.AutoScalingGroupArn = aws.String(arn)

	if mtp := p["managed_draining"].(string); len(mtp) > 0 {
		prov.ManagedDraining = awstypes.ManagedDraining(mtp)
	}

	prov.ManagedScaling = expandManagedScaling(p["managed_scaling"])

	if mtp := p["managed_termination_protection"].(string); len(mtp) > 0 {
		prov.ManagedTerminationProtection = awstypes.ManagedTerminationProtection(mtp)
	}

	return &prov
}

func expandAutoScalingGroupProviderUpdate(configured any) *awstypes.AutoScalingGroupProviderUpdate {
	if configured == nil {
		return nil
	}

	if len(configured.([]any)) == 0 {
		return nil
	}

	prov := awstypes.AutoScalingGroupProviderUpdate{}
	p := configured.([]any)[0].(map[string]any)

	if mtp := p["managed_draining"].(string); len(mtp) > 0 {
		prov.ManagedDraining = awstypes.ManagedDraining(mtp)
	}

	prov.ManagedScaling = expandManagedScaling(p["managed_scaling"])

	if mtp := p["managed_termination_protection"].(string); len(mtp) > 0 {
		prov.ManagedTerminationProtection = awstypes.ManagedTerminationProtection(mtp)
	}

	return &prov
}

func expandManagedScaling(configured any) *awstypes.ManagedScaling {
	if configured == nil {
		return nil
	}

	if len(configured.([]any)) == 0 {
		return nil
	}

	tfMap := configured.([]any)[0].(map[string]any)

	managedScaling := awstypes.ManagedScaling{}

	if v, ok := tfMap["instance_warmup_period"].(int); ok {
		managedScaling.InstanceWarmupPeriod = aws.Int32(int32(v))
	}
	if v, ok := tfMap["maximum_scaling_step_size"].(int); ok && v != 0 {
		managedScaling.MaximumScalingStepSize = aws.Int32(int32(v))
	}
	if v, ok := tfMap["minimum_scaling_step_size"].(int); ok && v != 0 {
		managedScaling.MinimumScalingStepSize = aws.Int32(int32(v))
	}
	if v, ok := tfMap[names.AttrStatus].(string); ok && len(v) > 0 {
		managedScaling.Status = awstypes.ManagedScalingStatus(v)
	}
	if v, ok := tfMap["target_capacity"].(int); ok && v != 0 {
		managedScaling.TargetCapacity = aws.Int32(int32(v))
	}

	return &managedScaling
}

func flattenAutoScalingGroupProvider(provider *awstypes.AutoScalingGroupProvider) []map[string]any {
	if provider == nil {
		return nil
	}

	p := map[string]any{
		"auto_scaling_group_arn":         aws.ToString(provider.AutoScalingGroupArn),
		"managed_draining":               provider.ManagedDraining,
		"managed_scaling":                []map[string]any{},
		"managed_termination_protection": provider.ManagedTerminationProtection,
	}

	if provider.ManagedScaling != nil {
		m := map[string]any{
			"instance_warmup_period":    aws.ToInt32(provider.ManagedScaling.InstanceWarmupPeriod),
			"maximum_scaling_step_size": aws.ToInt32(provider.ManagedScaling.MaximumScalingStepSize),
			"minimum_scaling_step_size": aws.ToInt32(provider.ManagedScaling.MinimumScalingStepSize),
			names.AttrStatus:            provider.ManagedScaling.Status,
			"target_capacity":           aws.ToInt32(provider.ManagedScaling.TargetCapacity),
		}

		p["managed_scaling"] = []map[string]any{m}
	}

	result := []map[string]any{p}
	return result
}

func expandManagedInstancesProviderCreate(configured any) *awstypes.CreateManagedInstancesProviderConfiguration {
	if configured == nil {
		return nil
	}

	if len(configured.([]any)) == 0 {
		return nil
	}

	tfMap := configured.([]any)[0].(map[string]any)
	apiObject := &awstypes.CreateManagedInstancesProviderConfiguration{}

	if v, ok := tfMap["infrastructure_optimization"].([]any); ok && len(v) > 0 {
		apiObject.InfrastructureOptimization = expandInfrastructureOptimization(v)
	}

	if v, ok := tfMap["infrastructure_role_arn"].(string); ok && v != "" {
		apiObject.InfrastructureRoleArn = aws.String(v)
	}

	if v, ok := tfMap["instance_launch_template"].([]any); ok && len(v) > 0 {
		apiObject.InstanceLaunchTemplate = expandInstanceLaunchTemplateCreate(v)
	}

	if v, ok := tfMap[names.AttrPropagateTags].(string); ok && v != "" {
		apiObject.PropagateTags = awstypes.PropagateMITags(v)
	}

	return apiObject
}

func expandManagedInstancesProviderUpdate(configured any) *awstypes.UpdateManagedInstancesProviderConfiguration {
	if configured == nil {
		return nil
	}

	if len(configured.([]any)) == 0 {
		return nil
	}

	tfMap := configured.([]any)[0].(map[string]any)
	apiObject := &awstypes.UpdateManagedInstancesProviderConfiguration{}

	if v, ok := tfMap["infrastructure_optimization"].([]any); ok && len(v) > 0 {
		apiObject.InfrastructureOptimization = expandInfrastructureOptimization(v)
	}

	if v, ok := tfMap["infrastructure_role_arn"].(string); ok && v != "" {
		apiObject.InfrastructureRoleArn = aws.String(v)
	}

	if v, ok := tfMap["instance_launch_template"].([]any); ok && len(v) > 0 {
		apiObject.InstanceLaunchTemplate = expandInstanceLaunchTemplateUpdate(v)
	}

	if v, ok := tfMap[names.AttrPropagateTags].(string); ok && v != "" {
		apiObject.PropagateTags = awstypes.PropagateMITags(v)
	}

	return apiObject
}

func expandInfrastructureOptimization(tfList []any) *awstypes.InfrastructureOptimization {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]any)
	apiObject := &awstypes.InfrastructureOptimization{}

	if v, ok := tfMap["scale_in_after"].(int); ok {
		apiObject.ScaleInAfter = aws.Int32(int32(v))
	}

	return apiObject
}

func expandInstanceLaunchTemplateCreate(tfList []any) *awstypes.InstanceLaunchTemplate {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]any)
	apiObject := &awstypes.InstanceLaunchTemplate{}

	if v, ok := tfMap["capacity_option_type"].(string); ok && v != "" {
		apiObject.CapacityOptionType = awstypes.CapacityOptionType(v)
	}

	if v, ok := tfMap["ec2_instance_profile_arn"].(string); ok && v != "" {
		apiObject.Ec2InstanceProfileArn = aws.String(v)
	}

	if v, ok := tfMap["instance_requirements"].([]any); ok && len(v) > 0 {
		apiObject.InstanceRequirements = expandInstanceRequirementsRequest(v)
	}

	if v, ok := tfMap["monitoring"].(string); ok && v != "" {
		apiObject.Monitoring = awstypes.ManagedInstancesMonitoringOptions(v)
	}

	if v, ok := tfMap[names.AttrNetworkConfiguration].([]any); ok && len(v) > 0 {
		apiObject.NetworkConfiguration = expandManagedInstancesNetworkConfiguration(v)
	}

	if v, ok := tfMap["storage_configuration"].([]any); ok && len(v) > 0 {
		apiObject.StorageConfiguration = expandManagedInstancesStorageConfiguration(v)
	}

	return apiObject
}

func expandInstanceLaunchTemplateUpdate(tfList []any) *awstypes.InstanceLaunchTemplateUpdate {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]any)
	apiObject := &awstypes.InstanceLaunchTemplateUpdate{}

	if v, ok := tfMap["ec2_instance_profile_arn"].(string); ok && v != "" {
		apiObject.Ec2InstanceProfileArn = aws.String(v)
	}

	if v, ok := tfMap["instance_requirements"].([]any); ok && len(v) > 0 {
		apiObject.InstanceRequirements = expandInstanceRequirementsRequest(v)
	}

	if v, ok := tfMap["monitoring"].(string); ok && v != "" {
		apiObject.Monitoring = awstypes.ManagedInstancesMonitoringOptions(v)
	}

	if v, ok := tfMap[names.AttrNetworkConfiguration].([]any); ok && len(v) > 0 {
		apiObject.NetworkConfiguration = expandManagedInstancesNetworkConfiguration(v)
	}

	if v, ok := tfMap["storage_configuration"].([]any); ok && len(v) > 0 {
		apiObject.StorageConfiguration = expandManagedInstancesStorageConfiguration(v)
	}

	return apiObject
}

func expandManagedInstancesNetworkConfiguration(tfList []any) *awstypes.ManagedInstancesNetworkConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]any)
	apiObject := &awstypes.ManagedInstancesNetworkConfiguration{}

	if v, ok := tfMap[names.AttrSecurityGroups].(*schema.Set); ok && v.Len() > 0 {
		apiObject.SecurityGroups = flex.ExpandStringValueSet(v)
	}

	if v, ok := tfMap[names.AttrSubnets].(*schema.Set); ok && v.Len() > 0 {
		apiObject.Subnets = flex.ExpandStringValueSet(v)
	}

	return apiObject
}

func expandManagedInstancesStorageConfiguration(tfList []any) *awstypes.ManagedInstancesStorageConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]any)
	apiObject := &awstypes.ManagedInstancesStorageConfiguration{}

	if v, ok := tfMap["storage_size_gib"].(int); ok && v > 0 {
		apiObject.StorageSizeGiB = aws.Int32(int32(v))
	}

	return apiObject
}

func expandInstanceRequirementsRequest(tfList []any) *awstypes.InstanceRequirementsRequest {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]any)
	apiObject := &awstypes.InstanceRequirementsRequest{}

	if v, ok := tfMap["accelerator_count"].([]any); ok && len(v) > 0 {
		apiObject.AcceleratorCount = expandAcceleratorCountRequest(v)
	}

	if v, ok := tfMap["accelerator_manufacturers"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.AcceleratorManufacturers = flex.ExpandStringyValueSet[awstypes.AcceleratorManufacturer](v)
	}

	if v, ok := tfMap["accelerator_names"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.AcceleratorNames = flex.ExpandStringyValueSet[awstypes.AcceleratorName](v)
	}

	if v, ok := tfMap["accelerator_total_memory_mib"].([]any); ok && len(v) > 0 {
		apiObject.AcceleratorTotalMemoryMiB = expandAcceleratorTotalMemoryMiBRequest(v)
	}

	if v, ok := tfMap["accelerator_types"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.AcceleratorTypes = flex.ExpandStringyValueSet[awstypes.AcceleratorType](v)
	}

	if v, ok := tfMap["allowed_instance_types"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.AllowedInstanceTypes = flex.ExpandStringValueSet(v)
	}

	if v, ok := tfMap["bare_metal"].(string); ok && v != "" {
		apiObject.BareMetal = awstypes.BareMetal(v)
	}

	if v, ok := tfMap["baseline_ebs_bandwidth_mbps"].([]any); ok && len(v) > 0 {
		apiObject.BaselineEbsBandwidthMbps = expandBaselineEBSBandwidthMbpsRequest(v)
	}

	if v, ok := tfMap["burstable_performance"].(string); ok && v != "" {
		apiObject.BurstablePerformance = awstypes.BurstablePerformance(v)
	}

	if v, ok := tfMap["cpu_manufacturers"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.CpuManufacturers = flex.ExpandStringyValueSet[awstypes.CpuManufacturer](v)
	}

	if v, ok := tfMap["excluded_instance_types"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.ExcludedInstanceTypes = flex.ExpandStringValueSet(v)
	}

	if v, ok := tfMap["instance_generations"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.InstanceGenerations = flex.ExpandStringyValueSet[awstypes.InstanceGeneration](v)
	}

	if v, ok := tfMap["local_storage"].(string); ok && v != "" {
		apiObject.LocalStorage = awstypes.LocalStorage(v)
	}

	if v, ok := tfMap["local_storage_types"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.LocalStorageTypes = flex.ExpandStringyValueSet[awstypes.LocalStorageType](v)
	}

	if v, ok := tfMap["max_spot_price_as_percentage_of_optimal_on_demand_price"].(int); ok && v > 0 {
		apiObject.MaxSpotPriceAsPercentageOfOptimalOnDemandPrice = aws.Int32(int32(v))
	}

	if v, ok := tfMap["memory_gib_per_vcpu"].([]any); ok && len(v) > 0 {
		apiObject.MemoryGiBPerVCpu = expandMemoryGiBPerVCPURequest(v)
	}

	if v, ok := tfMap["memory_mib"].([]any); ok && len(v) > 0 {
		apiObject.MemoryMiB = expandMemoryMiBRequest(v)
	}

	if v, ok := tfMap["network_bandwidth_gbps"].([]any); ok && len(v) > 0 {
		apiObject.NetworkBandwidthGbps = expandNetworkBandwidthGbpsRequest(v)
	}

	if v, ok := tfMap["network_interface_count"].([]any); ok && len(v) > 0 {
		apiObject.NetworkInterfaceCount = expandNetworkInterfaceCountRequest(v)
	}

	if v, ok := tfMap["on_demand_max_price_percentage_over_lowest_price"].(int); ok && v > 0 {
		apiObject.OnDemandMaxPricePercentageOverLowestPrice = aws.Int32(int32(v))
	}

	if v, ok := tfMap["require_hibernate_support"].(bool); ok {
		apiObject.RequireHibernateSupport = aws.Bool(v)
	}

	if v, ok := tfMap["spot_max_price_percentage_over_lowest_price"].(int); ok && v > 0 {
		apiObject.SpotMaxPricePercentageOverLowestPrice = aws.Int32(int32(v))
	}

	if v, ok := tfMap["total_local_storage_gb"].([]any); ok && len(v) > 0 {
		apiObject.TotalLocalStorageGB = expandTotalLocalStorageGBRequest(v)
	}

	if v, ok := tfMap["vcpu_count"].([]any); ok && len(v) > 0 {
		apiObject.VCpuCount = expandVCPUCountRangeRequest(v)
	}

	return apiObject
}

func expandVCPUCountRangeRequest(tfList []any) *awstypes.VCpuCountRangeRequest {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]any)
	apiObject := &awstypes.VCpuCountRangeRequest{}

	if v, ok := tfMap[names.AttrMin].(int); ok && v > 0 {
		apiObject.Min = aws.Int32(int32(v))
	}

	if v, ok := tfMap[names.AttrMax].(int); ok && v > 0 {
		apiObject.Max = aws.Int32(int32(v))
	}

	return apiObject
}

func expandMemoryMiBRequest(tfList []any) *awstypes.MemoryMiBRequest {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]any)
	apiObject := &awstypes.MemoryMiBRequest{}

	if v, ok := tfMap[names.AttrMin].(int); ok && v > 0 {
		apiObject.Min = aws.Int32(int32(v))
	}

	if v, ok := tfMap[names.AttrMax].(int); ok && v > 0 {
		apiObject.Max = aws.Int32(int32(v))
	}

	return apiObject
}

func expandMemoryGiBPerVCPURequest(tfList []any) *awstypes.MemoryGiBPerVCpuRequest {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]any)
	apiObject := &awstypes.MemoryGiBPerVCpuRequest{}

	if v, ok := tfMap[names.AttrMin].(float64); ok && v > 0 {
		apiObject.Min = aws.Float64(v)
	}

	if v, ok := tfMap[names.AttrMax].(float64); ok && v > 0 {
		apiObject.Max = aws.Float64(v)
	}

	return apiObject
}

func expandNetworkBandwidthGbpsRequest(tfList []any) *awstypes.NetworkBandwidthGbpsRequest {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]any)
	apiObject := &awstypes.NetworkBandwidthGbpsRequest{}

	if v, ok := tfMap[names.AttrMin].(float64); ok && v > 0 {
		apiObject.Min = aws.Float64(v)
	}

	if v, ok := tfMap[names.AttrMax].(float64); ok && v > 0 {
		apiObject.Max = aws.Float64(v)
	}

	return apiObject
}

func expandNetworkInterfaceCountRequest(tfList []any) *awstypes.NetworkInterfaceCountRequest {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]any)
	apiObject := &awstypes.NetworkInterfaceCountRequest{}

	if v, ok := tfMap[names.AttrMin].(int); ok && v > 0 {
		apiObject.Min = aws.Int32(int32(v))
	}

	if v, ok := tfMap[names.AttrMax].(int); ok && v > 0 {
		apiObject.Max = aws.Int32(int32(v))
	}

	return apiObject
}

func expandTotalLocalStorageGBRequest(tfList []any) *awstypes.TotalLocalStorageGBRequest {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]any)
	apiObject := &awstypes.TotalLocalStorageGBRequest{}

	if v, ok := tfMap[names.AttrMin].(float64); ok && v > 0 {
		apiObject.Min = aws.Float64(v)
	}

	if v, ok := tfMap[names.AttrMax].(float64); ok && v > 0 {
		apiObject.Max = aws.Float64(v)
	}

	return apiObject
}

func expandBaselineEBSBandwidthMbpsRequest(tfList []any) *awstypes.BaselineEbsBandwidthMbpsRequest {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]any)
	apiObject := &awstypes.BaselineEbsBandwidthMbpsRequest{}

	if v, ok := tfMap[names.AttrMin].(int); ok && v > 0 {
		apiObject.Min = aws.Int32(int32(v))
	}

	if v, ok := tfMap[names.AttrMax].(int); ok && v > 0 {
		apiObject.Max = aws.Int32(int32(v))
	}

	return apiObject
}

func expandAcceleratorCountRequest(tfList []any) *awstypes.AcceleratorCountRequest {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]any)
	apiObject := &awstypes.AcceleratorCountRequest{}

	if v, ok := tfMap[names.AttrMin].(int); ok && v >= 0 {
		apiObject.Min = aws.Int32(int32(v))
	}

	if v, ok := tfMap[names.AttrMax].(int); ok && v >= 0 {
		apiObject.Max = aws.Int32(int32(v))
	}

	return apiObject
}

func expandAcceleratorTotalMemoryMiBRequest(tfList []any) *awstypes.AcceleratorTotalMemoryMiBRequest {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]any)
	apiObject := &awstypes.AcceleratorTotalMemoryMiBRequest{}

	if v, ok := tfMap[names.AttrMin].(int); ok && v >= 0 {
		apiObject.Min = aws.Int32(int32(v))
	}

	if v, ok := tfMap[names.AttrMax].(int); ok && v >= 0 {
		apiObject.Max = aws.Int32(int32(v))
	}

	return apiObject
}

func flattenManagedInstancesProvider(provider *awstypes.ManagedInstancesProvider) []map[string]any {
	if provider == nil {
		return nil
	}

	tfMap := map[string]any{
		"infrastructure_role_arn": aws.ToString(provider.InfrastructureRoleArn),
		names.AttrPropagateTags:   provider.PropagateTags,
	}

	if provider.InstanceLaunchTemplate != nil {
		tfMap["instance_launch_template"] = flattenInstanceLaunchTemplate(provider.InstanceLaunchTemplate)
	}

	if provider.InfrastructureOptimization != nil {
		tfMap["infrastructure_optimization"] = flattenInfrastructureOptimization(provider.InfrastructureOptimization)
	}

	return []map[string]any{tfMap}
}

func flattenInfrastructureOptimization(apiObject *awstypes.InfrastructureOptimization) []map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{
		"scale_in_after": aws.ToInt32(apiObject.ScaleInAfter),
	}

	return []map[string]any{tfMap}
}

func flattenInstanceLaunchTemplate(template *awstypes.InstanceLaunchTemplate) []map[string]any {
	if template == nil {
		return nil
	}

	tfMap := map[string]any{
		"capacity_option_type":     template.CapacityOptionType,
		"ec2_instance_profile_arn": aws.ToString(template.Ec2InstanceProfileArn),
		"monitoring":               template.Monitoring,
	}

	if template.InstanceRequirements != nil {
		tfMap["instance_requirements"] = flattenInstanceRequirementsRequest(template.InstanceRequirements)
	}

	if template.NetworkConfiguration != nil {
		networkConfig := map[string]any{
			names.AttrSubnets: template.NetworkConfiguration.Subnets,
		}
		if template.NetworkConfiguration.SecurityGroups != nil {
			networkConfig[names.AttrSecurityGroups] = template.NetworkConfiguration.SecurityGroups
		}
		tfMap[names.AttrNetworkConfiguration] = []map[string]any{networkConfig}
	}

	if template.StorageConfiguration != nil {
		tfMap["storage_configuration"] = []map[string]any{{
			"storage_size_gib": aws.ToInt32(template.StorageConfiguration.StorageSizeGiB),
		}}
	}

	return []map[string]any{tfMap}
}

func flattenInstanceRequirementsRequest(req *awstypes.InstanceRequirementsRequest) []map[string]any {
	if req == nil {
		return nil
	}

	tfMap := map[string]any{
		"bare_metal":            req.BareMetal,
		"burstable_performance": req.BurstablePerformance,
		"local_storage":         req.LocalStorage,
		"max_spot_price_as_percentage_of_optimal_on_demand_price": aws.ToInt32(req.MaxSpotPriceAsPercentageOfOptimalOnDemandPrice),
		"on_demand_max_price_percentage_over_lowest_price":        aws.ToInt32(req.OnDemandMaxPricePercentageOverLowestPrice),
		"require_hibernate_support":                               aws.ToBool(req.RequireHibernateSupport),
		"spot_max_price_percentage_over_lowest_price":             aws.ToInt32(req.SpotMaxPricePercentageOverLowestPrice),
	}

	if req.AcceleratorCount != nil {
		tfMap["accelerator_count"] = []map[string]any{{
			names.AttrMin: aws.ToInt32(req.AcceleratorCount.Min),
			names.AttrMax: aws.ToInt32(req.AcceleratorCount.Max),
		}}
	}

	if req.AcceleratorManufacturers != nil {
		tfMap["accelerator_manufacturers"] = req.AcceleratorManufacturers
	}

	if req.AcceleratorNames != nil {
		tfMap["accelerator_names"] = req.AcceleratorNames
	}

	if req.AcceleratorTotalMemoryMiB != nil {
		tfMap["accelerator_total_memory_mib"] = []map[string]any{{
			names.AttrMin: aws.ToInt32(req.AcceleratorTotalMemoryMiB.Min),
			names.AttrMax: aws.ToInt32(req.AcceleratorTotalMemoryMiB.Max),
		}}
	}

	if req.AcceleratorTypes != nil {
		tfMap["accelerator_types"] = req.AcceleratorTypes
	}

	if req.AllowedInstanceTypes != nil {
		tfMap["allowed_instance_types"] = req.AllowedInstanceTypes
	}

	if req.BaselineEbsBandwidthMbps != nil {
		tfMap["baseline_ebs_bandwidth_mbps"] = []map[string]any{{
			names.AttrMin: aws.ToInt32(req.BaselineEbsBandwidthMbps.Min),
			names.AttrMax: aws.ToInt32(req.BaselineEbsBandwidthMbps.Max),
		}}
	}

	if req.CpuManufacturers != nil {
		tfMap["cpu_manufacturers"] = req.CpuManufacturers
	}

	if req.ExcludedInstanceTypes != nil {
		tfMap["excluded_instance_types"] = req.ExcludedInstanceTypes
	}

	if req.InstanceGenerations != nil {
		tfMap["instance_generations"] = req.InstanceGenerations
	}

	if req.LocalStorageTypes != nil {
		tfMap["local_storage_types"] = req.LocalStorageTypes
	}

	if req.MemoryGiBPerVCpu != nil {
		tfMap["memory_gib_per_vcpu"] = []map[string]any{{
			names.AttrMin: aws.ToFloat64(req.MemoryGiBPerVCpu.Min),
			names.AttrMax: aws.ToFloat64(req.MemoryGiBPerVCpu.Max),
		}}
	}

	if req.MemoryMiB != nil {
		tfMap["memory_mib"] = []map[string]any{{
			names.AttrMin: aws.ToInt32(req.MemoryMiB.Min),
			names.AttrMax: aws.ToInt32(req.MemoryMiB.Max),
		}}
	}

	if req.NetworkBandwidthGbps != nil {
		tfMap["network_bandwidth_gbps"] = []map[string]any{{
			names.AttrMin: aws.ToFloat64(req.NetworkBandwidthGbps.Min),
			names.AttrMax: aws.ToFloat64(req.NetworkBandwidthGbps.Max),
		}}
	}

	if req.NetworkInterfaceCount != nil {
		tfMap["network_interface_count"] = []map[string]any{{
			names.AttrMin: aws.ToInt32(req.NetworkInterfaceCount.Min),
			names.AttrMax: aws.ToInt32(req.NetworkInterfaceCount.Max),
		}}
	}

	if req.TotalLocalStorageGB != nil {
		tfMap["total_local_storage_gb"] = []map[string]any{{
			names.AttrMin: aws.ToFloat64(req.TotalLocalStorageGB.Min),
			names.AttrMax: aws.ToFloat64(req.TotalLocalStorageGB.Max),
		}}
	}

	if req.VCpuCount != nil {
		tfMap["vcpu_count"] = []map[string]any{{
			names.AttrMin: aws.ToInt32(req.VCpuCount.Min),
			names.AttrMax: aws.ToInt32(req.VCpuCount.Max),
		}}
	}

	return []map[string]any{tfMap}
}
