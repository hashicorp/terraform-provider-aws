// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lightsail

import (
	"context"
	"fmt"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lightsail"
	"github.com/aws/aws-sdk-go-v2/service/lightsail/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_lightsail_lb_https_redirection_policy")
func ResourceLoadBalancerHTTPSRedirectionPolicy() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceLoadBalancerHTTPSRedirectionPolicyCreate,
		ReadWithoutTimeout:   resourceLoadBalancerHTTPSRedirectionPolicyRead,
		UpdateWithoutTimeout: resourceLoadBalancerHTTPSRedirectionPolicyUpdate,
		DeleteWithoutTimeout: resourceLoadBalancerHTTPSRedirectionPolicyDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrEnabled: {
				Type:     schema.TypeBool,
				Required: true,
			},
			"lb_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(2, 255),
					validation.StringMatch(regexache.MustCompile(`^[A-Za-z]`), "must begin with an alphabetic character"),
					validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z_.-]+[^_.-]$`), "must contain only alphanumeric characters, underscores, hyphens, and dots"),
				),
			},
		},
	}
}

func resourceLoadBalancerHTTPSRedirectionPolicyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).LightsailClient(ctx)
	lbName := d.Get("lb_name").(string)
	in := lightsail.UpdateLoadBalancerAttributeInput{
		LoadBalancerName: aws.String(lbName),
		AttributeName:    types.LoadBalancerAttributeNameHttpsRedirectionEnabled,
		AttributeValue:   aws.String(fmt.Sprint(d.Get(names.AttrEnabled).(bool))),
	}

	out, err := conn.UpdateLoadBalancerAttribute(ctx, &in)

	if err != nil {
		return create.AppendDiagError(diags, names.Lightsail, string(types.OperationTypeUpdateLoadBalancerAttribute), ResLoadBalancerHTTPSRedirectionPolicy, lbName, err)
	}

	diag := expandOperations(ctx, conn, out.Operations, types.OperationTypeUpdateLoadBalancerAttribute, ResLoadBalancerHTTPSRedirectionPolicy, lbName)

	if diag != nil {
		return diag
	}

	d.SetId(lbName)

	return append(diags, resourceLoadBalancerHTTPSRedirectionPolicyRead(ctx, d, meta)...)
}

func resourceLoadBalancerHTTPSRedirectionPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).LightsailClient(ctx)

	out, err := FindLoadBalancerHTTPSRedirectionPolicyById(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		create.LogNotFoundRemoveState(names.Lightsail, create.ErrActionReading, ResLoadBalancerHTTPSRedirectionPolicy, d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.Lightsail, create.ErrActionReading, ResLoadBalancerHTTPSRedirectionPolicy, d.Id(), err)
	}

	d.Set(names.AttrEnabled, out)
	d.Set("lb_name", d.Id())

	return diags
}

func resourceLoadBalancerHTTPSRedirectionPolicyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).LightsailClient(ctx)
	lbName := d.Get("lb_name").(string)
	if d.HasChange(names.AttrEnabled) {
		in := lightsail.UpdateLoadBalancerAttributeInput{
			LoadBalancerName: aws.String(lbName),
			AttributeName:    types.LoadBalancerAttributeNameHttpsRedirectionEnabled,
			AttributeValue:   aws.String(fmt.Sprint(d.Get(names.AttrEnabled).(bool))),
		}

		out, err := conn.UpdateLoadBalancerAttribute(ctx, &in)

		if err != nil {
			return create.AppendDiagError(diags, names.Lightsail, string(types.OperationTypeUpdateLoadBalancerAttribute), ResLoadBalancerHTTPSRedirectionPolicy, lbName, err)
		}

		diag := expandOperations(ctx, conn, out.Operations, types.OperationTypeUpdateLoadBalancerAttribute, ResLoadBalancerHTTPSRedirectionPolicy, lbName)

		if diag != nil {
			return diag
		}
	}

	return append(diags, resourceLoadBalancerHTTPSRedirectionPolicyRead(ctx, d, meta)...)
}

func resourceLoadBalancerHTTPSRedirectionPolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).LightsailClient(ctx)
	lbName := d.Get("lb_name").(string)
	in := lightsail.UpdateLoadBalancerAttributeInput{
		LoadBalancerName: aws.String(lbName),
		AttributeName:    types.LoadBalancerAttributeNameHttpsRedirectionEnabled,
		AttributeValue:   aws.String("false"),
	}

	out, err := conn.UpdateLoadBalancerAttribute(ctx, &in)

	if err != nil {
		return create.AppendDiagError(diags, names.Lightsail, string(types.OperationTypeUpdateLoadBalancerAttribute), ResLoadBalancerHTTPSRedirectionPolicy, lbName, err)
	}

	diag := expandOperations(ctx, conn, out.Operations, types.OperationTypeUpdateLoadBalancerAttribute, ResLoadBalancerHTTPSRedirectionPolicy, lbName)

	if diag != nil {
		return diag
	}

	return diags
}

func FindLoadBalancerHTTPSRedirectionPolicyById(ctx context.Context, conn *lightsail.Client, id string) (*bool, error) {
	in := &lightsail.GetLoadBalancerInput{LoadBalancerName: aws.String(id)}
	out, err := conn.GetLoadBalancer(ctx, in)

	if IsANotFoundError(err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}

	if err != nil {
		return nil, err
	}

	if out == nil || out.LoadBalancer.HttpsRedirectionEnabled == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out.LoadBalancer.HttpsRedirectionEnabled, nil
}
