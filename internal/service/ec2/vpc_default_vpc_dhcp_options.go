// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_default_vpc_dhcp_options", name="DHCP Options")
// @Tags(identifierAttribute="id")
// @Testing(tagsTest=false)
func resourceDefaultVPCDHCPOptions() *schema.Resource {
	//lintignore:R011
	return &schema.Resource{
		CreateWithoutTimeout: resourceDefaultVPCDHCPOptionsCreate,
		ReadWithoutTimeout:   resourceVPCDHCPOptionsRead,
		UpdateWithoutTimeout: resourceVPCDHCPOptionsUpdate,
		DeleteWithoutTimeout: schema.NoopContext,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		CustomizeDiff: verify.SetTagsDiff,

		// Keep in sync with aws_vpc_dhcp_options' schema with the following changes:
		//   - domain_name is Computed-only
		//   - domain_name_servers is Computed-only and is TypeString
		//   - ipv6_address_preferred_lease_time is Computed-only and is TypeString
		//   - netbios_name_servers is Computed-only and is TypeString
		//   - netbios_node_type is Computed-only
		//   - ntp_servers is Computed-only and is TypeString
		//   - owner_id is Optional/Computed
		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDomainName: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"domain_name_servers": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ipv6_address_preferred_lease_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"netbios_name_servers": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"netbios_node_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ntp_servers": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrOwnerID: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourceDefaultVPCDHCPOptionsCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	input := &ec2.DescribeDhcpOptionsInput{}

	input.Filters = append(input.Filters,
		newFilter(names.AttrKey, []string{"domain-name"}),
		newFilter(names.AttrValue, []string{meta.(*conns.AWSClient).EC2RegionalPrivateDNSSuffix(ctx)}),
		newFilter(names.AttrKey, []string{"domain-name-servers"}),
		newFilter(names.AttrValue, []string{"AmazonProvidedDNS"}),
	)

	if v, ok := d.GetOk(names.AttrOwnerID); ok {
		input.Filters = append(input.Filters, newAttributeFilterList(map[string]string{
			"owner-id": v.(string),
		})...)
	}

	dhcpOptions, err := findDHCPOptions(ctx, conn, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 Default DHCP Options Set: %s", err)
	}

	d.SetId(aws.ToString(dhcpOptions.DhcpOptionsId))

	return append(diags, resourceVPCDHCPOptionsUpdate(ctx, d, meta)...)
}
