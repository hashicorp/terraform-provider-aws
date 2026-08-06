// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package kafka

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kafka"
	awstypes "github.com/aws/aws-sdk-go-v2/service/kafka/types"
	"github.com/hashicorp/terraform-plugin-framework/types"

	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
)

const (
	channelTestClusterARN = "arn:aws:kafka:us-east-1:123456789012:cluster/example/00000000-0000-0000-0000-000000000000-1"
	channelTestTopicARN   = "arn:aws:kafka:us-east-1:123456789012:topic/example/00000000-0000-0000-0000-000000000000-1/example"
	channelTestRoleARN    = "arn:aws:iam::123456789012:role/example"
	channelTestBucketARN  = "arn:aws:s3:::example-dest"
	channelTestDLQARN     = "arn:aws:s3:::example-dlq"
)

// TestChannelExpandS3 verifies an S3-destination model expands into a CreateChannelInput with
// the required fields set and the unused destination left nil. This exercises the AutoFlex
// field-name mapping between channelResourceModel and kafka.CreateChannelInput.
func TestChannelExpandS3(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	model := channelResourceModel{
		ChannelName: types.StringValue("example"),
		ClusterARN:  fwtypes.ARNValue(channelTestClusterARN),
		TopicConfigurationList: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &topicConfigurationModel{
			TopicARN: fwtypes.ARNValue(channelTestTopicARN),
			RecordConverter: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &recordConverterModel{
				ValueConverter: fwtypes.StringEnumValue(awstypes.ValueConverterByteArray),
			}),
			RecordSchema: fwtypes.NewListNestedObjectValueOfNull[recordSchemaModel](ctx),
		}),
		IcebergDestinationConfiguration: fwtypes.NewListNestedObjectValueOfNull[icebergDestinationModel](ctx),
		S3DestinationConfiguration: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &s3DestinationModel{
			DataFreshnessInSeconds:  types.Int32Null(),
			ServiceExecutionRoleARN: fwtypes.ARNValue(channelTestRoleARN),
			DeadLetterQueueS3: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &deadLetterQueueS3Model{
				BucketARN:           fwtypes.ARNValue(channelTestDLQARN),
				ErrorOutputPrefix:   types.StringNull(),
				ExpectedBucketOwner: types.StringNull(),
			}),
			Storage: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &s3StorageModel{
				BucketARN:           fwtypes.ARNValue(channelTestBucketARN),
				CompressionType:     fwtypes.StringEnumValue(awstypes.S3CompressionTypeNone),
				ExpectedBucketOwner: types.StringNull(),
				OutputKeyTemplate:   types.StringNull(),
				OutputPrefix:        types.StringNull(),
				StorageClass:        fwtypes.StringEnumValue(awstypes.S3StorageClassStandard),
			}),
		}),
		EncryptionConfiguration: fwtypes.NewListNestedObjectValueOfNull[encryptionConfigModel](ctx),
		LoggingInfo:             fwtypes.NewListNestedObjectValueOfNull[channelLoggingInfoModel](ctx),
	}

	var input kafka.CreateChannelInput
	if diags := fwflex.Expand(ctx, model, &input); diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	if got, want := aws.ToString(input.ChannelName), "example"; got != want {
		t.Errorf("ChannelName = %q, want %q", got, want)
	}
	if got, want := aws.ToString(input.ClusterArn), channelTestClusterARN; got != want {
		t.Errorf("ClusterArn = %q, want %q", got, want)
	}
	if got := len(input.TopicConfigurationList); got != 1 {
		t.Fatalf("len(TopicConfigurationList) = %d, want 1", got)
	}
	if got, want := aws.ToString(input.TopicConfigurationList[0].TopicArn), channelTestTopicARN; got != want {
		t.Errorf("TopicConfigurationList[0].TopicArn = %q, want %q", got, want)
	}
	if got, want := input.TopicConfigurationList[0].RecordConverter.ValueConverter, awstypes.ValueConverterByteArray; got != want {
		t.Errorf("RecordConverter.ValueConverter = %q, want %q", got, want)
	}
	if input.IcebergDestinationConfiguration != nil {
		t.Error("IcebergDestinationConfiguration is non-nil, want nil")
	}
	if input.S3DestinationConfiguration == nil {
		t.Fatal("S3DestinationConfiguration is nil, want non-nil")
	}
	if got, want := aws.ToString(input.S3DestinationConfiguration.ServiceExecutionRoleArn), channelTestRoleARN; got != want {
		t.Errorf("S3DestinationConfiguration.ServiceExecutionRoleArn = %q, want %q", got, want)
	}
	if input.S3DestinationConfiguration.Storage == nil {
		t.Fatal("S3DestinationConfiguration.Storage is nil, want non-nil")
	}
	if got, want := input.S3DestinationConfiguration.Storage.CompressionType, awstypes.S3CompressionTypeNone; got != want {
		t.Errorf("Storage.CompressionType = %q, want %q", got, want)
	}
	if got, want := input.S3DestinationConfiguration.Storage.StorageClass, awstypes.S3StorageClassStandard; got != want {
		t.Errorf("Storage.StorageClass = %q, want %q", got, want)
	}
	if input.S3DestinationConfiguration.DeadLetterQueueS3 == nil {
		t.Fatal("S3DestinationConfiguration.DeadLetterQueueS3 is nil, want non-nil")
	}
	if got, want := aws.ToString(input.S3DestinationConfiguration.DeadLetterQueueS3.BucketArn), channelTestDLQARN; got != want {
		t.Errorf("DeadLetterQueueS3.BucketArn = %q, want %q", got, want)
	}
}

