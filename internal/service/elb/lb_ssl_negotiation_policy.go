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
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKResource("aws_lb_ssl_negotiation_policy")
func ResourceSSLNegotiationPolicy() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceSSLNegotiationPolicyCreate,
		ReadWithoutTimeout:   resourceSSLNegotiationPolicyRead,
		DeleteWithoutTimeout: resourceSSLNegotiationPolicyDelete,

		Schema: map[string]*schema.Schema{
			"attribute": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"value": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
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
			"triggers": {
				Type:     schema.TypeMap,
				Optional: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func resourceSSLNegotiationPolicyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ELBConn(ctx)

	lbName := d.Get("load_balancer").(string)
	lbPort := d.Get("lb_port").(int)
	policyName := d.Get("name").(string)
	id := SSLNegotiationPolicyCreateResourceID(lbName, lbPort, policyName)

	{
		input := &elb.CreateLoadBalancerPolicyInput{
			LoadBalancerName: aws.String(lbName),
			PolicyName:       aws.String(policyName),
			PolicyTypeName:   aws.String("SSLNegotiationPolicyType"),
		}

		if v, ok := d.GetOk("attribute"); ok && v.(*schema.Set).Len() > 0 {
			input.PolicyAttributes = ExpandPolicyAttributes(v.(*schema.Set).List())
		}

		_, err := conn.CreateLoadBalancerPolicyWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "creating ELB Classic SSL Negotiation Policy (%s): %s", id, err)
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
			return sdkdiag.AppendErrorf(diags, "setting ELB Classic SSL Negotiation Policy (%s): %s", id, err)
		}
	}

	d.SetId(id)

	return append(diags, resourceSSLNegotiationPolicyRead(ctx, d, meta)...)
}

func resourceSSLNegotiationPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ELBConn(ctx)

	lbName, lbPort, policyName, err := SSLNegotiationPolicyParseResourceID(d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "parsing resource ID: %s", err)
	}

	_, err = FindLoadBalancerListenerPolicyByThreePartKey(ctx, conn, lbName, lbPort, policyName)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] ELB Classic SSL Negotiation Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ELB Classic SSL Negotiation Policy (%s): %s", d.Id(), err)
	}

	d.Set("lb_port", lbPort)
	d.Set("load_balancer", lbName)
	d.Set("name", policyName)

	// TODO: fix attribute
	// This was previously erroneously setting "attributes", however this cannot
	// be changed without introducing problematic side effects. The ELB service
	// automatically expands the results to include all SSL attributes
	// (unordered, so we'd need to switch to TypeSet anyways), which we would be
	// quite impractical to force practitioners to write out and potentially
	// update each time the API updates since there is nearly 100 attributes.

	// We can get away with this because there's only one policy returned
	// policyDesc := getResp.PolicyDescriptions[0]
	// attributes := FlattenPolicyAttributes(policyDesc.PolicyAttributeDescriptions)
	// if err := d.Set("attribute", attributes); err != nil {
	// 	return fmt.Errorf("setting attribute: %s", err)
	// }

	return diags
}

func resourceSSLNegotiationPolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ELBConn(ctx)

	lbName, lbPort, policyName, err := SSLNegotiationPolicyParseResourceID(d.Id())

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
		return sdkdiag.AppendErrorf(diags, "setting ELB Classic SSL Negotiation Policy (%s): %s", d.Id(), err)
	}

	_, err = conn.DeleteLoadBalancerPolicyWithContext(ctx, &elb.DeleteLoadBalancerPolicyInput{
		LoadBalancerName: aws.String(lbName),
		PolicyName:       aws.String(policyName),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting ELB Classic SSL Negotiation Policy (%s): %s", d.Id(), err)
	}

	return diags
}

const sslNegotiationPolicyResourceIDSeparator = ":"

func SSLNegotiationPolicyCreateResourceID(lbName string, lbPort int, policyName string) string {
	parts := []string{lbName, strconv.Itoa(lbPort), policyName}
	id := strings.Join(parts, sslNegotiationPolicyResourceIDSeparator)

	return id
}

func SSLNegotiationPolicyParseResourceID(id string) (string, int, string, error) {
	parts := strings.Split(id, sslNegotiationPolicyResourceIDSeparator)

	if len(parts) == 3 && parts[0] != "" && parts[1] != "" && parts[2] != "" {
		v, err := strconv.Atoi(parts[1])

		if err != nil {
			return "", 0, "", err
		}

		return parts[0], v, parts[2], nil
	}

	return "", 0, "", fmt.Errorf("unexpected format for ID (%[1]s), expected LBNAME%[2]sLBPORT%[2]sPOLICYNAME", id, sslNegotiationPolicyResourceIDSeparator)
}
