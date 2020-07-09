package aws

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/wafv2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

const (
	Wafv2WebACLCreateTimeout = 5 * time.Minute
	Wafv2WebACLUpdateTimeout = 5 * time.Minute
	Wafv2WebACLDeleteTimeout = 5 * time.Minute
)

func resourceAwsWafv2WebACL() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsWafv2WebACLCreate,
		Read:   resourceAwsWafv2WebACLRead,
		Update: resourceAwsWafv2WebACLUpdate,
		Delete: resourceAwsWafv2WebACLDelete,
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
			"default_action": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"allow": wafv2EmptySchema(),
						"block": wafv2EmptySchema(),
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
			"rule":              wafv2RuleSchema(),
			"tags":              tagsSchema(),
			"visibility_config": wafv2VisibilityConfigSchema(),
		},
	}
}

func resourceAwsWafv2WebACLCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).wafv2conn
	var resp *wafv2.CreateWebACLOutput

	params := &wafv2.CreateWebACLInput{
		Name:             aws.String(d.Get("name").(string)),
		Scope:            aws.String(d.Get("scope").(string)),
		DefaultAction:    expandWafv2DefaultAction(d.Get("default_action").([]interface{})),
		Rules:            expandWafv2Rules(d.Get("rule").(*schema.Set).List()),
		VisibilityConfig: expandWafv2VisibilityConfig(d.Get("visibility_config").([]interface{})),
	}

	if v, ok := d.GetOk("description"); ok {
		params.Description = aws.String(v.(string))
	}

	if v := d.Get("tags").(map[string]interface{}); len(v) > 0 {
		params.Tags = keyvaluetags.New(v).IgnoreAws().Wafv2Tags()
	}

	err := resource.Retry(Wafv2WebACLCreateTimeout, func() *resource.RetryError {
		var err error
		resp, err = conn.CreateWebACL(params)
		if err != nil {
			if isAWSErr(err, wafv2.ErrCodeWAFUnavailableEntityException, "") {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})

	if isResourceTimeoutError(err) {
		_, err = conn.CreateWebACL(params)
	}

	if err != nil {
		return fmt.Errorf("Error creating WAFv2 WebACL: %w", err)
	}

	if resp == nil || resp.Summary == nil {
		return fmt.Errorf("Error creating WAFv2 WebACL")
	}

	d.SetId(aws.StringValue(resp.Summary.Id))

	return resourceAwsWafv2WebACLRead(d, meta)
}

func resourceAwsWafv2WebACLRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).wafv2conn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	params := &wafv2.GetWebACLInput{
		Id:    aws.String(d.Id()),
		Name:  aws.String(d.Get("name").(string)),
		Scope: aws.String(d.Get("scope").(string)),
	}

	resp, err := conn.GetWebACL(params)
	if err != nil {
		if isAWSErr(err, wafv2.ErrCodeWAFNonexistentItemException, "") {
			log.Printf("[WARN] WAFv2 WebACL (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return err
	}

	if resp == nil || resp.WebACL == nil {
		return fmt.Errorf("Error getting WAFv2 WebACL")
	}

	d.Set("name", aws.StringValue(resp.WebACL.Name))
	d.Set("capacity", aws.Int64Value(resp.WebACL.Capacity))
	d.Set("description", aws.StringValue(resp.WebACL.Description))
	d.Set("arn", aws.StringValue(resp.WebACL.ARN))
	d.Set("lock_token", aws.StringValue(resp.LockToken))

	if err := d.Set("default_action", flattenWafv2DefaultAction(resp.WebACL.DefaultAction)); err != nil {
		return fmt.Errorf("Error setting default_action: %w", err)
	}

	if err := d.Set("rule", flattenWafv2Rules(resp.WebACL.Rules)); err != nil {
		return fmt.Errorf("Error setting rule: %w", err)
	}

	if err := d.Set("visibility_config", flattenWafv2VisibilityConfig(resp.WebACL.VisibilityConfig)); err != nil {
		return fmt.Errorf("Error setting visibility_config: %w", err)
	}

	arn := aws.StringValue(resp.WebACL.ARN)
	tags, err := keyvaluetags.Wafv2ListTags(conn, arn)
	if err != nil {
		return fmt.Errorf("Error listing tags for WAFv2 WebACL (%s): %w", arn, err)
	}

	if err := d.Set("tags", tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("Error setting tags: %w", err)
	}

	return nil
}

func resourceAwsWafv2WebACLUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).wafv2conn

	if d.HasChanges("default_action", "description", "rule", "visibility_config") {
		u := &wafv2.UpdateWebACLInput{
			Id:               aws.String(d.Id()),
			Name:             aws.String(d.Get("name").(string)),
			Scope:            aws.String(d.Get("scope").(string)),
			LockToken:        aws.String(d.Get("lock_token").(string)),
			DefaultAction:    expandWafv2DefaultAction(d.Get("default_action").([]interface{})),
			Rules:            expandWafv2Rules(d.Get("rule").(*schema.Set).List()),
			VisibilityConfig: expandWafv2VisibilityConfig(d.Get("visibility_config").([]interface{})),
		}

		if v, ok := d.GetOk("description"); ok && len(v.(string)) > 0 {
			u.Description = aws.String(v.(string))
		}

		err := resource.Retry(Wafv2WebACLUpdateTimeout, func() *resource.RetryError {
			_, err := conn.UpdateWebACL(u)
			if err != nil {
				if isAWSErr(err, wafv2.ErrCodeWAFUnavailableEntityException, "") {
					return resource.RetryableError(err)
				}
				return resource.NonRetryableError(err)
			}
			return nil
		})

		if isResourceTimeoutError(err) {
			_, err = conn.UpdateWebACL(u)
		}

		if err != nil {
			if isAWSErr(err, wafv2.ErrCodeWAFOptimisticLockException, "") {
				return fmt.Errorf("Error updating WAFv2 WebACL, resource has changed since last refresh please run a new plan before applying again: %w", err)
			}
			return fmt.Errorf("Error updating WAFv2 WebACL: %w", err)
		}
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")
		if err := keyvaluetags.Wafv2UpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating tags: %w", err)
		}
	}

	return resourceAwsWafv2WebACLRead(d, meta)
}

