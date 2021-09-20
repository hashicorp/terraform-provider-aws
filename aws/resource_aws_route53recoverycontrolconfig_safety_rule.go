package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	r53rcc "github.com/aws/aws-sdk-go/service/route53recoverycontrolconfig"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/service/route53recoverycontrolconfig/waiter"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func resourceAwsRoute53RecoveryControlConfigSafetyRule() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsRoute53RecoveryControlConfigSafetyRuleCreate,
		Read:   resourceAwsRoute53RecoveryControlConfigSafetyRuleRead,
		Update: resourceAwsRoute53RecoveryControlConfigSafetyRuleUpdate,
		Delete: resourceAwsRoute53RecoveryControlConfigSafetyRuleDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"asserted_controls": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				ExactlyOneOf: []string{
					"asserted_controls",
					"gating_controls",
				},
			},
			"control_panel_arn": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"gating_controls": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				ExactlyOneOf: []string{
					"asserted_controls",
					"gating_controls",
				},
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"rule_config": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"inverted": {
							Type:     schema.TypeBool,
							Required: true,
						},
						"threshold": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(r53rcc.RuleType_Values(), true),
						},
					},
				},
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"target_controls": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				RequiredWith: []string{
					"gating_controls",
				},
			},
			"wait_period_ms": {
				Type:     schema.TypeInt,
				Required: true,
			},
		},
	}
}

func resourceAwsRoute53RecoveryControlConfigSafetyRuleCreate(d *schema.ResourceData, meta interface{}) error {
	if _, ok := d.GetOk("asserted_controls"); ok {
		return createAssertionRule(d, meta)
	}

	if _, ok := d.GetOk("gating_controls"); ok {
		return createGatingRule(d, meta)
	}

	return fmt.Errorf("Expecting either asserted_controls or gating_controls,  but none found")
}

func resourceAwsRoute53RecoveryControlConfigSafetyRuleRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).Route53RecoveryControlConfigConn

	input := &r53rcc.DescribeSafetyRuleInput{
		SafetyRuleArn: aws.String(d.Id()),
	}

	output, err := conn.DescribeSafetyRule(input)

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, r53rcc.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] Route53 Recovery Control Config Safety Rule (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("Error describing Route53 Recovery Control Config Safety Rule: %s", err)
	}

	if output == nil {
		return fmt.Errorf("Error describing Route53 Recovery Control Config Safety Rule: %s", "empty response")
	}

	if output.AssertionRule != nil {
		result := output.AssertionRule
		d.Set("arn", result.SafetyRuleArn)
		d.Set("control_panel_arn", result.ControlPanelArn)
		d.Set("name", result.Name)
		d.Set("status", result.Status)
		d.Set("wait_period_ms", result.WaitPeriodMs)

		if err := d.Set("asserted_controls", flex.FlattenStringList(result.AssertedControls)); err != nil {
			return fmt.Errorf("Error setting asserted_controls: %w", err)
		}

		if result.RuleConfig != nil {
			d.Set("rule_config", []interface{}{flattenRoute53RecoveryControlConfigRuleConfig(result.RuleConfig)})
		} else {
			d.Set("rule_config", nil)
		}

		return nil
	}

	if output.GatingRule != nil {
		result := output.GatingRule
		d.Set("arn", result.SafetyRuleArn)
		d.Set("control_panel_arn", result.ControlPanelArn)
		d.Set("name", result.Name)
		d.Set("status", result.Status)
		d.Set("wait_period_ms", result.WaitPeriodMs)

		if err := d.Set("gating_controls", flex.FlattenStringList(result.GatingControls)); err != nil {
			return fmt.Errorf("Error setting gating_controls: %w", err)
		}

		if err := d.Set("target_controls", flex.FlattenStringList(result.TargetControls)); err != nil {
			return fmt.Errorf("Error setting target_controls: %w", err)
		}

		if result.RuleConfig != nil {
			d.Set("rule_config", []interface{}{flattenRoute53RecoveryControlConfigRuleConfig(result.RuleConfig)})
		} else {
			d.Set("rule_config", nil)
		}
	}

	return nil
}

func resourceAwsRoute53RecoveryControlConfigSafetyRuleUpdate(d *schema.ResourceData, meta interface{}) error {
	if _, ok := d.GetOk("asserted_controls"); ok {
		return updateAssertionRule(d, meta)
	}

	if _, ok := d.GetOk("gating_controls"); ok {
		return updateGatingRule(d, meta)
	}

	return nil
}

func resourceAwsRoute53RecoveryControlConfigSafetyRuleDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).Route53RecoveryControlConfigConn

	input := &r53rcc.DeleteSafetyRuleInput{
		SafetyRuleArn: aws.String(d.Id()),
	}

	_, err := conn.DeleteSafetyRule(input)

	if tfawserr.ErrCodeEquals(err, r53rcc.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("Error deleting Route53 Recovery Control Config Safety Rule: %s", err)
	}

	_, err = waiter.Route53RecoveryControlConfigSafetyRuleDeleted(conn, d.Id())

	if tfawserr.ErrCodeEquals(err, r53rcc.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("Error waiting for Route53 Recovery Control Config Safety Rule (%s) to be deleted: %w", d.Id(), err)
	}

	return nil
}

