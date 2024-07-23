// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	vpcCIDRMaxIPv4Netmask = 28
	vpcCIDRMinIPv4Netmask = 16
	vpcCIDRMaxIPv6Netmask = 56
)

// @SDKResource("aws_vpc", name="VPC")
// @Tags(identifierAttribute="id")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/ec2/types;awstypes;awstypes.Vpc")
func resourceVPC() *schema.Resource {
	//lintignore:R011
	return &schema.Resource{
		CreateWithoutTimeout: resourceVPCCreate,
		ReadWithoutTimeout:   resourceVPCRead,
		UpdateWithoutTimeout: resourceVPCUpdate,
		DeleteWithoutTimeout: resourceVPCDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceVPCImport,
		},

		CustomizeDiff: customdiff.All(
			resourceVPCCustomizeDiff,
			verify.SetTagsDiff,
		),

		SchemaVersion: 1,
		MigrateState:  vpcMigrateState,

		// Keep in sync with aws_default_vpc's schema.
		// See notes in default_vpc.go.
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
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ValidateFunc:  validation.IsCIDRNetwork(vpcCIDRMinIPv4Netmask, vpcCIDRMaxIPv4Netmask),
				ConflictsWith: []string{"ipv4_netmask_length"},
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
				Computed: true,
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
			"instance_tenancy": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      types.TenancyDefault,
				ValidateFunc: validation.StringInSlice(enum.Slice(types.TenancyDefault, types.TenancyDedicated), false),
			},
			"ipv4_ipam_pool_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"ipv4_netmask_length": {
				Type:          schema.TypeInt,
				Optional:      true,
				ForceNew:      true,
				ValidateFunc:  validation.IntBetween(vpcCIDRMinIPv4Netmask, vpcCIDRMaxIPv4Netmask),
				ConflictsWith: []string{names.AttrCIDRBlock},
				RequiredWith:  []string{"ipv4_ipam_pool_id"},
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

func resourceVPCCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	input := &ec2.CreateVpcInput{
		AmazonProvidedIpv6CidrBlock: aws.Bool(d.Get("assign_generated_ipv6_cidr_block").(bool)),
		InstanceTenancy:             types.Tenancy(d.Get("instance_tenancy").(string)),
		TagSpecifications:           getTagSpecificationsIn(ctx, types.ResourceTypeVpc),
	}

	if v, ok := d.GetOk(names.AttrCIDRBlock); ok {
		input.CidrBlock = aws.String(v.(string))
	}

	if v, ok := d.GetOk("ipv4_ipam_pool_id"); ok {
		input.Ipv4IpamPoolId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("ipv4_netmask_length"); ok {
		input.Ipv4NetmaskLength = aws.Int32(int32(v.(int)))
	}

	if v, ok := d.GetOk("ipv6_cidr_block"); ok {
		input.Ipv6CidrBlock = aws.String(v.(string))
	}

	if v, ok := d.GetOk("ipv6_cidr_block_network_border_group"); ok {
		input.Ipv6CidrBlockNetworkBorderGroup = aws.String(v.(string))
	}

	if v, ok := d.GetOk("ipv6_ipam_pool_id"); ok {
		input.Ipv6IpamPoolId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("ipv6_netmask_length"); ok {
		input.Ipv6NetmaskLength = aws.Int32(int32(v.(int)))
	}

	// "UnsupportedOperation: The operation AllocateIpamPoolCidr is not supported. Account 123456789012 is not monitored by IPAM ipam-07b079e3392782a55."
	outputRaw, err := tfresource.RetryWhenAWSErrMessageContains(ctx, ec2PropagationTimeout, func() (interface{}, error) {
		return conn.CreateVpc(ctx, input)
	}, errCodeUnsupportedOperation, "is not monitored by IPAM")

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EC2 VPC: %s", err)
	}

	output := outputRaw.(*ec2.CreateVpcOutput)

	d.SetId(aws.ToString(output.Vpc.VpcId))

	vpc, err := waitVPCCreated(ctx, conn, d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EC2 VPC (%s) create: %s", d.Id(), err)
	}

	if len(vpc.Ipv6CidrBlockAssociationSet) > 0 {
		associationID := aws.ToString(output.Vpc.Ipv6CidrBlockAssociationSet[0].AssociationId)

		if _, err := waitVPCIPv6CIDRBlockAssociationCreated(ctx, conn, associationID, vpcIPv6CIDRBlockAssociationCreatedTimeout); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for EC2 VPC (%s) IPv6 CIDR block (%s) associated: %s", d.Id(), associationID, err)
		}
	}

	vpcInfo := vpcInfo{
		vpc:                              vpc,
		enableDnsHostnames:               false,
		enableDnsSupport:                 true,
		enableNetworkAddressUsageMetrics: false,
	}

	if err := modifyVPCAttributesOnCreate(ctx, conn, d, &vpcInfo); err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EC2 VPC: %s", err)
	}

	return append(diags, resourceVPCRead(ctx, d, meta)...)
}

func resourceVPCRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	outputRaw, err := tfresource.RetryWhenNewResourceNotFound(ctx, ec2PropagationTimeout, func() (interface{}, error) {
		return findVPCByID(ctx, conn, d.Id())
	}, d.IsNewResource())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 VPC %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 VPC (%s): %s", d.Id(), err)
	}

	vpc := outputRaw.(*types.Vpc)

	ownerID := aws.ToString(vpc.OwnerId)
	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   names.EC2,
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: ownerID,
		Resource:  fmt.Sprintf("vpc/%s", d.Id()),
	}.String()
	d.Set(names.AttrARN, arn)
	d.Set(names.AttrCIDRBlock, vpc.CidrBlock)
	d.Set("dhcp_options_id", vpc.DhcpOptionsId)
	d.Set("instance_tenancy", vpc.InstanceTenancy)
	d.Set(names.AttrOwnerID, ownerID)

	if v, err := tfresource.RetryWhenNewResourceNotFound(ctx, ec2PropagationTimeout, func() (interface{}, error) {
		return findVPCAttribute(ctx, conn, d.Id(), types.VpcAttributeNameEnableDnsHostnames)
	}, d.IsNewResource()); err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 VPC (%s) Attribute (%s): %s", d.Id(), types.VpcAttributeNameEnableDnsHostnames, err)
	} else {
		d.Set("enable_dns_hostnames", v)
	}

	if v, err := tfresource.RetryWhenNewResourceNotFound(ctx, ec2PropagationTimeout, func() (interface{}, error) {
		return findVPCAttribute(ctx, conn, d.Id(), types.VpcAttributeNameEnableDnsSupport)
	}, d.IsNewResource()); err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 VPC (%s) Attribute (%s): %s", d.Id(), types.VpcAttributeNameEnableDnsSupport, err)
	} else {
		d.Set("enable_dns_support", v)
	}

	if v, err := tfresource.RetryWhenNewResourceNotFound(ctx, ec2PropagationTimeout, func() (interface{}, error) {
		return findVPCAttribute(ctx, conn, d.Id(), types.VpcAttributeNameEnableNetworkAddressUsageMetrics)
	}, d.IsNewResource()); err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 VPC (%s) Attribute (%s): %s", d.Id(), types.VpcAttributeNameEnableNetworkAddressUsageMetrics, err)
	} else {
		d.Set("enable_network_address_usage_metrics", v)
	}

	if v, err := findVPCDefaultNetworkACL(ctx, conn, d.Id()); err != nil {
		log.Printf("[WARN] Error reading EC2 VPC (%s) default NACL: %s", d.Id(), err)
	} else {
		d.Set("default_network_acl_id", v.NetworkAclId)
	}

	if v, err := findVPCMainRouteTable(ctx, conn, d.Id()); err != nil {
		log.Printf("[WARN] Error reading EC2 VPC (%s) main Route Table: %s", d.Id(), err)
		d.Set("default_route_table_id", nil)
		d.Set("main_route_table_id", nil)
	} else {
		d.Set("default_route_table_id", v.RouteTableId)
		d.Set("main_route_table_id", v.RouteTableId)
	}

	if v, err := findVPCDefaultSecurityGroup(ctx, conn, d.Id()); err != nil {
		log.Printf("[WARN] Error reading EC2 VPC (%s) default Security Group: %s", d.Id(), err)
		d.Set("default_security_group_id", nil)
	} else {
		d.Set("default_security_group_id", v.GroupId)
	}

	if ipv6CIDRBlockAssociation := defaultIPv6CIDRBlockAssociation(vpc, d.Get("ipv6_association_id").(string)); ipv6CIDRBlockAssociation == nil {
		d.Set("assign_generated_ipv6_cidr_block", nil)
		d.Set("ipv6_association_id", nil)
		d.Set("ipv6_cidr_block", nil)
		d.Set("ipv6_cidr_block_network_border_group", nil)
		d.Set("ipv6_ipam_pool_id", nil)
		d.Set("ipv6_netmask_length", nil)
	} else {
		cidrBlock := aws.ToString(ipv6CIDRBlockAssociation.Ipv6CidrBlock)
		ipv6PoolID := aws.ToString(ipv6CIDRBlockAssociation.Ipv6Pool)
		isAmazonIPv6Pool := ipv6PoolID == amazonIPv6PoolID
		d.Set("assign_generated_ipv6_cidr_block", isAmazonIPv6Pool)
		d.Set("ipv6_association_id", ipv6CIDRBlockAssociation.AssociationId)
		d.Set("ipv6_cidr_block", cidrBlock)
		d.Set("ipv6_cidr_block_network_border_group", ipv6CIDRBlockAssociation.NetworkBorderGroup)
		if isAmazonIPv6Pool {
			d.Set("ipv6_ipam_pool_id", nil)
		} else {
			if ipv6PoolID == ipamManagedIPv6PoolID {
				d.Set("ipv6_ipam_pool_id", d.Get("ipv6_ipam_pool_id"))
			} else {
				d.Set("ipv6_ipam_pool_id", ipv6PoolID)
			}
		}
		d.Set("ipv6_netmask_length", nil)
		if ipv6PoolID != "" && !isAmazonIPv6Pool {
			parts := strings.Split(cidrBlock, "/")
			if len(parts) == 2 {
				if v, err := strconv.Atoi(parts[1]); err == nil {
					d.Set("ipv6_netmask_length", v)
				} else {
					log.Printf("[WARN] Unable to parse CIDR (%s) netmask length: %s", cidrBlock, err)
				}
			} else {
				log.Printf("[WARN] Invalid CIDR block format: %s", cidrBlock)
			}
		}
	}

	setTagsOut(ctx, vpc.Tags)

	return diags
}

func resourceVPCUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	if d.HasChange("enable_dns_hostnames") {
		if err := modifyVPCDNSHostnames(ctx, conn, d.Id(), d.Get("enable_dns_hostnames").(bool)); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating EC2 VPC (%s): %s", d.Id(), err)
		}
	}

	if d.HasChange("enable_dns_support") {
		if err := modifyVPCDNSSupport(ctx, conn, d.Id(), d.Get("enable_dns_support").(bool)); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating EC2 VPC (%s): %s", d.Id(), err)
		}
	}

	if d.HasChange("enable_network_address_usage_metrics") {
		if err := modifyVPCNetworkAddressUsageMetrics(ctx, conn, d.Id(), d.Get("enable_network_address_usage_metrics").(bool)); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating EC2 VPC (%s): %s", d.Id(), err)
		}
	}

	if d.HasChange("instance_tenancy") {
		if err := modifyVPCTenancy(ctx, conn, d.Id(), d.Get("instance_tenancy").(string)); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating EC2 VPC (%s): %s", d.Id(), err)
		}
	}

	if d.HasChanges("assign_generated_ipv6_cidr_block", "ipv6_cidr_block_network_border_group") {
		associationID, err := modifyVPCIPv6CIDRBlockAssociation(ctx, conn, d.Id(),
			d.Get("ipv6_association_id").(string),
			d.Get("assign_generated_ipv6_cidr_block").(bool),
			"",
			"",
			0,
			d.Get("ipv6_cidr_block_network_border_group").(string))

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating EC2 VPC (%s): %s", d.Id(), err)
		}

		d.Set("ipv6_association_id", associationID)
	}

	if d.HasChanges("ipv6_cidr_block", "ipv6_ipam_pool_id") {
		associationID, err := modifyVPCIPv6CIDRBlockAssociation(ctx, conn, d.Id(),
			d.Get("ipv6_association_id").(string),
			false,
			d.Get("ipv6_cidr_block").(string),
			d.Get("ipv6_ipam_pool_id").(string),
			d.Get("ipv6_netmask_length").(int),
			"")

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating EC2 VPC (%s): %s", d.Id(), err)
		}

		d.Set("ipv6_association_id", associationID)
	}

	return append(diags, resourceVPCRead(ctx, d, meta)...)
}

func resourceVPCDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	input := &ec2.DeleteVpcInput{
		VpcId: aws.String(d.Id()),
	}

	log.Printf("[INFO] Deleting EC2 VPC: %s", d.Id())
	_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, d.Timeout(schema.TimeoutDelete), func() (interface{}, error) {
		return conn.DeleteVpc(ctx, input)
	}, errCodeDependencyViolation)

	if tfawserr.ErrCodeEquals(err, errCodeInvalidVPCIDNotFound) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting EC2 VPC (%s): %s", d.Id(), err)
	}

	_, err = tfresource.RetryUntilNotFound(ctx, d.Timeout(schema.TimeoutDelete), func() (interface{}, error) {
		return findVPCByID(ctx, conn, d.Id())
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EC2 VPC (%s) delete: %s", d.Id(), err)
	}

	// If the VPC's CIDR block was allocated from an IPAM pool, wait for the allocation to disappear.
	var ipamPoolID string
	if v, ok := d.GetOk("ipv4_ipam_pool_id"); ok {
		ipamPoolID = v.(string)
	}
	if ipamPoolID == "" {
		if v, ok := d.GetOk("ipv6_ipam_pool_id"); ok {
			ipamPoolID = v.(string)
		}
	}
	if ipamPoolID != "" && ipamPoolID != amazonIPv6PoolID {
		const (
			timeout = 35 * time.Minute // IPAM eventual consistency. It can take ~30 min to release allocations.
		)
		_, err := tfresource.RetryUntilNotFound(ctx, timeout, func() (interface{}, error) {
			return findIPAMPoolAllocationsForVPC(ctx, conn, ipamPoolID, d.Id())
		})

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for EC2 VPC (%s) IPAM Pool (%s) Allocation delete: %s", d.Id(), ipamPoolID, err)
		}
	}

	return diags
}

func resourceVPCImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	d.Set("assign_generated_ipv6_cidr_block", false)
	return []*schema.ResourceData{d}, nil
}

func resourceVPCCustomizeDiff(_ context.Context, diff *schema.ResourceDiff, v interface{}) error {
	if diff.HasChange("assign_generated_ipv6_cidr_block") {
		if err := diff.SetNewComputed("ipv6_association_id"); err != nil {
			return fmt.Errorf("setting ipv6_association_id to computed: %s", err)
		}
		if err := diff.SetNewComputed("ipv6_cidr_block"); err != nil {
			return fmt.Errorf("setting ipv6_cidr_block to computed: %s", err)
		}
	}

	if diff.HasChange("instance_tenancy") {
		old, new := diff.GetChange("instance_tenancy")
		if old.(string) != string(types.TenancyDedicated) || new.(string) != string(types.TenancyDefault) {
			diff.ForceNew("instance_tenancy")
		}
	}

	// cidr_block can be set by a value returned from IPAM or explicitly in config.
	if diff.Id() != "" && diff.HasChange(names.AttrCIDRBlock) {
		// If netmask is set then cidr_block is derived from IPAM, ignore changes.
		if diff.Get("ipv4_netmask_length") != 0 {
			return diff.Clear(names.AttrCIDRBlock)
		}
		return diff.ForceNew(names.AttrCIDRBlock)
	}

	return nil
}

