package kinesis

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceStream() *schema.Resource {
	return &schema.Resource{
		Create: resourceStreamCreate,
		Read:   resourceStreamRead,
		Update: resourceStreamUpdate,
		Delete: resourceStreamDelete,
		Importer: &schema.ResourceImporter{
			State: resourceStreamImport,
		},

		CustomizeDiff: customdiff.Sequence(
			verify.SetTagsDiff,
			func(_ context.Context, diff *schema.ResourceDiff, meta interface{}) error {
				shardCount := diff.Get("shard_count").(int)
				streamMode := kinesis.StreamModeProvisioned
				if v, ok := diff.GetOk("stream_mode_details.0.stream_mode"); ok {
					streamMode = v.(string)
				}
				switch streamMode {
				case kinesis.StreamModeOnDemand:
					if shardCount > 0 {
						return fmt.Errorf("shard_count must not be set when stream_mode is %s", streamMode)
					}
				case kinesis.StreamModeProvisioned:
					if shardCount < 1 {
						return fmt.Errorf("shard_count must be at least 1 when stream_mode is %s", streamMode)
					}
				default:
					return fmt.Errorf("unsupported stream mode %s", streamMode)
				}

				return nil
			},
		),

		SchemaVersion: 1,
		StateUpgraders: []schema.StateUpgrader{
			{
				Type:    resourceStreamResourceV0().CoreConfigSchema().ImpliedType(),
				Upgrade: StreamStateUpgradeV0,
				Version: 0,
			},
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(120 * time.Minute),
			Delete: schema.DefaultTimeout(120 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"encryption_type": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      kinesis.EncryptionTypeNone,
				ValidateFunc: validation.StringInSlice(kinesis.EncryptionType_Values(), true),
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return strings.EqualFold(old, new)
				},
			},
			"enforce_consumer_deletion": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"kms_key_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"retention_period": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      24,
				ValidateFunc: validation.IntBetween(24, 8760),
			},
			"shard_count": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"shard_level_metrics": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"stream_mode_details": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"stream_mode": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(kinesis.StreamMode_Values(), false),
						},
					},
				},
				DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},
	}
}

func resourceStreamCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).KinesisConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	name := d.Get("name").(string)
	input := &kinesis.CreateStreamInput{
		StreamName: aws.String(name),
	}

	if v, ok := d.GetOk("stream_mode_details"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.StreamModeDetails = expandStreamModeDetails(v.([]interface{})[0].(map[string]interface{}))
	}

	if streamMode := getStreamMode(d); streamMode == kinesis.StreamModeProvisioned {
		input.ShardCount = aws.Int64(int64(d.Get("shard_count").(int)))
	}

	log.Printf("[DEBUG] Creating Kinesis Stream: %s", input)
	_, err := conn.CreateStream(input)

	if err != nil {
		return fmt.Errorf("error creating Kinesis Stream (%s): %w", name, err)
	}

	streamDescription, err := waitStreamCreated(conn, name, d.Timeout(schema.TimeoutCreate))

	if streamDescription != nil {
		d.SetId(aws.StringValue(streamDescription.StreamARN))
	}

	if err != nil {
		return fmt.Errorf("error waiting for Kinesis Stream (%s) create: %w", name, err)
	}

	if v, ok := d.GetOk("retention_period"); ok && v.(int) > 0 {
		input := &kinesis.IncreaseStreamRetentionPeriodInput{
			RetentionPeriodHours: aws.Int64(int64(v.(int))),
			StreamName:           aws.String(name),
		}

		log.Printf("[DEBUG] Increasing Kinesis Stream retention period: %s", input)
		_, err := conn.IncreaseStreamRetentionPeriod(input)

		if err != nil {
			return fmt.Errorf("error increasing Kinesis Stream (%s) retention period: %w", name, err)
		}

		_, err = waitStreamUpdated(conn, name, d.Timeout(schema.TimeoutCreate))

		if err != nil {
			return fmt.Errorf("error waiting for Kinesis Stream (%s) update (IncreaseStreamRetentionPeriod): %w", name, err)
		}
	}

	if v, ok := d.GetOk("shard_level_metrics"); ok && v.(*schema.Set).Len() > 0 {
		input := &kinesis.EnableEnhancedMonitoringInput{
			ShardLevelMetrics: flex.ExpandStringSet(v.(*schema.Set)),
			StreamName:        aws.String(name),
		}

		log.Printf("[DEBUG] Enabling Kinesis Stream enhanced monitoring: %s", input)
		_, err := conn.EnableEnhancedMonitoring(input)

		if err != nil {
			return fmt.Errorf("error enabling Kinesis Stream (%s) enhanced monitoring: %w", name, err)
		}

		_, err = waitStreamUpdated(conn, name, d.Timeout(schema.TimeoutCreate))

		if err != nil {
			return fmt.Errorf("error waiting for Kinesis Stream (%s) update (EnableEnhancedMonitoring): %w", name, err)
		}
	}

	if v, ok := d.GetOk("encryption_type"); ok && v.(string) == kinesis.EncryptionTypeKms {
		if _, ok := d.GetOk("kms_key_id"); !ok {
			return fmt.Errorf("KMS Key Id required when setting encryption_type is not set as NONE")
		}

		input := &kinesis.StartStreamEncryptionInput{
			EncryptionType: aws.String(v.(string)),
			KeyId:          aws.String(d.Get("kms_key_id").(string)),
			StreamName:     aws.String(name),
		}

		log.Printf("[DEBUG] Starting Kinesis Stream encryption: %s", input)
		_, err := conn.StartStreamEncryption(input)

		if err != nil {
			return fmt.Errorf("error starting Kinesis Stream (%s) encryption: %w", name, err)
		}

		_, err = waitStreamUpdated(conn, name, d.Timeout(schema.TimeoutCreate))

		if err != nil {
			return fmt.Errorf("error waiting for Kinesis Stream (%s) update (StartStreamEncryption): %w", name, err)
		}
	}

	if len(tags) > 0 {
		if err := UpdateTags(conn, name, nil, tags); err != nil {
			return fmt.Errorf("error adding Kinesis Stream (%s) tags: %w", name, err)
		}
	}

	return resourceStreamRead(d, meta)
}

func resourceStreamRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).KinesisConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig
	name := d.Get("name").(string)

	stream, err := FindStreamByName(conn, name)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Kinesis Stream (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Kinesis Stream (%s): %w", name, err)
	}

	d.Set("arn", stream.StreamARN)
	d.Set("encryption_type", stream.EncryptionType)
	d.Set("kms_key_id", stream.KeyId)
	d.Set("name", stream.StreamName)
	d.Set("retention_period", stream.RetentionPeriodHours)

	streamMode := kinesis.StreamModeProvisioned
	if details := stream.StreamModeDetails; details != nil {
		streamMode = aws.StringValue(details.StreamMode)
	}
	if streamMode == kinesis.StreamModeProvisioned {
		d.Set("shard_count", stream.OpenShardCount)
	} else {
		d.Set("shard_count", nil)
	}

	var shardLevelMetrics []*string
	for _, v := range stream.EnhancedMonitoring {
		shardLevelMetrics = append(shardLevelMetrics, v.ShardLevelMetrics...)
	}
	d.Set("shard_level_metrics", aws.StringValueSlice(shardLevelMetrics))

	if details := stream.StreamModeDetails; details != nil {
		if err := d.Set("stream_mode_details", []interface{}{flattenStreamModeDetails(details)}); err != nil {
			return fmt.Errorf("error setting stream_mode_details: %w", err)
		}
	} else {
		d.Set("stream_mode_details", nil)
	}

	tags, err := ListTags(conn, name)

	if err != nil {
		return fmt.Errorf("error listing tags for Kinesis Stream (%s): %w", name, err)
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
	conn := meta.(*conns.AWSClient).KinesisConn
	name := d.Get("name").(string)

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, name, o, n); err != nil {
			return fmt.Errorf("error updating Kinesis Stream (%s) tags: %w", name, err)
		}
	}

	if d.HasChange("stream_mode_details.0.stream_mode") {
		input := &kinesis.UpdateStreamModeInput{
			StreamARN: aws.String(d.Id()),
			StreamModeDetails: &kinesis.StreamModeDetails{
				StreamMode: aws.String(d.Get("stream_mode_details.0.stream_mode").(string)),
			},
		}

		log.Printf("[DEBUG] Updating Kinesis Stream stream mode: %s", input)
		_, err := conn.UpdateStreamMode(input)

		if err != nil {
			return fmt.Errorf("error updating Kinesis Stream (%s) stream mode: %w", name, err)
		}

		_, err = waitStreamUpdated(conn, name, d.Timeout(schema.TimeoutUpdate))

		if err != nil {
			return fmt.Errorf("error waiting for Kinesis Stream (%s) update (UpdateStreamMode): %w", name, err)
		}
	}

	if streamMode := getStreamMode(d); streamMode == kinesis.StreamModeProvisioned && d.HasChange("shard_count") {
		input := &kinesis.UpdateShardCountInput{
			ScalingType:      aws.String(kinesis.ScalingTypeUniformScaling),
			StreamName:       aws.String(name),
			TargetShardCount: aws.Int64(int64(d.Get("shard_count").(int))),
		}

		log.Printf("[DEBUG] Updating Kinesis Stream shard count: %s", input)
		_, err := conn.UpdateShardCount(input)

		if err != nil {
			return fmt.Errorf("error updating Kinesis Stream (%s) shard count: %w", name, err)
		}

		_, err = waitStreamUpdated(conn, name, d.Timeout(schema.TimeoutUpdate))

		if err != nil {
			return fmt.Errorf("error waiting for Kinesis Stream (%s) update (UpdateShardCount): %w", name, err)
		}
	}

	if d.HasChange("retention_period") {
		oraw, nraw := d.GetChange("retention_period")
		o := oraw.(int)
		n := nraw.(int)

		if n > o {
			input := &kinesis.IncreaseStreamRetentionPeriodInput{
				RetentionPeriodHours: aws.Int64(int64(n)),
				StreamName:           aws.String(name),
			}

			log.Printf("[DEBUG] Increasing Kinesis Stream retention period: %s", input)
			_, err := conn.IncreaseStreamRetentionPeriod(input)

			if err != nil {
				return fmt.Errorf("error increasing Kinesis Stream (%s) retention period: %w", name, err)
			}

			_, err = waitStreamUpdated(conn, name, d.Timeout(schema.TimeoutUpdate))

			if err != nil {
				return fmt.Errorf("error waiting for Kinesis Stream (%s) update (IncreaseStreamRetentionPeriod): %w", name, err)
			}
		} else if n != 0 {
			input := &kinesis.DecreaseStreamRetentionPeriodInput{
				RetentionPeriodHours: aws.Int64(int64(n)),
				StreamName:           aws.String(name),
			}

			log.Printf("[DEBUG] Decreasing Kinesis Stream retention period: %s", input)
			_, err := conn.DecreaseStreamRetentionPeriod(input)

			if err != nil {
				return fmt.Errorf("error decreasing Kinesis Stream (%s) retention period: %w", name, err)
			}

			_, err = waitStreamUpdated(conn, name, d.Timeout(schema.TimeoutUpdate))

			if err != nil {
				return fmt.Errorf("error waiting for Kinesis Stream (%s) update (DecreaseStreamRetentionPeriod): %w", name, err)
			}
		}
	}

	if d.HasChange("shard_level_metrics") {
		o, n := d.GetChange("shard_level_metrics")
		os := o.(*schema.Set)
		ns := n.(*schema.Set)

		if del := os.Difference(ns); del.Len() > 0 {
			input := &kinesis.DisableEnhancedMonitoringInput{
				ShardLevelMetrics: flex.ExpandStringSet(del),
				StreamName:        aws.String(name),
			}

			log.Printf("[DEBUG] Disabling Kinesis Stream enhanced monitoring: %s", input)
			_, err := conn.DisableEnhancedMonitoring(input)

			if err != nil {
				return fmt.Errorf("error disabling Kinesis Stream (%s) enhanced monitoring: %w", name, err)
			}

			_, err = waitStreamUpdated(conn, name, d.Timeout(schema.TimeoutUpdate))

			if err != nil {
				return fmt.Errorf("error waiting for Kinesis Stream (%s) update (DisableEnhancedMonitoring): %w", name, err)
			}
		}

		if add := ns.Difference(os); add.Len() > 0 {
			input := &kinesis.EnableEnhancedMonitoringInput{
				ShardLevelMetrics: flex.ExpandStringSet(add),
				StreamName:        aws.String(name),
			}

			log.Printf("[DEBUG] Enabling Kinesis Stream enhanced monitoring: %s", input)
			_, err := conn.EnableEnhancedMonitoring(input)

			if err != nil {
				return fmt.Errorf("error enabling Kinesis Stream (%s) enhanced monitoring: %w", name, err)
			}

			_, err = waitStreamUpdated(conn, name, d.Timeout(schema.TimeoutUpdate))

			if err != nil {
				return fmt.Errorf("error waiting for Kinesis Stream (%s) update (EnableEnhancedMonitoring): %w", name, err)
			}
		}
	}

	if d.HasChanges("encryption_type", "kms_key_id") {
		oldEncryptionType, newEncryptionType := d.GetChange("encryption_type")
		oldKeyID, newKeyID := d.GetChange("kms_key_id")

		switch newEncryptionType, newKeyID := newEncryptionType.(string), newKeyID.(string); newEncryptionType {
		case kinesis.EncryptionTypeKms:
			if newKeyID == "" {
				return fmt.Errorf("KMS Key Id required when setting encryption_type is not set as NONE")
			}

			input := &kinesis.StartStreamEncryptionInput{
				EncryptionType: aws.String(newEncryptionType),
				KeyId:          aws.String(newKeyID),
				StreamName:     aws.String(name),
			}

			log.Printf("[DEBUG] Starting Kinesis Stream encryption: %s", input)
			_, err := conn.StartStreamEncryption(input)

			if err != nil {
				return fmt.Errorf("error starting Kinesis Stream (%s) encryption: %w", name, err)
			}

			_, err = waitStreamUpdated(conn, name, d.Timeout(schema.TimeoutUpdate))

			if err != nil {
				return fmt.Errorf("error waiting for Kinesis Stream (%s) update (StartStreamEncryption): %w", name, err)
			}

		case kinesis.EncryptionTypeNone:
			input := &kinesis.StopStreamEncryptionInput{
				EncryptionType: aws.String(oldEncryptionType.(string)),
				KeyId:          aws.String(oldKeyID.(string)),
				StreamName:     aws.String(name),
			}

			log.Printf("[DEBUG] Stopping Kinesis Stream encryption: %s", input)
			_, err := conn.StopStreamEncryption(input)

			if err != nil {
				return fmt.Errorf("error stopping Kinesis Stream (%s) encryption: %w", name, err)
			}

			_, err = waitStreamUpdated(conn, name, d.Timeout(schema.TimeoutUpdate))

			if err != nil {
				return fmt.Errorf("error waiting for Kinesis Stream (%s) update (StopStreamEncryption): %w", name, err)
			}

		default:
			return fmt.Errorf("unsupported encryption type: %s", newEncryptionType)
		}
	}

	return resourceStreamRead(d, meta)
}

func resourceStreamDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).KinesisConn
	name := d.Get("name").(string)

	log.Printf("[DEBUG] Deleting Kinesis Stream: (%s)", name)
	_, err := conn.DeleteStream(&kinesis.DeleteStreamInput{
		EnforceConsumerDeletion: aws.Bool(d.Get("enforce_consumer_deletion").(bool)),
		StreamName:              aws.String(name),
	})

	if tfawserr.ErrCodeEquals(err, kinesis.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting Kinesis Stream (%s): %w", name, err)
	}

	_, err = waitStreamDeleted(conn, name, d.Timeout(schema.TimeoutDelete))

	if err != nil {
		return fmt.Errorf("error waiting for Kinesis Stream (%s) delete: %w", name, err)
	}

	return nil
}

func resourceStreamImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	conn := meta.(*conns.AWSClient).KinesisConn

	output, err := FindStreamByName(conn, d.Id())

	if err != nil {
		return nil, err
	}

	d.SetId(aws.StringValue(output.StreamARN))
	d.Set("name", output.StreamName)
	return []*schema.ResourceData{d}, nil
}

func FindStreamByName(conn *kinesis.Kinesis, name string) (*kinesis.StreamDescriptionSummary, error) {
	input := &kinesis.DescribeStreamSummaryInput{
		StreamName: aws.String(name),
	}

	output, err := conn.DescribeStreamSummary(input)

	if tfawserr.ErrCodeEquals(err, kinesis.ErrCodeResourceNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.StreamDescriptionSummary == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.StreamDescriptionSummary, nil
}

func streamStatus(conn *kinesis.Kinesis, name string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindStreamByName(conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.StreamStatus), nil
	}
}

