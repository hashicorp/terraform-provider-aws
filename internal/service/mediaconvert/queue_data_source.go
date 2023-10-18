package mediaconvert

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/mediaconvert"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

// @SDKDataSource("aws_media_convert_queue", name="Queue")
func DataSourceQueue() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceQueueRead,

		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceQueueRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn, err := GetAccountClient(ctx, meta.(*conns.AWSClient))
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "error getting Media Convert Account Client: %s", err)
	}

	id := d.Get("id").(string)

	resp, err := conn.GetQueueWithContext(ctx, &mediaconvert.GetQueueInput{
		Name: aws.String(id),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "error getting Media Convert Queue (%s): %s", id, err)
	}

	if resp == nil || resp.Queue == nil {
		return sdkdiag.AppendErrorf(diags, "error getting Media Convert Queue (%s): empty response", id)
	}

	d.SetId(aws.StringValue(resp.Queue.Name))
	d.Set("name", resp.Queue.Name)
	d.Set("arn", resp.Queue.Arn)
	d.Set("status", resp.Queue.Status)

	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig
	tags, err := listTags(ctx, conn, aws.StringValue(resp.Queue.Arn))

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "error listing tags for Media Convert Queue (%s): %s", id, err)
	}

	if err := d.Set("tags", tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "error setting tags: %s", err)
	}

	return diags
}
