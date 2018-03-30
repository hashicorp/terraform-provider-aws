package aws

import (
	"fmt"
	"net"
	"time"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/directconnect"
)

func resourceAwsDxPrivateVif() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsDxPrivateVifCreate,
		Read:   resourceAwsDxPrivateVifRead,
		Update: resourceAwsDxPrivateVifUpdate,
		Delete: resourceAwsDxPrivateVifDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},

			"connection_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"address_family": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
					value := v.(string)
					if value != "ipv4" && value != "ipv6" {
						errors = append(errors, fmt.Errorf(
							"%q must be one of 'ipv4', 'ipv6'", k))
					}
					return
				},
			},

			"amazon_address": &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Computed:     true,
				ValidateFunc: validateCIDRAddress,
			},

			"asn": &schema.Schema{
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},

			"auth_key": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},

			"customer_address": &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Computed:     true,
				ValidateFunc: validateCIDRAddress,
			},

			"interface_name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"vlan": &schema.Schema{
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},

			"virtual_gateway_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"tags": tagsSchema(),
		},
	}
}

func resourceAwsDxPrivateVifGet(dxconn *directconnect.DirectConnect, id string) (*directconnect.VirtualInterface, error) {
	res, err := dxconn.DescribeVirtualInterfaces(&directconnect.DescribeVirtualInterfacesInput{
		VirtualInterfaceId: aws.String(id),
	})

	if err != nil {
		return nil, err
	}

	if len(res.VirtualInterfaces) == 0 {
		return nil, nil
	}

	vif := res.VirtualInterfaces[0]

	if *vif.VirtualInterfaceType != "private" {
		return nil, fmt.Errorf("Virtual interface %q is not private", id)
	}

	if *vif.VirtualInterfaceState == "deleted" {
		return nil, nil
	}

	return vif, nil
}

func resourceAwsDxPrivateVifCreate(d *schema.ResourceData, meta interface{}) error {
	var (
		vif *directconnect.VirtualInterface
		err error
	)

	dxconn := meta.(*AWSClient).dxconn

	vifSpec := &directconnect.NewPrivateVirtualInterface{
		AddressFamily:        aws.String(d.Get("address_family").(string)),
		Asn:                  aws.Int64(int64(d.Get("asn").(int))),
		VirtualInterfaceName: aws.String(d.Get("interface_name").(string)),
		Vlan:                 aws.Int64(int64(d.Get("vlan").(int))),
		VirtualGatewayId:     aws.String(d.Get("virtual_gateway_id").(string)),
	}

	if attr, ok := d.GetOk("amazon_address"); ok {
		vifSpec.AmazonAddress = aws.String(attr.(string))
	}

	if attr, ok := d.GetOk("auth_key"); ok {
		vifSpec.AuthKey = aws.String(attr.(string))
	}

	if attr, ok := d.GetOk("customer_address"); ok {
		vifSpec.CustomerAddress = aws.String(attr.(string))
	}

	vif, err = dxconn.CreatePrivateVirtualInterface(&directconnect.CreatePrivateVirtualInterfaceInput{
		ConnectionId: aws.String(d.Get("connection_id").(string)),

		NewPrivateVirtualInterface: vifSpec,
	})

	if err != nil {
		return err
	}

	vif, err = waitForAwsDxPrivateVif(dxconn, *vif.VirtualInterfaceId, []string{"available", "down"})

	if err != nil {
		return err
	}

	d.SetId(*vif.VirtualInterfaceId)

	return resourceAwsDxPrivateVifUpdate(d, meta)
}

func waitForAwsDxPrivateVif(dxconn *directconnect.DirectConnect, id string, target []string) (*directconnect.VirtualInterface, error) {
	wait := resource.StateChangeConf{
		Delay:      15 * time.Second,
		Pending:    []string{"pending"},
		Target:     target,
		Timeout:    30 * time.Minute,
		MinTimeout: 5 * time.Second,
		Refresh: func() (result interface{}, state string, err error) {
			vif, err := resourceAwsDxPrivateVifGet(dxconn, id)

			if err == nil && vif == nil {
				err = fmt.Errorf("Private virtual interface %q not found", id)
			}

			if err != nil {
				return nil, "UNKNOWN", err
			}

			return vif, *vif.VirtualInterfaceState, nil
		},
	}

	vif, err := wait.WaitForState()

	if err != nil {
		return nil, err
	}

	return vif.(*directconnect.VirtualInterface), err
}

