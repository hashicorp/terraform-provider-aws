// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dynamodb

import (
	"context"
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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
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
		},
	}
}

func (a *createBackupAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var config createBackupActionModel

	// Parse configuration
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get AWS client
	conn := a.Meta().DynamoDBClient(ctx)

	// Extract table name and backup name from configuration
	tableName := config.TableName.ValueString()
	backupName := config.BackupName.ValueString()

	// Generate backup name if not provided
	if backupName == "" {
		backupName = fmt.Sprintf("%s-backup-%s", tableName, id.UniqueId())
	}

	// Log operation start
	tflog.Info(ctx, "Starting DynamoDB create backup action", map[string]any{
		names.AttrTableName: tableName,
		"backup_name":       backupName,
	})

	// Send initial progress message
	resp.SendProgress(action.InvokeProgressEvent{
		Message: fmt.Sprintf("Starting backup creation for DynamoDB table %s...", tableName),
	})

	// Build CreateBackup input
	input := &dynamodb.CreateBackupInput{
		TableName:  aws.String(tableName),
		BackupName: aws.String(backupName),
	}

	// Call CreateBackup API with retry logic for transient errors
	const (
		createBackupTimeout = 10 * time.Minute
	)
	var output *dynamodb.CreateBackupOutput
	var err error
	for l := backoff.NewLoop(createBackupTimeout); l.Continue(ctx); {
		output, err = conn.CreateBackup(ctx, input)

		if err != nil {
			// Retry if backups are still being enabled for a newly created table
			if errs.IsAErrorMessageContains[*awstypes.ContinuousBackupsUnavailableException](err, "Backups are being enabled") {
				continue
			}

			// Retry if another backup operation is in progress or rate limit is exceeded
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

	// Extract backup details from API response
	backupDetails := output.BackupDetails
	backupArn := aws.ToString(backupDetails.BackupArn)

	// Send success progress message with backup ARN
	resp.SendProgress(action.InvokeProgressEvent{
		Message: fmt.Sprintf("Backup %s created successfully for DynamoDB table %s", backupArn, tableName),
	})

	// Output detailed backup information
	backupInfo := fmt.Sprintf("Backup Details:\n"+
		"  BackupArn: %s\n"+
		"  BackupCreationDateTime: %s\n"+
		"  BackupStatus: %s\n"+
		"  BackupType: %s",
		backupArn,
		backupDetails.BackupCreationDateTime.Format(time.RFC3339),
		backupDetails.BackupStatus,
		backupDetails.BackupType,
	)

	// Add optional fields if present
	if backupDetails.BackupExpiryDateTime != nil {
		backupInfo += fmt.Sprintf("\n  BackupExpiryDateTime: %s", backupDetails.BackupExpiryDateTime.Format(time.RFC3339))
	}
	if backupDetails.BackupSizeBytes != nil {
		backupInfo += fmt.Sprintf("\n  BackupSizeBytes: %d", *backupDetails.BackupSizeBytes)
	}

	resp.SendProgress(action.InvokeProgressEvent{
		Message: backupInfo,
	})

	// Log completion with table name and backup ARN
	tflog.Info(ctx, "DynamoDB create backup action completed successfully", map[string]any{
		names.AttrTableName: tableName,
		"backup_arn":        backupArn,
	})
}
