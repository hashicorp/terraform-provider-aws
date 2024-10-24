// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_ec2_traffic_mirror_target", name="Traffic Mirror Target")
// @Tags(identifierAttribute="id")
// @Testing(tagsTest=false)
func resourceTrafficMirrorTarget() *schema.Resource {
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
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
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
					names.AttrNetworkInterfaceID,
					"network_load_balancer_arn",
				},
			},
			names.AttrNetworkInterfaceID: {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				ExactlyOneOf: []string{
					"gateway_load_balancer_endpoint_id",
					names.AttrNetworkInterfaceID,
					"network_load_balancer_arn",
				},
			},
			"network_load_balancer_arn": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				ExactlyOneOf: []string{
					"gateway_load_balancer_endpoint_id",
					names.AttrNetworkInterfaceID,
					"network_load_balancer_arn",
				},
				ValidateFunc: verify.ValidARN,
			},
			names.AttrOwnerID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourceTrafficMirrorTargetCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	input := &ec2.CreateTrafficMirrorTargetInput{
		ClientToken:       aws.String(id.UniqueId()),
		TagSpecifications: getTagSpecificationsIn(ctx, awstypes.ResourceTypeTrafficMirrorTarget),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("gateway_load_balancer_endpoint_id"); ok {
		input.GatewayLoadBalancerEndpointId = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrNetworkInterfaceID); ok {
		input.NetworkInterfaceId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("network_load_balancer_arn"); ok {
		input.NetworkLoadBalancerArn = aws.String(v.(string))
	}

	output, err := conn.CreateTrafficMirrorTarget(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EC2 Traffic Mirror Target: %s", err)
	}

	d.SetId(aws.ToString(output.TrafficMirrorTarget.TrafficMirrorTargetId))

	return append(diags, resourceTrafficMirrorTargetRead(ctx, d, meta)...)
}

func resourceTrafficMirrorTargetRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	target, err := findTrafficMirrorTargetByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 Traffic Mirror Target %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 Traffic Mirror Target (%s): %s", d.Id(), err)
	}

	ownerID := aws.ToString(target.OwnerId)
	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   names.EC2,
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: ownerID,
		Resource:  fmt.Sprintf("traffic-mirror-target/%s", d.Id()),
	}.String()
	d.Set(names.AttrARN, arn)
	d.Set(names.AttrDescription, target.Description)
	d.Set("gateway_load_balancer_endpoint_id", target.GatewayLoadBalancerEndpointId)
	d.Set(names.AttrNetworkInterfaceID, target.NetworkInterfaceId)
	d.Set("network_load_balancer_arn", target.NetworkLoadBalancerArn)
	d.Set(names.AttrOwnerID, ownerID)

	setTagsOut(ctx, target.Tags)

	return diags
}

func resourceTrafficMirrorTargetUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// Tags only

	return append(diags, resourceTrafficMirrorTargetRead(ctx, d, meta)...)
}

func resourceTrafficMirrorTargetDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	log.Printf("[DEBUG] Deleting EC2 Traffic Mirror Target: %s", d.Id())
	_, err := conn.DeleteTrafficMirrorTarget(ctx, &ec2.DeleteTrafficMirrorTargetInput{
		TrafficMirrorTargetId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidTrafficMirrorTargetIdNotFound) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting EC2 Traffic Mirror Target (%s): %s", d.Id(), err)
	}

	return diags
}
