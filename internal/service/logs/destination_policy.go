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
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceDestinationPolicy() *schema.Resource {
	return &schema.Resource{
		Create: resourceDestinationPolicyPut,
		Read:   resourceDestinationPolicyRead,
		Update: resourceDestinationPolicyPut,
		Delete: schema.Noop,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
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
			"destination_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
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

	name := d.Get("destination_name").(string)
	input := &cloudwatchlogs.PutDestinationPolicyInput{
		AccessPolicy:    aws.String(d.Get("access_policy").(string)),
		DestinationName: aws.String(name),
	}

	if v, ok := d.GetOk("force_update"); ok {
		input.ForceUpdate = aws.Bool(v.(bool))
	}

	_, err := conn.PutDestinationPolicy(input)

	if err != nil {
		return fmt.Errorf("putting CloudWatch Logs Destination Policy (%s): %w", name, err)
	}

	if d.IsNewResource() {
		d.SetId(name)
	}

	return resourceDestinationPolicyRead(d, meta)
}

func resourceDestinationPolicyRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LogsConn

	destination, err := FindDestinationByName(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] CloudWatch Logs Destination Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("reading CloudWatch Logs Destination Policy (%s): %w", d.Id(), err)
	}

	d.Set("access_policy", destination.AccessPolicy)
	d.Set("destination_name", destination.DestinationName)

	return nil
}
