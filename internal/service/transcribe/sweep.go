//go:build sweep
// +build sweep

package transcribe

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_transcribe_vocabulary_filter", &resource.Sweeper{
		Name: "aws_transcribe_vocabulary_filter",
		F:    sweepVocabularyFilters,
		Dependencies: []string{
			"aws_s3_bucket",
		},
	})
}

func sweepVocabularyFilters(region string) error {
	client, err := sweep.SharedRegionalASweepClient(region)
	if err != nil {
		fmt.Errorf("error getting client: %s", err)
	}

	return nil
}
