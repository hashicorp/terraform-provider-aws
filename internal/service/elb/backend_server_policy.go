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
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKResource("aws_load_balancer_backend_server_policy", name="Backend Server Policy")
func resourceBackendServerPolicy() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceBackendServerPolicySet,
		ReadWithoutTimeout:   resourceBackendServerPolicyRead,
		UpdateWithoutTimeout: resourceBackendServerPolicySet,
		DeleteWithoutTimeout: resourceBackendServerPolicyDelete,

		Schema: map[string]*schema.Schema{
			"instance_port": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"load_balancer_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"policy_names": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Optional: true,
			},
		},
	}
}

func resourceBackendServerPolicySet(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ELBClient(ctx)

	instancePort := d.Get("instance_port").(int)
	lbName := d.Get("load_balancer_name").(string)
	id := backendServerPolicyCreateResourceID(lbName, instancePort)
	input := &elasticloadbalancing.SetLoadBalancerPoliciesForBackendServerInput{
		InstancePort:     aws.Int32(int32(instancePort)),
		LoadBalancerName: aws.String(lbName),
	}

	if v, ok := d.GetOk("policy_names"); ok && v.(*schema.Set).Len() > 0 {
		input.PolicyNames = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	_, err := conn.SetLoadBalancerPoliciesForBackendServer(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "setting ELB Classic Backend Server Policy (%s): %s", id, err)
	}

	if d.IsNewResource() {
		d.SetId(id)
	}

	return append(diags, resourceBackendServerPolicyRead(ctx, d, meta)...)
}

func resourceBackendServerPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ELBClient(ctx)

	lbName, instancePort, err := backendServerPolicyParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	policyNames, err := findLoadBalancerBackendServerPolicyByTwoPartKey(ctx, conn, lbName, instancePort)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] ELB Classic Backend Server Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ELB Classic Backend Server Policy (%s): %s", d.Id(), err)
	}

	d.Set("instance_port", instancePort)
	d.Set("load_balancer_name", lbName)
	d.Set("policy_names", policyNames)

	return diags
}

func resourceBackendServerPolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ELBClient(ctx)

	lbName, instancePort, err := backendServerPolicyParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[DEBUG] Deleting ELB Classic Backend Server Policy: %s", d.Id())
	_, err = conn.SetLoadBalancerPoliciesForBackendServer(ctx, &elasticloadbalancing.SetLoadBalancerPoliciesForBackendServerInput{
		InstancePort:     aws.Int32(int32(instancePort)),
		LoadBalancerName: aws.String(lbName),
		PolicyNames:      []string{},
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "setting ELB Classic Backend Server Policy (%s): %s", d.Id(), err)
	}

	return diags
}

func findLoadBalancerBackendServerPolicyByTwoPartKey(ctx context.Context, conn *elasticloadbalancing.Client, lbName string, instancePort int) ([]string, error) {
	lb, err := findLoadBalancerByName(ctx, conn, lbName)

	if err != nil {
		return nil, err
	}

	var policyNames []string

	for _, v := range lb.BackendServerDescriptions {
		if aws.ToInt32(v.InstancePort) != int32(instancePort) {
			continue
		}

		policyNames = append(policyNames, v.PolicyNames...)
	}

	return policyNames, nil
}

const backendServerPolicyResourceIDSeparator = ":"

func backendServerPolicyCreateResourceID(lbName string, instancePort int) string {
	parts := []string{lbName, strconv.Itoa(instancePort)}
	id := strings.Join(parts, backendServerPolicyResourceIDSeparator)

	return id
}

func backendServerPolicyParseResourceID(id string) (string, int, error) {
	parts := strings.Split(id, backendServerPolicyResourceIDSeparator)

	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		v, err := strconv.Atoi(parts[1])

		if err != nil {
			return "", 0, err
		}

		return parts[0], v, nil
	}

	return "", 0, fmt.Errorf("unexpected format for ID (%[1]s), expected LBNAME%[2]sINSTANCEPORT", id, backendServerPolicyResourceIDSeparator)
}
