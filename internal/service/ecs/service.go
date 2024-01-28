// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ecs

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"math"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_ecs_service", name="Service")
// @Tags(identifierAttribute="id")
func resourceService() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceServiceCreate,
		ReadWithoutTimeout:   resourceServiceRead,
		UpdateWithoutTimeout: resourceServiceUpdate,
		DeleteWithoutTimeout: resourceServiceDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceServiceImport,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(20 * time.Minute),
			Update: schema.DefaultTimeout(20 * time.Minute),
			Delete: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"alarms": {
				Type:             schema.TypeList,
				Optional:         true,
				MaxItems:         1,
				DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"alarm_names": {
							Type:     schema.TypeSet,
							Required: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"enable": {
							Type:     schema.TypeBool,
							Required: true,
						},
						"rollback": {
							Type:     schema.TypeBool,
							Required: true,
						},
					},
				},
			},
			"capacity_provider_strategy": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"base": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntBetween(0, 100000),
						},
						"capacity_provider": {
							Type:     schema.TypeString,
							Required: true,
						},
						"weight": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntBetween(0, 1000),
						},
					},
				},
			},
			"cluster": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"deployment_circuit_breaker": {
				Type:             schema.TypeList,
				Optional:         true,
				MaxItems:         1,
				DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enable": {
							Type:     schema.TypeBool,
							Required: true,
						},
						"rollback": {
							Type:     schema.TypeBool,
							Required: true,
						},
					},
				},
			},
			"deployment_controller": {
				Type:             schema.TypeList,
				Optional:         true,
				MaxItems:         1,
				DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:             schema.TypeString,
							ForceNew:         true,
							Optional:         true,
							Default:          types.DeploymentControllerTypeEcs,
							ValidateDiagFunc: enum.Validate[types.DeploymentControllerType](),
						},
					},
				},
			},
			"deployment_maximum_percent": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  200,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					if d.Get("scheduling_strategy") == types.SchedulingStrategyDaemon && new == "200" {
						return true
					}
					return false
				},
			},
			"deployment_minimum_healthy_percent": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  100,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					if d.Get("scheduling_strategy") == types.SchedulingStrategyDaemon && new == "100" {
						return true
					}
					return false
				},
			},
			"desired_count": {
				Type:     schema.TypeInt,
				Optional: true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return d.Get("scheduling_strategy") == types.SchedulingStrategyDaemon
				},
			},
			"enable_ecs_managed_tags": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"enable_execute_command": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"force_new_deployment": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"health_check_grace_period_seconds": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(0, math.MaxInt32),
			},
			"iam_role": {
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
				Computed: true,
			},
			"launch_type": {
				Type:             schema.TypeString,
				ForceNew:         true,
				Optional:         true,
				Computed:         true,
				ValidateDiagFunc: enum.Validate[types.LaunchType](),
			},
			"load_balancer": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"container_name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"container_port": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntBetween(0, 65536),
						},
						"elb_name": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"target_group_arn": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: verify.ValidARN,
						},
					},
				},
				Set: resourceLoadBalancerHash,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"network_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"assign_public_ip": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"security_groups": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"subnets": {
							Type:     schema.TypeSet,
							Required: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			"ordered_placement_strategy": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 5,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"field": {
							Type:     schema.TypeString,
							Optional: true,
							StateFunc: func(v interface{}) string {
								value := v.(string)
								if value == "host" {
									return "instanceId"
								}
								return value
							},
							DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
								return strings.EqualFold(old, new)
							},
						},
						"type": {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[types.PlacementStrategyType](),
						},
					},
				},
			},
			"placement_constraints": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 10,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"expression": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"type": {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[types.PlacementConstraintType](),
						},
					},
				},
			},
			"platform_version": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"propagate_tags": {
				Type:     schema.TypeString,
				Optional: true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					if old == "NONE" && new == "" {
						return true
					}
					return false
				},
				ValidateDiagFunc: enum.Validate[types.PropagateTags](),
			},
			"scheduling_strategy": {
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				Default:          types.SchedulingStrategyReplica,
				ValidateDiagFunc: enum.Validate[types.SchedulingStrategy](),
			},
			"service_connect_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enabled": {
							Type:     schema.TypeBool,
							Required: true,
						},
						"log_configuration": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"log_driver": {
										Type:             schema.TypeString,
										Required:         true,
										ValidateDiagFunc: enum.Validate[types.LogDriver](),
									},
									"options": {
										Type:     schema.TypeMap,
										Optional: true,
										Computed: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									"secret_option": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"name": {
													Type:     schema.TypeString,
													Required: true,
												},
												"value_from": {
													Type:     schema.TypeString,
													Required: true,
												},
											},
										},
									},
								},
							},
						},
						"namespace": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"service": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"client_alias": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"dns_name": {
													Type:     schema.TypeString,
													Optional: true,
												},
												"port": {
													Type:         schema.TypeInt,
													Required:     true,
													ValidateFunc: validation.IntBetween(0, 65535),
												},
											},
										},
									},
									"discovery_name": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"ingress_port_override": {
										Type:         schema.TypeInt,
										Optional:     true,
										ValidateFunc: validation.IntBetween(0, 65535),
									},
									"port_name": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},
					},
				},
			},
			"service_registries": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"container_name": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"container_port": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntBetween(0, 65536),
						},
						"port": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntBetween(0, 65536),
						},
						"registry_arn": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: verify.ValidARN,
						},
					},
				},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"task_definition": {
				Type:     schema.TypeString,
				Optional: true,
			},
			// modeled after null_resource & aws_api_gateway_deployment
			// only for _updates in-place_ rather than replacements
			"triggers": {
				Type:     schema.TypeMap,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"wait_for_steady_state": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
		},

		CustomizeDiff: customdiff.Sequence(
			verify.SetTagsDiff,
			capacityProviderStrategyCustomizeDiff,
			triggersCustomizeDiff,
		),
	}
}

func resourceServiceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ECSClient(ctx)
	partition := meta.(*conns.AWSClient).Partition

	deploymentController := expandDeploymentController(d.Get("deployment_controller").([]interface{}))
	deploymentMinimumHealthyPercent := d.Get("deployment_minimum_healthy_percent").(int)
	name := d.Get("name").(string)
	schedulingStrategy := types.SchedulingStrategy(d.Get("scheduling_strategy").(string))
	input := ecs.CreateServiceInput{
		CapacityProviderStrategy: expandCapacityProviderStrategy(d.Get("capacity_provider_strategy").(*schema.Set)),
		ClientToken:              aws.String(id.UniqueId()),
		DeploymentConfiguration:  &types.DeploymentConfiguration{},
		DeploymentController:     deploymentController,
		EnableECSManagedTags:     d.Get("enable_ecs_managed_tags").(bool),
		EnableExecuteCommand:     d.Get("enable_execute_command").(bool),
		NetworkConfiguration:     expandNetworkConfiguration(d.Get("network_configuration").([]interface{})),
		SchedulingStrategy:       schedulingStrategy,
		ServiceName:              aws.String(name),
		Tags:                     getTagsIn(ctx),
	}

	if v, ok := d.GetOk("alarms"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.DeploymentConfiguration.Alarms = expandAlarms(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("cluster"); ok {
		input.Cluster = aws.String(v.(string))
	}

	if schedulingStrategy == types.SchedulingStrategyDaemon && deploymentMinimumHealthyPercent != 100 {
		input.DeploymentConfiguration.MinimumHealthyPercent = aws.Int32(int32(deploymentMinimumHealthyPercent))
	} else if schedulingStrategy == types.SchedulingStrategyReplica {
		input.DeploymentConfiguration.MaximumPercent = aws.Int32(int32(d.Get("deployment_maximum_percent").(int)))
		input.DeploymentConfiguration.MinimumHealthyPercent = aws.Int32(int32(deploymentMinimumHealthyPercent))
		input.DesiredCount = aws.Int32(int32(d.Get("desired_count").(int)))
	}

	if v, ok := d.GetOk("deployment_circuit_breaker"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.DeploymentConfiguration.DeploymentCircuitBreaker = expandDeploymentCircuitBreaker(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("health_check_grace_period_seconds"); ok {
		input.HealthCheckGracePeriodSeconds = aws.Int32(int32(v.(int)))
	}

	if v, ok := d.GetOk("iam_role"); ok {
		input.Role = aws.String(v.(string))
	}

	if v, ok := d.GetOk("launch_type"); ok {
		input.LaunchType = types.LaunchType(v.(string))
		// When creating a service that uses the EXTERNAL deployment controller,
		// you can specify only parameters that aren't controlled at the task set level
		// hence you cannot set LaunchType, not changing the default launch_type from EC2 to empty
		// string to have backward compatibility
		if deploymentController != nil && deploymentController.Type == types.DeploymentControllerTypeExternal {
			input.LaunchType = ""
		}
	}

	loadBalancers := expandLoadBalancers(d.Get("load_balancer").(*schema.Set).List())
	if len(loadBalancers) > 0 {
		input.LoadBalancers = loadBalancers
	}

	if v, ok := d.GetOk("ordered_placement_strategy"); ok {
		ps, err := expandPlacementStrategy(v.([]interface{}))

		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}

		input.PlacementStrategy = ps
	}

	if v, ok := d.Get("placement_constraints").(*schema.Set); ok {
		pc, err := expandPlacementConstraints(v.List())

		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}

		input.PlacementConstraints = pc
	}

	if v, ok := d.GetOk("platform_version"); ok {
		input.PlatformVersion = aws.String(v.(string))
	}

	if v, ok := d.GetOk("propagate_tags"); ok {
		input.PropagateTags = types.PropagateTags(v.(string))
	}

	if v, ok := d.GetOk("service_connect_configuration"); ok && len(v.([]interface{})) > 0 {
		input.ServiceConnectConfiguration = expandServiceConnectConfiguration(v.([]interface{}))
	}

	serviceRegistries := d.Get("service_registries").([]interface{})
	if len(serviceRegistries) > 0 {
		srs := make([]types.ServiceRegistry, 0, len(serviceRegistries))
		for _, v := range serviceRegistries {
			raw := v.(map[string]interface{})
			sr := &types.ServiceRegistry{
				RegistryArn: aws.String(raw["registry_arn"].(string)),
			}
			if port, ok := raw["port"].(int); ok && port != 0 {
				sr.Port = aws.Int32(int32(port))
			}
			if raw, ok := raw["container_port"].(int); ok && raw != 0 {
				sr.ContainerPort = aws.Int32(int32(raw))
			}
			if raw, ok := raw["container_name"].(string); ok && raw != "" {
				sr.ContainerName = aws.String(raw)
			}

			srs = append(srs, *sr)
		}
		input.ServiceRegistries = srs
	}

	if v, ok := d.GetOk("task_definition"); ok {
		input.TaskDefinition = aws.String(v.(string))
	}

	output, err := serviceCreateWithRetry(ctx, conn, input)

	// Some partitions (e.g. ISO) may not support tag-on-create.
	if input.Tags != nil && errs.IsUnsupportedOperationInPartitionError(partition, err) {
		input.Tags = nil

		output, err = serviceCreateWithRetry(ctx, conn, input)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating ECS Service (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.Service.ServiceArn))

	fn := waitServiceActive
	if d.Get("wait_for_steady_state").(bool) {
		fn = waitServiceStable
	}
	if _, err := fn(ctx, conn, d.Id(), d.Get("cluster").(string), partition, d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for ECS Service (%s) create: %s", d.Id(), err)
	}

	// For partitions not supporting tag-on-create, attempt tag after create.
	if tags := getTagsIn(ctx); input.Tags == nil && len(tags) > 0 {
		err := createTags(ctx, conn, d.Id(), tags)

		// If default tags only, continue. Otherwise, error.
		if v, ok := d.GetOk(names.AttrTags); (!ok || len(v.(map[string]interface{})) == 0) && errs.IsUnsupportedOperationInPartitionError(partition, err) {
			return append(diags, resourceServiceRead(ctx, d, meta)...)
		}

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "setting ECS Service (%s) tags: %s", d.Id(), err)
		}
	}

	return append(diags, resourceServiceRead(ctx, d, meta)...)
}

func resourceServiceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ECSClient(ctx)
	partition := meta.(*conns.AWSClient).Partition

	cluster := d.Get("cluster").(string)

	service, err := findServiceByIDWaitForActive(ctx, conn, d.Id(), cluster, partition)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] ECS Service (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if errs.IsA[*types.ClusterNotFoundException](err) {
		log.Printf("[WARN] ECS Service (%s) parent cluster (%s) not found, removing from state.", d.Id(), cluster)
		d.SetId("")
		return diags
	}

	var ea *expectActiveError
	if errors.As(err, &ea) {
		log.Printf("[WARN] ECS Service (%s) in status %q, removing from state.", d.Id(), ea.status)
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ECS service (%s): %s", d.Id(), err)
	}

	d.SetId(aws.ToString(service.ServiceArn))
	d.Set("name", service.ServiceName)

	// When creating a service that uses the EXTERNAL deployment controller,
	// you can specify only parameters that aren't controlled at the task set level
	// hence TaskDefinition will not be set by aws sdk
	if service.TaskDefinition != nil {
		// Save task definition in the same format
		if strings.HasPrefix(d.Get("task_definition").(string), "arn:"+meta.(*conns.AWSClient).Partition+":ecs:") {
			d.Set("task_definition", service.TaskDefinition)
		} else {
			taskDefinition := buildFamilyAndRevisionFromARN(aws.ToString(service.TaskDefinition))
			d.Set("task_definition", taskDefinition)
		}
	}

	d.Set("scheduling_strategy", service.SchedulingStrategy)
	d.Set("desired_count", service.DesiredCount)
	d.Set("health_check_grace_period_seconds", service.HealthCheckGracePeriodSeconds)
	d.Set("launch_type", service.LaunchType)
	d.Set("enable_ecs_managed_tags", service.EnableECSManagedTags)
	d.Set("propagate_tags", service.PropagateTags)
	d.Set("platform_version", service.PlatformVersion)
	d.Set("enable_execute_command", service.EnableExecuteCommand)

	d.Set("triggers", d.Get("triggers"))

	// Save cluster in the same format
	if strings.HasPrefix(d.Get("cluster").(string), "arn:"+meta.(*conns.AWSClient).Partition+":ecs:") {
		d.Set("cluster", service.ClusterArn)
	} else {
		clusterARN := GetClusterNameFromARN(aws.ToString(service.ClusterArn))
		d.Set("cluster", clusterARN)
	}

	// Save IAM role in the same format
	if service.RoleArn != nil {
		if strings.HasPrefix(d.Get("iam_role").(string), "arn:"+meta.(*conns.AWSClient).Partition+":iam:") {
			d.Set("iam_role", service.RoleArn)
		} else {
			roleARN := GetRoleNameFromARN(aws.ToString(service.RoleArn))
			d.Set("iam_role", roleARN)
		}
	}

	if service.DeploymentConfiguration != nil {
		d.Set("deployment_maximum_percent", service.DeploymentConfiguration.MaximumPercent)
		d.Set("deployment_minimum_healthy_percent", service.DeploymentConfiguration.MinimumHealthyPercent)

		if service.DeploymentConfiguration.Alarms != nil {
			if err := d.Set("alarms", []interface{}{flattenAlarms(service.DeploymentConfiguration.Alarms)}); err != nil {
				return sdkdiag.AppendErrorf(diags, "setting alarms: %s", err)
			}
		} else {
			d.Set("alarms", nil)
		}

		if service.DeploymentConfiguration.DeploymentCircuitBreaker != nil {
			if err := d.Set("deployment_circuit_breaker", []interface{}{flattenDeploymentCircuitBreaker(service.DeploymentConfiguration.DeploymentCircuitBreaker)}); err != nil {
				return sdkdiag.AppendErrorf(diags, "setting deployment_circuit_breaker: %s", err)
			}
		} else {
			d.Set("deployment_circuit_breaker", nil)
		}
	}

	if err := d.Set("deployment_controller", flattenDeploymentController(service.DeploymentController)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting deployment_controller: %s", err)
	}

	if service.LoadBalancers != nil {
		if err := d.Set("load_balancer", flattenLoadBalancers(service.LoadBalancers)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting load_balancer: %s", err)
		}
	}

	if err := d.Set("capacity_provider_strategy", flattenCapacityProviderStrategy(service.CapacityProviderStrategy)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting capacity_provider_strategy: %s", err)
	}

	if err := d.Set("ordered_placement_strategy", flattenPlacementStrategy(service.PlacementStrategy)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting ordered_placement_strategy: %s", err)
	}

	if err := d.Set("placement_constraints", flattenServicePlacementConstraints(service.PlacementConstraints)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting placement_constraints: %s", err)
	}

	if err := d.Set("network_configuration", flattenNetworkConfiguration(service.NetworkConfiguration)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting network_configuration: %s", err)
	}

	// if err := d.Set("service_connect_configuration", flattenServiceConnectConfiguration(service.ServiceConnectConfiguration)); err != nil {
	// 	return fmt.Errorf("setting service_connect_configuration for (%s): %w", d.Id(), err)
	// }

	if err := d.Set("service_registries", flattenServiceRegistries(service.ServiceRegistries)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting service_registries: %s", err)
	}

	setTagsOut(ctx, service.Tags)

	return diags
}

func resourceServiceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ECSClient(ctx)
	partition := meta.(*conns.AWSClient).Partition

	if d.HasChangesExcept("tags", "tags_all") {
		input := &ecs.UpdateServiceInput{
			Cluster:            aws.String(d.Get("cluster").(string)),
			ForceNewDeployment: d.Get("force_new_deployment").(bool),
			Service:            aws.String(d.Id()),
		}

		if d.HasChange("alarms") {
			if input.DeploymentConfiguration == nil {
				input.DeploymentConfiguration = &types.DeploymentConfiguration{}
			}

			if v, ok := d.GetOk("alarms"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
				input.DeploymentConfiguration.Alarms = expandAlarms(v.([]interface{})[0].(map[string]interface{}))
			}
		}

		if d.HasChange("capacity_provider_strategy") {
			input.CapacityProviderStrategy = expandCapacityProviderStrategy(d.Get("capacity_provider_strategy").(*schema.Set))
		}

		if d.HasChange("deployment_circuit_breaker") {
			if input.DeploymentConfiguration == nil {
				input.DeploymentConfiguration = &types.DeploymentConfiguration{}
			}

			// To remove an existing deployment circuit breaker, specify an empty object.
			input.DeploymentConfiguration.DeploymentCircuitBreaker = &types.DeploymentCircuitBreaker{}

			if v, ok := d.GetOk("deployment_circuit_breaker"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
				input.DeploymentConfiguration.DeploymentCircuitBreaker = expandDeploymentCircuitBreaker(v.([]interface{})[0].(map[string]interface{}))
			}
		}

		switch schedulingStrategy := types.SchedulingStrategy(d.Get("scheduling_strategy").(string)); schedulingStrategy {
		case types.SchedulingStrategyDaemon:
			if d.HasChange("deployment_minimum_healthy_percent") {
				if input.DeploymentConfiguration == nil {
					input.DeploymentConfiguration = &types.DeploymentConfiguration{}
				}

				input.DeploymentConfiguration.MinimumHealthyPercent = aws.Int32(int32(d.Get("deployment_minimum_healthy_percent").(int)))
			}
		case types.SchedulingStrategyReplica:
			if d.HasChanges("deployment_maximum_percent", "deployment_minimum_healthy_percent") {
				if input.DeploymentConfiguration == nil {
					input.DeploymentConfiguration = &types.DeploymentConfiguration{}
				}

				input.DeploymentConfiguration.MaximumPercent = aws.Int32(int32(d.Get("deployment_maximum_percent").(int)))
				input.DeploymentConfiguration.MinimumHealthyPercent = aws.Int32(int32(d.Get("deployment_minimum_healthy_percent").(int)))
			}

			if d.HasChange("desired_count") {
				input.DesiredCount = aws.Int32(int32(d.Get("desired_count").(int)))
			}
		}

		if d.HasChange("enable_ecs_managed_tags") {
			input.EnableECSManagedTags = aws.Bool(d.Get("enable_ecs_managed_tags").(bool))
		}

		if d.HasChange("enable_execute_command") {
			input.EnableExecuteCommand = aws.Bool(d.Get("enable_execute_command").(bool))
		}

		if d.HasChange("health_check_grace_period_seconds") {
			input.HealthCheckGracePeriodSeconds = aws.Int32(int32(d.Get("health_check_grace_period_seconds").(int)))
		}

		if d.HasChange("load_balancer") {
			if v, ok := d.Get("load_balancer").(*schema.Set); ok && v != nil {
				input.LoadBalancers = expandLoadBalancers(v.List())
			}
		}

		if d.HasChange("network_configuration") {
			input.NetworkConfiguration = expandNetworkConfiguration(d.Get("network_configuration").([]interface{}))
		}

		if d.HasChange("ordered_placement_strategy") {
			// Reference: https://docs.aws.amazon.com/AmazonECS/latest/APIReference/API_UpdateService.html#ECS-UpdateService-request-placementStrategy
			// To remove an existing placement strategy, specify an empty object.
			input.PlacementStrategy = []types.PlacementStrategy{}

			if v, ok := d.GetOk("ordered_placement_strategy"); ok && len(v.([]interface{})) > 0 {
				ps, err := expandPlacementStrategy(v.([]interface{}))

				if err != nil {
					return sdkdiag.AppendFromErr(diags, err)
				}

				input.PlacementStrategy = ps
			}
		}

		if d.HasChange("placement_constraints") {
			// Reference: https://docs.aws.amazon.com/AmazonECS/latest/APIReference/API_UpdateService.html#ECS-UpdateService-request-placementConstraints
			// To remove all existing placement constraints, specify an empty array.
			input.PlacementConstraints = []types.PlacementConstraint{}

			if v, ok := d.Get("placement_constraints").(*schema.Set); ok && v.Len() > 0 {
				pc, err := expandPlacementConstraints(v.List())

				if err != nil {
					return sdkdiag.AppendFromErr(diags, err)
				}

				input.PlacementConstraints = pc
			}
		}

		if d.HasChange("platform_version") {
			input.PlatformVersion = aws.String(d.Get("platform_version").(string))
		}

		if d.HasChange("propagate_tags") {
			input.PropagateTags = types.PropagateTags(d.Get("propagate_tags").(string))
		}

		if d.HasChange("service_connect_configuration") {
			input.ServiceConnectConfiguration = expandServiceConnectConfiguration(d.Get("service_connect_configuration").([]interface{}))
		}

		if d.HasChange("service_registries") {
			input.ServiceRegistries = expandServiceRegistries(d.Get("service_registries").([]interface{}))
		}

		if d.HasChange("task_definition") {
			input.TaskDefinition = aws.String(d.Get("task_definition").(string))
		}

		// Retry due to IAM eventual consistency
		err := retry.RetryContext(ctx, propagationTimeout+serviceUpdateTimeout, func() *retry.RetryError {
			_, err := conn.UpdateService(ctx, input)

			if err != nil {
				if errs.IsAErrorMessageContains[*types.InvalidParameterException](err, "verify that the ECS service role being passed has the proper permissions") {
					return retry.RetryableError(err)
				}

				if errs.IsAErrorMessageContains[*types.InvalidParameterException](err, "does not have an associated load balancer") {
					return retry.RetryableError(err)
				}

				return retry.NonRetryableError(err)
			}
			return nil
		})

		if tfresource.TimedOut(err) {
			_, err = conn.UpdateService(ctx, input)
		}

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating ECS Service (%s): %s", d.Id(), err)
		}

		fn := waitServiceActive
		if d.Get("wait_for_steady_state").(bool) {
			fn = waitServiceStable
		}
		if _, err := fn(ctx, conn, d.Id(), d.Get("cluster").(string), partition, d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for ECS Service (%s) update: %s", d.Id(), err)
		}
	}

	return append(diags, resourceServiceRead(ctx, d, meta)...)
}

func resourceServiceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ECSClient(ctx)
	partition := meta.(*conns.AWSClient).Partition

	service, err := findServiceNoTagsByID(ctx, conn, d.Id(), d.Get("cluster").(string), partition)

	if tfresource.NotFound(err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ECS Service (%s): %s", d.Id(), err)
	}

	if aws.ToString(service.Status) == serviceStatusInactive {
		return diags
	}

	// Drain the ECS service
	if aws.ToString(service.Status) != serviceStatusDraining && service.SchedulingStrategy != types.SchedulingStrategyDaemon {
		_, err := conn.UpdateService(ctx, &ecs.UpdateServiceInput{
			Service:      aws.String(d.Id()),
			Cluster:      aws.String(d.Get("cluster").(string)),
			DesiredCount: aws.Int32(0),
		})

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "draining ECS Service (%s): %s", d.Id(), err)
		}
	}

	input := ecs.DeleteServiceInput{
		Service: aws.String(d.Id()),
		Cluster: aws.String(d.Get("cluster").(string)),
	}
	log.Printf("[DEBUG] Deleting ECS Service: %s", d.Id())
	err = retry.RetryContext(ctx, d.Timeout(schema.TimeoutDelete), func() *retry.RetryError {
		_, err := conn.DeleteService(ctx, &input)

		if err != nil {
			if errs.IsAErrorMessageContains[*types.InvalidParameterException](err, "The service cannot be stopped while deployments are active.") {
				return retry.RetryableError(err)
			}

			if errs.IsAErrorMessageContains[*types.InvalidParameterException](err, "has a dependent object") {
				return retry.RetryableError(err)
			}

			return retry.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		_, err = conn.DeleteService(ctx, &input)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting ECS Service (%s): %s", d.Id(), err)
	}

	if err := waitServiceInactive(ctx, conn, d.Id(), d.Get("cluster").(string), partition, d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for ECS Service (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func resourceServiceImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	if len(strings.Split(d.Id(), "/")) != 2 {
		return []*schema.ResourceData{}, fmt.Errorf("wrong format of resource: %s, expecting 'cluster-name/service-name'", d.Id())
	}
	cluster := strings.Split(d.Id(), "/")[0]
	name := strings.Split(d.Id(), "/")[1]
	log.Printf("[DEBUG] Importing ECS service %s from cluster %s", name, cluster)

	d.SetId(name)
	clusterArn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Region:    meta.(*conns.AWSClient).Region,
		Service:   "ecs",
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("cluster/%s", cluster),
	}.String()
	d.Set("cluster", clusterArn)
	return []*schema.ResourceData{d}, nil
}

func triggersCustomizeDiff(_ context.Context, d *schema.ResourceDiff, meta interface{}) error {
	// clears diff to avoid extraneous diffs but lets it pass for triggering update
	fnd := false
	if v, ok := d.GetOk("force_new_deployment"); ok {
		fnd = v.(bool)
	}

	if d.HasChange("triggers") && !fnd {
		return d.Clear("triggers")
	}

	if d.HasChange("triggers") && fnd {
		o, n := d.GetChange("triggers")
		if len(o.(map[string]interface{})) > 0 && len(n.(map[string]interface{})) == 0 {
			return d.Clear("triggers")
		}

		return nil
	}

	return nil
}

func capacityProviderStrategyCustomizeDiff(_ context.Context, d *schema.ResourceDiff, meta interface{}) error {
	// to be backward compatible, should ForceNew almost always (previous behavior), unless:
	//   force_new_deployment is true and
	//   neither the old set nor new set is 0 length
	if v := d.Get("force_new_deployment").(bool); !v {
		return capacityProviderStrategyForceNew(d)
	}

	old, new := d.GetChange("capacity_provider_strategy")

	ol := old.(*schema.Set).Len()
	nl := new.(*schema.Set).Len()

	if (ol == 0 && nl > 0) || (ol > 0 && nl == 0) {
		return capacityProviderStrategyForceNew(d)
	}

	return nil
}

func capacityProviderStrategyForceNew(d *schema.ResourceDiff) error {
	for _, key := range d.GetChangedKeysPrefix("capacity_provider_strategy") {
		if d.HasChange(key) {
			if err := d.ForceNew(key); err != nil {
				return fmt.Errorf("while attempting to force a new ECS service for capacity_provider_strategy: %w", err)
			}
		}
	}
	return nil
}

func expandAlarms(tfMap map[string]interface{}) *types.DeploymentAlarms {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.DeploymentAlarms{}

	if v, ok := tfMap["enable"].(bool); ok {
		apiObject.Enable = v
	}

	if v, ok := tfMap["enable"].(bool); ok {
		apiObject.Rollback = v
	}

	if v, ok := tfMap["alarm_names"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.AlarmNames = flex.ExpandStringValueSet(v)
	}

	return apiObject
}

func flattenAlarms(apiObject *types.DeploymentAlarms) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.AlarmNames; v != nil {
		tfMap["alarm_names"] = aws.StringSlice(v)
	}

	tfMap["enable"] = apiObject.Enable
	tfMap["rollback"] = apiObject.Rollback

	return tfMap
}

func expandDeploymentController(l []interface{}) *types.DeploymentController {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	deploymentController := &types.DeploymentController{
		Type: types.DeploymentControllerType(m["type"].(string)),
	}

	return deploymentController
}

func flattenDeploymentController(deploymentController *types.DeploymentController) []interface{} {
	m := map[string]interface{}{
		"type": types.DeploymentControllerTypeEcs,
	}

	if deploymentController == nil {
		return []interface{}{m}
	}

	m["type"] = deploymentController.Type

	return []interface{}{m}
}

func expandDeploymentCircuitBreaker(tfMap map[string]interface{}) *types.DeploymentCircuitBreaker {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.DeploymentCircuitBreaker{}

	apiObject.Enable = tfMap["enable"].(bool)
	apiObject.Rollback = tfMap["rollback"].(bool)

	return apiObject
}

func flattenDeploymentCircuitBreaker(apiObject *types.DeploymentCircuitBreaker) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	tfMap["enable"] = apiObject.Enable
	tfMap["rollback"] = apiObject.Rollback

	return tfMap
}

