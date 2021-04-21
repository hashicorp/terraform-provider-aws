package aws

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/batch"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/naming"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/batch/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/batch/waiter"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

func resourceAwsBatchComputeEnvironment() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsBatchComputeEnvironmentCreate,
		Read:   resourceAwsBatchComputeEnvironmentRead,
		Update: resourceAwsBatchComputeEnvironmentUpdate,
		Delete: resourceAwsBatchComputeEnvironmentDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"compute_environment_name": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"compute_environment_name_prefix"},
				ValidateFunc:  validateBatchName,
			},
			"compute_environment_name_prefix": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"compute_environment_name"},
				ValidateFunc:  validateBatchPrefix,
			},
			// TODO Required for upper(type) == MANAGED
			"compute_resources": {
				Type:     schema.TypeList,
				Optional: true,
				MinItems: 0,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"allocation_strategy": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
							StateFunc: func(val interface{}) string {
								return strings.ToUpper(val.(string))
							},
							ValidateFunc: validation.StringInSlice(batch.CRAllocationStrategy_Values(), true),
						},
						"bid_percentage": {
							Type:     schema.TypeInt,
							Optional: true,
							ForceNew: true,
						},
						"desired_vcpus": {
							Type:     schema.TypeInt,
							Optional: true,
							Computed: true,
						},
						"ec2_key_pair": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"image_id": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						// TODO Required for EC2
						"instance_role": {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validateArn,
						},
						// TODO Required for EC2
						"instance_type": {
							Type:     schema.TypeSet,
							Optional: true,
							ForceNew: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"launch_template": {
							Type:     schema.TypeList,
							Optional: true,
							ForceNew: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"launch_template_id": {
										Type:          schema.TypeString,
										Optional:      true,
										ConflictsWith: []string{"compute_resources.0.launch_template.0.launch_template_name"},
										ForceNew:      true,
									},
									"launch_template_name": {
										Type:          schema.TypeString,
										Optional:      true,
										ConflictsWith: []string{"compute_resources.0.launch_template.0.launch_template_id"},
										ForceNew:      true,
									},
									"version": {
										Type:     schema.TypeString,
										Optional: true,
										ForceNew: true,
									},
								},
							},
						},
						"max_vcpus": {
							Type:     schema.TypeInt,
							Required: true,
						},
						// TODO Required for SPOT
						"min_vcpus": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						// TODO Can be updated for FARGATE
						"security_group_ids": {
							Type:     schema.TypeSet,
							Required: true,
							ForceNew: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						// TODO Required for SPOT
						"spot_iam_fleet_role": {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validateArn,
						},
						// TODO Can be updated for FARGATE
						"subnets": {
							Type:     schema.TypeSet,
							Required: true,
							ForceNew: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"tags": tagsSchemaForceNew(),
						"type": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
							StateFunc: func(val interface{}) string {
								return strings.ToUpper(val.(string))
							},
							ValidateFunc: validation.StringInSlice(batch.CRType_Values(), true),
						},
					},
				},
			},
			"service_role": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateArn,
			},
			"state": {
				Type:     schema.TypeString,
				Optional: true,
				StateFunc: func(val interface{}) string {
					return strings.ToUpper(val.(string))
				},
				ValidateFunc: validation.StringInSlice(batch.CEState_Values(), true),
				Default:      batch.CEStateEnabled,
			},
			"tags":     tagsSchema(),
			"tags_all": tagsSchemaComputed(),
			"type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				StateFunc: func(val interface{}) string {
					return strings.ToUpper(val.(string))
				},
				ValidateFunc: validation.StringInSlice(batch.CEType_Values(), true),
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ecs_cluster_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status_reason": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},

		CustomizeDiff: SetTagsDiff,
	}
}

func resourceAwsBatchComputeEnvironmentCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).batchconn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(keyvaluetags.New(d.Get("tags").(map[string]interface{})))

	computeEnvironmentName := naming.Generate(d.Get("compute_environment_name").(string), d.Get("compute_environment_name_prefix").(string))

	serviceRole := d.Get("service_role").(string)
	computeEnvironmentType := d.Get("type").(string)

	input := &batch.CreateComputeEnvironmentInput{
		ComputeEnvironmentName: aws.String(computeEnvironmentName),
		ServiceRole:            aws.String(serviceRole),
		Type:                   aws.String(computeEnvironmentType),
	}

	if v, ok := d.GetOk("state"); ok {
		input.State = aws.String(v.(string))
	}

	if len(tags) > 0 {
		input.Tags = tags.IgnoreAws().BatchTags()
	}

	if strings.ToUpper(computeEnvironmentType) == batch.CETypeManaged {
		computeResources := d.Get("compute_resources").([]interface{})
		if len(computeResources) == 0 {
			return fmt.Errorf("One compute environment is expected, but no compute environments are set")
		}
		computeResource := computeResources[0].(map[string]interface{})

		maxvCpus := int64(computeResource["max_vcpus"].(int))
		computeResourceType := computeResource["type"].(string)

		var instanceTypes []*string
		for _, v := range computeResource["instance_type"].(*schema.Set).List() {
			instanceTypes = append(instanceTypes, aws.String(v.(string)))
		}

		var securityGroupIds []*string
		for _, v := range computeResource["security_group_ids"].(*schema.Set).List() {
			securityGroupIds = append(securityGroupIds, aws.String(v.(string)))
		}

		var subnets []*string
		for _, v := range computeResource["subnets"].(*schema.Set).List() {
			subnets = append(subnets, aws.String(v.(string)))
		}

		input.ComputeResources = &batch.ComputeResource{
			InstanceTypes:    instanceTypes,
			MaxvCpus:         aws.Int64(maxvCpus),
			SecurityGroupIds: securityGroupIds,
			Subnets:          subnets,
			Type:             aws.String(computeResourceType),
		}

		if v, ok := computeResource["allocation_strategy"]; ok && len(v.(string)) > 0 {
			input.ComputeResources.AllocationStrategy = aws.String(v.(string))
		}
		if v, ok := computeResource["bid_percentage"]; ok && v.(int) > 0 {
			input.ComputeResources.BidPercentage = aws.Int64(int64(v.(int)))
		}
		if v, ok := computeResource["desired_vcpus"]; ok && v.(int) > 0 {
			input.ComputeResources.DesiredvCpus = aws.Int64(int64(v.(int)))
		}
		if v, ok := computeResource["ec2_key_pair"]; ok && len(v.(string)) > 0 {
			input.ComputeResources.Ec2KeyPair = aws.String(v.(string))
		}
		if v, ok := computeResource["image_id"]; ok && len(v.(string)) > 0 {
			input.ComputeResources.ImageId = aws.String(v.(string))
		}
		if v, ok := computeResource["instance_role"]; ok && len(v.(string)) > 0 {
			input.ComputeResources.InstanceRole = aws.String(v.(string))
		}
		if v, ok := computeResource["min_vcpus"]; ok && v.(int) > 0 {
			input.ComputeResources.MinvCpus = aws.Int64(int64(v.(int)))
		} else if computeResourceType == batch.CRTypeEc2 || computeResourceType == batch.CRTypeSpot {
			input.ComputeResources.MinvCpus = aws.Int64(0)
		}
		if v, ok := computeResource["spot_iam_fleet_role"]; ok && len(v.(string)) > 0 {
			input.ComputeResources.SpotIamFleetRole = aws.String(v.(string))
		}
		if v, ok := computeResource["tags"]; ok {
			input.ComputeResources.Tags = keyvaluetags.New(v.(map[string]interface{})).IgnoreAws().BatchTags()
		}

		if raw, ok := computeResource["launch_template"]; ok && len(raw.([]interface{})) > 0 {
			input.ComputeResources.LaunchTemplate = &batch.LaunchTemplateSpecification{}
			launchTemplate := raw.([]interface{})[0].(map[string]interface{})
			if v, ok := launchTemplate["launch_template_id"]; ok {
				input.ComputeResources.LaunchTemplate.LaunchTemplateId = aws.String(v.(string))
			}
			if v, ok := launchTemplate["launch_template_name"]; ok {
				input.ComputeResources.LaunchTemplate.LaunchTemplateName = aws.String(v.(string))
			}
			if v, ok := launchTemplate["version"]; ok {
				input.ComputeResources.LaunchTemplate.Version = aws.String(v.(string))
			}
		}
	}

	log.Printf("[DEBUG] Create compute environment %s.\n", input)

	if _, err := conn.CreateComputeEnvironment(input); err != nil {
		return err
	}

	d.SetId(computeEnvironmentName)

	if _, err := waiter.ComputeEnvironmentCreated(conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return fmt.Errorf("error waiting for Batch Compute Environment (%s) create: %w", d.Id(), err)
	}

	return resourceAwsBatchComputeEnvironmentRead(d, meta)
}

func resourceAwsBatchComputeEnvironmentRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).batchconn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	computeEnvironment, err := finder.ComputeEnvironmentDetailByName(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Batch Compute Environment (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Batch Compute Environment (%s): %w", d.Id(), err)
	}

	if aws.StringValue(computeEnvironment.Type) == batch.CETypeManaged {
		if err := d.Set("compute_resources", flattenBatchComputeResources(computeEnvironment.ComputeResources)); err != nil {
			return fmt.Errorf("error setting compute_resources: %w", err)
		}
	}

	d.Set("arn", computeEnvironment.ComputeEnvironmentArn)
	d.Set("compute_environment_name", computeEnvironment.ComputeEnvironmentName)
	d.Set("compute_environment_name_prefix", naming.NamePrefixFromName(aws.StringValue(computeEnvironment.ComputeEnvironmentName)))
	d.Set("ecs_cluster_arn", computeEnvironment.EcsClusterArn)
	d.Set("service_role", computeEnvironment.ServiceRole)
	d.Set("state", computeEnvironment.State)
	d.Set("status", computeEnvironment.Status)
	d.Set("status_reason", computeEnvironment.StatusReason)
	d.Set("type", computeEnvironment.Type)

	tags := keyvaluetags.BatchKeyValueTags(computeEnvironment.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceAwsBatchComputeEnvironmentUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).batchconn

	if d.HasChanges("compute_resources", "service_role", "state") {
		computeEnvironmentName := d.Get("compute_environment_name").(string)

		input := &batch.UpdateComputeEnvironmentInput{
			ComputeEnvironment: aws.String(computeEnvironmentName),
		}

		if d.HasChange("service_role") {
			input.ServiceRole = aws.String(d.Get("service_role").(string))
		}
		if d.HasChange("state") {
			input.State = aws.String(d.Get("state").(string))
		}

		if d.HasChange("compute_resources") {
			computeResources := d.Get("compute_resources").([]interface{})
			if len(computeResources) == 0 {
				return fmt.Errorf("One compute environment is expected, but no compute environments are set")
			}
			computeResource := computeResources[0].(map[string]interface{})

			input.ComputeResources = &batch.ComputeResourceUpdate{}

			if d.HasChange("compute_resources.0.desired_vcpus") {
				input.ComputeResources.DesiredvCpus = aws.Int64(int64(computeResource["desired_vcpus"].(int)))
			}

			input.ComputeResources.MaxvCpus = aws.Int64(int64(computeResource["max_vcpus"].(int)))
			computeResourceType := computeResource["type"].(string)
			if computeResourceType == batch.CRTypeEc2 || computeResourceType == batch.CRTypeSpot {
				input.ComputeResources.MinvCpus = aws.Int64(int64(computeResource["min_vcpus"].(int)))
			}
		}

		log.Printf("[DEBUG] Update compute environment %s.\n", input)

		if _, err := conn.UpdateComputeEnvironment(input); err != nil {
			return fmt.Errorf("error updating Batch Compute Environment (%s): %w", d.Id(), err)
		}

		if _, err := waiter.ComputeEnvironmentUpdated(conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return fmt.Errorf("error waiting for Batch Compute Environment (%s) update: %w", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := keyvaluetags.BatchUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating tags: %w", err)
		}
	}

	return resourceAwsBatchComputeEnvironmentRead(d, meta)
}

func resourceAwsBatchComputeEnvironmentDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).batchconn

	log.Printf("[DEBUG] Disabling Batch Compute Environment (%s)", d.Id())
	{
		input := &batch.UpdateComputeEnvironmentInput{
			ComputeEnvironment: aws.String(d.Id()),
			State:              aws.String(batch.CEStateDisabled),
		}

		if _, err := conn.UpdateComputeEnvironment(input); err != nil {
			return fmt.Errorf("error disabling Batch Compute Environment (%s): %w", d.Id(), err)
		}

		if _, err := waiter.ComputeEnvironmentDisabled(conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
			return fmt.Errorf("error waiting for Batch Compute Environment (%s) disable: %w", d.Id(), err)
		}
	}

	log.Printf("[DEBUG] Deleting Batch Compute Environment (%s)", d.Id())
	{
		input := &batch.DeleteComputeEnvironmentInput{
			ComputeEnvironment: aws.String(d.Id()),
		}

		if _, err := conn.DeleteComputeEnvironment(input); err != nil {
			return fmt.Errorf("error deleting Batch Compute Environment (%s): %w", d.Id(), err)
		}

		if _, err := waiter.ComputeEnvironmentDeleted(conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
			return fmt.Errorf("error waiting for Batch Compute Environment (%s) delete: %w", d.Id(), err)
		}
	}

	return nil
}

