// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53recoverycontrolconfig

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	r53rcc "github.com/aws/aws-sdk-go-v2/service/route53recoverycontrolconfig"
	awstypes "github.com/aws/aws-sdk-go-v2/service/route53recoverycontrolconfig/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_route53recoverycontrolconfig_safety_rule", name="Safety Rule")
func resourceSafetyRule() *schema.Resource {
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
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[awstypes.RuleType](),
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

func resourceSafetyRuleCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	if _, ok := d.GetOk("asserted_controls"); ok {
		return append(diags, createAssertionRule(ctx, d, meta)...)
	}

	if _, ok := d.GetOk("gating_controls"); ok {
		return append(diags, createGatingRule(ctx, d, meta)...)
	}

	return diags
}

func resourceSafetyRuleRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53RecoveryControlConfigClient(ctx)

	output, err := findSafetyRuleByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Route53 Recovery Control Config Safety Rule (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "describing Route53 Recovery Control Config Safety Rule: %s", err)
	}

	if output.AssertionRule != nil {
		result := output.AssertionRule
		d.Set(names.AttrARN, result.SafetyRuleArn)
		d.Set("control_panel_arn", result.ControlPanelArn)
		d.Set(names.AttrName, result.Name)
		d.Set(names.AttrStatus, result.Status)
		d.Set("wait_period_ms", result.WaitPeriodMs)

		if err := d.Set("asserted_controls", flex.FlattenStringValueList(result.AssertedControls)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting asserted_controls: %s", err)
		}

		if result.RuleConfig != nil {
			d.Set("rule_config", []any{flattenRuleConfig(result.RuleConfig)})
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

		if err := d.Set("gating_controls", flex.FlattenStringValueList(result.GatingControls)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting gating_controls: %s", err)
		}

		if err := d.Set("target_controls", flex.FlattenStringValueList(result.TargetControls)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting target_controls: %s", err)
		}

		if result.RuleConfig != nil {
			d.Set("rule_config", []any{flattenRuleConfig(result.RuleConfig)})
		} else {
			d.Set("rule_config", nil)
		}
	}

	return diags
}

func resourceSafetyRuleUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	if _, ok := d.GetOk("asserted_controls"); ok {
		return append(diags, updateAssertionRule(ctx, d, meta)...)
	}

	if _, ok := d.GetOk("gating_controls"); ok {
		return append(diags, updateGatingRule(ctx, d, meta)...)
	}

	return diags
}

func resourceSafetyRuleDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53RecoveryControlConfigClient(ctx)

	log.Printf("[INFO] Deleting Route53 Recovery Control Config Safety Rule: %s", d.Id())
	_, err := conn.DeleteSafetyRule(ctx, &r53rcc.DeleteSafetyRuleInput{
		SafetyRuleArn: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Route53 Recovery Control Config Safety Rule: %s", err)
	}

	_, err = waitSafetyRuleDeleted(ctx, conn, d.Id())

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Route53 Recovery Control Config Safety Rule (%s) to be deleted: %s", d.Id(), err)
	}

	return diags
}

func createAssertionRule(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53RecoveryControlConfigClient(ctx)

	assertionRule := &awstypes.NewAssertionRule{
		Name:             aws.String(d.Get(names.AttrName).(string)),
		ControlPanelArn:  aws.String(d.Get("control_panel_arn").(string)),
		WaitPeriodMs:     aws.Int32(int32(d.Get("wait_period_ms").(int))),
		RuleConfig:       testAccSafetyRuleConfig_expandRule(d.Get("rule_config").([]any)[0].(map[string]any)),
		AssertedControls: flex.ExpandStringValueList(d.Get("asserted_controls").([]any)),
	}

	input := &r53rcc.CreateSafetyRuleInput{
		ClientToken:   aws.String(id.UniqueId()),
		AssertionRule: assertionRule,
	}

	output, err := conn.CreateSafetyRule(ctx, input)
	result := output.AssertionRule

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Route53 Recovery Control Config Assertion Rule: %s", err)
	}

	if result == nil {
		return sdkdiag.AppendErrorf(diags, "creating Route53 Recovery Control Config Assertion Rule empty response")
	}

	d.SetId(aws.ToString(result.SafetyRuleArn))

	if _, err := waitSafetyRuleCreated(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Route53 Recovery Control Config Assertion Rule (%s) to be Deployed: %s", d.Id(), err)
	}

	return append(diags, resourceSafetyRuleRead(ctx, d, meta)...)
}

