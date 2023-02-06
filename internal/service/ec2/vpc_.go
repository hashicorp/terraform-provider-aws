package ec2

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

const (
	VPCCIDRMaxIPv4 = 28
	VPCCIDRMinIPv4 = 16
	VPCCIDRMaxIPv6 = 56
)

func ResourceVPC() *schema.Resource {
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
		MigrateState:  VPCMigrateState,

		// Keep in sync with aws_default_vpc's schema.
		// See notes in default_vpc.go.
		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"assign_generated_ipv6_cidr_block": {
				Type:          schema.TypeBool,
				Optional:      true,
				ConflictsWith: []string{"ipv6_ipam_pool_id"},
			},
			"cidr_block": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ValidateFunc:  validation.IsCIDRNetwork(VPCCIDRMinIPv4, VPCCIDRMaxIPv4),
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
			"enable_classiclink": {
				Type:       schema.TypeBool,
				Optional:   true,
				Computed:   true,
				Deprecated: `With the retirement of EC2-Classic the enable_classiclink attribute has been deprecated and will be removed in a future version.`,
			},
			"enable_classiclink_dns_support": {
				Type:       schema.TypeBool,
				Optional:   true,
				Computed:   true,
				Deprecated: `With the retirement of EC2-Classic the enable_classiclink_dns_support attribute has been deprecated and will be removed in a future version.`,
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
				Default:      ec2.TenancyDefault,
				ValidateFunc: validation.StringInSlice([]string{ec2.TenancyDefault, ec2.TenancyDedicated}, false),
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
				ValidateFunc:  validation.IntBetween(VPCCIDRMinIPv4, VPCCIDRMaxIPv4),
				ConflictsWith: []string{"cidr_block"},
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
					validation.IsCIDRNetwork(VPCCIDRMaxIPv6, VPCCIDRMaxIPv6)),
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
				ValidateFunc:  validation.IntInSlice([]int{VPCCIDRMaxIPv6}),
				ConflictsWith: []string{"ipv6_cidr_block"},
				RequiredWith:  []string{"ipv6_ipam_pool_id"},
			},
			"main_route_table_id": {
				Type:     schema.TypeString,
				Computed: true,
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

func resourceVPCCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	if _, ok := d.GetOk("enable_classiclink"); ok {
		return sdkdiag.AppendErrorf(diags, `with the retirement of EC2-Classic no new VPCs can be created with ClassicLink enabled`)
	}

	input := &ec2.CreateVpcInput{
		AmazonProvidedIpv6CidrBlock: aws.Bool(d.Get("assign_generated_ipv6_cidr_block").(bool)),
		InstanceTenancy:             aws.String(d.Get("instance_tenancy").(string)),
		TagSpecifications:           tagSpecificationsFromKeyValueTags(tags, ec2.ResourceTypeVpc),
	}

	if v, ok := d.GetOk("cidr_block"); ok {
		input.CidrBlock = aws.String(v.(string))
	}

	if v, ok := d.GetOk("ipv4_ipam_pool_id"); ok {
		input.Ipv4IpamPoolId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("ipv4_netmask_length"); ok {
		input.Ipv4NetmaskLength = aws.Int64(int64(v.(int)))
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
		input.Ipv6NetmaskLength = aws.Int64(int64(v.(int)))
	}

	log.Printf("[DEBUG] Creating EC2 VPC: %s", input)
	output, err := conn.CreateVpcWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EC2 VPC: %s", err)
	}

	d.SetId(aws.StringValue(output.Vpc.VpcId))

	vpc, err := WaitVPCCreated(ctx, conn, d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EC2 VPC (%s) create: %s", d.Id(), err)
	}

	if len(vpc.Ipv6CidrBlockAssociationSet) > 0 && vpc.Ipv6CidrBlockAssociationSet[0] != nil {
		associationID := aws.StringValue(output.Vpc.Ipv6CidrBlockAssociationSet[0].AssociationId)

		_, err = WaitVPCIPv6CIDRBlockAssociationCreated(ctx, conn, associationID, vpcIPv6CIDRBlockAssociationCreatedTimeout)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for EC2 VPC (%s) IPv6 CIDR block (%s) to become associated: %s", d.Id(), associationID, err)
		}
	}

	vpcInfo := vpcInfo{
		vpc:                              vpc,
		enableClassicLink:                false,
		enableClassicLinkDNSSupport:      false,
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
	conn := meta.(*conns.AWSClient).EC2Conn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	outputRaw, err := tfresource.RetryWhenNewResourceNotFound(ctx, propagationTimeout, func() (interface{}, error) {
		return FindVPCByID(ctx, conn, d.Id())
	}, d.IsNewResource())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 VPC %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 VPC (%s): %s", d.Id(), err)
	}

	vpc := outputRaw.(*ec2.Vpc)

	ownerID := aws.StringValue(vpc.OwnerId)
	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   ec2.ServiceName,
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: ownerID,
		Resource:  fmt.Sprintf("vpc/%s", d.Id()),
	}.String()
	d.Set("arn", arn)
	d.Set("cidr_block", vpc.CidrBlock)
	d.Set("dhcp_options_id", vpc.DhcpOptionsId)
	d.Set("instance_tenancy", vpc.InstanceTenancy)
	d.Set("owner_id", ownerID)

	if v, err := FindVPCClassicLinkEnabled(ctx, conn, d.Id()); err != nil {
		if tfresource.NotFound(err) {
			d.Set("enable_classiclink", nil)
		} else {
			return sdkdiag.AppendErrorf(diags, "reading EC2 VPC (%s) ClassicLinkEnabled: %s", d.Id(), err)
		}
	} else {
		d.Set("enable_classiclink", v)
	}

	if v, err := FindVPCClassicLinkDNSSupported(ctx, conn, d.Id()); err != nil {
		if tfresource.NotFound(err) {
			d.Set("enable_classiclink_dns_support", nil)
		} else {
			return sdkdiag.AppendErrorf(diags, "reading EC2 VPC (%s) ClassicLinkDnsSupported: %s", d.Id(), err)
		}
	} else {
		d.Set("enable_classiclink_dns_support", v)
	}

	if v, err := FindVPCAttribute(ctx, conn, d.Id(), ec2.VpcAttributeNameEnableDnsHostnames); err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 VPC (%s) Attribute (%s): %s", d.Id(), ec2.VpcAttributeNameEnableDnsHostnames, err)
	} else {
		d.Set("enable_dns_hostnames", v)
	}

	if v, err := FindVPCAttribute(ctx, conn, d.Id(), ec2.VpcAttributeNameEnableDnsSupport); err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 VPC (%s) Attribute (%s): %s", d.Id(), ec2.VpcAttributeNameEnableDnsSupport, err)
	} else {
		d.Set("enable_dns_support", v)
	}

	if v, err := FindVPCAttribute(ctx, conn, d.Id(), ec2.VpcAttributeNameEnableNetworkAddressUsageMetrics); err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 VPC (%s) Attribute (%s): %s", d.Id(), ec2.VpcAttributeNameEnableNetworkAddressUsageMetrics, err)
	} else {
		d.Set("enable_network_address_usage_metrics", v)
	}

	if v, err := FindVPCDefaultNetworkACL(ctx, conn, d.Id()); err != nil {
		log.Printf("[WARN] Error reading EC2 VPC (%s) default NACL: %s", d.Id(), err)
	} else {
		d.Set("default_network_acl_id", v.NetworkAclId)
	}

	if v, err := FindVPCMainRouteTable(ctx, conn, d.Id()); err != nil {
		log.Printf("[WARN] Error reading EC2 VPC (%s) main Route Table: %s", d.Id(), err)
		d.Set("default_route_table_id", nil)
		d.Set("main_route_table_id", nil)
	} else {
		d.Set("default_route_table_id", v.RouteTableId)
		d.Set("main_route_table_id", v.RouteTableId)
	}

	if v, err := FindVPCDefaultSecurityGroup(ctx, conn, d.Id()); err != nil {
		log.Printf("[WARN] Error reading EC2 VPC (%s) default Security Group: %s", d.Id(), err)
		d.Set("default_security_group_id", nil)
	} else {
		d.Set("default_security_group_id", v.GroupId)
	}

	d.Set("assign_generated_ipv6_cidr_block", nil)
	d.Set("ipv6_cidr_block", nil)
	d.Set("ipv6_cidr_block_network_border_group", nil)
	d.Set("ipv6_ipam_pool_id", nil)
	d.Set("ipv6_netmask_length", nil)

	ipv6CIDRBlockAssociation := defaultIPv6CIDRBlockAssociation(vpc, d.Get("ipv6_association_id").(string))

	if ipv6CIDRBlockAssociation == nil {
		d.Set("ipv6_association_id", nil)
	} else {
		cidrBlock := aws.StringValue(ipv6CIDRBlockAssociation.Ipv6CidrBlock)
		ipv6PoolID := aws.StringValue(ipv6CIDRBlockAssociation.Ipv6Pool)
		isAmazonIPv6Pool := ipv6PoolID == AmazonIPv6PoolID
		d.Set("assign_generated_ipv6_cidr_block", isAmazonIPv6Pool)
		d.Set("ipv6_association_id", ipv6CIDRBlockAssociation.AssociationId)
		d.Set("ipv6_cidr_block", cidrBlock)
		d.Set("ipv6_cidr_block_network_border_group", ipv6CIDRBlockAssociation.NetworkBorderGroup)
		if !isAmazonIPv6Pool {
			d.Set("ipv6_ipam_pool_id", ipv6PoolID)
		}
		if ipv6PoolID != "" && !isAmazonIPv6Pool {
			parts := strings.Split(cidrBlock, "/")
			if len(parts) == 2 {
				if v, err := strconv.Atoi(parts[1]); err != nil {
					d.Set("ipv6_netmask_length", v)
				} else {
					log.Printf("[WARN] Unable to parse CIDR (%s) netmask length: %s", cidrBlock, err)
				}
			} else {
				log.Printf("[WARN] Invalid CIDR block format: %s", cidrBlock)
			}
		}
	}

	tags := KeyValueTags(vpc.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags_all: %s", err)
	}

	return diags
}

func resourceVPCUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn()

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

	if d.HasChange("enable_classiclink") {
		if err := modifyVPCClassicLink(ctx, conn, d.Id(), d.Get("enable_classiclink").(bool)); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating EC2 VPC (%s): %s", d.Id(), err)
		}
	}

	if d.HasChange("enable_classiclink_dns_support") {
		if err := modifyVPCClassicLinkDNSSupport(ctx, conn, d.Id(), d.Get("enable_classiclink_dns_support").(bool)); err != nil {
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

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(ctx, conn, d.Id(), o, n); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating EC2 VPC (%s): updating tags: %s", d.Id(), err)
		}
	}

	return append(diags, resourceVPCRead(ctx, d, meta)...)
}

func resourceVPCDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn()

	input := &ec2.DeleteVpcInput{
		VpcId: aws.String(d.Id()),
	}

	log.Printf("[INFO] Deleting EC2 VPC: %s", d.Id())
	_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, vpcDeletedTimeout, func() (interface{}, error) {
		return conn.DeleteVpcWithContext(ctx, input)
	}, errCodeDependencyViolation)

	if tfawserr.ErrCodeEquals(err, errCodeInvalidVPCIDNotFound) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting EC2 VPC (%s): %s", d.Id(), err)
	}

	_, err = tfresource.RetryUntilNotFound(ctx, vpcDeletedTimeout, func() (interface{}, error) {
		return FindVPCByID(ctx, conn, d.Id())
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EC2 VPC (%s) delete: %s", d.Id(), err)
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
			return fmt.Errorf("error setting ipv6_association_id to computed: %s", err)
		}
		if err := diff.SetNewComputed("ipv6_cidr_block"); err != nil {
			return fmt.Errorf("error setting ipv6_cidr_block to computed: %s", err)
		}
	}

	if diff.HasChange("instance_tenancy") {
		old, new := diff.GetChange("instance_tenancy")
		if old.(string) != ec2.TenancyDedicated || new.(string) != ec2.TenancyDefault {
			diff.ForceNew("instance_tenancy")
		}
	}

	// cidr_block can be set by a value returned from IPAM or explicitly in config.
	if diff.Id() != "" && diff.HasChange("cidr_block") {
		// If netmask is set then cidr_block is derived from IPAM, ignore changes.
		if diff.Get("ipv4_netmask_length") != 0 {
			return diff.Clear("cidr_block")
		}
		return diff.ForceNew("cidr_block")
	}

	return nil
}

