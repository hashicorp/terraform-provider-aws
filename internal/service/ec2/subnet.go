package ec2

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceSubnet() *schema.Resource {
	//lintignore:R011
	return &schema.Resource{
		Create: resourceSubnetCreate,
		Read:   resourceSubnetRead,
		Update: resourceSubnetUpdate,
		Delete: resourceSubnetDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		CustomizeDiff: verify.SetTagsDiff,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(20 * time.Minute),
		},

		SchemaVersion: 1,
		MigrateState:  SubnetMigrateState,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"assign_ipv6_address_on_creation": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"availability_zone": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"availability_zone_id"},
			},
			"availability_zone_id": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"availability_zone"},
			},
			"cidr_block": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidIPv4CIDRNetworkAddress,
			},
			"customer_owned_ipv4_pool": {
				Type:         schema.TypeString,
				Optional:     true,
				RequiredWith: []string{"map_customer_owned_ip_on_launch", "outpost_arn"},
			},
			"enable_dns64": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
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
			"ipv6_cidr_block": {
				Type:         schema.TypeString,
				Optional:     true,
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
				Default:  false,
			},
			"outpost_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"owner_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"private_dns_hostname_type_on_launch": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice(ec2.HostnameType_Values(), false),
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
			"vpc_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceSubnetCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := &ec2.CreateSubnetInput{
		TagSpecifications: ec2TagSpecificationsFromKeyValueTags(tags, ec2.ResourceTypeSubnet),
		VpcId:             aws.String(d.Get("vpc_id").(string)),
	}

	if v, ok := d.GetOk("availability_zone"); ok {
		input.AvailabilityZone = aws.String(v.(string))
	}

	if v, ok := d.GetOk("availability_zone_id"); ok {
		input.AvailabilityZoneId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("cidr_block"); ok {
		input.CidrBlock = aws.String(v.(string))
	}

	if v, ok := d.GetOk("ipv6_cidr_block"); ok {
		input.Ipv6CidrBlock = aws.String(v.(string))
	}

	if v, ok := d.GetOk("ipv6_native"); ok {
		input.Ipv6Native = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("outpost_arn"); ok {
		input.OutpostArn = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating EC2 Subnet: %s", input)
	output, err := conn.CreateSubnet(input)

	if err != nil {
		return fmt.Errorf("error creating EC2 Subnet: %w", err)
	}

	d.SetId(aws.StringValue(output.Subnet.SubnetId))

	_, err = WaitSubnetAvailable(conn, d.Id(), d.Timeout(schema.TimeoutCreate))

	if err != nil {
		return fmt.Errorf("error waiting for EC2 Subnet (%s) to become available: %w", d.Id(), err)
	}

	// You cannot modify multiple subnet attributes in the same request,
	// except CustomerOwnedIpv4Pool and MapCustomerOwnedIpOnLaunch.
	// Reference: https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_ModifySubnetAttribute.html

	if d.Get("assign_ipv6_address_on_creation").(bool) {
		input := &ec2.ModifySubnetAttributeInput{
			AssignIpv6AddressOnCreation: &ec2.AttributeBooleanValue{
				Value: aws.Bool(true),
			},
			SubnetId: aws.String(d.Id()),
		}

		if _, err := conn.ModifySubnetAttribute(input); err != nil {
			return fmt.Errorf("error setting EC2 Subnet (%s) AssignIpv6AddressOnCreation: %w", d.Id(), err)
		}
	}

	if v, ok := d.GetOk("customer_owned_ipv4_pool"); ok {
		input := &ec2.ModifySubnetAttributeInput{
			CustomerOwnedIpv4Pool: aws.String(v.(string)),
			MapCustomerOwnedIpOnLaunch: &ec2.AttributeBooleanValue{
				Value: aws.Bool(d.Get("map_customer_owned_ip_on_launch").(bool)),
			},
			SubnetId: aws.String(d.Id()),
		}

		if _, err := conn.ModifySubnetAttribute(input); err != nil {
			return fmt.Errorf("error setting EC2 Subnet (%s) CustomerOwnedIpv4Pool/MapCustomerOwnedIpOnLaunch: %w", d.Id(), err)
		}

		if _, err := WaitSubnetMapCustomerOwnedIPOnLaunchUpdated(conn, d.Id(), d.Get("map_customer_owned_ip_on_launch").(bool)); err != nil {
			return fmt.Errorf("error waiting for EC2 Subnet (%s) MapCustomerOwnedIpOnLaunch update: %w", d.Id(), err)
		}
	}

	if d.Get("enable_dns64").(bool) {
		input := &ec2.ModifySubnetAttributeInput{
			EnableDns64: &ec2.AttributeBooleanValue{
				Value: aws.Bool(true),
			},
			SubnetId: aws.String(d.Id()),
		}

		if _, err := conn.ModifySubnetAttribute(input); err != nil {
			return fmt.Errorf("error setting EC2 Subnet (%s) EnableDns64: %w", d.Id(), err)
		}

		if _, err := WaitSubnetEnableDns64Updated(conn, d.Id(), d.Get("enable_dns64").(bool)); err != nil {
			return fmt.Errorf("error waiting for EC2 Subnet (%s) EnableDns64 update: %w", d.Id(), err)
		}
	}

	if d.Get("enable_resource_name_dns_aaaa_record_on_launch").(bool) {
		input := &ec2.ModifySubnetAttributeInput{
			EnableResourceNameDnsAAAARecordOnLaunch: &ec2.AttributeBooleanValue{
				Value: aws.Bool(true),
			},
			SubnetId: aws.String(d.Id()),
		}

		if _, err := conn.ModifySubnetAttribute(input); err != nil {
			return fmt.Errorf("error setting EC2 Subnet (%s) EnableResourceNameDnsAAAARecordOnLaunch: %w", d.Id(), err)
		}

		if _, err := WaitSubnetEnableResourceNameDnsAAAARecordOnLaunchUpdated(conn, d.Id(), d.Get("enable_resource_name_dns_aaaa_record_on_launch").(bool)); err != nil {
			return fmt.Errorf("error waiting for EC2 Subnet (%s) EnableResourceNameDnsAAAARecordOnLaunch update: %w", d.Id(), err)
		}
	}

	if d.Get("enable_resource_name_dns_a_record_on_launch").(bool) {
		input := &ec2.ModifySubnetAttributeInput{
			EnableResourceNameDnsARecordOnLaunch: &ec2.AttributeBooleanValue{
				Value: aws.Bool(true),
			},
			SubnetId: aws.String(d.Id()),
		}

		if _, err := conn.ModifySubnetAttribute(input); err != nil {
			return fmt.Errorf("error setting EC2 Subnet (%s) EnableResourceNameDnsARecordOnLaunch: %w", d.Id(), err)
		}

		if _, err := WaitSubnetEnableResourceNameDnsARecordOnLaunchUpdated(conn, d.Id(), d.Get("enable_resource_name_dns_a_record_on_launch").(bool)); err != nil {
			return fmt.Errorf("error waiting for EC2 Subnet (%s) EnableResourceNameDnsARecordOnLaunch update: %w", d.Id(), err)
		}
	}

	if d.Get("map_public_ip_on_launch").(bool) {
		input := &ec2.ModifySubnetAttributeInput{
			MapPublicIpOnLaunch: &ec2.AttributeBooleanValue{
				Value: aws.Bool(true),
			},
			SubnetId: aws.String(d.Id()),
		}

		if _, err := conn.ModifySubnetAttribute(input); err != nil {
			return fmt.Errorf("error setting EC2 Subnet (%s) MapPublicIpOnLaunch: %w", d.Id(), err)
		}

		if _, err := WaitSubnetMapPublicIPOnLaunchUpdated(conn, d.Id(), d.Get("map_public_ip_on_launch").(bool)); err != nil {
			return fmt.Errorf("error waiting for EC2 Subnet (%s) MapPublicIpOnLaunch update: %w", d.Id(), err)
		}
	}

	if v, ok := d.GetOk("private_dns_hostname_type_on_launch"); ok {
		input := &ec2.ModifySubnetAttributeInput{
			PrivateDnsHostnameTypeOnLaunch: aws.String(v.(string)),
			SubnetId:                       aws.String(d.Id()),
		}

		if _, err := conn.ModifySubnetAttribute(input); err != nil {
			return fmt.Errorf("error setting EC2 Subnet (%s) PrivateDnsHostnameTypeOnLaunch: %w", d.Id(), err)
		}

		if _, err := WaitSubnetPrivateDNSHostnameTypeOnLaunchUpdated(conn, d.Id(), d.Get("private_dns_hostname_type_on_launch").(string)); err != nil {
			return fmt.Errorf("error waiting for EC2 Subnet (%s) PrivateDnsHostnameTypeOnLaunch update: %w", d.Id(), err)
		}
	}

	return resourceSubnetRead(d, meta)
}

func resourceSubnetRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	outputRaw, err := tfresource.RetryWhenNewResourceNotFound(SubnetPropagationTimeout, func() (interface{}, error) {
		return FindSubnetByID(conn, d.Id())
	}, d.IsNewResource())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Subnet (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Subnet (%s): %w", d.Id(), err)
	}

	subnet := outputRaw.(*ec2.Subnet)

	d.Set("arn", subnet.SubnetArn)
	d.Set("assign_ipv6_address_on_creation", subnet.AssignIpv6AddressOnCreation)
	d.Set("availability_zone", subnet.AvailabilityZone)
	d.Set("availability_zone_id", subnet.AvailabilityZoneId)
	d.Set("cidr_block", subnet.CidrBlock)
	d.Set("customer_owned_ipv4_pool", subnet.CustomerOwnedIpv4Pool)
	d.Set("enable_dns64", subnet.EnableDns64)
	d.Set("ipv6_native", subnet.Ipv6Native)
	d.Set("map_customer_owned_ip_on_launch", subnet.MapCustomerOwnedIpOnLaunch)
	d.Set("map_public_ip_on_launch", subnet.MapPublicIpOnLaunch)
	d.Set("outpost_arn", subnet.OutpostArn)
	d.Set("owner_id", subnet.OwnerId)
	d.Set("vpc_id", subnet.VpcId)

	// Make sure those values are set, if an IPv6 block exists it'll be set in the loop.
	d.Set("ipv6_cidr_block_association_id", nil)
	d.Set("ipv6_cidr_block", nil)

	for _, v := range subnet.Ipv6CidrBlockAssociationSet {
		if aws.StringValue(v.Ipv6CidrBlockState.State) == ec2.SubnetCidrBlockStateCodeAssociated { //we can only ever have 1 IPv6 block associated at once
			d.Set("ipv6_cidr_block_association_id", v.AssociationId)
			d.Set("ipv6_cidr_block", v.Ipv6CidrBlock)
			break
		}
	}

	if subnet.PrivateDnsNameOptionsOnLaunch != nil {
		d.Set("enable_resource_name_dns_aaaa_record_on_launch", subnet.PrivateDnsNameOptionsOnLaunch.EnableResourceNameDnsAAAARecord)
		d.Set("enable_resource_name_dns_a_record_on_launch", subnet.PrivateDnsNameOptionsOnLaunch.EnableResourceNameDnsARecord)
		d.Set("private_dns_hostname_type_on_launch", subnet.PrivateDnsNameOptionsOnLaunch.HostnameType)
	} else {
		d.Set("enable_resource_name_dns_aaaa_record_on_launch", nil)
		d.Set("enable_resource_name_dns_a_record_on_launch", nil)
		d.Set("private_dns_hostname_type_on_launch", nil)
	}

	tags := KeyValueTags(subnet.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceSubnetUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating EC2 Subnet (%s) tags: %w", d.Id(), err)
		}
	}

	// You cannot modify multiple subnet attributes in the same request,
	// except CustomerOwnedIpv4Pool and MapCustomerOwnedIpOnLaunch.
	// Reference: https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_ModifySubnetAttribute.html

	if d.HasChanges("customer_owned_ipv4_pool", "map_customer_owned_ip_on_launch") {
		input := &ec2.ModifySubnetAttributeInput{
			MapCustomerOwnedIpOnLaunch: &ec2.AttributeBooleanValue{
				Value: aws.Bool(d.Get("map_customer_owned_ip_on_launch").(bool)),
			},
			SubnetId: aws.String(d.Id()),
		}

		if v, ok := d.GetOk("customer_owned_ipv4_pool"); ok {
			input.CustomerOwnedIpv4Pool = aws.String(v.(string))
		}

		if _, err := conn.ModifySubnetAttribute(input); err != nil {
			return fmt.Errorf("error setting EC2 Subnet (%s) CustomerOwnedIpv4Pool/MapCustomerOwnedIpOnLaunch: %w", d.Id(), err)
		}

		if _, err := WaitSubnetMapCustomerOwnedIPOnLaunchUpdated(conn, d.Id(), d.Get("map_customer_owned_ip_on_launch").(bool)); err != nil {
			return fmt.Errorf("error waiting for EC2 Subnet (%s) MapCustomerOwnedIpOnLaunch update: %w", d.Id(), err)
		}
	}

	if d.HasChange("enable_dns64") {
		input := &ec2.ModifySubnetAttributeInput{
			EnableDns64: &ec2.AttributeBooleanValue{
				Value: aws.Bool(d.Get("enable_dns64").(bool)),
			},
			SubnetId: aws.String(d.Id()),
		}

		if _, err := conn.ModifySubnetAttribute(input); err != nil {
			return fmt.Errorf("error setting EC2 Subnet (%s) EnableDns64: %w", d.Id(), err)
		}

		if _, err := WaitSubnetEnableDns64Updated(conn, d.Id(), d.Get("enable_dns64").(bool)); err != nil {
			return fmt.Errorf("error waiting for EC2 Subnet (%s) EnableDns64 update: %w", d.Id(), err)
		}
	}

	if d.HasChange("enable_resource_name_dns_aaaa_record_on_launch") {
		input := &ec2.ModifySubnetAttributeInput{
			EnableResourceNameDnsAAAARecordOnLaunch: &ec2.AttributeBooleanValue{
				Value: aws.Bool(d.Get("enable_resource_name_dns_aaaa_record_on_launch").(bool)),
			},
			SubnetId: aws.String(d.Id()),
		}

		if _, err := conn.ModifySubnetAttribute(input); err != nil {
			return fmt.Errorf("error setting EC2 Subnet (%s) EnableResourceNameDnsAAAARecordOnLaunch: %w", d.Id(), err)
		}

		if _, err := WaitSubnetEnableResourceNameDnsAAAARecordOnLaunchUpdated(conn, d.Id(), d.Get("enable_resource_name_dns_aaaa_record_on_launch").(bool)); err != nil {
			return fmt.Errorf("error waiting for EC2 Subnet (%s) EnableResourceNameDnsAAAARecordOnLaunch update: %w", d.Id(), err)
		}
	}

	if d.HasChange("enable_resource_name_dns_a_record_on_launch") {
		input := &ec2.ModifySubnetAttributeInput{
			EnableResourceNameDnsARecordOnLaunch: &ec2.AttributeBooleanValue{
				Value: aws.Bool(d.Get("enable_resource_name_dns_a_record_on_launch").(bool)),
			},
			SubnetId: aws.String(d.Id()),
		}

		if _, err := conn.ModifySubnetAttribute(input); err != nil {
			return fmt.Errorf("error setting EC2 Subnet (%s) EnableResourceNameDnsARecordOnLaunch: %w", d.Id(), err)
		}

		if _, err := WaitSubnetEnableResourceNameDnsARecordOnLaunchUpdated(conn, d.Id(), d.Get("enable_resource_name_dns_a_record_on_launch").(bool)); err != nil {
			return fmt.Errorf("error waiting for EC2 Subnet (%s) EnableResourceNameDnsARecordOnLaunch update: %w", d.Id(), err)
		}
	}

	if d.HasChange("map_public_ip_on_launch") {
		input := &ec2.ModifySubnetAttributeInput{
			MapPublicIpOnLaunch: &ec2.AttributeBooleanValue{
				Value: aws.Bool(d.Get("map_public_ip_on_launch").(bool)),
			},
			SubnetId: aws.String(d.Id()),
		}

		if _, err := conn.ModifySubnetAttribute(input); err != nil {
			return fmt.Errorf("error setting EC2 Subnet (%s) MapPublicIpOnLaunch: %w", d.Id(), err)
		}

		if _, err := WaitSubnetMapPublicIPOnLaunchUpdated(conn, d.Id(), d.Get("map_public_ip_on_launch").(bool)); err != nil {
			return fmt.Errorf("error waiting for EC2 Subnet (%s) MapPublicIpOnLaunch update: %w", d.Id(), err)
		}
	}

	if d.HasChange("private_dns_hostname_type_on_launch") {
		input := &ec2.ModifySubnetAttributeInput{
			PrivateDnsHostnameTypeOnLaunch: aws.String(d.Get("private_dns_hostname_type_on_launch").(string)),
			SubnetId:                       aws.String(d.Id()),
		}

		if _, err := conn.ModifySubnetAttribute(input); err != nil {
			return fmt.Errorf("error setting EC2 Subnet (%s) PrivateDnsHostnameTypeOnLaunch: %w", d.Id(), err)
		}

		if _, err := WaitSubnetPrivateDNSHostnameTypeOnLaunchUpdated(conn, d.Id(), d.Get("private_dns_hostname_type_on_launch").(string)); err != nil {
			return fmt.Errorf("error waiting for EC2 Subnet (%s) PrivateDnsHostnameTypeOnLaunch update: %w", d.Id(), err)
		}
	}

	if d.HasChange("ipv6_cidr_block") {
		// We need to handle that we disassociate the IPv6 CIDR block before we try to associate the new one
		// This could be an issue as, we could error out when we try to add the new one
		// We may need to roll back the state and reattach the old one if this is the case
		if v, ok := d.GetOk("ipv6_cidr_block_association_id"); ok {
			if !d.Get("assign_ipv6_address_on_creation").(bool) {
				input := &ec2.ModifySubnetAttributeInput{
					AssignIpv6AddressOnCreation: &ec2.AttributeBooleanValue{
						Value: aws.Bool(false),
					},
					SubnetId: aws.String(d.Id()),
				}

				if _, err := conn.ModifySubnetAttribute(input); err != nil {
					return fmt.Errorf("error setting EC2 Subnet (%s) AssignIpv6AddressOnCreation: %w", d.Id(), err)
				}
			}

			associationID := v.(string)

			//Firstly we have to disassociate the old IPv6 CIDR Block
			input := &ec2.DisassociateSubnetCidrBlockInput{
				AssociationId: aws.String(associationID),
			}

			_, err := conn.DisassociateSubnetCidrBlock(input)

			if err != nil {
				return fmt.Errorf("error disassociating EC2 Subnet (%s) CIDR block (%s): %w", d.Id(), associationID, err)
			}

			_, err = WaitSubnetIPv6CIDRBlockAssociationDeleted(conn, associationID)

			if err != nil {
				return fmt.Errorf("error waiting for EC2 Subnet (%s) CIDR block (%s) to become disassociated: %w", d.Id(), associationID, err)
			}
		}

		if newIpv6 := d.Get("ipv6_cidr_block").(string); newIpv6 != "" {
			//Now we need to try to associate the new CIDR block
			input := &ec2.AssociateSubnetCidrBlockInput{
				Ipv6CidrBlock: aws.String(newIpv6),
				SubnetId:      aws.String(d.Id()),
			}

			output, err := conn.AssociateSubnetCidrBlock(input)

			if err != nil {
				//The big question here is, do we want to try to reassociate the old one??
				//If we have a failure here, then we may be in a situation that we have nothing associated
				return fmt.Errorf("error associating EC2 Subnet (%s) CIDR block (%s): %w", d.Id(), newIpv6, err)
			}

			associationID := aws.StringValue(output.Ipv6CidrBlockAssociation.AssociationId)

			_, err = WaitSubnetIPv6CIDRBlockAssociationCreated(conn, associationID)

			if err != nil {
				return fmt.Errorf("error waiting for EC2 Subnet (%s) CIDR block (%s) to become associated: %w", d.Id(), associationID, err)
			}
		}
	}

	if d.HasChange("assign_ipv6_address_on_creation") {
		input := &ec2.ModifySubnetAttributeInput{
			AssignIpv6AddressOnCreation: &ec2.AttributeBooleanValue{
				Value: aws.Bool(d.Get("assign_ipv6_address_on_creation").(bool)),
			},
			SubnetId: aws.String(d.Id()),
		}

		if _, err := conn.ModifySubnetAttribute(input); err != nil {
			return fmt.Errorf("error enabling EC2 Subnet (%s) AssignIpv6AddressOnCreation: %w", d.Id(), err)
		}
	}

	return resourceSubnetRead(d, meta)
}

func resourceSubnetDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	log.Printf("[INFO] Deleting EC2 Subnet: %s", d.Id())

	if err := deleteLingeringLambdaENIs(conn, "subnet-id", d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return fmt.Errorf("error deleting Lambda ENIs for EC2 Subnet (%s): %w", d.Id(), err)
	}

	_, err := tfresource.RetryWhenAWSErrCodeEquals(d.Timeout(schema.TimeoutDelete), func() (interface{}, error) {
		return conn.DeleteSubnet(&ec2.DeleteSubnetInput{
			SubnetId: aws.String(d.Id()),
		})
	}, ErrCodeDependencyViolation)

	if tfawserr.ErrCodeEquals(err, ErrCodeInvalidSubnetIDNotFound) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting EC2 Subnet (%s): %w", d.Id(), err)
	}

	return nil
}
