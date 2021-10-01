package aws

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
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/detective/waiter"
)

func resourceAwsDetectiveGraph() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceDetectiveAccountCreate,
		ReadWithoutTimeout:   resourceDetectiveAccountRead,
		DeleteWithoutTimeout: resourceDetectiveAccountDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"tags":     tagsSchemaForceNew(),
			"tags_all": tagsSchemaComputed(),
			"created_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceDetectiveAccountCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).detectiveconn

	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(keyvaluetags.New(d.Get("tags").(map[string]interface{})))

	input := &detective.CreateGraphInput{}

	if len(tags) > 0 {
		input.Tags = tags.IgnoreAws().DetectiveTags()
	}

	var output *detective.CreateGraphOutput
	var err error
	err = resource.RetryContext(ctx, waiter.DetectiveOperationTimeout, func() *resource.RetryError {
		output, err = conn.CreateGraphWithContext(ctx, input)
		if err != nil {
			if tfawserr.ErrCodeEquals(err, detective.ErrCodeInternalServerException) {
				return resource.RetryableError(err)
			}

			return resource.NonRetryableError(err)
		}

		return nil
	})

	if isResourceTimeoutError(err) {
		output, err = conn.CreateGraphWithContext(ctx, input)
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating detective Graph: %w", err))
	}

	d.SetId(aws.StringValue(output.GraphArn))

	return resourceDetectiveAccountRead(ctx, d, meta)
}

func resourceDetectiveAccountRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).detectiveconn

	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	resp, err := getDetectiveGraphArn(conn, ctx, d.Id())

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, detective.ErrCodeResourceNotFoundException) || resp == nil {
		d.SetId("")
		return nil
	}
	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading detective Graph (%s): %w", d.Id(), err))
	}

	d.Set("created_time", aws.TimeValue(resp.CreatedTime).Format(time.RFC3339))

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
	tags := keyvaluetags.DetectiveKeyValueTags(tg.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig)

	if err = d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return diag.FromErr(fmt.Errorf("error setting `%s` for Detective Graph (%s): %w", "tags", d.Id(), err))
	}

	if err = d.Set("tags_all", tags.Map()); err != nil {
		return diag.FromErr(fmt.Errorf("error setting `%s` for Detective Graph (%s): %w", "tags_all", d.Id(), err))
	}

	return nil
}

func resourceDetectiveAccountDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).detectiveconn

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

func getDetectiveGraphArn(conn *detective.Detective, ctx context.Context, graphArn string) (*detective.Graph, error) {
	var res *detective.Graph

	err := conn.ListGraphsPagesWithContext(ctx, &detective.ListGraphsInput{}, func(page *detective.ListGraphsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, graph := range page.GraphList {
			if graph == nil {
				continue
			}

			if aws.StringValue(graph.Arn) == graphArn {
				res = graph
				return false
			}
		}
		return !lastPage
	})

	return res, err
}
