package logs

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceDestinationPolicy() *schema.Resource {
	return &schema.Resource{
		Create: resourceDestinationPolicyPut,
		Update: resourceDestinationPolicyPut,
		Read:   resourceDestinationPolicyRead,
		Delete: resourceDestinationPolicyDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"destination_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"access_policy": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateFunc:     validation.StringIsJSON,
				DiffSuppressFunc: verify.SuppressEquivalentPolicyDiffs,
				StateFunc: func(v interface{}) string {
					json, _ := structure.NormalizeJsonString(v)
					return json
				},
			},

			"force_update": {
				Type:     schema.TypeBool,
				Optional: true,
			},
		},
	}
}

func resourceDestinationPolicyPut(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LogsConn

	destination_name := d.Get("destination_name").(string)

	params := &cloudwatchlogs.PutDestinationPolicyInput{
		DestinationName: aws.String(destination_name),
		AccessPolicy:    aws.String(d.Get("access_policy").(string)),
	}

	if v, ok := d.GetOk("force_update"); ok {
		params.ForceUpdate = aws.Bool(v.(bool))
	}

	_, err := conn.PutDestinationPolicy(params)

	if err != nil {
		return fmt.Errorf("Error creating CloudWatch Log Destination Policy with destination_name %s: %#v", destination_name, err)
	}

	d.SetId(destination_name)
	return resourceDestinationPolicyRead(d, meta)
}

func resourceDestinationPolicyRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LogsConn
	destination, exists, err := LookupDestination(conn, d.Id(), nil)
	if err != nil {
		return err
	}

	if !exists || destination.AccessPolicy == nil {
		log.Printf("[WARN] CloudWatch Log Destination Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("access_policy", destination.AccessPolicy)
	d.Set("destination_name", destination.DestinationName)

	return nil
}

func resourceDestinationPolicyDelete(d *schema.ResourceData, meta interface{}) error {
	return nil
}