// defaultIPv6CIDRBlockAssociation returns the "default" IPv6 CIDR block.
// Try and find IPv6 CIDR block information, first by any stored association ID.
// Then if no IPv6 CIDR block information is available, use the first associated IPv6 CIDR block.
func defaultIPv6CIDRBlockAssociation(vpc *ec2.Vpc, associationID string) *ec2.VpcIpv6CidrBlockAssociation {
	var ipv6CIDRBlockAssociation *ec2.VpcIpv6CidrBlockAssociation

	if associationID != "" {
		for _, v := range vpc.Ipv6CidrBlockAssociationSet {
			if state := aws.StringValue(v.Ipv6CidrBlockState.State); state == ec2.VpcCidrBlockStateCodeAssociated && aws.StringValue(v.AssociationId) == associationID {
				ipv6CIDRBlockAssociation = v

				break
			}
		}
	}

	if ipv6CIDRBlockAssociation == nil {
		for _, v := range vpc.Ipv6CidrBlockAssociationSet {
			if aws.StringValue(v.Ipv6CidrBlockState.State) == ec2.VpcCidrBlockStateCodeAssociated {
				ipv6CIDRBlockAssociation = v
			}
		}
	}

	return ipv6CIDRBlockAssociation
}

type vpcInfo struct {
	vpc                              *ec2.Vpc
	enableClassicLink                bool
	enableClassicLinkDNSSupport      bool
	enableDnsHostnames               bool
	enableDnsSupport                 bool
	enableNetworkAddressUsageMetrics bool
}