func flattenNetworkConfiguration(nc *types.NetworkConfiguration) []interface{} {
	if nc == nil {
		return nil
	}

	result := make(map[string]interface{})
	result["security_groups"] = flex.FlattenStringValueSet(nc.AwsvpcConfiguration.SecurityGroups)
	result["subnets"] = flex.FlattenStringValueSet(nc.AwsvpcConfiguration.Subnets)

	if nc.AwsvpcConfiguration.AssignPublicIp != "" {
		result["assign_public_ip"] = nc.AwsvpcConfiguration.AssignPublicIp == types.AssignPublicIpEnabled
	}

	return []interface{}{result}
}

func expandNetworkConfiguration(nc []interface{}) *types.NetworkConfiguration {
	if len(nc) == 0 {
		return nil
	}
	awsVpcConfig := &types.AwsVpcConfiguration{}
	raw := nc[0].(map[string]interface{})
	if val, ok := raw["security_groups"]; ok {
		awsVpcConfig.SecurityGroups = flex.ExpandStringValueSet(val.(*schema.Set))
	}
	awsVpcConfig.Subnets = flex.ExpandStringValueSet(raw["subnets"].(*schema.Set))
	if val, ok := raw["assign_public_ip"].(bool); ok {
		awsVpcConfig.AssignPublicIp = types.AssignPublicIpDisabled
		if val {
			awsVpcConfig.AssignPublicIp = types.AssignPublicIpEnabled
		}
	}

	return &types.NetworkConfiguration{AwsvpcConfiguration: awsVpcConfig}
}

