// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_ec2_traffic_mirror_filter", name="Traffic Mirror Filter")
// @Tags(identifierAttribute="id")
// @Testing(tagsTest=false)
func resourceTrafficMirrorFilter() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceTrafficMirrorFilterCreate,
		ReadWithoutTimeout:   resourceTrafficMirrorFilterRead,
		UpdateWithoutTimeout: resourceTrafficMirrorFilterUpdate,
		DeleteWithoutTimeout: resourceTrafficMirrorFilterDelete,

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
			"network_services": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type:             schema.TypeString,
					ValidateDiagFunc: enum.Validate[awstypes.TrafficMirrorNetworkService](),
					ValidateFunc: validation.StringInSlice([]string{
						"amazon-dns",
					}, false),
				},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourceTrafficMirrorFilterCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	input := &ec2.CreateTrafficMirrorFilterInput{
		ClientToken:       aws.String(id.UniqueId()),
		TagSpecifications: getTagSpecificationsIn(ctx, awstypes.ResourceTypeTrafficMirrorFilter),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	output, err := conn.CreateTrafficMirrorFilter(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EC2 Traffic Mirror Filter: %s", err)
	}

	d.SetId(aws.ToString(output.TrafficMirrorFilter.TrafficMirrorFilterId))

	if v, ok := d.GetOk("network_services"); ok && v.(*schema.Set).Len() > 0 {
		input := &ec2.ModifyTrafficMirrorFilterNetworkServicesInput{
			AddNetworkServices:    flex.ExpandStringyValueSet[awstypes.TrafficMirrorNetworkService](v.(*schema.Set)),
			TrafficMirrorFilterId: aws.String(d.Id()),
		}

		_, err := conn.ModifyTrafficMirrorFilterNetworkServices(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating EC2 Traffic Mirror Filter (%s) network services: %s", d.Id(), err)
		}
	}

	return append(diags, resourceTrafficMirrorFilterRead(ctx, d, meta)...)
}

func resourceTrafficMirrorFilterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	trafficMirrorFilter, err := findTrafficMirrorFilterByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 Traffic Mirror Filter %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 Traffic Mirror Filter (%s): %s", d.Id(), err)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "ec2",
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  "traffic-mirror-filter/" + d.Id(),
	}.String()
	d.Set(names.AttrARN, arn)
	d.Set(names.AttrDescription, trafficMirrorFilter.Description)
	d.Set("network_services", trafficMirrorFilter.NetworkServices)

	setTagsOut(ctx, trafficMirrorFilter.Tags)

	return diags
}

func resourceTrafficMirrorFilterUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	if d.HasChange("network_services") {
		input := &ec2.ModifyTrafficMirrorFilterNetworkServicesInput{
			TrafficMirrorFilterId: aws.String(d.Id()),
		}

		o, n := d.GetChange("network_services")
		add := n.(*schema.Set).Difference(o.(*schema.Set))
		if add.Len() > 0 {
			input.AddNetworkServices = flex.ExpandStringyValueSet[awstypes.TrafficMirrorNetworkService](add)
		}
		del := o.(*schema.Set).Difference(n.(*schema.Set))
		if del.Len() > 0 {
			input.RemoveNetworkServices = flex.ExpandStringyValueSet[awstypes.TrafficMirrorNetworkService](del)
		}

		_, err := conn.ModifyTrafficMirrorFilterNetworkServices(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating EC2 Traffic Mirror Filter (%s) network services: %s", d.Id(), err)
		}
	}

	return append(diags, resourceTrafficMirrorFilterRead(ctx, d, meta)...)
}

func resourceTrafficMirrorFilterDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	log.Printf("[DEBUG] Deleting EC2 Traffic Mirror Filter: %s", d.Id())
	_, err := conn.DeleteTrafficMirrorFilter(ctx, &ec2.DeleteTrafficMirrorFilterInput{
		TrafficMirrorFilterId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidTrafficMirrorFilterIdNotFound) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting EC2 Traffic Mirror Filter (%s): %s", d.Id(), err)
	}

	return diags
}
