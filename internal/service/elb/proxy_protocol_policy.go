package elb

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func ResourceProxyProtocolPolicy() *schema.Resource {
	return &schema.Resource{
		Create: resourceProxyProtocolPolicyCreate,
		Read:   resourceProxyProtocolPolicyRead,
		Update: resourceProxyProtocolPolicyUpdate,
		Delete: resourceProxyProtocolPolicyDelete,

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

func resourceProxyProtocolPolicyCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ELBConn
	elbname := aws.String(d.Get("load_balancer").(string))

	input := &elb.CreateLoadBalancerPolicyInput{
		LoadBalancerName: elbname,
		PolicyAttributes: []*elb.PolicyAttribute{
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

	if _, err := conn.CreateLoadBalancerPolicy(input); err != nil {
		return fmt.Errorf("Error creating a policy %s: %s",
			*input.PolicyName, err)
	}

	d.SetId(fmt.Sprintf("%s:%s", *elbname, *input.PolicyName))
	log.Printf("[INFO] ELB PolicyName: %s", *input.PolicyName)

	return resourceProxyProtocolPolicyUpdate(d, meta)
}

func resourceProxyProtocolPolicyRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ELBConn
	elbname := d.Get("load_balancer").(string)

	// Retrieve the current ELB policies for updating the state
	req := &elb.DescribeLoadBalancersInput{
		LoadBalancerNames: []*string{aws.String(elbname)},
	}
	resp, err := conn.DescribeLoadBalancers(req)
	if err != nil {
		if IsNotFound(err) {
			// The ELB is gone now, so just remove it from the state
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error retrieving ELB attributes: %s", err)
	}

	backends := flattenBackendPolicies(resp.LoadBalancerDescriptions[0].BackendServerDescriptions)

	ports := []*string{}
	for ip := range backends {
		ipstr := strconv.Itoa(int(ip))
		ports = append(ports, &ipstr)
	}
	d.Set("instance_ports", ports)
	d.Set("load_balancer", elbname)
	return nil
}

func resourceProxyProtocolPolicyUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ELBConn
	elbname := aws.String(d.Get("load_balancer").(string))

	// Retrieve the current ELB policies for updating the state
	req := &elb.DescribeLoadBalancersInput{
		LoadBalancerNames: []*string{elbname},
	}
	resp, err := conn.DescribeLoadBalancers(req)
	if err != nil {
		if IsNotFound(err) {
			// The ELB is gone now, so just remove it from the state
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error retrieving ELB attributes: %s", err)
	}

	backends := flattenBackendPolicies(resp.LoadBalancerDescriptions[0].BackendServerDescriptions)
	policyName := resourceProxyProtocolPolicyParseID(d.Id())

	if d.HasChange("instance_ports") {
		o, n := d.GetChange("instance_ports")
		os := o.(*schema.Set)
		ns := n.(*schema.Set)
		remove := os.Difference(ns).List()
		add := ns.Difference(os).List()

		inputs := []*elb.SetLoadBalancerPoliciesForBackendServerInput{}

		i, err := resourceProxyProtocolPolicyRemove(policyName, remove, backends)
		if err != nil {
			return err
		}
		inputs = append(inputs, i...)

		i, err = resourceProxyProtocolPolicyAdd(policyName, add, backends)
		if err != nil {
			return err
		}
		inputs = append(inputs, i...)

		for _, input := range inputs {
			input.LoadBalancerName = elbname
			if _, err := conn.SetLoadBalancerPoliciesForBackendServer(input); err != nil {
				return fmt.Errorf("Error setting policy for backend: %s", err)
			}
		}
	}

	return resourceProxyProtocolPolicyRead(d, meta)
}

func resourceProxyProtocolPolicyDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ELBConn
	elbname := aws.String(d.Get("load_balancer").(string))

	// Retrieve the current ELB policies for updating the state
	req := &elb.DescribeLoadBalancersInput{
		LoadBalancerNames: []*string{elbname},
	}
	var err error
	resp, err := conn.DescribeLoadBalancers(req)
	if err != nil {
		if IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("Error retrieving ELB attributes: %s", err)
	}

	backends := flattenBackendPolicies(resp.LoadBalancerDescriptions[0].BackendServerDescriptions)
	ports := d.Get("instance_ports").(*schema.Set).List()
	policyName := resourceProxyProtocolPolicyParseID(d.Id())

	inputs, err := resourceProxyProtocolPolicyRemove(policyName, ports, backends)
	if err != nil {
		return fmt.Errorf("Error detaching a policy from backend: %s", err)
	}
	for _, input := range inputs {
		input.LoadBalancerName = elbname
		if _, err := conn.SetLoadBalancerPoliciesForBackendServer(input); err != nil {
			return fmt.Errorf("Error setting policy for backend: %s", err)
		}
	}

	pOpt := &elb.DeleteLoadBalancerPolicyInput{
		LoadBalancerName: elbname,
		PolicyName:       aws.String(policyName),
	}
	if _, err := conn.DeleteLoadBalancerPolicy(pOpt); err != nil {
		return fmt.Errorf("Error removing a policy from load balancer: %s", err)
	}

	return nil
}

func resourceProxyProtocolPolicyRemove(policyName string, ports []interface{}, backends map[int64][]string) ([]*elb.SetLoadBalancerPoliciesForBackendServerInput, error) {
	inputs := make([]*elb.SetLoadBalancerPoliciesForBackendServerInput, 0, len(ports))
	for _, p := range ports {
		ip, err := strconv.ParseInt(p.(string), 10, 64)
		if err != nil {
			return nil, fmt.Errorf("Error detaching the policy: %s", err)
		}

		newPolicies := []*string{}
		curPolicies, found := backends[ip]
		if !found {
			// No policy for this instance port found, just skip it.
			continue
		}

		for _, p := range curPolicies {
			if p == policyName {
				// remove the policy
				continue
			}
			newPolicies = append(newPolicies, aws.String(p))
		}

		inputs = append(inputs, &elb.SetLoadBalancerPoliciesForBackendServerInput{
			InstancePort: &ip,
			PolicyNames:  newPolicies,
		})
	}
	return inputs, nil
}

func resourceProxyProtocolPolicyAdd(policyName string, ports []interface{}, backends map[int64][]string) ([]*elb.SetLoadBalancerPoliciesForBackendServerInput, error) {
	inputs := make([]*elb.SetLoadBalancerPoliciesForBackendServerInput, 0, len(ports))
	for _, p := range ports {
		ip, err := strconv.ParseInt(p.(string), 10, 64)
		if err != nil {
			return nil, fmt.Errorf("Error attaching the policy: %s", err)
		}

		newPolicies := []*string{}
		curPolicies := backends[ip]
		for _, p := range curPolicies {
			if p == policyName {
				// Just remove it for now. It will be back later.
				continue
			}
			newPolicies = append(newPolicies, aws.String(p))
		}
		newPolicies = append(newPolicies, aws.String(policyName))

		inputs = append(inputs, &elb.SetLoadBalancerPoliciesForBackendServerInput{
			InstancePort: &ip,
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
