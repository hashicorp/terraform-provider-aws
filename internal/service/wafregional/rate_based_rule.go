// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package wafregional

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/waf"
	"github.com/aws/aws-sdk-go/service/wafregional"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfwaf "github.com/hashicorp/terraform-provider-aws/internal/service/waf"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_wafregional_rate_based_rule", name="Rate Based Rule")
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
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"metric_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validMetricName,
			},
			"predicate": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"negated": {
							Type:     schema.TypeBool,
							Required: true,
						},
						"data_id": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(0, 128),
						},
						names.AttrType: {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(wafregional.PredicateType_Values(), false),
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
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceRateBasedRuleCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFRegionalConn(ctx)
	region := meta.(*conns.AWSClient).Region

	name := d.Get(names.AttrName).(string)
	outputRaw, err := NewRetryer(conn, region).RetryWithToken(ctx, func(token *string) (interface{}, error) {
		input := &waf.CreateRateBasedRuleInput{
			ChangeToken: token,
			MetricName:  aws.String(d.Get("metric_name").(string)),
			Name:        aws.String(name),
			RateKey:     aws.String(d.Get("rate_key").(string)),
			RateLimit:   aws.Int64(int64(d.Get("rate_limit").(int))),
			Tags:        getTagsIn(ctx),
		}

		return conn.CreateRateBasedRuleWithContext(ctx, input)
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating WAF Regional Rate Based Rule (%s): %s", name, err)
	}

	d.SetId(aws.StringValue(outputRaw.(*waf.CreateRateBasedRuleOutput).Rule.RuleId))

	if newPredicates := d.Get("predicate").(*schema.Set).List(); len(newPredicates) > 0 {
		var oldPredicates []interface{}

		if err := updateRateBasedRuleResourceWR(ctx, conn, region, d.Id(), d.Get("rate_limit").(int), oldPredicates, newPredicates); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating WAF Regional Rate Based Rule (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceRateBasedRuleRead(ctx, d, meta)...)
}

func resourceRateBasedRuleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFRegionalConn(ctx)

	params := &waf.GetRateBasedRuleInput{
		RuleId: aws.String(d.Id()),
	}

	resp, err := conn.GetRateBasedRuleWithContext(ctx, params)
	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, wafregional.ErrCodeWAFNonexistentItemException) {
		log.Printf("[WARN] WAF Regional Rate Based Rule (%s) not found, removing from state", d.Get(names.AttrName).(string))
		d.SetId("")
		return diags
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "getting WAF Regional Rate Based Rule (%s): %s", d.Get(names.AttrName).(string), err)
	}

	var predicates []map[string]interface{}

	for _, predicateSet := range resp.Rule.MatchPredicates {
		predicates = append(predicates, map[string]interface{}{
			"negated":      *predicateSet.Negated,
			names.AttrType: *predicateSet.Type,
			"data_id":      *predicateSet.DataId,
		})
	}

	arn := arn.ARN{
		AccountID: meta.(*conns.AWSClient).AccountID,
		Partition: meta.(*conns.AWSClient).Partition,
		Region:    meta.(*conns.AWSClient).Region,
		Resource:  fmt.Sprintf("ratebasedrule/%s", d.Id()),
		Service:   "waf-regional",
	}.String()
	d.Set(names.AttrARN, arn)
	d.Set("predicate", predicates)
	d.Set(names.AttrName, resp.Rule.Name)
	d.Set("metric_name", resp.Rule.MetricName)
	d.Set("rate_key", resp.Rule.RateKey)
	d.Set("rate_limit", resp.Rule.RateLimit)

	return diags
}

func resourceRateBasedRuleUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFRegionalConn(ctx)
	region := meta.(*conns.AWSClient).Region

	if d.HasChanges("predicate", "rate_limit") {
		o, n := d.GetChange("predicate")
		oldPredicates, newPredicates := o.(*schema.Set).List(), n.(*schema.Set).List()

		if err := updateRateBasedRuleResourceWR(ctx, conn, region, d.Id(), d.Get("rate_limit").(int), oldPredicates, newPredicates); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating WAF Regional Rate Based Rule (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceRateBasedRuleRead(ctx, d, meta)...)
}

func resourceRateBasedRuleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFRegionalConn(ctx)
	region := meta.(*conns.AWSClient).Region

	if oldPredicates := d.Get("predicate").(*schema.Set).List(); len(oldPredicates) > 0 {
		var newPredicates []interface{}

		err := updateRateBasedRuleResourceWR(ctx, conn, region, d.Id(), d.Get("rate_limit").(int), oldPredicates, newPredicates)

		if tfawserr.ErrCodeEquals(err, wafregional.ErrCodeWAFNonexistentContainerException, wafregional.ErrCodeWAFNonexistentItemException) {
			return diags
		}

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating WAF Regional Rate Based Rule (%s): %s", d.Id(), err)
		}
	}

	_, err := NewRetryer(conn, region).RetryWithToken(ctx, func(token *string) (interface{}, error) {
		req := &waf.DeleteRateBasedRuleInput{
			ChangeToken: token,
			RuleId:      aws.String(d.Id()),
		}
		log.Printf("[INFO] Deleting WAF Regional Rate Based Rule")
		return conn.DeleteRateBasedRuleWithContext(ctx, req)
	})

	if tfawserr.ErrCodeEquals(err, wafregional.ErrCodeWAFNonexistentItemException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting WAF Regional Rate Based Rule (%s): %s", d.Id(), err)
	}

	return diags
}

func updateRateBasedRuleResourceWR(ctx context.Context, conn *wafregional.WAFRegional, region, ruleID string, rateLimit int, oldP, newP []interface{}) error {
	_, err := NewRetryer(conn, region).RetryWithToken(ctx, func(token *string) (interface{}, error) {
		input := &waf.UpdateRateBasedRuleInput{
			ChangeToken: token,
			RateLimit:   aws.Int64(int64(rateLimit)),
			RuleId:      aws.String(ruleID),
			Updates:     tfwaf.DiffRulePredicates(oldP, newP),
		}

		return conn.UpdateRateBasedRuleWithContext(ctx, input)
	})

	return err
}