// modifyVPCAttributesOnCreate sets VPC attributes on resource Create.
// Called after new VPC creation or existing default VPC adoption.
func modifyVPCAttributesOnCreate(ctx context.Context, conn *ec2.EC2, d *schema.ResourceData, vpcInfo *vpcInfo) error {
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

	if new, old := d.Get("enable_classiclink").(bool), vpcInfo.enableClassicLink; old != new {
		if err := modifyVPCClassicLink(ctx, conn, d.Id(), new); err != nil {
			return err
		}
	}

	if new, old := d.Get("enable_classiclink_dns_support").(bool), vpcInfo.enableClassicLinkDNSSupport; old != new {
		if err := modifyVPCClassicLinkDNSSupport(ctx, conn, d.Id(), new); err != nil {
			return err
		}
	}

	return nil
}

func modifyVPCClassicLink(ctx context.Context, conn *ec2.EC2, vpcID string, v bool) error {
	if v {
		input := &ec2.EnableVpcClassicLinkInput{
			VpcId: aws.String(vpcID),
		}

		if _, err := conn.EnableVpcClassicLinkWithContext(ctx, input); err != nil {
			return fmt.Errorf("enabling EnableDnsHostnames: %w", err)
		}
	} else {
		input := &ec2.DisableVpcClassicLinkInput{
			VpcId: aws.String(vpcID),
		}

		if _, err := conn.DisableVpcClassicLinkWithContext(ctx, input); err != nil {
			return fmt.Errorf("disabling EnableDnsHostnames: %w", err)
		}
	}

	return nil
}

func modifyVPCClassicLinkDNSSupport(ctx context.Context, conn *ec2.EC2, vpcID string, v bool) error {
	if v {
		input := &ec2.EnableVpcClassicLinkDnsSupportInput{
			VpcId: aws.String(vpcID),
		}

		if _, err := conn.EnableVpcClassicLinkDnsSupportWithContext(ctx, input); err != nil {
			return fmt.Errorf("enabling ClassicLinkDnsSupport: %w", err)
		}
	} else {
		input := &ec2.DisableVpcClassicLinkDnsSupportInput{
			VpcId: aws.String(vpcID),
		}

		if _, err := conn.DisableVpcClassicLinkDnsSupportWithContext(ctx, input); err != nil {
			return fmt.Errorf("disabling ClassicLinkDnsSupport: %w", err)
		}
	}

	return nil
}

