package elb

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func ResourcePolicy() *schema.Resource {
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
						"name": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"value": {
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
	conn := meta.(*conns.AWSClient).ELBConn()

	lbName := d.Get("load_balancer_name").(string)
	policyName := d.Get("policy_name").(string)
	id := PolicyCreateResourceID(lbName, policyName)
	input := &elb.CreateLoadBalancerPolicyInput{
		LoadBalancerName: aws.String(lbName),
		PolicyName:       aws.String(policyName),
		PolicyTypeName:   aws.String(d.Get("policy_type_name").(string)),
	}

	if v, ok := d.GetOk("policy_attribute"); ok && v.(*schema.Set).Len() > 0 {
		input.PolicyAttributes = ExpandPolicyAttributes(v.(*schema.Set).List())
	}

	_, err := conn.CreateLoadBalancerPolicyWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating ELB Classic Load Balancer Policy (%s): %s", id, err)
	}

	d.SetId(id)

	return append(diags, resourcePolicyRead(ctx, d, meta)...)
}

func resourcePolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ELBConn()

	lbName, policyName, err := PolicyParseResourceID(d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "parsing resource ID: %s", err)
	}

	policy, err := FindLoadBalancerPolicyByTwoPartKey(ctx, conn, lbName, policyName)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] ELB Classic Load Balancer Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ELB Classic Load Balancer Policy (%s): %s", d.Id(), err)
	}

	d.Set("load_balancer_name", lbName)
	if err := d.Set("policy_attribute", FlattenPolicyAttributes(policy.PolicyAttributeDescriptions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting policy_attribute: %s", err)
	}
	d.Set("policy_name", policyName)
	d.Set("policy_type_name", policy.PolicyTypeName)

	return diags
}

func resourcePolicyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ELBConn()
	reassignments := Reassignment{}

	lbName, policyName, err := PolicyParseResourceID(d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "parsing resource ID: %s", err)
	}

	assigned, err := resourcePolicyAssigned(ctx, policyName, lbName, conn)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "determining assignment status of Load Balancer Policy %s: %s", policyName, err)
	}

	if assigned {
		reassignments, err = resourcePolicyUnassign(ctx, policyName, lbName, conn)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "unassigning Load Balancer Policy %s: %s", policyName, err)
		}
	}

	request := &elb.DeleteLoadBalancerPolicyInput{
		LoadBalancerName: aws.String(lbName),
		PolicyName:       aws.String(policyName),
	}

	if _, err := conn.DeleteLoadBalancerPolicyWithContext(ctx, request); err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Load Balancer Policy %s: %s", d.Id(), err)
	}

	diags = append(diags, sdkdiag.WrapDiagsf(resourcePolicyCreate(ctx, d, meta), "updating ELB Classic Policy (%s)", d.Id())...)
	if diags.HasError() {
		return diags
	}

	for _, listenerAssignment := range reassignments.listenerPolicies {
		if _, err := conn.SetLoadBalancerPoliciesOfListenerWithContext(ctx, listenerAssignment); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting LoadBalancerPoliciesOfListener: %s", err)
		}
	}

	for _, backendServerAssignment := range reassignments.backendServerPolicies {
		if _, err := conn.SetLoadBalancerPoliciesForBackendServerWithContext(ctx, backendServerAssignment); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting LoadBalancerPoliciesForBackendServer: %s", err)
		}
	}

	return append(diags, resourcePolicyRead(ctx, d, meta)...)
}

func resourcePolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ELBConn()

	lbName, policyName, err := PolicyParseResourceID(d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "parsing resource ID: %s", err)
	}

	assigned, err := resourcePolicyAssigned(ctx, policyName, lbName, conn)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "determining assignment status of Load Balancer Policy %s: %s", policyName, err)
	}

	if assigned {
		_, err := resourcePolicyUnassign(ctx, policyName, lbName, conn)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "unassigning Load Balancer Policy %s: %s", policyName, err)
		}
	}

	request := &elb.DeleteLoadBalancerPolicyInput{
		LoadBalancerName: aws.String(lbName),
		PolicyName:       aws.String(policyName),
	}

	if _, err := conn.DeleteLoadBalancerPolicyWithContext(ctx, request); err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting ELB Classic Load Balancer Policy (%s): %s", d.Id(), err)
	}

	return diags
}

