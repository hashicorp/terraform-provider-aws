package ecs

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceTaskSet() *schema.Resource {
	return &schema.Resource{
		Create: resourceTaskSetCreate,
		Read:   resourceTaskSetRead,
		Update: resourceTaskSetUpdate,
		Delete: resourceTaskSetDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"service": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"cluster": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"external_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},

			"task_definition": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"task_set_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"network_configuration": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"security_groups": {
							Type:     schema.TypeSet,
							MaxItems: 5,
							Optional: true,
							ForceNew: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"subnets": {
							Type:     schema.TypeSet,
							MaxItems: 16,
							Required: true,
							ForceNew: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"assign_public_ip": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
							ForceNew: true,
						},
					},
				},
			},

			// If you are using the CodeDeploy or an external deployment controller,
			// multiple target groups are not supported.
			// https://docs.aws.amazon.com/AmazonECS/latest/developerguide/register-multiple-targetgroups.html
			"load_balancer": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"load_balancer_name": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"target_group_arn": {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: verify.ValidARN,
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
							ValidateFunc: validation.IsPortNumber,
						},
					},
				},
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
							ForceNew: true,
						},
						"container_port": {
							Type:         schema.TypeInt,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validation.IsPortNumber,
						},
						"port": {
							Type:         schema.TypeInt,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validation.IsPortNumber,
						},
						"registry_arn": {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: verify.ValidARN,
						},
					},
				},
			},

			"launch_type": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				Computed:      true,
				ValidateFunc:  validation.StringInSlice(ecs.LaunchType_Values(), false),
				ConflictsWith: []string{"capacity_provider_strategy"},
			},

			"capacity_provider_strategy": {
				Type:          schema.TypeSet,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"launch_type"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"base": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntBetween(0, 100000),
							ForceNew:     true,
						},

						"capacity_provider": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},

						"weight": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntBetween(0, 1000),
							ForceNew:     true,
						},
					},
				},
			},

			"platform_version": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"scale": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"unit": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      ecs.ScaleUnitPercent,
							ValidateFunc: validation.StringInSlice(ecs.ScaleUnit_Values(), false),
						},
						"value": {
							Type:         schema.TypeFloat,
							Optional:     true,
							ValidateFunc: validation.FloatBetween(0.0, 100.0),
						},
					},
				},
			},

			"force_delete": {
				Type:     schema.TypeBool,
				Optional: true,
			},

			"stability_status": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"tags": tftags.TagsSchema(),

			"tags_all": tftags.TagsSchemaComputed(),

			"wait_until_stable": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			"wait_until_stable_timeout": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "10m",
				ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
					value := v.(string)
					duration, err := time.ParseDuration(value)
					if err != nil {
						errors = append(errors, fmt.Errorf(
							"%q cannot be parsed as a duration: %w", k, err))
					}
					if duration < 0 {
						errors = append(errors, fmt.Errorf(
							"%q must be greater than zero", k))
					}
					return
				},
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceTaskSetCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ECSConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	cluster := d.Get("cluster").(string)
	service := d.Get("service").(string)
	input := &ecs.CreateTaskSetInput{
		ClientToken:    aws.String(resource.UniqueId()),
		Cluster:        aws.String(cluster),
		Service:        aws.String(service),
		TaskDefinition: aws.String(d.Get("task_definition").(string)),
	}

	if len(tags) > 0 {
		input.Tags = Tags(tags.IgnoreAWS())
	}

	if v, ok := d.GetOk("capacity_provider_strategy"); ok && v.(*schema.Set).Len() > 0 {
		input.CapacityProviderStrategy = expandCapacityProviderStrategy(v.(*schema.Set))
	}

	if v, ok := d.GetOk("external_id"); ok {
		input.ExternalId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("launch_type"); ok {
		input.LaunchType = aws.String(v.(string))
	}

	if v, ok := d.GetOk("load_balancer"); ok && v.(*schema.Set).Len() > 0 {
		input.LoadBalancers = expandTaskSetLoadBalancers(v.(*schema.Set).List())
	}

	if v, ok := d.GetOk("network_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.NetworkConfiguration = expandNetworkConfiguration(v.([]interface{}))
	}

	if v, ok := d.GetOk("platform_version"); ok {
		input.PlatformVersion = aws.String(v.(string))
	}

	if v, ok := d.GetOk("scale"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.Scale = expandScale(v.([]interface{}))
	}

	if v, ok := d.GetOk("service_registries"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.ServiceRegistries = expandServiceRegistries(v.([]interface{}))
	}

	output, err := retryTaskSetCreate(conn, input)

	// Some partitions (i.e., ISO) may not support tag-on-create
	if input.Tags != nil && verify.CheckISOErrorTagsUnsupported(conn.PartitionID, err) {
		log.Printf("[WARN] ECS tagging failed creating Task Set with tags: %s. Trying create without tags.", err)
		input.Tags = nil

		output, err = retryTaskSetCreate(conn, input)
	}

	if err != nil {
		return fmt.Errorf("error creating ECS TaskSet: %w", err)
	}

	taskSetId := aws.StringValue(output.TaskSet.Id)

	d.SetId(fmt.Sprintf("%s,%s,%s", taskSetId, service, cluster))

	if d.Get("wait_until_stable").(bool) {
		timeout, _ := time.ParseDuration(d.Get("wait_until_stable_timeout").(string))
		if err := waitTaskSetStable(conn, timeout, taskSetId, service, cluster); err != nil {
			return fmt.Errorf("error waiting for ECS TaskSet (%s) to be stable: %w", d.Id(), err)
		}
	}

	// Some partitions (i.e., ISO) may not support tag-on-create, attempt tag after create
	if input.Tags == nil && len(tags) > 0 {
		err := UpdateTags(conn, aws.StringValue(output.TaskSet.TaskSetArn), nil, tags)

		if v, ok := d.GetOk("tags"); (!ok || len(v.(map[string]interface{})) == 0) && verify.CheckISOErrorTagsUnsupported(conn.PartitionID, err) {
			// If default tags only, log and continue. Otherwise, error.
			log.Printf("[WARN] ECS tagging failed adding tags after create for Task Set (%s): %s", d.Id(), err)
			return resourceTaskSetRead(d, meta)
		}

		if err != nil {
			return fmt.Errorf("ECS tagging failed adding tags after create for Task Set (%s): %w", d.Id(), err)
		}
	}

	return resourceTaskSetRead(d, meta)
}

func resourceTaskSetRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ECSConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	taskSetId, service, cluster, err := TaskSetParseID(d.Id())

	if err != nil {
		return err
	}

	input := &ecs.DescribeTaskSetsInput{
		Cluster:  aws.String(cluster),
		Include:  aws.StringSlice([]string{ecs.TaskSetFieldTags}),
		Service:  aws.String(service),
		TaskSets: aws.StringSlice([]string{taskSetId}),
	}

	out, err := conn.DescribeTaskSets(input)

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, ecs.ErrCodeClusterNotFoundException, ecs.ErrCodeServiceNotFoundException, ecs.ErrCodeTaskSetNotFoundException) {
		log.Printf("[WARN] ECS TaskSet (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	// Some partitions (i.e., ISO) may not support tagging, giving error
	if verify.CheckISOErrorTagsUnsupported(conn.PartitionID, err) {
		log.Printf("[WARN] ECS tagging failed describing Task Set (%s) with tags: %s; retrying without tags", d.Id(), err)

		input.Include = nil
		out, err = conn.DescribeTaskSets(input)
	}

	if err != nil {
		return fmt.Errorf("error reading ECS TaskSet (%s): %w", d.Id(), err)
	}

	if out == nil || len(out.TaskSets) == 0 {
		if d.IsNewResource() {
			return fmt.Errorf("error reading ECS TaskSet (%s): empty output after creation", d.Id())
		}
		log.Printf("[WARN] ECS TaskSet (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	taskSet := out.TaskSets[0]

	d.Set("arn", taskSet.TaskSetArn)
	d.Set("cluster", cluster)
	d.Set("launch_type", taskSet.LaunchType)
	d.Set("platform_version", taskSet.PlatformVersion)
	d.Set("external_id", taskSet.ExternalId)
	d.Set("service", service)
	d.Set("status", taskSet.Status)
	d.Set("stability_status", taskSet.StabilityStatus)
	d.Set("task_definition", taskSet.TaskDefinition)
	d.Set("task_set_id", taskSet.Id)

	if err := d.Set("capacity_provider_strategy", flattenCapacityProviderStrategy(taskSet.CapacityProviderStrategy)); err != nil {
		return fmt.Errorf("error setting capacity_provider_strategy: %w", err)
	}

	if err := d.Set("load_balancer", flattenTaskSetLoadBalancers(taskSet.LoadBalancers)); err != nil {
		return fmt.Errorf("error setting load_balancer: %w", err)
	}

	if err := d.Set("network_configuration", flattenNetworkConfiguration(taskSet.NetworkConfiguration)); err != nil {
		return fmt.Errorf("error setting network_configuration: %w", err)
	}

	if err := d.Set("scale", flattenScale(taskSet.Scale)); err != nil {
		return fmt.Errorf("error setting scale: %w", err)
	}

	if err := d.Set("service_registries", flattenServiceRegistries(taskSet.ServiceRegistries)); err != nil {
		return fmt.Errorf("error setting service_registries: %w", err)
	}

	tags := KeyValueTags(taskSet.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceTaskSetUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ECSConn

	if d.HasChangesExcept("tags", "tags_all") {
		taskSetId, service, cluster, err := TaskSetParseID(d.Id())

		if err != nil {
			return err
		}

		input := &ecs.UpdateTaskSetInput{
			Cluster: aws.String(cluster),
			Service: aws.String(service),
			TaskSet: aws.String(taskSetId),
			Scale:   expandScale(d.Get("scale").([]interface{})),
		}

		_, err = conn.UpdateTaskSet(input)

		if err != nil {
			return fmt.Errorf("error updating ECS TaskSet (%s): %w", d.Id(), err)
		}

		if d.Get("wait_until_stable").(bool) {
			timeout, _ := time.ParseDuration(d.Get("wait_until_stable_timeout").(string))
			if err := waitTaskSetStable(conn, timeout, taskSetId, service, cluster); err != nil {
				return fmt.Errorf("error waiting for ECS TaskSet (%s) to be stable after update: %w", d.Id(), err)
			}
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		err := UpdateTags(conn, d.Get("arn").(string), o, n)

		// Some partitions (i.e., ISO) may not support tagging, giving error
		if verify.CheckISOErrorTagsUnsupported(conn.PartitionID, err) {
			log.Printf("[WARN] ECS tagging failed updating tags for Task Set (%s): %s", d.Id(), err)
			return resourceTaskSetRead(d, meta)
		}

		if err != nil {
			return fmt.Errorf("ECS tagging failed updating tags for Task Set (%s): %w", d.Id(), err)
		}
	}

	return resourceTaskSetRead(d, meta)
}

func resourceTaskSetDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ECSConn

	taskSetId, service, cluster, err := TaskSetParseID(d.Id())

	if err != nil {
		return err
	}

	input := &ecs.DeleteTaskSetInput{
		Cluster: aws.String(cluster),
		Service: aws.String(service),
		TaskSet: aws.String(taskSetId),
		Force:   aws.Bool(d.Get("force_delete").(bool)),
	}

	_, err = conn.DeleteTaskSet(input)

	if tfawserr.ErrCodeEquals(err, ecs.ErrCodeTaskSetNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting ECS TaskSet (%s): %w", d.Id(), err)
	}

	if err := waitTaskSetDeleted(conn, taskSetId, service, cluster); err != nil {
		if tfawserr.ErrCodeEquals(err, ecs.ErrCodeTaskSetNotFoundException) {
			return nil
		}
		return fmt.Errorf("error waiting for ECS TaskSet (%s) to delete: %w", d.Id(), err)
	}

	return nil
}

func TaskSetParseID(id string) (string, string, string, error) {
	parts := strings.Split(id, ",")

	if len(parts) != 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
		return "", "", "", fmt.Errorf("unexpected format of ID (%q), expected TASK_SET_ID,SERVICE,CLUSTER", id)
	}

	return parts[0], parts[1], parts[2], nil
}

func retryTaskSetCreate(conn *ecs.ECS, input *ecs.CreateTaskSetInput) (*ecs.CreateTaskSetOutput, error) {
	outputRaw, err := tfresource.RetryWhen(
		propagationTimeout+taskSetCreateTimeout,
		func() (interface{}, error) {
			return conn.CreateTaskSet(input)
		},
		func(err error) (bool, error) {
			if tfawserr.ErrCodeEquals(err, ecs.ErrCodeClusterNotFoundException, ecs.ErrCodeServiceNotFoundException, ecs.ErrCodeTaskSetNotFoundException) ||
				tfawserr.ErrMessageContains(err, ecs.ErrCodeInvalidParameterException, "does not have an associated load balancer") {
				return true, err
			}
			return false, err
		},
	)

	output, ok := outputRaw.(*ecs.CreateTaskSetOutput)
	if !ok || output == nil || output.TaskSet == nil {
		return nil, fmt.Errorf("error creating ECS TaskSet: empty output")
	}

	return output, err
}
