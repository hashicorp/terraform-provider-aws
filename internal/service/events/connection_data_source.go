package events

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/eventbridge"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

func DataSourceConnection() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceConnectionRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"authorization_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"secret_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceConnectionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	d.SetId(d.Get("name").(string))

	conn := meta.(*conns.AWSClient).EventsConn()

	input := &eventbridge.DescribeConnectionInput{
		Name: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Reading EventBridge connection (%s)", d.Id())
	output, err := conn.DescribeConnectionWithContext(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "getting EventBridge connection (%s): %s", d.Id(), err)
	}

	if output == nil {
		return sdkdiag.AppendErrorf(diags, "getting EventBridge connection (%s): empty response", d.Id())
	}

	log.Printf("[DEBUG] Found EventBridge connection: %#v", *output)
	d.Set("arn", output.ConnectionArn)
	d.Set("secret_arn", output.SecretArn)
	d.Set("name", output.Name)
	d.Set("authorization_type", output.AuthorizationType)
	return diags
}
