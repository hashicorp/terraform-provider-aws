// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	awstypes "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @Action(aws_s3_empty_bucket, name="Empty Bucket")
func newEmptyBucketAction(_ context.Context) (action.ActionWithConfigure, error) {
	return &emptyBucketAction{}, nil
}

type emptyBucketAction struct {
	framework.ActionWithModel[emptyBucketModel]
}

type emptyBucketModel struct {
	framework.WithRegionModel
	BucketName types.String `tfsdk:"bucket_name"`
	Prefix     types.String `tfsdk:"prefix"`
	BatchSize  types.Int64  `tfsdk:"batch_size"`
	Timeout    types.Int64  `tfsdk:"timeout"`
}

func (a *emptyBucketAction) Schema(ctx context.Context, req action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Empties an S3 bucket by deleting all objects and versions. Useful for preparing buckets for deletion or cleanup operations.",
		Attributes: map[string]schema.Attribute{
			"bucket_name": schema.StringAttribute{
				Description: "Name of the S3 bucket to empty",
				Required:    true,
			},
			"prefix": schema.StringAttribute{
				Description: "Only delete objects whose keys begin with this prefix. If not specified, all objects will be deleted.",
				Optional:    true,
			},
			"batch_size": schema.Int64Attribute{
				Description: "Number of objects to delete per batch operation (1-1000, default: 1000)",
				Optional:    true,
				Validators: []validator.Int64{
					int64validator.Between(1, 1000),
				},
			},
			names.AttrTimeout: schema.Int64Attribute{
				Description: "Timeout in seconds for the empty operation (60-7200, default: 1800)",
				Optional:    true,
				Validators: []validator.Int64{
					int64validator.Between(60, 7200),
				},
			},
		},
	}
}

func (a *emptyBucketAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var config emptyBucketModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := a.Meta().S3Client(ctx)
	bucketName := config.BucketName.ValueString()

	// Set defaults
	batchSize := int32(1000)
	if !config.BatchSize.IsNull() {
		batchSize = int32(config.BatchSize.ValueInt64())
	}

	timeout := 1800 * time.Second
	if !config.Timeout.IsNull() {
		timeout = time.Duration(config.Timeout.ValueInt64()) * time.Second
	}

	var prefix string
	if !config.Prefix.IsNull() {
		prefix = config.Prefix.ValueString()
	}

	tflog.Info(ctx, "Starting S3 empty bucket action", map[string]any{
		"bucket_name": bucketName,
		"prefix":      prefix,
		"batch_size":  batchSize,
		"timeout":     timeout.String(),
	})

	// Check if bucket exists
	resp.SendProgress(action.InvokeProgressEvent{
		Message: fmt.Sprintf("Checking if S3 bucket %s exists...", bucketName),
	})

	_, err := conn.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Bucket Not Found",
			fmt.Sprintf("S3 bucket %s was not found or is not accessible: %s", bucketName, err),
		)
		return
	}

	// Check if bucket has versioning enabled
	versioningEnabled, err := a.checkBucketVersioning(ctx, conn, bucketName)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to Check Bucket Versioning",
			fmt.Sprintf("Could not check versioning status for bucket %s: %s", bucketName, err),
		)
		return
	}

	startTime := time.Now()
	totalDeleted := 0

	if prefix != "" {
		resp.SendProgress(action.InvokeProgressEvent{
			Message: fmt.Sprintf("Starting to empty S3 bucket %s with prefix '%s'...", bucketName, prefix),
		})
	} else {
		resp.SendProgress(action.InvokeProgressEvent{
			Message: fmt.Sprintf("Starting to empty S3 bucket %s...", bucketName),
		})
	}

	for {
		if time.Since(startTime) > timeout {
			resp.Diagnostics.AddError(
				"Timeout Emptying Bucket",
				fmt.Sprintf("Bucket emptying operation timed out after %v. Deleted %d objects.", timeout, totalDeleted),
			)
			return
		}

		var objectIds []awstypes.ObjectIdentifier
		var err error

		if versioningEnabled {
			objectIds, err = a.listObjectVersions(ctx, conn, bucketName, prefix, batchSize)
		} else {
			objectIds, err = a.listObjects(ctx, conn, bucketName, prefix, batchSize)
		}

		if err != nil {
			resp.Diagnostics.AddError(
				"Failed to List Objects",
				fmt.Sprintf("Could not list objects in bucket %s: %s", bucketName, err),
			)
			return
		}

		if len(objectIds) == 0 {
			break // No more objects to delete
		}

		// Delete batch
		deleted, err := a.deleteObjects(ctx, conn, bucketName, objectIds)
		if err != nil {
			resp.Diagnostics.AddError(
				"Failed to Delete Objects",
				fmt.Sprintf("Could not delete objects from bucket %s: %s", bucketName, err),
			)
			return
		}

		totalDeleted += deleted
		
		if prefix != "" {
			resp.SendProgress(action.InvokeProgressEvent{
				Message: fmt.Sprintf("Deleted %d objects from S3 bucket %s with prefix '%s' (total: %d)", deleted, bucketName, prefix, totalDeleted),
			})
		} else {
			resp.SendProgress(action.InvokeProgressEvent{
				Message: fmt.Sprintf("Deleted %d objects from S3 bucket %s (total: %d)", deleted, bucketName, totalDeleted),
			})
		}
	}

	if prefix != "" {
		resp.SendProgress(action.InvokeProgressEvent{
			Message: fmt.Sprintf("Successfully emptied S3 bucket %s with prefix '%s', deleted %d objects", bucketName, prefix, totalDeleted),
		})
	} else {
		resp.SendProgress(action.InvokeProgressEvent{
			Message: fmt.Sprintf("Successfully emptied S3 bucket %s, deleted %d objects", bucketName, totalDeleted),
		})
	}

	tflog.Info(ctx, "S3 empty bucket action completed successfully", map[string]any{
		"bucket_name":    bucketName,
		"total_deleted":  totalDeleted,
		"duration":       time.Since(startTime).String(),
	})
}

