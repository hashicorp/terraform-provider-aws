//go:build sweep
// +build sweep

package codegurureviewer

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codegurureviewer"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_codegurureviewer", &resource.Sweeper{
		Name: "aws_codegurureviewer",
		F:    sweepAssociations,
	})
}

func sweepAssociations(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	input := &codegurureviewer.ListRepositoryAssociationsInput{}
	conn := client.(*conns.AWSClient).CodeGuruReviewerConn()

	sweepResources := make([]sweep.Sweepable, 0)

	err = conn.ListRepositoryAssociationsPagesWithContext(ctx, input, func(page *codegurureviewer.ListRepositoryAssociationsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.RepositoryAssociationSummaries {
			r := ResourceRepositoryAssociation()
			d := r.Data(nil)

			d.SetId(aws.StringValue(v.Name))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping CodeGuruReviewer Association sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing CodeGuruReviewer Associations (%s): %w", region, err)
	}

	err = sweep.SweepOrchestratorWithContext(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping CodeGuruReviewer Associations (%s): %w", region, err)
	}

	return nil
}
