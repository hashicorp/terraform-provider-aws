package aws

import (
	"fmt"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directconnect"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsDxGateway() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsDxGatewayCreate,
		Read:   resourceAwsDxGatewayRead,
		Delete: resourceAwsDxGatewayDelete,
		Importer: &schema.ResourceImporter{
			State: resourceAwsDxGatewayImportState,
		},

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"amazon_side_asn": &schema.Schema{
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateAmazonSideAsn,
			},
		},
	}
}

func resourceAwsDxGatewayCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).dxconn

	input := &directconnect.CreateDirectConnectGatewayInput{
		DirectConnectGatewayName: aws.String(d.Get("name").(string)),
	}

	if asn, ok := d.GetOk("amazon_side_asn"); ok {
		i, err := strconv.ParseInt(asn.(string), 10, 64)
		if err != nil {
			return err
		}
		input.AmazonSideAsn = aws.Int64(i)
	}

	resp, err := conn.CreateDirectConnectGateway(input)
	if err != nil {
		return err
	}
	gatewayId := aws.StringValue(resp.DirectConnectGateway.DirectConnectGatewayId)

	stateConf := &resource.StateChangeConf{
		Pending:    []string{directconnect.GatewayStatePending},
		Target:     []string{directconnect.GatewayStateAvailable},
		Refresh:    dxGatewayRefreshStateFunc(conn, gatewayId),
		Timeout:    10 * time.Minute,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}
	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("Error waiting for Direct Connect Gateway (%s) to become available: %s", gatewayId, err)
	}

	d.SetId(gatewayId)
	return resourceAwsDxGatewayRead(d, meta)
}

func resourceAwsDxGatewayRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).dxconn

	resp, err := conn.DescribeDirectConnectGateways(&directconnect.DescribeDirectConnectGatewaysInput{
		DirectConnectGatewayId: aws.String(d.Id()),
	})
	if err != nil {
		return err
	}

	if len(resp.DirectConnectGateways) < 1 {
		d.SetId("")
		return nil
	}
	if len(resp.DirectConnectGateways) != 1 {
		return fmt.Errorf("[ERROR] Number of Direct Connect Gateways (%s) isn't one, got %d", d.Id(), len(resp.DirectConnectGateways))
	}
	gateway := resp.DirectConnectGateways[0]

	if d.Id() != aws.StringValue(gateway.DirectConnectGatewayId) {
		return fmt.Errorf("[ERROR] Direct Connect Gateway (%s) not found", d.Id())
	}

	d.Set("name", aws.StringValue(gateway.DirectConnectGatewayName))
	d.Set("amazon_side_asn", strconv.FormatInt(aws.Int64Value(gateway.AmazonSideAsn), 10))

	return nil
}

func resourceAwsDxGatewayDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).dxconn

	_, err := conn.DeleteDirectConnectGateway(&directconnect.DeleteDirectConnectGatewayInput{
		DirectConnectGatewayId: aws.String(d.Id()),
	})
	if err != nil {
		return err
	}
	stateConf := &resource.StateChangeConf{
		Pending:    []string{directconnect.GatewayStatePending, directconnect.GatewayStateAvailable, directconnect.GatewayStateDeleting},
		Target:     []string{directconnect.GatewayStateDeleted},
		Refresh:    dxGatewayRefreshStateFunc(conn, d.Id()),
		Timeout:    10 * time.Minute,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}
	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("Error waiting for Direct Connect Gateway (%s) to be deleted: %s", d.Id(), err)
	}
	d.SetId("")
	return nil
}

func dxGatewayRefreshStateFunc(conn *directconnect.DirectConnect, gatewayId string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		resp, err := conn.DescribeDirectConnectGateways(&directconnect.DescribeDirectConnectGatewaysInput{
			DirectConnectGatewayId: aws.String(gatewayId),
		})
		if err != nil {
			return nil, "failed", err
		}
		return resp, *resp.DirectConnectGateways[0].DirectConnectGatewayState, nil
	}
}
