package elb

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func ResourceAppCookieStickinessPolicy() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceAppCookieStickinessPolicyCreate,
		ReadWithoutTimeout:   resourceAppCookieStickinessPolicyRead,
		DeleteWithoutTimeout: resourceAppCookieStickinessPolicyDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"cookie_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"lb_port": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},
			"load_balancer": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
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
		},
	}
}

func resourceAppCookieStickinessPolicyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ELBConn()

	lbName := d.Get("load_balancer").(string)
	lbPort := d.Get("lb_port").(int)
	policyName := d.Get("name").(string)
	id := AppCookieStickinessPolicyCreateResourceID(lbName, lbPort, policyName)
	{
		input := &elb.CreateAppCookieStickinessPolicyInput{
			CookieName:       aws.String(d.Get("cookie_name").(string)),
			LoadBalancerName: aws.String(lbName),
			PolicyName:       aws.String(policyName),
		}

		if _, err := conn.CreateAppCookieStickinessPolicyWithContext(ctx, input); err != nil {
			return sdkdiag.AppendErrorf(diags, "creating ELB Classic App Cookie Stickiness Policy (%s): %s", id, err)
		}
	}

	{
		input := &elb.SetLoadBalancerPoliciesOfListenerInput{
			LoadBalancerName: aws.String(lbName),
			LoadBalancerPort: aws.Int64(int64(lbPort)),
			PolicyNames:      aws.StringSlice([]string{policyName}),
		}

		if _, err := conn.SetLoadBalancerPoliciesOfListenerWithContext(ctx, input); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting ELB Classic App Cookie Stickiness Policy (%s): %s", id, err)
		}
	}

	d.SetId(id)

	return append(diags, resourceAppCookieStickinessPolicyRead(ctx, d, meta)...)
}

func resourceAppCookieStickinessPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ELBConn()

	lbName, lbPort, policyName, err := AppCookieStickinessPolicyParseResourceID(d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "parsing resource ID: %s", err)
	}

	policy, err := FindLoadBalancerListenerPolicyByThreePartKey(ctx, conn, lbName, lbPort, policyName)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] ELB Classic App Cookie Stickiness Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ELB Classic App Cookie Stickiness Policy (%s): %s", d.Id(), err)
	}

	if len(policy.PolicyAttributeDescriptions) != 1 || aws.StringValue(policy.PolicyAttributeDescriptions[0].AttributeName) != "CookieName" {
		return sdkdiag.AppendErrorf(diags, "cookie not found")
	}
	cookieAttr := policy.PolicyAttributeDescriptions[0]
	d.Set("cookie_name", cookieAttr.AttributeValue)
	d.Set("lb_port", lbPort)
	d.Set("load_balancer", lbName)
	d.Set("name", policyName)

	return diags
}

func resourceAppCookieStickinessPolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ELBConn()

	lbName, lbPort, policyName, err := AppCookieStickinessPolicyParseResourceID(d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "parsing resource ID: %s", err)
	}

	// Perversely, if we Set an empty list of PolicyNames, we detach the
	// policies attached to a listener, which is required to delete the
	// policy itself.
	input := &elb.SetLoadBalancerPoliciesOfListenerInput{
		LoadBalancerName: aws.String(lbName),
		LoadBalancerPort: aws.Int64(int64(lbPort)),
		PolicyNames:      aws.StringSlice([]string{}),
	}

	_, err = conn.SetLoadBalancerPoliciesOfListenerWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "setting ELB Classic App Cookie Stickiness Policy (%s): %s", d.Id(), err)
	}

	log.Printf("[DEBUG] Deleting ELB Classic App Cookie Stickiness Policy: %s", d.Id())
	_, err = conn.DeleteLoadBalancerPolicyWithContext(ctx, &elb.DeleteLoadBalancerPolicyInput{
		LoadBalancerName: aws.String(lbName),
		PolicyName:       aws.String(policyName),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting ELB Classic App Cookie Stickiness Policy (%s): %s", d.Id(), err)
	}

	return diags
}

func FindLoadBalancerPolicyByTwoPartKey(ctx context.Context, conn *elb.ELB, lbName, policyName string) (*elb.PolicyDescription, error) {
	input := &elb.DescribeLoadBalancerPoliciesInput{
		LoadBalancerName: aws.String(lbName),
		PolicyNames:      aws.StringSlice([]string{policyName}),
	}

	output, err := conn.DescribeLoadBalancerPoliciesWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, elb.ErrCodePolicyNotFoundException, elb.ErrCodeAccessPointNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.PolicyDescriptions) == 0 || output.PolicyDescriptions[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output.PolicyDescriptions); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output.PolicyDescriptions[0], nil
}

func FindLoadBalancerListenerPolicyByThreePartKey(ctx context.Context, conn *elb.ELB, lbName string, lbPort int, policyName string) (*elb.PolicyDescription, error) {
	policy, err := FindLoadBalancerPolicyByTwoPartKey(ctx, conn, lbName, policyName)

	if err != nil {
		return nil, err
	}

	lb, err := FindLoadBalancerByName(ctx, conn, lbName)

	if err != nil {
		return nil, err
	}

	for _, v := range lb.ListenerDescriptions {
		if v == nil || v.Listener == nil {
			continue
		}

		if aws.Int64Value(v.Listener.LoadBalancerPort) != int64(lbPort) {
			continue
		}

		for _, v := range v.PolicyNames {
			if aws.StringValue(v) == policyName {
				return policy, nil
			}
		}
	}

	return nil, &resource.NotFoundError{}
}

const appCookieStickinessPolicyResourceIDSeparator = ":"

func AppCookieStickinessPolicyCreateResourceID(lbName string, lbPort int, policyName string) string {
	parts := []string{lbName, strconv.Itoa(lbPort), policyName}
	id := strings.Join(parts, appCookieStickinessPolicyResourceIDSeparator)

	return id
}

func AppCookieStickinessPolicyParseResourceID(id string) (string, int, string, error) {
	parts := strings.Split(id, appCookieStickinessPolicyResourceIDSeparator)

	if len(parts) == 3 && parts[0] != "" && parts[1] != "" && parts[2] != "" {
		v, err := strconv.Atoi(parts[1])

		if err != nil {
			return "", 0, "", err
		}

		return parts[0], v, parts[2], nil
	}

	return "", 0, "", fmt.Errorf("unexpected format for ID (%[1]s), expected LBNAME%[2]sLBPORT%[2]sPOLICYNAME", id, appCookieStickinessPolicyResourceIDSeparator)
}
