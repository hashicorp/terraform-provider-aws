//go:build sweep
// +build sweep

package kms

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func init() {
	resource.AddTestSweepers("aws_kms_key", &resource.Sweeper{
		Name: "aws_kms_key",
		F:    sweepKeys,
	})
}

func sweepKeys(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	input := &kms.ListKeysInput{
		Limit: aws.Int64(1000),
	}
	conn := client.(*conns.AWSClient).KMSConn()
	var sweeperErrs *multierror.Error
	sweepResources := make([]sweep.Sweepable, 0)

	err = conn.ListKeysPagesWithContext(ctx, input, func(page *kms.ListKeysOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Keys {
			keyID := aws.StringValue(v.KeyId)
			key, err := FindKeyByID(ctx, conn, keyID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error reading KMS Key (%s): %w", keyID, err))
				continue
			}

			if aws.StringValue(key.KeyManager) == kms.KeyManagerTypeAws {
				log.Printf("[DEBUG] Skipping KMS Key (%s): managed by AWS", keyID)
				continue
			}
			if aws.StringValue(key.KeyState) == kms.KeyStatePendingDeletion {
				log.Printf("[DEBUG] Skipping KMS Key (%s): pending deletion", keyID)
				continue
			}

			r := ResourceKey()
			d := r.Data(nil)
			d.SetId(keyID)
			d.Set("key_id", keyID)
			d.Set("deletion_window_in_days", "7")

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping KMS Key sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing KMS Keys (%s): %w", region, err))
	}

	err = sweep.SweepOrchestratorWithContext(ctx, sweepResources)

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error sweeping KMS Keys (%s): %w", region, err))
	}

	return sweeperErrs.ErrorOrNil()
}
