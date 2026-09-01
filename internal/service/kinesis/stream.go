// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package kinesis

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kinesis"
	"github.com/aws/aws-sdk-go-v2/service/kinesis/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/provider/sdkv2/importer"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/sdkv2"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_kinesis_stream", name="Stream")
// @IdentityAttribute("name")
// @CustomImport
// @Tags(identifierAttribute="name", resourceType="Stream")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/kinesis/types;awstypes;awstypes.StreamDescriptionSummary")
// @Testing(preIdentityVersion="v6.47.0")
// @Testing(tagsTest=false)
// @Testing(importIgnore="enforce_consumer_deletion")
// @Testing(importStateIdAttribute="name")
// @Testing(plannableImportAction="NoOp")
func resourceStream() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceStreamCreate,
		ReadWithoutTimeout:   resourceStreamRead,
		UpdateWithoutTimeout: resourceStreamUpdate,
		DeleteWithoutTimeout: resourceStreamDelete,

		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
				if err := importer.Import(ctx, d, meta); err != nil {
					return nil, err
				}

				conn := meta.(*conns.AWSClient).KinesisClient(ctx)

				// Region may be overridden in the import block.
				var optFns []func(*kinesis.Options)
				if v, ok := d.GetOk(names.AttrRegion); ok {
					optFns = append(optFns, func(o *kinesis.Options) {
						o.Region = v.(string)
					})
				}
				output, err := findStreamByName(ctx, conn, d.Id(), optFns...)

				if err != nil {
					return nil, err
				}

				d.SetId(aws.ToString(output.StreamARN))
				return []*schema.ResourceData{d}, nil
			},
		},

		CustomizeDiff: customdiff.Sequence(
			func(_ context.Context, diff *schema.ResourceDiff, meta any) error {
				switch streamMode, shardCount := getStreamMode(diff), diff.Get("shard_count").(int); streamMode {
				case types.StreamModeOnDemand:
					if shardCount > 0 {
						return fmt.Errorf("shard_count must not be set when stream_mode is %s", streamMode)
					}
				case types.StreamModeProvisioned:
					if shardCount < 1 {
						return fmt.Errorf("shard_count must be at least 1 when stream_mode is %s", streamMode)
					}
				}

				return nil
			},
			func(ctx context.Context, diff *schema.ResourceDiff, meta any) error {
				conn := meta.(*conns.AWSClient).KinesisClient(ctx)

				output, err := findLimits(ctx, conn)

				if err != nil {
					return nil //nolint:nilerr // Explicitly OK if IAM permissions not set (or any other error)
				}

				switch streamMode := getStreamMode(diff); streamMode {
				case types.StreamModeOnDemand:
					if diff.Id() == "" {
						if streamCount, streamLimit := aws.ToInt32(output.OnDemandStreamCount)+1, aws.ToInt32(output.OnDemandStreamCountLimit); streamCount > streamLimit {
							return fmt.Errorf("on-demand stream count (%d) would exceed the Kinesis account limit (%d)", streamCount, streamLimit)
						}
					}
				case types.StreamModeProvisioned:
					o, n := diff.GetChange("shard_count")
					if shardCount, shardLimit := aws.ToInt32(output.OpenShardCount)+int32(n.(int)-o.(int)), aws.ToInt32(output.ShardLimit); shardCount > shardLimit {
						return fmt.Errorf("open shard count (%d) would exceed the Kinesis account limit (%d)", shardCount, shardLimit)
					}
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

		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
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
					DiffSuppressFunc: sdkv2.SuppressEquivalentStringCaseInsensitive,
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
				"max_record_size_in_kib": {
					Type:         schema.TypeInt,
					Optional:     true,
					Computed:     true,
					ValidateFunc: validation.IntBetween(1024, 10240),
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
					Type:          schema.TypeInt,
					Optional:      true,
					ConflictsWith: []string{"warm_throughput_mib_ps"},
				},
				"shard_level_metrics": {
					Type:     schema.TypeSet,
					Optional: true,
					Elem: &schema.Schema{
						Type:             schema.TypeString,
						ValidateDiagFunc: enum.Validate[types.MetricsName](),
					},
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
				"warm_throughput_mib_ps": {
					Type:          schema.TypeInt,
					Optional:      true,
					ConflictsWith: []string{"shard_count"},
				},
			}
		},
	}
}

func resourceStreamCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KinesisClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := kinesis.CreateStreamInput{
		StreamName: aws.String(name),
	}

	if v, ok := d.GetOk("max_record_size_in_kib"); ok {
		input.MaxRecordSizeInKiB = aws.Int32(int32(v.(int)))
	}

	if streamMode := getStreamMode(d); streamMode == types.StreamModeProvisioned {
		input.ShardCount = aws.Int32(int32(d.Get("shard_count").(int)))
	}

	if v, ok := d.GetOk("stream_mode_details"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		input.StreamModeDetails = expandStreamModeDetails(v.([]any)[0].(map[string]any))
	}

	if tags := keyValueTags(ctx, getTagsIn(ctx)).Map(); len(tags) > 0 {
		input.Tags = tags
	}

	if v, ok := d.GetOk("warm_throughput_mib_ps"); ok {
		input.WarmThroughputMiBps = aws.Int32(int32(v.(int)))
	}

	_, err := conn.CreateStream(ctx, &input)

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
		input := kinesis.IncreaseStreamRetentionPeriodInput{
			RetentionPeriodHours: aws.Int32(int32(v.(int))),
			StreamName:           aws.String(name),
		}

		_, err := conn.IncreaseStreamRetentionPeriod(ctx, &input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "increasing Kinesis Stream (%s) retention period: %s", name, err)
		}

		if _, err := waitStreamUpdated(ctx, conn, name, d.Timeout(schema.TimeoutCreate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Kinesis Stream (%s) update (IncreaseStreamRetentionPeriod): %s", name, err)
		}
	}

	if v, ok := d.GetOk("shard_level_metrics"); ok && v.(*schema.Set).Len() > 0 {
		input := kinesis.EnableEnhancedMonitoringInput{
			ShardLevelMetrics: flex.ExpandStringyValueSet[types.MetricsName](v.(*schema.Set)),
			StreamName:        aws.String(name),
		}

		_, err := conn.EnableEnhancedMonitoring(ctx, &input)

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

			input := kinesis.StartStreamEncryptionInput{
				EncryptionType: v,
				KeyId:          aws.String(kmsKeyID.(string)),
				StreamName:     aws.String(name),
			}

			_, err := conn.StartStreamEncryption(ctx, &input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "starting Kinesis Stream (%s) encryption: %s", name, err)
			}

			if _, err := waitStreamUpdated(ctx, conn, name, d.Timeout(schema.TimeoutCreate)); err != nil {
				return sdkdiag.AppendErrorf(diags, "waiting for Kinesis Stream (%s) update (StartStreamEncryption): %s", name, err)
			}
		}
	}

	return append(diags, resourceStreamRead(ctx, d, meta)...)
}

func resourceStreamRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KinesisClient(ctx)

	name := d.Get(names.AttrName).(string)
	stream, err := findStreamByName(ctx, conn, name)

	if !d.IsNewResource() && retry.NotFound(err) {
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
	d.Set("max_record_size_in_kib", stream.MaxRecordSizeInKiB)
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
	if v := stream.StreamModeDetails; v != nil {
		if err := d.Set("stream_mode_details", []any{flattenStreamModeDetails(v)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting stream_mode_details: %s", err)
		}
	} else {
		d.Set("stream_mode_details", nil)
	}
	if v := stream.WarmThroughput; v != nil {
		d.Set("warm_throughput_mib_ps", v.CurrentMiBps)
	} else {
		d.Set("warm_throughput_mib_ps", nil)
	}

	return diags
}

func resourceStreamUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KinesisClient(ctx)
	name := d.Get(names.AttrName).(string)

	if d.HasChange("stream_mode_details.0.stream_mode") {
		input := kinesis.UpdateStreamModeInput{
			StreamARN: aws.String(d.Id()),
			StreamModeDetails: &types.StreamModeDetails{
				StreamMode: types.StreamMode(d.Get("stream_mode_details.0.stream_mode").(string)),
			},
		}

		_, err := conn.UpdateStreamMode(ctx, &input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Kinesis Stream (%s) stream mode: %s", name, err)
		}

		if _, err := waitStreamUpdated(ctx, conn, name, d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Kinesis Stream (%s) update (UpdateStreamMode): %s", name, err)
		}
	}

	if streamMode := getStreamMode(d); streamMode == types.StreamModeProvisioned && d.HasChange("shard_count") {
		input := kinesis.UpdateShardCountInput{
			ScalingType:      types.ScalingTypeUniformScaling,
			StreamName:       aws.String(name),
			TargetShardCount: aws.Int32(int32(d.Get("shard_count").(int))),
		}

		_, err := conn.UpdateShardCount(ctx, &input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Kinesis Stream (%s) shard count: %s", name, err)
		}

		if _, err := waitStreamUpdated(ctx, conn, name, d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Kinesis Stream (%s) update (UpdateShardCount): %s", name, err)
		}
	}

	if d.HasChange(names.AttrRetentionPeriod) {
		o, n := d.GetChange(names.AttrRetentionPeriod)
		oi, ni := int32(o.(int)), int32(n.(int))

		if ni > oi {
			input := kinesis.IncreaseStreamRetentionPeriodInput{
				RetentionPeriodHours: aws.Int32(ni),
				StreamName:           aws.String(name),
			}

			_, err := conn.IncreaseStreamRetentionPeriod(ctx, &input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "increasing Kinesis Stream (%s) retention period: %s", name, err)
			}

			if _, err := waitStreamUpdated(ctx, conn, name, d.Timeout(schema.TimeoutUpdate)); err != nil {
				return sdkdiag.AppendErrorf(diags, "waiting for Kinesis Stream (%s) update (IncreaseStreamRetentionPeriod): %s", name, err)
			}
		} else if ni != 0 {
			input := kinesis.DecreaseStreamRetentionPeriodInput{
				RetentionPeriodHours: aws.Int32(ni),
				StreamName:           aws.String(name),
			}

			_, err := conn.DecreaseStreamRetentionPeriod(ctx, &input)

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
		os, ns := o.(*schema.Set), n.(*schema.Set)

		if del := os.Difference(ns); del.Len() > 0 {
			input := kinesis.DisableEnhancedMonitoringInput{
				ShardLevelMetrics: flex.ExpandStringyValueSet[types.MetricsName](del),
				StreamName:        aws.String(name),
			}

			_, err := conn.DisableEnhancedMonitoring(ctx, &input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "disabling Kinesis Stream (%s) enhanced monitoring: %s", name, err)
			}

			if _, err := waitStreamUpdated(ctx, conn, name, d.Timeout(schema.TimeoutUpdate)); err != nil {
				return sdkdiag.AppendErrorf(diags, "waiting for Kinesis Stream (%s) update (DisableEnhancedMonitoring): %s", name, err)
			}
		}

		if add := ns.Difference(os); add.Len() > 0 {
			input := kinesis.EnableEnhancedMonitoringInput{
				ShardLevelMetrics: flex.ExpandStringyValueSet[types.MetricsName](add),
				StreamName:        aws.String(name),
			}

			_, err := conn.EnableEnhancedMonitoring(ctx, &input)

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

		switch oldEncryptionType, newEncryptionType, oldKeyID, newKeyID := types.EncryptionType(oldEncryptionType.(string)), types.EncryptionType(newEncryptionType.(string)), oldKeyID.(string), newKeyID.(string); newEncryptionType {
		case types.EncryptionTypeKms:
			if newKeyID == "" {
				return sdkdiag.AppendErrorf(diags, "KMS Key ID required when setting encryption_type is not set as NONE")
			}

			input := kinesis.StartStreamEncryptionInput{
				EncryptionType: newEncryptionType,
				KeyId:          aws.String(newKeyID),
				StreamName:     aws.String(name),
			}

			_, err := conn.StartStreamEncryption(ctx, &input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "starting Kinesis Stream (%s) encryption: %s", name, err)
			}

			if _, err := waitStreamUpdated(ctx, conn, name, d.Timeout(schema.TimeoutUpdate)); err != nil {
				return sdkdiag.AppendErrorf(diags, "waiting for Kinesis Stream (%s) update (StartStreamEncryption): %s", name, err)
			}

		case types.EncryptionTypeNone:
			input := kinesis.StopStreamEncryptionInput{
				EncryptionType: oldEncryptionType,
				KeyId:          aws.String(oldKeyID),
				StreamName:     aws.String(name),
			}

			_, err := conn.StopStreamEncryption(ctx, &input)

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

	if d.HasChange("max_record_size_in_kib") {
		_, n := d.GetChange("max_record_size_in_kib")

		input := kinesis.UpdateMaxRecordSizeInput{
			MaxRecordSizeInKiB: aws.Int32(int32(n.(int))),
			StreamARN:          aws.String(d.Id()),
		}

		_, err := conn.UpdateMaxRecordSize(ctx, &input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "update Kinesis Stream (%s) max record size: %s", name, err)
		}

		if _, err := waitStreamUpdated(ctx, conn, name, d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Kinesis Stream (%s) update (UpdateMaxRecordSize): %s", name, err)
		}
	}

	if d.HasChange("warm_throughput_mib_ps") {
		_, n := d.GetChange("warm_throughput_mib_ps")

		input := kinesis.UpdateStreamWarmThroughputInput{
			StreamARN:           aws.String(d.Id()),
			WarmThroughputMiBps: aws.Int32(int32(n.(int))),
		}

		_, err := conn.UpdateStreamWarmThroughput(ctx, &input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "update Kinesis Stream (%s) warm throughput: %s", name, err)
		}

		if _, err := waitStreamUpdated(ctx, conn, name, d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Kinesis Stream (%s) update (UpdateStreamWarmThroughput): %s", name, err)
		}
	}

	return append(diags, resourceStreamRead(ctx, d, meta)...)
}

func resourceStreamDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KinesisClient(ctx)

	log.Printf("[DEBUG] Deleting Kinesis Stream: (%s)", d.Id())
	name := d.Get(names.AttrName).(string)
	input := kinesis.DeleteStreamInput{
		EnforceConsumerDeletion: aws.Bool(d.Get("enforce_consumer_deletion").(bool)),
		StreamName:              aws.String(name),
	}
	_, err := conn.DeleteStream(ctx, &input)

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

func findLimits(ctx context.Context, conn *kinesis.Client) (*kinesis.DescribeLimitsOutput, error) {
	var input kinesis.DescribeLimitsInput
	output, err := conn.DescribeLimits(ctx, &input)

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output, nil
}

func findStreamByName(ctx context.Context, conn *kinesis.Client, name string, optFns ...func(*kinesis.Options)) (*types.StreamDescriptionSummary, error) {
	input := kinesis.DescribeStreamSummaryInput{
		StreamName: aws.String(name),
	}

	return findStreamSummary(ctx, conn, &input, optFns...)
}

func findStreamSummary(ctx context.Context, conn *kinesis.Client, input *kinesis.DescribeStreamSummaryInput, optFns ...func(*kinesis.Options)) (*types.StreamDescriptionSummary, error) {
	output, err := conn.DescribeStreamSummary(ctx, input, optFns...)

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.StreamDescriptionSummary == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output.StreamDescriptionSummary, nil
}

func streamStatus(conn *kinesis.Client, name string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findStreamByName(ctx, conn, name)

		if retry.NotFound(err) {
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
		Refresh:    streamStatus(conn, name),
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
		Refresh:    streamStatus(conn, name),
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
		Refresh:    streamStatus(conn, name),
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

func getStreamMode(d sdkv2.ResourceDiffer) types.StreamMode {
	streamMode, ok := d.GetOk("stream_mode_details.0.stream_mode")
	if !ok {
		return types.StreamModeProvisioned
	}

	return types.StreamMode(streamMode.(string))
}

func expandStreamModeDetails(d map[string]any) *types.StreamModeDetails {
	if d == nil {
		return nil
	}

	apiObject := &types.StreamModeDetails{}

	if v, ok := d["stream_mode"]; ok && len(v.(string)) > 0 {
		apiObject.StreamMode = types.StreamMode(v.(string))
	}

	return apiObject
}

func flattenStreamModeDetails(apiObject *types.StreamModeDetails) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{
		"stream_mode": apiObject.StreamMode,
	}

	return tfMap
}

func flattenWarmThroughputObject(apiObject *types.WarmThroughputObject) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{
		"current_mib_ps": aws.ToInt32(apiObject.CurrentMiBps),
		"target_mib_ps":  aws.ToInt32(apiObject.TargetMiBps),
	}

	return tfMap
}
