package wafv2

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/wafv2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

const (
	webACLCreateTimeout = 5 * time.Minute
	webACLUpdateTimeout = 5 * time.Minute
	webACLDeleteTimeout = 5 * time.Minute
)

func ResourceWebACL() *schema.Resource {
	return &schema.Resource{
		Create: resourceWebACLCreate,
		Read:   resourceWebACLRead,
		Update: resourceWebACLUpdate,
		Delete: resourceWebACLDelete,
		Importer: &schema.ResourceImporter{
			State: func(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				idParts := strings.Split(d.Id(), "/")
				if len(idParts) != 3 || idParts[0] == "" || idParts[1] == "" || idParts[2] == "" {
					return nil, fmt.Errorf("Unexpected format of ID (%q), expected ID/NAME/SCOPE", d.Id())
				}
				id := idParts[0]
				name := idParts[1]
				scope := idParts[2]
				d.SetId(id)
				d.Set("name", name)
				d.Set("scope", scope)
				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"capacity": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"custom_response_body": customResponseBodySchema(),
			"default_action": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"allow": allowConfigSchema(),
						"block": blockConfigSchema(),
					},
				},
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 256),
			},
			"lock_token": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 128),
					validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9-_]+$`), "must contain only alphanumeric hyphen and underscore characters"),
				),
			},
			"scope": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					wafv2.ScopeCloudfront,
					wafv2.ScopeRegional,
				}, false),
			},
			"rule": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"action": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"allow": allowConfigSchema(),
									"block": blockConfigSchema(),
									"count": countConfigSchema(),
								},
							},
						},
						"name": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(1, 128),
						},
						"override_action": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"count": emptySchema(),
									"none":  emptySchema(),
								},
							},
						},
						"priority": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"rule_label":        ruleLabelsSchema(),
						"statement":         webACLRootStatementSchema(webACLRootStatementSchemaLevel),
						"visibility_config": visibilityConfigSchema(),
					},
				},
			},
			"tags":              tftags.TagsSchema(),
			"tags_all":          tftags.TagsSchemaComputed(),
			"visibility_config": visibilityConfigSchema(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceWebACLCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).WAFV2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))
	var resp *wafv2.CreateWebACLOutput

	params := &wafv2.CreateWebACLInput{
		Name:             aws.String(d.Get("name").(string)),
		Scope:            aws.String(d.Get("scope").(string)),
		DefaultAction:    expandDefaultAction(d.Get("default_action").([]interface{})),
		Rules:            expandWebACLRules(d.Get("rule").(*schema.Set).List()),
		VisibilityConfig: expandVisibilityConfig(d.Get("visibility_config").([]interface{})),
	}

	if v, ok := d.GetOk("custom_response_body"); ok && v.(*schema.Set).Len() > 0 {
		params.CustomResponseBodies = expandCustomResponseBodies(v.(*schema.Set).List())
	}

	if v, ok := d.GetOk("description"); ok {
		params.Description = aws.String(v.(string))
	}

	if len(tags) > 0 {
		params.Tags = Tags(tags.IgnoreAWS())
	}

	err := resource.Retry(webACLCreateTimeout, func() *resource.RetryError {
		var err error
		resp, err = conn.CreateWebACL(params)
		if err != nil {
			if tfawserr.ErrCodeEquals(err, wafv2.ErrCodeWAFUnavailableEntityException) {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})

	if tfresource.TimedOut(err) {
		resp, err = conn.CreateWebACL(params)
	}

	if err != nil {
		return fmt.Errorf("Error creating WAFv2 WebACL: %w", err)
	}

	if resp == nil || resp.Summary == nil {
		return fmt.Errorf("Error creating WAFv2 WebACL")
	}

	d.SetId(aws.StringValue(resp.Summary.Id))

	return resourceWebACLRead(d, meta)
}

func resourceWebACLRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).WAFV2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	params := &wafv2.GetWebACLInput{
		Id:    aws.String(d.Id()),
		Name:  aws.String(d.Get("name").(string)),
		Scope: aws.String(d.Get("scope").(string)),
	}

	resp, err := conn.GetWebACL(params)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, wafv2.ErrCodeWAFNonexistentItemException) {
			log.Printf("[WARN] WAFv2 WebACL (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return err
	}

	if resp == nil || resp.WebACL == nil {
		return fmt.Errorf("Error getting WAFv2 WebACL")
	}

	d.Set("name", resp.WebACL.Name)
	d.Set("capacity", resp.WebACL.Capacity)
	d.Set("description", resp.WebACL.Description)
	d.Set("arn", resp.WebACL.ARN)
	d.Set("lock_token", resp.LockToken)

	if err := d.Set("custom_response_body", flattenCustomResponseBodies(resp.WebACL.CustomResponseBodies)); err != nil {
		return fmt.Errorf("Error setting custom_response_body: %w", err)
	}

	if err := d.Set("default_action", flattenDefaultAction(resp.WebACL.DefaultAction)); err != nil {
		return fmt.Errorf("Error setting default_action: %w", err)
	}

	if err := d.Set("rule", flattenWebACLRules(resp.WebACL.Rules)); err != nil {
		return fmt.Errorf("Error setting rule: %w", err)
	}

	if err := d.Set("visibility_config", flattenVisibilityConfig(resp.WebACL.VisibilityConfig)); err != nil {
		return fmt.Errorf("Error setting visibility_config: %w", err)
	}

	arn := aws.StringValue(resp.WebACL.ARN)
	tags, err := ListTags(conn, arn)
	if err != nil {
		return fmt.Errorf("Error listing tags for WAFv2 WebACL (%s): %w", arn, err)
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

func resourceWebACLUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).WAFV2Conn

	if d.HasChanges("custom_response_body", "default_action", "description", "rule", "visibility_config") {
		u := &wafv2.UpdateWebACLInput{
			Id:               aws.String(d.Id()),
			Name:             aws.String(d.Get("name").(string)),
			Scope:            aws.String(d.Get("scope").(string)),
			LockToken:        aws.String(d.Get("lock_token").(string)),
			DefaultAction:    expandDefaultAction(d.Get("default_action").([]interface{})),
			Rules:            expandWebACLRules(d.Get("rule").(*schema.Set).List()),
			VisibilityConfig: expandVisibilityConfig(d.Get("visibility_config").([]interface{})),
		}

		if v, ok := d.GetOk("custom_response_body"); ok && v.(*schema.Set).Len() > 0 {
			u.CustomResponseBodies = expandCustomResponseBodies(v.(*schema.Set).List())
		}

		if v, ok := d.GetOk("description"); ok {
			u.Description = aws.String(v.(string))
		}

		err := resource.Retry(webACLUpdateTimeout, func() *resource.RetryError {
			_, err := conn.UpdateWebACL(u)
			if err != nil {
				if tfawserr.ErrCodeEquals(err, wafv2.ErrCodeWAFUnavailableEntityException) {
					return resource.RetryableError(err)
				}
				return resource.NonRetryableError(err)
			}
			return nil
		})

		if tfresource.TimedOut(err) {
			_, err = conn.UpdateWebACL(u)
		}

		if err != nil {
			if tfawserr.ErrCodeEquals(err, wafv2.ErrCodeWAFOptimisticLockException) {
				return fmt.Errorf("Error updating WAFv2 WebACL, resource has changed since last refresh please run a new plan before applying again: %w", err)
			}
			return fmt.Errorf("Error updating WAFv2 WebACL: %w", err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")
		if err := UpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating tags: %w", err)
		}
	}

	return resourceWebACLRead(d, meta)
}

func resourceWebACLDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).WAFV2Conn

	log.Printf("[INFO] Deleting WAFv2 WebACL %s", d.Id())

	r := &wafv2.DeleteWebACLInput{
		Id:        aws.String(d.Id()),
		Name:      aws.String(d.Get("name").(string)),
		Scope:     aws.String(d.Get("scope").(string)),
		LockToken: aws.String(d.Get("lock_token").(string)),
	}

	err := resource.Retry(webACLDeleteTimeout, func() *resource.RetryError {
		_, err := conn.DeleteWebACL(r)
		if err != nil {
			if tfawserr.ErrCodeEquals(err, wafv2.ErrCodeWAFAssociatedItemException) {
				return resource.RetryableError(err)
			}
			if tfawserr.ErrCodeEquals(err, wafv2.ErrCodeWAFUnavailableEntityException) {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})

	if tfresource.TimedOut(err) {
		_, err = conn.DeleteWebACL(r)
	}

	if tfawserr.ErrCodeEquals(err, wafv2.ErrCodeWAFNonexistentItemException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("Error deleting WAFv2 WebACL: %w", err)
	}

	return nil
}

func webACLRootStatementSchema(level int) *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Required: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"and_statement":                         statementSchema(level),
				"byte_match_statement":                  byteMatchStatementSchema(),
				"geo_match_statement":                   geoMatchStatementSchema(),
				"ip_set_reference_statement":            ipSetReferenceStatementSchema(),
				"label_match_statement":                 labelMatchStatementSchema(),
				"managed_rule_group_statement":          managedRuleGroupStatementSchema(level),
				"not_statement":                         statementSchema(level),
				"or_statement":                          statementSchema(level),
				"rate_based_statement":                  rateBasedStatementSchema(level),
				"regex_pattern_set_reference_statement": regexPatternSetReferenceStatementSchema(),
				"rule_group_reference_statement":        ruleGroupReferenceStatementSchema(),
				"size_constraint_statement":             sizeConstraintSchema(),
				"sqli_match_statement":                  sqliMatchStatementSchema(),
				"xss_match_statement":                   xssMatchStatementSchema(),
			},
		},
	}
}

func managedRuleGroupStatementSchema(level int) *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"excluded_rule": excludedRuleSchema(),
				"name": {
					Type:         schema.TypeString,
					Required:     true,
					ValidateFunc: validation.StringLenBetween(1, 128),
				},
				"scope_down_statement": scopeDownStatementSchema(level - 1),
				"vendor_name": {
					Type:         schema.TypeString,
					Required:     true,
					ValidateFunc: validation.StringLenBetween(1, 128),
				},
				"version": {
					Type:         schema.TypeString,
					Optional:     true,
					ValidateFunc: validation.StringLenBetween(1, 128),
				},
			},
		},
	}
}

func excludedRuleSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"name": {
					Type:         schema.TypeString,
					Required:     true,
					ValidateFunc: validation.StringLenBetween(1, 128),
				},
			},
		},
	}
}

func rateBasedStatementSchema(level int) *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				// Required field
				"aggregate_key_type": {
					Type:         schema.TypeString,
					Optional:     true,
					Default:      wafv2.RateBasedStatementAggregateKeyTypeIp,
					ValidateFunc: validation.StringInSlice(wafv2.RateBasedStatementAggregateKeyType_Values(), false),
				},
				"forwarded_ip_config": forwardedIPConfig(),
				"limit": {
					Type:         schema.TypeInt,
					Required:     true,
					ValidateFunc: validation.IntBetween(100, 2000000000),
				},
				"scope_down_statement": scopeDownStatementSchema(level - 1),
			},
		},
	}
}

func scopeDownStatementSchema(level int) *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"and_statement":                         statementSchema(level),
				"byte_match_statement":                  byteMatchStatementSchema(),
				"geo_match_statement":                   geoMatchStatementSchema(),
				"label_match_statement":                 labelMatchStatementSchema(),
				"ip_set_reference_statement":            ipSetReferenceStatementSchema(),
				"not_statement":                         statementSchema(level),
				"or_statement":                          statementSchema(level),
				"regex_pattern_set_reference_statement": regexPatternSetReferenceStatementSchema(),
				"size_constraint_statement":             sizeConstraintSchema(),
				"sqli_match_statement":                  sqliMatchStatementSchema(),
				"xss_match_statement":                   xssMatchStatementSchema(),
			},
		},
	}
}

func ruleGroupReferenceStatementSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"arn": {
					Type:         schema.TypeString,
					Required:     true,
					ValidateFunc: verify.ValidARN,
				},
				"excluded_rule": excludedRuleSchema(),
			},
		},
	}
}

func expandWebACLRules(l []interface{}) []*wafv2.Rule {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	rules := make([]*wafv2.Rule, 0)

	for _, rule := range l {
		if rule == nil {
			continue
		}
		rules = append(rules, expandWebACLRule(rule.(map[string]interface{})))
	}

	return rules
}

func expandWebACLRule(m map[string]interface{}) *wafv2.Rule {
	if m == nil {
		return nil
	}

	rule := &wafv2.Rule{
		Name:             aws.String(m["name"].(string)),
		Priority:         aws.Int64(int64(m["priority"].(int))),
		Action:           expandRuleAction(m["action"].([]interface{})),
		OverrideAction:   expandOverrideAction(m["override_action"].([]interface{})),
		Statement:        expandWebACLRootStatement(m["statement"].([]interface{})),
		VisibilityConfig: expandVisibilityConfig(m["visibility_config"].([]interface{})),
	}

	if v, ok := m["rule_label"].(*schema.Set); ok && v.Len() > 0 {
		rule.RuleLabels = expandRuleLabels(v.List())
	}

	return rule
}

func expandOverrideAction(l []interface{}) *wafv2.OverrideAction {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})
	action := &wafv2.OverrideAction{}

	if v, ok := m["count"]; ok && len(v.([]interface{})) > 0 {
		action.Count = &wafv2.CountAction{}
	}

	if v, ok := m["none"]; ok && len(v.([]interface{})) > 0 {
		action.None = &wafv2.NoneAction{}
	}

	return action
}

func expandDefaultAction(l []interface{}) *wafv2.DefaultAction {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})
	action := &wafv2.DefaultAction{}

	if v, ok := m["allow"]; ok && len(v.([]interface{})) > 0 {
		action.Allow = expandAllowAction(v.([]interface{}))
	}

	if v, ok := m["block"]; ok && len(v.([]interface{})) > 0 {
		action.Block = expandBlockAction(v.([]interface{}))
	}

	return action
}

func expandWebACLRootStatement(l []interface{}) *wafv2.Statement {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	return expandWebACLStatement(m)
}

func expandWebACLStatement(m map[string]interface{}) *wafv2.Statement {
	if m == nil {
		return nil
	}

	statement := &wafv2.Statement{}

	if v, ok := m["and_statement"]; ok {
		statement.AndStatement = expandAndStatement(v.([]interface{}))
	}

	if v, ok := m["byte_match_statement"]; ok {
		statement.ByteMatchStatement = expandByteMatchStatement(v.([]interface{}))
	}

	if v, ok := m["ip_set_reference_statement"]; ok {
		statement.IPSetReferenceStatement = expandIPSetReferenceStatement(v.([]interface{}))
	}

	if v, ok := m["geo_match_statement"]; ok {
		statement.GeoMatchStatement = expandGeoMatchStatement(v.([]interface{}))
	}

	if v, ok := m["label_match_statement"]; ok {
		statement.LabelMatchStatement = expandLabelMatchStatement(v.([]interface{}))
	}

	if v, ok := m["managed_rule_group_statement"]; ok {
		statement.ManagedRuleGroupStatement = expandManagedRuleGroupStatement(v.([]interface{}))
	}

	if v, ok := m["not_statement"]; ok {
		statement.NotStatement = expandNotStatement(v.([]interface{}))
	}

	if v, ok := m["or_statement"]; ok {
		statement.OrStatement = expandOrStatement(v.([]interface{}))
	}

	if v, ok := m["rate_based_statement"]; ok {
		statement.RateBasedStatement = expandRateBasedStatement(v.([]interface{}))
	}

	if v, ok := m["regex_pattern_set_reference_statement"]; ok {
		statement.RegexPatternSetReferenceStatement = expandRegexPatternSetReferenceStatement(v.([]interface{}))
	}

	if v, ok := m["rule_group_reference_statement"]; ok {
		statement.RuleGroupReferenceStatement = expandRuleGroupReferenceStatement(v.([]interface{}))
	}

	if v, ok := m["size_constraint_statement"]; ok {
		statement.SizeConstraintStatement = expandSizeConstraintStatement(v.([]interface{}))
	}

	if v, ok := m["sqli_match_statement"]; ok {
		statement.SqliMatchStatement = expandSQLiMatchStatement(v.([]interface{}))
	}

	if v, ok := m["xss_match_statement"]; ok {
		statement.XssMatchStatement = expandXSSMatchStatement(v.([]interface{}))
	}

	return statement
}

func expandManagedRuleGroupStatement(l []interface{}) *wafv2.ManagedRuleGroupStatement {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})
	r := &wafv2.ManagedRuleGroupStatement{
		ExcludedRules: expandExcludedRules(m["excluded_rule"].([]interface{})),
		Name:          aws.String(m["name"].(string)),
		VendorName:    aws.String(m["vendor_name"].(string)),
	}

	if s, ok := m["scope_down_statement"].([]interface{}); ok && len(s) > 0 && s[0] != nil {
		r.ScopeDownStatement = expandStatement(s[0].(map[string]interface{}))
	}

	if v, ok := m["version"]; ok && v != "" {
		r.Version = aws.String(v.(string))
	}

	return r
}

func expandRateBasedStatement(l []interface{}) *wafv2.RateBasedStatement {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})
	r := &wafv2.RateBasedStatement{
		AggregateKeyType: aws.String(m["aggregate_key_type"].(string)),
		Limit:            aws.Int64(int64(m["limit"].(int))),
	}

	if v, ok := m["forwarded_ip_config"]; ok {
		r.ForwardedIPConfig = expandForwardedIPConfig(v.([]interface{}))
	}

	s := m["scope_down_statement"].([]interface{})
	if len(s) > 0 && s[0] != nil {
		r.ScopeDownStatement = expandStatement(s[0].(map[string]interface{}))
	}

	return r
}

func expandRuleGroupReferenceStatement(l []interface{}) *wafv2.RuleGroupReferenceStatement {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	return &wafv2.RuleGroupReferenceStatement{
		ARN:           aws.String(m["arn"].(string)),
		ExcludedRules: expandExcludedRules(m["excluded_rule"].([]interface{})),
	}
}

func expandExcludedRules(l []interface{}) []*wafv2.ExcludedRule {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	rules := make([]*wafv2.ExcludedRule, 0)

	for _, rule := range l {
		if rule == nil {
			continue
		}
		rules = append(rules, expandExcludedRule(rule.(map[string]interface{})))
	}

	return rules
}

func expandExcludedRule(m map[string]interface{}) *wafv2.ExcludedRule {
	if m == nil {
		return nil
	}

	return &wafv2.ExcludedRule{
		Name: aws.String(m["name"].(string)),
	}
}

func flattenWebACLRootStatement(s *wafv2.Statement) interface{} {
	if s == nil {
		return []interface{}{}
	}

	return []interface{}{flattenWebACLStatement(s)}
}

func flattenWebACLStatement(s *wafv2.Statement) map[string]interface{} {
	if s == nil {
		return map[string]interface{}{}
	}

	m := map[string]interface{}{}

	if s.AndStatement != nil {
		m["and_statement"] = flattenAndStatement(s.AndStatement)
	}

	if s.ByteMatchStatement != nil {
		m["byte_match_statement"] = flattenByteMatchStatement(s.ByteMatchStatement)
	}

	if s.IPSetReferenceStatement != nil {
		m["ip_set_reference_statement"] = flattenIPSetReferenceStatement(s.IPSetReferenceStatement)
	}

	if s.GeoMatchStatement != nil {
		m["geo_match_statement"] = flattenGeoMatchStatement(s.GeoMatchStatement)
	}

	if s.LabelMatchStatement != nil {
		m["label_match_statement"] = flattenLabelMatchStatement(s.LabelMatchStatement)
	}

	if s.ManagedRuleGroupStatement != nil {
		m["managed_rule_group_statement"] = flattenManagedRuleGroupStatement(s.ManagedRuleGroupStatement)
	}

	if s.NotStatement != nil {
		m["not_statement"] = flattenNotStatement(s.NotStatement)
	}

	if s.OrStatement != nil {
		m["or_statement"] = flattenOrStatement(s.OrStatement)
	}

	if s.RateBasedStatement != nil {
		m["rate_based_statement"] = flattenRateBasedStatement(s.RateBasedStatement)
	}

	if s.RegexPatternSetReferenceStatement != nil {
		m["regex_pattern_set_reference_statement"] = flattenRegexPatternSetReferenceStatement(s.RegexPatternSetReferenceStatement)
	}

	if s.RuleGroupReferenceStatement != nil {
		m["rule_group_reference_statement"] = flattenRuleGroupReferenceStatement(s.RuleGroupReferenceStatement)
	}

	if s.SizeConstraintStatement != nil {
		m["size_constraint_statement"] = flattenSizeConstraintStatement(s.SizeConstraintStatement)
	}

	if s.SqliMatchStatement != nil {
		m["sqli_match_statement"] = flattenSQLiMatchStatement(s.SqliMatchStatement)
	}

	if s.XssMatchStatement != nil {
		m["xss_match_statement"] = flattenXSSMatchStatement(s.XssMatchStatement)
	}

	return m
}

func flattenWebACLRules(r []*wafv2.Rule) interface{} {
	out := make([]map[string]interface{}, len(r))
	for i, rule := range r {
		m := make(map[string]interface{})
		m["action"] = flattenRuleAction(rule.Action)
		m["override_action"] = flattenOverrideAction(rule.OverrideAction)
		m["name"] = aws.StringValue(rule.Name)
		m["priority"] = int(aws.Int64Value(rule.Priority))
		m["rule_label"] = flattenRuleLabels(rule.RuleLabels)
		m["statement"] = flattenWebACLRootStatement(rule.Statement)
		m["visibility_config"] = flattenVisibilityConfig(rule.VisibilityConfig)
		out[i] = m
	}

	return out
}

func flattenOverrideAction(a *wafv2.OverrideAction) interface{} {
	if a == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{}

	if a.Count != nil {
		m["count"] = make([]map[string]interface{}, 1)
	}

	if a.None != nil {
		m["none"] = make([]map[string]interface{}, 1)
	}

	return []interface{}{m}
}

func flattenDefaultAction(a *wafv2.DefaultAction) interface{} {
	if a == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{}

	if a.Allow != nil {
		m["allow"] = flattenAllow(a.Allow)
	}

	if a.Block != nil {
		m["block"] = flattenBlock(a.Block)
	}

	return []interface{}{m}
}

func flattenManagedRuleGroupStatement(apiObject *wafv2.ManagedRuleGroupStatement) interface{} {
	if apiObject == nil {
		return []interface{}{}
	}

	tfMap := map[string]interface{}{}

	if apiObject.ExcludedRules != nil {
		tfMap["excluded_rule"] = flattenExcludedRules(apiObject.ExcludedRules)
	}

	if apiObject.Name != nil {
		tfMap["name"] = aws.StringValue(apiObject.Name)
	}

	if apiObject.ScopeDownStatement != nil {
		tfMap["scope_down_statement"] = []interface{}{flattenStatement(apiObject.ScopeDownStatement)}
	}

	if apiObject.VendorName != nil {
		tfMap["vendor_name"] = aws.StringValue(apiObject.VendorName)
	}

	if apiObject.Version != nil {
		tfMap["version"] = aws.StringValue(apiObject.Version)
	}

	return []interface{}{tfMap}
}

func flattenRateBasedStatement(apiObject *wafv2.RateBasedStatement) interface{} {
	if apiObject == nil {
		return []interface{}{}
	}

	tfMap := map[string]interface{}{}

	if apiObject.AggregateKeyType != nil {
		tfMap["aggregate_key_type"] = aws.StringValue(apiObject.AggregateKeyType)
	}

	if apiObject.ForwardedIPConfig != nil {
		tfMap["forwarded_ip_config"] = flattenForwardedIPConfig(apiObject.ForwardedIPConfig)
	}

	if apiObject.Limit != nil {
		tfMap["limit"] = int(aws.Int64Value(apiObject.Limit))
	}

	if apiObject.ScopeDownStatement != nil {
		tfMap["scope_down_statement"] = []interface{}{flattenStatement(apiObject.ScopeDownStatement)}
	}

	return []interface{}{tfMap}
}

func flattenRuleGroupReferenceStatement(r *wafv2.RuleGroupReferenceStatement) interface{} {
	if r == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"excluded_rule": flattenExcludedRules(r.ExcludedRules),
		"arn":           aws.StringValue(r.ARN),
	}

	return []interface{}{m}
}

func flattenExcludedRules(r []*wafv2.ExcludedRule) interface{} {
	out := make([]map[string]interface{}, len(r))
	for i, rule := range r {
		m := make(map[string]interface{})
		m["name"] = aws.StringValue(rule.Name)
		out[i] = m
	}

	return out
}
