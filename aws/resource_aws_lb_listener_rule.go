package aws

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"sort"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/hashicorp/errwrap"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
)

func resourceAwsLbbListenerRule() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsLbListenerRuleCreate,
		Read:   resourceAwsLbListenerRuleRead,
		Update: resourceAwsLbListenerRuleUpdate,
		Delete: resourceAwsLbListenerRuleDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"listener_arn": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"priority": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validateAwsLbListenerRulePriority,
			},
			"action": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"target_group_arn": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validateLbListenerActionType(),
						},
						"order": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntBetween(1, 50000),
						},
						"authenticate_cognito_config": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"authentication_request_extra_params": {
										Type:     schema.TypeMap,
										Optional: true,
									},
									"on_unauthenticated_request": {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
										ValidateFunc: validation.StringInSlice([]string{
											elbv2.AuthenticateCognitoActionConditionalBehaviorEnumDeny,
											elbv2.AuthenticateCognitoActionConditionalBehaviorEnumAllow,
											elbv2.AuthenticateCognitoActionConditionalBehaviorEnumAuthenticate,
										}, true),
									},
									"scope": {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
									},
									"session_cookie_name": {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
									},
									"session_timeout": {
										Type:     schema.TypeInt,
										Optional: true,
										Computed: true,
									},
									"user_pool_arn": {
										Type:     schema.TypeString,
										Required: true,
									},
									"user_pool_client": {
										Type:     schema.TypeString,
										Required: true,
									},
									"user_pool_domain": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},
						"authenticate_oidc_config": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"authentication_request_extra_params": {
										Type:     schema.TypeMap,
										Optional: true,
									},
									"authorization_endpoint": {
										Type:     schema.TypeString,
										Required: true,
									},
									"client_id": {
										Type:     schema.TypeString,
										Required: true,
									},
									"client_secret": {
										Type:      schema.TypeString,
										Required:  true,
										Sensitive: true,
									},
									"issuer": {
										Type:     schema.TypeString,
										Required: true,
									},
									"on_unauthenticated_request": {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
										ValidateFunc: validation.StringInSlice([]string{
											elbv2.AuthenticateOidcActionConditionalBehaviorEnumDeny,
											elbv2.AuthenticateOidcActionConditionalBehaviorEnumAllow,
											elbv2.AuthenticateOidcActionConditionalBehaviorEnumAuthenticate,
										}, true),
									},
									"scope": {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
									},
									"session_cookie_name": {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
									},
									"session_timeout": {
										Type:     schema.TypeInt,
										Optional: true,
										Computed: true,
									},
									"token_endpoint": {
										Type:     schema.TypeString,
										Required: true,
									},
									"user_info_endpoint": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},
					},
				},
			},
			"condition": {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"field": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validateMaxLength(64),
						},
						"values": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Optional: true,
						},
					},
				},
			},
		},
	}
}

func resourceAwsLbListenerRuleCreate(d *schema.ResourceData, meta interface{}) error {
	elbconn := meta.(*AWSClient).elbv2conn
	listenerArn := d.Get("listener_arn").(string)

	params := &elbv2.CreateRuleInput{
		ListenerArn: aws.String(listenerArn),
	}

	actions := d.Get("action").([]interface{})
	params.Actions = make([]*elbv2.Action, len(actions))
	for i, action := range actions {
		actionMap := action.(map[string]interface{})

		actionType := actionMap["type"].(string)
		action := &elbv2.Action{
			Type: aws.String(actionType),
		}
		if v, ok := actionMap["order"].(int); ok && v != 0 {
			action.Order = aws.Int64(int64(v))
		}

		switch actionType {
		case elbv2.ActionTypeEnumForward:
			if v, ok := actionMap["target_group_arn"].(string); ok && v != "" {
				action.TargetGroupArn = aws.String(v)
			}
		case elbv2.ActionTypeEnumAuthenticateOidc:
			if v, ok := actionMap["authenticate_oidc_config"].([]interface{}); ok {
				action.AuthenticateOidcConfig = expandELbAuthenticateOidcActionConfig(v[0].(map[string]interface{}))
			}
		case elbv2.ActionTypeEnumAuthenticateCognito:
			if v, ok := actionMap["authenticate_cognito_config"].([]interface{}); ok {
				action.AuthenticateCognitoConfig = expandELbAuthenticateCognitoActionConfig(v[0].(map[string]interface{}))
			}
		}
		params.Actions[i] = action
	}

	conditions := d.Get("condition").(*schema.Set).List()
	params.Conditions = make([]*elbv2.RuleCondition, len(conditions))
	for i, condition := range conditions {
		conditionMap := condition.(map[string]interface{})
		values := conditionMap["values"].([]interface{})
		params.Conditions[i] = &elbv2.RuleCondition{
			Field:  aws.String(conditionMap["field"].(string)),
			Values: make([]*string, len(values)),
		}
		for j, value := range values {
			params.Conditions[i].Values[j] = aws.String(value.(string))
		}
	}

	var resp *elbv2.CreateRuleOutput
	if v, ok := d.GetOk("priority"); ok {
		var err error
		params.Priority = aws.Int64(int64(v.(int)))
		resp, err = elbconn.CreateRule(params)
		if err != nil {
			return fmt.Errorf("Error creating LB Listener Rule: %v", err)
		}
	} else {
		err := resource.Retry(5*time.Minute, func() *resource.RetryError {
			var err error
			priority, err := highestListenerRulePriority(elbconn, listenerArn)
			if err != nil {
				return resource.NonRetryableError(err)
			}
			params.Priority = aws.Int64(priority + 1)
			resp, err = elbconn.CreateRule(params)
			if err != nil {
				if isAWSErr(err, elbv2.ErrCodePriorityInUseException, "") {
					return resource.RetryableError(err)
				}
				return resource.NonRetryableError(err)
			}
			return nil
		})
		if err != nil {
			return fmt.Errorf("Error creating LB Listener Rule: %v", err)
		}
	}

	if len(resp.Rules) == 0 {
		return errors.New("Error creating LB Listener Rule: no rules returned in response")
	}

	d.SetId(*resp.Rules[0].RuleArn)

	return resourceAwsLbListenerRuleRead(d, meta)
}

