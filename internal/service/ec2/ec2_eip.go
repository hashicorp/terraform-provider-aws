package ec2

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

const (
	// Maximum amount of time to wait for EIP association with EC2-Classic instances
	addressAssociationClassicTimeout = 2 * time.Minute
)

func ResourceEIP() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceEIPCreate,
		ReadWithoutTimeout:   resourceEIPRead,
		UpdateWithoutTimeout: resourceEIPUpdate,
		DeleteWithoutTimeout: resourceEIPDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		CustomizeDiff: verify.SetTagsDiff,

		Timeouts: &schema.ResourceTimeout{
			Read:   schema.DefaultTimeout(15 * time.Minute),
			Update: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(3 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"address": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"allocation_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"associate_with_private_ip": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"association_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"carrier_ip": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"customer_owned_ip": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"customer_owned_ipv4_pool": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"domain": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"instance": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"network_border_group": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"network_interface": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"private_dns": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"private_ip": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"public_dns": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"public_ip": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"public_ipv4_pool": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
			"vpc": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},
		},
	}
}

func resourceEIPCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := &ec2.AllocateAddressInput{}

	if v := d.Get("vpc"); v != nil && v.(bool) {
		input.Domain = aws.String(ec2.DomainTypeVpc)
	}

	if v, ok := d.GetOk("address"); ok {
		input.Address = aws.String(v.(string))
	}

	if v, ok := d.GetOk("customer_owned_ipv4_pool"); ok {
		input.CustomerOwnedIpv4Pool = aws.String(v.(string))
	}

	if v, ok := d.GetOk("network_border_group"); ok {
		input.NetworkBorderGroup = aws.String(v.(string))
	}

	if v, ok := d.GetOk("public_ipv4_pool"); ok {
		input.PublicIpv4Pool = aws.String(v.(string))
	}

	if len(tags) > 0 {
		input.TagSpecifications = tagSpecificationsFromKeyValueTags(tags, ec2.ResourceTypeElasticIp)
	}

	log.Printf("[DEBUG] Creating EC2 EIP: %s", input)
	output, err := conn.AllocateAddressWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EC2 EIP: %s", err)
	}

	d.SetId(aws.StringValue(output.AllocationId))

	_, err = tfresource.RetryWhenNotFound(ctx, d.Timeout(schema.TimeoutCreate), func() (interface{}, error) {
		return FindEIPByAllocationID(ctx, conn, d.Id())
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EC2 EIP (%s) create: %s", d.Id(), err)
	}

	instanceID := d.Get("instance").(string)
	eniID := d.Get("network_interface").(string)

	if instanceID != "" || eniID != "" {
		_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, d.Timeout(schema.TimeoutCreate),
			func() (interface{}, error) {
				return nil, associateEIP(ctx, conn, d.Id(), instanceID, eniID, d.Get("associate_with_private_ip").(string))
			}, errCodeInvalidAllocationIDNotFound)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "creating EC2 EIP (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceEIPRead(ctx, d, meta)...)
}

func resourceEIPRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	var err error
	var address *ec2.Address

	if eipID(d.Id()).IsVPC() {
		address, err = FindEIPByAllocationID(ctx, conn, d.Id())
	} else {
		address, err = FindEIPByPublicIP(ctx, conn, d.Id())
	}

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 EIP (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 EIP (%s): %s", d.Id(), err)
	}

	d.Set("allocation_id", address.AllocationId)
	d.Set("association_id", address.AssociationId)
	d.Set("carrier_ip", address.CarrierIp)
	d.Set("customer_owned_ip", address.CustomerOwnedIp)
	d.Set("customer_owned_ipv4_pool", address.CustomerOwnedIpv4Pool)
	d.Set("domain", address.Domain)
	d.Set("instance", address.InstanceId)
	d.Set("network_border_group", address.NetworkBorderGroup)
	d.Set("network_interface", address.NetworkInterfaceId)
	d.Set("public_ipv4_pool", address.PublicIpv4Pool)
	d.Set("vpc", aws.StringValue(address.Domain) == ec2.DomainTypeVpc)

	d.Set("private_ip", address.PrivateIpAddress)
	if v := aws.StringValue(address.PrivateIpAddress); v != "" {
		d.Set("private_dns", PrivateDNSNameForIP(meta.(*conns.AWSClient), v))
	}

	d.Set("public_ip", address.PublicIp)
	if v := aws.StringValue(address.PublicIp); v != "" {
		d.Set("public_dns", PublicDNSNameForIP(meta.(*conns.AWSClient), v))
	}

	// Force ID to be an Allocation ID if we're on a VPC.
	// This allows users to import the EIP based on the IP if they are in a VPC.
	if aws.StringValue(address.Domain) == ec2.DomainTypeVpc && net.ParseIP(d.Id()) != nil {
		d.SetId(aws.StringValue(address.AllocationId))
	}

	tags := KeyValueTags(address.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags_all: %s", err)
	}

	return diags
}

func resourceEIPUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn()

	if d.HasChanges("associate_with_private_ip", "instance", "network_interface") {
		o, n := d.GetChange("instance")
		oldInstanceID, newInstanceID := o.(string), n.(string)
		associationID := d.Get("association_id").(string)

		if oldInstanceID != "" || associationID != "" {
			if err := disassociateEIP(ctx, conn, d.Id(), associationID); err != nil {
				return sdkdiag.AppendErrorf(diags, "updating EC2 EIP (%s): %s", d.Id(), err)
			}
		}

		newNetworkInterfaceID := d.Get("network_interface").(string)

		if newInstanceID != "" || newNetworkInterfaceID != "" {
			if err := associateEIP(ctx, conn, d.Id(), newInstanceID, newNetworkInterfaceID, d.Get("associate_with_private_ip").(string)); err != nil {
				return sdkdiag.AppendErrorf(diags, "updating EC2 EIP (%s): %s", d.Id(), err)
			}
		}
	}

	if d.HasChange("tags_all") {
		if d.Get("domain").(string) == ec2.DomainTypeStandard {
			return sdkdiag.AppendErrorf(diags, "tags cannot be set for a standard-domain EIP - must be a VPC-domain EIP")
		}

		o, n := d.GetChange("tags_all")

		if err := UpdateTags(ctx, conn, d.Id(), o, n); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating EC2 EIP (%s) tags: %s", d.Id(), err)
		}
	}

	return append(diags, resourceEIPRead(ctx, d, meta)...)
}

func resourceEIPDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn()

	// If we are attached to an instance or interface, detach first.
	if associationID := d.Get("association_id").(string); associationID != "" || d.Get("instance").(string) != "" {
		if err := disassociateEIP(ctx, conn, d.Id(), associationID); err != nil {
			return sdkdiag.AppendErrorf(diags, "deleting EC2 EIP (%s): %s", d.Id(), err)
		}
	}

	input := &ec2.ReleaseAddressInput{}

	if eipID(d.Id()).IsVPC() {
		input.AllocationId = aws.String(d.Id())

		if v, ok := d.GetOk("network_border_group"); ok {
			input.NetworkBorderGroup = aws.String(v.(string))
		}
	} else {
		input.PublicIp = aws.String(d.Id())
	}

	_, err := conn.ReleaseAddressWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, errCodeInvalidAllocationIDNotFound) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting EC2 EIP (%s): %s", d.Id(), err)
	}

	return diags
}

type eipID string

// IsVPC returns whether or not the EIP is in the VPC domain.
func (id eipID) IsVPC() bool {
	return strings.HasPrefix(string(id), "eipalloc-")
}

func associateEIP(ctx context.Context, conn *ec2.EC2, id, instanceID, networkInterfaceID, privateIPAddress string) error {
	input := &ec2.AssociateAddressInput{}

	if eipID(id).IsVPC() {
		input.AllocationId = aws.String(id)
	} else {
		input.PublicIp = aws.String(id)
	}

	if instanceID != "" {
		input.InstanceId = aws.String(instanceID)
	}

	if networkInterfaceID != "" {
		input.NetworkInterfaceId = aws.String(networkInterfaceID)
	}

	if privateIPAddress != "" {
		input.PrivateIpAddress = aws.String(privateIPAddress)
	}

	output, err := conn.AssociateAddressWithContext(ctx, input)

	if err != nil {
		return fmt.Errorf("associating: %w", err)
	}

	if associationID := aws.StringValue(output.AssociationId); associationID != "" {
		_, err := tfresource.RetryWhen(ctx, propagationTimeout,
			func() (interface{}, error) {
				return FindEIPByAssociationID(ctx, conn, associationID)
			},
			func(err error) (bool, error) {
				if tfresource.NotFound(err) {
					return true, err
				}

				// "InvalidInstanceID: The pending instance 'i-0504e5b44ea06d599' is not in a valid state for this operation."
				if tfawserr.ErrMessageContains(err, errCodeInvalidInstanceID, "pending instance") {
					return true, err
				}

				return false, err
			},
		)

		if err != nil {
			return fmt.Errorf("associating: waiting for completion: %w", err)
		}
	} else {
		if err := waitForAddressAssociationClassic(ctx, conn, id, instanceID); err != nil {
			return fmt.Errorf("associating: waiting for completion: %w", err)
		}
	}

	return nil
}

func disassociateEIP(ctx context.Context, conn *ec2.EC2, id, associationID string) error {
	input := &ec2.DisassociateAddressInput{}

	if eipID(id).IsVPC() {
		if associationID == "" {
			return nil
		}

		input.AssociationId = aws.String(associationID)
	} else {
		input.PublicIp = aws.String(id)
	}

	_, err := conn.DisassociateAddressWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, errCodeInvalidAssociationIDNotFound) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("disassociating: %w", err)
	}

	return nil
}

// waitForAddressAssociationClassic ensures the correct Instance is associated with an Address
//
// This can take a few seconds to appear correctly for EC2-Classic addresses.
func waitForAddressAssociationClassic(ctx context.Context, conn *ec2.EC2, publicIP, instanceID string) error {
	err := resource.RetryContext(ctx, addressAssociationClassicTimeout, func() *resource.RetryError {
		address, err := FindEIPByPublicIP(ctx, conn, publicIP)

		if tfresource.NotFound(err) {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		if aws.StringValue(address.InstanceId) != instanceID {
			return resource.RetryableError(errors.New("not associated"))
		}

		return nil
	})

	if tfresource.TimedOut(err) { // nosemgrep:ci.helper-schema-TimeoutError-check-doesnt-return-output
		_, err = FindEIPByPublicIP(ctx, conn, publicIP)
	}

	return err
}

func ConvertIPToDashIP(ip string) string {
	return strings.Replace(ip, ".", "-", -1)
}

func PrivateDNSNameForIP(client *conns.AWSClient, ip string) string {
	return fmt.Sprintf("ip-%s.%s", ConvertIPToDashIP(ip), RegionalPrivateDNSSuffix(client.Region))
}

func PublicDNSNameForIP(client *conns.AWSClient, ip string) string {
	return client.PartitionHostname(fmt.Sprintf("ec2-%s.%s", ConvertIPToDashIP(ip), RegionalPublicDNSSuffix(client.Region)))
}
