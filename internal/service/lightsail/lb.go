package lightsail

import (
	"context"
	"errors"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lightsail"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

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
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"dns_name": {
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
			"ip_address_type": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "dualstack",
				ValidateFunc: validation.StringInSlice([]string{
					"dualstack",
					"ipv4",
				}, false),
			},
			"protocol": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"public_ports": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeInt},
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(2, 255),
					validation.StringMatch(regexp.MustCompile(`^[a-zA-Z]`), "must begin with an alphabetic character"),
					validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9_\-.]+[^._\-]$`), "must contain only alphanumeric characters, underscores, hyphens, and dots"),
				),
			},
			"support_code": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},
		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceLoadBalancerCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).LightsailConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	in := lightsail.CreateLoadBalancerInput{
		HealthCheckPath:  aws.String(d.Get("health_check_path").(string)),
		InstancePort:     aws.Int64(int64(d.Get("instance_port").(int))),
		LoadBalancerName: aws.String(d.Get("name").(string)),
	}

	if len(tags) > 0 {
		in.Tags = Tags(tags.IgnoreAWS())
	}

	out, err := conn.CreateLoadBalancerWithContext(ctx, &in)

	if err != nil {
		return create.DiagError(names.Lightsail, lightsail.OperationTypeCreateLoadBalancer, ResLoadBalancer, d.Get("name").(string), err)
	}

	if len(out.Operations) == 0 {
		return create.DiagError(names.Lightsail, lightsail.OperationTypeCreateLoadBalancer, ResLoadBalancer, d.Get("name").(string), errors.New("No operations found for Create Load Balancer request"))
	}

	op := out.Operations[0]
	d.SetId(d.Get("name").(string))

	err = waitOperation(ctx, conn, op.Id)
	if err != nil {
		return create.DiagError(names.Lightsail, lightsail.OperationTypeCreateLoadBalancer, ResLoadBalancer, d.Get("name").(string), errors.New("Error waiting for Create Load Balancer request operation"))
	}

	return resourceLoadBalancerRead(ctx, d, meta)
}

func resourceLoadBalancerRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).LightsailConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	lb, err := FindLoadBalancerByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		create.LogNotFoundRemoveState(names.Lightsail, create.ErrActionReading, ResLoadBalancer, d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return create.DiagError(names.Lightsail, create.ErrActionReading, ResLoadBalancer, d.Id(), err)
	}

	d.Set("arn", lb.Arn)
	d.Set("created_at", lb.CreatedAt.Format(time.RFC3339))
	d.Set("dns_name", lb.DnsName)
	d.Set("health_check_path", lb.HealthCheckPath)
	d.Set("instance_port", lb.InstancePort)
	d.Set("ip_address_type", lb.IpAddressType)
	d.Set("protocol", lb.Protocol)
	d.Set("public_ports", lb.PublicPorts)
	d.Set("name", lb.Name)
	d.Set("support_code", lb.SupportCode)

	tags := KeyValueTags(lb.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return create.DiagError(names.Lightsail, create.ErrActionReading, ResLoadBalancer, d.Id(), err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return create.DiagError(names.Lightsail, create.ErrActionReading, ResLoadBalancer, d.Id(), err)
	}

	return nil
}

func resourceLoadBalancerUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).LightsailConn()

	in := &lightsail.UpdateLoadBalancerAttributeInput{
		LoadBalancerName: aws.String(d.Get("name").(string)),
	}

	if d.HasChange("health_check_path") {
		healthCheckIn := in
		healthCheckIn.AttributeName = aws.String("HealthCheckPath")
		healthCheckIn.AttributeValue = aws.String(d.Get("health_check_path").(string))

		out, err := conn.UpdateLoadBalancerAttributeWithContext(ctx, healthCheckIn)

		if err != nil {
			return create.DiagError(names.Lightsail, lightsail.OperationTypeCreateLoadBalancer, ResLoadBalancer, d.Get("name").(string), err)
		}

		if len(out.Operations) == 0 {
			return create.DiagError(names.Lightsail, lightsail.OperationTypeCreateLoadBalancer, ResLoadBalancer, d.Get("name").(string), errors.New("No operations found for Create Load Balancer request"))
		}

		op := out.Operations[0]

		err = waitOperation(ctx, conn, op.Id)
		if err != nil {
			return create.DiagError(names.Lightsail, lightsail.OperationTypeCreateLoadBalancer, ResLoadBalancer, d.Get("name").(string), errors.New("Error waiting for Create Load Balancer request operation"))
		}
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := UpdateTags(ctx, conn, d.Id(), o, n); err != nil {
			return create.DiagError(names.Lightsail, create.ErrActionUpdating, ResLoadBalancer, d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(ctx, conn, d.Id(), o, n); err != nil {
			return create.DiagError(names.Lightsail, create.ErrActionUpdating, ResLoadBalancer, d.Id(), err)
		}
	}

	return resourceLoadBalancerRead(ctx, d, meta)
}

func resourceLoadBalancerDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).LightsailConn()

	out, err := conn.DeleteLoadBalancerWithContext(ctx, &lightsail.DeleteLoadBalancerInput{
		LoadBalancerName: aws.String(d.Id()),
	})

	if err != nil {
		return create.DiagError(names.Lightsail, lightsail.OperationTypeCreateLoadBalancer, ResLoadBalancer, d.Get("name").(string), err)
	}

	if len(out.Operations) == 0 {
		return create.DiagError(names.Lightsail, lightsail.OperationTypeCreateLoadBalancer, ResLoadBalancer, d.Get("name").(string), errors.New("No operations found for Create Load Balancer request"))
	}

	op := out.Operations[0]

	err = waitOperation(ctx, conn, op.Id)
	if err != nil {
		return create.DiagError(names.Lightsail, lightsail.OperationTypeCreateLoadBalancer, ResLoadBalancer, d.Get("name").(string), errors.New("Error waiting for Create Load Balancer request operation"))
	}

	return nil
}
