package aws

import (
	"fmt"
	"time"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/directconnect"
)

func resourceAwsDxPrivateVifAllocation() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsDxPrivateVifAllocationCreate,
		Read:   resourceAwsDxPrivateVifAllocationRead,
		Delete: resourceAwsDxPrivateVifAllocationDelete,

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

			"owner_account": &schema.Schema{
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateAwsAccountId,
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
		},
	}
}

func resourceAwsDxPrivateVifAllocationCreate(d *schema.ResourceData, meta interface{}) error {
	var (
		vif *directconnect.VirtualInterface
		err error
	)

	dxconn := meta.(*AWSClient).dxconn

	if d.Get("owner_account").(string) == meta.(*AWSClient).accountid {
		return fmt.Errorf("'owner_account' and connection account cannot be same for vif allocation")
	}

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
		OwnerAccount: aws.String(d.Get("owner_account").(string)),

		NewPrivateVirtualInterfaceAllocation: vifSpec,
	})

	if err != nil {
		return err
	}

	vif, err = waitForAwsDxPrivateVif(dxconn, *vif.VirtualInterfaceId, []string{"confirming", "available", "down"})

	if err != nil {
		return err
	}

	d.SetId(*vif.VirtualInterfaceId)

	return resourceAwsDxPrivateVifAllocationRead(d, meta)
}

func resourceAwsDxPrivateVifAllocationRead(d *schema.ResourceData, meta interface{}) error {
	dxconn := meta.(*AWSClient).dxconn

	vif, err := resourceAwsDxPrivateVifGet(dxconn, d.Id())

	if err != nil {
		return err
	}

	if vif == nil {
		d.SetId("")
		return nil
	}

	if *vif.OwnerAccount == meta.(*AWSClient).accountid {
		return fmt.Errorf("'owner_account' and connection account cannot be same for vif allocation")
	}

	d.Set("connection_id", *vif.ConnectionId)
	d.Set("address_family", *vif.AddressFamily)
	d.Set("asn", *vif.Asn)
	d.Set("interface_name", *vif.VirtualInterfaceName)
	d.Set("vlan", *vif.Vlan)
	d.Set("owner_account", *vif.OwnerAccount)

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
		AccountID: *vif.OwnerAccount,
		Resource:  fmt.Sprintf("dxvif/%s", d.Id()),
	}.String()

	d.Set("arn", arn)

	return nil
}

func resourceAwsDxPrivateVifAllocationDelete(d *schema.ResourceData, meta interface{}) error {
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
