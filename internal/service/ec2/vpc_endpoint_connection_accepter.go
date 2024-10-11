// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_vpc_endpoint_connection_accepter", name="VPC Endpoint Connection Accepter")
func resourceVPCEndpointConnectionAccepter() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceVPCEndpointConnectionAccepterCreate,
		ReadWithoutTimeout:   resourceVPCEndpointConnectionAccepterRead,
		DeleteWithoutTimeout: resourceVPCEndpointConnectionAccepterDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrVPCEndpointID: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"vpc_endpoint_service_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"vpc_endpoint_state": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceVPCEndpointConnectionAccepterCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	serviceID := d.Get("vpc_endpoint_service_id").(string)
	vpcEndpointID := d.Get(names.AttrVPCEndpointID).(string)
	id := vpcEndpointConnectionAccepterCreateResourceID(serviceID, vpcEndpointID)
	input := &ec2.AcceptVpcEndpointConnectionsInput{
		ServiceId:      aws.String(serviceID),
		VpcEndpointIds: []string{vpcEndpointID},
	}

	_, err := conn.AcceptVpcEndpointConnections(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "accepting VPC Endpoint Connection (%s): %s", id, err)
	}

	d.SetId(id)

	if _, err := waitVPCEndpointConnectionAccepted(ctx, conn, serviceID, vpcEndpointID, d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for VPC Endpoint Connection (%s) accept: %s", d.Id(), err)
	}

	return append(diags, resourceVPCEndpointConnectionAccepterRead(ctx, d, meta)...)
}

func resourceVPCEndpointConnectionAccepterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	serviceID, vpcEndpointID, err := vpcEndpointConnectionAccepterParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	vpcEndpointConnection, err := findVPCEndpointConnectionByServiceIDAndVPCEndpointID(ctx, conn, serviceID, vpcEndpointID)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] VPC Endpoint Connection Accepter %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading VPC Endpoint Connection (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrVPCEndpointID, vpcEndpointConnection.VpcEndpointId)
	d.Set("vpc_endpoint_service_id", vpcEndpointConnection.ServiceId)
	d.Set("vpc_endpoint_state", vpcEndpointConnection.VpcEndpointState)

	return diags
}

func resourceVPCEndpointConnectionAccepterDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	serviceID, vpcEndpointID, err := vpcEndpointConnectionAccepterParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[DEBUG] Rejecting VPC Endpoint Connection: %s", d.Id())
	_, err = conn.RejectVpcEndpointConnections(ctx, &ec2.RejectVpcEndpointConnectionsInput{
		ServiceId:      aws.String(serviceID),
		VpcEndpointIds: []string{vpcEndpointID},
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidVPCEndpointServiceIdNotFound) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "rejecting VPC Endpoint Connection (%s): %s", d.Id(), err)
	}

	return diags
}

const vpcEndpointConnectionAccepterResourceIDSeparator = "_"

func vpcEndpointConnectionAccepterCreateResourceID(serviceID, vpcEndpointID string) string {
	parts := []string{serviceID, vpcEndpointID}
	id := strings.Join(parts, vpcEndpointConnectionAccepterResourceIDSeparator)

	return id
}

func vpcEndpointConnectionAccepterParseResourceID(id string) (string, string, error) {
	parts := strings.Split(id, vpcEndpointConnectionAccepterResourceIDSeparator)

	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected VPCEndpointServiceID%[2]sVPCEndpointID", id, vpcEndpointConnectionAccepterResourceIDSeparator)
}
