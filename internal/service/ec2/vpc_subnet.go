// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	fdiag "github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/list"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
	"go.opentelemetry.io/contrib/instrumentation/github.com/aws/aws-sdk-go-v2/otelaws"
	"go.opentelemetry.io/otel/attribute"
)

// @SDKResource("aws_subnet", name="Subnet")
// @Tags(identifierAttribute="id")
// @IdentityAttribute("id")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/ec2/types;awstypes;awstypes.Subnet")
// @Testing(generator=false)
// @Testing(preIdentityVersion="v6.8.0")
func resourceSubnet() *schema.Resource {
	//lintignore:R011
	return &schema.Resource{
		CreateWithoutTimeout: resourceSubnetCreate,
		ReadWithoutTimeout:   resourceSubnetRead,
		UpdateWithoutTimeout: resourceSubnetUpdate,
		DeleteWithoutTimeout: resourceSubnetDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(20 * time.Minute),
		},

		SchemaVersion: 1,
		MigrateState:  subnetMigrateState,

		// Keep in sync with aws_default_subnet's schema.
		// See notes in vpc_default_subnet.go.
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
				ConflictsWith: []string{names.AttrAvailabilityZone},
			},
			names.AttrCIDRBlock: {
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
			"enable_lni_at_device_index": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.NoZeroValues,
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
				Required: true,
				ForceNew: true,
			},
		},
	}
}

// @SDKListResource("aws_subnet")
func subnetResourceAsListResource() inttypes.ListResourceForSDK {
	l := subnetListResource{}
	l.SetResourceSchema(resourceSubnet())

	return &l
}

func resourceSubnetCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	input := &ec2.CreateSubnetInput{
		TagSpecifications: getTagSpecificationsIn(ctx, awstypes.ResourceTypeSubnet),
		VpcId:             aws.String(d.Get(names.AttrVPCID).(string)),
	}

	if v, ok := d.GetOk(names.AttrAvailabilityZone); ok {
		input.AvailabilityZone = aws.String(v.(string))
	}

	if v, ok := d.GetOk("availability_zone_id"); ok {
		input.AvailabilityZoneId = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrCIDRBlock); ok {
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

	output, err := conn.CreateSubnet(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EC2 Subnet: %s", err)
	}

	d.SetId(aws.ToString(output.Subnet.SubnetId))

	subnet, err := waitSubnetAvailable(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate))

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EC2 Subnet (%s) create: %s", d.Id(), err)
	}

	for i, v := range subnet.Ipv6CidrBlockAssociationSet {
		if v.Ipv6CidrBlockState.State == awstypes.SubnetCidrBlockStateCodeAssociating { //we can only ever have 1 IPv6 block associated at once
			associationID := aws.ToString(v.AssociationId)

			subnetCidrBlockState, err := waitSubnetIPv6CIDRBlockAssociationCreated(ctx, conn, associationID)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "waiting for EC2 Subnet (%s) IPv6 CIDR block (%s) to become associated: %s", d.Id(), associationID, err)
			}

			subnet.Ipv6CidrBlockAssociationSet[i].Ipv6CidrBlockState = subnetCidrBlockState
		}
	}

	if err := modifySubnetAttributesOnCreate(ctx, conn, d, subnet, false); err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	return append(diags, resourceSubnetRead(ctx, d, meta)...)
}

func resourceSubnetRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	subnet, err := tfresource.RetryWhenNewResourceNotFound(ctx, ec2PropagationTimeout, func(ctx context.Context) (*awstypes.Subnet, error) {
		return findSubnetByID(ctx, conn, d.Id())
	}, d.IsNewResource())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 Subnet (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 Subnet (%s): %s", d.Id(), err)
	}

	resourceSubnetFlatten(ctx, subnet, d)

	return diags
}

func resourceSubnetUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	// You cannot modify multiple subnet attributes in the same request,
	// except CustomerOwnedIpv4Pool and MapCustomerOwnedIpOnLaunch.
	// Reference: https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_ModifySubnetAttribute.html

	if d.HasChanges("customer_owned_ipv4_pool", "map_customer_owned_ip_on_launch") {
		if err := modifySubnetOutpostRackAttributes(ctx, conn, d.Id(), d.Get("customer_owned_ipv4_pool").(string), d.Get("map_customer_owned_ip_on_launch").(bool)); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	// If we're disabling IPv6 assignment for new ENIs, do that before modifying the IPv6 CIDR block.
	if d.HasChange("assign_ipv6_address_on_creation") && !d.Get("assign_ipv6_address_on_creation").(bool) {
		if err := modifySubnetAssignIPv6AddressOnCreation(ctx, conn, d.Id(), false); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	// If we're enabling dns64 and resource_name_dns_aaaa_record_on_launch, do that after modifying the IPv6 CIDR block.
	if d.HasChange("ipv6_cidr_block") {
		if err := modifySubnetIPv6CIDRBlockAssociation(ctx, conn, d.Id(), d.Get("ipv6_cidr_block_association_id").(string), d.Get("ipv6_cidr_block").(string)); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	if d.HasChange("enable_dns64") {
		if err := modifySubnetEnableDNS64(ctx, conn, d.Id(), d.Get("enable_dns64").(bool)); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	if d.HasChange("enable_lni_at_device_index") {
		if err := modifySubnetEnableLniAtDeviceIndex(ctx, conn, d.Id(), int32(d.Get("enable_lni_at_device_index").(int))); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	if d.HasChange("enable_resource_name_dns_aaaa_record_on_launch") {
		if err := modifySubnetEnableResourceNameDNSAAAARecordOnLaunch(ctx, conn, d.Id(), d.Get("enable_resource_name_dns_aaaa_record_on_launch").(bool)); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	if d.HasChange("enable_resource_name_dns_a_record_on_launch") {
		if err := modifySubnetEnableResourceNameDNSARecordOnLaunch(ctx, conn, d.Id(), d.Get("enable_resource_name_dns_a_record_on_launch").(bool)); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	if d.HasChange("map_public_ip_on_launch") {
		if err := modifySubnetMapPublicIPOnLaunch(ctx, conn, d.Id(), d.Get("map_public_ip_on_launch").(bool)); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	if d.HasChange("private_dns_hostname_type_on_launch") {
		if err := modifySubnetPrivateDNSHostnameTypeOnLaunch(ctx, conn, d.Id(), d.Get("private_dns_hostname_type_on_launch").(string)); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	// If we're enabling IPv6 assignment for new ENIs, do that after modifying the IPv6 CIDR block.
	if d.HasChange("assign_ipv6_address_on_creation") && d.Get("assign_ipv6_address_on_creation").(bool) {
		if err := modifySubnetAssignIPv6AddressOnCreation(ctx, conn, d.Id(), true); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	return append(diags, resourceSubnetRead(ctx, d, meta)...)
}

func resourceSubnetDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	ctx = tflog.SetField(ctx, logging.KeyResourceId, d.Id())

	tflog.Info(ctx, "Deleting EC2 Subnet")

	if err := deleteLingeringENIs(ctx, meta.(*conns.AWSClient).EC2Client(ctx), "subnet-id", d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting ENIs for EC2 Subnet (%s): %s", d.Id(), err)
	}

	_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, d.Timeout(schema.TimeoutDelete), func(ctx context.Context) (any, error) {
		return conn.DeleteSubnet(ctx, &ec2.DeleteSubnetInput{
			SubnetId: aws.String(d.Id()),
		})
	}, errCodeDependencyViolation)

	if tfawserr.ErrCodeEquals(err, errCodeInvalidSubnetIDNotFound) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting EC2 Subnet (%s): %s", d.Id(), err)
	}

	return diags
}

// modifySubnetAttributesOnCreate sets subnet attributes on resource Create.
// Called after new subnet creation or existing default subnet adoption.
func modifySubnetAttributesOnCreate(ctx context.Context, conn *ec2.Client, d *schema.ResourceData, subnet *awstypes.Subnet, computedIPv6CidrBlock bool) error {
	// If we're disabling IPv6 assignment for new ENIs, do that before modifying the IPv6 CIDR block.
	if new, old := d.Get("assign_ipv6_address_on_creation").(bool), aws.ToBool(subnet.AssignIpv6AddressOnCreation); old != new && !new {
		if err := modifySubnetAssignIPv6AddressOnCreation(ctx, conn, d.Id(), false); err != nil {
			return err
		}
	}

	// If we're disabling DNS64, do that before modifying the IPv6 CIDR block.
	if new, old := d.Get("enable_dns64").(bool), aws.ToBool(subnet.EnableDns64); old != new && !new {
		if err := modifySubnetEnableDNS64(ctx, conn, d.Id(), false); err != nil {
			return err
		}
	}

	// Creating a new IPv6-native default subnet assigns a computed IPv6 CIDR block.
	// Don't attempt to do anything with it.
	if !computedIPv6CidrBlock {
		var oldAssociationID, oldIPv6CIDRBlock string
		for _, v := range subnet.Ipv6CidrBlockAssociationSet {
			if v.Ipv6CidrBlockState.State == awstypes.SubnetCidrBlockStateCodeAssociated { //we can only ever have 1 IPv6 block associated at once
				oldAssociationID = aws.ToString(v.AssociationId)
				oldIPv6CIDRBlock = aws.ToString(v.Ipv6CidrBlock)

				break
			}
		}
		if new := d.Get("ipv6_cidr_block").(string); oldIPv6CIDRBlock != new {
			if err := modifySubnetIPv6CIDRBlockAssociation(ctx, conn, d.Id(), oldAssociationID, new); err != nil {
				return err
			}
		}
	}

	// If we're enabling IPv6 assignment for new ENIs, do that after modifying the IPv6 CIDR block.
	if new, old := d.Get("assign_ipv6_address_on_creation").(bool), aws.ToBool(subnet.AssignIpv6AddressOnCreation); old != new && new {
		if err := modifySubnetAssignIPv6AddressOnCreation(ctx, conn, d.Id(), true); err != nil {
			return err
		}
	}

	if newCustomerOwnedIPOnLaunch, oldCustomerOwnedIPOnLaunch, newMapCustomerOwnedIPOnLaunch, oldMapCustomerOwnedIPOnLaunch :=
		d.Get("customer_owned_ipv4_pool").(string), aws.ToString(subnet.CustomerOwnedIpv4Pool), d.Get("map_customer_owned_ip_on_launch").(bool), aws.ToBool(subnet.MapCustomerOwnedIpOnLaunch); oldCustomerOwnedIPOnLaunch != newCustomerOwnedIPOnLaunch || oldMapCustomerOwnedIPOnLaunch != newMapCustomerOwnedIPOnLaunch {
		if err := modifySubnetOutpostRackAttributes(ctx, conn, d.Id(), newCustomerOwnedIPOnLaunch, newMapCustomerOwnedIPOnLaunch); err != nil {
			return err
		}
	}

	// If we're enabling DNS64, do that after modifying the IPv6 CIDR block.
	if new, old := d.Get("enable_dns64").(bool), aws.ToBool(subnet.EnableDns64); old != new && new {
		if err := modifySubnetEnableDNS64(ctx, conn, d.Id(), true); err != nil {
			return err
		}
	}

	if new, old := int32(d.Get("enable_lni_at_device_index").(int)), aws.ToInt32(subnet.EnableLniAtDeviceIndex); old != new && new != 0 {
		if err := modifySubnetEnableLniAtDeviceIndex(ctx, conn, d.Id(), new); err != nil {
			return err
		}
	}

	if subnet.PrivateDnsNameOptionsOnLaunch != nil {
		if new, old := d.Get("enable_resource_name_dns_aaaa_record_on_launch").(bool), aws.ToBool(subnet.PrivateDnsNameOptionsOnLaunch.EnableResourceNameDnsAAAARecord); old != new {
			if err := modifySubnetEnableResourceNameDNSAAAARecordOnLaunch(ctx, conn, d.Id(), new); err != nil {
				return err
			}
		}

		if new, old := d.Get("enable_resource_name_dns_a_record_on_launch").(bool), aws.ToBool(subnet.PrivateDnsNameOptionsOnLaunch.EnableResourceNameDnsARecord); old != new {
			if err := modifySubnetEnableResourceNameDNSARecordOnLaunch(ctx, conn, d.Id(), new); err != nil {
				return err
			}
		}

		// private_dns_hostname_type_on_launch is Computed, so only modify if the new value is set.
		if new, old := d.Get("private_dns_hostname_type_on_launch").(string), string(subnet.PrivateDnsNameOptionsOnLaunch.HostnameType); old != new && new != "" {
			if err := modifySubnetPrivateDNSHostnameTypeOnLaunch(ctx, conn, d.Id(), new); err != nil {
				return err
			}
		}
	}

	if new, old := d.Get("map_public_ip_on_launch").(bool), aws.ToBool(subnet.MapPublicIpOnLaunch); old != new {
		if err := modifySubnetMapPublicIPOnLaunch(ctx, conn, d.Id(), new); err != nil {
			return err
		}
	}

	return nil
}

func modifySubnetAssignIPv6AddressOnCreation(ctx context.Context, conn *ec2.Client, subnetID string, v bool) error {
	input := &ec2.ModifySubnetAttributeInput{
		AssignIpv6AddressOnCreation: &awstypes.AttributeBooleanValue{
			Value: aws.Bool(v),
		},
		SubnetId: aws.String(subnetID),
	}

	if _, err := conn.ModifySubnetAttribute(ctx, input); err != nil {
		return fmt.Errorf("setting EC2 Subnet (%s) AssignIpv6AddressOnCreation: %w", subnetID, err)
	}

	if _, err := waitSubnetAssignIPv6AddressOnCreationUpdated(ctx, conn, subnetID, v); err != nil {
		return fmt.Errorf("waiting for EC2 Subnet (%s) AssignIpv6AddressOnCreation update: %w", subnetID, err)
	}

	return nil
}

func modifySubnetEnableDNS64(ctx context.Context, conn *ec2.Client, subnetID string, v bool) error {
	input := &ec2.ModifySubnetAttributeInput{
		EnableDns64: &awstypes.AttributeBooleanValue{
			Value: aws.Bool(v),
		},
		SubnetId: aws.String(subnetID),
	}

	if _, err := conn.ModifySubnetAttribute(ctx, input); err != nil {
		return fmt.Errorf("modifying EC2 Subnet (%s) EnableDns64: %w", subnetID, err)
	}

	if _, err := waitSubnetEnableDNS64Updated(ctx, conn, subnetID, v); err != nil {
		return fmt.Errorf("waiting for EC2 Subnet (%s) EnableDns64 update: %w", subnetID, err)
	}

	return nil
}

func modifySubnetEnableLniAtDeviceIndex(ctx context.Context, conn *ec2.Client, subnetID string, deviceIndex int32) error {
	input := &ec2.ModifySubnetAttributeInput{
		EnableLniAtDeviceIndex: aws.Int32(deviceIndex),
		SubnetId:               aws.String(subnetID),
	}

	if _, err := conn.ModifySubnetAttribute(ctx, input); err != nil {
		return fmt.Errorf("modifying EC2 Subnet (%s) EnableLniAtDeviceIndex: %w", subnetID, err)
	}

	if _, err := waitSubnetEnableLniAtDeviceIndexUpdated(ctx, conn, subnetID, deviceIndex); err != nil {
		return fmt.Errorf("waiting for EC2 Subnet (%s) EnableLniAtDeviceIndex update: %w", subnetID, err)
	}

	return nil
}

func modifySubnetEnableResourceNameDNSAAAARecordOnLaunch(ctx context.Context, conn *ec2.Client, subnetID string, v bool) error {
	input := &ec2.ModifySubnetAttributeInput{
		EnableResourceNameDnsAAAARecordOnLaunch: &awstypes.AttributeBooleanValue{
			Value: aws.Bool(v),
		},
		SubnetId: aws.String(subnetID),
	}

	if _, err := conn.ModifySubnetAttribute(ctx, input); err != nil {
		return fmt.Errorf("modifying EC2 Subnet (%s) EnableResourceNameDnsAAAARecordOnLaunch: %w", subnetID, err)
	}

	if _, err := waitSubnetEnableResourceNameDNSAAAARecordOnLaunchUpdated(ctx, conn, subnetID, v); err != nil {
		return fmt.Errorf("waiting for EC2 Subnet (%s) EnableResourceNameDnsAAAARecordOnLaunch update: %w", subnetID, err)
	}

	return nil
}

func modifySubnetEnableResourceNameDNSARecordOnLaunch(ctx context.Context, conn *ec2.Client, subnetID string, v bool) error {
	input := &ec2.ModifySubnetAttributeInput{
		EnableResourceNameDnsARecordOnLaunch: &awstypes.AttributeBooleanValue{
			Value: aws.Bool(v),
		},
		SubnetId: aws.String(subnetID),
	}

	if _, err := conn.ModifySubnetAttribute(ctx, input); err != nil {
		return fmt.Errorf("modifying EC2 Subnet (%s) EnableResourceNameDnsARecordOnLaunch: %w", subnetID, err)
	}

	if _, err := waitSubnetEnableResourceNameDNSARecordOnLaunchUpdated(ctx, conn, subnetID, v); err != nil {
		return fmt.Errorf("waiting for EC2 Subnet (%s) EnableResourceNameDnsARecordOnLaunch update: %w", subnetID, err)
	}

	return nil
}

func modifySubnetIPv6CIDRBlockAssociation(ctx context.Context, conn *ec2.Client, subnetID, associationID, cidrBlock string) error {
	// We need to handle that we disassociate the IPv6 CIDR block before we try to associate the new one
	// This could be an issue as, we could error out when we try to add the new one
	// We may need to roll back the state and reattach the old one if this is the case
	if associationID != "" {
		input := &ec2.DisassociateSubnetCidrBlockInput{
			AssociationId: aws.String(associationID),
		}

		_, err := conn.DisassociateSubnetCidrBlock(ctx, input)

		if err != nil {
			return fmt.Errorf("disassociating EC2 Subnet (%s) IPv6 CIDR block (%s): %w", subnetID, associationID, err)
		}

		_, err = waitSubnetIPv6CIDRBlockAssociationDeleted(ctx, conn, associationID)

		if err != nil {
			return fmt.Errorf("waiting for EC2 Subnet (%s) IPv6 CIDR block (%s) to become disassociated: %w", subnetID, associationID, err)
		}
	}

	if cidrBlock != "" {
		input := &ec2.AssociateSubnetCidrBlockInput{
			Ipv6CidrBlock: aws.String(cidrBlock),
			SubnetId:      aws.String(subnetID),
		}

		output, err := conn.AssociateSubnetCidrBlock(ctx, input)

		if err != nil {
			//The big question here is, do we want to try to reassociate the old one??
			//If we have a failure here, then we may be in a situation that we have nothing associated
			return fmt.Errorf("associating EC2 Subnet (%s) IPv6 CIDR block (%s): %w", subnetID, cidrBlock, err)
		}

		associationID := aws.ToString(output.Ipv6CidrBlockAssociation.AssociationId)

		_, err = waitSubnetIPv6CIDRBlockAssociationCreated(ctx, conn, associationID)

		if err != nil {
			return fmt.Errorf("waiting for EC2 Subnet (%s) IPv6 CIDR block (%s) to become associated: %w", subnetID, associationID, err)
		}
	}

	return nil
}

func modifySubnetMapPublicIPOnLaunch(ctx context.Context, conn *ec2.Client, subnetID string, v bool) error {
	input := &ec2.ModifySubnetAttributeInput{
		MapPublicIpOnLaunch: &awstypes.AttributeBooleanValue{
			Value: aws.Bool(v),
		},
		SubnetId: aws.String(subnetID),
	}

	if _, err := conn.ModifySubnetAttribute(ctx, input); err != nil {
		return fmt.Errorf("modifying EC2 Subnet (%s) MapPublicIpOnLaunch: %w", subnetID, err)
	}

	if _, err := waitSubnetMapPublicIPOnLaunchUpdated(ctx, conn, subnetID, v); err != nil {
		return fmt.Errorf("waiting for EC2 Subnet (%s) MapPublicIpOnLaunch update: %w", subnetID, err)
	}

	return nil
}

func modifySubnetOutpostRackAttributes(ctx context.Context, conn *ec2.Client, subnetID string, customerOwnedIPv4Pool string, mapCustomerOwnedIPOnLaunch bool) error {
	input := &ec2.ModifySubnetAttributeInput{
		MapCustomerOwnedIpOnLaunch: &awstypes.AttributeBooleanValue{
			Value: aws.Bool(mapCustomerOwnedIPOnLaunch),
		},
		SubnetId: aws.String(subnetID),
	}

	if customerOwnedIPv4Pool != "" {
		input.CustomerOwnedIpv4Pool = aws.String(customerOwnedIPv4Pool)
	}

	if _, err := conn.ModifySubnetAttribute(ctx, input); err != nil {
		return fmt.Errorf("modifying EC2 Subnet (%s) CustomerOwnedIpv4Pool/MapCustomerOwnedIpOnLaunch: %w", subnetID, err)
	}

	if _, err := waitSubnetMapCustomerOwnedIPOnLaunchUpdated(ctx, conn, subnetID, mapCustomerOwnedIPOnLaunch); err != nil {
		return fmt.Errorf("waiting for EC2 Subnet (%s) MapCustomerOwnedIpOnLaunch update: %w", subnetID, err)
	}

	return nil
}

func modifySubnetPrivateDNSHostnameTypeOnLaunch(ctx context.Context, conn *ec2.Client, subnetID string, v string) error {
	input := &ec2.ModifySubnetAttributeInput{
		PrivateDnsHostnameTypeOnLaunch: awstypes.HostnameType(v),
		SubnetId:                       aws.String(subnetID),
	}

	if _, err := conn.ModifySubnetAttribute(ctx, input); err != nil {
		return fmt.Errorf("modifying EC2 Subnet (%s) PrivateDnsHostnameTypeOnLaunch: %w", subnetID, err)
	}

	if _, err := waitSubnetPrivateDNSHostnameTypeOnLaunchUpdated(ctx, conn, subnetID, v); err != nil {
		return fmt.Errorf("waiting for EC2 Subnet (%s) PrivateDnsHostnameTypeOnLaunch update: %w", subnetID, err)
	}

	return nil
}

func resourceSubnetFlatten(ctx context.Context, subnet *awstypes.Subnet, rd *schema.ResourceData) {
	rd.Set(names.AttrARN, subnet.SubnetArn)
	rd.Set("assign_ipv6_address_on_creation", subnet.AssignIpv6AddressOnCreation)
	rd.Set(names.AttrAvailabilityZone, subnet.AvailabilityZone)
	rd.Set("availability_zone_id", subnet.AvailabilityZoneId)
	rd.Set(names.AttrCIDRBlock, subnet.CidrBlock)
	rd.Set("customer_owned_ipv4_pool", subnet.CustomerOwnedIpv4Pool)
	rd.Set("enable_dns64", subnet.EnableDns64)
	rd.Set("enable_lni_at_device_index", subnet.EnableLniAtDeviceIndex)
	rd.Set("ipv6_native", subnet.Ipv6Native)
	rd.Set("map_customer_owned_ip_on_launch", subnet.MapCustomerOwnedIpOnLaunch)
	rd.Set("map_public_ip_on_launch", subnet.MapPublicIpOnLaunch)
	rd.Set("outpost_arn", subnet.OutpostArn)
	rd.Set(names.AttrOwnerID, subnet.OwnerId)
	rd.Set(names.AttrVPCID, subnet.VpcId)

	// Make sure those values are set, if an IPv6 block exists it'll be set in the loop.
	rd.Set("ipv6_cidr_block_association_id", nil)
	rd.Set("ipv6_cidr_block", nil)

	for _, v := range subnet.Ipv6CidrBlockAssociationSet {
		if v.Ipv6CidrBlockState.State == awstypes.SubnetCidrBlockStateCodeAssociated { //we can only ever have 1 IPv6 block associated at once
			rd.Set("ipv6_cidr_block_association_id", v.AssociationId)
			rd.Set("ipv6_cidr_block", v.Ipv6CidrBlock)
			break
		}
	}

	if subnet.PrivateDnsNameOptionsOnLaunch != nil {
		rd.Set("enable_resource_name_dns_aaaa_record_on_launch", subnet.PrivateDnsNameOptionsOnLaunch.EnableResourceNameDnsAAAARecord)
		rd.Set("enable_resource_name_dns_a_record_on_launch", subnet.PrivateDnsNameOptionsOnLaunch.EnableResourceNameDnsARecord)
		rd.Set("private_dns_hostname_type_on_launch", subnet.PrivateDnsNameOptionsOnLaunch.HostnameType)
	} else {
		rd.Set("enable_resource_name_dns_aaaa_record_on_launch", nil)
		rd.Set("enable_resource_name_dns_a_record_on_launch", nil)
		rd.Set("private_dns_hostname_type_on_launch", nil)
	}

	setTagsOut(ctx, subnet.Tags)
}

var _ list.ListResourceWithRawV5Schemas = &subnetListResource{}

type subnetListResource struct {
	framework.ResourceWithConfigure
	framework.ListResourceWithSDKv2Resource
	framework.ListResourceWithSDKv2Tags
}

type subnetListResourceModel struct {
	framework.WithRegionModel
	SubnetIDs fwtypes.ListValueOf[types.String] `tfsdk:"subnet_ids"`
	Filters   customListFilters                 `tfsdk:"filter"`
}

func (l *subnetListResource) ListResourceConfigSchema(ctx context.Context, request list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = listschema.Schema{
		Attributes: map[string]listschema.Attribute{
			names.AttrSubnetIDs: listschema.ListAttribute{
				CustomType:  fwtypes.ListOfStringType,
				ElementType: types.StringType,
				Optional:    true,
			},
		},
		Blocks: map[string]listschema.Block{
			names.AttrFilter: listschema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[customListFilterModel](ctx),
				NestedObject: listschema.NestedBlockObject{
					Attributes: map[string]listschema.Attribute{
						names.AttrName: listschema.StringAttribute{
							Required: true,
							Validators: []validator.String{
								notDefaultForAZValidator{},
							},
						},
						names.AttrValues: listschema.ListAttribute{
							CustomType:  fwtypes.ListOfStringType,
							ElementType: types.StringType,
							Required:    true,
						},
					},
				},
			},
		},
	}
}

var _ validator.String = notDefaultForAZValidator{}

type notDefaultForAZValidator struct{}

func (v notDefaultForAZValidator) Description(ctx context.Context) string {
	return v.MarkdownDescription(ctx)
}

func (v notDefaultForAZValidator) MarkdownDescription(_ context.Context) string {
	return ""
}

func (v notDefaultForAZValidator) ValidateString(ctx context.Context, request validator.StringRequest, response *validator.StringResponse) {
	if request.ConfigValue.IsNull() || request.ConfigValue.IsUnknown() {
		return
	}

	value := request.ConfigValue

	if value.ValueString() == "default-for-az" {
		response.Diagnostics.Append(fdiag.NewAttributeErrorDiagnostic(
			request.Path,
			"Invalid Attribute Value",
			`The filter "default-for-az" is not supported. To list default Subnets, use the resource type "aws_default_subnet".`,
		))
	}
}

func (l *subnetListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	awsClient := l.Meta()
	conn := awsClient.EC2Client(ctx)

	attributes := []attribute.KeyValue{
		otelaws.RegionAttr(awsClient.Region(ctx)),
	}
	for _, attribute := range attributes {
		ctx = tflog.SetField(ctx, string(attribute.Key), attribute.Value.AsInterface())
	}

	var query subnetListResourceModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	var input ec2.DescribeSubnetsInput
	if diags := fwflex.Expand(ctx, query, &input); diags.HasError() {
		stream.Results = list.ListResultsStreamDiagnostics(diags)
		return
	}

	input.Filters = append(input.Filters, awstypes.Filter{
		Name:   aws.String("default-for-az"),
		Values: []string{"false"},
	})

	tflog.Info(ctx, "Listing resources")

	stream.Results = func(yield func(list.ListResult) bool) {
		pages := ec2.NewDescribeSubnetsPaginator(conn, &input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			for _, subnet := range page.Subnets {
				ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrID), aws.ToString(subnet.SubnetId))

				result := request.NewListResult(ctx)

				tags := keyValueTags(ctx, subnet.Tags)

				rd := l.ResourceData()
				rd.SetId(aws.ToString(subnet.SubnetId))

				tflog.Info(ctx, "Reading resource")
				resourceSubnetFlatten(ctx, &subnet, rd)

				// set tags
				err = l.SetTags(ctx, awsClient, rd)
				if err != nil {
					result = fwdiag.NewListResultErrorDiagnostic(err)
					yield(result)
					return
				}

				if v, ok := tags["Name"]; ok {
					result.DisplayName = fmt.Sprintf("%s (%s)", v.ValueString(), aws.ToString(subnet.SubnetId))
				} else {
					result.DisplayName = aws.ToString(subnet.SubnetId)
				}

				l.SetResult(ctx, awsClient, request.IncludeResource, &result, rd)
				if result.Diagnostics.HasError() {
					yield(result)
					return
				}

				if !yield(result) {
					return
				}
			}
		}
	}
}
