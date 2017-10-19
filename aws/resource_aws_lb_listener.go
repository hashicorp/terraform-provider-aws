package aws

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/hashicorp/errwrap"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
)

const healthCheckOnlyHostname = "health-check-only.terraform.localhost"

func resourceAwsLbListener() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsLbListenerCreate,
		Read:   resourceAwsLbListenerRead,
		Update: resourceAwsLbListenerUpdate,
		Delete: resourceAwsLbListenerDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"load_balancer_arn": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"port": {
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validateAwsLbListenerPort,
			},

			"protocol": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "HTTP",
				StateFunc: func(v interface{}) string {
					return strings.ToUpper(v.(string))
				},
				ValidateFunc: validateAwsLbListenerProtocol,
			},

			"ssl_policy": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"certificate_arn": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"wait_for_capacity_timeout": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "10m",
				ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
					value := v.(string)
					duration, err := time.ParseDuration(value)
					if err != nil {
						errors = append(errors, fmt.Errorf(
							"%q cannot be parsed as a duration: %s", k, err))
					}
					if duration < 0 {
						errors = append(errors, fmt.Errorf(
							"%q must be greater than zero", k))
					}
					return
				},
			},

			"wait_for_target_group_capacity": {
				Type:     schema.TypeInt,
				Optional: true,
			},

			"default_action": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"target_group_arn": {
							Type:     schema.TypeString,
							Required: true,
						},
						"type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validateAwsLbListenerActionType,
						},
					},
				},
			},
		},
	}
}

