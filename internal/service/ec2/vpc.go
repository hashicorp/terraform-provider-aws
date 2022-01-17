package ec2

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
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
		Create: resourceVPCCreate,
		Read:   resourceVPCRead,
		Update: resourceVPCUpdate,
		Delete: resourceVPCDelete,
		Importer: &schema.ResourceImporter{
			State: resourceVPCImport,
		},

		CustomizeDiff: customdiff.All(
			resourceVPCCustomizeDiff,
			verify.SetTagsDiff,
			func(_ context.Context, diff *schema.ResourceDiff, v interface{}) error {
				// cidr_block can be set by a value returned from IPAM or explicitly in config
				if diff.Id() != "" && diff.HasChange("cidr_block") {
					// if netmask is set then cidr_block is derived from ipam, ignore changes
					if diff.Get("ipv4_netmask_length") != 0 {
						return diff.Clear("cidr_block")
					}
					return diff.ForceNew("cidr_block")
				}
				return nil
			},
		),

		SchemaVersion: 1,
		MigrateState:  VPCMigrateState,

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
			"dhcp_options_id": {
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
			"enable_classiclink": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"enable_classiclink_dns_support": {
				Type:     schema.TypeBool,
				Optional: true,
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
				ValidateFunc: validation.Any(
					validation.StringIsEmpty,
					validation.All(
						verify.ValidIPv6CIDRNetworkAddress,
						validation.IsCIDRNetwork(VPCCIDRMaxIPv6, VPCCIDRMaxIPv6)),
				),
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

func resourceVPCCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := &ec2.CreateVpcInput{
		AmazonProvidedIpv6CidrBlock: aws.Bool(d.Get("assign_generated_ipv6_cidr_block").(bool)),
		InstanceTenancy:             aws.String(d.Get("instance_tenancy").(string)),
		TagSpecifications:           ec2TagSpecificationsFromKeyValueTags(tags, ec2.ResourceTypeVpc),
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
	output, err := conn.CreateVpc(input)

	if err != nil {
		return fmt.Errorf("error creating EC2 VPC: %w", err)
	}

	d.SetId(aws.StringValue(output.Vpc.VpcId))

	vpc, err := WaitVPCCreated(conn, d.Id())

	if err != nil {
		return fmt.Errorf("error waiting for EC2 VPC (%s) create: %w", d.Id(), err)
	}

	if len(vpc.Ipv6CidrBlockAssociationSet) > 0 && vpc.Ipv6CidrBlockAssociationSet[0] != nil {
		associationID := aws.StringValue(output.Vpc.Ipv6CidrBlockAssociationSet[0].AssociationId)

		_, err = WaitVPCIPv6CIDRBlockAssociationCreated(conn, associationID)

		if err != nil {
			return fmt.Errorf("error waiting for EC2 VPC (%s) CIDR block (%s) to become associated: %w", d.Id(), associationID, err)
		}
	}

	vpcInfo := vpcInfo{
		vpc:                         vpc,
		enableClassicLink:           false,
		enableClassicLinkDNSSupport: false,
		enableDnsHostnames:          false,
		enableDnsSupport:            true,
	}

	if err := modifyVPCAttributesOnCreate(conn, d, &vpcInfo); err != nil {
		return err
	}

	return resourceVPCRead(d, meta)
}

func resourceVPCRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	outputRaw, err := tfresource.RetryWhenNewResourceNotFound(PropagationTimeout, func() (interface{}, error) {
		return FindVPCByID(conn, d.Id())
	}, d.IsNewResource())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 VPC %s not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading EC2 VPC (%s): %w", d.Id(), err)
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

	if v, err := FindVPCClassicLinkEnabled(conn, d.Id()); err != nil {
		return fmt.Errorf("error reading EC2 VPC (%s) ClassicLinkEnabled: %w", d.Id(), err)
	} else {
		d.Set("enable_classiclink", v)
	}

	if v, err := FindVPCClassicLinkDnsSupported(conn, d.Id()); err != nil {
		return fmt.Errorf("error reading EC2 VPC (%s) ClassicLinkDnsSupported: %w", d.Id(), err)
	} else {
		d.Set("enable_classiclink_dns_support", v)
	}

	if v, err := FindVPCAttribute(conn, d.Id(), ec2.VpcAttributeNameEnableDnsHostnames); err != nil {
		return fmt.Errorf("error reading EC2 VPC (%s) Attribute (%s): %w", d.Id(), ec2.VpcAttributeNameEnableDnsHostnames, err)
	} else {
		d.Set("enable_dns_hostnames", v)
	}

	if v, err := FindVPCAttribute(conn, d.Id(), ec2.VpcAttributeNameEnableDnsSupport); err != nil {
		return fmt.Errorf("error reading EC2 VPC (%s) Attribute (%s): %w", d.Id(), ec2.VpcAttributeNameEnableDnsSupport, err)
	} else {
		d.Set("enable_dns_support", v)
	}

	nacl, err := FindVPCDefaultNetworkACL(conn, d.Id())

	if err != nil {
		return fmt.Errorf("error reading EC2 VPC (%s) default NACL: %w", d.Id(), err)
	}

	d.Set("default_network_acl_id", nacl.NetworkAclId)

	routeTable, err := FindVPCMainRouteTable(conn, d.Id())

	if err != nil {
		return fmt.Errorf("error reading EC2 VPC (%s) main Route Table: %w", d.Id(), err)
	}

	d.Set("default_route_table_id", routeTable.RouteTableId)
	d.Set("main_route_table_id", routeTable.RouteTableId)

	securityGroup, err := FindVPCDefaultSecurityGroup(conn, d.Id())

	if err != nil {
		return fmt.Errorf("error reading EC2 VPC (%s) default Security Group: %w", d.Id(), err)
	}

	d.Set("default_security_group_id", securityGroup.GroupId)

	d.Set("assign_generated_ipv6_cidr_block", nil)
	d.Set("ipv6_cidr_block", nil)
	d.Set("ipv6_cidr_block_network_border_group", nil)
	d.Set("ipv6_ipam_pool_id", nil)
	d.Set("ipv6_netmask_length", nil)

	// Try and find IPv6 CIDR block information, first by any stored association ID.
	// Then if no IPv6 CIDR block information is available, use the first associated IPv6 CIDR block.
	var ipv6CIDRBlockAssociation *ec2.VpcIpv6CidrBlockAssociation
	if associationID := d.Get("ipv6_association_id").(string); associationID != "" {
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
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceVPCUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	if d.HasChange("enable_dns_hostnames") {
		if err := modifyVPCDnsHostnames(conn, d.Id(), d.Get("enable_dns_hostnames").(bool)); err != nil {
			return err
		}
	}

	if d.HasChange("enable_dns_support") {
		if err := modifyVPCDnsSupport(conn, d.Id(), d.Get("enable_dns_support").(bool)); err != nil {
			return err
		}
	}

	if d.HasChange("enable_classiclink") {
		if err := modifyVPCClassicLink(conn, d.Id(), d.Get("enable_classiclink").(bool)); err != nil {
			return err
		}
	}

	if d.HasChange("enable_classiclink_dns_support") {
		if err := modifyVPCClassicLinkDnsSupport(conn, d.Id(), d.Get("enable_classiclink_dns_support").(bool)); err != nil {
			return err
		}
	}

	if d.HasChange("instance_tenancy") {
		if err := modifyVPCTenancy(conn, d.Id(), d.Get("instance_tenancy").(string)); err != nil {
			return err
		}
	}

	vpcid := d.Id()

	if d.HasChanges("assign_generated_ipv6_cidr_block", "ipv6_cidr_block_network_border_group") {
		associationID, err := modifyVPCIPv6CIDRBlockAssociation(conn, d.Id(),
			d.Get("ipv6_association_id").(string),
			d.Get("assign_generated_ipv6_cidr_block").(bool),
			"",
			"",
			0,
			d.Get("ipv6_cidr_block_network_border_group").(string))

		if err != nil {
			return err
		}

		d.Set("ipv6_association_id", associationID)
		// toAssign := d.Get("assign_generated_ipv6_cidr_block").(bool)
		// borderNetworkGroup := d.Get("ipv6_cidr_block_network_border_group").(string)
		// existingCIDR := d.Get("ipv6_cidr_block").(string)

		// log.Printf("[INFO] Modifying assign_generated_ipv6_cidr_block to %#v", toAssign)

		// if toAssign && borderNetworkGroup != "" {
		// 	// if an existing IPv6 CIDR block is assigned, we need to unassign it first
		// 	if existingCIDR != "" {
		// 		associationID := d.Get("ipv6_association_id").(string)
		// 		modifyOpts := &ec2.DisassociateVpcCidrBlockInput{
		// 			AssociationId: aws.String(associationID),
		// 		}
		// 		log.Printf("[INFO] Disabling assign_generated_ipv6_cidr_block vpc attribute for %s: %#v",
		// 			d.Id(), modifyOpts)
		// 		if _, err := conn.DisassociateVpcCidrBlock(modifyOpts); err != nil {
		// 			return err
		// 		}

		// 		log.Printf("[DEBUG] Waiting for EC2 VPC (%s) IPv6 CIDR to become disassociated", d.Id())

		// 		if err := waitForEc2VpcIpv6CidrBlockAssociationDelete(conn, d.Id(), associationID); err != nil {
		// 			return fmt.Errorf("error waiting for EC2 VPC (%s) IPv6 CIDR to become disassociated: %s", d.Id(), err)
		// 		}
		// 	}

		// 	modifyOpts := &ec2.AssociateVpcCidrBlockInput{
		// 		VpcId:                           &vpcid,
		// 		AmazonProvidedIpv6CidrBlock:     aws.Bool(toAssign),
		// 		Ipv6CidrBlockNetworkBorderGroup: aws.String(borderNetworkGroup),
		// 	}
		// 	log.Printf("[INFO] Enabling assign_generated_ipv6_cidr_block vpc attribute for %s: %#v with border network group %s",
		// 		d.Id(), modifyOpts, borderNetworkGroup)

		// 	resp, err := conn.AssociateVpcCidrBlock(modifyOpts)

		// 	if err != nil {
		// 		return err
		// 	}

		// 	log.Printf("[DEBUG] Waiting for EC2 VPC (%s) IPv6 CIDR to become associated", d.Id())

		// 	if err := waitForEc2VpcIpv6CidrBlockAssociationCreate(conn, d.Id(), aws.StringValue(resp.Ipv6CidrBlockAssociation.AssociationId)); err != nil {
		// 		return fmt.Errorf("error waiting for EC2 VPC (%s) IPv6 CIDR to become associated: %s", d.Id(), err)
		// 	}
		// }
		// // if no IPv6 CIDR block is assigned, we need to unassign the existing one
		// if !toAssign {
		// 	associationID := d.Get("ipv6_association_id").(string)
		// 	modifyOpts := &ec2.DisassociateVpcCidrBlockInput{
		// 		AssociationId: aws.String(associationID),
		// 	}
		// 	log.Printf("[INFO] Disabling assign_generated_ipv6_cidr_block vpc attribute for %s: %#v",
		// 		d.Id(), modifyOpts)
		// 	if _, err := conn.DisassociateVpcCidrBlock(modifyOpts); err != nil {
		// 		return err
		// 	}

		// 	log.Printf("[DEBUG] Waiting for EC2 VPC (%s) IPv6 CIDR to become disassociated", d.Id())
		// 	if err := waitForEc2VpcIpv6CidrBlockAssociationDelete(conn, d.Id(), associationID); err != nil {
		// 		return fmt.Errorf("error waiting for EC2 VPC (%s) IPv6 CIDR to become disassociated: %s", d.Id(), err)
		// 	}
		// }
		// // if an IPv6 CIDR blosk is to be assigned and no network border group is specified
		// // just create the new association and remove the existing one if a border group is configured
		// if toAssign && borderNetworkGroup == "" {
		// 	log.Printf("[INFO] Modifying IPv6 Block Network Border Group")
		// 	modifyOpts := &ec2.AssociateVpcCidrBlockInput{
		// 		VpcId:                       &vpcid,
		// 		AmazonProvidedIpv6CidrBlock: aws.Bool(d.Get("assign_generated_ipv6_cidr_block").(bool)),
		// 	}
		// 	if val := d.Get("ipv6_cidr_block"); val != "" {
		// 		log.Printf("[INFO] Disabling assign_generated_ipv6_cidr_block vpc attribute for %s: %#v",
		// 			d.Id(), modifyOpts)
		// 		disassociationID := d.Get("ipv6_association_id").(string)
		// 		disModifyOpts := &ec2.DisassociateVpcCidrBlockInput{
		// 			AssociationId: aws.String(disassociationID),
		// 		}
		// 		log.Printf("[INFO] Dissaociating IPv6 Block Network Border Group")
		// 		if _, err := conn.DisassociateVpcCidrBlock(disModifyOpts); err != nil {
		// 			return err
		// 		}
		// 	}

		// 	if v := d.Get("ipv6_cidr_block_network_border_group"); v != "" {
		// 		modifyOpts.Ipv6CidrBlockNetworkBorderGroup = aws.String(v.(string))
		// 		log.Printf("[INFO] Trying to associate IPv6 Block Network Border Group")
		// 		if _, err := conn.AssociateVpcCidrBlock(modifyOpts); err != nil {
		// 			return err
		// 		}
		// 	}
		// 	if v := d.Get("ipv6_cidr_block_network_border_group"); v == "" {
		// 		associationID := d.Get("ipv6_association_id").(string)
		// 		modifyOpts := &ec2.DisassociateVpcCidrBlockInput{
		// 			AssociationId: aws.String(associationID),
		// 		}
		// 		log.Printf("[INFO] Dissaociating IPv6 Block Network Border Group")
		// 		if _, err := conn.DisassociateVpcCidrBlock(modifyOpts); err != nil {
		// 			return err
		// 		}
		// 		if d.Get("assign_generated_ipv6_cidr_block").(bool) {
		// 			log.Printf("[INFO] Trying to associate IPv6 Block Network Border Group")
		// 			modifyOpts := &ec2.AssociateVpcCidrBlockInput{
		// 				VpcId:                       &vpcid,
		// 				AmazonProvidedIpv6CidrBlock: aws.Bool(d.Get("assign_generated_ipv6_cidr_block").(bool)),
		// 			}
		// 			if _, err := conn.AssociateVpcCidrBlock(modifyOpts); err != nil {
		// 				return err
		// 			}
		// 		}
		// 	}
		// }
	}

	if d.HasChanges("ipv6_cidr_block", "ipv6_ipam_pool_id") {
		log.Printf("[INFO] Modifying ipam IPv6 CIDR")

		// if assoc id exists it needs to be disassociated
		if v, ok := d.GetOk("ipv6_association_id"); ok {
			if err := ipv6DisassociateCidrBlock(conn, d.Id(), v.(string)); err != nil {
				return err
			}
		}
		if v := d.Get("ipv6_ipam_pool_id"); v != "" {
			modifyOpts := &ec2.AssociateVpcCidrBlockInput{
				VpcId:          &vpcid,
				Ipv6IpamPoolId: aws.String(v.(string)),
			}

			if v := d.Get("ipv6_netmask_length"); v != 0 {
				modifyOpts.Ipv6NetmaskLength = aws.Int64(int64(v.(int)))
			}

			if v := d.Get("ipv6_cidr_block"); v != "" {
				modifyOpts.Ipv6CidrBlock = aws.String(v.(string))
			}

			resp, err := conn.AssociateVpcCidrBlock(modifyOpts)
			if err != nil {
				return err
			}
			if err := waitForEc2VpcIpv6CidrBlockAssociationCreate(conn, d.Id(), aws.StringValue(resp.Ipv6CidrBlockAssociation.AssociationId)); err != nil {
				return fmt.Errorf("error waiting for EC2 VPC (%s) IPv6 CIDR to become associated: %w", d.Id(), err)
			}
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating EC2 VPC (%s) tags: %w", d.Id(), err)
		}
	}

	return resourceVPCRead(d, meta)
}

func resourceVPCDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	input := &ec2.DeleteVpcInput{
		VpcId: aws.String(d.Id()),
	}

	log.Printf("[INFO] Deleting EC2 VPC: %s", d.Id())
	_, err := tfresource.RetryWhenAWSErrCodeEquals(vpcDeletedTimeout, func() (interface{}, error) {
		return conn.DeleteVpc(input)
	}, ErrCodeDependencyViolation)

	if tfawserr.ErrCodeEquals(err, ErrCodeInvalidVpcIDNotFound) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting EC2 VPC (%s): %w", d.Id(), err)
	}

	return nil
}

