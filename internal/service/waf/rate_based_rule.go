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

// @SDKResource("aws_waf_rate_based_rule", name="Rate Based Rule")
// @Tags(identifierAttribute="arn")
func resourceRateBasedRule() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceRateBasedRuleCreate,
		ReadWithoutTimeout:   resourceRateBasedRuleRead,
		UpdateWithoutTimeout: resourceRateBasedRuleUpdate,
		DeleteWithoutTimeout: resourceRateBasedRuleDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
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
			"predicates": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"data_id": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(0, 128),
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
			"rate_key": {
				Type:     schema.TypeString,
				Required: true,
			},
			"rate_limit": {
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntAtLeast(100),
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceRateBasedRuleCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &waf.CreateRateBasedRuleInput{
		MetricName: aws.String(d.Get(names.AttrMetricName).(string)),
		Name:       aws.String(name),
		RateKey:    awstypes.RateKey(d.Get("rate_key").(string)),
		RateLimit:  aws.Int64(int64(d.Get("rate_limit").(int))),
		Tags:       getTagsIn(ctx),
	}

	output, err := newRetryer(conn).RetryWithToken(ctx, func(token *string) (interface{}, error) {
		input.ChangeToken = token

		return conn.CreateRateBasedRule(ctx, input)
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating WAF Rate Based Rule (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.(*waf.CreateRateBasedRuleOutput).Rule.RuleId))

	if newPredicates := d.Get("predicates").(*schema.Set).List(); len(newPredicates) > 0 {
		if err := updateRateBasedRule(ctx, conn, d.Id(), []interface{}{}, newPredicates, d.Get("rate_limit")); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	return append(diags, resourceRateBasedRuleRead(ctx, d, meta)...)
}

func resourceRateBasedRuleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFClient(ctx)

	rule, err := findRateBasedRuleByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] WAF Rate Based Rule (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading WAF Rate Based Rule (%s): %s", d.Id(), err)
	}

	var predicates []map[string]interface{}

	for _, predicateSet := range rule.MatchPredicates {
		predicates = append(predicates, map[string]interface{}{
			"data_id":      aws.ToString(predicateSet.DataId),
			"negated":      aws.ToBool(predicateSet.Negated),
			names.AttrType: predicateSet.Type,
		})
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "waf",
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  "ratebasedrule/" + d.Id(),
	}.String()
	d.Set(names.AttrARN, arn)
	d.Set(names.AttrMetricName, rule.MetricName)
	d.Set(names.AttrName, rule.Name)
	if err := d.Set("predicates", predicates); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting predicates: %s", err)
	}
	d.Set("rate_key", rule.RateKey)
	d.Set("rate_limit", rule.RateLimit)

	return diags
}

func resourceRateBasedRuleUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFClient(ctx)

	if d.HasChanges("predicates", "rate_limit") {
		o, n := d.GetChange("predicates")
		oldP, newP := o.(*schema.Set).List(), n.(*schema.Set).List()
		if err := updateRateBasedRule(ctx, conn, d.Id(), oldP, newP, d.Get("rate_limit")); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	return append(diags, resourceRateBasedRuleRead(ctx, d, meta)...)
}

func resourceRateBasedRuleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFClient(ctx)

	if oldPredicates := d.Get("predicates").(*schema.Set).List(); len(oldPredicates) > 0 {
		noPredicates := []interface{}{}
		rateLimit := d.Get("rate_limit")
		if err := updateRateBasedRule(ctx, conn, d.Id(), oldPredicates, noPredicates, rateLimit); err != nil && !errs.IsA[*awstypes.WAFNonexistentItemException](err) && !errs.IsA[*awstypes.WAFNonexistentContainerException](err) {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	log.Printf("[INFO] Deleting WAF Rate Based Rule: %s", d.Id())
	_, err := newRetryer(conn).RetryWithToken(ctx, func(token *string) (interface{}, error) {
		input := &waf.DeleteRateBasedRuleInput{
			ChangeToken: token,
			RuleId:      aws.String(d.Id()),
		}

		return conn.DeleteRateBasedRule(ctx, input)
	})

	if errs.IsA[*awstypes.WAFNonexistentItemException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting WAF Rate Based Rule (%s): %s", d.Id(), err)
	}

	return diags
}

func findRateBasedRuleByID(ctx context.Context, conn *waf.Client, id string) (*awstypes.RateBasedRule, error) {
	input := &waf.GetRateBasedRuleInput{
		RuleId: aws.String(id),
	}

	output, err := conn.GetRateBasedRule(ctx, input)

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

func updateRateBasedRule(ctx context.Context, conn *waf.Client, id string, oldP, newP []interface{}, rateLimit interface{}) error {
	input := &waf.UpdateRateBasedRuleInput{
		RateLimit: aws.Int64(int64(rateLimit.(int))),
		RuleId:    aws.String(id),
		Updates:   diffRulePredicates(oldP, newP),
	}

	_, err := newRetryer(conn).RetryWithToken(ctx, func(token *string) (interface{}, error) {
		input.ChangeToken = token

		return conn.UpdateRateBasedRule(ctx, input)
	})

	if err != nil {
		return fmt.Errorf("updating WAF Rate Based Rule (%s): %w", id, err)
	}

	return nil
}
