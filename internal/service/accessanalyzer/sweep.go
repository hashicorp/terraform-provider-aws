//go:build sweep
// +build sweep

package accessanalyzer

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/accessanalyzer"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_accessanalyzer_analyzer", &resource.Sweeper{
		Name: "aws_accessanalyzer_analyzer",
		F:    sweepAnalyzers,
	})
}

func sweepAnalyzers(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).AccessAnalyzerConn()
	input := &accessanalyzer.ListAnalyzersInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = conn.ListAnalyzersPagesWithContext(ctx, input, func(page *accessanalyzer.ListAnalyzersOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, analyzer := range page.Analyzers {
			r := ResourceAnalyzer()
			d := r.Data(nil)
			d.SetId(aws.StringValue(analyzer.Name))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Access Analyzer Analyzer sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing Access Analyzer Analyzers (%s): %w", region, err)
	}

	err = sweep.SweepOrchestratorWithContext(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Access Analyzer Analyzers (%s): %w", region, err)
	}

	return nil
}
