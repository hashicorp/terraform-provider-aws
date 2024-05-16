// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elb

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKResource("aws_lb_cookie_stickiness_policy")
func ResourceCookieStickinessPolicy() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceCookieStickinessPolicyCreate,
		ReadWithoutTimeout:   resourceCookieStickinessPolicyRead,
		DeleteWithoutTimeout: resourceCookieStickinessPolicyDelete,

		Schema: map[string]*schema.Schema{
			"cookie_expiration_period": {
				Type:         schema.TypeInt,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.IntAtLeast(0),
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
			},
		},
	}
}

func resourceCookieStickinessPolicyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ELBConn(ctx)

	lbName := d.Get("load_balancer").(string)
	lbPort := d.Get("lb_port").(int)
	policyName := d.Get("name").(string)
	id := LBCookieStickinessPolicyCreateResourceID(lbName, lbPort, policyName)
	{
		input := &elb.CreateLBCookieStickinessPolicyInput{
			LoadBalancerName: aws.String(lbName),
			PolicyName:       aws.String(policyName),
		}

		if v, ok := d.GetOk("cookie_expiration_period"); ok {
			input.CookieExpirationPeriod = aws.Int64(int64(v.(int)))
		}

		_, err := conn.CreateLBCookieStickinessPolicyWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "creating ELB Classic LB Cookie Stickiness Policy (%s): %s", id, err)
		}
	}

	{
		input := &elb.SetLoadBalancerPoliciesOfListenerInput{
			LoadBalancerName: aws.String(lbName),
			LoadBalancerPort: aws.Int64(int64(lbPort)),
			PolicyNames:      aws.StringSlice([]string{policyName}),
		}

		_, err := conn.SetLoadBalancerPoliciesOfListenerWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "setting ELB Classic LB Cookie Stickiness Policy (%s): %s", id, err)
		}
	}

	d.SetId(id)

	return append(diags, resourceCookieStickinessPolicyRead(ctx, d, meta)...)
}

func resourceCookieStickinessPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ELBConn(ctx)

	lbName, lbPort, policyName, err := LBCookieStickinessPolicyParseResourceID(d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "parsing resource ID: %s", err)
	}

	policy, err := FindLoadBalancerListenerPolicyByThreePartKey(ctx, conn, lbName, lbPort, policyName)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] ELB Classic LB Cookie Stickiness Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ELB Classic LB Cookie Stickiness Policy (%s): %s", d.Id(), err)
	}

	if len(policy.PolicyAttributeDescriptions) != 1 || aws.StringValue(policy.PolicyAttributeDescriptions[0].AttributeName) != "CookieExpirationPeriod" {
		return sdkdiag.AppendErrorf(diags, "cookie expiration period not found")
	}
	if v, err := strconv.Atoi(aws.StringValue(policy.PolicyAttributeDescriptions[0].AttributeValue)); err != nil {
		return sdkdiag.AppendErrorf(diags, "parsing cookie expiration period: %s", err)
	} else {
		d.Set("cookie_expiration_period", v)
	}
	d.Set("lb_port", lbPort)
	d.Set("load_balancer", lbName)
	d.Set("name", policyName)

	return diags
}

func resourceCookieStickinessPolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ELBConn(ctx)

	lbName, lbPort, policyName, err := LBCookieStickinessPolicyParseResourceID(d.Id())

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
		return sdkdiag.AppendErrorf(diags, "setting ELB Classic LB Cookie Stickiness Policy (%s): %s", d.Id(), err)
	}

	log.Printf("[DEBUG] Deleting ELB Classic LB Cookie Stickiness Policy: %s", d.Id())
	_, err = conn.DeleteLoadBalancerPolicyWithContext(ctx, &elb.DeleteLoadBalancerPolicyInput{
		LoadBalancerName: aws.String(lbName),
		PolicyName:       aws.String(policyName),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting ELB Classic LB Cookie Stickiness Policy (%s): %s", d.Id(), err)
	}

	return diags
}

const lbCookieStickinessPolicyResourceIDSeparator = ":"

func LBCookieStickinessPolicyCreateResourceID(lbName string, lbPort int, policyName string) string {
	parts := []string{lbName, strconv.Itoa(lbPort), policyName}
	id := strings.Join(parts, lbCookieStickinessPolicyResourceIDSeparator)

	return id
}

func LBCookieStickinessPolicyParseResourceID(id string) (string, int, string, error) {
	parts := strings.Split(id, lbCookieStickinessPolicyResourceIDSeparator)

	if len(parts) == 3 && parts[0] != "" && parts[1] != "" && parts[2] != "" {
		v, err := strconv.Atoi(parts[1])

		if err != nil {
			return "", 0, "", err
		}

		return parts[0], v, parts[2], nil
	}

	return "", 0, "", fmt.Errorf("unexpected format for ID (%[1]s), expected LBNAME%[2]sLBPORT%[2]sPOLICYNAME", id, lbCookieStickinessPolicyResourceIDSeparator)
}
