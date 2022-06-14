package elb

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func ResourcePolicy() *schema.Resource {
	return &schema.Resource{
		Create: resourcePolicyCreate,
		Read:   resourcePolicyRead,
		Update: resourcePolicyUpdate,
		Delete: resourcePolicyDelete,

		Schema: map[string]*schema.Schema{
			"load_balancer_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
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
		},
	}
}

func resourcePolicyCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ELBConn

	lbspOpts := &elb.CreateLoadBalancerPolicyInput{
		LoadBalancerName: aws.String(d.Get("load_balancer_name").(string)),
		PolicyName:       aws.String(d.Get("policy_name").(string)),
		PolicyTypeName:   aws.String(d.Get("policy_type_name").(string)),
	}

	if v, ok := d.GetOk("policy_attribute"); ok && v.(*schema.Set).Len() > 0 {
		lbspOpts.PolicyAttributes = ExpandPolicyAttributes(v.(*schema.Set).List())
	}

	if _, err := conn.CreateLoadBalancerPolicy(lbspOpts); err != nil {
		return fmt.Errorf("Error creating LoadBalancerPolicy: %s", err)
	}

	d.SetId(fmt.Sprintf("%s:%s",
		*lbspOpts.LoadBalancerName,
		*lbspOpts.PolicyName))
	return resourcePolicyRead(d, meta)
}

func resourcePolicyRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ELBConn

	loadBalancerName, policyName := PolicyParseID(d.Id())

	request := &elb.DescribeLoadBalancerPoliciesInput{
		LoadBalancerName: aws.String(loadBalancerName),
		PolicyNames:      []*string{aws.String(policyName)},
	}

	getResp, err := conn.DescribeLoadBalancerPolicies(request)

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, "LoadBalancerNotFound") {
		log.Printf("[WARN] Load Balancer (%s) not found, removing from state", loadBalancerName)
		d.SetId("")
		return nil
	}
	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, elb.ErrCodePolicyNotFoundException) {
		log.Printf("[WARN] Load Balancer Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("Error retrieving policy: %s", err)
	}

	if len(getResp.PolicyDescriptions) != 1 {
		return fmt.Errorf("Unable to find policy %#v", getResp.PolicyDescriptions)
	}

	policyDesc := getResp.PolicyDescriptions[0]
	policyTypeName := policyDesc.PolicyTypeName
	policyAttributes := policyDesc.PolicyAttributeDescriptions

	d.Set("policy_name", policyName)
	d.Set("policy_type_name", policyTypeName)
	d.Set("load_balancer_name", loadBalancerName)
	if err := d.Set("policy_attribute", FlattenPolicyAttributes(policyAttributes)); err != nil {
		return fmt.Errorf("error setting policy_attribute: %w", err)
	}

	return nil
}

func resourcePolicyUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ELBConn
	reassignments := Reassignment{}

	loadBalancerName, policyName := PolicyParseID(d.Id())

	assigned, err := resourcePolicyAssigned(policyName, loadBalancerName, conn)
	if err != nil {
		return fmt.Errorf("Error determining assignment status of Load Balancer Policy %s: %s", policyName, err)
	}

	if assigned {
		reassignments, err = resourcePolicyUnassign(policyName, loadBalancerName, conn)
		if err != nil {
			return fmt.Errorf("Error unassigning Load Balancer Policy %s: %s", policyName, err)
		}
	}

	request := &elb.DeleteLoadBalancerPolicyInput{
		LoadBalancerName: aws.String(loadBalancerName),
		PolicyName:       aws.String(policyName),
	}

	if _, err := conn.DeleteLoadBalancerPolicy(request); err != nil {
		return fmt.Errorf("Error deleting Load Balancer Policy %s: %s", d.Id(), err)
	}

	err = resourcePolicyCreate(d, meta)
	if err != nil {
		return err
	}

	for _, listenerAssignment := range reassignments.listenerPolicies {
		if _, err := conn.SetLoadBalancerPoliciesOfListener(listenerAssignment); err != nil {
			return fmt.Errorf("Error setting LoadBalancerPoliciesOfListener: %s", err)
		}
	}

	for _, backendServerAssignment := range reassignments.backendServerPolicies {
		if _, err := conn.SetLoadBalancerPoliciesForBackendServer(backendServerAssignment); err != nil {
			return fmt.Errorf("Error setting LoadBalancerPoliciesForBackendServer: %s", err)
		}
	}

	return resourcePolicyRead(d, meta)
}

func resourcePolicyDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ELBConn

	loadBalancerName, policyName := PolicyParseID(d.Id())

	assigned, err := resourcePolicyAssigned(policyName, loadBalancerName, conn)
	if err != nil {
		return fmt.Errorf("Error determining assignment status of Load Balancer Policy %s: %s", policyName, err)
	}

	if assigned {
		_, err := resourcePolicyUnassign(policyName, loadBalancerName, conn)
		if err != nil {
			return fmt.Errorf("Error unassigning Load Balancer Policy %s: %s", policyName, err)
		}
	}

	request := &elb.DeleteLoadBalancerPolicyInput{
		LoadBalancerName: aws.String(loadBalancerName),
		PolicyName:       aws.String(policyName),
	}

	if _, err := conn.DeleteLoadBalancerPolicy(request); err != nil {
		return fmt.Errorf("Error deleting Load Balancer Policy %s: %s", d.Id(), err)
	}

	return nil
}

func PolicyParseID(id string) (string, string) {
	parts := strings.SplitN(id, ":", 2)
	return parts[0], parts[1]
}

func resourcePolicyAssigned(policyName, loadBalancerName string, conn *elb.ELB) (bool, error) {
	describeElbOpts := &elb.DescribeLoadBalancersInput{
		LoadBalancerNames: []*string{aws.String(loadBalancerName)},
	}

	describeResp, err := conn.DescribeLoadBalancers(describeElbOpts)

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

func resourcePolicyUnassign(policyName, loadBalancerName string, conn *elb.ELB) (Reassignment, error) {
	reassignments := Reassignment{}

	describeElbOpts := &elb.DescribeLoadBalancersInput{
		LoadBalancerNames: []*string{aws.String(loadBalancerName)},
	}

	describeResp, err := conn.DescribeLoadBalancers(describeElbOpts)

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

			_, err = conn.SetLoadBalancerPoliciesForBackendServer(setOpts)
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

			_, err = conn.SetLoadBalancerPoliciesOfListener(setOpts)
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
