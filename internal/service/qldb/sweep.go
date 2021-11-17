//go:build sweep
// +build sweep

package qldb

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/qldb"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_qldb_ledger", &resource.Sweeper{
		Name: "aws_qldb_ledger",
		F:    sweepLedgers,
	})
	resource.AddTestSweepers("aws_qldb_stream", &resource.Sweeper{
		Name: "aws_qldb_stream",
		F:    testSweepQLDBStreams,
	})

}

func sweepLedgers(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.(*conns.AWSClient).QLDBConn
	input := &qldb.ListLedgersInput{}
	page, err := conn.ListLedgers(input)

	if err != nil {
		if sweep.SkipSweepError(err) {
			log.Printf("[WARN] Skipping QLDB Ledger sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error listing QLDB Ledgers: %s", err)
	}

	for _, item := range page.Ledgers {
		input := &qldb.DeleteLedgerInput{
			Name: item.Name,
		}
		name := aws.StringValue(item.Name)

		log.Printf("[INFO] Deleting QLDB Ledger: %s", name)
		_, err = conn.DeleteLedger(input)

		if err != nil {
			log.Printf("[ERROR] Failed to delete QLDB Ledger %s: %s", name, err)
			continue
		}

		if err := WaitForLedgerDeletion(conn, name); err != nil {
			log.Printf("[ERROR] Error waiting for QLDB Ledger (%s) deletion: %s", name, err)
		}
	}

	return nil
}

func testSweepQLDBStreams(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)(region)

	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.(*AWSClient).qldbconn
	input := &qldb.ListLedgersInput{}
	page, err := conn.ListLedgers(input)

	if err != nil {
		if sweep.SkipSweepError(err) {
			log.Printf("[WARN] Skipping QLDB Stream sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error listing QLDB Ledgers for QLDB Stream Sweep: %s", err)
	}

	for _, item := range page.Ledgers {
		ledgerName := aws.StringValue(item.Name)
		err := testSweepQLDBLedgerStreams(conn, region, ledgerName)
		if err != nil {
			log.Printf("[ERROR] Failed to delete QLDB Stream for Ledger %s: %s", ledgerName, err)
			continue
		}
	}

	return nil
}

func testSweepQLDBLedgerStreams(conn *qldb.QLDB, region string, ledgerName string) error {
	input := &qldb.ListJournalKinesisStreamsForLedgerInput{
		LedgerName: aws.String(ledgerName),
	}
	page, err := conn.ListJournalKinesisStreamsForLedger(input)

	if err != nil {
		if sweep.SkipSweepError(err) {
			log.Printf("[WARN] Skipping QLDB Stream sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error listing QLDB Streams: %s", err)
	}

	for _, item := range page.Streams {
		input := &qldb.CancelJournalKinesisStreamInput{
			LedgerName: item.LedgerName,
			StreamId:   item.StreamId,
		}

		ledgerName := aws.StringValue(item.LedgerName)
		streamID := aws.StringValue(item.StreamId)

		log.Printf("[INFO] Cancelling QLDB Stream: (%s, %s)", ledgerName, streamID)
		_, err = conn.CancelJournalKinesisStream(input)

		if err != nil {
			log.Printf("[ERROR] Failed to cancel QLDB Ledger: (%s, %s): %s", ledgerName, streamID, err)
			continue
		}

		if err := waitForQLDBStreamCancellation(conn, ledgerName, streamID); err != nil {
			log.Printf("[ERROR] Error waiting for QLDB Ledger (%s, %s) deletion: %s", ledgerName, streamID, err)
		}
	}

	return nil
}

func waitForQLDBStreamCancellation(conn *qldb.QLDB, ledgerName string, streamID string) error {
	stateConf := resource.StateChangeConf{
		Pending: []string{
			qldb.StreamStatusActive,
			qldb.StreamStatusImpaired,
		},
		Target: []string{
			qldb.StreamStatusCanceled,
			qldb.StreamStatusCompleted,
			qldb.StreamStatusFailed,
		},
		Timeout:    5 * time.Minute,
		MinTimeout: 1 * time.Second,
		Refresh: func() (interface{}, string, error) {
			resp, err := conn.DescribeJournalKinesisStream(&qldb.DescribeJournalKinesisStreamInput{
				LedgerName: aws.String(ledgerName),
				StreamId:   aws.String(streamID),
			})

			if tfawserr.ErrMessageContains(err, qldb.ErrCodeResourceNotFoundException, "") {
				return 1, "", nil
			}

			if err != nil {
				return nil, qldb.ErrCodeResourceInUseException, err
			}

			return resp, aws.StringValue(resp.Stream.Status), nil
		},
	}

	_, err := stateConf.WaitForState()

	return err
}
