package ec2

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/aws/awserr"
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

// acceptance tests for byoip related tests are in vpc_byoip_test.go
func ResourceVPC() *schema.Resource {
	//lintignore:R011
	return &schema.Resource{
		Create: resourceVPCCreate,
		Read:   resourceVPCRead,
		Update: resourceVPCUpdate,
		Delete: resourceVPCDelete,
		Importer: &schema.ResourceImporter{
			State: resourceVPCInstanceImport,
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
			"default_security_group_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"default_route_table_id": {
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

	// Create the VPC
	createOpts := &ec2.CreateVpcInput{
		InstanceTenancy:             aws.String(d.Get("instance_tenancy").(string)),
		AmazonProvidedIpv6CidrBlock: aws.Bool(d.Get("assign_generated_ipv6_cidr_block").(bool)),
		TagSpecifications:           ec2TagSpecificationsFromKeyValueTags(tags, ec2.ResourceTypeVpc),
	}

	if v, ok := d.GetOk("cidr_block"); ok {
		createOpts.CidrBlock = aws.String(v.(string))
	}

	if v, ok := d.GetOk("ipv4_ipam_pool_id"); ok {
		createOpts.Ipv4IpamPoolId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("ipv4_netmask_length"); ok {
		createOpts.Ipv4NetmaskLength = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("ipv6_ipam_pool_id"); ok {
		createOpts.Ipv6IpamPoolId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("ipv6_cidr_block"); ok {
		createOpts.Ipv6CidrBlock = aws.String(v.(string))
	}

	if v, ok := d.GetOk("ipv6_netmask_length"); ok {
		createOpts.Ipv6NetmaskLength = aws.Int64(int64(v.(int)))
	}

	log.Printf("[DEBUG] VPC create config: %#v", *createOpts)
	vpcResp, err := conn.CreateVpc(createOpts)
	if err != nil {
		return fmt.Errorf("Error creating VPC: %s", err)
	}

	// Get the ID and store it
	vpc := vpcResp.Vpc
	d.SetId(aws.StringValue(vpc.VpcId))
	log.Printf("[INFO] VPC ID: %s", d.Id())

	// Wait for the VPC to become available
	log.Printf(
		"[DEBUG] Waiting for VPC (%s) to become available",
		d.Id())
	stateConf := &resource.StateChangeConf{
		Pending: []string{"pending"},
		Target:  []string{"available"},
		Refresh: VPCStateRefreshFunc(conn, d.Id()),
		Timeout: 10 * time.Minute,
	}
	if _, err := stateConf.WaitForState(); err != nil {
		return fmt.Errorf(
			"Error waiting for VPC (%s) to become available: %s",
			d.Id(), err)
	}

	if len(vpc.Ipv6CidrBlockAssociationSet) > 0 && vpc.Ipv6CidrBlockAssociationSet[0] != nil {
		log.Printf("[DEBUG] Waiting for EC2 VPC (%s) IPv6 CIDR to become associated", d.Id())
		if err := waitForEc2VpcIpv6CidrBlockAssociationCreate(conn, d.Id(), aws.StringValue(vpcResp.Vpc.Ipv6CidrBlockAssociationSet[0].AssociationId)); err != nil {
			return fmt.Errorf("error waiting for EC2 VPC (%s) IPv6 CIDR to become associated: %s", d.Id(), err)
		}
	}

	// You cannot modify the DNS resolution and DNS hostnames attributes in the same request. Use separate requests for each attribute.
	// Reference: https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_ModifyVpcAttribute.html

	if d.Get("enable_dns_hostnames").(bool) {
		input := &ec2.ModifyVpcAttributeInput{
			EnableDnsHostnames: &ec2.AttributeBooleanValue{
				Value: aws.Bool(true),
			},
			VpcId: aws.String(d.Id()),
		}

		if _, err := conn.ModifyVpcAttribute(input); err != nil {
			return fmt.Errorf("error enabling EC2 VPC (%s) DNS Hostnames: %w", d.Id(), err)
		}

		if _, err := WaitVPCAttributeUpdated(conn, d.Id(), ec2.VpcAttributeNameEnableDnsHostnames, d.Get("enable_dns_hostnames").(bool)); err != nil {
			return fmt.Errorf("error waiting for EC2 VPC (%s) DNS Hostnames to enable: %w", d.Id(), err)
		}
	}

	// By default, only the enableDnsSupport attribute is set to true in a VPC created any other way.
	// Reference: https://docs.aws.amazon.com/vpc/latest/userguide/vpc-dns.html#vpc-dns-support

	if !d.Get("enable_dns_support").(bool) {
		input := &ec2.ModifyVpcAttributeInput{
			EnableDnsSupport: &ec2.AttributeBooleanValue{
				Value: aws.Bool(false),
			},
			VpcId: aws.String(d.Id()),
		}

		if _, err := conn.ModifyVpcAttribute(input); err != nil {
			return fmt.Errorf("error disabling EC2 VPC (%s) DNS Support: %w", d.Id(), err)
		}

		if _, err := WaitVPCAttributeUpdated(conn, d.Id(), ec2.VpcAttributeNameEnableDnsSupport, d.Get("enable_dns_support").(bool)); err != nil {
			return fmt.Errorf("error waiting for EC2 VPC (%s) DNS Support to disable: %w", d.Id(), err)
		}
	}

	if d.Get("enable_classiclink").(bool) {
		input := &ec2.EnableVpcClassicLinkInput{
			VpcId: aws.String(d.Id()),
		}

		if _, err := conn.EnableVpcClassicLink(input); err != nil {
			return fmt.Errorf("error enabling VPC (%s) ClassicLink: %s", d.Id(), err)
		}
	}

	if d.Get("enable_classiclink_dns_support").(bool) {
		input := &ec2.EnableVpcClassicLinkDnsSupportInput{
			VpcId: aws.String(d.Id()),
		}

		if _, err := conn.EnableVpcClassicLinkDnsSupport(input); err != nil {
			return fmt.Errorf("error enabling VPC (%s) ClassicLink DNS support: %s", d.Id(), err)
		}
	}

	return resourceVPCRead(d, meta)
}

func resourceVPCRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	var vpc *ec2.Vpc

	err := resource.Retry(VPCPropagationTimeout, func() *resource.RetryError {
		var err error

		vpc, err = FindVPCByID(conn, d.Id())

		if d.IsNewResource() && tfawserr.ErrCodeEquals(err, "InvalidVpcID.NotFound") {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		if d.IsNewResource() && vpc == nil {
			return resource.RetryableError(&resource.NotFoundError{
				LastError: fmt.Errorf("EC2 VPC (%s) not found", d.Id()),
			})
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		vpc, err = FindVPCByID(conn, d.Id())
	}

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, "InvalidVpcID.NotFound") {
		log.Printf("[WARN] EC2 VPC (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading EC2 VPC (%s): %w", d.Id(), err)
	}

	if vpc == nil {
		if d.IsNewResource() {
			return fmt.Errorf("error reading EC2 VPC (%s): not found after creation", d.Id())
		}

		log.Printf("[WARN] EC2 VPC (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	vpcid := d.Id()
	d.Set("cidr_block", vpc.CidrBlock)
	d.Set("dhcp_options_id", vpc.DhcpOptionsId)
	d.Set("instance_tenancy", vpc.InstanceTenancy)

	// ARN
	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   ec2.ServiceName,
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: aws.StringValue(vpc.OwnerId),
		Resource:  fmt.Sprintf("vpc/%s", d.Id()),
	}.String()
	d.Set("arn", arn)

	tags := KeyValueTags(vpc.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %s", err)
	}

	d.Set("owner_id", vpc.OwnerId)

	// Make sure those values are set, if an IPv6 block exists it'll be set in the loop
	d.Set("ipv6_association_id", "")
	d.Set("ipv6_cidr_block", "")
	// assign_generated_ipv6_cidr_block is not returned by the API
	// leave unassigned if not referenced
	if v := d.Get("assign_generated_ipv6_cidr_block"); v != "" {
		d.Set("assign_generated_ipv6_cidr_block", aws.Bool(v.(bool)))
	}
	for _, a := range vpc.Ipv6CidrBlockAssociationSet {
		if aws.StringValue(a.Ipv6CidrBlockState.State) == ec2.VpcCidrBlockStateCodeAssociated { //we can only ever have 1 IPv6 block associated at once
			d.Set("ipv6_association_id", a.AssociationId)
			d.Set("ipv6_cidr_block", a.Ipv6CidrBlock)
		}
	}

	enableDnsHostnames, err := FindVPCAttribute(conn, aws.StringValue(vpc.VpcId), ec2.VpcAttributeNameEnableDnsHostnames)

	if err != nil {
		return fmt.Errorf("error reading EC2 VPC (%s) Attribute (%s): %w", aws.StringValue(vpc.VpcId), ec2.VpcAttributeNameEnableDnsHostnames, err)
	}

	d.Set("enable_dns_hostnames", enableDnsHostnames)

	enableDnsSupport, err := FindVPCAttribute(conn, aws.StringValue(vpc.VpcId), ec2.VpcAttributeNameEnableDnsSupport)

	if err != nil {
		return fmt.Errorf("error reading EC2 VPC (%s) Attribute (%s): %w", aws.StringValue(vpc.VpcId), ec2.VpcAttributeNameEnableDnsSupport, err)
	}

	d.Set("enable_dns_support", enableDnsSupport)

	describeClassiclinkOpts := &ec2.DescribeVpcClassicLinkInput{
		VpcIds: []*string{&vpcid},
	}

	// Classic Link is only available in regions that support EC2 Classic
	respClassiclink, err := conn.DescribeVpcClassicLink(describeClassiclinkOpts)
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() == "UnsupportedOperation" {
			log.Printf("[WARN] VPC Classic Link is not supported in this region")
		} else if tfawserr.ErrMessageContains(err, "InvalidVpcID.NotFound", "") {
			log.Printf("[WARN] VPC Classic Link functionality you requested is not available for this VPC")
		} else {
			return err
		}
	} else {
		classiclink_enabled := false
		for _, v := range respClassiclink.Vpcs {
			if aws.StringValue(v.VpcId) == vpcid {
				if v.ClassicLinkEnabled != nil {
					classiclink_enabled = aws.BoolValue(v.ClassicLinkEnabled)
				}
				break
			}
		}
		d.Set("enable_classiclink", classiclink_enabled)
	}

	describeClassiclinkDnsOpts := &ec2.DescribeVpcClassicLinkDnsSupportInput{
		VpcIds: []*string{&vpcid},
	}

	respClassiclinkDnsSupport, err := conn.DescribeVpcClassicLinkDnsSupport(describeClassiclinkDnsOpts)
	if err != nil {
		if tfawserr.ErrMessageContains(err, "UnsupportedOperation", "The functionality you requested is not available in this region") ||
			tfawserr.ErrMessageContains(err, "AuthFailure", "This request has been administratively disabled") {
			log.Printf("[WARN] VPC Classic Link DNS Support is not supported in this region")
		} else {
			return err
		}
	} else {
		classiclinkdns_enabled := false
		for _, v := range respClassiclinkDnsSupport.Vpcs {
			if aws.StringValue(v.VpcId) == vpcid {
				if v.ClassicLinkDnsSupported != nil {
					classiclinkdns_enabled = aws.BoolValue(v.ClassicLinkDnsSupported)
				}
				break
			}
		}
		d.Set("enable_classiclink_dns_support", classiclinkdns_enabled)
	}

	routeTableId, err := resourceVPCSetMainRouteTable(conn, vpcid)
	if err != nil {
		log.Printf("[WARN] Unable to set Main Route Table: %s", err)
	}
	d.Set("main_route_table_id", routeTableId)

	if err := resourceVPCSetDefaultNetworkACL(conn, d); err != nil {
		log.Printf("[WARN] Unable to set Default Network ACL: %s", err)
	}
	if err := resourceVPCSetDefaultSecurityGroup(conn, d); err != nil {
		log.Printf("[WARN] Unable to set Default Security Group: %s", err)
	}
	if err := resourceVPCSetDefaultRouteTable(conn, d); err != nil {
		log.Printf("[WARN] Unable to set Default Route Table: %s", err)
	}

	return nil
}

func resourceVPCUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	vpcid := d.Id()
	if d.HasChange("enable_dns_hostnames") {
		input := &ec2.ModifyVpcAttributeInput{
			VpcId: aws.String(d.Id()),
			EnableDnsHostnames: &ec2.AttributeBooleanValue{
				Value: aws.Bool(d.Get("enable_dns_hostnames").(bool)),
			},
		}

		if _, err := conn.ModifyVpcAttribute(input); err != nil {
			return fmt.Errorf("error updating EC2 VPC (%s) DNS Hostnames: %w", d.Id(), err)
		}

		if _, err := WaitVPCAttributeUpdated(conn, d.Id(), ec2.VpcAttributeNameEnableDnsHostnames, d.Get("enable_dns_hostnames").(bool)); err != nil {
			return fmt.Errorf("error waiting for EC2 VPC (%s) DNS Hostnames update: %w", d.Id(), err)
		}
	}

	_, hasEnableDnsSupportOption := d.GetOk("enable_dns_support")

	if !hasEnableDnsSupportOption || d.HasChange("enable_dns_support") {
		input := &ec2.ModifyVpcAttributeInput{
			VpcId: aws.String(d.Id()),
			EnableDnsSupport: &ec2.AttributeBooleanValue{
				Value: aws.Bool(d.Get("enable_dns_support").(bool)),
			},
		}

		if _, err := conn.ModifyVpcAttribute(input); err != nil {
			return fmt.Errorf("error updating EC2 VPC (%s) DNS Support: %w", d.Id(), err)
		}

		if _, err := WaitVPCAttributeUpdated(conn, d.Id(), ec2.VpcAttributeNameEnableDnsSupport, d.Get("enable_dns_support").(bool)); err != nil {
			return fmt.Errorf("error waiting for EC2 VPC (%s) DNS Support update: %w", d.Id(), err)
		}
	}

	if d.HasChange("enable_classiclink") {
		val := d.Get("enable_classiclink").(bool)
		if val {
			modifyOpts := &ec2.EnableVpcClassicLinkInput{
				VpcId: &vpcid,
			}
			log.Printf(
				"[INFO] Modifying enable_classiclink vpc attribute for %s: %#v",
				d.Id(), modifyOpts)
			if _, err := conn.EnableVpcClassicLink(modifyOpts); err != nil {
				return err
			}
		} else {
			modifyOpts := &ec2.DisableVpcClassicLinkInput{
				VpcId: &vpcid,
			}
			log.Printf(
				"[INFO] Modifying enable_classiclink vpc attribute for %s: %#v",
				d.Id(), modifyOpts)
			if _, err := conn.DisableVpcClassicLink(modifyOpts); err != nil {
				return err
			}
		}
	}

	if d.HasChange("enable_classiclink_dns_support") {
		val := d.Get("enable_classiclink_dns_support").(bool)
		if val {
			modifyOpts := &ec2.EnableVpcClassicLinkDnsSupportInput{
				VpcId: &vpcid,
			}
			log.Printf(
				"[INFO] Modifying enable_classiclink_dns_support vpc attribute for %s: %#v",
				d.Id(), modifyOpts)
			if _, err := conn.EnableVpcClassicLinkDnsSupport(modifyOpts); err != nil {
				return err
			}
		} else {
			modifyOpts := &ec2.DisableVpcClassicLinkDnsSupportInput{
				VpcId: &vpcid,
			}
			log.Printf(
				"[INFO] Modifying enable_classiclink_dns_support vpc attribute for %s: %#v",
				d.Id(), modifyOpts)
			if _, err := conn.DisableVpcClassicLinkDnsSupport(modifyOpts); err != nil {
				return err
			}
		}
	}

	if d.HasChange("assign_generated_ipv6_cidr_block") {
		toAssign := d.Get("assign_generated_ipv6_cidr_block").(bool)

		log.Printf("[INFO] Modifying assign_generated_ipv6_cidr_block to %#v", toAssign)

		if toAssign {
			modifyOpts := &ec2.AssociateVpcCidrBlockInput{
				VpcId:                       &vpcid,
				AmazonProvidedIpv6CidrBlock: aws.Bool(toAssign),
			}
			log.Printf("[INFO] Enabling assign_generated_ipv6_cidr_block vpc attribute for %s: %#v",
				d.Id(), modifyOpts)
			resp, err := conn.AssociateVpcCidrBlock(modifyOpts)
			if err != nil {
				return err
			}

			log.Printf("[DEBUG] Waiting for EC2 VPC (%s) IPv6 CIDR to become associated", d.Id())
			if err := waitForEc2VpcIpv6CidrBlockAssociationCreate(conn, d.Id(), aws.StringValue(resp.Ipv6CidrBlockAssociation.AssociationId)); err != nil {
				return fmt.Errorf("error waiting for EC2 VPC (%s) IPv6 CIDR to become associated: %s", d.Id(), err)
			}
		} else {
			associationID := d.Get("ipv6_association_id").(string)
			modifyOpts := &ec2.DisassociateVpcCidrBlockInput{
				AssociationId: aws.String(associationID),
			}
			log.Printf("[INFO] Disabling assign_generated_ipv6_cidr_block vpc attribute for %s: %#v",
				d.Id(), modifyOpts)
			if _, err := conn.DisassociateVpcCidrBlock(modifyOpts); err != nil {
				return err
			}

			log.Printf("[DEBUG] Waiting for EC2 VPC (%s) IPv6 CIDR to become disassociated", d.Id())
			if err := waitForEc2VpcIpv6CidrBlockAssociationDelete(conn, d.Id(), associationID); err != nil {
				return fmt.Errorf("error waiting for EC2 VPC (%s) IPv6 CIDR to become disassociated: %s", d.Id(), err)
			}
		}
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

	if d.HasChange("instance_tenancy") {
		modifyOpts := &ec2.ModifyVpcTenancyInput{
			VpcId:           aws.String(vpcid),
			InstanceTenancy: aws.String(d.Get("instance_tenancy").(string)),
		}
		log.Printf(
			"[INFO] Modifying instance_tenancy vpc attribute for %s: %#v",
			d.Id(), modifyOpts)
		if _, err := conn.ModifyVpcTenancy(modifyOpts); err != nil {
			return err
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating tags: %s", err)
		}
	}

	return resourceVPCRead(d, meta)
}

func resourceVPCDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	vpcID := d.Id()
	deleteVpcOpts := &ec2.DeleteVpcInput{
		VpcId: &vpcID,
	}
	log.Printf("[INFO] Deleting VPC: %s", d.Id())

	err := resource.Retry(5*time.Minute, func() *resource.RetryError {
		_, err := conn.DeleteVpc(deleteVpcOpts)
		if err == nil {
			return nil
		}

		if tfawserr.ErrMessageContains(err, "InvalidVpcID.NotFound", "") {
			return nil
		}
		if tfawserr.ErrMessageContains(err, "DependencyViolation", "") {
			return resource.RetryableError(err)
		}
		return resource.NonRetryableError(fmt.Errorf("Error deleting VPC: %s", err))
	})
	if tfresource.TimedOut(err) {
		_, err = conn.DeleteVpc(deleteVpcOpts)
		if tfawserr.ErrMessageContains(err, "InvalidVpcID.NotFound", "") {
			return nil
		}
	}

	if err != nil {
		return fmt.Errorf("Error deleting VPC: %s", err)
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

	return nil
}

// VPCStateRefreshFunc returns a resource.StateRefreshFunc that is used to watch
// a VPC.
func VPCStateRefreshFunc(conn *ec2.EC2, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		describeVpcOpts := &ec2.DescribeVpcsInput{
			VpcIds: []*string{aws.String(id)},
		}
		resp, err := conn.DescribeVpcs(describeVpcOpts)
		if err != nil {
			if ec2err, ok := err.(awserr.Error); ok && ec2err.Code() == "InvalidVpcID.NotFound" {
				resp = nil
			} else {
				log.Printf("Error on VPCStateRefresh: %s", err)
				return nil, "", err
			}
		}

		if resp == nil {
			// Sometimes AWS just has consistency issues and doesn't see
			// our instance yet. Return an empty state.
			return nil, "", nil
		}

		vpc := resp.Vpcs[0]
		return vpc, *vpc.State, nil
	}
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

func resourceVPCSetDefaultNetworkACL(conn *ec2.EC2, d *schema.ResourceData) error {
	filter1 := &ec2.Filter{
		Name:   aws.String("default"),
		Values: []*string{aws.String("true")},
	}
	filter2 := &ec2.Filter{
		Name:   aws.String("vpc-id"),
		Values: []*string{aws.String(d.Id())},
	}
	describeNetworkACLOpts := &ec2.DescribeNetworkAclsInput{
		Filters: []*ec2.Filter{filter1, filter2},
	}
	networkAclResp, err := conn.DescribeNetworkAcls(describeNetworkACLOpts)

	if err != nil {
		return err
	}
	if v := networkAclResp.NetworkAcls; len(v) > 0 {
		d.Set("default_network_acl_id", v[0].NetworkAclId)
	}

	return nil
}

func resourceVPCSetDefaultSecurityGroup(conn *ec2.EC2, d *schema.ResourceData) error {
	filter1 := &ec2.Filter{
		Name:   aws.String("group-name"),
		Values: []*string{aws.String("default")},
	}
	filter2 := &ec2.Filter{
		Name:   aws.String("vpc-id"),
		Values: []*string{aws.String(d.Id())},
	}
	describeSgOpts := &ec2.DescribeSecurityGroupsInput{
		Filters: []*ec2.Filter{filter1, filter2},
	}
	securityGroupResp, err := conn.DescribeSecurityGroups(describeSgOpts)

	if err != nil {
		return err
	}
	if v := securityGroupResp.SecurityGroups; len(v) > 0 {
		d.Set("default_security_group_id", v[0].GroupId)
	}

	return nil
}

func resourceVPCSetDefaultRouteTable(conn *ec2.EC2, d *schema.ResourceData) error {
	filter1 := &ec2.Filter{
		Name:   aws.String("association.main"),
		Values: []*string{aws.String("true")},
	}
	filter2 := &ec2.Filter{
		Name:   aws.String("vpc-id"),
		Values: []*string{aws.String(d.Id())},
	}

	findOpts := &ec2.DescribeRouteTablesInput{
		Filters: []*ec2.Filter{filter1, filter2},
	}

	resp, err := conn.DescribeRouteTables(findOpts)
	if err != nil {
		return err
	}

	if len(resp.RouteTables) < 1 || resp.RouteTables[0] == nil {
		return fmt.Errorf("Default Route table not found")
	}

	// There Can Be Only 1 ... Default Route Table
	d.Set("default_route_table_id", resp.RouteTables[0].RouteTableId)

	return nil
}

func resourceVPCSetMainRouteTable(conn *ec2.EC2, vpcid string) (string, error) {
	filter1 := &ec2.Filter{
		Name:   aws.String("association.main"),
		Values: []*string{aws.String("true")},
	}
	filter2 := &ec2.Filter{
		Name:   aws.String("vpc-id"),
		Values: []*string{aws.String(vpcid)},
	}

	findOpts := &ec2.DescribeRouteTablesInput{
		Filters: []*ec2.Filter{filter1, filter2},
	}

	resp, err := conn.DescribeRouteTables(findOpts)
	if err != nil {
		return "", err
	}

	if len(resp.RouteTables) < 1 || resp.RouteTables[0] == nil {
		return "", fmt.Errorf("Main Route table not found")
	}

	// There Can Be Only 1 Main Route Table for a VPC
	return aws.StringValue(resp.RouteTables[0].RouteTableId), nil
}

func resourceVPCInstanceImport(
	d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
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
