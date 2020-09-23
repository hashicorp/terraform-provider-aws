package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssoadmin"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsSsoPermissionSet() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsSsoPermissionSetCreate,
		Read:   resourceAwsSsoPermissionSetRead,
		Update: resourceAwsSsoPermissionSetUpdate,
		Delete: resourceAwsSsoPermissionSetDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"instance_arn": {
				Type: schema.TypeString,
			},
			
			"permission_set_arn": {
				Type: schema.TypeString,
			},

			"created_date": {
				Type: schema.TypeString,
			},

			"description": {
				Type: schema.TypeString,
			},

			"name": {
				Type: schema.TypeString,
			},

			"relay_state": {
				Type: schema.TypeString,
			},

			"session_duration": {
				Type: schema.TypeString,
			},

			"inline_policy": {
				Type: schema.TypeString,
			},

			"managed_policies": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString
				},
			},

			"tags": tagsSchema(),
		},
	}
}

func resourceAwsSsoPermissionSetCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ssoadminconn
	// TODO
	// d.SetId(*resp.PermissionSetArn)
	return resourceAwsSsoPermissionSetRead(d, meta)
}

func resourceAwsSsoPermissionSetRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ssoadminconn
	// TODO
	return nil
}

func resourceAwsSsoPermissionSetUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ssoadminconn
	// TODO
	return resourceAwsSsoPermissionSetRead(d, meta)
}

func resourceAwsSsoPermissionSetDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ssoadminconn
	// TODO
	return nil
}

func waitForPermissionSetProvisioning(conn *identitystore.IdentityStore, arn string) error {

}
