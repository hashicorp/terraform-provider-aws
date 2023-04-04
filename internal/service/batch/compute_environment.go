package batch

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/batch"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceComputeEnvironment() *schema.Resource {
	return &schema.Resource{
		Create: resourceComputeEnvironmentCreate,
		Read:   resourceComputeEnvironmentRead,
		Update: resourceComputeEnvironmentUpdate,
		Delete: resourceComputeEnvironmentDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		CustomizeDiff: customdiff.Sequence(
			resourceComputeEnvironmentCustomizeDiff,
			verify.SetTagsDiff,
		),

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"compute_environment_name": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"compute_environment_name_prefix"},
				ValidateFunc:  validName,
			},
			"compute_environment_name_prefix": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"compute_environment_name"},
				ValidateFunc:  validPrefix,
			},
			"compute_resources": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				MinItems: 0,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"allocation_strategy": {
							Type:     schema.TypeString,
							Optional: true,
							StateFunc: func(val interface{}) string {
								return strings.ToUpper(val.(string))
							},
							ValidateFunc: validation.StringInSlice(batch.CRAllocationStrategy_Values(), true),
						},
						"bid_percentage": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"desired_vcpus": {
							Type:     schema.TypeInt,
							Optional: true,
							Computed: true,
						},
						"ec2_configuration": {
							Type:     schema.TypeList,
							Optional: true,
							Computed: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"image_id_override": {
										Type:         schema.TypeString,
										Optional:     true,
										Computed:     true,
										ValidateFunc: validation.StringLenBetween(1, 256),
									},
									"image_type": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringLenBetween(1, 256),
									},
								},
							},
						},
						"ec2_key_pair": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"image_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"instance_role": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: verify.ValidARN,
						},
						"instance_type": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"launch_template": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"launch_template_id": {
										Type:          schema.TypeString,
										Optional:      true,
										ConflictsWith: []string{"compute_resources.0.launch_template.0.launch_template_name"},
									},
									"launch_template_name": {
										Type:          schema.TypeString,
										Optional:      true,
										ConflictsWith: []string{"compute_resources.0.launch_template.0.launch_template_id"},
									},
									"version": {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
									},
								},
							},
						},
						"max_vcpus": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"min_vcpus": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"security_group_ids": {
							Type:     schema.TypeSet,
							Required: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"spot_iam_fleet_role": {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: verify.ValidARN,
						},
						"subnets": {
							Type:     schema.TypeSet,
							Required: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"tags": tftags.TagsSchema(),
						"type": {
							Type:     schema.TypeString,
							Required: true,
							StateFunc: func(val interface{}) string {
								return strings.ToUpper(val.(string))
							},
							ValidateFunc: validation.StringInSlice(batch.CRType_Values(), true),
						},
					},
				},
			},
			"ecs_cluster_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"service_role": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: verify.ValidARN,
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
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status_reason": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
			"type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				StateFunc: func(val interface{}) string {
					return strings.ToUpper(val.(string))
				},
				ValidateFunc: validation.StringInSlice(batch.CEType_Values(), true),
			},
		},
	}
}

func resourceComputeEnvironmentCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).BatchConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	computeEnvironmentName := create.Name(d.Get("compute_environment_name").(string), d.Get("compute_environment_name_prefix").(string))
	computeEnvironmentType := d.Get("type").(string)

	input := &batch.CreateComputeEnvironmentInput{
		ComputeEnvironmentName: aws.String(computeEnvironmentName),
		ServiceRole:            aws.String(d.Get("service_role").(string)),
		Type:                   aws.String(computeEnvironmentType),
	}

	if v, ok := d.GetOk("compute_resources"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.ComputeResources = expandComputeResource(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("state"); ok {
		input.State = aws.String(v.(string))
	}

	if len(tags) > 0 {
		input.Tags = Tags(tags.IgnoreAWS())
	}

	log.Printf("[DEBUG] Creating Batch Compute Environment: %s", input)
	output, err := conn.CreateComputeEnvironment(input)

	if err != nil {
		return fmt.Errorf("error creating Batch Compute Environment (%s): %w", computeEnvironmentName, err)
	}

	d.SetId(aws.StringValue(output.ComputeEnvironmentName))

	if _, err := waitComputeEnvironmentCreated(conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return fmt.Errorf("error waiting for Batch Compute Environment (%s) create: %w", d.Id(), err)
	}

	return resourceComputeEnvironmentRead(d, meta)
}

func resourceComputeEnvironmentRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).BatchConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	computeEnvironment, err := FindComputeEnvironmentDetailByName(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Batch Compute Environment (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Batch Compute Environment (%s): %w", d.Id(), err)
	}

	computeEnvironmentType := aws.StringValue(computeEnvironment.Type)

	d.Set("arn", computeEnvironment.ComputeEnvironmentArn)
	d.Set("compute_environment_name", computeEnvironment.ComputeEnvironmentName)
	d.Set("compute_environment_name_prefix", create.NamePrefixFromName(aws.StringValue(computeEnvironment.ComputeEnvironmentName)))
	d.Set("ecs_cluster_arn", computeEnvironment.EcsClusterArn)
	d.Set("service_role", computeEnvironment.ServiceRole)
	d.Set("state", computeEnvironment.State)
	d.Set("status", computeEnvironment.Status)
	d.Set("status_reason", computeEnvironment.StatusReason)
	d.Set("type", computeEnvironmentType)

	if computeEnvironment.ComputeResources != nil {
		if err := d.Set("compute_resources", []interface{}{flattenComputeResource(computeEnvironment.ComputeResources)}); err != nil {
			return fmt.Errorf("error setting compute_resources: %w", err)
		}
	} else {
		d.Set("compute_resources", nil)
	}

	tags := KeyValueTags(computeEnvironment.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceComputeEnvironmentUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).BatchConn

	if d.HasChangesExcept("tags", "tags_all") {
		input := &batch.UpdateComputeEnvironmentInput{
			ComputeEnvironment: aws.String(d.Id()),
		}

		if d.HasChange("service_role") {
			input.ServiceRole = aws.String(d.Get("service_role").(string))
		}

		if d.HasChange("state") {
			input.State = aws.String(d.Get("state").(string))
		}

		if computeEnvironmentType := strings.ToUpper(d.Get("type").(string)); computeEnvironmentType == batch.CETypeManaged {
			// "At least one compute-resources attribute must be specified"
			computeResourceUpdate := &batch.ComputeResourceUpdate{
				MaxvCpus: aws.Int64(int64(d.Get("compute_resources.0.max_vcpus").(int))),
			}

			if d.HasChange("compute_resources.0.security_group_ids") {
				computeResourceUpdate.SecurityGroupIds = flex.ExpandStringSet(d.Get("compute_resources.0.security_group_ids").(*schema.Set))
			}

			if d.HasChange("compute_resources.0.subnets") {
				computeResourceUpdate.Subnets = flex.ExpandStringSet(d.Get("compute_resources.0.subnets").(*schema.Set))
			}

			if d.HasChange("compute_resources.0.allocation_strategy") {
				if allocationStrategy, ok := d.GetOk("compute_resources.0.allocation_strategy"); ok {
					computeResourceUpdate.AllocationStrategy = aws.String(allocationStrategy.(string))
				} else {
					computeResourceUpdate.AllocationStrategy = aws.String("")
				}
			}

			computeEnvironmentType := d.Get("compute_resources.0.type").(string)

			if d.HasChange("compute_resources.0.type") {
				computeResourceUpdate.Type = aws.String(computeEnvironmentType)
			}

			if !isFargateType(computeEnvironmentType) {

				if d.HasChange("compute_resources.0.desired_vcpus") {
					if desiredvCpus, ok := d.GetOk("compute_resources.0.desired_vcpus"); ok {
						computeResourceUpdate.DesiredvCpus = aws.Int64(int64(desiredvCpus.(int)))
					} else {
						computeResourceUpdate.DesiredvCpus = aws.Int64(0)
					}
				}

				if d.HasChange("compute_resources.0.min_vcpus") {
					if minVcpus, ok := d.GetOk("compute_resources.0.min_vcpus"); ok {
						computeResourceUpdate.MinvCpus = aws.Int64(int64(minVcpus.(int)))
					} else {
						computeResourceUpdate.MinvCpus = aws.Int64(0)
					}
				}

				if d.HasChange("compute_resources.0.bid_percentage") {
					if bidPercentage, ok := d.GetOk("compute_resources.0.bid_percentage"); ok {
						computeResourceUpdate.BidPercentage = aws.Int64(int64(bidPercentage.(int)))
					} else {
						computeResourceUpdate.BidPercentage = aws.Int64(0)
					}
				}

				if d.HasChange("compute_resources.0.ec2_configuration") {
					defaultImageType := "ECS_AL2" // TODO remember to update this when adding EKS support
					ec2Configuration := d.Get("compute_resources.0.ec2_configuration").([]interface{})
					computeResourceUpdate.Ec2Configuration = ExpandComputeEnvironmentEc2ConfigurationUpdate(ec2Configuration, defaultImageType)
				}

				if d.HasChange("compute_resources.0.ec2_key_pair") {
					if keyPair, ok := d.GetOk("compute_resources.0.ec2_key_pair"); ok {
						computeResourceUpdate.Ec2KeyPair = aws.String(keyPair.(string))
					} else {
						computeResourceUpdate.Ec2KeyPair = aws.String("")
					}
				}

				if d.HasChange("compute_resources.0.image_id") {
					if imageId, ok := d.GetOk("compute_resources.0.image_id"); ok {
						computeResourceUpdate.ImageId = aws.String(imageId.(string))
					} else {
						computeResourceUpdate.ImageId = aws.String("")
					}
				}

				if d.HasChange("compute_resources.0.instance_role") {
					if instanceRole, ok := d.GetOk("compute_resources.0.instance_role"); ok {
						computeResourceUpdate.InstanceRole = aws.String(instanceRole.(string))
					} else {
						computeResourceUpdate.InstanceRole = aws.String("")
					}
				}

				if d.HasChange("compute_resources.0.instance_type") {
					computeResourceUpdate.InstanceTypes = flex.ExpandStringSet(d.Get("compute_resources.0.instance_type").(*schema.Set))
				}

				if d.HasChange("compute_resources.0.launch_template") {
					launchTemplate := d.Get("compute_resources.0.launch_template").([]interface{})
					computeResourceUpdate.LaunchTemplate = ExpandComputeEnvironmentLaunchTemplateUpdate(launchTemplate)
				}

				if d.HasChange("compute_resources.0.tags") {
					if tags, ok := d.GetOk("compute_resources.0.tags"); ok {
						computeResourceUpdate.Tags = Tags(tftags.New(tags.(map[string]interface{})).IgnoreAWS())
					} else {
						computeResourceUpdate.Tags = aws.StringMap(map[string]string{})
					}
				}
			}

			input.ComputeResources = computeResourceUpdate
		}

		log.Printf("[DEBUG] Updating Batch Compute Environment: %s", input)
		if _, err := conn.UpdateComputeEnvironment(input); err != nil {
			return fmt.Errorf("error updating Batch Compute Environment (%s): %w", d.Id(), err)
		}

		if _, err := waitComputeEnvironmentUpdated(conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return fmt.Errorf("error waiting for Batch Compute Environment (%s) update: %w", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating tags: %w", err)
		}
	}

	return resourceComputeEnvironmentRead(d, meta)
}

func ExpandComputeEnvironmentEc2ConfigurationUpdate(ec2Configuration []interface{}, defaultImageType string) []*batch.Ec2Configuration {
	if len(ec2Configuration) == 0 {
		return []*batch.Ec2Configuration{
			{
				ImageType: aws.String(defaultImageType),
			},
		}
	}

	results := make([]*batch.Ec2Configuration, 0)
	for _, ec2 := range ec2Configuration {
		result := &batch.Ec2Configuration{}
		m := ec2.(map[string]interface{})
		if v, ok := m["image_id_override"]; ok {
			result.ImageIdOverride = aws.String(v.(string))
		}
		if v, ok := m["image_type"]; ok {
			result.ImageType = aws.String(v.(string))
		}
		results = append(results, result)
	}

	return results
}

func ExpandComputeEnvironmentLaunchTemplateUpdate(launchTemplate []interface{}) *batch.LaunchTemplateSpecification {
	if len(launchTemplate) == 0 {
		// delete any existing launch template configuration
		return &batch.LaunchTemplateSpecification{
			LaunchTemplateId: aws.String(""),
		}
	}

	lts := &batch.LaunchTemplateSpecification{}
	m := launchTemplate[0].(map[string]interface{})
	if id, ok := m["launch_template_id"]; ok {
		lts.LaunchTemplateId = aws.String(id.(string))
	}
	if name, ok := m["launch_template_name"]; ok {
		lts.LaunchTemplateName = aws.String(name.(string))
	}
	if version, ok := m["version"]; ok {
		lts.Version = aws.String(version.(string))
	} else {
		lts.Version = aws.String("")
	}
	return lts
}

func resourceComputeEnvironmentDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).BatchConn

	log.Printf("[DEBUG] Disabling Batch Compute Environment (%s)", d.Id())
	{
		input := &batch.UpdateComputeEnvironmentInput{
			ComputeEnvironment: aws.String(d.Id()),
			State:              aws.String(batch.CEStateDisabled),
		}

		if _, err := conn.UpdateComputeEnvironment(input); err != nil {
			return fmt.Errorf("error disabling Batch Compute Environment (%s): %w", d.Id(), err)
		}

		if _, err := waitComputeEnvironmentDisabled(conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
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

		if _, err := waitComputeEnvironmentDeleted(conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
			return fmt.Errorf("error waiting for Batch Compute Environment (%s) delete: %w", d.Id(), err)
		}
	}

	return nil
}

