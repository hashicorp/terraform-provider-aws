package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53resolver"
)

const (
	Route53ResolverRuleStatusCreating = "CREATING"
	Route53ResolverRuleStatusDeleted  = "DELETED"
)

func resourceAwsRoute53ResolverRule() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsRoute53ResolverRuleCreate,
		Read:   resourceAwsRoute53ResolverRuleRead,
		Update: resourceAwsRoute53ResolverRuleUpdate,
		Delete: resourceAwsRoute53ResolverRuleDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"domain_name": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				DiffSuppressFunc: suppressRoute53ZoneNameWithTrailingDot,
				ValidateFunc:     validation.StringLenBetween(1, 256),
			},

			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 64),
			},

			"resolver_endpoint_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 64),
			},

			"rule_type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					route53resolver.RuleTypeOptionForward,
					route53resolver.RuleTypeOptionSystem,
					route53resolver.RuleTypeOptionRecursive,
				}, false),
			},

			"target_ip": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ip": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.SingleIP(),
						},
						"port": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      53,
							ValidateFunc: validation.IntBetween(0, 65535),
						},
					},
				},
			},

			"tags": tagsSchema(),
		},
	}
}

func resourceAwsRoute53ResolverRuleCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).r53resolverconn

	req := &route53resolver.CreateResolverRuleInput{
		CreatorRequestId: aws.String(resource.UniqueId()),
		DomainName:       aws.String(d.Get("domain_name").(string)),
		RuleType:         aws.String(d.Get("rule_type").(string)),
	}

	if v, ok := d.GetOk("name"); ok {
		req.Name = aws.String(v.(string))
	}

	if v, ok := d.GetOk("resolver_endpoint_id"); ok {
		req.ResolverEndpointId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("target_ip"); ok {
		req.TargetIps = expandTargetIps(v.([]interface{}))
	}

	if v, ok := d.GetOk("tags"); ok {
		req.Tags = tagsFromMapRoute53Resolver(v.(map[string]interface{}))
	}

	log.Printf("[DEBUG] Creating Route 53 Resolver rule: %s", req)

	res, err := conn.CreateResolverRule(req)
	if err != nil {
		return fmt.Errorf("Error creating Route 53 Resolver rule: %s", err)
	}

	d.SetId(*res.ResolverRule.Id)

	stateConf := &resource.StateChangeConf{
		Pending: []string{Route53ResolverRuleStatusCreating},
		Target:  []string{route53resolver.ResolverRuleStatusComplete},
		Refresh: resourceAwsRoute53ResolverRuleStateRefresh(conn, d.Id()),
		Timeout: d.Timeout(schema.TimeoutCreate),
	}
	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("Error waiting for Route 53 Resolver rule (%s) to be created: %s", d.Id(), err)
	}

	return resourceAwsRoute53ResolverRuleRead(d, meta)
}

func resourceAwsRoute53ResolverRuleRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).r53resolverconn

	req := &route53resolver.GetResolverRuleInput{
		ResolverRuleId: aws.String(d.Id()),
	}

	res, err := conn.GetResolverRule(req)
	if err != nil {
		if isAWSErr(err, route53resolver.ErrCodeResourceNotFoundException, "") {
			log.Printf("[WARN] No Route 53 Resolver rule by Id (%s) found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error reading Route 53 Resolver rule %s: %s", d.Id(), err)
	}

	rule := res.ResolverRule

	d.Set("arn", rule.Arn)
	d.Set("domain_name", rule.DomainName)
	d.Set("name", rule.Name)
	d.Set("resolver_endpoint_id", rule.ResolverEndpointId)
	d.Set("rule_type", rule.RuleType)
	d.Set("target_ip", flattenTargetIps(rule.TargetIps))

	err = getTagsRoute53Resolver(conn, d, d.Get("arn").(string))
	if err != nil {
		return fmt.Errorf("Error reading Route 53 Resolver rule tags %s: %s", d.Id(), err)
	}

	return nil
}