func ipv6DisassociateCidrBlock(conn *ec2.EC2, id, allocationId string) error {
	log.Printf("[INFO] Disassociating IPv6 CIDR association id: %s", allocationId)
	modifyOpts := &ec2.DisassociateVpcCidrBlockInput{
		AssociationId: aws.String(allocationId),
	}
	if _, err := conn.DisassociateVpcCidrBlock(modifyOpts); err != nil {
		return err
	}
	log.Printf("[DEBUG] Waiting for EC2 VPC (%s) IPv6 CIDR to become disassociated", id)
	if err := waitForEc2VpcIpv6CidrBlockAssociationDelete(conn, id, allocationId); err != nil {
		return fmt.Errorf("error waiting for EC2 VPC (%s) IPv6 CIDR to become disassociated: %w", id, err)
	}

	return nil
}

func resourceVPCCustomizeDiff(_ context.Context, diff *schema.ResourceDiff, v interface{}) error {
	if diff.HasChange("instance_tenancy") {
		old, new := diff.GetChange("instance_tenancy")
		if old.(string) != ec2.TenancyDedicated || new.(string) != ec2.TenancyDefault {
			diff.ForceNew("instance_tenancy")
		}
	}

	return nil
}

func Ipv6CidrStateRefreshFunc(conn *ec2.EC2, id string, associationId string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		describeVpcOpts := &ec2.DescribeVpcsInput{
			VpcIds: []*string{aws.String(id)},
		}
		resp, err := conn.DescribeVpcs(describeVpcOpts)

		if tfawserr.ErrMessageContains(err, "InvalidVpcID.NotFound", "") {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		if resp == nil || len(resp.Vpcs) == 0 || resp.Vpcs[0] == nil || resp.Vpcs[0].Ipv6CidrBlockAssociationSet == nil {
			// Sometimes AWS just has consistency issues and doesn't see
			// our instance yet. Return an empty state.
			return nil, "", nil
		}

		for _, association := range resp.Vpcs[0].Ipv6CidrBlockAssociationSet {
			if aws.StringValue(association.AssociationId) == associationId {
				return association, aws.StringValue(association.Ipv6CidrBlockState.State), nil
			}
		}

		return nil, "", nil
	}
}

func resourceVPCImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	d.Set("assign_generated_ipv6_cidr_block", false)
	return []*schema.ResourceData{d}, nil
}

// vpcDescribe returns EC2 API information about the specified VPC.
// If the VPC doesn't exist, return nil.
func vpcDescribe(conn *ec2.EC2, vpcId string) (*ec2.Vpc, error) {
	resp, err := conn.DescribeVpcs(&ec2.DescribeVpcsInput{
		VpcIds: aws.StringSlice([]string{vpcId}),
	})
	if err != nil {
		if !tfawserr.ErrMessageContains(err, "InvalidVpcID.NotFound", "") {
			return nil, err
		}
		resp = nil
	}

	if resp == nil {
		return nil, nil
	}

	n := len(resp.Vpcs)
	switch n {
	case 0:
		return nil, nil

	case 1:
		return resp.Vpcs[0], nil

	default:
		return nil, fmt.Errorf("Found %d VPCs for %s, expected 1", n, vpcId)
	}
}

func waitForEc2VpcIpv6CidrBlockAssociationCreate(conn *ec2.EC2, vpcID, associationID string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			ec2.VpcCidrBlockStateCodeAssociating,
			ec2.VpcCidrBlockStateCodeDisassociated,
		},
		Target:  []string{ec2.VpcCidrBlockStateCodeAssociated},
		Refresh: Ipv6CidrStateRefreshFunc(conn, vpcID, associationID),
		Timeout: 10 * time.Minute,
	}
	_, err := stateConf.WaitForState()

	return err
}