// TestChannelExpandIceberg verifies an Iceberg-destination model expands into a
// CreateChannelInput with the Iceberg fields set (append-only, catalog warehouse, destination
// table, record schema) and the S3 destination left nil.
func TestChannelExpandIceberg(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	const gsrARN = "arn:aws:glue:us-east-1:123456789012:schema/example/example"
	const warehouseARN = "arn:aws:s3tables:us-east-1:123456789012:bucket/example"

	model := channelResourceModel{
		ChannelName: types.StringValue("example"),
		ClusterARN:  fwtypes.ARNValue(channelTestClusterARN),
		TopicConfigurationList: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &topicConfigurationModel{
			TopicARN: fwtypes.ARNValue(channelTestTopicARN),
			RecordConverter: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &recordConverterModel{
				ValueConverter: fwtypes.StringEnumValue(awstypes.ValueConverterJson),
			}),
			RecordSchema: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &recordSchemaModel{
				GSRARN: fwtypes.ARNValue(gsrARN),
			}),
		}),
		S3DestinationConfiguration: fwtypes.NewListNestedObjectValueOfNull[s3DestinationModel](ctx),
		IcebergDestinationConfiguration: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &icebergDestinationModel{
			AppendOnly:              types.BoolValue(true),
			CompressionType:         fwtypes.StringEnumValue(awstypes.IcebergCompressionTypeZstd),
			DataFreshnessInSeconds:  types.Int32Value(300),
			ServiceExecutionRoleARN: fwtypes.ARNValue(channelTestRoleARN),
			Catalog: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &catalogModel{
				CatalogARN:        fwtypes.ARNNull(),
				WarehouseLocation: fwtypes.ARNValue(warehouseARN),
			}),
			DeadLetterQueueS3: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &deadLetterQueueS3Model{
				BucketARN:           fwtypes.ARNValue(channelTestDLQARN),
				ErrorOutputPrefix:   types.StringNull(),
				ExpectedBucketOwner: types.StringNull(),
			}),
			DestinationTableList: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &destinationTableModel{
				DestinationDatabaseName: types.StringValue("example_namespace"),
				DestinationTableName:    types.StringValue("example_table"),
				PartitionSpec:           fwtypes.NewListNestedObjectValueOfNull[partitionSpecModel](ctx),
			}),
			SchemaEvolution: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &schemaEvolutionModel{
				EnableSchemaEvolution: types.BoolValue(false),
			}),
			TableCreation: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tableCreationModel{
				EnableTableCreation: types.BoolValue(true),
			}),
		}),
		EncryptionConfiguration: fwtypes.NewListNestedObjectValueOfNull[encryptionConfigModel](ctx),
		LoggingInfo:             fwtypes.NewListNestedObjectValueOfNull[channelLoggingInfoModel](ctx),
	}

	var input kafka.CreateChannelInput
	if diags := fwflex.Expand(ctx, model, &input); diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	if input.S3DestinationConfiguration != nil {
		t.Error("S3DestinationConfiguration is non-nil, want nil")
	}
	if input.IcebergDestinationConfiguration == nil {
		t.Fatal("IcebergDestinationConfiguration is nil, want non-nil")
	}
	iceberg := input.IcebergDestinationConfiguration
	if got := aws.ToBool(iceberg.AppendOnly); !got {
		t.Error("IcebergDestinationConfiguration.AppendOnly = false, want true")
	}
	if got, want := aws.ToInt32(iceberg.DataFreshnessInSeconds), int32(300); got != want {
		t.Errorf("IcebergDestinationConfiguration.DataFreshnessInSeconds = %d, want %d", got, want)
	}
	if iceberg.Catalog == nil {
		t.Fatal("IcebergDestinationConfiguration.Catalog is nil, want non-nil")
	}
	if got, want := aws.ToString(iceberg.Catalog.WarehouseLocation), warehouseARN; got != want {
		t.Errorf("Catalog.WarehouseLocation = %q, want %q", got, want)
	}
	if got := len(iceberg.DestinationTableList); got != 1 {
		t.Fatalf("len(DestinationTableList) = %d, want 1", got)
	}
	if got, want := aws.ToString(iceberg.DestinationTableList[0].DestinationTableName), "example_table"; got != want {
		t.Errorf("DestinationTableList[0].DestinationTableName = %q, want %q", got, want)
	}
	if got := len(input.TopicConfigurationList); got != 1 {
		t.Fatalf("len(TopicConfigurationList) = %d, want 1", got)
	}
	if input.TopicConfigurationList[0].RecordSchema == nil {
		t.Fatal("TopicConfigurationList[0].RecordSchema is nil, want non-nil")
	}
	if got, want := aws.ToString(input.TopicConfigurationList[0].RecordSchema.GsrArn), gsrARN; got != want {
		t.Errorf("RecordSchema.GsrArn = %q, want %q", got, want)
	}
}

