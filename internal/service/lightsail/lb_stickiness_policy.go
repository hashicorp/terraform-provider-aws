// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lightsail

import (
	"context"
	"fmt"
	"regexp"
	"strconv"

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

// @SDKResource("aws_lightsail_lb_stickiness_policy")
func ResourceLoadBalancerStickinessPolicy() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceLoadBalancerStickinessPolicyCreate,
		ReadWithoutTimeout:   resourceLoadBalancerStickinessPolicyRead,
		UpdateWithoutTimeout: resourceLoadBalancerStickinessPolicyUpdate,
		DeleteWithoutTimeout: resourceLoadBalancerStickinessPolicyDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"cookie_duration": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"enabled": {
				Type:     schema.TypeBool,
				Required: true,
			},
			"lb_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(2, 255),
					validation.StringMatch(regexp.MustCompile(`^[a-zA-Z]`), "must begin with an alphabetic character"),
					validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9_\-.]+[^._\-]$`), "must contain only alphanumeric characters, underscores, hyphens, and dots"),
				),
			},
		},
	}
}

func resourceLoadBalancerStickinessPolicyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).LightsailClient(ctx)
	lbName := d.Get("lb_name").(string)
	for _, v := range []string{"enabled", "cookie_duration"} {
		in := lightsail.UpdateLoadBalancerAttributeInput{
			LoadBalancerName: aws.String(lbName),
		}

		if v == "enabled" {
			in.AttributeName = types.LoadBalancerAttributeNameSessionStickinessEnabled
			in.AttributeValue = aws.String(fmt.Sprint(d.Get("enabled").(bool)))
		}

		if v == "cookie_duration" {
			in.AttributeName = types.LoadBalancerAttributeNameSessionStickinessLbCookieDurationSeconds
			in.AttributeValue = aws.String(fmt.Sprint(d.Get("cookie_duration").(int)))
		}

		out, err := conn.UpdateLoadBalancerAttribute(ctx, &in)

		if err != nil {
			return create.DiagError(names.Lightsail, string(types.OperationTypeUpdateLoadBalancerAttribute), ResLoadBalancerStickinessPolicy, lbName, err)
		}

		diag := expandOperations(ctx, conn, out.Operations, types.OperationTypeUpdateLoadBalancerAttribute, ResLoadBalancerStickinessPolicy, lbName)

		if diag != nil {
			return diag
		}
	}

	d.SetId(lbName)

	return resourceLoadBalancerStickinessPolicyRead(ctx, d, meta)
}

func resourceLoadBalancerStickinessPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).LightsailClient(ctx)

	out, err := FindLoadBalancerStickinessPolicyById(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		create.LogNotFoundRemoveState(names.Lightsail, create.ErrActionReading, ResLoadBalancerStickinessPolicy, d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return create.DiagError(names.Lightsail, create.ErrActionReading, ResLoadBalancerStickinessPolicy, d.Id(), err)
	}

	boolValue, err := strconv.ParseBool(out[string(types.LoadBalancerAttributeNameSessionStickinessEnabled)])
	if err != nil {
		return create.DiagError(names.Lightsail, create.ErrActionReading, ResLoadBalancerStickinessPolicy, d.Id(), err)
	}

	intValue, err := strconv.Atoi(out[string(types.LoadBalancerAttributeNameSessionStickinessLbCookieDurationSeconds)])
	if err != nil {
		return create.DiagError(names.Lightsail, create.ErrActionReading, ResLoadBalancerStickinessPolicy, d.Id(), err)
	}

	d.Set("cookie_duration", intValue)
	d.Set("enabled", boolValue)
	d.Set("lb_name", d.Id())

	return nil
}

func resourceLoadBalancerStickinessPolicyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).LightsailClient(ctx)
	lbName := d.Get("lb_name").(string)
	if d.HasChange("enabled") {
		in := lightsail.UpdateLoadBalancerAttributeInput{
			LoadBalancerName: aws.String(lbName),
			AttributeName:    types.LoadBalancerAttributeNameSessionStickinessEnabled,
			AttributeValue:   aws.String(fmt.Sprint(d.Get("enabled").(bool))),
		}

		out, err := conn.UpdateLoadBalancerAttribute(ctx, &in)

		if err != nil {
			return create.DiagError(names.Lightsail, string(types.OperationTypeUpdateLoadBalancerAttribute), ResLoadBalancerStickinessPolicy, lbName, err)
		}

		diag := expandOperations(ctx, conn, out.Operations, types.OperationTypeUpdateLoadBalancerAttribute, ResLoadBalancerStickinessPolicy, lbName)

		if diag != nil {
			return diag
		}
	}

	if d.HasChange("cookie_duration") {
		in := lightsail.UpdateLoadBalancerAttributeInput{
			LoadBalancerName: aws.String(lbName),
			AttributeName:    types.LoadBalancerAttributeNameSessionStickinessLbCookieDurationSeconds,
			AttributeValue:   aws.String(fmt.Sprint(d.Get("cookie_duration").(int))),
		}

		out, err := conn.UpdateLoadBalancerAttribute(ctx, &in)

		if err != nil {
			return create.DiagError(names.Lightsail, string(types.OperationTypeUpdateLoadBalancerAttribute), ResLoadBalancerStickinessPolicy, lbName, err)
		}

		diag := expandOperations(ctx, conn, out.Operations, types.OperationTypeUpdateLoadBalancerAttribute, ResLoadBalancerStickinessPolicy, lbName)

		if diag != nil {
			return diag
		}
	}

	return resourceLoadBalancerStickinessPolicyRead(ctx, d, meta)
}

func resourceLoadBalancerStickinessPolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).LightsailClient(ctx)
	lbName := d.Get("lb_name").(string)
	in := lightsail.UpdateLoadBalancerAttributeInput{
		LoadBalancerName: aws.String(lbName),
		AttributeName:    types.LoadBalancerAttributeNameSessionStickinessEnabled,
		AttributeValue:   aws.String("false"),
	}

	out, err := conn.UpdateLoadBalancerAttribute(ctx, &in)

	if err != nil {
		return create.DiagError(names.Lightsail, string(types.OperationTypeUpdateLoadBalancerAttribute), ResLoadBalancerStickinessPolicy, lbName, err)
	}

	diag := expandOperations(ctx, conn, out.Operations, types.OperationTypeUpdateLoadBalancerAttribute, ResLoadBalancerStickinessPolicy, lbName)

	if diag != nil {
		return diag
	}

	return nil
}

func FindLoadBalancerStickinessPolicyById(ctx context.Context, conn *lightsail.Client, id string) (map[string]string, error) {
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

	if out == nil || out.LoadBalancer.ConfigurationOptions == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out.LoadBalancer.ConfigurationOptions, nil
}
