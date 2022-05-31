package devicefarm

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/devicefarm"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceDevicePool() *schema.Resource {
	return &schema.Resource{
		Create: resourceDevicePoolCreate,
		Read:   resourceDevicePoolRead,
		Update: resourceDevicePoolUpdate,
		Delete: resourceDevicePoolDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(0, 256),
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 16384),
			},
			"max_devices": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"project_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			"rule": {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"attribute": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice(devicefarm.DeviceAttribute_Values(), false),
						},
						"operator": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice(devicefarm.RuleOperator_Values(), false),
						},
						"value": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},
		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceDevicePoolCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DeviceFarmConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	name := d.Get("name").(string)
	input := &devicefarm.CreateDevicePoolInput{
		Name:       aws.String(name),
		ProjectArn: aws.String(d.Get("project_arn").(string)),
		Rules:      expandDevicePoolRules(d.Get("rule").(*schema.Set)),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("max_devices"); ok {
		input.MaxDevices = aws.Int64(int64(v.(int)))
	}

	log.Printf("[DEBUG] Creating DeviceFarm DevicePool: %s", name)
	out, err := conn.CreateDevicePool(input)
	if err != nil {
		return fmt.Errorf("Error creating DeviceFarm DevicePool: %w", err)
	}

	arn := aws.StringValue(out.DevicePool.Arn)
	log.Printf("[DEBUG] Successsfully Created DeviceFarm DevicePool: %s", arn)
	d.SetId(arn)

	if len(tags) > 0 {
		if err := UpdateTags(conn, arn, nil, tags); err != nil {
			return fmt.Errorf("error updating DeviceFarm DevicePool (%s) tags: %w", arn, err)
		}
	}

	return resourceDevicePoolRead(d, meta)
}

func resourceDevicePoolRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DeviceFarmConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	devicePool, err := FindDevicepoolByArn(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] DeviceFarm DevicePool (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading DeviceFarm DevicePool (%s): %w", d.Id(), err)
	}

	arn := aws.StringValue(devicePool.Arn)
	d.Set("name", devicePool.Name)
	d.Set("arn", arn)
	d.Set("description", devicePool.Description)
	d.Set("max_devices", devicePool.MaxDevices)

	projectArn, err := decodeProjectARN(arn, "devicepool", meta)
	if err != nil {
		return fmt.Errorf("error decoding project_arn (%s): %w", arn, err)
	}

	d.Set("project_arn", projectArn)

	if err := d.Set("rule", flattenDevicePoolRules(devicePool.Rules)); err != nil {
		return fmt.Errorf("error setting rule: %w", err)
	}

	tags, err := ListTags(conn, arn)

	if err != nil {
		return fmt.Errorf("error listing tags for DeviceFarm DevicePool (%s): %w", arn, err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceDevicePoolUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DeviceFarmConn

	if d.HasChangesExcept("tags", "tags_all") {
		input := &devicefarm.UpdateDevicePoolInput{
			Arn: aws.String(d.Id()),
		}

		if d.HasChange("name") {
			input.Name = aws.String(d.Get("name").(string))
		}

		if d.HasChange("description") {
			input.Description = aws.String(d.Get("description").(string))
		}

		if d.HasChange("rule") {
			input.Rules = expandDevicePoolRules(d.Get("rule").(*schema.Set))
		}

		if d.HasChange("max_devices") {
			if v, ok := d.GetOk("max_devices"); ok {
				input.MaxDevices = aws.Int64(int64(v.(int)))
			} else {
				input.ClearMaxDevices = aws.Bool(true)
			}
		}

		log.Printf("[DEBUG] Updating DeviceFarm DevicePool: %s", d.Id())
		_, err := conn.UpdateDevicePool(input)
		if err != nil {
			return fmt.Errorf("Error Updating DeviceFarm DevicePool: %w", err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating DeviceFarm DevicePool (%s) tags: %w", d.Get("arn").(string), err)
		}
	}

	return resourceDevicePoolRead(d, meta)
}

func resourceDevicePoolDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DeviceFarmConn

	input := &devicefarm.DeleteDevicePoolInput{
		Arn: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Deleting DeviceFarm DevicePool: %s", d.Id())
	_, err := conn.DeleteDevicePool(input)

	if tfawserr.ErrCodeEquals(err, devicefarm.ErrCodeNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("Error deleting DeviceFarm DevicePool: %w", err)
	}

	return nil
}

func expandDevicePoolRules(s *schema.Set) []*devicefarm.Rule {
	rules := make([]*devicefarm.Rule, 0)

	for _, r := range s.List() {
		rule := &devicefarm.Rule{}
		tfMap := r.(map[string]interface{})

		if v, ok := tfMap["attribute"].(string); ok && v != "" {
			rule.Attribute = aws.String(v)
		}

		if v, ok := tfMap["operator"].(string); ok && v != "" {
			rule.Operator = aws.String(v)
		}

		if v, ok := tfMap["value"].(string); ok && v != "" {
			rule.Value = aws.String(v)
		}

		rules = append(rules, rule)
	}
	return rules
}

func flattenDevicePoolRules(list []*devicefarm.Rule) []map[string]interface{} {
	if len(list) == 0 {
		return nil
	}

	result := make([]map[string]interface{}, 0, len(list))
	for _, setting := range list {
		l := map[string]interface{}{}

		if setting.Attribute != nil {
			l["attribute"] = aws.StringValue(setting.Attribute)
		}

		if setting.Operator != nil {
			l["operator"] = aws.StringValue(setting.Operator)
		}

		if setting.Value != nil {
			l["value"] = aws.StringValue(setting.Value)
		}

		result = append(result, l)
	}
	return result
}

func decodeProjectARN(id, typ string, meta interface{}) (string, error) {
	poolArn, err := arn.Parse(id)
	if err != nil {
		return "", fmt.Errorf("Error parsing '%s': %w", id, err)
	}

	poolArnResouce := poolArn.Resource
	parts := strings.Split(strings.TrimPrefix(poolArnResouce, fmt.Sprintf("%s:", typ)), "/")
	if len(parts) != 2 {
		return "", fmt.Errorf("Unexpected format of ID (%q), expected project-id/%q-id", poolArnResouce, typ)
	}

	projectId := parts[0]
	projectArn := arn.ARN{
		AccountID: meta.(*conns.AWSClient).AccountID,
		Partition: meta.(*conns.AWSClient).Partition,
		Region:    meta.(*conns.AWSClient).Region,
		Resource:  fmt.Sprintf("project:%s", projectId),
		Service:   devicefarm.ServiceName,
	}.String()

	return projectArn, nil
}
