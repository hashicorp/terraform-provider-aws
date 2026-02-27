// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package dynamodb

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	awstypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	sdkid "github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-provider-aws/internal/actionwait"
	"github.com/hashicorp/terraform-provider-aws/internal/backoff"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @Action(aws_dynamodb_create_backup, name="Create Backup")
func newCreateBackupAction(_ context.Context) (action.ActionWithConfigure, error) {
	return &createBackupAction{}, nil
}

var (
	_ action.Action = (*createBackupAction)(nil)
)

type createBackupAction struct {
	framework.ActionWithModel[createBackupActionModel]
}

type createBackupActionModel struct {
	framework.WithRegionModel
	TableName  types.String `tfsdk:"table_name"`
	BackupName types.String `tfsdk:"backup_name"`
	Timeout    types.Int64  `tfsdk:"timeout"`
}

func (a *createBackupAction) Schema(ctx context.Context, req action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Creates an on-demand backup of a DynamoDB table. The backup is created asynchronously and typically completes within minutes.",
		Attributes: map[string]schema.Attribute{
			names.AttrTableName: schema.StringAttribute{
				Description: "The name or ARN of the DynamoDB table to backup",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 1024),
				},
			},
			"backup_name": schema.StringAttribute{
				Description: "Name for the backup. If not provided, a name will be generated automatically using the table name and a unique identifier",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(3, 255),
					stringvalidator.RegexMatches(
						regexache.MustCompile(`^[a-zA-Z0-9_.-]+$`),
						"must contain only alphanumeric characters, underscores, periods, and hyphens",
					),
				},
			},
			names.AttrTimeout: schema.Int64Attribute{
				Description: "Timeout in minutes for the backup operation. Defaults to 10 minutes",
				Optional:    true,
			},
		},
	}
}

func (a *createBackupAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var config createBackupActionModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := a.Meta().DynamoDBClient(ctx)

	timeout := 10 * time.Minute
	if !config.Timeout.IsNull() {
		timeout = time.Duration(config.Timeout.ValueInt64()) * time.Minute
	}

	tableName := config.TableName.ValueString()
	backupName := config.BackupName.ValueString()

	if backupName == "" {
		backupName = fmt.Sprintf("%s-backup-%s", tableName, sdkid.UniqueId())
	}

	tflog.Info(ctx, "Starting DynamoDB create backup action", map[string]any{
		names.AttrTableName: tableName,
		"backup_name":       backupName,
	})

	resp.SendProgress(action.InvokeProgressEvent{
		Message: fmt.Sprintf("Starting backup creation for DynamoDB table %s...", tableName),
	})

	input := &dynamodb.CreateBackupInput{
		TableName:  aws.String(tableName),
		BackupName: aws.String(backupName),
	}

	var output *dynamodb.CreateBackupOutput
	var err error
	for l := backoff.NewLoop(timeout); l.Continue(ctx); {
		output, err = conn.CreateBackup(ctx, input)

		if err != nil {
			if errs.IsAErrorMessageContains[*awstypes.ContinuousBackupsUnavailableException](err, "Backups are being enabled") {
				continue
			}

			if errs.IsA[*awstypes.BackupInUseException](err) || errs.IsA[*awstypes.LimitExceededException](err) {
				continue
			}
		}

		break
	}

	if err != nil {
		resp.Diagnostics.AddError("creating DynamoDB backup", err.Error())
		return
	}

	backupArn := aws.ToString(output.BackupDetails.BackupArn)

	resp.SendProgress(action.InvokeProgressEvent{
		Message: "Backup started, waiting for completion...",
	})

	result, err := actionwait.WaitForStatus(ctx, func(ctx context.Context) (actionwait.FetchResult[*awstypes.BackupDescription], error) {
		input := &dynamodb.DescribeBackupInput{BackupArn: aws.String(backupArn)}
		output, err := conn.DescribeBackup(ctx, input)
		if err != nil {
			return actionwait.FetchResult[*awstypes.BackupDescription]{}, err
		}
		desc := output.BackupDescription
		return actionwait.FetchResult[*awstypes.BackupDescription]{Status: actionwait.Status(desc.BackupDetails.BackupStatus), Value: desc}, nil
	}, actionwait.Options[*awstypes.BackupDescription]{
		Timeout:            timeout,
		Interval:           actionwait.WithBackoffDelay(backoff.DefaultSDKv2HelperRetryCompatibleDelay()),
		ProgressInterval:   30 * time.Second,
		SuccessStates:      []actionwait.Status{actionwait.Status(awstypes.BackupStatusAvailable)},
		TransitionalStates: []actionwait.Status{actionwait.Status(awstypes.BackupStatusCreating)},
		FailureStates:      []actionwait.Status{actionwait.Status(awstypes.BackupStatusDeleted)},
		ProgressSink: func(fr actionwait.FetchResult[any], meta actionwait.ProgressMeta) {
			resp.SendProgress(action.InvokeProgressEvent{Message: "Backup currently in state: " + string(fr.Status)})
		},
	})
	if err != nil {
		var timeoutErr *actionwait.TimeoutError
		var failureErr *actionwait.FailureStateError
		var unexpectedErr *actionwait.UnexpectedStateError
		if errors.As(err, &timeoutErr) {
			resp.Diagnostics.AddError("Backup timeout", "Backup did not complete within the specified timeout")
		} else if errors.As(err, &failureErr) {
			resp.Diagnostics.AddError("Backup failed", "Backup completed with status: "+err.Error())
		} else if errors.As(err, &unexpectedErr) {
			resp.Diagnostics.AddError("Unexpected backup status", err.Error())
		} else {
			resp.Diagnostics.AddError("Error waiting for backup", err.Error())
		}
		return
	}

	backupDetails := result.Value.BackupDetails
	backupInfo := fmt.Sprintf("Backup completed successfully\n"+
		"  ARN: %s\n"+
		"  Created: %s\n"+
		"  Size: %d bytes",
		aws.ToString(backupDetails.BackupArn),
		backupDetails.BackupCreationDateTime.Format(time.RFC3339),
		aws.ToInt64(backupDetails.BackupSizeBytes),
	)
	resp.SendProgress(action.InvokeProgressEvent{Message: backupInfo})

	tflog.Info(ctx, "DynamoDB create backup action completed successfully", map[string]any{
		names.AttrTableName: tableName,
		"backup_arn":        backupArn,
	})
}
