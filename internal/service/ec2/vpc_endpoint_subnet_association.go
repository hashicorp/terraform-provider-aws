// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_vpc_endpoint_subnet_association", name="VPC Endpoint Subnet Association")
func resourceVPCEndpointSubnetAssociation() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceVPCEndpointSubnetAssociationCreate,
		ReadWithoutTimeout:   resourceVPCEndpointSubnetAssociationRead,
		DeleteWithoutTimeout: resourceVPCEndpointSubnetAssociationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVPCEndpointSubnetAssociationImport,
		},

		Schema: map[string]*schema.Schema{
			names.AttrSubnetID: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrVPCEndpointID: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
	}
}

func resourceVPCEndpointSubnetAssociationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	endpointID := d.Get(names.AttrVPCEndpointID).(string)
	subnetID := d.Get(names.AttrSubnetID).(string)
	// Human friendly ID for error messages since d.Id() is non-descriptive
	id := fmt.Sprintf("%s/%s", endpointID, subnetID)

	input := &ec2.ModifyVpcEndpointInput{
		VpcEndpointId: aws.String(endpointID),
		AddSubnetIds:  []string{subnetID},
	}

	log.Printf("[DEBUG] Creating VPC Endpoint Subnet Association: %v", input)

	// See https://github.com/hashicorp/terraform-provider-aws/issues/3382.
	// Prevent concurrent subnet association requests and delay between requests.
	mk := "vpc_endpoint_subnet_association_" + endpointID
	conns.GlobalMutexKV.Lock(mk)
	defer conns.GlobalMutexKV.Unlock(mk)

	c := &retry.StateChangeConf{
		Delay:   1 * time.Minute,
		Timeout: 3 * time.Minute,
		Target:  []string{"ok"},
		Refresh: func() (interface{}, string, error) {
			output, err := conn.ModifyVpcEndpoint(ctx, input)

			return output, "ok", err
		},
	}
	_, err := c.WaitForStateContext(ctx)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating VPC Endpoint Subnet Association (%s): %s", id, err)
	}

	d.SetId(vpcEndpointSubnetAssociationCreateID(endpointID, subnetID))

	_, err = waitVPCEndpointAvailable(ctx, conn, endpointID, d.Timeout(schema.TimeoutCreate))

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for VPC Endpoint (%s) to become available: %s", endpointID, err)
	}

	return append(diags, resourceVPCEndpointSubnetAssociationRead(ctx, d, meta)...)
}

func resourceVPCEndpointSubnetAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	endpointID := d.Get(names.AttrVPCEndpointID).(string)
	subnetID := d.Get(names.AttrSubnetID).(string)
	// Human friendly ID for error messages since d.Id() is non-descriptive
	id := fmt.Sprintf("%s/%s", endpointID, subnetID)

	err := findVPCEndpointSubnetAssociationExists(ctx, conn, endpointID, subnetID)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] VPC Endpoint Subnet Association (%s) not found, removing from state", id)
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading VPC Endpoint Subnet Association (%s): %s", id, err)
	}

	return diags
}

func resourceVPCEndpointSubnetAssociationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	endpointID := d.Get(names.AttrVPCEndpointID).(string)
	subnetID := d.Get(names.AttrSubnetID).(string)
	// Human friendly ID for error messages since d.Id() is non-descriptive
	id := fmt.Sprintf("%s/%s", endpointID, subnetID)

	input := &ec2.ModifyVpcEndpointInput{
		VpcEndpointId:   aws.String(endpointID),
		RemoveSubnetIds: []string{subnetID},
	}

	log.Printf("[DEBUG] Deleting VPC Endpoint Subnet Association: %s", id)
	_, err := conn.ModifyVpcEndpoint(ctx, input)

	if tfawserr.ErrCodeEquals(err, errCodeInvalidVPCEndpointIdNotFound) || tfawserr.ErrCodeEquals(err, errCodeInvalidSubnetIdNotFound) || tfawserr.ErrCodeEquals(err, errCodeInvalidParameter) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting VPC Endpoint Subnet Association (%s): %s", id, err)
	}

	_, err = waitVPCEndpointAvailable(ctx, conn, endpointID, d.Timeout(schema.TimeoutDelete))

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for VPC Endpoint (%s) to become available: %s", endpointID, err)
	}

	return diags
}

func resourceVPCEndpointSubnetAssociationImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	parts := strings.Split(d.Id(), "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("wrong format of import ID (%s), use: 'vpc-endpoint-id/subnet-id'", d.Id())
	}

	endpointID := parts[0]
	subnetID := parts[1]
	log.Printf("[DEBUG] Importing VPC Endpoint (%s) Subnet (%s) Association", endpointID, subnetID)

	d.SetId(vpcEndpointSubnetAssociationCreateID(endpointID, subnetID))
	d.Set(names.AttrVPCEndpointID, endpointID)
	d.Set(names.AttrSubnetID, subnetID)

	return []*schema.ResourceData{d}, nil
}
