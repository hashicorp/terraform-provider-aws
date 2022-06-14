package detective

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/detective"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceGraph() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceGraphCreate,
		ReadContext:   resourceGraphRead,
		UpdateContext: resourceGraphUpdate,
		DeleteContext: resourceGraphDelete,
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
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},
		CustomizeDiff: verify.SetTagsDiff,
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
	err = resource.RetryContext(ctx, GraphOperationTimeout, func() *resource.RetryError {
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
		return diag.Errorf("error creating detective Graph: %s", err)
	}

	d.SetId(aws.StringValue(output.GraphArn))

	return resourceGraphRead(ctx, d, meta)
}

func resourceGraphRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).DetectiveConn

	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	resp, err := FindGraphByARN(conn, ctx, d.Id())

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, detective.ErrCodeResourceNotFoundException) || resp == nil {
		d.SetId("")
		return nil
	}
	if err != nil {
		return diag.Errorf("error reading detective Graph (%s): %s", d.Id(), err)
	}

	d.Set("created_time", aws.TimeValue(resp.CreatedTime).Format(time.RFC3339))
	d.Set("graph_arn", resp.Arn)

	tags, err := ListTags(conn, aws.StringValue(resp.Arn))

	if err != nil {
		return diag.Errorf("error listing tags for Detective Graph (%s): %s", d.Id(), err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	if err = d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return diag.Errorf("error setting `%s` for Detective Graph (%s): %s", "tags", d.Id(), err)
	}

	if err = d.Set("tags_all", tags.Map()); err != nil {
		return diag.Errorf("error setting `%s` for Detective Graph (%s): %s", "tags_all", d.Id(), err)
	}

	return nil
}

func resourceGraphUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).DetectiveConn

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")
		if err := UpdateTags(conn, d.Id(), o, n); err != nil {
			return diag.Errorf("error updating detective Graph tags (%s): %s", d.Id(), err)
		}
	}

	return resourceGraphRead(ctx, d, meta)
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
		return diag.Errorf("error deleting detective Graph (%s): %s", d.Id(), err)
	}

	return nil
}
