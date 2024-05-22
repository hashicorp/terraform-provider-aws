// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package waf

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/waf"
	awstypes "github.com/aws/aws-sdk-go-v2/service/waf/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_waf_rule_group", name="Rule Group")
// @Tags(identifierAttribute="arn")
func resourceRuleGroup() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceRuleGroupCreate,
		ReadWithoutTimeout:   resourceRuleGroupRead,
		UpdateWithoutTimeout: resourceRuleGroupUpdate,
		DeleteWithoutTimeout: resourceRuleGroupDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"activated_rule": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrAction: {
							Type:     schema.TypeList,
							MaxItems: 1,
							Required: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrType: {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},
						names.AttrPriority: {
							Type:     schema.TypeInt,
							Required: true,
						},
						"rule_id": {
							Type:     schema.TypeString,
							Required: true,
						},
						names.AttrType: {
							Type:     schema.TypeString,
							Optional: true,
							Default:  awstypes.WafRuleTypeRegular,
						},
					},
				},
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrMetricName: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validMetricName,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceRuleGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFClient(ctx)

	name := d.Get(names.AttrName).(string)
	output, err := newRetryer(conn).RetryWithToken(ctx, func(token *string) (interface{}, error) {
		input := &waf.CreateRuleGroupInput{
			ChangeToken: token,
			MetricName:  aws.String(d.Get(names.AttrMetricName).(string)),
			Name:        aws.String(name),
			Tags:        getTagsIn(ctx),
		}

		return conn.CreateRuleGroup(ctx, input)
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating WAF Rule Group (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.(*waf.CreateRuleGroupOutput).RuleGroup.RuleGroupId))

	if activatedRules := d.Get("activated_rule").(*schema.Set).List(); len(activatedRules) > 0 {
		noActivatedRules := []interface{}{}
		if err := updateRuleGroup(ctx, conn, d.Id(), noActivatedRules, activatedRules); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	return append(diags, resourceRuleGroupRead(ctx, d, meta)...)
}

func resourceRuleGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFClient(ctx)

	ruleGroup, err := findRuleGroupByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] WAF Rule Group (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading WAF Rule Group (%s): %s", d.Id(), err)
	}

	var activatedRules []awstypes.ActivatedRule
	input := &waf.ListActivatedRulesInRuleGroupInput{
		RuleGroupId: aws.String(d.Id()),
	}

	err = listActivatedRulesInRuleGroupPages(ctx, conn, input, func(page *waf.ListActivatedRulesInRuleGroupOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		activatedRules = append(activatedRules, page.ActivatedRules...)

		return !lastPage
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing WAF Rule Group (%s) activated rules: %s", d.Id(), err)
	}

	if err := d.Set("activated_rule", flattenActivatedRules(activatedRules)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting activated_rule: %s", err)
	}
	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "waf",
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  "rulegroup/" + d.Id(),
	}.String()
	d.Set(names.AttrARN, arn)
	d.Set(names.AttrMetricName, ruleGroup.MetricName)
	d.Set(names.AttrName, ruleGroup.Name)

	return diags
}

func resourceRuleGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFClient(ctx)

	if d.HasChange("activated_rule") {
		o, n := d.GetChange("activated_rule")
		oldRules, newRules := o.(*schema.Set).List(), n.(*schema.Set).List()
		if err := updateRuleGroup(ctx, conn, d.Id(), oldRules, newRules); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	return append(diags, resourceRuleGroupRead(ctx, d, meta)...)
}

func resourceRuleGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFClient(ctx)

	if oldRules := d.Get("activated_rule").(*schema.Set).List(); len(oldRules) > 0 {
		noRules := []interface{}{}
		if err := updateRuleGroup(ctx, conn, d.Id(), oldRules, noRules); err != nil && !errs.IsA[*awstypes.WAFNonexistentItemException](err) && !errs.IsA[*awstypes.WAFNonexistentContainerException](err) {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	log.Printf("[INFO] Deleting WAF Rule Group: %s", d.Id())
	_, err := newRetryer(conn).RetryWithToken(ctx, func(token *string) (interface{}, error) {
		input := &waf.DeleteRuleGroupInput{
			ChangeToken: token,
			RuleGroupId: aws.String(d.Id()),
		}

		return conn.DeleteRuleGroup(ctx, input)
	})

	if errs.IsA[*awstypes.WAFNonexistentItemException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting WAF Rule Group (%s): %s", d.Id(), err)
	}

	return diags
}

func findRuleGroupByID(ctx context.Context, conn *waf.Client, id string) (*awstypes.RuleGroup, error) {
	input := &waf.GetRuleGroupInput{
		RuleGroupId: aws.String(id),
	}

	output, err := conn.GetRuleGroup(ctx, input)

	if errs.IsA[*awstypes.WAFNonexistentItemException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.RuleGroup == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.RuleGroup, nil
}

func updateRuleGroup(ctx context.Context, conn *waf.Client, id string, oldRules, newRules []interface{}) error {
	_, err := newRetryer(conn).RetryWithToken(ctx, func(token *string) (interface{}, error) {
		input := &waf.UpdateRuleGroupInput{
			ChangeToken: token,
			RuleGroupId: aws.String(id),
			Updates:     diffRuleGroupActivatedRules(oldRules, newRules),
		}

		return conn.UpdateRuleGroup(ctx, input)
	})

	if err != nil {
		return fmt.Errorf("updating WAF Rule Group (%s): %w", id, err)
	}

	return nil
}

func diffRuleGroupActivatedRules(oldRules, newRules []interface{}) []awstypes.RuleGroupUpdate {
	updates := make([]awstypes.RuleGroupUpdate, 0)

	for _, op := range oldRules {
		rule := op.(map[string]interface{})

		if idx, contains := sliceContainsMap(newRules, rule); contains {
			newRules = append(newRules[:idx], newRules[idx+1:]...)
			continue
		}

		updates = append(updates, awstypes.RuleGroupUpdate{
			Action:        awstypes.ChangeActionDelete,
			ActivatedRule: expandActivatedRule(rule),
		})
	}

	for _, np := range newRules {
		rule := np.(map[string]interface{})

		updates = append(updates, awstypes.RuleGroupUpdate{
			Action:        awstypes.ChangeActionInsert,
			ActivatedRule: expandActivatedRule(rule),
		})
	}
	return updates
}

func flattenActivatedRules(activatedRules []awstypes.ActivatedRule) []interface{} {
	out := make([]interface{}, len(activatedRules))
	for i, ar := range activatedRules {
		rule := map[string]interface{}{
			names.AttrPriority: aws.ToInt32(ar.Priority),
			"rule_id":          aws.ToString(ar.RuleId),
			names.AttrType:     ar.Type,
		}
		if ar.Action != nil {
			rule[names.AttrAction] = []interface{}{
				map[string]interface{}{
					names.AttrType: string(ar.Action.Type),
				},
			}
		}
		out[i] = rule
	}
	return out
}

func expandActivatedRule(rule map[string]interface{}) *awstypes.ActivatedRule {
	r := &awstypes.ActivatedRule{
		Priority: aws.Int32(int32(rule[names.AttrPriority].(int))),
		RuleId:   aws.String(rule["rule_id"].(string)),
		Type:     awstypes.WafRuleType(rule[names.AttrType].(string)),
	}

	if a, ok := rule[names.AttrAction].([]interface{}); ok && len(a) > 0 {
		m := a[0].(map[string]interface{})
		r.Action = &awstypes.WafAction{
			Type: awstypes.WafActionType(m[names.AttrType].(string)),
		}
	}
	return r
}
