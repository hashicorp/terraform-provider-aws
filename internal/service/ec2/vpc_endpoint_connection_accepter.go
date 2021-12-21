package ec2

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func ResourceVPCEndpointConnectionAccepter() *schema.Resource {
	return &schema.Resource{
		Create: resourceVPCEndpointConnectionAccepterCreate,
		Read:   resourceVPCEndpointConnectionAccepterRead,
		Delete: resourceVPCEndpointConnectionAccepterDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"state": {
				Type:     schema.TypeString,
				Computed: true,
			},
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
		},
	}
}

func resourceVPCEndpointConnectionAccepterCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	serviceID := d.Get("vpc_endpoint_service_id").(string)
	vpcEndpointID := d.Get("vpc_endpoint_id").(string)
	input := &ec2.AcceptVpcEndpointConnectionsInput{
		ServiceId:      aws.String(serviceID),
		VpcEndpointIds: aws.StringSlice([]string{vpcEndpointID}),
	}

	log.Printf("[DEBUG] Accepting VPC Endpoint Connection: %s", input)
	_, err := conn.AcceptVpcEndpointConnections(input)

	if err != nil {
		return fmt.Errorf("error accepting VPC Endpoint Connection: %w", err)
	}

	d.SetId(vpcEndpointConnectionAccepterCreateResourceID(serviceID, vpcEndpointID))

	stateConf := &resource.StateChangeConf{
		Pending:    []string{"pendingAcceptance", "pending"},
		Target:     []string{"available"},
		Refresh:    vpcEndpointConnectionRefresh(conn, serviceID, vpcEndpointID),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      5 * time.Second,
		MinTimeout: 5 * time.Second,
	}

	vpceConn, err := stateConf.WaitForStateContext(context.Background())
	if err != nil {
		return fmt.Errorf("error waiting for VPC Endpoint (%s) to be accepted by VPC Endpoint Service (%s): %s", vpcEndpointID, serviceID, err)
	}

	d.Set("state", vpceConn.(ec2.VpcEndpointConnection).VpcEndpointState)

	return resourceVPCEndpointConnectionAccepterRead(d, meta)
}

func vpcEndpointConnectionRefresh(conn *ec2.EC2, svcID, vpceID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		log.Printf("[DEBUG] Reading VPC Endpoint Connections for VPC Endpoint Service (%s)", svcID)

		input := &ec2.DescribeVpcEndpointConnectionsInput{
			Filters: BuildAttributeFilterList(map[string]string{"service-id": svcID}),
		}

		var vpceConn *ec2.VpcEndpointConnection

		paginator := func(page *ec2.DescribeVpcEndpointConnectionsOutput, lastPage bool) bool {
			for _, c := range page.VpcEndpointConnections {
				if aws.StringValue(c.VpcEndpointId) == vpceID {
					log.Printf("[DEBUG] Found VPC Endpoint Connection from VPC Endpoint Service (%s) to VPC Endpoint (%s): %s", svcID, vpceID, *c)
					vpceConn = c
					return false
				}
			}
			return !lastPage
		}

		if err := conn.DescribeVpcEndpointConnectionsPages(input, paginator); err != nil {
			return nil, "", err
		}

		if vpceConn == nil {
			return nil, "", fmt.Errorf("VPC Endpoint Connection from VPC Endpoint Service (%s) to VPC Endpoint (%s) not found", svcID, vpceID)
		}

		state := aws.StringValue(vpceConn.VpcEndpointState)
		log.Printf("[DEBUG] state %s", state)

		// No point in retrying if the endpoint is in a failed state.
		if state == "failed" {
			return nil, state, fmt.Errorf("VPC Endpoint Connection state %q", state)
		}

		return *vpceConn, state, nil
	}
}

func resourceVPCEndpointConnectionAccepterRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	serviceID, vpcEndpointID, err := vpcEndpointConnectionAccepterParseResourceID(d.Id())

	if err != nil {
		return err
	}

	vpceConn, state, err := vpcEndpointConnectionRefresh(conn, serviceID, vpcEndpointID)()
	if err != nil && state != "failed" {
		return fmt.Errorf("error reading VPC Endpoint Connection from VPC Endpoint Service (%s) to VPC Endpoint (%s): %s", serviceID, vpcEndpointID, err)
	}

	d.Set("state", vpceConn.(ec2.VpcEndpointConnection).VpcEndpointState)
	d.Set("vpc_endpoint_id", vpcEndpointID)
	d.Set("vpc_endpoint_service_id", serviceID)

	return nil
}

func resourceVPCEndpointConnectionAccepterDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	serviceID, vpcEndpointID, err := vpcEndpointConnectionAccepterParseResourceID(d.Id())

	if err != nil {
		return err
	}

	input := &ec2.RejectVpcEndpointConnectionsInput{
		ServiceId:      aws.String(serviceID),
		VpcEndpointIds: aws.StringSlice([]string{vpcEndpointID}),
	}

	if _, err := conn.RejectVpcEndpointConnections(input); err != nil {
		return fmt.Errorf("error rejecting VPC Endpoint Connection from VPC Endpoint (%s) to VPC Endpoint Service (%s): %s", serviceID, vpcEndpointID, err)
	}

	return nil
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
