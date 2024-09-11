// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lightsail

import (
	"context"
	"time"

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
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_lightsail_lb", name="LB")
// @Tags(identifierAttribute="id", resourceType="LB")
func ResourceLoadBalancer() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceLoadBalancerCreate,
		ReadWithoutTimeout:   resourceLoadBalancerRead,
		UpdateWithoutTimeout: resourceLoadBalancerUpdate,
		DeleteWithoutTimeout: resourceLoadBalancerDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrCreatedAt: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDNSName: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"health_check_path": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "/",
			},
			"instance_port": {
				Type:         schema.TypeInt,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.IntBetween(0, 65535),
			},
			names.AttrIPAddressType: {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "dualstack",
				ValidateFunc: validation.StringInSlice([]string{
					"dualstack",
					"ipv4",
				}, false),
			},
			names.AttrProtocol: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"public_ports": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeInt},
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(2, 255),
					validation.StringMatch(regexache.MustCompile(`^[A-Za-z]`), "must begin with an alphabetic character"),
					validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z_.-]+[^_.-]$`), "must contain only alphanumeric characters, underscores, hyphens, and dots"),
				),
			},
			"support_code": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceLoadBalancerCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).LightsailClient(ctx)

	lbName := d.Get(names.AttrName).(string)
	in := lightsail.CreateLoadBalancerInput{
		InstancePort:     int32(d.Get("instance_port").(int)),
		LoadBalancerName: aws.String(lbName),
		Tags:             getTagsIn(ctx),
	}

	if d.Get("health_check_path").(string) != "/" {
		in.HealthCheckPath = aws.String(d.Get("health_check_path").(string))
	}

	out, err := conn.CreateLoadBalancer(ctx, &in)

	if err != nil {
		return create.AppendDiagError(diags, names.Lightsail, string(types.OperationTypeCreateLoadBalancer), ResLoadBalancer, lbName, err)
	}

	diag := expandOperations(ctx, conn, out.Operations, types.OperationTypeCreateLoadBalancer, ResLoadBalancer, lbName)

	if diag != nil {
		return diag
	}

	d.SetId(lbName)

	return append(diags, resourceLoadBalancerRead(ctx, d, meta)...)
}

func resourceLoadBalancerRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).LightsailClient(ctx)

	lb, err := FindLoadBalancerById(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		create.LogNotFoundRemoveState(names.Lightsail, create.ErrActionReading, ResLoadBalancer, d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.Lightsail, create.ErrActionReading, ResLoadBalancer, d.Id(), err)
	}

	d.Set(names.AttrARN, lb.Arn)
	d.Set(names.AttrCreatedAt, lb.CreatedAt.Format(time.RFC3339))
	d.Set(names.AttrDNSName, lb.DnsName)
	d.Set("health_check_path", lb.HealthCheckPath)
	d.Set("instance_port", lb.InstancePort)
	d.Set(names.AttrIPAddressType, lb.IpAddressType)
	d.Set(names.AttrProtocol, lb.Protocol)
	d.Set("public_ports", lb.PublicPorts)
	d.Set(names.AttrName, lb.Name)
	d.Set("support_code", lb.SupportCode)

	setTagsOut(ctx, lb.Tags)

	return diags
}

func resourceLoadBalancerUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).LightsailClient(ctx)
	lbName := d.Get(names.AttrName).(string)

	in := &lightsail.UpdateLoadBalancerAttributeInput{
		LoadBalancerName: aws.String(lbName),
	}

	if d.HasChange("health_check_path") {
		healthCheckIn := in
		healthCheckIn.AttributeName = types.LoadBalancerAttributeNameHealthCheckPath
		healthCheckIn.AttributeValue = aws.String(d.Get("health_check_path").(string))

		out, err := conn.UpdateLoadBalancerAttribute(ctx, healthCheckIn)

		if err != nil {
			return create.AppendDiagError(diags, names.Lightsail, string(types.OperationTypeUpdateLoadBalancerAttribute), ResLoadBalancer, lbName, err)
		}

		diag := expandOperations(ctx, conn, out.Operations, types.OperationTypeUpdateLoadBalancerAttribute, ResLoadBalancer, lbName)

		if diag != nil {
			return diag
		}
	}

	return append(diags, resourceLoadBalancerRead(ctx, d, meta)...)
}

func resourceLoadBalancerDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).LightsailClient(ctx)
	lbName := d.Get(names.AttrName).(string)

	out, err := conn.DeleteLoadBalancer(ctx, &lightsail.DeleteLoadBalancerInput{
		LoadBalancerName: aws.String(d.Id()),
	})

	if err != nil && errs.IsA[*types.NotFoundException](err) {
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.Lightsail, string(types.OperationTypeDeleteLoadBalancer), ResLoadBalancer, lbName, err)
	}

	diag := expandOperations(ctx, conn, out.Operations, types.OperationTypeDeleteLoadBalancer, ResLoadBalancer, lbName)

	if diag != nil {
		return diag
	}

	return diags
}

func FindLoadBalancerById(ctx context.Context, conn *lightsail.Client, name string) (*types.LoadBalancer, error) {
	in := &lightsail.GetLoadBalancerInput{LoadBalancerName: aws.String(name)}
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

	if out == nil || out.LoadBalancer == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	lb := out.LoadBalancer

	return lb, nil
}
