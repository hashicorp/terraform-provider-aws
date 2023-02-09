package lightsail

import (
	"context"
	"errors"
	"fmt"
	"regexp"

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

func resourceLoadBalancerHTTPSRedirectionPolicyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).LightsailConn()

	in := lightsail.UpdateLoadBalancerAttributeInput{
		LoadBalancerName: aws.String(d.Get("lb_name").(string)),
		AttributeName:    aws.String(lightsail.LoadBalancerAttributeNameHttpsRedirectionEnabled),
		AttributeValue:   aws.String(fmt.Sprint(d.Get("enabled").(bool))),
	}

	out, err := conn.UpdateLoadBalancerAttributeWithContext(ctx, &in)

	if err != nil {
		return create.DiagError(names.Lightsail, lightsail.OperationTypeUpdateLoadBalancerAttribute, ResLoadBalancerHTTPSRedirectionPolicy, d.Get("lb_name").(string), err)
	}

	if len(out.Operations) == 0 {
		return create.DiagError(names.Lightsail, lightsail.OperationTypeUpdateLoadBalancerAttribute, ResLoadBalancerHTTPSRedirectionPolicy, d.Get("lb_name").(string), errors.New("No operations found for Update Load Balancer Attribute request"))
	}

	op := out.Operations[0]
	err = waitOperation(ctx, conn, op.Id)

	if err != nil {
		return create.DiagError(names.Lightsail, lightsail.OperationTypeUpdateLoadBalancerAttribute, ResLoadBalancerHTTPSRedirectionPolicy, d.Get("lb_name").(string), errors.New("Error waiting for Update Load Balancer Attribute request operation"))
	}

	d.SetId(d.Get("lb_name").(string))

	return resourceLoadBalancerHTTPSRedirectionPolicyRead(ctx, d, meta)
}

func resourceLoadBalancerHTTPSRedirectionPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).LightsailConn()

	out, err := FindLoadBalancerHTTPSRedirectionPolicyById(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		create.LogNotFoundRemoveState(names.Lightsail, create.ErrActionReading, ResLoadBalancerHTTPSRedirectionPolicy, d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return create.DiagError(names.Lightsail, create.ErrActionReading, ResLoadBalancerHTTPSRedirectionPolicy, d.Id(), err)
	}

	d.Set("enabled", out)
	d.Set("lb_name", d.Id())

	return nil
}

func resourceLoadBalancerHTTPSRedirectionPolicyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).LightsailConn()

	if d.HasChange("enabled") {
		in := lightsail.UpdateLoadBalancerAttributeInput{
			LoadBalancerName: aws.String(d.Get("lb_name").(string)),
			AttributeName:    aws.String(lightsail.LoadBalancerAttributeNameHttpsRedirectionEnabled),
			AttributeValue:   aws.String(fmt.Sprint(d.Get("enabled").(bool))),
		}

		out, err := conn.UpdateLoadBalancerAttributeWithContext(ctx, &in)

		if err != nil {
			return create.DiagError(names.Lightsail, lightsail.OperationTypeUpdateLoadBalancerAttribute, ResLoadBalancerHTTPSRedirectionPolicy, d.Get("lb_name").(string), err)
		}

		if len(out.Operations) == 0 {
			return create.DiagError(names.Lightsail, lightsail.OperationTypeUpdateLoadBalancerAttribute, ResLoadBalancerHTTPSRedirectionPolicy, d.Get("lb_name").(string), errors.New("No operations found for Update Load Balancer Attribute request"))
		}

		op := out.Operations[0]
		err = waitOperation(ctx, conn, op.Id)

		if err != nil {
			return create.DiagError(names.Lightsail, lightsail.OperationTypeUpdateLoadBalancerAttribute, ResLoadBalancerHTTPSRedirectionPolicy, d.Get("lb_name").(string), errors.New("Error waiting for Update Load Balancer Attribute request operation"))
		}
	}

	return resourceLoadBalancerHTTPSRedirectionPolicyRead(ctx, d, meta)
}

func resourceLoadBalancerHTTPSRedirectionPolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).LightsailConn()

	in := lightsail.UpdateLoadBalancerAttributeInput{
		LoadBalancerName: aws.String(d.Get("lb_name").(string)),
		AttributeName:    aws.String(lightsail.LoadBalancerAttributeNameHttpsRedirectionEnabled),
		AttributeValue:   aws.String("false"),
	}

	out, err := conn.UpdateLoadBalancerAttributeWithContext(ctx, &in)

	if err != nil {
		return create.DiagError(names.Lightsail, lightsail.OperationTypeUpdateLoadBalancerAttribute, ResLoadBalancerHTTPSRedirectionPolicy, d.Get("lb_name").(string), err)
	}

	if len(out.Operations) == 0 {
		return create.DiagError(names.Lightsail, lightsail.OperationTypeUpdateLoadBalancerAttribute, ResLoadBalancerHTTPSRedirectionPolicy, d.Get("lb_name").(string), errors.New("No operations found for Update Load Balancer Attribute request"))
	}

	op := out.Operations[0]
	err = waitOperation(ctx, conn, op.Id)

	if err != nil {
		return create.DiagError(names.Lightsail, lightsail.OperationTypeUpdateLoadBalancerAttribute, ResLoadBalancerHTTPSRedirectionPolicy, d.Get("lb_name").(string), errors.New("Error waiting for Update Load Balancer Attribute request operation"))
	}

	return nil
}
