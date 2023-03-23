package elb

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
)

func ResourceBackendServerPolicy() *schema.Resource {
	return &schema.Resource{
		Create: resourceBackendServerPolicyCreate,
		Read:   resourceBackendServerPolicyRead,
		Update: resourceBackendServerPolicyCreate,
		Delete: resourceBackendServerPolicyDelete,

		Schema: map[string]*schema.Schema{
			"load_balancer_name": {
				Type:     schema.TypeString,
				Required: true,
			},

			"policy_names": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Optional: true,
				Set:      schema.HashString,
			},

			"instance_port": {
				Type:     schema.TypeInt,
				Required: true,
			},
		},
	}
}

func resourceBackendServerPolicyCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ELBConn

	loadBalancerName := d.Get("load_balancer_name")

	policyNames := []*string{}
	if v, ok := d.GetOk("policy_names"); ok {
		policyNames = flex.ExpandStringSet(v.(*schema.Set))
	}

	setOpts := &elb.SetLoadBalancerPoliciesForBackendServerInput{
		LoadBalancerName: aws.String(loadBalancerName.(string)),
		InstancePort:     aws.Int64(int64(d.Get("instance_port").(int))),
		PolicyNames:      policyNames,
	}

	if _, err := conn.SetLoadBalancerPoliciesForBackendServer(setOpts); err != nil {
		return fmt.Errorf("Error setting LoadBalancerPoliciesForBackendServer: %s", err)
	}

	d.SetId(fmt.Sprintf("%s:%s", *setOpts.LoadBalancerName, strconv.FormatInt(*setOpts.InstancePort, 10)))
	return resourceBackendServerPolicyRead(d, meta)
}

func resourceBackendServerPolicyRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ELBConn

	loadBalancerName, instancePort := BackendServerPoliciesParseID(d.Id())

	describeElbOpts := &elb.DescribeLoadBalancersInput{
		LoadBalancerNames: []*string{aws.String(loadBalancerName)},
	}

	describeResp, err := conn.DescribeLoadBalancers(describeElbOpts)

	if err != nil {
		if ec2err, ok := err.(awserr.Error); ok {
			if ec2err.Code() == "LoadBalancerNotFound" {
				d.SetId("")
				return fmt.Errorf("LoadBalancerNotFound: %s", err)
			}
		}
		return fmt.Errorf("Error retrieving ELB description: %s", err)
	}

	if len(describeResp.LoadBalancerDescriptions) != 1 {
		return fmt.Errorf("Unable to find ELB: %#v", describeResp.LoadBalancerDescriptions)
	}

	lb := describeResp.LoadBalancerDescriptions[0]

	policyNames := []*string{}
	for _, backendServer := range lb.BackendServerDescriptions {
		if instancePort != strconv.Itoa(int(aws.Int64Value(backendServer.InstancePort))) {
			continue
		}

		policyNames = append(policyNames, backendServer.PolicyNames...)
	}

	d.Set("load_balancer_name", loadBalancerName)
	instancePortVal, err := strconv.ParseInt(instancePort, 10, 64)
	if err != nil {
		return fmt.Errorf("error parsing instance port: %s", err)
	}
	d.Set("instance_port", instancePortVal)
	d.Set("policy_names", flex.FlattenStringList(policyNames))

	return nil
}

func resourceBackendServerPolicyDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ELBConn

	loadBalancerName, instancePort := BackendServerPoliciesParseID(d.Id())

	instancePortInt, err := strconv.ParseInt(instancePort, 10, 64)
	if err != nil {
		return fmt.Errorf("Error parsing instancePort as integer: %s", err)
	}

	setOpts := &elb.SetLoadBalancerPoliciesForBackendServerInput{
		LoadBalancerName: aws.String(loadBalancerName),
		InstancePort:     aws.Int64(instancePortInt),
		PolicyNames:      []*string{},
	}

	if _, err := conn.SetLoadBalancerPoliciesForBackendServer(setOpts); err != nil {
		return fmt.Errorf("Error setting LoadBalancerPoliciesForBackendServer: %s", err)
	}

	return nil
}

func BackendServerPoliciesParseID(id string) (string, string) {
	parts := strings.SplitN(id, ":", 2)
	return parts[0], parts[1]
}