func createGatingRule(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53RecoveryControlConfigClient(ctx)

	gatingRule := &awstypes.NewGatingRule{
		Name:            aws.String(d.Get(names.AttrName).(string)),
		ControlPanelArn: aws.String(d.Get("control_panel_arn").(string)),
		WaitPeriodMs:    aws.Int32(int32(d.Get("wait_period_ms").(int))),
		RuleConfig:      testAccSafetyRuleConfig_expandRule(d.Get("rule_config").([]any)[0].(map[string]any)),
		GatingControls:  flex.ExpandStringValueList(d.Get("gating_controls").([]any)),
		TargetControls:  flex.ExpandStringValueList(d.Get("target_controls").([]any)),
	}

	input := &r53rcc.CreateSafetyRuleInput{
		ClientToken: aws.String(id.UniqueId()),
		GatingRule:  gatingRule,
	}

	output, err := conn.CreateSafetyRule(ctx, input)
	result := output.GatingRule

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Route53 Recovery Control Config Gating Rule: %s", err)
	}

	if result == nil {
		return sdkdiag.AppendErrorf(diags, "creating Route53 Recovery Control Config Gating Rule empty response")
	}

	d.SetId(aws.ToString(result.SafetyRuleArn))

	if _, err := waitSafetyRuleCreated(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Route53 Recovery Control Config Assertion Rule (%s) to be Deployed: %s", d.Id(), err)
	}

	return append(diags, resourceSafetyRuleRead(ctx, d, meta)...)
}

func updateAssertionRule(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53RecoveryControlConfigClient(ctx)

	assertionRuleUpdate := &awstypes.AssertionRuleUpdate{
		SafetyRuleArn: aws.String(d.Get(names.AttrARN).(string)),
	}

	if d.HasChange(names.AttrName) {
		assertionRuleUpdate.Name = aws.String(d.Get(names.AttrName).(string))
	}

	if d.HasChange("wait_period_ms") {
		assertionRuleUpdate.WaitPeriodMs = aws.Int32(int32(d.Get("wait_period_ms").(int)))
	}

	input := &r53rcc.UpdateSafetyRuleInput{
		AssertionRuleUpdate: assertionRuleUpdate,
	}

	_, err := conn.UpdateSafetyRule(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Route53 Recovery Control Config Assertion Rule: %s", err)
	}

	return append(diags, sdkdiag.WrapDiagsf(resourceControlPanelRead(ctx, d, meta), "updating Route53 Recovery Control Config Assertion Rule")...)
}

func updateGatingRule(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53RecoveryControlConfigClient(ctx)

	gatingRuleUpdate := &awstypes.GatingRuleUpdate{
		SafetyRuleArn: aws.String(d.Get(names.AttrARN).(string)),
	}

	if d.HasChange(names.AttrName) {
		gatingRuleUpdate.Name = aws.String(d.Get(names.AttrName).(string))
	}

	if d.HasChange("wait_period_ms") {
		gatingRuleUpdate.WaitPeriodMs = aws.Int32(int32(d.Get("wait_period_ms").(int)))
	}

	input := &r53rcc.UpdateSafetyRuleInput{
		GatingRuleUpdate: gatingRuleUpdate,
	}

	_, err := conn.UpdateSafetyRule(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Route53 Recovery Control Config Gating Rule: %s", err)
	}

	return append(diags, sdkdiag.WrapDiagsf(resourceControlPanelRead(ctx, d, meta), "updating Route53 Recovery Control Config Gating Rule")...)
}

func findSafetyRuleByARN(ctx context.Context, conn *r53rcc.Client, arn string) (*r53rcc.DescribeSafetyRuleOutput, error) {
	input := &r53rcc.DescribeSafetyRuleInput{
		SafetyRuleArn: aws.String(arn),
	}

	output, err := conn.DescribeSafetyRule(ctx, input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}
	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func testAccSafetyRuleConfig_expandRule(tfMap map[string]any) *awstypes.RuleConfig {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.RuleConfig{}

	if v, ok := tfMap["inverted"].(bool); ok {
		apiObject.Inverted = aws.Bool(v)
	}

	if v, ok := tfMap["threshold"].(int); ok {
		apiObject.Threshold = aws.Int32(int32(v))
	}

	if v, ok := tfMap[names.AttrType].(string); ok && v != "" {
		apiObject.Type = awstypes.RuleType(v)
	}
	return apiObject
}

func flattenRuleConfig(apiObject *awstypes.RuleConfig) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if v := apiObject.Inverted; v != nil {
		tfMap["inverted"] = aws.ToBool(v)
	}

	if v := apiObject.Threshold; v != nil {
		tfMap["threshold"] = aws.ToInt32(v)
	}

	tfMap[names.AttrType] = apiObject.Type

	return tfMap
}
