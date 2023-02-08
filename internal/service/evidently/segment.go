package evidently

import (
	"context"
	"log"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatchevidently"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceSegment() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceSegmentCreate,
		ReadWithoutTimeout:   resourceSegmentRead,
		UpdateWithoutTimeout: resourceSegmentUpdate,
		DeleteWithoutTimeout: resourceSegmentDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 160),
			},
			"experiment_count": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"last_updated_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"launch_count": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 64),
					validation.StringMatch(regexp.MustCompile(`^[-a-zA-Z0-9._]*$`), "alphanumeric and can contain hyphens, underscores, and periods"),
				),
			},
			"pattern": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 1024),
					validation.StringIsJSON,
				),
				DiffSuppressFunc: verify.SuppressEquivalentJSONDiffs,
				StateFunc: func(v interface{}) string {
					json, _ := structure.NormalizeJsonString(v)
					return json
				},
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceSegmentCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EvidentlyConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	name := d.Get("name").(string)
	input := &cloudwatchevidently.CreateSegmentInput{
		Name:    aws.String(name),
		Pattern: aws.String(d.Get("pattern").(string)),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if len(tags) > 0 {
		input.Tags = Tags(tags.IgnoreAWS())
	}

	log.Printf("[DEBUG] Creating CloudWatch Evidently Segment: %s", input)
	output, err := conn.CreateSegmentWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("creating CloudWatch Evidently Segment (%s): %s", name, err)
	}

	d.SetId(aws.StringValue(output.Segment.Arn))

	return resourceSegmentRead(ctx, d, meta)
}

func resourceSegmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EvidentlyConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	segment, err := FindSegmentByNameOrARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] CloudWatch Evidently Segment (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading CloudWatch Evidently Segment (%s): %s", d.Id(), err)
	}

	d.Set("arn", segment.Arn)
	d.Set("created_time", aws.TimeValue(segment.CreatedTime).Format(time.RFC3339))
	d.Set("description", segment.Description)
	d.Set("experiment_count", segment.ExperimentCount)
	d.Set("last_updated_time", aws.TimeValue(segment.LastUpdatedTime).Format(time.RFC3339))
	d.Set("launch_count", segment.LaunchCount)
	d.Set("name", segment.Name)
	d.Set("pattern", segment.Pattern)

	tags := KeyValueTags(segment.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return diag.Errorf("setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return diag.Errorf("setting tags_all: %s", err)
	}

	return nil
}

func resourceSegmentUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EvidentlyConn()

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(ctx, conn, d.Id(), o, n); err != nil {
			return diag.Errorf("updating tags: %s", err)
		}
	}

	return resourceSegmentRead(ctx, d, meta)
}

func resourceSegmentDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EvidentlyConn()

	log.Printf("[DEBUG] Deleting CloudWatch Evidently Segment: %s", d.Id())
	_, err := conn.DeleteSegmentWithContext(ctx, &cloudwatchevidently.DeleteSegmentInput{
		Segment: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, cloudwatchevidently.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return diag.Errorf("deleting CloudWatch Evidently Segment (%s): %s", d.Id(), err)
	}

	return nil
}