func waitStreamCreated(conn *kinesis.Kinesis, name string, timeout time.Duration) (*kinesis.StreamDescriptionSummary, error) {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{kinesis.StreamStatusCreating},
		Target:     []string{kinesis.StreamStatusActive},
		Refresh:    streamStatus(conn, name),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*kinesis.StreamDescriptionSummary); ok {
		return output, err
	}

	return nil, err
}

func waitStreamDeleted(conn *kinesis.Kinesis, name string, timeout time.Duration) (*kinesis.StreamDescriptionSummary, error) {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{kinesis.StreamStatusDeleting},
		Target:     []string{},
		Refresh:    streamStatus(conn, name),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*kinesis.StreamDescriptionSummary); ok {
		return output, err
	}

	return nil, err
}

func waitStreamUpdated(conn *kinesis.Kinesis, name string, timeout time.Duration) (*kinesis.StreamDescriptionSummary, error) { //nolint:unparam
	stateConf := &resource.StateChangeConf{
		Pending:    []string{kinesis.StreamStatusUpdating},
		Target:     []string{kinesis.StreamStatusActive},
		Refresh:    streamStatus(conn, name),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*kinesis.StreamDescriptionSummary); ok {
		return output, err
	}

	return nil, err
}

func getStreamMode(d *schema.ResourceData) string {
	streamMode, ok := d.GetOk("stream_mode_details.0.stream_mode")
	if !ok {
		return kinesis.StreamModeProvisioned
	}

	return streamMode.(string)
}

func expandStreamModeDetails(d map[string]interface{}) *kinesis.StreamModeDetails {
	if d == nil {
		return nil
	}

	apiObject := &kinesis.StreamModeDetails{}

	if v, ok := d["stream_mode"]; ok && len(v.(string)) > 0 {
		apiObject.StreamMode = aws.String(v.(string))
	}

	return apiObject
}

func flattenStreamModeDetails(apiObject *kinesis.StreamModeDetails) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.StreamMode; v != nil {
		tfMap["stream_mode"] = aws.StringValue(v)
	}

	return tfMap
}
