package ec2

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_ec2_instance_connect_endpoint")
// @Tags(identifierAttribute="id")
func ResourceInstanceConnectEndpoint() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceInstanceConnectEndpointCreate,
		ReadWithoutTimeout:   resourceInstanceConnectEndpointRead,
		UpdateWithoutTimeout: resourceInstanceConnectEndpointUpdate,
		DeleteWithoutTimeout: resourceInstanceConnectEndpointDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(1 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"availability_zone": {
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
			"fips_dns_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"subnet_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"network_interface_ids": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"owner_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"security_group_ids": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"state": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"preserve_client_ip": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
				Default:  true,
			},
			"vpc_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceInstanceConnectEndpointCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn(ctx)

	input := &ec2.CreateInstanceConnectEndpointInput{
		SubnetId:          aws.String(d.Get("subnet_id").(string)),
		TagSpecifications: getTagSpecificationsIn(ctx, ec2.ResourceTypeInstanceConnectEndpoint),
	}

	if attr, ok := d.GetOk("autoscaling_groups"); ok {
		input.SecurityGroupIds = flex.ExpandStringSet(attr.(*schema.Set))
	}

	if v, ok := d.GetOk("disable_rollback"); ok {
		input.PreserveClientIp = aws.Bool(v.(bool))
	}

	output, err := conn.CreateInstanceConnectEndpointWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EC2 Instance Connect Endpoint: %s", err)
	}

	d.SetId(aws.StringValue(output.InstanceConnectEndpoint.InstanceConnectEndpointId))

	return append(diags, resourceInstanceConnectEndpointRead(ctx, d, meta)...)
}

func resourceInstanceConnectEndpointRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn(ctx)

	output, err := conn.DescribeInstanceConnectEndpointsWithContext(ctx, &ec2.DescribeInstanceConnectEndpointsInput{
		InstanceConnectEndpointIds: []*string{aws.String(d.Id())},
	})

	if !d.IsNewResource() && tfresource.NotFound(err) {
		create.LogNotFoundRemoveState(names.EC2, create.ErrActionReading, ResInstanceConnectEndpointState, d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return create.DiagError(names.EC2, create.ErrActionReading, ResInstanceConnectEndpointState, d.Id(), err)
	}

	d.Set("id", output.InstanceConnectEndpoints[0].InstanceConnectEndpointId)
	d.Set("arn", output.InstanceConnectEndpoints[0].InstanceConnectEndpointArn)

	d.Set("availability_zone", output.InstanceConnectEndpoints[0].AvailabilityZone)
	d.Set("created_at", output.InstanceConnectEndpoints[0].CreatedAt)
	d.Set("dns_name", output.InstanceConnectEndpoints[0].DnsName)
	d.Set("fips_dns_name", output.InstanceConnectEndpoints[0].FipsDnsName)

	if err := d.Set("network_interface_ids", flex.FlattenStringSet(output.InstanceConnectEndpoints[0].NetworkInterfaceIds)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting network_interface_ids for EC2 Instance Connect Endpoint (%s): %s", d.Id(), err)
	}

	d.Set("owner_id", output.InstanceConnectEndpoints[0].OwnerId)
	d.Set("preserve_client_ip", output.InstanceConnectEndpoints[0].PreserveClientIp)

	if err := d.Set("security_group_ids", flex.FlattenStringSet(output.InstanceConnectEndpoints[0].SecurityGroupIds)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting security_group_ids for EC2 Instance Connect Endpoint (%s): %s", d.Id(), err)
	}

	d.Set("state", output.InstanceConnectEndpoints[0].State)
	d.Set("subnet_id", output.InstanceConnectEndpoints[0].SubnetId)
	d.Set("vpc_id", output.InstanceConnectEndpoints[0].VpcId)
	setTagsOut(ctx, output.InstanceConnectEndpoints[0].Tags)

	return diags
}

func resourceInstanceConnectEndpointUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// Tags only.
	return append(diags, resourceInstanceConnectEndpointRead(ctx, d, meta)...)
}

func resourceInstanceConnectEndpointDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn(ctx)

	log.Printf("[DEBUG] Deleting EC2 Instance Connect Endpoint: %s", d.Id())
	_, err := conn.DeleteInstanceConnectEndpointWithContext(ctx, &ec2.DeleteInstanceConnectEndpointInput{
		InstanceConnectEndpointId: aws.String(d.Id()),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting EC2 Instance Connect Endpoint (%s): %s", d.Id(), err)
	}

	return diags
}
