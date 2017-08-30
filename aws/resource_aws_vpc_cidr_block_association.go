package aws

import "github.com/hashicorp/terraform/helper/schema"

func resourceAwsCidrBlockAssociation() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsCidrBlockAssociationCreate,
    Read: resourceAwsCidrBlockAssociationRead,
    Update: resourceAwsCidrBlockAssociationUpdate,
    Delete: resourceAwsCidrBlockAssociationDelete,

    Schema: map[string]*schema.Schema{
      "vpc_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
      "cidr_block": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
    }
	}
}

func resourceAwsCidrBlockAssociationCreate(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func resourceAwsCidrBlockAssociationRead(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func resourceAwsCidrBlockAssociationUpdate(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func resourceAwsCidrBlockAssociationDelete(d *schema.ResourceData, meta interface{}) error {
	return nil
}
