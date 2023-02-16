package ec2

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceTrafficMirrorTarget() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceTrafficMirrorTargetCreate,
		ReadWithoutTimeout:   resourceTrafficMirrorTargetRead,
		UpdateWithoutTimeout: resourceTrafficMirrorTargetUpdate,
		DeleteWithoutTimeout: resourceTrafficMirrorTargetDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		CustomizeDiff: verify.SetTagsDiff,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"gateway_load_balancer_endpoint_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				ExactlyOneOf: []string{
					"gateway_load_balancer_endpoint_id",
					"network_interface_id",
					"network_load_balancer_arn",
				},
			},
			"network_interface_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				ExactlyOneOf: []string{
					"gateway_load_balancer_endpoint_id",
					"network_interface_id",
					"network_load_balancer_arn",
				},
			},
			"network_load_balancer_arn": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				ExactlyOneOf: []string{
					"gateway_load_balancer_endpoint_id",
					"network_interface_id",
					"network_load_balancer_arn",
				},
				ValidateFunc: verify.ValidARN,
			},
			"owner_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},
	}
}

func resourceTrafficMirrorTargetCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := &ec2.CreateTrafficMirrorTargetInput{}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("gateway_load_balancer_endpoint_id"); ok {
		input.GatewayLoadBalancerEndpointId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("network_interface_id"); ok {
		input.NetworkInterfaceId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("network_load_balancer_arn"); ok {
		input.NetworkLoadBalancerArn = aws.String(v.(string))
	}

	if len(tags) > 0 {
		input.TagSpecifications = tagSpecificationsFromKeyValueTags(tags, ec2.ResourceTypeTrafficMirrorTarget)
	}

	output, err := conn.CreateTrafficMirrorTargetWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EC2 Traffic Mirror Target: %s", err)
	}

	d.SetId(aws.StringValue(output.TrafficMirrorTarget.TrafficMirrorTargetId))

	return append(diags, resourceTrafficMirrorTargetRead(ctx, d, meta)...)
}

func resourceTrafficMirrorTargetRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	target, err := FindTrafficMirrorTargetByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 Traffic Mirror Target %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 Traffic Mirror Target (%s): %s", d.Id(), err)
	}

	ownerID := aws.StringValue(target.OwnerId)
	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   ec2.ServiceName,
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: ownerID,
		Resource:  fmt.Sprintf("traffic-mirror-target/%s", d.Id()),
	}.String()
	d.Set("arn", arn)
	d.Set("description", target.Description)
	d.Set("gateway_load_balancer_endpoint_id", target.GatewayLoadBalancerEndpointId)
	d.Set("network_interface_id", target.NetworkInterfaceId)
	d.Set("network_load_balancer_arn", target.NetworkLoadBalancerArn)
	d.Set("owner_id", ownerID)

	tags := KeyValueTags(target.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags_all: %s", err)
	}

	return diags
}

func resourceTrafficMirrorTargetUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn()

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(ctx, conn, d.Id(), o, n); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating EC2 Traffic Mirror Target (%s) tags: %s", d.Id(), err)
		}
	}

	return append(diags, resourceTrafficMirrorTargetRead(ctx, d, meta)...)
}

func resourceTrafficMirrorTargetDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn()

	log.Printf("[DEBUG] Deleting EC2 Traffic Mirror Target: %s", d.Id())
	_, err := conn.DeleteTrafficMirrorTargetWithContext(ctx, &ec2.DeleteTrafficMirrorTargetInput{
		TrafficMirrorTargetId: aws.String(d.Id()),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting EC2 Traffic Mirror Target (%s): %s", d.Id(), err)
	}

	return diags
}
