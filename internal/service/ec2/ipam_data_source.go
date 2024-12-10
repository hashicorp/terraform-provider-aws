// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_vpc_ipam", name="IPAM")
// @Tags
// @Testing(tagsTest=false)
func dataSourceIPAM() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceIPAMRead,

		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"default_resource_discovery_association_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"default_resource_discovery_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"enable_private_gua": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"ipam_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"ipam_region": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrFilter: customFiltersSchema(),
			"owner_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"public_default_scope_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"private_default_scope_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"resource_discovery_association_count": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"scope_count": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"tier": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrState: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceIPAMRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	input := &ec2.DescribeIpamsInput{}

	if v, ok := d.GetOk("ipam_id"); ok {
		input.IpamIds = []string{v.(string)}
	}

	input.Filters = append(input.Filters, newCustomFilterList(
		d.Get(names.AttrFilter).(*schema.Set),
	)...)

	if len(input.Filters) == 0 {
		input.Filters = nil
	}

	ipam, err := findIPAM(ctx, conn, input)

	if err != nil {
		return sdkdiag.AppendFromErr(diags, tfresource.SingularDataSourceFindError("IPAM", err))
	}

	d.SetId(aws.ToString(ipam.IpamId))
	d.Set("default_resource_discovery_association_id", ipam.DefaultResourceDiscoveryAssociationId)
	d.Set("default_resource_discovery_id", ipam.DefaultResourceDiscoveryId)
	d.Set("enable_private_gua", ipam.EnablePrivateGua)
	d.Set("ipam_region", ipam.IpamRegion)
	d.Set(names.AttrARN, ipam.IpamArn)
	d.Set("owner_id", ipam.OwnerId)
	d.Set(names.AttrDescription, ipam.Description)
	d.Set("public_default_scope_id", ipam.PublicDefaultScopeId)
	d.Set("private_default_scope_id", ipam.PrivateDefaultScopeId)
	d.Set("resource_discovery_association_count", ipam.ResourceDiscoveryAssociationCount)
	d.Set("scope_count", ipam.ScopeCount)
	d.Set("tier", ipam.Tier)
	d.Set(names.AttrState, ipam.State)

	setTagsOut(ctx, ipam.Tags)

	return diags
}
