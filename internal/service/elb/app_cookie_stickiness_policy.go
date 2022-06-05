package elb

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func ResourceAppCookieStickinessPolicy() *schema.Resource {
	return &schema.Resource{
		// There is no concept of "updating" an App Stickiness policy in
		// the AWS API.
		Create: resourceAppCookieStickinessPolicyCreate,
		Read:   resourceAppCookieStickinessPolicyRead,
		Delete: resourceAppCookieStickinessPolicyDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: func(v interface{}, k string) (ws []string, es []error) {
					value := v.(string)
					if !regexp.MustCompile(`^[0-9A-Za-z-]+$`).MatchString(value) {
						es = append(es, fmt.Errorf(
							"only alphanumeric characters and hyphens allowed in %q", k))
					}
					return
				},
			},

			"load_balancer": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"lb_port": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},

			"cookie_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceAppCookieStickinessPolicyCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ELBConn

	// Provision the AppStickinessPolicy
	acspOpts := &elb.CreateAppCookieStickinessPolicyInput{
		CookieName:       aws.String(d.Get("cookie_name").(string)),
		LoadBalancerName: aws.String(d.Get("load_balancer").(string)),
		PolicyName:       aws.String(d.Get("name").(string)),
	}

	if _, err := conn.CreateAppCookieStickinessPolicy(acspOpts); err != nil {
		return fmt.Errorf("Error creating AppCookieStickinessPolicy: %s", err)
	}

	setLoadBalancerOpts := &elb.SetLoadBalancerPoliciesOfListenerInput{
		LoadBalancerName: aws.String(d.Get("load_balancer").(string)),
		LoadBalancerPort: aws.Int64(int64(d.Get("lb_port").(int))),
		PolicyNames:      []*string{aws.String(d.Get("name").(string))},
	}

	if _, err := conn.SetLoadBalancerPoliciesOfListener(setLoadBalancerOpts); err != nil {
		return fmt.Errorf("Error setting AppCookieStickinessPolicy: %s", err)
	}

	d.SetId(fmt.Sprintf("%s:%d:%s",
		*acspOpts.LoadBalancerName,
		*setLoadBalancerOpts.LoadBalancerPort,
		*acspOpts.PolicyName))
	return nil
}

func resourceAppCookieStickinessPolicyRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ELBConn

	lbName, lbPort, policyName := AppCookieStickinessPolicyParseID(d.Id())

	request := &elb.DescribeLoadBalancerPoliciesInput{
		LoadBalancerName: aws.String(lbName),
		PolicyNames:      []*string{aws.String(policyName)},
	}

	getResp, err := conn.DescribeLoadBalancerPolicies(request)

	if err != nil {
		if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, elb.ErrCodePolicyNotFoundException) {
			log.Printf("[WARN] ELB Classic LB (%s) App Cookie Policy (%s) not found, removing from state", lbName, policyName)
			d.SetId("")
			return nil
		}
		if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, elb.ErrCodeAccessPointNotFoundException) {
			log.Printf("[WARN] ELB Classic LB (%s) not found, removing App Cookie Policy (%s) from state", lbName, policyName)
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error retrieving ELB Classic (%s) App Cookie Policy (%s): %w", lbName, policyName, err)
	}

	if len(getResp.PolicyDescriptions) != 1 {
		return fmt.Errorf("Unable to find policy %#v", getResp.PolicyDescriptions)
	}

	// we know the policy exists now, but we have to check if it's assigned to a listener
	assigned, err := resourceSticknessPolicyAssigned(policyName, lbName, lbPort, conn)
	if err != nil {
		return err
	}
	if !d.IsNewResource() && !assigned {
		log.Printf("[WARN] ELB Classic LB (%s) App Cookie Policy (%s) exists, but isn't assigned to a listener", lbName, policyName)
		d.SetId("")
		return nil
	}

	// We can get away with this because there's only one attribute, the
	// cookie expiration, in these descriptions.
	policyDesc := getResp.PolicyDescriptions[0]
	cookieAttr := policyDesc.PolicyAttributeDescriptions[0]
	if aws.StringValue(cookieAttr.AttributeName) != "CookieName" {
		return fmt.Errorf("Unable to find cookie Name.")
	}

	d.Set("cookie_name", cookieAttr.AttributeValue)
	d.Set("name", policyName)
	d.Set("load_balancer", lbName)

	lbPortInt, _ := strconv.Atoi(lbPort)
	d.Set("lb_port", lbPortInt)

	return nil
}

// Determine if a particular policy is assigned to an ELB listener
func resourceSticknessPolicyAssigned(policyName, lbName, lbPort string, conn *elb.ELB) (bool, error) {
	describeElbOpts := &elb.DescribeLoadBalancersInput{
		LoadBalancerNames: []*string{aws.String(lbName)},
	}
	describeResp, err := conn.DescribeLoadBalancers(describeElbOpts)

	if tfawserr.ErrCodeEquals(err, elb.ErrCodeAccessPointNotFoundException) {
		return false, nil
	}

	if err != nil {
		return false, fmt.Errorf("Error retrieving ELB description: %s", err)
	}

	if len(describeResp.LoadBalancerDescriptions) != 1 {
		return false, fmt.Errorf("Unable to find ELB: %#v", describeResp.LoadBalancerDescriptions)
	}

	lb := describeResp.LoadBalancerDescriptions[0]
	assigned := false
	for _, listener := range lb.ListenerDescriptions {
		if listener == nil || listener.Listener == nil || lbPort != strconv.Itoa(int(aws.Int64Value(listener.Listener.LoadBalancerPort))) {
			continue
		}

		for _, name := range listener.PolicyNames {
			if policyName == aws.StringValue(name) {
				assigned = true
				break
			}
		}
	}

	return assigned, nil
}

func resourceAppCookieStickinessPolicyDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ELBConn

	lbName, _, policyName := AppCookieStickinessPolicyParseID(d.Id())

	// Perversely, if we Set an empty list of PolicyNames, we detach the
	// policies attached to a listener, which is required to delete the
	// policy itself.
	setLoadBalancerOpts := &elb.SetLoadBalancerPoliciesOfListenerInput{
		LoadBalancerName: aws.String(d.Get("load_balancer").(string)),
		LoadBalancerPort: aws.Int64(int64(d.Get("lb_port").(int))),
		PolicyNames:      []*string{},
	}

	if _, err := conn.SetLoadBalancerPoliciesOfListener(setLoadBalancerOpts); err != nil {
		return fmt.Errorf("Error removing AppCookieStickinessPolicy: %s", err)
	}

	request := &elb.DeleteLoadBalancerPolicyInput{
		LoadBalancerName: aws.String(lbName),
		PolicyName:       aws.String(policyName),
	}

	if _, err := conn.DeleteLoadBalancerPolicy(request); err != nil {
		return fmt.Errorf("Error deleting App stickiness policy %s: %s", d.Id(), err)
	}
	return nil
}

// AppCookieStickinessPolicyParseID takes an ID and parses it into
// it's constituent parts. You need three axes (LB name, policy name, and LB
// port) to create or identify a stickiness policy in AWS's API.
func AppCookieStickinessPolicyParseID(id string) (string, string, string) {
	parts := strings.SplitN(id, ":", 3)
	return parts[0], parts[1], parts[2]
}