func modifyVPCDNSHostnames(ctx context.Context, conn *ec2.EC2, vpcID string, v bool) error {
	input := &ec2.ModifyVpcAttributeInput{
		EnableDnsHostnames: &ec2.AttributeBooleanValue{
			Value: aws.Bool(v),
		},
		VpcId: aws.String(vpcID),
	}

	if _, err := conn.ModifyVpcAttributeWithContext(ctx, input); err != nil {
		return fmt.Errorf("modifying EnableDnsHostnames: %w", err)
	}

	if _, err := WaitVPCAttributeUpdated(ctx, conn, vpcID, ec2.VpcAttributeNameEnableDnsHostnames, v); err != nil {
		return fmt.Errorf("modifying EnableDnsHostnames: waiting for completion: %w", err)
	}

	return nil
}

func modifyVPCDNSSupport(ctx context.Context, conn *ec2.EC2, vpcID string, v bool) error {
	input := &ec2.ModifyVpcAttributeInput{
		EnableDnsSupport: &ec2.AttributeBooleanValue{
			Value: aws.Bool(v),
		},
		VpcId: aws.String(vpcID),
	}

	if _, err := conn.ModifyVpcAttributeWithContext(ctx, input); err != nil {
		return fmt.Errorf("modifying EnableDnsSupport: %w", err)
	}

	if _, err := WaitVPCAttributeUpdated(ctx, conn, vpcID, ec2.VpcAttributeNameEnableDnsSupport, v); err != nil {
		return fmt.Errorf("modifying EnableDnsSupport: waiting for completion: %w", err)
	}

	return nil
}

func modifyVPCNetworkAddressUsageMetrics(ctx context.Context, conn *ec2.EC2, vpcID string, v bool) error {
	input := &ec2.ModifyVpcAttributeInput{
		EnableNetworkAddressUsageMetrics: &ec2.AttributeBooleanValue{
			Value: aws.Bool(v),
		},
		VpcId: aws.String(vpcID),
	}

	if _, err := conn.ModifyVpcAttributeWithContext(ctx, input); err != nil {
		return fmt.Errorf("modifying EnableNetworkAddressUsageMetrics: %w", err)
	}

	if _, err := WaitVPCAttributeUpdated(ctx, conn, vpcID, ec2.VpcAttributeNameEnableNetworkAddressUsageMetrics, v); err != nil {
		return fmt.Errorf("modifying EnableNetworkAddressUsageMetrics: waiting for completion: %w", err)
	}

	return nil
}

// modifyVPCIPv6CIDRBlockAssociation modify's a VPC's IPv6 CIDR block association.
// Any exiting association is deleted and any new association's ID is returned.
func modifyVPCIPv6CIDRBlockAssociation(ctx context.Context, conn *ec2.EC2, vpcID, associationID string, amazonProvidedCIDRBlock bool, cidrBlock, ipamPoolID string, netmaskLength int, networkBorderGroup string) (string, error) {
	if associationID != "" {
		input := &ec2.DisassociateVpcCidrBlockInput{
			AssociationId: aws.String(associationID),
		}

		_, err := conn.DisassociateVpcCidrBlockWithContext(ctx, input)

		if err != nil {
			return "", fmt.Errorf("disassociating IPv6 CIDR block (%s): %w", associationID, err)
		}

		_, err = WaitVPCIPv6CIDRBlockAssociationDeleted(ctx, conn, associationID, vpcIPv6CIDRBlockAssociationDeletedTimeout)

		if err != nil {
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
			input.Ipv6NetmaskLength = aws.Int64(int64(netmaskLength))
		}

		output, err := conn.AssociateVpcCidrBlockWithContext(ctx, input)

		if err != nil {
			return "", fmt.Errorf("associating IPv6 CIDR block: %w", err)
		}

		associationID = aws.StringValue(output.Ipv6CidrBlockAssociation.AssociationId)

		_, err = WaitVPCIPv6CIDRBlockAssociationCreated(ctx, conn, associationID, vpcIPv6CIDRBlockAssociationCreatedTimeout)

		if err != nil {
			return "", fmt.Errorf("associating IPv6 CIDR block: waiting for completion: %w", err)
		}
	}

	return associationID, nil
}

func modifyVPCTenancy(ctx context.Context, conn *ec2.EC2, vpcID string, v string) error {
	input := &ec2.ModifyVpcTenancyInput{
		InstanceTenancy: aws.String(v),
		VpcId:           aws.String(vpcID),
	}

	if _, err := conn.ModifyVpcTenancyWithContext(ctx, input); err != nil {
		return fmt.Errorf("modifying Tenancy: %w", err)
	}

	return nil
}
