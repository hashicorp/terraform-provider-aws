//go:build sweep
// +build sweep

package kms

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_kms_key", &resource.Sweeper{
		Name: "aws_kms_key",
		F:    sweepKeys,
	})
}

func sweepKeys(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.(*conns.AWSClient).KMSConn

	err = conn.ListKeysPages(&kms.ListKeysInput{Limit: aws.Int64(1000)}, func(out *kms.ListKeysOutput, lastPage bool) bool {
		for _, k := range out.Keys {
			kKeyId := aws.StringValue(k.KeyId)
			kOut, err := conn.DescribeKey(&kms.DescribeKeyInput{
				KeyId: k.KeyId,
			})
			if err != nil {
				log.Printf("Error: Failed to describe key %q: %s", kKeyId, err)
				return false
			}
			if aws.StringValue(kOut.KeyMetadata.KeyManager) == kms.KeyManagerTypeAws {
				// Skip (default) keys which are managed by AWS
				continue
			}
			if aws.StringValue(kOut.KeyMetadata.KeyState) == kms.KeyStatePendingDeletion {
				// Skip keys which are already scheduled for deletion
				continue
			}

			r := ResourceKey()
			d := r.Data(nil)
			d.SetId(kKeyId)
			d.Set("key_id", kKeyId)
			d.Set("deletion_window_in_days", "7")
			err = r.Delete(d, client)
			if err != nil {
				log.Printf("Error: Failed to schedule key %q for deletion: %s", kKeyId, err)
				return false
			}
		}
		return !lastPage
	})
	if err != nil {
		if sweep.SkipSweepError(err) {
			log.Printf("[WARN] Skipping KMS Key sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error describing KMS keys: %w", err)
	}

	return nil
}
