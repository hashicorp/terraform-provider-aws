package qldb

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/qldb"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceStream() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceStreamCreate,
		ReadWithoutTimeout:   resourceStreamRead,
		UpdateWithoutTimeout: resourceStreamUpdate,
		DeleteWithoutTimeout: resourceStreamDelete,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"exclusive_end_time": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.IsRFC3339Time,
			},
			"inclusive_start_time": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.IsRFC3339Time,
			},
			"kinesis_configuration": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"aggregation_enabled": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
							ForceNew: true,
						},
						"stream_arn": {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: verify.ValidARN,
						},
					},
				},
			},
			"ledger_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 32),
				),
			},
			"role_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"stream_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 32),
				),
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceStreamCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).QLDBConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	ledgerName := d.Get("ledger_name").(string)
	name := d.Get("stream_name").(string)
	input := &qldb.StreamJournalToKinesisInput{
		LedgerName: aws.String(ledgerName),
		RoleArn:    aws.String(d.Get("role_arn").(string)),
		StreamName: aws.String(name),
		Tags:       Tags(tags.IgnoreAWS()),
	}

	if v, ok := d.GetOk("exclusive_end_time"); ok {
		v, _ := time.Parse(time.RFC3339, v.(string))
		input.ExclusiveEndTime = aws.Time(v)
	}

	if v, ok := d.GetOk("inclusive_start_time"); ok {
		v, _ := time.Parse(time.RFC3339, v.(string))
		input.InclusiveStartTime = aws.Time(v)
	}

	if v, ok := d.GetOk("kinesis_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.KinesisConfiguration = expandKinesisConfiguration(v.([]interface{})[0].(map[string]interface{}))
	}

	log.Printf("[DEBUG] Creating QLDB Stream: %s", input)
	output, err := conn.StreamJournalToKinesisWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("creating QLDB Stream (%s): %s", name, err)
	}

	d.SetId(aws.StringValue(output.StreamId))

	if _, err := waitStreamCreated(ctx, conn, ledgerName, d.Id()); err != nil {
		return diag.Errorf("waiting for QLDB Stream (%s) create: %s", d.Id(), err)
	}

	return resourceStreamRead(ctx, d, meta)
}

func resourceStreamRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).QLDBConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	ledgerName := d.Get("ledger_name").(string)
	stream, err := FindStream(ctx, conn, ledgerName, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] QLDB Stream %s not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading QLDB Stream (%s): %s", d.Id(), err)
	}

	d.Set("arn", stream.Arn)
	if stream.ExclusiveEndTime != nil {
		d.Set("exclusive_end_time", aws.TimeValue(stream.ExclusiveEndTime).Format(time.RFC3339))
	} else {
		d.Set("exclusive_end_time", nil)
	}
	if stream.InclusiveStartTime != nil {
		d.Set("inclusive_start_time", aws.TimeValue(stream.InclusiveStartTime).Format(time.RFC3339))
	} else {
		d.Set("inclusive_start_time", nil)
	}
	if stream.KinesisConfiguration != nil {
		if err := d.Set("kinesis_configuration", []interface{}{flattenKinesisConfiguration(stream.KinesisConfiguration)}); err != nil {
			return diag.Errorf("setting kinesis_configuration: %s", err)
		}
	} else {
		d.Set("kinesis_configuration", nil)
	}
	d.Set("ledger_name", stream.LedgerName)
	d.Set("role_arn", stream.RoleArn)
	d.Set("stream_name", stream.StreamName)

	tags, err := ListTags(ctx, conn, d.Get("arn").(string))

	if err != nil {
		return diag.Errorf("listing tags for QLDB Stream (%s): %s", d.Id(), err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return diag.Errorf("setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return diag.Errorf("setting tags_all: %s", err)
	}

	return nil
}

func resourceStreamUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).QLDBConn()

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := UpdateTags(ctx, conn, d.Get("arn").(string), o, n); err != nil {
			return diag.Errorf("updating tags: %s", err)
		}
	}

	return resourceStreamRead(ctx, d, meta)
}

func resourceStreamDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).QLDBConn()

	ledgerName := d.Get("ledger_name").(string)
	input := &qldb.CancelJournalKinesisStreamInput{
		LedgerName: aws.String(ledgerName),
		StreamId:   aws.String(d.Id()),
	}

	log.Printf("[INFO] Deleting QLDB Stream: %s", d.Id())
	_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, 5*time.Minute,
		func() (interface{}, error) {
			return conn.CancelJournalKinesisStreamWithContext(ctx, input)
		}, qldb.ErrCodeResourceInUseException)

	if tfawserr.ErrCodeEquals(err, qldb.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return diag.Errorf("deleting QLDB Stream (%s): %s", d.Id(), err)
	}

	if _, err := waitStreamDeleted(ctx, conn, ledgerName, d.Id()); err != nil {
		return diag.Errorf("waiting for QLDB Stream (%s) delete: %s", d.Id(), err)
	}

	return nil
}

func FindStream(ctx context.Context, conn *qldb.QLDB, ledgerName, streamID string) (*qldb.JournalKinesisStreamDescription, error) {
	input := &qldb.DescribeJournalKinesisStreamInput{
		LedgerName: aws.String(ledgerName),
		StreamId:   aws.String(streamID),
	}

	output, err := findJournalKinesisStream(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// See https://docs.aws.amazon.com/qldb/latest/developerguide/streams.create.html#streams.create.states.
	switch status := aws.StringValue(output.Status); status {
	case qldb.StreamStatusCompleted, qldb.StreamStatusCanceled, qldb.StreamStatusFailed:
		return nil, &resource.NotFoundError{
			Message:     status,
			LastRequest: input,
		}
	}

	return output, nil
}

func findJournalKinesisStream(ctx context.Context, conn *qldb.QLDB, input *qldb.DescribeJournalKinesisStreamInput) (*qldb.JournalKinesisStreamDescription, error) {
	output, err := conn.DescribeJournalKinesisStreamWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, qldb.ErrCodeResourceNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Stream == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Stream, nil
}

func statusStreamCreated(ctx context.Context, conn *qldb.QLDB, ledgerName, streamID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		// Don't call FindStream as it maps useful statuses to NotFoundError.
		output, err := findJournalKinesisStream(ctx, conn, &qldb.DescribeJournalKinesisStreamInput{
			LedgerName: aws.String(ledgerName),
			StreamId:   aws.String(streamID),
		})

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status), nil
	}
}

func waitStreamCreated(ctx context.Context, conn *qldb.QLDB, ledgerName, streamID string) (*qldb.JournalKinesisStreamDescription, error) {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{qldb.StreamStatusImpaired},
		Target:     []string{qldb.StreamStatusActive},
		Refresh:    statusStreamCreated(ctx, conn, ledgerName, streamID),
		Timeout:    8 * time.Minute,
		MinTimeout: 3 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*qldb.JournalKinesisStreamDescription); ok {
		tfresource.SetLastError(err, errors.New(aws.StringValue(output.ErrorCause)))

		return output, err
	}

	return nil, err
}

func statusStreamDeleted(ctx context.Context, conn *qldb.QLDB, ledgerName, streamID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindStream(ctx, conn, ledgerName, streamID)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status), nil
	}
}

func waitStreamDeleted(ctx context.Context, conn *qldb.QLDB, ledgerName, streamID string) (*qldb.JournalKinesisStreamDescription, error) {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{qldb.StreamStatusActive, qldb.StreamStatusImpaired},
		Target:     []string{},
		Refresh:    statusStreamDeleted(ctx, conn, ledgerName, streamID),
		Timeout:    5 * time.Minute,
		MinTimeout: 1 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*qldb.JournalKinesisStreamDescription); ok {
		tfresource.SetLastError(err, errors.New(aws.StringValue(output.ErrorCause)))

		return output, err
	}

	return nil, err
}

func expandKinesisConfiguration(tfMap map[string]interface{}) *qldb.KinesisConfiguration {
	if tfMap == nil {
		return nil
	}

	apiObject := &qldb.KinesisConfiguration{}

	if v, ok := tfMap["aggregation_enabled"].(bool); ok {
		apiObject.AggregationEnabled = aws.Bool(v)
	}

	if v, ok := tfMap["stream_arn"].(string); ok && v != "" {
		apiObject.StreamArn = aws.String(v)
	}

	return apiObject
}

func flattenKinesisConfiguration(apiObject *qldb.KinesisConfiguration) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.AggregationEnabled; v != nil {
		tfMap["aggregation_enabled"] = aws.BoolValue(v)
	}

	if v := apiObject.StreamArn; v != nil {
		tfMap["stream_arn"] = aws.StringValue(v)
	}

	return tfMap
}
