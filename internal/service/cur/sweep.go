//go:build sweep
// +build sweep

package cur

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	cur "github.com/aws/aws-sdk-go/service/costandusagereportservice"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_cur_report_definition", &resource.Sweeper{
		Name: "aws_cur_report_definition",
		F:    sweepReportDefinitions,
	})
}

func sweepReportDefinitions(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).CURConn()
	input := &cur.DescribeReportDefinitionsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = conn.DescribeReportDefinitionsPagesWithContext(ctx, input, func(page *cur.DescribeReportDefinitionsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, reportDefinition := range page.ReportDefinitions {
			r := ResourceReportDefinition()
			d := r.Data(nil)
			d.SetId(aws.StringValue(reportDefinition.ReportName))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping EC2 Cost And Usage Report Definition sweep for %s: %s", region, err)
		return nil
	}
	if err != nil {
		return fmt.Errorf("error listing Cost And Usage Report Definitions (%s): %w", region, err)
	}

	err = sweep.SweepOrchestratorWithContext(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Cost And Usage Report Definitions (%s): %w", region, err)
	}

	return nil
}
