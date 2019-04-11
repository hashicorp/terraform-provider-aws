package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kinesisvideo"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsKinesisVideoStream() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsKinesisVideoStreamCreate,
		Read:   resourceAwsKinesisVideoStreamRead,
		Update: resourceAwsKinesisVideoStreamUpdate,
		Delete: resourceAwsKinesisVideoStreamDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

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
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateAwsKinesisVideoStreamDeviceName,
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

			"tags": tagsSchema(),
		},
	}
}

func resourceAwsKinesisVideoStreamCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).kinesisvideoconn

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

	if v, ok := d.GetOk("tags"); ok {
		createOpts.Tags = tagsFromMapGeneric(v.(map[string]interface{}))
	}

	resp, err := conn.CreateStream(createOpts)
	if err != nil {
		return fmt.Errorf("Error creating Kinesis Video Stream: %s", err)
	}

	arn := aws.StringValue(resp.StreamARN)

	stateConf := &resource.StateChangeConf{
		Pending:    []string{kinesisvideo.StatusCreating},
		Target:     []string{kinesisvideo.StatusActive},
		Refresh:    kinesisVideoStreamStateRefresh(conn, arn),
		Timeout:    15 * time.Minute,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	if _, err = stateConf.WaitForState(); err != nil {
		return fmt.Errorf("Error waiting for creating Kinesis Video Stream (%s): %s", d.Id(), err)
	}

	d.SetId(arn)

	return resourceAwsKinesisVideoStreamRead(d, meta)
}

func resourceAwsKinesisVideoStreamRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).kinesisvideoconn

	descOpts := &kinesisvideo.DescribeStreamInput{
		StreamARN: aws.String(d.Id()),
	}

	resp, err := conn.DescribeStream(descOpts)
	if isAWSErr(err, kinesisvideo.ErrCodeResourceNotFoundException, "") {
		log.Printf("[WARN] Kinesis Video Stream (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("Error describing Kinesis Video Stream (%s): %s", d.Id(), err)
	}

	d.Set("name", resp.StreamInfo.StreamName)
	d.Set("data_retention_in_hours", resp.StreamInfo.DataRetentionInHours)
	d.Set("device_name", resp.StreamInfo.DeviceName)
	d.Set("kms_key_id", resp.StreamInfo.KmsKeyId)
	d.Set("media_type", resp.StreamInfo.MediaType)
	d.Set("arn", resp.StreamInfo.StreamARN)
	if err := d.Set("creation_time", resp.StreamInfo.CreationTime.Format(time.RFC3339)); err != nil {
		return fmt.Errorf("error setting creation_time: %s", err)
	}
	d.Set("version", resp.StreamInfo.Version)

	if err := saveTagsKinesisVideoStream(conn, d, d.Id()); err != nil {
		if isAWSErr(err, kinesisvideo.ErrCodeResourceNotFoundException, "") {
			log.Printf("[WARN] Kinesis Video Stream (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}

		return fmt.Errorf("error saving tags for %s: %s", d.Id(), err)
	}

	return nil
}

func resourceAwsKinesisVideoStreamUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).kinesisvideoconn

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
		if isAWSErr(err, kinesisvideo.ErrCodeResourceNotFoundException, "") {
			log.Printf("[WARN] Kinesis Video Stream (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error updating Kinesis Video Stream (%s): %s", d.Id(), err)
	}

	if err := setTagsKinesisVideoStream(conn, d, d.Id()); err != nil {
		if isAWSErr(err, kinesisvideo.ErrCodeResourceNotFoundException, "") {
			log.Printf("[WARN] Kinesis Video Stream (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}

		return fmt.Errorf("error setting tags for %s: %s", d.Id(), err)
	}

	stateConf := &resource.StateChangeConf{
		Pending:    []string{kinesisvideo.StatusUpdating},
		Target:     []string{kinesisvideo.StatusActive},
		Refresh:    kinesisVideoStreamStateRefresh(conn, d.Id()),
		Timeout:    15 * time.Minute,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	if _, err := stateConf.WaitForState(); err != nil {
		return fmt.Errorf("Error waiting for updating Kinesis Video Stream (%s): %s", d.Id(), err)
	}

	return resourceAwsKinesisVideoStreamRead(d, meta)
}

func resourceAwsKinesisVideoStreamDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).kinesisvideoconn

	if _, err := conn.DeleteStream(&kinesisvideo.DeleteStreamInput{
		StreamARN:      aws.String(d.Id()),
		CurrentVersion: aws.String(d.Get("version").(string)),
	}); err != nil {
		if isAWSErr(err, kinesisvideo.ErrCodeResourceNotFoundException, "") {
			return nil
		}
		return fmt.Errorf("Error deleting Kinesis Video Stream (%s): %s", d.Id(), err)
	}

	stateConf := &resource.StateChangeConf{
		Pending:    []string{kinesisvideo.StatusDeleting},
		Target:     []string{"DELETED"},
		Refresh:    kinesisVideoStreamStateRefresh(conn, d.Id()),
		Timeout:    15 * time.Minute,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	if _, err := stateConf.WaitForState(); err != nil {
		return fmt.Errorf("Error waiting for deleting Kinesis Video Stream (%s): %s", d.Id(), err)
	}

	return nil
}

func kinesisVideoStreamStateRefresh(conn *kinesisvideo.KinesisVideo, arn string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		emptyResp := &kinesisvideo.DescribeStreamInput{}

		resp, err := conn.DescribeStream(&kinesisvideo.DescribeStreamInput{
			StreamARN: aws.String(arn),
		})
		if isAWSErr(err, kinesisvideo.ErrCodeResourceNotFoundException, "") {
			return emptyResp, "DELETED", nil
		}
		if err != nil {
			return nil, "", err
		}

		return resp, aws.StringValue(resp.StreamInfo.Status), nil
	}
}