// defaultIPv6CIDRBlockAssociation returns the "default" IPv6 CIDR block.
// Try and find IPv6 CIDR block information, first by any stored association ID.
// Then if no IPv6 CIDR block information is available, use the first associated IPv6 CIDR block.
func defaultIPv6CIDRBlockAssociation(vpc *types.Vpc, associationID string) *types.VpcIpv6CidrBlockAssociation {
	var ipv6CIDRBlockAssociation types.VpcIpv6CidrBlockAssociation

	if associationID != "" {
		for _, v := range vpc.Ipv6CidrBlockAssociationSet {
			if state := string(v.Ipv6CidrBlockState.State); state == string(types.VpcCidrBlockStateCodeAssociated) && aws.ToString(v.AssociationId) == associationID {
				ipv6CIDRBlockAssociation = v
				break
			}
		}
	}

	if ipv6CIDRBlockAssociation == (types.VpcIpv6CidrBlockAssociation{}) {
		for _, v := range vpc.Ipv6CidrBlockAssociationSet {
			if string(v.Ipv6CidrBlockState.State) == string(types.VpcCidrBlockStateCodeAssociated) {
				ipv6CIDRBlockAssociation = v
			}
		}
	}

	return &ipv6CIDRBlockAssociation
}

type vpcInfo struct {
	vpc                              *types.Vpc
	enableDnsHostnames               bool
	enableDnsSupport                 bool
	enableNetworkAddressUsageMetrics bool
}

// modifyVPCAttributesOnCreate sets VPC attributes on resource Create.
// Called after new VPC creation or existing default VPC adoption.
func modifyVPCAttributesOnCreate(ctx context.Context, conn *ec2.Client, d *schema.ResourceData, vpcInfo *vpcInfo) error {
	if new, old := d.Get("enable_dns_hostnames").(bool), vpcInfo.enableDnsHostnames; old != new {
		if err := modifyVPCDNSHostnames(ctx, conn, d.Id(), new); err != nil {
			return err
		}
	}

	if new, old := d.Get("enable_dns_support").(bool), vpcInfo.enableDnsSupport; old != new {
		if err := modifyVPCDNSSupport(ctx, conn, d.Id(), new); err != nil {
			return err
		}
	}

	if new, old := d.Get("enable_network_address_usage_metrics").(bool), vpcInfo.enableNetworkAddressUsageMetrics; old != new {
		if err := modifyVPCNetworkAddressUsageMetrics(ctx, conn, d.Id(), new); err != nil {
			return err
		}
	}

	return nil
}

func modifyVPCDNSHostnames(ctx context.Context, conn *ec2.Client, vpcID string, v bool) error {
	input := &ec2.ModifyVpcAttributeInput{
		EnableDnsHostnames: &types.AttributeBooleanValue{
			Value: aws.Bool(v),
		},
		VpcId: aws.String(vpcID),
	}

	if _, err := conn.ModifyVpcAttribute(ctx, input); err != nil {
		return fmt.Errorf("modifying EnableDnsHostnames: %w", err)
	}

	if _, err := waitVPCAttributeUpdated(ctx, conn, vpcID, types.VpcAttributeNameEnableDnsHostnames, v); err != nil {
		return fmt.Errorf("modifying EnableDnsHostnames: waiting for completion: %w", err)
	}

	return nil
}

func modifyVPCDNSSupport(ctx context.Context, conn *ec2.Client, vpcID string, v bool) error {
	input := &ec2.ModifyVpcAttributeInput{
		EnableDnsSupport: &types.AttributeBooleanValue{
			Value: aws.Bool(v),
		},
		VpcId: aws.String(vpcID),
	}

	if _, err := conn.ModifyVpcAttribute(ctx, input); err != nil {
		return fmt.Errorf("modifying EnableDnsSupport: %w", err)
	}

	if _, err := waitVPCAttributeUpdated(ctx, conn, vpcID, types.VpcAttributeNameEnableDnsSupport, v); err != nil {
		return fmt.Errorf("modifying EnableDnsSupport: waiting for completion: %w", err)
	}

	return nil
}

