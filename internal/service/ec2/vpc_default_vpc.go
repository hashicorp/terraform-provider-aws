// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
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

// @SDKResource("aws_default_vpc", name="Default VPC")
// @Tags(identifierAttribute="id")
// @Testing(tagsTest=false)
func resourceDefaultVPC() *schema.Resource {
	//lintignore:R011
	return &schema.Resource{
		CreateWithoutTimeout: resourceDefaultVPCCreate,
		ReadWithoutTimeout:   resourceVPCRead,
		UpdateWithoutTimeout: resourceVPCUpdate,
		DeleteWithoutTimeout: resourceDefaultVPCDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceVPCImport,
		},

		CustomizeDiff: verify.SetTagsDiff,

		SchemaVersion: 1,
		MigrateState:  vpcMigrateState,

		// Keep in sync with aws_vpc's schema with the following changes:
		//   - cidr_block is Computed-only
		//   - enable_dns_hostnames is not Computed has a Default of true
		//   - instance_tenancy is Computed-only
		//   - ipv4_ipam_pool_id is omitted as it's not set in resourceVPCRead
		//   - ipv4_netmask_length is omitted as it's not set in resourceVPCRead
		// and additions:
		//   - existing_default_vpc Computed-only, set in resourceDefaultVPCCreate
		//   - force_destroy Optional
		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"assign_generated_ipv6_cidr_block": {
				Type:          schema.TypeBool,
				Optional:      true,
				ConflictsWith: []string{"ipv6_ipam_pool_id"},
			},
			names.AttrCIDRBlock: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"default_network_acl_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"default_route_table_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"default_security_group_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"dhcp_options_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"enable_dns_hostnames": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"enable_dns_support": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"enable_network_address_usage_metrics": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"existing_default_vpc": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			names.AttrForceDestroy: {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"instance_tenancy": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ipv6_association_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ipv6_cidr_block": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"ipv6_netmask_length", "assign_generated_ipv6_cidr_block"},
				RequiredWith:  []string{"ipv6_ipam_pool_id"},
				ValidateFunc: validation.All(
					verify.ValidIPv6CIDRNetworkAddress,
					validation.IsCIDRNetwork(vpcCIDRMaxIPv6Netmask, vpcCIDRMaxIPv6Netmask)),
			},
			"ipv6_cidr_block_network_border_group": {
				Type:         schema.TypeString,
				Computed:     true,
				Optional:     true,
				RequiredWith: []string{"assign_generated_ipv6_cidr_block"},
			},
			"ipv6_ipam_pool_id": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"assign_generated_ipv6_cidr_block"},
			},
			"ipv6_netmask_length": {
				Type:          schema.TypeInt,
				Optional:      true,
				ValidateFunc:  validation.IntInSlice([]int{vpcCIDRMaxIPv6Netmask}),
				ConflictsWith: []string{"ipv6_cidr_block"},
				RequiredWith:  []string{"ipv6_ipam_pool_id"},
			},
			"main_route_table_id": {
				Type:     schema.TypeString,
				Computed: true,
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

func resourceDefaultVPCCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics { // nosemgrep:ci.semgrep.tags.calling-UpdateTags-in-resource-create
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	input := &ec2.DescribeVpcsInput{
		Filters: newAttributeFilterList(
			map[string]string{
				"isDefault": "true",
			},
		),
	}

	vpcInfo := &vpcInfo{}
	vpc, err := findVPC(ctx, conn, input)

	if err == nil {
		d.SetId(aws.ToString(vpc.VpcId))
		d.Set("existing_default_vpc", true)

		vpcInfo.vpc = vpc

		if v, err := findVPCAttribute(ctx, conn, d.Id(), awstypes.VpcAttributeNameEnableDnsHostnames); err != nil {
			return sdkdiag.AppendErrorf(diags, "reading EC2 VPC (%s) Attribute (%s): %s", d.Id(), awstypes.VpcAttributeNameEnableDnsHostnames, err)
		} else {
			vpcInfo.enableDnsHostnames = v
		}

		if v, err := findVPCAttribute(ctx, conn, d.Id(), awstypes.VpcAttributeNameEnableDnsSupport); err != nil {
			return sdkdiag.AppendErrorf(diags, "reading EC2 VPC (%s) Attribute (%s): %s", d.Id(), awstypes.VpcAttributeNameEnableDnsSupport, err)
		} else {
			vpcInfo.enableDnsSupport = v
		}
		if v, err := findVPCAttribute(ctx, conn, d.Id(), awstypes.VpcAttributeNameEnableNetworkAddressUsageMetrics); err != nil {
			return sdkdiag.AppendErrorf(diags, "reading EC2 VPC (%s) Attribute (%s): %s", d.Id(), awstypes.VpcAttributeNameEnableNetworkAddressUsageMetrics, err)
		} else {
			vpcInfo.enableNetworkAddressUsageMetrics = v
		}
	} else if tfresource.NotFound(err) {
		input := &ec2.CreateDefaultVpcInput{}

		output, err := conn.CreateDefaultVpc(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "creating EC2 Default VPC: %s", err)
		}

		vpc = output.Vpc

		d.SetId(aws.ToString(vpc.VpcId))
		d.Set("existing_default_vpc", false)

		vpc, err = waitVPCCreated(ctx, conn, d.Id())

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for EC2 Default VPC (%s) create: %s", d.Id(), err)
		}

		vpcInfo.vpc = vpc
		vpcInfo.enableDnsHostnames = true
		vpcInfo.enableDnsSupport = true
		vpcInfo.enableNetworkAddressUsageMetrics = false
	} else {
		return sdkdiag.AppendErrorf(diags, "reading EC2 Default VPC (%s): %s", d.Id(), err)
	}

	if err := modifyVPCAttributesOnCreate(ctx, conn, d, vpcInfo); err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EC2 Default VPC: %s", err)
	}

	// Configure IPv6.
	var associationID, oldIPv6PoolID, oldIPv6CIDRBlock, oldIPv6CIDRBlockNetworkBorderGroup string
	var oldAssignGeneratedIPv6CIDRBlock bool

	if v := defaultIPv6CIDRBlockAssociation(vpcInfo.vpc, ""); v != nil {
		associationID = aws.ToString(v.AssociationId)
		oldIPv6CIDRBlock = aws.ToString(v.Ipv6CidrBlock)
		oldIPv6CIDRBlockNetworkBorderGroup = aws.ToString(v.NetworkBorderGroup)
		if ipv6PoolID := aws.ToString(v.Ipv6Pool); ipv6PoolID == amazonIPv6PoolID {
			oldAssignGeneratedIPv6CIDRBlock = true
		} else {
			oldIPv6PoolID = ipv6PoolID
		}
	}

	if newAssignGeneratedIPv6CIDRBlock, newIPv6CIDRBlockNetworkBorderGroup := d.Get("assign_generated_ipv6_cidr_block").(bool), d.Get("ipv6_cidr_block_network_border_group").(string); oldAssignGeneratedIPv6CIDRBlock != newAssignGeneratedIPv6CIDRBlock || (newIPv6CIDRBlockNetworkBorderGroup != "" && oldIPv6CIDRBlockNetworkBorderGroup != newIPv6CIDRBlockNetworkBorderGroup) {
		associationID, err := modifyVPCIPv6CIDRBlockAssociation(ctx, conn, d.Id(),
			associationID,
			newAssignGeneratedIPv6CIDRBlock,
			"",
			"",
			0,
			newIPv6CIDRBlockNetworkBorderGroup)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "creating EC2 Default VPC: %s", err)
		}

		d.Set("ipv6_association_id", associationID)
	}

	if newAssignGeneratedIPv6CIDRBlock, newIPv6CIDRBlock, newIPv6PoolID := d.Get("assign_generated_ipv6_cidr_block").(bool), d.Get("ipv6_cidr_block").(string), d.Get("ipv6_ipam_pool_id").(string); !newAssignGeneratedIPv6CIDRBlock && (oldIPv6CIDRBlock != newIPv6CIDRBlock || oldIPv6PoolID != newIPv6PoolID) {
		associationID, err := modifyVPCIPv6CIDRBlockAssociation(ctx, conn, d.Id(),
			associationID,
			false,
			newIPv6CIDRBlock,
			newIPv6PoolID,
			d.Get("ipv6_netmask_length").(int),
			"")

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "creating EC2 Default VPC: %s", err)
		}

		d.Set("ipv6_association_id", associationID)
	}

	// Configure tags.
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig
	newTags := keyValueTags(ctx, getTagsIn(ctx))
	oldTags := keyValueTags(ctx, vpc.Tags).IgnoreSystem(names.EC2).IgnoreConfig(ignoreTagsConfig)

	if !oldTags.Equal(newTags) {
		if err := updateTags(ctx, conn, d.Id(), oldTags, newTags); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating EC2 Default VPC (%s) tags: %s", d.Id(), err)
		}
	}

	return append(diags, resourceVPCRead(ctx, d, meta)...)
}

func resourceDefaultVPCDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	if d.Get(names.AttrForceDestroy).(bool) {
		return append(diags, resourceVPCDelete(ctx, d, meta)...)
	}

	log.Printf("[WARN] EC2 Default VPC (%s) not deleted, removing from state", d.Id())

	return diags
}
