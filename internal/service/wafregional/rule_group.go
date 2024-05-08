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
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfwaf "github.com/hashicorp/terraform-provider-aws/internal/service/waf"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_wafregional_rule_group", name="Rule Group")
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
									names.AttrType: {
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
						names.AttrType: {
							Type:     schema.TypeString,
							Optional: true,
							Default:  wafregional.WafRuleTypeRegular,
						},
					},
				},
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

func resourceRuleGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFRegionalConn(ctx)
	region := meta.(*conns.AWSClient).Region

	name := d.Get(names.AttrName).(string)
	outputRaw, err := NewRetryer(conn, region).RetryWithToken(ctx, func(token *string) (interface{}, error) {
		input := &waf.CreateRuleGroupInput{
			ChangeToken: token,
			MetricName:  aws.String(d.Get("metric_name").(string)),
			Name:        aws.String(name),
			Tags:        getTagsIn(ctx),
		}

		return conn.CreateRuleGroupWithContext(ctx, input)
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating WAF Regional Rule Group (%s): %s", name, err)
	}

	d.SetId(aws.StringValue(outputRaw.(*waf.CreateRuleGroupOutput).RuleGroup.RuleGroupId))

	activatedRule := d.Get("activated_rule").(*schema.Set).List()
	if len(activatedRule) > 0 {
		noActivatedRules := []interface{}{}

		if err := updateRuleGroupResourceWR(ctx, conn, region, d.Id(), noActivatedRules, activatedRule); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating WAF Regional Rule Group (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceRuleGroupRead(ctx, d, meta)...)
}

func resourceRuleGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFRegionalConn(ctx)

	params := &waf.GetRuleGroupInput{
		RuleGroupId: aws.String(d.Id()),
	}

	resp, err := conn.GetRuleGroupWithContext(ctx, params)
	if err != nil {
		if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, wafregional.ErrCodeWAFNonexistentItemException) {
			log.Printf("[WARN] WAF Regional Rule Group (%s) not found, removing from state", d.Id())
			d.SetId("")
			return diags
		}

		return sdkdiag.AppendErrorf(diags, "reading WAF Regional Rule Group (%s): %s", d.Id(), err)
	}

	rResp, err := conn.ListActivatedRulesInRuleGroupWithContext(ctx, &waf.ListActivatedRulesInRuleGroupInput{
		RuleGroupId: aws.String(d.Id()),
	})
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing activated rules in WAF Regional Rule Group (%s): %s", d.Id(), err)
	}

	arn := arn.ARN{
		AccountID: meta.(*conns.AWSClient).AccountID,
		Partition: meta.(*conns.AWSClient).Partition,
		Region:    meta.(*conns.AWSClient).Region,
		Resource:  fmt.Sprintf("rulegroup/%s", d.Id()),
		Service:   "waf-regional",
	}.String()
	d.Set(names.AttrARN, arn)
	d.Set("activated_rule", tfwaf.FlattenActivatedRules(rResp.ActivatedRules))
	d.Set(names.AttrName, resp.RuleGroup.Name)
	d.Set("metric_name", resp.RuleGroup.MetricName)

	return diags
}

func resourceRuleGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFRegionalConn(ctx)
	region := meta.(*conns.AWSClient).Region

	if d.HasChange("activated_rule") {
		o, n := d.GetChange("activated_rule")
		oldRules, newRules := o.(*schema.Set).List(), n.(*schema.Set).List()

		if err := updateRuleGroupResourceWR(ctx, conn, region, d.Id(), oldRules, newRules); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating WAF Regional Rule Group (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceRuleGroupRead(ctx, d, meta)...)
}

func resourceRuleGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFRegionalConn(ctx)
	region := meta.(*conns.AWSClient).Region

	oldRules := d.Get("activated_rule").(*schema.Set).List()
	err := DeleteRuleGroup(ctx, d.Id(), oldRules, conn, region)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting WAF Regional Rule Group (%s): %s", d.Id(), err)
	}

	return diags
}

func DeleteRuleGroup(ctx context.Context, ruleGroupID string, oldRules []interface{}, conn *wafregional.WAFRegional, region string) error {
	if len(oldRules) > 0 {
		noRules := []interface{}{}

		err := updateRuleGroupResourceWR(ctx, conn, region, ruleGroupID, oldRules, noRules)

		if tfawserr.ErrCodeEquals(err, wafregional.ErrCodeWAFNonexistentContainerException, wafregional.ErrCodeWAFNonexistentItemException) {
			return nil
		}

		if err != nil {
			return fmt.Errorf("updating WAF Regional Rule Group (%s): %s", ruleGroupID, err)
		}
	}

	log.Printf("[INFO] Deleting WAF Regional Rule Group: %s", ruleGroupID)
	_, err := NewRetryer(conn, region).RetryWithToken(ctx, func(token *string) (interface{}, error) {
		input := &waf.DeleteRuleGroupInput{
			ChangeToken: token,
			RuleGroupId: aws.String(ruleGroupID),
		}

		return conn.DeleteRuleGroupWithContext(ctx, input)
	})

	if tfawserr.ErrCodeEquals(err, waf.ErrCodeNonexistentItemException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("deleting WAF Regional Rule Group (%s): %s", ruleGroupID, err)
	}

	return nil
}

func updateRuleGroupResourceWR(ctx context.Context, conn *wafregional.WAFRegional, region, ruleGroupID string, oldRules, newRules []interface{}) error {
	_, err := NewRetryer(conn, region).RetryWithToken(ctx, func(token *string) (interface{}, error) {
		input := &waf.UpdateRuleGroupInput{
			ChangeToken: token,
			RuleGroupId: aws.String(ruleGroupID),
			Updates:     tfwaf.DiffRuleGroupActivatedRules(oldRules, newRules),
		}

		return conn.UpdateRuleGroupWithContext(ctx, input)
	})

	return err
}
