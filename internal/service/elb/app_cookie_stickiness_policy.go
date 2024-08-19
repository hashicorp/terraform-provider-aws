// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elb

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing"
	awstypes "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_app_cookie_stickiness_policy", name="App Cookie Stickiness Policy")
func resourceAppCookieStickinessPolicy() *schema.Resource {
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
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: func(v interface{}, k string) (ws []string, es []error) {
					value := v.(string)
					if !regexache.MustCompile(`^[0-9A-Za-z-]+$`).MatchString(value) {
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
	conn := meta.(*conns.AWSClient).ELBClient(ctx)

	lbName := d.Get("load_balancer").(string)
	lbPort := d.Get("lb_port").(int)
	policyName := d.Get(names.AttrName).(string)
	id := appCookieStickinessPolicyCreateResourceID(lbName, lbPort, policyName)
	{
		input := &elasticloadbalancing.CreateAppCookieStickinessPolicyInput{
			CookieName:       aws.String(d.Get("cookie_name").(string)),
			LoadBalancerName: aws.String(lbName),
			PolicyName:       aws.String(policyName),
		}

		_, err := conn.CreateAppCookieStickinessPolicy(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "creating ELB Classic App Cookie Stickiness Policy (%s): %s", id, err)
		}
	}

	d.SetId(id)

	{
		input := &elasticloadbalancing.SetLoadBalancerPoliciesOfListenerInput{
			LoadBalancerName: aws.String(lbName),
			LoadBalancerPort: int32(lbPort),
			PolicyNames:      []string{policyName},
		}

		_, err := conn.SetLoadBalancerPoliciesOfListener(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "setting ELB Classic App Cookie Stickiness Policy (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceAppCookieStickinessPolicyRead(ctx, d, meta)...)
}

func resourceAppCookieStickinessPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ELBClient(ctx)

	lbName, lbPort, policyName, err := appCookieStickinessPolicyParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	policy, err := findLoadBalancerListenerPolicyByThreePartKey(ctx, conn, lbName, lbPort, policyName)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] ELB Classic App Cookie Stickiness Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ELB Classic App Cookie Stickiness Policy (%s): %s", d.Id(), err)
	}

	if len(policy.PolicyAttributeDescriptions) != 1 || aws.ToString(policy.PolicyAttributeDescriptions[0].AttributeName) != "CookieName" {
		return sdkdiag.AppendErrorf(diags, "cookie not found")
	}

	cookieAttr := policy.PolicyAttributeDescriptions[0]
	d.Set("cookie_name", cookieAttr.AttributeValue)
	d.Set("lb_port", lbPort)
	d.Set("load_balancer", lbName)
	d.Set(names.AttrName, policyName)

	return diags
}

func resourceAppCookieStickinessPolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ELBClient(ctx)

	lbName, lbPort, policyName, err := appCookieStickinessPolicyParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	// Perversely, if we Set an empty list of PolicyNames, we detach the
	// policies attached to a listener, which is required to delete the
	// policy itself.
	input := &elasticloadbalancing.SetLoadBalancerPoliciesOfListenerInput{
		LoadBalancerName: aws.String(lbName),
		LoadBalancerPort: int32(lbPort),
		PolicyNames:      []string{},
	}

	_, err = conn.SetLoadBalancerPoliciesOfListener(ctx, input)

	if tfawserr.ErrCodeEquals(err, errCodeLoadBalancerNotFound) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "setting ELB Classic App Cookie Stickiness Policy (%s): %s", d.Id(), err)
	}

	log.Printf("[DEBUG] Deleting ELB Classic App Cookie Stickiness Policy: %s", d.Id())
	_, err = conn.DeleteLoadBalancerPolicy(ctx, &elasticloadbalancing.DeleteLoadBalancerPolicyInput{
		LoadBalancerName: aws.String(lbName),
		PolicyName:       aws.String(policyName),
	})

	if tfawserr.ErrCodeEquals(err, errCodeLoadBalancerNotFound) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting ELB Classic App Cookie Stickiness Policy (%s): %s", d.Id(), err)
	}

	return diags
}

func findLoadBalancerPolicyByTwoPartKey(ctx context.Context, conn *elasticloadbalancing.Client, lbName, policyName string) (*awstypes.PolicyDescription, error) {
	input := &elasticloadbalancing.DescribeLoadBalancerPoliciesInput{
		LoadBalancerName: aws.String(lbName),
		PolicyNames:      []string{policyName},
	}

	output, err := conn.DescribeLoadBalancerPolicies(ctx, input)

	if errs.IsA[*awstypes.PolicyNotFoundException](err) || errs.IsA[*awstypes.AccessPointNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output.PolicyDescriptions)
}

func findLoadBalancerListenerPolicyByThreePartKey(ctx context.Context, conn *elasticloadbalancing.Client, lbName string, lbPort int, policyName string) (*awstypes.PolicyDescription, error) {
	policy, err := findLoadBalancerPolicyByTwoPartKey(ctx, conn, lbName, policyName)

	if err != nil {
		return nil, err
	}

	lb, err := findLoadBalancerByName(ctx, conn, lbName)

	if err != nil {
		return nil, err
	}

	for _, v := range lb.ListenerDescriptions {
		if v.Listener == nil {
			continue
		}

		if v.Listener.LoadBalancerPort != int32(lbPort) {
			continue
		}

		for _, v := range v.PolicyNames {
			if v == policyName {
				return policy, nil
			}
		}
	}

	return nil, &retry.NotFoundError{}
}

const appCookieStickinessPolicyResourceIDSeparator = ":"

func appCookieStickinessPolicyCreateResourceID(lbName string, lbPort int, policyName string) string {
	parts := []string{lbName, strconv.Itoa(lbPort), policyName}
	id := strings.Join(parts, appCookieStickinessPolicyResourceIDSeparator)

	return id
}

func appCookieStickinessPolicyParseResourceID(id string) (string, int, string, error) {
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
