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
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKResource("aws_load_balancer_listener_policy")
func ResourceListenerPolicy() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceListenerPolicySet,
		ReadWithoutTimeout:   resourceListenerPolicyRead,
		UpdateWithoutTimeout: resourceListenerPolicySet,
		DeleteWithoutTimeout: resourceListenerPolicyDelete,

		Schema: map[string]*schema.Schema{
			"load_balancer_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"load_balancer_port": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"policy_names": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"triggers": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func resourceListenerPolicySet(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ELBConn(ctx)

	lbName := d.Get("load_balancer_name").(string)
	lbPort := d.Get("load_balancer_port").(int)
	id := ListenerPolicyCreateResourceID(lbName, lbPort)
	input := &elb.SetLoadBalancerPoliciesOfListenerInput{
		LoadBalancerName: aws.String(lbName),
		LoadBalancerPort: aws.Int64(int64(lbPort)),
	}

	if v, ok := d.GetOk("policy_names"); ok && v.(*schema.Set).Len() > 0 {
		input.PolicyNames = flex.ExpandStringSet(v.(*schema.Set))
	}

	_, err := conn.SetLoadBalancerPoliciesOfListenerWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "setting ELB Classic Listener Policy (%s): %s", id, err)
	}

	d.SetId(id)

	return append(diags, resourceListenerPolicyRead(ctx, d, meta)...)
}

func resourceListenerPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ELBConn(ctx)

	lbName, lbPort, err := ListenerPolicyParseResourceID(d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "parsing resource ID: %s", err)
	}

	policyNames, err := FindLoadBalancerListenerPolicyByTwoPartKey(ctx, conn, lbName, lbPort)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] ELB Classic Listener Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ELB Classic Listener Policy (%s): %s", d.Id(), err)
	}

	d.Set("load_balancer_name", lbName)
	d.Set("load_balancer_port", lbPort)
	d.Set("policy_names", policyNames)

	return diags
}

func resourceListenerPolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ELBConn(ctx)

	lbName, lbPort, err := ListenerPolicyParseResourceID(d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "parsing resource ID: %s", err)
	}

	input := &elb.SetLoadBalancerPoliciesOfListenerInput{
		LoadBalancerName: aws.String(lbName),
		LoadBalancerPort: aws.Int64(int64(lbPort)),
		PolicyNames:      aws.StringSlice([]string{}),
	}

	log.Printf("[DEBUG] Deleting ELB Classic Listener Policy: %s", d.Id())
	_, err = conn.SetLoadBalancerPoliciesOfListenerWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "setting ELB Classic Listener Policy (%s): %s", d.Id(), err)
	}

	return diags
}

func FindLoadBalancerListenerPolicyByTwoPartKey(ctx context.Context, conn *elb.ELB, lbName string, lbPort int) ([]string, error) {
	lb, err := FindLoadBalancerByName(ctx, conn, lbName)

	if err != nil {
		return nil, err
	}

	var policyNames []string

	for _, v := range lb.ListenerDescriptions {
		if v == nil {
			continue
		}

		if aws.Int64Value(v.Listener.LoadBalancerPort) != int64(lbPort) {
			continue
		}

		policyNames = append(policyNames, aws.StringValueSlice(v.PolicyNames)...)
	}

	return policyNames, nil
}

const listenerPolicyResourceIDSeparator = ":"

func ListenerPolicyCreateResourceID(lbName string, lbPort int) string {
	parts := []string{lbName, strconv.Itoa(lbPort)}
	id := strings.Join(parts, listenerPolicyResourceIDSeparator)

	return id
}

func ListenerPolicyParseResourceID(id string) (string, int, error) {
	parts := strings.Split(id, listenerPolicyResourceIDSeparator)

	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		v, err := strconv.Atoi(parts[1])

		if err != nil {
			return "", 0, err
		}

		return parts[0], v, nil
	}

	return "", 0, fmt.Errorf("unexpected format for ID (%[1]s), expected LBNAME%[2]sLBPORT", id, listenerPolicyResourceIDSeparator)
}