// TestChannelFlattenS3 verifies a DescribeChannelOutput flattens into the resource model,
// populating the computed identifiers and the S3 destination while leaving the Iceberg
// destination null. This exercises the AutoFlex mapping used by Read.
func TestChannelFlattenS3(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	output := &kafka.DescribeChannelOutput{
		ChannelArn:      aws.String("arn:aws:kafka:us-east-1:123456789012:channel/example/00000000-0000-0000-0000-000000000000-1/abcd1234"),
		ChannelName:     aws.String("example"),
		CreationTime:    aws.Time(time.Unix(0, 0).UTC()),
		DestinationType: awstypes.ChannelDestinationTypeS3,
		Status:          awstypes.ChannelStatusActive,
		TopicConfigurationList: []awstypes.TopicConfiguration{{
			TopicArn:        aws.String(channelTestTopicARN),
			RecordConverter: &awstypes.RecordConverter{ValueConverter: awstypes.ValueConverterByteArray},
		}},
		S3DestinationConfiguration: &awstypes.S3DestinationConfiguration{
			ServiceExecutionRoleArn: aws.String(channelTestRoleARN),
			DeadLetterQueueS3:       &awstypes.DeadLetterQueueS3{BucketArn: aws.String(channelTestDLQARN)},
			Storage: &awstypes.S3Storage{
				BucketArn:       aws.String(channelTestBucketARN),
				CompressionType: awstypes.S3CompressionTypeNone,
				StorageClass:    awstypes.S3StorageClassStandard,
			},
		},
	}

	var model channelResourceModel
	if diags := fwflex.Flatten(ctx, output, &model); diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	if got, want := model.ChannelARN.ValueString(), aws.ToString(output.ChannelArn); got != want {
		t.Errorf("ChannelARN = %q, want %q", got, want)
	}
	if got, want := model.ChannelName.ValueString(), "example"; got != want {
		t.Errorf("ChannelName = %q, want %q", got, want)
	}
	if got, want := model.DestinationType.ValueString(), string(awstypes.ChannelDestinationTypeS3); got != want {
		t.Errorf("DestinationType = %q, want %q", got, want)
	}
	if got, want := model.Status.ValueString(), string(awstypes.ChannelStatusActive); got != want {
		t.Errorf("Status = %q, want %q", got, want)
	}
	if model.S3DestinationConfiguration.IsNull() {
		t.Error("S3DestinationConfiguration is null, want set")
	}
	if !model.IcebergDestinationConfiguration.IsNull() {
		t.Error("IcebergDestinationConfiguration is set, want null")
	}
}
