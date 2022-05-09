package ec2

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceClientVPNRoute() *schema.Resource {
	return &schema.Resource{
		Create: resourceClientVPNRouteCreate,
		Read:   resourceClientVPNRouteRead,
		Delete: resourceClientVPNRouteDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(ClientVPNRouteCreatedTimeout),
			Delete: schema.DefaultTimeout(ClientVPNRouteDeletedTimeout),
		},

		Schema: map[string]*schema.Schema{
			"client_vpn_endpoint_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"destination_cidr_block": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidIPv4CIDRNetworkAddress,
			},
			"origin": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"target_vpc_subnet_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"type": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceClientVPNRouteCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	endpointID := d.Get("client_vpn_endpoint_id").(string)
	targetSubnetID := d.Get("target_vpc_subnet_id").(string)
	destinationCIDR := d.Get("destination_cidr_block").(string)
	id := ClientVPNRouteCreateResourceID(endpointID, targetSubnetID, destinationCIDR)
	input := &ec2.CreateClientVpnRouteInput{
		ClientVpnEndpointId:  aws.String(endpointID),
		DestinationCidrBlock: aws.String(destinationCIDR),
		TargetVpcSubnetId:    aws.String(targetSubnetID),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating EC2 Client VPN Route: %s", input)
	_, err := tfresource.RetryWhenAWSErrCodeEquals(propagationTimeout, func() (interface{}, error) {
		return conn.CreateClientVpnRoute(input)
	}, ErrCodeInvalidClientVpnActiveAssociationNotFound)

	if err != nil {
		return fmt.Errorf("error creating EC2 Client VPN Route (%s): %w", id, err)
	}

	d.SetId(id)

	if _, err := WaitClientVPNRouteCreated(conn, endpointID, targetSubnetID, destinationCIDR, d.Timeout(schema.TimeoutCreate)); err != nil {
		return fmt.Errorf("error waiting for EC2 Client VPN Route (%s) create: %w", d.Id(), err)
	}

	return resourceClientVPNRouteRead(d, meta)
}

func resourceClientVPNRouteRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	endpointID, targetSubnetID, destinationCIDR, err := ClientVPNRouteParseResourceID(d.Id())

	if err != nil {
		return err
	}

	route, err := FindClientVPNRouteByThreePartKey(conn, endpointID, targetSubnetID, destinationCIDR)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 Client VPN Route (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading EC2 Client VPN Route (%s): %w", d.Id(), err)
	}

	d.Set("client_vpn_endpoint_id", route.ClientVpnEndpointId)
	d.Set("description", route.Description)
	d.Set("destination_cidr_block", route.DestinationCidr)
	d.Set("origin", route.Origin)
	d.Set("target_vpc_subnet_id", route.TargetSubnet)
	d.Set("type", route.Type)

	return nil
}

func resourceClientVPNRouteDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	endpointID, targetSubnetID, destinationCIDR, err := ClientVPNRouteParseResourceID(d.Id())

	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Deleting EC2 Client VPN Route: %s", d.Id())
	_, err = conn.DeleteClientVpnRoute(&ec2.DeleteClientVpnRouteInput{
		ClientVpnEndpointId:  aws.String(endpointID),
		DestinationCidrBlock: aws.String(destinationCIDR),
		TargetVpcSubnetId:    aws.String(targetSubnetID),
	})

	if tfawserr.ErrCodeEquals(err, ErrCodeInvalidClientVpnEndpointIdNotFound, ErrCodeInvalidClientVpnRouteNotFound) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting EC2 Client VPN Route (%s): %w", d.Id(), err)
	}

	if _, err := WaitClientVPNRouteDeleted(conn, endpointID, targetSubnetID, destinationCIDR, d.Timeout(schema.TimeoutDelete)); err != nil {
		return fmt.Errorf("error waiting for EC2 Client VPN Route (%s) delete: %w", d.Id(), err)
	}

	return nil
}

const clientVPNRouteIDSeparator = ","

func ClientVPNRouteCreateResourceID(endpointID, targetSubnetID, destinationCIDR string) string {
	parts := []string{endpointID, targetSubnetID, destinationCIDR}
	id := strings.Join(parts, clientVPNRouteIDSeparator)

	return id
}

func ClientVPNRouteParseResourceID(id string) (string, string, string, error) {
	parts := strings.Split(id, clientVPNRouteIDSeparator)

	if len(parts) == 3 && parts[0] != "" && parts[1] != "" && parts[2] != "" {
		return parts[0], parts[1], parts[2], nil
	}

	return "", "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected EndpointID%[2]sTargetSubnetID%[2]sDestinationCIDRBlock", id, clientVPNRouteIDSeparator)
}