func waitForEc2VpcIpv6CidrBlockAssociationDelete(conn *ec2.EC2, vpcID, associationID string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			ec2.VpcCidrBlockStateCodeAssociated,
			ec2.VpcCidrBlockStateCodeDisassociating,
		},
		Target:         []string{ec2.VpcCidrBlockStateCodeDisassociated},
		Refresh:        Ipv6CidrStateRefreshFunc(conn, vpcID, associationID),
		Timeout:        5 * time.Minute,
		NotFoundChecks: 1,
	}
	_, err := stateConf.WaitForState()

	return err
}

type vpcInfo struct {
	vpc                         *ec2.Vpc
	enableClassicLink           bool
	enableClassicLinkDNSSupport bool
	enableDnsHostnames          bool
	enableDnsSupport            bool
}

// modifyVPCAttributesOnCreate sets VPC attributes on resource Create.
// Called after new VPC creation or existing default VPC adoption.
func modifyVPCAttributesOnCreate(conn *ec2.EC2, d *schema.ResourceData, vpcInfo *vpcInfo) error {
	if new, old := d.Get("enable_dns_hostnames").(bool), vpcInfo.enableDnsHostnames; old != new {
		if err := modifyVPCDnsHostnames(conn, d.Id(), new); err != nil {
			return err
		}
	}

	if new, old := d.Get("enable_dns_support").(bool), vpcInfo.enableDnsSupport; old != new {
		if err := modifyVPCDnsSupport(conn, d.Id(), new); err != nil {
			return err
		}
	}

	if new, old := d.Get("enable_classiclink").(bool), vpcInfo.enableClassicLink; old != new {
		if err := modifyVPCClassicLink(conn, d.Id(), new); err != nil {
			return err
		}
	}

	if new, old := d.Get("enable_classiclink_dns_support").(bool), vpcInfo.enableClassicLinkDNSSupport; old != new {
		if err := modifyVPCClassicLinkDnsSupport(conn, d.Id(), new); err != nil {
			return err
		}
	}

	return nil
}

