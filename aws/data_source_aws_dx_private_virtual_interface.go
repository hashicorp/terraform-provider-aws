package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/directconnect"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
)

func dataSourceAwsDxPrivateVirtualInterface() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsDxPrivateVirtualInterfaceRead,

		Schema: map[string]*schema.Schema{
			"virtual_interface_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"connection_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"vpn_gateway_id": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"dx_gateway_id"},
			},
			"dx_gateway_id": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"vpn_gateway_id"},
			},
			"vlan": {
				Type:         schema.TypeInt,
				Computed:     true,
				ValidateFunc: validation.IntBetween(1, 4094),
			},
			"bgp_asn": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"bgp_auth_key": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"address_family": {
				Type:         schema.TypeString,
				Computed:     true,
				ValidateFunc: validation.StringInSlice([]string{directconnect.AddressFamilyIpv4, directconnect.AddressFamilyIpv6}, false),
			},
			"customer_address": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"amazon_address": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"mtu": {
				Type:         schema.TypeInt,
				Default:      1500,
				Computed:     true,
				ValidateFunc: validateIntegerInSlice([]int{1500, 9001}),
			},
			"jumbo_frame_capable": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"tags": tagsSchema(),
		},
	}
}

func dataSourceAwsDxPrivateVirtualInterfaceRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).dxconn

	vif, err := dxVirtualInterfaceRead(d.Get("virtual_interface_id").(string), conn)
	if err != nil {
		return err
	}
	if vif == nil {
		log.Printf("[WARN] Direct Connect virtual interface (%s) not found", d.Get("virtual_interface_id").(string))
		return nil
	}

	arn := arn.ARN{
		Partition: meta.(*AWSClient).partition,
		Region:    meta.(*AWSClient).region,
		Service:   "directconnect",
		AccountID: meta.(*AWSClient).accountid,
		Resource:  fmt.Sprintf("dxvif/%s", d.Get("virtual_interface_id")),
	}.String()

	d.Set("arn", arn)
	d.Set("connection_id", vif.ConnectionId)
	d.Set("name", vif.VirtualInterfaceName)
	d.Set("vlan", vif.Vlan)
	d.Set("bgp_asn", vif.Asn)
	d.Set("bgp_auth_key", vif.AuthKey)
	d.Set("address_family", vif.AddressFamily)
	d.Set("customer_address", vif.CustomerAddress)
	d.Set("amazon_address", vif.AmazonAddress)
	d.Set("vpn_gateway_id", vif.VirtualGatewayId)
	d.Set("dx_gateway_id", vif.DirectConnectGatewayId)
	d.Set("mtu", vif.Mtu)
	d.Set("jumbo_frame_capable", vif.JumboFrameCapable)

	err1 := getTagsDX(conn, d, arn)

	return err1
}
