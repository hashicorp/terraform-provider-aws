package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/batch"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
)

const (
	MANAGED   = "MANAGED"
	UNMANAGED = "UNMANAGED"
	EC2       = "EC2"
	SPOT      = "SPOT"
	ENABLED   = "ENABLED"
	DISABLED  = "DISABLED"
)

func resourceAwsBatchComputeEnvironment() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsBatchComputeEnvironmentCreate,
		Read:   resourceAwsBatchComputeEnvironmentRead,
		Update: resourceAwsBatchComputeEnvironmentUpdate,
		Delete: resourceAwsBatchComputeEnvironmentDelete,

		Schema: map[string]*schema.Schema{
			"compute_environment_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"compute_resources": {
				Type:     schema.TypeList,
				Optional: true,
				MinItems: 0,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"bid_percentage": {
							Type:     schema.TypeInt,
							Optional: true,
							ForceNew: true,
						},
						"desired_vcpus": {
							Type:     schema.TypeInt,
							Optional: true,
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
						"instance_role": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"instance_type": {
							Type:     schema.TypeSet,
							Required: true,
							ForceNew: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Set:      schema.HashString,
						},
						"max_vcpus": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"min_vcpus": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"security_group_ids": {
							Type:     schema.TypeSet,
							Required: true,
							ForceNew: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Set:      schema.HashString,
						},
						"spot_iam_fleet_role": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"subnets": {
							Type:     schema.TypeSet,
							Required: true,
							ForceNew: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Set:      schema.HashString,
						},
						"tags": tagsSchema(),
						"type": {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringInSlice([]string{EC2, SPOT}, true),
						},
					},
				},
			},
			"service_role": {
				Type:     schema.TypeString,
				Required: true,
			},
			"state": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{ENABLED, DISABLED}, true),
				Default:      "ENABLED",
			},
			"type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{MANAGED, UNMANAGED}, true),
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ecc_cluster_arn": {
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
	}
}

func resourceAwsBatchComputeEnvironmentCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).batchconn

	computeEnvironmentName := d.Get("compute_environment_name").(string)

	log.Printf("[DEBUG] Create compute environment \"%s\".\n", computeEnvironmentName)

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

	if computeEnvironmentType == MANAGED {
		computeResources := d.Get("compute_resources").([]interface{})
		if len(computeResources) == 0 {
			return fmt.Errorf("One compute environment is expected, but no compute environments are set")
		}
		computeResource := computeResources[0].(map[string]interface{})

		instanceRole := computeResource["instance_role"].(string)
		maxvCpus := int64(computeResource["max_vcpus"].(int))
		minvCpus := int64(computeResource["min_vcpus"].(int))
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
			InstanceRole:     aws.String(instanceRole),
			InstanceTypes:    instanceTypes,
			MaxvCpus:         aws.Int64(maxvCpus),
			MinvCpus:         aws.Int64(minvCpus),
			SecurityGroupIds: securityGroupIds,
			Subnets:          subnets,
			Type:             aws.String(computeResourceType),
		}

		if v, ok := computeResource["bid_percentage"]; ok {
			input.ComputeResources.BidPercentage = aws.Int64(int64(v.(int)))
		}
		if v, ok := computeResource["desired_vcpus"]; ok {
			input.ComputeResources.DesiredvCpus = aws.Int64(int64(v.(int)))
		}
		if v, ok := computeResource["ec2_key_pair"]; ok {
			input.ComputeResources.Ec2KeyPair = aws.String(v.(string))
		}
		if v, ok := computeResource["image_id"]; ok {
			input.ComputeResources.ImageId = aws.String(v.(string))
		}
		if v, ok := computeResource["spot_iam_fleet_role"]; ok {
			input.ComputeResources.SpotIamFleetRole = aws.String(v.(string))
		}
		if v, ok := computeResource["tags"]; ok {
			input.ComputeResources.Tags = tagsFromMapGeneric(v.(map[string]interface{}))
		}
	}

	if _, err := conn.CreateComputeEnvironment(input); err != nil {
		return err
	}

	d.SetId(computeEnvironmentName)

	if err := waitComputeEnvironmentValid(d, meta); err != nil {
		return err
	}

	return resourceAwsBatchComputeEnvironmentRead(d, meta)
}

func resourceAwsBatchComputeEnvironmentRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).batchconn

	computeEnvironmentName := d.Get("compute_environment_name").(string)

	input := &batch.DescribeComputeEnvironmentsInput{
		ComputeEnvironments: []*string{
			aws.String(computeEnvironmentName),
		},
	}

	result, err := conn.DescribeComputeEnvironments(input)
	if err != nil {
		return err
	}

	if len(result.ComputeEnvironments) == 0 {
		return fmt.Errorf("One compute environment is expected, but AWS return no compute environment")
	} else if len(result.ComputeEnvironments) >= 2 {
		return fmt.Errorf("One compute environment is expected, but AWS return too many compute environment")
	}
	computeEnvironment := result.ComputeEnvironments[0]

	d.Set("service_role", computeEnvironment.ServiceRole)
	d.Set("state", computeEnvironment.State)
	d.Set("type", computeEnvironment.Type)

	if *(computeEnvironment.Type) == "MANAGED" {
		computeResource := computeEnvironment.ComputeResources

		d.Set("compute_resources", []interface{}{
			map[string]interface{}{
				"bid_percentage":      computeResource.BidPercentage,
				"desired_vcpus":       computeResource.DesiredvCpus,
				"ec2_key_pair":        computeResource.Ec2KeyPair,
				"image_id":            computeResource.ImageId,
				"instance_role":       computeResource.InstanceRole,
				"instance_type":       schema.NewSet(schema.HashString, flattenStringList(computeResource.InstanceTypes)),
				"max_vcpus":           computeResource.MaxvCpus,
				"min_vcpus":           computeResource.MinvCpus,
				"security_group_ids":  schema.NewSet(schema.HashString, flattenStringList(computeResource.SecurityGroupIds)),
				"spot_iam_fleet_role": computeResource.SpotIamFleetRole,
				"subnets":             schema.NewSet(schema.HashString, flattenStringList(computeResource.Subnets)),
				"tags":                tagsToMapGeneric(computeResource.Tags),
				"type":                computeResource.Type,
			},
		})
	}

	d.Set("arn", computeEnvironment.ComputeEnvironmentArn)
	d.Set("ecc_cluster_arn", computeEnvironment.ComputeEnvironmentArn)
	d.Set("status", computeEnvironment.Status)
	d.Set("status_reason", computeEnvironment.StatusReason)

	return nil
}

func resourceAwsBatchComputeEnvironmentDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).batchconn

	computeEnvironmentName := d.Get("compute_environment_name").(string)

	log.Printf("[DEBUG] Delete compute environment \"%s\".\n", computeEnvironmentName)

	updateInput := &batch.UpdateComputeEnvironmentInput{
		ComputeEnvironment: aws.String(computeEnvironmentName),
		State:              aws.String(DISABLED),
	}

	if _, err := conn.UpdateComputeEnvironment(updateInput); err != nil {
		return err
	}

	if err := waitComputeEnvironmentValid(d, meta); err != nil {
		return err
	}

	input := &batch.DeleteComputeEnvironmentInput{
		ComputeEnvironment: aws.String(computeEnvironmentName),
	}

	if _, err := conn.DeleteComputeEnvironment(input); err != nil {
		return err
	}

	if err := waitComputeEnvironmentDeleted(d, meta); err != nil {
		return err
	}

	d.SetId("")

	return nil
}

func resourceAwsBatchComputeEnvironmentUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).batchconn

	computeEnvironmentName := d.Get("compute_environment_name").(string)

	log.Printf("[DEBUG] Update compute environment \"%s\".\n", computeEnvironmentName)

	input := &batch.UpdateComputeEnvironmentInput{
		ComputeEnvironment: aws.String(computeEnvironmentName),
		ComputeResources:   &batch.ComputeResourceUpdate{},
	}

	if d.HasChange("service_role") {
		log.Printf("[DEBUG] Service role of \"%s\" has change.\n", computeEnvironmentName)
		input.ServiceRole = aws.String(d.Get("service_role").(string))
	}
	if d.HasChange("state") {
		log.Printf("[DEBUG] State of \"%s\" has change.\n", computeEnvironmentName)
		input.ServiceRole = aws.String(d.Get("state").(string))
	}

	if d.HasChange("compute_resources") {
		log.Printf("[DEBUG] Compute resources of \"%s\" has change.\n", computeEnvironmentName)
		computeResources := d.Get("compute_resources").([]interface{})
		if len(computeResources) == 0 {
			return fmt.Errorf("One compute environment is expected, but no compute environments are set")
		}
		computeResource := computeResources[0].(map[string]interface{})

		if d.HasChange("compute_resources.0.desired_vcpus") {
			log.Printf("[DEBUG] Desired vCpus of \"%s\" has change.\n", computeEnvironmentName)
			input.ComputeResources.DesiredvCpus = aws.Int64(int64(computeResource["desired_vcpus"].(int)))
		}
		if d.HasChange("compute_resources.0.max_vcpus") {
			log.Printf("[DEBUG] Max vCpus of \"%s\" has change.\n", computeEnvironmentName)
			input.ComputeResources.MaxvCpus = aws.Int64(int64(computeResource["max_vcpus"].(int)))
		}
		if d.HasChange("compute_resources.0.min_vcpus") {
			log.Printf("[DEBUG] Min vCpus of \"%s\" has change.\n", computeEnvironmentName)
			input.ComputeResources.MinvCpus = aws.Int64(int64(computeResource["min_vcpus"].(int)))
		}
	}

	if _, err := conn.UpdateComputeEnvironment(input); err != nil {
		return err
	}

	return resourceAwsBatchComputeEnvironmentRead(d, meta)
}

func waitComputeEnvironmentValid(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).batchconn

	computeEnvironmentName := d.Get("compute_environment_name").(string)

	for {
		result, err := conn.DescribeComputeEnvironments(&batch.DescribeComputeEnvironmentsInput{
			ComputeEnvironments: []*string{
				aws.String(computeEnvironmentName),
			},
		})
		if err != nil {
			return err
		}

		if len(result.ComputeEnvironments) == 0 {
			return fmt.Errorf("One compute environment is expected, but AWS return no compute environment")
		} else if len(result.ComputeEnvironments) >= 2 {
			return fmt.Errorf("One compute environment is expected, but AWS return too many compute environment")
		}
		computeEnvironment := result.ComputeEnvironments[0]
		if *(computeEnvironment.Status) == "VALID" {
			break
		} else {
			time.Sleep(1 * time.Second)
		}
	}

	log.Printf("[DEBUG] Status of compute environment \"%s\" get VALID.\n", computeEnvironmentName)
	return nil
}

func waitComputeEnvironmentDeleted(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).batchconn

	computeEnvironmentName := d.Get("compute_environment_name").(string)

	for {
		result, err := conn.DescribeComputeEnvironments(&batch.DescribeComputeEnvironmentsInput{
			ComputeEnvironments: []*string{
				aws.String(computeEnvironmentName),
			},
		})
		if err != nil {
			return err
		}

		if len(result.ComputeEnvironments) == 0 {
			break
		} else {
			time.Sleep(1 * time.Second)
		}
	}

	log.Printf("[DEBUG] Compute environment \"%s\" deleted.\n", computeEnvironmentName)
	return nil
}
