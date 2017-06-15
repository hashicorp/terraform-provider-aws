package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directconnect"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsDxVirtualInterfaceConfirm() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsDxVirtualInterfaceConfirmCreate,
		Read:   resourceAwsDxVirtualInterfaceConfirmRead,
		Delete: resourceAwsDxVirtualInterfaceConfirmDelete,

		Importer: &schema.ResourceImporter{
			State: func(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				d.Set("allow_down_state", false)
				d.Set("virtual_interface_id", d.Id())
				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"virtual_interface_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"interface_type": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"virtual_gateway_id": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},

			"allow_down_state": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
				Default:  false,
			},

			"connection_id": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"asn": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"virtual_interface_name": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"vlan": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"amazon_address": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"customer_address": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"owner_account_id": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"auth_key": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"route_filter_prefixes": &schema.Schema{
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
		},
	}
}

func resourceAwsDxVirtualInterfaceConfirmCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).dxconn

	if v, ok := d.GetOk("interface_type"); ok && v.(string) == "public" {
		params := &directconnect.ConfirmPublicVirtualInterfaceInput{
			VirtualInterfaceId: aws.String(d.Get("virtual_interface_id").(string)),
		}
		_, err := conn.ConfirmPublicVirtualInterface(params)
		if err != nil {
			return fmt.Errorf("Error creating DirectConnect Virtual Interface: %s", err)
		}
	} else {
		params := &directconnect.ConfirmPrivateVirtualInterfaceInput{
			VirtualInterfaceId: aws.String(d.Get("virtual_interface_id").(string)),
		}
		if v, ok := d.GetOk("virtual_gateway_id"); ok {
			params.VirtualGatewayId = aws.String(v.(string))
		} else {
			return fmt.Errorf("Error virtual_gateway_id is required for private Virtual Interface with id: %s", d.Get("virtual_interface_id").(string))
		}
		_, err := conn.ConfirmPrivateVirtualInterface(params)
		if err != nil {
			return fmt.Errorf("Error creating DirectConnect Virtual Interface: %s", err)
		}
	}

	stateConf := &resource.StateChangeConf{
		Pending:    []string{"pending", "verifying", "down"},
		Target:     []string{"available"},
		Refresh:    DxVirtualInterfaceRefreshFunc(conn, d.Get("virtual_interface_id").(string)),
		Timeout:    30 * time.Minute,
		Delay:      10 * time.Second,
		MinTimeout: 10 * time.Second,
	}

	if v, ok := d.GetOk("allow_down_state"); ok && v.(bool) {
		stateConf.Pending = []string{"pending", "verifying"}
		stateConf.Target = []string{"available", "down"}
	}

	_, stateErr := stateConf.WaitForState()
	if stateErr != nil {
		return fmt.Errorf(
			"Error waiting for DirectConnect Virtual Interface (%s) to become ready: %s",
			d.Get("virtual_interface_id"), stateErr)
	}

	d.SetId(d.Get("virtual_interface_id").(string))

	return resourceAwsDxVirtualInterfaceConfirmRead(d, meta)
}

func DxVirtualInterfaceRefreshFunc(conn *directconnect.DirectConnect, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		resp, err := conn.DescribeVirtualInterfaces(&directconnect.DescribeVirtualInterfacesInput{
			VirtualInterfaceId: aws.String(id),
		})

		if err != nil {
			log.Printf("Error on DxVirtualInterfaceRefresh: %s", err)
			return nil, "", err
		}

		if resp == nil || len(resp.VirtualInterfaces) == 0 {
			return nil, "", nil
		}

		virtualInterface := resp.VirtualInterfaces[0]
		return virtualInterface, *virtualInterface.VirtualInterfaceState, nil
	}
}

func resourceAwsDxVirtualInterfaceConfirmRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).dxconn

	resp, err := conn.DescribeVirtualInterfaces(&directconnect.DescribeVirtualInterfacesInput{
		VirtualInterfaceId: aws.String(d.Id()),
	})

	if err != nil {
		return fmt.Errorf("Error reading DirectConnect Virtual Interface: %s", err)
	}

	if len(resp.VirtualInterfaces) > 1 {
		return fmt.Errorf("More than one DirectConnect Virtual Interface returned")
	}

	if len(resp.VirtualInterfaces) == 0 {
		d.SetId("")
		return nil
	}

	virtualInterface := resp.VirtualInterfaces[0]

	d.Set("interface_type", virtualInterface.VirtualInterfaceType)
	d.Set("virtual_gateway_id", virtualInterface.VirtualGatewayId)
	d.Set("connection_id", virtualInterface.ConnectionId)
	d.Set("asn", virtualInterface.Asn)
	d.Set("virtual_interface_name", virtualInterface.VirtualInterfaceName)
	d.Set("vlan", virtualInterface.Vlan)
	d.Set("amazon_address", virtualInterface.AmazonAddress)
	d.Set("customer_address", virtualInterface.CustomerAddress)
	d.Set("owner_account_id", virtualInterface.OwnerAccount)
	d.Set("auth_key", virtualInterface.AuthKey)

	set := &schema.Set{F: schema.HashString}
	for _, val := range virtualInterface.RouteFilterPrefixes {
		set.Add(*val.Cidr)
	}
	d.Set("route_filter_prefixes", set)

	d.SetId(*virtualInterface.VirtualInterfaceId)

	return nil
}

func resourceAwsDxVirtualInterfaceConfirmDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).dxconn

	_, err := conn.DeleteVirtualInterface(&directconnect.DeleteVirtualInterfaceInput{
		VirtualInterfaceId: aws.String(d.Id()),
	})

	if err != nil {
		return fmt.Errorf("Error deleting DirectConnect Virtual Interface: %s", err)
	}

	stateConf := &resource.StateChangeConf{
		Pending:    []string{"deleting"},
		Target:     []string{"deleted"},
		Refresh:    DxVirtualInterfaceRefreshFunc(conn, d.Id()),
		Timeout:    10 * time.Minute,
		Delay:      10 * time.Second,
		MinTimeout: 10 * time.Second,
	}

	_, stateErr := stateConf.WaitForState()
	if stateErr != nil {
		return fmt.Errorf(
			"Error waiting for DirectConnect Virtual Interface (%s) to be deleted: %s",
			d.Get("virtual_interface_id"), stateErr)
	}

	return nil
}
