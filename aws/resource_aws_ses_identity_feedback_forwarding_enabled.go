package aws

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceAwsSesIdentityFeedbackForwardingEnabled() *schema.Resource {
	return &schema.Resource{
		//Create: resourceAwsSesIdentityFeedbackForwardingEnabledCreate,
		//Read:   resourceAwsSesIdentityFeedbackForwardingEnabledRead,
		//Delete: resourceAwsSesIdentityFeedbackForwardingEnabledDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"identity": {
				Type:     schema.TypeString,
				Required: true,
			},
			"enabled": {
				Type:     schema.TypeBool,
				Required: true,
			},
		},
	}
}

//func resourceAwsSesIdentityFeedbackForwardingEnabledCreate(d *schema.ResourceData, meta interface{}) error {
//}
//
//func resourceAwsSesIdentityFeedbackForwardingEnabledRead(d *schema.ResourceData, meta interface{}) error {
//}
//
//func resourceAwsSesIdentityFeedbackForwardingEnabledDelete(d *schema.ResourceData, meta interface{}) error {
//
//}
