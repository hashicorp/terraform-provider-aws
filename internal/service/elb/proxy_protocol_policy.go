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
	awstypes "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tfmaps "github.com/hashicorp/terraform-provider-aws/internal/maps"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKResource("aws_proxy_protocol_policy", name="Proxy Protocol Policy")
func resourceProxyProtocolPolicy() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceProxyProtocolPolicyCreate,
		ReadWithoutTimeout:   resourceProxyProtocolPolicyRead,
		UpdateWithoutTimeout: resourceProxyProtocolPolicyUpdate,
		DeleteWithoutTimeout: resourceProxyProtocolPolicyDelete,

		Schema: map[string]*schema.Schema{
			"instance_ports": {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: verify.StringIsInt32,
				},
			},
			"load_balancer": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceProxyProtocolPolicyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ELBClient(ctx)

	lbName := d.Get("load_balancer").(string)
	input := &elasticloadbalancing.CreateLoadBalancerPolicyInput{
		LoadBalancerName: aws.String(lbName),
		PolicyAttributes: []awstypes.PolicyAttribute{
			{
				AttributeName:  aws.String("ProxyProtocol"),
				AttributeValue: aws.String("True"),
			},
		},
		PolicyName:     aws.String("TFEnableProxyProtocol"),
		PolicyTypeName: aws.String("ProxyProtocolPolicyType"),
	}

	_, err := conn.CreateLoadBalancerPolicy(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating ELB Classic Proxy Protocol Policy (%s): %s", lbName, err)
	}

	d.SetId(proxyProtocolPolicyCreateResourceID(lbName, aws.ToString(input.PolicyName)))

	return append(diags, resourceProxyProtocolPolicyUpdate(ctx, d, meta)...)
}

func resourceProxyProtocolPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ELBClient(ctx)

	lbName, _, err := proxyProtocolPolicyParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	lb, err := findLoadBalancerByName(ctx, conn, lbName)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] ELB Classic Proxy Protocol Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ELB Classic Load Balancer (%s): %s", lbName, err)
	}

	ports := tfslices.ApplyToAll(tfmaps.Keys(flattenBackendServerDescriptionPolicies(lb.BackendServerDescriptions)), flex.Int32ValueToStringValue)
	d.Set("instance_ports", ports)
	d.Set("load_balancer", lbName)

	return diags
}

func resourceProxyProtocolPolicyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ELBClient(ctx)

	lbName, policyName, err := proxyProtocolPolicyParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	lb, err := findLoadBalancerByName(ctx, conn, lbName)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ELB Classic Load Balancer (%s): %s", lbName, err)
	}

	backendPolicies := flattenBackendServerDescriptionPolicies(lb.BackendServerDescriptions)

	if d.HasChange("instance_ports") {
		o, n := d.GetChange("instance_ports")
		os, ns := o.(*schema.Set), n.(*schema.Set)
		add, del := ns.Difference(os), os.Difference(ns)

		var inputs []*elasticloadbalancing.SetLoadBalancerPoliciesForBackendServerInput
		inputs = append(inputs, expandRemoveProxyProtocolPolicyInputs(policyName, flex.ExpandStringValueSet(del), backendPolicies)...)
		inputs = append(inputs, expandAddProxyProtocolPolicyInputs(policyName, flex.ExpandStringValueSet(add), backendPolicies)...)

		for _, input := range inputs {
			input.LoadBalancerName = aws.String(lbName)

			_, err := conn.SetLoadBalancerPoliciesForBackendServer(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "setting ELB Classic Backend Server Policy (%s): %s", lbName, err)
			}
		}
	}

	return append(diags, resourceProxyProtocolPolicyRead(ctx, d, meta)...)
}

func resourceProxyProtocolPolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ELBClient(ctx)

	lbName, policyName, err := proxyProtocolPolicyParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	lb, err := findLoadBalancerByName(ctx, conn, lbName)

	if tfresource.NotFound(err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ELB Classic Load Balancer (%s): %s", lbName, err)
	}

	backendPolicies := flattenBackendServerDescriptionPolicies(lb.BackendServerDescriptions)
	ports := flex.ExpandStringValueSet(d.Get("instance_ports").(*schema.Set))

	for _, input := range expandRemoveProxyProtocolPolicyInputs(policyName, ports, backendPolicies) {
		input.LoadBalancerName = aws.String(lbName)

		_, err := conn.SetLoadBalancerPoliciesForBackendServer(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "setting ELB Classic Backend Server Policy (%s): %s", lbName, err)
		}
	}

	_, err = conn.DeleteLoadBalancerPolicy(ctx, &elasticloadbalancing.DeleteLoadBalancerPolicyInput{
		LoadBalancerName: aws.String(lbName),
		PolicyName:       aws.String(policyName),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting ELB Classic Proxy Protocol Policy (%s): %s", lbName, err)
	}

	return diags
}

func expandAddProxyProtocolPolicyInputs(policyName string, ports []string, backendPolicies map[int32][]string) []*elasticloadbalancing.SetLoadBalancerPoliciesForBackendServerInput {
	apiObjects := make([]*elasticloadbalancing.SetLoadBalancerPoliciesForBackendServerInput, 0, len(ports))

	for _, p := range ports {
		port := flex.StringValueToInt32Value(p)

		newPolicies := []string{}
		curPolicies := backendPolicies[port]
		for _, p := range curPolicies {
			if p == policyName {
				// Just remove it for now. It will be back later.
				continue
			}

			newPolicies = append(newPolicies, p)
		}
		newPolicies = append(newPolicies, policyName)

		apiObjects = append(apiObjects, &elasticloadbalancing.SetLoadBalancerPoliciesForBackendServerInput{
			InstancePort: aws.Int32(port),
			PolicyNames:  newPolicies,
		})
	}
	return apiObjects
}

func expandRemoveProxyProtocolPolicyInputs(policyName string, ports []string, backendPolicies map[int32][]string) []*elasticloadbalancing.SetLoadBalancerPoliciesForBackendServerInput {
	apiObjects := make([]*elasticloadbalancing.SetLoadBalancerPoliciesForBackendServerInput, 0, len(ports))

	for _, p := range ports {
		port := flex.StringValueToInt32Value(p)

		newPolicies := []string{}
		curPolicies, found := backendPolicies[port]
		if !found {
			// No policy for this instance port found, just skip it.
			continue
		}

		for _, p := range curPolicies {
			if p == policyName {
				// remove the policy
				continue
			}
			newPolicies = append(newPolicies, p)
		}

		apiObjects = append(apiObjects, &elasticloadbalancing.SetLoadBalancerPoliciesForBackendServerInput{
			InstancePort: aws.Int32(port),
			PolicyNames:  newPolicies,
		})
	}

	return apiObjects
}

const proxyProtocolPolicyResourceIDSeparator = ":"

func proxyProtocolPolicyCreateResourceID(lbName, policyName string) string {
	parts := []string{lbName, policyName}
	id := strings.Join(parts, proxyProtocolPolicyResourceIDSeparator)

	return id
}

func proxyProtocolPolicyParseResourceID(id string) (string, string, error) {
	parts := strings.SplitN(id, proxyProtocolPolicyResourceIDSeparator, 2)

	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected LBNAME%[2]sPOLICYNAME", id, proxyProtocolPolicyResourceIDSeparator)
}
