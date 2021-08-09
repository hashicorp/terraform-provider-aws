package aws

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/detective"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
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
			"tags": tagsSchemaForceNew(),
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
	err = resource.RetryContext(ctx, 4*time.Minute, func() *resource.RetryError {
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

	resp, err := getDetectiveGraphArn(conn, ctx, d.Id())

	if tfawserr.ErrCodeEquals(err, detective.ErrCodeResourceNotFoundException) {
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading detective Graph (%s): %w", d.Id(), err))
	}

	d.Set("created_time", aws.TimeValue(resp.CreatedTime).Format(time.RFC3339))

	return nil
}

func resourceDetectiveAccountDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).detectiveconn

	input := &detective.DeleteGraphInput{
		GraphArn: aws.String(d.Id()),
	}

	err := resource.RetryContext(ctx, 4*time.Minute, func() *resource.RetryError {
		_, err := conn.DeleteGraphWithContext(ctx, input)

		if err != nil {
			if tfawserr.ErrCodeEquals(err, detective.ErrCodeResourceNotFoundException) {
				return nil
			}
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if isResourceTimeoutError(err) {
		_, err = conn.DeleteGraphWithContext(ctx, input)
	}

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
