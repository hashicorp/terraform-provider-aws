package kinesisvideo

import (
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kinesisvideo"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceStream() *schema.Resource {
	return &schema.Resource{
		Create: resourceStreamCreate,
		Read:   resourceStreamRead,
		Update: resourceStreamUpdate,
		Delete: resourceStreamDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		CustomizeDiff: verify.SetTagsDiff,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(120 * time.Minute),
			Delete: schema.DefaultTimeout(120 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"data_retention_in_hours": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
				Default:  0,
			},

			"device_name": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(0, 128),
					validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9_.-]+$`), "must only include alphanumeric, underscore, period, or hyphen characters"),
				),
			},

			"kms_key_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"media_type": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"creation_time": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"version": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"tags": tftags.TagsSchema(),

			"tags_all": tftags.TagsSchemaComputed(),
		},
	}
}

func resourceStreamCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).KinesisVideoConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	createOpts := &kinesisvideo.CreateStreamInput{
		StreamName:           aws.String(d.Get("name").(string)),
		DataRetentionInHours: aws.Int64(int64(d.Get("data_retention_in_hours").(int))),
	}

	if v, ok := d.GetOk("device_name"); ok {
		createOpts.DeviceName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("kms_key_id"); ok {
		createOpts.KmsKeyId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("media_type"); ok {
		createOpts.MediaType = aws.String(v.(string))
	}

	if len(tags) > 0 {
		createOpts.Tags = Tags(tags.IgnoreAWS())
	}

	resp, err := conn.CreateStream(createOpts)
	if err != nil {
		return fmt.Errorf("Error creating Kinesis Video Stream: %s", err)
	}

	arn := aws.StringValue(resp.StreamARN)
	d.SetId(arn)

	stateConf := &resource.StateChangeConf{
		Pending:    []string{kinesisvideo.StatusCreating},
		Target:     []string{kinesisvideo.StatusActive},
		Refresh:    StreamStateRefresh(conn, arn),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	if _, err = stateConf.WaitForState(); err != nil {
		return fmt.Errorf("Error waiting for creating Kinesis Video Stream (%s): %s", d.Id(), err)
	}

	return resourceStreamRead(d, meta)
}

func resourceStreamRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).KinesisVideoConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	descOpts := &kinesisvideo.DescribeStreamInput{
		StreamARN: aws.String(d.Id()),
	}

	resp, err := conn.DescribeStream(descOpts)
	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, kinesisvideo.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] Kinesis Video Stream (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("Error describing Kinesis Video Stream (%s): %w", d.Id(), err)
	}

	d.Set("name", resp.StreamInfo.StreamName)
	d.Set("data_retention_in_hours", resp.StreamInfo.DataRetentionInHours)
	d.Set("device_name", resp.StreamInfo.DeviceName)
	d.Set("kms_key_id", resp.StreamInfo.KmsKeyId)
	d.Set("media_type", resp.StreamInfo.MediaType)
	d.Set("arn", resp.StreamInfo.StreamARN)
	if err := d.Set("creation_time", resp.StreamInfo.CreationTime.Format(time.RFC3339)); err != nil {
		return fmt.Errorf("error setting creation_time: %w", err)
	}
	d.Set("version", resp.StreamInfo.Version)

	tags, err := ListTags(conn, d.Id())

	if err != nil {
		return fmt.Errorf("error listing tags for Kinesis Video Stream (%s): %w", d.Id(), err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceStreamUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).KinesisVideoConn

	updateOpts := &kinesisvideo.UpdateStreamInput{
		StreamARN:      aws.String(d.Id()),
		CurrentVersion: aws.String(d.Get("version").(string)),
	}

	if v, ok := d.GetOk("device_name"); ok {
		updateOpts.DeviceName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("media_type"); ok {
		updateOpts.MediaType = aws.String(v.(string))
	}

	if _, err := conn.UpdateStream(updateOpts); err != nil {
		return fmt.Errorf("Error updating Kinesis Video Stream (%s): %w", d.Id(), err)
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating Kinesis Video Stream (%s) tags: %w", d.Id(), err)
		}
	}

	stateConf := &resource.StateChangeConf{
		Pending:    []string{kinesisvideo.StatusUpdating},
		Target:     []string{kinesisvideo.StatusActive},
		Refresh:    StreamStateRefresh(conn, d.Id()),
		Timeout:    d.Timeout(schema.TimeoutUpdate),
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	if _, err := stateConf.WaitForState(); err != nil {
		return fmt.Errorf("Error waiting for updating Kinesis Video Stream (%s): %w", d.Id(), err)
	}

	return resourceStreamRead(d, meta)
}

func resourceStreamDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).KinesisVideoConn

	if _, err := conn.DeleteStream(&kinesisvideo.DeleteStreamInput{
		StreamARN:      aws.String(d.Id()),
		CurrentVersion: aws.String(d.Get("version").(string)),
	}); err != nil {
		if tfawserr.ErrCodeEquals(err, kinesisvideo.ErrCodeResourceNotFoundException) {
			return nil
		}
		return fmt.Errorf("Error deleting Kinesis Video Stream (%s): %w", d.Id(), err)
	}

	stateConf := &resource.StateChangeConf{
		Pending:    []string{kinesisvideo.StatusDeleting},
		Target:     []string{"DELETED"},
		Refresh:    StreamStateRefresh(conn, d.Id()),
		Timeout:    d.Timeout(schema.TimeoutDelete),
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	if _, err := stateConf.WaitForState(); err != nil {
		return fmt.Errorf("Error waiting for deleting Kinesis Video Stream (%s): %w", d.Id(), err)
	}

	return nil
}

func StreamStateRefresh(conn *kinesisvideo.KinesisVideo, arn string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		emptyResp := &kinesisvideo.DescribeStreamInput{}

		resp, err := conn.DescribeStream(&kinesisvideo.DescribeStreamInput{
			StreamARN: aws.String(arn),
		})
		if tfawserr.ErrCodeEquals(err, kinesisvideo.ErrCodeResourceNotFoundException) {
			return emptyResp, "DELETED", nil
		}
		if err != nil {
			return nil, "", err
		}

		return resp, aws.StringValue(resp.StreamInfo.Status), nil
	}
}
