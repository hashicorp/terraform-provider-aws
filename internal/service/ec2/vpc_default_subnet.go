// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_default_subnet", name="Subnet")
// @Tags(identifierAttribute="id")
// @Testing(tagsTest=false)
func resourceDefaultSubnet() *schema.Resource {
	//lintignore:R011
	return &schema.Resource{
		CreateWithoutTimeout: resourceDefaultSubnetCreate,
		ReadWithoutTimeout:   resourceSubnetRead,
		UpdateWithoutTimeout: resourceSubnetUpdate,
		DeleteWithoutTimeout: resourceDefaultSubnetDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		CustomizeDiff: verify.SetTagsDiff,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(20 * time.Minute),
		},

		SchemaVersion: 1,
		MigrateState:  subnetMigrateState,

		// Keep in sync with aws_subnet's schema with the following changes:
		//   - availability_zone is Required/ForceNew
		//   - availability_zone_id is Computed-only
		//   - cidr_block is Computed-only
		//   - enable_lni_at_device_index is Computed-only
		//   - ipv6_cidr_block is Optional/Computed as it's automatically assigned if ipv6_native = true
		//   - map_public_ip_on_launch has a Default of true
		//   - outpost_arn is Computed-only
		//   - vpc_id is Computed-only
		// and additions:
		//   - existing_default_subnet Computed-only, set in resourceDefaultSubnetCreate
		//   - force_destroy Optional
		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"assign_ipv6_address_on_creation": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			names.AttrAvailabilityZone: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"availability_zone_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrCIDRBlock: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"customer_owned_ipv4_pool": {
				Type:         schema.TypeString,
				Optional:     true,
				RequiredWith: []string{"map_customer_owned_ip_on_launch"},
			},
			"enable_dns64": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"enable_lni_at_device_index": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"enable_resource_name_dns_aaaa_record_on_launch": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"enable_resource_name_dns_a_record_on_launch": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"existing_default_subnet": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			names.AttrForceDestroy: {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"ipv6_cidr_block": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: verify.ValidIPv6CIDRNetworkAddress,
			},
			"ipv6_cidr_block_association_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ipv6_native": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
				Default:  false,
			},
			"map_customer_owned_ip_on_launch": {
				Type:         schema.TypeBool,
				Optional:     true,
				RequiredWith: []string{"customer_owned_ipv4_pool", "outpost_arn"},
			},
			"map_public_ip_on_launch": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"outpost_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrOwnerID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"private_dns_hostname_type_on_launch": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ValidateDiagFunc: enum.Validate[awstypes.HostnameType](),
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			names.AttrVPCID: {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceDefaultSubnetCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics { // nosemgrep:ci.semgrep.tags.calling-UpdateTags-in-resource-create
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	availabilityZone := d.Get(names.AttrAvailabilityZone).(string)
	input := &ec2.DescribeSubnetsInput{
		Filters: newAttributeFilterList(
			map[string]string{
				"availabilityZone": availabilityZone,
				"defaultForAz":     "true",
			},
		),
	}

	var computedIPv6CIDRBlock bool
	subnet, err := findSubnet(ctx, conn, input)

	if err == nil {
		log.Printf("[INFO] Found existing EC2 Default Subnet (%s)", availabilityZone)
		d.SetId(aws.ToString(subnet.SubnetId))
		d.Set("existing_default_subnet", true)
	} else if tfresource.NotFound(err) {
		input := &ec2.CreateDefaultSubnetInput{
			AvailabilityZone: aws.String(availabilityZone),
		}

		var ipv6Native bool
		if v, ok := d.GetOk("ipv6_native"); ok {
			ipv6Native = v.(bool)
			input.Ipv6Native = aws.Bool(ipv6Native)
		}

		log.Printf("[DEBUG] Creating EC2 Default Subnet: %#v", input)
		output, err := conn.CreateDefaultSubnet(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "creating EC2 Default Subnet (%s): %s", availabilityZone, err)
		}

		subnet = output.Subnet

		d.SetId(aws.ToString(subnet.SubnetId))
		d.Set("existing_default_subnet", false)

		subnet, err = waitSubnetAvailable(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate))

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for EC2 Default Subnet (%s) create: %s", d.Id(), err)
		}

		// Creating an IPv6-native default subnets associates an IPv6 CIDR block.
		for i, v := range subnet.Ipv6CidrBlockAssociationSet {
			if v.Ipv6CidrBlockState.State == awstypes.SubnetCidrBlockStateCodeAssociating { //we can only ever have 1 IPv6 block associated at once
				associationID := aws.ToString(v.AssociationId)

				subnetCidrBlockState, err := waitSubnetIPv6CIDRBlockAssociationCreated(ctx, conn, associationID)

				if err != nil {
					return sdkdiag.AppendErrorf(diags, "waiting for EC2 Default Subnet (%s) IPv6 CIDR block (%s) to become associated: %s", d.Id(), associationID, err)
				}

				subnet.Ipv6CidrBlockAssociationSet[i].Ipv6CidrBlockState = subnetCidrBlockState
			}
		}

		if ipv6Native {
			computedIPv6CIDRBlock = true
		}
	} else {
		return sdkdiag.AppendErrorf(diags, "reading EC2 Default Subnet (%s): %s", availabilityZone, err)
	}

	if err := modifySubnetAttributesOnCreate(ctx, conn, d, subnet, computedIPv6CIDRBlock); err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	// Configure tags.
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig
	newTags := keyValueTags(ctx, getTagsIn(ctx))
	oldTags := keyValueTags(ctx, subnet.Tags).IgnoreSystem(names.EC2).IgnoreConfig(ignoreTagsConfig)

	if !oldTags.Equal(newTags) {
		if err := updateTags(ctx, conn, d.Id(), oldTags, newTags); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating EC2 Default Subnet (%s) tags: %s", d.Id(), err)
		}
	}

	return append(diags, resourceSubnetRead(ctx, d, meta)...)
}

func resourceDefaultSubnetDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	if d.Get(names.AttrForceDestroy).(bool) {
		return append(diags, resourceSubnetDelete(ctx, d, meta)...)
	}

	log.Printf("[WARN] EC2 Default Subnet (%s) not deleted, removing from state", d.Id())

	return diags
}
