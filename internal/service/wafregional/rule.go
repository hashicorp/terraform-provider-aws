// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package wafregional

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/wafregional"
	awstypes "github.com/aws/aws-sdk-go-v2/service/wafregional/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_wafregional_rule", name="Rule")
// @Tags(identifierAttribute="arn")
func resourceRule() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceRuleCreate,
		ReadWithoutTimeout:   resourceRuleRead,
		UpdateWithoutTimeout: resourceRuleUpdate,
		DeleteWithoutTimeout: resourceRuleDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrMetricName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"predicate": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"data_id": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(1, 128),
						},
						"negated": {
							Type:     schema.TypeBool,
							Required: true,
						},
						names.AttrType: {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[awstypes.PredicateType](),
						},
					},
				},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceRuleCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFRegionalClient(ctx)
	region := meta.(*conns.AWSClient).Region

	name := d.Get(names.AttrName).(string)
	outputRaw, err := newRetryer(conn, region).RetryWithToken(ctx, func(token *string) (interface{}, error) {
		input := &wafregional.CreateRuleInput{
			ChangeToken: token,
			MetricName:  aws.String(d.Get(names.AttrMetricName).(string)),
			Name:        aws.String(name),
			Tags:        getTagsIn(ctx),
		}

		return conn.CreateRule(ctx, input)
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating WAF Regional Rule (%s): %s", name, err)
	}

	d.SetId(aws.ToString(outputRaw.(*wafregional.CreateRuleOutput).Rule.RuleId))

	if newPredicates := d.Get("predicate").(*schema.Set).List(); len(newPredicates) > 0 {
		var oldPredicates []interface{}
		if err := updateRule(ctx, conn, region, d.Id(), oldPredicates, newPredicates); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	return append(diags, resourceRuleRead(ctx, d, meta)...)
}

func resourceRuleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFRegionalClient(ctx)

	rule, err := findRuleByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] WAF Regional Rule (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading WAF Regional Rule (%s): %s", d.Id(), err)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "waf-regional",
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  "rule/" + d.Id(),
	}.String()
	d.Set(names.AttrARN, arn)
	d.Set(names.AttrMetricName, rule.MetricName)
	d.Set(names.AttrName, rule.Name)
	if err := d.Set("predicate", flattenPredicates(rule.Predicates)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting predicate: %s", err)
	}

	return diags
}

func resourceRuleUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFRegionalClient(ctx)
	region := meta.(*conns.AWSClient).Region

	if d.HasChange("predicate") {
		o, n := d.GetChange("predicate")
		oldPredicates, newPredicates := o.(*schema.Set).List(), n.(*schema.Set).List()
		if err := updateRule(ctx, conn, region, d.Id(), oldPredicates, newPredicates); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	return append(diags, resourceRuleRead(ctx, d, meta)...)
}

func resourceRuleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFRegionalClient(ctx)
	region := meta.(*conns.AWSClient).Region

	if oldPredicates := d.Get("predicate").(*schema.Set).List(); len(oldPredicates) > 0 {
		var newPredicates []interface{}
		if err := updateRule(ctx, conn, region, d.Id(), oldPredicates, newPredicates); err != nil && !errs.IsA[*awstypes.WAFNonexistentItemException](err) && !errs.IsA[*awstypes.WAFNonexistentContainerException](err) {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	log.Printf("[INFO] Deleting WAF Regional Rule: %s", d.Id())
	_, err := newRetryer(conn, region).RetryWithToken(ctx, func(token *string) (interface{}, error) {
		input := &wafregional.DeleteRuleInput{
			ChangeToken: token,
			RuleId:      aws.String(d.Id()),
		}

		return conn.DeleteRule(ctx, input)
	})

	if errs.IsA[*awstypes.WAFNonexistentItemException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting WAF Regional Rule (%s): %s", d.Id(), err)
	}

	return diags
}

func findRuleByID(ctx context.Context, conn *wafregional.Client, id string) (*awstypes.Rule, error) {
	input := &wafregional.GetRuleInput{
		RuleId: aws.String(id),
	}

	output, err := conn.GetRule(ctx, input)

	if errs.IsA[*awstypes.WAFNonexistentItemException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Rule == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Rule, nil
}

func updateRule(ctx context.Context, conn *wafregional.Client, region, ruleID string, oldP, newP []interface{}) error {
	_, err := newRetryer(conn, region).RetryWithToken(ctx, func(token *string) (interface{}, error) {
		input := &wafregional.UpdateRuleInput{
			ChangeToken: token,
			RuleId:      aws.String(ruleID),
			Updates:     diffRulePredicates(oldP, newP),
		}

		return conn.UpdateRule(ctx, input)
	})

	if err != nil {
		return fmt.Errorf("updating WAF Regional Rule (%s): %w", ruleID, err)
	}

	return nil
}

func flattenPredicates(ts []awstypes.Predicate) []interface{} {
	out := make([]interface{}, len(ts))
	for i, p := range ts {
		m := make(map[string]interface{})
		m["negated"] = aws.ToBool(p.Negated)
		m[names.AttrType] = p.Type
		m["data_id"] = aws.ToString(p.DataId)
		out[i] = m
	}
	return out
}

func diffRulePredicates(oldP, newP []interface{}) []awstypes.RuleUpdate {
	updates := make([]awstypes.RuleUpdate, 0)

	for _, op := range oldP {
		predicate := op.(map[string]interface{})

		if idx, contains := sliceContainsMap(newP, predicate); contains {
			newP = append(newP[:idx], newP[idx+1:]...)
			continue
		}

		updates = append(updates, awstypes.RuleUpdate{
			Action: awstypes.ChangeActionDelete,
			Predicate: &awstypes.Predicate{
				Negated: aws.Bool(predicate["negated"].(bool)),
				Type:    awstypes.PredicateType(predicate[names.AttrType].(string)),
				DataId:  aws.String(predicate["data_id"].(string)),
			},
		})
	}

	for _, np := range newP {
		predicate := np.(map[string]interface{})

		updates = append(updates, awstypes.RuleUpdate{
			Action: awstypes.ChangeActionInsert,
			Predicate: &awstypes.Predicate{
				Negated: aws.Bool(predicate["negated"].(bool)),
				Type:    awstypes.PredicateType(predicate[names.AttrType].(string)),
				DataId:  aws.String(predicate["data_id"].(string)),
			},
		})
	}
	return updates
}
