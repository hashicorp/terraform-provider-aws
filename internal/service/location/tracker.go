package location

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/locationservice"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceTracker() *schema.Resource {
	return &schema.Resource{
		Create: resourceTrackerCreate,
		Read:   resourceTrackerRead,
		Update: resourceTrackerUpdate,
		Delete: resourceTrackerDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
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

func resourceTrackerCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LocationConn
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

	output, err := conn.CreateTracker(input)

	if err != nil {
		return fmt.Errorf("error creating Location Service Tracker: %w", err)
	}

	if output == nil {
		return fmt.Errorf("error creating Location Service Tracker: empty result")
	}

	d.SetId(aws.StringValue(output.TrackerName))

	return resourceTrackerRead(d, meta)
}

func resourceTrackerRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LocationConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	input := &locationservice.DescribeTrackerInput{
		TrackerName: aws.String(d.Id()),
	}

	output, err := conn.DescribeTracker(input)

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, locationservice.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] Location Service Tracker (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error getting Location Service Tracker (%s): %w", d.Id(), err)
	}

	if output == nil {
		return fmt.Errorf("error getting Location Service Map (%s): empty response", d.Id())
	}

	d.Set("create_time", aws.TimeValue(output.CreateTime).Format(time.RFC3339))
	d.Set("description", output.Description)
	d.Set("kms_key_id", output.KmsKeyId)
	d.Set("position_filtering", output.PositionFiltering)

	tags := KeyValueTags(output.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	d.Set("tracker_arn", output.TrackerArn)
	d.Set("tracker_name", output.TrackerName)
	d.Set("update_time", aws.TimeValue(output.UpdateTime).Format(time.RFC3339))

	return nil
}

func resourceTrackerUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LocationConn

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

		_, err := conn.UpdateTracker(input)

		if err != nil {
			return fmt.Errorf("error updating Location Service Tracker (%s): %w", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Get("tracker_arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating tags for Location Service Tracker (%s): %w", d.Id(), err)
		}
	}

	return resourceTrackerRead(d, meta)
}

func resourceTrackerDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LocationConn

	input := &locationservice.DeleteTrackerInput{
		TrackerName: aws.String(d.Id()),
	}

	_, err := conn.DeleteTracker(input)

	if tfawserr.ErrCodeEquals(err, locationservice.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting Location Service Tracker (%s): %w", d.Id(), err)
	}

	return nil
}
