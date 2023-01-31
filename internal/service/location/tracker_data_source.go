package location

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/locationservice"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

func DataSourceTracker() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceTrackerRead,
		Schema: map[string]*schema.Schema{
			"create_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"kms_key_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"position_filtering": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tftags.TagsSchemaComputed(),
			"tracker_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tracker_name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 100),
			},
			"update_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceTrackerRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LocationConn()

	input := &locationservice.DescribeTrackerInput{
		TrackerName: aws.String(d.Get("tracker_name").(string)),
	}

	output, err := conn.DescribeTrackerWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "getting Location Service Tracker: %s", err)
	}

	if output == nil {
		return sdkdiag.AppendErrorf(diags, "getting Location Service Tracker: empty response")
	}

	d.SetId(aws.StringValue(output.TrackerName))
	d.Set("create_time", aws.TimeValue(output.CreateTime).Format(time.RFC3339))
	d.Set("description", output.Description)
	d.Set("kms_key_id", output.KmsKeyId)
	d.Set("position_filtering", output.PositionFiltering)
	d.Set("tags", KeyValueTags(output.Tags).IgnoreAWS().IgnoreConfig(meta.(*conns.AWSClient).IgnoreTagsConfig).Map())
	d.Set("tracker_arn", output.TrackerArn)
	d.Set("tracker_name", output.TrackerName)
	d.Set("update_time", aws.TimeValue(output.UpdateTime).Format(time.RFC3339))

	return diags
}
