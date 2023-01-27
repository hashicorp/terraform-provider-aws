package elb

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

func ResourceCookieStickinessPolicy() *schema.Resource {
	return &schema.Resource{
		// There is no concept of "updating" an LB Stickiness policy in
		// the AWS API.
		CreateWithoutTimeout: resourceCookieStickinessPolicyCreate,
		ReadWithoutTimeout:   resourceCookieStickinessPolicyRead,
		DeleteWithoutTimeout: resourceCookieStickinessPolicyDelete,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
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

			"cookie_expiration_period": {
				Type:         schema.TypeInt,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.IntAtLeast(0),
			},
		},
	}
}

func resourceCookieStickinessPolicyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ELBConn()

	// Provision the LBStickinessPolicy
	lbspOpts := &elb.CreateLBCookieStickinessPolicyInput{
		LoadBalancerName: aws.String(d.Get("load_balancer").(string)),
		PolicyName:       aws.String(d.Get("name").(string)),
	}

	if v := d.Get("cookie_expiration_period").(int); v > 0 {
		lbspOpts.CookieExpirationPeriod = aws.Int64(int64(v))
	}

	log.Printf("[DEBUG] LB Cookie Stickiness Policy opts: %#v", lbspOpts)
	if _, err := conn.CreateLBCookieStickinessPolicyWithContext(ctx, lbspOpts); err != nil {
		return sdkdiag.AppendErrorf(diags, "creating LBCookieStickinessPolicy: %s", err)
	}

	setLoadBalancerOpts := &elb.SetLoadBalancerPoliciesOfListenerInput{
		LoadBalancerName: aws.String(d.Get("load_balancer").(string)),
		LoadBalancerPort: aws.Int64(int64(d.Get("lb_port").(int))),
		PolicyNames:      []*string{aws.String(d.Get("name").(string))},
	}

	log.Printf("[DEBUG] LB Cookie Stickiness create configuration: %#v", setLoadBalancerOpts)
	if _, err := conn.SetLoadBalancerPoliciesOfListenerWithContext(ctx, setLoadBalancerOpts); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting LBCookieStickinessPolicy: %s", err)
	}

	d.SetId(fmt.Sprintf("%s:%d:%s",
		*lbspOpts.LoadBalancerName,
		*setLoadBalancerOpts.LoadBalancerPort,
		*lbspOpts.PolicyName))
	return diags
}

func resourceCookieStickinessPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ELBConn()

	lbName, lbPort, policyName := CookieStickinessPolicyParseID(d.Id())

	request := &elb.DescribeLoadBalancerPoliciesInput{
		LoadBalancerName: aws.String(lbName),
		PolicyNames:      []*string{aws.String(policyName)},
	}

	getResp, err := conn.DescribeLoadBalancerPoliciesWithContext(ctx, request)
	if err != nil {
		if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, elb.ErrCodePolicyNotFoundException) {
			log.Printf("[WARN] ELB Classic LB (%s) LB Cookie Policy (%s) not found, removing from state", lbName, policyName)
			d.SetId("")
			return diags
		}
		if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, elb.ErrCodeAccessPointNotFoundException) {
			log.Printf("[WARN] ELB Classic LB (%s) not found, removing from state", lbName)
			d.SetId("")
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "retrieving ELB Classic (%s) LB Cookie Policy (%s): %s", lbName, policyName, err)
	}

	if len(getResp.PolicyDescriptions) != 1 {
		return sdkdiag.AppendErrorf(diags, "Unable to find policy %#v", getResp.PolicyDescriptions)
	}

	// we know the policy exists now, but we have to check if it's assigned to a listener
	assigned, err := resourceSticknessPolicyAssigned(ctx, conn, policyName, lbName, lbPort)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ELB Classic LB Cookie Stickiness Policy (%s): %s", d.Id(), err)
	}
	if !d.IsNewResource() && !assigned {
		// policy exists, but isn't assigned to a listener
		log.Printf("[WARN] ELB Classic LB (%s) LB Cookie Policy (%s) exists, but isn't assigned to a listener", lbName, policyName)
		d.SetId("")
		return diags
	}

	// We can get away with this because there's only one attribute, the
	// cookie expiration, in these descriptions.
	policyDesc := getResp.PolicyDescriptions[0]
	cookieAttr := policyDesc.PolicyAttributeDescriptions[0]
	if aws.StringValue(cookieAttr.AttributeName) != "CookieExpirationPeriod" {
		return sdkdiag.AppendErrorf(diags, "Unable to find cookie expiration period.")
	}
	cookieVal, err := strconv.Atoi(aws.StringValue(cookieAttr.AttributeValue))
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "parsing cookie expiration period: %s", err)
	}
	d.Set("cookie_expiration_period", cookieVal)

	d.Set("name", policyName)
	d.Set("load_balancer", lbName)
	lbPortInt, err := strconv.Atoi(lbPort)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ELB Classic LB Cookie Stickiness Policy (%s): parsing port number: %s", d.Id(), err)
	}
	d.Set("lb_port", lbPortInt)

	return diags
}

func resourceCookieStickinessPolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ELBConn()

	lbName, _, policyName := CookieStickinessPolicyParseID(d.Id())

	// Perversely, if we Set an empty list of PolicyNames, we detach the
	// policies attached to a listener, which is required to delete the
	// policy itself.
	setLoadBalancerOpts := &elb.SetLoadBalancerPoliciesOfListenerInput{
		LoadBalancerName: aws.String(d.Get("load_balancer").(string)),
		LoadBalancerPort: aws.Int64(int64(d.Get("lb_port").(int))),
		PolicyNames:      []*string{},
	}

	if _, err := conn.SetLoadBalancerPoliciesOfListenerWithContext(ctx, setLoadBalancerOpts); err != nil {
		return sdkdiag.AppendErrorf(diags, "removing LBCookieStickinessPolicy: %s", err)
	}

	request := &elb.DeleteLoadBalancerPolicyInput{
		LoadBalancerName: aws.String(lbName),
		PolicyName:       aws.String(policyName),
	}

	if _, err := conn.DeleteLoadBalancerPolicyWithContext(ctx, request); err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting LB stickiness policy %s: %s", d.Id(), err)
	}
	return diags
}

// CookieStickinessPolicyParseID takes an ID and parses it into
// it's constituent parts. You need three axes (LB name, policy name, and LB
// port) to create or identify a stickiness policy in AWS's API.
func CookieStickinessPolicyParseID(id string) (string, string, string) {
	parts := strings.SplitN(id, ":", 3)
	return parts[0], parts[1], parts[2]
}
