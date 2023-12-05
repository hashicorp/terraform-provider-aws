// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_ec2_transit_gateway_multicast_domain", name="Transit Gateway Multicast Domain")
// @Tags(identifierAttribute="id")
func ResourceTransitGatewayMulticastDomain() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceTransitGatewayMulticastDomainCreate,
		ReadWithoutTimeout:   resourceTransitGatewayMulticastDomainRead,
		UpdateWithoutTimeout: resourceTransitGatewayMulticastDomainUpdate,
		DeleteWithoutTimeout: resourceTransitGatewayMulticastDomainDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		CustomizeDiff: verify.SetTagsDiff,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"auto_accept_shared_associations": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      ec2.AutoAcceptSharedAssociationsValueDisable,
				ValidateFunc: validation.StringInSlice(ec2.AutoAcceptSharedAssociationsValue_Values(), false),
			},
			"igmpv2_support": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      ec2.Igmpv2SupportValueDisable,
				ValidateFunc: validation.StringInSlice(ec2.Igmpv2SupportValue_Values(), false),
			},
			"owner_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"static_sources_support": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      ec2.StaticSourcesSupportValueDisable,
				ValidateFunc: validation.StringInSlice(ec2.StaticSourcesSupportValue_Values(), false),
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"transit_gateway_id": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
		},
	}
}

func resourceTransitGatewayMulticastDomainCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).EC2Conn(ctx)

	input := &ec2.CreateTransitGatewayMulticastDomainInput{
		Options: &ec2.CreateTransitGatewayMulticastDomainRequestOptions{
			AutoAcceptSharedAssociations: aws.String(d.Get("auto_accept_shared_associations").(string)),
			Igmpv2Support:                aws.String(d.Get("igmpv2_support").(string)),
			StaticSourcesSupport:         aws.String(d.Get("static_sources_support").(string)),
		},
		TagSpecifications: getTagSpecificationsIn(ctx, ec2.ResourceTypeTransitGatewayMulticastDomain),
		TransitGatewayId:  aws.String(d.Get("transit_gateway_id").(string)),
	}

	log.Printf("[DEBUG] Creating EC2 Transit Gateway Multicast Domain: %s", input)
	output, err := conn.CreateTransitGatewayMulticastDomainWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EC2 Transit Gateway Multicast Domain: %s", err)
	}

	d.SetId(aws.StringValue(output.TransitGatewayMulticastDomain.TransitGatewayMulticastDomainId))

	if _, err := WaitTransitGatewayMulticastDomainCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EC2 Transit Gateway Multicast Domain (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceTransitGatewayMulticastDomainRead(ctx, d, meta)...)
}

func resourceTransitGatewayMulticastDomainRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).EC2Conn(ctx)

	multicastDomain, err := FindTransitGatewayMulticastDomainByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 Transit Gateway Multicast Domain %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 Transit Gateway Multicast Domain (%s): %s", d.Id(), err)
	}

	d.Set("arn", multicastDomain.TransitGatewayMulticastDomainArn)
	d.Set("auto_accept_shared_associations", multicastDomain.Options.AutoAcceptSharedAssociations)
	d.Set("igmpv2_support", multicastDomain.Options.Igmpv2Support)
	d.Set("owner_id", multicastDomain.OwnerId)
	d.Set("static_sources_support", multicastDomain.Options.StaticSourcesSupport)
	d.Set("transit_gateway_id", multicastDomain.TransitGatewayId)

	setTagsOut(ctx, multicastDomain.Tags)

	return diags
}

func resourceTransitGatewayMulticastDomainUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Tags only.
	return resourceTransitGatewayMulticastDomainRead(ctx, d, meta)
}

func resourceTransitGatewayMulticastDomainDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).EC2Conn(ctx)

	groups, err := FindTransitGatewayMulticastGroups(ctx, conn, &ec2.SearchTransitGatewayMulticastGroupsInput{
		TransitGatewayMulticastDomainId: aws.String(d.Id()),
	})

	if tfresource.NotFound(err) {
		err = nil
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing EC2 Transit Gateway Multicast Groups (%s): %s", d.Id(), err)
	}

	for _, v := range groups {
		if aws.BoolValue(v.GroupMember) {
			err := deregisterTransitGatewayMulticastGroupMember(ctx, conn, d.Id(), aws.StringValue(v.GroupIpAddress), aws.StringValue(v.NetworkInterfaceId))

			if err != nil {
				diags = sdkdiag.AppendFromErr(diags, err)
			}
		} else if aws.BoolValue(v.GroupSource) {
			err := deregisterTransitGatewayMulticastGroupSource(ctx, conn, d.Id(), aws.StringValue(v.GroupIpAddress), aws.StringValue(v.NetworkInterfaceId))

			if err != nil {
				diags = sdkdiag.AppendFromErr(diags, err)
			}
		}
	}

	if diags.HasError() {
		return diags
	}

	associations, err := FindTransitGatewayMulticastDomainAssociations(ctx, conn, &ec2.GetTransitGatewayMulticastDomainAssociationsInput{
		TransitGatewayMulticastDomainId: aws.String(d.Id()),
	})

	if tfresource.NotFound(err) {
		err = nil
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing EC2 Transit Gateway Multicast Domain Associations (%s): %s", d.Id(), err)
	}

	for _, v := range associations {
		err := disassociateTransitGatewayMulticastDomain(ctx, conn, d.Id(), aws.StringValue(v.TransitGatewayAttachmentId), aws.StringValue(v.Subnet.SubnetId), d.Timeout(schema.TimeoutDelete))

		if err != nil {
			diags = sdkdiag.AppendFromErr(diags, err)
		}
	}

	if diags.HasError() {
		return diags
	}

	log.Printf("[DEBUG] Deleting EC2 Transit Gateway Multicast Domain: %s", d.Id())
	_, err = conn.DeleteTransitGatewayMulticastDomainWithContext(ctx, &ec2.DeleteTransitGatewayMulticastDomainInput{
		TransitGatewayMulticastDomainId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidTransitGatewayMulticastDomainIdNotFound) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting EC2 Transit Gateway Multicast Domain (%s): %s", d.Id(), err)
	}

	if _, err := WaitTransitGatewayMulticastDomainDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EC2 Transit Gateway Multicast Domain (%s) delete: %s", d.Id(), err)
	}

	return diags
}
