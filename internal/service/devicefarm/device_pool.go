package devicefarm

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/devicefarm"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceDevicePool() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceDevicePoolCreate,
		ReadWithoutTimeout:   resourceDevicePoolRead,
		UpdateWithoutTimeout: resourceDevicePoolUpdate,
		DeleteWithoutTimeout: resourceDevicePoolDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
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

func resourceDevicePoolCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DeviceFarmConn()
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
	out, err := conn.CreateDevicePoolWithContext(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "Error creating DeviceFarm DevicePool: %s", err)
	}

	arn := aws.StringValue(out.DevicePool.Arn)
	log.Printf("[DEBUG] Successsfully Created DeviceFarm DevicePool: %s", arn)
	d.SetId(arn)

	if len(tags) > 0 {
		if err := UpdateTags(ctx, conn, arn, nil, tags); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating DeviceFarm DevicePool (%s) tags: %s", arn, err)
		}
	}

	return append(diags, resourceDevicePoolRead(ctx, d, meta)...)
}

func resourceDevicePoolRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DeviceFarmConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	devicePool, err := FindDevicePoolByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] DeviceFarm DevicePool (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading DeviceFarm DevicePool (%s): %s", d.Id(), err)
	}

	arn := aws.StringValue(devicePool.Arn)
	d.Set("name", devicePool.Name)
	d.Set("arn", arn)
	d.Set("description", devicePool.Description)
	d.Set("max_devices", devicePool.MaxDevices)

	projectArn, err := decodeProjectARN(arn, "devicepool", meta)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "decoding project_arn (%s): %s", arn, err)
	}

	d.Set("project_arn", projectArn)

	if err := d.Set("rule", flattenDevicePoolRules(devicePool.Rules)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting rule: %s", err)
	}

	tags, err := ListTags(ctx, conn, arn)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing tags for DeviceFarm DevicePool (%s): %s", arn, err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags_all: %s", err)
	}

	return diags
}

func resourceDevicePoolUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DeviceFarmConn()

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
		_, err := conn.UpdateDevicePoolWithContext(ctx, input)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "Error Updating DeviceFarm DevicePool: %s", err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(ctx, conn, d.Get("arn").(string), o, n); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating DeviceFarm DevicePool (%s) tags: %s", d.Get("arn").(string), err)
		}
	}

	return append(diags, resourceDevicePoolRead(ctx, d, meta)...)
}

func resourceDevicePoolDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DeviceFarmConn()

	input := &devicefarm.DeleteDevicePoolInput{
		Arn: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Deleting DeviceFarm DevicePool: %s", d.Id())
	_, err := conn.DeleteDevicePoolWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, devicefarm.ErrCodeNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "Error deleting DeviceFarm DevicePool: %s", err)
	}

	return diags
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