func expandPlacementConstraints(tfList []interface{}) ([]types.PlacementConstraint, error) {
	if len(tfList) == 0 {
		return nil, nil
	}

	var result []types.PlacementConstraint

	for _, tfMapRaw := range tfList {
		if tfMapRaw == nil {
			continue
		}

		tfMap := tfMapRaw.(map[string]interface{})

		apiObject := &types.PlacementConstraint{}

		if v, ok := tfMap["expression"].(string); ok && v != "" {
			apiObject.Expression = aws.String(v)
		}

		if v, ok := tfMap["type"].(string); ok && v != "" {
			apiObject.Type = types.PlacementConstraintType(v)
		}

		if err := validPlacementConstraint(string(apiObject.Type), aws.ToString(apiObject.Expression)); err != nil {
			return result, err
		}

		result = append(result, *apiObject)
	}

	return result, nil
}

func flattenServicePlacementConstraints(pcs []types.PlacementConstraint) []map[string]interface{} {
	if len(pcs) == 0 {
		return nil
	}
	results := make([]map[string]interface{}, 0)
	for _, pc := range pcs {
		c := make(map[string]interface{})
		c["type"] = pc.Type
		if pc.Expression != nil {
			c["expression"] = aws.ToString(pc.Expression)
		}

		results = append(results, c)
	}
	return results
}

func expandPlacementStrategy(s []interface{}) ([]types.PlacementStrategy, error) {
	if len(s) == 0 {
		return nil, nil
	}
	pss := make([]types.PlacementStrategy, 0)
	for _, raw := range s {
		p, ok := raw.(map[string]interface{})

		if !ok {
			continue
		}

		t, ok := p["type"].(string)

		if !ok {
			return nil, fmt.Errorf("missing type attribute in placement strategy configuration block")
		}

		f, ok := p["field"].(string)

		if !ok {
			return nil, fmt.Errorf("missing field attribute in placement strategy configuration block")
		}

		if err := validPlacementStrategy(t, f); err != nil {
			return nil, err
		}
		ps := &types.PlacementStrategy{
			Type: types.PlacementStrategyType(t),
		}
		if f != "" {
			// Field must be omitted (i.e. not empty string) for random strategy
			ps.Field = aws.String(f)
		}
		pss = append(pss, *ps)
	}
	return pss, nil
}

func flattenPlacementStrategy(pss []types.PlacementStrategy) []interface{} {
	if len(pss) == 0 {
		return nil
	}
	results := make([]interface{}, 0, len(pss))
	for _, ps := range pss {
		c := make(map[string]interface{})
		c["type"] = ps.Type

		if ps.Field != nil {
			c["field"] = aws.ToString(ps.Field)

			// for some fields the API requires lowercase for creation but will return uppercase on query
			if aws.ToString(ps.Field) == "MEMORY" || aws.ToString(ps.Field) == "CPU" {
				c["field"] = strings.ToLower(aws.ToString(ps.Field))
			}
		}

		results = append(results, c)
	}
	return results
}

func expandServiceConnectConfiguration(sc []interface{}) *types.ServiceConnectConfiguration {
	if len(sc) == 0 {
		return &types.ServiceConnectConfiguration{
			Enabled: false,
		}
	}
	raw := sc[0].(map[string]interface{})

	config := &types.ServiceConnectConfiguration{}
	if v, ok := raw["enabled"].(bool); ok {
		config.Enabled = v
	}

	if v, ok := raw["log_configuration"].([]interface{}); ok && len(v) > 0 {
		config.LogConfiguration = expandLogConfiguration(v)
	}

	if v, ok := raw["namespace"].(string); ok && v != "" {
		config.Namespace = aws.String(v)
	}

	if v, ok := raw["service"].([]interface{}); ok && len(v) > 0 {
		config.Services = expandServices(v)
	}

	return config
}

func expandLogConfiguration(lc []interface{}) *types.LogConfiguration {
	if len(lc) == 0 {
		return &types.LogConfiguration{}
	}
	raw := lc[0].(map[string]interface{})

	config := &types.LogConfiguration{}
	if v, ok := raw["log_driver"].(string); ok && v != "" {
		config.LogDriver = types.LogDriver(v)
	}
	if v, ok := raw["options"].(map[string]interface{}); ok && len(v) > 0 {
		config.Options = flex.ExpandStringValueMap(v)
	}
	if v, ok := raw["secret_option"].([]interface{}); ok && len(v) > 0 {
		config.SecretOptions = expandSecretOptions(v)
	}

	return config
}

