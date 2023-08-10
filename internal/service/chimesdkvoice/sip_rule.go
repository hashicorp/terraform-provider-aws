// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package chimesdkvoice

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/chimesdkvoice"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

// @SDKResource("aws_chimesdkvoice_sip_rule", name="Sip Rule")
func ResourceSipRule() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceSipRuleCreate,
		ReadWithoutTimeout:   resourceSipRuleRead,
		UpdateWithoutTimeout: resourceSipRuleUpdate,
		DeleteWithoutTimeout: resourceSipRuleDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"disabled": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 256),
			},
			"target_applications": {
				Type:     schema.TypeSet,
				Required: true,
				MinItems: 1,
				MaxItems: 25,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"aws_region": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"priority": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntAtLeast(1),
						},
						"sip_media_application_id": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(1, 256),
						},
					},
				},
			},
			"trigger_type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(chimesdkvoice.SipRuleTriggerType_Values(), false),
			},
			"trigger_value": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceSipRuleCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ChimeSDKVoiceConn(ctx)

	input := &chimesdkvoice.CreateSipRuleInput{
		Name:               aws.String(d.Get("name").(string)),
		TriggerType:        aws.String(d.Get("trigger_type").(string)),
		TriggerValue:       aws.String(d.Get("trigger_value").(string)),
		TargetApplications: expandSipRuleTargetApplications(d.Get("target_applications").(*schema.Set).List()),
	}

	if v, ok := d.GetOk("disabled"); ok {
		input.Disabled = aws.Bool(v.(bool))
	}

	resp, err := conn.CreateSipRuleWithContext(ctx, input)

	if err != nil || resp.SipRule == nil {
		return sdkdiag.AppendErrorf(diags, "creating ChimeSKVoice Sip Rule: %s", err)
	}

	d.SetId(aws.StringValue(resp.SipRule.SipRuleId))

	return append(diags, resourceSipRuleRead(ctx, d, meta)...)
}

func resourceSipRuleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ChimeSDKVoiceConn(ctx)

	getInput := &chimesdkvoice.GetSipRuleInput{
		SipRuleId: aws.String(d.Id()),
	}

	resp, err := conn.GetSipRuleWithContext(ctx, getInput)
	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, chimesdkvoice.ErrCodeNotFoundException) {
		log.Printf("[WARN] ChimeSDKVoice Sip Rule %s not found", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil || resp.SipRule == nil {
		return sdkdiag.AppendErrorf(diags, "getting Sip Rule (%s): %s", d.Id(), err)
	}

	d.Set("name", resp.SipRule.Name)
	d.Set("disabled", resp.SipRule.Disabled)
	d.Set("trigger_type", resp.SipRule.TriggerType)
	d.Set("trigger_value", resp.SipRule.TriggerValue)
	d.Set("target_applications", flattenSipRuleTargetApplications(resp.SipRule.TargetApplications))
	return diags
}

func resourceSipRuleUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ChimeSDKVoiceConn(ctx)

	updateInput := &chimesdkvoice.UpdateSipRuleInput{
		SipRuleId: aws.String(d.Id()),
		Name:      aws.String(d.Get("name").(string)),
	}

	if d.HasChanges("target_applications") {
		updateInput.TargetApplications = expandSipRuleTargetApplications(d.Get("target_applications").(*schema.Set).List())
	}

	if d.HasChanges("disabled") {
		updateInput.Disabled = aws.Bool(d.Get("disabled").(bool))
	}

	if _, err := conn.UpdateSipRuleWithContext(ctx, updateInput); err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Sip Rule (%s): %s", d.Id(), err)
	}

	return append(diags, resourceSipRuleRead(ctx, d, meta)...)
}

func resourceSipRuleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ChimeSDKVoiceConn(ctx)

	input := &chimesdkvoice.DeleteSipRuleInput{
		SipRuleId: aws.String(d.Id()),
	}

	if _, err := conn.DeleteSipRuleWithContext(ctx, input); err != nil {
		if tfawserr.ErrCodeEquals(err, chimesdkvoice.ErrCodeNotFoundException) {
			log.Printf("[WARN] ChimeSDKVoice Sip Rule %s not found", d.Id())
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "deleting Sip Rule (%s)", d.Id())
	}

	return diags
}

func expandSipRuleTargetApplications(data []interface{}) []*chimesdkvoice.SipRuleTargetApplication {
	var targetApplications []*chimesdkvoice.SipRuleTargetApplication

	for _, rItem := range data {
		item := rItem.(map[string]interface{})
		application := &chimesdkvoice.SipRuleTargetApplication{
			SipMediaApplicationId: aws.String(item["sip_media_application_id"].(string)),
			Priority:              aws.Int64(int64(item["priority"].(int))),
			AwsRegion:             aws.String(item["aws_region"].(string)),
		}

		targetApplications = append(targetApplications, application)
	}

	return targetApplications
}

func flattenSipRuleTargetApplications(apiObject []*chimesdkvoice.SipRuleTargetApplication) []interface{} {
	var rawSipRuleTargetApplications []interface{}

	for _, e := range apiObject {
		rawTargetApplication := map[string]interface{}{
			"sip_media_application_id": aws.StringValue(e.SipMediaApplicationId),
			"priority":                 aws.Int64Value(e.Priority),
			"aws_region":               aws.StringValue(e.AwsRegion),
		}

		rawSipRuleTargetApplications = append(rawSipRuleTargetApplications, rawTargetApplication)
	}
	return rawSipRuleTargetApplications
}
