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

func DataSourceSource() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceSourceRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name_prefix": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_by": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"state": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceSourceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EventsConn()

	input := &eventbridge.ListEventSourcesInput{}
	if v, ok := d.GetOk("name_prefix"); ok {
		input.NamePrefix = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Listing EventBridge sources: %s", input)

	resp, err := conn.ListEventSourcesWithContext(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing EventBridge sources: %s", err)
	}

	if resp == nil || len(resp.EventSources) == 0 {
		return sdkdiag.AppendErrorf(diags, "no matching partner event source")
	}
	if len(resp.EventSources) > 1 {
		return sdkdiag.AppendErrorf(diags, "multiple event sources matched; use additional constraints to reduce matches to a single event source")
	}

	es := resp.EventSources[0]

	d.SetId(aws.StringValue(es.Name))
	d.Set("arn", es.Arn)
	d.Set("created_by", es.CreatedBy)
	d.Set("name", es.Name)
	d.Set("state", es.State)

	return diags
}
