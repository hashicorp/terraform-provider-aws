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
func ResourceRateBasedRule() *schema.Resource {
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
			"metric_name": {
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
		MetricName: aws.String(d.Get("metric_name").(string)),
		Name:       aws.String(name),
		RateKey:    awstypes.RateKey(d.Get("rate_key").(string)),
		RateLimit:  aws.Int64(int64(d.Get("rate_limit").(int))),
		Tags:       getTagsIn(ctx),
	}

	wr := NewRetryer(conn)
	outputRaw, err := wr.RetryWithToken(ctx, func(token *string) (interface{}, error) {
		input.ChangeToken = token

		return conn.CreateRateBasedRule(ctx, input)
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating WAF Rate Based Rule (%s): %s", name, err)
	}

	output := outputRaw.(*waf.CreateRateBasedRuleOutput)

	d.SetId(aws.ToString(output.Rule.RuleId))

	newPredicates := d.Get("predicates").(*schema.Set).List()
	if len(newPredicates) > 0 {
		err := updateRateBasedRuleResource(ctx, conn, aws.ToString(output.Rule.RuleId), []interface{}{}, newPredicates, d.Get("rate_limit"))

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating WAF Rate Based Rule (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceRateBasedRuleRead(ctx, d, meta)...)
}

func resourceRateBasedRuleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFClient(ctx)

	rule, err := FindRateBasedRuleByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] WAF Rate Based Rule (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading WAF Rate Based Rule (%s): %s", d.Id(), err)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "waf",
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("ratebasedrule/%s", d.Id()),
	}.String()
	d.Set(names.AttrARN, arn)
	d.Set(names.AttrName, rule.Name)
	d.Set("metric_name", rule.MetricName)
	d.Set("rate_key", rule.RateKey)
	d.Set("rate_limit", rule.RateLimit)

	var predicates []map[string]interface{}

	for _, predicateSet := range rule.MatchPredicates {
		predicate := map[string]interface{}{
			"negated":      *predicateSet.Negated,
			names.AttrType: predicateSet.Type,
			"data_id":      *predicateSet.DataId,
		}
		predicates = append(predicates, predicate)
	}

	if err := d.Set("predicates", predicates); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting predicates: %s", err)
	}

	return diags
}

func resourceRateBasedRuleUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFClient(ctx)

	if d.HasChanges("predicates", "rate_limit") {
		o, n := d.GetChange("predicates")
		oldP, newP := o.(*schema.Set).List(), n.(*schema.Set).List()
		rateLimit := d.Get("rate_limit")

		err := updateRateBasedRuleResource(ctx, conn, d.Id(), oldP, newP, rateLimit)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating WAF Rate Based Rule (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceRateBasedRuleRead(ctx, d, meta)...)
}

func resourceRateBasedRuleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFClient(ctx)

	oldPredicates := d.Get("predicates").(*schema.Set).List()
	if len(oldPredicates) > 0 {
		noPredicates := []interface{}{}
		rateLimit := d.Get("rate_limit")

		err := updateRateBasedRuleResource(ctx, conn, d.Id(), oldPredicates, noPredicates, rateLimit)

		if err != nil {
			if !errs.IsA[*awstypes.WAFNonexistentItemException](err) && !errs.IsA[*awstypes.WAFNonexistentContainerException](err) {
				return sdkdiag.AppendErrorf(diags, "updating WAF Rate Based Rule (%s): %s", d.Id(), err)
			}
		}
	}

	log.Printf("[INFO] Deleting WAF Rate Based Rule: %s", d.Id())
	wr := NewRetryer(conn)
	_, err := wr.RetryWithToken(ctx, func(token *string) (interface{}, error) {
		return conn.DeleteRateBasedRule(ctx, &waf.DeleteRateBasedRuleInput{
			ChangeToken: token,
			RuleId:      aws.String(d.Id()),
		})
	})
	if errs.IsA[*awstypes.WAFNonexistentItemException](err) {
		return diags
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting WAF Rate Based Rule (%s): %s", d.Id(), err)
	}

	return diags
}

func updateRateBasedRuleResource(ctx context.Context, conn *waf.Client, id string, oldP, newP []interface{}, rateLimit interface{}) error {
	input := &waf.UpdateRateBasedRuleInput{
		RateLimit: aws.Int64(int64(rateLimit.(int))),
		RuleId:    aws.String(id),
		Updates:   DiffRulePredicates(oldP, newP),
	}

	wr := NewRetryer(conn)
	_, err := wr.RetryWithToken(ctx, func(token *string) (interface{}, error) {
		input.ChangeToken = token
		return conn.UpdateRateBasedRule(ctx, input)
	})
	return err
}

func FindRateBasedRuleByID(ctx context.Context, conn *waf.Client, id string) (*awstypes.RateBasedRule, error) {
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