func resourceAwsRoute53ResolverRuleUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).r53resolverconn

	d.Partial(true)

	if d.HasChange("name") || d.HasChange("resolver_endpoint_id") || d.HasChange("target_ip") {
		config := &route53resolver.ResolverRuleConfig{}

		if v, ok := d.GetOk("name"); ok {
			config.Name = aws.String(v.(string))
		}

		if v, ok := d.GetOk("resolver_endpoint_id"); ok {
			config.ResolverEndpointId = aws.String(v.(string))
		}

		if v, ok := d.GetOk("target_ip"); ok {
			config.TargetIps = expandTargetIps(v.([]interface{}))
		}

		req := &route53resolver.UpdateResolverRuleInput{
			ResolverRuleId: aws.String(d.Id()),
			Config:         config,
		}

		_, err := conn.UpdateResolverRule(req)
		if err != nil {
			if isAWSErr(err, route53resolver.ErrCodeUnknownResourceException, "") {
				log.Printf("[WARN] No Route 53 Resolver rule by Id (%s) found, removing from state", d.Id())
				d.SetId("")
				return nil
			}
			return fmt.Errorf("Error updating Route 53 Resolver rule %s: %s", d.Id(), err)
		}

		d.SetPartial("name")
		d.SetPartial("resolver_endpoint_id")
		d.SetPartial("target_ip")
	}

	stateConf := &resource.StateChangeConf{
		Pending: []string{route53resolver.ResolverRuleStatusUpdating},
		Target:  []string{route53resolver.ResolverRuleStatusComplete},
		Refresh: resourceAwsRoute53ResolverRuleStateRefresh(conn, d.Id()),
		Timeout: d.Timeout(schema.TimeoutUpdate),
	}
	_, err := stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("Error waiting for Route 53 Resolver rule (%s) to be updated: %s", d.Id(), err)
	}

	if d.HasChange("tags") {
		err := setTagsRoute53Resolver(conn, d, d.Get("arn").(string))
		if err != nil {
			return fmt.Errorf("Error updating Route 53 Resolver rule tags: %s", err)
		}
		d.SetPartial("tags")
	}

	d.Partial(false)

	return resourceAwsRoute53ResolverRuleRead(d, meta)
}

func resourceAwsRoute53ResolverRuleDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).r53resolverconn

	req := &route53resolver.DeleteResolverRuleInput{
		ResolverRuleId: aws.String(d.Id()),
	}

	_, err := conn.DeleteResolverRule(req)
	if err != nil {
		if isAWSErr(err, route53resolver.ErrCodeResourceNotFoundException, "") {
			return nil
		}
		return fmt.Errorf("Error deleting Route 53 Resolver rule %s: %s", d.Id(), err)
	}

	stateConf := &resource.StateChangeConf{
		Pending: []string{route53resolver.ResolverRuleStatusDeleting},
		Target:  []string{Route53ResolverRuleStatusDeleted},
		Refresh: resourceAwsRoute53ResolverRuleStateRefresh(conn, d.Id()),
		Timeout: d.Timeout(schema.TimeoutDelete),
	}
	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("Error waiting for Route 53 Resolver rule (%s) to be deleted: %s", d.Id(), err)
	}

	return nil
}

func resourceAwsRoute53ResolverRuleStateRefresh(conn *route53resolver.Route53Resolver, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {

		req := &route53resolver.GetResolverRuleInput{
			ResolverRuleId: aws.String(id),
		}

		res, err := conn.GetResolverRule(req)
		if err != nil {
			if isAWSErr(err, route53resolver.ErrCodeResourceNotFoundException, "") {
				return "", Route53ResolverRuleStatusDeleted, nil
			}
			return nil, "", err
		}

		return res.ResolverRule, aws.StringValue(res.ResolverRule.Status), nil
	}
}

func expandTargetIps(tips []interface{}) []*route53resolver.TargetAddress {
	tas := make([]*route53resolver.TargetAddress, len(tips), len(tips))

	for i, tip := range tips {
		ta := tip.(map[string]interface{})
		tas[i] = &route53resolver.TargetAddress{
			Ip:   aws.String(ta["ip"].(string)),
			Port: aws.Int64(int64(ta["port"].(int))),
		}
	}

	return tas
}

func flattenTargetIps(tas []*route53resolver.TargetAddress) []interface{} {
	tips := make([]interface{}, len(tas), len(tas))

	for i, ta := range tas {
		tips[i] = map[string]interface{}{
			"ip":   *ta.Ip,
			"port": *ta.Port,
		}
	}
	return tips
}
