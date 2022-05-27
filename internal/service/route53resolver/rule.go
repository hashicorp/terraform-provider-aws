package route53resolver

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53resolver"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

const (
	RuleStatusDeleted = "DELETED"
)

const (
	ruleCreatedDefaultTimeout = 10 * time.Minute
	ruleUpdatedDefaultTimeout = 10 * time.Minute
	ruleDeletedDefaultTimeout = 10 * time.Minute
)

func ResourceRule() *schema.Resource {
	return &schema.Resource{
		Create: resourceRuleCreate,
		Read:   resourceRuleRead,
		Update: resourceRuleUpdate,
		Delete: resourceRuleDelete,
		CustomizeDiff: customdiff.Sequence(
			resourceRuleCustomizeDiff,
			verify.SetTagsDiff,
		),
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(ruleCreatedDefaultTimeout),
			Update: schema.DefaultTimeout(ruleUpdatedDefaultTimeout),
			Delete: schema.DefaultTimeout(ruleDeletedDefaultTimeout),
		},

		Schema: map[string]*schema.Schema{
			"domain_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 256),
				StateFunc:    trimTrailingPeriod,
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

			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validResolverName,
			},

			"resolver_endpoint_id": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"target_ip": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ip": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.IsIPAddress,
						},
						"port": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      53,
							ValidateFunc: validation.IntBetween(1, 65535),
						},
					},
				},
				Set: ruleHashTargetIP,
			},

			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),

			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"owner_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"share_status": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceRuleCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).Route53ResolverConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	req := &route53resolver.CreateResolverRuleInput{
		CreatorRequestId: aws.String(resource.PrefixedUniqueId("tf-r53-resolver-rule-")),
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
		req.TargetIps = expandRuleTargetIPs(v.(*schema.Set))
	}
	if v, ok := d.GetOk("tags"); ok && len(v.(map[string]interface{})) > 0 {
		req.Tags = Tags(tags.IgnoreAWS())
	}

	log.Printf("[DEBUG] Creating Route 53 Resolver rule: %s", req)
	resp, err := conn.CreateResolverRule(req)
	if err != nil {
		return fmt.Errorf("error creating Route 53 Resolver rule: %s", err)
	}

	d.SetId(aws.StringValue(resp.ResolverRule.Id))

	err = RuleWaitUntilTargetState(conn, d.Id(), d.Timeout(schema.TimeoutCreate),
		[]string{}, // Should go straight to COMPLETE
		[]string{route53resolver.ResolverRuleStatusComplete})
	if err != nil {
		return err
	}

	return resourceRuleRead(d, meta)
}

func resourceRuleRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).Route53ResolverConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	ruleRaw, state, err := ruleRefresh(conn, d.Id())()
	if err != nil {
		return fmt.Errorf("error getting Route53 Resolver rule (%s): %s", d.Id(), err)
	}
	if state == RuleStatusDeleted {
		log.Printf("[WARN] Route53 Resolver rule (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	rule := ruleRaw.(*route53resolver.ResolverRule)
	d.Set("arn", rule.Arn)
	// To be consistent with other AWS services that do not accept a trailing period,
	// we remove the suffix from the Domain Name returned from the API
	d.Set("domain_name", trimTrailingPeriod(aws.StringValue(rule.DomainName)))
	d.Set("name", rule.Name)
	d.Set("owner_id", rule.OwnerId)
	d.Set("resolver_endpoint_id", rule.ResolverEndpointId)
	d.Set("rule_type", rule.RuleType)
	d.Set("share_status", rule.ShareStatus)
	if err := d.Set("target_ip", schema.NewSet(ruleHashTargetIP, flattenRuleTargetIPs(rule.TargetIps))); err != nil {
		return err
	}

	tags, err := ListTags(conn, d.Get("arn").(string))

	if err != nil {
		return fmt.Errorf("error listing tags for Route53 Resolver rule (%s): %s", d.Get("arn").(string), err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceRuleUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).Route53ResolverConn

	if d.HasChanges("name", "resolver_endpoint_id", "target_ip") {
		req := &route53resolver.UpdateResolverRuleInput{
			ResolverRuleId: aws.String(d.Id()),
			Config:         &route53resolver.ResolverRuleConfig{},
		}
		if v, ok := d.GetOk("name"); ok {
			req.Config.Name = aws.String(v.(string))
		}
		if v, ok := d.GetOk("resolver_endpoint_id"); ok {
			req.Config.ResolverEndpointId = aws.String(v.(string))
		}
		if v, ok := d.GetOk("target_ip"); ok {
			req.Config.TargetIps = expandRuleTargetIPs(v.(*schema.Set))
		}

		log.Printf("[DEBUG] Updating Route53 Resolver rule: %#v", req)
		_, err := conn.UpdateResolverRule(req)
		if err != nil {
			return fmt.Errorf("error updating Route 53 Resolver rule (%s): %s", d.Id(), err)
		}

		err = RuleWaitUntilTargetState(conn, d.Id(), d.Timeout(schema.TimeoutUpdate),
			[]string{route53resolver.ResolverRuleStatusUpdating},
			[]string{route53resolver.ResolverRuleStatusComplete})
		if err != nil {
			return err
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")
		if err := UpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating Route53 Resolver rule (%s) tags: %s", d.Get("arn").(string), err)
		}
	}

	return resourceRuleRead(d, meta)
}

func resourceRuleDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).Route53ResolverConn

	log.Printf("[DEBUG] Deleting Route53 Resolver rule: %s", d.Id())
	_, err := conn.DeleteResolverRule(&route53resolver.DeleteResolverRuleInput{
		ResolverRuleId: aws.String(d.Id()),
	})
	if tfawserr.ErrCodeEquals(err, route53resolver.ErrCodeResourceNotFoundException) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("error deleting Route 53 Resolver rule (%s): %s", d.Id(), err)
	}

	err = RuleWaitUntilTargetState(conn, d.Id(), d.Timeout(schema.TimeoutDelete),
		[]string{route53resolver.ResolverRuleStatusDeleting},
		[]string{RuleStatusDeleted})
	if err != nil {
		return err
	}

	return nil
}

func resourceRuleCustomizeDiff(_ context.Context, diff *schema.ResourceDiff, v interface{}) error {
	if diff.Id() != "" {
		if diff.HasChange("resolver_endpoint_id") {
			if _, n := diff.GetChange("resolver_endpoint_id"); n.(string) == "" {
				if err := diff.ForceNew("resolver_endpoint_id"); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func ruleRefresh(conn *route53resolver.Route53Resolver, ruleId string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		resp, err := conn.GetResolverRule(&route53resolver.GetResolverRuleInput{
			ResolverRuleId: aws.String(ruleId),
		})
		if tfawserr.ErrCodeEquals(err, route53resolver.ErrCodeResourceNotFoundException) {
			return "", RuleStatusDeleted, nil
		}
		if err != nil {
			return nil, "", err
		}

		if statusMessage := aws.StringValue(resp.ResolverRule.StatusMessage); statusMessage != "" {
			log.Printf("[INFO] Route 53 Resolver rule (%s) status message: %s", ruleId, statusMessage)
		}

		return resp.ResolverRule, aws.StringValue(resp.ResolverRule.Status), nil
	}
}

func RuleWaitUntilTargetState(conn *route53resolver.Route53Resolver, ruleId string, timeout time.Duration, pending, target []string) error {
	stateConf := &resource.StateChangeConf{
		Pending:    pending,
		Target:     target,
		Refresh:    ruleRefresh(conn, ruleId),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 5 * time.Second,
	}
	if _, err := stateConf.WaitForState(); err != nil {
		return fmt.Errorf("error waiting for Route53 Resolver rule (%s) to reach target state: %s", ruleId, err)
	}

	return nil
}

func ruleHashTargetIP(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	buf.WriteString(fmt.Sprintf("%s-%d-", m["ip"].(string), m["port"].(int)))
	return create.StringHashcode(buf.String())
}

// trimTrailingPeriod is used to remove the trailing period
// of "name" or "domain name" attributes often returned from
// the Route53 API or provided as user input.
// The single dot (".") domain name is returned as-is.
func trimTrailingPeriod(v interface{}) string {
	var str string
	switch value := v.(type) {
	case *string:
		str = aws.StringValue(value)
	case string:
		str = value
	default:
		return ""
	}

	if str == "." {
		return str
	}

	return strings.TrimSuffix(str, ".")
}
