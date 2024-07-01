// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elb

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing"
	awstypes "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

// @SDKResource("aws_proxy_protocol_policy")
func ResourceProxyProtocolPolicy() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceProxyProtocolPolicyCreate,
		ReadWithoutTimeout:   resourceProxyProtocolPolicyRead,
		UpdateWithoutTimeout: resourceProxyProtocolPolicyUpdate,
		DeleteWithoutTimeout: resourceProxyProtocolPolicyDelete,

		Schema: map[string]*schema.Schema{
			"load_balancer": {
				Type:     schema.TypeString,
				Required: true,
			},

			"instance_ports": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Required: true,
				Set:      schema.HashString,
			},
		},
	}
}

func resourceProxyProtocolPolicyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ELBClient(ctx)
	elbname := aws.String(d.Get("load_balancer").(string))

	input := &elasticloadbalancing.CreateLoadBalancerPolicyInput{
		LoadBalancerName: elbname,
		PolicyAttributes: []awstypes.PolicyAttribute{
			{
				AttributeName:  aws.String("ProxyProtocol"),
				AttributeValue: aws.String("True"),
			},
		},
		PolicyName:     aws.String("TFEnableProxyProtocol"),
		PolicyTypeName: aws.String("ProxyProtocolPolicyType"),
	}

	// Create a policy
	log.Printf("[DEBUG] ELB create a policy %s from policy type %s",
		*input.PolicyName, *input.PolicyTypeName)

	if _, err := conn.CreateLoadBalancerPolicy(ctx, input); err != nil {
		return sdkdiag.AppendErrorf(diags, "creating a policy %s: %s", aws.ToString(input.PolicyName), err)
	}

	d.SetId(fmt.Sprintf("%s:%s", *elbname, *input.PolicyName))

	return append(diags, resourceProxyProtocolPolicyUpdate(ctx, d, meta)...)
}

func resourceProxyProtocolPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ELBClient(ctx)
	elbname := d.Get("load_balancer").(string)

	// Retrieve the current ELB policies for updating the state
	req := &elasticloadbalancing.DescribeLoadBalancersInput{
		LoadBalancerNames: []string{elbname},
	}
	resp, err := conn.DescribeLoadBalancers(ctx, req)
	if err != nil {
		if !d.IsNewResource() && errs.IsA[*awstypes.AccessPointNotFoundException](err) {
			log.Printf("[WARN] ELB Classic Proxy Protocol Policy (%s) not found, removing from state", d.Id())
			d.SetId("")
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "retrieving ELB attributes: %s", err)
	}

	backends := flattenBackendPolicies(resp.LoadBalancerDescriptions[0].BackendServerDescriptions)

	ports := []*string{}
	for ip := range backends {
		ipstr := strconv.Itoa(int(ip))
		ports = append(ports, &ipstr)
	}
	d.Set("instance_ports", ports)
	d.Set("load_balancer", elbname)
	return diags
}

func resourceProxyProtocolPolicyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ELBClient(ctx)
	elbname := aws.String(d.Get("load_balancer").(string))

	// Retrieve the current ELB policies for updating the state
	req := &elasticloadbalancing.DescribeLoadBalancersInput{
		LoadBalancerNames: []string{*elbname},
	}
	resp, err := conn.DescribeLoadBalancers(ctx, req)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "retrieving ELB attributes: %s", err)
	}

	backends := flattenBackendPolicies(resp.LoadBalancerDescriptions[0].BackendServerDescriptions)
	policyName := resourceProxyProtocolPolicyParseID(d.Id())

	if d.HasChange("instance_ports") {
		o, n := d.GetChange("instance_ports")
		os := o.(*schema.Set)
		ns := n.(*schema.Set)
		remove := os.Difference(ns).List()
		add := ns.Difference(os).List()

		inputs := []*elasticloadbalancing.SetLoadBalancerPoliciesForBackendServerInput{}

		i, err := resourceProxyProtocolPolicyRemove(policyName, remove, backends)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating ELB Classic Proxy Protocol Policy (%s): %s", d.Id(), err)
		}
		inputs = append(inputs, i...)

		i, err = resourceProxyProtocolPolicyAdd(policyName, add, backends)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating ELB Classic Proxy Protocol Policy (%s): %s", d.Id(), err)
		}
		inputs = append(inputs, i...)

		for _, input := range inputs {
			input.LoadBalancerName = elbname
			if _, err := conn.SetLoadBalancerPoliciesForBackendServer(ctx, input); err != nil {
				return sdkdiag.AppendErrorf(diags, "setting policy for backend: %s", err)
			}
		}
	}

	return append(diags, resourceProxyProtocolPolicyRead(ctx, d, meta)...)
}

func resourceProxyProtocolPolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ELBClient(ctx)
	elbname := aws.String(d.Get("load_balancer").(string))

	// Retrieve the current ELB policies for updating the state
	req := &elasticloadbalancing.DescribeLoadBalancersInput{
		LoadBalancerNames: []string{*elbname},
	}
	resp, err := conn.DescribeLoadBalancers(ctx, req)
	if err != nil {
		if errs.IsA[*awstypes.AccessPointNotFoundException](err) {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "retrieving ELB attributes: %s", err)
	}

	backends := flattenBackendPolicies(resp.LoadBalancerDescriptions[0].BackendServerDescriptions)
	ports := d.Get("instance_ports").(*schema.Set).List()
	policyName := resourceProxyProtocolPolicyParseID(d.Id())

	inputs, err := resourceProxyProtocolPolicyRemove(policyName, ports, backends)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting ELB Classic Proxy Protocol Policy (%s): %s", d.Id(), err)
	}
	for _, input := range inputs {
		input.LoadBalancerName = elbname
		if _, err := conn.SetLoadBalancerPoliciesForBackendServer(ctx, input); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting policy for backend: %s", err)
		}
	}

	pOpt := &elasticloadbalancing.DeleteLoadBalancerPolicyInput{
		LoadBalancerName: elbname,
		PolicyName:       aws.String(policyName),
	}
	if _, err := conn.DeleteLoadBalancerPolicy(ctx, pOpt); err != nil {
		return sdkdiag.AppendErrorf(diags, "removing a policy from load balancer: %s", err)
	}

	return diags
}

func resourceProxyProtocolPolicyRemove(policyName string, ports []interface{}, backends map[int32][]string) ([]*elasticloadbalancing.SetLoadBalancerPoliciesForBackendServerInput, error) {
	inputs := make([]*elasticloadbalancing.SetLoadBalancerPoliciesForBackendServerInput, 0, len(ports))
	for _, p := range ports {
		ip, err := strconv.ParseInt(p.(string), 10, 32)
		ip32 := int32(ip)
		if err != nil {
			return nil, fmt.Errorf("detaching the policy: %s", err)
		}

		newPolicies := []string{}
		curPolicies, found := backends[ip32]
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

		inputs = append(inputs, &elasticloadbalancing.SetLoadBalancerPoliciesForBackendServerInput{
			InstancePort: &ip32,
			PolicyNames:  newPolicies,
		})
	}
	return inputs, nil
}

func resourceProxyProtocolPolicyAdd(policyName string, ports []interface{}, backends map[int32][]string) ([]*elasticloadbalancing.SetLoadBalancerPoliciesForBackendServerInput, error) {
	inputs := make([]*elasticloadbalancing.SetLoadBalancerPoliciesForBackendServerInput, 0, len(ports))
	for _, p := range ports {
		ip, err := strconv.ParseInt(p.(string), 10, 64)
		ip32 := int32(ip)

		if err != nil {
			return nil, fmt.Errorf("attaching the policy: %s", err)
		}

		newPolicies := []string{}
		curPolicies := backends[ip32]
		for _, p := range curPolicies {
			if p == policyName {
				// Just remove it for now. It will be back later.
				continue
			}
			newPolicies = append(newPolicies, p)
		}
		newPolicies = append(newPolicies, policyName)

		inputs = append(inputs, &elasticloadbalancing.SetLoadBalancerPoliciesForBackendServerInput{
			InstancePort: &ip32,
			PolicyNames:  newPolicies,
		})
	}
	return inputs, nil
}

// resourceProxyProtocolPolicyParseID takes an ID and parses it into
// it's constituent parts. You need two axes (LB name, policy name)
// to create or identify a proxy protocol policy in AWS's API.
func resourceProxyProtocolPolicyParseID(id string) string {
	parts := strings.SplitN(id, ":", 2)
	// We currently omit the ELB name as it is not currently used anywhere
	return parts[1]
}
