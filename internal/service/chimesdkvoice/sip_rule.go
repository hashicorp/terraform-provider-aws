// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package chimesdkvoice

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/chimesdkvoice"
	awstypes "github.com/aws/aws-sdk-go-v2/service/chimesdkvoice/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
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
			names.AttrName: {
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
						names.AttrPriority: {
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
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: enum.Validate[awstypes.SipRuleTriggerType](),
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
	conn := meta.(*conns.AWSClient).ChimeSDKVoiceClient(ctx)

	input := &chimesdkvoice.CreateSipRuleInput{
		Name:               aws.String(d.Get(names.AttrName).(string)),
		TriggerType:        awstypes.SipRuleTriggerType(d.Get("trigger_type").(string)),
		TriggerValue:       aws.String(d.Get("trigger_value").(string)),
		TargetApplications: expandSipRuleTargetApplications(d.Get("target_applications").(*schema.Set).List()),
	}

	if v, ok := d.GetOk("disabled"); ok {
		input.Disabled = aws.Bool(v.(bool))
	}

	resp, err := conn.CreateSipRule(ctx, input)

	if err != nil || resp.SipRule == nil {
		return sdkdiag.AppendErrorf(diags, "creating ChimeSKVoice Sip Rule: %s", err)
	}

	d.SetId(aws.ToString(resp.SipRule.SipRuleId))

	return append(diags, resourceSipRuleRead(ctx, d, meta)...)
}

func resourceSipRuleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ChimeSDKVoiceClient(ctx)

	resp, err := FindSIPResourceWithRetry(ctx, d.IsNewResource(), func() (*awstypes.SipRule, error) {
		return findSIPRuleByID(ctx, conn, d.Id())
	})

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] ChimeSDKVoice Sip Rule %s not found", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ChimeSKVoice Sip Rule (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrName, resp.Name)
	d.Set("disabled", resp.Disabled)
	d.Set("trigger_type", resp.TriggerType)
	d.Set("trigger_value", resp.TriggerValue)
	d.Set("target_applications", flattenSipRuleTargetApplications(resp.TargetApplications))
	return diags
}

func resourceSipRuleUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ChimeSDKVoiceClient(ctx)

	updateInput := &chimesdkvoice.UpdateSipRuleInput{
		SipRuleId: aws.String(d.Id()),
		Name:      aws.String(d.Get(names.AttrName).(string)),
	}

	if d.HasChanges("target_applications") {
		updateInput.TargetApplications = expandSipRuleTargetApplications(d.Get("target_applications").(*schema.Set).List())
	}

	if d.HasChanges("disabled") {
		updateInput.Disabled = aws.Bool(d.Get("disabled").(bool))
	}

	if _, err := conn.UpdateSipRule(ctx, updateInput); err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Sip Rule (%s): %s", d.Id(), err)
	}

	return append(diags, resourceSipRuleRead(ctx, d, meta)...)
}

func resourceSipRuleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ChimeSDKVoiceClient(ctx)

	input := &chimesdkvoice.DeleteSipRuleInput{
		SipRuleId: aws.String(d.Id()),
	}

	if _, err := conn.DeleteSipRule(ctx, input); err != nil {
		if errs.IsA[*awstypes.NotFoundException](err) {
			log.Printf("[WARN] ChimeSDKVoice Sip Rule %s not found", d.Id())
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "deleting Sip Rule (%s)", d.Id())
	}

	return diags
}

func expandSipRuleTargetApplications(data []interface{}) []awstypes.SipRuleTargetApplication {
	var targetApplications []awstypes.SipRuleTargetApplication

	for _, rItem := range data {
		item := rItem.(map[string]interface{})
		application := awstypes.SipRuleTargetApplication{
			SipMediaApplicationId: aws.String(item["sip_media_application_id"].(string)),
			Priority:              aws.Int32(int32(item[names.AttrPriority].(int))),
			AwsRegion:             aws.String(item["aws_region"].(string)),
		}

		targetApplications = append(targetApplications, application)
	}

	return targetApplications
}

func flattenSipRuleTargetApplications(apiObject []awstypes.SipRuleTargetApplication) []interface{} {
	var rawSipRuleTargetApplications []interface{}

	for _, e := range apiObject {
		rawTargetApplication := map[string]interface{}{
			"sip_media_application_id": aws.ToString(e.SipMediaApplicationId),
			names.AttrPriority:         e.Priority,
			"aws_region":               aws.ToString(e.AwsRegion),
		}

		rawSipRuleTargetApplications = append(rawSipRuleTargetApplications, rawTargetApplication)
	}
	return rawSipRuleTargetApplications
}

func findSIPRuleByID(ctx context.Context, conn *chimesdkvoice.Client, id string) (*awstypes.SipRule, error) {
	in := &chimesdkvoice.GetSipRuleInput{
		SipRuleId: aws.String(id),
	}

	resp, err := conn.GetSipRule(ctx, in)

	if errs.IsA[*awstypes.NotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}

	if resp == nil || resp.SipRule == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	if err != nil {
		return nil, err
	}

	return resp.SipRule, nil
}