func modifyVPCClassicLink(conn *ec2.EC2, vpcID string, v bool) error {
	if v {
		input := &ec2.EnableVpcClassicLinkInput{
			VpcId: aws.String(vpcID),
		}

		if _, err := conn.EnableVpcClassicLink(input); err != nil {
			return fmt.Errorf("error enabling EC2 VPC (%s) ClassicLink: %w", vpcID, err)
		}
	} else {
		input := &ec2.DisableVpcClassicLinkInput{
			VpcId: aws.String(vpcID),
		}

		if _, err := conn.DisableVpcClassicLink(input); err != nil {
			return fmt.Errorf("error disabling EC2 VPC (%s) ClassicLink: %w", vpcID, err)
		}
	}

	return nil
}

func modifyVPCClassicLinkDnsSupport(conn *ec2.EC2, vpcID string, v bool) error {
	if v {
		input := &ec2.EnableVpcClassicLinkDnsSupportInput{
			VpcId: aws.String(vpcID),
		}

		if _, err := conn.EnableVpcClassicLinkDnsSupport(input); err != nil {
			return fmt.Errorf("error enabling EC2 VPC (%s) ClassicLinkDnsSupport: %w", vpcID, err)
		}
	} else {
		input := &ec2.DisableVpcClassicLinkDnsSupportInput{
			VpcId: aws.String(vpcID),
		}

		if _, err := conn.DisableVpcClassicLinkDnsSupport(input); err != nil {
			return fmt.Errorf("error disabling EC2 VPC (%s) ClassicLinkDnsSupport: %w", vpcID, err)
		}
	}

	return nil
}