func (a *emptyBucketAction) checkBucketVersioning(ctx context.Context, conn *s3.Client, bucketName string) (bool, error) {
	input := &s3.GetBucketVersioningInput{
		Bucket: aws.String(bucketName),
	}

	output, err := conn.GetBucketVersioning(ctx, input)
	if err != nil {
		return false, err
	}

	return output.Status == awstypes.BucketVersioningStatusEnabled, nil
}

func (a *emptyBucketAction) listObjects(ctx context.Context, conn *s3.Client, bucketName, prefix string, maxKeys int32) ([]awstypes.ObjectIdentifier, error) {
	input := &s3.ListObjectsV2Input{
		Bucket:  aws.String(bucketName),
		MaxKeys: aws.Int32(maxKeys),
	}

	if prefix != "" {
		input.Prefix = aws.String(prefix)
	}

	output, err := conn.ListObjectsV2(ctx, input)
	if err != nil {
		return nil, err
	}

	var objectIds []awstypes.ObjectIdentifier
	for _, obj := range output.Contents {
		objectIds = append(objectIds, awstypes.ObjectIdentifier{
			Key: obj.Key,
		})
	}

	return objectIds, nil
}

func (a *emptyBucketAction) listObjectVersions(ctx context.Context, conn *s3.Client, bucketName, prefix string, maxKeys int32) ([]awstypes.ObjectIdentifier, error) {
	input := &s3.ListObjectVersionsInput{
		Bucket:  aws.String(bucketName),
		MaxKeys: aws.Int32(maxKeys),
	}

	if prefix != "" {
		input.Prefix = aws.String(prefix)
	}

	output, err := conn.ListObjectVersions(ctx, input)
	if err != nil {
		return nil, err
	}

	var objectIds []awstypes.ObjectIdentifier

	// Add object versions
	for _, version := range output.Versions {
		objectIds = append(objectIds, awstypes.ObjectIdentifier{
			Key:       version.Key,
			VersionId: version.VersionId,
		})
	}

	// Add delete markers
	for _, marker := range output.DeleteMarkers {
		objectIds = append(objectIds, awstypes.ObjectIdentifier{
			Key:       marker.Key,
			VersionId: marker.VersionId,
		})
	}

	return objectIds, nil
}

func (a *emptyBucketAction) deleteObjects(ctx context.Context, conn *s3.Client, bucketName string, objectIds []awstypes.ObjectIdentifier) (int, error) {
	if len(objectIds) == 0 {
		return 0, nil
	}

	input := &s3.DeleteObjectsInput{
		Bucket: aws.String(bucketName),
		Delete: &awstypes.Delete{
			Objects: objectIds,
			Quiet:   aws.Bool(true),
		},
	}

	output, err := conn.DeleteObjects(ctx, input)
	if err != nil {
		return 0, err
	}

	// Check for errors in deletion
	if len(output.Errors) > 0 {
		return len(output.Deleted), fmt.Errorf("failed to delete %d objects", len(output.Errors))
	}

	return len(output.Deleted), nil
}