func flattenBatchComputeResources(computeResource *batch.ComputeResource) []map[string]interface{} {
	result := make([]map[string]interface{}, 0)
	m := make(map[string]interface{})

	m["allocation_strategy"] = aws.StringValue(computeResource.AllocationStrategy)
	m["bid_percentage"] = int(aws.Int64Value(computeResource.BidPercentage))
	m["desired_vcpus"] = int(aws.Int64Value(computeResource.DesiredvCpus))
	m["ec2_key_pair"] = aws.StringValue(computeResource.Ec2KeyPair)
	m["image_id"] = aws.StringValue(computeResource.ImageId)
	m["instance_role"] = aws.StringValue(computeResource.InstanceRole)
	m["instance_type"] = flattenStringSet(computeResource.InstanceTypes)
	m["max_vcpus"] = int(aws.Int64Value(computeResource.MaxvCpus))
	m["min_vcpus"] = int(aws.Int64Value(computeResource.MinvCpus))
	m["security_group_ids"] = flattenStringSet(computeResource.SecurityGroupIds)
	m["spot_iam_fleet_role"] = aws.StringValue(computeResource.SpotIamFleetRole)
	m["subnets"] = flattenStringSet(computeResource.Subnets)
	m["tags"] = keyvaluetags.BatchKeyValueTags(computeResource.Tags).IgnoreAws().Map()
	m["type"] = aws.StringValue(computeResource.Type)

	if launchTemplate := computeResource.LaunchTemplate; launchTemplate != nil {
		lt := make(map[string]interface{})
		lt["launch_template_id"] = aws.StringValue(launchTemplate.LaunchTemplateId)
		lt["launch_template_name"] = aws.StringValue(launchTemplate.LaunchTemplateName)
		lt["version"] = aws.StringValue(launchTemplate.Version)
		m["launch_template"] = []map[string]interface{}{lt}
	}

	result = append(result, m)
	return result
}