func resourceAwsLbListenerRuleRead(d *schema.ResourceData, meta interface{}) error {
	elbconn := meta.(*AWSClient).elbv2conn

	resp, err := elbconn.DescribeRules(&elbv2.DescribeRulesInput{
		RuleArns: []*string{aws.String(d.Id())},
	})
	if err != nil {
		if isRuleNotFound(err) {
			log.Printf("[WARN] DescribeRules - removing %s from state", d.Id())
			d.SetId("")
			return nil
		}
		return errwrap.Wrapf(fmt.Sprintf("Error retrieving Rules for listener %s: {{err}}", d.Id()), err)
	}

	if len(resp.Rules) != 1 {
		return fmt.Errorf("Error retrieving Rule %q", d.Id())
	}

	rule := resp.Rules[0]

	d.Set("arn", rule.RuleArn)

	// The listener arn isn't in the response but can be derived from the rule arn
	d.Set("listener_arn", lbListenerARNFromRuleARN(*rule.RuleArn))

	// Rules are evaluated in priority order, from the lowest value to the highest value. The default rule has the lowest priority.
	if *rule.Priority == "default" {
		d.Set("priority", 99999)
	} else {
		if priority, err := strconv.Atoi(*rule.Priority); err != nil {
			return fmt.Errorf("Cannot convert rule priority %q to int: {{err}}", err)
		} else {
			d.Set("priority", priority)
		}
	}

	sortedActions := sortActionsBasedonTypeinTFFile("action", rule.Actions, d)
	actions := make([]interface{}, 0, len(sortedActions))
	for i, action := range sortedActions {
		m := make(map[string]interface{})
		if action.Order != nil {
			m["order"] = int(aws.Int64Value(action.Order))
		}
		actionType := aws.StringValue(action.Type)
		m["type"] = actionType

		switch actionType {
		case elbv2.ActionTypeEnumForward:
			m["target_group_arn"] = aws.StringValue(action.TargetGroupArn)
		case elbv2.ActionTypeEnumAuthenticateOidc:
			// Since the client_secret is never returned from the API ignore it and use whats already in the state
			client_secret := d.Get("action." + strconv.Itoa(i) + ".authenticate_oidc_config.0.client_secret").(string)
			m["authenticate_oidc_config"] = flattenELbAuthenticateOidcActionConfig(action.AuthenticateOidcConfig, client_secret)
		case elbv2.ActionTypeEnumAuthenticateCognito:
			m["authenticate_cognito_config"] = flattenELbAuthenticateCognitoActionConfig(action.AuthenticateCognitoConfig)
		}
		actions = append(actions, m)
	}
	d.Set("action", actions)

	conditions := make([]interface{}, len(rule.Conditions))
	for i, condition := range rule.Conditions {
		conditionMap := make(map[string]interface{})
		conditionMap["field"] = *condition.Field
		conditionValues := make([]string, len(condition.Values))
		for k, value := range condition.Values {
			conditionValues[k] = *value
		}
		conditionMap["values"] = conditionValues
		conditions[i] = conditionMap
	}
	d.Set("condition", conditions)

	return nil
}

