// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elb

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_lb_ssl_negotiation_policy", name="SSL Negotiation Policy")
func resourceSSLNegotiationPolicy() *schema.Resource {
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
						names.AttrName: {
							Type:     schema.TypeString,
							Required: true,
						},
						names.AttrValue: {
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
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrTriggers: {
				Type:     schema.TypeMap,
				Optional: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func resourceSSLNegotiationPolicyCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ELBClient(ctx)

	lbName := d.Get("load_balancer").(string)
	lbPort := int32(d.Get("lb_port").(int))
	policyName := d.Get(names.AttrName).(string)
	id := sslNegotiationPolicyCreateResourceID(lbName, lbPort, policyName)

	{
		input := elasticloadbalancing.CreateLoadBalancerPolicyInput{
			LoadBalancerName: aws.String(lbName),
			PolicyName:       aws.String(policyName),
			PolicyTypeName:   aws.String("SSLNegotiationPolicyType"),
		}

		if v, ok := d.GetOk("attribute"); ok && v.(*schema.Set).Len() > 0 {
			input.PolicyAttributes = expandPolicyAttributes(v.(*schema.Set).List())
		}

		_, err := conn.CreateLoadBalancerPolicy(ctx, &input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "creating ELB Classic SSL Negotiation Policy (%s): %s", id, err)
		}
	}

	{
		input := elasticloadbalancing.SetLoadBalancerPoliciesOfListenerInput{
			LoadBalancerName: aws.String(lbName),
			LoadBalancerPort: lbPort,
			PolicyNames:      []string{policyName},
		}

		_, err := conn.SetLoadBalancerPoliciesOfListener(ctx, &input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "setting ELB Classic SSL Negotiation Policy (%s): %s", id, err)
		}
	}

	d.SetId(id)

	return append(diags, resourceSSLNegotiationPolicyRead(ctx, d, meta)...)
}

func resourceSSLNegotiationPolicyRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ELBClient(ctx)

	lbName, lbPort, policyName, err := sslNegotiationPolicyParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	_, err = findLoadBalancerListenerPolicyByThreePartKey(ctx, conn, lbName, lbPort, policyName)

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
	d.Set(names.AttrName, policyName)

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

func resourceSSLNegotiationPolicyDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ELBClient(ctx)

	lbName, lbPort, policyName, err := sslNegotiationPolicyParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	{
		// Perversely, if we Set an empty list of PolicyNames, we detach the
		// policies attached to a listener, which is required to delete the
		// policy itself.
		input := elasticloadbalancing.SetLoadBalancerPoliciesOfListenerInput{
			LoadBalancerName: aws.String(lbName),
			LoadBalancerPort: lbPort,
			PolicyNames:      []string{},
		}

		_, err = conn.SetLoadBalancerPoliciesOfListener(ctx, &input)

		if tfawserr.ErrCodeEquals(err, errCodeLoadBalancerNotFound) {
			return diags
		}

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "setting ELB Classic SSL Negotiation Policy (%s): %s", d.Id(), err)
		}
	}

	{
		log.Printf("[DEBUG] Deleting ELB Classic SSL Negotiation Policy: %s", d.Id())
		input := elasticloadbalancing.DeleteLoadBalancerPolicyInput{
			LoadBalancerName: aws.String(lbName),
			PolicyName:       aws.String(policyName),
		}
		_, err = conn.DeleteLoadBalancerPolicy(ctx, &input)

		if tfawserr.ErrCodeEquals(err, errCodeLoadBalancerNotFound) {
			return diags
		}

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "deleting ELB Classic SSL Negotiation Policy (%s): %s", d.Id(), err)
		}
	}

	return diags
}

const sslNegotiationPolicyResourceIDSeparator = ":"

func sslNegotiationPolicyCreateResourceID(lbName string, lbPort int32, policyName string) string {
	parts := []string{lbName, flex.Int32ValueToStringValue(lbPort), policyName}
	id := strings.Join(parts, sslNegotiationPolicyResourceIDSeparator)

	return id
}

func sslNegotiationPolicyParseResourceID(id string) (string, int32, string, error) {
	parts := strings.Split(id, sslNegotiationPolicyResourceIDSeparator)

	if len(parts) == 3 && parts[0] != "" && parts[1] != "" && parts[2] != "" {
		return parts[0], flex.StringValueToInt32Value(parts[1]), parts[2], nil
	}

	return "", 0, "", fmt.Errorf("unexpected format for ID (%[1]s), expected LBNAME%[2]sLBPORT%[2]sPOLICYNAME", id, sslNegotiationPolicyResourceIDSeparator)
}
