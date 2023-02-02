package lightsail

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lightsail"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

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
	conn := meta.(*conns.AWSClient).LightsailConn()

	for _, v := range []string{"enabled", "cookie_duration"} {
		in := lightsail.UpdateLoadBalancerAttributeInput{
			LoadBalancerName: aws.String(d.Get("lb_name").(string)),
		}

		if v == "enabled" {
			in.AttributeName = aws.String(lightsail.LoadBalancerAttributeNameSessionStickinessEnabled)
			in.AttributeValue = aws.String(fmt.Sprint(d.Get("enabled").(bool)))
		}

		if v == "cookie_duration" {
			in.AttributeName = aws.String(lightsail.LoadBalancerAttributeNameSessionStickinessLbCookieDurationSeconds)
			in.AttributeValue = aws.String(fmt.Sprint(d.Get("cookie_duration").(int)))
		}

		out, err := conn.UpdateLoadBalancerAttributeWithContext(ctx, &in)

		if err != nil {
			return create.DiagError(names.Lightsail, lightsail.OperationTypeUpdateLoadBalancerAttribute, ResLoadBalancerStickinessPolicy, d.Get("lb_name").(string), err)
		}

		if len(out.Operations) == 0 {
			return create.DiagError(names.Lightsail, lightsail.OperationTypeUpdateLoadBalancerAttribute, ResLoadBalancerStickinessPolicy, d.Get("lb_name").(string), errors.New("No operations found for Update Load Balancer Attribute request"))
		}

		op := out.Operations[0]
		err = waitOperation(ctx, conn, op.Id)

		if err != nil {
			return create.DiagError(names.Lightsail, lightsail.OperationTypeUpdateLoadBalancerAttribute, ResLoadBalancerStickinessPolicy, d.Get("lb_name").(string), errors.New("Error waiting for Update Load Balancer Attribute request operation"))
		}
	}

	d.SetId(d.Get("lb_name").(string))

	return resourceLoadBalancerStickinessPolicyRead(ctx, d, meta)
}

func resourceLoadBalancerStickinessPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).LightsailConn()

	out, err := FindLoadBalancerStickinessPolicyById(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		create.LogNotFoundRemoveState(names.Lightsail, create.ErrActionReading, ResLoadBalancerStickinessPolicy, d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return create.DiagError(names.Lightsail, create.ErrActionReading, ResLoadBalancerStickinessPolicy, d.Id(), err)
	}

	boolValue, err := strconv.ParseBool(*out[lightsail.LoadBalancerAttributeNameSessionStickinessEnabled])
	if err != nil {
		return create.DiagError(names.Lightsail, create.ErrActionReading, ResLoadBalancerStickinessPolicy, d.Id(), err)
	}

	intValue, err := strconv.Atoi(*out[lightsail.LoadBalancerAttributeNameSessionStickinessLbCookieDurationSeconds])
	if err != nil {
		return create.DiagError(names.Lightsail, create.ErrActionReading, ResLoadBalancerStickinessPolicy, d.Id(), err)
	}

	d.Set("cookie_duration", intValue)
	d.Set("enabled", boolValue)
	d.Set("lb_name", d.Id())

	return nil
}

func resourceLoadBalancerStickinessPolicyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).LightsailConn()

	if d.HasChange("enabled") {
		in := lightsail.UpdateLoadBalancerAttributeInput{
			LoadBalancerName: aws.String(d.Get("lb_name").(string)),
			AttributeName:    aws.String(lightsail.LoadBalancerAttributeNameSessionStickinessEnabled),
			AttributeValue:   aws.String(fmt.Sprint(d.Get("enabled").(bool))),
		}

		out, err := conn.UpdateLoadBalancerAttributeWithContext(ctx, &in)

		if err != nil {
			return create.DiagError(names.Lightsail, lightsail.OperationTypeUpdateLoadBalancerAttribute, ResLoadBalancerStickinessPolicy, d.Get("lb_name").(string), err)
		}

		if len(out.Operations) == 0 {
			return create.DiagError(names.Lightsail, lightsail.OperationTypeUpdateLoadBalancerAttribute, ResLoadBalancerStickinessPolicy, d.Get("lb_name").(string), errors.New("No operations found for Update Load Balancer Attribute request"))
		}

		op := out.Operations[0]
		err = waitOperation(ctx, conn, op.Id)

		if err != nil {
			return create.DiagError(names.Lightsail, lightsail.OperationTypeUpdateLoadBalancerAttribute, ResLoadBalancerStickinessPolicy, d.Get("lb_name").(string), errors.New("Error waiting for Update Load Balancer Attribute request operation"))
		}
	}

	if d.HasChange("cookie_duration") {
		in := lightsail.UpdateLoadBalancerAttributeInput{
			LoadBalancerName: aws.String(d.Get("lb_name").(string)),
			AttributeName:    aws.String(lightsail.LoadBalancerAttributeNameSessionStickinessLbCookieDurationSeconds),
			AttributeValue:   aws.String(fmt.Sprint(d.Get("cookie_duration").(int))),
		}

		out, err := conn.UpdateLoadBalancerAttributeWithContext(ctx, &in)

		if err != nil {
			return create.DiagError(names.Lightsail, lightsail.OperationTypeUpdateLoadBalancerAttribute, ResLoadBalancerStickinessPolicy, d.Get("lb_name").(string), err)
		}

		if len(out.Operations) == 0 {
			return create.DiagError(names.Lightsail, lightsail.OperationTypeUpdateLoadBalancerAttribute, ResLoadBalancerStickinessPolicy, d.Get("lb_name").(string), errors.New("No operations found for Update Load Balancer Attribute request"))
		}

		op := out.Operations[0]
		err = waitOperation(ctx, conn, op.Id)

		if err != nil {
			return create.DiagError(names.Lightsail, lightsail.OperationTypeUpdateLoadBalancerAttribute, ResLoadBalancerStickinessPolicy, d.Get("lb_name").(string), errors.New("Error waiting for Update Load Balancer Attribute request operation"))
		}
	}

	return resourceLoadBalancerStickinessPolicyRead(ctx, d, meta)
}

func resourceLoadBalancerStickinessPolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).LightsailConn()

	in := lightsail.UpdateLoadBalancerAttributeInput{
		LoadBalancerName: aws.String(d.Get("lb_name").(string)),
		AttributeName:    aws.String(lightsail.LoadBalancerAttributeNameSessionStickinessEnabled),
		AttributeValue:   aws.String("false"),
	}

	out, err := conn.UpdateLoadBalancerAttributeWithContext(ctx, &in)

	if err != nil {
		return create.DiagError(names.Lightsail, lightsail.OperationTypeUpdateLoadBalancerAttribute, ResLoadBalancerStickinessPolicy, d.Get("lb_name").(string), err)
	}

	if len(out.Operations) == 0 {
		return create.DiagError(names.Lightsail, lightsail.OperationTypeUpdateLoadBalancerAttribute, ResLoadBalancerStickinessPolicy, d.Get("lb_name").(string), errors.New("No operations found for Update Load Balancer Attribute request"))
	}

	op := out.Operations[0]
	err = waitOperation(ctx, conn, op.Id)

	if err != nil {
		return create.DiagError(names.Lightsail, lightsail.OperationTypeUpdateLoadBalancerAttribute, ResLoadBalancerStickinessPolicy, d.Get("lb_name").(string), errors.New("Error waiting for Update Load Balancer Attribute request operation"))
	}

	return nil
}
