// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"time"

	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_vpc_peering_connection_accepter", name="VPC Peering Connection")
// @Tags(identifierAttribute="id")
// @Testing(tagsTest=false)
func resourceVPCPeeringConnectionAccepter() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceVPCPeeringAccepterCreate,
		ReadWithoutTimeout:   resourceVPCPeeringConnectionRead,
		UpdateWithoutTimeout: resourceVPCPeeringConnectionUpdate,
		DeleteWithoutTimeout: schema.NoopContext,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(1 * time.Minute),
			Update: schema.DefaultTimeout(1 * time.Minute),
		},

		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, m interface{}) (result []*schema.ResourceData, err error) {
				d.Set("vpc_peering_connection_id", d.Id())

				return []*schema.ResourceData{d}, nil
			},
		},

		// Keep in sync with aws_vpc_peering_connections's schema with the following changes:
		//   - peer_owner_id is Computed-only
		//   - peer_region is Computed-only
		//   - peer_vpc_id is Computed-only
		//   - vpc_id is Computed-only
		// and additions:
		//   - vpc_peering_connection_id Required/ForceNew
		Schema: map[string]*schema.Schema{
			"accept_status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"accepter": vpcPeeringConnectionOptionsSchema,
			"auto_accept": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"peer_owner_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"peer_region": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"peer_vpc_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"requester":       vpcPeeringConnectionOptionsSchema,
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			names.AttrVPCID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"vpc_peering_connection_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceVPCPeeringAccepterCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	vpcPeeringConnectionID := d.Get("vpc_peering_connection_id").(string)
	vpcPeeringConnection, err := findVPCPeeringConnectionByID(ctx, conn, vpcPeeringConnectionID)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 VPC Peering Connection (%s): %s", vpcPeeringConnectionID, err)
	}

	d.SetId(vpcPeeringConnectionID)

	if _, ok := d.GetOk("auto_accept"); ok && vpcPeeringConnection.Status.Code == awstypes.VpcPeeringConnectionStateReasonCodePendingAcceptance {
		vpcPeeringConnection, err = acceptVPCPeeringConnection(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate))

		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	if err := modifyVPCPeeringConnectionOptions(ctx, conn, d, vpcPeeringConnection, true); err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	if err := createTags(ctx, conn, d.Id(), getTagsIn(ctx)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting EC2 VPC Peering Connection (%s) tags: %s", d.Id(), err)
	}

	return append(diags, resourceVPCPeeringConnectionRead(ctx, d, meta)...)
}
