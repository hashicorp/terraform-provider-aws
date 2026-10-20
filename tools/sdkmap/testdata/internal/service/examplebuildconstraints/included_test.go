// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package examplebuildconstraints_test

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/examplebuildconstraints"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccExampleBuildConstraintsIncluded_basic(t *testing.T) {
	ctx := acctest.Context(t)

	acctest.ParallelTest(ctx, t, acctest.TestCase{
		PreCheck:     func() { acctest.PreCheck(ctx, t) },
		CheckDestroy: testAccCheckIncludedDestroy(ctx, t),
	})
}

func testAccCheckIncludedDestroy(ctx context.Context, t *testing.T) func() error {
	return func() error {
		conn := acctest.ProviderMeta(ctx, t).ExampleBuildConstraintsClient(ctx)

		input := examplebuildconstraints.ListThingsInput{
			NamePrefix: aws.String(acctest.ResourcePrefix),
		}

		_, err := conn.ListThings(ctx, &input)

		return err
	}
}