func resourceAwsLbListenerRuleUpdate(d *schema.ResourceData, meta interface{}) error {
	elbconn := meta.(*AWSClient).elbv2conn

	d.Partial(true)

	if d.HasChange("priority") {
		params := &elbv2.SetRulePrioritiesInput{
			RulePriorities: []*elbv2.RulePriorityPair{
				{
					RuleArn:  aws.String(d.Id()),
					Priority: aws.Int64(int64(d.Get("priority").(int))),
				},
			},
		}

		_, err := elbconn.SetRulePriorities(params)
		if err != nil {
			return err
		}

		d.SetPartial("priority")
	}

	requestUpdate := false
	params := &elbv2.ModifyRuleInput{
		RuleArn: aws.String(d.Id()),
	}

	if d.HasChange("action") {
		actions := d.Get("action").([]interface{})
		params.Actions = make([]*elbv2.Action, len(actions))
		for i, action := range actions {
			actionMap := action.(map[string]interface{})

			actionType := actionMap["type"].(string)
			action := &elbv2.Action{
				Type: aws.String(actionType),
			}
			if v, ok := actionMap["order"].(int); ok && v != 0 {
				action.Order = aws.Int64(int64(v))
			}

			switch actionType {
			case elbv2.ActionTypeEnumForward:
				if v, ok := actionMap["target_group_arn"].(string); ok && v != "" {
					action.TargetGroupArn = aws.String(v)
				}
			case elbv2.ActionTypeEnumAuthenticateOidc:
				if v, ok := actionMap["authenticate_oidc_config"].([]interface{}); ok {
					action.AuthenticateOidcConfig = expandELbAuthenticateOidcActionConfig(v[0].(map[string]interface{}))
				}
			case elbv2.ActionTypeEnumAuthenticateCognito:
				if v, ok := actionMap["authenticate_cognito_config"].([]interface{}); ok {
					action.AuthenticateCognitoConfig = expandELbAuthenticateCognitoActionConfig(v[0].(map[string]interface{}))
				}
			}
			params.Actions[i] = action
		}
		requestUpdate = true
		d.SetPartial("action")
	}

	if d.HasChange("condition") {
		conditions := d.Get("condition").(*schema.Set).List()
		params.Conditions = make([]*elbv2.RuleCondition, len(conditions))
		for i, condition := range conditions {
			conditionMap := condition.(map[string]interface{})
			values := conditionMap["values"].([]interface{})
			params.Conditions[i] = &elbv2.RuleCondition{
				Field:  aws.String(conditionMap["field"].(string)),
				Values: make([]*string, len(values)),
			}
			for j, value := range values {
				params.Conditions[i].Values[j] = aws.String(value.(string))
			}
		}
		requestUpdate = true
		d.SetPartial("condition")
	}

	if requestUpdate {
		resp, err := elbconn.ModifyRule(params)
		if err != nil {
			return errwrap.Wrapf("Error modifying LB Listener Rule: {{err}}", err)
		}

		if len(resp.Rules) == 0 {
			return errors.New("Error modifying creating LB Listener Rule: no rules returned in response")
		}
	}

	d.Partial(false)

	return resourceAwsLbListenerRuleRead(d, meta)
}

func resourceAwsLbListenerRuleDelete(d *schema.ResourceData, meta interface{}) error {
	elbconn := meta.(*AWSClient).elbv2conn

	_, err := elbconn.DeleteRule(&elbv2.DeleteRuleInput{
		RuleArn: aws.String(d.Id()),
	})
	if err != nil && !isRuleNotFound(err) {
		return errwrap.Wrapf("Error deleting LB Listener Rule: {{err}}", err)
	}
	return nil
}

func validateAwsLbListenerRulePriority(v interface{}, k string) (ws []string, errors []error) {
	value := v.(int)
	if value < 1 || (value > 50000 && value != 99999) {
		errors = append(errors, fmt.Errorf("%q must be in the range 1-50000 for normal rule or 99999 for default rule", k))
	}
	return
}

// from arn:
// arn:aws:elasticloadbalancing:us-east-1:012345678912:listener-rule/app/name/0123456789abcdef/abcdef0123456789/456789abcedf1234
// select submatches:
// (arn:aws:elasticloadbalancing:us-east-1:012345678912:listener)-rule(/app/name/0123456789abcdef/abcdef0123456789)/456789abcedf1234
// concat to become:
// arn:aws:elasticloadbalancing:us-east-1:012345678912:listener/app/name/0123456789abcdef/abcdef0123456789
var lbListenerARNFromRuleARNRegexp = regexp.MustCompile(`^(arn:.+:listener)-rule(/.+)/[^/]+$`)

func lbListenerARNFromRuleARN(ruleArn string) string {
	if arnComponents := lbListenerARNFromRuleARNRegexp.FindStringSubmatch(ruleArn); len(arnComponents) > 1 {
		return arnComponents[1] + arnComponents[2]
	}

	return ""
}

func isRuleNotFound(err error) bool {
	elberr, ok := err.(awserr.Error)
	return ok && elberr.Code() == "RuleNotFound"
}

func highestListenerRulePriority(conn *elbv2.ELBV2, arn string) (priority int64, err error) {
	var priorities []int
	var nextMarker *string

	for {
		out, aerr := conn.DescribeRules(&elbv2.DescribeRulesInput{
			ListenerArn: aws.String(arn),
			Marker:      nextMarker,
		})
		if aerr != nil {
			err = aerr
			return
		}
		for _, rule := range out.Rules {
			if *rule.Priority != "default" {
				p, _ := strconv.Atoi(*rule.Priority)
				priorities = append(priorities, p)
			}
		}
		if out.NextMarker == nil {
			break
		}
		nextMarker = out.NextMarker
	}

	if len(priorities) == 0 {
		priority = 0
		return
	}

	sort.IntSlice(priorities).Sort()
	priority = int64(priorities[len(priorities)-1])

	return
}
