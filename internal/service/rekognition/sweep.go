//go:build sweep
// +build sweep

package rekognition

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rekognition"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_rekognition_collection", &resource.Sweeper{
		Name: "aws_rekognition_collection",
		F:    sweepCollections,
	})
}

func sweepCollections(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.(*conns.AWSClient).RekognitionConn()
	sweepResources := make([]sweep.Sweepable, 0)
	var errs *multierror.Error

	err = conn.ListCollectionsPages(&rekognition.ListCollectionsInput{}, func(resp *rekognition.ListCollectionsOutput, lastPage bool) bool {
		if len(resp.CollectionIds) == 0 {
			log.Print("[DEBUG] No Rekognition Collections to sweep")
			return !lastPage
		}

		for _, c := range resp.CollectionIds {
			r := ResourceCollection()
			d := r.Data(nil)
			d.SetId(aws.StringValue(c))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error describing Rekognition Collections: %w", err))
		// in case work can be done, don't jump out yet
	}

	if err = sweep.SweepOrchestratorWithContext(ctx, sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping Rekognition Collections for %s: %w", region, err))
	}

	if sweep.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping Rekognition Collections sweep for %s: %s", region, err)
		return nil
	}

	return errs.ErrorOrNil()
}
