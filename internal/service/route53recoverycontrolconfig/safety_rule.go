// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53recoverycontrolconfig

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	r53rcc "github.com/aws/aws-sdk-go/service/route53recoverycontrolconfig"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_route53recoverycontrolconfig_safety_rule")
func ResourceSafetyRule() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceSafetyRuleCreate,
		ReadWithoutTimeout:   resourceSafetyRuleRead,
		UpdateWithoutTimeout: resourceSafetyRuleUpdate,
		DeleteWithoutTimeout: resourceSafetyRuleDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
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
			names.AttrName: {
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
						names.AttrType: {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(r53rcc.RuleType_Values(), true),
						},
					},
				},
			},
			names.AttrStatus: {
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

func resourceSafetyRuleCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	if _, ok := d.GetOk("asserted_controls"); ok {
		return append(diags, createAssertionRule(ctx, d, meta)...)
	}

	if _, ok := d.GetOk("gating_controls"); ok {
		return append(diags, createGatingRule(ctx, d, meta)...)
	}

	return diags
}

func resourceSafetyRuleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53RecoveryControlConfigConn(ctx)

	input := &r53rcc.DescribeSafetyRuleInput{
		SafetyRuleArn: aws.String(d.Id()),
	}

	output, err := conn.DescribeSafetyRuleWithContext(ctx, input)

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, r53rcc.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] Route53 Recovery Control Config Safety Rule (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "describing Route53 Recovery Control Config Safety Rule: %s", err)
	}

	if output == nil {
		return sdkdiag.AppendErrorf(diags, "describing Route53 Recovery Control Config Safety Rule: %s", "empty response")
	}

	if output.AssertionRule != nil {
		result := output.AssertionRule
		d.Set(names.AttrARN, result.SafetyRuleArn)
		d.Set("control_panel_arn", result.ControlPanelArn)
		d.Set(names.AttrName, result.Name)
		d.Set(names.AttrStatus, result.Status)
		d.Set("wait_period_ms", result.WaitPeriodMs)

		if err := d.Set("asserted_controls", flex.FlattenStringList(result.AssertedControls)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting asserted_controls: %s", err)
		}

		if result.RuleConfig != nil {
			d.Set("rule_config", []interface{}{flattenRuleConfig(result.RuleConfig)})
		} else {
			d.Set("rule_config", nil)
		}

		return diags
	}

	if output.GatingRule != nil {
		result := output.GatingRule
		d.Set(names.AttrARN, result.SafetyRuleArn)
		d.Set("control_panel_arn", result.ControlPanelArn)
		d.Set(names.AttrName, result.Name)
		d.Set(names.AttrStatus, result.Status)
		d.Set("wait_period_ms", result.WaitPeriodMs)

		if err := d.Set("gating_controls", flex.FlattenStringList(result.GatingControls)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting gating_controls: %s", err)
		}

		if err := d.Set("target_controls", flex.FlattenStringList(result.TargetControls)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting target_controls: %s", err)
		}

		if result.RuleConfig != nil {
			d.Set("rule_config", []interface{}{flattenRuleConfig(result.RuleConfig)})
		} else {
			d.Set("rule_config", nil)
		}
	}

	return diags
}

func resourceSafetyRuleUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	if _, ok := d.GetOk("asserted_controls"); ok {
		return append(diags, updateAssertionRule(ctx, d, meta)...)
	}

	if _, ok := d.GetOk("gating_controls"); ok {
		return append(diags, updateGatingRule(ctx, d, meta)...)
	}

	return diags
}

func resourceSafetyRuleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53RecoveryControlConfigConn(ctx)

	log.Printf("[INFO] Deleting Route53 Recovery Control Config Safety Rule: %s", d.Id())
	_, err := conn.DeleteSafetyRuleWithContext(ctx, &r53rcc.DeleteSafetyRuleInput{
		SafetyRuleArn: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, r53rcc.ErrCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Route53 Recovery Control Config Safety Rule: %s", err)
	}

	_, err = waitSafetyRuleDeleted(ctx, conn, d.Id())

	if tfawserr.ErrCodeEquals(err, r53rcc.ErrCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Route53 Recovery Control Config Safety Rule (%s) to be deleted: %s", d.Id(), err)
	}

	return diags
}

