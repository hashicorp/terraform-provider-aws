// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kinesis

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kinesis"
	"github.com/aws/aws-sdk-go-v2/service/kinesis/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_kinesis_stream", name="Stream")
// @Tags(identifierAttribute="name")
func resourceStream() *schema.Resource {
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
				streamMode := types.StreamModeProvisioned
				if v, ok := diff.GetOk("stream_mode_details.0.stream_mode"); ok {
					streamMode = types.StreamMode(v.(string))
				}
				switch streamMode {
				case types.StreamModeOnDemand:
					if shardCount > 0 {
						return fmt.Errorf("shard_count must not be set when stream_mode is %s", streamMode)
					}
				case types.StreamModeProvisioned:
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
			names.AttrARN: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"encryption_type": {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          types.EncryptionTypeNone,
				ValidateDiagFunc: enum.Validate[types.EncryptionType](),
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return strings.EqualFold(old, new)
				},
			},
			"enforce_consumer_deletion": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			names.AttrKMSKeyID: {
				Type:     schema.TypeString,
				Optional: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrRetentionPeriod: {
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
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[types.StreamMode](),
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
	conn := meta.(*conns.AWSClient).KinesisClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &kinesis.CreateStreamInput{
		StreamName: aws.String(name),
	}

	if v, ok := d.GetOk("stream_mode_details"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.StreamModeDetails = expandStreamModeDetails(v.([]interface{})[0].(map[string]interface{}))
	}

	if streamMode := getStreamMode(d); streamMode == types.StreamModeProvisioned {
		input.ShardCount = aws.Int32(int32(d.Get("shard_count").(int)))
	}

	_, err := conn.CreateStream(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Kinesis Stream (%s): %s", name, err)
	}

	streamDescription, err := waitStreamCreated(ctx, conn, name, d.Timeout(schema.TimeoutCreate))

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Kinesis Stream (%s) create: %s", name, err)
	}

	if streamDescription != nil {
		d.SetId(aws.ToString(streamDescription.StreamARN))
	}

	if v, ok := d.GetOk(names.AttrRetentionPeriod); ok && v.(int) > 0 {
		input := &kinesis.IncreaseStreamRetentionPeriodInput{
			RetentionPeriodHours: aws.Int32(int32(v.(int))),
			StreamName:           aws.String(name),
		}

		_, err := conn.IncreaseStreamRetentionPeriod(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "increasing Kinesis Stream (%s) retention period: %s", name, err)
		}

		if _, err := waitStreamUpdated(ctx, conn, name, d.Timeout(schema.TimeoutCreate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Kinesis Stream (%s) update (IncreaseStreamRetentionPeriod): %s", name, err)
		}
	}

	if v, ok := d.GetOk("shard_level_metrics"); ok && v.(*schema.Set).Len() > 0 {
		input := &kinesis.EnableEnhancedMonitoringInput{
			ShardLevelMetrics: flex.ExpandStringyValueSet[types.MetricsName](v.(*schema.Set)),
			StreamName:        aws.String(name),
		}

		_, err := conn.EnableEnhancedMonitoring(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "enabling Kinesis Stream (%s) enhanced monitoring: %s", name, err)
		}

		if _, err := waitStreamUpdated(ctx, conn, name, d.Timeout(schema.TimeoutCreate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Kinesis Stream (%s) update (EnableEnhancedMonitoring): %s", name, err)
		}
	}

	if v, ok := d.GetOk("encryption_type"); ok {
		if v := types.EncryptionType(v.(string)); v == types.EncryptionTypeKms {
			kmsKeyID, ok := d.GetOk(names.AttrKMSKeyID)
			if !ok {
				return sdkdiag.AppendErrorf(diags, "KMS Key ID required when setting encryption_type is not set as NONE")
			}

			input := &kinesis.StartStreamEncryptionInput{
				EncryptionType: v,
				KeyId:          aws.String(kmsKeyID.(string)),
				StreamName:     aws.String(name),
			}

			_, err := conn.StartStreamEncryption(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "starting Kinesis Stream (%s) encryption: %s", name, err)
			}

			if _, err := waitStreamUpdated(ctx, conn, name, d.Timeout(schema.TimeoutCreate)); err != nil {
				return sdkdiag.AppendErrorf(diags, "waiting for Kinesis Stream (%s) update (StartStreamEncryption): %s", name, err)
			}
		}
	}

	if err := createTags(ctx, conn, name, getTagsIn(ctx)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting Kinesis Stream (%s) tags: %s", name, err)
	}

	return append(diags, resourceStreamRead(ctx, d, meta)...)
}

func resourceStreamRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KinesisClient(ctx)

	name := d.Get(names.AttrName).(string)
	stream, err := findStreamByName(ctx, conn, name)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Kinesis Stream (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Kinesis Stream (%s): %s", name, err)
	}

	d.Set(names.AttrARN, stream.StreamARN)
	d.Set("encryption_type", stream.EncryptionType)
	d.Set(names.AttrKMSKeyID, stream.KeyId)
	d.Set(names.AttrName, stream.StreamName)
	d.Set(names.AttrRetentionPeriod, stream.RetentionPeriodHours)
	streamMode := types.StreamModeProvisioned
	if details := stream.StreamModeDetails; details != nil {
		streamMode = details.StreamMode
	}
	if streamMode == types.StreamModeProvisioned {
		d.Set("shard_count", stream.OpenShardCount)
	} else {
		d.Set("shard_count", nil)
	}
	var shardLevelMetrics []types.MetricsName
	for _, v := range stream.EnhancedMonitoring {
		shardLevelMetrics = append(shardLevelMetrics, v.ShardLevelMetrics...)
	}
	d.Set("shard_level_metrics", shardLevelMetrics)
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
	conn := meta.(*conns.AWSClient).KinesisClient(ctx)
	name := d.Get(names.AttrName).(string)

	if d.HasChange("stream_mode_details.0.stream_mode") {
		input := &kinesis.UpdateStreamModeInput{
			StreamARN: aws.String(d.Id()),
			StreamModeDetails: &types.StreamModeDetails{
				StreamMode: types.StreamMode(d.Get("stream_mode_details.0.stream_mode").(string)),
			},
		}

		_, err := conn.UpdateStreamMode(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Kinesis Stream (%s) stream mode: %s", name, err)
		}

		if _, err := waitStreamUpdated(ctx, conn, name, d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Kinesis Stream (%s) update (UpdateStreamMode): %s", name, err)
		}
	}

	if streamMode := getStreamMode(d); streamMode == types.StreamModeProvisioned && d.HasChange("shard_count") {
		input := &kinesis.UpdateShardCountInput{
			ScalingType:      types.ScalingTypeUniformScaling,
			StreamName:       aws.String(name),
			TargetShardCount: aws.Int32(int32(d.Get("shard_count").(int))),
		}

		_, err := conn.UpdateShardCount(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Kinesis Stream (%s) shard count: %s", name, err)
		}

		if _, err := waitStreamUpdated(ctx, conn, name, d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Kinesis Stream (%s) update (UpdateShardCount): %s", name, err)
		}
	}

	if d.HasChange(names.AttrRetentionPeriod) {
		oraw, nraw := d.GetChange(names.AttrRetentionPeriod)
		o := oraw.(int)
		n := nraw.(int)

		if n > o {
			input := &kinesis.IncreaseStreamRetentionPeriodInput{
				RetentionPeriodHours: aws.Int32(int32(n)),
				StreamName:           aws.String(name),
			}

			_, err := conn.IncreaseStreamRetentionPeriod(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "increasing Kinesis Stream (%s) retention period: %s", name, err)
			}

			if _, err := waitStreamUpdated(ctx, conn, name, d.Timeout(schema.TimeoutUpdate)); err != nil {
				return sdkdiag.AppendErrorf(diags, "waiting for Kinesis Stream (%s) update (IncreaseStreamRetentionPeriod): %s", name, err)
			}
		} else if n != 0 {
			input := &kinesis.DecreaseStreamRetentionPeriodInput{
				RetentionPeriodHours: aws.Int32(int32(n)),
				StreamName:           aws.String(name),
			}

			_, err := conn.DecreaseStreamRetentionPeriod(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "decreasing Kinesis Stream (%s) retention period: %s", name, err)
			}

			if _, err := waitStreamUpdated(ctx, conn, name, d.Timeout(schema.TimeoutUpdate)); err != nil {
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
				ShardLevelMetrics: flex.ExpandStringyValueSet[types.MetricsName](del),
				StreamName:        aws.String(name),
			}

			_, err := conn.DisableEnhancedMonitoring(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "disabling Kinesis Stream (%s) enhanced monitoring: %s", name, err)
			}

			if _, err := waitStreamUpdated(ctx, conn, name, d.Timeout(schema.TimeoutUpdate)); err != nil {
				return sdkdiag.AppendErrorf(diags, "waiting for Kinesis Stream (%s) update (DisableEnhancedMonitoring): %s", name, err)
			}
		}

		if add := ns.Difference(os); add.Len() > 0 {
			input := &kinesis.EnableEnhancedMonitoringInput{
				ShardLevelMetrics: flex.ExpandStringyValueSet[types.MetricsName](add),
				StreamName:        aws.String(name),
			}

			_, err := conn.EnableEnhancedMonitoring(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "enabling Kinesis Stream (%s) enhanced monitoring: %s", name, err)
			}

			if _, err := waitStreamUpdated(ctx, conn, name, d.Timeout(schema.TimeoutUpdate)); err != nil {
				return sdkdiag.AppendErrorf(diags, "waiting for Kinesis Stream (%s) update (EnableEnhancedMonitoring): %s", name, err)
			}
		}
	}

	if d.HasChanges("encryption_type", names.AttrKMSKeyID) {
		oldEncryptionType, newEncryptionType := d.GetChange("encryption_type")
		oldKeyID, newKeyID := d.GetChange(names.AttrKMSKeyID)

		switch oldEncryptionType, newEncryptionType, newKeyID := types.EncryptionType(oldEncryptionType.(string)), types.EncryptionType(newEncryptionType.(string)), newKeyID.(string); newEncryptionType {
		case types.EncryptionTypeKms:
			if newKeyID == "" {
				return sdkdiag.AppendErrorf(diags, "KMS Key ID required when setting encryption_type is not set as NONE")
			}

			input := &kinesis.StartStreamEncryptionInput{
				EncryptionType: newEncryptionType,
				KeyId:          aws.String(newKeyID),
				StreamName:     aws.String(name),
			}

			_, err := conn.StartStreamEncryption(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "starting Kinesis Stream (%s) encryption: %s", name, err)
			}

			if _, err := waitStreamUpdated(ctx, conn, name, d.Timeout(schema.TimeoutUpdate)); err != nil {
				return sdkdiag.AppendErrorf(diags, "waiting for Kinesis Stream (%s) update (StartStreamEncryption): %s", name, err)
			}

		case types.EncryptionTypeNone:
			input := &kinesis.StopStreamEncryptionInput{
				EncryptionType: oldEncryptionType,
				KeyId:          aws.String(oldKeyID.(string)),
				StreamName:     aws.String(name),
			}

			_, err := conn.StopStreamEncryption(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "stopping Kinesis Stream (%s) encryption: %s", name, err)
			}

			if _, err := waitStreamUpdated(ctx, conn, name, d.Timeout(schema.TimeoutUpdate)); err != nil {
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
	conn := meta.(*conns.AWSClient).KinesisClient(ctx)
	name := d.Get(names.AttrName).(string)

	log.Printf("[DEBUG] Deleting Kinesis Stream: (%s)", name)
	_, err := conn.DeleteStream(ctx, &kinesis.DeleteStreamInput{
		EnforceConsumerDeletion: aws.Bool(d.Get("enforce_consumer_deletion").(bool)),
		StreamName:              aws.String(name),
	})

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Kinesis Stream (%s): %s", name, err)
	}

	if _, err := waitStreamDeleted(ctx, conn, name, d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Kinesis Stream (%s) delete: %s", name, err)
	}

	return diags
}

func resourceStreamImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	conn := meta.(*conns.AWSClient).KinesisClient(ctx)

	output, err := findStreamByName(ctx, conn, d.Id())

	if err != nil {
		return nil, err
	}

	d.SetId(aws.ToString(output.StreamARN))
	d.Set(names.AttrName, output.StreamName)
	return []*schema.ResourceData{d}, nil
}

func findStreamByName(ctx context.Context, conn *kinesis.Client, name string) (*types.StreamDescriptionSummary, error) {
	input := &kinesis.DescribeStreamSummaryInput{
		StreamName: aws.String(name),
	}

	output, err := conn.DescribeStreamSummary(ctx, input)

	if errs.IsA[*types.ResourceNotFoundException](err) {
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

func streamStatus(ctx context.Context, conn *kinesis.Client, name string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findStreamByName(ctx, conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.StreamStatus), nil
	}
}

func waitStreamCreated(ctx context.Context, conn *kinesis.Client, name string, timeout time.Duration) (*types.StreamDescriptionSummary, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(types.StreamStatusCreating),
		Target:     enum.Slice(types.StreamStatusActive),
		Refresh:    streamStatus(ctx, conn, name),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.StreamDescriptionSummary); ok {
		return output, err
	}

	return nil, err
}

func waitStreamDeleted(ctx context.Context, conn *kinesis.Client, name string, timeout time.Duration) (*types.StreamDescriptionSummary, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(types.StreamStatusDeleting),
		Target:     []string{},
		Refresh:    streamStatus(ctx, conn, name),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.StreamDescriptionSummary); ok {
		return output, err
	}

	return nil, err
}

func waitStreamUpdated(ctx context.Context, conn *kinesis.Client, name string, timeout time.Duration) (*types.StreamDescriptionSummary, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(types.StreamStatusUpdating),
		Target:     enum.Slice(types.StreamStatusActive),
		Refresh:    streamStatus(ctx, conn, name),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.StreamDescriptionSummary); ok {
		return output, err
	}

	return nil, err
}

func getStreamMode(d *schema.ResourceData) types.StreamMode {
	streamMode, ok := d.GetOk("stream_mode_details.0.stream_mode")
	if !ok {
		return types.StreamModeProvisioned
	}

	return types.StreamMode(streamMode.(string))
}

func expandStreamModeDetails(d map[string]interface{}) *types.StreamModeDetails {
	if d == nil {
		return nil
	}

	apiObject := &types.StreamModeDetails{}

	if v, ok := d["stream_mode"]; ok && len(v.(string)) > 0 {
		apiObject.StreamMode = types.StreamMode(v.(string))
	}

	return apiObject
}

func flattenStreamModeDetails(apiObject *types.StreamModeDetails) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"stream_mode": apiObject.StreamMode,
	}

	return tfMap
}
