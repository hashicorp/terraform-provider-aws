package kinesisvideo

import (
	"context"
	"log"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kinesisvideo"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceStream() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceStreamCreate,
		ReadWithoutTimeout:   resourceStreamRead,
		UpdateWithoutTimeout: resourceStreamUpdate,
		DeleteWithoutTimeout: resourceStreamDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
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

func resourceStreamCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KinesisVideoConn()
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

	resp, err := conn.CreateStreamWithContext(ctx, createOpts)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Kinesis Video Stream: %s", err)
	}

	arn := aws.StringValue(resp.StreamARN)
	d.SetId(arn)

	stateConf := &resource.StateChangeConf{
		Pending:    []string{kinesisvideo.StatusCreating},
		Target:     []string{kinesisvideo.StatusActive},
		Refresh:    StreamStateRefresh(ctx, conn, arn),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	if _, err = stateConf.WaitForStateContext(ctx); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for creating Kinesis Video Stream (%s): %s", d.Id(), err)
	}

	return append(diags, resourceStreamRead(ctx, d, meta)...)
}

func resourceStreamRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KinesisVideoConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	descOpts := &kinesisvideo.DescribeStreamInput{
		StreamARN: aws.String(d.Id()),
	}

	resp, err := conn.DescribeStreamWithContext(ctx, descOpts)
	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, kinesisvideo.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] Kinesis Video Stream (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "describing Kinesis Video Stream (%s): %s", d.Id(), err)
	}

	d.Set("name", resp.StreamInfo.StreamName)
	d.Set("data_retention_in_hours", resp.StreamInfo.DataRetentionInHours)
	d.Set("device_name", resp.StreamInfo.DeviceName)
	d.Set("kms_key_id", resp.StreamInfo.KmsKeyId)
	d.Set("media_type", resp.StreamInfo.MediaType)
	d.Set("arn", resp.StreamInfo.StreamARN)
	if err := d.Set("creation_time", resp.StreamInfo.CreationTime.Format(time.RFC3339)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting creation_time: %s", err)
	}
	d.Set("version", resp.StreamInfo.Version)

	tags, err := ListTags(ctx, conn, d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing tags for Kinesis Video Stream (%s): %s", d.Id(), err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags_all: %s", err)
	}

	return diags
}

func resourceStreamUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KinesisVideoConn()

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

	if _, err := conn.UpdateStreamWithContext(ctx, updateOpts); err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Kinesis Video Stream (%s): %s", d.Id(), err)
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(ctx, conn, d.Id(), o, n); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Kinesis Video Stream (%s) tags: %s", d.Id(), err)
		}
	}

	stateConf := &resource.StateChangeConf{
		Pending:    []string{kinesisvideo.StatusUpdating},
		Target:     []string{kinesisvideo.StatusActive},
		Refresh:    StreamStateRefresh(ctx, conn, d.Id()),
		Timeout:    d.Timeout(schema.TimeoutUpdate),
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	if _, err := stateConf.WaitForStateContext(ctx); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for updating Kinesis Video Stream (%s): %s", d.Id(), err)
	}

	return append(diags, resourceStreamRead(ctx, d, meta)...)
}

func resourceStreamDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KinesisVideoConn()

	if _, err := conn.DeleteStreamWithContext(ctx, &kinesisvideo.DeleteStreamInput{
		StreamARN:      aws.String(d.Id()),
		CurrentVersion: aws.String(d.Get("version").(string)),
	}); err != nil {
		if tfawserr.ErrCodeEquals(err, kinesisvideo.ErrCodeResourceNotFoundException) {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "deleting Kinesis Video Stream (%s): %s", d.Id(), err)
	}

	stateConf := &resource.StateChangeConf{
		Pending:    []string{kinesisvideo.StatusDeleting},
		Target:     []string{"DELETED"},
		Refresh:    StreamStateRefresh(ctx, conn, d.Id()),
		Timeout:    d.Timeout(schema.TimeoutDelete),
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	if _, err := stateConf.WaitForStateContext(ctx); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for deleting Kinesis Video Stream (%s): %s", d.Id(), err)
	}

	return diags
}

func StreamStateRefresh(ctx context.Context, conn *kinesisvideo.KinesisVideo, arn string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		emptyResp := &kinesisvideo.DescribeStreamInput{}

		resp, err := conn.DescribeStreamWithContext(ctx, &kinesisvideo.DescribeStreamInput{
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
