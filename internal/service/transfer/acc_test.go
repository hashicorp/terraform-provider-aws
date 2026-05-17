// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package transfer_test

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/transfer"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).TransferClient(ctx)

	input := &transfer.ListServersInput{}

	_, err := conn.ListServers(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}
