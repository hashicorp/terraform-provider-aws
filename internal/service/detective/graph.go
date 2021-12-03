package detective

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/detective"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func ResourceGraph() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceGraphCreate,
		ReadWithoutTimeout:   resourceGraphRead,
		DeleteWithoutTimeout: resourceGraphDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"created_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"graph_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags":     tftags.TagsSchemaForceNew(),
			"tags_all": tftags.TagsSchemaComputed(),
		},
	}
}

func resourceGraphCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).DetectiveConn

	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := &detective.CreateGraphInput{}

	if len(tags) > 0 {
		input.Tags = Tags(tags.IgnoreAWS())
	}

	var output *detective.CreateGraphOutput
	var err error
	err = resource.RetryContext(ctx, DetectiveOperationTimeout, func() *resource.RetryError {
		output, err = conn.CreateGraphWithContext(ctx, input)
		if err != nil {
			if tfawserr.ErrCodeEquals(err, detective.ErrCodeInternalServerException) {
				return resource.RetryableError(err)
			}

			return resource.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		output, err = conn.CreateGraphWithContext(ctx, input)
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating detective Graph: %w", err))
	}

	d.SetId(aws.StringValue(output.GraphArn))

	return resourceGraphRead(ctx, d, meta)
}

func resourceGraphRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).DetectiveConn

	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	resp, err := FindDetectiveGraphByArn(conn, ctx, d.Id())

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, detective.ErrCodeResourceNotFoundException) || resp == nil {
		d.SetId("")
		return nil
	}
	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading detective Graph (%s): %w", d.Id(), err))
	}

	d.Set("created_time", aws.TimeValue(resp.CreatedTime).Format(time.RFC3339))
	d.Set("graph_arn", aws.StringValue(resp.Arn))

	tg, err := conn.ListTagsForResource(&detective.ListTagsForResourceInput{
		ResourceArn: resp.Arn,
	})
	if err != nil {
		return diag.FromErr(fmt.Errorf("error listing tags for Detective Graph (%s): %w", d.Id(), err))
	}
	if tg.Tags == nil {
		log.Printf("[DEBUG] Detective Graph tags (%s) not found", d.Id())
		return nil
	}
	tags := KeyValueTags(tg.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	if err = d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return diag.FromErr(fmt.Errorf("error setting `%s` for Detective Graph (%s): %w", "tags", d.Id(), err))
	}

	if err = d.Set("tags_all", tags.Map()); err != nil {
		return diag.FromErr(fmt.Errorf("error setting `%s` for Detective Graph (%s): %w", "tags_all", d.Id(), err))
	}

	return nil
}

func resourceGraphDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).DetectiveConn

	input := &detective.DeleteGraphInput{
		GraphArn: aws.String(d.Id()),
	}

	_, err := conn.DeleteGraphWithContext(ctx, input)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, detective.ErrCodeResourceNotFoundException) {
			return nil
		}
		return diag.FromErr(fmt.Errorf("error deleting detective Graph (%s): %w", d.Id(), err))
	}

	return nil
}
