package lightsail

import (
	"context"
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
	lbName := d.Get("lb_name").(string)
	in := lightsail.UpdateLoadBalancerAttributeInput{
		LoadBalancerName: aws.String(lbName),
		AttributeName:    aws.String(lightsail.LoadBalancerAttributeNameHttpsRedirectionEnabled),
		AttributeValue:   aws.String(fmt.Sprint(d.Get("enabled").(bool))),
	}

	out, err := conn.UpdateLoadBalancerAttributeWithContext(ctx, &in)

	if err != nil {
		return create.DiagError(names.Lightsail, lightsail.OperationTypeUpdateLoadBalancerAttribute, ResLoadBalancerHTTPSRedirectionPolicy, lbName, err)
	}

	diag := expandOperations(ctx, conn, out.Operations, lightsail.OperationTypeUpdateLoadBalancerAttribute, ResLoadBalancerHTTPSRedirectionPolicy, lbName)

	if diag != nil {
		return diag
	}

	d.SetId(lbName)

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
	lbName := d.Get("lb_name").(string)
	if d.HasChange("enabled") {
		in := lightsail.UpdateLoadBalancerAttributeInput{
			LoadBalancerName: aws.String(lbName),
			AttributeName:    aws.String(lightsail.LoadBalancerAttributeNameHttpsRedirectionEnabled),
			AttributeValue:   aws.String(fmt.Sprint(d.Get("enabled").(bool))),
		}

		out, err := conn.UpdateLoadBalancerAttributeWithContext(ctx, &in)

		if err != nil {
			return create.DiagError(names.Lightsail, lightsail.OperationTypeUpdateLoadBalancerAttribute, ResLoadBalancerHTTPSRedirectionPolicy, lbName, err)
		}

		diag := expandOperations(ctx, conn, out.Operations, lightsail.OperationTypeUpdateLoadBalancerAttribute, ResLoadBalancerHTTPSRedirectionPolicy, lbName)

		if diag != nil {
			return diag
		}
	}

	return resourceLoadBalancerHTTPSRedirectionPolicyRead(ctx, d, meta)
}

func resourceLoadBalancerHTTPSRedirectionPolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).LightsailConn()
	lbName := d.Get("lb_name").(string)
	in := lightsail.UpdateLoadBalancerAttributeInput{
		LoadBalancerName: aws.String(lbName),
		AttributeName:    aws.String(lightsail.LoadBalancerAttributeNameHttpsRedirectionEnabled),
		AttributeValue:   aws.String("false"),
	}

	out, err := conn.UpdateLoadBalancerAttributeWithContext(ctx, &in)

	if err != nil {
		return create.DiagError(names.Lightsail, lightsail.OperationTypeUpdateLoadBalancerAttribute, ResLoadBalancerHTTPSRedirectionPolicy, lbName, err)
	}

	diag := expandOperations(ctx, conn, out.Operations, lightsail.OperationTypeUpdateLoadBalancerAttribute, ResLoadBalancerHTTPSRedirectionPolicy, lbName)

	if diag != nil {
		return diag
	}

	return nil
}
