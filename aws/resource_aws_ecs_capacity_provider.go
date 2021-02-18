package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/ecs/waiter"
)

const (
	ecsCapacityProviderTimeoutUpdate = 10 * time.Minute
)

func resourceAwsEcsCapacityProvider() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsEcsCapacityProviderCreate,
		Read:   resourceAwsEcsCapacityProviderRead,
		Update: resourceAwsEcsCapacityProviderUpdate,
		Delete: resourceAwsEcsCapacityProviderDelete,
		Importer: &schema.ResourceImporter{
			State: resourceAwsEcsCapacityProviderImport,
		},

		CustomizeDiff: SetTagsDiff,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
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
							ValidateFunc: validateArn,
							ForceNew:     true,
						},
						"managed_termination_protection": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
							ValidateFunc: validation.StringInSlice([]string{
								ecs.ManagedTerminationProtectionEnabled,
								ecs.ManagedTerminationProtectionDisabled,
							}, false),
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
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
										ValidateFunc: validation.StringInSlice([]string{
											ecs.ManagedScalingStatusEnabled,
											ecs.ManagedScalingStatusDisabled,
										}, false)},
									"target_capacity": {
										Type:         schema.TypeInt,
										Optional:     true,
										Computed:     true,
										ValidateFunc: validation.IntBetween(1, 100),
									},
								},
							},
						},
					},
				},
			},
			"tags":     tagsSchema(),
			"tags_all": tagsSchemaComputed(),
		},
	}
}

func resourceAwsEcsCapacityProviderCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ecsconn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(keyvaluetags.New(d.Get("tags").(map[string]interface{})))

	input := ecs.CreateCapacityProviderInput{
		Name:                     aws.String(d.Get("name").(string)),
		AutoScalingGroupProvider: expandAutoScalingGroupProviderCreate(d.Get("auto_scaling_group_provider")),
	}

	// `CreateCapacityProviderInput` does not accept an empty array of tags
	if len(tags) > 0 {
		input.Tags = tags.IgnoreAws().EcsTags()
	}

	out, err := conn.CreateCapacityProvider(&input)

	if err != nil {
		return fmt.Errorf("error creating ECS Capacity Provider: %s", err)
	}

	provider := *out.CapacityProvider

	log.Printf("[DEBUG] ECS Capacity Provider created: %s", aws.StringValue(provider.CapacityProviderArn))
	d.SetId(aws.StringValue(provider.CapacityProviderArn))

	return resourceAwsEcsCapacityProviderRead(d, meta)
}

func resourceAwsEcsCapacityProviderRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ecsconn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	input := &ecs.DescribeCapacityProvidersInput{
		CapacityProviders: []*string{aws.String(d.Id())},
		Include:           []*string{aws.String(ecs.CapacityProviderFieldTags)},
	}

	output, err := conn.DescribeCapacityProviders(input)

	if err != nil {
		return fmt.Errorf("error reading ECS Capacity Provider (%s): %s", d.Id(), err)
	}

	var provider *ecs.CapacityProvider
	for _, cp := range output.CapacityProviders {
		if aws.StringValue(cp.CapacityProviderArn) == d.Id() {
			provider = cp
			break
		}
	}

	if provider == nil {
		log.Printf("[WARN] ECS Capacity Provider (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if aws.StringValue(provider.Status) == ecs.CapacityProviderStatusInactive {
		log.Printf("[WARN] ECS Capacity Provider (%s) is INACTIVE, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("arn", provider.CapacityProviderArn)
	d.Set("name", provider.Name)

	tags := keyvaluetags.EcsKeyValueTags(provider.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	if err := d.Set("auto_scaling_group_provider", flattenAutoScalingGroupProvider(provider.AutoScalingGroupProvider)); err != nil {
		return fmt.Errorf("error setting autoscaling group provider: %s", err)
	}

	return nil
}

func resourceAwsEcsCapacityProviderUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ecsconn

	input := &ecs.UpdateCapacityProviderInput{
		Name: aws.String(d.Get("name").(string)),
	}

	if d.HasChange("auto_scaling_group_provider") {
		input.AutoScalingGroupProvider = expandAutoScalingGroupProviderUpdate(d.Get("auto_scaling_group_provider"))

		err := resource.Retry(ecsCapacityProviderTimeoutUpdate, func() *resource.RetryError {
			_, err := conn.UpdateCapacityProvider(input)
			if err != nil {
				if isAWSErr(err, ecs.ErrCodeUpdateInProgressException, "") {
					return resource.RetryableError(err)
				}
				return resource.NonRetryableError(err)
			}
			return nil
		})
		if isResourceTimeoutError(err) {
			_, err = conn.UpdateCapacityProvider(input)
		}
		if err != nil {
			return fmt.Errorf("error updating ECS Capacity Provider (%s): %s", d.Id(), err)
		}

		if _, err = waiter.CapacityProviderUpdate(conn, d.Id()); err != nil {
			return fmt.Errorf("error waiting for ECS Capacity Provider (%s) update: %s", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := keyvaluetags.EcsUpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating ECS Capacity Provider (%s) tags: %s", d.Id(), err)
		}
	}

	return resourceAwsEcsCapacityProviderRead(d, meta)
}

func resourceAwsEcsCapacityProviderDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ecsconn

	input := &ecs.DeleteCapacityProviderInput{
		CapacityProvider: aws.String(d.Id()),
	}

	_, err := conn.DeleteCapacityProvider(input)

	if err != nil {
		return fmt.Errorf("error deleting ECS Capacity Provider (%s): %w", d.Id(), err)
	}

	if _, err := waiter.CapacityProviderInactive(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for ECS Capacity Provider (%s) to delete: %w", d.Id(), err)
	}

	return nil
}

func resourceAwsEcsCapacityProviderImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	d.Set("name", d.Id())
	d.SetId(arn.ARN{
		Partition: meta.(*AWSClient).partition,
		Region:    meta.(*AWSClient).region,
		AccountID: meta.(*AWSClient).accountid,
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
