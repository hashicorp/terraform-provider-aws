package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/service/ssoadmin"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceAwsSsoInstance() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsSsoInstanceRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"identity_store_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceAwsSsoInstanceRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ssoadminconn

	log.Printf("[DEBUG] Reading AWS SSO Instances")
	resp, err := conn.ListInstances(&ssoadmin.ListInstancesInput{})
	if err != nil {
		return fmt.Errorf("Error getting AWS SSO Instances: %s", err)
	}

	if resp == nil || len(resp.Instances) == 0 {
		return fmt.Errorf("No AWS SSO Instance found")
	}

	instance := resp.Instances[0]
	log.Printf("[DEBUG] Received AWS SSO Instance: %s", instance)

	d.SetId(time.Now().UTC().String())
	d.Set("arn", instance.InstanceArn)
	d.Set("identity_store_id", instance.IdentityStoreId)

	return nil
}
