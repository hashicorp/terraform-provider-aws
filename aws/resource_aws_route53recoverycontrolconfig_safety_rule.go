package aws

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53recoverycontrolconfig"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/route53recoverycontrolconfig"
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
			"safety_rule_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"control_panel_arn": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"wait_period_ms": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"inverted": {
				Type:     schema.TypeBool,
				Required: true,
			},
			"threshold": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"rule_type": {
				Type:     schema.TypeString,
				Required: true,
				StateFunc: func(val interface{}) string {
					return strings.ToUpper(val.(string))
				},
				ValidateFunc: validation.StringInSlice([]string{
					route53recoverycontrolconfig.RuleTypeAtleast,
					route53recoverycontrolconfig.RuleTypeAnd,
					route53recoverycontrolconfig.RuleTypeOr,
				}, true),
			},
			"asserted_controls": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type:     schema.TypeString,
					Required: true,
				},
			},
			"gating_controls": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type:     schema.TypeString,
					Required: true,
				},
			},
			"target_controls": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type:     schema.TypeString,
					Required: true,
				},
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

	return nil
}

func resourceAwsRoute53RecoveryControlConfigSafetyRuleRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).route53recoverycontrolconfigconn

	input := &route53recoverycontrolconfig.DescribeSafetyRuleInput{
		SafetyRuleArn: aws.String(d.Id()),
	}

	output, err := conn.DescribeSafetyRule(input)

	if err != nil {
		return fmt.Errorf("Error describing Route53 Recovery Control Config Safety Rule: %s", err)
	}

	if !d.IsNewResource() && output == nil {
		log.Printf("[WARN] Route53 Recovery Control Config Safety Rule (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if output.AssertionRule != nil {
		result := output.AssertionRule
		d.Set("safety_rule_arn", result.SafetyRuleArn)
		d.Set("control_panel_arn", result.ControlPanelArn)
		d.Set("name", result.Name)
		d.Set("status", result.Status)
		d.Set("wait_period_ms", result.WaitPeriodMs)

		if err := d.Set("asserted_controls", flattenStringList(result.AssertedControls)); err != nil {
			return fmt.Errorf("Error setting asserted_controls: %w", err)
		}

		if result.RuleConfig != nil {
			d.Set("inverted", result.RuleConfig.Inverted)
			d.Set("threshold", result.RuleConfig.Threshold)
			d.Set("type", result.RuleConfig.Type)
		}

	}

	if output.GatingRule != nil {
		result := output.GatingRule
		d.Set("safety_rule_arn", result.SafetyRuleArn)
		d.Set("control_panel_arn", result.ControlPanelArn)
		d.Set("name", result.Name)
		d.Set("status", result.Status)
		d.Set("wait_period_ms", result.WaitPeriodMs)

		if err := d.Set("gating_controls", flattenStringList(result.GatingControls)); err != nil {
			return fmt.Errorf("Error setting gating_controls: %w", err)
		}

		if err := d.Set("target_controls", flattenStringList(result.TargetControls)); err != nil {
			return fmt.Errorf("Error setting target_controls: %w", err)
		}

		if result.RuleConfig != nil {
			d.Set("inverted", result.RuleConfig.Inverted)
			d.Set("threshold", result.RuleConfig.Threshold)
			d.Set("type", result.RuleConfig.Type)
		}

	}
	return nil
}

func resourceAwsRoute53RecoveryControlConfigSafetyRuleUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).route53recoverycontrolconfigconn

	// First check type of rule
	describeRuleInput := &route53recoverycontrolconfig.DescribeSafetyRuleInput{
		SafetyRuleArn: aws.String(d.Get("safety_rule_arn").(string)),
	}

	output, err := conn.DescribeSafetyRule(describeRuleInput)

	if err != nil {
		return fmt.Errorf("Error describing Route53 Recovery Control Config Safety Rule: %s", err)
	}

	if output.AssertionRule != nil {
		return updateAssertionRule(d, meta)
	}

	if output.GatingRule != nil {
		return updateGatingRule(d, meta)
	}

	return nil
}

