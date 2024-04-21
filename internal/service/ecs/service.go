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

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
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
func ResourceService() *schema.Resource {
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
						names.AttrWeight: {
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
						names.AttrType: {
							Type:         schema.TypeString,
							ForceNew:     true,
							Optional:     true,
							Default:      ecs.DeploymentControllerTypeEcs,
							ValidateFunc: validation.StringInSlice(ecs.DeploymentControllerType_Values(), false),
						},
					},
				},
			},
			"deployment_maximum_percent": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  200,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					if d.Get("scheduling_strategy").(string) == ecs.SchedulingStrategyDaemon && new == "200" {
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
					if d.Get("scheduling_strategy").(string) == ecs.SchedulingStrategyDaemon && new == "100" {
						return true
					}
					return false
				},
			},
			"desired_count": {
				Type:     schema.TypeInt,
				Optional: true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return d.Get("scheduling_strategy").(string) == ecs.SchedulingStrategyDaemon
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
				Type:         schema.TypeString,
				ForceNew:     true,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice(ecs.LaunchType_Values(), false),
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
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrNetworkConfiguration: {
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
						names.AttrSecurityGroups: {
							Type:     schema.TypeSet,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						names.AttrSubnets: {
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
						names.AttrField: {
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
						names.AttrType: {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(ecs.PlacementStrategyType_Values(), false),
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
						names.AttrExpression: {
							Type:     schema.TypeString,
							Optional: true,
						},
						names.AttrType: {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(ecs.PlacementConstraintType_Values(), false),
						},
					},
				},
			},
			"platform_version": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			names.AttrPropagateTags: {
				Type:     schema.TypeString,
				Optional: true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					if old == "NONE" && new == "" {
						return true
					}
					return false
				},
				ValidateFunc: validation.StringInSlice(ecs.PropagateTags_Values(), false),
			},
			"scheduling_strategy": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      ecs.SchedulingStrategyReplica,
				ValidateFunc: validation.StringInSlice(ecs.SchedulingStrategy_Values(), false),
			},
			"service_connect_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrEnabled: {
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
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringInSlice(ecs.LogDriver_Values(), false),
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
												names.AttrName: {
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
						names.AttrNamespace: {
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
												names.AttrDNSName: {
													Type:     schema.TypeString,
													Optional: true,
												},
												names.AttrPort: {
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
									names.AttrTimeout: {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"idle_timeout_seconds": {
													Type:         schema.TypeInt,
													Optional:     true,
													ValidateFunc: validation.IntBetween(0, 2147483647),
												},
												"per_request_timeout_seconds": {
													Type:         schema.TypeInt,
													Optional:     true,
													ValidateFunc: validation.IntBetween(0, 2147483647),
												},
											},
										},
									},
									"tls": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"issuer_cert_authority": {
													Type:     schema.TypeList,
													Required: true,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"aws_pca_authority_arn": {
																Type:         schema.TypeString,
																Required:     true,
																ValidateFunc: verify.ValidARN,
															},
														},
													},
												},
												names.AttrKMSKey: {
													Type:     schema.TypeString,
													Optional: true,
												},
												names.AttrRoleARN: {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: verify.ValidARN,
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
						names.AttrPort: {
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
			names.AttrTriggers: {
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
			"volume_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"managed_ebs_volume": {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"role_arn": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: verify.ValidARN,
									},
									"encrypted": {
										Type:     schema.TypeBool,
										Optional: true,
										Default:  true,
									},
									"file_system_type": {
										Type:         schema.TypeString,
										Optional:     true,
										Default:      ecs.TaskFilesystemTypeXfs,
										ValidateFunc: validation.StringInSlice(ecs.TaskFilesystemType_Values(), false),
									},
									"iops": {
										Type:     schema.TypeInt,
										Optional: true,
									},
									"kms_key_id": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"size_in_gb": {
										Type:     schema.TypeInt,
										Optional: true,
									},
									"snapshot_id": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"throughput": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.IntBetween(0, 1000),
									},
									"volume_type": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
					},
				},
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
	conn := meta.(*conns.AWSClient).ECSConn(ctx)

	deploymentController := expandDeploymentController(d.Get("deployment_controller").([]interface{}))
	deploymentMinimumHealthyPercent := d.Get("deployment_minimum_healthy_percent").(int)
	name := d.Get(names.AttrName).(string)
	schedulingStrategy := d.Get("scheduling_strategy").(string)
	input := ecs.CreateServiceInput{
		CapacityProviderStrategy: expandCapacityProviderStrategy(d.Get("capacity_provider_strategy").(*schema.Set)),
		ClientToken:              aws.String(id.UniqueId()),
		DeploymentConfiguration:  &ecs.DeploymentConfiguration{},
		DeploymentController:     deploymentController,
		EnableECSManagedTags:     aws.Bool(d.Get("enable_ecs_managed_tags").(bool)),
		EnableExecuteCommand:     aws.Bool(d.Get("enable_execute_command").(bool)),
		NetworkConfiguration:     expandNetworkConfiguration(d.Get(names.AttrNetworkConfiguration).([]interface{})),
		SchedulingStrategy:       aws.String(schedulingStrategy),
		ServiceName:              aws.String(name),
		Tags:                     getTagsIn(ctx),
	}

	if v, ok := d.GetOk("alarms"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.DeploymentConfiguration.Alarms = expandAlarms(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("cluster"); ok {
		input.Cluster = aws.String(v.(string))
	}

	if schedulingStrategy == ecs.SchedulingStrategyDaemon && deploymentMinimumHealthyPercent != 100 {
		input.DeploymentConfiguration.MinimumHealthyPercent = aws.Int64(int64(deploymentMinimumHealthyPercent))
	} else if schedulingStrategy == ecs.SchedulingStrategyReplica {
		input.DeploymentConfiguration.MaximumPercent = aws.Int64(int64(d.Get("deployment_maximum_percent").(int)))
		input.DeploymentConfiguration.MinimumHealthyPercent = aws.Int64(int64(deploymentMinimumHealthyPercent))
		input.DesiredCount = aws.Int64(int64(d.Get("desired_count").(int)))
	}

	if v, ok := d.GetOk("deployment_circuit_breaker"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.DeploymentConfiguration.DeploymentCircuitBreaker = expandDeploymentCircuitBreaker(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("health_check_grace_period_seconds"); ok {
		input.HealthCheckGracePeriodSeconds = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("iam_role"); ok {
		input.Role = aws.String(v.(string))
	}

	if v, ok := d.GetOk("launch_type"); ok {
		input.LaunchType = aws.String(v.(string))
		// When creating a service that uses the EXTERNAL deployment controller,
		// you can specify only parameters that aren't controlled at the task set level
		// hence you cannot set LaunchType, not changing the default launch_type from EC2 to empty
		// string to have backward compatibility
		if deploymentController != nil && aws.StringValue(deploymentController.Type) == ecs.DeploymentControllerTypeExternal {
			input.LaunchType = aws.String("")
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

	if v, ok := d.GetOk(names.AttrPropagateTags); ok {
		input.PropagateTags = aws.String(v.(string))
	}

	if v, ok := d.GetOk("service_connect_configuration"); ok && len(v.([]interface{})) > 0 {
		input.ServiceConnectConfiguration = expandServiceConnectConfiguration(v.([]interface{}))
	}

	if v, ok := d.GetOk("volume_configuration"); ok && len(v.([]interface{})) > 0 {
		input.VolumeConfigurations = expandVolumeConfigurations(v.([]interface{}))
	}

	serviceRegistries := d.Get("service_registries").([]interface{})
	if len(serviceRegistries) > 0 {
		srs := make([]*ecs.ServiceRegistry, 0, len(serviceRegistries))
		for _, v := range serviceRegistries {
			raw := v.(map[string]interface{})
			sr := &ecs.ServiceRegistry{
				RegistryArn: aws.String(raw["registry_arn"].(string)),
			}
			if port, ok := raw[names.AttrPort].(int); ok && port != 0 {
				sr.Port = aws.Int64(int64(port))
			}
			if raw, ok := raw["container_port"].(int); ok && raw != 0 {
				sr.ContainerPort = aws.Int64(int64(raw))
			}
			if raw, ok := raw["container_name"].(string); ok && raw != "" {
				sr.ContainerName = aws.String(raw)
			}

			srs = append(srs, sr)
		}
		input.ServiceRegistries = srs
	}

	if v, ok := d.GetOk("task_definition"); ok {
		input.TaskDefinition = aws.String(v.(string))
	}

	output, err := serviceCreateWithRetry(ctx, conn, input)

	// Some partitions (e.g. ISO) may not support tag-on-create.
	if input.Tags != nil && errs.IsUnsupportedOperationInPartitionError(conn.PartitionID, err) {
		input.Tags = nil

		output, err = serviceCreateWithRetry(ctx, conn, input)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating ECS Service (%s): %s", name, err)
	}

	d.SetId(aws.StringValue(output.Service.ServiceArn))

	fn := waitServiceActive
	if d.Get("wait_for_steady_state").(bool) {
		fn = waitServiceStable
	}
	if _, err := fn(ctx, conn, d.Id(), d.Get("cluster").(string), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for ECS Service (%s) create: %s", d.Id(), err)
	}

	// For partitions not supporting tag-on-create, attempt tag after create.
	if tags := getTagsIn(ctx); input.Tags == nil && len(tags) > 0 {
		err := createTags(ctx, conn, d.Id(), tags)

		// If default tags only, continue. Otherwise, error.
		if v, ok := d.GetOk(names.AttrTags); (!ok || len(v.(map[string]interface{})) == 0) && errs.IsUnsupportedOperationInPartitionError(conn.PartitionID, err) {
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
	conn := meta.(*conns.AWSClient).ECSConn(ctx)

	cluster := d.Get("cluster").(string)

	service, err := FindServiceByIDWaitForActive(ctx, conn, d.Id(), cluster)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] ECS Service (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if tfawserr.ErrCodeEquals(err, ecs.ErrCodeClusterNotFoundException) {
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

	d.SetId(aws.StringValue(service.ServiceArn))
	d.Set(names.AttrName, service.ServiceName)

	// When creating a service that uses the EXTERNAL deployment controller,
	// you can specify only parameters that aren't controlled at the task set level
	// hence TaskDefinition will not be set by aws sdk
	if service.TaskDefinition != nil {
		// Save task definition in the same format
		if strings.HasPrefix(d.Get("task_definition").(string), "arn:"+meta.(*conns.AWSClient).Partition+":ecs:") {
			d.Set("task_definition", service.TaskDefinition)
		} else {
			taskDefinition := buildFamilyAndRevisionFromARN(aws.StringValue(service.TaskDefinition))
			d.Set("task_definition", taskDefinition)
		}
	}

	d.Set("scheduling_strategy", service.SchedulingStrategy)
	d.Set("desired_count", service.DesiredCount)
	d.Set("health_check_grace_period_seconds", service.HealthCheckGracePeriodSeconds)
	d.Set("launch_type", service.LaunchType)
	d.Set("enable_ecs_managed_tags", service.EnableECSManagedTags)
	d.Set(names.AttrPropagateTags, service.PropagateTags)
	d.Set("platform_version", service.PlatformVersion)
	d.Set("enable_execute_command", service.EnableExecuteCommand)

	d.Set(names.AttrTriggers, d.Get(names.AttrTriggers))

	// Save cluster in the same format
	if strings.HasPrefix(d.Get("cluster").(string), "arn:"+meta.(*conns.AWSClient).Partition+":ecs:") {
		d.Set("cluster", service.ClusterArn)
	} else {
		clusterARN := GetClusterNameFromARN(aws.StringValue(service.ClusterArn))
		d.Set("cluster", clusterARN)
	}

	// Save IAM role in the same format
	if service.RoleArn != nil {
		if strings.HasPrefix(d.Get("iam_role").(string), "arn:"+meta.(*conns.AWSClient).Partition+":iam:") {
			d.Set("iam_role", service.RoleArn)
		} else {
			roleARN := GetRoleNameFromARN(aws.StringValue(service.RoleArn))
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

	if err := d.Set(names.AttrNetworkConfiguration, flattenNetworkConfiguration(service.NetworkConfiguration)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting network_configuration: %s", err)
	}

	//if err := d.Set("service_connect_configuration", flattenServiceConnectConfiguration(service.ServiceConnectConfiguration)); err != nil {
	//	return sdkdiag.AppendErrorf(diags, "setting service_connect_configuration: %s", err)
	//}

	if err := d.Set("service_registries", flattenServiceRegistries(service.ServiceRegistries)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting service_registries: %s", err)
	}

	setTagsOut(ctx, service.Tags)

	return diags
}

func resourceServiceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ECSConn(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := &ecs.UpdateServiceInput{
			Cluster:            aws.String(d.Get("cluster").(string)),
			ForceNewDeployment: aws.Bool(d.Get("force_new_deployment").(bool)),
			Service:            aws.String(d.Id()),
		}

		if d.HasChange("alarms") {
			if input.DeploymentConfiguration == nil {
				input.DeploymentConfiguration = &ecs.DeploymentConfiguration{}
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
				input.DeploymentConfiguration = &ecs.DeploymentConfiguration{}
			}

			// To remove an existing deployment circuit breaker, specify an empty object.
			input.DeploymentConfiguration.DeploymentCircuitBreaker = &ecs.DeploymentCircuitBreaker{}

			if v, ok := d.GetOk("deployment_circuit_breaker"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
				input.DeploymentConfiguration.DeploymentCircuitBreaker = expandDeploymentCircuitBreaker(v.([]interface{})[0].(map[string]interface{}))
			}
		}

		switch schedulingStrategy := d.Get("scheduling_strategy").(string); schedulingStrategy {
		case ecs.SchedulingStrategyDaemon:
			if d.HasChange("deployment_minimum_healthy_percent") {
				if input.DeploymentConfiguration == nil {
					input.DeploymentConfiguration = &ecs.DeploymentConfiguration{}
				}

				input.DeploymentConfiguration.MinimumHealthyPercent = aws.Int64(int64(d.Get("deployment_minimum_healthy_percent").(int)))
			}
		case ecs.SchedulingStrategyReplica:
			if d.HasChanges("deployment_maximum_percent", "deployment_minimum_healthy_percent") {
				if input.DeploymentConfiguration == nil {
					input.DeploymentConfiguration = &ecs.DeploymentConfiguration{}
				}

				input.DeploymentConfiguration.MaximumPercent = aws.Int64(int64(d.Get("deployment_maximum_percent").(int)))
				input.DeploymentConfiguration.MinimumHealthyPercent = aws.Int64(int64(d.Get("deployment_minimum_healthy_percent").(int)))
			}

			if d.HasChange("desired_count") {
				input.DesiredCount = aws.Int64(int64(d.Get("desired_count").(int)))
			}
		}

		if d.HasChange("enable_ecs_managed_tags") {
			input.EnableECSManagedTags = aws.Bool(d.Get("enable_ecs_managed_tags").(bool))
		}

		if d.HasChange("enable_execute_command") {
			input.EnableExecuteCommand = aws.Bool(d.Get("enable_execute_command").(bool))
		}

		if d.HasChange("health_check_grace_period_seconds") {
			input.HealthCheckGracePeriodSeconds = aws.Int64(int64(d.Get("health_check_grace_period_seconds").(int)))
		}

		if d.HasChange("load_balancer") {
			if v, ok := d.Get("load_balancer").(*schema.Set); ok && v != nil {
				input.LoadBalancers = expandLoadBalancers(v.List())
			}
		}

		if d.HasChange(names.AttrNetworkConfiguration) {
			input.NetworkConfiguration = expandNetworkConfiguration(d.Get(names.AttrNetworkConfiguration).([]interface{}))
		}

		if d.HasChange("ordered_placement_strategy") {
			// Reference: https://docs.aws.amazon.com/AmazonECS/latest/APIReference/API_UpdateService.html#ECS-UpdateService-request-placementStrategy
			// To remove an existing placement strategy, specify an empty object.
			input.PlacementStrategy = []*ecs.PlacementStrategy{}

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
			input.PlacementConstraints = []*ecs.PlacementConstraint{}

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

		if d.HasChange(names.AttrPropagateTags) {
			input.PropagateTags = aws.String(d.Get(names.AttrPropagateTags).(string))
		}

		if d.HasChange("service_connect_configuration") {
			input.ServiceConnectConfiguration = expandServiceConnectConfiguration(d.Get("service_connect_configuration").([]interface{}))
		}

		if d.HasChange("volume_configuration") {
			input.VolumeConfigurations = expandVolumeConfigurations(d.Get("volume_configuration").([]interface{}))
		}

		if d.HasChange("service_registries") {
			input.ServiceRegistries = expandServiceRegistries(d.Get("service_registries").([]interface{}))
		}

		if d.HasChange("task_definition") {
			input.TaskDefinition = aws.String(d.Get("task_definition").(string))
		}

		// Retry due to IAM eventual consistency
		err := retry.RetryContext(ctx, propagationTimeout+serviceUpdateTimeout, func() *retry.RetryError {
			_, err := conn.UpdateServiceWithContext(ctx, input)

			if err != nil {
				if tfawserr.ErrMessageContains(err, ecs.ErrCodeInvalidParameterException, "verify that the ECS service role being passed has the proper permissions") {
					return retry.RetryableError(err)
				}

				if tfawserr.ErrMessageContains(err, ecs.ErrCodeInvalidParameterException, "does not have an associated load balancer") {
					return retry.RetryableError(err)
				}

				return retry.NonRetryableError(err)
			}
			return nil
		})

		if tfresource.TimedOut(err) {
			_, err = conn.UpdateServiceWithContext(ctx, input)
		}

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating ECS Service (%s): %s", d.Id(), err)
		}

		fn := waitServiceActive
		if d.Get("wait_for_steady_state").(bool) {
			fn = waitServiceStable
		}
		if _, err := fn(ctx, conn, d.Id(), d.Get("cluster").(string), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for ECS Service (%s) update: %s", d.Id(), err)
		}
	}

	return append(diags, resourceServiceRead(ctx, d, meta)...)
}

func resourceServiceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ECSConn(ctx)

	service, err := FindServiceNoTagsByID(ctx, conn, d.Id(), d.Get("cluster").(string))

	if tfresource.NotFound(err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ECS Service (%s): %s", d.Id(), err)
	}

	if aws.StringValue(service.Status) == serviceStatusInactive {
		return diags
	}

	// Drain the ECS service
	if aws.StringValue(service.Status) != serviceStatusDraining && aws.StringValue(service.SchedulingStrategy) != ecs.SchedulingStrategyDaemon {
		_, err := conn.UpdateServiceWithContext(ctx, &ecs.UpdateServiceInput{
			Service:      aws.String(d.Id()),
			Cluster:      aws.String(d.Get("cluster").(string)),
			DesiredCount: aws.Int64(0),
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
		_, err := conn.DeleteServiceWithContext(ctx, &input)

		if err != nil {
			if tfawserr.ErrMessageContains(err, ecs.ErrCodeInvalidParameterException, "The service cannot be stopped while deployments are active.") {
				return retry.RetryableError(err)
			}

			if tfawserr.ErrMessageContains(err, "DependencyViolation", "has a dependent object") {
				return retry.RetryableError(err)
			}

			return retry.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		_, err = conn.DeleteServiceWithContext(ctx, &input)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting ECS Service (%s): %s", d.Id(), err)
	}

	if err := waitServiceInactive(ctx, conn, d.Id(), d.Get("cluster").(string), d.Timeout(schema.TimeoutDelete)); err != nil {
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

	if d.HasChange(names.AttrTriggers) && !fnd {
		return d.Clear(names.AttrTriggers)
	}

	if d.HasChange(names.AttrTriggers) && fnd {
		o, n := d.GetChange(names.AttrTriggers)
		if len(o.(map[string]interface{})) > 0 && len(n.(map[string]interface{})) == 0 {
			return d.Clear(names.AttrTriggers)
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

func expandAlarms(tfMap map[string]interface{}) *ecs.DeploymentAlarms {
	if tfMap == nil {
		return nil
	}

	apiObject := &ecs.DeploymentAlarms{}

	if v, ok := tfMap["enable"].(bool); ok {
		apiObject.Enable = aws.Bool(v)
	}

	if v, ok := tfMap["rollback"].(bool); ok {
		apiObject.Rollback = aws.Bool(v)
	}

	if v, ok := tfMap["alarm_names"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.AlarmNames = flex.ExpandStringSet(v)
	}

	return apiObject
}

func flattenAlarms(apiObject *ecs.DeploymentAlarms) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.AlarmNames; v != nil {
		tfMap["alarm_names"] = aws.StringValueSlice(v)
	}

	if v := apiObject.Enable; v != nil {
		tfMap["enable"] = aws.BoolValue(v)
	}

	if v := apiObject.Rollback; v != nil {
		tfMap["rollback"] = aws.BoolValue(v)
	}

	return tfMap
}

func expandDeploymentController(l []interface{}) *ecs.DeploymentController {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	deploymentController := &ecs.DeploymentController{
		Type: aws.String(m[names.AttrType].(string)),
	}

	return deploymentController
}

func flattenDeploymentController(deploymentController *ecs.DeploymentController) []interface{} {
	m := map[string]interface{}{
		names.AttrType: ecs.DeploymentControllerTypeEcs,
	}

	if deploymentController == nil {
		return []interface{}{m}
	}

	m[names.AttrType] = aws.StringValue(deploymentController.Type)

	return []interface{}{m}
}

func expandDeploymentCircuitBreaker(tfMap map[string]interface{}) *ecs.DeploymentCircuitBreaker {
	if tfMap == nil {
		return nil
	}

	apiObject := &ecs.DeploymentCircuitBreaker{}

	apiObject.Enable = aws.Bool(tfMap["enable"].(bool))
	apiObject.Rollback = aws.Bool(tfMap["rollback"].(bool))

	return apiObject
}

func flattenDeploymentCircuitBreaker(apiObject *ecs.DeploymentCircuitBreaker) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	tfMap["enable"] = aws.BoolValue(apiObject.Enable)
	tfMap["rollback"] = aws.BoolValue(apiObject.Rollback)

	return tfMap
}

func flattenNetworkConfiguration(nc *ecs.NetworkConfiguration) []interface{} {
	if nc == nil {
		return nil
	}

	result := make(map[string]interface{})
	result[names.AttrSecurityGroups] = flex.FlattenStringSet(nc.AwsvpcConfiguration.SecurityGroups)
	result[names.AttrSubnets] = flex.FlattenStringSet(nc.AwsvpcConfiguration.Subnets)

	if nc.AwsvpcConfiguration.AssignPublicIp != nil {
		result["assign_public_ip"] = aws.StringValue(nc.AwsvpcConfiguration.AssignPublicIp) == ecs.AssignPublicIpEnabled
	}

	return []interface{}{result}
}

func expandNetworkConfiguration(nc []interface{}) *ecs.NetworkConfiguration {
	if len(nc) == 0 {
		return nil
	}
	awsVpcConfig := &ecs.AwsVpcConfiguration{}
	raw := nc[0].(map[string]interface{})
	if val, ok := raw[names.AttrSecurityGroups]; ok {
		awsVpcConfig.SecurityGroups = flex.ExpandStringSet(val.(*schema.Set))
	}
	awsVpcConfig.Subnets = flex.ExpandStringSet(raw[names.AttrSubnets].(*schema.Set))
	if val, ok := raw["assign_public_ip"].(bool); ok {
		awsVpcConfig.AssignPublicIp = aws.String(ecs.AssignPublicIpDisabled)
		if val {
			awsVpcConfig.AssignPublicIp = aws.String(ecs.AssignPublicIpEnabled)
		}
	}

	return &ecs.NetworkConfiguration{AwsvpcConfiguration: awsVpcConfig}
}

func expandPlacementConstraints(tfList []interface{}) ([]*ecs.PlacementConstraint, error) {
	if len(tfList) == 0 {
		return nil, nil
	}

	var result []*ecs.PlacementConstraint

	for _, tfMapRaw := range tfList {
		if tfMapRaw == nil {
			continue
		}

		tfMap := tfMapRaw.(map[string]interface{})

		apiObject := &ecs.PlacementConstraint{}

		if v, ok := tfMap[names.AttrExpression].(string); ok && v != "" {
			apiObject.Expression = aws.String(v)
		}

		if v, ok := tfMap[names.AttrType].(string); ok && v != "" {
			apiObject.Type = aws.String(v)
		}

		if err := validPlacementConstraint(aws.StringValue(apiObject.Type), aws.StringValue(apiObject.Expression)); err != nil {
			return result, err
		}

		result = append(result, apiObject)
	}

	return result, nil
}

func flattenServicePlacementConstraints(pcs []*ecs.PlacementConstraint) []map[string]interface{} {
	if len(pcs) == 0 {
		return nil
	}
	results := make([]map[string]interface{}, 0)
	for _, pc := range pcs {
		c := make(map[string]interface{})
		c[names.AttrType] = aws.StringValue(pc.Type)
		if pc.Expression != nil {
			c[names.AttrExpression] = aws.StringValue(pc.Expression)
		}

		results = append(results, c)
	}
	return results
}

func expandPlacementStrategy(s []interface{}) ([]*ecs.PlacementStrategy, error) {
	if len(s) == 0 {
		return nil, nil
	}
	pss := make([]*ecs.PlacementStrategy, 0)
	for _, raw := range s {
		p, ok := raw.(map[string]interface{})

		if !ok {
			continue
		}

		t, ok := p[names.AttrType].(string)

		if !ok {
			return nil, fmt.Errorf("missing type attribute in placement strategy configuration block")
		}

		f, ok := p[names.AttrField].(string)

		if !ok {
			return nil, fmt.Errorf("missing field attribute in placement strategy configuration block")
		}

		if err := validPlacementStrategy(t, f); err != nil {
			return nil, err
		}
		ps := &ecs.PlacementStrategy{
			Type: aws.String(t),
		}
		if f != "" {
			// Field must be omitted (i.e. not empty string) for random strategy
			ps.Field = aws.String(f)
		}
		pss = append(pss, ps)
	}
	return pss, nil
}

func flattenPlacementStrategy(pss []*ecs.PlacementStrategy) []interface{} {
	if len(pss) == 0 {
		return nil
	}
	results := make([]interface{}, 0, len(pss))
	for _, ps := range pss {
		c := make(map[string]interface{})
		c[names.AttrType] = aws.StringValue(ps.Type)

		if ps.Field != nil {
			c[names.AttrField] = aws.StringValue(ps.Field)

			// for some fields the API requires lowercase for creation but will return uppercase on query
			if aws.StringValue(ps.Field) == "MEMORY" || aws.StringValue(ps.Field) == "CPU" {
				c[names.AttrField] = strings.ToLower(aws.StringValue(ps.Field))
			}
		}

		results = append(results, c)
	}
	return results
}

func expandServiceConnectConfiguration(sc []interface{}) *ecs.ServiceConnectConfiguration {
	if len(sc) == 0 {
		return &ecs.ServiceConnectConfiguration{
			Enabled: aws.Bool(false),
		}
	}
	raw := sc[0].(map[string]interface{})

	config := &ecs.ServiceConnectConfiguration{}
	if v, ok := raw[names.AttrEnabled].(bool); ok {
		config.Enabled = aws.Bool(v)
	}

	if v, ok := raw["log_configuration"].([]interface{}); ok && len(v) > 0 {
		config.LogConfiguration = expandLogConfiguration(v)
	}

	if v, ok := raw[names.AttrNamespace].(string); ok && v != "" {
		config.Namespace = aws.String(v)
	}

	if v, ok := raw["service"].([]interface{}); ok && len(v) > 0 {
		config.Services = expandServices(v)
	}

	return config
}

func expandLogConfiguration(lc []interface{}) *ecs.LogConfiguration {
	if len(lc) == 0 {
		return &ecs.LogConfiguration{}
	}
	raw := lc[0].(map[string]interface{})

	config := &ecs.LogConfiguration{}
	if v, ok := raw["log_driver"].(string); ok && v != "" {
		config.LogDriver = aws.String(v)
	}
	if v, ok := raw["options"].(map[string]interface{}); ok && len(v) > 0 {
		config.Options = flex.ExpandStringMap(v)
	}
	if v, ok := raw["secret_option"].([]interface{}); ok && len(v) > 0 {
		config.SecretOptions = expandSecretOptions(v)
	}

	return config
}

func expandSecretOptions(sop []interface{}) []*ecs.Secret {
	if len(sop) == 0 {
		return nil
	}

	var out []*ecs.Secret
	for _, item := range sop {
		raw, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		var config ecs.Secret
		if v, ok := raw[names.AttrName].(string); ok && v != "" {
			config.Name = aws.String(v)
		}
		if v, ok := raw["value_from"].(string); ok && v != "" {
			config.ValueFrom = aws.String(v)
		}

		out = append(out, &config)
	}

	return out
}

func expandVolumeConfigurations(vc []interface{}) []*ecs.ServiceVolumeConfiguration {
	if len(vc) == 0 {
		return nil
	}

	vcs := make([]*ecs.ServiceVolumeConfiguration, 0)

	for _, raw := range vc {
		p := raw.(map[string]interface{})

		config := &ecs.ServiceVolumeConfiguration{
			Name: aws.String(p["name"].(string)),
		}

		if v, ok := p["managed_ebs_volume"].([]interface{}); ok && len(v) > 0 {
			config.ManagedEBSVolume = expandManagedEBSVolume(v)
		}
		vcs = append(vcs, config)
	}

	return vcs
}

func expandManagedEBSVolume(ebs []interface{}) *ecs.ServiceManagedEBSVolumeConfiguration {
	if len(ebs) == 0 {
		return &ecs.ServiceManagedEBSVolumeConfiguration{}
	}
	raw := ebs[0].(map[string]interface{})

	config := &ecs.ServiceManagedEBSVolumeConfiguration{}
	if v, ok := raw["role_arn"].(string); ok && v != "" {
		config.RoleArn = aws.String(v)
	}
	if v, ok := raw["encrypted"].(bool); ok {
		config.Encrypted = aws.Bool(v)
	}
	if v, ok := raw["file_system_type"].(string); ok && v != "" {
		config.FilesystemType = aws.String(v)
	}
	if v, ok := raw["iops"].(int); ok && v != 0 {
		config.Iops = aws.Int64(int64(v))
	}
	if v, ok := raw["kms_key_id"].(string); ok && v != "" {
		config.KmsKeyId = aws.String(v)
	}
	if v, ok := raw["size_in_gb"].(int); ok && v != 0 {
		config.SizeInGiB = aws.Int64(int64(v))
	}
	if v, ok := raw["snapshot_id"].(string); ok && v != "" {
		config.SnapshotId = aws.String(v)
	}
	if v, ok := raw["throughput"].(int); ok && v != 0 {
		config.Throughput = aws.Int64(int64(v))
	}
	if v, ok := raw["volume_type"].(string); ok && v != "" {
		config.VolumeType = aws.String(v)
	}

	return config
}

func expandServices(srv []interface{}) []*ecs.ServiceConnectService {
	if len(srv) == 0 {
		return nil
	}

	var out []*ecs.ServiceConnectService
	for _, item := range srv {
		raw, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		var config ecs.ServiceConnectService
		if v, ok := raw["client_alias"].([]interface{}); ok && len(v) > 0 {
			config.ClientAliases = expandClientAliases(v)
		}
		if v, ok := raw["discovery_name"].(string); ok && v != "" {
			config.DiscoveryName = aws.String(v)
		}
		if v, ok := raw["ingress_port_override"].(int); ok && v != 0 {
			config.IngressPortOverride = aws.Int64(int64(v))
		}
		if v, ok := raw["port_name"].(string); ok && v != "" {
			config.PortName = aws.String(v)
		}

		if v, ok := raw[names.AttrTimeout].([]interface{}); ok && len(v) > 0 {
			config.Timeout = expandTimeout(v)
		}

		if v, ok := raw["tls"].([]interface{}); ok && len(v) > 0 {
			config.Tls = expandTLS(v)
		}

		out = append(out, &config)
	}

	return out
}

func expandTimeout(timeout []interface{}) *ecs.TimeoutConfiguration {
	if len(timeout) == 0 {
		return nil
	}

	raw, ok := timeout[0].(map[string]interface{})
	if !ok {
		return nil
	}
	timeoutConfig := &ecs.TimeoutConfiguration{}
	if v, ok := raw["idle_timeout_seconds"].(int); ok {
		timeoutConfig.IdleTimeoutSeconds = aws.Int64(int64(v))
	}
	if v, ok := raw["per_request_timeout_seconds"].(int); ok {
		timeoutConfig.PerRequestTimeoutSeconds = aws.Int64(int64(v))
	}
	return timeoutConfig
}

func expandTLS(tls []interface{}) *ecs.ServiceConnectTlsConfiguration {
	if len(tls) == 0 {
		return nil
	}

	raw, ok := tls[0].(map[string]interface{})
	if !ok {
		return nil
	}
	tlsConfig := &ecs.ServiceConnectTlsConfiguration{}
	if v, ok := raw["issuer_cert_authority"].([]interface{}); ok && len(v) > 0 {
		tlsConfig.IssuerCertificateAuthority = expandIssuerCertAuthority(v)
	}
	if v, ok := raw[names.AttrKMSKey].(string); ok && v != "" {
		tlsConfig.KmsKey = aws.String(v)
	}
	if v, ok := raw[names.AttrRoleARN].(string); ok && v != "" {
		tlsConfig.RoleArn = aws.String(v)
	}
	return tlsConfig
}

func expandIssuerCertAuthority(pca []interface{}) *ecs.ServiceConnectTlsCertificateAuthority {
	if len(pca) == 0 {
		return nil
	}

	raw, ok := pca[0].(map[string]interface{})
	if !ok {
		return nil
	}
	config := &ecs.ServiceConnectTlsCertificateAuthority{}

	if v, ok := raw["aws_pca_authority_arn"].(string); ok && v != "" {
		config.AwsPcaAuthorityArn = aws.String(v)
	}
	return config
}

func expandClientAliases(srv []interface{}) []*ecs.ServiceConnectClientAlias {
	if len(srv) == 0 {
		return nil
	}

	var out []*ecs.ServiceConnectClientAlias
	for _, item := range srv {
		raw, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		var config ecs.ServiceConnectClientAlias
		if v, ok := raw[names.AttrPort].(int); ok {
			config.Port = aws.Int64(int64(v))
		}
		if v, ok := raw[names.AttrDNSName].(string); ok && v != "" {
			config.DnsName = aws.String(v)
		}

		out = append(out, &config)
	}

	return out
}

func flattenServiceRegistries(srs []*ecs.ServiceRegistry) []map[string]interface{} {
	if len(srs) == 0 {
		return nil
	}
	results := make([]map[string]interface{}, 0)
	for _, sr := range srs {
		c := map[string]interface{}{
			"registry_arn": aws.StringValue(sr.RegistryArn),
		}
		if sr.Port != nil {
			c[names.AttrPort] = int(aws.Int64Value(sr.Port))
		}
		if sr.ContainerPort != nil {
			c["container_port"] = int(aws.Int64Value(sr.ContainerPort))
		}
		if sr.ContainerName != nil {
			c["container_name"] = aws.StringValue(sr.ContainerName)
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

func serviceCreateWithRetry(ctx context.Context, conn *ecs.ECS, input ecs.CreateServiceInput) (*ecs.CreateServiceOutput, error) {
	var output *ecs.CreateServiceOutput
	err := retry.RetryContext(ctx, propagationTimeout+serviceCreateTimeout, func() *retry.RetryError {
		var err error
		output, err = conn.CreateServiceWithContext(ctx, &input)

		if err != nil {
			if tfawserr.ErrCodeEquals(err, ecs.ErrCodeClusterNotFoundException) {
				return retry.RetryableError(err)
			}

			if tfawserr.ErrMessageContains(err, ecs.ErrCodeInvalidParameterException, "verify that the ECS service role being passed has the proper permissions") {
				return retry.RetryableError(err)
			}

			if tfawserr.ErrMessageContains(err, ecs.ErrCodeInvalidParameterException, "does not have an associated load balancer") {
				return retry.RetryableError(err)
			}

			if tfawserr.ErrMessageContains(err, ecs.ErrCodeInvalidParameterException, "Unable to assume the service linked role") {
				return retry.RetryableError(err)
			}

			return retry.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		output, err = conn.CreateServiceWithContext(ctx, &input)
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