func resourceAwsDxPrivateVifUpdate(d *schema.ResourceData, meta interface{}) error {
	dxconn := meta.(*AWSClient).dxconn

	arn := arn.ARN{
		Partition: meta.(*AWSClient).partition,
		Region:    meta.(*AWSClient).region,
		Service:   "directconnect",
		AccountID: meta.(*AWSClient).accountid,
		Resource:  fmt.Sprintf("dxvif/%s", d.Id()),
	}.String()

	if err := setTagsDX(dxconn, d, arn); err != nil {
		return err
	}

	return resourceAwsDxPrivateVifRead(d, meta)
}

func resourceAwsDxPrivateVifRead(d *schema.ResourceData, meta interface{}) error {
	dxconn := meta.(*AWSClient).dxconn

	vif, err := resourceAwsDxPrivateVifGet(dxconn, d.Id())

	if err != nil {
		return err
	}

	if vif == nil {
		d.SetId("")
		return nil
	}

	if *vif.OwnerAccount != meta.(*AWSClient).accountid {
		return fmt.Errorf("Private virtual interface does not belong to current account")
	}

	d.Set("connection_id", *vif.ConnectionId)
	d.Set("address_family", *vif.AddressFamily)
	d.Set("asn", *vif.Asn)
	d.Set("interface_name", *vif.VirtualInterfaceName)
	d.Set("vlan", *vif.Vlan)

	if vif.VirtualGatewayId != nil {
		d.Set("virtual_gateway_id", *vif.VirtualGatewayId)
	}

	if vif.AmazonAddress != nil {
		d.Set("amazon_address", *vif.AmazonAddress)
	}

	if vif.AuthKey != nil {
		d.Set("auth_key", *vif.AuthKey)
	}

	if vif.CustomerAddress != nil {
		d.Set("customer_address", *vif.CustomerAddress)
	}

	arn := arn.ARN{
		Partition: meta.(*AWSClient).partition,
		Region:    meta.(*AWSClient).region,
		Service:   "directconnect",
		AccountID: meta.(*AWSClient).accountid,
		Resource:  fmt.Sprintf("dxvif/%s", d.Id()),
	}.String()

	d.Set("arn", arn)

	if err := getTagsDX(dxconn, d, arn); err != nil {
		return err
	}

	return nil
}

func resourceAwsDxPrivateVifDelete(d *schema.ResourceData, meta interface{}) error {
	dxconn := meta.(*AWSClient).dxconn

	id := d.Id()

	input := &directconnect.DeleteVirtualInterfaceInput{
		VirtualInterfaceId: aws.String(id),
	}

	_, err := dxconn.DeleteVirtualInterface(input)

	if err != nil {
		return err
	}

	wait := resource.StateChangeConf{
		Delay:      15 * time.Second,
		Pending:    []string{"deleting"},
		Target:     []string{"deleted"},
		Timeout:    30 * time.Minute,
		MinTimeout: 5 * time.Second,
		Refresh: func() (result interface{}, state string, err error) {
			vif, err := resourceAwsDxPrivateVifGet(dxconn, id)

			if err != nil {
				return nil, "UNKNOWN", err
			}

			if vif == nil {
				return &directconnect.VirtualInterface{}, "deleted", nil
			}

			return vif, *vif.VirtualInterfaceState, nil
		},
	}

	if _, err = wait.WaitForState(); err != nil {
		return err
	}

	d.SetId("")

	return nil
}

func validateCIDRAddress(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)
	_, _, err := net.ParseCIDR(value)
	if err != nil {
		errors = append(errors, fmt.Errorf(
			"%q must contain a valid CIDR, got error parsing: %s", k, err))
	}
	return
}