func resourceComputeEnvironmentCustomizeDiff(_ context.Context, diff *schema.ResourceDiff, meta interface{}) error {
	if computeEnvironmentType := strings.ToUpper(diff.Get("type").(string)); computeEnvironmentType == batch.CETypeUnmanaged {
		// UNMANAGED compute environments can have no compute_resources configured.
		if v, ok := diff.GetOk("compute_resources"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			return fmt.Errorf("no `compute_resources` can be specified when `type` is %q", computeEnvironmentType)
		}
	}

	if diff.Id() != "" {
		// Update.

		fargateComputeResources := isFargateType(diff.Get("compute_resources.0.type").(string))

		if !isUpdatableComputeEnvironment(diff) {
			if diff.HasChange("compute_resources.0.security_group_ids") && !fargateComputeResources {
				if err := diff.ForceNew("compute_resources.0.security_group_ids"); err != nil {
					return err
				}
			}

			if diff.HasChange("compute_resources.0.subnets") && !fargateComputeResources {
				if err := diff.ForceNew("compute_resources.0.subnets"); err != nil {
					return err
				}
			}

			if diff.HasChange("compute_resources.0.allocation_strategy") {
				if err := diff.ForceNew("compute_resources.0.allocation_strategy"); err != nil {
					return err
				}
			}

			if diff.HasChange("compute_resources.0.bid_percentage") {
				if err := diff.ForceNew("compute_resources.0.bid_percentage"); err != nil {
					return err
				}
			}

			if diff.HasChange("compute_resources.0.ec2_configuration.#") {
				if err := diff.ForceNew("compute_resources.0.ec2_configuration.#"); err != nil {
					return err
				}
			}

			if diff.HasChange("compute_resources.0.ec2_configuration.0.image_id_override") {
				if err := diff.ForceNew("compute_resources.0.ec2_configuration.0.image_id_override"); err != nil {
					return err
				}
			}

			if diff.HasChange("compute_resources.0.ec2_configuration.0.image_type") {
				if err := diff.ForceNew("compute_resources.0.ec2_configuration.0.image_type"); err != nil {
					return err
				}
			}

			if diff.HasChange("compute_resources.0.ec2_key_pair") {
				if err := diff.ForceNew("compute_resources.0.ec2_key_pair"); err != nil {
					return err
				}
			}

			if diff.HasChange("compute_resources.0.image_id") {
				if err := diff.ForceNew("compute_resources.0.image_id"); err != nil {
					return err
				}
			}

			if diff.HasChange("compute_resources.0.instance_role") {
				if err := diff.ForceNew("compute_resources.0.instance_role"); err != nil {
					return err
				}
			}

			if diff.HasChange("compute_resources.0.instance_type") {
				if err := diff.ForceNew("compute_resources.0.instance_type"); err != nil {
					return err
				}
			}

			if diff.HasChange("compute_resources.0.launch_template.#") {
				if err := diff.ForceNew("compute_resources.0.launch_template.#"); err != nil {
					return err
				}
			}

			if diff.HasChange("compute_resources.0.launch_template.0.launch_template_id") {
				if err := diff.ForceNew("compute_resources.0.launch_template.0.launch_template_id"); err != nil {
					return err
				}
			}

			if diff.HasChange("compute_resources.0.launch_template.0.launch_template_name") {
				if err := diff.ForceNew("compute_resources.0.launch_template.0.launch_template_name"); err != nil {
					return err
				}
			}

			if diff.HasChange("compute_resources.0.launch_template.0.version") {
				if err := diff.ForceNew("compute_resources.0.launch_template.0.version"); err != nil {
					return err
				}
			}

			if diff.HasChange("compute_resources.0.tags") {
				if err := diff.ForceNew("compute_resources.0.tags"); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func isFargateType(computeResourceType string) bool {
	if computeResourceType == batch.CRTypeFargate || computeResourceType == batch.CRTypeFargateSpot {
		return true
	}
	return false
}

func isUpdatableComputeEnvironment(diff *schema.ResourceDiff) bool {
	if !isServiceLinkedRoleDiff(diff) {
		return false
	}
	if !isUpdatableAllocationStrategyDiff(diff) {
		return false
	}
	return true
}

func isServiceLinkedRoleDiff(diff *schema.ResourceDiff) bool {
	var before, after string
	if diff.HasChange("service_role") {
		beforeRaw, afterRaw := diff.GetChange("service_role")
		before, _ = beforeRaw.(string)
		after, _ := afterRaw.(string)
		return isServiceLinkedRole(before) && isServiceLinkedRole(after)
	}
	afterRaw, _ := diff.GetOk("service_role")
	after, _ = afterRaw.(string)
	return isServiceLinkedRole(after)
}

func isServiceLinkedRole(roleArn string) bool {
	if roleArn == "" {
		// Empty role ARN defaults to AWS service-linked role
		return true
	}
	re := regexp.MustCompile(`arn:[^:]+:iam::\d{12}:role/aws-service-role/batch\.amazonaws\.com/*`)
	return re.MatchString(roleArn)
}

func isUpdatableAllocationStrategyDiff(diff *schema.ResourceDiff) bool {
	var before, after string
	if computeResourcesCount, ok := diff.Get("compute_resources.#").(int); ok {
		if computeResourcesCount > 0 {
			if diff.HasChange("compute_resources.0.allocation_strategy") {
				beforeRaw, afterRaw := diff.GetChange("compute_resources.0.allocation_strategy")
				before, _ = beforeRaw.(string)
				after, _ = afterRaw.(string)
				return isUpdatableAllocationStrategy(before) && isUpdatableAllocationStrategy(after)
			}
			afterRaw, _ := diff.GetOk("compute_resources.0.allocation_strategy")
			after, _ := afterRaw.(string)
			return isUpdatableAllocationStrategy(after)
		}
	}
	return false
}

func isUpdatableAllocationStrategy(allocationStrategy string) bool {
	return allocationStrategy == batch.CRAllocationStrategyBestFitProgressive || allocationStrategy == batch.CRAllocationStrategySpotCapacityOptimized
}

func expandComputeResource(tfMap map[string]interface{}) *batch.ComputeResource {
	if tfMap == nil {
		return nil
	}

	var computeResourceType string

	if v, ok := tfMap["type"].(string); ok && v != "" {
		computeResourceType = v
	}

	apiObject := &batch.ComputeResource{}

	if v, ok := tfMap["allocation_strategy"].(string); ok && v != "" {
		apiObject.AllocationStrategy = aws.String(v)
	}

	if v, ok := tfMap["bid_percentage"].(int); ok && v != 0 {
		apiObject.BidPercentage = aws.Int64(int64(v))
	}

	if v, ok := tfMap["desired_vcpus"].(int); ok && v != 0 {
		apiObject.DesiredvCpus = aws.Int64(int64(v))
	}

	if v, ok := tfMap["ec2_configuration"].([]interface{}); ok && len(v) > 0 {
		apiObject.Ec2Configuration = expandEC2Configurations(v)
	}

	if v, ok := tfMap["ec2_key_pair"].(string); ok && v != "" {
		apiObject.Ec2KeyPair = aws.String(v)
	}

	if v, ok := tfMap["image_id"].(string); ok && v != "" {
		apiObject.ImageId = aws.String(v)
	}

	if v, ok := tfMap["instance_role"].(string); ok && v != "" {
		apiObject.InstanceRole = aws.String(v)
	}

	if v, ok := tfMap["instance_type"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.InstanceTypes = flex.ExpandStringSet(v)
	}

	if v, ok := tfMap["launch_template"].([]interface{}); ok && len(v) > 0 {
		apiObject.LaunchTemplate = expandLaunchTemplateSpecification(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["max_vcpus"].(int); ok && v != 0 {
		apiObject.MaxvCpus = aws.Int64(int64(v))
	}

	if v, ok := tfMap["min_vcpus"].(int); ok && v != 0 {
		apiObject.MinvCpus = aws.Int64(int64(v))
	} else if computeResourceType := strings.ToUpper(computeResourceType); computeResourceType == batch.CRTypeEc2 || computeResourceType == batch.CRTypeSpot {
		apiObject.MinvCpus = aws.Int64(0)
	}

	if v, ok := tfMap["security_group_ids"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.SecurityGroupIds = flex.ExpandStringSet(v)
	}

	if v, ok := tfMap["spot_iam_fleet_role"].(string); ok && v != "" {
		apiObject.SpotIamFleetRole = aws.String(v)
	}

	if v, ok := tfMap["subnets"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.Subnets = flex.ExpandStringSet(v)
	}

	if v, ok := tfMap["tags"].(map[string]interface{}); ok && len(v) > 0 {
		apiObject.Tags = Tags(tftags.New(v).IgnoreAWS())
	}

	if computeResourceType != "" {
		apiObject.Type = aws.String(computeResourceType)
	}

	return apiObject
}

func expandEC2Configuration(tfMap map[string]interface{}) *batch.Ec2Configuration {
	if tfMap == nil {
		return nil
	}

	apiObject := &batch.Ec2Configuration{}

	if v, ok := tfMap["image_id_override"].(string); ok && v != "" {
		apiObject.ImageIdOverride = aws.String(v)
	}

	if v, ok := tfMap["image_type"].(string); ok && v != "" {
		apiObject.ImageType = aws.String(v)
	}

	return apiObject
}

func expandEC2Configurations(tfList []interface{}) []*batch.Ec2Configuration {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*batch.Ec2Configuration

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandEC2Configuration(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandLaunchTemplateSpecification(tfMap map[string]interface{}) *batch.LaunchTemplateSpecification {
	if tfMap == nil {
		return nil
	}

	apiObject := &batch.LaunchTemplateSpecification{}

	if v, ok := tfMap["launch_template_id"].(string); ok && v != "" {
		apiObject.LaunchTemplateId = aws.String(v)
	}

	if v, ok := tfMap["launch_template_name"].(string); ok && v != "" {
		apiObject.LaunchTemplateName = aws.String(v)
	}

	if v, ok := tfMap["version"].(string); ok && v != "" {
		apiObject.Version = aws.String(v)
	}

	return apiObject
}

func flattenComputeResource(apiObject *batch.ComputeResource) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.AllocationStrategy; v != nil {
		tfMap["allocation_strategy"] = aws.StringValue(v)
	}

	if v := apiObject.BidPercentage; v != nil {
		tfMap["bid_percentage"] = aws.Int64Value(v)
	}

	if v := apiObject.DesiredvCpus; v != nil {
		tfMap["desired_vcpus"] = aws.Int64Value(v)
	}

	if v := apiObject.Ec2Configuration; v != nil {
		tfMap["ec2_configuration"] = flattenEC2Configurations(v)
	}

	if v := apiObject.Ec2KeyPair; v != nil {
		tfMap["ec2_key_pair"] = aws.StringValue(v)
	}

	if v := apiObject.ImageId; v != nil {
		tfMap["image_id"] = aws.StringValue(v)
	}

	if v := apiObject.InstanceRole; v != nil {
		tfMap["instance_role"] = aws.StringValue(v)
	}

	if v := apiObject.InstanceTypes; v != nil {
		tfMap["instance_type"] = aws.StringValueSlice(v)
	}

	if v := apiObject.LaunchTemplate; v != nil {
		tfMap["launch_template"] = []interface{}{flattenLaunchTemplateSpecification(v)}
	}

	if v := apiObject.MaxvCpus; v != nil {
		tfMap["max_vcpus"] = aws.Int64Value(v)
	}

	if v := apiObject.MinvCpus; v != nil {
		tfMap["min_vcpus"] = aws.Int64Value(v)
	}

	if v := apiObject.SecurityGroupIds; v != nil {
		tfMap["security_group_ids"] = aws.StringValueSlice(v)
	}

	if v := apiObject.SpotIamFleetRole; v != nil {
		tfMap["spot_iam_fleet_role"] = aws.StringValue(v)
	}

	if v := apiObject.Subnets; v != nil {
		tfMap["subnets"] = aws.StringValueSlice(v)
	}

	if v := apiObject.Tags; v != nil {
		tfMap["tags"] = KeyValueTags(v).IgnoreAWS().Map()
	}

	if v := apiObject.Type; v != nil {
		tfMap["type"] = aws.StringValue(v)
	}

	return tfMap
}

func flattenEC2Configuration(apiObject *batch.Ec2Configuration) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.ImageIdOverride; v != nil {
		tfMap["image_id_override"] = aws.StringValue(v)
	}

	if v := apiObject.ImageType; v != nil {
		tfMap["image_type"] = aws.StringValue(v)
	}

	return tfMap
}

func flattenEC2Configurations(apiObjects []*batch.Ec2Configuration) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenEC2Configuration(apiObject))
	}

	return tfList
}

func flattenLaunchTemplateSpecification(apiObject *batch.LaunchTemplateSpecification) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.LaunchTemplateId; v != nil {
		tfMap["launch_template_id"] = aws.StringValue(v)
	}

	if v := apiObject.LaunchTemplateName; v != nil {
		tfMap["launch_template_name"] = aws.StringValue(v)
	}

	if v := apiObject.Version; v != nil {
		tfMap["version"] = aws.StringValue(v)
	}

	return tfMap
}
