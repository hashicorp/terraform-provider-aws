// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
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

// @SDKResource("aws_eip", name="EIP")
// @Tags(identifierAttribute="id")
// @Testing(tagsTest=false)
func resourceEIP() *schema.Resource {
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
			names.AttrAddress: {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"allocation_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"associate_with_private_ip": {
				Type:     schema.TypeString,
				Optional: true,
			},
			names.AttrAssociationID: {
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
			names.AttrDomain: {
				Type:             schema.TypeString,
				ForceNew:         true,
				Optional:         true,
				Computed:         true,
				ValidateDiagFunc: enum.Validate[types.DomainType](),
				ConflictsWith:    []string{"vpc"},
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
			"ptr_record": {
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
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"vpc": {
				Type:          schema.TypeBool,
				Optional:      true,
				ForceNew:      true,
				Computed:      true,
				Deprecated:    "use domain attribute instead",
				ConflictsWith: []string{names.AttrDomain},
			},
		},
	}
}

func resourceEIPCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	input := &ec2.AllocateAddressInput{
		TagSpecifications: getTagSpecificationsIn(ctx, types.ResourceTypeElasticIp),
	}

	if v, ok := d.GetOk(names.AttrAddress); ok {
		input.Address = aws.String(v.(string))
	}

	if v, ok := d.GetOk("customer_owned_ipv4_pool"); ok {
		input.CustomerOwnedIpv4Pool = aws.String(v.(string))
	}

	if v := d.Get(names.AttrDomain); v != nil && v.(string) != "" {
		input.Domain = types.DomainType(v.(string))
	}

	if v := d.Get("vpc"); v != nil && v.(bool) {
		input.Domain = types.DomainTypeVpc
	}

	if v, ok := d.GetOk("network_border_group"); ok {
		input.NetworkBorderGroup = aws.String(v.(string))
	}

	if v, ok := d.GetOk("public_ipv4_pool"); ok {
		input.PublicIpv4Pool = aws.String(v.(string))
	}

	output, err := conn.AllocateAddress(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EC2 EIP: %s", err)
	}

	d.SetId(aws.ToString(output.AllocationId))

	_, err = tfresource.RetryWhenNotFound(ctx, d.Timeout(schema.TimeoutCreate), func() (interface{}, error) {
		return findEIPByAllocationID(ctx, conn, d.Id())
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EC2 EIP (%s) create: %s", d.Id(), err)
	}

	if instanceID, eniID := d.Get("instance").(string), d.Get("network_interface").(string); instanceID != "" || eniID != "" {
		_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, d.Timeout(schema.TimeoutCreate),
			func() (interface{}, error) {
				return nil, associateEIP(ctx, conn, d.Id(), instanceID, eniID, d.Get("associate_with_private_ip").(string))
			}, errCodeInvalidAllocationIDNotFound)

		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	return append(diags, resourceEIPRead(ctx, d, meta)...)
}

func resourceEIPRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	if !eipID(d.Id()).IsVPC() {
		return sdkdiag.AppendErrorf(diags, `with the retirement of EC2-Classic %s domain EC2 EIPs are no longer supported`, types.DomainTypeStandard)
	}

	outputRaw, err := tfresource.RetryWhenNewResourceNotFound(ctx, ec2PropagationTimeout, func() (interface{}, error) {
		return findEIPByAllocationID(ctx, conn, d.Id())
	}, d.IsNewResource())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 EIP (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 EIP (%s): %s", d.Id(), err)
	}

	address := outputRaw.(*types.Address)
	allocationID := aws.ToString(address.AllocationId)
	d.Set("allocation_id", allocationID)
	d.Set(names.AttrARN, eipARN(meta.(*conns.AWSClient), allocationID))
	d.Set(names.AttrAssociationID, address.AssociationId)
	d.Set("carrier_ip", address.CarrierIp)
	d.Set("customer_owned_ip", address.CustomerOwnedIp)
	d.Set("customer_owned_ipv4_pool", address.CustomerOwnedIpv4Pool)
	d.Set(names.AttrDomain, address.Domain)
	d.Set("instance", address.InstanceId)
	d.Set("network_border_group", address.NetworkBorderGroup)
	d.Set("network_interface", address.NetworkInterfaceId)
	d.Set("public_ipv4_pool", address.PublicIpv4Pool)
	d.Set("private_ip", address.PrivateIpAddress)
	if v := aws.ToString(address.PrivateIpAddress); v != "" {
		d.Set("private_dns", meta.(*conns.AWSClient).EC2PrivateDNSNameForIP(ctx, v))
	}
	d.Set("public_ip", address.PublicIp)
	if v := aws.ToString(address.PublicIp); v != "" {
		d.Set("public_dns", meta.(*conns.AWSClient).EC2PublicDNSNameForIP(ctx, v))
	}
	d.Set("vpc", address.Domain == types.DomainTypeVpc)

	// Force ID to be an Allocation ID if we're on a VPC.
	// This allows users to import the EIP based on the IP if they are in a VPC.
	if address.Domain == types.DomainTypeVpc && net.ParseIP(d.Id()) != nil {
		d.SetId(aws.ToString(address.AllocationId))
	}

	addressAttr, err := findEIPDomainNameAttributeByAllocationID(ctx, conn, d.Id())

	switch {
	case err == nil:
		d.Set("ptr_record", strings.TrimSuffix(aws.ToString(addressAttr.PtrRecord), "."))
	case tfresource.NotFound(err):
		d.Set("ptr_record", nil)
	default:
		return sdkdiag.AppendErrorf(diags, "reading EC2 EIP (%s) domain name attribute: %s", d.Id(), err)
	}

	setTagsOut(ctx, address.Tags)

	return diags
}

func resourceEIPUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	if d.HasChanges("associate_with_private_ip", "instance", "network_interface") {
		o, n := d.GetChange("instance")
		oldInstanceID, newInstanceID := o.(string), n.(string)

		if associationID := d.Get(names.AttrAssociationID).(string); oldInstanceID != "" || associationID != "" {
			if err := disassociateEIP(ctx, conn, associationID); err != nil {
				return sdkdiag.AppendFromErr(diags, err)
			}
		}

		if newNetworkInterfaceID := d.Get("network_interface").(string); newInstanceID != "" || newNetworkInterfaceID != "" {
			if err := associateEIP(ctx, conn, d.Id(), newInstanceID, newNetworkInterfaceID, d.Get("associate_with_private_ip").(string)); err != nil {
				return sdkdiag.AppendFromErr(diags, err)
			}
		}
	}

	return append(diags, resourceEIPRead(ctx, d, meta)...)
}

func resourceEIPDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	if !eipID(d.Id()).IsVPC() {
		return sdkdiag.AppendErrorf(diags, `with the retirement of EC2-Classic %s domain EC2 EIPs are no longer supported`, types.DomainTypeStandard)
	}

	// If we are attached to an instance or interface, detach first.
	if associationID := d.Get(names.AttrAssociationID).(string); associationID != "" || d.Get("instance").(string) != "" {
		if err := disassociateEIP(ctx, conn, associationID); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	input := &ec2.ReleaseAddressInput{
		AllocationId: aws.String(d.Id()),
	}

	if v, ok := d.GetOk("network_border_group"); ok {
		input.NetworkBorderGroup = aws.String(v.(string))
	}

	log.Printf("[INFO] Deleting EC2 EIP: %s", d.Id())
	_, err := conn.ReleaseAddress(ctx, input)

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

func associateEIP(ctx context.Context, conn *ec2.Client, allocationID, instanceID, networkInterfaceID, privateIPAddress string) error {
	input := &ec2.AssociateAddressInput{
		AllocationId: aws.String(allocationID),
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

	output, err := conn.AssociateAddress(ctx, input)

	if err != nil {
		return fmt.Errorf("associating EC2 EIP (%s): %w", allocationID, err)
	}

	_, err = tfresource.RetryWhen(ctx, ec2PropagationTimeout,
		func() (interface{}, error) {
			return findEIPByAssociationID(ctx, conn, aws.ToString(output.AssociationId))
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
		return fmt.Errorf("waiting for EC2 EIP (%s) association: %w", allocationID, err)
	}

	return nil
}

func disassociateEIP(ctx context.Context, conn *ec2.Client, associationID string) error {
	if associationID == "" {
		return nil
	}

	input := &ec2.DisassociateAddressInput{
		AssociationId: aws.String(associationID),
	}

	_, err := conn.DisassociateAddress(ctx, input)

	if tfawserr.ErrCodeEquals(err, errCodeInvalidAssociationIDNotFound) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("disassociating EC2 EIP (%s): %w", associationID, err)
	}

	return nil
}

func eipARN(c *conns.AWSClient, allocationID string) string {
	return arn.ARN{
		Partition: c.Partition,
		Service:   names.EC2,
		Region:    c.Region,
		AccountID: c.AccountID,
		Resource:  "elastic-ip/" + allocationID,
	}.String()
}
