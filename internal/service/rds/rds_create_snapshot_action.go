// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	awstypes "github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/actionwait"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @Action(aws_rds_create_snapshot, name="Create DB Snapshot")
func newCreateSnapshotAction(_ context.Context) (action.ActionWithConfigure, error) {
	return &createSnapshotAction{}, nil
}

type createSnapshotAction struct {
	framework.ActionWithModel[createSnapshotModel]
}

type createSnapshotModel struct {
	framework.WithRegionModel
	DBInstanceIdentifier types.String            `tfsdk:"db_instance_identifier"`
	SnapshotIdentifier   types.String            `tfsdk:"snapshot_identifier"`
	Tags                 fwtypes.MapOfString     `tfsdk:"tags"`
	Timeout              types.Int64             `tfsdk:"timeout"`
}

func (a *createSnapshotAction) Schema(ctx context.Context, req action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Creates a manual snapshot of an Amazon RDS database instance for backup and disaster recovery purposes.",
		Attributes: map[string]schema.Attribute{
			"db_instance_identifier": schema.StringAttribute{
				Description: "The DB instance identifier of the RDS instance to create a snapshot of.",
				Required:    true,
			},
			"snapshot_identifier": schema.StringAttribute{
				Description: "The identifier for the DB snapshot. Must be unique within your AWS account in the selected region.",
				Required:    true,
			},
			names.AttrTags: schema.MapAttribute{
				CustomType:  fwtypes.MapOfStringType,
				Description: "A map of tags to assign to the snapshot.",
				Optional:    true,
				ElementType: types.StringType,
			},
			names.AttrTimeout: schema.Int64Attribute{
				Description: "Timeout in seconds to wait for the snapshot to complete (300-7200, default: 3600).",
				Optional:    true,
				Validators: []validator.Int64{
					int64validator.Between(300, 7200),
				},
			},
		},
	}
}

func (a *createSnapshotAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var config createSnapshotModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := a.Meta().RDSClient(ctx)

	dbInstanceId := config.DBInstanceIdentifier.ValueString()
	snapshotId := config.SnapshotIdentifier.ValueString()

	timeout := 3600 * time.Second
	if !config.Timeout.IsNull() {
		timeout = time.Duration(config.Timeout.ValueInt64()) * time.Second
	}

	tflog.Info(ctx, "Starting RDS create snapshot action", map[string]any{
		"db_instance_identifier": dbInstanceId,
		"snapshot_identifier":    snapshotId,
		names.AttrTimeout:        timeout.String(),
	})

	resp.SendProgress(action.InvokeProgressEvent{
		Message: fmt.Sprintf("Creating snapshot %s for RDS instance %s...", snapshotId, dbInstanceId),
	})

	// Create snapshot
	input := &rds.CreateDBSnapshotInput{
		DBInstanceIdentifier: aws.String(dbInstanceId),
		DBSnapshotIdentifier: aws.String(snapshotId),
	}

	// Add tags if specified
	if !config.Tags.IsNull() {
		var tags []awstypes.Tag
		for k, v := range config.Tags.Elements() {
			tags = append(tags, awstypes.Tag{
				Key:   aws.String(k),
				Value: aws.String(v.(types.String).ValueString()),
			})
		}
		input.Tags = tags
	}

	_, err := conn.CreateDBSnapshot(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to Create DB Snapshot",
			fmt.Sprintf("Could not create snapshot %s for RDS instance %s: %s", snapshotId, dbInstanceId, err),
		)
		return
	}

	resp.SendProgress(action.InvokeProgressEvent{
		Message: fmt.Sprintf("Snapshot %s creation initiated, waiting for completion...", snapshotId),
	})

	// Wait for snapshot to be available
	_, err = actionwait.WaitForStatus(ctx, func(ctx context.Context) (actionwait.FetchResult[struct{}], error) {
		input := rds.DescribeDBSnapshotsInput{
			DBSnapshotIdentifier: aws.String(snapshotId),
		}
		output, err := conn.DescribeDBSnapshots(ctx, &input)
		if err != nil {
			return actionwait.FetchResult[struct{}]{}, fmt.Errorf("describing snapshot: %w", err)
		}
		if len(output.DBSnapshots) == 0 {
			return actionwait.FetchResult[struct{}]{}, fmt.Errorf("snapshot %s not found", snapshotId)
		}
		status := aws.ToString(output.DBSnapshots[0].Status)
		return actionwait.FetchResult[struct{}]{Status: actionwait.Status(status)}, nil
	}, actionwait.Options[struct{}]{
		Timeout:          timeout,
		Interval:         actionwait.FixedInterval(30 * time.Second),
		ProgressInterval: 2 * time.Minute,
		SuccessStates:    []actionwait.Status{"available"},
		TransitionalStates: []actionwait.Status{
			"creating",
		},
		FailureStates: []actionwait.Status{
			"error",
			"failed",
		},
		ProgressSink: func(fr actionwait.FetchResult[any], meta actionwait.ProgressMeta) {
			resp.SendProgress(action.InvokeProgressEvent{
				Message: fmt.Sprintf("Snapshot %s is currently '%s', continuing to wait...", snapshotId, fr.Status),
			})
		},
	})

	if err != nil {
		var timeoutErr *actionwait.TimeoutError
		var failureErr *actionwait.FailureStateError
		var unexpectedErr *actionwait.UnexpectedStateError

		if errors.As(err, &timeoutErr) {
			resp.Diagnostics.AddError(
				"Timeout Creating DB Snapshot",
				fmt.Sprintf("Snapshot %s did not complete within %s", snapshotId, timeout),
			)
		} else if errors.As(err, &failureErr) {
			resp.Diagnostics.AddError(
				"DB Snapshot Creation Failed",
				fmt.Sprintf("Snapshot %s failed with status: %s", snapshotId, failureErr.Status),
			)
		} else if errors.As(err, &unexpectedErr) {
			resp.Diagnostics.AddError(
				"Unexpected Snapshot Status",
				fmt.Sprintf("Snapshot %s entered unexpected status: %s", snapshotId, unexpectedErr.Status),
			)
		} else {
			resp.Diagnostics.AddError(
				"Error Creating DB Snapshot",
				fmt.Sprintf("Error while creating snapshot %s: %s", snapshotId, err),
			)
		}
		return
	}

	resp.SendProgress(action.InvokeProgressEvent{
		Message: fmt.Sprintf("Snapshot %s created successfully for RDS instance %s", snapshotId, dbInstanceId),
	})

	tflog.Info(ctx, "RDS create snapshot action completed successfully", map[string]any{
		"db_instance_identifier": dbInstanceId,
		"snapshot_identifier":    snapshotId,
	})
}
