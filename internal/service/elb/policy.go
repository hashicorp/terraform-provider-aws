// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elb

import (
	"context"
	"fmt"
	"log"
	"slices"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing"
	awstypes "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_load_balancer_policy", name="Load Balancer Policy")
func resourcePolicy() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourcePolicyCreate,
		ReadWithoutTimeout:   resourcePolicyRead,
		UpdateWithoutTimeout: resourcePolicyUpdate,
		DeleteWithoutTimeout: resourcePolicyDelete,

		Schema: map[string]*schema.Schema{
			"load_balancer_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"policy_attribute": {
				Type:     schema.TypeSet,
				Optional: true,
				// If policy_attribute(s) are not specified,
				// default values per policy type (see https://awscli.amazonaws.com/v2/documentation/api/latest/reference/elb/describe-load-balancer-policies.html)
				// will be returned by the API; thus, this TypeSet is marked as Computed.
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrName: {
							Type:     schema.TypeString,
							Optional: true,
						},
						names.AttrValue: {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
				// For policy types like "SSLNegotiationPolicyType" that can reference predefined policies
				// via the "Reference-Security-Policy" policy_attribute (https://docs.aws.amazon.com/elasticloadbalancing/latest/classic/elb-security-policy-table.html),
				// differences caused by additional attributes returned by the API are suppressed.
				DiffSuppressFunc: suppressPolicyAttributeDiffs,
			},
			"policy_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"policy_type_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourcePolicyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ELBClient(ctx)

	lbName := d.Get("load_balancer_name").(string)
	policyName := d.Get("policy_name").(string)
	id := policyCreateResourceID(lbName, policyName)
	input := &elasticloadbalancing.CreateLoadBalancerPolicyInput{
		LoadBalancerName: aws.String(lbName),
		PolicyName:       aws.String(policyName),
		PolicyTypeName:   aws.String(d.Get("policy_type_name").(string)),
	}

	if v, ok := d.GetOk("policy_attribute"); ok && v.(*schema.Set).Len() > 0 {
		input.PolicyAttributes = expandPolicyAttributes(v.(*schema.Set).List())
	}

	_, err := conn.CreateLoadBalancerPolicy(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating ELB Classic Load Balancer Policy (%s): %s", id, err)
	}

	d.SetId(id)

	return append(diags, resourcePolicyRead(ctx, d, meta)...)
}

func resourcePolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ELBClient(ctx)

	lbName, policyName, err := policyParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	policy, err := findLoadBalancerPolicyByTwoPartKey(ctx, conn, lbName, policyName)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] ELB Classic Load Balancer Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ELB Classic Load Balancer Policy (%s): %s", d.Id(), err)
	}

	d.Set("load_balancer_name", lbName)
	if err := d.Set("policy_attribute", flattenPolicyAttributeDescriptions(policy.PolicyAttributeDescriptions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting policy_attribute: %s", err)
	}
	d.Set("policy_name", policyName)
	d.Set("policy_type_name", policy.PolicyTypeName)

	return diags
}

func resourcePolicyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ELBClient(ctx)

	lbName, policyName, err := policyParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	reassignments := &policyReassignments{}

	err = findPolicyAttachmentByTwoPartKey(ctx, conn, lbName, policyName)
	switch {
	case tfresource.NotFound(err):
		// Policy not attached.
	case err != nil:
		return sdkdiag.AppendErrorf(diags, "reading ELB Classic Load Balancer Policy Attachment (%s/%s): %s", lbName, policyName, err)
	default:
		reassignments, err = unassignPolicy(ctx, conn, lbName, policyName)

		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	input := &elasticloadbalancing.DeleteLoadBalancerPolicyInput{
		LoadBalancerName: aws.String(lbName),
		PolicyName:       aws.String(policyName),
	}

	_, err = conn.DeleteLoadBalancerPolicy(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting ELB Classic Load Balancer Policy (%s): %s", d.Id(), err)
	}

	diags = append(diags, sdkdiag.WrapDiagsf(resourcePolicyCreate(ctx, d, meta), "updating ELB Classic Policy (%s)", d.Id())...)

	if diags.HasError() {
		return diags
	}

	for _, input := range reassignments.listenerPolicies {
		_, err := conn.SetLoadBalancerPoliciesOfListener(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "setting ELB Classic Listener Policy (%s): %s", lbName, err)
		}
	}

	for _, input := range reassignments.backendServerPolicies {
		_, err := conn.SetLoadBalancerPoliciesForBackendServer(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "setting ELB Classic Backend Server Policy (%s): %s", lbName, err)
		}
	}

	return append(diags, resourcePolicyRead(ctx, d, meta)...)
}

func resourcePolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ELBClient(ctx)

	lbName, policyName, err := policyParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	err = findPolicyAttachmentByTwoPartKey(ctx, conn, lbName, policyName)
	switch {
	case tfresource.NotFound(err):
		// Policy not attached.
	case err != nil:
		return sdkdiag.AppendErrorf(diags, "reading ELB Classic Load Balancer Policy Attachment (%s/%s): %s", lbName, policyName, err)
	default:
		if _, err := unassignPolicy(ctx, conn, lbName, policyName); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	log.Printf("[DEBUG] Deleting ELB Classic Load Balancer Policy: %s", d.Id())
	_, err = conn.DeleteLoadBalancerPolicy(ctx, &elasticloadbalancing.DeleteLoadBalancerPolicyInput{
		LoadBalancerName: aws.String(lbName),
		PolicyName:       aws.String(policyName),
	})

	if tfawserr.ErrCodeEquals(err, errCodeLoadBalancerNotFound) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting ELB Classic Load Balancer Policy (%s): %s", d.Id(), err)
	}

	return diags
}

func findPolicyAttachmentByTwoPartKey(ctx context.Context, conn *elasticloadbalancing.Client, lbName, policyName string) error {
	lb, err := findLoadBalancerByName(ctx, conn, lbName)

	if err != nil {
		return err
	}

	attached := slices.ContainsFunc(lb.BackendServerDescriptions, func(v awstypes.BackendServerDescription) bool {
		return slices.Contains(v.PolicyNames, policyName)
	})

	if attached {
		return nil
	}

	attached = slices.ContainsFunc(lb.ListenerDescriptions, func(v awstypes.ListenerDescription) bool {
		return slices.Contains(v.PolicyNames, policyName)
	})

	if attached {
		return nil
	}

	return &retry.NotFoundError{}
}

type policyReassignments struct {
	backendServerPolicies []*elasticloadbalancing.SetLoadBalancerPoliciesForBackendServerInput
	listenerPolicies      []*elasticloadbalancing.SetLoadBalancerPoliciesOfListenerInput
}

func unassignPolicy(ctx context.Context, conn *elasticloadbalancing.Client, lbName, policyName string) (*policyReassignments, error) {
	reassignments := &policyReassignments{}

	lb, err := findLoadBalancerByName(ctx, conn, lbName)

	if tfresource.NotFound(err) {
		return reassignments, nil
	}

	if err != nil {
		return nil, err
	}

	for _, v := range lb.BackendServerDescriptions {
		policies := tfslices.Filter(v.PolicyNames, func(v string) bool {
			return v != policyName
		})

		if len(v.PolicyNames) != len(policies) {
			reassignments.backendServerPolicies = append(reassignments.backendServerPolicies, &elasticloadbalancing.SetLoadBalancerPoliciesForBackendServerInput{
				InstancePort:     v.InstancePort,
				LoadBalancerName: aws.String(lbName),
				PolicyNames:      v.PolicyNames,
			})

			input := &elasticloadbalancing.SetLoadBalancerPoliciesForBackendServerInput{
				InstancePort:     v.InstancePort,
				LoadBalancerName: aws.String(lbName),
				PolicyNames:      policies,
			}

			_, err = conn.SetLoadBalancerPoliciesForBackendServer(ctx, input)

			if err != nil {
				return nil, fmt.Errorf("setting ELB Classic Backend Server Policy (%s): %w", lbName, err)
			}
		}
	}

	for _, v := range lb.ListenerDescriptions {
		policies := tfslices.Filter(v.PolicyNames, func(v string) bool {
			return v != policyName
		})

		if len(v.PolicyNames) != len(policies) {
			reassignments.listenerPolicies = append(reassignments.listenerPolicies, &elasticloadbalancing.SetLoadBalancerPoliciesOfListenerInput{
				LoadBalancerName: aws.String(lbName),
				LoadBalancerPort: v.Listener.LoadBalancerPort,
				PolicyNames:      v.PolicyNames,
			})

			input := &elasticloadbalancing.SetLoadBalancerPoliciesOfListenerInput{
				LoadBalancerName: aws.String(lbName),
				LoadBalancerPort: v.Listener.LoadBalancerPort,
				PolicyNames:      policies,
			}

			_, err = conn.SetLoadBalancerPoliciesOfListener(ctx, input)

			if err != nil {
				return reassignments, fmt.Errorf("setting ELB Classic Listener Policy (%s): %w", lbName, err)
			}
		}
	}

	return reassignments, nil
}

func suppressPolicyAttributeDiffs(k, old, new string, d *schema.ResourceData) bool {
	// Show difference for new resource
	if d.Id() == "" {
		return false
	}

	// Show differences if configured attributes are not in state
	if old == "0" && new != "0" {
		return false
	}

	o, n := d.GetChange("policy_attribute")
	oldAttributes := o.(*schema.Set)
	newAttributes := n.(*schema.Set)

	// Suppress differences if the attributes returned from the API contain those configured
	return oldAttributes.Intersection(newAttributes).Len() == newAttributes.Len()
}

const policyResourceIDSeparator = ":"

func policyCreateResourceID(lbName, policyName string) string {
	parts := []string{lbName, policyName}
	id := strings.Join(parts, policyResourceIDSeparator)

	return id
}

func policyParseResourceID(id string) (string, string, error) {
	parts := strings.Split(id, backendServerPolicyResourceIDSeparator)

	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected LBNAME%[2]sPOLICYNAME", id, policyResourceIDSeparator)
}
