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
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func ResourceClientVPNNetworkAssociation() *schema.Resource {
	return &schema.Resource{
		Create: resourceClientVPNNetworkAssociationCreate,
		Read:   resourceClientVPNNetworkAssociationRead,
		Update: resourceClientVPNNetworkAssociationUpdate,
		Delete: resourceClientVPNNetworkAssociationDelete,

		Importer: &schema.ResourceImporter{
			State: resourceClientVPNNetworkAssociationImport,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(ClientVPNNetworkAssociationCreatedTimeout),
			Delete: schema.DefaultTimeout(ClientVPNNetworkAssociationDeletedTimeout),
		},

		Schema: map[string]*schema.Schema{
			"association_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"client_vpn_endpoint_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"security_groups": {
				Type:       schema.TypeSet,
				MinItems:   1,
				MaxItems:   5,
				Optional:   true,
				Computed:   true,
				Elem:       &schema.Schema{Type: schema.TypeString},
				Set:        schema.HashString,
				Deprecated: "Use the `security_group_ids` attribute of the `aws_ec2_client_vpn_endpoint` resource instead.",
			},
			"status": {
				Type:       schema.TypeString,
				Computed:   true,
				Deprecated: `This attribute has been deprecated.`,
			},
			"subnet_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"vpc_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceClientVPNNetworkAssociationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	endpointID := d.Get("client_vpn_endpoint_id").(string)
	input := &ec2.AssociateClientVpnTargetNetworkInput{
		ClientVpnEndpointId: aws.String(endpointID),
		SubnetId:            aws.String(d.Get("subnet_id").(string)),
	}

	log.Printf("[DEBUG] Creating EC2 Client VPN Network Association: %s", input)

	output, err := conn.AssociateClientVpnTargetNetwork(input)

	if err != nil {
		return fmt.Errorf("error creating EC2 Client VPN Network Association: %w", err)
	}

	d.SetId(aws.StringValue(output.AssociationId))

	targetNetwork, err := WaitClientVPNNetworkAssociationCreated(conn, d.Id(), endpointID, d.Timeout(schema.TimeoutCreate))

	if err != nil {
		return fmt.Errorf("error waiting for EC2 Client VPN Network Association (%s) create: %w", d.Id(), err)
	}

	if v, ok := d.GetOk("security_groups"); ok {
		input := &ec2.ApplySecurityGroupsToClientVpnTargetNetworkInput{
			ClientVpnEndpointId: aws.String(endpointID),
			SecurityGroupIds:    flex.ExpandStringSet(v.(*schema.Set)),
			VpcId:               targetNetwork.VpcId,
		}

		_, err := conn.ApplySecurityGroupsToClientVpnTargetNetwork(input)

		if err != nil {
			return fmt.Errorf("error applying Security Groups to EC2 Client VPN Network Association (%s): %w", d.Id(), err)
		}
	}

	return resourceClientVPNNetworkAssociationRead(d, meta)
}

func resourceClientVPNNetworkAssociationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	endpointID := d.Get("client_vpn_endpoint_id").(string)
	network, err := FindClientVPNNetworkAssociationByIDs(conn, d.Id(), endpointID)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 Client VPN Network Association (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading EC2 Client VPN Network Association (%s): %w", d.Id(), err)
	}

	d.Set("association_id", network.AssociationId)
	d.Set("client_vpn_endpoint_id", network.ClientVpnEndpointId)
	d.Set("security_groups", aws.StringValueSlice(network.SecurityGroups))
	d.Set("status", network.Status.Code)
	d.Set("subnet_id", network.TargetNetworkId)
	d.Set("vpc_id", network.VpcId)

	return nil
}

func resourceClientVPNNetworkAssociationUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	if d.HasChange("security_groups") {
		input := &ec2.ApplySecurityGroupsToClientVpnTargetNetworkInput{
			ClientVpnEndpointId: aws.String(d.Get("client_vpn_endpoint_id").(string)),
			SecurityGroupIds:    flex.ExpandStringSet(d.Get("security_groups").(*schema.Set)),
			VpcId:               aws.String(d.Get("vpc_id").(string)),
		}

		if _, err := conn.ApplySecurityGroupsToClientVpnTargetNetwork(input); err != nil {
			return fmt.Errorf("error applying Security Groups to EC2 Client VPN Network Association (%s): %w", d.Id(), err)
		}
	}

	return resourceClientVPNNetworkAssociationRead(d, meta)
}

func resourceClientVPNNetworkAssociationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	endpointID := d.Get("client_vpn_endpoint_id").(string)

	log.Printf("[DEBUG] Deleting EC2 Client VPN Network Association: %s", d.Id())
	_, err := conn.DisassociateClientVpnTargetNetwork(&ec2.DisassociateClientVpnTargetNetworkInput{
		ClientVpnEndpointId: aws.String(endpointID),
		AssociationId:       aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidClientVPNAssociationIdNotFound, errCodeInvalidClientVPNEndpointIdNotFound) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error disassociating EC2 Client VPN Network Association (%s): %w", d.Id(), err)
	}

	if _, err := WaitClientVPNNetworkAssociationDeleted(conn, d.Id(), endpointID, d.Timeout(schema.TimeoutDelete)); err != nil {
		return fmt.Errorf("error waiting for EC2 Client VPN Network Association (%s) delete: %w", d.Id(), err)
	}

	return nil
}

func resourceClientVPNNetworkAssociationImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	parts := strings.Split(d.Id(), ",")

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return nil, fmt.Errorf("unexpected format for ID (%[1]s), expected EndpointID%[2]sAssociationID", d.Id(), ",")
	}

	d.SetId(parts[1])
	d.Set("client_vpn_endpoint_id", parts[0])

	return []*schema.ResourceData{d}, nil
}
