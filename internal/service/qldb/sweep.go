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
