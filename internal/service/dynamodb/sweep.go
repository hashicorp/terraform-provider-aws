// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dynamodb

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	awstypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/hashicorp/go-multierror"
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
	var errs *multierror.Error
	var g multierror.Group
	var mutex = &sync.Mutex{}

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

			// read concurrently and gather errors
			g.Go(func() error {
				// Need to Read first to fill in `replica` attribute
				err := sdk.ReadResource(ctx, r, d, client)

				if err != nil {
					return err
				}

				// In case it was already deleted
				if d.Id() == "" {
					return nil
				}

				mutex.Lock()
				defer mutex.Unlock()
				sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))

				return nil
			})
		}
	}

	if err := g.Wait().ErrorOrNil(); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error concurrently reading DynamoDB Tables: %w", err))
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping DynamoDB Tables (%s): %w", region, err))
	}

	return errs.ErrorOrNil()
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
	sweepables := make([]sweep.Sweepable, 0)
	var errs *multierror.Error
	var g multierror.Group

	err = listBackupsPages(ctx, conn, input, func(page *dynamodb.ListBackupsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.BackupSummaries {
			if v.BackupType == awstypes.BackupTypeSystem {
				log.Printf("[DEBUG] Skipping DynamoDB Backup %q, cannot delete %q backups", aws.ToString(v.BackupArn), v.BackupType)
				continue
			}

			sweepables = append(sweepables, backupSweeper{
				conn: conn,
				arn:  v.BackupArn,
			})
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("listing DynamoDB Backups (%s): %w", region, err))
	}

	if err := g.Wait().ErrorOrNil(); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("reading DynamoDB Backups: %w", err))
	}

	if err := sweep.SweepOrchestrator(ctx, sweepables); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("sweeping DynamoDB Backups (%s): %w", region, err))
	}

	return errs.ErrorOrNil()
}

type backupSweeper struct {
	conn *dynamodb.Client
	arn  *string
}

func (bs backupSweeper) Delete(ctx context.Context, timeout time.Duration, optFns ...tfresource.OptionsFunc) error {
	input := &dynamodb.DeleteBackupInput{
		BackupArn: bs.arn,
	}
	err := tfresource.Retry(ctx, timeout, func() *retry.RetryError {
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
