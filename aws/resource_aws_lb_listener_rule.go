package aws

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elbv2"
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
						"type": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								elbv2.ActionTypeEnumFixedResponse,
								elbv2.ActionTypeEnumForward,
								elbv2.ActionTypeEnumRedirect,
							}, true),
						},

						"target_group_arn": {
							Type:             schema.TypeString,
							Optional:         true,
							DiffSuppressFunc: suppressIfActionTypeNot("forward"),
						},

						"redirect": {
							Type:             schema.TypeList,
							Optional:         true,
							DiffSuppressFunc: suppressIfActionTypeNot("redirect"),
							MaxItems:         1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"host": {
										Type:     schema.TypeString,
										Optional: true,
										Default:  "#{host}",
									},

									"path": {
										Type:     schema.TypeString,
										Optional: true,
										Default:  "/#{path}",
									},

									"port": {
										Type:     schema.TypeString,
										Optional: true,
										Default:  "#{port}",
									},

									"protocol": {
										Type:     schema.TypeString,
										Optional: true,
										Default:  "#{protocol}",
										ValidateFunc: validation.StringInSlice([]string{
											"#{protocol}",
											"HTTP",
											"HTTPS",
										}, false),
									},

									"query": {
										Type:     schema.TypeString,
										Optional: true,
										Default:  "#{query}",
									},

									"status_code": {
										Type:     schema.TypeString,
										Required: true,
										ValidateFunc: validation.StringInSlice([]string{
											"HTTP_301",
											"HTTP_302",
										}, false),
									},
								},
							},
						},

						"fixed_response": {
							Type:             schema.TypeList,
							Optional:         true,
							DiffSuppressFunc: suppressIfActionTypeNot("fixed-response"),
							MaxItems:         1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"content_type": {
										Type:     schema.TypeString,
										Required: true,
										ValidateFunc: validation.StringInSlice([]string{
											"text/plain",
											"text/css",
											"text/html",
											"application/javascript",
											"application/json",
										}, false),
									},

									"message_body": {
										Type:     schema.TypeString,
										Optional: true,
									},

									"status_code": {
										Type:         schema.TypeString,
										Optional:     true,
										Computed:     true,
										ValidateFunc: validation.StringMatch(regexp.MustCompile(`^[245]\d\d$`), ""),
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

func suppressIfActionTypeNot(t string) schema.SchemaDiffSuppressFunc {
	return func(k, old, new string, d *schema.ResourceData) bool {
		take := 2
		i := strings.IndexFunc(k, func(r rune) bool {
			if r == '.' {
				take -= 1
				return take == 0
			}
			return false
		})
		at := k[:i+1] + "type"
		return d.Get(at).(string) != t
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

		action := &elbv2.Action{}
		action.Type = aws.String(actionMap["type"].(string))

		switch actionMap["type"].(string) {
		case "forward":
			action.TargetGroupArn = aws.String(actionMap["target_group_arn"].(string))

		case "redirect":
			redirectList := actionMap["redirect"].([]interface{})

			if len(redirectList) == 1 {
				redirectMap := redirectList[0].(map[string]interface{})

				action.RedirectConfig = &elbv2.RedirectActionConfig{
					Host:       aws.String(redirectMap["host"].(string)),
					Path:       aws.String(redirectMap["path"].(string)),
					Port:       aws.String(redirectMap["port"].(string)),
					Protocol:   aws.String(redirectMap["protocol"].(string)),
					Query:      aws.String(redirectMap["query"].(string)),
					StatusCode: aws.String(redirectMap["status_code"].(string)),
				}
			} else {
				return errors.New("for actions of type 'redirect', you must specify a 'redirect' block")
			}

		case "fixed-response":
			fixedResponseList := actionMap["fixed_response"].([]interface{})

			if len(fixedResponseList) == 1 {
				fixedResponseMap := fixedResponseList[0].(map[string]interface{})

				action.FixedResponseConfig = &elbv2.FixedResponseActionConfig{
					ContentType: aws.String(fixedResponseMap["content_type"].(string)),
					MessageBody: aws.String(fixedResponseMap["message_body"].(string)),
					StatusCode:  aws.String(fixedResponseMap["status_code"].(string)),
				}
			} else {
				return errors.New("for actions of type 'fixed-response', you must specify a 'fixed_response' block")
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

	d.SetId(aws.StringValue(resp.Rules[0].RuleArn))

	return resourceAwsLbListenerRuleRead(d, meta)
}

func resourceAwsLbListenerRuleRead(d *schema.ResourceData, meta interface{}) error {
	elbconn := meta.(*AWSClient).elbv2conn

	resp, err := elbconn.DescribeRules(&elbv2.DescribeRulesInput{
		RuleArns: []*string{aws.String(d.Id())},
	})
	if err != nil {
		if isAWSErr(err, elbv2.ErrCodeRuleNotFoundException, "") {
			log.Printf("[WARN] DescribeRules - removing %s from state", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error retrieving Rules for listener %q: %s", d.Id(), err)
	}

	if len(resp.Rules) != 1 {
		return fmt.Errorf("Error retrieving Rule %q", d.Id())
	}

	rule := resp.Rules[0]

	d.Set("arn", rule.RuleArn)

	// The listener arn isn't in the response but can be derived from the rule arn
	d.Set("listener_arn", lbListenerARNFromRuleARN(aws.StringValue(rule.RuleArn)))

	// Rules are evaluated in priority order, from the lowest value to the highest value. The default rule has the lowest priority.
	if aws.StringValue(rule.Priority) == "default" {
		d.Set("priority", 99999)
	} else {
		if priority, err := strconv.Atoi(aws.StringValue(rule.Priority)); err != nil {
			return fmt.Errorf("Cannot convert rule priority %q to int: %s", aws.StringValue(rule.Priority), err)
		} else {
			d.Set("priority", priority)
		}
	}

	actions := make([]interface{}, len(rule.Actions))
	for i, action := range rule.Actions {
		actionMap := make(map[string]interface{})
		actionMap["type"] = aws.StringValue(action.Type)

		switch actionMap["type"] {
		case "forward":
			actionMap["target_group_arn"] = aws.StringValue(action.TargetGroupArn)

		case "redirect":
			actionMap["redirect"] = []map[string]interface{}{
				{
					"host":        aws.StringValue(action.RedirectConfig.Host),
					"path":        aws.StringValue(action.RedirectConfig.Path),
					"port":        aws.StringValue(action.RedirectConfig.Port),
					"protocol":    aws.StringValue(action.RedirectConfig.Protocol),
					"query":       aws.StringValue(action.RedirectConfig.Query),
					"status_code": aws.StringValue(action.RedirectConfig.StatusCode),
				},
			}

		case "fixed-response":
			actionMap["fixed_response"] = []map[string]interface{}{
				{
					"content_type": aws.StringValue(action.FixedResponseConfig.ContentType),
					"message_body": aws.StringValue(action.FixedResponseConfig.MessageBody),
					"status_code":  aws.StringValue(action.FixedResponseConfig.StatusCode),
				},
			}
		}

		actions[i] = actionMap
	}
	d.Set("action", actions)

	conditions := make([]interface{}, len(rule.Conditions))
	for i, condition := range rule.Conditions {
		conditionMap := make(map[string]interface{})
		conditionMap["field"] = aws.StringValue(condition.Field)
		conditionValues := make([]string, len(condition.Values))
		for k, value := range condition.Values {
			conditionValues[k] = aws.StringValue(value)
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

			action := &elbv2.Action{}
			action.Type = aws.String(actionMap["type"].(string))

			switch actionMap["type"].(string) {
			case "forward":
				action.TargetGroupArn = aws.String(actionMap["target_group_arn"].(string))

			case "redirect":
				redirectList := actionMap["redirect"].([]interface{})

				if len(redirectList) == 1 {
					redirectMap := redirectList[0].(map[string]interface{})

					action.RedirectConfig = &elbv2.RedirectActionConfig{
						Host:       aws.String(redirectMap["host"].(string)),
						Path:       aws.String(redirectMap["path"].(string)),
						Port:       aws.String(redirectMap["port"].(string)),
						Protocol:   aws.String(redirectMap["protocol"].(string)),
						Query:      aws.String(redirectMap["query"].(string)),
						StatusCode: aws.String(redirectMap["status_code"].(string)),
					}
				} else {
					return errors.New("for actions of type 'redirect', you must specify a 'redirect' block")
				}

			case "fixed-response":
				fixedResponseList := actionMap["fixed_response"].([]interface{})

				if len(fixedResponseList) == 1 {
					fixedResponseMap := fixedResponseList[0].(map[string]interface{})

					action.FixedResponseConfig = &elbv2.FixedResponseActionConfig{
						ContentType: aws.String(fixedResponseMap["content_type"].(string)),
						MessageBody: aws.String(fixedResponseMap["message_body"].(string)),
						StatusCode:  aws.String(fixedResponseMap["status_code"].(string)),
					}
				} else {
					return errors.New("for actions of type 'fixed-response', you must specify a 'fixed_response' block")
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
			return fmt.Errorf("Error modifying LB Listener Rule: %s", err)
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
	if err != nil && !isAWSErr(err, elbv2.ErrCodeRuleNotFoundException, "") {
		return fmt.Errorf("Error deleting LB Listener Rule: %s", err)
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
			if aws.StringValue(rule.Priority) != "default" {
				p, _ := strconv.Atoi(aws.StringValue(rule.Priority))
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
