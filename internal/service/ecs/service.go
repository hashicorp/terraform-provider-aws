package ecs

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"math"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceService() *schema.Resource {
	return &schema.Resource{
		Create: resourceServiceCreate,
		Read:   resourceServiceRead,
		Update: resourceServiceUpdate,
		Delete: resourceServiceDelete,
		Importer: &schema.ResourceImporter{
			State: resourceServiceImport,
		},

		Timeouts: &schema.ResourceTimeout{
			Delete: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
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
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				// Ignore missing configuration block
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					if old == "1" && new == "0" {
						return true
					}
					return false
				},
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
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				// Ignore missing configuration block
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					if old == "1" && new == "0" {
						return true
					}
					return false
				},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
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
							Set:      schema.HashString,
						},
						"subnets": {
							Type:     schema.TypeSet,
							Required: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Set:      schema.HashString,
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
						"expression": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"type": {
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
			"propagate_tags": {
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
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
			"task_definition": {
				Type:     schema.TypeString,
				Optional: true,
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
		),
	}
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

func resourceServiceImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
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

func resourceServiceCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ECSConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	deploymentMinimumHealthyPercent := d.Get("deployment_minimum_healthy_percent").(int)
	schedulingStrategy := d.Get("scheduling_strategy").(string)
	deploymentController := expandDeploymentController(d.Get("deployment_controller").([]interface{}))

	input := ecs.CreateServiceInput{
		ClientToken:          aws.String(resource.UniqueId()),
		DeploymentController: deploymentController,
		SchedulingStrategy:   aws.String(schedulingStrategy),
		ServiceName:          aws.String(d.Get("name").(string)),
		TaskDefinition:       aws.String(d.Get("task_definition").(string)),
		EnableECSManagedTags: aws.Bool(d.Get("enable_ecs_managed_tags").(bool)),
		EnableExecuteCommand: aws.Bool(d.Get("enable_execute_command").(bool)),
	}

	if schedulingStrategy == ecs.SchedulingStrategyDaemon && deploymentMinimumHealthyPercent != 100 {
		input.DeploymentConfiguration = &ecs.DeploymentConfiguration{
			MinimumHealthyPercent: aws.Int64(int64(deploymentMinimumHealthyPercent)),
		}
	} else if schedulingStrategy == ecs.SchedulingStrategyReplica {
		input.DeploymentConfiguration = &ecs.DeploymentConfiguration{
			MaximumPercent:        aws.Int64(int64(d.Get("deployment_maximum_percent").(int))),
			MinimumHealthyPercent: aws.Int64(int64(deploymentMinimumHealthyPercent)),
		}

		input.DesiredCount = aws.Int64(int64(d.Get("desired_count").(int)))
	}

	if v, ok := d.GetOk("deployment_circuit_breaker"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.DeploymentConfiguration = &ecs.DeploymentConfiguration{}
		input.DeploymentConfiguration.DeploymentCircuitBreaker = expandDeploymentCircuitBreaker(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("cluster"); ok {
		input.Cluster = aws.String(v.(string))
	}

	if v, ok := d.GetOk("health_check_grace_period_seconds"); ok {
		input.HealthCheckGracePeriodSeconds = aws.Int64(int64(v.(int)))
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

	if v, ok := d.GetOk("propagate_tags"); ok {
		input.PropagateTags = aws.String(v.(string))
	}

	if v, ok := d.GetOk("platform_version"); ok {
		input.PlatformVersion = aws.String(v.(string))
	}

	input.CapacityProviderStrategy = expandCapacityProviderStrategy(d.Get("capacity_provider_strategy").(*schema.Set))

	loadBalancers := expandLoadBalancers(d.Get("load_balancer").(*schema.Set).List())
	if len(loadBalancers) > 0 {
		log.Printf("[DEBUG] Adding ECS load balancers: %s", loadBalancers)
		input.LoadBalancers = loadBalancers
	}
	if v, ok := d.GetOk("iam_role"); ok {
		input.Role = aws.String(v.(string))
	}

	input.NetworkConfiguration = expandNetworkConfiguration(d.Get("network_configuration").([]interface{}))

	if v, ok := d.GetOk("ordered_placement_strategy"); ok {
		ps, err := expandPlacementStrategy(v.([]interface{}))

		if err != nil {
			return err
		}

		input.PlacementStrategy = ps
	}

	if v, ok := d.Get("placement_constraints").(*schema.Set); ok {
		pc, err := expandPlacementConstraints(v.List())

		if err != nil {
			return err
		}

		input.PlacementConstraints = pc
	}

	serviceRegistries := d.Get("service_registries").([]interface{})
	if len(serviceRegistries) > 0 {
		srs := make([]*ecs.ServiceRegistry, 0, len(serviceRegistries))
		for _, v := range serviceRegistries {
			raw := v.(map[string]interface{})
			sr := &ecs.ServiceRegistry{
				RegistryArn: aws.String(raw["registry_arn"].(string)),
			}
			if port, ok := raw["port"].(int); ok && port != 0 {
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

	if len(tags) > 0 {
		input.Tags = Tags(tags.IgnoreAWS()) // tags field doesn't exist in all partitions
	}

	log.Printf("[DEBUG] Creating ECS Service: %s", input)

	output, err := retryServiceCreate(conn, input)

	// Some partitions (i.e., ISO) may not support tag-on-create
	if input.Tags != nil && verify.CheckISOErrorTagsUnsupported(err) {
		log.Printf("[WARN] ECS tagging failed creating Service (%s) with tags: %s. Trying create without tags.", d.Get("name").(string), err)
		input.Tags = nil

		output, err = retryServiceCreate(conn, input)
	}

	if err != nil {
		return fmt.Errorf("failed creating ECS service (%s): %w", d.Get("name").(string), err)
	}

	if output == nil || output.Service == nil {
		return fmt.Errorf("error creating ECS service: empty response")
	}

	log.Printf("[DEBUG] ECS service created: %s", aws.StringValue(output.Service.ServiceArn))
	d.SetId(aws.StringValue(output.Service.ServiceArn))

	cluster := d.Get("cluster").(string)

	if d.Get("wait_for_steady_state").(bool) {
		if err := waitServiceStable(conn, d.Id(), cluster); err != nil {
			return fmt.Errorf("error waiting for ECS service (%s) to reach steady state after creation: %w", d.Id(), err)
		}
	} else {
		if _, err := waitServiceDescribeReady(conn, d.Id(), cluster); err != nil {
			return fmt.Errorf("error waiting for ECS service (%s) to become active after creation: %w", d.Id(), err)
		}
	}

	// Some partitions (i.e., ISO) may not support tag-on-create, attempt tag after create
	if input.Tags == nil && len(tags) > 0 {
		err := UpdateTags(conn, d.Id(), nil, tags)

		// If default tags only, log and continue. Otherwise, error.
		if v, ok := d.GetOk("tags"); (!ok || len(v.(map[string]interface{})) == 0) && verify.CheckISOErrorTagsUnsupported(err) {
			log.Printf("[WARN] ECS tagging failed adding tags after create for Service (%s): %s", d.Id(), err)
			return resourceServiceRead(d, meta)
		}

		if err != nil {
			return fmt.Errorf("ECS tagging failed adding tags after create for Service (%s): %w", d.Id(), err)
		}
	}

	return resourceServiceRead(d, meta)
}

func resourceServiceRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ECSConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	log.Printf("[DEBUG] Reading ECS service %s", d.Id())
	input := ecs.DescribeServicesInput{
		Cluster:  aws.String(d.Get("cluster").(string)),
		Include:  aws.StringSlice([]string{ecs.ServiceFieldTags}),
		Services: aws.StringSlice([]string{d.Id()}),
	}

	output, err := conn.DescribeServices(&input)

	// Some partitions (i.e., ISO) may not support tagging, giving error
	if verify.CheckISOErrorTagsUnsupported(err) {
		log.Printf("[WARN] ECS tagging failed describing Service (%s) with tags: %s; retrying without tags", d.Id(), err)

		input.Include = nil
		output, err = conn.DescribeServices(&input)
	}

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, ecs.ErrCodeServiceNotFoundException) {
		log.Printf("[WARN] ECS service (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		log.Printf("[DEBUG] Waiting for ECS Service (%s) to become active", d.Id())
		output, err = waitServiceDescribeReady(conn, d.Id(), d.Get("cluster").(string))
	}

	if tfawserr.ErrCodeEquals(err, ecs.ErrCodeClusterNotFoundException) {
		log.Printf("[WARN] ECS Service %s parent cluster %s not found, removing from state.", d.Id(), d.Get("cluster").(string))
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading ECS service (%s): %w", d.Id(), err)
	}

	if len(output.Services) < 1 {
		if d.IsNewResource() {
			return fmt.Errorf("ECS service not created: %q", d.Id())
		}
		log.Printf("[WARN] Removing ECS service %s (%s) because it's gone", d.Get("name").(string), d.Id())
		d.SetId("")
		return nil
	}

	service := output.Services[0]

	// Status==INACTIVE means deleted service
	if aws.StringValue(service.Status) == serviceStatusInactive {
		log.Printf("[WARN] Removing ECS service %q because it's INACTIVE", aws.StringValue(service.ServiceArn))
		d.SetId("")
		return nil
	}

	log.Printf("[DEBUG] Received ECS service %s", service)

	d.SetId(aws.StringValue(service.ServiceArn))
	d.Set("name", service.ServiceName)

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
	d.Set("propagate_tags", service.PropagateTags)
	d.Set("platform_version", service.PlatformVersion)
	d.Set("enable_execute_command", service.EnableExecuteCommand)

	// Save cluster in the same format
	if strings.HasPrefix(d.Get("cluster").(string), "arn:"+meta.(*conns.AWSClient).Partition+":ecs:") {
		d.Set("cluster", service.ClusterArn)
	} else {
		clusterARN := getNameFromARN(aws.StringValue(service.ClusterArn))
		d.Set("cluster", clusterARN)
	}

	// Save IAM role in the same format
	if service.RoleArn != nil {
		if strings.HasPrefix(d.Get("iam_role").(string), "arn:"+meta.(*conns.AWSClient).Partition+":iam:") {
			d.Set("iam_role", service.RoleArn)
		} else {
			roleARN := getNameFromARN(aws.StringValue(service.RoleArn))
			d.Set("iam_role", roleARN)
		}
	}

	if service.DeploymentConfiguration != nil {
		d.Set("deployment_maximum_percent", service.DeploymentConfiguration.MaximumPercent)
		d.Set("deployment_minimum_healthy_percent", service.DeploymentConfiguration.MinimumHealthyPercent)

		if service.DeploymentConfiguration.DeploymentCircuitBreaker != nil {
			if err := d.Set("deployment_circuit_breaker", []interface{}{flattenDeploymentCircuitBreaker(service.DeploymentConfiguration.DeploymentCircuitBreaker)}); err != nil {
				return fmt.Errorf("error setting deployment_circuit_break: %w", err)
			}
		} else {
			d.Set("deployment_circuit_breaker", nil)
		}
	}

	if err := d.Set("deployment_controller", flattenDeploymentController(service.DeploymentController)); err != nil {
		return fmt.Errorf("error setting deployment_controller for (%s): %w", d.Id(), err)
	}

	if service.LoadBalancers != nil {
		d.Set("load_balancer", flattenLoadBalancers(service.LoadBalancers))
	}

	if err := d.Set("capacity_provider_strategy", flattenCapacityProviderStrategy(service.CapacityProviderStrategy)); err != nil {
		return fmt.Errorf("error setting capacity_provider_strategy: %w", err)
	}

	if err := d.Set("ordered_placement_strategy", flattenPlacementStrategy(service.PlacementStrategy)); err != nil {
		return fmt.Errorf("error setting ordered_placement_strategy: %w", err)
	}

	if err := d.Set("placement_constraints", flattenServicePlacementConstraints(service.PlacementConstraints)); err != nil {
		log.Printf("[ERR] Error setting placement_constraints for (%s): %s", d.Id(), err)
	}

	if err := d.Set("network_configuration", flattenNetworkConfiguration(service.NetworkConfiguration)); err != nil {
		return fmt.Errorf("error setting network_configuration for (%s): %w", d.Id(), err)
	}

	if err := d.Set("service_registries", flattenServiceRegistries(service.ServiceRegistries)); err != nil {
		return fmt.Errorf("error setting service_registries for (%s): %w", d.Id(), err)
	}

	tags := KeyValueTags(service.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func expandDeploymentController(l []interface{}) *ecs.DeploymentController {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	deploymentController := &ecs.DeploymentController{
		Type: aws.String(m["type"].(string)),
	}

	return deploymentController
}

func flattenDeploymentController(deploymentController *ecs.DeploymentController) []interface{} {
	m := map[string]interface{}{
		"type": ecs.DeploymentControllerTypeEcs,
	}

	if deploymentController == nil {
		return []interface{}{m}
	}

	m["type"] = aws.StringValue(deploymentController.Type)

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
	result["security_groups"] = flex.FlattenStringSet(nc.AwsvpcConfiguration.SecurityGroups)
	result["subnets"] = flex.FlattenStringSet(nc.AwsvpcConfiguration.Subnets)

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
	if val, ok := raw["security_groups"]; ok {
		awsVpcConfig.SecurityGroups = flex.ExpandStringSet(val.(*schema.Set))
	}
	awsVpcConfig.Subnets = flex.ExpandStringSet(raw["subnets"].(*schema.Set))
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

		if v, ok := tfMap["expression"].(string); ok && v != "" {
			apiObject.Expression = aws.String(v)
		}

		if v, ok := tfMap["type"].(string); ok && v != "" {
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
		c["type"] = aws.StringValue(pc.Type)
		if pc.Expression != nil {
			c["expression"] = aws.StringValue(pc.Expression)
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
		c["type"] = aws.StringValue(ps.Type)

		if ps.Field != nil {
			c["field"] = aws.StringValue(ps.Field)

			// for some fields the API requires lowercase for creation but will return uppercase on query
			if aws.StringValue(ps.Field) == "MEMORY" || aws.StringValue(ps.Field) == "CPU" {
				c["field"] = strings.ToLower(aws.StringValue(ps.Field))
			}
		}

		results = append(results, c)
	}
	return results
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
			c["port"] = int(aws.Int64Value(sr.Port))
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

func resourceServiceUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ECSConn

	if d.HasChangesExcept("tags", "tags_all") {
		input := &ecs.UpdateServiceInput{
			Cluster:            aws.String(d.Get("cluster").(string)),
			ForceNewDeployment: aws.Bool(d.Get("force_new_deployment").(bool)),
			Service:            aws.String(d.Id()),
		}

		schedulingStrategy := d.Get("scheduling_strategy").(string)

		if schedulingStrategy == ecs.SchedulingStrategyDaemon {
			if d.HasChange("deployment_minimum_healthy_percent") {
				input.DeploymentConfiguration = &ecs.DeploymentConfiguration{
					MinimumHealthyPercent: aws.Int64(int64(d.Get("deployment_minimum_healthy_percent").(int))),
				}
			}
		} else if schedulingStrategy == ecs.SchedulingStrategyReplica {
			if d.HasChange("desired_count") {
				input.DesiredCount = aws.Int64(int64(d.Get("desired_count").(int)))
			}

			if d.HasChanges("deployment_maximum_percent", "deployment_minimum_healthy_percent") {
				input.DeploymentConfiguration = &ecs.DeploymentConfiguration{
					MaximumPercent:        aws.Int64(int64(d.Get("deployment_maximum_percent").(int))),
					MinimumHealthyPercent: aws.Int64(int64(d.Get("deployment_minimum_healthy_percent").(int))),
				}
			}
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

		if d.HasChange("ordered_placement_strategy") {
			// Reference: https://docs.aws.amazon.com/AmazonECS/latest/APIReference/API_UpdateService.html#ECS-UpdateService-request-placementStrategy
			// To remove an existing placement strategy, specify an empty object.
			input.PlacementStrategy = []*ecs.PlacementStrategy{}

			if v, ok := d.GetOk("ordered_placement_strategy"); ok && len(v.([]interface{})) > 0 {
				ps, err := expandPlacementStrategy(v.([]interface{}))

				if err != nil {
					return err
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
					return err
				}

				input.PlacementConstraints = pc
			}
		}

		if d.HasChange("platform_version") {
			input.PlatformVersion = aws.String(d.Get("platform_version").(string))
		}

		if d.HasChange("health_check_grace_period_seconds") {
			input.HealthCheckGracePeriodSeconds = aws.Int64(int64(d.Get("health_check_grace_period_seconds").(int)))
		}

		if d.HasChange("task_definition") {
			input.TaskDefinition = aws.String(d.Get("task_definition").(string))
		}

		if d.HasChange("network_configuration") {
			input.NetworkConfiguration = expandNetworkConfiguration(d.Get("network_configuration").([]interface{}))
		}

		if d.HasChange("capacity_provider_strategy") {
			input.CapacityProviderStrategy = expandCapacityProviderStrategy(d.Get("capacity_provider_strategy").(*schema.Set))
		}

		if d.HasChange("enable_execute_command") {
			input.EnableExecuteCommand = aws.Bool(d.Get("enable_execute_command").(bool))
		}

		if d.HasChange("enable_ecs_managed_tags") {
			input.EnableECSManagedTags = aws.Bool(d.Get("enable_ecs_managed_tags").(bool))
		}

		if d.HasChange("load_balancer") {
			if v, ok := d.Get("load_balancer").(*schema.Set); ok && v != nil {
				input.LoadBalancers = expandLoadBalancers(v.List())
			}
		}

		if d.HasChange("propagate_tags") {
			input.PropagateTags = aws.String(d.Get("propagate_tags").(string))
		}

		if d.HasChange("service_registries") {
			input.ServiceRegistries = expandServiceRegistries(d.Get("service_registries").([]interface{}))
		}

		log.Printf("[DEBUG] Updating ECS Service (%s): %s", d.Id(), input)
		// Retry due to IAM eventual consistency
		err := resource.Retry(propagationTimeout+serviceUpdateTimeout, func() *resource.RetryError {
			_, err := conn.UpdateService(input)

			if err != nil {
				if tfawserr.ErrMessageContains(err, ecs.ErrCodeInvalidParameterException, "verify that the ECS service role being passed has the proper permissions") {
					return resource.RetryableError(err)
				}

				if tfawserr.ErrMessageContains(err, ecs.ErrCodeInvalidParameterException, "does not have an associated load balancer") {
					return resource.RetryableError(err)
				}

				return resource.NonRetryableError(err)
			}
			return nil
		})

		if tfresource.TimedOut(err) {
			_, err = conn.UpdateService(input)
		}

		if err != nil {
			return fmt.Errorf("error updating ECS Service (%s): %w", d.Id(), err)
		}

		cluster := d.Get("cluster").(string)
		if d.Get("wait_for_steady_state").(bool) {
			if err := waitServiceStable(conn, d.Id(), cluster); err != nil {
				return fmt.Errorf("error waiting for ECS service (%s) to reach steady state after update: %w", d.Id(), err)
			}
		} else {
			if _, err := waitServiceDescribeReady(conn, d.Id(), cluster); err != nil {
				return fmt.Errorf("error waiting for ECS service (%s) to become active after update: %w", d.Id(), err)
			}
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		err := UpdateTags(conn, d.Id(), o, n)

		// Some partitions (i.e., ISO) may not support tagging, giving error
		if verify.CheckISOErrorTagsUnsupported(err) {
			log.Printf("[WARN] ECS tagging failed updating tags for Service (%s): %s", d.Id(), err)
			return resourceServiceRead(d, meta)
		}

		if err != nil {
			return fmt.Errorf("ECS tagging failed updating tags for Service (%s): %w", d.Id(), err)
		}
	}

	return resourceServiceRead(d, meta)
}

func resourceServiceDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ECSConn

	// Check if it's not already gone
	output, err := conn.DescribeServices(&ecs.DescribeServicesInput{
		Services: aws.StringSlice([]string{d.Id()}),
		Cluster:  aws.String(d.Get("cluster").(string)),
	})

	if err != nil {
		if tfawserr.ErrCodeEquals(err, ecs.ErrCodeServiceNotFoundException) {
			log.Printf("[DEBUG] Removing ECS Service from state, %q is already gone", d.Id())
			return nil
		}
		return err
	}

	if len(output.Services) == 0 {
		log.Printf("[DEBUG] Removing ECS Service from state, %q is already gone", d.Id())
		return nil
	}

	log.Printf("[DEBUG] ECS service %s is currently %s", d.Id(), aws.StringValue(output.Services[0].Status))

	if aws.StringValue(output.Services[0].Status) == "INACTIVE" {
		return nil
	}

	// Drain the ECS service
	if aws.StringValue(output.Services[0].Status) != "DRAINING" && aws.StringValue(output.Services[0].SchedulingStrategy) != ecs.SchedulingStrategyDaemon {
		log.Printf("[DEBUG] Draining ECS service %s", d.Id())
		_, err = conn.UpdateService(&ecs.UpdateServiceInput{
			Service:      aws.String(d.Id()),
			Cluster:      aws.String(d.Get("cluster").(string)),
			DesiredCount: aws.Int64(0),
		})
		if err != nil {
			return err
		}
	}

	input := ecs.DeleteServiceInput{
		Service: aws.String(d.Id()),
		Cluster: aws.String(d.Get("cluster").(string)),
	}
	// Wait until the ECS service is drained
	err = resource.Retry(d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		_, err := conn.DeleteService(&input)

		if err != nil {
			if tfawserr.ErrMessageContains(err, ecs.ErrCodeInvalidParameterException, "The service cannot be stopped while deployments are active.") {
				return resource.RetryableError(err)
			}

			if tfawserr.ErrMessageContains(err, "DependencyViolation", "has a dependent object") {
				return resource.RetryableError(err)
			}

			return resource.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		_, err = conn.DeleteService(&input)
	}

	if err != nil {
		return fmt.Errorf("error deleting ECS service (%s): %w", d.Id(), err)
	}

	if err := waitServiceInactive(conn, d.Id(), d.Get("cluster").(string)); err != nil {
		return fmt.Errorf("error deleting ECS service (%s): %w", d.Id(), err)
	}

	log.Printf("[DEBUG] ECS service %s deleted.", d.Id())
	return nil
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

func retryServiceCreate(conn *ecs.ECS, input ecs.CreateServiceInput) (*ecs.CreateServiceOutput, error) {
	var output *ecs.CreateServiceOutput
	err := resource.Retry(propagationTimeout+serviceCreateTimeout, func() *resource.RetryError {
		var err error
		output, err = conn.CreateService(&input)

		if err != nil {
			if tfawserr.ErrCodeEquals(err, ecs.ErrCodeClusterNotFoundException) {
				return resource.RetryableError(err)
			}

			if tfawserr.ErrMessageContains(err, ecs.ErrCodeInvalidParameterException, "verify that the ECS service role being passed has the proper permissions") {
				return resource.RetryableError(err)
			}

			if tfawserr.ErrMessageContains(err, ecs.ErrCodeInvalidParameterException, "does not have an associated load balancer") {
				return resource.RetryableError(err)
			}

			if tfawserr.ErrMessageContains(err, ecs.ErrCodeInvalidParameterException, "Unable to assume the service linked role") {
				return resource.RetryableError(err)
			}

			return resource.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		output, err = conn.CreateService(&input)
	}

	return output, err
}

func buildFamilyAndRevisionFromARN(arn string) string {
	return strings.Split(arn, "/")[1]
}

// Expects the following ARNs:
// arn:aws:iam::0123456789:role/EcsService
// arn:aws:ecs:us-west-2:0123456789:cluster/radek-cluster
func getNameFromARN(arn string) string {
	return strings.Split(arn, "/")[1]
}