func resourceAwsWafv2WebACLDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).wafv2conn

	log.Printf("[INFO] Deleting WAFv2 WebACL %s", d.Id())

	r := &wafv2.DeleteWebACLInput{
		Id:        aws.String(d.Id()),
		Name:      aws.String(d.Get("name").(string)),
		Scope:     aws.String(d.Get("scope").(string)),
		LockToken: aws.String(d.Get("lock_token").(string)),
	}

	err := resource.Retry(Wafv2WebACLDeleteTimeout, func() *resource.RetryError {
		_, err := conn.DeleteWebACL(r)
		if err != nil {
			if isAWSErr(err, wafv2.ErrCodeWAFAssociatedItemException, "") {
				return resource.RetryableError(err)
			}
			if isAWSErr(err, wafv2.ErrCodeWAFUnavailableEntityException, "") {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})

	if isResourceTimeoutError(err) {
		_, err = conn.DeleteWebACL(r)
	}

	if err != nil {
		return fmt.Errorf("Error deleting WAFv2 WebACL: %w", err)
	}

	return nil
}

func wafv2ManagedRuleGroupStatementSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"excluded_rule": wafv2ExcludedRuleSchema(),
				"name": {
					Type:         schema.TypeString,
					Required:     true,
					ValidateFunc: validation.StringLenBetween(1, 128),
				},
				"vendor_name": {
					Type:         schema.TypeString,
					Required:     true,
					ValidateFunc: validation.StringLenBetween(1, 128),
				},
			},
		},
	}
}

func wafv2ExcludedRuleSchema() *schema.Schema {
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

func wafv2RateBasedStatementSchema(level int) *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				// Required field but currently only supports "IP"
				"aggregate_key_type": {
					Type:     schema.TypeString,
					Optional: true,
					Default:  wafv2.RateBasedStatementAggregateKeyTypeIp,
					ValidateFunc: validation.StringInSlice([]string{
						wafv2.RateBasedStatementAggregateKeyTypeIp,
					}, false),
				},
				"limit": {
					Type:         schema.TypeInt,
					Required:     true,
					ValidateFunc: validation.IntBetween(100, 2000000000),
				},
				"scope_down_statement": wafv2ScopeDownStatementSchema(level - 1),
			},
		},
	}
}

func wafv2ScopeDownStatementSchema(level int) *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"and_statement":                         wafv2StatementSchema(level),
				"byte_match_statement":                  wafv2ByteMatchStatementSchema(),
				"geo_match_statement":                   wafv2GeoMatchStatementSchema(),
				"ip_set_reference_statement":            wafv2IpSetReferenceStatementSchema(),
				"not_statement":                         wafv2StatementSchema(level),
				"or_statement":                          wafv2StatementSchema(level),
				"regex_pattern_set_reference_statement": wafv2RegexPatternSetReferenceStatementSchema(),
				"size_constraint_statement":             wafv2SizeConstraintSchema(),
				"sqli_match_statement":                  wafv2SqliMatchStatementSchema(),
				"xss_match_statement":                   wafv2XssMatchStatementSchema(),
			},
		},
	}
}

func wafv2RuleGroupReferenceStatementSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"arn": {
					Type:         schema.TypeString,
					Required:     true,
					ValidateFunc: validateArn,
				},
				"excluded_rule": wafv2ExcludedRuleSchema(),
			},
		},
	}
}

func expandWafv2DefaultAction(l []interface{}) *wafv2.DefaultAction {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})
	action := &wafv2.DefaultAction{}

	if v, ok := m["allow"]; ok && len(v.([]interface{})) > 0 {
		action.Allow = &wafv2.AllowAction{}
	}

	if v, ok := m["block"]; ok && len(v.([]interface{})) > 0 {
		action.Block = &wafv2.BlockAction{}
	}

	return action
}
