package aws

import (
	"bytes"
	"fmt"
	"log"
	"math"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/hashicorp/terraform/helper/hashcode"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
)

func resourceAwsEcsTaskSet() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsEcsTaskSetCreate,
		Read:   resourceAwsEcsTaskSetRead,
		Update: resourceAwsEcsTaskSetUpdate,
		Delete: resourceAwsEcsTaskSetDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"cluster": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"external_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},

			"launch_type": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					ecs.LaunchTypeEc2,
					ecs.LaunchTypeFargate,
				}, false),
			},

			"load_balancers": {
				Type:     schema.TypeSet,
				MaxItems: 1,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"elb_name": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"target_group_arn": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"container_name": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"container_port": {
							Type:         schema.TypeInt,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validation.IntBetween(0, 65536),
						},
					},
				},
				Set: resourceAwsEcsLoadBalancerHash,
			},

			"network_configuration": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"awsvpc_configuration": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"assign_public_ip": {
										Type:     schema.TypeString,
										Optional: true,
										Default:  ecs.AssignPublicIpDisabled,
										ValidateFunc: validation.StringInSlice([]string{
											ecs.AssignPublicIpEnabled,
											ecs.AssignPublicIpDisabled,
										}, false),
									},
									"security_groups": {
										Type:     schema.TypeSet,
										MaxItems: 5,
										Optional: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
										Set:      schema.HashString,
									},
									"subnets": {
										Type:     schema.TypeSet,
										MaxItems: 16,
										Required: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
										Set:      schema.HashString,
									},
								},
							},
						},
					},
				},
			},

			"platform_version": {
				Type:     schema.TypeString,
				Optional: true,
				// ForceNew: true,
				Computed: true,
				// Default:  "LATEST",
			},

			"scale": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"unit": {
							Type:     schema.TypeString,
							Optional: true,
							ValidateFunc: validation.StringInSlice([]string{
								ecs.ScaleUnitPercent,
							}, false),
						},
						"value": {
							Type:     schema.TypeFloat,
							Optional: true,
							ValidateFunc: func(i interface{}, k string) (s []string, es []error) {
								v, ok := i.(float64)
								if !ok {
									es = append(es, fmt.Errorf("expected type of %s to be float64", k))
									return
								}

								if v < 0.0 || v > 100.0 {
									es = append(es, fmt.Errorf("expected %s to be in the range (0.0 - 100.0), got %f", k, v))
									return
								}

								return
							},
						},
					},
				},
			},

			"service": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"service_registries": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				ForceNew: true,
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
							Type:     schema.TypeString,
							Optional: true,
							ValidateFunc: validateArn,
						},
					},
				},
			},

			"task_definition": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"force_delete": {
				Type:     schema.TypeBool,
				Optional: true,
			},
		},
	}
}

func resourceAwsEcsTaskSetCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ecsconn

	cluster := d.Get("cluster").(string)
	service := d.Get("service").(string)
	input := &ecs.CreateTaskSetInput{
		ClientToken:    aws.String(resource.UniqueId()),
		Cluster:        aws.String(cluster),
		Service:        aws.String(service),
		TaskDefinition: aws.String(d.Get("task_definition").(string)),
	}

	if v, ok := d.GetOk("external_id"); ok {
		input.ExternalId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("launch_type"); ok {
		input.LaunchType = aws.String(v.(string))
	}

	loadBalancers := expandEcsLoadBalancers(d.Get("load_balancers").(*schema.Set).List())
	if len(loadBalancers) > 0 {
		log.Printf("[DEBUG] Adding ECS load balancers: %s", loadBalancers)
		input.LoadBalancers = loadBalancers
	}

	input.NetworkConfiguration = expandEcsNetworkConfiguration(d.Get("network_configuration").([]interface{}))

	if v, ok := d.GetOk("platform_version"); ok {
		input.PlatformVersion = aws.String(v.(string))
	}

	scale := d.Get("scale").([]interface{})
	if len(scale) > 0 {
		input.Scale = expandAwsEcsScale(scale[0].(map[string]interface{}))
	}

	serviceRegistries := d.Get("service_registries").(*schema.Set).List()

	// serviceRegistries := d.Get("service_registries").([]interface{})
	if len(serviceRegistries) > 0 {
		input.ServiceRegistries = expandAwsEcsServiceRegistries(serviceRegistries)
	}

	log.Printf("[DEBUG] Creating ECS Task set: %s", input)

	// Retry due to AWS IAM & ECS eventual consistency
	var out *ecs.CreateTaskSetOutput
	var err error
	err = resource.Retry(2*time.Minute, func() *resource.RetryError {
		out, err = conn.CreateTaskSet(&input)

		if err != nil {
			if isAWSErr(err, ecs.ErrCodeClusterNotFoundException, "") {
				return resource.RetryableError(err)
			}
			if isAWSErr(err, ecs.ErrCodeServiceNotFoundException, "") {
				return resource.RetryableError(err)
			}
			if isAWSErr(err, ecs.ErrCodeInvalidParameterException, "does not have an associated load balancer") {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("Error creating ECS TaskSet: %s", err)
	}

	taskSet := *out.TaskSet

	if taskSet == nil {
		return fmt.Errorf("Error creating ECS TaskSet: invalid response from AWS")
	}

	log.Printf("[DEBUG] ECS Task set created: %s", *taskSet.TaskSetArn)

	d.SetId(*service.TaskSetArn)

	// d.SetId(fmt.Sprintf("%s|%s|%s", cluser, service, *taskSet.Id))
	return resourceAwsEcsTaskSetRead(d, meta)
}

func resourceAwsEcsTaskSetRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ecsconn

	log.Printf("[DEBUG] Reading ECS task set %s", d.Id())

	cluster, service, taskSetId, err := decodeEcsTaskSetId(d.Id())
	if err != nil {
		return fmt.Errorf("Error reading ECS TaskSet (%s): %s", d.Id(), err)
	}
	input := &ecs.DescribeTaskSetsInput{
		Cluster:  aws.String(cluster),
		Service:  aws.String(service),
		TaskSets: []*string{aws.String(taskSetId)},
	}

	var out *ecs.DescribeTaskSetsOutput
	err := resource.Retry(2*time.Minute, func() *resource.RetryError {
		var err error
		out, err = conn.DescribeTaskSets(&input)
		if err != nil {
			if d.IsNewResource() && isAWSErr(err, ecs.ErrCodeServiceNotFoundException, "") || isAWSErr(err, ecs.ErrCodeClusterNotFoundException, "") {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}

		if len(out.TaskSets) < 1 {
			if d.IsNewResource() {
				return resource.RetryableError(fmt.Errorf("ECS task set not created yet: %q", d.Id()))
			}
			log.Printf("[WARN] ECS Task Set %s not found, removing from state.", d.Id())
			d.SetId("")
			return nil
		}

		return nil
	})

	// after retrying
	if err != nil {
		if isAWSErr(err, ecs.ErrCodeClusterNotFoundException, "") {
			log.Printf("[WARN] ECS TaskSet (%s) not found because cluster(%s) isn't found , removing from state", d.Id(), cluster)
			d.SetId("")
			return nil
		}
		if isAWSErr(err, ecs.ErrCodeServiceNotFoundException, "") {
			log.Printf("[WARN] ECS TaskSet (%s) not found because service(%s) isn't found , removing from state", d.Id(), service)
			d.SetId("")
			return nil
		}
		return err
	}


	if len(out.TaskSets) < 1 {
		log.Printf("[WARN] Removing ECS task set %s because it's gone", d.Id())
		d.SetId("")
		return nil
	}

	if len(resp.TaskSets) != 1 {
		return fmt.Errorf("Error reading # of ECS TaskSet (%s) expected 1, got %d", d.Id(), len(out.TaskSets))
	}

	taskSet := out.TaskSets[0]

	log.Printf("[DEBUG] Received ECS task set %s", taskSet)

	d.Set("arn", taskSet.TaskSetArn)
	d.Set("desired_count", taskSet.ComputedDesiredCount)
	d.Set("launch_type", taskSet.LaunchType)
	d.Set("platform_version", taskSet.PlatformVersion)
	d.Set("external_id", taskSet.ExternalId)

	// Save cluster in the same format
	if strings.HasPrefix(d.Get("cluster").(string), "arn:"+meta.(*AWSClient).partition+":ecs:") {
		d.Set("cluster", taskSet.ClusterArn)
	} else {
		clusterARN := getNameFromARN(*taskSet.ClusterArn)
		d.Set("cluster", clusterARN)
	}

	// Save service in the same format
	if strings.HasPrefix(d.Get("service").(string), "arn:"+meta.(*AWSClient).partition+":ecs:") {
		d.Set("service", taskSet.ServiceArn)
	} else {
		clusterARN := getNameFromARN(*taskSet.ServiceArn)
		d.Set("service", ServiceArn)
	}

	// Save task definition in the same format
	if strings.HasPrefix(d.Get("task_definition").(string), "arn:"+meta.(*AWSClient).partition+":ecs:") {
		d.Set("task_definition", taskSet.TaskDefinition)
	} else {
		taskDefinition := buildFamilyAndRevisionFromARN(*service.TaskDefinition)
		d.Set("task_definition", taskDefinition)
	}

	if taskSet.LoadBalancers != nil {
		d.Set("load_balancer", flattenEcsLoadBalancers(taskSet.LoadBalancers))
	}

	if err := d.Set("scale", flattenAwsEcsScale(service.NetworkConfiguration)); err != nil {
		return fmt.Errorf("Error setting scale for (%s): %s", d.Id(), err)
	}

	if err := d.Set("network_configuration", flattenEcsNetworkConfiguration(service.NetworkConfiguration)); err != nil {
		return fmt.Errorf("Error setting network_configuration for (%s): %s", d.Id(), err)
	}

	if err := d.Set("service_registries", flattenServiceRegistries(service.ServiceRegistries)); err != nil {
		return fmt.Errorf("Error setting service_registries for (%s): %s", d.Id(), err)
	}

	return nil
}

func resourceAwsEcsTaskSetUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ecsconn
	return resourceAwsEcsTaskSetRead(d, meta)
}

func resourceAwsEcsTaskSetDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ecsconn

	// Check if it's not already gone
	resp, err := conn.DescribeTaskSets(&ecs.DescribeTaskSetsInput{
		TaskSets: []*string{aws.String(d.Id())},
		Services: aws.String(d.Get("service").(string)),
		Cluster:  aws.String(d.Get("cluster").(string)),
	})

	if err != nil {
		return fmt.Errorf("Error deleting ECS TaskSet (%s): %s", d.Id(), err)
	}

	// cluster, service, taskSetId, err := decodeEcsTaskSetId(d.Id())

	if len(resp.TaskSets) == 0 {
		log.Printf("[DEBUG] Removing ECS TaskSet from state, %q is already gone", d.Id())
		return nil
	}

	log.Printf("[DEBUG] ECS TaskSet %s is currently %s", d.Id(), *resp.TaskSets[0].Status)

	input := &ecs.DeleteTaskSetInput{
		Cluster: aws.String(d.Get("cluster").(string)),
		Service: aws.String(d.Get("service").(string)),
		TaskSet: aws.String(d.Id()),
	}

	if v, ok := d.GetOk("force_bool"); ok && v.(bool) {
		input.Force = aws.Bool(v.(bool))
	}

	// Wait until the ECS task set is drained
	err = resource.Retry(5*time.Minute, func() *resource.RetryError {
		log.Printf("[DEBUG] Trying to delete ECS task set %s", input)
		_, err := conn.DeleteTaskSet(&input)
		if err != nil {
			if isAWSErr(err, ecs.ErrCodeInvalidParameterException, "The service cannot be stopped while deployments are active.") {
				return resource.RetryableError(err)
			}
			if isAWSErr(err, ecs.ErrCodeTaskSetNotFoundException, "") {
				return nil
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})

	if err != nil {
		if isAWSErr(err, ecs.ErrCodeTaskSetNotFoundException, "") {
			return nil
		}
		return err
	}

	log.Printf("[DEBUG] ECS TaskSet %s deleted.", d.Id())
	return nil
}




func expandEcsDeploymentController(l []interface{}) *ecs.DeploymentController {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	deploymentController := &ecs.DeploymentController{
		Type: aws.String(m["type"].(string)),
	}

	return deploymentController
}

func flattenEcsDeploymentController(deploymentController *ecs.DeploymentController) []interface{} {
	m := map[string]interface{}{
		"type": ecs.DeploymentControllerTypeEcs,
	}

	if deploymentController == nil {
		return []interface{}{m}
	}

	m["type"] = aws.StringValue(deploymentController.Type)

	return []interface{}{m}
}

func flattenEcsNetworkConfiguration(nc *ecs.NetworkConfiguration) []interface{} {
	if nc == nil {
		return nil
	}

	result := make(map[string]interface{})
	result["security_groups"] = schema.NewSet(schema.HashString, flattenStringList(nc.AwsvpcConfiguration.SecurityGroups))
	result["subnets"] = schema.NewSet(schema.HashString, flattenStringList(nc.AwsvpcConfiguration.Subnets))

	if nc.AwsvpcConfiguration.AssignPublicIp != nil {
		result["assign_public_ip"] = *nc.AwsvpcConfiguration.AssignPublicIp == ecs.AssignPublicIpEnabled
	}

	return []interface{}{result}
}

func expandEcsNetworkConfiguration(nc []interface{}) *ecs.NetworkConfiguration {
	if len(nc) == 0 {
		return nil
	}
	awsVpcConfig := &ecs.AwsVpcConfiguration{}
	raw := nc[0].(map[string]interface{})
	if val, ok := raw["security_groups"]; ok {
		awsVpcConfig.SecurityGroups = expandStringSet(val.(*schema.Set))
	}
	awsVpcConfig.Subnets = expandStringSet(raw["subnets"].(*schema.Set))
	if val, ok := raw["assign_public_ip"].(bool); ok {
		awsVpcConfig.AssignPublicIp = aws.String(ecs.AssignPublicIpDisabled)
		if val {
			awsVpcConfig.AssignPublicIp = aws.String(ecs.AssignPublicIpEnabled)
		}
	}

	return &ecs.NetworkConfiguration{AwsvpcConfiguration: awsVpcConfig}
}

func flattenServicePlacementConstraints(pcs []*ecs.PlacementConstraint) []map[string]interface{} {
	if len(pcs) == 0 {
		return nil
	}
	results := make([]map[string]interface{}, 0)
	for _, pc := range pcs {
		c := make(map[string]interface{})
		c["type"] = *pc.Type
		if pc.Expression != nil {
			c["expression"] = *pc.Expression
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
		p := raw.(map[string]interface{})
		t := p["type"].(string)
		f := p["field"].(string)
		if err := validateAwsEcsPlacementStrategy(t, f); err != nil {
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
		c["type"] = *ps.Type

		if ps.Field != nil {
			c["field"] = *ps.Field

			// for some fields the API requires lowercase for creation but will return uppercase on query
			if *ps.Field == "MEMORY" || *ps.Field == "CPU" {
				c["field"] = strings.ToLower(*ps.Field)
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

func resourceAwsEcsServiceUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ecsconn
	updateService := false

	input := ecs.UpdateServiceInput{
		Service: aws.String(d.Id()),
		Cluster: aws.String(d.Get("cluster").(string)),
	}

	schedulingStrategy := d.Get("scheduling_strategy").(string)

	if schedulingStrategy == ecs.SchedulingStrategyDaemon {
		if d.HasChange("deployment_minimum_healthy_percent") {
			updateService = true
			input.DeploymentConfiguration = &ecs.DeploymentConfiguration{
				MinimumHealthyPercent: aws.Int64(int64(d.Get("deployment_minimum_healthy_percent").(int))),
			}
		}
	} else if schedulingStrategy == ecs.SchedulingStrategyReplica {
		if d.HasChange("desired_count") {
			updateService = true
			input.DesiredCount = aws.Int64(int64(d.Get("desired_count").(int)))
		}

		if d.HasChange("deployment_maximum_percent") || d.HasChange("deployment_minimum_healthy_percent") {
			updateService = true
			input.DeploymentConfiguration = &ecs.DeploymentConfiguration{
				MaximumPercent:        aws.Int64(int64(d.Get("deployment_maximum_percent").(int))),
				MinimumHealthyPercent: aws.Int64(int64(d.Get("deployment_minimum_healthy_percent").(int))),
			}
		}
	}

	if d.HasChange("platform_version") {
		updateService = true
		input.PlatformVersion = aws.String(d.Get("platform_version").(string))
	}

	if d.HasChange("health_check_grace_period_seconds") {
		updateService = true
		input.HealthCheckGracePeriodSeconds = aws.Int64(int64(d.Get("health_check_grace_period_seconds").(int)))
	}

	if d.HasChange("task_definition") {
		updateService = true
		input.TaskDefinition = aws.String(d.Get("task_definition").(string))
	}

	if d.HasChange("network_configuration") {
		updateService = true
		input.NetworkConfiguration = expandEcsNetworkConfiguration(d.Get("network_configuration").([]interface{}))
	}

	if updateService {
		log.Printf("[DEBUG] Updating ECS Service (%s): %s", d.Id(), input)
		// Retry due to IAM eventual consistency
		err := resource.Retry(2*time.Minute, func() *resource.RetryError {
			out, err := conn.UpdateService(&input)
			if err != nil {
				if isAWSErr(err, ecs.ErrCodeInvalidParameterException, "Please verify that the ECS service role being passed has the proper permissions.") {
					return resource.RetryableError(err)
				}
				if isAWSErr(err, ecs.ErrCodeInvalidParameterException, "does not have an associated load balancer") {
					return resource.RetryableError(err)
				}
				return resource.NonRetryableError(err)
			}

			log.Printf("[DEBUG] Updated ECS service %s", out.Service)
			return nil
		})
		if err != nil {
			return fmt.Errorf("error updating ECS Service (%s): %s", d.Id(), err)
		}
	}

	if d.HasChange("tags") {
		oldTagsRaw, newTagsRaw := d.GetChange("tags")
		oldTagsMap := oldTagsRaw.(map[string]interface{})
		newTagsMap := newTagsRaw.(map[string]interface{})
		createTags, removeTags := diffTagsECS(tagsFromMapECS(oldTagsMap), tagsFromMapECS(newTagsMap))

		if len(removeTags) > 0 {
			removeTagKeys := make([]*string, len(removeTags))
			for i, removeTag := range removeTags {
				removeTagKeys[i] = removeTag.Key
			}

			input := &ecs.UntagResourceInput{
				ResourceArn: aws.String(d.Id()),
				TagKeys:     removeTagKeys,
			}

			log.Printf("[DEBUG] Untagging ECS Cluster: %s", input)
			if _, err := conn.UntagResource(input); err != nil {
				return fmt.Errorf("error untagging ECS Cluster (%s): %s", d.Id(), err)
			}
		}

		if len(createTags) > 0 {
			input := &ecs.TagResourceInput{
				ResourceArn: aws.String(d.Id()),
				Tags:        createTags,
			}

			log.Printf("[DEBUG] Tagging ECS Cluster: %s", input)
			if _, err := conn.TagResource(input); err != nil {
				return fmt.Errorf("error tagging ECS Cluster (%s): %s", d.Id(), err)
			}
		}
	}

	return resourceAwsEcsServiceRead(d, meta)
}

func resourceAwsEcsServiceDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ecsconn

	// Check if it's not already gone
	resp, err := conn.DescribeServices(&ecs.DescribeServicesInput{
		Services: []*string{aws.String(d.Id())},
		Cluster:  aws.String(d.Get("cluster").(string)),
	})
	if err != nil {
		if isAWSErr(err, ecs.ErrCodeServiceNotFoundException, "") {
			log.Printf("[DEBUG] Removing ECS Service from state, %q is already gone", d.Id())
			return nil
		}
		return err
	}

	if len(resp.Services) == 0 {
		log.Printf("[DEBUG] Removing ECS Service from state, %q is already gone", d.Id())
		return nil
	}

	log.Printf("[DEBUG] ECS service %s is currently %s", d.Id(), *resp.Services[0].Status)

	if *resp.Services[0].Status == "INACTIVE" {
		return nil
	}

	// Drain the ECS service
	if *resp.Services[0].Status != "DRAINING" && aws.StringValue(resp.Services[0].SchedulingStrategy) != ecs.SchedulingStrategyDaemon {
		log.Printf("[DEBUG] Draining ECS service %s", d.Id())
		_, err = conn.UpdateService(&ecs.UpdateServiceInput{
			Service:      aws.String(d.Id()),
			Cluster:      aws.String(d.Get("cluster").(string)),
			DesiredCount: aws.Int64(int64(0)),
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
	err = resource.Retry(5*time.Minute, func() *resource.RetryError {
		log.Printf("[DEBUG] Trying to delete ECS service %s", input)
		_, err := conn.DeleteService(&input)
		if err != nil {
			if isAWSErr(err, ecs.ErrCodeInvalidParameterException, "The service cannot be stopped while deployments are active.") {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})

	if err != nil {
		return err
	}

	// Wait until it's deleted
	wait := resource.StateChangeConf{
		Pending:    []string{"ACTIVE", "DRAINING"},
		Target:     []string{"INACTIVE"},
		Timeout:    10 * time.Minute,
		MinTimeout: 1 * time.Second,
		Refresh: func() (interface{}, string, error) {
			log.Printf("[DEBUG] Checking if ECS service %s is INACTIVE", d.Id())
			resp, err := conn.DescribeServices(&ecs.DescribeServicesInput{
				Services: []*string{aws.String(d.Id())},
				Cluster:  aws.String(d.Get("cluster").(string)),
			})
			if err != nil {
				return resp, "FAILED", err
			}

			log.Printf("[DEBUG] ECS service (%s) is currently %q", d.Id(), *resp.Services[0].Status)
			return resp, *resp.Services[0].Status, nil
		},
	}

	_, err = wait.WaitForState()
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] ECS service %s deleted.", d.Id())
	return nil
}

func resourceAwsEcsLoadBalancerHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})

	buf.WriteString(fmt.Sprintf("%s-", m["elb_name"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["container_name"].(string)))
	buf.WriteString(fmt.Sprintf("%d-", m["container_port"].(int)))

	if s := m["target_group_arn"].(string); s != "" {
		buf.WriteString(fmt.Sprintf("%s-", s))
	}

	return hashcode.String(buf.String())
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
