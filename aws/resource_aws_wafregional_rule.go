package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/service/wafregional"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/waf"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
)

func resourceAwsWafRegionalRule() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsWafRegionalRuleCreate,
		Read:   resourceAwsWafRegionalRuleRead,
		Update: resourceAwsWafRegionalRuleUpdate,
		Delete: resourceAwsWafRegionalRuleDelete,

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"metric_name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"predicate": &schema.Schema{
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"negated": &schema.Schema{
							Type:     schema.TypeBool,
							Required: true,
						},
						"data_id": &schema.Schema{
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(1, 128),
						},
						"type": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								"IPMatch",
								"ByteMatch",
								"SqlInjectionMatch",
								"SizeConstraint",
								"XssMatch",
							}, false),
						},
					},
				},
			},
		},
	}
}

func resourceAwsWafRegionalRuleCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).wafregionalconn
	region := meta.(*AWSClient).region

	wr := newWafRegionalRetryer(conn, region)
	out, err := wr.RetryWithToken(func(token *string) (interface{}, error) {
		params := &waf.CreateRuleInput{
			ChangeToken: token,
			MetricName:  aws.String(d.Get("metric_name").(string)),
			Name:        aws.String(d.Get("name").(string)),
		}

		return conn.CreateRule(params)
	})
	if err != nil {
		return err
	}
	resp := out.(*waf.CreateRuleOutput)
	d.SetId(*resp.Rule.RuleId)
	return resourceAwsWafRegionalRuleUpdate(d, meta)
}

func resourceAwsWafRegionalRuleRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).wafregionalconn

	params := &waf.GetRuleInput{
		RuleId: aws.String(d.Id()),
	}

	resp, err := conn.GetRule(params)
	if err != nil {
		if isAWSErr(err, wafregional.ErrCodeWAFNonexistentItemException, "") {
			log.Printf("[WARN] WAF Rule (%s) not found, error code (404)", d.Id())
			d.SetId("")
			return nil
		}

		return err
	}

	d.Set("predicate", flattenWafPredicates(resp.Rule.Predicates))
	d.Set("name", resp.Rule.Name)
	d.Set("metric_name", resp.Rule.MetricName)

	return nil
}

func resourceAwsWafRegionalRuleUpdate(d *schema.ResourceData, meta interface{}) error {
	if d.HasChange("predicate") {
		err := updateWafRegionalRuleResource(d, meta, waf.ChangeActionInsert)
		if err != nil {
			return fmt.Errorf("Error Updating WAF Rule: %s", err)
		}
	}
	return resourceAwsWafRegionalRuleRead(d, meta)
}

func resourceAwsWafRegionalRuleDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).wafregionalconn
	region := meta.(*AWSClient).region

	if v, ok := d.GetOk("predicate"); ok {
		predicates := v.(*schema.Set).List()
		if len(predicates) > 0 {
			err := updateWafRegionalRuleResource(d, meta, waf.ChangeActionDelete)
			if err != nil {
				return fmt.Errorf("Error Removing WAF Rule Predicates: %s", err)
			}
		}
	}

	wr := newWafRegionalRetryer(conn, region)
	_, err := wr.RetryWithToken(func(token *string) (interface{}, error) {
		req := &waf.DeleteRuleInput{
			ChangeToken: token,
			RuleId:      aws.String(d.Id()),
		}
		log.Printf("[INFO] Deleting WAF Rule")
		return conn.DeleteRule(req)
	})
	if err != nil {
		return fmt.Errorf("Error deleting WAF Rule: %s", err)
	}

	return nil
}

func updateWafRegionalRuleResource(d *schema.ResourceData, meta interface{}, ChangeAction string) error {
	conn := meta.(*AWSClient).wafregionalconn
	region := meta.(*AWSClient).region

	wr := newWafRegionalRetryer(conn, region)
	_, err := wr.RetryWithToken(func(token *string) (interface{}, error) {
		req := &waf.UpdateRuleInput{
			ChangeToken: token,
			RuleId:      aws.String(d.Id()),
		}

		predicatesSet := d.Get("predicate").(*schema.Set)
		for _, predicateI := range predicatesSet.List() {
			predicate := predicateI.(map[string]interface{})
			updatePredicate := &waf.RuleUpdate{
				Action: aws.String(ChangeAction),
				Predicate: &waf.Predicate{
					Negated: aws.Bool(predicate["negated"].(bool)),
					Type:    aws.String(predicate["type"].(string)),
					DataId:  aws.String(predicate["data_id"].(string)),
				},
			}
			req.Updates = append(req.Updates, updatePredicate)
		}

		return conn.UpdateRule(req)
	})
	if err != nil {
		return fmt.Errorf("Error Updating WAF Rule: %s", err)
	}

	return nil
}

func flattenWafPredicates(ts []*waf.Predicate) []interface{} {
	out := make([]interface{}, len(ts), len(ts))
	for i, p := range ts {
		m := make(map[string]interface{})
		m["negated"] = *p.Negated
		m["type"] = *p.Type
		m["data_id"] = *p.DataId
		out[i] = m
	}
	return out
}
