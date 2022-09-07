package ecs

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
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

func ResourceCapacityProvider() *schema.Resource {
	return &schema.Resource{
		Create: resourceCapacityProviderCreate,
		Read:   resourceCapacityProviderRead,
		Update: resourceCapacityProviderUpdate,
		Delete: resourceCapacityProviderDelete,
		Importer: &schema.ResourceImporter{
			State: resourceCapacityProviderImport,
		},

		CustomizeDiff: verify.SetTagsDiff,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"auto_scaling_group_provider": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Required: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"auto_scaling_group_arn": {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: verify.ValidARN,
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
										ValidateFunc: validation.IntBetween(1, 10000),
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
									"status": {
										Type:         schema.TypeString,
										Optional:     true,
										Computed:     true,
										ValidateFunc: validation.StringInSlice(ecs.ManagedScalingStatus_Values(), false)},
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
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.StringInSlice(ecs.ManagedTerminationProtection_Values(), false),
						},
					},
				},
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},
	}
}

func resourceCapacityProviderCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ECSConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	name := d.Get("name").(string)
	input := ecs.CreateCapacityProviderInput{
		Name:                     aws.String(name),
		AutoScalingGroupProvider: expandAutoScalingGroupProviderCreate(d.Get("auto_scaling_group_provider")),
	}

	// `CreateCapacityProviderInput` does not accept an empty array of tags
	if len(tags) > 0 {
		input.Tags = Tags(tags.IgnoreAWS())
	}

	log.Printf("[DEBUG] Creating ECS Capacity Provider: %s", input)
	output, err := conn.CreateCapacityProvider(&input)

	// Some partitions (i.e., ISO) may not support tag-on-create
	if input.Tags != nil && verify.CheckISOErrorTagsUnsupported(conn.PartitionID, err) {
		log.Printf("[WARN] ECS tagging failed creating Capacity Provider (%s) with tags: %s. Trying create without tags.", name, err)
		input.Tags = nil

		output, err = conn.CreateCapacityProvider(&input)
	}

	if err != nil {
		return fmt.Errorf("failed creating ECS Capacity Provider (%s): %w", name, err)
	}

	d.SetId(aws.StringValue(output.CapacityProvider.CapacityProviderArn))

	// Some partitions (i.e., ISO) may not support tag-on-create, attempt tag after create
	if input.Tags == nil && len(tags) > 0 {
		err := UpdateTags(conn, d.Id(), nil, tags)

		if v, ok := d.GetOk("tags"); (!ok || len(v.(map[string]interface{})) == 0) && verify.CheckISOErrorTagsUnsupported(conn.PartitionID, err) {
			// If default tags only, log and continue. Otherwise, error.
			log.Printf("[WARN] ECS tagging failed adding tags after create for Capacity Provider (%s): %s", d.Id(), err)
			return resourceCapacityProviderRead(d, meta)
		}

		if err != nil {
			return fmt.Errorf("ECS tagging failed adding tags after create for Capacity Provider (%s): %w", d.Id(), err)
		}
	}

	return resourceCapacityProviderRead(d, meta)
}

func resourceCapacityProviderRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ECSConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	output, err := FindCapacityProviderByARN(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] ECS Capacity Provider (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading ECS Capacity Provider (%s): %w", d.Id(), err)
	}

	d.Set("arn", output.CapacityProviderArn)

	if err := d.Set("auto_scaling_group_provider", flattenAutoScalingGroupProvider(output.AutoScalingGroupProvider)); err != nil {
		return fmt.Errorf("error setting auto_scaling_group_provider: %w", err)
	}

	d.Set("name", output.Name)

	tags := KeyValueTags(output.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceCapacityProviderUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ECSConn

	if d.HasChangesExcept("tags", "tags_all") {
		input := &ecs.UpdateCapacityProviderInput{
			AutoScalingGroupProvider: expandAutoScalingGroupProviderUpdate(d.Get("auto_scaling_group_provider")),
			Name:                     aws.String(d.Get("name").(string)),
		}

		log.Printf("[DEBUG] Updating ECS Capacity Provider: %s", input)
		err := resource.Retry(capacityProviderUpdateTimeout, func() *resource.RetryError {
			_, err := conn.UpdateCapacityProvider(input)

			if tfawserr.ErrCodeEquals(err, ecs.ErrCodeUpdateInProgressException) {
				return resource.RetryableError(err)
			}

			if err != nil {
				return resource.NonRetryableError(err)
			}

			return nil
		})

		if tfresource.TimedOut(err) {
			_, err = conn.UpdateCapacityProvider(input)
		}

		if err != nil {
			return fmt.Errorf("error updating ECS Capacity Provider (%s): %w", d.Id(), err)
		}

		if _, err = waitCapacityProviderUpdated(conn, d.Id()); err != nil {
			return fmt.Errorf("error waiting for ECS Capacity Provider (%s) to update: %w", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		err := UpdateTags(conn, d.Id(), o, n)

		// Some partitions (i.e., ISO) may not support tagging, giving error
		if verify.CheckISOErrorTagsUnsupported(conn.PartitionID, err) {
			log.Printf("[WARN] ECS tagging failed updating tags for Capacity Provider (%s): %s", d.Id(), err)
			return resourceCapacityProviderRead(d, meta)
		}

		if err != nil {
			return fmt.Errorf("ECS tagging failed updating tags for Capacity Provider (%s): %w", d.Id(), err)
		}
	}

	return resourceCapacityProviderRead(d, meta)
}

func resourceCapacityProviderDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ECSConn

	log.Printf("[DEBUG] Deleting ECS Capacity Provider (%s)", d.Id())
	_, err := conn.DeleteCapacityProvider(&ecs.DeleteCapacityProviderInput{
		CapacityProvider: aws.String(d.Id()),
	})

	// "An error occurred (ClientException) when calling the DeleteCapacityProvider operation: The specified capacity provider does not exist. Specify a valid name or ARN and try again."
	if tfawserr.ErrMessageContains(err, ecs.ErrCodeClientException, "capacity provider does not exist") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting ECS Capacity Provider (%s): %w", d.Id(), err)
	}

	if _, err := waitCapacityProviderDeleted(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for ECS Capacity Provider (%s) to delete: %w", d.Id(), err)
	}

	return nil
}

func resourceCapacityProviderImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	d.Set("name", d.Id())
	d.SetId(arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Service:   "ecs",
		Resource:  fmt.Sprintf("capacity-provider/%s", d.Id()),
	}.String())
	return []*schema.ResourceData{d}, nil
}

func expandAutoScalingGroupProviderCreate(configured interface{}) *ecs.AutoScalingGroupProvider {
	if configured == nil {
		return nil
	}

	if configured.([]interface{}) == nil || len(configured.([]interface{})) == 0 {
		return nil
	}

	prov := ecs.AutoScalingGroupProvider{}
	p := configured.([]interface{})[0].(map[string]interface{})
	arn := p["auto_scaling_group_arn"].(string)
	prov.AutoScalingGroupArn = aws.String(arn)

	if mtp := p["managed_termination_protection"].(string); len(mtp) > 0 {
		prov.ManagedTerminationProtection = aws.String(mtp)
	}

	prov.ManagedScaling = expandManagedScaling(p["managed_scaling"])

	return &prov
}

func expandAutoScalingGroupProviderUpdate(configured interface{}) *ecs.AutoScalingGroupProviderUpdate {
	if configured == nil {
		return nil
	}

	if configured.([]interface{}) == nil || len(configured.([]interface{})) == 0 {
		return nil
	}

	prov := ecs.AutoScalingGroupProviderUpdate{}
	p := configured.([]interface{})[0].(map[string]interface{})

	if mtp := p["managed_termination_protection"].(string); len(mtp) > 0 {
		prov.ManagedTerminationProtection = aws.String(mtp)
	}

	prov.ManagedScaling = expandManagedScaling(p["managed_scaling"])

	return &prov
}

func expandManagedScaling(configured interface{}) *ecs.ManagedScaling {
	if configured == nil {
		return nil
	}

	if configured.([]interface{}) == nil || len(configured.([]interface{})) == 0 {
		return nil
	}

	p := configured.([]interface{})[0].(map[string]interface{})

	managedScaling := ecs.ManagedScaling{}

	if val, ok := p["instance_warmup_period"].(int); ok && val != 0 {
		managedScaling.InstanceWarmupPeriod = aws.Int64(int64(val))
	}
	if val, ok := p["maximum_scaling_step_size"].(int); ok && val != 0 {
		managedScaling.MaximumScalingStepSize = aws.Int64(int64(val))
	}
	if val, ok := p["minimum_scaling_step_size"].(int); ok && val != 0 {
		managedScaling.MinimumScalingStepSize = aws.Int64(int64(val))
	}
	if val, ok := p["status"].(string); ok && len(val) > 0 {
		managedScaling.Status = aws.String(val)
	}
	if val, ok := p["target_capacity"].(int); ok && val != 0 {
		managedScaling.TargetCapacity = aws.Int64(int64(val))
	}

	return &managedScaling
}

func flattenAutoScalingGroupProvider(provider *ecs.AutoScalingGroupProvider) []map[string]interface{} {
	if provider == nil {
		return nil
	}

	p := map[string]interface{}{
		"auto_scaling_group_arn":         aws.StringValue(provider.AutoScalingGroupArn),
		"managed_termination_protection": aws.StringValue(provider.ManagedTerminationProtection),
		"managed_scaling":                []map[string]interface{}{},
	}

	if provider.ManagedScaling != nil {
		m := map[string]interface{}{
			"instance_warmup_period":    aws.Int64Value(provider.ManagedScaling.InstanceWarmupPeriod),
			"maximum_scaling_step_size": aws.Int64Value(provider.ManagedScaling.MaximumScalingStepSize),
			"minimum_scaling_step_size": aws.Int64Value(provider.ManagedScaling.MinimumScalingStepSize),
			"status":                    aws.StringValue(provider.ManagedScaling.Status),
			"target_capacity":           aws.Int64Value(provider.ManagedScaling.TargetCapacity),
		}

		p["managed_scaling"] = []map[string]interface{}{m}
	}

	result := []map[string]interface{}{p}
	return result
}
