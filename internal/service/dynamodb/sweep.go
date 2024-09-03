// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dynamodb

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	awstypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/sdk"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func RegisterSweepers() {
	resource.AddTestSweepers("aws_dynamodb_table", &resource.Sweeper{
		Name: "aws_dynamodb_table",
		F:    sweepTables,
	})

	resource.AddTestSweepers("aws_dynamodb_backup", &resource.Sweeper{
		Name: "aws_dynamodb_backup",
		F:    sweepBackups,
	})
}

func sweepTables(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.DynamoDBClient(ctx)
	input := &dynamodb.ListTablesInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := dynamodb.NewListTablesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping DynamoDB Table sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing DynamoDB Tables (%s): %w", region, err)
		}

		for _, v := range page.TableNames {
			_, err := conn.UpdateTable(ctx, &dynamodb.UpdateTableInput{
				DeletionProtectionEnabled: aws.Bool(false),
				TableName:                 aws.String(v),
			})

			if err != nil {
				log.Printf("[WARN] DynamoDB Table (%s): %s", v, err)
			}

			r := resourceTable()
			d := r.Data(nil)
			d.SetId(v)
			// Refresh replicas.
			if err := sdk.ReadResource(ctx, r, d, client); err != nil {
				log.Printf("[WARN] Skipping DynamoDB Table %s: %s", v, err)
				continue
			}
			if d.Id() == "" {
				continue
			}

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping DynamoDB Tables (%s): %w", region, err)
	}

	return nil
}

func sweepBackups(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.DynamoDBClient(ctx)
	input := &dynamodb.ListBackupsInput{
		BackupType: awstypes.BackupTypeFilterAll,
	}
	sweepResources := make([]sweep.Sweepable, 0)

	err = listBackupsPages(ctx, conn, input, func(page *dynamodb.ListBackupsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.BackupSummaries {
			arn := aws.ToString(v.BackupArn)

			if v.BackupType == awstypes.BackupTypeSystem {
				log.Printf("[DEBUG] Skipping DynamoDB Backup %s: BackupType=%s", arn, v.BackupType)
				continue
			}

			sweepResources = append(sweepResources, backupSweeper{
				conn: conn,
				arn:  arn,
			})
		}

		return !lastPage
	})

	if awsv2.SkipSweepError(err) {
		log.Printf("[WARN] Skipping DynamoDB Backup sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing DynamoDB Backups (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping DynamoDB Backups (%s): %w", region, err)
	}

	return nil
}

type backupSweeper struct {
	conn *dynamodb.Client
	arn  string
}

func (bs backupSweeper) Delete(ctx context.Context, timeout time.Duration, optFns ...tfresource.OptionsFunc) error {
	input := &dynamodb.DeleteBackupInput{
		BackupArn: aws.String(bs.arn),
	}

	err := tfresource.Retry(ctx, timeout, func() *retry.RetryError {
		log.Printf("[DEBUG] Deleting DynamoDB Backup: %s", bs.arn)
		_, err := bs.conn.DeleteBackup(ctx, input)
		if errs.IsA[*awstypes.BackupNotFoundException](err) {
			return nil
		}
		if errs.IsA[*awstypes.BackupInUseException](err) || errs.IsA[*awstypes.LimitExceededException](err) {
			return retry.RetryableError(err)
		}
		if err != nil {
			return retry.NonRetryableError(err)
		}

		return nil
	}, optFns...)
	if tfresource.TimedOut(err) {
		_, err = bs.conn.DeleteBackup(ctx, input)
	}

	return err
}
