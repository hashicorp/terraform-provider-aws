// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKResource("aws_eip_association")
func ResourceEIPAssociation() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceEIPAssociationCreate,
		ReadWithoutTimeout:   resourceEIPAssociationRead,
		DeleteWithoutTimeout: resourceEIPAssociationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"allocation_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"allow_reassociation": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
			},
			"instance_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"network_interface_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"private_ip_address": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"public_ip": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
		},
	}
}

func resourceEIPAssociationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn(ctx)

	input := &ec2.AssociateAddressInput{}

	if v, ok := d.GetOk("allocation_id"); ok {
		input.AllocationId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("allow_reassociation"); ok {
		input.AllowReassociation = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("instance_id"); ok {
		input.InstanceId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("network_interface_id"); ok {
		input.NetworkInterfaceId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("private_ip_address"); ok {
		input.PrivateIpAddress = aws.String(v.(string))
	}

	if v, ok := d.GetOk("public_ip"); ok {
		input.PublicIp = aws.String(v.(string))
	}

	output, err := conn.AssociateAddressWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EC2 EIP Association: %s", err)
	}

	d.SetId(aws.StringValue(output.AssociationId))

	_, err = tfresource.RetryWhen(ctx, ec2PropagationTimeout,
		func() (interface{}, error) {
			return FindEIPByAssociationID(ctx, conn, d.Id())
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
		return sdkdiag.AppendErrorf(diags, "waiting for EC2 EIP Association (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceEIPAssociationRead(ctx, d, meta)...)
}

func resourceEIPAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn(ctx)

	if !eipAssociationID(d.Id()).IsVPC() {
		return sdkdiag.AppendErrorf(diags, `with the retirement of EC2-Classic %s domain EC2 EIPs are no longer supported`, ec2.DomainTypeStandard)
	}

	address, err := FindEIPByAssociationID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 EIP Association (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 EIP Association (%s): %s", d.Id(), err)
	}

	d.Set("allocation_id", address.AllocationId)
	d.Set("instance_id", address.InstanceId)
	d.Set("network_interface_id", address.NetworkInterfaceId)
	d.Set("private_ip_address", address.PrivateIpAddress)
	d.Set("public_ip", address.PublicIp)

	return diags
}

func resourceEIPAssociationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn(ctx)

	if !eipAssociationID(d.Id()).IsVPC() {
		return sdkdiag.AppendErrorf(diags, `with the retirement of EC2-Classic %s domain EC2 EIPs are no longer supported`, ec2.DomainTypeStandard)
	}

	input := &ec2.DisassociateAddressInput{
		AssociationId: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Deleting EC2 EIP Association: %s", d.Id())
	_, err := conn.DisassociateAddressWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, errCodeInvalidAssociationIDNotFound) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting EC2 EIP Association (%s): %s", d.Id(), err)
	}

	return diags
}

type eipAssociationID string

// IsVPC returns whether or not the associated EIP is in the VPC domain.
func (id eipAssociationID) IsVPC() bool {
	return strings.HasPrefix(string(id), "eipassoc-")
}