func modifyVPCDnsHostnames(conn *ec2.EC2, vpcID string, v bool) error {
	input := &ec2.ModifyVpcAttributeInput{
		EnableDnsHostnames: &ec2.AttributeBooleanValue{
			Value: aws.Bool(v),
		},
		VpcId: aws.String(vpcID),
	}

	if _, err := conn.ModifyVpcAttribute(input); err != nil {
		return fmt.Errorf("error modifying EC2 VPC (%s) EnableDnsHostnames: %w", vpcID, err)
	}

	if _, err := WaitVPCAttributeUpdated(conn, vpcID, ec2.VpcAttributeNameEnableDnsHostnames, v); err != nil {
		return fmt.Errorf("error waiting for EC2 VPC (%s) EnableDnsHostnames update: %w", vpcID, err)
	}

	return nil
}

func modifyVPCDnsSupport(conn *ec2.EC2, vpcID string, v bool) error {
	input := &ec2.ModifyVpcAttributeInput{
		EnableDnsSupport: &ec2.AttributeBooleanValue{
			Value: aws.Bool(v),
		},
		VpcId: aws.String(vpcID),
	}

	if _, err := conn.ModifyVpcAttribute(input); err != nil {
		return fmt.Errorf("error modifying EC2 VPC (%s) EnableDnsSupport: %w", vpcID, err)
	}

	if _, err := WaitVPCAttributeUpdated(conn, vpcID, ec2.VpcAttributeNameEnableDnsSupport, v); err != nil {
		return fmt.Errorf("error waiting for EC2 VPC (%s) EnableDnsSupport update: %w", vpcID, err)
	}

	return nil
}

