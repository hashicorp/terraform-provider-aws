package logs

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func init() {
	_sp.registerSDKResourceFactory("aws_cloudwatch_log_destination_policy", resourceDestinationPolicy)
}

func resourceDestinationPolicy() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceDestinationPolicyPut,
		ReadWithoutTimeout:   resourceDestinationPolicyRead,
		UpdateWithoutTimeout: resourceDestinationPolicyPut,
		DeleteWithoutTimeout: schema.NoopContext,

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

func resourceDestinationPolicyPut(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).LogsConn()

	name := d.Get("destination_name").(string)
	input := &cloudwatchlogs.PutDestinationPolicyInput{
		AccessPolicy:    aws.String(d.Get("access_policy").(string)),
		DestinationName: aws.String(name),
	}

	if v, ok := d.GetOk("force_update"); ok {
		input.ForceUpdate = aws.Bool(v.(bool))
	}

	_, err := conn.PutDestinationPolicyWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("putting CloudWatch Logs Destination Policy (%s): %s", name, err)
	}

	if d.IsNewResource() {
		d.SetId(name)
	}

	return resourceDestinationPolicyRead(ctx, d, meta)
}

func resourceDestinationPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).LogsConn()

	destination, err := FindDestinationByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] CloudWatch Logs Destination Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading CloudWatch Logs Destination Policy (%s): %s", d.Id(), err)
	}

	d.Set("access_policy", destination.AccessPolicy)
	d.Set("destination_name", destination.DestinationName)

	return nil
}