func resourceAwsLbListenerCreate(d *schema.ResourceData, meta interface{}) error {
	elbconn := meta.(*AWSClient).elbv2conn

	lbArn := d.Get("load_balancer_arn").(string)

	params := &elbv2.CreateListenerInput{
		LoadBalancerArn: aws.String(lbArn),
		Port:            aws.Int64(int64(d.Get("port").(int))),
		Protocol:        aws.String(d.Get("protocol").(string)),
	}

	if sslPolicy, ok := d.GetOk("ssl_policy"); ok {
		params.SslPolicy = aws.String(sslPolicy.(string))
	}

	if certificateArn, ok := d.GetOk("certificate_arn"); ok {
		params.Certificates = make([]*elbv2.Certificate, 1)
		params.Certificates[0] = &elbv2.Certificate{
			CertificateArn: aws.String(certificateArn.(string)),
		}
	}

	if defaultActions := d.Get("default_action").([]interface{}); len(defaultActions) == 1 {
		params.DefaultActions = make([]*elbv2.Action, len(defaultActions))

		for i, defaultAction := range defaultActions {
			defaultActionMap := defaultAction.(map[string]interface{})

			params.DefaultActions[i] = &elbv2.Action{
				TargetGroupArn: aws.String(defaultActionMap["target_group_arn"].(string)),
				Type:           aws.String(defaultActionMap["type"].(string)),
			}
		}
	}

	var resp *elbv2.CreateListenerOutput

	err := resource.Retry(5*time.Minute, func() *resource.RetryError {
		var err error
		log.Printf("[DEBUG] Creating LB listener for ARN: %s", d.Get("load_balancer_arn").(string))
		resp, err = elbconn.CreateListener(params)
		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr.Code() == "CertificateNotFound" {
				log.Printf("[WARN] Got an error while trying to create LB listener for ARN: %s: %s", lbArn, err)
				return resource.RetryableError(err)
			}
		}
		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if err != nil {
		return errwrap.Wrapf("Error creating LB Listener: {{err}}", err)
	}

	if len(resp.Listeners) == 0 {
		return errors.New("Error creating LB Listener: no listeners returned in response")
	}

	d.SetId(*resp.Listeners[0].ListenerArn)

	return resourceAwsLbListenerRead(d, meta)
}

func resourceAwsLbListenerRead(d *schema.ResourceData, meta interface{}) error {
	elbconn := meta.(*AWSClient).elbv2conn

	resp, err := elbconn.DescribeListeners(&elbv2.DescribeListenersInput{
		ListenerArns: []*string{aws.String(d.Id())},
	})
	if err != nil {
		if isListenerNotFound(err) {
			log.Printf("[WARN] DescribeListeners - removing %s from state", d.Id())
			d.SetId("")
			return nil
		}
		return errwrap.Wrapf("Error retrieving Listener: {{err}}", err)
	}

	if len(resp.Listeners) != 1 {
		return fmt.Errorf("Error retrieving Listener %q", d.Id())
	}

	listener := resp.Listeners[0]

	d.Set("arn", listener.ListenerArn)
	d.Set("load_balancer_arn", listener.LoadBalancerArn)
	d.Set("port", listener.Port)
	d.Set("protocol", listener.Protocol)
	d.Set("ssl_policy", listener.SslPolicy)

	if listener.Certificates != nil && len(listener.Certificates) == 1 {
		d.Set("certificate_arn", listener.Certificates[0].CertificateArn)
	}

	defaultActions := make([]map[string]interface{}, 0)
	if listener.DefaultActions != nil && len(listener.DefaultActions) > 0 {
		for _, defaultAction := range listener.DefaultActions {
			action := map[string]interface{}{
				"target_group_arn": *defaultAction.TargetGroupArn,
				"type":             *defaultAction.Type,
			}
			defaultActions = append(defaultActions, action)
		}
	}
	d.Set("default_action", defaultActions)

	return nil
}

func resourceAwsLbListenerUpdate(d *schema.ResourceData, meta interface{}) error {
	elbconn := meta.(*AWSClient).elbv2conn

	shouldWaitForCapacity := false

	params := &elbv2.ModifyListenerInput{
		ListenerArn: aws.String(d.Id()),
		Port:        aws.Int64(int64(d.Get("port").(int))),
		Protocol:    aws.String(d.Get("protocol").(string)),
	}

	if sslPolicy, ok := d.GetOk("ssl_policy"); ok {
		params.SslPolicy = aws.String(sslPolicy.(string))
	}

	if certificateArn, ok := d.GetOk("certificate_arn"); ok {
		params.Certificates = make([]*elbv2.Certificate, 1)
		params.Certificates[0] = &elbv2.Certificate{
			CertificateArn: aws.String(certificateArn.(string)),
		}
	}

	if defaultActions := d.Get("default_action").([]interface{}); len(defaultActions) == 1 {
		params.DefaultActions = make([]*elbv2.Action, len(defaultActions))

		// TODO: does this only execute on a change to the target_group_arn ?
		shouldWaitForCapacity = true

		for i, defaultAction := range defaultActions {
			defaultActionMap := defaultAction.(map[string]interface{})

			params.DefaultActions[i] = &elbv2.Action{
				TargetGroupArn: aws.String(defaultActionMap["target_group_arn"].(string)),
				Type:           aws.String(defaultActionMap["type"].(string)),
			}
		}
	}

	if shouldWaitForCapacity {

		err := addHealthCheckOnlyRule(params, elbconn, d.Get("arn").(string))
		if err != nil {
			return errwrap.Wrapf("Error adding health-check only rule to listener: {{err}}", err)
		}

		if err := waitForListenerTargetGroupCapacity(d, meta, func(d *schema.ResourceData, current int, target int) (bool, string) {
			if current < target {
				return false, fmt.Sprintf("Need at least %d healthy instances in target group, have %d", current, target)
			}
			return true, ""
		}); err != nil {
			return errwrap.Wrapf("Error waiting for Target Group Capacity: {{err}}", err)
		}
	}

	_, err := elbconn.ModifyListener(params)
	if err != nil {
		return errwrap.Wrapf("Error modifying LB Listener: {{err}}", err)
	}

	err = removeHealthCheckOnlyRule(params, elbconn, d.Get("arn").(string))
	if err != nil {
		return errwrap.Wrapf("Error modifying ALB Listener: {{err}}", err)
	}

	return resourceAwsLbListenerRead(d, meta)
}

func resourceAwsLbListenerDelete(d *schema.ResourceData, meta interface{}) error {
	elbconn := meta.(*AWSClient).elbv2conn

	_, err := elbconn.DeleteListener(&elbv2.DeleteListenerInput{
		ListenerArn: aws.String(d.Id()),
	})
	if err != nil {
		return errwrap.Wrapf("Error deleting Listener: {{err}}", err)
	}

	return nil
}

func validateAwsLbListenerPort(v interface{}, k string) (ws []string, errors []error) {
	port := v.(int)
	if port < 1 || port > 65536 {
		errors = append(errors, fmt.Errorf("%q must be a valid port number (1-65536)", k))
	}
	return
}

func validateAwsLbListenerProtocol(v interface{}, k string) (ws []string, errors []error) {
	value := strings.ToLower(v.(string))
	if value == "http" || value == "https" || value == "tcp" {
		return
	}

	errors = append(errors, fmt.Errorf("%q must be either %q, %q or %q", k, "HTTP", "HTTPS", "TCP"))
	return
}

func validateAwsLbListenerActionType(v interface{}, k string) (ws []string, errors []error) {
	value := strings.ToLower(v.(string))
	if value != "forward" {
		errors = append(errors, fmt.Errorf("%q must have the value %q", k, "forward"))
	}
	return
}

func isListenerNotFound(err error) bool {
	elberr, ok := err.(awserr.Error)
	return ok && elberr.Code() == "ListenerNotFound"
}

// Adds a custom rule to a listener for the sole purpose of
// enabling health checks on a target group
func addHealthCheckOnlyRule(params *elbv2.ModifyListenerInput, elbconn *elbv2.ELBV2, listenerArn string) error {
	targetGroupArn := params.DefaultActions[0].TargetGroupArn

	resp, err := elbconn.CreateRule(&elbv2.CreateRuleInput{
		ListenerArn: aws.String(listenerArn),
		Actions: []*elbv2.Action{&elbv2.Action{
			TargetGroupArn: targetGroupArn,
			Type:           aws.String("forward"),
		}},
		Conditions: []*elbv2.RuleCondition{&elbv2.RuleCondition{
			Field:  aws.String("host-header"),
			Values: aws.StringSlice([]string{healthCheckOnlyHostname}),
		}},
		Priority: aws.Int64(100),
	})

	log.Printf("[DEBUG] resp.Rules.length: %d", len(resp.Rules))

	if err != nil {
		return errwrap.Wrapf("Error creating temporary listener rule: {{err}}", err)
	}
	log.Printf("[DEBUG] addHealthCheckOnlyRule: added rule")
	return nil
}

// Removes the custom health-check-only rule from a listener
func removeHealthCheckOnlyRule(params *elbv2.ModifyListenerInput, elbconn *elbv2.ELBV2, listenerArn string) error {

	resp, err := elbconn.DescribeRules(&elbv2.DescribeRulesInput{
		ListenerArn: aws.String(listenerArn),
	})

	if err != nil {
		return errwrap.Wrapf("Error describing temporary listener rules: {{err}}", err)
	}

Rules:
	for _, rule := range resp.Rules {
		log.Printf("[DEBUG] removeHealthCheckOnlyRule: testing rule %v", rule)
		if !aws.BoolValue(rule.IsDefault) {
			for _, cond := range rule.Conditions {
				field := aws.StringValue(cond.Field)
				values := aws.StringValueSlice(cond.Values)
				if field == "host-header" && len(values) == 1 && values[0] == healthCheckOnlyHostname {
					log.Printf("[DEBUG] removeHealthCheckOnlyRule: removing rule %v", rule)
					_, err = elbconn.DeleteRule(&elbv2.DeleteRuleInput{
						RuleArn: rule.RuleArn,
					})

					if err != nil {
						return errwrap.Wrapf("Error removing temporary listener rule: {{err}}", err)
					}
					continue Rules
				}
			}
		}
	}

	return nil
}