// modifyVPCIPv6CIDRBlockAssociation modify's a VPC's IPv6 CIDR block association.
// Any exiting association is deleted and any new association's ID is returned.
func modifyVPCIPv6CIDRBlockAssociation(conn *ec2.EC2, vpcID, associationID string, amazonProvidedCIDRBlock bool, cidrBlock, ipamPoolID string, netmaskLength int, networkBorderGroup string) (string, error) {
	if associationID != "" {
		input := &ec2.DisassociateVpcCidrBlockInput{
			AssociationId: aws.String(associationID),
		}

		_, err := conn.DisassociateVpcCidrBlock(input)

		if err != nil {
			return "", fmt.Errorf("error disassociating EC2 VPC (%s) CIDR block (%s): %w", vpcID, associationID, err)
		}

		_, err = WaitVPCIPv6CIDRBlockAssociationDeleted(conn, associationID)

		if err != nil {
			return "", fmt.Errorf("error waiting for EC2 VPC (%s) CIDR block (%s) to become disassociated: %w", vpcID, associationID, err)
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

		output, err := conn.AssociateVpcCidrBlock(input)

		if err != nil {
			return "", fmt.Errorf("error associating EC2 VPC (%s) IPv6 CIDR block: %w", vpcID, err)
		}

		associationID = aws.StringValue(output.Ipv6CidrBlockAssociation.AssociationId)

		_, err = WaitVPCIPv6CIDRBlockAssociationCreated(conn, associationID)

		if err != nil {
			return "", fmt.Errorf("error waiting for EC2 VPC (%s) CIDR block (%s) to become associated: %w", vpcID, associationID, err)
		}
	}

	return associationID, nil
}

func modifyVPCTenancy(conn *ec2.EC2, vpcID string, v string) error {
	input := &ec2.ModifyVpcTenancyInput{
		InstanceTenancy: aws.String(v),
		VpcId:           aws.String(vpcID),
	}

	if _, err := conn.ModifyVpcTenancy(input); err != nil {
		return fmt.Errorf("error modifying EC2 VPC (%s) Tenancy: %w", vpcID, err)
	}

	return nil
}
