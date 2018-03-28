package aws

import (
	"fmt"
	"time"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directconnect"
)

func resourceAwsDxPrivateVirtualInterface() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsDxPrivateVirtualInterfaceCreate,
		Read:   resourceAwsDxPrivateVirtualInterfaceRead,
		Delete: resourceAwsDxPrivateVirtualInterfaceDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"connection_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"owner_account": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"address_family": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"amazon_address": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
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
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
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

			"owner": &schema.Schema{
				Type:     schema.TypeBool,
				Computed: true,
			},

			"virtual_gateway_id": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,

				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					// this field is ignored, when owner of the interface
					// is not the same, as connection owner
					owner, ok := d.GetOkExists("owner")
					return ok && !owner.(bool)
				},
			},
		},
	}
}

func resourceAwsDxPrivateVirtualInterfaceGet(dxconn *directconnect.DirectConnect, id string) (*directconnect.VirtualInterface, error) {
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

func resourceAwsDxPrivateVirtualInterfaceCreate(d *schema.ResourceData, meta interface{}) error {
	var (
		vif              *directconnect.VirtualInterface
		err              error
		ownerAccount     string
		virtualGatewayId string
	)

	dxconn := meta.(*AWSClient).dxconn

	ownerAccount = d.Get("owner_account").(string)

	if ownerAccount == meta.(*AWSClient).accountid {
		if attr, ok := d.GetOk("virtual_gateway_id"); ok {
			virtualGatewayId = attr.(string)
		} else {
			return fmt.Errorf("virtual_gateway_id is required if connection and interface owners are same")
		}

		vifSpec := &directconnect.NewPrivateVirtualInterface{
			AddressFamily:        aws.String(d.Get("address_family").(string)),
			Asn:                  aws.Int64(int64(d.Get("asn").(int))),
			VirtualInterfaceName: aws.String(d.Get("interface_name").(string)),
			Vlan:                 aws.Int64(int64(d.Get("vlan").(int))),
			VirtualGatewayId:     aws.String(virtualGatewayId),
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
	} else {
		vifSpec := &directconnect.NewPrivateVirtualInterfaceAllocation{
			AddressFamily:        aws.String(d.Get("address_family").(string)),
			Asn:                  aws.Int64(int64(d.Get("asn").(int))),
			VirtualInterfaceName: aws.String(d.Get("interface_name").(string)),
			Vlan:                 aws.Int64(int64(d.Get("vlan").(int))),
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

		vif, err = dxconn.AllocatePrivateVirtualInterface(&directconnect.AllocatePrivateVirtualInterfaceInput{
			ConnectionId: aws.String(d.Get("connection_id").(string)),
			OwnerAccount: aws.String(ownerAccount),

			NewPrivateVirtualInterfaceAllocation: vifSpec,
		})
	}

	if err != nil {
		return err
	}

	vif, err = waitForAwsDxPrivateVirtualInterface(dxconn, *vif.VirtualInterfaceId, []string{"confirming", "available", "down"})

	if err != nil {
		return err
	}

	d.SetId(*vif.VirtualInterfaceId)

	return resourceAwsDxPrivateVirtualInterfaceUpdate(vif, d, meta.(*AWSClient).accountid)
}

func waitForAwsDxPrivateVirtualInterface(dxconn *directconnect.DirectConnect, id string, target []string) (*directconnect.VirtualInterface, error) {
	wait := resource.StateChangeConf{
		Delay:      15 * time.Second,
		Pending:    []string{"pending"},
		Target:     target,
		Timeout:    30 * time.Minute,
		MinTimeout: 5 * time.Second,
		Refresh: func() (result interface{}, state string, err error) {
			vif, err := resourceAwsDxPrivateVirtualInterfaceGet(dxconn, id)

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

func resourceAwsDxPrivateVirtualInterfaceUpdate(vif *directconnect.VirtualInterface, d *schema.ResourceData, accountid string) error {
	d.Set("connection_id", *vif.ConnectionId)
	d.Set("address_family", *vif.AddressFamily)
	d.Set("asn", *vif.Asn)
	d.Set("interface_name", *vif.VirtualInterfaceName)
	d.Set("vlan", *vif.Vlan)
	d.Set("owner_account", *vif.OwnerAccount)

	d.Set("owner", *vif.OwnerAccount == accountid)

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

	return nil
}

func resourceAwsDxPrivateVirtualInterfaceRead(d *schema.ResourceData, meta interface{}) error {
	dxconn := meta.(*AWSClient).dxconn

	vif, err := resourceAwsDxPrivateVirtualInterfaceGet(dxconn, d.Id())

	if err != nil {
		return err
	}

	if vif == nil {
		d.SetId("")
		return nil
	}

	return resourceAwsDxPrivateVirtualInterfaceUpdate(vif, d, meta.(*AWSClient).accountid)
}

func resourceAwsDxPrivateVirtualInterfaceDelete(d *schema.ResourceData, meta interface{}) error {
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
			vif, err := resourceAwsDxPrivateVirtualInterfaceGet(dxconn, id)

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