func createAssertionRule(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).Route53RecoveryControlConfigConn

	assertionRule := &r53rcc.NewAssertionRule{
		Name:             aws.String(d.Get("name").(string)),
		ControlPanelArn:  aws.String(d.Get("control_panel_arn").(string)),
		WaitPeriodMs:     aws.Int64(int64(d.Get("wait_period_ms").(int))),
		RuleConfig:       expandRoute53RecoveryControlConfigRuleConfig(d.Get("rule_config").([]interface{})[0].(map[string]interface{})),
		AssertedControls: flex.ExpandStringList(d.Get("asserted_controls").([]interface{})),
	}

	input := &r53rcc.CreateSafetyRuleInput{
		ClientToken:   aws.String(resource.UniqueId()),
		AssertionRule: assertionRule,
	}

	output, err := conn.CreateSafetyRule(input)
	result := output.AssertionRule

	if err != nil {
		return fmt.Errorf("Error creating Route53 Recovery Control Config Assertion Rule: %w", err)
	}

	if result == nil {
		return fmt.Errorf("Error creating Route53 Recovery Control Config Assertion Rule empty response")
	}

	d.SetId(aws.StringValue(result.SafetyRuleArn))

	if _, err := waiter.Route53RecoveryControlConfigSafetyRuleCreated(conn, d.Id()); err != nil {
		return fmt.Errorf("Error waiting for Route53 Recovery Control Config Assertion Rule (%s) to be Deployed: %w", d.Id(), err)
	}

	return resourceAwsRoute53RecoveryControlConfigSafetyRuleRead(d, meta)
}

func createGatingRule(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).Route53RecoveryControlConfigConn

	gatingRule := &r53rcc.NewGatingRule{
		Name:            aws.String(d.Get("name").(string)),
		ControlPanelArn: aws.String(d.Get("control_panel_arn").(string)),
		WaitPeriodMs:    aws.Int64(int64(d.Get("wait_period_ms").(int))),
		RuleConfig:      expandRoute53RecoveryControlConfigRuleConfig(d.Get("rule_config").([]interface{})[0].(map[string]interface{})),
		GatingControls:  flex.ExpandStringList(d.Get("gating_controls").([]interface{})),
		TargetControls:  flex.ExpandStringList(d.Get("target_controls").([]interface{})),
	}

	input := &r53rcc.CreateSafetyRuleInput{
		ClientToken: aws.String(resource.UniqueId()),
		GatingRule:  gatingRule,
	}

	output, err := conn.CreateSafetyRule(input)
	result := output.GatingRule

	if err != nil {
		return fmt.Errorf("Error creating Route53 Recovery Control Config Gating Rule: %w", err)
	}

	if result == nil {
		return fmt.Errorf("Error creating Route53 Recovery Control Config Gating Rule empty response")
	}

	d.SetId(aws.StringValue(result.SafetyRuleArn))

	if _, err := waiter.Route53RecoveryControlConfigSafetyRuleCreated(conn, d.Id()); err != nil {
		return fmt.Errorf("Error waiting for Route53 Recovery Control Config Assertion Rule (%s) to be Deployed: %w", d.Id(), err)
	}

	return resourceAwsRoute53RecoveryControlConfigSafetyRuleRead(d, meta)
}

func updateAssertionRule(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).Route53RecoveryControlConfigConn

	assertionRuleUpdate := &r53rcc.AssertionRuleUpdate{
		SafetyRuleArn: aws.String(d.Get("arn").(string)),
	}

	if d.HasChange("name") {
		assertionRuleUpdate.Name = aws.String(d.Get("name").(string))
	}

	if d.HasChange("wait_period_ms") {
		assertionRuleUpdate.WaitPeriodMs = aws.Int64(int64(d.Get("wait_period_ms").(int)))
	}

	input := &r53rcc.UpdateSafetyRuleInput{
		AssertionRuleUpdate: assertionRuleUpdate,
	}

	_, err := conn.UpdateSafetyRule(input)

	if err != nil {
		return fmt.Errorf("error updating Route53 Recovery Control Config Assertion Rule: %s", err)
	}

	return resourceAwsRoute53RecoveryControlConfigControlPanelRead(d, meta)
}

func updateGatingRule(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).Route53RecoveryControlConfigConn

	gatingRuleUpdate := &r53rcc.GatingRuleUpdate{
		SafetyRuleArn: aws.String(d.Get("arn").(string)),
	}

	if d.HasChange("name") {
		gatingRuleUpdate.Name = aws.String(d.Get("name").(string))
	}

	if d.HasChange("wait_period_ms") {
		gatingRuleUpdate.WaitPeriodMs = aws.Int64(int64(d.Get("wait_period_ms").(int)))
	}

	input := &r53rcc.UpdateSafetyRuleInput{
		GatingRuleUpdate: gatingRuleUpdate,
	}

	_, err := conn.UpdateSafetyRule(input)

	if err != nil {
		return fmt.Errorf("error updating Route53 Recovery Control Config Gating Rule: %s", err)
	}

	return resourceAwsRoute53RecoveryControlConfigControlPanelRead(d, meta)
}

func expandRoute53RecoveryControlConfigRuleConfig(tfMap map[string]interface{}) *r53rcc.RuleConfig {
	if tfMap == nil {
		return nil
	}

	apiObject := &r53rcc.RuleConfig{}

	if v, ok := tfMap["inverted"].(bool); ok {
		apiObject.Inverted = aws.Bool(v)
	}

	if v, ok := tfMap["threshold"].(int); ok {
		apiObject.Threshold = aws.Int64(int64(v))
	}

	if v, ok := tfMap["type"].(string); ok && v != "" {
		apiObject.Type = aws.String(v)
	}
	return apiObject
}

func flattenRoute53RecoveryControlConfigRuleConfig(apiObject *r53rcc.RuleConfig) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Inverted; v != nil {
		tfMap["inverted"] = aws.BoolValue(v)
	}

	if v := apiObject.Threshold; v != nil {
		tfMap["threshold"] = aws.Int64Value(v)
	}

	if v := apiObject.Type; v != nil {
		tfMap["type"] = aws.StringValue(v)
	}

	return tfMap
}
