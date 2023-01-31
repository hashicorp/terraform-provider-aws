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
)

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
							Default:  wafregional.WafRuleTypeRegular,
						},
					},
				},
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
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
	conn := meta.(*conns.AWSClient).WAFRegionalConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))
	region := meta.(*conns.AWSClient).Region

	wr := NewRetryer(conn, region)
	out, err := wr.RetryWithToken(ctx, func(token *string) (interface{}, error) {
		params := &waf.CreateRuleGroupInput{
			ChangeToken: token,
			MetricName:  aws.String(d.Get("metric_name").(string)),
			Name:        aws.String(d.Get("name").(string)),
		}

		if len(tags) > 0 {
			params.Tags = Tags(tags.IgnoreAWS())
		}

		return conn.CreateRuleGroupWithContext(ctx, params)
	})
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating WAF Regional Rule Group (%s): %s", d.Get("name").(string), err)
	}
	resp := out.(*waf.CreateRuleGroupOutput)
	d.SetId(aws.StringValue(resp.RuleGroup.RuleGroupId))

	activatedRule := d.Get("activated_rule").(*schema.Set).List()
	if len(activatedRule) > 0 {
		noActivatedRules := []interface{}{}

		err := updateRuleGroupResourceWR(ctx, d.Id(), noActivatedRules, activatedRule, conn, region)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating WAF Regional Rule Group: %s", err)
		}
	}

	return append(diags, resourceRuleGroupRead(ctx, d, meta)...)
}

func resourceRuleGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFRegionalConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

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
	d.Set("arn", arn)

	tags, err := ListTags(ctx, conn, arn)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing tags for WAF Regional Rule Group (%s): %s", arn, err)
	}
	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags_all: %s", err)
	}

	d.Set("activated_rule", tfwaf.FlattenActivatedRules(rResp.ActivatedRules))
	d.Set("name", resp.RuleGroup.Name)
	d.Set("metric_name", resp.RuleGroup.MetricName)

	return diags
}

func resourceRuleGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFRegionalConn()
	region := meta.(*conns.AWSClient).Region

	if d.HasChange("activated_rule") {
		o, n := d.GetChange("activated_rule")
		oldRules, newRules := o.(*schema.Set).List(), n.(*schema.Set).List()

		err := updateRuleGroupResourceWR(ctx, d.Id(), oldRules, newRules, conn, region)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating WAF Regional Rule Group (%s): %s", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(ctx, conn, d.Get("arn").(string), o, n); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating tags: %s", err)
		}
	}

	return append(diags, resourceRuleGroupRead(ctx, d, meta)...)
}

func resourceRuleGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFRegionalConn()
	region := meta.(*conns.AWSClient).Region

	oldRules := d.Get("activated_rule").(*schema.Set).List()
	err := DeleteRuleGroup(ctx, d.Id(), oldRules, conn, region)

	return sdkdiag.AppendErrorf(diags, "deleting WAF Regional Rule Group (%s): %s", d.Id(), err)
}

func DeleteRuleGroup(ctx context.Context, id string, oldRules []interface{}, conn *wafregional.WAFRegional, region string) error {
	if len(oldRules) > 0 {
		noRules := []interface{}{}
		err := updateRuleGroupResourceWR(ctx, id, oldRules, noRules, conn, region)
		if err != nil {
			return fmt.Errorf("updating WAF Regional Rule Group Predicates: %s", err)
		}
	}

	wr := NewRetryer(conn, region)
	_, err := wr.RetryWithToken(ctx, func(token *string) (interface{}, error) {
		req := &waf.DeleteRuleGroupInput{
			ChangeToken: token,
			RuleGroupId: aws.String(id),
		}
		log.Printf("[INFO] Deleting WAF Regional Rule Group")
		return conn.DeleteRuleGroupWithContext(ctx, req)
	})
	if err != nil {
		return fmt.Errorf("deleting WAF Regional Rule Group: %s", err)
	}
	return nil
}

func updateRuleGroupResourceWR(ctx context.Context, id string, oldRules, newRules []interface{}, conn *wafregional.WAFRegional, region string) error {
	wr := NewRetryer(conn, region)
	_, err := wr.RetryWithToken(ctx, func(token *string) (interface{}, error) {
		req := &waf.UpdateRuleGroupInput{
			ChangeToken: token,
			RuleGroupId: aws.String(id),
			Updates:     tfwaf.DiffRuleGroupActivatedRules(oldRules, newRules),
		}

		return conn.UpdateRuleGroupWithContext(ctx, req)
	})
	if err != nil {
		return fmt.Errorf("updating WAF Regional Rule Group: %s", err)
	}

	return nil
}