func expandSecretOptions(sop []interface{}) []types.Secret {
	if len(sop) == 0 {
		return nil
	}

	var out []types.Secret
	for _, item := range sop {
		raw, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		var config types.Secret
		if v, ok := raw["name"].(string); ok && v != "" {
			config.Name = aws.String(v)
		}
		if v, ok := raw["value_from"].(string); ok && v != "" {
			config.ValueFrom = aws.String(v)
		}

		out = append(out, config)
	}

	return out
}

func expandServices(srv []interface{}) []types.ServiceConnectService {
	if len(srv) == 0 {
		return nil
	}

	var out []types.ServiceConnectService
	for _, item := range srv {
		raw, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		var config types.ServiceConnectService
		if v, ok := raw["client_alias"].([]interface{}); ok && len(v) > 0 {
			config.ClientAliases = expandClientAliases(v)
		}
		if v, ok := raw["discovery_name"].(string); ok && v != "" {
			config.DiscoveryName = aws.String(v)
		}
		if v, ok := raw["ingress_port_override"].(int); ok && v != 0 {
			config.IngressPortOverride = aws.Int32(int32(v))
		}
		if v, ok := raw["port_name"].(string); ok && v != "" {
			config.PortName = aws.String(v)
		}

		out = append(out, config)
	}

	return out
}

func expandClientAliases(srv []interface{}) []types.ServiceConnectClientAlias {
	if len(srv) == 0 {
		return nil
	}

	var out []types.ServiceConnectClientAlias
	for _, item := range srv {
		raw, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		var config types.ServiceConnectClientAlias
		if v, ok := raw["port"].(int); ok {
			config.Port = aws.Int32(int32(v))
		}
		if v, ok := raw["dns_name"].(string); ok && v != "" {
			config.DnsName = aws.String(v)
		}

		out = append(out, config)
	}

	return out
}

func flattenServiceRegistries(srs []types.ServiceRegistry) []map[string]interface{} {
	if len(srs) == 0 {
		return nil
	}
	results := make([]map[string]interface{}, 0)
	for _, sr := range srs {
		c := map[string]interface{}{
			"registry_arn": aws.ToString(sr.RegistryArn),
		}
		if sr.Port != nil {
			c["port"] = sr.Port
		}
		if sr.ContainerPort != nil {
			c["container_port"] = sr.ContainerPort
		}
		if sr.ContainerName != nil {
			c["container_name"] = aws.ToString(sr.ContainerName)
		}
		results = append(results, c)
	}
	return results
}

func resourceLoadBalancerHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})

	buf.WriteString(fmt.Sprintf("%s-", m["elb_name"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["container_name"].(string)))
	buf.WriteString(fmt.Sprintf("%d-", m["container_port"].(int)))

	if s := m["target_group_arn"].(string); s != "" {
		buf.WriteString(fmt.Sprintf("%s-", s))
	}

	return create.StringHashcode(buf.String())
}

func serviceCreateWithRetry(ctx context.Context, conn *ecs.Client, input ecs.CreateServiceInput) (*ecs.CreateServiceOutput, error) {
	var output *ecs.CreateServiceOutput
	err := retry.RetryContext(ctx, propagationTimeout+serviceCreateTimeout, func() *retry.RetryError {
		var err error
		output, err = conn.CreateService(ctx, &input)

		if err != nil {
			if errs.IsA[*types.ClusterNotFoundException](err) {
				return retry.RetryableError(err)
			}

			if errs.IsAErrorMessageContains[*types.InvalidParameterException](err, "verify that the ECS service role being passed has the proper permissions") {
				return retry.RetryableError(err)
			}

			if errs.IsAErrorMessageContains[*types.InvalidParameterException](err, "does not have an associated load balancer") {
				return retry.RetryableError(err)
			}

			if errs.IsAErrorMessageContains[*types.InvalidParameterException](err, "Unable to assume the service linked role") {
				return retry.RetryableError(err)
			}

			return retry.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		output, err = conn.CreateService(ctx, &input)
	}

	return output, err
}

func buildFamilyAndRevisionFromARN(arn string) string {
	return strings.Split(arn, "/")[1]
}

// GetRoleNameFromARN parses a role name from a fully qualified ARN
//
// When providing a role name with a path, it must be prefixed with the full path
// including a leading `/`.
// See: https://docs.aws.amazon.com/AmazonECS/latest/APIReference/API_CreateService.html#ECS-CreateService-request-role
//
// Expects an IAM role ARN:
//
//	arn:aws:iam::0123456789:role/EcsService
//	arn:aws:iam::0123456789:role/group/my-role
func GetRoleNameFromARN(arn string) string {
	if parts := strings.Split(arn, "/"); len(parts) == 2 {
		return parts[1]
	} else if len(parts) > 2 {
		return fmt.Sprintf("/%s", strings.Join(parts[1:], "/"))
	}
	return ""
}

// GetClusterNameFromARN parses a cluster name from a fully qualified ARN
//
// Expects an ECS cluster ARN:
//
//	arn:aws:ecs:us-west-2:0123456789:cluster/my-cluster
func GetClusterNameFromARN(arn string) string {
	if parts := strings.Split(arn, "/"); len(parts) == 2 {
		return parts[1]
	}
	return ""
}
