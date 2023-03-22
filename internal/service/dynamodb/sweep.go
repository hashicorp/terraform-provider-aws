//go:build sweep
// +build sweep

package dynamodb

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func init() {
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
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.(*conns.AWSClient).DynamoDBConn()
	sweepResources := make([]sweep.Sweepable, 0)
	var errs *multierror.Error
	var g multierror.Group
	var mutex = &sync.Mutex{}

	err = conn.ListTablesPagesWithContext(ctx, &dynamodb.ListTablesInput{}, func(page *dynamodb.ListTablesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, tableName := range page.TableNames {
			r := ResourceTable()
			d := r.Data(nil)

			id := aws.StringValue(tableName)
			d.SetId(id)

			// read concurrently and gather errors
			g.Go(func() error {
				// Need to Read first to fill in `replica` attribute
				err := sweep.ReadResource(ctx, r, d, client)

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

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error listing DynamoDB Tables for %s: %w", region, err))
	}

	if err = g.Wait().ErrorOrNil(); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error concurrently reading DynamoDB Tables: %w", err))
	}

	if err = sweep.SweepOrchestratorWithContext(ctx, sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping DynamoDB Tables for %s: %w", region, err))
	}

	if sweep.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping DynamoDB Tables sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func sweepBackups(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.(*conns.AWSClient).DynamoDBConn()
	sweepables := make([]sweep.Sweepable, 0)
	var errs *multierror.Error
	var g multierror.Group

	input := &dynamodb.ListBackupsInput{
		BackupType: aws.String(dynamodb.BackupTypeFilterAll),
	}
	err = listBackupsPages(ctx, conn, input, func(page *dynamodb.ListBackupsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, backup := range page.BackupSummaries {
			if aws.StringValue(backup.BackupType) == dynamodb.BackupTypeFilterSystem {
				log.Printf("[DEBUG] Skipping DynamoDB Backup %q, cannot delete %q backups", aws.StringValue(backup.BackupArn), dynamodb.BackupTypeFilterSystem)
				continue
			}

			sweepables = append(sweepables, backupSweeper{
				conn: conn,
				arn:  backup.BackupArn,
			})
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("listing DynamoDB Backups for %s: %w", region, err))
	}

	if err = g.Wait().ErrorOrNil(); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("reading DynamoDB Backups: %w", err))
	}

	if err = sweep.SweepOrchestratorWithContext(ctx, sweepables); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("sweeping DynamoDB Backups for %s: %w", region, err))
	}

	if sweep.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping DynamoDB Backups sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

type backupSweeper struct {
	conn *dynamodb.DynamoDB
	arn  *string
}

func (bs backupSweeper) Delete(ctx context.Context, timeout time.Duration, optFns ...tfresource.OptionsFunc) error {
	input := &dynamodb.DeleteBackupInput{
		BackupArn: bs.arn,
	}
	err := tfresource.Retry(ctx, timeout, func() *resource.RetryError {
		_, err := bs.conn.DeleteBackupWithContext(ctx, input)
		if tfawserr.ErrCodeEquals(err, dynamodb.ErrCodeBackupNotFoundException) {
			return nil
		}
		if tfawserr.ErrCodeEquals(err, dynamodb.ErrCodeBackupInUseException, dynamodb.ErrCodeLimitExceededException) {
			return resource.RetryableError(err)
		}
		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	}, optFns...)
	if tfresource.TimedOut(err) {
		_, err = bs.conn.DeleteBackupWithContext(ctx, input)
	}

	return err
}
