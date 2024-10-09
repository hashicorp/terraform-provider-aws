// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package waf

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/YakDriver/regexache"
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

// @SDKResource("aws_waf_rule", name="Rule")
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
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z]+$`), "must contain only alphanumeric characters"),
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
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceRuleCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFClient(ctx)

	name := d.Get(names.AttrName).(string)
	output, err := newRetryer(conn).RetryWithToken(ctx, func(token *string) (interface{}, error) {
		input := &waf.CreateRuleInput{
			ChangeToken: token,
			MetricName:  aws.String(d.Get(names.AttrMetricName).(string)),
			Name:        aws.String(name),
			Tags:        getTagsIn(ctx),
		}

		return conn.CreateRule(ctx, input)
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating WAF Rule (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.(*waf.CreateRuleOutput).Rule.RuleId))

	if newPredicates := d.Get("predicates").(*schema.Set).List(); len(newPredicates) > 0 {
		noPredicates := []interface{}{}
		if err := updateRule(ctx, conn, d.Id(), noPredicates, newPredicates); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	return append(diags, resourceRuleRead(ctx, d, meta)...)
}

func resourceRuleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFClient(ctx)

	rule, err := findRuleByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] WAF Rule (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading WAF Rule (%s): %s", d.Id(), err)
	}

	var predicates []map[string]interface{}

	for _, predicateSet := range rule.Predicates {
		predicate := map[string]interface{}{
			"data_id":      aws.ToString(predicateSet.DataId),
			"negated":      aws.ToBool(predicateSet.Negated),
			names.AttrType: predicateSet.Type,
		}
		predicates = append(predicates, predicate)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "waf",
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  "rule/" + d.Id(),
	}.String()
	d.Set(names.AttrARN, arn)
	d.Set(names.AttrMetricName, rule.MetricName)
	d.Set(names.AttrName, rule.Name)
	if err := d.Set("predicates", predicates); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting predicates: %s", err)
	}

	return diags
}

func resourceRuleUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFClient(ctx)

	if d.HasChange("predicates") {
		o, n := d.GetChange("predicates")
		oldP, newP := o.(*schema.Set).List(), n.(*schema.Set).List()
		if err := updateRule(ctx, conn, d.Id(), oldP, newP); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	return append(diags, resourceRuleRead(ctx, d, meta)...)
}

func resourceRuleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFClient(ctx)

	if oldPredicates := d.Get("predicates").(*schema.Set).List(); len(oldPredicates) > 0 {
		noPredicates := []interface{}{}
		if err := updateRule(ctx, conn, d.Id(), oldPredicates, noPredicates); err != nil && !errs.IsA[*awstypes.WAFNonexistentItemException](err) && !errs.IsA[*awstypes.WAFNonexistentContainerException](err) {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	const (
		timeout = 1 * time.Minute
	)
	_, err := tfresource.RetryWhenIsA[*awstypes.WAFReferencedItemException](ctx, timeout, func() (interface{}, error) {
		return newRetryer(conn).RetryWithToken(ctx, func(token *string) (interface{}, error) {
			input := &waf.DeleteRuleInput{
				ChangeToken: token,
				RuleId:      aws.String(d.Id()),
			}

			return conn.DeleteRule(ctx, input)
		})
	})

	if errs.IsA[*awstypes.WAFNonexistentItemException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting WAF Rule (%s): %s", d.Id(), err)
	}

	return diags
}

func findRuleByID(ctx context.Context, conn *waf.Client, id string) (*awstypes.Rule, error) {
	input := &waf.GetRuleInput{
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

func updateRule(ctx context.Context, conn *waf.Client, id string, oldP, newP []interface{}) error {
	_, err := newRetryer(conn).RetryWithToken(ctx, func(token *string) (interface{}, error) {
		input := &waf.UpdateRuleInput{
			ChangeToken: token,
			RuleId:      aws.String(id),
			Updates:     diffRulePredicates(oldP, newP),
		}

		return conn.UpdateRule(ctx, input)
	})

	if err != nil {
		return fmt.Errorf("updating WAF Rule (%s): %w", id, err)
	}

	return nil
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
