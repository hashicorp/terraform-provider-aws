// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kms

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	awstypes "github.com/aws/aws-sdk-go-v2/service/kms/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/sdk"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
	resource.AddTestSweepers("aws_kms_key", &resource.Sweeper{
		Name: "aws_kms_key",
		F:    sweepKeys,
	})
}

func sweepKeys(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.KMSClient(ctx)
	input := &kms.ListKeysInput{
		Limit: aws.Int32(1000),
	}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := kms.NewListKeysPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping KMS Key sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing KMS Keys (%s): %w", region, err)
		}

		for _, v := range page.Keys {
			keyID := aws.ToString(v.KeyId)
			key, err := findKeyByID(ctx, conn, keyID)

			if tfresource.NotFound(err) {
				continue
			}

			if tfawserr.ErrMessageContains(err, "AccessDeniedException", "is not authorized to perform") {
				log.Printf("[DEBUG] Skipping KMS Key (%s): %s", keyID, err)
				continue
			}

			if err != nil {
				continue
			}

			if key.KeyManager == awstypes.KeyManagerTypeAws {
				log.Printf("[DEBUG] Skipping KMS Key (%s): managed by AWS", keyID)
				continue
			}

			r := resourceKey()
			d := r.Data(nil)
			d.SetId(keyID)
			d.Set(names.AttrKeyID, keyID)
			d.Set("deletion_window_in_days", 7) //nolint:mnd // 7 days is the minimum value

			sweepResources = append(sweepResources, sdk.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping KMS Keys (%s): %w", region, err)
	}

	return nil
}