func resourcePolicyAssigned(ctx context.Context, policyName, loadBalancerName string, conn *elb.ELB) (bool, error) {
	describeElbOpts := &elb.DescribeLoadBalancersInput{
		LoadBalancerNames: []*string{aws.String(loadBalancerName)},
	}

	describeResp, err := conn.DescribeLoadBalancersWithContext(ctx, describeElbOpts)

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
	for _, backendServer := range lb.BackendServerDescriptions {
		for _, name := range backendServer.PolicyNames {
			if policyName == aws.StringValue(name) {
				assigned = true
				break
			}
		}
	}

	for _, listener := range lb.ListenerDescriptions {
		for _, name := range listener.PolicyNames {
			if policyName == aws.StringValue(name) {
				assigned = true
				break
			}
		}
	}

	return assigned, nil
}

type Reassignment struct {
	backendServerPolicies []*elb.SetLoadBalancerPoliciesForBackendServerInput
	listenerPolicies      []*elb.SetLoadBalancerPoliciesOfListenerInput
}

func resourcePolicyUnassign(ctx context.Context, policyName, loadBalancerName string, conn *elb.ELB) (Reassignment, error) {
	reassignments := Reassignment{}

	describeElbOpts := &elb.DescribeLoadBalancersInput{
		LoadBalancerNames: []*string{aws.String(loadBalancerName)},
	}

	describeResp, err := conn.DescribeLoadBalancersWithContext(ctx, describeElbOpts)

	if tfawserr.ErrCodeEquals(err, elb.ErrCodeAccessPointNotFoundException) {
		return reassignments, nil
	}

	if err != nil {
		return reassignments, fmt.Errorf("Error retrieving ELB description: %s", err)
	}

	if len(describeResp.LoadBalancerDescriptions) != 1 {
		return reassignments, fmt.Errorf("Unable to find ELB: %#v", describeResp.LoadBalancerDescriptions)
	}

	lb := describeResp.LoadBalancerDescriptions[0]

	for _, backendServer := range lb.BackendServerDescriptions {
		policies := []*string{}

		for _, name := range backendServer.PolicyNames {
			if policyName != aws.StringValue(name) {
				policies = append(policies, name)
			}
		}

		if len(backendServer.PolicyNames) != len(policies) {
			setOpts := &elb.SetLoadBalancerPoliciesForBackendServerInput{
				LoadBalancerName: aws.String(loadBalancerName),
				InstancePort:     aws.Int64(*backendServer.InstancePort),
				PolicyNames:      policies,
			}

			reassignOpts := &elb.SetLoadBalancerPoliciesForBackendServerInput{
				LoadBalancerName: aws.String(loadBalancerName),
				InstancePort:     aws.Int64(*backendServer.InstancePort),
				PolicyNames:      backendServer.PolicyNames,
			}

			reassignments.backendServerPolicies = append(reassignments.backendServerPolicies, reassignOpts)

			_, err = conn.SetLoadBalancerPoliciesForBackendServerWithContext(ctx, setOpts)
			if err != nil {
				return reassignments, fmt.Errorf("Error Setting Load Balancer Policies for Backend Server: %s", err)
			}
		}
	}

	for _, listener := range lb.ListenerDescriptions {
		policies := []*string{}

		for _, name := range listener.PolicyNames {
			if policyName != aws.StringValue(name) {
				policies = append(policies, name)
			}
		}

		if len(listener.PolicyNames) != len(policies) {
			setOpts := &elb.SetLoadBalancerPoliciesOfListenerInput{
				LoadBalancerName: aws.String(loadBalancerName),
				LoadBalancerPort: aws.Int64(*listener.Listener.LoadBalancerPort),
				PolicyNames:      policies,
			}

			reassignOpts := &elb.SetLoadBalancerPoliciesOfListenerInput{
				LoadBalancerName: aws.String(loadBalancerName),
				LoadBalancerPort: aws.Int64(*listener.Listener.LoadBalancerPort),
				PolicyNames:      listener.PolicyNames,
			}

			reassignments.listenerPolicies = append(reassignments.listenerPolicies, reassignOpts)

			_, err = conn.SetLoadBalancerPoliciesOfListenerWithContext(ctx, setOpts)
			if err != nil {
				return reassignments, fmt.Errorf("Error Setting Load Balancer Policies of Listener: %s", err)
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

func PolicyCreateResourceID(lbName, policyName string) string {
	parts := []string{lbName, policyName}
	id := strings.Join(parts, policyResourceIDSeparator)

	return id
}

func PolicyParseResourceID(id string) (string, string, error) {
	parts := strings.Split(id, backendServerPolicyResourceIDSeparator)

	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected LBNAME%[2]sPOLICYNAME", id, policyResourceIDSeparator)
}
