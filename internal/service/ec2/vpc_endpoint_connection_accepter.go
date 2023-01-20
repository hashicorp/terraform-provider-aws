package ec2

import (
	"context"
	"fmt"
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

func ResourceVPCEndpointConnectionAccepter() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceVPCEndpointConnectionAccepterCreate,
		ReadWithoutTimeout:   resourceVPCEndpointConnectionAccepterRead,
		DeleteWithoutTimeout: resourceVPCEndpointConnectionAccepterDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"vpc_endpoint_id": {
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
	conn := meta.(*conns.AWSClient).EC2Conn()

	serviceID := d.Get("vpc_endpoint_service_id").(string)
	vpcEndpointID := d.Get("vpc_endpoint_id").(string)
	id := VPCEndpointConnectionAccepterCreateResourceID(serviceID, vpcEndpointID)
	input := &ec2.AcceptVpcEndpointConnectionsInput{
		ServiceId:      aws.String(serviceID),
		VpcEndpointIds: aws.StringSlice([]string{vpcEndpointID}),
	}

	log.Printf("[DEBUG] Accepting VPC Endpoint Connection: %s", input)
	_, err := conn.AcceptVpcEndpointConnectionsWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "accepting VPC Endpoint Connection (%s): %s", id, err)
	}

	d.SetId(id)

	_, err = waitVPCEndpointConnectionAccepted(ctx, conn, serviceID, vpcEndpointID, d.Timeout(schema.TimeoutCreate))

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for VPC Endpoint Connection (%s) to be accepted: %s", d.Id(), err)
	}

	return append(diags, resourceVPCEndpointConnectionAccepterRead(ctx, d, meta)...)
}

func resourceVPCEndpointConnectionAccepterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn()

	serviceID, vpcEndpointID, err := VPCEndpointConnectionAccepterParseResourceID(d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading VPC Endpoint Connection (%s): %s", d.Id(), err)
	}

	vpcEndpointConnection, err := FindVPCEndpointConnectionByServiceIDAndVPCEndpointID(ctx, conn, serviceID, vpcEndpointID)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] VPC Endpoint Connection %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading VPC Endpoint Connection (%s): %s", d.Id(), err)
	}

	d.Set("vpc_endpoint_id", vpcEndpointConnection.VpcEndpointId)
	d.Set("vpc_endpoint_service_id", vpcEndpointConnection.ServiceId)
	d.Set("vpc_endpoint_state", vpcEndpointConnection.VpcEndpointState)

	return diags
}

func resourceVPCEndpointConnectionAccepterDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn()

	serviceID, vpcEndpointID, err := VPCEndpointConnectionAccepterParseResourceID(d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting VPC Endpoint Connection (%s): %s", d.Id(), err)
	}

	input := &ec2.RejectVpcEndpointConnectionsInput{
		ServiceId:      aws.String(serviceID),
		VpcEndpointIds: aws.StringSlice([]string{vpcEndpointID}),
	}

	_, err = conn.RejectVpcEndpointConnectionsWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, errCodeInvalidVPCEndpointServiceIdNotFound) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "rejecting VPC Endpoint Connection (%s): %s", d.Id(), err)
	}

	return diags
}

const vpcEndpointConnectionAccepterResourceIDSeparator = "_"

func VPCEndpointConnectionAccepterCreateResourceID(serviceID, vpcEndpointID string) string {
	parts := []string{serviceID, vpcEndpointID}
	id := strings.Join(parts, vpcEndpointConnectionAccepterResourceIDSeparator)

	return id
}

func VPCEndpointConnectionAccepterParseResourceID(id string) (string, string, error) {
	parts := strings.Split(id, vpcEndpointConnectionAccepterResourceIDSeparator)

	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected VPCEndpointServiceID%[2]sVPCEndpointID", id, vpcEndpointConnectionAccepterResourceIDSeparator)
}