func modifyVPCNetworkAddressUsageMetrics(ctx context.Context, conn *ec2.Client, vpcID string, v bool) error {
	input := &ec2.ModifyVpcAttributeInput{
		EnableNetworkAddressUsageMetrics: &types.AttributeBooleanValue{
			Value: aws.Bool(v),
		},
		VpcId: aws.String(vpcID),
	}

	if _, err := conn.ModifyVpcAttribute(ctx, input); err != nil {
		return fmt.Errorf("modifying EnableNetworkAddressUsageMetrics: %w", err)
	}

	if _, err := waitVPCAttributeUpdated(ctx, conn, vpcID, types.VpcAttributeNameEnableNetworkAddressUsageMetrics, v); err != nil {
		return fmt.Errorf("modifying EnableNetworkAddressUsageMetrics: waiting for completion: %w", err)
	}

	return nil
}

// modifyVPCIPv6CIDRBlockAssociation modify's a VPC's IPv6 CIDR block association.
// Any exiting association is deleted and any new association's ID is returned.
func modifyVPCIPv6CIDRBlockAssociation(ctx context.Context, conn *ec2.Client, vpcID, associationID string, amazonProvidedCIDRBlock bool, cidrBlock, ipamPoolID string, netmaskLength int, networkBorderGroup string) (string, error) {
	if associationID != "" {
		input := &ec2.DisassociateVpcCidrBlockInput{
			AssociationId: aws.String(associationID),
		}

		_, err := conn.DisassociateVpcCidrBlock(ctx, input)

		if err != nil {
			return "", fmt.Errorf("disassociating IPv6 CIDR block (%s): %w", associationID, err)
		}

		if err := waitVPCIPv6CIDRBlockAssociationDeleted(ctx, conn, associationID, vpcIPv6CIDRBlockAssociationDeletedTimeout); err != nil {
			return "", fmt.Errorf("disassociating IPv6 CIDR block (%s): waiting for completion: %w", associationID, err)
		}
	}

	if amazonProvidedCIDRBlock || cidrBlock != "" || ipamPoolID != "" {
		input := &ec2.AssociateVpcCidrBlockInput{
			VpcId: aws.String(vpcID),
		}

		if amazonProvidedCIDRBlock {
			input.AmazonProvidedIpv6CidrBlock = aws.Bool(amazonProvidedCIDRBlock)
		}

		if cidrBlock != "" {
			input.Ipv6CidrBlock = aws.String(cidrBlock)
		}

		if networkBorderGroup != "" {
			input.Ipv6CidrBlockNetworkBorderGroup = aws.String(networkBorderGroup)
		}

		if ipamPoolID != "" {
			input.Ipv6IpamPoolId = aws.String(ipamPoolID)
		}

		if netmaskLength > 0 {
			input.Ipv6NetmaskLength = aws.Int32(int32(netmaskLength))
		}

		output, err := conn.AssociateVpcCidrBlock(ctx, input)

		if err != nil {
			return "", fmt.Errorf("associating IPv6 CIDR block: %w", err)
		}

		associationID = aws.ToString(output.Ipv6CidrBlockAssociation.AssociationId)

		if _, err := waitVPCIPv6CIDRBlockAssociationCreated(ctx, conn, associationID, vpcIPv6CIDRBlockAssociationCreatedTimeout); err != nil {
			return "", fmt.Errorf("associating IPv6 CIDR block: waiting for completion: %w", err)
		}
	}

	return associationID, nil
}

func modifyVPCTenancy(ctx context.Context, conn *ec2.Client, vpcID string, v string) error {
	input := &ec2.ModifyVpcTenancyInput{
		InstanceTenancy: types.VpcTenancy(v),
		VpcId:           aws.String(vpcID),
	}

	if _, err := conn.ModifyVpcTenancy(ctx, input); err != nil {
		return fmt.Errorf("modifying Tenancy: %w", err)
	}

	return nil
}

func findIPAMPoolAllocationsForVPC(ctx context.Context, conn *ec2.Client, poolID, vpcID string) ([]types.IpamPoolAllocation, error) {
	input := &ec2.GetIpamPoolAllocationsInput{
		IpamPoolId: aws.String(poolID),
	}

	output, err := findIPAMPoolAllocations(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	output = slices.Filter(output, func(v types.IpamPoolAllocation) bool {
		return string(v.ResourceType) == string(types.IpamPoolAllocationResourceTypeVpc) && aws.ToString(v.ResourceId) == vpcID
	})

	if len(output) == 0 {
		return nil, &retry.NotFoundError{}
	}

	return output, nil
}