func createAssertionRule(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53RecoveryControlConfigConn(ctx)

	assertionRule := &r53rcc.NewAssertionRule{
		Name:             aws.String(d.Get(names.AttrName).(string)),
		ControlPanelArn:  aws.String(d.Get("control_panel_arn").(string)),
		WaitPeriodMs:     aws.Int64(int64(d.Get("wait_period_ms").(int))),
		RuleConfig:       testAccSafetyRuleConfig_expandRule(d.Get("rule_config").([]interface{})[0].(map[string]interface{})),
		AssertedControls: flex.ExpandStringList(d.Get("asserted_controls").([]interface{})),
	}

	input := &r53rcc.CreateSafetyRuleInput{
		ClientToken:   aws.String(id.UniqueId()),
		AssertionRule: assertionRule,
	}

	output, err := conn.CreateSafetyRuleWithContext(ctx, input)
	result := output.AssertionRule

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Route53 Recovery Control Config Assertion Rule: %s", err)
	}

	if result == nil {
		return sdkdiag.AppendErrorf(diags, "creating Route53 Recovery Control Config Assertion Rule empty response")
	}

	d.SetId(aws.StringValue(result.SafetyRuleArn))

	if _, err := waitSafetyRuleCreated(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Route53 Recovery Control Config Assertion Rule (%s) to be Deployed: %s", d.Id(), err)
	}

	return append(diags, resourceSafetyRuleRead(ctx, d, meta)...)
}

func createGatingRule(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53RecoveryControlConfigConn(ctx)

	gatingRule := &r53rcc.NewGatingRule{
		Name:            aws.String(d.Get(names.AttrName).(string)),
		ControlPanelArn: aws.String(d.Get("control_panel_arn").(string)),
		WaitPeriodMs:    aws.Int64(int64(d.Get("wait_period_ms").(int))),
		RuleConfig:      testAccSafetyRuleConfig_expandRule(d.Get("rule_config").([]interface{})[0].(map[string]interface{})),
		GatingControls:  flex.ExpandStringList(d.Get("gating_controls").([]interface{})),
		TargetControls:  flex.ExpandStringList(d.Get("target_controls").([]interface{})),
	}

	input := &r53rcc.CreateSafetyRuleInput{
		ClientToken: aws.String(id.UniqueId()),
		GatingRule:  gatingRule,
	}

	output, err := conn.CreateSafetyRuleWithContext(ctx, input)
	result := output.GatingRule

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Route53 Recovery Control Config Gating Rule: %s", err)
	}

	if result == nil {
		return sdkdiag.AppendErrorf(diags, "creating Route53 Recovery Control Config Gating Rule empty response")
	}

	d.SetId(aws.StringValue(result.SafetyRuleArn))

	if _, err := waitSafetyRuleCreated(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Route53 Recovery Control Config Assertion Rule (%s) to be Deployed: %s", d.Id(), err)
	}

	return append(diags, resourceSafetyRuleRead(ctx, d, meta)...)
}

func updateAssertionRule(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53RecoveryControlConfigConn(ctx)

	assertionRuleUpdate := &r53rcc.AssertionRuleUpdate{
		SafetyRuleArn: aws.String(d.Get(names.AttrARN).(string)),
	}

	if d.HasChange(names.AttrName) {
		assertionRuleUpdate.Name = aws.String(d.Get(names.AttrName).(string))
	}

	if d.HasChange("wait_period_ms") {
		assertionRuleUpdate.WaitPeriodMs = aws.Int64(int64(d.Get("wait_period_ms").(int)))
	}

	input := &r53rcc.UpdateSafetyRuleInput{
		AssertionRuleUpdate: assertionRuleUpdate,
	}

	_, err := conn.UpdateSafetyRuleWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Route53 Recovery Control Config Assertion Rule: %s", err)
	}

	return append(diags, sdkdiag.WrapDiagsf(resourceControlPanelRead(ctx, d, meta), "updating Route53 Recovery Control Config Assertion Rule")...)
}

func updateGatingRule(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53RecoveryControlConfigConn(ctx)

	gatingRuleUpdate := &r53rcc.GatingRuleUpdate{
		SafetyRuleArn: aws.String(d.Get(names.AttrARN).(string)),
	}

	if d.HasChange(names.AttrName) {
		gatingRuleUpdate.Name = aws.String(d.Get(names.AttrName).(string))
	}

	if d.HasChange("wait_period_ms") {
		gatingRuleUpdate.WaitPeriodMs = aws.Int64(int64(d.Get("wait_period_ms").(int)))
	}

	input := &r53rcc.UpdateSafetyRuleInput{
		GatingRuleUpdate: gatingRuleUpdate,
	}

	_, err := conn.UpdateSafetyRuleWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Route53 Recovery Control Config Gating Rule: %s", err)
	}

	return append(diags, sdkdiag.WrapDiagsf(resourceControlPanelRead(ctx, d, meta), "updating Route53 Recovery Control Config Gating Rule")...)
}

func testAccSafetyRuleConfig_expandRule(tfMap map[string]interface{}) *r53rcc.RuleConfig {
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

	if v, ok := tfMap[names.AttrType].(string); ok && v != "" {
		apiObject.Type = aws.String(v)
	}
	return apiObject
}

func flattenRuleConfig(apiObject *r53rcc.RuleConfig) map[string]interface{} {
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
		tfMap[names.AttrType] = aws.StringValue(v)
	}

	return tfMap
}
