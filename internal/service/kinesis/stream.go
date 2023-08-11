// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

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
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_kinesis_stream", name="Stream")
// @Tags(identifierAttribute="name")
func ResourceStream() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceStreamCreate,
		ReadWithoutTimeout:   resourceStreamRead,
		UpdateWithoutTimeout: resourceStreamUpdate,
		DeleteWithoutTimeout: resourceStreamDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceStreamImport,
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
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourceStreamCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KinesisConn(ctx)

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
	_, err := conn.CreateStreamWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Kinesis Stream (%s): %s", name, err)
	}

	streamDescription, err := waitStreamCreated(ctx, conn, name, d.Timeout(schema.TimeoutCreate))

	if streamDescription != nil {
		d.SetId(aws.StringValue(streamDescription.StreamARN))
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Kinesis Stream (%s) create: %s", name, err)
	}

	if v, ok := d.GetOk("retention_period"); ok && v.(int) > 0 {
		input := &kinesis.IncreaseStreamRetentionPeriodInput{
			RetentionPeriodHours: aws.Int64(int64(v.(int))),
			StreamName:           aws.String(name),
		}

		log.Printf("[DEBUG] Increasing Kinesis Stream retention period: %s", input)
		_, err := conn.IncreaseStreamRetentionPeriodWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "increasing Kinesis Stream (%s) retention period: %s", name, err)
		}

		_, err = waitStreamUpdated(ctx, conn, name, d.Timeout(schema.TimeoutCreate))

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Kinesis Stream (%s) update (IncreaseStreamRetentionPeriod): %s", name, err)
		}
	}

	if v, ok := d.GetOk("shard_level_metrics"); ok && v.(*schema.Set).Len() > 0 {
		input := &kinesis.EnableEnhancedMonitoringInput{
			ShardLevelMetrics: flex.ExpandStringSet(v.(*schema.Set)),
			StreamName:        aws.String(name),
		}

		log.Printf("[DEBUG] Enabling Kinesis Stream enhanced monitoring: %s", input)
		_, err := conn.EnableEnhancedMonitoringWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "enabling Kinesis Stream (%s) enhanced monitoring: %s", name, err)
		}

		_, err = waitStreamUpdated(ctx, conn, name, d.Timeout(schema.TimeoutCreate))

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Kinesis Stream (%s) update (EnableEnhancedMonitoring): %s", name, err)
		}
	}

	if v, ok := d.GetOk("encryption_type"); ok && v.(string) == kinesis.EncryptionTypeKms {
		if _, ok := d.GetOk("kms_key_id"); !ok {
			return sdkdiag.AppendErrorf(diags, "KMS Key Id required when setting encryption_type is not set as NONE")
		}

		input := &kinesis.StartStreamEncryptionInput{
			EncryptionType: aws.String(v.(string)),
			KeyId:          aws.String(d.Get("kms_key_id").(string)),
			StreamName:     aws.String(name),
		}

		log.Printf("[DEBUG] Starting Kinesis Stream encryption: %s", input)
		_, err := conn.StartStreamEncryptionWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "starting Kinesis Stream (%s) encryption: %s", name, err)
		}

		_, err = waitStreamUpdated(ctx, conn, name, d.Timeout(schema.TimeoutCreate))

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Kinesis Stream (%s) update (StartStreamEncryption): %s", name, err)
		}
	}

	if err := createTags(ctx, conn, name, getTagsIn(ctx)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting Kinesis Stream (%s) tags: %s", name, err)
	}

	return append(diags, resourceStreamRead(ctx, d, meta)...)
}

func resourceStreamRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KinesisConn(ctx)

	name := d.Get("name").(string)
	stream, err := FindStreamByName(ctx, conn, name)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Kinesis Stream (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Kinesis Stream (%s): %s", name, err)
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
			return sdkdiag.AppendErrorf(diags, "setting stream_mode_details: %s", err)
		}
	} else {
		d.Set("stream_mode_details", nil)
	}

	return diags
}

func resourceStreamUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KinesisConn(ctx)
	name := d.Get("name").(string)

	if d.HasChange("stream_mode_details.0.stream_mode") {
		input := &kinesis.UpdateStreamModeInput{
			StreamARN: aws.String(d.Id()),
			StreamModeDetails: &kinesis.StreamModeDetails{
				StreamMode: aws.String(d.Get("stream_mode_details.0.stream_mode").(string)),
			},
		}

		log.Printf("[DEBUG] Updating Kinesis Stream stream mode: %s", input)
		_, err := conn.UpdateStreamModeWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Kinesis Stream (%s) stream mode: %s", name, err)
		}

		_, err = waitStreamUpdated(ctx, conn, name, d.Timeout(schema.TimeoutUpdate))

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Kinesis Stream (%s) update (UpdateStreamMode): %s", name, err)
		}
	}

	if streamMode := getStreamMode(d); streamMode == kinesis.StreamModeProvisioned && d.HasChange("shard_count") {
		input := &kinesis.UpdateShardCountInput{
			ScalingType:      aws.String(kinesis.ScalingTypeUniformScaling),
			StreamName:       aws.String(name),
			TargetShardCount: aws.Int64(int64(d.Get("shard_count").(int))),
		}

		log.Printf("[DEBUG] Updating Kinesis Stream shard count: %s", input)
		_, err := conn.UpdateShardCountWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Kinesis Stream (%s) shard count: %s", name, err)
		}

		_, err = waitStreamUpdated(ctx, conn, name, d.Timeout(schema.TimeoutUpdate))

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Kinesis Stream (%s) update (UpdateShardCount): %s", name, err)
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
			_, err := conn.IncreaseStreamRetentionPeriodWithContext(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "increasing Kinesis Stream (%s) retention period: %s", name, err)
			}

			_, err = waitStreamUpdated(ctx, conn, name, d.Timeout(schema.TimeoutUpdate))

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "waiting for Kinesis Stream (%s) update (IncreaseStreamRetentionPeriod): %s", name, err)
			}
		} else if n != 0 {
			input := &kinesis.DecreaseStreamRetentionPeriodInput{
				RetentionPeriodHours: aws.Int64(int64(n)),
				StreamName:           aws.String(name),
			}

			log.Printf("[DEBUG] Decreasing Kinesis Stream retention period: %s", input)
			_, err := conn.DecreaseStreamRetentionPeriodWithContext(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "decreasing Kinesis Stream (%s) retention period: %s", name, err)
			}

			_, err = waitStreamUpdated(ctx, conn, name, d.Timeout(schema.TimeoutUpdate))

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "waiting for Kinesis Stream (%s) update (DecreaseStreamRetentionPeriod): %s", name, err)
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
			_, err := conn.DisableEnhancedMonitoringWithContext(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "disabling Kinesis Stream (%s) enhanced monitoring: %s", name, err)
			}

			_, err = waitStreamUpdated(ctx, conn, name, d.Timeout(schema.TimeoutUpdate))

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "waiting for Kinesis Stream (%s) update (DisableEnhancedMonitoring): %s", name, err)
			}
		}

		if add := ns.Difference(os); add.Len() > 0 {
			input := &kinesis.EnableEnhancedMonitoringInput{
				ShardLevelMetrics: flex.ExpandStringSet(add),
				StreamName:        aws.String(name),
			}

			log.Printf("[DEBUG] Enabling Kinesis Stream enhanced monitoring: %s", input)
			_, err := conn.EnableEnhancedMonitoringWithContext(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "enabling Kinesis Stream (%s) enhanced monitoring: %s", name, err)
			}

			_, err = waitStreamUpdated(ctx, conn, name, d.Timeout(schema.TimeoutUpdate))

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "waiting for Kinesis Stream (%s) update (EnableEnhancedMonitoring): %s", name, err)
			}
		}
	}

	if d.HasChanges("encryption_type", "kms_key_id") {
		oldEncryptionType, newEncryptionType := d.GetChange("encryption_type")
		oldKeyID, newKeyID := d.GetChange("kms_key_id")

		switch newEncryptionType, newKeyID := newEncryptionType.(string), newKeyID.(string); newEncryptionType {
		case kinesis.EncryptionTypeKms:
			if newKeyID == "" {
				return sdkdiag.AppendErrorf(diags, "KMS Key Id required when setting encryption_type is not set as NONE")
			}

			input := &kinesis.StartStreamEncryptionInput{
				EncryptionType: aws.String(newEncryptionType),
				KeyId:          aws.String(newKeyID),
				StreamName:     aws.String(name),
			}

			log.Printf("[DEBUG] Starting Kinesis Stream encryption: %s", input)
			_, err := conn.StartStreamEncryptionWithContext(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "starting Kinesis Stream (%s) encryption: %s", name, err)
			}

			_, err = waitStreamUpdated(ctx, conn, name, d.Timeout(schema.TimeoutUpdate))

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "waiting for Kinesis Stream (%s) update (StartStreamEncryption): %s", name, err)
			}

		case kinesis.EncryptionTypeNone:
			input := &kinesis.StopStreamEncryptionInput{
				EncryptionType: aws.String(oldEncryptionType.(string)),
				KeyId:          aws.String(oldKeyID.(string)),
				StreamName:     aws.String(name),
			}

			log.Printf("[DEBUG] Stopping Kinesis Stream encryption: %s", input)
			_, err := conn.StopStreamEncryptionWithContext(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "stopping Kinesis Stream (%s) encryption: %s", name, err)
			}

			_, err = waitStreamUpdated(ctx, conn, name, d.Timeout(schema.TimeoutUpdate))

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "waiting for Kinesis Stream (%s) update (StopStreamEncryption): %s", name, err)
			}

		default:
			return sdkdiag.AppendErrorf(diags, "unsupported encryption type: %s", newEncryptionType)
		}
	}

	return append(diags, resourceStreamRead(ctx, d, meta)...)
}

func resourceStreamDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KinesisConn(ctx)
	name := d.Get("name").(string)

	log.Printf("[DEBUG] Deleting Kinesis Stream: (%s)", name)
	_, err := conn.DeleteStreamWithContext(ctx, &kinesis.DeleteStreamInput{
		EnforceConsumerDeletion: aws.Bool(d.Get("enforce_consumer_deletion").(bool)),
		StreamName:              aws.String(name),
	})

	if tfawserr.ErrCodeEquals(err, kinesis.ErrCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Kinesis Stream (%s): %s", name, err)
	}

	_, err = waitStreamDeleted(ctx, conn, name, d.Timeout(schema.TimeoutDelete))

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Kinesis Stream (%s) delete: %s", name, err)
	}

	return diags
}

func resourceStreamImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	conn := meta.(*conns.AWSClient).KinesisConn(ctx)

	output, err := FindStreamByName(ctx, conn, d.Id())

	if err != nil {
		return nil, err
	}

	d.SetId(aws.StringValue(output.StreamARN))
	d.Set("name", output.StreamName)
	return []*schema.ResourceData{d}, nil
}

func FindStreamByName(ctx context.Context, conn *kinesis.Kinesis, name string) (*kinesis.StreamDescriptionSummary, error) {
	input := &kinesis.DescribeStreamSummaryInput{
		StreamName: aws.String(name),
	}

	output, err := conn.DescribeStreamSummaryWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, kinesis.ErrCodeResourceNotFoundException) {
		return nil, &retry.NotFoundError{
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

func streamStatus(ctx context.Context, conn *kinesis.Kinesis, name string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindStreamByName(ctx, conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.StreamStatus), nil
	}
}

func waitStreamCreated(ctx context.Context, conn *kinesis.Kinesis, name string, timeout time.Duration) (*kinesis.StreamDescriptionSummary, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{kinesis.StreamStatusCreating},
		Target:     []string{kinesis.StreamStatusActive},
		Refresh:    streamStatus(ctx, conn, name),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*kinesis.StreamDescriptionSummary); ok {
		return output, err
	}

	return nil, err
}

func waitStreamDeleted(ctx context.Context, conn *kinesis.Kinesis, name string, timeout time.Duration) (*kinesis.StreamDescriptionSummary, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{kinesis.StreamStatusDeleting},
		Target:     []string{},
		Refresh:    streamStatus(ctx, conn, name),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*kinesis.StreamDescriptionSummary); ok {
		return output, err
	}

	return nil, err
}

func waitStreamUpdated(ctx context.Context, conn *kinesis.Kinesis, name string, timeout time.Duration) (*kinesis.StreamDescriptionSummary, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending:    []string{kinesis.StreamStatusUpdating},
		Target:     []string{kinesis.StreamStatusActive},
		Refresh:    streamStatus(ctx, conn, name),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

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
