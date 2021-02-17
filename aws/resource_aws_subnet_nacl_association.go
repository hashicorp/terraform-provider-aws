package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceAwsSubnetNaclAssociation() *schema.Resource {
	//lintignore:R011
	return &schema.Resource{
		Create: resourceAwsSubnetNaclAssociationCreate,
		Read:   resourceAwsSubnetNaclAssociationRead,
		Update: resourceAwsSubnetNaclAssociationUpdate,
		Delete: resourceAwsSubnetNaclAssociationDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(20 * time.Minute),
		},

		SchemaVersion: 1,

		Schema: map[string]*schema.Schema{
			"subnet_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: false,
			},

			"network_acl_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: false,
			},
		},
	}
}

func resourceAwsSubnetNaclAssociationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	// find all current nacls with subnet id
	subnetId := d.Get("subnet_id").(string)
	existingAssociation := findNetworkAclAssociation(subnetId, conn)

	log.Printf(existingAssociation.(string))
	createOpts := &ec2.ReplaceNetworkAclAssociationInput{
		AssociationId: aws.String(existingAssociation.NetworkAclAssociationId.(string)),
		NetworkAclId:  aws.String(d.Get("network_acl_id").(string)),
	}

	var err error
	resp, err := conn.ReplaceNetworkAclAssociation(createOpts)

	if err != nil {
		return fmt.Errorf("error replacing subnet network acl association: %w", err)
	}

	// Get the ID and store it
	associationId := aws.StringValue(resp.NewAssociationId)
	d.SetId(associationId.(string))
	log.Printf("[INFO] New Association ID: %s", associationId)

	_, err = stateConf.WaitForState()

	if err != nil {
		return fmt.Errorf("error waiting for subnet (%s) to become ready: %w", d.Id(), err)
	}

	return resourceAwsSubnetNaclAssociationRead(d, meta)
}

func resourceAwsSubnetNaclAssociationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	subnetId := d.Get("subnet_id").(string)
	existingAssociation, err := findNetworkAclAssociation(subnetId, conn)

	if err != nil {
		if isAWSErr(err, "InvalidNetworkAclSubnetAssociation.NotFound", "") {
			log.Printf("[WARN] Network Acl Association for Subnet Id (%s) not found, removing from state", d.Get("network_acl_id"), d.Get("subnet_id"))
			d.SetId("")
			return nil
		}
		return err
	}

	if d.Get("subnet_id") == aws.StringValue(existingAssociation.SubnetId) {
		d.SetId(aws.StringValue(existingAssociation.NetworkAclAssociationId))
		d.Set("network_acl_id", aws.StringValue(existingAssociation.NetworkAclId))
		return nil
	}

	return nil
}

func resourceAwsSubnetNaclAssociationUpdate(d *schema.ResourceData, meta interface{}) error {
	return resourceAwsSubnetNaclAssociationCreate(d, meta)
}

func resourceAwsSubnetNaclAssociationDelete(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[WARN] Cannot destroy Network ACL Association. Terraform will remove this resource from the state file, however resources may remain.")
	return nil
}
