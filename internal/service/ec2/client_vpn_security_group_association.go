package ec2

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ResourceClientVPNSecurityGroupsAssociation() *schema.Resource {
	return &schema.Resource{
		Create: resourceClientVPNSecurityGroupsAssociationCreate,
		Read:   resourceClientVPNSecurityGroupsAssociationRead,
		Update: resourceClientVPNSecurityGroupsAssociationUpdate,
		Delete: resourceClientVPNSecurityGroupsAssociationDelete,

		Schema: map[string]*schema.Schema{
			"client_vpn_endpoint_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"security_groups": {
				Type:     schema.TypeSet,
				MinItems: 1,
				MaxItems: 5,
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"vpc_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceClientVPNSecurityGroupsAssociationCreate(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func resourceClientVPNSecurityGroupsAssociationRead(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func resourceClientVPNSecurityGroupsAssociationUpdate(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func resourceClientVPNSecurityGroupsAssociationDelete(d *schema.ResourceData, meta interface{}) error {
	return nil
}
