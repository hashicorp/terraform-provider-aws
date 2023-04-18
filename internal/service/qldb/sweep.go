//go:build sweep
// +build sweep

package qldb

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/qldb"
	multierror "github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_qldb_ledger", &resource.Sweeper{
		Name: "aws_qldb_ledger",
		F:    sweepLedgers,
		Dependencies: []string{
			"aws_qldb_stream",
		},
	})

	resource.AddTestSweepers("aws_qldb_stream", &resource.Sweeper{
		Name: "aws_qldb_stream",
		F:    sweepStreams,
	})

}

func sweepLedgers(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.(*conns.AWSClient).QLDBConn()
	input := &qldb.ListLedgersInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = conn.ListLedgersPagesWithContext(ctx, input, func(page *qldb.ListLedgersOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Ledgers {
			r := ResourceLedger()
			d := r.Data(nil)
			d.SetId(aws.StringValue(v.Name))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping QLDB Ledger sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing QLDB Ledgers (%s): %w", region, err)
	}

	err = sweep.SweepOrchestratorWithContext(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping QLDB Ledgers (%s): %w", region, err)
	}

	return nil
}

func sweepStreams(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.(*conns.AWSClient).QLDBConn()
	input := &qldb.ListLedgersInput{}
	var sweeperErrs *multierror.Error
	sweepResources := make([]sweep.Sweepable, 0)

	err = conn.ListLedgersPagesWithContext(ctx, input, func(page *qldb.ListLedgersOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Ledgers {
			input := &qldb.ListJournalKinesisStreamsForLedgerInput{
				LedgerName: v.Name,
			}

			err := conn.ListJournalKinesisStreamsForLedgerPagesWithContext(ctx, input, func(page *qldb.ListJournalKinesisStreamsForLedgerOutput, lastPage bool) bool {
				if page == nil {
					return !lastPage
				}

				for _, v := range page.Streams {
					r := ResourceStream()
					d := r.Data(nil)
					d.SetId(aws.StringValue(v.StreamId))
					d.Set("ledger_name", v.LedgerName)

					sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
				}

				return !lastPage
			})

			if sweep.SkipSweepError(err) {
				continue
			}

			if err != nil {
				sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing QLDB Streams (%s): %w", region, err))
			}
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping QLDB Stream sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing QLDB Ledgers (%s): %w", region, err))
	}

	err = sweep.SweepOrchestratorWithContext(ctx, sweepResources)

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error sweeping QLDB Streams (%s): %w", region, err))
	}

	return sweeperErrs.ErrorOrNil()
}
