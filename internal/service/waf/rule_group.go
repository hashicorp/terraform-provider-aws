// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package waf

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/waf"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_waf_rule_group", name="Rule Group")
// @Tags(identifierAttribute="arn")
func ResourceRuleGroup() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceRuleGroupCreate,
		ReadWithoutTimeout:   resourceRuleGroupRead,
		UpdateWithoutTimeout: resourceRuleGroupUpdate,
		DeleteWithoutTimeout: resourceRuleGroupDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
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
			"activated_rule": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"action": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Required: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"type": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},
						"priority": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"rule_id": {
							Type:     schema.TypeString,
							Required: true,
						},
						"type": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  waf.WafRuleTypeRegular,
						},
					},
				},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceRuleGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFConn(ctx)

	wr := NewRetryer(conn)
	out, err := wr.RetryWithToken(ctx, func(token *string) (interface{}, error) {
		input := &waf.CreateRuleGroupInput{
			ChangeToken: token,
			MetricName:  aws.String(d.Get("metric_name").(string)),
			Name:        aws.String(d.Get("name").(string)),
			Tags:        getTagsIn(ctx),
		}

		return conn.CreateRuleGroupWithContext(ctx, input)
	})
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating WAF Rule Group (%s): %s", d.Get("name").(string), err)
	}
	resp := out.(*waf.CreateRuleGroupOutput)
	d.SetId(aws.StringValue(resp.RuleGroup.RuleGroupId))

	activatedRules := d.Get("activated_rule").(*schema.Set).List()
	if len(activatedRules) > 0 {
		noActivatedRules := []interface{}{}

		err := updateRuleGroupResource(ctx, d.Id(), noActivatedRules, activatedRules, conn)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating WAF Rule Group: %s", err)
		}
	}

	return append(diags, resourceRuleGroupRead(ctx, d, meta)...)
}

func resourceRuleGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFConn(ctx)

	params := &waf.GetRuleGroupInput{
		RuleGroupId: aws.String(d.Id()),
	}

	resp, err := conn.GetRuleGroupWithContext(ctx, params)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, waf.ErrCodeNonexistentItemException) {
			log.Printf("[WARN] WAF Rule Group (%s) not found, removing from state", d.Id())
			d.SetId("")
			return diags
		}

		return sdkdiag.AppendErrorf(diags, "reading WAF Rule Group (%s): %s", d.Id(), err)
	}

	rResp, err := conn.ListActivatedRulesInRuleGroupWithContext(ctx, &waf.ListActivatedRulesInRuleGroupInput{
		RuleGroupId: aws.String(d.Id()),
	})
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing activated rules in WAF Rule Group (%s): %s", d.Id(), err)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "waf",
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("rulegroup/%s", d.Id()),
	}.String()
	d.Set("arn", arn)
	d.Set("activated_rule", FlattenActivatedRules(rResp.ActivatedRules))
	d.Set("name", resp.RuleGroup.Name)
	d.Set("metric_name", resp.RuleGroup.MetricName)

	return diags
}

func resourceRuleGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFConn(ctx)

	if d.HasChange("activated_rule") {
		o, n := d.GetChange("activated_rule")
		oldRules, newRules := o.(*schema.Set).List(), n.(*schema.Set).List()

		err := updateRuleGroupResource(ctx, d.Id(), oldRules, newRules, conn)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating WAF Rule Group (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceRuleGroupRead(ctx, d, meta)...)
}

func resourceRuleGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFConn(ctx)

	oldRules := d.Get("activated_rule").(*schema.Set).List()
	if err := deleteRuleGroup(ctx, d.Id(), oldRules, conn); err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting WAF Rule Group (%s): %s", d.Id(), err)
	}
	return diags
}

func deleteRuleGroup(ctx context.Context, id string, oldRules []interface{}, conn *waf.WAF) error {
	if len(oldRules) > 0 {
		noRules := []interface{}{}
		err := updateRuleGroupResource(ctx, id, oldRules, noRules, conn)
		if err != nil {
			return fmt.Errorf("updating WAF Rule Group Predicates: %s", err)
		}
	}

	wr := NewRetryer(conn)
	_, err := wr.RetryWithToken(ctx, func(token *string) (interface{}, error) {
		req := &waf.DeleteRuleGroupInput{
			ChangeToken: token,
			RuleGroupId: aws.String(id),
		}
		log.Printf("[INFO] Deleting WAF Rule Group")
		return conn.DeleteRuleGroupWithContext(ctx, req)
	})
	if err != nil {
		return fmt.Errorf("deleting WAF Rule Group: %s", err)
	}
	return nil
}

func updateRuleGroupResource(ctx context.Context, id string, oldRules, newRules []interface{}, conn *waf.WAF) error {
	wr := NewRetryer(conn)
	_, err := wr.RetryWithToken(ctx, func(token *string) (interface{}, error) {
		req := &waf.UpdateRuleGroupInput{
			ChangeToken: token,
			RuleGroupId: aws.String(id),
			Updates:     DiffRuleGroupActivatedRules(oldRules, newRules),
		}

		return conn.UpdateRuleGroupWithContext(ctx, req)
	})
	if err != nil {
		return fmt.Errorf("Updating WAF Rule Group: %s", err)
	}

	return nil
}
