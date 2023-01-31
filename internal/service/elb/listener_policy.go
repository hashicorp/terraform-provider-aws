package elb

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
)

func ResourceListenerPolicy() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceListenerPolicyCreate,
		ReadWithoutTimeout:   resourceListenerPolicyRead,
		UpdateWithoutTimeout: resourceListenerPolicyCreate,
		DeleteWithoutTimeout: resourceListenerPolicyDelete,

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

			"load_balancer_port": {
				Type:     schema.TypeInt,
				Required: true,
			},
		},
	}
}

func resourceListenerPolicyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ELBConn()

	loadBalancerName := d.Get("load_balancer_name")

	policyNames := []*string{}
	if v, ok := d.GetOk("policy_names"); ok {
		policyNames = flex.ExpandStringSet(v.(*schema.Set))
	}

	setOpts := &elb.SetLoadBalancerPoliciesOfListenerInput{
		LoadBalancerName: aws.String(loadBalancerName.(string)),
		LoadBalancerPort: aws.Int64(int64(d.Get("load_balancer_port").(int))),
		PolicyNames:      policyNames,
	}

	if _, err := conn.SetLoadBalancerPoliciesOfListenerWithContext(ctx, setOpts); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting LoadBalancerPoliciesOfListener: %s", err)
	}

	d.SetId(fmt.Sprintf("%s:%s", *setOpts.LoadBalancerName, strconv.FormatInt(*setOpts.LoadBalancerPort, 10)))
	return append(diags, resourceListenerPolicyRead(ctx, d, meta)...)
}

func resourceListenerPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ELBConn()

	loadBalancerName, loadBalancerPort := ListenerPoliciesParseID(d.Id())

	describeElbOpts := &elb.DescribeLoadBalancersInput{
		LoadBalancerNames: []*string{aws.String(loadBalancerName)},
	}

	describeResp, err := conn.DescribeLoadBalancersWithContext(ctx, describeElbOpts)

	if err != nil {
		if ec2err, ok := err.(awserr.Error); ok {
			if ec2err.Code() == "LoadBalancerNotFound" {
				d.SetId("")
				return sdkdiag.AppendErrorf(diags, "LoadBalancerNotFound: %s", err)
			}
		}
		return sdkdiag.AppendErrorf(diags, "retrieving ELB description: %s", err)
	}

	if len(describeResp.LoadBalancerDescriptions) != 1 {
		return sdkdiag.AppendErrorf(diags, "Unable to find ELB: %#v", describeResp.LoadBalancerDescriptions)
	}

	lb := describeResp.LoadBalancerDescriptions[0]

	policyNames := []*string{}
	for _, listener := range lb.ListenerDescriptions {
		if loadBalancerPort != strconv.Itoa(int(aws.Int64Value(listener.Listener.LoadBalancerPort))) {
			continue
		}

		policyNames = append(policyNames, listener.PolicyNames...)
	}

	d.Set("load_balancer_name", loadBalancerName)
	loadBalancerPortVal, err := strconv.ParseInt(loadBalancerPort, 10, 64)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "parsing load balancer port: %s", err)
	}
	d.Set("load_balancer_port", loadBalancerPortVal)
	d.Set("policy_names", flex.FlattenStringList(policyNames))

	return diags
}

func resourceListenerPolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ELBConn()

	loadBalancerName, loadBalancerPort := ListenerPoliciesParseID(d.Id())

	loadBalancerPortInt, err := strconv.ParseInt(loadBalancerPort, 10, 64)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "parsing loadBalancerPort as integer: %s", err)
	}

	setOpts := &elb.SetLoadBalancerPoliciesOfListenerInput{
		LoadBalancerName: aws.String(loadBalancerName),
		LoadBalancerPort: aws.Int64(loadBalancerPortInt),
		PolicyNames:      []*string{},
	}

	if _, err := conn.SetLoadBalancerPoliciesOfListenerWithContext(ctx, setOpts); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting LoadBalancerPoliciesOfListener: %s", err)
	}

	return diags
}

func ListenerPoliciesParseID(id string) (string, string) {
	parts := strings.SplitN(id, ":", 2)
	return parts[0], parts[1]
}
