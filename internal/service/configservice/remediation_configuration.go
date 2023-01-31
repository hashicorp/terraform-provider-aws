package configservice

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/configservice"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	// Maximum amount of time to wait for Config service eventual consistency on deletion
	remediationConfigurationDeletionTimeout = 2 * time.Minute
)

func ResourceRemediationConfiguration() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceRemediationConfigurationPut,
		ReadWithoutTimeout:   resourceRemediationConfigurationRead,
		UpdateWithoutTimeout: resourceRemediationConfigurationPut,
		DeleteWithoutTimeout: resourceRemediationConfigurationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"automatic": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"config_rule_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 64),
			},
			"execution_controls": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ssm_controls": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"concurrent_execution_rate_percentage": {
										Type:         schema.TypeInt,
										Optional:     true,
										ValidateFunc: validation.IntBetween(1, 100),
									},
									"error_percentage": {
										Type:         schema.TypeInt,
										Optional:     true,
										ValidateFunc: validation.IntBetween(1, 100),
									},
								},
							},
						},
					},
				},
			},
			"maximum_automatic_attempts": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(1, 25),
			},
			"parameter": {
				Type:     schema.TypeSet,
				MaxItems: 25,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"resource_value": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(0, 256),
						},
						"static_value": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"static_values": {
							Type:     schema.TypeList,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			"resource_type": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"retry_attempt_seconds": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(1, 2678000),
			},
			"target_id": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 256),
			},
			"target_type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(configservice.RemediationTargetType_Values(), false),
			},
			"target_version": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func resourceRemediationConfigurationPut(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConfigServiceConn()

	name := d.Get("config_rule_name").(string)
	input := configservice.RemediationConfiguration{
		ConfigRuleName: aws.String(name),
	}

	if v, ok := d.GetOk("parameter"); ok && v.(*schema.Set).Len() > 0 {
		input.Parameters = expandRemediationParameterValues(v.(*schema.Set).List())
	}
	if v, ok := d.GetOk("resource_type"); ok {
		input.ResourceType = aws.String(v.(string))
	}
	if v, ok := d.GetOk("target_id"); ok {
		input.TargetId = aws.String(v.(string))
	}
	if v, ok := d.GetOk("target_type"); ok {
		input.TargetType = aws.String(v.(string))
	}
	if v, ok := d.GetOk("target_version"); ok {
		input.TargetVersion = aws.String(v.(string))
	}
	if v, ok := d.GetOk("automatic"); ok {
		input.Automatic = aws.Bool(v.(bool))
	}
	if v, ok := d.GetOk("maximum_automatic_attempts"); ok {
		input.MaximumAutomaticAttempts = aws.Int64(int64(v.(int)))
	}
	if v, ok := d.GetOk("retry_attempt_seconds"); ok {
		input.RetryAttemptSeconds = aws.Int64(int64(v.(int)))
	}
	if v, ok := d.GetOk("execution_controls"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.ExecutionControls = expandExecutionControls(v.([]interface{})[0].(map[string]interface{}))
	}

	inputs := configservice.PutRemediationConfigurationsInput{
		RemediationConfigurations: []*configservice.RemediationConfiguration{&input},
	}

	log.Printf("[DEBUG] Creating AWSConfig remediation configuration: %s", inputs)
	_, err := conn.PutRemediationConfigurationsWithContext(ctx, &inputs)
	if err != nil {
		return create.DiagError(names.ConfigService, create.ErrActionCreating, ResNameRemediationConfiguration, fmt.Sprintf("%+v", inputs), err)
	}

	d.SetId(name)

	log.Printf("[DEBUG] AWSConfig config remediation configuration for rule %q created", name)

	return append(diags, resourceRemediationConfigurationRead(ctx, d, meta)...)
}

func resourceRemediationConfigurationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConfigServiceConn()
	out, err := conn.DescribeRemediationConfigurationsWithContext(ctx, &configservice.DescribeRemediationConfigurationsInput{
		ConfigRuleNames: []*string{aws.String(d.Id())},
	})

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, configservice.ErrCodeNoSuchConfigRuleException) {
		log.Printf("[WARN] Config Rule %q is gone (NoSuchConfigRuleException)", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return create.DiagError(names.ConfigService, create.ErrActionReading, ResNameRemediationConfiguration, d.Id(), err)
	}

	numberOfRemediationConfigurations := len(out.RemediationConfigurations)
	if !d.IsNewResource() && numberOfRemediationConfigurations < 1 {
		log.Printf("[WARN] No Remediation Configuration for Config Rule %q (no remediation configuration found)", d.Id())
		d.SetId("")
		return diags
	}

	if d.IsNewResource() && numberOfRemediationConfigurations < 1 {
		return create.DiagError(names.ConfigService, create.ErrActionReading, ResNameRemediationConfiguration, d.Id(), errors.New("none found after creation"))
	}

	log.Printf("[DEBUG] AWS Config remediation configurations received: %s", out)

	remediationConfiguration := out.RemediationConfigurations[0]
	d.Set("arn", remediationConfiguration.Arn)
	d.Set("config_rule_name", remediationConfiguration.ConfigRuleName)
	d.Set("resource_type", remediationConfiguration.ResourceType)
	d.Set("target_id", remediationConfiguration.TargetId)
	d.Set("target_type", remediationConfiguration.TargetType)
	d.Set("target_version", remediationConfiguration.TargetVersion)
	d.Set("automatic", remediationConfiguration.Automatic)
	d.Set("maximum_automatic_attempts", remediationConfiguration.MaximumAutomaticAttempts)
	d.Set("retry_attempt_seconds", remediationConfiguration.RetryAttemptSeconds)
	d.Set("maximum_automatic_attempts", remediationConfiguration.MaximumAutomaticAttempts)

	if err := d.Set("execution_controls", flattenExecutionControls(remediationConfiguration.ExecutionControls)); err != nil {
		return create.DiagError(names.ConfigService, create.ErrActionReading, ResNameRemediationConfiguration, d.Id(), err)
	}

	if err := d.Set("parameter", flattenRemediationParameterValues(remediationConfiguration.Parameters)); err != nil {
		return create.DiagError(names.ConfigService, create.ErrActionReading, ResNameRemediationConfiguration, d.Id(), err)
	}

	return diags
}

func resourceRemediationConfigurationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConfigServiceConn()

	name := d.Get("config_rule_name").(string)

	input := &configservice.DeleteRemediationConfigurationInput{
		ConfigRuleName: aws.String(name),
	}

	if v, ok := d.GetOk("resource_type"); ok {
		input.ResourceType = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Deleting AWS Config remediation configurations for rule %q", name)
	err := resource.RetryContext(ctx, remediationConfigurationDeletionTimeout, func() *resource.RetryError {
		_, err := conn.DeleteRemediationConfigurationWithContext(ctx, input)

		if tfawserr.ErrCodeEquals(err, configservice.ErrCodeResourceInUseException) {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		_, err = conn.DeleteRemediationConfigurationWithContext(ctx, input)
	}

	if err != nil {
		return create.DiagError(names.ConfigService, create.ErrActionDeleting, ResNameRemediationConfiguration, d.Id(), err)
	}

	return diags
}

func expandRemediationParameterValue(tfMap map[string]interface{}) *configservice.RemediationParameterValue {
	if tfMap == nil {
		return nil
	}

	apiObject := &configservice.RemediationParameterValue{}

	if v, ok := tfMap["resource_value"].(string); ok && v != "" {
		apiObject.ResourceValue = &configservice.ResourceValue{
			Value: aws.String(v),
		}
	}

	if v, ok := tfMap["static_value"].(string); ok && v != "" {
		apiObject.StaticValue = &configservice.StaticValue{
			Values: aws.StringSlice([]string{v}),
		}
	}

	if v, ok := tfMap["static_values"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.StaticValue = &configservice.StaticValue{
			Values: flex.ExpandStringList(v),
		}
	}

	return apiObject
}

func expandRemediationParameterValues(tfList []interface{}) map[string]*configservice.RemediationParameterValue {
	if len(tfList) == 0 {
		return nil
	}

	apiObjects := make(map[string]*configservice.RemediationParameterValue)

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		if v, ok := tfMap["name"].(string); !ok || v == "" {
			continue
		}

		apiObjects[tfMap["name"].(string)] = expandRemediationParameterValue(tfMap)
	}

	return apiObjects
}

func expandSSMControls(tfMap map[string]interface{}) *configservice.SsmControls {
	if tfMap == nil {
		return nil
	}

	apiObject := &configservice.SsmControls{}

	if v, ok := tfMap["concurrent_execution_rate_percentage"].(int); ok && v != 0 {
		apiObject.ConcurrentExecutionRatePercentage = aws.Int64(int64(v))
	}

	if v, ok := tfMap["error_percentage"].(int); ok && v != 0 {
		apiObject.ErrorPercentage = aws.Int64(int64(v))
	}

	return apiObject
}

func expandExecutionControls(tfMap map[string]interface{}) *configservice.ExecutionControls {
	if tfMap == nil {
		return nil
	}

	apiObject := &configservice.ExecutionControls{}

	if v, ok := tfMap["ssm_controls"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.SsmControls = expandSSMControls(v[0].(map[string]interface{}))
	}

	return apiObject
}

func flattenRemediationParameterValues(parameters map[string]*configservice.RemediationParameterValue) []interface{} {
	var items []interface{}

	for key, value := range parameters {
		item := make(map[string]interface{})
		item["name"] = key
		if v := value.ResourceValue; v != nil {
			item["resource_value"] = aws.StringValue(v.Value)
		}
		if v := value.StaticValue; v != nil && len(v.Values) > 1 {
			item["static_values"] = aws.StringValueSlice(v.Values)
		}
		if v := value.StaticValue; v != nil && len(v.Values) == 1 {
			item["static_value"] = aws.StringValue(v.Values[0])
		}

		items = append(items, item)
	}

	return items
}

func flattenExecutionControls(controls *configservice.ExecutionControls) []interface{} {
	if controls == nil {
		return nil
	}
	return []interface{}{map[string]interface{}{
		"ssm_controls": flattenSSMControls(controls.SsmControls),
	}}
}

func flattenSSMControls(controls *configservice.SsmControls) []interface{} {
	if controls == nil {
		return nil
	}
	m := make(map[string]interface{})
	if controls.ConcurrentExecutionRatePercentage != nil {
		m["concurrent_execution_rate_percentage"] = controls.ConcurrentExecutionRatePercentage
	}
	if controls.ErrorPercentage != nil {
		m["error_percentage"] = controls.ErrorPercentage
	}
	return []interface{}{m}
}
