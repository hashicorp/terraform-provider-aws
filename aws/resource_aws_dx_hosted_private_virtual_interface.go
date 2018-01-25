package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directconnect"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsDxHostedPrivateVirtualInterface() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsDxHostedPrivateVirtualInterfaceCreate,
		Read:   resourceAwsDxHostedPrivateVirtualInterfaceRead,
		Delete: resourceAwsDxHostedPrivateVirtualInterfaceDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: mergeSchemas(
			dxVirtualInterfaceSchema,
			map[string]*schema.Schema{
				"owner_account_id": {
					Type:         schema.TypeString,
					Required:     true,
					ForceNew:     true,
					ValidateFunc: validateAwsAccountId,
				},
			},
		),

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
	}
}

func resourceAwsDxHostedPrivateVirtualInterfaceCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).dxconn

	req := &directconnect.AllocatePrivateVirtualInterfaceInput{
		ConnectionId: aws.String(d.Get("connection_id").(string)),
		OwnerAccount: aws.String(d.Get("owner_account_id").(string)),
		NewPrivateVirtualInterfaceAllocation: &directconnect.NewPrivateVirtualInterfaceAllocation{
			VirtualInterfaceName: aws.String(d.Get("name").(string)),
			Vlan:                 aws.Int64(int64(d.Get("vlan").(int))),
			Asn:                  aws.Int64(int64(d.Get("bgp_asn").(int))),
			AddressFamily:        aws.String(d.Get("address_family").(string)),
		},
	}
	if v, ok := d.GetOk("bgp_auth_key"); ok {
		req.NewPrivateVirtualInterfaceAllocation.AuthKey = aws.String(v.(string))
	}
	if v, ok := d.GetOk("customer_address"); ok {
		req.NewPrivateVirtualInterfaceAllocation.CustomerAddress = aws.String(v.(string))
	}
	if v, ok := d.GetOk("amazon_address"); ok {
		req.NewPrivateVirtualInterfaceAllocation.AmazonAddress = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating Direct Connect hosted private virtual interface: %#v", req)
	resp, err := conn.AllocatePrivateVirtualInterface(req)
	if err != nil {
		return fmt.Errorf("Error creating Direct Connect hosted private virtual interface: %s", err.Error())
	}

	d.SetId(aws.StringValue(resp.VirtualInterfaceId))

	if err := dxHostedPrivateVirtualInterfaceWaitUntilAvailable(d, conn); err != nil {
		return err
	}

	return resourceAwsDxHostedPrivateVirtualInterfaceRead(d, meta)
}

func resourceAwsDxHostedPrivateVirtualInterfaceRead(d *schema.ResourceData, meta interface{}) error {
	vif, err := dxVirtualInterfaceRead(d, meta)
	if err != nil {
		return err
	}
	if vif == nil {
		return nil
	}

	d.Set("owner_account_id", vif.OwnerAccount)
	return dxPrivateVirtualInterfaceAttributes(d, meta, vif)
}

func resourceAwsDxHostedPrivateVirtualInterfaceDelete(d *schema.ResourceData, meta interface{}) error {
	return dxVirtualInterfaceDelete(d, meta)
}

func dxHostedPrivateVirtualInterfaceWaitUntilAvailable(d *schema.ResourceData, conn *directconnect.DirectConnect) error {
	return dxVirtualInterfaceWaitUntilAvailable(
		d,
		conn,
		[]string{
			directconnect.VirtualInterfaceStatePending,
		},
		[]string{
			directconnect.VirtualInterfaceStateAvailable,
			directconnect.VirtualInterfaceStateConfirming,
			directconnect.VirtualInterfaceStateDown,
		})
}
