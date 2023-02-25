package location

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/locationservice"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceTracker() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceTrackerCreate,
		ReadWithoutTimeout:   resourceTrackerRead,
		UpdateWithoutTimeout: resourceTrackerUpdate,
		DeleteWithoutTimeout: resourceTrackerDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"create_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 1000),
			},
			"kms_key_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 2048),
			},
			"position_filtering": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      locationservice.PositionFilteringTimeBased,
				ValidateFunc: validation.StringInSlice(locationservice.PositionFiltering_Values(), false),
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
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
		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceTrackerCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LocationConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := &locationservice.CreateTrackerInput{}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("kms_key_id"); ok {
		input.KmsKeyId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("position_filtering"); ok {
		input.PositionFiltering = aws.String(v.(string))
	}

	if len(tags) > 0 {
		input.Tags = Tags(tags.IgnoreAWS())
	}

	if v, ok := d.GetOk("tracker_name"); ok {
		input.TrackerName = aws.String(v.(string))
	}

	output, err := conn.CreateTrackerWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Location Service Tracker: %s", err)
	}

	if output == nil {
		return sdkdiag.AppendErrorf(diags, "creating Location Service Tracker: empty result")
	}

	d.SetId(aws.StringValue(output.TrackerName))

	return append(diags, resourceTrackerRead(ctx, d, meta)...)
}

func resourceTrackerRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LocationConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	input := &locationservice.DescribeTrackerInput{
		TrackerName: aws.String(d.Id()),
	}

	output, err := conn.DescribeTrackerWithContext(ctx, input)

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, locationservice.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] Location Service Tracker (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "getting Location Service Tracker (%s): %s", d.Id(), err)
	}

	if output == nil {
		return sdkdiag.AppendErrorf(diags, "getting Location Service Map (%s): empty response", d.Id())
	}

	d.Set("create_time", aws.TimeValue(output.CreateTime).Format(time.RFC3339))
	d.Set("description", output.Description)
	d.Set("kms_key_id", output.KmsKeyId)
	d.Set("position_filtering", output.PositionFiltering)

	tags := KeyValueTags(output.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags_all: %s", err)
	}

	d.Set("tracker_arn", output.TrackerArn)
	d.Set("tracker_name", output.TrackerName)
	d.Set("update_time", aws.TimeValue(output.UpdateTime).Format(time.RFC3339))

	return diags
}

func resourceTrackerUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LocationConn()

	if d.HasChanges("description", "position_filtering") {
		input := &locationservice.UpdateTrackerInput{
			TrackerName: aws.String(d.Id()),
		}

		if v, ok := d.GetOk("description"); ok {
			input.Description = aws.String(v.(string))
		}

		if v, ok := d.GetOk("position_filtering"); ok {
			input.PositionFiltering = aws.String(v.(string))
		}

		_, err := conn.UpdateTrackerWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Location Service Tracker (%s): %s", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(ctx, conn, d.Get("tracker_arn").(string), o, n); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating tags for Location Service Tracker (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceTrackerRead(ctx, d, meta)...)
}

func resourceTrackerDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LocationConn()

	input := &locationservice.DeleteTrackerInput{
		TrackerName: aws.String(d.Id()),
	}

	_, err := conn.DeleteTrackerWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, locationservice.ErrCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Location Service Tracker (%s): %s", d.Id(), err)
	}

	return diags
}