func resourceAwsRoute53RecoveryControlConfigSafetyRuleDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).route53recoverycontrolconfigconn

	input := &route53recoverycontrolconfig.DeleteSafetyRuleInput{
		SafetyRuleArn: aws.String(d.Id()),
	}

	_, err := conn.DeleteSafetyRule(input)

	if err != nil {
		if isAWSErr(err, route53recoverycontrolconfig.ErrCodeResourceNotFoundException, "") {
			return nil
		}
		return fmt.Errorf("Error deleting Route53 Recovery Control Config Safety Rule: %s", err)
	}

	if _, err := waiter.Route53RecoveryControlConfigSafetyRuleDeleted(conn, d.Id()); err != nil {
		if isResourceNotFoundError(err) {
			return nil
		}
		return fmt.Errorf("Error waiting for Route53 Recovery Control Config Safety Rule (%s) to be deleted: %w", d.Id(), err)
	}

	return nil
}

func createAssertionRule(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).route53recoverycontrolconfigconn

	ruleConfig := &route53recoverycontrolconfig.RuleConfig{
		Inverted:  aws.Bool(d.Get("inverted").(bool)),
		Threshold: aws.Int64(int64(d.Get("threshold").(int))),
		Type:      aws.String(d.Get("rule_type").(string)),
	}

	assertionRule := &route53recoverycontrolconfig.NewAssertionRule{
		Name:             aws.String(d.Get("name").(string)),
		ControlPanelArn:  aws.String(d.Get("control_panel_arn").(string)),
		WaitPeriodMs:     aws.Int64(int64(d.Get("wait_period_ms").(int))),
		RuleConfig:       ruleConfig,
		AssertedControls: expandStringList(d.Get("asserted_controls").([]interface{})),
	}

	input := &route53recoverycontrolconfig.CreateSafetyRuleInput{
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
	conn := meta.(*AWSClient).route53recoverycontrolconfigconn

	ruleConfig := &route53recoverycontrolconfig.RuleConfig{
		Inverted:  aws.Bool(d.Get("inverted").(bool)),
		Threshold: aws.Int64(int64(d.Get("threshold").(int))),
		Type:      aws.String(d.Get("rule_type").(string)),
	}

	gatingRule := &route53recoverycontrolconfig.NewGatingRule{
		Name:            aws.String(d.Get("name").(string)),
		ControlPanelArn: aws.String(d.Get("control_panel_arn").(string)),
		WaitPeriodMs:    aws.Int64(int64(d.Get("wait_period_ms").(int))),
		RuleConfig:      ruleConfig,
		GatingControls:  expandStringList(d.Get("gating_controls").([]interface{})),
		TargetControls:  expandStringList(d.Get("target_controls").([]interface{})),
	}

	input := &route53recoverycontrolconfig.CreateSafetyRuleInput{
		ClientToken: aws.String(resource.UniqueId()),
		GatingRule:  gatingRule,
	}

	output, err := conn.CreateSafetyRule(input)
	result := output.AssertionRule

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
	conn := meta.(*AWSClient).route53recoverycontrolconfigconn

	assertionRuleUpdate := &route53recoverycontrolconfig.AssertionRuleUpdate{
		SafetyRuleArn: aws.String(d.Get("safety_rule_arn").(string)),
	}

	if d.HasChange("name") {
		assertionRuleUpdate.Name = aws.String(d.Get("name").(string))
	}

	if d.HasChange("wait_period_ms") {
		assertionRuleUpdate.WaitPeriodMs = aws.Int64(int64(d.Get("wait_period_ms").(int)))
	}

	input := &route53recoverycontrolconfig.UpdateSafetyRuleInput{
		AssertionRuleUpdate: assertionRuleUpdate,
	}

	_, err := conn.UpdateSafetyRule(input)
	if err != nil {
		return fmt.Errorf("error updating Route53 Recovery Control Config Assertion Rule: %s", err)
	}

	return resourceAwsRoute53RecoveryControlConfigControlPanelRead(d, meta)
}

func updateGatingRule(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).route53recoverycontrolconfigconn

	gatingRuleUpdate := &route53recoverycontrolconfig.GatingRuleUpdate{
		SafetyRuleArn: aws.String(d.Get("safety_rule_arn").(string)),
	}

	if d.HasChange("name") {
		gatingRuleUpdate.Name = aws.String(d.Get("name").(string))
	}

	if d.HasChange("wait_period_ms") {
		gatingRuleUpdate.WaitPeriodMs = aws.Int64(int64(d.Get("wait_period_ms").(int)))
	}

	input := &route53recoverycontrolconfig.UpdateSafetyRuleInput{
		GatingRuleUpdate: gatingRuleUpdate,
	}

	_, err := conn.UpdateSafetyRule(input)
	if err != nil {
		return fmt.Errorf("error updating Route53 Recovery Control Config Gating Rule: %s", err)
	}

	return resourceAwsRoute53RecoveryControlConfigControlPanelRead(d, meta)
}
